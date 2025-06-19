#!/bin/bash

# GitHub Actions Release Asset Validation Script
# This script mirrors the exact validation used in the CI/CD workflow

set -e

APP_NAME="azure-tui"
VERSION="${1:-v1.0.1}"  # Default to v1.0.1, but can be overridden

echo "🔍 Validating release assets for $VERSION..."
echo "============================================="
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "Justfile" ]; then
    echo "❌ Not in Azure TUI project directory"
    echo "Please run this script from the root of the azure-tui project"
    exit 1
fi

echo "✅ In Azure TUI project directory"
echo ""

# Define required files (same as in GitHub Actions workflow)
REQUIRED_FILES=(
    "${APP_NAME}-${VERSION}-release.tar.gz"
    "release/${APP_NAME}-linux-amd64"
    "release/${APP_NAME}-windows-amd64.exe"
    "release/${APP_NAME}-darwin-amd64"
    "release/${APP_NAME}-darwin-arm64"
    "release/checksums.txt"
)

echo "📋 Checking required files:"
echo "=========================="

# Track validation status
all_valid=true

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        size=$(du -h "$file" | cut -f1)
        echo "✅ $file exists ($size)"
    else
        echo "❌ $file missing"
        all_valid=false
    fi
done

echo ""

if [ "$all_valid" = true ]; then
    echo "🎉 SUCCESS: All release assets validated for $VERSION!"
    echo ""
    echo "📊 Summary:"
    echo "==========="
    
    # Show detailed information
    for file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            # Get file type
            if [[ "$file" == *.tar.gz ]]; then
                type="Release Archive"
            elif [[ "$file" == *linux* ]]; then
                type="Linux Binary"
            elif [[ "$file" == *windows* ]]; then
                type="Windows Binary"
            elif [[ "$file" == *darwin*amd64 ]]; then
                type="macOS Intel Binary"
            elif [[ "$file" == *darwin*arm64 ]]; then
                type="macOS Apple Silicon Binary"
            elif [[ "$file" == *checksums* ]]; then
                type="SHA256 Checksums"
            else
                type="Unknown"
            fi
            printf "%-40s %-25s %s\n" "$file" "$type" "$size"
        fi
    done
    
    echo ""
    echo "🚀 Ready for GitHub Actions workflow!"
    echo ""
    echo "Next steps:"
    echo "1. git add . && git commit -m 'Add release assets for $VERSION'"
    echo "2. git push origin main"
    echo "3. git push origin $VERSION  # Push the tag to trigger release workflow"
    echo "4. Monitor at: https://github.com/olafkfreund/azure-tui/actions"
    
else
    echo "❌ FAILURE: Some release assets are missing for $VERSION"
    echo ""
    echo "🔧 To create missing assets:"
    echo "============================"
    echo ""
    echo "1. Build all platforms:"
    echo "   just build-all"
    echo ""
    echo "2. Create release directory and copy binaries:"
    echo "   mkdir -p release"
    echo "   cp build/${APP_NAME}-* release/"
    echo ""
    echo "3. Add documentation:"
    echo "   cp README.md release/"
    echo "   cp LICENSE release/"
    echo ""
    echo "4. Create checksums:"
    echo "   cd release && sha256sum ${APP_NAME}-* > checksums.txt && cd .."
    echo ""
    echo "5. Create release archive:"
    echo "   cd release && tar -czf ../${APP_NAME}-${VERSION}-release.tar.gz . && cd .."
    echo ""
    echo "6. Re-run this validation:"
    echo "   ./scripts/validate-release-assets.sh $VERSION"
    
    exit 1
fi
