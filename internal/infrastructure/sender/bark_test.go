package sender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
)

func TestBarkSender_SendSuccess(t *testing.T) {
	var got map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/push" {
			t.Errorf("expected request path /push, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"success"}`))
	}))
	defer server.Close()

	sender := NewBarkSender(config.BarkConfig{
		Enabled:   true,
		DeviceKey: "server-configured-key",
	})
	sender.serverURL = server.URL
	notification, err := domain.NewNotification(domain.ChannelBark, "server-configured-key", "Deploy", "Done", map[string]interface{}{
		"group":       "deployments",
		"sound":       "minuet",
		"badge":       float64(1),
		"ttl":         float64(3600),
		"body":        "client-controlled-body",
		"unknown":     "ignored",
		"device_key":  "client-controlled-key",
		"device_keys": `["client-controlled-key"]`,
	})
	if err != nil {
		t.Fatalf("unexpected notification error: %v", err)
	}

	if err := sender.Send(context.Background(), notification); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if got["title"] != "Deploy" {
		t.Errorf("expected title Deploy, got %s", got["title"])
	}
	if got["body"] != "Done" {
		t.Errorf("expected body Done, got %s", got["body"])
	}
	if got["group"] != "deployments" {
		t.Errorf("expected group deployments, got %s", got["group"])
	}
	if got["sound"] != "minuet" {
		t.Errorf("expected sound minuet, got %s", got["sound"])
	}
	if got["badge"] != float64(1) {
		t.Errorf("expected numeric badge 1, got %#v", got["badge"])
	}
	if got["ttl"] != float64(3600) {
		t.Errorf("expected numeric ttl 3600, got %#v", got["ttl"])
	}
	if _, ok := got["unknown"]; ok {
		t.Error("unknown Bark metadata parameter should be ignored")
	}
	if got["device_key"] != "server-configured-key" {
		t.Error("metadata must not override server-configured Bark device key")
	}
	if _, ok := got["device_keys"]; ok {
		t.Error("metadata must not override server-configured Bark device keys")
	}
}

func TestBarkSender_SendBatchSuccess(t *testing.T) {
	var got map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"message":"success"}`))
	}))
	defer server.Close()

	sender := NewBarkSender(config.BarkConfig{
		Enabled:   true,
		DeviceKey: "key1, key2",
	})
	sender.serverURL = server.URL
	notification, err := domain.NewNotification(domain.ChannelBark, "2 device keys", "", "Done", nil)
	if err != nil {
		t.Fatalf("unexpected notification error: %v", err)
	}

	if err := sender.Send(context.Background(), notification); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if _, ok := got["device_key"]; ok {
		t.Error("single device key should be ignored when batch keys are configured")
	}
	deviceKeys, ok := got["device_keys"].([]interface{})
	if !ok {
		t.Fatalf("expected device_keys array, got %T", got["device_keys"])
	}
	if len(deviceKeys) != 2 || deviceKeys[0] != "key1" || deviceKeys[1] != "key2" {
		t.Errorf("unexpected device_keys: %#v", deviceKeys)
	}
}

func TestBarkSender_SendAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":400,"message":"bad request"}`))
	}))
	defer server.Close()

	sender := NewBarkSender(config.BarkConfig{
		Enabled:   true,
		DeviceKey: "server-configured-key",
	})
	sender.serverURL = server.URL
	notification, err := domain.NewNotification(domain.ChannelBark, "server-configured-key", "", "Done", nil)
	if err != nil {
		t.Fatalf("unexpected notification error: %v", err)
	}

	err = sender.Send(context.Background(), notification)
	if err == nil {
		t.Fatal("expected send error")
	}

	domainErr, ok := err.(*domain.Error)
	if !ok {
		t.Fatalf("expected domain error, got %T", err)
	}
	if domainErr.Code != domain.ErrCodeSendFailed {
		t.Errorf("expected error code %s, got %s", domain.ErrCodeSendFailed, domainErr.Code)
	}
}
