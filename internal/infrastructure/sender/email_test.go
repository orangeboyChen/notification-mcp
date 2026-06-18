package sender

import (
	"testing"

	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
)

func TestEmailSender_BuildMessageUsesConfiguredFromName(t *testing.T) {
	sender := NewEmailSender(config.EmailConfig{
		From:     "sender@example.com",
		FromName: "Default Bot",
	})
	notification, err := domain.NewNotification(domain.ChannelEmail, "to@example.com", "Subject", "Body", nil)
	if err != nil {
		t.Fatalf("unexpected notification error: %v", err)
	}

	message := sender.buildMessage(notification)
	from := message.GetHeader("From")

	if len(from) != 1 {
		t.Fatalf("expected one From header, got %d", len(from))
	}
	if from[0] != `"Default Bot" <sender@example.com>` {
		t.Errorf("unexpected From header: %s", from[0])
	}
}

func TestEmailSender_BuildMessageMetadataOverridesFromName(t *testing.T) {
	sender := NewEmailSender(config.EmailConfig{
		From:     "sender@example.com",
		FromName: "Default Bot",
	})
	notification, err := domain.NewNotification(domain.ChannelEmail, "to@example.com", "Subject", "Body", map[string]interface{}{
		"from_name": "Deploy Bot",
	})
	if err != nil {
		t.Fatalf("unexpected notification error: %v", err)
	}

	message := sender.buildMessage(notification)
	from := message.GetHeader("From")

	if len(from) != 1 {
		t.Fatalf("expected one From header, got %d", len(from))
	}
	if from[0] != `"Deploy Bot" <sender@example.com>` {
		t.Errorf("unexpected From header: %s", from[0])
	}
}
