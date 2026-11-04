package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Autumn-27/artex/internal/artexcontrol/client"
	"github.com/Autumn-27/artex/internal/artexcontrol/control"
	"github.com/Autumn-27/artex/internal/artexcontrol/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds validated startup configuration for artex-mcp.
type Config struct {
	BaseURL      string
	TokenFile    string
	WriteEnabled bool
	AuditFile    string
}

// LoadConfigFromEnv reads and validates configuration from environment variables.
func LoadConfigFromEnv(lookup func(string) string) (*Config, error) {
	baseURL := strings.TrimSpace(lookup("ARTEX_BASE_URL"))
	if baseURL == "" {
		return nil, errors.New("ARTEX_BASE_URL is required")
	}

	tokenFile := strings.TrimSpace(lookup("ARTEX_TOKEN_FILE"))
	if tokenFile == "" {
		return nil, errors.New("ARTEX_TOKEN_FILE is required")
	}

	writeEnabledStr := strings.TrimSpace(lookup("ARTEX_MCP_WRITE_ENABLED"))
	writeEnabled := false
	if writeEnabledStr != "" {
		if writeEnabledStr != "true" && writeEnabledStr != "false" {
			return nil, errors.New("ARTEX_MCP_WRITE_ENABLED must be 'true' or 'false'")
		}
		writeEnabled = writeEnabledStr == "true"
	}

	auditFile := strings.TrimSpace(lookup("ARTEX_MCP_AUDIT_FILE"))
	if writeEnabled {
		if auditFile == "" {
			return nil, errors.New("ARTEX_MCP_AUDIT_FILE is required when write mode is enabled")
		}
		if !filepath.IsAbs(auditFile) {
			return nil, errors.New("ARTEX_MCP_AUDIT_FILE must be an absolute path")
		}
	}

	return &Config{
		BaseURL:      baseURL,
		TokenFile:    tokenFile,
		WriteEnabled: writeEnabled,
		AuditFile:    auditFile,
	}, nil
}

func main() {
	// Stderr logging only; stdout is strictly reserved for JSON-RPC transport
	logger := log.New(os.Stderr, "[artex-mcp] ", log.LstdFlags)

	cfg, err := LoadConfigFromEnv(os.Getenv)
	if err != nil {
		logger.Fatalf("configuration error: %v", err)
	}

	apiClient, err := client.New(client.Options{
		BaseURL:   cfg.BaseURL,
		TokenFile: cfg.TokenFile,
		Timeout:   15 * time.Second,
	})
	if err != nil {
		logger.Fatalf("client initialization error: %v", err)
	}

	svc := control.NewService(apiClient)

	var mgr *control.ChangeManager
	if cfg.WriteEnabled {
		audit, err := control.NewFileAudit(cfg.AuditFile)
		if err != nil {
			logger.Fatalf("audit initialization error: %v", err)
		}
		mgr = control.NewChangeManager(svc, audit, time.Now)
	}

	server := mcpserver.New(svc, mgr, mcpserver.Options{
		WriteEnabled: cfg.WriteEnabled,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Printf("starting ARTEX control MCP companion (write_enabled=%t)", cfg.WriteEnabled)

	transport := &mcp.StdioTransport{}
	if err := server.Run(ctx, transport); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("server exited: %v", err)
		os.Exit(1)
	}

	logger.Println("server cleanly shut down")
}
