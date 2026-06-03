package domain

import "context"

// ChannelParamConfig describes a parameter for a channel.
type ChannelParamConfig struct {
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// ChannelConfig describes the configuration and capabilities of a channel.
type ChannelConfig struct {
	Channel          Channel                       `json:"channel"`
	Enabled          bool                          `json:"enabled"`
	DefaultRecipient string                        `json:"default_recipient,omitempty"`
	SupportedParams  map[string]ChannelParamConfig `json:"supported_params"`
}

// NotificationSender defines the port for sending notifications through a specific channel.
type NotificationSender interface {
	// Send delivers the notification through the channel.
	Send(ctx context.Context, notification *Notification) error

	// Channel returns the channel this sender handles.
	Channel() Channel

	// IsEnabled returns whether this sender is properly configured and enabled.
	IsEnabled() bool

	// DefaultRecipient returns the configured default recipient for this channel.
	DefaultRecipient() string

	// GetChannelConfig returns the parameter schema and configuration for this channel.
	GetChannelConfig() ChannelConfig
}

// NotificationResult represents the result of a notification send operation.
type NotificationResult struct {
	Success   bool
	MessageID string
	Channel   Channel
	Error     *Error
}

// NewSuccessResult creates a successful notification result.
func NewSuccessResult(channel Channel, messageID string) *NotificationResult {
	return &NotificationResult{
		Success:   true,
		MessageID: messageID,
		Channel:   channel,
	}
}

// NewFailureResult creates a failed notification result.
func NewFailureResult(channel Channel, err *Error) *NotificationResult {
	return &NotificationResult{
		Success: false,
		Channel: channel,
		Error:   err,
	}
}
