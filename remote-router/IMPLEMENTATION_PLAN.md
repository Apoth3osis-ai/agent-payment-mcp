# AgentPMT Remote-First MCP Router - Implementation Plan

**Goal:** Build a minimal, secure, signed native MCP server that acts as a stdio JSON-RPC router forwarding requests to the remote AgentPMT HTTPS API. No privileged OS access, no local shell execution.

**Status:** Planning Phase
**Created:** October 16, 2025
**Based on:** wasm.md specifications

---

## Executive Summary

This implementation transforms the current AgentPMT MCP server into a lightweight remote-first router:

- **Tiny binaries** (<10MB) with stripped symbols
- **Zero privileged operations** - HTTPS outbound only
- **Raw JSON schema preservation** - no SDK re-marshaling
- **Optional SSE streaming** - for real-time responses
- **Code-signed** Windows binaries with optional MSIX packaging
- **Cross-platform** - Windows, macOS (Intel/ARM), Linux (x64/ARM)

---

## Architecture Overview

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│  Claude/Cursor  │ stdio   │  MCP Router      │  HTTPS  │  AgentPMT API   │
│  (JSON-RPC)     │◄──────►│  (Go binary)     │◄──────►│  (Remote)       │
└─────────────────┘         └──────────────────┘         └─────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │  config.json     │
                            │  (API keys)      │
                            └──────────────────┘
```

**Key Principles:**
1. Server speaks stdio JSON-RPC to client
2. All tool operations proxy to HTTPS API
3. No local code execution or file system access
4. Raw JSON schemas preserved from API
5. Deterministic, signed binaries

---

## Implementation Steps (12 Phases)

### Phase 0: Project Setup ✅

**Folder Structure:**
```
remote-router/
├── cmd/
│   └── agent-payment-router/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── api/
│   │   ├── client.go
│   │   └── sse.go
│   └── mcp/
│       ├── server.go
│       └── rpc.go
├── scripts/
│   ├── build-all.sh
│   ├── sign-windows.ps1
│   └── package-msix.ps1
├── windows/
│   └── AppxManifest.xml
├── distribution/
│   ├── binaries/
│   └── packages/
├── tests/
│   └── e2e_test.go
├── go.mod
├── go.sum
└── README.md
```

---

### Phase 1: Config and Environment Variables

**File:** `internal/config/config.go`

**Objectives:**
- Default `APIURL` to `https://api.agentpmt.com`
- Support environment variable overrides (non-breaking)
- Maintain backward compatibility with `config.json`

**Environment Variables:**
```go
AGENTPMT_API_URL      // Overrides APIURL
AGENTPMT_API_KEY      // Overrides APIKey
AGENTPMT_BUDGET_KEY   // Overrides BudgetKey
```

**Implementation:**
```go
func Load() (*Config, error) {
    // Load from config.json first
    cfg := loadFromFile()

    // Environment overrides (optional, non-breaking)
    if v := os.Getenv("AGENTPMT_API_URL"); v != "" {
        cfg.APIURL = v
    }
    if v := os.Getenv("AGENTPMT_API_KEY"); v != "" {
        cfg.APIKey = v
    }
    if v := os.Getenv("AGENTPMT_BUDGET_KEY"); v != "" {
        cfg.BudgetKey = v
    }

    // Set default if still empty
    if cfg.APIURL == "" {
        cfg.APIURL = "https://api.agentpmt.com"
    }

    return cfg, nil
}
```

**Acceptance Criteria:**
- ✅ `config.json` continues to work unchanged
- ✅ Environment variables can override for CI/testing
- ✅ Defaults to production API URL when empty

**Commit Message:** `feat(config): add client constructor, UA, env overrides`

---

### Phase 2: HTTP Client Hardening

**File:** `internal/api/client.go`

**Objectives:**
- Create proper HTTP client with timeouts
- Set clear User-Agent header
- Keep outbound HTTPS only (no local listeners)

**Implementation:**
```go
var DefaultUA = "AgentPMT-MCP/1.0"

type Client struct {
    baseURL    string
    apiKey     string
    budgetKey  string
    http       *http.Client
}

func NewClient(baseURL, apiKey, budgetKey string) *Client {
    if baseURL == "" {
        baseURL = "https://api.agentpmt.com"
    }

    return &Client{
        baseURL:   baseURL,
        apiKey:    apiKey,
        budgetKey: budgetKey,
        http: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
    req.Header.Set("User-Agent", DefaultUA)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.apiKey)
    req.Header.Set("X-Budget-Key", c.budgetKey)
    return c.http.Do(req)
}
```

**Security Considerations:**
- ✅ HTTPS only (no HTTP fallback)
- ✅ 60-second timeout prevents hanging
- ✅ Clear User-Agent for API analytics
- ✅ Consistent header injection

