# Notification MCP

A multi-channel notification service built with the Model Context Protocol (MCP), following Domain-Driven Design (DDD) architecture principles.

## Features

- 🔔 **Multi-channel notifications** - Send messages via Telegram and Email
- 🏗️ **DDD Architecture** - Clean separation of concerns with domain, application, infrastructure, and interface layers
- 🔌 **MCP Protocol** - Standard Model Context Protocol interface for AI assistant integration
- 🐳 **Docker Ready** - Multi-stage build with health checks
- 🔒 **Optional Authentication** - Token-based MCP authentication
- ✅ **Configuration Validation** - Startup checks for required settings

## Quick Start

### Prerequisites

- Go 1.21+
- A Telegram Bot Token (from [@BotFather](https://t.me/BotFather)) and/or SMTP credentials

### Installation

```bash
git clone https://github.com/orangeboy/notification-mcp.git
cd notification-mcp
go mod tidy
```

### Configuration

Copy the example environment file and configure your channels:

```bash
cp .env.example .env
# Edit .env with your credentials
```

### Running

```bash
# Build and run
go build -o notification-mcp ./cmd/server
./notification-mcp

# Or run directly
go run ./cmd/server
```

### Docker

```bash
# Build
docker build -t notification-mcp .

# Run
docker run -d --env-file .env -p 3000:3000 notification-mcp
```

The MCP server will be available at `http://localhost:3000/mcp`.

## MCP Tools

### `send-notification`

Send a notification through a specified channel. Recipients are configured server-side to prevent abuse.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `channel` | string | ✅ | Channel to use (`telegram`, `email`) |
| `body` | string | ✅ | Message body content |
| `title` | string | ❌ | Message title/subject |
| `metadata` | object | ❌ | Extra options (e.g., `{"format": "html"}`) |

**Example:**
```json
{
  "channel": "telegram",
  "body": "Service v1.2.3 deployed successfully to production."
}
```

**Example (HTML email):**
```json
{
  "channel": "email",
  "body": "<h1>Deploy Complete</h1><p>All services healthy.</p>",
  "title": "Deploy Complete",
  "metadata": {"format": "html"}
}
```

### `get-channel-config`

Get the supported channels, their parameter schema, and configuration. Use this to discover available channels before sending.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `channel` | string | ❌ | Filter by channel name. If omitted, returns all channels |

**Returns:**
```json
{
  "channels": [
    {
      "channel": "telegram",
      "enabled": true,
      "default_recipient": "123456789",
      "supported_params": {
        "body": {"required": true, "description": "Message content, supports Markdown format"},
        "title": {"required": false, "description": "Bold title prepended to message body"},
        "metadata": {"required": false, "description": "Optional key-value pairs"}
      }
    },
    {
      "channel": "email",
      "enabled": true,
      "default_recipient": "team@example.com",
      "supported_params": {
        "body": {"required": true, "description": "Email body content. Supports plain text or HTML"},
        "title": {"required": false, "description": "Email subject line"},
        "metadata": {"required": false, "description": "Optional, e.g. {\"format\": \"html\"}"}
      }
    }
  ],
  "total": 2
}
```

## Architecture

```
notification-mcp/
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Domain models, value objects, ports
│   ├── application/     # Application services (use cases)
│   ├── infrastructure/  # Technical implementations
│   │   ├── config/      # Configuration management
│   │   └── sender/      # Channel sender implementations
│   └── interfaces/      # Interface adapters
│       └── mcp/         # MCP protocol server
├── .env.example         # Configuration template
├── Dockerfile           # Multi-stage Docker build
└── .github/workflows/   # CI/CD pipeline
```

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_ENABLED` | `false` | Enable Telegram channel |
| `TELEGRAM_BOT_TOKEN` | - | Telegram Bot API token |
| `TELEGRAM_CHAT_ID` | - | Default Telegram chat ID |
| `EMAIL_ENABLED` | `false` | Enable Email channel |
| `EMAIL_SMTP_HOST` | - | SMTP server hostname |
| `EMAIL_SMTP_PORT` | `587` | SMTP server port |
| `EMAIL_SMTP_USERNAME` | - | SMTP authentication username |
| `EMAIL_SMTP_PASSWORD` | - | SMTP authentication password |
| `EMAIL_FROM` | - | Sender email address |
| `EMAIL_TO` | - | Recipient email address (defaults to EMAIL_FROM if not set) |
| `EMAIL_USE_TLS` | `true` | Use TLS for SMTP |
| `MCP_AUTH_TOKEN` | - | Optional MCP authentication token |
| `MCP_PORT` | `3000` | MCP HTTP server port (also serves health check) |

## Testing

```bash
go test -v ./...
```

## License

MIT
