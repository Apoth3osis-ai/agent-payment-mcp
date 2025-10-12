# MCP Server Framework & Standalone Executable Research Report

**Date:** October 9, 2025
**Purpose:** Research alternatives for building MCP servers as standalone executables requiring no user-installed dependencies

---

## Executive Summary

After comprehensive research, **Go with the official MCP SDK** emerges as the optimal choice for creating standalone MCP server executables. Go produces the smallest binaries (5-10MB optimized), has trivial cross-compilation, native performance, and the simplest distribution model. For your specific use case (dynamic tool fetching from REST API with API key authentication), Go provides the best balance of simplicity, performance, and user experience.

**Recommended Stack:**
- **Framework:** Go with official MCP SDK (`github.com/modelcontextprotocol/go-sdk`)
- **Packaging:** Native Go compilation (cross-platform built-in)
- **Distribution:** Pre-built binaries with external config file
- **Final Size:** 5-10MB per platform

---

## 1. MCP Server Framework Comparison

### 1.1 Python Options

#### FastMCP 2.0
**Package:** `fastmcp` (maintained by Prefect)
**Latest Version:** 2.6+
**GitHub:** https://github.com/jlowin/fastmcp

**Pros:**
- Most Pythonic, decorator-based API (`@mcp.tool`)
- Excellent documentation at https://gofastmcp.com/
- Dynamic tool registration via mounting/proxying
- Enterprise auth support (Google, GitHub, Azure, Auth0)
- Server composition with `mcp.mount("prefix", sub_server)`
- Active development and community

**Cons:**
- Requires packaging into large executables (12-150MB)
- PyInstaller adds significant overhead
- Python runtime bundling complexity

**Code Example:**
```python
from fastmcp import FastMCP

mcp = FastMCP("My Server")

@mcp.tool()
def add_numbers(a: int, b: int) -> int:
    """Add two numbers together"""
    return a + b

if __name__ == "__main__":
    mcp.run(transport='stdio')
```

**Dynamic Tool Registration:**
```python
# Fetch tools from REST API and mount dynamically
api_mcp = FastMCP("API Tools")

# Fetch tool definitions from REST API
tools = fetch_tools_from_api()

for tool_def in tools:
    @api_mcp.tool()
    def dynamic_tool(**kwargs):
        # Proxy to REST API
        return call_api(tool_def, kwargs)

# Mount to main server
mcp.mount("api", api_mcp)
```

**Ease of Use:** 9/10 - Simplest Python option
**Documentation:** 9/10 - Excellent
**Stability:** 8/10 - Actively maintained

---

#### Official Python MCP SDK
**Package:** `mcp`
**Latest Version:** 1.7.1
**GitHub:** https://github.com/modelcontextprotocol/python-sdk

**Pros:**
- Official implementation
- Lower-level control
- FastMCP 1.0 was incorporated into this SDK

**Cons:**
- More boilerplate than FastMCP 2.0
- Less developer-friendly API
- Same packaging challenges as FastMCP

**Code Example:**
```python
from mcp.server import Server
from mcp.server.stdio import stdio_server

server = Server("my-server")

@server.list_tools()
async def list_tools():
    return [{"name": "add", "description": "Add numbers"}]

@server.call_tool()
async def call_tool(name, arguments):
    if name == "add":
        return arguments["a"] + arguments["b"]

async def main():
    async with stdio_server() as streams:
        await server.run(
            streams[0], streams[1],
            server.create_initialization_options()
        )
```

**Ease of Use:** 6/10 - More verbose
**Documentation:** 7/10 - Adequate but less polished
**Stability:** 9/10 - Official SDK

---

### 1.2 TypeScript/Node.js Options

#### Official TypeScript MCP SDK
**Package:** `@modelcontextprotocol/sdk`
**Requirements:** Node.js 18+
**GitHub:** https://github.com/modelcontextprotocol/typescript-sdk

**Pros:**
- Official implementation
- Good TypeScript support
- Clean async API
- Standard transports (stdio, SSE)

**Cons:**
- Requires packaging with pkg/nexe
- Node.js runtime increases file size (~50MB+)
- More complex than Go for binaries

**Code Example:**
```typescript
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';

const server = new McpServer({
  name: 'demo-server',
  version: '1.0.0'
});

server.registerTool(
  'add',
  {
    title: 'Addition Tool',
    description: 'Add two numbers',
    inputSchema: { a: z.number(), b: z.number() }
  },
  async ({ a, b }) => ({
    content: [{ type: 'text', text: String(a + b) }]
  })
);

const transport = new StdioServerTransport();
await server.connect(transport);
```

