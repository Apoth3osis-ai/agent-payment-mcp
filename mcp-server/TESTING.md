# Testing Guide

Complete testing strategy for the Agent Payment MCP Server.

## Testing Levels

### 1. Compilation Test

Verifies that the code compiles with the official Go MCP SDK.

```bash
cd /home/richard/Documents/agentpmt/local_mcp/mcp-server

# Download dependencies
go mod tidy

# Verify no errors
go build ./cmd/agent-payment-server

# Check binary was created
ls -lh agent-payment-server
```

**Expected Output**:
```
go: downloading github.com/modelcontextprotocol/go-sdk v0.1.0
go: downloading github.com/modelcontextprotocol/go-sdk/jsonschema v0.1.0
...
-rwxr-xr-x 1 user user 7.2M Oct 9 16:00 agent-payment-server
```

**Success Criteria**:
- ✅ No compilation errors
- ✅ Binary created
- ✅ Binary size ~6-8MB

---

### 2. Unit Tests

Test individual components in isolation.

Create `internal/api/client_test.go`:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchTools(t *testing.T) {
	// Mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("Missing or incorrect API key")
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"tools": [
				{
					"type": "function",
					"function": {
						"name": "test_tool",
						"description": "A test tool",
						"parameters": {
							"type": "object",
							"properties": {
								"param1": {"type": "string"}
							}
						}
					}
				}
			]
		}`))
	}))
	defer server.Close()

	// Create client with mock URL
	client := NewClient("test-key", "test-budget")
	client.baseURL = server.URL

	// Test fetch
	resp, err := client.FetchTools(1, 10)
	if err != nil {
		t.Fatalf("FetchTools failed: %v", err)
	}

	// Verify response
	if !resp.Success {
		t.Error("Expected success=true")
	}
	if len(resp.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(resp.Tools))
	}
	if resp.Tools[0].Function.Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", resp.Tools[0].Function.Name)
	}
}

func TestExecuteTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"result": "Tool executed successfully",
			"cost": 0.01,
			"balance": 9.99
		}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "test-budget")
	client.baseURL = server.URL

	resp, err := client.ExecuteTool("test_tool", map[string]interface{}{"param1": "value1"})
	if err != nil {
		t.Fatalf("ExecuteTool failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}
	if resp.Cost != 0.01 {
		t.Errorf("Expected cost=0.01, got %f", resp.Cost)
	}
}
```

Run tests:
```bash
go test ./internal/api/...
```

**Expected Output**:
```
ok      github.com/agentpmt/agent-payment-mcp-server/internal/api       0.123s
```

---

### 3. Integration Test - Server Startup

Test that the server starts correctly and fetches tools.

```bash
# Set test credentials (use actual API keys)
export AGENT_PAYMENT_API_KEY=your-test-api-key
export AGENT_PAYMENT_BUDGET_KEY=your-test-budget-key

# Run server (will run until interrupted)
./agent-payment-server
```

**Expected Output**:
```
2025/10/09 16:00:00 Fetching tools from Agent Payment API...
2025/10/09 16:00:01 Fetched 50 tools from API
2025/10/09 16:00:01 Successfully registered 50 tools
2025/10/09 16:00:01 Starting MCP server on stdio transport...
```

**Success Criteria**:
- ✅ No error messages
- ✅ Tools fetched count > 0
- ✅ Registered tools count matches fetched count
- ✅ Server starts on stdio transport

Press `Ctrl+C` to stop:
```
^C2025/10/09 16:00:30 Received signal interrupt, shutting down gracefully...
2025/10/09 16:00:30 Server shutdown complete
```

**Failure Scenarios**:

If missing credentials:
```
Error: AGENT_PAYMENT_API_KEY and AGENT_PAYMENT_BUDGET_KEY environment variables must be set

Usage:
  export AGENT_PAYMENT_API_KEY=your-api-key
  export AGENT_PAYMENT_BUDGET_KEY=your-budget-key
  agent-payment-server
```

If API is unreachable:
```
2025/10/09 16:00:00 Fetching tools from Agent Payment API...
2025/10/09 16:00:30 Failed to create server: failed to fetch tools: Post "https://api.agentpmt.com/products/fetch": dial tcp: lookup api.agentpmt.com: no such host
```

---

### 4. MCP Inspector Test

Test the server using the official MCP Inspector tool.

