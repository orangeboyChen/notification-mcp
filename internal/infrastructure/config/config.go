package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	Telegram TelegramConfig
	Email    EmailConfig
	Auth     AuthConfig
	Server   ServerConfig
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	Enabled bool
	Token   string
	ChatID  string
}

// EmailConfig holds SMTP email configuration.
type EmailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       string
	UseTLS   bool
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled bool
	Token   string
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	MCPPort int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Telegram: TelegramConfig{
			Enabled: getEnvBool("TELEGRAM_ENABLED", false),
			Token:   os.Getenv("TELEGRAM_BOT_TOKEN"),
			ChatID:  os.Getenv("TELEGRAM_CHAT_ID"),
		},
		Email: EmailConfig{
			Enabled:  getEnvBool("EMAIL_ENABLED", false),
			Host:     os.Getenv("EMAIL_SMTP_HOST"),
			Port:     getEnvInt("EMAIL_SMTP_PORT", 587),
			Username: os.Getenv("EMAIL_SMTP_USERNAME"),
			Password: os.Getenv("EMAIL_SMTP_PASSWORD"),
			From:     os.Getenv("EMAIL_FROM"),
			To:       os.Getenv("EMAIL_TO"),
			UseTLS:   getEnvBool("EMAIL_USE_TLS", true),
		},
		Auth: AuthConfig{
			Enabled: os.Getenv("MCP_AUTH_TOKEN") != "",
			Token:   os.Getenv("MCP_AUTH_TOKEN"),
		},
		Server: ServerConfig{
			MCPPort: getEnvInt("MCP_PORT", 3000),
		},
	}

	return cfg, nil
}

// Validate checks the configuration for completeness based on enabled channels.
func (c *Config) Validate() []error {
	var errs []error

	if c.Telegram.Enabled {
		if c.Telegram.Token == "" {
			errs = append(errs, fmt.Errorf("TELEGRAM_BOT_TOKEN is required when Telegram is enabled"))
		}
		if c.Telegram.ChatID == "" {
			errs = append(errs, fmt.Errorf("TELEGRAM_CHAT_ID is required when Telegram is enabled"))
		}
	}

	if c.Email.Enabled {
		if c.Email.Host == "" {
			errs = append(errs, fmt.Errorf("EMAIL_SMTP_HOST is required when Email is enabled"))
		}
		if c.Email.Username == "" {
			errs = append(errs, fmt.Errorf("EMAIL_SMTP_USERNAME is required when Email is enabled"))
		}
		if c.Email.Password == "" {
			errs = append(errs, fmt.Errorf("EMAIL_SMTP_PASSWORD is required when Email is enabled"))
		}
		if c.Email.From == "" {
			errs = append(errs, fmt.Errorf("EMAIL_FROM is required when Email is enabled"))
		}
		if c.Email.Port <= 0 || c.Email.Port > 65535 {
			errs = append(errs, fmt.Errorf("EMAIL_SMTP_PORT must be between 1 and 65535"))
		}
	}

	if !c.Telegram.Enabled && !c.Email.Enabled {
		errs = append(errs, fmt.Errorf("at least one notification channel must be enabled"))
	}

	return errs
}

// EnabledChannels returns a list of enabled channel names.
func (c *Config) EnabledChannels() []string {
	var channels []string
	if c.Telegram.Enabled {
		channels = append(channels, "telegram")
	}
	if c.Email.Enabled {
		channels = append(channels, "email")
	}
	return channels
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return strings.ToLower(val) == "true" || val == "1"
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
