package app

import (
	chatsHandlers "github.com/pzsp-teams/cli/internal/handlers/chats"
	chatsMessagesHandlers "github.com/pzsp-teams/cli/internal/handlers/chats/messages"
)

func init() {
	chatsCmd := CommandDef{
		Use:   "chats",
		Short: "Manage chats",
		SubCommands: []CommandDef{
			{
				Use:     "list",
				Short:   "List all chats",
				Handler: chatsHandlers.ListChats,
			},
			{
				Use:   "messages",
				Short: "Manage chat messages",
				SubCommands: []CommandDef{
					{
						Use:   "get",
						Short: "Get messages",
						Flags: []FlagDef{
							{Name: "start", Usage: "Start time", Type: InputDate},
							{Name: "end", Usage: "End time", Type: InputDate},
							{Name: "file", Usage: "Output file", Type: InputFile},
							{
								Name:       "format",
								Usage:      "Output format",
								Type:       InputChoice,
								Options:    []string{"plain", "markdown"},
								DefaultVal: "plain",
							},
						},
						Handler: chatsMessagesHandlers.GetMessages,
					},
					{
						Use:   "send",
						Short: "Send message",
						Flags: []FlagDef{
							{Name: "chats", Usage: "Chats list", Type: InputString},
							{
								Name:          "message",
								Usage:         "Message content",
								Type:          InputString,
								ConflictsWith: []string{"message-file", "template"},
							},
							{
								Name:          "message-file",
								Usage:         "Message file",
								Type:          InputFile,
								ConflictsWith: []string{"message", "template"},
							},
							{
								Name:          "template",
								Usage:         "Template file",
								Type:          InputFile,
								RequiresFlags: []string{"data"},
								ConflictsWith: []string{"message", "message-file"},
							},
							{Name: "data", Usage: "Template data", Type: InputFile},
							{Name: "dry-run", Usage: "Preview only", Type: InputBool},
							{Name: "ignore-errors", Usage: "Continue on error", Type: InputBool},
						},
						Handler: chatsMessagesHandlers.SendMessages,
					},
				},
			},
		},
	}

	Registry = append(Registry, chatsCmd)
}
