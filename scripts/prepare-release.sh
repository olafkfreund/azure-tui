#!/bin/bash

# Test and fix release workflow for Azure TUI
# This script helps prepare for a successful release

set -e

APP_NAME="azure-tui"
VERSION=${1:-"v1.0.0"}

echo "🚀 Azure TUI Release Preparation Script"
echo "========================================"
echo "Version: $VERSION"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "🔍 Checking prerequisites..."

if ! command_exists git; then
    echo "❌ Git not found"
    exit 1
fi

if ! command_exists go; then
    echo "❌ Go not found"
    exit 1
fi

if ! command_exists just; then
    echo "❌ Just not found. Install with: curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/bin"
    exit 1
fi

echo "✅ Prerequisites satisfied"

# Clean and build
echo ""
echo "🏗️  Building application..."
just clean
just build-all

# Verify builds
echo ""
echo "📦 Verifying builds..."
if [ ! -d "build" ]; then
    echo "❌ Build directory not found"
    exit 1
fi

EXPECTED_BINARIES=(
    "azure-tui-linux-amd64"
    "azure-tui-windows-amd64.exe"
    "azure-tui-darwin-amd64"
    "azure-tui-darwin-arm64"
)

for binary in "${EXPECTED_BINARIES[@]}"; do
    if [ -f "build/$binary" ]; then
        echo "✅ $binary"
    else
        echo "❌ $binary missing"
        exit 1
    fi
done

# Create release directory structure
echo ""
echo "📁 Creating release directory structure..."
mkdir -p release

# Copy binaries
for binary in "${EXPECTED_BINARIES[@]}"; do
    cp "build/$binary" "release/"
done

# Copy documentation
cp README.md release/
[ -f LICENSE ] && cp LICENSE release/ || echo "⚠️  No LICENSE file found"
[ -d docs ] && cp -r docs release/ || echo "⚠️  No docs directory found"

# Create checksums
echo ""
echo "🔐 Creating checksums..."
cd release
sha256sum ${APP_NAME}-* > checksums.txt
cd ..

# Create release archive
echo ""
echo "📦 Creating release archive..."
cd release
tar -czf "../${APP_NAME}-${VERSION}-release.tar.gz" .
cd ..

echo ""
echo "✅ Release preparation completed!"
echo ""
echo "📋 Release artifacts created:"
ls -la "${APP_NAME}-${VERSION}-release.tar.gz"
echo ""
echo "📁 Release directory contents:"
ls -la release/

# Git operations
echo ""
echo "🏷️  Git operations..."

# Check if tag exists
if git tag | grep -q "^${VERSION}$"; then
    echo "⚠️  Tag $VERSION already exists"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$VERSION" || true
        git push origin ":refs/tags/$VERSION" || true
        echo "🗑️  Deleted existing tag"
    else
        echo "❌ Aborting due to existing tag"
        exit 1
    fi
fi

# Create and push tag
echo "🏷️  Creating tag $VERSION..."
git tag "$VERSION"

echo ""
echo "🎯 Ready to release!"
echo ""
echo "Next steps:"
echo "1. Push the tag to trigger the release workflow:"
echo "   git push origin $VERSION"
echo ""
echo "2. Monitor the GitHub Actions workflow:"
echo "   https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\([^/]*\/[^/.]*\).*/\1/')/actions"
echo ""
echo "3. If the workflow fails, check:"
echo "   - Repository permissions (Settings > Actions > General)"
echo "   - GITHUB_TOKEN permissions"
echo "   - Workflow file syntax"
echo ""

# Offer to push tag
read -p "Push tag $VERSION now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🚀 Pushing tag..."
    git push origin "$VERSION"
    echo "✅ Tag pushed! Check GitHub Actions for workflow execution."
else
    echo "⏳ Tag created but not pushed. Push manually when ready:"
    echo "   git push origin $VERSION"
fi
