# Build script for eid-pd-reader
# Ensures correct Go version is used

$ErrorActionPreference = "Stop"

# Set Go 1.27.3 paths
$env:GOROOT = "C:\Program Files\Go"
$env:PATH = "$env:GOROOT\bin;" + $env:PATH

Write-Host "=== Building eid-pd-reader ===" -ForegroundColor Cyan
Write-Host "Go version: " -NoNewline
go version

Write-Host "`nCleaning old build..." -ForegroundColor Yellow
Remove-Item eid-pd-reader.exe -ErrorAction SilentlyContinue

Write-Host "Building..." -ForegroundColor Yellow
go build -o eid-pd-reader.exe ./cmd/eid-pd-reader

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✓ Build successful!" -ForegroundColor Green
    Write-Host "`nVersion info:" -ForegroundColor Cyan
    .\eid-pd-reader.exe --version
} else {
    Write-Host "`n✗ Build failed!" -ForegroundColor Red
    exit 1
}