**Acceptance Criteria:**
- ✅ All requests include User-Agent header
- ✅ API key and budget key sent on every request
- ✅ No local listeners or privileged operations

**Commit Message:** `feat(api): add client constructor, UA, env overrides`

---

### Phase 3: Tools List Pass-Through (Raw JSON Schemas)

**Files:** `internal/api/client.go`, `internal/mcp/server.go`

**Objectives:**
- Fetch tools from `/products/fetch` endpoint
- Preserve raw JSON schemas without re-marshaling
- Map 1:1 to MCP `tools/list` response

**API Types:**
```go
type ToolDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  json.RawMessage `json:"parameters"` // Raw pass-through
}

type FetchToolsResponse struct {
    Success bool             `json:"success"`
    Tools   []ToolDefinition `json:"tools"`
    Error   string           `json:"error,omitempty"`
}
```

**API Client Method:**
```go
func (c *Client) FetchTools() ([]ToolDefinition, error) {
    req, _ := http.NewRequest("GET", c.baseURL+"/products/fetch", nil)
    resp, err := c.do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode/100 != 2 {
        return nil, fmt.Errorf("fetch failed: %s", string(body))
    }

    var out FetchToolsResponse
    if err := json.Unmarshal(body, &out); err != nil {
        return nil, err
    }

    if !out.Success {
        return nil, fmt.Errorf("fetch error: %s", out.Error)
    }

    return out.Tools, nil
}
```

**MCP Handler:**
```go
func (s *Server) handleToolsList(id interface{}) JSONRPCResponse {
    tools, err := s.apiClient.FetchTools()
    if err != nil {
        return jsonErr(id, -32000, err.Error())
    }

    // Map to MCP response with raw schema
    var mcpTools []map[string]any
    for _, t := range tools {
        m := map[string]any{
            "name":        t.Name,
            "description": t.Description,
            "inputSchema": json.RawMessage(t.Parameters), // Raw pass-through!
        }
        mcpTools = append(mcpTools, m)
    }

    return jsonOK(id, map[string]any{"tools": mcpTools})
}
```

**Acceptance Criteria:**
- ✅ Tools fetched from remote API
- ✅ JSON schemas preserved exactly as received
- ✅ Claude/Cursor display tools with original schemas
- ✅ No SDK re-marshaling

**Commit Message:** `feat(mcp): preserve raw JSON schemas in tools/list`

---

### Phase 4: Tool Invocation + Optional SSE Streaming

**Files:** `internal/api/client.go`, `internal/api/sse.go`

#### 4A. Synchronous Purchase

**Types:**
```go
type PurchaseRequest struct {
    ProductID  string          `json:"product_id"`
    Parameters json.RawMessage `json:"parameters"`
}

type PurchaseResponse struct {
    Success bool   `json:"success"`
    Output  string `json:"output,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

**Implementation:**
```go
func (c *Client) Purchase(req PurchaseRequest) (*PurchaseResponse, error) {
    payload, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", c.baseURL+"/products/purchase", bytes.NewReader(payload))

    resp, err := c.do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode/100 != 2 {
        return nil, fmt.Errorf("purchase failed: %s", string(body))
    }

    var out PurchaseResponse
    if err := json.Unmarshal(body, &out); err != nil {
        return nil, err
    }

    if !out.Success {
        return nil, fmt.Errorf(out.Error)
    }

    return &out, nil
}
```

#### 4B. Optional SSE Streaming

**File:** `internal/api/sse.go`

**Research Finding:** Use `tmaxmax/go-sse` library (modern, Go 1.23 iterators)

```go
package api

import (
    "context"
    "net/http"

    "github.com/tmaxmax/go-sse"
)

func (c *Client) StreamPurchase(ctx context.Context, req PurchaseRequest, onChunk func(string)) error {
    payload, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/products/purchase?stream=true", bytes.NewReader(payload))
    httpReq.Header.Set("Accept", "text/event-stream")

    resp, err := c.do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Parse SSE stream
    for event, err := range sse.Read(resp.Body, nil) {
        if err != nil {
            return err
        }

        onChunk(event.Data)
    }

    return nil
}
```

**MCP Integration:**
```go
func (s *Server) handleToolsCall(id interface{}, params map[string]interface{}) JSONRPCResponse {
    toolName := params["name"].(string)
    args := params["arguments"].(map[string]interface{})

    // Check if streaming requested
    streaming := false
    if s, ok := args["stream"].(bool); ok {
        streaming = s
    }

    req := PurchaseRequest{
        ProductID:  toolName,
        Parameters: json.RawMessage(mustMarshal(args)),
    }

    if streaming {
        // Stream response
        var chunks []string
        err := s.apiClient.StreamPurchase(context.Background(), req, func(chunk string) {
            chunks = append(chunks, chunk)
            // TODO: Optionally send partial MCP responses
        })

        if err != nil {
            return jsonErr(id, -32000, err.Error())
        }

        return jsonOK(id, map[string]any{
            "content": []map[string]any{
                {"type": "text", "text": strings.Join(chunks, "")},
            },
        })
    }

    // Synchronous response
    resp, err := s.apiClient.Purchase(req)
    if err != nil {
        return jsonErr(id, -32000, err.Error())
    }

    return jsonOK(id, map[string]any{
        "content": []map[string]any{
            {"type": "text", "text": resp.Output},
        },
    })
}
```

