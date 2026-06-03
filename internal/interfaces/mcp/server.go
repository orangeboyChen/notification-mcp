package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/orangeboy/notification-mcp/internal/application"
)

// Server wraps the MCP server and exposes notification tools.
type Server struct {
	mcpServer   *server.MCPServer
	notifySvc   *application.NotificationService
	authToken   string
	authEnabled bool
}

// NewServer creates a new MCP notification server.
func NewServer(notifySvc *application.NotificationService, authToken string) *Server {
	s := &Server{
		notifySvc:   notifySvc,
		authToken:   authToken,
		authEnabled: authToken != "",
	}

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		"notification-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Register tools
	s.registerTools()

	return s
}

// MCPServer returns the underlying MCP server instance.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// registerTools registers all MCP tools.
func (s *Server) registerTools() {
	// send-notification tool
	sendNotificationTool := mcp.NewTool(
		"send-notification",
		mcp.WithDescription("Send a notification through a specified channel. Recipients are configured server-side to prevent abuse."),
		mcp.WithString("channel",
			mcp.Required(),
			mcp.Description("The notification channel to use (telegram, email)"),
		),
		mcp.WithString("body",
			mcp.Required(),
			mcp.Description("The message body content"),
		),
		mcp.WithString("title",
			mcp.Description("Optional message title/subject"),
		),
		mcp.WithObject("metadata",
			mcp.Description("Optional metadata key-value pairs (e.g., {\"format\": \"html\"})"),
		),
	)

	s.mcpServer.AddTool(sendNotificationTool, s.handleSendNotification)

	// get-channel-config tool
	getChannelConfigTool := mcp.NewTool(
		"get-channel-config",
		mcp.WithDescription("Get the supported channels, their parameter schema, and configuration. Use this to discover available channels before sending."),
		mcp.WithString("channel",
			mcp.Description("Optional: specify a channel name to get config for only that channel. If omitted, returns all channels."),
		),
	)

	s.mcpServer.AddTool(getChannelConfigTool, s.handleGetChannelConfig)
}

// handleSendNotification handles the send-notification tool call.
func (s *Server) handleSendNotification(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters using SDK helper methods
	channel := request.GetString("channel", "")
	title := request.GetString("title", "")
	body := request.GetString("body", "")

	// Extract metadata from arguments map
	var metadata map[string]string
	args := request.GetArguments()
	if args != nil {
		if metaRaw, ok := args["metadata"]; ok && metaRaw != nil {
			if metaMap, ok := metaRaw.(map[string]interface{}); ok {
				metadata = make(map[string]string)
				for k, v := range metaMap {
					metadata[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	// Build request — recipient is always resolved from server config
	req := &application.SendNotificationRequest{
		Channel:  channel,
		Title:    title,
		Body:     body,
		Metadata: metadata,
	}

	// Send notification
	log.Printf("MCP tool call: send-notification channel=%s", channel)
	resp := s.notifySvc.SendNotification(ctx, req)

	// Format response
	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response"), nil
	}

	if !resp.Success {
		return mcp.NewToolResultError(string(respJSON)), nil
	}

	return mcp.NewToolResultText(string(respJSON)), nil
}

// handleGetChannelConfig handles the get-channel-config tool call.
func (s *Server) handleGetChannelConfig(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelFilter := request.GetString("channel", "")

	configs := s.notifySvc.GetChannelConfigs()

	// Filter by channel if specified
	if channelFilter != "" {
		var filtered []interface{}
		for _, cfg := range configs {
			if string(cfg.Channel) == channelFilter {
				filtered = append(filtered, cfg)
			}
		}
		if len(filtered) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("channel '%s' not found or not registered", channelFilter)), nil
		}
		respJSON, err := json.MarshalIndent(filtered[0], "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to marshal response"), nil
		}
		return mcp.NewToolResultText(string(respJSON)), nil
	}

	// Return all channel configs
	result := map[string]interface{}{
		"channels": configs,
		"total":    len(configs),
	}

	respJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal response"), nil
	}

	return mcp.NewToolResultText(string(respJSON)), nil
}
