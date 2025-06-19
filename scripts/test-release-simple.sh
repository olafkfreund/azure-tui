#!/bin/bash

# Simple release test script for Azure TUI
set -e

APP_NAME="azure-tui"
VERSION="v1.0.0"

echo "🚀 Testing Azure TUI Release Process"
echo "====================================="
echo "Version: $VERSION"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "Justfile" ]; then
    echo "❌ Not in Azure TUI project directory"
    exit 1
fi

echo "✅ In Azure TUI project directory"

# Check if builds exist
echo ""
echo "🔍 Checking build artifacts..."

if [ ! -d "build" ]; then
    echo "❌ Build directory missing. Running build..."
    just build-all
fi

EXPECTED_BINARIES=(
    "azure-tui-linux-amd64"
    "azure-tui-windows-amd64.exe"
    "azure-tui-darwin-amd64"
)

echo "📦 Build artifacts:"
for binary in "${EXPECTED_BINARIES[@]}"; do
    if [ -f "build/$binary" ]; then
        echo "✅ $binary ($(du -h "build/$binary" | cut -f1))"
    else
        echo "❌ $binary missing"
    fi
done

# Test release preparation
echo ""
echo "📁 Testing release preparation..."

# Clean and create release directory
rm -rf release
mkdir -p release

# Copy binaries
echo "📋 Copying binaries..."
for binary in "${EXPECTED_BINARIES[@]}"; do
    if [ -f "build/$binary" ]; then
        cp "build/$binary" "release/"
        echo "✅ Copied $binary"
    fi
done

# Copy documentation
echo "📖 Copying documentation..."
cp README.md release/
[ -f LICENSE ] && cp LICENSE release/ && echo "✅ Copied LICENSE"
[ -d docs ] && cp -r docs release/ && echo "✅ Copied docs"

# Create checksums
echo "🔐 Creating checksums..."
cd release
if ls ${APP_NAME}-* 1> /dev/null 2>&1; then
    sha256sum ${APP_NAME}-* > checksums.txt
    echo "✅ Created checksums.txt"
    echo "📋 Checksums:"
    cat checksums.txt
else
    echo "❌ No binaries found for checksums"
fi
cd ..

# Create release archive
echo ""
echo "📦 Creating release archive..."
cd release
tar -czf "../${APP_NAME}-${VERSION}-release.tar.gz" .
cd ..

if [ -f "${APP_NAME}-${VERSION}-release.tar.gz" ]; then
    echo "✅ Created release archive: $(du -h "${APP_NAME}-${VERSION}-release.tar.gz" | cut -f1)"
else
    echo "❌ Failed to create release archive"
    exit 1
fi

# List all release files
echo ""
echo "📋 Release files prepared:"
echo "=========================="
ls -la "${APP_NAME}-${VERSION}-release.tar.gz"
echo ""
echo "📁 Release directory contents:"
ls -la release/

echo ""
echo "🎯 Release preparation successful!"
echo ""
echo "Files ready for GitHub release:"
echo "- ${APP_NAME}-${VERSION}-release.tar.gz"
echo "- release/${APP_NAME}-linux-amd64"
echo "- release/${APP_NAME}-windows-amd64.exe" 
echo "- release/${APP_NAME}-darwin-amd64"
echo "- release/checksums.txt"

echo ""
echo "✅ Release test completed successfully!"