**Dependencies:**
```bash
go get github.com/tmaxmax/go-sse
```

**Acceptance Criteria:**
- ✅ Synchronous purchase works for all tools
- ✅ Optional streaming with `stream: true` parameter
- ✅ MCP receives text content in standard format
- ✅ Errors returned as MCP error objects (not Go errors)

**Commit Message:** `feat(api): implement purchase call + optional SSE streaming`

---

### Phase 5: Stdio JSON-RPC Loop

**Files:** `internal/mcp/rpc.go`, `internal/mcp/server.go`

**Objectives:**
- Minimal stdio loop (no SDK dependencies)
- Newline-delimited JSON-RPC 2.0
- Support: `initialize`, `tools/list`, `tools/call`, `notifications/initialized`

**Main Loop:**
```go
func (s *Server) HandleStdioTransport() error {
    scanner := bufio.NewScanner(os.Stdin)
    encoder := json.NewEncoder(os.Stdout)

    for scanner.Scan() {
        line := scanner.Bytes()

        var req JSONRPCRequest
        if err := json.Unmarshal(line, &req); err != nil {
            log.Printf("Error parsing request: %v", err)
            continue
        }

        var response JSONRPCResponse

        switch req.Method {
        case "initialize":
            response = s.handleInitialize(req.ID)
        case "tools/list":
            response = s.handleToolsList(req.ID)
        case "tools/call":
            response = s.handleToolsCall(req.ID, req.Params)
        case "notifications/initialized":
            // No response for notifications
            continue
        case "resources/list":
            // Return empty (not supported)
            response = jsonOK(req.ID, map[string]any{"resources": []any{}})
        default:
            response = jsonErr(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
        }

        if err := encoder.Encode(response); err != nil {
            log.Printf("Error encoding response: %v", err)
        }
    }

    return scanner.Err()
}
```

**Initialize Handler:**
```go
func (s *Server) handleInitialize(id interface{}) JSONRPCResponse {
    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      id,
        Result: map[string]interface{}{
            "protocolVersion": "2025-03-26",
            "capabilities": map[string]interface{}{
                "tools": map[string]interface{}{
                    "listChanged": true,
                },
            },
            "serverInfo": map[string]interface{}{
                "name":    "agent-payment-router",
                "version": version.Version,
            },
        },
    }
}
```

**Logging Best Practice:**
```go
// All logging to stderr (stdout reserved for JSON-RPC)
log.SetOutput(os.Stderr)
log.SetPrefix("[AgentPMT] ")

// Redact secrets
func sanitizeLog(msg string, apiKey string) string {
    if apiKey != "" {
        msg = strings.ReplaceAll(msg, apiKey, "***REDACTED***")
    }
    return msg
}
```

**Acceptance Criteria:**
- ✅ Works with Claude Desktop/Cursor/VS Code
- ✅ All responses are newline-delimited JSON
- ✅ Logging goes to stderr only
- ✅ Secrets redacted from logs

**Commit Message:** `feat(mcp): implement minimal stdio JSON-RPC loop`

---

### Phase 6: Build Small, Deterministic Binaries

**File:** `scripts/build-all.sh`

**Objectives:**
- Cross-compile for 5 platforms
- Strip symbols (`-s -w`)
- Embed version string
- Binaries < 10MB each

**Build Script:**
```bash
#!/bin/bash
set -e

VERSION=${1:-dev}

echo "Building AgentPMT Router v${VERSION}..."

# Ensure output directories exist
mkdir -p distribution/binaries/{windows-amd64,linux-amd64,linux-arm64,darwin-amd64,darwin-arm64}

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION}"

# Windows AMD64
echo "Building Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
  -trimpath \
  -ldflags="${LDFLAGS}" \
  -o distribution/binaries/windows-amd64/agent-payment-router.exe \
  ./cmd/agent-payment-router

# Linux AMD64
echo "Building Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="${LDFLAGS}" \
  -o distribution/binaries/linux-amd64/agent-payment-router \
  ./cmd/agent-payment-router

# Linux ARM64
echo "Building Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
  -trimpath \
  -ldflags="${LDFLAGS}" \
  -o distribution/binaries/linux-arm64/agent-payment-router \
  ./cmd/agent-payment-router

# macOS Intel
echo "Building macOS Intel..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
  -trimpath \
  -ldflags="${LDFLAGS}" \
  -o distribution/binaries/darwin-amd64/agent-payment-router \
  ./cmd/agent-payment-router

# macOS Apple Silicon
echo "Building macOS ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
  -trimpath \
  -ldflags="${LDFLAGS}" \
  -o distribution/binaries/darwin-arm64/agent-payment-router \
  ./cmd/agent-payment-router

echo "Build complete!"
ls -lh distribution/binaries/*/agent-payment-router*
```

