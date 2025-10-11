#!/bin/bash
set -e

echo "================================================"
echo "Local Testing Setup"
echo "================================================"
echo ""

# Ensure Go is in PATH
export PATH=$PATH:/usr/local/go/bin

# Verify Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go not found in PATH"
    echo "Please run: source ~/.bashrc"
    echo "Then try again"
    exit 1
fi

# Build Go server
echo "1. Building Go MCP server..."
cd mcp-server
go mod tidy || echo "Warning: go mod tidy failed (may need to update go.mod with correct SDK path)"
go build -o ../distribution/binaries/linux-amd64/agent-payment-server ./cmd/agent-payment-server || {
    echo "Error: Go build failed. You may need to:"
    echo "  - Update go.mod with correct MCP SDK import path"
    echo "  - Run: go get github.com/modelcontextprotocol/go-sdk@latest"
    exit 1
}
cd ..

echo "✅ Go server built: distribution/binaries/linux-amd64/agent-payment-server"
echo ""

# Install PWA dependencies
echo "2. Installing PWA dependencies..."
cd pwa
if [ ! -d "node_modules" ]; then
    npm install
else
    echo "   (node_modules already exists, skipping)"
fi
cd ..

echo "✅ PWA dependencies installed"
echo ""

# Instructions
echo "================================================"
echo "✅ Setup Complete!"
echo "================================================"
echo ""
echo "Next steps:"
echo ""
echo "1. Start PWA dev server:"
echo "   cd pwa && npm run dev"
echo "   Visit http://localhost:5173"
echo ""
echo "2. In the PWA:"
echo "   - Settings tab: Enter your API credentials"
echo "   - Tools tab: Browse available tools"
echo "   - Install tab: Download installer package"
echo ""
echo "3. Install the package:"
echo "   - Extract downloaded ZIP"
echo "   - Run install script: ./install-linux.sh claude"
echo "   - Restart Claude Desktop"
echo ""
echo "4. Test in Claude:"
echo "   - Open Claude Desktop"
echo "   - Look for Agent Payment tools"
echo "   - Try executing a tool"
echo ""
