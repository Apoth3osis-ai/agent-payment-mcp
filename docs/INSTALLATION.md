# Installation Guide

This guide provides detailed installation instructions for Agent Payment MCP.

## Quick Install (Recommended)

For most users, follow the quick installation steps in the [main README](../README.md#quick-install):

1. Get API keys from [agentpmt.com](https://agentpmt.com)
2. Download the installer for your platform
3. Run the installer and enter your credentials
4. Restart your AI tool

## Platform-Specific Instructions

### Windows

1. Download `agent-payment-installer-windows-amd64.exe`
2. Double-click to run (Windows Defender may warn - see [troubleshooting](../README.md#troubleshooting))
3. Enter API credentials in the browser window
4. Select your AI tools
5. Click Install
6. If needed, add Windows Defender exclusion for `%USERPROFILE%\.agent-payment`
7. Restart Claude Desktop, Cursor, or your preferred AI tool

### macOS

1. Download the appropriate installer:
   - Apple Silicon (M1/M2/M3): `agent-payment-installer-macos-arm64`
   - Intel: `agent-payment-installer-macos-intel`
2. Open Terminal and make executable: `chmod +x agent-payment-installer-*`
3. Run: `./agent-payment-installer-*`
4. Browser opens automatically - enter credentials
5. Select AI tools and install
6. Restart your AI tools

### Linux

1. Download `agent-payment-installer-linux-amd64`
2. Make executable: `chmod +x agent-payment-installer-linux-amd64`
3. Run: `./agent-payment-installer-linux-amd64`
4. Browser opens automatically - enter credentials
5. Select AI tools and install
6. Restart your AI tools

## Manual Installation

If you need to install manually without the installer:

### 1. Download MCP Server Binary

Get the appropriate binary from the [releases page](https://github.com/Apoth3osis-ai/agent-payment-mcp/releases):
- `agent-payment-server-linux-amd64`
- `agent-payment-server-darwin-arm64`
- `agent-payment-server-darwin-amd64`
- `agent-payment-server-windows-amd64.exe`

### 2. Install Binary

```bash
# Create installation directory
mkdir -p ~/.agent-payment

# Move binary (replace with your platform's binary name)
mv agent-payment-server-* ~/.agent-payment/agent-payment-server
chmod +x ~/.agent-payment/agent-payment-server
```

### 3. Create Configuration

Create `~/.agent-payment/config.json`:

```json
{
  "api_key": "your-api-key-here",
  "budget_key": "your-budget-key-here",
  "api_url": "https://api.agentpmt.com"
}
```

### 4. Configure AI Tool

#### Claude Desktop

Edit your config file:
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

Add:
```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/username/.agent-payment/agent-payment-server"
    }
  }
}
```

#### Cursor

Edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/username/.agent-payment/agent-payment-server"
    }
  }
}
```

#### VS Code

Edit your MCP config:
- Windows: `%APPDATA%\Code\User\mcp.json`
- macOS: `~/Library/Application Support/Code/User/mcp.json`
- Linux: `~/.config/Code/User/mcp.json`

Add:
```json
{
  "servers": {
    "agent-payment": {
      "command": "/home/username/.agent-payment/agent-payment-server"
    }
  }
}
```

#### Zed

Edit `~/.config/zed/settings.json`:

```json
{
  "context_servers": {
    "agent-payment": {
      "command": "/home/username/.agent-payment/agent-payment-server"
    }
  }
}
```

#### Windsurf

Edit `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/username/.agent-payment/agent-payment-server"
    }
  }
}
```

### 5. Restart AI Tool

Close and reopen your AI tool completely for changes to take effect.

## Verification

After installation, verify the MCP server is working:

1. Open your AI tool
2. Check for Agent Payment tools in the tools list
3. Try using a tool (e.g., "Smart Math Interpreter")
4. Check logs if tools don't appear (see troubleshooting below)

## Troubleshooting

For troubleshooting steps, see the [main README troubleshooting section](../README.md#troubleshooting).

Common issues:
- Windows Defender blocking: Add exclusion for `%USERPROFILE%\.agent-payment`
- Tools not showing: Check MCP logs in your AI tool's log directory
- Connection errors: Verify API keys are correct in `config.json`

## Advanced Configuration

### Test Environment

To use the test API environment, change `api_url` in `config.json`:

```json
{
  "api_url": "https://test.api.agentpmt.com"
}
```

### Multiple Installations

You can install on multiple machines using the same API keys. Your budget key controls spending across all installations.

## Support

- Documentation: [Main README](../README.md)
- Issues: [GitHub Issues](https://github.com/Apoth3osis-ai/agent-payment-mcp/issues)
- Website: [agentpmt.com](https://agentpmt.com)