**Ease of Use:** 7/10 - TypeScript adds complexity
**Documentation:** 8/10 - Well documented
**Stability:** 9/10 - Official SDK

---

### 1.3 Compiled Language Options

#### Go MCP SDK (Official)
**Package:** `github.com/modelcontextprotocol/go-sdk`
**Maintained by:** Google collaboration
**Community Alternatives:** `mark3labs/mcp-go`, `metoro-io/mcp-golang`

**Pros:**
- **BEST for standalone executables** - native compilation
- Smallest binary sizes (5-10MB optimized)
- **Trivial cross-compilation** - single command
- Fast startup, low memory usage
- No runtime dependencies
- Type-safe with Go structs
- Simple deployment

**Cons:**
- Less dynamic than Python/JS (compile-time)
- Smaller ecosystem than Python/Node
- More verbose than Python

**Code Example (Official SDK):**
```go
package main

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/modelcontextprotocol/go-sdk/transport/stdio"
)

func main() {
    server := mcp.NewServer("demo-server", "1.0.0")

    server.RegisterTool("add", mcp.Tool{
        Description: "Add two numbers",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "a": map[string]string{"type": "number"},
                "b": map[string]string{"type": "number"},
            },
        },
    }, func(args map[string]interface{}) (interface{}, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return a + b, nil
    })

    transport := stdio.NewStdioTransport()
    server.Serve(context.Background(), transport)
}
```

**Build & Cross-Compile:**
```bash
# Build for current platform
go build -ldflags "-s -w" -o mcp-server

# Cross-compile for all platforms
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o mcp-server-windows.exe
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o mcp-server-macos-amd64
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o mcp-server-macos-arm64
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o mcp-server-linux
```

**Ease of Use:** 8/10 - Simple once you know Go
**Documentation:** 8/10 - Good official docs
**Stability:** 9/10 - Official + Google maintained

---

#### Rust MCP SDK (Official)
**Package:** `rmcp` crate
**GitHub:** https://github.com/modelcontextprotocol/rust-sdk

**Pros:**
- **Smallest possible binaries** (158KB-3MB stripped)
- Highest performance (4,700+ QPS documented)
- Memory safety guarantees
- Macro-based tools (`#[tool]`)
- Both stdio and HTTP transports

**Cons:**
- **Steepest learning curve**
- Longer compile times
- Smaller MCP ecosystem
- More complex than Go for most use cases

**Code Example:**
```rust
use rmcp::{tool, Server};

#[tool]
fn add(a: i32, b: i32) -> i32 {
    a + b
}

#[tokio::main]
async fn main() {
    let server = Server::new("demo-server")
        .tool(add);

    server.run_stdio().await.unwrap();
}
```

**Build:**
```bash
cargo build --release
strip target/release/mcp-server  # 158KB - 600KB
```

**Ease of Use:** 5/10 - Rust learning curve
**Documentation:** 7/10 - Growing
**Stability:** 8/10 - Official but newer

---

## 2. Framework Comparison Table

| Framework | Language | Ease of Use | Binary Size | Cross-Compile | Dynamic Tools | Docs Quality | Recommendation |
|-----------|----------|-------------|-------------|---------------|---------------|--------------|----------------|
| **FastMCP 2.0** | Python | 9/10 | 12-150MB | Via PyInstaller | Excellent | 9/10 | Use for rapid prototyping |
| Official Python SDK | Python | 6/10 | 12-150MB | Via PyInstaller | Good | 7/10 | Use if you need official support |
| TypeScript SDK | TypeScript | 7/10 | 50MB+ | Via pkg | Good | 8/10 | Use if team knows Node.js |
| **Go SDK (Official)** | Go | 8/10 | **5-10MB** | **Built-in** | Moderate | 8/10 | **BEST for distribution** |
| Rust SDK | Rust | 5/10 | 0.2-3MB | Built-in | Moderate | 7/10 | Use for max performance |

---

## 3. Standalone Executable Packaging Comparison

### 3.1 Python Packaging Options

#### PyInstaller
**Maturity:** Most mature (since 2005)
**Python Support:** 3.8-3.14

**Pros:**
- Most widely used
- Best package compatibility
- Handles binary dependencies well
- Cross-platform (Windows, macOS, Linux)
- Simple CLI: `pyinstaller --onefile script.py`

**Cons:**
- Large file sizes (12-150MB typical)
- Must build on each target platform (no cross-compile)
- Slower startup (extracts to temp dir)
- Anti-virus false positives

**File Size Examples:**
- Hello World: 12MB
- FastMCP server: 20-50MB
- Complex app with dependencies: 100-150MB

