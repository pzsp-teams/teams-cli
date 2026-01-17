# Message templates

Teams CLI supports templated messages using structured data in JSON, YAML, or
TOML. A parser is automatically selected based on the data file extension.

The parser converts newlines to `<br>` in plain text, but preserves HTML as-is
when HTML tags are detected.

## Mention syntax

Use `@@identifier@@` to create mentions:

- Email: `@@user@domain.com@@`
- Special mentions for bulk sends: `@@channel@@`, `@@team@@`, `@@this@@`,
  `@@everyone@@`

Special mentions are resolved to the receiver automatically in bulk sends.

## Plain text example

Template (`template.txt`):

```
Hello {{.name}}!

Your order #{{.order}} is ready.
Thank you!
```

Data (`data.json`):

```json
{
  "alex": {
    "name": "Alex",
    "order": "12345"
  },
  "blair": {
    "name": "Blair",
    "order": "67890"
  }
}
```

Data (`data.yaml`):

```yaml
alex:
  name: Alex
  order: "12345"
blair:
  name: Blair
  order: "67890"
```

Data (`data.toml`):

```toml
[alex]
name = "Alex"
order = "12345"

[blair]
name = "Blair"
order = "67890"
```

## HTML content example

Template (`template-html.txt`):

```
<p>Hello <b>{{.name}}</b>!</p>

<p>Your order #{{.order}} is ready.</p>
<p>Click <a href="https://example.com/orders/{{.order}}">here</a> for details.</p>
```

## HTML tag detection

The parser treats content as HTML if it finds any of:

- `<p>` and `</p>`
- `<b>` and `</b>`
- `<i>` and `</i>`
- `<br>`
- `<a href="...">` and `</a>`

When HTML tags are detected, newlines are preserved.

## Channel message sample

Template (`channel_message.txt`):

```
Hello {{.name}}!
This is a channel mention @@channel@@
This is a personal mention {{.mention}}
This is a team mention @@team@@

Your order #{{.order}} is ready.
Thank you!
```

Data (`channel_message_data.yaml`):

```yaml
general:
  name: Alex
  order: "12345"
  mention: "@@user.one@example.com@@"
announcements:
  name: Blair
  order: "67890"
  mention: "@@user.two@example.com@@"
```

## Chat message sample

Template (`chat_message.txt`):

```
Hello {{.name}}!
This is a this mention @@this@@

Your order #{{.order}} is ready.
Thank you!
```

Data (`chat_message_data.yaml`):

```yaml
user.one@example.com:
  name: Alex
  order: "12345"
user.two@example.com:
  name: Blair
  order: "67890"
```

## Group chat sample

Template (`group_chat_message.txt`):

```
Hello {{.name}}!
This is a group chat mention @@everyone@@
This is a personal mention {{.mention}}
This is a duplicate group chat mention @@everyone@@

Your order #{{.order}} is ready.
Thank you!
```

Data (`group_chat_message_data.yaml`):

```yaml
project chat:
  name: Alex
  order: "12345"
  mention: "@@user.one@example.com@@"
marketing chat:
  name: Blair
  order: "67890"
  mention: "@@user.two@example.com@@"
```
