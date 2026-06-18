package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
)

// BarkSender implements the NotificationSender interface for Bark push.
type BarkSender struct {
	config    config.BarkConfig
	client    *http.Client
	serverURL string
}

const defaultBarkServerURL = "https://api.day.app"

// NewBarkSender creates a new BarkSender.
func NewBarkSender(cfg config.BarkConfig) *BarkSender {
	return &BarkSender{
		config:    cfg,
		client:    &http.Client{},
		serverURL: defaultBarkServerURL,
	}
}

// Channel returns the channel type.
func (b *BarkSender) Channel() domain.Channel {
	return domain.ChannelBark
}

// IsEnabled returns whether the Bark sender is properly configured.
func (b *BarkSender) IsEnabled() bool {
	return b.config.Enabled && b.config.DeviceKey != ""
}

// DefaultRecipient returns the configured Bark recipient key or key count.
func (b *BarkSender) DefaultRecipient() string {
	deviceKeys := b.deviceKeys()
	if len(deviceKeys) > 1 {
		return fmt.Sprintf("%d device keys", len(deviceKeys))
	}
	if len(deviceKeys) == 1 {
		return deviceKeys[0]
	}
	return ""
}

// GetChannelConfig returns the parameter schema for Bark.
func (b *BarkSender) GetChannelConfig() domain.ChannelConfig {
	return domain.ChannelConfig{
		Channel:          domain.ChannelBark,
		Enabled:          b.IsEnabled(),
		DefaultRecipient: b.DefaultRecipient(),
		SupportedParams: map[string]domain.ChannelParamConfig{
			"body": {
				Required:    true,
				Description: "Bark notification body content",
			},
			"title": {
				Required:    false,
				Description: "Bark notification title",
			},
			"metadata": {
				Required:    false,
				Description: "Optional Bark parameters: subtitle, markdown, level, volume, badge, call, autoCopy, copy, sound, icon, image, group, ciphertext, isArchive, ttl, url, action, id, delete",
			},
		},
	}
}

// Send delivers a notification via Bark push.
func (b *BarkSender) Send(ctx context.Context, notification *domain.Notification) error {
	if !b.IsEnabled() {
		return domain.NewDomainError(domain.ErrCodeChannelDisabled, "bark channel is not enabled")
	}

	payload := barkPayload{
		"body": notification.Body,
	}
	if notification.Title != "" {
		payload["title"] = notification.Title
	}
	for key, value := range notification.Metadata {
		if !isAllowedBarkMetadataParam(key) {
			continue
		}
		payload[key] = value
	}
	deviceKeys := b.deviceKeys()
	if len(deviceKeys) > 1 {
		payload["device_keys"] = deviceKeys
	} else {
		payload["device_key"] = deviceKeys[0]
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeValidation, "failed to encode bark payload", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(b.serverURL, "/")+"/push", bytes.NewReader(requestBody))
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to create bark request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to send bark notification", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeNetworkError, "failed to read bark response", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.NewDomainError(domain.ErrCodeSendFailed,
			fmt.Sprintf("bark webhook error (status %d): %s", resp.StatusCode, string(body)))
	}

	var result barkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeSendFailed, "failed to parse bark response", err)
	}

	if result.Code != http.StatusOK {
		message := result.Message
		if message == "" {
			message = string(body)
		}
		return domain.NewDomainError(domain.ErrCodeSendFailed,
			fmt.Sprintf("bark webhook error (code %d): %s", result.Code, message))
	}

	return nil
}

type barkPayload map[string]interface{}

func (b *BarkSender) deviceKeys() []string {
	parts := strings.Split(b.config.DeviceKey, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

func isAllowedBarkMetadataParam(key string) bool {
	_, ok := allowedBarkMetadataParams[key]
	return ok
}

var allowedBarkMetadataParams = map[string]struct{}{
	"subtitle":   {},
	"markdown":   {},
	"level":      {},
	"volume":     {},
	"badge":      {},
	"call":       {},
	"autoCopy":   {},
	"copy":       {},
	"sound":      {},
	"icon":       {},
	"image":      {},
	"group":      {},
	"ciphertext": {},
	"isArchive":  {},
	"ttl":        {},
	"url":        {},
	"action":     {},
	"id":         {},
	"delete":     {},
}

// barkResponse represents the Bark webhook response.
type barkResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
