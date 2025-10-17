#!/bin/bash
# Quick stdio smoke test for the router

set -e

BINARY=${1:-./agent-payment-router}

if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found: $BINARY"
    echo "Usage: $0 [path-to-binary]"
    exit 1
fi

echo "Testing stdio interface of $BINARY..."

# Set test environment
export AGENTPMT_API_KEY="test-api-key"
export AGENTPMT_BUDGET_KEY="test-budget-key"
export AGENTPMT_API_URL="https://api.agentpmt.com"

# Test initialize
echo "Test 1: Initialize..."
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | $BINARY 2>/dev/null)

if echo "$RESPONSE" | jq -e '.result.protocolVersion' > /dev/null 2>&1; then
    echo "✓ Initialize successful"
    echo "  Protocol version: $(echo "$RESPONSE" | jq -r '.result.protocolVersion')"
    echo "  Server: $(echo "$RESPONSE" | jq -r '.result.serverInfo.name') v$(echo "$RESPONSE" | jq -r '.result.serverInfo.version')"
else
    echo "✗ Initialize failed"
    echo "  Response: $RESPONSE"
    exit 1
fi

# Test tools/list (will fail if API is not reachable, but that's ok for smoke test)
echo ""
echo "Test 2: Tools List (may fail if API unreachable)..."
RESPONSE=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | $BINARY 2>/dev/null || true)

if echo "$RESPONSE" | jq -e '.result.tools' > /dev/null 2>&1; then
    TOOL_COUNT=$(echo "$RESPONSE" | jq '.result.tools | length')
    echo "✓ Tools list successful ($TOOL_COUNT tools)"
elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo "⚠ Tools list failed (expected if API not configured)"
    echo "  Error: $(echo "$RESPONSE" | jq -r '.error.message')"
else
    echo "⚠ Tools list returned unexpected response"
fi

# Test resources/list
echo ""
echo "Test 3: Resources List..."
RESPONSE=$(echo '{"jsonrpc":"2.0","id":3,"method":"resources/list","params":{}}' | $BINARY 2>/dev/null)

if echo "$RESPONSE" | jq -e '.result.resources' > /dev/null 2>&1; then
    echo "✓ Resources list successful (empty as expected)"
else
    echo "✗ Resources list failed"
    exit 1
fi

echo ""
echo "✓ All stdio tests passed!"
