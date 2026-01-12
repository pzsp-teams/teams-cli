package app

import (
	channelsHandlers "github.com/pzsp-teams/cli/internal/handlers/channels"
	channelsMessagesHandlers "github.com/pzsp-teams/cli/internal/handlers/channels/messages"
)

var (
	channelsLong = `Commands for interacting with Microsoft Teams channels`

	channelsCreateLong = `Create multiple Teams channels with members from a data file (YAML/JSON/TOML/CSV).

The data file should contain channel definitions with team_ref, channel_ref, role, and user_ref.

Examples:
  # Create channels from YAML file
  teams-cli channels create --team myteam --data channels.yaml

  # Create channels from JSON file
  teams-cli channels create --team myteam --data channels.json

  # Dry run to preview
  teams-cli channels create --team myteam --data channels.yaml --dry-run
  
  # Ensure members are added to channels if they already exist
  teams-cli channels create --team myteam --data channels.yaml --ensure-in-channels

  # Ensure members are memebers of the team
  teams-cli channels create --team myteam --data channels.yaml --ensure-in-team`

	channelsMessagesLong = `Commands for retrieving and managing messages in channels`

	channelsMessagesGetLong = `Retrieve messages from all channels you have access to within the specified time range.

Examples:
  # Last 24 hours (default)
  teams-cli channels messages get

  # From 2 hours ago till now
  teams-cli channels messages get --start "2 hours ago"

  # Specific time window
  teams-cli channels messages get --start "2024-01-01 10:00" --end "2024-01-01 11:00"

  # Yesterday
  teams-cli channels messages get --start yesterday --end now
	
  # Filter by team
  teams-cli channels messages get --team "My Team"

  # Filter by channel and team
  teams-cli channels messages get --team "My Team" --channel "General"`

	channelsMessagesSendLong = `Send messages to one or more Teams channels using templates, raw strings, or text files.

Examples:
  # Send templated messages
  teams-cli channels messages send --template msg.txt --data recipients.yaml --team MyTeam

  # Send raw message to specific channels
  teams-cli channels messages send --message "Hello team!" --channels General,Announcements --team MyTeam

  # Send message from file
  teams-cli channels messages send --message-file msg.txt --channels General --team MyTeam

  # Dry run to preview
  teams-cli channels messages send --template msg.txt --data recipients.yaml --team MyTeam --dry-run`

	channelsMessagesReplyLong = `Send a reply to a specific message in a Teams channel using raw strings or text files.

Examples:
  # Send a reply with raw message
  teams-cli channels messages reply --team MyTeam --channel General --message-id 123456 --message "Thanks for the update!"

  # Send a reply from file
  teams-cli channels messages reply --team MyTeam --channel General --message-id 123456 --message-file reply.txt`
)

func init() {
	channelsCmd := CommandDef{
		Use:   "channels",
		Short: "Manage channels",
		Long:  channelsLong,
		SubCommands: []CommandDef{
			{
				Use:   "create",
				Short: "Create channels from file",
				Long:  channelsCreateLong,
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
				Long:  channelsMessagesLong,
				SubCommands: []CommandDef{
					{
						Use:   "get",
						Short: "Get messages",
						Long:  channelsMessagesGetLong,
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
						Long:  channelsMessagesSendLong,
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
						Long:  channelsMessagesReplyLong,
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