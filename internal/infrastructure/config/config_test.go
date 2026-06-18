package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Telegram.Enabled {
		t.Error("telegram should be disabled by default")
	}
	if cfg.Email.Enabled {
		t.Error("email should be disabled by default")
	}
	if cfg.Bark.Enabled {
		t.Error("bark should be disabled by default")
	}
	if cfg.Email.Port != 587 {
		t.Errorf("expected default port 587, got %d", cfg.Email.Port)
	}
	if cfg.Server.MCPPort != 3000 {
		t.Errorf("expected default MCP port 3000, got %d", cfg.Server.MCPPort)
	}
}

func TestValidate_NoChannelsEnabled(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{Enabled: false},
		Email:    EmailConfig{Enabled: false},
		Bark:     BarkConfig{Enabled: false},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Error("expected validation error when no channels enabled")
	}
}

func TestValidate_TelegramMissingToken(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{
			Enabled: true,
			Token:   "",
			ChatID:  "123",
		},
	}

	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Error() == "TELEGRAM_BOT_TOKEN is required when Telegram is enabled" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for missing telegram token")
	}
}

func TestValidate_EmailMissingHost(t *testing.T) {
	cfg := &Config{
		Email: EmailConfig{
			Enabled:  true,
			Host:     "",
			Port:     587,
			Username: "user",
			Password: "pass",
			From:     "from@test.com",
		},
	}

	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Error() == "EMAIL_SMTP_HOST is required when Email is enabled" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for missing email host")
	}
}

func TestLoad_BarkDeviceKey(t *testing.T) {
	os.Clearenv()
	t.Setenv("BARK_DEVICE_KEY", "key1,key2")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Bark.DeviceKey != "key1,key2" {
		t.Errorf("unexpected bark device key: %s", cfg.Bark.DeviceKey)
	}
}

func TestValidate_BarkMissingDeviceKey(t *testing.T) {
	cfg := &Config{
		Bark: BarkConfig{
			Enabled: true,
		},
	}

	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Error() == "BARK_DEVICE_KEY is required when Bark is enabled" {
			found = true
		}
	}
	if !found {
		t.Error("expected validation error for missing bark device key")
	}
}

func TestEnabledChannels(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{Enabled: true},
		Email:    EmailConfig{Enabled: true},
		Bark:     BarkConfig{Enabled: true, DeviceKey: "key"},
	}

	channels := cfg.EnabledChannels()
	if len(channels) != 3 {
		t.Errorf("expected 3 channels, got %d", len(channels))
	}
}
