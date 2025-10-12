#!/bin/bash
# Quick test to verify MCP server starts and logs schema sanitization

echo "=========================================="
echo "Testing MCP Server with Schema Fix"
echo "=========================================="
echo ""

# Check if API keys are configured
if [ -z "$AGENT_PAYMENT_API_KEY" ] || [ -z "$AGENT_PAYMENT_BUDGET_KEY" ]; then
    echo "⚠️  API keys not set in environment"
    echo ""
    echo "Please set:"
    echo "  export AGENT_PAYMENT_API_KEY='your-api-key'"
    echo "  export AGENT_PAYMENT_BUDGET_KEY='your-budget-key'"
    echo ""
    echo "Or the server will start but fail to fetch tools."
    echo ""
fi

# Test the server binary
echo "Starting MCP server (will auto-exit after 5 seconds)..."
echo ""
echo "Look for this message in the output:"
echo "  'Sanitized schema for tool <name> (fixed required fields)'"
echo ""
echo "=========================================="

# Start server and kill after 5 seconds
timeout 5 /home/richard/.agent-payment/agent-payment-server 2>&1 | head -30

echo ""
echo "=========================================="
echo "Test complete!"
echo ""
echo "If you saw 'Sanitized schema' messages above,"
echo "the fix is working! ✅"
echo ""
echo "Now try Claude CLI - it should work without errors."
echo "=========================================="
