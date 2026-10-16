#!/usr/bin/env bash
# ========================================================
# Raven - Automated Build with Git Version Verification
# ========================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$APP_DIR"

VERSION="${1:-}"
FORCE="${2:-false}"

if [ -z "$VERSION" ]; then
    if [ -f "$APP_DIR/VERSION" ]; then
        VERSION="$(cat "$APP_DIR/VERSION" | tr -d '[:space:]')"
    else
        VERSION="v1.0.0"
    fi
fi

if [[ "$VERSION" != v* ]]; then
    VERSION="v$VERSION"
fi

echo "=========================================="
echo "         Raven Automated Build            "
echo "=========================================="
echo "Target Version : $VERSION"

# Version check: Ensure version is not repeated
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    if [ "$FORCE" != "true" ] && [ "$FORCE" != "--force" ]; then
        echo ""
        echo "====================================================================="
        echo " [ERROR] VERSION COLLISION: Git tag '$VERSION' already exists!"
        echo " Versions must be strictly unique and not repeated."
        echo " To rebuild this version, run: bash scripts/build.sh $VERSION --force"
        echo "====================================================================="
        echo ""
        exit 1
    else
        echo " [WARNING] Building existing version '$VERSION' (--force enabled)."
    fi
else
    echo " [OK] Version '$VERSION' is unique and unreleased."
fi

GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "dev")"
BUILD_DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

echo "Git Commit     : $GIT_COMMIT"
echo "Build Timestamp: $BUILD_DATE"

echo ""
echo "[1/3] Running Go test suite..."
go test -v ./...
echo "      All tests passed."

echo ""
echo "[2/3] Compiling binary with linker flags..."
LDFLAGS="-s -w -X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT -X main.BuildDate=$BUILD_DATE"
go build -ldflags="$LDFLAGS" -o raven .
echo "      Binary created: $APP_DIR/raven"

echo ""
echo "[3/3] Validating compiled executable..."
./raven -version

echo ""
echo "=========================================="
echo "       Build Completed Successfully!      "
echo "=========================================="
