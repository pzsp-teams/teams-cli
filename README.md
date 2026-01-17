[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

# Teams CLI

A command-line tool for managing Microsoft Teams teams, channels, chats, and
messages. Includes both CLI and TUI modes for easier workflows.

## Features

- List, create, archive, unarchive, and delete teams.
- Create channels with members from data files.
- Send, reply to, and fetch channel messages.
- List chats and send chat messages (including templates).

## Installation

- Go install: `go install github.com/pzsp-teams/teams-cli@latest`
- Download a binary from the GitHub releases page.

## Configuration

Default config locations:

| OS      | Location                                                       |
| ------- | -------------------------------------------------------------- |
| Linux   | `~/.config/teams-cli`                                          |
| macOS   | `~/Library/Application Support/teams-cli`                      |
| Windows | `%LOCALAPPDATA%\teams-cli` (fallback `LocalAppData\teams-cli`) |

A default config file is created under the config directory on first run.

### Config file structure

```toml
[auth]
  client_id = ""
  tenant_id = ""
  email = ""
  auth_method = "interactive"

[sender]
  max_retries = 3
  next_retry_delay = 2
  timeout = 10

[cache]
  mode = "async"
```

## Documentation

[Detailed docs](https://pzsp-teams.github.io/teams-cli/)
