# Notification MCP

Notification MCP exposes Telegram, Email, and Bark notifications as MCP tools.
Recipients are configured on the server to prevent clients from sending to arbitrary destinations.

## Quick Start

```bash
docker pull ghcr.io/orangeboychen/notification-mcp:{version}
docker run -d --env-file .env -p 3000:3000 ghcr.io/orangeboychen/notification-mcp:{version}
```

MCP endpoint:

```text
http://localhost:3000/mcp
```

Health check:

```bash
curl http://localhost:3000/health
```

## MCP Client Configuration

```json
{
  "mcpServers": {
    "notification": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

## Environment Variables

At least one notification channel must be enabled.

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_PORT` | `3000` | MCP HTTP server port. Also serves `/health` |
| `MCP_AUTH_TOKEN` | - | Optional MCP authentication token. Empty means authentication is disabled |

### Telegram

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_ENABLED` | `false` | Enable Telegram channel |
| `TELEGRAM_BOT_TOKEN` | - | Telegram Bot API token |
| `TELEGRAM_CHAT_ID` | - | Server-controlled Telegram chat ID |

### Email

| Variable | Default | Description |
|----------|---------|-------------|
| `EMAIL_ENABLED` | `false` | Enable Email channel |
| `EMAIL_SMTP_HOST` | - | SMTP server hostname |
| `EMAIL_SMTP_PORT` | `587` | SMTP server port |
| `EMAIL_SMTP_USERNAME` | - | SMTP authentication username |
| `EMAIL_SMTP_PASSWORD` | - | SMTP authentication password |
| `EMAIL_FROM` | - | Sender email address |
| `EMAIL_FROM_NAME` | - | Default sender display name. Can be overridden per message with `metadata.from_name` |
| `EMAIL_TO` | `EMAIL_FROM` | Recipient email address |
| `EMAIL_USE_TLS` | `true` | Use TLS for SMTP |

### Bark

The service sends Bark notifications to the official JSON endpoint:

```text
POST https://api.day.app/push
```

| Variable | Default | Description |
|----------|---------|-------------|
| `BARK_ENABLED` | `false` | Enable Bark channel |
| `BARK_DEVICE_KEY` | - | Bark device key. Use comma-separated keys for batch push, e.g. `key1,key2,key3` |

## Tools

### `send-notification`

Send a notification through `telegram`, `email`, or `bark`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `channel` | string | yes | Channel name: `telegram`, `email`, or `bark` |
| `body` | string | yes | Message body |
| `title` | string | no | Message title or email subject |
| `metadata` | object | no | Channel-specific options |

Email supports:

| Metadata key | Description |
|--------------|-------------|
| `format` | Set to `html` to send HTML email |
| `from_name` | Sender display name for this message. Overrides `EMAIL_FROM_NAME` |

Bark supports:

| Metadata key | Description |
|--------------|-------------|
| `subtitle` | Push subtitle |
| `markdown` | Markdown body. Bark ignores `body` when this is set |
| `level` | Interruption level: `critical`, `active`, `timeSensitive`, or `passive` |
| `volume` | Critical alert volume, `0` to `10` |
| `badge` | App badge number |
| `call` | Set to `"1"` to repeat the notification ringtone |
| `autoCopy` | Set to `"1"` to copy push content automatically where iOS allows it |
| `copy` | Text copied from the notification |
| `sound` | Bark notification sound |
| `icon` | Custom icon URL |
| `image` | Push image URL |
| `group` | Notification group |
| `ciphertext` | Encrypted push ciphertext |
| `isArchive` | Set to `1` to save push history, other values disable saving |
| `ttl` | Archive retention time in seconds |
| `url` | URL opened when the notification is tapped |
| `action` | Set to `"alert"` to show an action popup in Bark |
| `id` | Same ID updates an existing notification |
| `delete` | Set to `"1"` with `id` to delete the notification |

`device_key` and `device_keys` are always derived from `BARK_DEVICE_KEY`.

Example:

```json
{
  "channel": "bark",
  "title": "Deploy Complete",
  "body": "All services are healthy.",
  "metadata": {
    "group": "deployments",
    "sound": "minuet",
    "badge": 1,
    "url": "https://example.com/deploys/123"
  }
}
```

### `get-channel-config`

Returns registered channels, enabled state, server-configured recipient, and supported parameters.

```json
{
  "channel": "optional channel name"
}
```

## Local Development

```bash
go build -o notification-mcp ./cmd/server
./notification-mcp
```

Run checks before committing:

```bash
make check
```

## License

MIT