**Build Process:**
```bash
pip install pyinstaller
pyinstaller --onefile \
    --name mcp-server \
    --hidden-import fastmcp \
    server.py
```

**Recommendation:** Best Python option, but still produces large binaries

---

#### Nuitka
**Maturity:** Active, compiles Python to C
**Performance:** 2-4x faster execution

**Pros:**
- Compiles Python to C (performance boost)
- Genuine compilation vs bundling
- Supports all Python constructs

**Cons:**
- **Larger than PyInstaller** (21-52MB vs 12-20MB)
- Longer build times (10-30 minutes)
- More complex configuration
- Occasional compatibility issues

**File Size Examples:**
- Hello World: 21MB (vs PyInstaller's 12MB)
- Complex app: 50-150MB

**Recommendation:** Only use if you need performance boost, not for distribution

---

#### PyOxidizer
**Maturity:** Newer (since 2019), Rust-based
**Approach:** Embeds Python in Rust binary

**Pros:**
- Modern approach
- Fast startup (no extraction)
- Single file output

**Cons:**
- Complex configuration (TOML/Python)
- Must build binary dependencies from source
- Smaller ecosystem than PyInstaller
- Steeper learning curve

**Recommendation:** Interesting but not worth complexity vs PyInstaller

---

### 3.2 Node.js Packaging Options

#### pkg (Vercel)
**Status:** Maintenance mode (last update 2021)
**Node Support:** Up to Node 18

**Pros:**
- Simple CLI: `pkg server.js`
- Cross-compilation built-in
- Single executable output

**Cons:**
- **No longer actively maintained**
- Large binaries (50MB+)
- Limited to Node 18
- TypeScript requires pre-compilation

**File Size:** 50-70MB typical

**Build Process:**
```bash
npm install -g pkg
pkg server.js --targets node18-win-x64,node18-macos-x64,node18-linux-x64
```

**Recommendation:** Avoid due to maintenance status

---

#### nexe
**Status:** Alternative to pkg
**Approach:** Similar bundling

**Pros:**
- Active development
- Similar to pkg

**Cons:**
- Same large binary sizes
- Complex with native modules

**Recommendation:** Better than pkg but still large binaries

---

### 3.3 Compiled Language Packaging

#### Go Native Compilation
**Maturity:** Built into Go toolchain
**Simplicity:** Single command

**Pros:**
- **Built into language** - no external tools
- **Trivial cross-compilation** - set env vars
- **Smallest practical binaries** (5-10MB)
- Fast compilation (seconds)
- No runtime dependencies
- **No anti-virus issues**

**Cons:**
- Must learn Go (if not already familiar)

**File Size with Optimization:**
```bash
# Basic build: ~20-30MB
go build -o server

# Optimized: 5-10MB
go build -ldflags "-s -w" -o server
# -s: strip symbol table
# -w: strip DWARF debug info

# With UPX compression: 2-5MB (optional)
upx --best server
```

**Cross-Compilation:**
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o server.exe

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o server-macos-intel

# macOS ARM (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o server-macos-arm

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o server-linux
```

**Recommendation:** BEST option for standalone executables

---

#### Rust Native Compilation
**Maturity:** Built into Rust toolchain
**Binary Size:** Smallest possible

**Pros:**
- **Smallest binaries** (158KB-3MB stripped)
- Highest performance
- Memory safety

**Cons:**
- Steeper learning curve
- Longer compile times (minutes vs seconds)

**File Size with Optimization:**
```toml
# Cargo.toml
[profile.release]
opt-level = "z"      # Optimize for size
lto = true           # Link-time optimization
codegen-units = 1    # Better optimization
strip = true         # Strip symbols
panic = "abort"      # Smaller panic handler
```

```bash
cargo build --release
# Result: 600KB-3MB
```

**Recommendation:** Best for size/performance, but overkill for most MCP servers

---

### 3.4 Cross-Platform Frameworks

#### Tauri
**Maturity:** Modern (since 2020)
**Approach:** Rust backend + web frontend

**Use Case:** System tray app with MCP server background process

**Pros:**
- **Small binaries** (5-15MB)
- System tray integration
- Web UI for configuration
- Modern DX (similar to Electron)

**Cons:**
- Complexity (full GUI app)
- Overkill for command-line server
- Requires Rust knowledge

**File Size:** 5-15MB

**Recommendation:** Only if you need GUI configuration panel

---

#### Electron
**Maturity:** Very mature
**Approach:** Chromium + Node.js

**Pros:**
- Full GUI capabilities
- Large ecosystem

**Cons:**
- **Massive binaries** (80-150MB minimum)
- High memory usage
- Slow startup

**Recommendation:** Avoid for MCP servers

---

## 4. Packaging Method Comparison Table

| Method | Language | Binary Size | Build Complexity | Cross-Compile | Startup Time | Recommendation |
|--------|----------|-------------|------------------|---------------|--------------|----------------|
| PyInstaller | Python | 12-150MB | Low | Manual (per platform) | Slow (extraction) | Best Python option |
| Nuitka | Python | 21-150MB | Medium | Manual (per platform) | Fast | For performance only |
| PyOxidizer | Python | 15-100MB | High | Manual (per platform) | Fast | Skip, too complex |
| pkg (Vercel) | Node.js | 50-70MB | Low | Built-in | Medium | Avoid (unmaintained) |
| nexe | Node.js | 50-70MB | Low | Built-in | Medium | Better than pkg |
| **Go Native** | Go | **5-10MB** | **Very Low** | **Built-in** | **Instant** | **BEST CHOICE** |
| Rust Native | Rust | 0.2-3MB | Medium | Built-in | Instant | Best for size/perf |
| Tauri | Rust+Web | 5-15MB | High | Built-in | Fast | Only if GUI needed |
| Electron | JS+Chromium | 80-150MB | Medium | Built-in | Slow | Avoid |

---

## 5. Recommended Architecture

### 5.1 Framework Choice: **Go with Official MCP SDK**

**Justification:**
1. **Smallest practical binaries** (5-10MB) - 95% smaller than Python
2. **Trivial cross-compilation** - build all platforms in seconds
3. **No runtime dependencies** - true standalone
4. **Fast startup** - instant vs Python's extraction delay
5. **Official support** - maintained in collaboration with Google
6. **Simple distribution** - single binary, no installers
7. **No anti-virus issues** - native code, not packed Python

**Trade-offs:**
- Less dynamic than Python (but sufficient for your use case)
- Team needs Go knowledge (but Go is simpler than Rust)
- Slightly more verbose than FastMCP Python

---

### 5.2 Complete Implementation Example

#### Go MCP Server with Dynamic REST API Tools

```go
// server.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/modelcontextprotocol/go-sdk/transport/stdio"
)