**Version Embedding:**
```go
// cmd/agent-payment-router/main.go
package main

var Version = "dev" // Set by -ldflags at build time

func main() {
    log.Printf("AgentPMT Router v%s starting...", Version)
    // ...
}
```

**Acceptance Criteria:**
- ✅ All binaries < 10MB
- ✅ Version embedded and displayed at startup
- ✅ Deterministic builds (same input = same output)
- ✅ No CGO dependencies

**Commit Message:** `chore(build): strip symbols, embed version, cross-compile`

---

### Phase 7: Windows Packaging & Signing

**Research Findings:**
- Azure Trusted Signing: $9.99/month, instant SmartScreen reputation (US/Canada only)
- Traditional EV certs: No longer provide instant reputation bypass
- MSIX: Code signing mandatory, better Windows integration
- SmartScreen reputation: 2-8 weeks building period

#### 7A. Code Signing Script

**File:** `scripts/sign-windows.ps1`

```powershell
param(
    [string]$File = "distribution/binaries/windows-amd64/agent-payment-router.exe",
    [string]$CertThumbprint = $env:CODESIGN_THUMBPRINT,
    [string]$TimestampUrl = "http://timestamp.digicert.com"
)

if (-not $CertThumbprint) {
    Write-Error "Certificate thumbprint required. Set CODESIGN_THUMBPRINT env var or pass -CertThumbprint"
    exit 1
}

Write-Host "Signing $File..."

# Sign with SHA256
& signtool sign `
    /fd SHA256 `
    /tr $TimestampUrl `
    /td SHA256 `
    /sha1 $CertThumbprint `
    $File

