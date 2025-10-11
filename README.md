# Agent Payment MCP

Access AI-powered tools from Agent Payment API in Claude Desktop, Cursor, and VS Code.

## Overview

Agent Payment MCP provides a seamless way to integrate premium AI tools into your desktop workflows. The system consists of:

- **PWA (Progressive Web App)** - Web interface to browse tools and generate installers
- **Go MCP Server** - Lightweight standalone server (6-8MB) that proxies tools to desktop clients
- **Installation Scripts** - Automated setup for Claude Desktop, Cursor, and VS Code

## Quick Start

### For End Users

1. **Get API Credentials**
   - Visit [agentpmt.com](https://agentpmt.com) to get your API and budget keys

2. **Visit the PWA**
   - Go to [install.agentpmt.com](https://install.agentpmt.com) (or your deployment URL)
   - Enter your API credentials in Settings
   - Browse available tools in the Tools tab
   - Download installer for your platform from the Install tab

3. **Install & Use**
   - Run the downloaded installer
   - Restart your desktop client (Claude/Cursor/VS Code)
   - Tools will appear in your MCP tools list

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

## Documentation

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
