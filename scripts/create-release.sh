#!/bin/bash

# Azure TUI Release Creation Script
# Usage: ./scripts/create-release.sh v1.0.0 "Release description"

set -e

VERSION="$1"
DESCRIPTION="$2"

if [ -z "$VERSION" ]; then
    echo "❌ Error: Version tag is required"
    echo "Usage: $0 <version> [description]"
    echo "Example: $0 v1.0.0 'Major release with search functionality'"
    exit 1
fi

if [ -z "$DESCRIPTION" ]; then
    DESCRIPTION="Azure TUI Release $VERSION"
fi

echo "🏷️  Creating release: $VERSION"
echo "📝 Description: $DESCRIPTION"
echo ""

# Ensure we're on main branch
echo "🔄 Switching to main branch..."
git checkout main
git pull origin main

# Ensure working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Working directory is not clean. Please commit or stash changes."
    git status
    exit 1
fi

# Create and push tag
echo "🏷️  Creating tag: $VERSION"
git tag -a "$VERSION" -m "$DESCRIPTION"

echo "📤 Pushing tag to GitHub..."
git push origin "$VERSION"

echo ""
echo "✅ Tag created and pushed successfully!"
echo ""
echo "🚀 The CI/CD pipeline will now:"
echo "   1. Run all tests and quality checks"
echo "   2. Build binaries for all platforms (Linux, macOS, Windows)"
echo "   3. Create GitHub release with artifacts"
echo "   4. Generate comprehensive release notes"
echo ""
echo "📍 Monitor the progress at:"
echo "   https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
echo ""
echo "🎉 Release will be available at:"
echo "   https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases/tag/$VERSION"
