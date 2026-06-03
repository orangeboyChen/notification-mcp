# AGENTS.md - MCP Client Integration Guide

## Overview

This service implements the Model Context Protocol (MCP) and exposes notification capabilities as MCP tools. Any MCP-compatible client (Claude Desktop, Cursor, etc.) can integrate with this service.

## Integration

### Claude Desktop Configuration

First start the notification-mcp server:

```bash
# Set environment variables and run
export TELEGRAM_ENABLED=true
export TELEGRAM_BOT_TOKEN=your-bot-token
export TELEGRAM_CHAT_ID=your-chat-id
./notification-mcp
```

Then add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "notification": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

### Docker-based Integration

```bash
# Start the remote MCP server
docker run -d --env-file .env -p 3000:3000 ghcr.io/orangeboy/notification-mcp:latest
```

Then configure your MCP client to connect to the remote endpoint:

```json
{
  "mcpServers": {
    "notification": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

## Available Tools

### send-notification

Sends a notification through a specified channel. Recipients are configured server-side to prevent abuse.

**Input Schema:**
```json
{
  "channel": "telegram | email",
  "body": "message content",
  "title": "optional title",
  "metadata": {"format": "html"}
}
```

**Example:**
```json
{
  "channel": "telegram",
  "body": "Deployment complete!"
}
```

**Success Response:**
```json
{
  "success": true,
  "message_id": "20240101120000-abc12345",
  "channel": "telegram"
}
```

**Error Response:**
```json
{
  "success": false,
  "channel": "telegram",
  "error": "channel telegram is not enabled or not properly configured",
  "error_code": "CHANNEL_DISABLED"
}
```

### get-channel-config

Returns the supported channels, their parameter schema, and configuration. Use this to discover available channels before calling `send-notification`.

**Input Schema:**
```json
{
  "channel": "optional: filter by channel name"
}
```

**Response:**
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
    }
  ],
  "total": 1
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_CHANNEL` | The specified channel is not supported |
| `VALIDATION_ERROR` | Request parameters failed validation |
| `CHANNEL_DISABLED` | Channel is not enabled or configured |
| `CONFIG_MISSING` | Required configuration is missing |
| `SEND_FAILED` | Message delivery failed |
| `NETWORK_ERROR` | Network connectivity issue |
| `AUTH_ERROR` | Authentication failed |
| `RATE_LIMIT` | Rate limit exceeded |

## Authentication

If `MCP_AUTH_TOKEN` is set, clients must provide the token for authentication. When not set, authentication is disabled.

## Health Check

The service exposes a health endpoint on the same port as the MCP server (configurable via `MCP_PORT`, default `3000`).

```bash
curl http://localhost:3000/health
# {"status":"healthy"}
```
