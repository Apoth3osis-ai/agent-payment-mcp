# AgentPMT MCP Router Installer for Windows
# Installs the router binary and configures MCP for Claude Desktop or Cursor

param(
    [string]$Client = "auto",  # "claude", "cursor", or "auto"
    [string]$ApiKey = "",
    [string]$BudgetKey = ""
)

$ErrorActionPreference = "Stop"

Write-Host "AgentPMT MCP Router Installer for Windows" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Detect client if auto
if ($Client -eq "auto") {
    $claudePath = "$env:APPDATA\Claude"
    $cursorPath = "$env:USERPROFILE\.cursor"

    if (Test-Path $claudePath) {
        $Client = "claude"
        Write-Host "✓ Detected Claude Desktop" -ForegroundColor Green
    } elseif (Test-Path $cursorPath) {
        $Client = "cursor"
        Write-Host "✓ Detected Cursor" -ForegroundColor Green
    } else {
        Write-Host "⚠ Could not auto-detect client. Please specify -Client claude or -Client cursor" -ForegroundColor Yellow
        exit 1
    }
}

# Set installation paths
if ($Client -eq "claude") {
    $installPath = "$env:APPDATA\Claude\servers\agent-payment"
    $configPath = "$env:APPDATA\Claude\claude_desktop_config.json"
} elseif ($Client -eq "cursor") {
    $installPath = "$env:USERPROFILE\.cursor\mcp\servers\agent-payment"
    $configPath = "$env:USERPROFILE\.cursor\mcp_config.json"
} else {
    Write-Error "Invalid client: $Client. Must be 'claude' or 'cursor'"
    exit 1
}

Write-Host "Installation path: $installPath" -ForegroundColor Gray
Write-Host ""

# Create directory
New-Item -ItemType Directory -Path $installPath -Force | Out-Null

# Download binary (or use local if available)
$binaryUrl = "https://github.com/Apoth3osis-ai/agent-payment-mcp/releases/latest/download/agent-payment-router-windows-amd64.exe"
$binaryPath = "$installPath\agent-payment-router.exe"

Write-Host "Downloading router binary..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $binaryUrl -OutFile $binaryPath -UseBasicParsing
    Write-Host "✓ Downloaded binary" -ForegroundColor Green
} catch {
    Write-Error "Failed to download binary: $_"
    exit 1
}

# Create config.json (optional, env vars preferred)
$configJson = @{
    APIURL = "https://api.agentpmt.com"
    APIKey = $ApiKey
    BudgetKey = $BudgetKey
}

if ($ApiKey -and $BudgetKey) {
    $configJson | ConvertTo-Json | Set-Content "$installPath\config.json"
    Write-Host "✓ Created config.json with provided keys" -ForegroundColor Green
} else {
    Write-Host "⚠ No API keys provided - you'll need to set environment variables or create config.json manually" -ForegroundColor Yellow
}

# Update MCP config
Write-Host ""
Write-Host "Updating MCP configuration..." -ForegroundColor Cyan

$mcpConfig = @{
    mcpServers = @{
        "agent-payment" = @{
            command = $binaryPath
            args = @()
            env = @{}
        }
    }
}

# Add env vars if keys provided
if ($ApiKey) {
    $mcpConfig.mcpServers."agent-payment".env.AGENTPMT_API_KEY = $ApiKey
}
if ($BudgetKey) {
    $mcpConfig.mcpServers."agent-payment".env.AGENTPMT_BUDGET_KEY = $BudgetKey
}

# Merge with existing config if present
if (Test-Path $configPath) {
    $existing = Get-Content $configPath | ConvertFrom-Json
    if ($existing.mcpServers) {
        $existing.mcpServers."agent-payment" = $mcpConfig.mcpServers."agent-payment"
        $mcpConfig = $existing
    }
}

# Write config
$mcpConfig | ConvertTo-Json -Depth 10 | Set-Content $configPath
Write-Host "✓ Updated $configPath" -ForegroundColor Green

Write-Host ""
Write-Host "✓ Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Restart $Client" -ForegroundColor White
if (-not $ApiKey -or -not $BudgetKey) {
    Write-Host "  2. Get your API keys from https://agentpmt.com" -ForegroundColor White
    Write-Host "  3. Edit $configPath" -ForegroundColor White
    Write-Host "     and add your AGENTPMT_API_KEY and AGENTPMT_BUDGET_KEY" -ForegroundColor White
}
Write-Host ""
