#!/bin/bash

# Debug script for GitHub Actions release issues
# This script helps diagnose common problems with the release workflow

set -e

echo "🔍 Azure TUI Release Debugging Script"
echo "=====================================
"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

echo "✅ Git repository detected"

# Check if we have a remote origin
if ! git remote get-url origin > /dev/null 2>&1; then
    echo "❌ No origin remote found"
    exit 1
fi

REPO_URL=$(git remote get-url origin)
echo "✅ Repository URL: $REPO_URL"

# Extract owner and repo from URL
if [[ $REPO_URL =~ github\.com[:/]([^/]+)/([^/.]+) ]]; then
    OWNER=${BASH_REMATCH[1]}
    REPO=${BASH_REMATCH[2]}
    echo "✅ Repository: $OWNER/$REPO"
else
    echo "❌ Could not parse GitHub repository URL"
    exit 1
fi

# Check current branch and tags
CURRENT_BRANCH=$(git branch --show-current)
echo "📝 Current branch: $CURRENT_BRANCH"

echo ""
echo "🏷️  Recent tags:"
git tag --sort=-version:refname | head -5 || echo "No tags found"

echo ""
echo "📊 Repository status:"
git status --porcelain

# Check if GitHub CLI is available
if command -v gh > /dev/null 2>&1; then
    echo ""
    echo "🔧 GitHub CLI found, checking authentication..."
    if gh auth status > /dev/null 2>&1; then
        echo "✅ GitHub CLI authenticated"
        
        echo ""
        echo "🔍 Checking repository permissions..."
        gh api repos/$OWNER/$REPO --jq '.permissions' 2>/dev/null || echo "Could not check permissions"
        
        echo ""
        echo "📦 Recent releases:"
        gh release list --limit 3 2>/dev/null || echo "No releases found or insufficient permissions"
        
    else
        echo "❌ GitHub CLI not authenticated. Run: gh auth login"
    fi
else
    echo "⚠️  GitHub CLI not found. Install with: curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg"
fi

echo ""
echo "🔍 Checking workflow files..."
if [ -f ".github/workflows/ci.yml" ]; then
    echo "✅ CI workflow found"
    
    # Check for common issues in workflow
    if grep -q "softprops/action-gh-release@v1" .github/workflows/ci.yml; then
        echo "✅ Using action-gh-release@v1"
    elif grep -q "softprops/action-gh-release@v2" .github/workflows/ci.yml; then
        echo "⚠️  Using action-gh-release@v2 (consider v1 for better compatibility)"
    fi
    
    if grep -q "contents: write" .github/workflows/ci.yml; then
        echo "✅ Has contents write permission"
    else
        echo "❌ Missing contents write permission"
    fi
    
else
    echo "❌ No CI workflow found at .github/workflows/ci.yml"
fi

echo ""
echo "🏗️  Checking build artifacts..."
if [ -d "build" ]; then
    echo "✅ Build directory exists"
    ls -la build/ | grep azure-tui || echo "No azure-tui binaries found"
else
    echo "❌ No build directory found"
fi

echo ""
echo "📋 Release Troubleshooting Tips:"
echo "================================="
echo ""
echo "1. **Permission Issues:**"
echo "   - Ensure repository has 'Actions' enabled in Settings > Actions"
echo "   - Check that workflow has 'contents: write' permission"
echo "   - Verify GITHUB_TOKEN has necessary scopes"
echo ""
echo "2. **Tag Issues:**"
echo "   - Create tag: git tag v1.0.0 && git push origin v1.0.0"
echo "   - Ensure tag follows semantic versioning (v1.0.0, not 1.0.0)"
echo ""
echo "3. **Workflow Trigger Issues:**"
echo "   - Check workflow triggers: push tags, release events"
echo "   - Verify branch protection rules don't block workflow"
echo ""
echo "4. **File Path Issues:**"
echo "   - Ensure all files in 'files:' section exist"
echo "   - Check file paths are relative to workflow workspace"
echo ""
echo "5. **Action Version Issues:**"
echo "   - Use softprops/action-gh-release@v1 for stability"
echo "   - Consider pinning to specific commit SHA"
echo ""

echo "🔧 Suggested fixes for current setup:"
echo "======================================"
echo ""

# Suggest creating a tag if none exist
if ! git tag | head -1 > /dev/null 2>&1; then
    echo "📌 Create your first tag:"
    echo "   git tag v1.0.0"
    echo "   git push origin v1.0.0"
    echo ""
fi

# Suggest building if no artifacts
if [ ! -d "build" ] || [ -z "$(ls -A build 2>/dev/null)" ]; then
    echo "🏗️  Build your application:"
    echo "   just build-all"
    echo "   # or"
    echo "   go build -o build/azure-tui-linux-amd64 ./cmd/main.go"
    echo ""
fi

echo "✅ Debug script completed!"
echo ""
echo "Next steps:"
echo "1. Fix any issues identified above"
echo "2. Commit and push changes"
echo "3. Create/push a tag to trigger the release workflow"
echo "4. Monitor the Actions tab for workflow execution"