// Config loaded from config.json next to executable
type Config struct {
    APIKey    string `json:"api_key"`
    BudgetKey string `json:"budget_key"`
    APIURL    string `json:"api_url"`
}

// Tool definition from REST API
type APITool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// Load config from file next to executable
func loadConfig() (*Config, error) {
    exePath, _ := os.Executable()
    exeDir := filepath.Dir(exePath)
    configPath := filepath.Join(exeDir, "config.json")

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return &config, nil
}

// Fetch available tools from REST API
func fetchTools(config *Config) ([]APITool, error) {
    req, _ := http.NewRequest("GET", config.APIURL+"/products/fetch", nil)
    req.Header.Set("X-API-Key", config.APIKey)
    req.Header.Set("X-Budget-Key", config.BudgetKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var tools []APITool
    if err := json.Unmarshal(body, &tools); err != nil {
        return nil, err
    }

    return tools, nil
}

// Execute tool by calling REST API
func executeTool(config *Config, toolName string, args map[string]interface{}) (interface{}, error) {
    payload := map[string]interface{}{
        "tool": toolName,
        "args": args,
    }

    jsonData, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", config.APIURL+"/products/purchase", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", config.APIKey)
    req.Header.Set("X-Budget-Key", config.BudgetKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var result interface{}
    json.Unmarshal(body, &result)

    return result, nil
}

func main() {
    // Load configuration
    config, err := loadConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Fetch tools from REST API
    tools, err := fetchTools(config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching tools: %v\n", err)
        os.Exit(1)
    }

    // Create MCP server
    server := mcp.NewServer("dynamic-api-server", "1.0.0")

    // Register each tool dynamically
    for _, tool := range tools {
        toolName := tool.Name // Capture for closure
        toolDesc := tool.Description
        toolSchema := tool.InputSchema

        server.RegisterTool(toolName, mcp.Tool{
            Description: toolDesc,
            InputSchema: toolSchema,
        }, func(args map[string]interface{}) (interface{}, error) {
            return executeTool(config, toolName, args)
        })
    }

    // Start server with stdio transport
    transport := stdio.NewStdioTransport()
    server.Serve(context.Background(), transport)
}
```

#### Config File Format (config.json)

```json
{
  "api_key": "user-api-key-here",
  "budget_key": "user-budget-key-here",
  "api_url": "https://api.example.com"
}
```

---

### 5.3 Build Process

#### Build Script (build.sh)

```bash
#!/bin/bash
set -e

echo "Building MCP Server for all platforms..."

# Build flags for size optimization
LDFLAGS="-s -w"

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mcp-server-windows.exe

# Build for macOS Intel
echo "Building for macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mcp-server-macos-intel

# Build for macOS ARM (M1/M2/M3)
echo "Building for macOS ARM..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/mcp-server-macos-arm

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/mcp-server-linux

echo "Builds complete!"
ls -lh dist/
```

#### GitHub Actions Workflow (.github/workflows/release.yml)

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build binaries
        run: |
          mkdir -p dist

          # Windows
          GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/mcp-server-windows.exe

          # macOS Intel
          GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/mcp-server-macos-intel

          # macOS ARM
          GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/mcp-server-macos-arm

          # Linux
          GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/mcp-server-linux

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 5.4 Distribution Strategy: **Pre-Built Binaries with External Config**

#### Option A: Pre-Built Binaries with Config File (RECOMMENDED)

**How it works:**
1. PWA generates `config.json` with user's API keys
2. User downloads platform-specific binary + config file as ZIP
3. User extracts ZIP and runs executable
4. Executable reads config.json from same directory

**Pros:**
- ✅ Simple implementation
- ✅ Fast download (5-10MB)
- ✅ User can edit config later
- ✅ No security issues (config outside binary)
- ✅ Can update config without re-downloading

**Cons:**
- Two files instead of one (but in ZIP)

**Implementation:**
```javascript
// PWA Backend (Node.js/Express)
app.post('/api/download-mcp-server', async (req, res) => {
  const { api_key, budget_key, platform } = req.body;

  // Generate config
  const config = {
    api_key,
    budget_key,
    api_url: 'https://api.example.com'
  };

  // Create ZIP with binary + config
  const archive = archiver('zip');

  // Add pre-built binary
  const binaryPath = `./binaries/mcp-server-${platform}`;
  archive.file(binaryPath, { name: 'mcp-server' });

  // Add config
  archive.append(JSON.stringify(config, null, 2), { name: 'config.json' });

  // Add README
  archive.append(readmeContent, { name: 'README.txt' });

  res.attachment('mcp-server.zip');
  archive.pipe(res);
  archive.finalize();
});
```

**File Structure in ZIP:**
```
mcp-server.zip
├── mcp-server           (or .exe on Windows)
├── config.json          (user-specific)
└── README.txt           (setup instructions)
```

---

#### Option B: Embedded Config (Alternative)

**How it works:**
1. PWA sends API keys to backend
2. Backend builds custom binary with embedded config
3. User downloads personalized executable

**Pros:**
- Single file
- Config can't be separated

**Cons:**
- ❌ Must recompile for each user (slow, 1-2 seconds per build)
- ❌ Increased server load
- ❌ User can't update config without re-downloading
- ❌ More complex implementation

**Recommendation:** Only use if single-file is absolutely required

---

#### Option C: Environment Variables

**How it works:**
1. User downloads generic binary
2. User sets environment variables
3. Executable reads from environment

**Pros:**
- Single generic binary
- Standard approach

**Cons:**
- ❌ Poor UX (users must set env vars)
- ❌ Platform-specific instructions
- ❌ Easy to mess up

**Recommendation:** Avoid for non-technical users

---

### 5.5 PWA User Flow (Recommended Approach)

```
1. User enters API Key + Budget Key in PWA form

2. User selects their platform:
   - Windows
   - macOS (Intel)
   - macOS (Apple Silicon)
   - Linux

3. PWA generates download:
   - Creates config.json with user's keys
   - Packages pre-built binary + config into ZIP
   - Serves ZIP download (~5-10MB)

4. User downloads and extracts ZIP

5. User follows simple setup:

   Windows:
   - Extract ZIP to desired location
   - Run setup.bat (adds to Claude Desktop config)
   - Or manually add to Claude Desktop

   macOS/Linux:
   - Extract ZIP
   - chmod +x mcp-server
   - Run ./setup.sh (adds to config)
   - Or manually add to Claude Desktop

6. User opens Claude Desktop
   - MCP server auto-starts via stdio
   - Tools appear in Claude interface
```

---

### 5.6 Claude Desktop Integration

#### Windows Config
**Path:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "my-api-server": {
      "command": "C:\\Users\\Username\\MCPServer\\mcp-server.exe"
    }
  }
}
```

#### macOS Config
**Path:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "my-api-server": {
      "command": "/Users/username/MCPServer/mcp-server"
    }
  }
}
```

#### Linux Config
**Path:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "my-api-server": {
      "command": "/home/username/MCPServer/mcp-server"
    }
  }
}
```

