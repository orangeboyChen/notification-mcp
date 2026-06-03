package sender

import (
	"context"
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
)

// EmailSender implements the NotificationSender interface for email via SMTP.
type EmailSender struct {
	config config.EmailConfig
}

// NewEmailSender creates a new EmailSender.
func NewEmailSender(cfg config.EmailConfig) *EmailSender {
	return &EmailSender{
		config: cfg,
	}
}

// Channel returns the channel type.
func (e *EmailSender) Channel() domain.Channel {
	return domain.ChannelEmail
}

// IsEnabled returns whether the Email sender is properly configured.
func (e *EmailSender) IsEnabled() bool {
	return e.config.Enabled && e.config.Host != "" && e.config.Username != "" && e.config.Password != ""
}

// DefaultRecipient returns the configured default email recipient.
func (e *EmailSender) DefaultRecipient() string {
	if e.config.To != "" {
		return e.config.To
	}
	return e.config.From
}

// GetChannelConfig returns the parameter schema for Email.
func (e *EmailSender) GetChannelConfig() domain.ChannelConfig {
	return domain.ChannelConfig{
		Channel:          domain.ChannelEmail,
		Enabled:          e.IsEnabled(),
		DefaultRecipient: e.DefaultRecipient(),
		SupportedParams: map[string]domain.ChannelParamConfig{
			"body": {
				Required:    true,
				Description: "Email body content. Supports plain text or HTML (set metadata.format to 'html')",
			},
			"title": {
				Required:    false,
				Description: "Email subject line. Defaults to 'Notification' if not provided",
			},
			"metadata": {
				Required:    false,
				Description: "Optional key-value pairs, e.g. {\"format\": \"html\"} for HTML emails",
			},
		},
	}
}

// Send delivers a notification via email.
func (e *EmailSender) Send(_ context.Context, notification *domain.Notification) error {
	if !e.IsEnabled() {
		return domain.NewDomainError(domain.ErrCodeChannelDisabled, "email channel is not enabled")
	}

	// Build the email message
	m := gomail.NewMessage()
	m.SetHeader("From", e.config.From)
	m.SetHeader("To", notification.Recipient)

	// Set subject (use title if available, otherwise use a default)
	subject := notification.Title
	if subject == "" {
		subject = "Notification"
	}
	m.SetHeader("Subject", subject)

	// Determine content type based on metadata
	contentType := "text/plain"
	if notification.Metadata != nil {
		if format, ok := notification.Metadata["format"]; ok && format == "html" {
			contentType = "text/html"
		}
	}
	m.SetBody(contentType, notification.Body)

	// Create dialer
	d := gomail.NewDialer(e.config.Host, e.config.Port, e.config.Username, e.config.Password)
	if e.config.UseTLS {
		d.TLSConfig = &tls.Config{
			ServerName: e.config.Host,
		}
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return domain.NewDomainErrorWithCause(domain.ErrCodeSendFailed,
			fmt.Sprintf("failed to send email to %s", notification.Recipient), err)
	}

	return nil
}
