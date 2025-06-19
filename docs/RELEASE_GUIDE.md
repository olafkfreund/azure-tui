# Release Creation Guide

## 🚀 How to Create a Release Using CI/CD

### Quick Reference

```bash
# Method 1: Using the release script
just create-release v1.0.0 "Major release with search functionality"

# Method 2: Manual git commands
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Method 3: GitHub web interface (Recommended)
# Go to GitHub → Releases → Create a new release
```

## 📋 Release Process Overview

### 1. **Pre-Release Checklist**

```bash
# Check if repository is ready for release
just check-release

# This will:
# ✅ Run full QA (tests, linting, security)
# ✅ Verify working directory is clean
# ✅ Show recent tags for version reference
```

### 2. **Release Trigger Methods**

#### **🎯 Method 1: GitHub Web Interface (Recommended)**

1. Go to your GitHub repository
2. Click **"Releases"** → **"Create a new release"**
3. **Tag version**: Enter version (e.g., `v1.0.0`)
4. **Release title**: e.g., "Azure TUI v1.0.0"
5. **Description**: The CI will auto-generate comprehensive notes
6. Click **"Publish release"**

#### **🎯 Method 2: Command Line (Justfile)**

```bash
# Create release with description
just create-release v1.0.0 "Major release with advanced search"

# Create release with default description
just create-release v1.0.0
```

#### **🎯 Method 3: Git Commands**

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0 - Advanced search functionality"

# Push tag to trigger CI
git push origin v1.0.0
```

### 3. **CI/CD Release Workflow**

When you create a release (publish on GitHub or push a tag), the CI automatically:

#### **🔄 Automated Steps:**

1. **Quality Assurance**
   - ✅ Runs all tests (unit, integration, search functionality)
   - ✅ Verifies code formatting and linting
   - ✅ Checks security with gosec (when enabled)

2. **Multi-Platform Builds**
   - 🐧 Linux AMD64
   - 🍎 macOS AMD64 & ARM64
   - 🪟 Windows AMD64
   - 📦 Creates checksums for all binaries

3. **Release Asset Creation**
   - 📁 Bundles all binaries
   - 📋 Copies documentation (README, docs/)
   - 🔒 Generates checksums.txt
   - 📦 Creates release archive

4. **GitHub Release**
   - 📤 Uploads all binary artifacts
   - 📝 Generates comprehensive release notes
   - 🔗 Provides download links and installation instructions
   - 📊 Includes feature documentation and usage examples

### 4. **Release Versioning**

#### **🏷️ Semantic Versioning (Recommended)**

```bash
# Major release (breaking changes)
just create-release v2.0.0 "Major rewrite with new TUI framework"

# Minor release (new features)
just create-release v1.1.0 "Added Kubernetes support and advanced search"

# Patch release (bug fixes)
just create-release v1.0.1 "Fixed memory leak in search engine"

# Pre-release versions
just create-release v1.1.0-beta.1 "Beta release with experimental features"
just create-release v1.1.0-rc.1 "Release candidate"
```

#### **📅 Date-based Versioning**

```bash
just create-release v2024.06.19 "Monthly release June 2024"
```

### 5. **Release Monitoring**

#### **📊 Check Release Status**

```bash
# Show current release status
just release-status

# List recent releases
just list-releases

# Monitor CI pipeline
# GitHub → Actions → View workflow runs
```

#### **🔍 Release Verification**

```bash
# After release is published, verify:
curl -s https://api.github.com/repos/YOUR_ORG/azure-tui/releases/latest | jq '.tag_name'

# Download and test released binary
wget https://github.com/YOUR_ORG/azure-tui/releases/download/v1.0.0/azure-tui-linux-amd64
chmod +x azure-tui-linux-amd64
./azure-tui-linux-amd64 --version
```

## 📦 What Gets Included in a Release

### **🎯 Automated Release Assets:**

- ✅ **Linux Binary** (AMD64)
- ✅ **Windows Executable** (AMD64)
- ✅ **macOS Binaries** (AMD64 + ARM64)
- ✅ **Checksums File** (SHA256 verification)
- ✅ **Release Archive** (All binaries + docs)
- ✅ **Installation Instructions**
- ✅ **Feature Documentation**

### **📋 Generated Release Notes Include:**

- 🚀 **Key Features Overview**
- 🔍 **Search Engine Documentation**
- 📥 **Installation Instructions**
- 🎮 **Quick Start Guide**
- 📦 **Download Matrix** (Platform/Architecture)
- 🔐 **Security & Verification**
- 📋 **System Requirements**
- 🐛 **Known Issues & Workarounds**
- 🆘 **Support Links**

## 🎯 Example Release Workflow

```bash
# 1. Ensure everything is ready
just check-release

# 2. Create and push release
just create-release v1.2.0 "Enhanced search with wildcard support and performance improvements"

# 3. Monitor CI progress
# → Go to GitHub Actions to watch the pipeline

# 4. Verify release
# → Check GitHub Releases page for new release
# → Download and test binaries
# → Verify checksums

# 5. Announce release
# → Update README if needed
# → Share with team/community
```

## 🚨 Troubleshooting

### **Common Issues:**

1. **CI Fails During Release**
   ```bash
   # Check what failed in the CI
   # Usually: tests, linting, or build issues
   just qa-full  # Run locally to debug
   ```

2. **Tag Already Exists**
   ```bash
   # Delete tag locally and remotely
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   
   # Create new tag
   just create-release v1.0.1 "Fixed release"
   ```

3. **Missing Binaries in Release**
   ```bash
   # Check if build matrix completed
   # Look at GitHub Actions → Build job → Matrix strategy
   ```

## 🎉 Success!

After a successful release:
- ✅ Binaries are available for download
- ✅ Release notes are automatically generated
- ✅ GitHub Releases page is updated
- ✅ Users can install with simple commands
- ✅ CI/CD pipeline validates everything

Your Azure TUI release is now live and ready for users! 🚀
