# Azure TUI - Complete User Manual 📚

## Table of Contents
1. [Quick Start](#quick-start)
2. [Real Azure Integration](#real-azure-integration)
3. [Tree View Navigation](#tree-view-navigation)
4. [Resource Management](#resource-management)
5. [AI-Powered Features](#ai-powered-features)
6. [Troubleshooting Guide](#troubleshooting-guide)
7. [Best Practices](#best-practices)

---

## Quick Start

### Installation & Setup

1. **Prerequisites**:
   ```bash
   # Ensure Azure CLI is installed and authenticated
   az --version
   az login
   az account show
   ```

2. **Build & Run**:
   ```bash
   git clone <repository-url>
   cd azure-tui
   go build -o aztui cmd/main.go
   ./aztui
   ```

### First Launch Experience

When you first launch Azure TUI, you'll see:

```
Starting Azure TUI...
Creating initial model...
Creating tea program...
Starting program...
```

The application will:
- ✅ Load demo data immediately (no hanging)
- 🔄 Connect to Azure CLI in background 
- 📊 Display real resource groups within 2-5 seconds
- 💡 Fall back to demo data if Azure is unavailable

---

## Real Azure Integration

### Authentication Status

Azure TUI automatically detects your Azure CLI authentication:

**✅ Authenticated State**:
```
☁️ Development   🏢 Demo Organization   📁 4 groups
```

**❌ Not Authenticated**:
```
☁️ Demo Mode   🏢 Demo Organization   📁 Loading groups...
```

### Real Data Loading

The application loads your actual Azure resources:

1. **Subscriptions**: Shows your real Azure subscriptions
2. **Resource Groups**: Displays actual resource groups from your subscription
3. **Resources**: Loads real resources when expanding groups
4. **Resource Details**: Shows actual resource properties and metadata

**Example Real Resource Groups Loaded**:
- `NetworkWatcherRG`
- `rg-fcaks-identity`
- `rg-fcaks-tfstate`
- `dem01_group`

---

## Tree View Navigation

### Basic Navigation

```
┌─ Azure Resources ───────────────┐ ┌─ Resource Details ──────────────────┐
│ ☁️  Azure Resources             │ │ Welcome to Azure TUI                │
│ ├─ 🗂️  NetworkWatcherRG         │ │                                     │
│ ├─ 🗂️  rg-fcaks-identity        │ │ TREE VIEW INTERFACE                 │
│ ├─ 🗂️  rg-fcaks-tfstate         │ │                                     │
│ └─ 🗂️  dem01_group              │ │ Navigate with:                      │
│    ├─ 🖥️  dem01                 │ │ • j/k or ↑↓ - Navigate tree        │
│    ├─ 💾 dem01groupdiag         │ │ • Space - Expand/collapse           │
│    ├─ 🌐 dem01-vnet             │ │ • Enter - Open resource             │
│    └─ 🔒 dem01-nsg              │ │ • ? - Show all shortcuts            │
└─────────────────────────────────┘ └─────────────────────────────────────┘
```

### Navigation Keys

| Key | Action | Description |
|-----|--------|-------------|
| `j` or `↓` | Move Down | Navigate down in tree |
| `k` or `↑` | Move Up | Navigate up in tree |
| `Space` | Expand/Collapse | Toggle resource group expansion |
| `Enter` | Open Resource | Open resource in content tab |

### Tree Expansion Behavior

**Expanding Resource Groups**:
1. Press `Space` on a resource group
2. Azure TUI loads real resources from that group
3. Shows loading indicator during fetch
4. Displays actual resources with proper icons

**Example Expansion**:
```bash
# Before expansion:
├─ 🗂️  dem01_group

# After pressing Space:
├─ 🗂️  dem01_group
│  ├─ 🖥️  dem01 (Virtual Machine)
│  ├─ 💾 dem01groupdiag (Storage Account)
│  ├─ 🌐 dem01-vnet (Virtual Network)
│  ├─ 🔒 dem01-nsg (Network Security Group)
│  ├─ 🌐 dem01-ip (Public IP Address)
│  └─ 🔗 dem01211_z1 (Network Interface)
```

---

## Resource Management

### Opening Resources

**Using Enter Key**:
1. Navigate to any resource in the tree
2. Press `Enter` to open in content tab
3. View detailed resource information
4. Access resource-specific actions

**Content Tab Features**:
- Resource properties and metadata
- Real-time status information
- Available management actions
- Resource-specific icons and styling

### Resource Actions

Press specific keys while a resource is selected:

| Key | Action | Description |
|-----|--------|-------------|
| `a` | AI Analysis | Get AI-powered insights |
| `M` | Metrics | View performance dashboard |
| `E` | Edit | Resource configuration editor |
| `T` | Terraform | Generate Terraform code |
| `B` | Bicep | Generate Bicep template |
| `O` | Optimize | Cost optimization analysis |
| `Ctrl+D` | Delete | Safe deletion with confirmation |

---

## AI-Powered Features

### AI Analysis (`a` key)

Get comprehensive AI insights for any resource:

```
🤖 AI Analysis: dem01 (Virtual Machine)

Configuration Summary:
• Size: Standard_B1s (1 vCPU, 1GB RAM)
• OS: Ubuntu 20.04 LTS
• Disk: Premium SSD (30GB)
• Network: Single NIC in dem01-vnet

Optimization Recommendations:
• Consider B2s for better performance if needed
• Enable automated backup for data protection  
• Review NSG rules for security compliance

Security Considerations:
• SSH access detected - ensure key management
• Consider enabling disk encryption
• Review network security group rules
```

### Cost Optimization (`O` key)

Get AI-driven cost savings analysis:

```
💰 Cost Optimization: dem01_group

Current Monthly Cost: ~$45

Savings Opportunities:
1. VM Right-sizing (Save $15/month):
   • dem01: Consider B1s instead of B2s
   • Low CPU utilization detected

2. Storage Optimization (Save $8/month):
   • Convert to Standard SSD for non-critical workloads
   
3. Reserved Instances (Save $12/month):
   • 1-year reservation available
```

---

## Troubleshooting Guide

### Common Issues

#### 1. Application Hanging on Startup

**Symptoms**: Application shows "Starting program..." and hangs

**Solution**: 
- ✅ **Fixed in current version**: App now starts immediately with demo data
- Real Azure data loads in background (2-5 seconds)
- No more hanging issues

#### 2. No Real Azure Data

**Symptoms**: Only shows demo data, no real resource groups

**Possible Causes**:
```bash
# Check Azure CLI authentication
az account show

# If not logged in:
az login

# Verify subscription access:
az account list --output table
```

**Debug Steps**:
1. Ensure Azure CLI is installed: `az --version`
2. Check authentication: `az account show`
3. Verify subscription access: `az group list --output table`
4. Check network connectivity to Azure

#### 3. Tree View Not Expanding

**Symptoms**: Pressing Space doesn't expand resource groups

**Solution**:
1. Ensure you're in tree view mode (default)
2. Select a resource group first (not a resource)
3. Press `Space` key (not Enter)
4. Wait for loading indicator

#### 4. Resource Actions Not Working

**Symptoms**: Pressing action keys (`a`, `M`, `T`) doesn't respond

**Requirements**:
- Resource must be selected
- For AI features: Set `OPENAI_API_KEY` environment variable
- For some actions: Proper Azure permissions required

### Performance Optimization

#### Large Resource Sets

If you have many resource groups or resources:

1. **Scrolling**: Use `j/k` keys for smooth navigation
2. **Search**: Use `/` to search (if implemented)
3. **Filtering**: Focus on specific resource groups
4. **Timeouts**: Default 5-second timeout prevents hanging

#### Memory Usage

For optimal performance:
- Close unused content tabs with `Ctrl+W`
- Use tree view mode (more efficient than traditional)
- Restart application if it becomes sluggish

---

## Best Practices

### Daily Workflow

#### 1. Resource Health Check
```bash
# Daily routine:
1. Launch Azure TUI
2. Navigate through critical resource groups
3. Check for any status issues
4. Use AI analysis (`a`) on key resources
```

#### 2. Cost Management
```bash
# Weekly cost review:
1. Use cost optimization (`O`) on expensive resource groups
2. Review AI recommendations
3. Implement suggested optimizations
4. Generate reports for stakeholders
```

#### 3. Infrastructure Documentation
```bash
# Monthly documentation:
1. Select key resources
2. Generate Terraform code (`T`)
3. Export Bicep templates (`B`)
4. Maintain infrastructure as code
```

### Security Best Practices

#### 1. Resource Access
- Regularly review AI security recommendations
- Use Azure TUI for security posture assessment
- Implement suggested security improvements

#### 2. Authentication
- Keep Azure CLI authentication current
- Use appropriate Azure RBAC permissions
- Avoid overprivileged access

### Team Collaboration

#### 1. Shared Configuration
Create team-wide Azure TUI config:

```yaml
# ~/.config/azure-tui/config.yaml
naming:
  standard: "team-{{type}}-{{env}}-{{name}}"
  environment: "prod"

ai:
  provider: "openai"
  model: "gpt-4"
  enabled: true

features:
  metrics_dashboard: true
  ai_analysis: true
  iac_generation: true
  cost_optimization: true
```

#### 2. Best Practices Documentation
- Document common workflows
- Share AI-generated insights
- Maintain team resource naming conventions
- Use generated IaC templates consistently

---

## Advanced Usage

### Keyboard Shortcuts Reference

| Category | Key | Action | Description |
|----------|-----|--------|-------------|
| **Navigation** | `j/k` or `↑/↓` | Move | Navigate tree/list |
| | `Space` | Expand | Toggle tree node |
| | `Enter` | Open | Open in content tab |
| **Tabs** | `Tab` | Next Tab | Switch to next tab |
| | `Shift+Tab` | Previous Tab | Switch to previous tab |
| | `Ctrl+W` | Close Tab | Close current tab |
| **Actions** | `a` | AI Analysis | Get AI insights |
| | `M` | Metrics | Performance dashboard |
| | `E` | Edit | Configuration editor |
| | `T` | Terraform | Generate code |
| | `B` | Bicep | Generate template |
| | `O` | Optimize | Cost analysis |
| | `Ctrl+D` | Delete | Safe deletion |
| **Interface** | `F2` | Toggle Mode | Tree/traditional view |
| | `?` | Help | Show shortcuts |
| | `Esc` | Close | Close dialogs |
| | `q` | Quit | Exit application |

### Configuration Options

Create `~/.config/azure-tui/config.yaml` for customization:

```yaml
# Display preferences
interface:
  theme: "azure"
  tree_view: true
  auto_expand: false

# AI integration
ai:
  provider: "openai"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4"
  timeout: 30

# Azure settings
azure:
  default_subscription: "Development"
  timeout: 10
  cache_ttl: 300

# Feature toggles
features:
  ai_analysis: true
  metrics_dashboard: true
  cost_optimization: true
  iac_generation: true
```

---

## Summary

Azure TUI provides a powerful, keyboard-driven interface for Azure resource management. With real Azure integration, AI-powered insights, and comprehensive resource management capabilities, it transforms the Azure experience from web-based to terminal-based efficiency.

**Key Benefits**:
- ⚡ **Fast**: Instant startup, real data in background
- 🎯 **Efficient**: Keyboard-driven workflow
- 🤖 **Intelligent**: AI-powered insights and recommendations
- 🔧 **Comprehensive**: Full resource management capabilities
- 💰 **Cost-Aware**: Built-in optimization recommendations

The application is production-ready and provides significant productivity improvements for DevOps professionals, developers, and Azure administrators.
