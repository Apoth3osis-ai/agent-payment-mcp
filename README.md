# Agent Payment MCP

Access AI-powered tools from Agent Payment API in Claude Desktop, Cursor, VS Code, Zed, IntelliJ IDEA, Windsurf, and more.

## 🤖 For AI Agents

If you're an AI agent helping a user install this MCP server, please read **[AGENTS.md](AGENTS.md)** for structured installation instructions. It contains step-by-step guidance for automated setup.

## Overview

Agent Payment MCP provides a seamless way to integrate premium AI tools into your desktop workflows. The system consists of:

- **PWA (Progressive Web App)** - Web interface to browse tools and generate installers
- **Go MCP Server** - Lightweight standalone server (6-8MB) that proxies tools to desktop clients
- **Universal Installer** - Automated setup for 8+ AI tools including Claude Desktop, Cursor, VS Code, Zed, JetBrains IDEs, Windsurf, and more

## Quick Start

### Automated Installation (Recommended)

1. **Clone this repository:**
   ```bash
   git clone https://github.com/Apoth3osis-ai/agent-payment-mcp
   cd agent-payment-mcp
   ```

2. **Configure API credentials:**
   ```bash
   cp .env.example .env
   # Edit .env and add your actual API_KEY and BUDGET_KEY from https://agentpmt.com
   ```

3. **Run universal installer:**
   ```bash
   ./scripts/install-mcp.sh
   ```

   This script automatically:
   - Detects installed AI tools (Claude Desktop, Cursor, VS Code, Zed, etc.)
   - Configures MCP server for each detected tool
   - Sets up the Go binary with your credentials

4. **Restart your AI tools** and verify the MCP server appears

### Manual Installation via PWA

1. **Get API Credentials**
   - Visit [agentpmt.com](https://agentpmt.com) to get your API and budget keys

2. **Visit the PWA**
   - Go to [install.agentpmt.com](https://install.agentpmt.com) (or your deployment URL)
   - Enter your API credentials in Settings
   - Browse available tools in the Tools tab
   - Download installer for your platform from the Install tab

3. **Install & Use**
   - Run the downloaded installer
   - Restart your desktop client
   - Tools will appear in your MCP tools list

### For AI Agents

If you're an AI coding assistant helping with installation, see **[AGENTS.md](AGENTS.md)** for detailed agent-friendly instructions.

### For Developers

See [Building from Source](#building-from-source) below.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PWA (Browser)                             │
│  - Enter API credentials (encrypted locally)                    │
│  - Browse available tools                                        │
│  - Download installers                                           │
└─────────────────────────────────────────────────────────────────┘
                              |
                              | Downloads
                              v
┌─────────────────────────────────────────────────────────────────┐
│                 Installer Package (ZIP/.mcpb)                    │
│  - Go MCP Server binary                                          │
│  - config.json (with user's API keys)                            │
│  - Install script                                                │
└─────────────────────────────────────────────────────────────────┘
                              |
                              | Installs to
                              v
┌─────────────────────────────────────────────────────────────────┐
│              Desktop Client (Claude/Cursor/VS Code)              │
│  - Runs Go MCP Server on startup                                │
│  - Displays tools in native UI                                  │
│  - Executes tools via MCP protocol                              │
└─────────────────────────────────────────────────────────────────┘
                              |
                              | API Calls
                              v
┌─────────────────────────────────────────────────────────────────┐
│                  Agent Payment API                               │
│  - Provides tool definitions                                     │
│  - Executes tool requests                                        │
│  - Manages billing/budgets                                       │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
agent-payment-system/
├── pwa/                    # Progressive Web App
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── routes/        # Page components
│   │   └── lib/           # Utilities (crypto, storage, API)
│   ├── public/            # Static assets
│   └── package.json
│
├── mcp-server/            # Go MCP Server
│   ├── cmd/               # Main application
│   ├── internal/          # Server logic
│   └── go.mod
│
├── distribution/          # Build outputs
│   ├── binaries/         # Compiled Go binaries
│   ├── packages/         # .mcpb and installer ZIPs
│   └── templates/        # Install scripts & configs
│
├── scripts/              # Build scripts
│   ├── build-all.sh
│   ├── package-mcpb.sh
│   └── package-installers.sh
│
└── .github/              # CI/CD
    └── workflows/
```

## Building from Source

### Prerequisites

- **Node.js 20+** (for PWA)
- **Go 1.21+** (for MCP server)
- **Git**

### Initial Setup

```bash
# Clone repository
git clone https://github.com/Apoth3osis-ai/agent-payment-mcp
cd agent-payment-mcp

# Install PWA dependencies
cd pwa
npm install
cd ..

# Initialize Go module
cd mcp-server
go mod download
cd ..
```

### Build Everything

```bash
# Build Go binaries for all platforms
./scripts/build-all.sh

# Build PWA
cd pwa
npm run build
cd ..

# Optional: Create distribution packages
./scripts/package-mcpb.sh        # .mcpb packages for Claude
./scripts/package-installers.sh  # Installer ZIPs for all editors
```

### Development

```bash
# Run PWA dev server
cd pwa
npm run dev
# Visit http://localhost:5173

# Build Go server for local testing
cd mcp-server
go build -o agent-payment-server ./cmd/agent-payment-server
./agent-payment-server  # Requires config.json
```

## Deployment

### PWA Deployment

The PWA can be deployed to any static hosting service:

- **Vercel**: `cd pwa && vercel deploy`
- **Netlify**: `cd pwa && netlify deploy --prod`
- **AWS S3**: Upload `pwa/dist/` to S3 bucket
- **GitHub Pages**: Use `.github/workflows/release.yml`

### Automated Releases

Push a git tag to trigger automated builds:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will:
1. Build Go binaries for all platforms
2. Build PWA
3. Create GitHub release with binaries
4. Deploy PWA to hosting (configure in workflow)

## Configuration

### PWA Environment Variables

Create `pwa/.env` for local development:

```env
VITE_API_BASE_URL=https://api.agentpmt.com
```

### MCP Server Configuration

The server reads `config.json`:

```json
{
  "api_key": "your-api-key",
  "budget_key": "your-budget-key",
  "api_url": "https://api.agentpmt.com"
}
```

## Supported AI Tools

This MCP server works with the following AI coding assistants:

- **Claude Desktop** - Full native MCP support
- **Claude Code CLI** - Command-line interface with MCP
- **Cursor** - VS Code fork with AI capabilities
- **Windsurf** - Codeium-powered IDE
- **VS Code** - With GitHub Copilot extension
- **Zed** - Modern code editor with AI
- **IntelliJ IDEA / JetBrains IDEs** (2025.1+) - Built-in MCP client
- **Sourcegraph Cody** - Code intelligence with MCP via OpenCtx

The universal installer automatically detects and configures all installed tools.

## Documentation

- **[AGENTS.md](AGENTS.md)** - AI agent installation guide (for automated setup)
- **[PWA Documentation](pwa/README.md)** - PWA development guide
- **[MCP Server Documentation](mcp-server/README.md)** - Server implementation details
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute
- **[Implementation Plan](IMPLEMENT_PLAN.md)** - Detailed implementation guide

## Support

- **Website**: [agentpmt.com](https://agentpmt.com)
- **Documentation**: [docs.agentpmt.com](https://docs.agentpmt.com)
- **Issues**: [GitHub Issues](https://github.com/Apoth3osis-ai/agent-payment-mcp/issues)

## License

MIT - See [LICENSE](LICENSE) file for details.
