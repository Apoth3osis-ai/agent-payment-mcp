#!/bin/bash
set -e

VERSION=${1:-dev}

echo "Building AgentPMT Router v${VERSION} for all platforms..."

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

echo ""
echo "Build complete! Binaries:"
ls -lh distribution/binaries/*/agent-payment-router* | awk '{print $9, "→", $5}'

echo ""
echo "Total sizes:"
du -sh distribution/binaries/*
