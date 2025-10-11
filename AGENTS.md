# AGENTS.md - AI Agent Installation Guide

This file provides structured instructions for AI coding agents (Claude Code, Cursor, Gemini CLI, Codex, etc.) to help users install and configure the Agent Payment MCP server.

## Project Overview

This repository contains a complete Model Context Protocol (MCP) server system that connects desktop AI clients to the Agent Payment API. The system consists of:

- **Go MCP Server**: Lightweight binary (6-8MB) that proxies Agent Payment tools to desktop clients
- **Universal installer script**: Automatically detects and configures all supported AI tools
- **Pre-built binaries**: Cross-platform executables for Windows, macOS (Intel/ARM), and Linux

## Prerequisites Check

Before proceeding with installation, verify these requirements:

1. **Go 1.21 or higher** (only needed if building from source)
   ```bash
   go version
   # Should show: go version go1.21.0 or higher
   ```

2. **API Credentials**: User must have Agent Payment API and Budget keys
   - Get from: https://agentpmt.com/dashboard/api-keys
   - Will be configured in `.env` file

3. **Operating System**: Linux, macOS 10.15+, or Windows 10+ (with WSL for some tools)

## Installation Steps

### Step 1: Clone Repository

```bash
git clone https://github.com/Apoth3osis-ai/agent-payment-mcp
cd agent-payment-mcp
```

### Step 2: Configure API Credentials

```bash
# Copy environment template
cp .env.example .env

# Edit .env file and add real credentials
# Replace 'your_api_key_here' with actual API key
# Replace 'your_budget_key_here' with actual Budget key
nano .env  # or use any text editor
```

**Required .env content:**
```bash
API_KEY=actual_api_key_from_agentpmt
BUDGET_KEY=actual_budget_key_from_agentpmt
```

### Step 3: Build MCP Server (or use pre-built binary)

**Option A: Use Pre-built Binary (Recommended)**

Pre-built binaries are available in `distribution/binaries/`:
- Windows: `distribution/binaries/windows-amd64/agent-payment-server.exe`
- Linux: `distribution/binaries/linux-amd64/agent-payment-server`
- macOS Intel: `distribution/binaries/darwin-amd64/agent-payment-server`
- macOS ARM: `distribution/binaries/darwin-arm64/agent-payment-server`

Make binary executable (Linux/macOS):
```bash
chmod +x distribution/binaries/linux-amd64/agent-payment-server
```

**Option B: Build from Source**

```bash
cd mcp-server
go build -o ../bin/agent-payment-server ./cmd/agent-payment-server
cd ..

# Make executable (Linux/macOS)
chmod +x bin/agent-payment-server
```

### Step 4: Run Universal Installer

```bash
# Run the auto-detection and configuration script
./scripts/install-mcp.sh
```

This script will:
1. Detect which AI tools are installed (Claude Desktop, Cursor, VS Code, etc.)
2. Build or copy the MCP server binary to appropriate location
3. Generate configuration for each detected tool
4. Create backup of existing configurations
5. Apply new MCP server configuration
6. Provide verification steps

### Step 5: Restart AI Tools

After installation completes, restart any AI tools that were configured:
- Close and reopen Claude Desktop
- Reload Cursor (Ctrl+Shift+P → "Developer: Reload Window")
- Reload VS Code (Ctrl+Shift+P → "Developer: Reload Window")
- Restart other configured tools

### Step 6: Verify Installation

**For Claude Desktop:**
- Look for hammer icon (🔨) in the chat input
- Click to see "agent-payment" server and its tools

**For Cursor:**
- Open Settings → MCP
- Verify "agent-payment" server shows green dot (connected)

**For VS Code:**
- Run command: "MCP: List Servers"
- Verify "agent-payment" appears with "Running" status

**For other tools:**
- Follow tool-specific verification steps shown by installer

## Troubleshooting

### Binary won't execute

**Linux/macOS:**
```bash
# Ensure binary has execute permissions
chmod +x bin/agent-payment-server

# Test execution
./bin/agent-payment-server --help
```

**Windows:**
- Check antivirus hasn't quarantined the .exe file
- Run from PowerShell with full path

### Configuration not found

```bash
# Verify .env file exists and has correct values
cat .env

# Should show:
# API_KEY=<your actual key>
# BUDGET_KEY=<your actual key>
```

### AI tool doesn't detect MCP server

1. **Check configuration file path:**
   - Each tool has specific config file location (see tool-specific docs)
   - Installer shows exact paths used

2. **Verify configuration syntax:**
   - Check JSON is valid (no trailing commas, proper quotes)
   - Use online JSON validator if needed

3. **Check logs:**
   - Claude Desktop: `~/Library/Logs/Claude/mcp*.log` (macOS) or `%LOCALAPPDATA%\Claude\Logs\` (Windows)
   - VS Code: Command Palette → "Output" → select "MCP Logs"
   - Cursor: Settings → MCP → View server logs

4. **Restart tool completely:**
   - Full quit and relaunch (not just close window)

### Server won't start

```bash
# Test server manually
cd mcp-server
go run ./cmd/agent-payment-server

# Check for errors in output
# Common issues:
# - Missing Go installation
# - Port already in use
# - Invalid .env configuration
```

## Building and Testing

### Run MCP Server Manually

```bash
# Load environment variables from .env
export $(cat .env | xargs)

# Run server
./bin/agent-payment-server