---

### 5.7 Auto-Setup Scripts

#### Windows Setup (setup.bat)

```batch
@echo off
echo Setting up MCP Server for Claude Desktop...

set CONFIG_PATH=%APPDATA%\Claude\claude_desktop_config.json
set SERVER_PATH=%~dp0mcp-server.exe

echo Server path: %SERVER_PATH%
echo Config path: %CONFIG_PATH%

REM Create config directory if needed
if not exist "%APPDATA%\Claude" mkdir "%APPDATA%\Claude"

REM Add to Claude config (simplified - use PowerShell for JSON parsing)
powershell -Command "$config = if (Test-Path '%CONFIG_PATH%') { Get-Content '%CONFIG_PATH%' | ConvertFrom-Json } else { @{} }; if (-not $config.mcpServers) { $config | Add-Member -NotePropertyName mcpServers -NotePropertyValue @{} }; $config.mcpServers.'my-api-server' = @{ command = '%SERVER_PATH%' }; $config | ConvertTo-Json -Depth 10 | Set-Content '%CONFIG_PATH%'"

echo.
echo Setup complete! Restart Claude Desktop to use the server.
pause
```

#### macOS/Linux Setup (setup.sh)

```bash
#!/bin/bash

echo "Setting up MCP Server for Claude Desktop..."

CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
SERVER_PATH="$(cd "$(dirname "$0")" && pwd)/mcp-server"

# Make executable
chmod +x "$SERVER_PATH"

echo "Server path: $SERVER_PATH"
echo "Config path: $CONFIG_PATH"

# Create config directory
mkdir -p "$(dirname "$CONFIG_PATH")"

# Add to Claude config (using jq for JSON manipulation)
if [ -f "$CONFIG_PATH" ]; then
  # Update existing config
  jq ".mcpServers[\"my-api-server\"] = {\"command\": \"$SERVER_PATH\"}" "$CONFIG_PATH" > "$CONFIG_PATH.tmp"
  mv "$CONFIG_PATH.tmp" "$CONFIG_PATH"
else
  # Create new config
  cat > "$CONFIG_PATH" << EOF
{
  "mcpServers": {
    "my-api-server": {
      "command": "$SERVER_PATH"
    }
  }
}
EOF
fi

echo ""
echo "Setup complete! Restart Claude Desktop to use the server."
```

