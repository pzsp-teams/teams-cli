Reply to message

### Synopsis

Send a reply to a specific message in a Teams channel using raw strings or text files.

### Usage

```
teams-cli channels messages reply [flags]
```

### Options

```
      --channel string        Channel name/ID
  -h, --help                  help for reply
      --message string        Reply content
      --message-file string   Reply file
      --message-id string     Message ID
      --team string           Team name/ID
```

### Examples

#### Send a reply with raw message

```bash
teams-cli channels messages reply --team MyTeam --channel General --message-id 123456 --message "Thanks for the update!"
```

#### Send a reply from file

```bash
teams-cli channels messages reply --team MyTeam --channel General --message-id 123456 --message-file reply.txt
```

### TUI Preview

![Channels messages reply TUI](../previews/channels-messages-reply-1.png)

![Channels messages reply TUI (alternate)](../previews/channels-messages-reply-2.png)

