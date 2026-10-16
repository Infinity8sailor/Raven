# ========================================================
# Raven - Windows Auto-Start & Build Installer
# Works without Administrator privileges.
# ========================================================

param (
    [int]$Port = 0
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$appDir = (Resolve-Path "$scriptDir\..").Path
$exePath = Join-Path $appDir "raven.exe"
$configFile = Join-Path $appDir "config.json"

# Resolve port from param, config.json, or fallback
$resolvedPort = 9471
if (Test-Path $configFile) {
    try {
        $cfgJson = Get-Content $configFile -Raw | ConvertFrom-Json
        if ($cfgJson.port) { $resolvedPort = $cfgJson.port }
    } catch {}
}

if ($Port -gt 0) {
    $resolvedPort = $Port
    try {
        if (Test-Path $configFile) {
            $cfgJson = Get-Content $configFile -Raw | ConvertFrom-Json
            $cfgJson.port = $Port
            $cfgJson | ConvertTo-Json -Depth 4 | Set-Content $configFile -Encoding UTF8
        } else {
            @{ port = $Port; data_dir = "data"; web_dir = "web" } | ConvertTo-Json | Set-Content $configFile -Encoding UTF8
        }
        Write-Host "Configured custom port in config.json: $Port" -ForegroundColor Green
    } catch {
        Write-Warning "Could not update config.json with port $Port: $_"
    }
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "        Raven Windows Installer          " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "App Directory: $appDir" -ForegroundColor Gray
Write-Host "HTTP Port    : $resolvedPort" -ForegroundColor Gray

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

# Step 3: Install Windows Auto-Start (Registry & Startup Folder)
Write-Host "[3/3] Setting up Windows Auto-Start..." -ForegroundColor Yellow

# 3A. Auto-check and overwrite HKCU Run registry key
$regPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
try {
    $existingReg = (Get-ItemProperty -Path $regPath -Name "RavenApp" -ErrorAction SilentlyContinue).RavenApp
    if ($existingReg) {
        Write-Host "      [Registry] Existing auto-start key detected: $existingReg" -ForegroundColor Cyan
        if ($existingReg -ne "`"$exePath`"") {
            Write-Host "      [Registry] Path differs! Overwriting existing key with: `"$exePath`"..." -ForegroundColor Yellow
        } else {
            Write-Host "      [Registry] Refreshing/re-verifying existing key for: `"$exePath`"..." -ForegroundColor DarkGray
        }
    } else {
        Write-Host "      [Registry] No prior auto-start key found. Creating new key..." -ForegroundColor DarkGray
    }
    Set-ItemProperty -Path $regPath -Name "RavenApp" -Value "`"$exePath`"" -Force -ErrorAction Stop
    Write-Host "      [Registry] Auto-start key active -> `"$exePath`"" -ForegroundColor Green
} catch {
    Write-Warning "      [Registry] Could not write HKCU Run registry entry: $_"
}

# 3B. Auto-check and overwrite Startup folder shortcut
$startupDir = [Environment]::GetFolderPath([Environment+SpecialFolder]::Startup)
$shortcutPath = Join-Path $startupDir "Raven.lnk"
try {
    $wscript = New-Object -ComObject WScript.Shell
    if (Test-Path $shortcutPath) {
        $oldShortcut = $wscript.CreateShortcut($shortcutPath)
        Write-Host "      [Shortcut] Existing Startup shortcut detected: $($oldShortcut.TargetPath)" -ForegroundColor Cyan
        Write-Host "      [Shortcut] Overwriting shortcut target with: $exePath..." -ForegroundColor Yellow
    }
    $shortcut = $wscript.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = $exePath
    $shortcut.WorkingDirectory = $appDir
    $shortcut.Description = "Raven - Minimal Offline Todo & Clipboard"
    $shortcut.IconLocation = "$exePath,0"
    $shortcut.Save()
    Write-Host "      [Shortcut] Startup shortcut active -> $shortcutPath" -ForegroundColor Green
} catch {
    Write-Warning "      [Shortcut] Could not configure shortcut in Startup folder: $_"
}

Write-Host ""
Write-Host "Raven successfully installed!" -ForegroundColor Green
Write-Host "Raven will start automatically whenever you log in to Windows." -ForegroundColor Cyan
Write-Host "Web Interface: http://localhost:$resolvedPort" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
