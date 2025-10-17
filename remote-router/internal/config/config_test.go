package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("AGENTPMT_API_URL", "https://test.api.com")
	os.Setenv("AGENTPMT_API_KEY", "test-api-key")
	os.Setenv("AGENTPMT_BUDGET_KEY", "test-budget-key")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.APIURL != "https://test.api.com" {
		t.Errorf("Expected APIURL to be https://test.api.com, got %s", cfg.APIURL)
	}
	if cfg.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be test-api-key, got %s", cfg.APIKey)
	}
	if cfg.BudgetKey != "test-budget-key" {
		t.Errorf("Expected BudgetKey to be test-budget-key, got %s", cfg.BudgetKey)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	os.Setenv("AGENTPMT_API_KEY", "test-api-key")
	os.Setenv("AGENTPMT_BUDGET_KEY", "test-budget-key")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.APIURL != DefaultAPIURL {
		t.Errorf("Expected APIURL to be %s, got %s", DefaultAPIURL, cfg.APIURL)
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configContent := `{
		"APIURL": "https://file.api.com",
		"APIKey": "file-api-key",
		"BudgetKey": "file-budget-key"
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.APIURL != "https://file.api.com" {
		t.Errorf("Expected APIURL from file, got %s", cfg.APIURL)
	}
	if cfg.APIKey != "file-api-key" {
		t.Errorf("Expected APIKey from file, got %s", cfg.APIKey)
	}
}

func TestEnvOverridesConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configContent := `{
		"APIURL": "https://file.api.com",
		"APIKey": "file-api-key",
		"BudgetKey": "file-budget-key"
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Set env vars to override
	os.Setenv("AGENTPMT_API_URL", "https://env.api.com")
	os.Setenv("AGENTPMT_API_KEY", "env-api-key")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Env vars should override file
	if cfg.APIURL != "https://env.api.com" {
		t.Errorf("Expected env var to override, got %s", cfg.APIURL)
	}
	if cfg.APIKey != "env-api-key" {
		t.Errorf("Expected env var to override, got %s", cfg.APIKey)
	}
	// This one shouldn't be overridden
	if cfg.BudgetKey != "file-budget-key" {
		t.Errorf("Expected budget key from file, got %s", cfg.BudgetKey)
	}
}

func TestSanitize(t *testing.T) {
	cfg := &Config{
		APIURL:    "https://api.example.com",
		APIKey:    "secret-api-key-1234",
		BudgetKey: "secret-budget-key-5678",
	}

	sanitized := cfg.Sanitize()

	if sanitized.APIURL != cfg.APIURL {
		t.Error("APIURL should not be redacted")
	}
	if sanitized.APIKey == cfg.APIKey {
		t.Error("APIKey should be redacted")
	}
	if sanitized.BudgetKey == cfg.BudgetKey {
		t.Error("BudgetKey should be redacted")
	}
	if sanitized.APIKey != "secr***1234" {
		t.Errorf("Expected redacted format, got %s", sanitized.APIKey)
	}
}

func TestMissingRequiredFields(t *testing.T) {
	// No env vars set, no config file
	os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when APIKey is missing")
	}

	// Set only API key
	os.Setenv("AGENTPMT_API_KEY", "test-key")
	_, err = Load()
	if err == nil {
		t.Error("Expected error when BudgetKey is missing")
	}
}
