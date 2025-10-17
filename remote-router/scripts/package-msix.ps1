# MSIX Packaging Script (Optional)
# Creates a signed MSIX package for Windows

param(
    [string]$Version = "1.0.0",
    [string]$CertThumbprint = $env:CODESIGN_THUMBPRINT
)

if (-not $CertThumbprint) {
    Write-Error "Certificate thumbprint required"
    exit 1
}

$Out = "distribution/packages/agent-payment-router-v${Version}.msix"
$Tmp = "distribution/msix-content"

Write-Host "Creating MSIX package for version $Version..." -ForegroundColor Cyan

# Clean and prepare
Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $Tmp | Out-Null

# Copy files
Write-Host "Copying files..." -ForegroundColor Gray
Copy-Item distribution/binaries/windows-amd64/agent-payment-router.exe $Tmp
Copy-Item ../../agent-payment-logo.png $Tmp -ErrorAction SilentlyContinue
Copy-Item windows/AppxManifest.xml $Tmp

# Update version in manifest
Write-Host "Updating manifest version..." -ForegroundColor Gray
(Get-Content "$Tmp/AppxManifest.xml") -replace '1\.0\.0\.0', "${Version}.0" | Set-Content "$Tmp/AppxManifest.xml"

# Ensure output directory exists
New-Item -ItemType Directory -Path (Split-Path $Out) -Force | Out-Null

# Create MSIX
Write-Host "Packaging MSIX..." -ForegroundColor Cyan
& makeappx.exe pack /d $Tmp /p $Out /l /o

if ($LASTEXITCODE -ne 0) {
    Write-Error "MSIX packaging failed"
    exit $LASTEXITCODE
}

# Sign MSIX
Write-Host "Signing MSIX package..." -ForegroundColor Cyan
$SignTool = "C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x64\signtool.exe"
if (-not (Test-Path $SignTool)) {
    $SignTool = (Get-Command signtool.exe -ErrorAction SilentlyContinue).Source
}

& $SignTool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /sha1 $CertThumbprint $Out

if ($LASTEXITCODE -ne 0) {
    Write-Error "MSIX signing failed"
    exit $LASTEXITCODE
}

Write-Host "✓ Successfully created and signed $Out" -ForegroundColor Green

# Clean up temp directory
Remove-Item -Recurse -Force $Tmp

# Show file info
Write-Host ""
Write-Host "Package details:" -ForegroundColor Cyan
Get-Item $Out | Format-List Name, Length, LastWriteTime
