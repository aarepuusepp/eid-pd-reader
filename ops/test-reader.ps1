# Test Smart Card Reader Detection
# Shows all connected PC/SC readers

$ErrorActionPreference = "Stop"

Write-Host "`n=== Smart Card Reader Detection Test ===" -ForegroundColor Cyan
Write-Host "This script tests if your card reader is detected by PC/SC`n" -ForegroundColor Yellow

# Check if Smart Card service is running
Write-Host "Checking Smart Card service..." -ForegroundColor Yellow
$service = Get-Service -Name "SCardSvr" -ErrorAction SilentlyContinue

if ($service) {
    if ($service.Status -eq "Running") {
        Write-Host "✓ Smart Card service is running" -ForegroundColor Green
    } else {
        Write-Host "✗ Smart Card service is stopped!" -ForegroundColor Red
        Write-Host "  Starting service..." -ForegroundColor Yellow
        Start-Service -Name "SCardSvr"
        Write-Host "✓ Service started" -ForegroundColor Green
    }
} else {
    Write-Host "✗ Smart Card service not found!" -ForegroundColor Red
    exit 1
}

# Test with eid-pd-reader
Write-Host "`nTesting with eid-pd-reader..." -ForegroundColor Yellow
$helperPath = "C:\Program Files\eid-pd-reader\eid-pd-reader.exe"

if (-not (Test-Path $helperPath)) {
    Write-Host "✗ eid-pd-reader not found!" -ForegroundColor Red
    Write-Host "  Please run install.cmd first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Running: eid-pd-reader.exe --dump" -ForegroundColor Gray
Write-Host ""

# Run eid-pd-reader with card detection
try {
    & $helperPath --dump 2>&1 | ForEach-Object {
        $line = $_.ToString()
        
        # Highlight important information
        if ($line -match "Connected to reader:") {
            Write-Host $line -ForegroundColor Green
        } elseif ($line -match "Card detected:") {
            Write-Host $line -ForegroundColor Green
        } elseif ($line -match "ATR:") {
            Write-Host $line -ForegroundColor Cyan
        } elseif ($line -match "error|Error|ERROR") {
            Write-Host $line -ForegroundColor Red
        } elseif ($line -match "Surname|Given|Personal Code|Document") {
            Write-Host $line -ForegroundColor White
        } else {
            Write-Host $line -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "✗ Error running eid-pd-reader: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Cyan
Write-Host "If you see 'Connected to reader: ACS ACR1252...' above," -ForegroundColor Yellow
Write-Host "then your NFC reader is working correctly!" -ForegroundColor Yellow

