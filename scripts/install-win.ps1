# ========================================================
# Raven - Windows Auto-Start & Build Installer
# Works without Administrator privileges.
# ========================================================

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$appDir = (Resolve-Path "$scriptDir\..").Path
$exePath = Join-Path $appDir "raven.exe"

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "        Raven Windows Installer          " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "App Directory: $appDir" -ForegroundColor Gray

# Step 1: Build the latest executable if Go is available
$goCmd = Get-Command "go" -ErrorAction SilentlyContinue
if ($goCmd) {
    Write-Host "[1/3] Compiling latest raven.exe using Go..." -ForegroundColor Yellow
    Push-Location $appDir
    try {
        & go build -ldflags="-s -w" -o raven.exe main.go
        if ($LASTEXITCODE -eq 0 -and (Test-Path $exePath)) {
            Write-Host "      Build succeeded: $exePath" -ForegroundColor Green
        } else {
            Write-Warning "      Go build finished with non-zero exit code ($LASTEXITCODE)."
        }
    } catch {
        Write-Warning "      Go build failed: $_. Checking for existing raven.exe..."
    } finally {
        Pop-Location
    }
} else {
    Write-Host "[1/3] Go compiler not found on PATH. Looking for pre-built raven.exe..." -ForegroundColor Yellow
}

if (-not (Test-Path $exePath)) {
    Write-Error "raven.exe was not found at $exePath! Please install Go to build it, or download/copy raven.exe to the app directory."
    exit 1
}

# Step 2: Remove any legacy scheduled task if present
try {
    $existingTask = Get-ScheduledTask -TaskName "RavenApp" -ErrorAction SilentlyContinue
    if ($existingTask) {
        Unregister-ScheduledTask -TaskName "RavenApp" -Confirm:$false -ErrorAction SilentlyContinue
        Write-Host "[2/3] Cleaned up legacy scheduled task 'RavenApp'." -ForegroundColor DarkGray
    } else {
        Write-Host "[2/3] Checking environment..." -ForegroundColor DarkGray
    }
} catch {
    Write-Host "[2/3] Checking environment..." -ForegroundColor DarkGray
}

# Step 3: Install Windows Startup Folder Shortcut (User-level, 100% reliable, zero admin rights needed)
Write-Host "[3/3] Setting up Windows Auto-Start..." -ForegroundColor Yellow

$startupDir = [Environment]::GetFolderPath([Environment+SpecialFolder]::Startup)
$shortcutPath = Join-Path $startupDir "Raven.lnk"

try {
    $wscript = New-Object -ComObject WScript.Shell
    $shortcut = $wscript.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = $exePath
    $shortcut.WorkingDirectory = $appDir
    $shortcut.Description = "Raven - Minimal Offline Todo & Clipboard"
    $shortcut.IconLocation = "$exePath,0"
    $shortcut.Save()
    Write-Host "      Installed Startup shortcut: $shortcutPath" -ForegroundColor Green
} catch {
    Write-Warning "      Could not create shortcut in Startup folder: $_"
}

# Also register in Current User Run Key for maximum compatibility
try {
    $regPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
    Set-ItemProperty -Path $regPath -Name "RavenApp" -Value "`"$exePath`"" -ErrorAction Stop
    Write-Host "      Registered in HKCU Run registry: HKCU\...\Run\RavenApp" -ForegroundColor Green
} catch {
    Write-Warning "      Could not write HKCU Run registry entry: $_"
}

Write-Host ""
Write-Host "Raven successfully installed!" -ForegroundColor Green
Write-Host "Raven will start automatically whenever you log in to Windows." -ForegroundColor Cyan
Write-Host "Web Interface: http://localhost:9471" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
