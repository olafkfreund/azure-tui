#!/bin/bash

# Test Release Workflow Script
# This script helps test and debug the release process locally

set -e

echo "🧪 Testing Release Workflow Components"
echo "======================================"
echo ""

# Test 1: Check if all required tools are available
echo "1️⃣ Checking required tools..."
MISSING_TOOLS=()

if ! command -v git &> /dev/null; then
    MISSING_TOOLS+=("git")
fi

if ! command -v just &> /dev/null; then
    MISSING_TOOLS+=("just")
fi

if ! command -v go &> /dev/null; then
    MISSING_TOOLS+=("go")
fi

if [ ${#MISSING_TOOLS[@]} -eq 0 ]; then
    echo "✅ All required tools are available"
else
    echo "❌ Missing tools: ${MISSING_TOOLS[*]}"
    exit 1
fi

# Test 2: Check git repository status
echo ""
echo "2️⃣ Checking git repository status..."
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  Working directory has uncommitted changes:"
    git status --short
    echo ""
    echo "💡 Consider committing changes before creating a release"
else
    echo "✅ Working directory is clean"
fi

# Test 3: Run build and tests
echo ""
echo "3️⃣ Testing build process..."
if just build > /dev/null 2>&1; then
    echo "✅ Build: Success"
else
    echo "❌ Build: Failed"
    echo "Run 'just build' to see detailed error"
    exit 1
fi

echo ""
echo "4️⃣ Testing test suite..."
if just test > /dev/null 2>&1; then
    echo "✅ Tests: Pass"
else
    echo "❌ Tests: Failed"
    echo "Run 'just test' to see detailed error"
    exit 1
fi

# Test 4: Check existing tags
echo ""
echo "5️⃣ Checking existing tags..."
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
echo "📋 Latest tag: $LATEST_TAG"

if [ "$LATEST_TAG" != "none" ]; then
    echo "📋 Recent tags:"
    git tag --sort=-version:refname | head -5
fi

# Test 5: Suggest next version
echo ""
echo "6️⃣ Version suggestions..."
if [ "$LATEST_TAG" != "none" ]; then
    # Extract version number and suggest next versions
    if [[ $LATEST_TAG =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        
        NEXT_PATCH="v$MAJOR.$MINOR.$((PATCH + 1))"
        NEXT_MINOR="v$MAJOR.$((MINOR + 1)).0"
        NEXT_MAJOR="v$((MAJOR + 1)).0.0"
        
        echo "💡 Suggested next versions:"
        echo "   Patch (bug fixes): $NEXT_PATCH"
        echo "   Minor (new features): $NEXT_MINOR"
        echo "   Major (breaking changes): $NEXT_MAJOR"
    fi
else
    echo "💡 Suggested first version: v1.0.0"
fi

# Test 6: Check GitHub connectivity (if in CI or has gh CLI)
echo ""
echo "7️⃣ Checking GitHub connectivity..."
if command -v gh &> /dev/null; then
    if gh auth status &> /dev/null; then
        echo "✅ GitHub CLI authenticated"
        REPO_INFO=$(gh repo view --json owner,name 2>/dev/null || echo "")
        if [ -n "$REPO_INFO" ]; then
            echo "📂 Repository: $(echo "$REPO_INFO" | jq -r '.owner.login + "/" + .name')"
        fi
    else
        echo "⚠️  GitHub CLI not authenticated"
        echo "💡 Run 'gh auth login' for enhanced GitHub integration"
    fi
else
    echo "⚠️  GitHub CLI not found"
    echo "💡 Install 'gh' CLI for enhanced GitHub integration"
fi

echo ""
echo "🎯 Release Process Test Complete!"
echo ""
echo "📝 Next Steps:"
echo "   1. Fix any issues mentioned above"
echo "   2. Choose your next version (see suggestions above)"
echo "   3. Create release:"
echo "      • GitHub UI: Go to Releases → Create new release"
echo "      • Command line: just create-release vX.Y.Z \"Description\""
echo "      • Git tag: git tag -a vX.Y.Z -m \"Release vX.Y.Z\" && git push origin vX.Y.Z"
echo ""
echo "🚀 The CI/CD pipeline will handle the rest automatically!"
