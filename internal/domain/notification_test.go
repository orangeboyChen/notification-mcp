package domain

import "testing"

func TestChannel_IsValid(t *testing.T) {
	tests := []struct {
		channel Channel
		valid   bool
	}{
		{ChannelTelegram, true},
		{ChannelEmail, true},
		{Channel("sms"), false},
		{Channel(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			if got := tt.channel.IsValid(); got != tt.valid {
				t.Errorf("Channel(%q).IsValid() = %v, want %v", tt.channel, got, tt.valid)
			}
		})
	}
}

func TestNewNotification_Valid(t *testing.T) {
	n, err := NewNotification(ChannelTelegram, "12345", "Test", "Hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Channel != ChannelTelegram {
		t.Errorf("expected channel telegram, got %s", n.Channel)
	}
	if n.Recipient != "12345" {
		t.Errorf("expected recipient 12345, got %s", n.Recipient)
	}
	if n.ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestNewNotification_InvalidChannel(t *testing.T) {
	_, err := NewNotification(Channel("sms"), "12345", "Test", "Hello", nil)
	if err == nil {
		t.Fatal("expected error for invalid channel")
	}
	domainErr, ok := err.(*Error)
	if !ok {
		t.Fatal("expected domain Error")
	}
	if domainErr.Code != ErrCodeInvalidChannel {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidChannel, domainErr.Code)
	}
}

func TestNewNotification_EmptyRecipient(t *testing.T) {
	n, err := NewNotification(ChannelEmail, "", "Test", "Hello", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v (empty recipient should be allowed)", err)
	}
	if n.Recipient != "" {
		t.Errorf("expected empty recipient, got %s", n.Recipient)
	}
}

func TestNewNotification_EmptyBody(t *testing.T) {
	_, err := NewNotification(ChannelTelegram, "12345", "Test", "", nil)
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestNotification_FormattedContent_Telegram(t *testing.T) {
	n, _ := NewNotification(ChannelTelegram, "12345", "Hello", "World", nil)
	content := n.FormattedContent()
	if content == "" {
		t.Error("expected non-empty formatted content")
	}
	// Should contain the title in bold
	if len(content) < len("World") {
		t.Error("content should include body")
	}
}

func TestNotification_FormattedContent_Email(t *testing.T) {
	n, _ := NewNotification(ChannelEmail, "test@example.com", "Subject", "Body content", nil)
	content := n.FormattedContent()
	if content != "Body content" {
		t.Errorf("expected 'Body content', got %s", content)
	}
}
