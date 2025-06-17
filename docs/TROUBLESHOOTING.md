# Troubleshooting Guide 🔧

## Overview
Comprehensive troubleshooting guide for Azure TUI covering common issues, diagnostic steps, and solutions.

---

## Table of Contents
1. [Quick Diagnostics](#quick-diagnostics)
2. [Application Startup Issues](#application-startup-issues)
3. [Azure Integration Problems](#azure-integration-problems)
4. [AI Feature Issues](#ai-feature-issues)
5. [Performance Problems](#performance-problems)
6. [UI and Display Issues](#ui-and-display-issues)
7. [Configuration Problems](#configuration-problems)
8. [Advanced Troubleshooting](#advanced-troubleshooting)

---

## Quick Diagnostics

### Health Check Commands
```bash
# Quick system check
aztui --health-check

# Detailed diagnostics
aztui --diagnose

# Check specific components
aztui --check-azure
aztui --check-ai
aztui --check-config
```

### Common Quick Fixes
```bash
# Clear cache and restart
rm -rf ~/.cache/aztui/
aztui

# Reset configuration
mv ~/.config/aztui/config.yaml ~/.config/aztui/config.yaml.backup
aztui --generate-config

# Update Azure CLI
az upgrade
az login
```

---

## Application Startup Issues

### 1. Application Hangs on Startup

#### Symptoms:
- Application starts but doesn't show UI
- Cursor shows loading but nothing happens
- Terminal becomes unresponsive

#### Causes & Solutions:

**Azure CLI Authentication Issues:**
```bash
# Check Azure CLI status
az account show

# Re-authenticate
az login

# Clear Azure CLI cache
az account clear
az login
```

**Long Azure API Response Times:**
```bash
# Start in demo mode
aztui --demo

# Use shorter timeout
aztui --azure-timeout=10s

# Disable auto-refresh
aztui --no-auto-refresh
```

**Terminal Compatibility:**
```bash
# Try different terminal
export TERM=xterm-256color
aztui

# Disable mouse support
aztui --no-mouse

# Force fallback mode
aztui --fallback-ui
```

### 2. "Command not found" Error

#### Solution:
```bash
# Check if binary is in PATH
which aztui

# Add to PATH if needed
export PATH=$PATH:/path/to/aztui

# Or run with full path
/path/to/aztui
```

### 3. Permission Denied

#### Solution:
```bash
# Make binary executable
chmod +x aztui

# Check file permissions
ls -la aztui

# Run with sudo (not recommended)
sudo ./aztui
```

---

## Azure Integration Problems

### 1. "Azure CLI not found"

#### Symptoms:
- Error message about Azure CLI
- Unable to load subscriptions
- Authentication failures

#### Solutions:
```bash
# Install Azure CLI (Ubuntu/Debian)
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Install Azure CLI (macOS)
brew install azure-cli

# Install Azure CLI (Windows)
# Download from: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli

# Verify installation
az --version
```

### 2. Authentication Failures

#### "Login required" Error:
```bash
# Interactive login
az login

# Device code login (for remote/headless)
az login --use-device-code

# Service principal login
az login --service-principal \
         --username $AZURE_CLIENT_ID \
         --password $AZURE_CLIENT_SECRET \
         --tenant $AZURE_TENANT_ID
```

#### "Token expired" Error:
```bash
# Refresh token
az account get-access-token --query accessToken --output tsv

# Re-authenticate
az logout
az login
```

### 3. Subscription Access Issues

#### "No subscriptions found":
```bash
# List available subscriptions
az account list

# Set default subscription
az account set --subscription "your-subscription-id"

# Check permissions
az role assignment list --assignee $(az account show --query user.name -o tsv)
```

#### "Access denied to subscription":
```bash
# Check current context
az account show

# Verify permissions
az resource list --query "length(@)"

# Contact Azure administrator for access
```

### 4. Resource Loading Problems

#### "No resource groups found":
```bash
# Verify resource groups exist
az group list

# Check resource group permissions
az group list --query "[].{Name:name, Location:location}" -o table

# Try different subscription
az account set --subscription "another-subscription"
```

#### "Resources not loading":
```bash
# Test resource listing
az resource list -g "resource-group-name"

# Check network connectivity
ping management.azure.com

# Increase timeout
aztui --azure-timeout=60s
```

---

## AI Feature Issues

### 1. AI Analysis Not Working

#### "API key not configured":
```bash
# Set API key
export AZURE_TUI_AI_API_KEY="your-api-key"

# Or in config file
echo "ai.api_key: your-api-key" >> ~/.config/aztui/config.yaml
```

#### "Invalid API key":
```bash
# Test API key directly
curl -H "Authorization: Bearer $AZURE_TUI_AI_API_KEY" \
     https://api.openai.com/v1/models

# Generate new API key from OpenAI dashboard
```

### 2. AI Requests Timing Out

#### Solutions:
```bash
# Increase AI timeout
aztui --ai-timeout=120s

# Use shorter model
export AZURE_TUI_AI_MODEL="gpt-3.5-turbo"

# Enable streaming (if supported)
aztui --ai-streaming
```

### 3. Rate Limiting Issues

#### "Rate limit exceeded":
```yaml
# Configure rate limiting in config.yaml
ai:
  rate_limiting:
    requests_per_minute: 30
    retry_attempts: 3
    backoff_multiplier: 2
```

### 4. Azure OpenAI Issues

#### Configuration:
```bash
# Azure OpenAI environment variables
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_API_KEY="your-key"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"

# Test Azure OpenAI connectivity
curl -H "api-key: $AZURE_OPENAI_API_KEY" \
     "$AZURE_OPENAI_ENDPOINT/openai/deployments?api-version=2023-03-15-preview"
```

---

## Performance Problems

### 1. Slow Loading Times

#### Large Azure Environments:
```yaml
# Optimize for large environments
performance:
  azure:
    lazy_loading: true
    max_resources_per_group: 500
    pagination_size: 50
  
ui:
  auto_refresh: false
  refresh_interval: 300s
```

#### Network Issues:
```bash
# Test network connectivity
ping management.azure.com
ping api.openai.com

# Use shorter timeouts
aztui --azure-timeout=15s --ai-timeout=30s
```

### 2. High Memory Usage

#### Solutions:
```yaml
# Memory optimization
performance:
  max_memory_usage: "256MB"
  gc_threshold: "128MB"
  resource_cache_size: 500
```

```bash
# Monitor memory usage
top -p $(pgrep aztui)

# Restart application periodically
# (for long-running sessions)
```

### 3. CPU Usage Issues

#### High CPU during startup:
```bash
# Disable intensive features
aztui --no-health-monitoring --no-auto-refresh

# Use minimal UI
aztui --minimal-ui
```

---

## UI and Display Issues

### 1. Display Corruption

#### Terminal Issues:
```bash
# Reset terminal
reset

# Clear screen
clear

# Try different TERM
export TERM=screen-256color
aztui
```

#### Unicode Problems:
```bash
# Check locale
locale

# Set UTF-8 locale
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8

# Disable Unicode
aztui --no-unicode
```

### 2. Color Issues

#### Terminal Color Support:
```bash
# Check color support
echo $COLORTERM
tput colors

# Force color mode
aztui --color=always

# Disable colors
aztui --no-color
```

### 3. Sizing Problems

#### Panel Sizing:
```yaml
# Adjust panel sizes in config.yaml
ui:
  panels:
    tree_width: 30    # Reduce tree panel
    content_width: 70 # Increase content panel
```

#### Small Terminal:
```bash
# Check terminal size
echo $COLUMNS x $LINES

# Use minimal interface
aztui --compact-ui

# Hide status bar
aztui --no-status-bar
```

---

## Configuration Problems

### 1. Config File Not Found

#### Solutions:
```bash
# Create config directory
mkdir -p ~/.config/aztui

# Generate default config
aztui --generate-config

# Use custom config location
aztui --config=/path/to/config.yaml
```

### 2. Invalid Configuration

#### YAML Syntax Errors:
```bash
# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('~/.config/aztui/config.yaml'))"

# Or use online YAML validator

# Reset to defaults
mv ~/.config/aztui/config.yaml ~/.config/aztui/config.yaml.broken
aztui --generate-config
```

### 3. Environment Variable Conflicts

#### Solutions:
```bash
# List all AZURE_TUI_* variables
env | grep AZURE_TUI_

# Clear conflicting variables
unset AZURE_TUI_THEME
unset AZURE_TUI_CONFIG

# Start fresh
env -i PATH=$PATH TERM=$TERM aztui
```

---

## Advanced Troubleshooting

### 1. Enable Debug Logging

#### Comprehensive Debug Mode:
```bash
# Enable all debug logging
export AZURE_TUI_DEBUG=true
export AZURE_TUI_LOG_LEVEL=debug
aztui --verbose 2>&1 | tee aztui-debug.log
```

#### Component-Specific Debugging:
```yaml
# In config.yaml
logging:
  level: debug
  components:
    azure_api: debug
    ai_requests: debug
    ui_events: debug
    performance: debug
```

### 2. Network Diagnostics

#### Azure Connectivity:
```bash
# Test Azure endpoints
curl -I https://management.azure.com
curl -I https://login.microsoftonline.com

# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY

# Test with different DNS
nslookup management.azure.com 8.8.8.8
```

#### AI Service Connectivity:
```bash
# Test OpenAI
curl -I https://api.openai.com

# Test with proxy
curl --proxy $HTTP_PROXY -I https://api.openai.com

# Check firewall
telnet api.openai.com 443
```

### 3. Process Diagnostics

#### System Resources:
```bash
# Check system resources
free -h
df -h
ulimit -a

# Monitor process
strace -p $(pgrep aztui)
lsof -p $(pgrep aztui)
```

#### Go Runtime Diagnostics:
```bash
# Enable Go debugging
export GODEBUG=gctrace=1,schedtrace=1000
aztui

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 4. File System Issues

#### Permissions:
```bash
# Check config directory permissions
ls -la ~/.config/aztui/

# Fix permissions
chmod 755 ~/.config/aztui/
chmod 644 ~/.config/aztui/config.yaml
```

#### Disk Space:
```bash
# Check available space
df -h ~/.config/aztui/
df -h ~/.cache/aztui/

# Clean cache
rm -rf ~/.cache/aztui/
```

---

## Getting Help

### 1. Collecting Debug Information

#### Debug Report Script:
```bash
#!/bin/bash
# Create debug report
echo "=== Azure TUI Debug Report ===" > aztui-debug-report.txt
echo "Date: $(date)" >> aztui-debug-report.txt
echo "System: $(uname -a)" >> aztui-debug-report.txt
echo "Terminal: $TERM" >> aztui-debug-report.txt
echo "Shell: $SHELL" >> aztui-debug-report.txt
echo "" >> aztui-debug-report.txt

echo "=== Azure CLI ===" >> aztui-debug-report.txt
az --version >> aztui-debug-report.txt 2>&1
az account show >> aztui-debug-report.txt 2>&1
echo "" >> aztui-debug-report.txt

echo "=== Environment Variables ===" >> aztui-debug-report.txt
env | grep -E "(AZURE|TERM|LANG|LC_)" >> aztui-debug-report.txt
echo "" >> aztui-debug-report.txt

echo "=== Config File ===" >> aztui-debug-report.txt
cat ~/.config/aztui/config.yaml >> aztui-debug-report.txt 2>&1
echo "" >> aztui-debug-report.txt

echo "=== Recent Logs ===" >> aztui-debug-report.txt
tail -n 100 ~/.local/share/aztui/logs/aztui.log >> aztui-debug-report.txt 2>&1

echo "Debug report saved to: aztui-debug-report.txt"
```

### 2. GitHub Issues

#### Before Creating an Issue:
1. Search existing issues
2. Try latest version
3. Follow issue template
4. Include debug report
5. Provide reproduction steps

#### Issue Template:
```markdown
**Bug Description**
Brief description of the issue

**Steps to Reproduce**
1. Step one
2. Step two
3. Expected vs actual behavior

**Environment**
- OS: [e.g., Ubuntu 22.04]
- Terminal: [e.g., gnome-terminal]
- Azure CLI version: [output of `az --version`]
- Go version: [if building from source]

**Configuration**
[Relevant parts of config.yaml]

**Logs**
[Debug output or error messages]
```

### 3. Community Support

- **GitHub Discussions**: General questions and feature requests
- **Stack Overflow**: Tag questions with `azure-tui`
- **Discord/Slack**: Real-time community support (if available)

---

## Known Issues and Workarounds

### 1. Terminal Compatibility
- **Issue**: Some terminals don't support all Unicode characters
- **Workaround**: Use `--no-unicode` flag

### 2. Large Resource Groups
- **Issue**: Loading 1000+ resources causes slowdown
- **Workaround**: Use `--max-resources-per-group=500`

### 3. Network Proxies
- **Issue**: Corporate proxies may block API calls
- **Workaround**: Configure proxy settings in environment variables

### 4. Windows Subsystem for Linux (WSL)
- **Issue**: Azure CLI authentication may not work properly
- **Workaround**: Use `az login --use-device-code`

---

This troubleshooting guide covers most common issues. For additional help, check the GitHub repository or create an issue with detailed information.