---

## 6. File Size Estimates

### Recommended Approach (Go)

| Platform | Binary Size | + Config | In ZIP | Notes |
|----------|-------------|----------|--------|-------|
| Windows x64 | 6-8MB | +1KB | ~3MB | Compressed in ZIP |
| macOS Intel | 6-8MB | +1KB | ~3MB | Compressed |
| macOS ARM | 6-8MB | +1KB | ~3MB | Compressed |
| Linux x64 | 6-8MB | +1KB | ~3MB | Compressed |

**Total storage for all platforms:** ~30MB (uncompressed), ~12MB (compressed)

---

### Alternative Approaches (For Comparison)

| Approach | Windows | macOS | Linux | Notes |
|----------|---------|-------|-------|-------|
| FastMCP + PyInstaller | 40-80MB | 40-80MB | 40-80MB | Per platform |
| TypeScript + pkg | 55MB | 55MB | 55MB | Per platform |
| Rust (optimized) | 1-2MB | 1-2MB | 1-2MB | Smallest but complex |
| Tauri (with GUI) | 8-12MB | 8-12MB | 8-12MB | If GUI needed |

---

## 7. Security Considerations

### API Key Storage

#### External Config File (Recommended)
**Security Level:** Medium-High

✅ **Pros:**
- Keys not embedded in binary
- User can secure config file with filesystem permissions
- Easy to rotate keys (edit config file)
- Config can be excluded from backups

⚠️ **Cons:**
- Keys stored in plaintext JSON
- User must protect file permissions

**Mitigation:**
- Set restrictive file permissions (600 on Unix, similar on Windows)
- Educate users to keep config file secure
- Consider encrypting config file (adds complexity)

---

#### Embedded in Binary
**Security Level:** Low

❌ **Cons:**
- Keys can be extracted via strings/hex editor
- Must rebuild binary to rotate keys
- Harder to secure

**Only use if:** Single-file distribution is absolutely required

---

#### Environment Variables
**Security Level:** Medium

✅ **Pros:**
- Not in version control
- Standard approach

⚠️ **Cons:**
- Visible in process list (`ps aux` shows env vars in some cases)
- Complex for non-technical users

---

### API Security Best Practices

1. **Rate Limiting:** Implement on backend API
2. **Key Restrictions:** Restrict API keys by IP/origin if possible
3. **Budget Limits:** Enforce on backend (budget_key parameter)
4. **Key Rotation:** Make it easy to update config.json
5. **Monitoring:** Log API usage for abuse detection

