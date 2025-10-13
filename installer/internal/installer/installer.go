package installer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed binaries/linux-amd64
var linuxBinary []byte

//go:embed binaries/darwin-amd64
var darwinAmd64Binary []byte

//go:embed binaries/darwin-arm64
var darwinArm64Binary []byte

//go:embed binaries/windows-amd64.exe
var windowsBinary []byte

const (
	installDir = ".agent-payment"
	binaryName = "agent-payment-server"
)

type InstallRequest struct {
	APIKey        string   `json:"apiKey"`
	BudgetKey     string   `json:"budgetKey"`
	Environment   string   `json:"environment"` // "live" or "test"
	SelectedTools []string `json:"selectedTools"`
}

type InstallProgress struct {
	Step     string `json:"step"`
	Message  string `json:"message"`
	Progress int    `json:"progress"` // 0-100
	Error    string `json:"error,omitempty"`
}

type Installer struct{}

func New() *Installer {
	return &Installer{}
}

func (inst *Installer) Install(req InstallRequest) InstallProgress {
	// Validate
	if req.APIKey == "" {
		return InstallProgress{Error: "API Key is required"}
	}
	if req.BudgetKey == "" {
		return InstallProgress{Error: "Budget Key is required"}
	}
	if len(req.SelectedTools) == 0 {
		return InstallProgress{Error: "Please select at least one AI tool"}
	}

	// IMPORTANT: Only Claude Desktop is supported in v1.0.0
	// Reject any other tools to prevent installation attempts
	for _, toolID := range req.SelectedTools {
		if toolID != "claude-desktop" {
			return InstallProgress{Error: fmt.Sprintf("Tool '%s' is not yet supported. Currently only Claude Desktop is available. Other tools coming soon!", toolID)}
		}
	}

	// Step 1: Copy/download binary
	progress := InstallProgress{
		Step:     "download",
		Message:  "Setting up MCP server binary...",
		Progress: 10,
	}

	binaryPath, err := inst.setupBinary()
	if err != nil {
		progress.Error = fmt.Sprintf("Failed to setup binary: %v", err)
		return progress
	}

	// Step 2: Create config.json
	progress.Step = "config"
	progress.Message = "Creating configuration file..."
	progress.Progress = 40

	apiURL := "https://api.agentpmt.com"
	if req.Environment == "test" {
		apiURL = "https://test.api.agentpmt.com"
	}

	err = inst.createConfig(filepath.Dir(binaryPath), req.APIKey, req.BudgetKey, apiURL)
	if err != nil {
		progress.Error = fmt.Sprintf("Failed to create config: %v", err)
		return progress
	}

	// Step 3: Configure tools
	progress.Step = "configure"
	progress.Message = "Configuring AI tools..."
	progress.Progress = 60

	for _, toolID := range req.SelectedTools {
		err := inst.configureTool(toolID, binaryPath)
		if err != nil {
			progress.Error = fmt.Sprintf("Failed to configure %s: %v", toolID, err)
			return progress
		}
	}

	progress.Step = "complete"
	progress.Message = "Installation complete!"
	progress.Progress = 100

	return progress
}

