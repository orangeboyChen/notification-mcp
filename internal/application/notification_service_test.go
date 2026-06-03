package application

import (
	"context"
	"testing"

	"github.com/orangeboy/notification-mcp/internal/domain"
)

// mockSender is a test double for domain.NotificationSender.
type mockSender struct {
	channel          domain.Channel
	enabled          bool
	sendErr          error
	sendCalls        int
	defaultRecipient string
}

func (m *mockSender) Send(_ context.Context, _ *domain.Notification) error {
	m.sendCalls++
	return m.sendErr
}

func (m *mockSender) Channel() domain.Channel {
	return m.channel
}

func (m *mockSender) IsEnabled() bool {
	return m.enabled
}

func (m *mockSender) DefaultRecipient() string {
	if m.defaultRecipient != "" {
		return m.defaultRecipient
	}
	return "default-recipient"
}

func (m *mockSender) GetChannelConfig() domain.ChannelConfig {
	return domain.ChannelConfig{
		Channel:          m.channel,
		Enabled:          m.enabled,
		DefaultRecipient: m.DefaultRecipient(),
		SupportedParams: map[string]domain.ChannelParamConfig{
			"body": {Required: true, Description: "Message content"},
		},
	}
}

func TestSendNotification_Success(t *testing.T) {
	mock := &mockSender{
		channel: domain.ChannelTelegram,
		enabled: true,
	}

	svc := NewNotificationService([]domain.NotificationSender{mock})

	resp := svc.SendNotification(context.Background(), &SendNotificationRequest{
		Channel: "telegram",
		Title:   "Test",
		Body:    "Hello, World!",
	})

	if !resp.Success {
		t.Errorf("expected success, got error: %s", resp.Error)
	}
	if resp.MessageID == "" {
		t.Error("expected message ID to be set")
	}
	if mock.sendCalls != 1 {
		t.Errorf("expected 1 send call, got %d", mock.sendCalls)
	}
}

func TestSendNotification_InvalidChannel(t *testing.T) {
	svc := NewNotificationService(nil)

	resp := svc.SendNotification(context.Background(), &SendNotificationRequest{
		Channel: "sms",
		Body:    "Hello",
	})

	if resp.Success {
		t.Error("expected failure for invalid channel")
	}
	if resp.ErrorCode != domain.ErrCodeInvalidChannel {
		t.Errorf("expected error code %s, got %s", domain.ErrCodeInvalidChannel, resp.ErrorCode)
	}
}

func TestSendNotification_ChannelDisabled(t *testing.T) {
	mock := &mockSender{
		channel: domain.ChannelEmail,
		enabled: false,
	}

	svc := NewNotificationService([]domain.NotificationSender{mock})

	resp := svc.SendNotification(context.Background(), &SendNotificationRequest{
		Channel: "email",
		Body:    "Hello",
	})

	if resp.Success {
		t.Error("expected failure for disabled channel")
	}
	if resp.ErrorCode != domain.ErrCodeChannelDisabled {
		t.Errorf("expected error code %s, got %s", domain.ErrCodeChannelDisabled, resp.ErrorCode)
	}
}

func TestSendNotification_SendError(t *testing.T) {
	mock := &mockSender{
		channel: domain.ChannelTelegram,
		enabled: true,
		sendErr: domain.NewDomainError(domain.ErrCodeNetworkError, "connection timeout"),
	}

	svc := NewNotificationService([]domain.NotificationSender{mock})

	resp := svc.SendNotification(context.Background(), &SendNotificationRequest{
		Channel: "telegram",
		Body:    "Hello",
	})

	if resp.Success {
		t.Error("expected failure on send error")
	}
	if resp.ErrorCode != domain.ErrCodeNetworkError {
		t.Errorf("expected error code %s, got %s", domain.ErrCodeNetworkError, resp.ErrorCode)
	}
}

func TestSendNotification_EmptyBody(t *testing.T) {
	svc := NewNotificationService(nil)

	resp := svc.SendNotification(context.Background(), &SendNotificationRequest{
		Channel: "telegram",
		Body:    "",
	})

	if resp.Success {
		t.Error("expected failure for empty body")
	}
	if resp.ErrorCode != domain.ErrCodeValidation {
		t.Errorf("expected error code %s, got %s", domain.ErrCodeValidation, resp.ErrorCode)
	}
}

func TestGetEnabledChannels(t *testing.T) {
	senders := []domain.NotificationSender{
		&mockSender{channel: domain.ChannelTelegram, enabled: true},
		&mockSender{channel: domain.ChannelEmail, enabled: false},
	}

	svc := NewNotificationService(senders)
	channels := svc.GetEnabledChannels()

	if len(channels) != 1 {
		t.Errorf("expected 1 enabled channel, got %d", len(channels))
	}
	if channels[0] != "telegram" {
		t.Errorf("expected telegram, got %s", channels[0])
	}
}
