#!/bin/bash
set -e

echo "================================================"
echo "Building Agent Payment MCP Server"
echo "================================================"

cd mcp-server

# Clean previous builds
rm -rf ../distribution/binaries
mkdir -p ../distribution/binaries

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o ../distribution/binaries/windows-amd64/agent-payment-server.exe \
  ./cmd/agent-payment-server

# Build for macOS Intel
echo "Building for macOS Intel..."
GOOS=darwin GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o ../distribution/binaries/darwin-amd64/agent-payment-server \
  ./cmd/agent-payment-server

# Build for macOS Apple Silicon
echo "Building for macOS Apple Silicon..."
GOOS=darwin GOARCH=arm64 go build \
  -ldflags="-s -w" \
  -o ../distribution/binaries/darwin-arm64/agent-payment-server \
  ./cmd/agent-payment-server

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o ../distribution/binaries/linux-amd64/agent-payment-server \
  ./cmd/agent-payment-server

echo ""
echo "================================================"
echo "Build complete!"
echo "================================================"
echo ""
echo "Binaries:"
ls -lh ../distribution/binaries/*/agent-payment-server*
echo ""
