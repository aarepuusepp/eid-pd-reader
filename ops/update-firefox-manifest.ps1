# Update Firefox manifest with actual extension ID
# This script copies the updated Firefox host manifest to the installation directory

$ErrorActionPreference = "Stop"

$manifestPath = "C:\Program Files\eid-pd-reader\ee.aarepuusepp.eid_pd_reader.firefox.json"
$sourcePath = "..\host-manifests\ee.aarepuusepp.eid_pd_reader.firefox.json"

# Check if source exists
if (-not (Test-Path $sourcePath)) {
    Write-Error "Source manifest not found: $sourcePath"
    exit 1
}

Write-Host "Copying updated Firefox manifest..." -ForegroundColor Cyan
Write-Host "From: $sourcePath" -ForegroundColor Gray
Write-Host "To:   $manifestPath" -ForegroundColor Gray

try {
    Copy-Item $sourcePath $manifestPath -Force
    Write-Host "✓ Manifest copied successfully" -ForegroundColor Green
} catch {
    Write-Error "Failed to copy manifest. Run as Administrator?"
    exit 1
}

Write-Host "`nVerifying content..." -ForegroundColor Cyan
Get-Content $manifestPath | Write-Host -ForegroundColor Gray

Write-Host "`n✓ Done!" -ForegroundColor Green
Write-Host "Firefox manifest updated at: $manifestPath" -ForegroundColor Cyan

