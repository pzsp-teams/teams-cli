Create single team

### Synopsis

This command creates a single team based on provided flags and a simplified data file.

### Data file format

CSV example:

```csv
role,user_ref
owner,user.one@example.com
owner,user.two@example.com
```

YAML example:

```yaml
description: "Example team"
owners:
  - user.one@example.com
members:
  - user.two@example.com
```

JSON and TOML formats are also supported with the same fields as the YAML example.

You can also provide the description via the `--description` flag instead of the file.

### Usage

```
teams-cli teams create-single [flags]
```

### Options

```
      --description string   Description of the team
      --dry-run              Preview only
  -f, --file string          Path to members data file
  -h, --help                 help for create-single
  -i, --include-me           Include current user
      --team-name string     Name of the team to create
      --visibility string    Visibility (private/public)
```

### TUI Preview

![Teams create single TUI](../previews/teams-create-simplified.png)

