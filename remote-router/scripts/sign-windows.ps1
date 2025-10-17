# Windows Code Signing Script
# Requires: SignTool.exe from Windows SDK

param(
    [string]$File = "distribution/binaries/windows-amd64/agent-payment-router.exe",
    [string]$CertThumbprint = $env:CODESIGN_THUMBPRINT,
    [string]$TimestampUrl = "http://timestamp.digicert.com"
)

# Validate parameters
if (-not $CertThumbprint) {
    Write-Error "Certificate thumbprint required. Set CODESIGN_THUMBPRINT environment variable or pass -CertThumbprint"
    exit 1
}

if (-not (Test-Path $File)) {
    Write-Error "File not found: $File"
    exit 1
}

Write-Host "Signing $File with certificate $CertThumbprint..." -ForegroundColor Cyan

# Find SignTool.exe
$SignTool = "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe"
if (-not (Test-Path $SignTool)) {
    # Try to find it in PATH
    $SignTool = (Get-Command signtool.exe -ErrorAction SilentlyContinue).Source
    if (-not $SignTool) {
        Write-Error "SignTool.exe not found. Install Windows SDK."
        exit 1
    }
}

Write-Host "Using SignTool: $SignTool" -ForegroundColor Gray

# Sign the file
& $SignTool sign `
    /fd SHA256 `
    /tr $TimestampUrl `
    /td SHA256 `
    /sha1 $CertThumbprint `
    $File

if ($LASTEXITCODE -ne 0) {
    Write-Error "Signing failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host "✓ Successfully signed $File" -ForegroundColor Green

# Verify signature
Write-Host "Verifying signature..." -ForegroundColor Cyan
& $SignTool verify /pa $File

if ($LASTEXITCODE -ne 0) {
    Write-Warning "Signature verification failed"
    exit $LASTEXITCODE
}

Write-Host "✓ Signature verified successfully" -ForegroundColor Green