func (inst *Installer) setupBinary() (string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	var binaryFilename string
	var binaryData []byte

	// Select the correct embedded binary
	switch osName {
	case "windows":
		binaryFilename = binaryName + ".exe"
		binaryData = windowsBinary
	case "linux":
		binaryFilename = binaryName
		binaryData = linuxBinary
	case "darwin":
		binaryFilename = binaryName
		if archName == "arm64" {
			binaryData = darwinArm64Binary
		} else {
			binaryData = darwinAmd64Binary
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", osName, archName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	installPath := filepath.Join(home, installDir)
	err = os.MkdirAll(installPath, 0755)
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(installPath, binaryFilename)

	// Write embedded binary to disk
	err = os.WriteFile(binaryPath, binaryData, 0755)
	if err != nil {
		return "", err
	}

	// On Windows, add Defender exclusion and unblock the file
	if osName == "windows" {
		// Add Windows Defender exclusion for the installation folder
		// This prevents Defender from deleting the binary as a false positive
		if err := addWindowsDefenderExclusion(installPath); err != nil {
			// Log but don't fail - user might have to add manually
			fmt.Fprintf(os.Stderr, "Warning: Could not add Windows Defender exclusion: %v\n", err)
			fmt.Fprintf(os.Stderr, "You may need to manually add '%s' to Windows Defender exclusions\n", installPath)
		}

		// Unblock the file to prevent "spawn UNKNOWN" errors
		if err := unblockWindowsFile(binaryPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not unblock file: %v\n", err)
		}
	}

	return binaryPath, nil
}

func (inst *Installer) createConfig(dir, apiKey, budgetKey, apiURL string) error {
	config := map[string]string{
		"api_key":    apiKey,
		"budget_key": budgetKey,
		"api_url":    apiURL,
	}

	configPath := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureTool(toolID, binaryPath string) error {
	switch toolID {
	case "claude-desktop":
		return inst.configureClaudeDesktop(binaryPath)
	case "claude-code":
		return inst.configureClaudeCode(binaryPath)
	case "cursor":
		return inst.configureCursor(binaryPath)
	case "vscode":
		return inst.configureVSCode(binaryPath)
	case "zed":
		return inst.configureZed(binaryPath)
	case "windsurf":
		return inst.configureWindsurf(binaryPath)
	case "jetbrains":
		return nil // GUI configuration required
	default:
		return fmt.Errorf("unknown tool: %s", toolID)
	}
}

func (inst *Installer) configureClaudeDesktop(binaryPath string) error {
	home, _ := os.UserHomeDir()
	var configPath string

	switch runtime.GOOS {
	case "darwin":
		configPath = filepath.Join(home, "Library/Application Support/Claude/claude_desktop_config.json")
	case "linux":
		configPath = filepath.Join(home, ".config/Claude/claude_desktop_config.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Claude/claude_desktop_config.json")
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	os.MkdirAll(filepath.Dir(configPath), 0755)

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureClaudeCode(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".claude.json")

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	// Configure at USER SCOPE (global)
	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureCursor(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".cursor/mcp.json")

	os.MkdirAll(filepath.Dir(configPath), 0755)

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureVSCode(binaryPath string) error {
	home, _ := os.UserHomeDir()
	var configPath string

	switch runtime.GOOS {
	case "darwin":
		configPath = filepath.Join(home, "Library/Application Support/Code/User/mcp.json")
	case "linux":
		configPath = filepath.Join(home, ".config/Code/User/mcp.json")
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), "Code/User/mcp.json")
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	os.MkdirAll(filepath.Dir(configPath), 0755)

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	// VS Code uses "servers" not "mcpServers"
	if config["servers"] == nil {
		config["servers"] = make(map[string]interface{})
	}

	servers := config["servers"].(map[string]interface{})
	servers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureZed(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/zed/settings.json")

	os.MkdirAll(filepath.Dir(configPath), 0755)

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	// Zed uses "context_servers"
	if config["context_servers"] == nil {
		config["context_servers"] = make(map[string]interface{})
	}

	contextServers := config["context_servers"].(map[string]interface{})
	contextServers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (inst *Installer) configureWindsurf(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codeium/windsurf/mcp_config.json")

	os.MkdirAll(filepath.Dir(configPath), 0755)

	config := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["agent-payment"] = map[string]interface{}{
		"command": binaryPath,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func addWindowsDefenderExclusion(path string) error {
	// Add the installation folder to Windows Defender exclusions
	// This prevents false positives where Defender deletes the Go binary
	// Requires admin privileges - will fail silently if not admin
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Add-MpPreference -ExclusionPath '%s'", path))
	return cmd.Run()
}

func unblockWindowsFile(path string) error {
	// Remove the Zone.Identifier alternate data stream that marks files as downloaded
	// This prevents "spawn UNKNOWN" errors when Claude Desktop tries to execute the binary
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Unblock-File -Path '%s'", path))
	return cmd.Run()
}