**Install Inspector** (one-time):
```bash
npm install -g @modelcontextprotocol/inspector
```

**Run Inspector**:
```bash
export AGENT_PAYMENT_API_KEY=your-api-key
export AGENT_PAYMENT_BUDGET_KEY=your-budget-key

npx @modelcontextprotocol/inspector /home/richard/Documents/agentpmt/local_mcp/mcp-server/agent-payment-server
```

**Expected Output**:
```
Model Context Protocol Inspector
Server: /home/richard/Documents/agentpmt/local_mcp/mcp-server/agent-payment-server
Opening inspector at http://localhost:3000
```

**In Browser** (http://localhost:3000):

1. **Initialize Connection**
   - Click "Initialize"
   - Should see "Connected" status
   - Should see server info: name="agent-payment", version="1.0.0"

2. **List Tools**
   - Click "List Tools"
   - Should see all tools from API
   - Each tool should have:
     - Name (e.g., "weather_check")
     - Description
     - Input schema with parameters

3. **Call a Tool**
   - Select a tool (e.g., "weather_check")
   - Fill in required parameters (e.g., city: "New York")
   - Click "Call Tool"
   - Should see result with:
     - Result text
     - Cost (if applicable)
     - Balance (if applicable)

4. **Verify Error Handling**
   - Call a tool with invalid parameters
   - Should see error response with `isError: true`
   - Should stay connected (not disconnect)

**Success Criteria**:
- ✅ Connection establishes successfully
- ✅ All tools appear in list
- ✅ Tool schemas are correct
- ✅ Tool calls execute successfully
- ✅ Results are formatted correctly
- ✅ Errors are handled gracefully

---

### 5. Claude Desktop Integration Test

Test the server in actual Claude Desktop environment.

**Step 1: Configure Claude Desktop**

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or equivalent:

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/richard/Documents/agentpmt/local_mcp/mcp-server/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key"
      }
    }
  }
}
```

**Step 2: Restart Claude Desktop**

Completely quit and restart Claude Desktop.

**Step 3: Verify Connection**

Check Claude Desktop logs:
```bash
# macOS
tail -f ~/Library/Logs/Claude/mcp*.log

# Linux
tail -f ~/.config/Claude/logs/mcp*.log
```

Look for:
```
[agent-payment] Fetching tools from Agent Payment API...
[agent-payment] Fetched 50 tools from API
[agent-payment] Successfully registered 50 tools
[agent-payment] Starting MCP server on stdio transport...
```

**Step 4: Test in Claude**

Open Claude Desktop and test the following scenarios:

**Test 1: Tool Discovery**
```
User: What tools do I have available?

Expected: Claude lists all Agent Payment tools with descriptions
```

**Test 2: Tool Execution**
```
User: [Use one of your available tools with valid parameters]

Expected: Claude calls the tool, receives result, and responds appropriately
```

**Test 3: Error Handling**
```
User: [Use a tool with invalid parameters]

Expected: Claude receives error message and explains the issue
```

**Test 4: Browse Tools UI**
```
1. Click the tools icon in Claude Desktop
2. Browse available tools
3. Verify all Agent Payment tools appear
4. Check that descriptions and parameters are visible
```

**Success Criteria**:
- ✅ Server connects without errors
- ✅ All tools appear in Claude
- ✅ Tool descriptions are accurate
- ✅ Tools can be called successfully
- ✅ Results are displayed correctly
- ✅ Errors are handled gracefully
- ✅ Server stays connected (doesn't disconnect on errors)

---

### 6. Cursor Integration Test

Similar to Claude Desktop, but with Cursor IDE.

**Configure Cursor**:

Edit Cursor's MCP settings (Settings → MCP):

```json
{
  "mcpServers": {
    "agent-payment": {
      "command": "/home/richard/Documents/agentpmt/local_mcp/mcp-server/agent-payment-server",
      "env": {
        "AGENT_PAYMENT_API_KEY": "your-api-key",
        "AGENT_PAYMENT_BUDGET_KEY": "your-budget-key"
      }
    }
  }
}
```

**Test in Cursor**:
```
1. Open Cursor AI chat
2. Ask "What tools are available?"
3. Verify Agent Payment tools appear
4. Test calling a tool
5. Verify results appear in chat
```

---

### 7. Load Testing

Test server performance under load.

Create `test/load_test.go`:

```go
package test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/agentpmt/agent-payment-mcp-server/internal/mcp"
)

