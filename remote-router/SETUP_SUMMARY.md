# Remote-First MCP Router - Setup Complete ✅

**Date:** October 16, 2025
**Status:** Project scaffolding complete, ready for implementation

## What Was Done

### 1. Research Completed ✅

Three specialized research agents investigated critical technical areas:

#### A. MCP stdio JSON-RPC Protocol
- **Finding:** MCP uses newline-delimited JSON-RPC 2.0 over stdio
- **Key Methods:** `initialize`, `tools/list`, `tools/call`, `notifications/initialized`
- **Schema Preservation:** Use `json.RawMessage` to prevent re-marshaling
- **Transport:** Read from stdin, write to stdout, log to stderr only

#### B. Windows Code Signing (2024/2025)
- **Azure Trusted Signing:** $9.99/month, instant SmartScreen reputation (US/Canada only)
- **Traditional Certs:** No longer provide instant reputation bypass
- **SmartScreen Reality:** 2-8 weeks reputation building required
- **Best Practice:** Sign with SHA256, always timestamp, submit to Microsoft

#### C. Server-Sent Events (SSE) in Go
- **Recommended Library:** `tmaxmax/go-sse` (modern, Go 1.23 iterators)
- **Alternative:** `r3labs/sse` (auto-reconnection, persistent connections)
- **Critical Setting:** `DisableCompression: true` on HTTP transport
- **Use Case:** Optional streaming for real-time API responses

### 2. Project Structure Created ✅

```
remote-router/
├── cmd/
│   └── agent-payment-router/       # Main entry point (empty, ready for code)
├── internal/
│   ├── config/                      # Configuration loading (to be implemented)
│   ├── api/                         # HTTP client + SSE (to be implemented)
│   └── mcp/                         # MCP JSON-RPC handlers (to be implemented)
├── scripts/                         # Build & signing scripts (to be created)
├── windows/                         # MSIX manifests (to be created)
├── distribution/
│   ├── binaries/                    # Built binaries (output folder)
│   ├── packages/                    # MSIX/MCPB packages (output folder)
│   └── templates/                   # Installer templates (to be created)
├── tests/                           # E2E tests (to be created)
├── go.mod                           # ✅ Go module initialized
├── .gitignore                       # ✅ Ignoring binaries, secrets, build artifacts
├── README.md                        # ✅ User-facing documentation
├── IMPLEMENTATION_PLAN.md           # ✅ Detailed 12-phase implementation plan
└── SETUP_SUMMARY.md                 # ✅ This file
```

### 3. Documentation Created ✅

#### IMPLEMENTATION_PLAN.md (Comprehensive)
- **12 Implementation Phases** with detailed steps
- **Code examples** for each component
- **Acceptance criteria** for each phase
- **Timeline estimate:** 35-40 hours (1 week focused development)
- **Risk assessment** and mitigation strategies
- **Security guardrails** and best practices
- **Commit plan** with semantic messages

