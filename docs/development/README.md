# Developer Documentation

This folder contains technical documentation for developers who want to contribute to or build Agent Payment MCP from source.

## Contents

- **[AGENTS.md](AGENTS.md)** - AI agent installation guide with structured instructions for automated setup
- **[GO_MCP_SDK_RESEARCH_AND_IMPLEMENTATION.md](GO_MCP_SDK_RESEARCH_AND_IMPLEMENTATION.md)** - Research and implementation details for the Go MCP SDK integration
- **[IMPLEMENTATION_COMPLETE.md](IMPLEMENTATION_COMPLETE.md)** - Implementation completion report and final architecture
- **[IMPLEMENT_PLAN.md](IMPLEMENT_PLAN.md)** - Detailed implementation plan and technical specifications
- **[MCP_SERVER_RESEARCH_REPORT.md](MCP_SERVER_RESEARCH_REPORT.md)** - Research findings on MCP server implementation approaches
- **[PARALLEL_EXECUTION_GUIDE.md](PARALLEL_EXECUTION_GUIDE.md)** - Guide for parallel tool execution and optimization
- **[PLAN.md](PLAN.md)** - Project planning and architecture decisions

## For End Users

If you're an end user looking to install Agent Payment MCP, please see the [main README](../../README.md) for simple installation instructions.

## Building from Source

### Prerequisites

- Go 1.21+
- Node.js 20+
- Git

### Build MCP Server

```bash
cd mcp-server
go build -o agent-payment-server ./cmd/agent-payment-server
```

### Build Installer

```bash
cd installer
go build -o agent-payment-installer ./cmd/installer
```

### Build PWA

```bash
cd pwa
npm install
npm run build
```

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## Architecture Overview

Agent Payment MCP consists of:

1. **MCP Server** (Go) - Proxies Agent Payment API tools to MCP clients
2. **Installer** (Go) - Cross-platform installer with embedded binaries
3. **PWA** (React) - Web interface for browsing tools and generating installers

## Support

- Issues: [GitHub Issues](https://github.com/Apoth3osis-ai/agent-payment-mcp/issues)
- Main Documentation: [README.md](../../README.md)
