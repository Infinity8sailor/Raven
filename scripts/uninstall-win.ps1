# ========================================================
# Raven - Windows Auto-Start Uninstaller
# Removes Startup shortcut and HKCU Run registry entry.
# ========================================================

$ErrorActionPreference = "SilentlyContinue"

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "       Raven Windows Uninstaller          " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# Remove Startup shortcut
$startupDir = [Environment]::GetFolderPath([Environment+SpecialFolder]::Startup)
$shortcutPath = Join-Path $startupDir "Raven.lnk"
if (Test-Path $shortcutPath) {
    Remove-Item -Path $shortcutPath -Force
    Write-Host "Removed Startup shortcut: $shortcutPath" -ForegroundColor Green
} else {
    Write-Host "Startup shortcut was not present." -ForegroundColor DarkGray
}

# Remove HKCU Run registry
$regPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
Remove-ItemProperty -Path $regPath -Name "RavenApp" -ErrorAction SilentlyContinue
Write-Host "Removed HKCU Run registry key." -ForegroundColor Green

# Remove legacy scheduled task if present
Unregister-ScheduledTask -TaskName "RavenApp" -Confirm:$false -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Raven auto-start has been removed." -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
