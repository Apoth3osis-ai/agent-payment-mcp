package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Apoth3osis-ai/agent-payment-mcp/remote-router/internal/api"
	"github.com/Apoth3osis-ai/agent-payment-mcp/remote-router/internal/config"
	"github.com/Apoth3osis-ai/agent-payment-mcp/remote-router/internal/mcp"
)

var Version = "dev" // Set by -ldflags at build time

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nPlease ensure:\n")
		fmt.Fprintf(os.Stderr, "  1. config.json exists next to the binary, OR\n")
		fmt.Fprintf(os.Stderr, "  2. Environment variables are set:\n")
		fmt.Fprintf(os.Stderr, "     AGENTPMT_API_KEY\n")
		fmt.Fprintf(os.Stderr, "     AGENTPMT_BUDGET_KEY\n")
		fmt.Fprintf(os.Stderr, "     AGENTPMT_API_URL (optional, defaults to https://api.agentpmt.com)\n")
		os.Exit(1)
	}

	// Setup logging with secret redaction
	mcp.SetupLogging(cfg.APIKey, cfg.BudgetKey)

	log.Printf("AgentPMT MCP Router v%s starting...", Version)
	log.Printf("API URL: %s", cfg.APIURL)
	log.Printf("Configuration loaded (keys: %s, %s)", redact(cfg.APIKey), redact(cfg.BudgetKey))

	// Create API client
	apiClient := api.NewClient(cfg.APIURL, cfg.APIKey, cfg.BudgetKey)

	// Create MCP server
	server := mcp.NewServer(apiClient, Version)

	log.Printf("MCP server ready, listening on stdio...")

	// Run stdio transport (blocks until stdin closes)
	if err := server.HandleStdioTransport(); err != nil {
		log.Fatalf("Transport error: %v", err)
	}

	log.Printf("MCP server shutting down")
}

// redact masks a secret for display
func redact(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}