func BenchmarkToolExecution(b *testing.B) {
	// Setup server
	server, err := mcp.NewServer(mcp.Config{
		APIKey:    "test-key",
		BudgetKey: "test-budget",
	})
	if err != nil {
		b.Fatal(err)
	}

	// Benchmark tool calls
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate tool call
			// (In real test, would call via MCP protocol)
		}
	})
}

func TestConcurrentToolCalls(t *testing.T) {
	// Test that server can handle concurrent calls
	server, err := mcp.NewServer(mcp.Config{
		APIKey:    "test-key",
		BudgetKey: "test-budget",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make 100 concurrent tool calls
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Call tool (would use actual MCP client in real test)
			// errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent call failed: %v", err)
		}
	}
}
```

Run load tests:
```bash
go test -bench=. -benchtime=10s ./test/...
```

**Expected Performance**:
- Tool registration: < 10ms per tool
- Tool execution: ~API latency + 5ms
- Memory: < 100MB under load
- Concurrent calls: 100+ simultaneous

---

### 8. End-to-End Test

Complete workflow test simulating real usage.

**Test Script**:
```bash
#!/bin/bash
set -e

echo "=== Agent Payment MCP Server E2E Test ==="

# 1. Build
echo "Building server..."
cd /home/richard/Documents/agentpmt/local_mcp/mcp-server
go build -o agent-payment-server ./cmd/agent-payment-server

# 2. Start server in background
echo "Starting server..."
export AGENT_PAYMENT_API_KEY=your-api-key
export AGENT_PAYMENT_BUDGET_KEY=your-budget-key
./agent-payment-server &
SERVER_PID=$!

# Give server time to start
sleep 2

# 3. Send test request via MCP protocol
echo "Sending test request..."
# (Would use actual MCP client library here)

# 4. Verify response
echo "Verifying response..."
# (Check response format)

# 5. Cleanup
echo "Cleaning up..."
kill $SERVER_PID

echo "=== E2E Test Passed ==="
```

---

## Debugging

### Enable Debug Logging

Modify `internal/mcp/server.go`:

```go
import "log"

// Add at start of createToolHandler:
log.Printf("DEBUG: Tool %s called with params: %+v", toolName, params)

// Add after API call:
log.Printf("DEBUG: API response: %+v", result)
```

Rebuild and run to see detailed logs.

### Common Issues

**Issue**: "No tools registered"
```
Solution: Check API response format matches expected structure.
Debug: Add log.Printf("DEBUG: API response: %+v", toolsResp) in NewServer()
```

**Issue**: "Tool execution failed"
```
Solution: Check API credentials and network connectivity.
Debug: Add logging to api.Client.ExecuteTool()
```

**Issue**: "Invalid JSON schema"
```
Solution: API schema format may not match expected structure.
Debug: Print schema before conversion in registerTool()
```

---

## Continuous Integration

Example GitHub Actions workflow:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v ./...

      - name: Build
        run: go build -v ./cmd/agent-payment-server

      - name: Upload binary
        uses: actions/upload-artifact@v3
        with:
          name: agent-payment-server
          path: agent-payment-server
```

---

## Test Checklist

Before releasing:

- [ ] Unit tests pass (`go test ./...`)
- [ ] Code compiles without errors
- [ ] Server starts successfully
- [ ] Tools are fetched from API
- [ ] Tools are registered correctly
- [ ] MCP Inspector shows all tools
- [ ] Tool calls execute successfully
- [ ] Errors are handled gracefully
- [ ] Claude Desktop integration works
- [ ] Cursor integration works
- [ ] Load test passes
- [ ] Binary size is acceptable (< 10MB)
- [ ] Memory usage is reasonable (< 50MB)
- [ ] No sensitive data in logs
- [ ] Documentation is up to date

---

## Performance Benchmarks

Expected performance metrics:

| Metric | Target | Measured |
|--------|--------|----------|
| Startup time | < 3s | TBD |
| Tool registration | < 10ms/tool | TBD |
| Tool execution | < API + 10ms | TBD |
| Memory (idle) | < 20MB | TBD |
| Memory (100 calls) | < 50MB | TBD |
| Binary size | < 10MB | TBD |
| Concurrent calls | 100+ | TBD |

Fill in "Measured" column after testing.
