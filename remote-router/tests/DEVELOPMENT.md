# Development Guide

## Project Structure

```
remote-router/
├── cmd/agent-payment-router/    # Main entry point
├── internal/
│   ├── config/                   # Configuration loading
│   ├── api/                      # HTTP client + SSE streaming
│   └── mcp/                      # MCP JSON-RPC server
├── scripts/                      # Build and deployment scripts
├── distribution/
│   ├── binaries/                 # Built binaries (output)
│   └── templates/                # Installer templates
├── tests/                        # E2E tests
└── windows/                      # MSIX packaging manifests
```

## Quick Start

### Prerequisites

- Go 1.23+
- Git

### Build

```bash
# Build for current platform
go build -o agent-payment-router ./cmd/agent-payment-router

# Build for all platforms
./scripts/build-all.sh 1.0.0
```

### Test

```bash
# Unit tests
go test ./...

# E2E tests
go test ./tests -v

# Stdio smoke test
./scripts/test-stdio.sh ./agent-payment-router
```

### Run Locally

```bash
# Set environment variables
export AGENTPMT_API_KEY="your-api-key"
export AGENTPMT_BUDGET_KEY="your-budget-key"

# Run
./agent-payment-router
```

Then send JSON-RPC requests via stdin:
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./agent-payment-router
```

## Architecture

The router is a minimal proxy between MCP clients (Claude/Cursor) and the AgentPMT API:

1. **Config Layer** (`internal/config`) - Loads configuration from file or environment
2. **API Layer** (`internal/api`) - HTTP client with SSE streaming support
3. **MCP Layer** (`internal/mcp`) - stdio JSON-RPC server implementation
4. **Main** (`cmd/agent-payment-router`) - Wires everything together

### Key Design Principles

1. **Zero privileged operations** - Only outbound HTTPS
2. **Raw schema preservation** - Use `json.RawMessage` to avoid re-marshaling
3. **Minimal dependencies** - Only `tmaxmax/go-sse` for streaming
4. **Deterministic builds** - Static compilation with stripped symbols

## Making Changes

### Adding a New MCP Method

1. Add method handler in `internal/mcp/server.go`
2. Add route in `HandleStdioTransport` switch statement
3. Add test in `internal/mcp/server_test.go`
4. Add E2E test in `tests/e2e_test.go`

### Adding a New API Endpoint

1. Add types in `internal/api/client.go`
2. Add method to `Client` struct
3. Add test in `internal/api/client_test.go`
4. Update `ClientInterface` if needed

### Modifying Configuration

1. Update `Config` struct in `internal/config/config.go`
2. Add environment variable handling in `Load()`
3. Add test in `internal/config/config_test.go`
4. Update README.md configuration section

## Testing

### Unit Tests

```bash
# All packages
go test ./...

# Specific package
go test ./internal/config -v
go test ./internal/api -v
go test ./internal/mcp -v
```

### E2E Tests

```bash
# Build first
go build -o agent-payment-router ./cmd/agent-payment-router

# Run E2E tests
go test ./tests -v
```

### Benchmarks

```bash
go test ./tests -bench=. -benchmem
```

## Release Process

### Manual Release

1. Tag the release:
   ```bash
   git tag router-v1.0.0
   git push origin router-v1.0.0
   ```

2. GitHub Actions will:
   - Build all platform binaries
   - Sign Windows binary (if configured)
   - Create GitHub release
   - Upload binaries and installers

### Testing Before Release

1. Build all platforms:
   ```bash
   ./scripts/build-all.sh 1.0.0-rc1
   ```

2. Test each binary:
   ```bash
   ./scripts/test-stdio.sh distribution/binaries/linux-amd64/agent-payment-router
   ./scripts/test-stdio.sh distribution/binaries/darwin-amd64/agent-payment-router
   # etc
   ```

3. Test installers (requires target OS):
   ```bash
   ./distribution/templates/install-linux.sh --client claude --api-key test --budget-key test
   ```

## Troubleshooting

### Build Failures

**CGO errors:**
- Ensure `CGO_ENABLED=0` for static builds

**Missing dependencies:**
```bash
go mod tidy
```

### Test Failures

**E2E tests can't find binary:**
```bash
# Build binary first
go build -o agent-payment-router ./cmd/agent-payment-router
```

**API tests timing out:**
- Check network connectivity
- Increase timeout in test

### Runtime Issues

**Config not found:**
- Place `config.json` next to binary
- OR set environment variables

**API authentication errors:**
- Check API key validity
- Ensure budget key is correct
- Verify API URL

## Code Style

- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions under 50 lines where possible
- Add tests for new functionality
- Update documentation with changes

## Contributing

See parent repository CONTRIBUTING.md for guidelines.
