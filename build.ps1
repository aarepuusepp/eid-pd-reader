# Build script for eid-pd-reader
# Ensures correct Go version is used

$ErrorActionPreference = "Stop"

# Set Go 1.27.3 paths
$env:GOROOT = "C:\Program Files\Go"
$env:PATH = "$env:GOROOT\bin;" + $env:PATH
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

# PKCS#11 support uses CGO. Prefer a modern MSYS2 MinGW-w64 compiler when present.
$preferredCompilers = @(
    "C:\msys64-2\ucrt64\bin\gcc.exe",
    "C:\msys64-2\mingw64\bin\gcc.exe",
    "C:\msys64-2\clang64\bin\clang.exe",
    "C:\msys64\ucrt64\bin\gcc.exe",
    "C:\msys64\mingw64\bin\gcc.exe",
    "C:\msys64\clang64\bin\clang.exe"
)

foreach ($compiler in $preferredCompilers) {
    if (Test-Path $compiler) {
        $env:CC = $compiler
        $env:PATH = (Split-Path $compiler) + ";" + $env:PATH
        break
    }
}

if (-not $env:CC) {
    $gcc = Get-Command gcc.exe -ErrorAction SilentlyContinue
    if ($gcc) {
        $env:CC = $gcc.Source
    }
}

Write-Host "=== Building eid-pd-reader ===" -ForegroundColor Cyan
Write-Host "Go version: " -NoNewline
go version
Write-Host "Target: $env:GOOS/$env:GOARCH, CGO_ENABLED=$env:CGO_ENABLED"
Write-Host "C compiler: $env:CC"
if ($env:CC -like "*TDM-GCC*") {
    Write-Host "`nWARNING: TDM-GCC has produced non-runnable CGO binaries in this project." -ForegroundColor Yellow
    Write-Host "Install MSYS2 and mingw-w64-ucrt-x86_64-gcc if the built exe fails to start." -ForegroundColor Yellow
}

Write-Host "`nCleaning old build..." -ForegroundColor Yellow
Remove-Item eid-pd-reader.exe -ErrorAction SilentlyContinue

Write-Host "Building..." -ForegroundColor Yellow
go build -o eid-pd-reader.exe ./cmd/eid-pd-reader

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild completed. Validating executable..." -ForegroundColor Yellow
    Write-Host "`nVersion info:" -ForegroundColor Cyan
    try {
        .\eid-pd-reader.exe --version
        Write-Host "`n✓ Build successful and executable starts." -ForegroundColor Green
    } catch {
        Write-Host "`n✗ Build output could not be started." -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
        Write-Host "`nThis usually means the CGO C compiler produced an invalid Windows binary." -ForegroundColor Yellow
        Write-Host "Recommended fix:" -ForegroundColor Yellow
        Write-Host "  1. Install MSYS2 from https://www.msys2.org/" -ForegroundColor Yellow
        Write-Host "  2. In MSYS2 UCRT64 shell run: pacman -S --needed mingw-w64-ucrt-x86_64-gcc" -ForegroundColor Yellow
        Write-Host "  3. Re-run this script from PowerShell." -ForegroundColor Yellow
        exit 1
    }
} else {
    Write-Host "`n✗ Build failed!" -ForegroundColor Red
    exit 1
}

