Send message

### Synopsis

Send messages to one or more Teams chats using templates, raw strings, or text files.
See [Message templates](../message-templates/index.md) for template syntax and examples.

### Usage

```
teams-cli chats messages send [flags]
```

### Options

```
      --chats strings         Chats list
      --data string           Template data
      --dry-run               Preview only
  -h, --help                  help for send
      --ignore-errors         Continue on error
      --message string        Message content
      --message-file string   Message file
      --template string       Template file
```

### Examples

#### Send templated messages

```bash
teams-cli chats messages send --template msg.txt --data recipients.yaml
```

#### Send raw message to specific chats

```bash
teams-cli chats messages send --message "Hello!" --chats user1@domain.com,user2@domain.com
```

#### Send message from file

```bash
teams-cli chats messages send --message-file msg.txt --chats user@domain.com
```

#### Dry run to preview

```bash
teams-cli chats messages send --template msg.txt --data recipients.yaml --dry-run
```

### TUI Preview

![Chats messages send TUI](../previews/chats-messges-send-1.png)

![Chats messages send TUI (alternate)](../previews/chats-messages-send-2.png)

![Chats messages send TUI (alternate 2)](../previews/chats-messages-send-3.png)


