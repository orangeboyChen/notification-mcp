package application

import (
	"context"
	"fmt"
	"log"

	"github.com/orangeboy/notification-mcp/internal/domain"
)

// SendNotificationRequest represents the input for sending a notification.
type SendNotificationRequest struct {
	Channel  string
	Title    string
	Body     string
	Metadata map[string]string
}

// SendNotificationResponse represents the output of a notification send operation.
type SendNotificationResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Channel   string `json:"channel"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// NotificationService is the application service for notification operations.
type NotificationService struct {
	senders map[domain.Channel]domain.NotificationSender
}

// NewNotificationService creates a new NotificationService with the given senders.
func NewNotificationService(senders []domain.NotificationSender) *NotificationService {
	senderMap := make(map[domain.Channel]domain.NotificationSender)
	for _, s := range senders {
		senderMap[s.Channel()] = s
	}
	return &NotificationService{
		senders: senderMap,
	}
}

// SendNotification processes a notification send request.
func (s *NotificationService) SendNotification(ctx context.Context, req *SendNotificationRequest) *SendNotificationResponse {
	channel := domain.Channel(req.Channel)

	// Find the appropriate sender
	sender, exists := s.senders[channel]
	if !exists {
		// Still try to create the notification for validation (e.g. invalid channel name)
		_, err := domain.NewNotification(channel, "", req.Title, req.Body, req.Metadata)
		if err != nil {
			domainErr, ok := err.(*domain.Error)
			if ok {
				return &SendNotificationResponse{
					Success:   false,
					Channel:   req.Channel,
					Error:     domainErr.Message,
					ErrorCode: domainErr.Code,
				}
			}
		}
		return &SendNotificationResponse{
			Success:   false,
			Channel:   req.Channel,
			Error:     fmt.Sprintf("no sender registered for channel: %s", req.Channel),
			ErrorCode: domain.ErrCodeInvalidChannel,
		}
	}

	// Resolve recipient from server config (not user-provided, to prevent abuse)
	recipient := sender.DefaultRecipient()

	// Create domain notification entity (includes validation)
	notification, err := domain.NewNotification(channel, recipient, req.Title, req.Body, req.Metadata)
	if err != nil {
		domainErr, ok := err.(*domain.Error)
		if ok {
			return &SendNotificationResponse{
				Success:   false,
				Channel:   req.Channel,
				Error:     domainErr.Message,
				ErrorCode: domainErr.Code,
			}
		}
		return &SendNotificationResponse{
			Success:   false,
			Channel:   req.Channel,
			Error:     err.Error(),
			ErrorCode: domain.ErrCodeValidation,
		}
	}

	// Check if the channel is enabled
	if !sender.IsEnabled() {
		return &SendNotificationResponse{
			Success:   false,
			Channel:   req.Channel,
			Error:     fmt.Sprintf("channel %s is not enabled or not properly configured", req.Channel),
			ErrorCode: domain.ErrCodeChannelDisabled,
		}
	}

	// Send the notification
	log.Printf("Sending notification via %s to %s", channel, recipient)
	if err := sender.Send(ctx, notification); err != nil {
		domainErr, ok := err.(*domain.Error)
		if ok {
			return &SendNotificationResponse{
				Success:   false,
				Channel:   req.Channel,
				Error:     domainErr.Message,
				ErrorCode: domainErr.Code,
			}
		}
		return &SendNotificationResponse{
			Success:   false,
			Channel:   req.Channel,
			Error:     err.Error(),
			ErrorCode: domain.ErrCodeSendFailed,
		}
	}

	log.Printf("Notification sent successfully via %s to %s (ID: %s)", channel, recipient, notification.ID)
	return &SendNotificationResponse{
		Success:   true,
		MessageID: notification.ID,
		Channel:   req.Channel,
	}
}

// GetEnabledChannels returns a list of currently enabled channels.
func (s *NotificationService) GetEnabledChannels() []string {
	var channels []string
	for ch, sender := range s.senders {
		if sender.IsEnabled() {
			channels = append(channels, string(ch))
		}
	}
	return channels
}

// GetChannelConfigs returns the configuration and parameter schema for all registered channels.
func (s *NotificationService) GetChannelConfigs() []domain.ChannelConfig {
	var configs []domain.ChannelConfig
	for _, sender := range s.senders {
		configs = append(configs, sender.GetChannelConfig())
	}
	return configs
}