---

### Executable Security

**Go advantages:**
- No anti-virus false positives (unlike PyInstaller)
- No interpreter to exploit
- Standard OS executable format
- Can be code-signed (Windows/macOS)

**Code Signing (Optional but Recommended):**
```bash
# Windows (requires certificate)
signtool sign /f certificate.pfx /p password mcp-server.exe

# macOS (requires Apple Developer account)
codesign --sign "Developer ID" mcp-server
```

---

## 8. Update Mechanism

### Simple Version Check

**server.go addition:**
```go
const VERSION = "1.0.0"
const UPDATE_CHECK_URL = "https://api.example.com/mcp-server/version"

func checkForUpdates() {
    resp, err := http.Get(UPDATE_CHECK_URL)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    latestVersion := string(body)

    if latestVersion != VERSION {
        fmt.Fprintf(os.Stderr, "⚠️  Update available: %s (you have %s)\n", latestVersion, VERSION)
        fmt.Fprintf(os.Stderr, "   Download: https://example.com/download\n")
    }
}

func main() {
    go checkForUpdates() // Run in background
    // ... rest of server code
}
```

**Trade-off:** Users must manually download new version

---

### Auto-Update (Optional, More Complex)

Use a library like `go-update` or `equinox`:
```go
import "github.com/inconshreveable/go-update"

func autoUpdate() error {
    resp, err := http.Get("https://example.com/binaries/latest/mcp-server")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    err = update.Apply(resp.Body, update.Options{})
    if err != nil {
        return err
    }

    fmt.Println("✓ Updated successfully! Restart the server.")
    return nil
}
```

**Recommendation:** Start with version check, add auto-update later if needed

---

## 9. Documentation & Examples

### Key Resources

#### Official MCP Docs
- Specification: https://modelcontextprotocol.io/
- Building servers: https://modelcontextprotocol.io/quickstart/server

#### Go MCP SDK
- Official SDK: https://github.com/modelcontextprotocol/go-sdk
- Community (mark3labs): https://github.com/mark3labs/mcp-go
- Community (metoro): https://github.com/metoro-io/mcp-golang

#### FastMCP (Python Alternative)
- Website: https://gofastmcp.com/
- GitHub: https://github.com/jlowin/fastmcp
- Tutorial: https://www.datacamp.com/tutorial/building-mcp-server-client-fastmcp

