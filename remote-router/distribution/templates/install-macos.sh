#!/bin/bash
# AgentPMT MCP Router Installer for macOS

set -e

# Parse arguments
CLIENT="auto"
API_KEY=""
BUDGET_KEY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --client)
            CLIENT="$2"
            shift 2
            ;;
        --api-key)
            API_KEY="$2"
            shift 2
            ;;
        --budget-key)
            BUDGET_KEY="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--client claude|cursor|auto] [--api-key KEY] [--budget-key KEY]"
            exit 1
            ;;
    esac
done

echo "AgentPMT MCP Router Installer for macOS"
echo "========================================"
echo ""

# Detect client
if [ "$CLIENT" = "auto" ]; then
    if [ -d "$HOME/Library/Application Support/Claude" ]; then
        CLIENT="claude"
        echo "✓ Detected Claude Desktop"
    elif [ -d "$HOME/.cursor" ]; then
        CLIENT="cursor"
        echo "✓ Detected Cursor"
    else
        echo "⚠ Could not auto-detect client. Use --client claude or --client cursor"
        exit 1
    fi
fi

# Set installation paths
if [ "$CLIENT" = "claude" ]; then
    INSTALL_PATH="$HOME/Library/Application Support/Claude/servers/agent-payment"
    CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
elif [ "$CLIENT" = "cursor" ]; then
    INSTALL_PATH="$HOME/.cursor/mcp/servers/agent-payment"
    CONFIG_PATH="$HOME/.cursor/mcp_config.json"
else
    echo "Invalid client: $CLIENT"
    exit 1
fi

echo "Installation path: $INSTALL_PATH"
echo ""

# Create directory
mkdir -p "$INSTALL_PATH"

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    BINARY_URL="https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-darwin-amd64"
elif [ "$ARCH" = "arm64" ]; then
    BINARY_URL="https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-darwin-arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

BINARY_PATH="$INSTALL_PATH/agent-payment-router"

# Download binary
echo "Downloading router binary for $ARCH..."
curl -L "$BINARY_URL" -o "$BINARY_PATH"
chmod +x "$BINARY_PATH"
echo "✓ Downloaded and installed binary"

# Remove quarantine attribute (macOS security)
echo "Removing quarantine attribute..."
xattr -d com.apple.quarantine "$BINARY_PATH" 2>/dev/null || true

# Create config.json (optional)
if [ -n "$API_KEY" ] && [ -n "$BUDGET_KEY" ]; then
    cat > "$INSTALL_PATH/config.json" <<EOF
{
  "APIURL": "https://api.agentpmt.com",
  "APIKey": "$API_KEY",
  "BudgetKey": "$BUDGET_KEY"
}
EOF
    echo "✓ Created config.json with provided keys"
else
    echo "⚠ No API keys provided - you'll need to set environment variables"
fi

# Update MCP config
echo ""
echo "Updating MCP configuration..."

# Create or update MCP config
if [ -f "$CONFIG_PATH" ]; then
    # Merge with existing config (requires jq)
    if command -v jq > /dev/null; then
        TMP_CONFIG=$(mktemp)
        jq ".mcpServers.\"agent-payment\" = {
            \"command\": \"$BINARY_PATH\",
            \"args\": [],
            \"env\": {
                \"AGENTPMT_API_KEY\": \"${API_KEY:-}\",
                \"AGENTPMT_BUDGET_KEY\": \"${BUDGET_KEY:-}\"
            }
        }" "$CONFIG_PATH" > "$TMP_CONFIG"
        mv "$TMP_CONFIG" "$CONFIG_PATH"
    else
        echo "⚠ jq not found, cannot merge with existing config"
        echo "   Install jq: brew install jq"
        echo "   Or manually add agent-payment to $CONFIG_PATH"
    fi
else
    # Create new config
    mkdir -p "$(dirname "$CONFIG_PATH")"
    cat > "$CONFIG_PATH" <<EOF
{
  "mcpServers": {
    "agent-payment": {
      "command": "$BINARY_PATH",
      "args": [],
      "env": {
        "AGENTPMT_API_KEY": "${API_KEY:-}",
        "AGENTPMT_BUDGET_KEY": "${BUDGET_KEY:-}"
      }
    }
  }
}
EOF
fi

echo "✓ Updated $CONFIG_PATH"

echo ""
echo "✓ Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Restart $CLIENT"
if [ -z "$API_KEY" ] || [ -z "$BUDGET_KEY" ]; then
    echo "  2. Get your API keys from https://agentpmt.com"
    echo "  3. Edit $CONFIG_PATH and add your keys"
fi
echo ""
echo "Note: If you see Gatekeeper warnings, the binary is not code-signed."
echo "      You may need to allow it in System Settings > Privacy & Security"
echo ""
