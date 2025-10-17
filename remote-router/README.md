# AgentPMT Remote MCP Router

> **Secure, lightweight MCP server that connects Claude Desktop and Cursor to 298+ AgentPMT tools**

[![Build Status](https://github.com/Apoth3osis-ai/agent-payment-mcp/actions/workflows/release-router.yml/badge.svg)](https://github.com/Apoth3osis-ai/agent-payment-mcp/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## What is This?

The AgentPMT Remote Router is a tiny, secure MCP server that gives Claude Desktop and Cursor access to 298+ powerful tools from the AgentPMT marketplace. Think of it as a bridge between your AI assistant and a world of capabilities.

**Key Features:**
- 🚀 **298+ Tools Available** - Math, code execution, data generation, and more
- 🔒 **100% Secure** - No shell access, no privileged operations, HTTPS only
- 📦 **Tiny Binary** - Just 5-6MB, starts in <100ms
- 🌍 **Cross-Platform** - Windows, macOS (Intel/ARM), Linux (x64/ARM)
- ⚡ **Optional Streaming** - Real-time responses for long-running operations
- 🎯 **Zero Config** - Works out of the box with API keys

---

## Quick Start (5 Minutes)

### Step 1: Get Your API Keys

Sign up at [agentpmt.com](https://agentpmt.com) to get:
- **API Key** - Your authentication token
- **Budget Key** - Your spending control token

### Step 2: Install the Router

Choose your platform:

#### **Linux / macOS**
```bash
# Download and run installer
curl -fsSL https://raw.githubusercontent.com/Apoth3osis-ai/agent-payment-mcp/main/remote-router/distribution/templates/install-linux.sh | bash

# Or manual installation:
mkdir -p ~/.agent-payment-router
cd ~/.agent-payment-router

# Download binary for your platform
# Linux x64:
curl -L -o agent-payment-router https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-linux-amd64

# macOS Intel:
curl -L -o agent-payment-router https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-darwin-amd64

# macOS ARM (M1/M2/M3):
curl -L -o agent-payment-router https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-darwin-arm64

# Make executable
chmod +x agent-payment-router
```

#### **Windows**
```powershell
# Download and run installer
irm https://raw.githubusercontent.com/Apoth3osis-ai/agent-payment-mcp/main/remote-router/distribution/templates/install-windows.ps1 | iex

# Or manual installation:
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agent-payment-router"
cd "$env:USERPROFILE\.agent-payment-router"

# Download binary
Invoke-WebRequest -Uri "https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-windows-amd64.exe" -OutFile "agent-payment-router.exe"
```

### Step 3: Configure API Keys

Create `config.json` in the installation directory:

**Linux/macOS:** `~/.agent-payment-router/config.json`
**Windows:** `%USERPROFILE%\.agent-payment-router\config.json`

```json
{
  "APIURL": "https://api.agentpmt.com",
  "APIKey": "your-api-key-here",
  "BudgetKey": "your-budget-key-here"
}
```

Or use environment variables:
```bash
export AGENTPMT_API_KEY="your-api-key"
export AGENTPMT_BUDGET_KEY="your-budget-key"
```

### Step 4: Connect to Claude Desktop or Cursor

#### **Claude Desktop**

Edit your Claude Desktop config:
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

Add this configuration:
```json
{
  "mcpServers": {
    "agentpmt": {
      "command": "/home/YOUR_USERNAME/.agent-payment-router/agent-payment-router"
    }
  }
}
```

**Windows users:** Use `"C:\\Users\\YOUR_USERNAME\\.agent-payment-router\\agent-payment-router.exe"`

#### **Cursor**

Edit your Cursor config:
- **macOS:** `~/Library/Application Support/Cursor/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`
- **Linux:** `~/.config/Cursor/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`
- **Windows:** `%APPDATA%\Cursor\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json`

Add the same configuration as above.

### Step 5: Restart and Test

**Restart Claude Desktop or Cursor**, then ask:
> "What tools do you have available?"

You should see **298 tools** with names like:
- `Smart-Math-Interpreter`
- `Secure-Python-Code-Sandbox`
- `Quantum-Random-Number-Generator`
- ... and 295 more!

---

## Migrating from Previous Installation

If you previously installed the old `agent-payment-server`, follow these steps:

### Step 1: Locate Old Installation

The old server was typically installed at:
- **Linux/macOS:** `~/.agent-payment/agent-payment-server`
- **Windows:** `%USERPROFILE%\.agent-payment\agent-payment-server.exe`

### Step 2: Save Your API Keys

Your old configuration is at:
- `~/.agent-payment/config.json` (Linux/macOS)
- `%USERPROFILE%\.agent-payment\config.json` (Windows)

Copy the `APIKey` and `BudgetKey` values - you'll need them for the new router.

### Step 3: Install New Router

Follow the "Quick Start" instructions above, using the same API keys from your old config.

### Step 4: Update Claude Desktop / Cursor Config

**Old configuration (REMOVE THIS):**
```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/YOUR_USERNAME/.agent-payment/agent-payment-server"
    }
  }
}
```

**New configuration (USE THIS):**
```json
{
  "mcpServers": {
    "agentpmt": {
      "command": "/home/YOUR_USERNAME/.agent-payment-router/agent-payment-router"
    }
  }
}
```

### Step 5: Restart and Verify

Restart Claude Desktop or Cursor and verify tools are available.

### Step 6: Remove Old Installation (Optional)

Once the new router is working:
```bash
# Linux/macOS
rm -rf ~/.agent-payment/

# Windows
Remove-Item -Recurse -Force "$env:USERPROFILE\.agent-payment"
```

---

## Uninstalling

To completely remove the AgentPMT Router:

### Step 1: Remove MCP Server Configuration

Edit your Claude Desktop or Cursor config file and **remove** the `agentpmt` entry:

```json
{
  "mcpServers": {
    "agentpmt": { ... }  ← DELETE THIS ENTIRE BLOCK
  }
}
```

### Step 2: Delete Router Files

**Linux/macOS:**
```bash
rm -rf ~/.agent-payment-router/
```

**Windows:**
```powershell
Remove-Item -Recurse -Force "$env:USERPROFILE\.agent-payment-router"
```

### Step 3: Restart

Restart Claude Desktop or Cursor to apply changes.

---

## Available Tools

The router provides access to 298 tools across multiple categories:

### 🔢 Mathematics & Computing
- **Smart-Math-Interpreter** - Universal math engine (SymPy, NumPy, SciPy)
- **Quantum-Random-Number-Generator** - True quantum randomness
- **Matrix-Operations-Engine** - Advanced linear algebra
- **Statistical-Analysis-Suite** - Comprehensive statistics

### 💻 Code Execution
- **Secure-Python-Code-Sandbox** - Execute Python code safely
- **JavaScript-Runtime-Environment** - Run JavaScript code
- **Code-Linter-and-Formatter** - Multi-language code quality

### 📊 Data & Finance
- **Synthetic-Financial-Data-Generator** - Realistic financial datasets
- **Time-Series-Forecasting-Engine** - Predictive analytics
- **Market-Simulation-Framework** - Financial modeling

### 🔐 Security & Cryptography
- **Quantum-Cryptographic-Seed-Generator** - Secure random seeds
- **Hash-Function-Library** - Multiple hashing algorithms
- **Encryption-Decryption-Suite** - AES, RSA, and more

### 📧 Communication & Email
- **Email-Verification-Service** - Validate email addresses
- **Email-Parser-and-Extractor** - Parse email content
- **SMTP-Email-Sender** - Send emails programmatically

**... and 270+ more tools!**

To see all tools, ask Claude: *"List all available tools"*

---

## Architecture

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│  Claude/Cursor  │ stdio   │  Router (5MB)    │  HTTPS  │  AgentPMT API   │
│  (MCP Client)   │◄──────►│  Go Binary       │◄──────►│  298+ Tools     │
└─────────────────┘         └──────────────────┘         └─────────────────┘

                            Zero Shell Access
                            Zero Privileged Ops
                            HTTPS Outbound Only
```

**How It Works:**
1. Claude Desktop sends JSON-RPC request via stdio
2. Router validates request and extracts tool name + parameters
3. Router maps tool name to product ID (e.g., "Smart-Math-Interpreter" → "689df4a...")
4. Router forwards request to AgentPMT API via HTTPS
5. API executes tool and returns result
6. Router returns result to Claude Desktop

**Security Model:**
- ✅ No shell command execution
- ✅ No local file system access (except reading config.json)
- ✅ No network listeners
- ✅ No elevated privileges required
- ✅ All secrets redacted from logs
- ✅ HTTPS communication only

---

## Configuration Options

### Config File Method

Create `~/.agent-payment-router/config.json`:
```json
{
  "APIURL": "https://api.agentpmt.com",
  "APIKey": "your-api-key",
  "BudgetKey": "your-budget-key"
}
```

### Environment Variables Method

Environment variables **override** config file values:
```bash
export AGENTPMT_API_URL="https://api.agentpmt.com"
export AGENTPMT_API_KEY="your-api-key"
export AGENTPMT_BUDGET_KEY="your-budget-key"
```

### Streaming (Optional)

Some tools support real-time streaming. To enable, tools automatically detect if streaming is available from the API.

No configuration needed - streaming is handled transparently!

---

## Troubleshooting

### Tools Not Appearing in Claude Desktop

**Symptoms:** Claude says "No tools available" or shows 0 tools

**Solutions:**
1. **Check API keys:** Ensure they're valid in `config.json`
2. **Restart Claude Desktop:** Close completely and reopen
3. **Check logs:** Run router manually to see errors:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ~/.agent-payment-router/agent-payment-router
   ```
4. **Verify connection:** Test API access:
   ```bash
   curl -H "X-API-Key: YOUR_API_KEY" -H "X-Budget-Key: YOUR_BUDGET_KEY" https://api.agentpmt.com/products/fetch?page=1&page_size=1
   ```

### Windows SmartScreen Warning

**Symptoms:** "Windows protected your PC" warning when running the binary

**Why This Happens:** New binaries need to build reputation with Microsoft

**Solutions:**
- Click "More info" → "Run anyway" (safe if downloaded from official GitHub releases)
- Wait 2-8 weeks as binary builds reputation
- Binary is code-signed but reputation takes time

**Note:** This is normal for all new Windows applications and does not indicate a security issue.

### Tool Execution Fails

**Symptoms:** Tool call returns error

**Check:**
1. **Budget balance:** Ensure your budget key has sufficient funds
2. **API status:** Check https://status.agentpmt.com
3. **Network:** Verify HTTPS access to api.agentpmt.com
4. **Tool parameters:** Ensure you're passing correct parameters

### "Pattern Validation Error"

**Symptoms:** Error about tool name pattern `^[a-zA-Z0-9_-]{1,64}$`

**Solution:** Update to latest version - this was fixed in recent releases:
```bash
cd ~/.agent-payment-router
curl -L -o agent-payment-router https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-linux-amd64
chmod +x agent-payment-router
```

---

## Advanced Usage

### Testing the Router Manually

Test stdio interface directly:
```bash
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ~/.agent-payment-router/agent-payment-router

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ~/.agent-payment-router/agent-payment-router | jq '.result.tools | length'

# Expected output: 298
```

### Using with Multiple MCP Servers

You can run multiple MCP servers simultaneously:
```json
{
  "mcpServers": {
    "agentpmt": {
      "command": "/home/user/.agent-payment-router/agent-payment-router"
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/Documents"]
    }
  }
}
```

### Custom API Endpoint

For testing or enterprise deployments:
```json
{
  "APIURL": "https://custom-api.example.com",
  "APIKey": "your-api-key",
  "BudgetKey": "your-budget-key"
}
```

---

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/Apoth3osis-ai/agent-payment-mcp.git
cd agent-payment-mcp/remote-router

# Install dependencies
go mod tidy

# Build for current platform
go build -o agent-payment-router ./cmd/agent-payment-router

# Test
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./agent-payment-router
```

### Cross-Platform Builds

```bash
# Build all platforms with version
./scripts/build-all.sh 1.0.0

# Binaries created in distribution/binaries/
```

### Running Tests

```bash
# Unit tests
go test ./...

# E2E tests
cd tests && go test -v

# Stdio smoke test
./scripts/test-stdio.sh
```

See [DEVELOPMENT.md](./DEVELOPMENT.md) for detailed developer documentation.

---

## Support

- 🐛 **Report Bugs:** [GitHub Issues](https://github.com/Apoth3osis-ai/agent-payment-mcp/issues)
- 📖 **Documentation:** [agentpmt.com/docs](https://agentpmt.com/docs)
- 💬 **Community:** [Discord](https://discord.gg/agentpmt) *(coming soon)*
- 📧 **Email:** support@agentpmt.com

---

## FAQ

**Q: Is this safe to use?**
A: Yes! The router has zero privileged operations, no shell access, and only makes HTTPS requests to AgentPMT API. All code is open source.

**Q: What tools are available?**
A: 298 tools across math, code execution, data generation, security, email, and more. Ask Claude: "List all available tools"

**Q: Does this cost money?**
A: Tools are pay-per-use based on AgentPMT pricing. Your budget key controls spending. Check [agentpmt.com/pricing](https://agentpmt.com/pricing)

**Q: Can I use this with Cursor?**
A: Yes! Works with Claude Desktop, Cursor, and any MCP-compatible client.

**Q: How is this different from the old server?**
A: The new router is 40% smaller, remote-first (no local execution), uses readable tool names, and supports dynamic pagination.

**Q: Can I self-host the API?**
A: Enterprise customers can run a private AgentPMT API instance. Contact sales@agentpmt.com

---

## License

MIT License - See [LICENSE](../LICENSE) for details.

---

## Changelog

### v1.0.0 (2025-10-16)
- ✅ Initial production release
- ✅ 298 tools with dynamic pagination
- ✅ MCP-compliant tool names (hyphens instead of spaces)
- ✅ Bidirectional mapping (readable names ↔ product IDs)
- ✅ Optional SSE streaming support
- ✅ Cross-platform binaries (5 platforms)
- ✅ Complete test coverage (29 tests)
- ✅ Windows code signing support

---

**Built with ❤️ using [Claude Code](https://claude.com/claude-code)**
