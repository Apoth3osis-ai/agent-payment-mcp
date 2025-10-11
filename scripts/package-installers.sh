#!/bin/bash
set -e

echo "================================================"
echo "Creating Installer Packages"
echo "================================================"
echo ""

# Ensure binaries exist
if [ ! -d "distribution/binaries" ]; then
    echo "Error: distribution/binaries not found. Run build-all.sh first."
    exit 1
fi

# Create packages directory
mkdir -p distribution/packages

# Create example config
cat > distribution/config.example.json <<EOF
{
  "api_key": "your-api-key-here",
  "budget_key": "your-budget-key-here",
  "api_url": "https://api.agentpmt.com"
}
EOF

# Function to create installer package
create_installer() {
    local platform=$1
    local binary_name=$2
    local install_script=$3
    local package_name=$4

    echo "Creating installer package for $platform..."

    local temp_dir="distribution/temp-installer-$platform"
    mkdir -p "$temp_dir"

    # Copy files
    cp "distribution/binaries/$platform/$binary_name" "$temp_dir/"
    cp "distribution/templates/$install_script" "$temp_dir/"
    cp "distribution/config.example.json" "$temp_dir/config.json"

    # Make scripts executable
    chmod +x "$temp_dir/$install_script"
    if [ "$binary_name" != "agent-payment-server.exe" ]; then
        chmod +x "$temp_dir/$binary_name"
    fi

    # Create README
    cat > "$temp_dir/README.md" <<EOF
# Agent Payment MCP Server Installer

Thank you for downloading Agent Payment!

## What's Included

- \`$binary_name\`: MCP server executable
- \`$install_script\`: Installation script
- \`config.json\`: Configuration file (update with your API keys)
- This README

## Prerequisites

You need API credentials from https://agentpmt.com

## Installation

### Step 1: Configure API Keys

Edit \`config.json\` and add your API credentials:
\`\`\`json
{
  "api_key": "your-api-key",
  "budget_key": "your-budget-key",
  "api_url": "https://api.agentpmt.com"
}
\`\`\`

### Step 2: Run Installer

**macOS/Linux:**
\`\`\`bash
chmod +x $install_script
./$install_script [claude|cursor|vscode]
\`\`\`

**Windows:**
Right-click \`$install_script\` and select "Run with PowerShell"

### Step 3: Restart Your Editor

Restart Claude Desktop, Cursor, or VS Code to load the MCP server.

## Supported Editors

- Claude Desktop
- Cursor
- VS Code (with Claude Code extension)

## Troubleshooting

### Server not starting
- Ensure the executable has execute permissions (macOS/Linux)
- Check that config.json is in the same directory
- Verify your API credentials are correct

### Tools not appearing
- Restart your editor completely
- Check the MCP server logs in your editor
- Verify config.json syntax is valid

## Support

- Website: https://agentpmt.com
- Documentation: https://docs.agentpmt.com
- Issues: https://github.com/your-repo/issues

## Security

- Keep your \`config.json\` file secure
- Do not share your API credentials
- The executable only connects to api.agentpmt.com
EOF

    # Create ZIP (preserve permissions)
    cd "$temp_dir"
    zip -r "../../distribution/packages/$package_name.zip" .
    cd - > /dev/null

    # Clean up
    rm -rf "$temp_dir"

    echo "✅ Created: distribution/packages/$package_name.zip"
}

# Create installer packages for each platform
create_installer "windows-amd64" "agent-payment-server.exe" "install-windows.ps1" "agent-payment-windows-installer"
create_installer "darwin-amd64" "agent-payment-server" "install-macos.sh" "agent-payment-macos-intel-installer"
create_installer "darwin-arm64" "agent-payment-server" "install-macos.sh" "agent-payment-macos-arm64-installer"
create_installer "linux-amd64" "agent-payment-server" "install-linux.sh" "agent-payment-linux-installer"

echo ""
echo "================================================"
echo "✅ All installer packages created!"
echo "================================================"
echo ""
echo "Packages:"
ls -lh distribution/packages/*.zip
echo ""
