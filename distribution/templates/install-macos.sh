#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "================================================"
echo "Agent Payment MCP Server Installer (macOS)"
echo "================================================"
echo ""

# Determine config path based on editor
EDITOR="${1:-claude}"
if [ "$EDITOR" = "claude" ]; then
    CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
elif [ "$EDITOR" = "cursor" ]; then
    CONFIG_PATH="$HOME/.cursor/mcp.json"
elif [ "$EDITOR" = "vscode" ]; then
    CONFIG_PATH="$HOME/.config/Code/User/globalStorage/claude-code/mcp.json"
else
    echo -e "${RED}Unknown editor: $EDITOR${NC}"
    echo "Usage: $0 [claude|cursor|vscode]"
    exit 1
fi

# Installation directory
INSTALL_DIR="$HOME/.agent-payment"

echo "Editor: $EDITOR"
echo "Config path: $CONFIG_PATH"
echo "Install directory: $INSTALL_DIR"
echo ""

# Create installation directory
echo -e "${GREEN}Creating installation directory...${NC}"
mkdir -p "$INSTALL_DIR"

# Copy server executable
echo -e "${GREEN}Installing server executable...${NC}"
cp agent-payment-server "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/agent-payment-server"

# Copy configuration
echo -e "${GREEN}Installing configuration...${NC}"
cp config.json "$INSTALL_DIR/"

# Configure editor
echo -e "${GREEN}Configuring $EDITOR...${NC}"

# Create config directory if needed
mkdir -p "$(dirname "$CONFIG_PATH")"

# Initialize config if doesn't exist
if [ ! -f "$CONFIG_PATH" ]; then
    if [ "$EDITOR" = "claude" ]; then
        echo '{"mcpServers":{}}' > "$CONFIG_PATH"
    else
        echo '{"servers":{}}' > "$CONFIG_PATH"
    fi
fi

# Check if jq is available
if command -v jq &> /dev/null; then
    # Use jq for JSON manipulation
    BACKUP="$CONFIG_PATH.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$CONFIG_PATH" "$BACKUP"
    echo -e "${GREEN}Backed up config to: $BACKUP${NC}"

    if [ "$EDITOR" = "claude" ]; then
        jq ".mcpServers[\"agent-payment\"] = {
            \"command\": \"$INSTALL_DIR/agent-payment-server\",
            \"args\": []
        }" "$CONFIG_PATH" > "$CONFIG_PATH.tmp"
    else
        jq ".servers[\"agent-payment\"] = {
            \"type\": \"stdio\",
            \"command\": \"$INSTALL_DIR/agent-payment-server\",
            \"args\": []
        }" "$CONFIG_PATH" > "$CONFIG_PATH.tmp"
    fi

    mv "$CONFIG_PATH.tmp" "$CONFIG_PATH"
    echo -e "${GREEN}✅ Configuration updated automatically${NC}"
else
    echo -e "${YELLOW}⚠️  jq not found. Manual configuration required:${NC}"
    echo ""
    echo "Add this to $CONFIG_PATH:"
    echo ""
    if [ "$EDITOR" = "claude" ]; then
        cat <<EOF
{
  "mcpServers": {
    "agent-payment": {
      "command": "$INSTALL_DIR/agent-payment-server",
      "args": []
    }
  }
}
EOF
    else
        cat <<EOF
{
  "servers": {
    "agent-payment": {
      "type": "stdio",
      "command": "$INSTALL_DIR/agent-payment-server",
      "args": []
    }
  }
}
EOF
    fi
    echo ""
fi

echo ""
echo "================================================"
echo -e "${GREEN}✅ Installation Complete!${NC}"
echo "================================================"
echo ""
echo "Server installed to: $INSTALL_DIR"
echo ""
echo -e "${YELLOW}⚠️  Important: Restart $EDITOR to apply changes${NC}"
echo ""
