<#
.SYNOPSIS
    Automated build process for Raven with Git version verification.

.DESCRIPTION
    1. Checks the target version against existing Git tags (prevents duplicate/repeated releases).
    2. Runs the full Go test suite before compilation.
    3. Injects Version, GitCommit, and BuildDate into the binary via Go linker flags.
    4. Validates the compiled binary via `raven.exe -version`.

.PARAMETER Version
    Target release version (e.g., 'v1.0.0'). Defaults to the content of the VERSION file.

.PARAMETER Force
    Allows rebuilding an existing version tag without aborting.

.PARAMETER SkipTests
    Skips test execution before compilation.

.PARAMETER Tag
    Creates an annotated Git tag for the version after successful compilation.
#>

param (
    [string]$Version,
    [switch]$Force,
    [switch]$SkipTests,
    [switch]$Tag
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$appDir = (Resolve-Path "$scriptDir\..").Path
Push-Location $appDir

try {
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "         Raven Automated Build            " -ForegroundColor Cyan
    Write-Host "==========================================" -ForegroundColor Cyan

    # 1. Determine Target Version
    if (-not $Version) {
        $versionFile = Join-Path $appDir "VERSION"
        if (Test-Path $versionFile) {
            $Version = (Get-Content $versionFile -Raw).Trim()
        } else {
            $Version = "v1.0.0"
        }
    }

    if (-not $Version.StartsWith("v") -and -not [string]::IsNullOrWhiteSpace($Version)) {
        $Version = "v$Version"
    }

    Write-Host "Target Version : $Version" -ForegroundColor Yellow

    # 2. Version Check: Ensure version is NOT repeated in Git
    $existingTag = $(git tag -l $Version 2>$null)
    if ($existingTag -and -not $Force) {
        Write-Host ""
        Write-Host "=====================================================================" -ForegroundColor Red
        Write-Host " [ERROR] VERSION COLLISION: Git tag '$Version' already exists!" -ForegroundColor Red
        Write-Host " Versions must be strictly unique and not repeated." -ForegroundColor Red
        Write-Host " Tagged commit: $(git rev-parse --short $Version 2>$null)" -ForegroundColor Yellow
        Write-Host ""
        Write-Host " Actions to resolve:" -ForegroundColor Yellow
        Write-Host "  1. Bump the version in the VERSION file (e.g., v1.0.1)." -ForegroundColor Gray
        Write-Host "  2. Pass a new version: .\scripts\build.ps1 -Version v1.0.1" -ForegroundColor Gray
        Write-Host "  3. Use -Force to intentionally rebuild an existing version." -ForegroundColor Gray
        Write-Host "=====================================================================" -ForegroundColor Red
        Write-Host ""
        exit 1
    } elseif ($existingTag -and $Force) {
        Write-Host " [WARNING] Building existing version '$Version' (-Force enabled)." -ForegroundColor DarkYellow
    } else {
        Write-Host " [OK] Version '$Version' is unique and unreleased." -ForegroundColor Green
    }

    # 3. Git Metadata
    $gitCommit = $(git rev-parse --short HEAD 2>$null)
    if (-not $gitCommit) { $gitCommit = "dev" }
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

    Write-Host "Git Commit     : $gitCommit" -ForegroundColor Gray
    Write-Host "Build Timestamp: $buildDate" -ForegroundColor Gray

    # 4. Run Test Suite
    if (-not $SkipTests) {
        Write-Host ""
        Write-Host "[1/3] Running Go test suite..." -ForegroundColor Yellow
        & go test -v ./...
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Test suite failed! Aborting compilation."
            exit 1
        }
        Write-Host "      All tests passed." -ForegroundColor Green
    } else {
        Write-Host "[1/3] Skipping tests (-SkipTests specified)." -ForegroundColor DarkGray
    }

    # 5. Compile Binary with Linker Flags
    Write-Host ""
    Write-Host "[2/3] Compiling binary with linker flags..." -ForegroundColor Yellow
    $ldflags = "-s -w -X main.Version=$Version -X main.GitCommit=$gitCommit -X main.BuildDate=$buildDate"
    & go build -ldflags "$ldflags" -o raven.exe .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Go build failed!"
        exit 1
    }
    Write-Host "      Binary created: $appDir\raven.exe" -ForegroundColor Green

    # 6. Verify Executable
    Write-Host ""
    Write-Host "[3/3] Validating compiled executable..." -ForegroundColor Yellow
    $verString = & .\raven.exe -version
    Write-Host "      $verString" -ForegroundColor Cyan

    # 7. Optional Git Tag Creation
    if ($Tag -and (-not $existingTag)) {
        Write-Host ""
        Write-Host "Creating Git tag '$Version'..." -ForegroundColor Yellow
        git tag -a $Version -m "Release $Version"
        Write-Host "Git tag '$Version' created successfully." -ForegroundColor Green
    }

    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "       Build Completed Successfully!      " -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host ""
}
finally {
    Pop-Location
}
