package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	APIURL    string `json:"APIURL"`
	APIKey    string `json:"APIKey"`
	BudgetKey string `json:"BudgetKey"`
}

// DefaultAPIURL is the default AgentPMT API endpoint
const DefaultAPIURL = "https://api.agentpmt.com"

// Load reads configuration from config.json and applies environment variable overrides
func Load() (*Config, error) {
	cfg := &Config{}

	// Try to load from config.json in the same directory as the executable
	configPath, err := findConfigFile()
	if err == nil {
		if err := loadFromFile(configPath, cfg); err != nil {
			// Config file exists but is invalid
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}
	// If config file doesn't exist, that's okay - we'll use env vars and defaults

	// Apply environment variable overrides (non-breaking)
	if v := os.Getenv("AGENTPMT_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("AGENTPMT_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("AGENTPMT_BUDGET_KEY"); v != "" {
		cfg.BudgetKey = v
	}

	// Set default API URL if still empty
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("APIKey is required (set in config.json or AGENTPMT_API_KEY env var)")
	}
	if cfg.BudgetKey == "" {
		return nil, fmt.Errorf("BudgetKey is required (set in config.json or AGENTPMT_BUDGET_KEY env var)")
	}

	return cfg, nil
}

// findConfigFile locates config.json relative to the executable or current directory
func findConfigFile() (string, error) {
	// Try current directory first
	if _, err := os.Stat("config.json"); err == nil {
		return "config.json", nil
	}

	// Try executable directory
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := filepath.Join(exeDir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("config.json not found")
}

// loadFromFile reads and parses config.json
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("invalid JSON in config file: %w", err)
	}

	return nil
}

// Sanitize returns a copy of the config with secrets redacted (for logging)
func (c *Config) Sanitize() *Config {
	return &Config{
		APIURL:    c.APIURL,
		APIKey:    redact(c.APIKey),
		BudgetKey: redact(c.BudgetKey),
	}
}

// redact masks a secret value for logging
func redact(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}
