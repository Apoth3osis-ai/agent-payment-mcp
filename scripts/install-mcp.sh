#!/usr/bin/env bash

# Agent Payment MCP Universal Installer
# Automatically detects and configures MCP server for all supported AI tools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Variables
DETECTED_TOOLS=()
CONFIGURED_TOOLS=()
FAILED_TOOLS=()

# Print functions
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check if .env file exists
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        print_error ".env file not found!"
        print_info "Please copy .env.example to .env and configure your API credentials:"
        echo "  cp .env.example .env"
        echo "  nano .env  # Edit and add your actual API_KEY and BUDGET_KEY"
        exit 1
    fi

    # Load environment variables
    set -a
    source "$PROJECT_ROOT/.env"
    set +a

    # Check if API credentials are configured
    if [ "$API_KEY" = "your_api_key_here" ] || [ -z "$API_KEY" ]; then
        print_error "API_KEY not configured in .env file!"
        print_info "Please edit .env and add your actual API key from https://agentpmt.com"
        exit 1
    fi

    if [ "$BUDGET_KEY" = "your_budget_key_here" ] || [ -z "$BUDGET_KEY" ]; then
        print_error "BUDGET_KEY not configured in .env file!"
        print_info "Please edit .env and add your actual budget key from https://agentpmt.com"
        exit 1
    fi

    print_success "Environment configured"
    print_info "API_KEY: ${API_KEY:0:8}..."
    print_info "BUDGET_KEY: ${BUDGET_KEY:0:8}..."
}

# Determine binary path
get_binary_path() {
    local os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch_name=$(uname -m)

    # Map architecture names
    case "$arch_name" in
        x86_64|amd64) arch_name="amd64" ;;
        arm64|aarch64) arch_name="arm64" ;;
        *) print_error "Unsupported architecture: $arch_name"; exit 1 ;;
    esac

    # Map OS names
    case "$os_name" in
        linux) os_name="linux" ;;
        darwin) os_name="darwin" ;;
        mingw*|msys*|cygwin*) os_name="windows" ;;
        *) print_error "Unsupported OS: $os_name"; exit 1 ;;
    esac

    local binary_name="agent-payment-server"
    if [ "$os_name" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi

    local binary_path="$PROJECT_ROOT/distribution/binaries/${os_name}-${arch_name}/${binary_name}"

    if [ ! -f "$binary_path" ]; then
        print_warning "Pre-built binary not found at: $binary_path"
        print_info "Building from source..."

        # Check if Go is installed
        if ! command -v go &> /dev/null; then
            print_error "Go is not installed. Please install Go 1.21+ or use pre-built binaries."
            exit 1
        fi

        # Build from source
        cd "$PROJECT_ROOT/mcp-server"
        go build -o "$PROJECT_ROOT/bin/agent-payment-server" ./cmd/agent-payment-server
        binary_path="$PROJECT_ROOT/bin/agent-payment-server"

        # Make executable
        chmod +x "$binary_path"

        print_success "Built binary at: $binary_path"
    else
        # Make sure pre-built binary is executable
        chmod +x "$binary_path"
        print_success "Using pre-built binary: $binary_path"
    fi

    echo "$binary_path"
}

# Detect Claude Desktop
detect_claude_desktop() {
    print_info "Checking for Claude Desktop..."

    if [[ "$OSTYPE" == "darwin"* ]]; then
        if [ -d "/Applications/Claude.app" ] || [ -d "$HOME/Applications/Claude.app" ]; then
            DETECTED_TOOLS+=("claude-desktop")
            print_success "Detected: Claude Desktop (macOS)"
            return 0
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Check common Linux locations
        if [ -d "$HOME/.local/share/claude" ] || command -v claude &> /dev/null; then
            DETECTED_TOOLS+=("claude-desktop")
            print_success "Detected: Claude Desktop (Linux)"
            return 0
        fi
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        if [ -d "$LOCALAPPDATA/Programs/claude-desktop" ]; then
            DETECTED_TOOLS+=("claude-desktop")
            print_success "Detected: Claude Desktop (Windows)"
            return 0
        fi
    fi

    return 1
}

