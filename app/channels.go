package app

import (
	channelsHandlers "github.com/pzsp-teams/cli/internal/handlers/channels"
	channelsMessagesHandlers "github.com/pzsp-teams/cli/internal/handlers/channels/messages"
)

func init() {
	channelsCmd := CommandDef{
		Use:   "channels",
		Short: "Manage channels",
		SubCommands: []CommandDef{
			{
				Use:   "create",
				Short: "Create channels from file",
				Flags: []FlagDef{
					{Name: "team", Usage: "Name of the team", Type: InputString, Required: true},
					{Name: "data", Usage: "Path to channels data file (YAML/JSON/TOML/CSV)", Type: InputFile, Required: true},
					{Name: "ensure-in-channels", Usage: "Ensure members are in channels", Type: InputBool},
					{Name: "ensure-in-team", Usage: "Ensure members are in team", Type: InputBool},
					{Name: "dry-run", Usage: "Preview only", Type: InputBool},
				},
				Handler: channelsHandlers.CreateChannels,
			},
			{
				Use:   "messages",
				Short: "Manage channel messages",
				SubCommands: []CommandDef{
					{
						Use:   "get",
						Short: "Get messages",
						Flags: []FlagDef{
							{Name: "start", Usage: "Start time", Type: InputDate},
							{Name: "end", Usage: "End time", Type: InputDate},
							{Name: "file", Usage: "Output file", Type: InputFile},
							{
								Name:       "team-ref",
								Usage:      "Team reference to filter messages",
								Type:       InputString,
								DefaultVal: "",
							},
							{
								Name:       "channel-ref",
								Usage:      "Channel reference to filter messages",
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
						Handler: channelsMessagesHandlers.GetMessages,
					},
					{
						Use:   "send",
						Short: "Send message",
						Flags: []FlagDef{
							{Name: "team", Usage: "Team name/ID", Type: InputString, Required: true},
							{
								Name:          "channels",
								Usage:         "Channels list",
								Type:          InputList,
								ConflictsWith: []string{"template", "data"},
							},
							{
								Name:          "message",
								Usage:         "Message content",
								Type:          InputLongString,
								ConflictsWith: []string{"message-file", "template", "data"},
								RequiresFlags: []string{"channels"},
							},
							{
								Name:          "message-file",
								Usage:         "Message file",
								Type:          InputFile,
								ConflictsWith: []string{"message", "template", "data"},
								RequiresFlags: []string{"channels"},
							},
							{
								Name:          "template",
								Usage:         "Template file",
								Type:          InputFile,
								RequiresFlags: []string{"data"},
								ConflictsWith: []string{"message", "message-file", "channels"},
							},
							{
								Name:          "data",
								Usage:         "Template data",
								Type:          InputFile,
								RequiresFlags: []string{"template"},
								ConflictsWith: []string{"message", "message-file", "channels"},
							},
							{Name: "dry-run", Usage: "Preview only", Type: InputBool},
							{Name: "ignore-errors", Usage: "Continue on error", Type: InputBool},
						},
						Handler: channelsMessagesHandlers.SendMessages,
					},
					{
						Use:   "reply",
						Short: "Reply to message",
						Flags: []FlagDef{
							{Name: "team", Usage: "Team name/ID", Type: InputString, Required: true},
							{Name: "channel", Usage: "Channel name/ID", Type: InputString, Required: true},
							{Name: "message-id", Usage: "Message ID", Type: InputString, Required: true},
							{
								Name:          "message",
								Usage:         "Reply content",
								Type:          InputLongString,
								ConflictsWith: []string{"message-file"},
							},
							{
								Name:          "message-file",
								Usage:         "Reply file",
								Type:          InputFile,
								ConflictsWith: []string{"message"},
							},
						},
						Handler: channelsMessagesHandlers.ReplyToMessage,
					},
				},
			},
		},
	}

	Registry = append(Registry, channelsCmd)
}
