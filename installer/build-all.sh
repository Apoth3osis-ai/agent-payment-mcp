#!/bin/bash

# Agent Payment MCP Installer Build Script
# Builds for all platforms

set -e

cd "$(dirname "$0")"

echo "Building Agent Payment MCP Installers..."

# Create build directory
mkdir -p build

# Linux AMD64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o build/agent-payment-installer-linux-amd64 ./cmd/installer

# Windows AMD64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o build/agent-payment-installer-windows-amd64.exe ./cmd/installer

# macOS AMD64 (Intel)
echo "Building for macOS AMD64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o build/agent-payment-installer-macos-intel ./cmd/installer

# macOS ARM64 (M1/M2)
echo "Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o build/agent-payment-installer-macos-arm64 ./cmd/installer

echo ""
echo "✓ Build complete!"
echo ""
ls -lh build/
