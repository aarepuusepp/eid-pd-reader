# Build eid-pd-reader with NFC support on Windows
# Requires MSYS2 with MinGW64 and OpenPACE installed

$ErrorActionPreference = "Stop"

Write-Host "=== Building eid-pd-reader with NFC/PACE support ===" -ForegroundColor Cyan
Write-Host ""

# Check MSYS2
$msys2Path = "C:\msys64\mingw64"
$gccPath = "$msys2Path\bin\gcc.exe"

if (-not (Test-Path $gccPath)) {
    Write-Host "ERROR: MSYS2 MinGW64 not found at $msys2Path!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install MSYS2 and MinGW64 first:" -ForegroundColor Yellow
    Write-Host "  1. Download from: https://www.msys2.org/" -ForegroundColor Yellow
    Write-Host "  2. Install to C:\msys64" -ForegroundColor Yellow
    Write-Host "  3. Run: pacman -S mingw-w64-x86_64-gcc" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host "✓ Found MinGW64 GCC at $gccPath" -ForegroundColor Green

# Check OpenPACE
$openpacePath = "$msys2Path\lib\libeac.a"

if (-not (Test-Path $openpacePath)) {
    Write-Host "ERROR: OpenPACE not found at $openpacePath!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please build and install OpenPACE first." -ForegroundColor Yellow
    Write-Host "See ops/build-openpace.md for instructions." -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host "✓ Found OpenPACE library" -ForegroundColor Green
Write-Host ""

# Set environment
$env:CGO_ENABLED = "1"
$env:CC = $gccPath
$env:PKG_CONFIG_PATH = "$msys2Path\lib\pkgconfig"

Write-Host "Environment:" -ForegroundColor Cyan
Write-Host "  CGO_ENABLED = $env:CGO_ENABLED"
Write-Host "  CC = $env:CC"
Write-Host "  PKG_CONFIG_PATH = $env:PKG_CONFIG_PATH"
Write-Host ""

# Clean previous builds
Write-Host "Cleaning previous builds..." -ForegroundColor Yellow
if (Test-Path "eid-pd-reader-nfc.exe") {
    Remove-Item "eid-pd-reader-nfc.exe" -Force
}

# Build
Write-Host "Building with NFC support (this may take a minute)..." -ForegroundColor Green
go build -tags openpace -o eid-pd-reader-nfc.exe .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✓ Build successful!" -ForegroundColor Green
    Write-Host ""
    
    # Show version
    Write-Host "Version information:" -ForegroundColor Cyan
    .\eid-pd-reader-nfc.exe --version
    Write-Host ""
    
    # Check capabilities
    Write-Host "Checking NFC capabilities..." -ForegroundColor Cyan
    $caps = '{"command":"capabilities"}' | .\eid-pd-reader-nfc.exe 2>$null | ConvertFrom-Json
    if ($caps.data.nfcSupported) {
        Write-Host "✓ NFC/PACE support: ENABLED" -ForegroundColor Green
    } else {
        Write-Host "⚠ NFC/PACE support: DISABLED" -ForegroundColor Yellow
    }
    Write-Host ""
    
    Write-Host "Output: eid-pd-reader-nfc.exe" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To test NFC reading:" -ForegroundColor Yellow
    Write-Host "  .\eid-pd-reader-nfc.exe --nfc --can YOUR_CAN --dump" -ForegroundColor Yellow
    
} else {
    Write-Host ""
    Write-Host "✗ Build failed!" -ForegroundColor Red
    exit 1
}

