#!/bin/bash
set -e

echo "================================================"
echo "Creating .mcpb Packages for Claude Desktop"
echo "================================================"
echo ""

# Ensure binaries exist
if [ ! -d "distribution/binaries" ]; then
    echo "Error: distribution/binaries not found. Run build-all.sh first."
    exit 1
fi

# Create packages directory
mkdir -p distribution/packages

# Function to create .mcpb package
create_mcpb() {
    local platform=$1
    local binary_name=$2
    local package_name=$3

    echo "Creating .mcpb package for $platform..."

    local temp_dir="distribution/temp-mcpb-$platform"
    mkdir -p "$temp_dir"

    # Copy files
    cp "distribution/binaries/$platform/$binary_name" "$temp_dir/"
    cp "distribution/templates/mcpb-manifest.json" "$temp_dir/manifest.json"
    cp "agent-payment-logo.png" "$temp_dir/"

    # Create README
    cat > "$temp_dir/README.md" <<EOF
# Agent Payment MCP Server

Thank you for installing Agent Payment!

## What's Included

- MCP server executable
- manifest.json with server configuration

## Installation

1. Double-click the .mcpb file
2. Claude Desktop will open automatically
3. Click "Install" when prompted
4. Restart Claude Desktop
5. Tools will appear in your MCP tools list

## Configuration

Before using the tools, you'll need to:
1. Get your API credentials from https://agentpmt.com
2. Configure the server with your credentials

## Support

- Website: https://agentpmt.com
- Documentation: https://docs.agentpmt.com
- Issues: https://github.com/your-repo/issues
EOF

    # Create ZIP
    cd "$temp_dir"
    zip -r "../../distribution/packages/$package_name.zip" .
    cd - > /dev/null

    # Rename to .mcpb
    mv "distribution/packages/$package_name.zip" "distribution/packages/$package_name.mcpb"

    # Clean up
    rm -rf "$temp_dir"

    echo "✅ Created: distribution/packages/$package_name.mcpb"
}

# Create packages for each platform
create_mcpb "windows-amd64" "agent-payment-server.exe" "agent-payment-windows-amd64"
create_mcpb "darwin-amd64" "agent-payment-server" "agent-payment-macos-intel"
create_mcpb "darwin-arm64" "agent-payment-server" "agent-payment-macos-arm64"
create_mcpb "linux-amd64" "agent-payment-server" "agent-payment-linux-amd64"

echo ""
echo "================================================"
echo "✅ All .mcpb packages created!"
echo "================================================"
echo ""
echo "Packages:"
ls -lh distribution/packages/*.mcpb
echo ""