# Detect Claude Code CLI
detect_claude_code() {
    print_info "Checking for Claude Code CLI..."

    if command -v claude &> /dev/null; then
        DETECTED_TOOLS+=("claude-code")
        print_success "Detected: Claude Code CLI"
        return 0
    fi

    return 1
}

# Detect Cursor
detect_cursor() {
    print_info "Checking for Cursor..."

    if command -v cursor &> /dev/null; then
        DETECTED_TOOLS+=("cursor")
        print_success "Detected: Cursor"
        return 0
    fi

    # Check common installation locations
    if [[ "$OSTYPE" == "darwin"* ]] && [ -d "/Applications/Cursor.app" ]; then
        DETECTED_TOOLS+=("cursor")
        print_success "Detected: Cursor (macOS)"
        return 0
    elif [[ "$OSTYPE" == "linux-gnu"* ]] && [ -d "$HOME/.cursor" ]; then
        DETECTED_TOOLS+=("cursor")
        print_success "Detected: Cursor (Linux)"
        return 0
    fi

    return 1
}

# Detect Windsurf
detect_windsurf() {
    print_info "Checking for Windsurf..."

    if [ -d "$HOME/.codeium/windsurf" ]; then
        DETECTED_TOOLS+=("windsurf")
        print_success "Detected: Windsurf"
        return 0
    fi

    return 1
}

# Detect VS Code
detect_vscode() {
    print_info "Checking for VS Code..."

    if command -v code &> /dev/null || command -v code-insiders &> /dev/null; then
        DETECTED_TOOLS+=("vscode")
        print_success "Detected: VS Code"
        return 0
    fi

    return 1
}

# Detect Zed
detect_zed() {
    print_info "Checking for Zed..."

    if command -v zed &> /dev/null; then
        DETECTED_TOOLS+=("zed")
        print_success "Detected: Zed"
        return 0
    fi

    # Check common locations
    if [[ "$OSTYPE" == "darwin"* ]] && [ -d "/Applications/Zed.app" ]; then
        DETECTED_TOOLS+=("zed")
        print_success "Detected: Zed (macOS)"
        return 0
    elif [[ "$OSTYPE" == "linux-gnu"* ]] && [ -d "$HOME/.config/zed" ]; then
        DETECTED_TOOLS+=("zed")
        print_success "Detected: Zed (Linux)"
        return 0
    fi

    return 1
}

# Detect IntelliJ IDEA / JetBrains IDEs
detect_jetbrains() {
    print_info "Checking for JetBrains IDEs..."

    local detected=false

    # Check for common JetBrains IDEs
    if [[ "$OSTYPE" == "darwin"* ]]; then
        [ -d "/Applications/IntelliJ IDEA.app" ] && detected=true
        [ -d "/Applications/PyCharm.app" ] && detected=true
        [ -d "/Applications/WebStorm.app" ] && detected=true
        [ -d "/Applications/GoLand.app" ] && detected=true
        [ -d "/Applications/PhpStorm.app" ] && detected=true
        [ -d "/Applications/Rider.app" ] && detected=true
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        [ -d "$HOME/.config/JetBrains" ] && detected=true
    fi

    if [ "$detected" = true ]; then
        DETECTED_TOOLS+=("jetbrains")
        print_success "Detected: JetBrains IDEs"
        return 0
    fi

    return 1
}