if ($LASTEXITCODE -ne 0) {
    Write-Error "Signing failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host "Successfully signed $File"

# Verify signature
& signtool verify /pa $File
```

#### 7B. MSIX Packaging (Optional)

**File:** `windows/AppxManifest.xml`

```xml
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10"
         xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10">
  <Identity Name="com.agentpmt.mcp-router"
            Publisher="CN=AgentPMT"
            Version="1.0.0.0" />

  <Properties>
    <DisplayName>AgentPMT MCP Router</DisplayName>
    <PublisherDisplayName>AgentPMT</PublisherDisplayName>
    <Logo>agent-payment-logo.png</Logo>
  </Properties>

  <Dependencies>
    <TargetDeviceFamily Name="Windows.Desktop" MinVersion="10.0.17763.0" MaxVersionTested="10.0.22621.0" />
  </Dependencies>

  <Resources>
    <Resource Language="en-us" />
  </Resources>

  <Applications>
    <Application Id="AgentPMTRouter" Executable="agent-payment-router.exe" EntryPoint="Windows.FullTrustApplication">
      <uap:VisualElements DisplayName="AgentPMT MCP Router"
                          Description="Secure USDC payments for AI agents"
                          Square150x150Logo="agent-payment-logo.png"
                          Square44x44Logo="agent-payment-logo.png"
                          BackgroundColor="transparent" />
    </Application>
  </Applications>

  <Capabilities>
    <Capability Name="internetClient" />
  </Capabilities>
</Package>
```

**File:** `scripts/package-msix.ps1`

```powershell
param(
    [string]$Version = "1.0.0",
    [string]$CertThumbprint = $env:CODESIGN_THUMBPRINT
)

$Out = "distribution/packages/agent-payment-router-v${Version}.msix"
$Tmp = "distribution/msix-content"

# Clean and prepare
Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $Tmp | Out-Null

# Copy files
Copy-Item distribution/binaries/windows-amd64/agent-payment-router.exe $Tmp
Copy-Item agent-payment-logo.png $Tmp
Copy-Item windows/AppxManifest.xml $Tmp

# Update version in manifest
(Get-Content "$Tmp/AppxManifest.xml") -replace '1\.0\.0\.0', "${Version}.0" | Set-Content "$Tmp/AppxManifest.xml"

# Create MSIX
Write-Host "Creating MSIX package..."
& makeappx.exe pack /d $Tmp /p $Out /l

if ($LASTEXITCODE -ne 0) {
    Write-Error "MSIX packaging failed"
    exit $LASTEXITCODE
}

# Sign MSIX
Write-Host "Signing MSIX..."
& signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /sha1 $CertThumbprint $Out

if ($LASTEXITCODE -ne 0) {
    Write-Error "MSIX signing failed"
    exit $LASTEXITCODE
}

Write-Host "Successfully created and signed $Out"
```

**Acceptance Criteria:**
- ✅ Windows EXE is code-signed with timestamping
- ✅ MSIX package created and signed (optional)
- ✅ Signature verification passes
- ✅ Files ready for distribution

**Commit Message:** `feat(win): add signing and optional MSIX packaging scripts`

---

### Phase 8: Installer Refresh

**Files:**
- `distribution/templates/install-windows.ps1`
- `distribution/templates/install-linux.sh`
- `distribution/templates/install-macos.sh`
- `distribution/templates/mcpb-manifest.json`

**Objectives:**
- Update to use new router binary names
- Place config.json next to binary (no hidden folders)
- Keep `.mcpb` packaging working
- Maintain detector compatibility

**Windows Installer Template:**
```powershell
# install-windows.ps1
$ErrorActionPreference = "Stop"

# Detect Claude/Cursor installation
$claudePath = "$env:APPDATA\Claude\servers\agent-payment"
$cursorPath = "$env:USERPROFILE\.cursor\mcp\servers\agent-payment"

# User choice or auto-detect
$installPath = $claudePath  # Or $cursorPath

# Create directory
New-Item -ItemType Directory -Path $installPath -Force | Out-Null

# Download binary (or copy from package)
$binaryUrl = "https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-windows-amd64.exe"
Invoke-WebRequest -Uri $binaryUrl -OutFile "$installPath\agent-payment-router.exe"

# Create config.json
@{
    APIURL = "https://api.agentpmt.com"
    APIKey = ""
    BudgetKey = ""
} | ConvertTo-Json | Set-Content "$installPath\config.json"

# Create MCP manifest
$manifest = @{
    mcpServers = @{
        "agent-payment" = @{
            command = "$installPath\agent-payment-router.exe"
            args = @()
            env = @{
                AGENTPMT_API_KEY = "your-api-key-here"
                AGENTPMT_BUDGET_KEY = "your-budget-key-here"
            }
        }
    }
}
$manifest | ConvertTo-Json -Depth 10 | Set-Content "$env:APPDATA\Claude\claude_desktop_config.json"

Write-Host "Installation complete! Restart Claude Desktop."
```

**Linux/macOS Installer Template:**
```bash
#!/bin/bash
set -e

# Detect installation path
CLAUDE_PATH="$HOME/.config/Claude/servers/agent-payment"
CURSOR_PATH="$HOME/.cursor/mcp/servers/agent-payment"

INSTALL_PATH="${CLAUDE_PATH}"  # Or $CURSOR_PATH

# Create directory
mkdir -p "$INSTALL_PATH"

# Download binary
BINARY_URL="https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
curl -L "$BINARY_URL" -o "$INSTALL_PATH/agent-payment-router"
chmod +x "$INSTALL_PATH/agent-payment-router"

# Create config.json
cat > "$INSTALL_PATH/config.json" <<EOF
{
  "APIURL": "https://api.agentpmt.com",
  "APIKey": "",
  "BudgetKey": ""
}
EOF

# Create MCP manifest
cat > "$HOME/.config/Claude/claude_desktop_config.json" <<EOF
{
  "mcpServers": {
    "agent-payment": {
      "command": "$INSTALL_PATH/agent-payment-router",
      "args": [],
      "env": {
        "AGENTPMT_API_KEY": "your-api-key-here",
        "AGENTPMT_BUDGET_KEY": "your-budget-key-here"
      }
    }
  }
}
EOF

echo "Installation complete! Restart Claude Desktop."
```

**MCPB Manifest Update:**
```json
{
  "name": "agent-payment",
  "version": "{{VERSION}}",
  "description": "Secure USDC payments for AI agents",
  "platforms": {
    "windows-amd64": {
      "binary": "agent-payment-router.exe",
      "sha256": "{{SHA256_WINDOWS}}"
    },
    "linux-amd64": {
      "binary": "agent-payment-router",
      "sha256": "{{SHA256_LINUX}}"
    },
    "darwin-amd64": {
      "binary": "agent-payment-router",
      "sha256": "{{SHA256_DARWIN_AMD64}}"
    },
    "darwin-arm64": {
      "binary": "agent-payment-router",
      "sha256": "{{SHA256_DARWIN_ARM64}}"
    }
  }
}
```

**Acceptance Criteria:**
- ✅ Installers drop binary + config in correct location
- ✅ MCP manifest points to new binary
- ✅ `.mcpb` packages build successfully
- ✅ Claude/Cursor detect server after restart

**Commit Message:** `chore(installer): point templates at new binaries; keep config side-by-side`

---

### Phase 9: CI Release Pipeline

**File:** `.github/workflows/release.yml`

**Objectives:**
- Build all platform binaries
- Sign Windows binaries (if secrets available)
- Package `.mcpb` bundles
- Create GitHub release
- Upload all artifacts

**Workflow:**
```yaml
name: Build and Release Remote Router

on:
  push:
    tags:
      - 'router-v*'
  workflow_dispatch:

jobs:
  build:
    name: Build Cross-Platform Binaries
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build all platforms
        run: |
          cd remote-router
          chmod +x scripts/build-all.sh
          ./scripts/build-all.sh ${GITHUB_REF#refs/tags/router-v}

      - name: Calculate SHA256 hashes
        id: hashes
        run: |
          cd remote-router/distribution/binaries
          echo "windows=$(sha256sum windows-amd64/agent-payment-router.exe | awk '{print $1}')" >> $GITHUB_OUTPUT
          echo "linux_amd64=$(sha256sum linux-amd64/agent-payment-router | awk '{print $1}')" >> $GITHUB_OUTPUT
          echo "linux_arm64=$(sha256sum linux-arm64/agent-payment-router | awk '{print $1}')" >> $GITHUB_OUTPUT
          echo "darwin_amd64=$(sha256sum darwin-amd64/agent-payment-router | awk '{print $1}')" >> $GITHUB_OUTPUT
          echo "darwin_arm64=$(sha256sum darwin-arm64/agent-payment-router | awk '{print $1}')" >> $GITHUB_OUTPUT

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: remote-router/distribution/binaries/

  sign-windows:
    name: Sign Windows Binary
    runs-on: windows-latest
    needs: build
    if: github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v4

      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          name: binaries
          path: remote-router/distribution/binaries/

      - name: Sign with Azure Trusted Signing
        uses: Azure/trusted-signing-action@v1
        with:
          azure-tenant-id: ${{ secrets.AZURE_TENANT_ID }}
          azure-client-id: ${{ secrets.AZURE_CLIENT_ID }}
          azure-client-secret: ${{ secrets.AZURE_CLIENT_SECRET }}
          endpoint: ${{ secrets.AZURE_CODE_SIGNING_ENDPOINT }}
          trusted-signing-account-name: ${{ secrets.AZURE_SIGNING_ACCOUNT }}
          certificate-profile-name: ${{ secrets.AZURE_CERT_PROFILE }}
          files-folder: remote-router/distribution/binaries/windows-amd64
          files-folder-filter: exe
          timestamp-rfc3161: http://timestamp.acs.microsoft.com
          timestamp-digest: SHA256

      - name: Upload signed binary
        uses: actions/upload-artifact@v4
        with:
          name: binaries-signed
          path: remote-router/distribution/binaries/

  release:
    name: Create GitHub Release
    runs-on: ubuntu-latest
    needs: [build, sign-windows]
    if: startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v4

      - name: Download signed binaries
        uses: actions/download-artifact@v4
        with:
          name: binaries-signed
          path: remote-router/distribution/binaries/

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            remote-router/distribution/binaries/windows-amd64/agent-payment-router.exe
            remote-router/distribution/binaries/linux-amd64/agent-payment-router
            remote-router/distribution/binaries/linux-arm64/agent-payment-router
            remote-router/distribution/binaries/darwin-amd64/agent-payment-router
            remote-router/distribution/binaries/darwin-arm64/agent-payment-router
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Required Secrets:**
- `AZURE_TENANT_ID`
- `AZURE_CLIENT_ID`
- `AZURE_CLIENT_SECRET`
- `AZURE_CODE_SIGNING_ENDPOINT`
- `AZURE_SIGNING_ACCOUNT`
- `AZURE_CERT_PROFILE`

**Acceptance Criteria:**
- ✅ Tagged builds trigger release
- ✅ Windows binary is code-signed
- ✅ All platform binaries uploaded to release
- ✅ SHA256 hashes calculated

**Commit Message:** `ci(release): sign, package .mcpb/.msix, upload assets`

---

### Phase 10: End-to-End Tests

**File:** `tests/e2e_test.go`

**Test Scenarios:**

#### A. Stdio Smoke Test
```go
func TestStdioBasic(t *testing.T) {
    cmd := exec.Command("../distribution/binaries/linux-amd64/agent-payment-router")

    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    cmd.Start()
    defer cmd.Process.Kill()

    // Send initialize
    stdin.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"))

    // Read response
    scanner := bufio.NewScanner(stdout)
    scanner.Scan()

    var resp map[string]interface{}
    json.Unmarshal(scanner.Bytes(), &resp)

    assert.Equal(t, "2.0", resp["jsonrpc"])
    assert.Equal(t, float64(1), resp["id"])
    assert.NotNil(t, resp["result"])
}
```

#### B. Tools List Test
```go
func TestToolsList(t *testing.T) {
    cmd := exec.Command("../distribution/binaries/linux-amd64/agent-payment-router")

    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    cmd.Start()
    defer cmd.Process.Kill()

    // Send tools/list
    stdin.Write([]byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"))

    scanner := bufio.NewScanner(stdout)
    scanner.Scan()

    var resp map[string]interface{}
    json.Unmarshal(scanner.Bytes(), &resp)

    result := resp["result"].(map[string]interface{})
    tools := result["tools"].([]interface{})

    assert.Greater(t, len(tools), 0, "Should have at least one tool")
}
```

#### C. HTTP API Test
```bash
#!/bin/bash
# tests/test-api.sh

API_KEY="test-key"
BUDGET_KEY="test-budget"

# Test fetch endpoint
curl -s -H "X-API-Key: ${API_KEY}" -H "X-Budget-Key: ${BUDGET_KEY}" \
  https://api.agentpmt.com/products/fetch | jq .

# Test purchase endpoint
curl -s -X POST -H "X-API-Key: ${API_KEY}" -H "X-Budget-Key: ${BUDGET_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"product_id":"test","parameters":{}}' \
  https://api.agentpmt.com/products/purchase | jq .
```

#### D. Integration Test (Claude Desktop)
```bash
#!/bin/bash
# tests/test-claude.sh

# Install to test environment
export TEST_INSTALL_PATH="$HOME/.claude-test/servers/agent-payment"
mkdir -p "$TEST_INSTALL_PATH"

# Copy binary
cp distribution/binaries/$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)/agent-payment-router \
   "$TEST_INSTALL_PATH/"

# Create config
cat > "$TEST_INSTALL_PATH/config.json" <<EOF
{
  "APIURL": "https://api.agentpmt.com",
  "APIKey": "${AGENTPMT_API_KEY}",
  "BudgetKey": "${AGENTPMT_BUDGET_KEY}"
}
EOF

# Test stdio directly
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | \
  "$TEST_INSTALL_PATH/agent-payment-router"
```

**Test Matrix:**
- ✅ Stdio JSON-RPC basic flow
- ✅ Tools list returns valid data
- ✅ Tool call executes successfully
- ✅ API endpoints respond correctly
- ✅ Binary works in Claude Desktop

**Commit Message:** `test(e2e): add stdio smoke and curl checks`

---

## Phase 11: Documentation

**Files to Create:**
- `remote-router/README.md` - Quick start guide
- `remote-router/ARCHITECTURE.md` - Technical architecture
- `remote-router/SECURITY.md` - Security model
- `remote-router/CONTRIBUTING.md` - Development guide

**README.md Outline:**
```markdown
# AgentPMT Remote-First MCP Router

Minimal, secure MCP server that routes stdio JSON-RPC to AgentPMT HTTPS API.

## Features
- 🔒 Zero privileged operations (HTTPS outbound only)
- 📦 Tiny binaries (<10MB, stripped)
- 🔑 Raw JSON schema preservation
- 🌊 Optional SSE streaming
- ✍️ Code-signed Windows binaries
- 🌍 Cross-platform (Windows, macOS, Linux)

## Quick Start
[Installation instructions]

## Configuration
[Config file and env vars]

## Security
[Security model and guarantees]

## Development
[Building, testing, contributing]
```

---

## Phase 12: Definition of Done (Checklist)

### Core Functionality
- [ ] Server speaks stdio JSON-RPC (newline-delimited)
- [ ] No OS-level shell execution
- [ ] No local listeners or privileged operations
- [ ] Raw JSON schemas passed through unmodified
- [ ] Tool calls POST to `/products/purchase` with params
- [ ] Output returned as MCP text content
- [ ] Optional SSE streaming works

### Build & Distribution
- [ ] Cross-platform binaries built (Windows, macOS, Linux)
- [ ] Binaries < 10MB each
- [ ] Version embedded via ldflags
- [ ] SHA256 hashes calculated
- [ ] Windows binary code-signed
- [ ] MSIX package created (optional)

### Installation
- [ ] Installers drop binary + config to correct locations
- [ ] `.mcpb` packages build successfully
- [ ] MCP manifest points to new binary
- [ ] Claude/Cursor detect server after restart

### CI/CD
- [ ] GitHub Actions builds all platforms
- [ ] Windows signing integrated (Azure Trusted Signing)
- [ ] Artifacts uploaded to releases
- [ ] Automated versioning works

### Testing
- [ ] Stdio smoke test passes
- [ ] Tools list test passes
- [ ] Tool invocation test passes
- [ ] API integration test passes
- [ ] Claude Desktop integration verified

### Documentation
- [ ] README with quick start
- [ ] Architecture documentation
- [ ] Security model documented
- [ ] Contributing guide created

---

## Security Guardrails

**What This Router DOES:**
- ✅ Reads JSON-RPC from stdin
- ✅ Makes HTTPS requests to AgentPMT API
- ✅ Returns JSON-RPC responses to stdout
- ✅ Logs to stderr (with secret redaction)

**What This Router DOES NOT DO:**
- ❌ Execute shell commands
- ❌ Access local file system (except config.json)
- ❌ Open network listeners
- ❌ Require elevated privileges
- ❌ Attempt to bypass Windows Defender/SmartScreen

**Legitimate Security Practices Only:**
- Code signing with valid certificates
- Timestamping for signature longevity
- MSIX packaging for Windows integration
- Submit to Microsoft for reputation review
- Build reputation organically (2-8 weeks)

---

## Dependencies

**Go Modules:**
```
github.com/tmaxmax/go-sse  // SSE client (optional streaming)
```

**Build Tools:**
- Go 1.23+
- Git
- Azure Trusted Signing CLI (Windows signing)
- MakeAppx.exe (MSIX packaging, optional)
- SignTool.exe (Windows signing)

**CI Requirements:**
- GitHub Actions (Ubuntu, Windows runners)
- Azure Trusted Signing account (or code signing certificate)

---

## Commit Plan (Git Messages)

1. `feat(config): add client constructor, UA, env overrides`
2. `feat(mcp): preserve raw JSON schemas in tools/list`
3. `feat(api): implement purchase call + optional SSE streaming`
4. `chore(build): strip symbols, embed version, cross-compile`
5. `feat(win): add signing and optional MSIX packaging scripts`
6. `chore(installer): point templates at new binaries; keep config side-by-side`
7. `ci(release): sign, package .mcpb/.msix, upload assets`
8. `test(e2e): add stdio smoke and curl checks`
9. `docs: add README, architecture, security docs`

---

## Timeline Estimate

| Phase | Estimated Time | Dependencies |
|-------|---------------|--------------|
| 0. Project Setup | 1 hour | None |
| 1. Config + Env Vars | 2 hours | Phase 0 |
| 2. HTTP Client | 2 hours | Phase 1 |
| 3. Tools List | 3 hours | Phase 2 |
| 4. Tool Invocation + SSE | 4 hours | Phase 3 |
| 5. Stdio JSON-RPC Loop | 3 hours | Phase 4 |
| 6. Build Scripts | 2 hours | Phase 5 |
| 7. Windows Signing | 4 hours | Phase 6, Certificates |
| 8. Installer Updates | 3 hours | Phase 6 |
| 9. CI Pipeline | 4 hours | Phase 7, 8 |
| 10. E2E Tests | 4 hours | Phase 9 |
| 11. Documentation | 3 hours | Phase 10 |
| 12. QA & Polish | 4 hours | All phases |

**Total: ~35-40 hours** (approximately 1 week of focused development)

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Azure Trusted Signing eligibility issues | Medium | High | Fallback to traditional code signing cert + 2-8 week reputation building |
| API schema changes break raw pass-through | Low | Medium | Version API responses, validate in tests |
| SSE library compatibility issues | Low | Medium | Keep synchronous path working, SSE optional |
| Windows SmartScreen warnings persist | High | Medium | Submit to Microsoft, build reputation, document expected behavior |
| Cross-platform binary issues | Low | High | Test on all platforms before release |
| MCP protocol changes | Low | High | Pin to MCP protocol version, monitor spec changes |

---

## Next Steps

1. **Review this plan** - Validate approach with stakeholders
2. **Set up Azure Trusted Signing** - Or acquire code signing certificate
3. **Create `remote-router` Go module** - Initialize project structure
4. **Begin Phase 1** - Config and environment variables
5. **Iterate through phases** - Following commit plan
6. **Test on all platforms** - Before release
7. **Submit to Microsoft** - For SmartScreen reputation
8. **Monitor usage** - Track errors and performance

---

## Questions & Decisions Needed

1. **Code Signing:** Do we have Azure Trusted Signing access? Or need to acquire EV certificate?
2. **MSIX Packaging:** Required or optional? (Adds complexity but improves Windows UX)
3. **Streaming:** Is SSE streaming a hard requirement or nice-to-have?
4. **Version Strategy:** Separate versioning from main MCP server? (`router-v1.0.0` vs `v1.0.0`)
5. **Migration Path:** How to transition existing users from old server to router?
6. **Testing Environment:** Do we have test API keys for automated testing?

---

## Success Criteria

**This implementation is successful when:**

1. ✅ Router binary is < 10MB per platform
2. ✅ No privileged operations required
3. ✅ Raw JSON schemas preserved exactly
4. ✅ Windows binary is code-signed
5. ✅ All platform binaries available in GitHub releases
6. ✅ Claude Desktop successfully loads and uses the router
7. ✅ Tool invocations complete successfully
8. ✅ Optional streaming works for supported endpoints
9. ✅ End-to-end tests pass on all platforms
10. ✅ Documentation is complete and accurate

---

**Plan Status:** DRAFT - Awaiting Approval
**Last Updated:** October 16, 2025
**Owner:** Development Team
**Reviewers:** [To be assigned]
