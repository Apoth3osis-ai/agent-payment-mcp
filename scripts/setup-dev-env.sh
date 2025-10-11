#!/bin/bash
set -e

echo "================================================"
echo "Agent Payment - Developer Environment Setup"
echo "================================================"
echo ""
echo "This script will automatically install all required dependencies:"
echo "  - Go 1.21+"
echo "  - Node.js 20+"
echo "  - npm"
echo "  - jq (for JSON manipulation)"
echo ""

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
    # Detect package manager
    if command -v apt-get &> /dev/null; then
        PKG_MGR="apt"
    elif command -v yum &> /dev/null; then
        PKG_MGR="yum"
    elif command -v dnf &> /dev/null; then
        PKG_MGR="dnf"
    else
        echo "Error: Unsupported Linux distribution (no apt/yum/dnf found)"
        exit 1
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
    PKG_MGR="brew"
else
    echo "Error: Unsupported OS: $OSTYPE"
    exit 1
fi

echo "Detected: $OS with $PKG_MGR"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Install Node.js
if command_exists node; then
    NODE_VERSION=$(node -v | sed 's/v//' | cut -d. -f1)
    if [ "$NODE_VERSION" -ge 20 ]; then
        echo "✅ Node.js $(node -v) already installed"
    else
        echo "⚠️  Node.js $(node -v) is too old (need 20+), upgrading..."
        INSTALL_NODE=true
    fi
else
    echo "📦 Installing Node.js 20..."
    INSTALL_NODE=true
fi

if [ "$INSTALL_NODE" = true ]; then
    if [ "$OS" = "linux" ]; then
        # Install NodeSource repository
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        if [ "$PKG_MGR" = "apt" ]; then
            sudo apt-get update
            sudo apt-get install -y nodejs
        elif [ "$PKG_MGR" = "yum" ] || [ "$PKG_MGR" = "dnf" ]; then
            sudo $PKG_MGR install -y nodejs
        fi
    elif [ "$OS" = "macos" ]; then
        if ! command_exists brew; then
            echo "Installing Homebrew first..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        fi
        brew install node@20
    fi
    echo "✅ Node.js installed"
fi

# Install Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//' | cut -d. -f1,2)
    GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
    if [ "$GO_MAJOR" -ge 1 ] && [ "$GO_MINOR" -ge 21 ]; then
        echo "✅ Go $(go version | awk '{print $3}') already installed"
    else
        echo "⚠️  Go $(go version | awk '{print $3}') is too old (need 1.21+), upgrading..."
        INSTALL_GO=true
    fi
else
    echo "📦 Installing Go 1.21..."
    INSTALL_GO=true
fi

if [ "$INSTALL_GO" = true ]; then
    if [ "$OS" = "linux" ]; then
        wget -q https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
        rm go1.21.13.linux-amd64.tar.gz

        # Add to PATH if not already there
        if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        export PATH=$PATH:/usr/local/go/bin
    elif [ "$OS" = "macos" ]; then
        brew install go
    fi
    echo "✅ Go installed"
fi

# Install jq
if command_exists jq; then
    echo "✅ jq already installed"
else
    echo "📦 Installing jq..."
    if [ "$OS" = "linux" ]; then
        if [ "$PKG_MGR" = "apt" ]; then
            sudo apt-get update
            sudo apt-get install -y jq
        elif [ "$PKG_MGR" = "yum" ] || [ "$PKG_MGR" = "dnf" ]; then
            sudo $PKG_MGR install -y jq
        fi
    elif [ "$OS" = "macos" ]; then
        brew install jq
    fi
    echo "✅ jq installed"
fi

# Verify installations
echo ""
echo "================================================"
echo "Verifying installations..."
echo "================================================"

if ! command_exists node; then
    echo "❌ Node.js installation failed"
    exit 1
fi

if ! command_exists npm; then
    echo "❌ npm installation failed"
    exit 1
fi

if ! command_exists go; then
    echo "❌ Go installation failed"
    echo "   Please run: source ~/.bashrc"
    echo "   Then run this script again"
    exit 1
fi

if ! command_exists jq; then
    echo "❌ jq installation failed"
    exit 1
fi

echo "✅ Node.js: $(node -v)"
echo "✅ npm: $(npm -v)"
echo "✅ Go: $(go version | awk '{print $3}')"
echo "✅ jq: $(jq --version)"

echo ""
echo "================================================"
echo "Installing Project Dependencies..."
echo "================================================"
echo ""

# Install PWA dependencies
cd "$(dirname "$0")/.."
echo "📦 Installing PWA dependencies..."
cd pwa
npm install
cd ..
echo "✅ PWA dependencies installed"

# Install Go dependencies
echo ""
echo "📦 Installing Go dependencies..."
cd mcp-server
go mod tidy
cd ..
echo "✅ Go dependencies installed"

echo ""
echo "================================================"
echo "✅ Development Environment Ready!"
echo "================================================"
echo ""
echo "Next steps:"
echo ""
echo "1. Build Go server:"
echo "   ./scripts/build-all.sh"
echo ""
echo "2. Start PWA dev server:"
echo "   cd pwa && npm run dev"
echo ""
echo "3. Or run the quick test:"
echo "   ./scripts/test-local.sh"
echo ""
