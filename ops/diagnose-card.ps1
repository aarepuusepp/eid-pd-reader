# PowerShell script to diagnose card reader and card issues

$ErrorActionPreference = "Stop"

Write-Host "=== PC/SC Card Reader Diagnostics ===" -ForegroundColor Cyan
Write-Host ""

# Check PC/SC service
Write-Host "1. Checking PC/SC Service..." -ForegroundColor Yellow
$service = Get-Service SCardSvr
Write-Host "   Status: $($service.Status)" -ForegroundColor $(if ($service.Status -eq 'Running') { 'Green' } else { 'Red' })
Write-Host "   StartType: $($service.StartType)"

if ($service.Status -ne 'Running') {
    Write-Host "   Attempting to start service..." -ForegroundColor Yellow
    try {
        Start-Service SCardSvr
        Write-Host "   ✓ Service started successfully" -ForegroundColor Green
    } catch {
        Write-Host "   ✗ Failed to start service: $_" -ForegroundColor Red
        Write-Host "   Try running as Administrator" -ForegroundColor Yellow
    }
}

Write-Host ""

# List readers using certutil
Write-Host "2. Checking Smart Card Readers..." -ForegroundColor Yellow
try {
    $readers = certutil -scinfo 2>&1 | Select-String "Reader:"
    if ($readers) {
        $readers | ForEach-Object {
            Write-Host "   $($_.Line)" -ForegroundColor Green
        }
    } else {
        Write-Host "   No readers found via certutil" -ForegroundColor Red
    }
} catch {
    Write-Host "   Could not query readers: $_" -ForegroundColor Red
}

Write-Host ""

# Check for cards
Write-Host "3. Checking for Smart Cards..." -ForegroundColor Yellow
try {
    $cardInfo = certutil -scinfo 2>&1
    if ($cardInfo -match "Card:") {
        Write-Host "   Card detected" -ForegroundColor Green
        $cardInfo | Select-String "Card:", "ATR:" | ForEach-Object {
            Write-Host "   $($_.Line)" -ForegroundColor Cyan
        }
    } else {
        Write-Host "   No card detected or card not responding" -ForegroundColor Red
    }
} catch {
    Write-Host "   Could not query card: $_" -ForegroundColor Red
}

Write-Host ""

# Try our helper
Write-Host "4. Testing with eid-pd-reader..." -ForegroundColor Yellow
$helperPath = Join-Path $PSScriptRoot "..\eid-pd-reader.exe"
if (Test-Path $helperPath) {
    Write-Host "   Running: eid-pd-reader.exe --dump" -ForegroundColor Cyan
    & $helperPath --dump
    Write-Host ""
} else {
    Write-Host "   eid-pd-reader.exe not found at $helperPath" -ForegroundColor Red
    Write-Host "   Build it first: ops\build-all.ps1" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Troubleshooting Steps ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "If card is not responding (error 80100066):" -ForegroundColor Yellow
Write-Host "  1. Remove the card completely"
Write-Host "  2. Wait 2-3 seconds"
Write-Host "  3. Clean the card contacts with a soft cloth"
Write-Host "  4. Reinsert the card firmly"
Write-Host "  5. Wait for Windows to detect it (may take a few seconds)"
Write-Host "  6. Try again"
Write-Host ""
Write-Host "If still not working:" -ForegroundColor Yellow
Write-Host "  1. Restart the PC/SC service:"
Write-Host "     Restart-Service SCardSvr"
Write-Host "  2. Try a different USB port"
Write-Host "  3. Check if card works with other software (e.g., DigiDoc4)"
Write-Host "  4. Card may be damaged - try another card"
Write-Host ""
Write-Host "For Idemia cards specifically:" -ForegroundColor Yellow
Write-Host "  - Older Idemia cards may require specific drivers"
Write-Host "  - Ensure Estonian ID-card middleware is installed"
Write-Host "  - Some older cards may not work with all readers"
Write-Host ""

