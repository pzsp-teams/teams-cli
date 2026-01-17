Create teams from file

### Synopsis

Create multiple Teams from a data file (YAML/JSON/CSV).

The data file should contain team definitions with display names, descriptions, owners, and members.

### Data file format

YAML example:

```yaml
engineering:
  description: "Engineering team"
  owners:
    - user.one@example.com
  members:
    - user.two@example.com
ops:
  description: "Operations team"
  visibility: "public"
  owners:
    - user.three@example.com
  members: []
```

JSON example:

```json
{
  "engineering": {
    "description": "Engineering team",
    "owners": ["user.one@example.com"],
    "members": ["user.two@example.com"]
  },
  "ops": {
    "description": "Operations team",
    "visibility": "public",
    "owners": ["user.three@example.com"],
    "members": []
  }
}
```

TOML example:

```toml
[engineering]
description = "Engineering team"
owners = ["user.one@example.com"]
members = ["user.two@example.com"]

[ops]
description = "Operations team"
visibility = "public"
owners = ["user.three@example.com"]
members = []
```

### Usage

```
teams-cli teams create [flags]
```

### Options

```
      --data string   Path to teams data file (YAML/JSON/CSV)
      --dry-run       Preview without creating teams
  -h, --help          help for create
```

### Examples

#### Create teams from YAML file

```bash
teams-cli teams create --data teams.yaml
```

#### Create teams from JSON file

```bash
teams-cli teams create --data teams.json
```

#### Dry run to preview

```bash
teams-cli teams create --data teams.yaml --dry-run
```

### TUI Preview

![Teams create TUI](../previews/teams-create.png)

