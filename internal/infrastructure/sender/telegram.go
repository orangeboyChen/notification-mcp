package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
)

// TelegramSender implements the NotificationSender interface for Telegram.
type TelegramSender struct {
	config config.TelegramConfig
	client *http.Client
}

// NewTelegramSender creates a new TelegramSender.
func NewTelegramSender(cfg config.TelegramConfig) *TelegramSender {
	return &TelegramSender{
		config: cfg,
		client: &http.Client{},
	}
}

// Channel returns the channel type.
func (t *TelegramSender) Channel() domain.Channel {
	return domain.ChannelTelegram
}

// IsEnabled returns whether the Telegram sender is properly configured.
func (t *TelegramSender) IsEnabled() bool {
	return t.config.Enabled && t.config.Token != "" && t.config.ChatID != ""
}

// DefaultRecipient returns the configured default chat ID.
func (t *TelegramSender) DefaultRecipient() string {
	return t.config.ChatID
}

// GetChannelConfig returns the parameter schema for Telegram.
func (t *TelegramSender) GetChannelConfig() domain.ChannelConfig {
	return domain.ChannelConfig{
		Channel:          domain.ChannelTelegram,
		Enabled:          t.IsEnabled(),
		DefaultRecipient: t.config.ChatID,
		SupportedParams: map[string]domain.ChannelParamConfig{
			"body": {
				Required:    true,
				Description: "Message content, supports Markdown format",
			},
			"title": {
				Required:    false,
				Description: "Bold title prepended to message body",
			},
			"metadata": {
				Required:    false,
				Description: "Optional key-value pairs, e.g. {\"parse_mode\": \"MarkdownV2\"}",
			},
		},
	}
}

// Send delivers a notification via Telegram Bot API.
func (t *TelegramSender) Send(ctx context.Context, notification *domain.Notification) error {
	if !t.IsEnabled() {
		return domain.NewDomainError(domain.ErrCodeChannelDisabled, "telegram channel is not enabled")
	}

	// Always use the configured ChatID (recipient is server-controlled)
	chatID := t.config.ChatID

	// Format the message content
	text := notification.FormattedContent()

	// Build the API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.config.Token)

	// Build request parameters
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("parse_mode", "Markdown")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to send telegram message", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to read response", err)
	}

	// Parse response
	var result telegramResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeSendFailed, "failed to parse telegram response", err)
	}

	if !result.OK {
		return domain.NewDomainError(domain.ErrCodeSendFailed,
			fmt.Sprintf("telegram API error (code %d): %s", result.ErrorCode, result.Description))
	}

	return nil
}

// telegramResponse represents the Telegram Bot API response.
type telegramResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}
