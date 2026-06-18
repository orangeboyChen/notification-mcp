package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Channel represents a notification delivery channel.
type Channel string

// Supported notification channels.
const (
	ChannelTelegram Channel = "telegram"
	ChannelEmail    Channel = "email"
	ChannelBark     Channel = "bark"
)

// ValidChannels returns all supported channels.
func ValidChannels() []Channel {
	return []Channel{ChannelTelegram, ChannelEmail, ChannelBark}
}

// IsValid checks if the channel is a supported channel.
func (c Channel) IsValid() bool {
	for _, valid := range ValidChannels() {
		if c == valid {
			return true
		}
	}
	return false
}

// Notification represents the core domain entity for a notification message.
type Notification struct {
	ID        string
	Channel   Channel
	Recipient string
	Title     string
	Body      string
	Metadata  map[string]interface{}
	CreatedAt time.Time
}

// NewNotification creates a new Notification with validation.
func NewNotification(channel Channel, recipient, title, body string, metadata map[string]interface{}) (*Notification, error) {
	if !channel.IsValid() {
		return nil, NewDomainError(ErrCodeInvalidChannel, "unsupported channel: "+string(channel))
	}
	if body == "" {
		return nil, NewDomainError(ErrCodeValidation, "body is required")
	}

	return &Notification{
		ID:        generateID(),
		Channel:   channel,
		Recipient: recipient,
		Title:     title,
		Body:      body,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}, nil
}

// SupportsTitle returns whether the channel supports a separate title field.
func (n *Notification) SupportsTitle() bool {
	switch n.Channel {
	case ChannelEmail:
		return true
	case ChannelTelegram:
		return true
	case ChannelBark:
		return true
	default:
		return false
	}
}

// FormattedContent returns the content adapted for the specific channel.
func (n *Notification) FormattedContent() string {
	switch n.Channel {
	case ChannelTelegram:
		if n.Title != "" {
			return "*" + escapeMarkdown(n.Title) + "*\n\n" + n.Body
		}
		return n.Body
	case ChannelEmail:
		return n.Body
	case ChannelBark:
		return n.Body
	default:
		return n.Body
	}
}

// escapeMarkdown escapes special Markdown characters for Telegram.
func escapeMarkdown(s string) string {
	// For MarkdownV2, escape special characters
	replacer := []struct {
		old, new string
	}{
		{"_", "\\_"},
		{"*", "\\*"},
		{"[", "\\["},
		{"]", "\\]"},
		{"(", "\\("},
		{")", "\\)"},
		{"~", "\\~"},
		{"`", "\\`"},
		{">", "\\>"},
		{"#", "\\#"},
		{"+", "\\+"},
		{"-", "\\-"},
		{"=", "\\="},
		{"|", "\\|"},
		{"{", "\\{"},
		{"}", "\\}"},
		{".", "\\."},
		{"!", "\\!"},
	}
	result := s
	for _, r := range replacer {
		result = replaceAll(result, r.old, r.new)
	}
	return result
}

func replaceAll(s, old, new string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result = append(result, []byte(new)...)
			i += len(old) - 1
		} else {
			result = append(result, s[i])
		}
	}
	return string(result)
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomSuffix()
}

func randomSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based suffix
		return fmt.Sprintf("%08x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
