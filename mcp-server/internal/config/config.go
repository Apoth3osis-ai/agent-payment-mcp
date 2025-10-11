package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all configuration for the MCP server
type Config struct {
	APIKey     string `json:"api_key"`
	BudgetKey  string `json:"budget_key"`
	APIURL     string `json:"api_url"`
	Auth       string `json:"auth,omitempty"`
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	// If path is empty, look for config.json in executable directory
	if path == "" {
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		path = filepath.Join(filepath.Dir(exePath), "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required in config")
	}
	if cfg.BudgetKey == "" {
		return nil, fmt.Errorf("budget_key is required in config")
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.agentpmt.com"
	}

	return &cfg, nil
}