# Configure Claude Desktop
configure_claude_desktop() {
    print_info "Configuring Claude Desktop..."

    local config_dir
    local config_file

    if [[ "$OSTYPE" == "darwin"* ]]; then
        config_dir="$HOME/Library/Application Support/Claude"
        config_file="$config_dir/claude_desktop_config.json"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        config_dir="$HOME/.config/Claude"
        config_file="$config_dir/claude_desktop_config.json"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        config_dir="$APPDATA/Claude"
        config_file="$config_dir/claude_desktop_config.json"
    else
        print_error "Unsupported OS for Claude Desktop"
        return 1
    fi

    # Create config directory if it doesn't exist
    mkdir -p "$config_dir"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'mcpServers' not in config:
    config['mcpServers'] = {}
config['mcpServers']['agent-payment'] = {
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured Claude Desktop"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("claude-desktop")
}

# Configure Claude Code CLI
configure_claude_code() {
    print_info "Configuring Claude Code CLI..."

    local config_file="$HOME/.claude.json"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'mcpServers' not in config:
    config['mcpServers'] = {}
config['mcpServers']['agent-payment'] = {
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured Claude Code CLI"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("claude-code")
}

# Configure Cursor
configure_cursor() {
    print_info "Configuring Cursor..."

    local config_dir="$HOME/.cursor"
    local config_file="$config_dir/mcp.json"

    # Create config directory if needed
    mkdir -p "$config_dir"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration (Cursor uses "mcpServers")
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'mcpServers' not in config:
    config['mcpServers'] = {}
config['mcpServers']['agent-payment'] = {
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured Cursor"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("cursor")
}

# Configure Windsurf
configure_windsurf() {
    print_info "Configuring Windsurf..."

    local config_dir="$HOME/.codeium/windsurf"
    local config_file="$config_dir/mcp_config.json"

    # Create config directory if needed
    mkdir -p "$config_dir"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'mcpServers' not in config:
    config['mcpServers'] = {}
config['mcpServers']['agent-payment'] = {
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured Windsurf"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("windsurf")
}

# Configure VS Code
configure_vscode() {
    print_info "Configuring VS Code..."

    local config_dir
    if [[ "$OSTYPE" == "darwin"* ]]; then
        config_dir="$HOME/Library/Application Support/Code/User"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        config_dir="$HOME/.config/Code/User"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        config_dir="$APPDATA/Code/User"
    else
        print_error "Unsupported OS for VS Code"
        return 1
    fi

    local config_file="$config_dir/mcp.json"

    # Create config directory if needed
    mkdir -p "$config_dir"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one (VS Code uses "servers" not "mcpServers")
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'servers' not in config:
    config['servers'] = {}
config['servers']['agent-payment'] = {
    'type': 'stdio',
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured VS Code"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("vscode")
}

# Configure Zed
configure_zed() {
    print_info "Configuring Zed..."

    local config_dir="$HOME/.config/zed"
    local config_file="$config_dir/settings.json"

    # Create config directory if needed
    mkdir -p "$config_dir"

    # Backup existing config
    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d_%H%M%S)"
        print_info "Backed up existing configuration"
    fi

    # Read existing config or create new one
    local existing_config="{}"
    if [ -f "$config_file" ]; then
        existing_config=$(cat "$config_file")
    fi

    # Add or update agent-payment server configuration (Zed uses context_servers)
    local new_config=$(echo "$existing_config" | python3 -c "
import sys, json
config = json.load(sys.stdin)
if 'context_servers' not in config:
    config['context_servers'] = {}
config['context_servers']['agent-payment'] = {
    'source': 'custom',
    'command': '$BINARY_PATH',
    'env': {
        'API_KEY': '$API_KEY',
        'BUDGET_KEY': '$BUDGET_KEY',
        'API_URL': '${API_URL:-https://api.agentpmt.com}'
    }
}
print(json.dumps(config, indent=2))
")

    echo "$new_config" > "$config_file"
    print_success "Configured Zed"
    print_info "Config file: $config_file"
    CONFIGURED_TOOLS+=("zed")
}

# Configure JetBrains IDEs
configure_jetbrains() {
    print_info "Configuring JetBrains IDEs..."

    print_warning "JetBrains IDEs require GUI configuration"
    print_info "Please follow these steps in your JetBrains IDE:"
    echo "  1. Open Settings (Cmd+, or Ctrl+Alt+S)"
    echo "  2. Go to: Tools → AI Assistant → Model Context Protocol"
    echo "  3. Click the '+' button to add a new server"
    echo "  4. Configure as follows:"
    echo "     - Name: agent-payment"
    echo "     - Command: $BINARY_PATH"
    echo "     - Environment Variables:"
    echo "       API_KEY=$API_KEY"
    echo "       BUDGET_KEY=$BUDGET_KEY"
    echo "       API_URL=${API_URL:-https://api.agentpmt.com}"
    echo "  5. Click Apply and OK"

    CONFIGURED_TOOLS+=("jetbrains-manual")
}

# Detect all tools
detect_all_tools() {
    print_header "Detecting Installed AI Tools"

    detect_claude_desktop || true
    detect_claude_code || true
    detect_cursor || true
    detect_windsurf || true
    detect_vscode || true
    detect_zed || true
    detect_jetbrains || true

    if [ ${#DETECTED_TOOLS[@]} -eq 0 ]; then
        print_warning "No supported AI tools detected"
        print_info "Supported tools: Claude Desktop, Claude Code CLI, Cursor, Windsurf, VS Code, Zed, JetBrains IDEs"
        exit 1
    fi

    print_success "Detected ${#DETECTED_TOOLS[@]} AI tool(s)"
}

# Configure all detected tools
configure_all_tools() {
    print_header "Configuring Detected Tools"

    for tool in "${DETECTED_TOOLS[@]}"; do
        case "$tool" in
            "claude-desktop")
                configure_claude_desktop || FAILED_TOOLS+=("claude-desktop")
                ;;
            "claude-code")
                configure_claude_code || FAILED_TOOLS+=("claude-code")
                ;;
            "cursor")
                configure_cursor || FAILED_TOOLS+=("cursor")
                ;;
            "windsurf")
                configure_windsurf || FAILED_TOOLS+=("windsurf")
                ;;
            "vscode")
                configure_vscode || FAILED_TOOLS+=("vscode")
                ;;
            "zed")
                configure_zed || FAILED_TOOLS+=("zed")
                ;;
            "jetbrains")
                configure_jetbrains
                ;;
        esac
    done
}

