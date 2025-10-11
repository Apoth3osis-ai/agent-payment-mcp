# Agent Payment MCP Server Installer (Windows)
param(
    [Parameter(Position=0)]
    [ValidateSet("claude", "cursor", "vscode")]
    [string]$Editor = "claude"
)

# Colors
function Write-Success { Write-Host "✅ $args" -ForegroundColor Green }
function Write-Error { Write-Host "❌ $args" -ForegroundColor Red }
function Write-Warning { Write-Host "⚠️  $args" -ForegroundColor Yellow }
function Write-Info { Write-Host "ℹ️  $args" -ForegroundColor Cyan }

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Agent Payment MCP Server Installer (Windows)" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Determine config path based on editor
$configPath = switch ($Editor) {
    "claude" { "$env:APPDATA\Claude\claude_desktop_config.json" }
    "cursor" { "$env:USERPROFILE\.cursor\mcp.json" }
    "vscode" { "$env:APPDATA\Code\User\globalStorage\claude-code\mcp.json" }
}

# Installation directory
$installDir = "$env:USERPROFILE\.agent-payment"

Write-Host "Editor: $Editor"
Write-Host "Config path: $configPath"
Write-Host "Install directory: $installDir"
Write-Host ""

# Create installation directory
Write-Info "Creating installation directory..."
New-Item -ItemType Directory -Path $installDir -Force | Out-Null

# Copy server executable
Write-Info "Installing server executable..."
Copy-Item "agent-payment-server.exe" "$installDir\" -Force

# Copy configuration
Write-Info "Installing configuration..."
Copy-Item "config.json" "$installDir\" -Force

# Configure editor
Write-Info "Configuring $Editor..."

# Create config directory
$configDir = Split-Path $configPath -Parent
New-Item -ItemType Directory -Path $configDir -Force | Out-Null

# Initialize config if doesn't exist
if (!(Test-Path $configPath)) {
    if ($Editor -eq "claude") {
        '{"mcpServers":{}}' | Out-File -FilePath $configPath -Encoding UTF8
    } else {
        '{"servers":{}}' | Out-File -FilePath $configPath -Encoding UTF8
    }
}

# Backup config
$backup = "$configPath.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"
Copy-Item $configPath $backup -Force
Write-Success "Backed up config to: $backup"

# Read and update config
try {
    $config = Get-Content $configPath -Raw | ConvertFrom-Json

    # Create server configuration
    $serverConfig = @{
        command = "$installDir\agent-payment-server.exe"
        args = @()
    }

    if ($Editor -ne "claude") {
        $serverConfig.type = "stdio"
    }

    # Add to config
    if ($Editor -eq "claude") {
        if (!$config.mcpServers) {
            $config | Add-Member -NotePropertyName "mcpServers" -NotePropertyValue @{} -Force
        }
        $config.mcpServers | Add-Member -NotePropertyName "agent-payment" `
            -NotePropertyValue $serverConfig -Force
    } else {
        if (!$config.servers) {
            $config | Add-Member -NotePropertyName "servers" -NotePropertyValue @{} -Force
        }
        $config.servers | Add-Member -NotePropertyName "agent-payment" `
            -NotePropertyValue $serverConfig -Force
    }

    # Save config
    $config | ConvertTo-Json -Depth 10 | Out-File -FilePath $configPath -Encoding UTF8
    Write-Success "Configuration updated automatically"

} catch {
    Write-Warning "Failed to update config automatically: $_"
    Write-Host ""
    Write-Host "Please add this to $configPath manually:" -ForegroundColor Yellow
    Write-Host ""
    if ($Editor -eq "claude") {
        Write-Host @"
{
  "mcpServers": {
    "agent-payment": {
      "command": "$installDir\agent-payment-server.exe",
      "args": []
    }
  }
}
"@
    } else {
        Write-Host @"
{
  "servers": {
    "agent-payment": {
      "type": "stdio",
      "command": "$installDir\agent-payment-server.exe",
      "args": []
    }
  }
}
"@
    }
    Write-Host ""
}

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Success "Installation Complete!"
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Server installed to: $installDir"
Write-Host ""
Write-Warning "Important: Restart $Editor to apply changes"
Write-Host ""
