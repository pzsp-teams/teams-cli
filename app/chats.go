package app

import (
	chatsHandlers "github.com/pzsp-teams/cli/internal/handlers/chats"
	chatsMessagesHandlers "github.com/pzsp-teams/cli/internal/handlers/chats/messages"
)

var (
	chatsLong = `Commands for interacting with Microsoft Teams chats`

	chatsListLong = `List all chats you have access to.

Examples:
  # List all chats
  teams-cli chats list`

	chatsMessagesLong = `Commands for retrieving and managing messages in chats`

	chatsMessagesGetLong = `Retrieve messages from all chats you have access to within the specified time range.

Examples:
  # Last 24 hours (default)
  teams-cli chats messages get

  # From 2 hours ago till now
  teams-cli chats messages get --start "2 hours ago"

  # Specific time window
  teams-cli chats messages get --start "2024-01-01 10:00" --end "2024-01-01 11:00"

  # Yesterday
  teams-cli chats messages get --start yesterday --end now
  
  # Filter by chat
  teams-cli chats messages get --chat-ref "<chat-reference>"`

	chatsMessagesSendLong = `Send messages to one or more Teams chats using templates, raw strings, or text files.

Examples:
  # Send templated messages
  teams-cli chats messages send --template msg.txt --data recipients.yaml

  # Send raw message to specific chats
  teams-cli chats messages send --message "Hello!" --chats user1@domain.com,user2@domain.com

  # Send message from file
  teams-cli chats messages send --message-file msg.txt --chats user@domain.com

  # Dry run to preview
  teams-cli chats messages send --template msg.txt --data recipients.yaml --dry-run`
)

func init() {
	chatsCmd := CommandDef{
		Use:   "chats",
		Short: "Manage chats",
		Long:  chatsLong,
		SubCommands: []CommandDef{
			{
				Use:     "list",
				Short:   "List all chats",
				Long:    chatsListLong,
				Handler: chatsHandlers.ListChats,
			},
			{
				Use:   "messages",
				Short: "Manage chat messages",
				Long:  chatsMessagesLong,
				SubCommands: []CommandDef{
					{
						Use:   "get",
						Short: "Get messages",
						Long:  chatsMessagesGetLong,
						Flags: []FlagDef{
							{Name: "start", Usage: "Start time", Type: InputDate},
							{Name: "end", Usage: "End time", Type: InputDate},
							{Name: "file", Usage: "Output file", Type: InputFile},
							{
								Name:       "chat-ref",
								Usage:      "Chat reference to filter messages",
								Type:       InputString,
								DefaultVal: "",
							},
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
						Long:  chatsMessagesSendLong,
						Flags: []FlagDef{
							{Name: "chats", Usage: "Chats list", Type: InputList},
							{
								Name:          "message",
								Usage:         "Message content",
								Type:          InputLongString,
								RequiresFlags: []string{"chats"},
								ConflictsWith: []string{"message-file", "template"},
							},
							{
								Name:          "message-file",
								Usage:         "Message file",
								Type:          InputFile,
								RequiresFlags: []string{"chats"},
								ConflictsWith: []string{"message", "template"},
							},
							{
								Name:          "template",
								Usage:         "Template file",
								Type:          InputFile,
								RequiresFlags: []string{"data"},
								ConflictsWith: []string{"message", "message-file", "chats"},
							},
							{
								Name:          "data",
								Usage:         "Template data",
								Type:          InputFile,
								RequiresFlags: []string{"template"},
								ConflictsWith: []string{"message", "message-file", "chats"},
							},
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