# Print summary and next steps
print_summary() {
    print_header "Installation Summary"

    echo ""
    print_info "Binary location: $BINARY_PATH"
    echo ""

    if [ ${#CONFIGURED_TOOLS[@]} -gt 0 ]; then
        print_success "Successfully configured ${#CONFIGURED_TOOLS[@]} tool(s):"
        for tool in "${CONFIGURED_TOOLS[@]}"; do
            echo "  - $tool"
        done
    fi

    if [ ${#FAILED_TOOLS[@]} -gt 0 ]; then
        echo ""
        print_error "Failed to configure ${#FAILED_TOOLS[@]} tool(s):"
        for tool in "${FAILED_TOOLS[@]}"; do
            echo "  - $tool"
        done
    fi

    echo ""
    print_header "Next Steps"
    echo ""
    print_info "1. Restart your AI tools for changes to take effect:"
    echo "   - Claude Desktop: Quit and relaunch"
    echo "   - Cursor: Reload window (Ctrl/Cmd+Shift+P → 'Developer: Reload Window')"
    echo "   - VS Code: Reload window (Ctrl/Cmd+Shift+P → 'Developer: Reload Window')"
    echo "   - Zed: Restart the editor"
    echo "   - Claude Code CLI: Run 'claude mcp list' to verify"
    echo ""
    print_info "2. Verify MCP server connection:"
    echo "   - Claude Desktop: Look for 🔨 hammer icon in chat"
    echo "   - Cursor: Settings → MCP → Check for green dot next to 'agent-payment'"
    echo "   - VS Code: Command Palette → 'MCP: List Servers'"
    echo "   - Zed: Open Agent Panel and check server status"
    echo ""
    print_info "3. Test the MCP server by asking your AI tool to use Agent Payment tools"
    echo ""
    print_info "For troubleshooting, see:"
    echo "  - AGENTS.md - Detailed installation guide"
    echo "  - GitHub Issues: https://github.com/Apoth3osis-ai/agent-payment-mcp/issues"
    echo ""
}

# Main execution
main() {
    clear
    print_header "Agent Payment MCP Universal Installer"
    echo ""

    # Check prerequisites
    check_prerequisites
    echo ""

    # Get binary path
    print_header "Preparing MCP Server Binary"
    BINARY_PATH=$(get_binary_path)
    echo ""

    # Detect tools
    detect_all_tools
    echo ""

    # Configure tools
    configure_all_tools
    echo ""

    # Print summary
    print_summary
}

# Run main function
main "$@"
