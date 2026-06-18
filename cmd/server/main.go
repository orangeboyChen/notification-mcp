package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mark3labs/mcp-go/server"

	"github.com/orangeboy/notification-mcp/internal/application"
	"github.com/orangeboy/notification-mcp/internal/domain"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/config"
	"github.com/orangeboy/notification-mcp/internal/infrastructure/sender"
	mcpserver "github.com/orangeboy/notification-mcp/internal/interfaces/mcp"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting notification-mcp service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if errs := cfg.Validate(); len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Configuration error: %v", e)
		}
		log.Fatalf("Configuration validation failed with %d error(s)", len(errs))
	}

	log.Printf("Enabled channels: %s", strings.Join(cfg.EnabledChannels(), ", "))

	// Initialize senders
	var senders []domain.NotificationSender

	if cfg.Telegram.Enabled {
		telegramSender := sender.NewTelegramSender(cfg.Telegram)
		senders = append(senders, telegramSender)
		log.Println("Telegram sender initialized")
	}

	if cfg.Email.Enabled {
		emailSender := sender.NewEmailSender(cfg.Email)
		senders = append(senders, emailSender)
		log.Println("Email sender initialized")
	}

	if cfg.Bark.Enabled {
		barkSender := sender.NewBarkSender(cfg.Bark)
		senders = append(senders, barkSender)
		log.Println("Bark sender initialized")
	}

	// Create application service
	notifySvc := application.NewNotificationService(senders)

	// Create MCP server
	mcpSrv := mcpserver.NewServer(notifySvc, cfg.Auth.Token)

	// Create StreamableHTTP transport for remote MCP
	httpServer := server.NewStreamableHTTPServer(mcpSrv.MCPServer())

	// Set up HTTP mux with health check and MCP endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"healthy"}`)
	})

	// Apply auth middleware to MCP endpoint
	mcpHandler := mcpserver.AuthMiddleware(cfg.Auth.Token)(httpServer)
	if cfg.Auth.Enabled {
		log.Println("MCP authentication enabled")
	} else {
		log.Println("MCP authentication disabled (no MCP_AUTH_TOKEN configured)")
	}
	mux.Handle("/mcp", mcpHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.MCPPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("MCP server listening on %s (endpoint: /mcp)", addr)
	log.Printf("Health check available at %s/health", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
