# PowerShell script to install native host manifest on Windows

$ErrorActionPreference = "Stop"

$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$HOST_MANIFEST_CHROME = Join-Path $SCRIPT_DIR "..\host-manifests\ee.aarepuusepp.eid_pd_reader.json"
$HOST_MANIFEST_FIREFOX = Join-Path $SCRIPT_DIR "..\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json"
$HOSTNAME = "ee.aarepuusepp.eid_pd_reader"

# Check if manifests exist
if (-not (Test-Path $HOST_MANIFEST_CHROME)) {
    Write-Error "Chrome/Edge manifest not found: $HOST_MANIFEST_CHROME"
    exit 1
}

if (-not (Test-Path $HOST_MANIFEST_FIREFOX)) {
    Write-Error "Firefox manifest not found: $HOST_MANIFEST_FIREFOX"
    exit 1
}

# Get absolute path to eid-pd-reader.exe (assume it's in repo root)
$REPO_ROOT = Split-Path -Parent $SCRIPT_DIR
$HELPER_EXE = Join-Path $REPO_ROOT "eid-pd-reader.exe"

if (-not (Test-Path $HELPER_EXE)) {
    Write-Warning "eid-pd-reader.exe not found at $HELPER_EXE"
    Write-Host "Please build first using: go build -o eid-pd-reader.exe ./cmd/eid-pd-reader"
    Write-Host "Or use the build script: .\build.ps1"
    Write-Host "Then update the 'path' field in manifests to point to the absolute path of eid-pd-reader.exe"
}

# Read and update Chrome/Edge manifest with absolute path
$manifest = Get-Content $HOST_MANIFEST_CHROME | ConvertFrom-Json
$manifest.path = $HELPER_EXE
$manifest | ConvertTo-Json -Depth 10 | Set-Content $HOST_MANIFEST_CHROME

# Read and update Firefox manifest with absolute path
$manifestFirefox = Get-Content $HOST_MANIFEST_FIREFOX | ConvertFrom-Json
$manifestFirefox.path = $HELPER_EXE
$manifestFirefox | ConvertTo-Json -Depth 10 | Set-Content $HOST_MANIFEST_FIREFOX

Write-Host "`n=== Registering Native Host ===" -ForegroundColor Cyan

# Register for Chrome
$chromeKey = "HKLM:\Software\Google\Chrome\NativeMessagingHosts\$HOSTNAME"
New-Item -Path $chromeKey -Force | Out-Null
Set-ItemProperty -Path $chromeKey -Name "(default)" -Value $HOST_MANIFEST_CHROME -Type String
Write-Host "[OK] Chrome:  $chromeKey" -ForegroundColor Green

# Register for Edge
$edgeKey = "HKLM:\Software\Microsoft\Edge\NativeMessagingHosts\$HOSTNAME"
New-Item -Path $edgeKey -Force | Out-Null
Set-ItemProperty -Path $edgeKey -Name "(default)" -Value $HOST_MANIFEST_CHROME -Type String
Write-Host "[OK] Edge:    $edgeKey" -ForegroundColor Green

# Register for Firefox
$firefoxKey = "HKLM:\Software\Mozilla\NativeMessagingHosts\$HOSTNAME"
New-Item -Path $firefoxKey -Force | Out-Null
Set-ItemProperty -Path $firefoxKey -Name "(default)" -Value $HOST_MANIFEST_FIREFOX -Type String
Write-Host "[OK] Firefox: $firefoxKey" -ForegroundColor Green

Write-Host "`n=== Installation Complete ===" -ForegroundColor Green
Write-Host "Chrome/Edge Manifest: $HOST_MANIFEST_CHROME"
Write-Host "Firefox Manifest:     $HOST_MANIFEST_FIREFOX"
Write-Host "Helper:               $HELPER_EXE"

