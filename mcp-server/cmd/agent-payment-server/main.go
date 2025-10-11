package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/agentpmt/agent-payment-mcp-server/internal/config"
	"github.com/agentpmt/agent-payment-mcp-server/internal/mcp"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var apiKey, budgetKey string

	// Try to load from config.json first (for .mcpb package installations)
	exePath, err := os.Executable()
	if err == nil {
		configPath := filepath.Join(filepath.Dir(exePath), "config.json")
		if _, err := os.Stat(configPath); err == nil {
			cfg, err := config.Load(configPath)
			if err == nil {
				apiKey = cfg.APIKey
				budgetKey = cfg.BudgetKey
				log.Printf("Loaded configuration from %s", configPath)
			}
		}
	}

	// Fall back to environment variables if config.json not found
	if apiKey == "" {
		apiKey = os.Getenv("AGENT_PAYMENT_API_KEY")
	}
	if budgetKey == "" {
		budgetKey = os.Getenv("AGENT_PAYMENT_BUDGET_KEY")
	}

	if apiKey == "" || budgetKey == "" {
		fmt.Fprintln(os.Stderr, "Error: No API credentials found")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Provide credentials via environment variables:")
		fmt.Fprintln(os.Stderr, "  export AGENT_PAYMENT_API_KEY=your-api-key")
		fmt.Fprintln(os.Stderr, "  export AGENT_PAYMENT_BUDGET_KEY=your-budget-key")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or place a config.json file next to the binary:")
		fmt.Fprintln(os.Stderr, `  {"api_key": "your-key", "budget_key": "your-budget-key"}`)
		os.Exit(1)
	}

	// Create server
	server, err := mcp.NewServer(mcp.Config{
		APIKey:    apiKey,
		BudgetKey: budgetKey,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}

	log.Println("Server shutdown complete")
}