# Or from source
cd mcp-server
go run ./cmd/agent-payment-server
```

### Test with MCP Inspector

```bash
# Install MCP Inspector (requires Node.js)
npm install -g @modelcontextprotocol/inspector

# Test your server
npx @modelcontextprotocol/inspector ./bin/agent-payment-server
```

This opens a web interface showing:
- Available tools
- Tool schemas
- Test tool execution
- Request/response inspection

### Build for Different Platforms

```bash
# Build for all platforms
./scripts/build-all.sh

# Or manually:
GOOS=linux GOARCH=amd64 go build -o dist/agent-payment-server-linux ./cmd/agent-payment-server
GOOS=darwin GOARCH=amd64 go build -o dist/agent-payment-server-macos-intel ./cmd/agent-payment-server
GOOS=darwin GOARCH=arm64 go build -o dist/agent-payment-server-macos-arm ./cmd/agent-payment-server
GOOS=windows GOARCH=amd64 go build -o dist/agent-payment-server.exe ./cmd/agent-payment-server
```

## Supported AI Tools

The following AI tools are automatically detected and configured:

1. **Claude Desktop** - Native MCP support
2. **Claude Code CLI** - Full MCP client
3. **Cursor** - MCP via settings
4. **Windsurf** - Codeium-based MCP
5. **VS Code** - GitHub Copilot MCP extension
6. **Zed** - Context servers
7. **IntelliJ IDEA / JetBrains IDEs** - Built-in MCP (2025.1+)
8. **Sourcegraph Cody** - OpenCtx MCP provider

## Configuration Files

Each tool stores MCP configuration in a specific location:

- **Claude Desktop**:
  - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
  - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

- **Cursor**: `~/.cursor/mcp.json`

- **Windsurf**: `~/.codeium/windsurf/mcp_config.json`

- **VS Code**: `.vscode/mcp.json` or User Settings JSON

- **Zed**: `~/.config/zed/settings.json` (context_servers section)

- **Claude Code CLI**: `~/.claude.json`

## Security Notes

1. **Never commit `.env` file** - Contains sensitive API keys
2. **Use `.gitignore`** - Ensure `.env` is ignored
3. **Protect API keys** - Treat like passwords
4. **Review permissions** - Understand what Agent Payment API access grants
5. **Use budget limits** - Set appropriate Budget Key limits in Agent Payment dashboard

## Getting Help

- **Issues**: https://github.com/Apoth3osis-ai/agent-payment-mcp/issues
- **Documentation**: https://docs.agentpmt.com
- **Agent Payment Support**: https://agentpmt.com/support

## Code Organization

```
agent-payment-mcp/
├── mcp-server/              # Go MCP server source
│   ├── cmd/                 # Main application
│   ├── internal/            # Internal packages
│   │   ├── api/            # Agent Payment API client
│   │   ├── config/         # Configuration management
│   │   └── mcp/            # MCP protocol implementation
│   ├── go.mod              # Go dependencies
│   └── README.md           # Server documentation
├── distribution/            # Pre-built binaries and scripts
│   ├── binaries/           # Platform-specific executables
│   └── templates/          # Installation script templates
├── scripts/                 # Build and installation scripts
│   ├── build-all.sh        # Cross-platform build script
│   └── install-mcp.sh      # Universal installer
├── .env.example            # Environment template
├── .gitignore              # Git ignore rules
├── AGENTS.md               # This file - AI agent instructions
└── README.md               # User-facing documentation
```

## Development Workflow

For AI agents helping with development:

1. **Making changes to MCP server:**
   ```bash
   cd mcp-server
   # Edit files in cmd/ or internal/
   go build -o ../bin/agent-payment-server ./cmd/agent-payment-server
   # Test changes
   ```

2. **Updating installer script:**
   ```bash
   # Edit scripts/install-mcp.sh
   # Test on current system
   ./scripts/install-mcp.sh
   ```

3. **Adding new tool support:**
   - Research tool's MCP configuration format
   - Add detection logic to `install-mcp.sh`
   - Add configuration template
   - Update documentation

4. **Running tests:**
   ```bash
   cd mcp-server
   go test ./...
   ```

## Command Reference

**Quick command cheatsheet for common operations:**

```bash
# Setup
git clone https://github.com/Apoth3osis-ai/agent-payment-mcp
cd agent-payment-mcp
cp .env.example .env
# Edit .env with real keys

# Build
cd mcp-server && go build -o ../bin/agent-payment-server ./cmd/agent-payment-server && cd ..

# Install
chmod +x scripts/install-mcp.sh
./scripts/install-mcp.sh

# Test
./bin/agent-payment-server --help
npx @modelcontextprotocol/inspector ./bin/agent-payment-server

# Verify
# Restart AI tool, then check for MCP server connection
```

## Important Notes for AI Agents

When helping users with this project:

1. **Always confirm they have API credentials** before proceeding
2. **Use pre-built binaries when possible** - faster than building from source
3. **Run the universal installer** - handles all configuration automatically
4. **Verify .env file** is properly configured with real keys (not example placeholders)
5. **Remind to restart AI tools** after configuration changes
6. **Check logs if issues occur** - most problems show clear error messages
7. **Use absolute paths** in configurations - relative paths often fail
8. **Windows users**: Some tools (Claude Code CLI) require WSL

## Future Enhancements

Planned features (check roadmap for status):
- GUI configuration tool
- Docker container deployment
- Kubernetes deployment manifests
- Additional AI tool support as they adopt MCP
- Enhanced OAuth support when Codex/Gemini add MCP