#### Real-World Examples
- GitHub MCP Server (Go): https://github.com/github/github-mcp-server
- SQL Server MCP (C#): https://github.com/ian-cowley/MCPSqlServer
- Awesome MCP Servers: https://github.com/punkpeye/awesome-mcp-servers

#### Go Resources
- Cross-compilation: https://opensource.com/article/21/1/go-cross-compiling
- Binary optimization: https://blog.howardjohn.info/posts/go-binary-size/

---

## 10. Implementation Timeline

### Phase 1: MVP (1-2 weeks)
- [ ] Set up Go project with MCP SDK
- [ ] Implement config loading from JSON
- [ ] Implement REST API tool fetching
- [ ] Implement tool execution via REST API
- [ ] Test with Claude Desktop locally
- [ ] Create build script for all platforms

### Phase 2: Distribution (1 week)
- [ ] Create PWA backend endpoint for generating downloads
- [ ] Implement ZIP packaging (binary + config)
- [ ] Create setup scripts (setup.bat, setup.sh)
- [ ] Write user documentation
- [ ] Test on all platforms

### Phase 3: CI/CD (3-4 days)
- [ ] Set up GitHub Actions workflow
- [ ] Automate builds on tag push
- [ ] Automated releases with binaries
- [ ] Version numbering strategy

### Phase 4: Polish (1 week)
- [ ] Add version check mechanism
- [ ] Improve error messages
- [ ] Add logging (to stderr, not stdout!)
- [ ] Code signing (Windows/macOS)
- [ ] Final testing & bug fixes

**Total:** 4-5 weeks for complete implementation

---

## 11. Alternative Recommendation: FastMCP for Rapid Prototyping

**If your team:**
- Already knows Python well
- Wants to prototype quickly
- Can accept 40-80MB binaries
- Prioritizes development speed over distribution size

**Then use:**
- Framework: FastMCP 2.0
- Packaging: PyInstaller with `--onefile`
- Distribution: Pre-built binaries per platform

**FastMCP advantages:**
- Faster initial development (more concise code)
- Better for dynamic tool registration
- Easier debugging
- Can switch to Go later if needed

**Example FastMCP Implementation:**

```python
# server.py
import os
import json
import httpx
from fastmcp import FastMCP

# Load config
with open('config.json') as f:
    config = json.load(f)

mcp = FastMCP("API Proxy Server")

# Fetch tools at startup
async def fetch_api_tools():
    async with httpx.AsyncClient() as client:
        resp = await client.get(
            f"{config['api_url']}/products/fetch",
            headers={
                "X-API-Key": config['api_key'],
                "X-Budget-Key": config['budget_key']
            }
        )
        return resp.json()

# Dynamic tool registration
tools = asyncio.run(fetch_api_tools())

for tool_def in tools:
    @mcp.tool()
    async def api_tool(**kwargs):
        """Dynamically registered tool"""
        async with httpx.AsyncClient() as client:
            resp = await client.post(
                f"{config['api_url']}/products/purchase",
                json={"tool": tool_def['name'], "args": kwargs},
                headers={
                    "X-API-Key": config['api_key'],
                    "X-Budget-Key": config['budget_key']
                }
            )
            return resp.json()

if __name__ == "__main__":
    mcp.run(transport='stdio')
```

**Build with PyInstaller:**
```bash
pip install pyinstaller fastmcp httpx
pyinstaller --onefile --name mcp-server server.py
```

**Result:** 40-80MB executable (vs Go's 6-8MB)

---

## 12. Final Recommendation Summary

### For Production (Recommended): **Go + Official MCP SDK**

**Why:**
- 5-10MB binaries (85-90% smaller than Python)
- Trivial cross-compilation (built into language)
- Professional user experience (fast, lightweight)
- No anti-virus issues
- Simple distribution (single binary + config)
- Official SDK with Google collaboration

**When to use:**
- You want the best user experience
- File size matters
- You can invest 1-2 weeks learning Go
- Production-ready from day one

---

### For Rapid Prototyping: **FastMCP + PyInstaller**

**Why:**
- Fastest development time
- Most concise code
- Best dynamic tool registration
- Great documentation

**When to use:**
- Team already knows Python
- Prototype/MVP stage
- Can accept 40-80MB binaries
- Want to move fast

**Migration path:** FastMCP → Go later if needed

---

### For Maximum Optimization: **Rust + rmcp**

**Why:**
- Smallest binaries (1-3MB)
- Highest performance
- Memory safety

**When to use:**
- File size is critical
- Performance is critical
- Team has Rust expertise
- Have time for longer development

---

## 13. Questions & Answers

**Q: Can Go handle dynamic tool registration from REST API?**
A: Yes, as shown in the example code. You fetch tool definitions at startup and register them dynamically.

**Q: How do users update their API keys?**
A: They edit `config.json` and restart the server. Simple and standard.

**Q: What if the REST API is slow to respond with tools?**
A: Add a loading state or cache the tools locally. You can also fetch tools asynchronously after server starts.

**Q: Can one executable work for all platforms?**
A: No, you need separate binaries per platform (Windows .exe, macOS, Linux). But Go makes this trivial with cross-compilation.

**Q: How do I handle API versioning?**
A: Include version in REST API response and in MCP server version string. Check compatibility at startup.

**Q: What about M1/M2/M3 Macs vs Intel Macs?**
A: Build two macOS binaries: one for `amd64` (Intel) and one for `arm64` (Apple Silicon). Rosetta can run Intel binaries on ARM, but native is better.

**Q: Can I use this for VS Code / Cursor / other MCP clients?**
A: Yes! stdio transport works with all MCP clients. Each client has its own config file location.

**Q: How do I debug the MCP server?**
A: Log to stderr (NOT stdout - that corrupts stdio transport). Use `fmt.Fprintf(os.Stderr, "...")`

**Q: Can I embed the config inside the binary?**
A: Yes, using `go:embed` directive, but external config file is better for security and flexibility.

---

## 14. Conclusion

**Bottom Line:** Use **Go with the official MCP SDK** for the best combination of simplicity, performance, and user experience.

**File size:** 5-10MB per platform (vs 40-150MB for Python/Node)
**Build time:** ~5 seconds for all platforms
**Runtime deps:** Zero
**User experience:** Download, extract, run - done

**Action Items:**
1. Learn basic Go (1-2 days if new to Go)
2. Implement MCP server with REST API proxy (3-5 days)
3. Set up build scripts and GitHub Actions (1-2 days)
4. Create PWA download endpoint (2-3 days)
5. Test on all platforms (1-2 days)
6. Launch! 🚀

**Total time to production:** 2-3 weeks with Go, including learning curve

---

**Report compiled:** October 9, 2025
**Research sources:** 40+ web searches across official docs, GitHub repos, tutorials, and community discussions
