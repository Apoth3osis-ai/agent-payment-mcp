package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Tool struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Detected   bool   `json:"detected"`
	ConfigPath string `json:"configPath"`
}

type Detector struct{}

func New() *Detector {
	return &Detector{}
}

func (d *Detector) DetectAll() []Tool {
	return []Tool{
		d.detectClaudeDesktop(),
		d.detectClaudeCode(),
		d.detectCursor(),
		d.detectVSCode(),
		d.detectZed(),
		d.detectWindsurf(),
		d.detectJetBrains(),
	}
}

func (d *Detector) detectClaudeDesktop() Tool {
	tool := Tool{
		ID:   "claude-desktop",
		Name: "Claude Desktop",
	}

	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		if pathExists("/Applications/Claude.app") || pathExists(filepath.Join(home, "Applications/Claude.app")) {
			tool.Detected = true
			tool.ConfigPath = filepath.Join(home, "Library/Application Support/Claude/claude_desktop_config.json")
		}
	case "linux":
		if pathExists(filepath.Join(home, ".local/share/claude")) {
			tool.Detected = true
			tool.ConfigPath = filepath.Join(home, ".config/Claude/claude_desktop_config.json")
		}
	case "windows":
		appdata := os.Getenv("APPDATA")
		localappdata := os.Getenv("LOCALAPPDATA")
		if pathExists(filepath.Join(localappdata, "Programs/claude-desktop")) {
			tool.Detected = true
			tool.ConfigPath = filepath.Join(appdata, "Claude/claude_desktop_config.json")
		}
	}

	return tool
}

func (d *Detector) detectClaudeCode() Tool {
	tool := Tool{
		ID:   "claude-code",
		Name: "Claude Code CLI",
	}

	if commandExists("claude") {
		tool.Detected = true
		home, _ := os.UserHomeDir()
		tool.ConfigPath = filepath.Join(home, ".claude.json")
	}

	return tool
}

func (d *Detector) detectCursor() Tool {
	tool := Tool{
		ID:   "cursor",
		Name: "Cursor",
	}

	home, _ := os.UserHomeDir()

	if commandExists("cursor") {
		tool.Detected = true
		tool.ConfigPath = filepath.Join(home, ".cursor/mcp.json")
	} else {
		switch runtime.GOOS {
		case "darwin":
			if pathExists("/Applications/Cursor.app") {
				tool.Detected = true
				tool.ConfigPath = filepath.Join(home, ".cursor/mcp.json")
			}
		case "linux":
			if pathExists(filepath.Join(home, ".cursor")) {
				tool.Detected = true
				tool.ConfigPath = filepath.Join(home, ".cursor/mcp.json")
			}
		}
	}

	return tool
}

func (d *Detector) detectVSCode() Tool {
	tool := Tool{
		ID:   "vscode",
		Name: "VS Code",
	}

	home, _ := os.UserHomeDir()

	if commandExists("code") || commandExists("code-insiders") {
		tool.Detected = true

		switch runtime.GOOS {
		case "darwin":
			tool.ConfigPath = filepath.Join(home, "Library/Application Support/Code/User/mcp.json")
		case "linux":
			tool.ConfigPath = filepath.Join(home, ".config/Code/User/mcp.json")
		case "windows":
			tool.ConfigPath = filepath.Join(os.Getenv("APPDATA"), "Code/User/mcp.json")
		}
	}

	return tool
}

func (d *Detector) detectZed() Tool {
	tool := Tool{
		ID:   "zed",
		Name: "Zed",
	}

	home, _ := os.UserHomeDir()

	if commandExists("zed") {
		tool.Detected = true
		tool.ConfigPath = filepath.Join(home, ".config/zed/settings.json")
	} else {
		switch runtime.GOOS {
		case "darwin":
			if pathExists("/Applications/Zed.app") {
				tool.Detected = true
				tool.ConfigPath = filepath.Join(home, ".config/zed/settings.json")
			}
		case "linux":
			if pathExists(filepath.Join(home, ".config/zed")) {
				tool.Detected = true
				tool.ConfigPath = filepath.Join(home, ".config/zed/settings.json")
			}
		}
	}

	return tool
}

func (d *Detector) detectWindsurf() Tool {
	tool := Tool{
		ID:   "windsurf",
		Name: "Windsurf",
	}

	home, _ := os.UserHomeDir()
	if pathExists(filepath.Join(home, ".codeium/windsurf")) {
		tool.Detected = true
		tool.ConfigPath = filepath.Join(home, ".codeium/windsurf/mcp_config.json")
	}

	return tool
}

func (d *Detector) detectJetBrains() Tool {
	tool := Tool{
		ID:   "jetbrains",
		Name: "JetBrains IDEs",
	}

	detected := false

	switch runtime.GOOS {
	case "darwin":
		apps := []string{
			"/Applications/IntelliJ IDEA.app",
			"/Applications/PyCharm.app",
			"/Applications/WebStorm.app",
			"/Applications/GoLand.app",
			"/Applications/PhpStorm.app",
			"/Applications/Rider.app",
		}
		for _, app := range apps {
			if pathExists(app) {
				detected = true
				break
			}
		}
	case "linux":
		home, _ := os.UserHomeDir()
		if pathExists(filepath.Join(home, ".config/JetBrains")) {
			detected = true
		}
	}

	if detected {
		tool.Detected = true
		tool.ConfigPath = "GUI Configuration Required"
	}

	return tool
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
