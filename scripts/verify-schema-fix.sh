#!/bin/bash
# Quick verification script to check if schema sanitization is working

echo "==========================================="
echo "Schema Sanitization Verification"
echo "==========================================="
echo ""

# Check if sanitizeJSONSchema function exists in server.go
if grep -q "func sanitizeJSONSchema" mcp-server/internal/mcp/server.go; then
    echo "✅ sanitizeJSONSchema function found in server.go"
else
    echo "❌ sanitizeJSONSchema function NOT found"
    exit 1
fi

# Check if function is being called in registerTool
if grep -q "sanitizedParams := sanitizeJSONSchema" mcp-server/internal/mcp/server.go; then
    echo "✅ sanitizeJSONSchema is called in registerTool"
else
    echo "❌ sanitizeJSONSchema is NOT called"
    exit 1
fi

# Check if tests exist
if [ -f "mcp-server/internal/mcp/schema_test.go" ]; then
    echo "✅ Unit tests found"
    
    # Run tests
    echo ""
    echo "Running unit tests..."
    cd mcp-server
    if go test -v ./internal/mcp -run TestSanitizeJSONSchema 2>&1 | grep -q "PASS"; then
        echo "✅ All unit tests pass"
    else
        echo "❌ Tests failed"
        exit 1
    fi
    cd ..
else
    echo "❌ Unit tests NOT found"
    exit 1
fi

# Check if binaries are built
echo ""
echo "Checking binaries..."
for binary in distribution/binaries/linux-amd64/agent-payment-server \
              distribution/binaries/darwin-amd64/agent-payment-server \
              distribution/binaries/darwin-arm64/agent-payment-server \
              distribution/binaries/windows-amd64/agent-payment-server.exe; do
    if [ -f "$binary" ]; then
        echo "✅ $binary exists"
    else
        echo "⚠️  $binary not found (run scripts/build-all.sh)"
    fi
done

echo ""
echo "==========================================="
echo "Verification Complete"
echo "==========================================="
echo ""
echo "Next steps:"
echo "1. Rebuild all binaries: bash scripts/build-all.sh"
echo "2. Reinstall MCP server using installer"
echo "3. Test with Claude Code CLI"
echo ""