#### README.md (User-Facing)
- Architecture overview with diagrams
- Quick start guide
- Configuration methods (config file + env vars)
- Security model (what router does/doesn't do)
- Troubleshooting guide
- Development instructions

### 4. Dependencies Identified ✅

**Go Modules:**
- `github.com/tmaxmax/go-sse` - SSE streaming client

**Build Tools:**
- Go 1.23+
- SignTool.exe (Windows signing)
- MakeAppx.exe (MSIX packaging, optional)

**CI Requirements:**
- GitHub Actions (ubuntu-latest, windows-latest)
- Azure Trusted Signing account or traditional code signing certificate

## Implementation Roadmap

### Phase 1: Config and Environment Variables (2 hours)
**File:** `internal/config/config.go`
- Load from `config.json`
- Support env var overrides (`AGENTPMT_API_URL`, `AGENTPMT_API_KEY`, `AGENTPMT_BUDGET_KEY`)
- Default to `https://api.agentpmt.com`

### Phase 2: HTTP Client Hardening (2 hours)
**File:** `internal/api/client.go`
- Constructor with timeouts
- User-Agent header: `AgentPMT-MCP/1.0`
- Consistent header injection (API keys)
- HTTPS only, no local listeners

### Phase 3: Tools List Pass-Through (3 hours)
**Files:** `internal/api/client.go`, `internal/mcp/server.go`
- Fetch from `GET /products/fetch`
- Preserve raw JSON schemas (`json.RawMessage`)
- Map 1:1 to MCP `tools/list` response

### Phase 4: Tool Invocation + SSE Streaming (4 hours)
**Files:** `internal/api/client.go`, `internal/api/sse.go`
- Synchronous: `POST /products/purchase`
- Optional streaming: `POST /products/purchase?stream=true` with SSE
- Parse responses to MCP format

### Phase 5: Stdio JSON-RPC Loop (3 hours)
**Files:** `internal/mcp/rpc.go`, `internal/mcp/server.go`
- Newline-delimited JSON-RPC 2.0
- Handle: `initialize`, `tools/list`, `tools/call`
- Log to stderr, secrets redacted

### Phase 6: Build Scripts (2 hours)
**File:** `scripts/build-all.sh`
- Cross-compile: Windows, macOS (Intel/ARM), Linux (x64/ARM)
- Strip symbols: `-ldflags "-s -w"`
- Embed version: `-X main.Version=${VERSION}`
- Target: <10MB per binary

### Phase 7: Windows Signing (4 hours)
**Files:** `scripts/sign-windows.ps1`, `scripts/package-msix.ps1`, `windows/AppxManifest.xml`
- Code sign with SHA256 + timestamping
- Optional MSIX packaging
- Azure Trusted Signing integration

### Phase 8: Installer Updates (3 hours)
**Files:** `distribution/templates/install-*.{sh,ps1}`, `mcpb-manifest.json`
- Update to new binary names
- Place config next to binary
- Maintain `.mcpb` packaging

### Phase 9: CI Pipeline (4 hours)
**File:** `.github/workflows/release-router.yml`
- Build all platforms
- Sign Windows binaries
- Package `.mcpb` bundles
- Upload to GitHub Releases

### Phase 10: E2E Tests (4 hours)
**File:** `tests/e2e_test.go`
- Stdio smoke test
- Tools list test
- Tool invocation test
- API integration test

### Phase 11: Documentation (3 hours)
- Architecture docs
- Security model
- Contributing guide
- API documentation

### Phase 12: QA & Polish (4 hours)
- Cross-platform testing
- Performance profiling
- Security audit
- User acceptance testing

**Total Estimated Time: 35-40 hours**

## Key Architectural Decisions

### 1. Remote-First Design
- **All logic in remote API** - Router is just a proxy
- **No local code execution** - Maximum security
- **Deterministic behavior** - Same API = same results

### 2. Raw Schema Preservation
- **Use `json.RawMessage`** - No SDK re-marshaling
- **Pass-through unchanged** - API schemas reach client intact
- **Dynamic tool discovery** - Tools defined by API, not code

### 3. Optional Streaming
- **Default: synchronous** - Simple request/response
- **Optional: SSE streaming** - For real-time updates
- **Graceful fallback** - Works without streaming support

### 4. Minimal Dependencies
- **Only one external lib** - `tmaxmax/go-sse` for streaming
- **Standard library first** - HTTP, JSON, bufio
- **No MCP SDK** - Custom stdio loop for full control

### 5. Cross-Platform First
- **5 platform targets** - Windows, macOS (Intel/ARM), Linux (x64/ARM)
- **Static binaries** - `CGO_ENABLED=0` for portability
- **Platform-specific signing** - Windows gets special treatment

## Security Guarantees

### ✅ What This Router Does
1. Reads JSON-RPC from stdin
2. Makes HTTPS requests to `api.agentpmt.com`
3. Returns JSON-RPC to stdout
4. Logs to stderr (secrets redacted)

### ❌ What This Router Does NOT Do
1. Execute shell commands
2. Access file system (except config.json read)
3. Open network listeners
4. Require elevated privileges
5. Bypass security controls

### 🔒 Legitimate Security Practices
1. Code signing with valid certificates
2. Timestamping for signature longevity
3. MSIX packaging for Windows integration
4. Microsoft submission for reputation
5. Organic reputation building (2-8 weeks)

## Questions & Decisions Needed

Before starting implementation, we need to decide:

### 1. Code Signing Approach
**Question:** Do we have Azure Trusted Signing access?
- **If Yes:** Use Azure Trusted Signing ($9.99/month, instant reputation)
- **If No:** Acquire traditional code signing cert + plan for 2-8 week reputation building

### 2. MSIX Packaging
**Question:** Is MSIX packaging required or optional?
- **Required:** Adds complexity but better Windows integration
- **Optional:** Skip for v1, add later if needed

### 3. Streaming Priority
**Question:** Is SSE streaming a hard requirement or nice-to-have?
- **Hard Requirement:** Implement in Phase 4
- **Nice-to-Have:** Defer to v2, keep synchronous only for v1

### 4. Version Strategy
**Question:** Separate versioning from main MCP server?
- **Option A:** `router-v1.0.0` (separate version namespace)
- **Option B:** `v2.0.0` (major version bump)
- **Option C:** `v1.5.0` (minor version bump)

### 5. Migration Path
**Question:** How to transition existing users?
- **Big Bang:** Replace old server entirely
- **Parallel:** Run both, deprecate old over time
- **Opt-In:** New users get router, existing keep old

### 6. Testing Environment
**Question:** Do we have test API keys?
- **Yes:** Can run automated E2E tests
- **No:** Need to create test environment or use mocks

## Next Steps

### Immediate (Today)
1. ✅ **Review this summary** - Validate approach
2. ⏳ **Answer decision questions** - See above
3. ⏳ **Set up code signing** - Azure Trusted Signing or acquire cert

### Short Term (This Week)
4. ⏳ **Initialize Go module** - Run `go mod tidy` in remote-router/
5. ⏳ **Implement Phase 1** - Config with env overrides
6. ⏳ **Implement Phase 2** - HTTP client hardening
7. ⏳ **Implement Phase 3** - Tools list pass-through

### Medium Term (Next Week)
8. ⏳ **Implement Phases 4-7** - Tool invocation, stdio loop, build, signing
9. ⏳ **Implement Phases 8-10** - Installers, CI, tests
10. ⏳ **Implement Phases 11-12** - Documentation, QA

### Release (Week 3)
11. ⏳ **Test on all platforms** - Windows, macOS, Linux
12. ⏳ **Submit to Microsoft** - For SmartScreen reputation
13. ⏳ **Create GitHub release** - Tagged version with all binaries
14. ⏳ **Update documentation** - Installation guides, migration docs

## Success Metrics

This implementation succeeds when:

1. ✅ Router binary < 10MB per platform
2. ✅ Zero privileged operations required
3. ✅ Raw JSON schemas preserved exactly
4. ✅ Windows binary code-signed
5. ✅ All platforms available in releases
6. ✅ Claude Desktop loads router successfully
7. ✅ Tool invocations complete successfully
8. ✅ Optional streaming works
9. ✅ E2E tests pass on all platforms
10. ✅ Documentation complete and accurate

## Resources Created

All files are in `/home/richard/Documents/agentpmt/local_mcp/remote-router/`:

1. **IMPLEMENTATION_PLAN.md** - 400+ line detailed implementation guide
2. **README.md** - User-facing documentation
3. **go.mod** - Go module definition
4. **.gitignore** - Git ignore rules
5. **SETUP_SUMMARY.md** - This summary
6. **Directory structure** - All folders created

## Research Artifacts

Three research agents provided comprehensive findings on:

1. **MCP stdio JSON-RPC** - Protocol specs, best practices, code examples
2. **Windows Code Signing** - Modern approaches, Azure Trusted Signing, SmartScreen reality
3. **SSE in Go** - Library recommendations, implementation patterns, best practices

All research findings are integrated into the implementation plan.

---

**Status:** ✅ Setup Complete - Ready for Implementation
**Next Action:** Review decision questions and begin Phase 1 implementation
**Estimated Completion:** 1 week of focused development (35-40 hours)
