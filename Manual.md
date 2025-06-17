# Azure TUI Manual 📚

## Table of Contents

1. [Getting Started](#getting-started)
2. [Interface Overview](#interface-overview)
3. [Real-World Examples](#real-world-examples)
4. [Advanced Features](#advanced-features)
5. [Troubleshooting](#troubleshooting)
6. [Best Practices](#best-practices)

---

## Getting Started

### Installation & Setup

#### Prerequisites

```bash
# Install Go 1.21+
curl -LO https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Azure CLI (optional for demo mode)
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
az login
```

#### Building the Application

```bash
git clone https://github.com/olafkfreund/azure-tui
cd azure-tui
go build -o aztui ./cmd
```

#### First Run

```bash
# With Azure credentials
./aztui

# Demo mode (no Azure setup required)
DEMO_MODE=true ./aztui
```

### Basic Navigation

**Tree View Navigation**:

- `j` or `↓` - Move down in tree
- `k` or `↑` - Move up in tree  
- `Space` - Expand/collapse resource groups
- `Enter` - Open resource in content tab

**Tab Management**:

- `Tab` - Switch to next content tab
- `Shift+Tab` - Switch to previous content tab
- `Ctrl+W` - Close current content tab

**General**:

- `F2` - Toggle between tree view and traditional mode
- `?` - Show keyboard shortcuts
- `Esc` - Close dialogs/popups
- `q` - Quit application

---

## Interface Overview

### Tree View Mode (Default)

```
┌─ Azure Resources ───────────────┐ ┌─ Resource Details ──────────────────┐
│ ☁️  Azure Resources             │ │ 🖥️ my-production-vm               │
│ ├─ 🗂️  prod-webapp-rg           │ │                                    │
│ │  ├─ 🌐 webapp-frontend        │ │ Name: my-production-vm             │
│ │  ├─ 🗄️  webapp-database       │ │ Type: Microsoft.Compute/VM         │
│ │  └─ 🔑 webapp-secrets         │ │ Location: West Europe              │
│ ├─ 🗂️  dev-environment-rg       │ │ Resource Group: prod-webapp-rg     │
│ │  ├─ 🖥️  dev-jumpbox           │ │ Status: Running                    │
│ │  └─ 🚢 dev-k8s-cluster        │ │                                    │
│ └─ 🗂️  monitoring-rg            │ │ Actions:                           │
│    ├─ 📊 central-logs           │ │ • Press 'a' for AI analysis        │
│    └─ 🚨 critical-alerts        │ │ • Press 'M' for metrics            │
└─────────────────────────────────┘ │ • Press 'E' to edit                │
                                     │ • Press 'T' for Terraform          │
                                     └─────────────────────────────────────┘
┌─ Status: ☁️ Production Subscription │ 🏢 Contoso Corp │ 📁 5 groups ──────┐
└─────────────────────────────────────────────────────────────────────────┘
```

### Traditional Mode (F2 to toggle)

```
┌─ Resource Groups ───────────────┐ ┌─ Resources in Group ───────────────┐
│ → 🗂️  prod-webapp-rg            │ │ → 🖥️  my-production-vm            │
│   🗂️  dev-environment-rg        │ │   🌐 webapp-frontend               │
│   🗂️  data-analytics-rg         │ │   🗄️  webapp-database              │
│   🗂️  monitoring-rg             │ │   🔑 webapp-secrets                │
│   🗂️  backup-storage-rg         │ │   💾 webappstorageacct             │
└─────────────────────────────────┘ └─────────────────────────────────────┘
```

---

## Real-World Examples

### Example 1: Daily Resource Discovery

**Scenario**: You're a DevOps engineer starting your day and need to check the status of your production environment.

```bash
# Launch Azure TUI
./aztui

# Navigation steps:
1. Use 'j/k' to navigate to your production resource group
2. Press 'Space' to expand the resource group
3. Navigate to a critical VM using 'j/k'
4. Press 'a' for AI analysis
```

**AI Analysis Output**:

```
🤖 AI Analysis: my-production-vm

Configuration Summary:
• VM Size: Standard_D4s_v3 (4 vCPUs, 16GB RAM)
• OS: Ubuntu 20.04 LTS
• Disk: Premium SSD, 128GB
• Network: Accelerated networking enabled
• Backup: Daily backup configured

Recommendations:
✅ VM is properly sized for current workload
⚠️  Consider upgrading to Ubuntu 22.04 LTS for latest security patches
💰 Potential 15% cost savings with Reserved Instance
🔒 Enable Azure Security Center standard tier for enhanced monitoring

Security Considerations:
• SSH access restricted to company IP ranges ✅
• OS updates are automatic ✅  
• Consider enabling disk encryption at rest
```

### Example 2: Infrastructure as Code Generation

**Scenario**: You need to replicate a production VM configuration for a new environment.

```bash
# Steps:
1. Navigate to the VM you want to replicate
2. Press 'Enter' to open it in a content tab
3. Press 'T' to generate Terraform code
```

**Generated Terraform Code**:

```hcl
# Generated by Azure TUI - my-production-vm
resource "azurerm_resource_group" "main" {
  name     = "prod-webapp-rg"
  location = "West Europe"
}

resource "azurerm_virtual_network" "main" {
  name                = "prod-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "main" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_linux_virtual_machine" "main" {
  name                = "my-production-vm"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_D4s_v3"
  
  disable_password_authentication = true
  
  admin_username = "adminuser"
  
  network_interface_ids = [
    azurerm_network_interface.main.id,
  ]
  
  admin_ssh_key {
    username   = "adminuser"
    public_key = file("~/.ssh/id_rsa.pub")
  }
  
  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }
  
  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-focal"
    sku       = "20_04-lts-gen2"
    version   = "latest"
  }
}
```

### Example 3: Cost Optimization Workflow

**Scenario**: Monthly cost review to identify savings opportunities.

```bash
# Steps:
1. Navigate to your most expensive resource group
2. Press 'O' for cost optimization analysis
```

**Cost Optimization Report**:

```
💰 Cost Optimization Analysis: prod-webapp-rg

Current Monthly Costs: $2,847

Savings Opportunities:

1. Right-Sizing Recommendations (Save $420/month):
   • my-production-vm: Downsize from D4s_v3 to D2s_v3
     - CPU utilization: 23% average
     - Memory usage: 45% average  
     - Estimated savings: $156/month

   • webapp-database: Reduce DTU from 200 to 100
     - Current utilization: 35% average
     - Estimated savings: $264/month

2. Reserved Instance Opportunities (Save $680/month):
   • 3x Standard_D2s_v3 VMs qualified for 1-year RI
   • Database qualified for 1-year Reserved Capacity
   
3. Storage Optimization (Save $95/month):
   • Convert infrequently accessed blobs to Cool tier
   • 847GB eligible for archive tier

4. Unused Resources (Save $234/month):
   • 2x orphaned disks (last attached 45 days ago)
   • 1x unused public IP address
   • 3x old snapshots (>90 days)

Total Potential Monthly Savings: $1,429 (50% reduction)
Annual Savings: $17,148
```

### Example 4: Multi-Resource Management

**Scenario**: You need to compare configurations across multiple VMs and apply consistent settings.

```bash
# Steps:
1. Navigate to first VM, press 'Enter' to open in tab
2. Navigate to second VM, press 'Enter' to open in new tab
3. Use 'Tab' to switch between resource tabs
4. Press 'E' on each to compare configurations
5. Apply consistent settings across VMs
```

**Configuration Comparison**:

```
Tab 1: prod-web-vm-01          Tab 2: prod-web-vm-02
┌─ Configuration ──────────┐    ┌─ Configuration ──────────┐
│ Size: Standard_D2s_v3    │    │ Size: Standard_D4s_v3    │ ⚠️
│ OS: Ubuntu 20.04         │    │ OS: Ubuntu 22.04         │ ⚠️
│ Backup: Enabled          │    │ Backup: Disabled         │ ⚠️
│ Monitoring: Enabled      │    │ Monitoring: Enabled      │ ✅
│ Auto-shutdown: 19:00     │    │ Auto-shutdown: None      │ ⚠️
└──────────────────────────┘    └──────────────────────────┘

Issues Detected:
• Inconsistent VM sizes (may indicate configuration drift)
• Different OS versions (security risk)
• Backup not configured on VM-02 (compliance issue)
• Auto-shutdown missing on VM-02 (cost issue)
```

### Example 5: Security Assessment

**Scenario**: Security team needs to assess the security posture of Azure resources.

```bash
# Steps:
1. Navigate to a resource group containing sensitive resources
2. Press 'a' on each critical resource for security analysis
3. Generate comprehensive security report
```

**Security Analysis Example**:

```
🔒 Security Analysis: webapp-database (SQL Server)

Security Score: 7.2/10 (Good)

✅ Strengths:
• TLS 1.2 enforced for all connections
• Firewall rules restrict access to known IPs
• Transparent Data Encryption (TDE) enabled
• Automatic backups with point-in-time restore
• SQL Threat Detection enabled

⚠️ Recommendations:
• Enable Azure AD authentication (currently using SQL auth)
• Configure Advanced Data Security for vulnerability assessment
• Implement dynamic data masking for sensitive columns
• Set up audit logging to storage account

🚨 Critical Issues:
• Allow Azure services firewall rule is too permissive
• Some user accounts have excessive privileges
• Password policy not enforced (recommend Azure AD integration)

Compliance Status:
✅ GDPR: Data encryption and backup retention compliant
⚠️ SOC 2: Audit logging needs enhancement
⚠️ PCI DSS: Additional controls needed for payment data
```

### Example 6: Real-Time Monitoring

**Scenario**: Monitor resource performance during peak traffic hours.

```bash
# Steps:
1. Navigate to your critical application VM
2. Press 'M' to open metrics dashboard
3. Monitor real-time performance metrics
```

**Metrics Dashboard**:

```
📊 Metrics Dashboard: webapp-frontend-vm
Refresh: Every 30s | Time Range: Last 1 hour

CPU Usage                     Memory Usage
████████████░░░░░░░░░  65%    ███████████████░░░░░  76%
Avg: 58% | Peak: 89%          Avg: 72% | Peak: 91%

Network I/O                   Disk I/O  
In:  ████████░░░░░░░░░  42Mbps  Read:  ███░░░░░░░░░░░░░  156 IOPS
Out: ██████░░░░░░░░░░░  28Mbps  Write: █████░░░░░░░░░░  287 IOPS

Alerts:
🟡 High memory usage detected (>75% for 10 minutes)
🟢 CPU usage within normal range
🟢 Disk performance healthy
🟡 Network traffic above baseline (expected during peak hours)

Recommendations:
• Consider scaling up memory if high usage persists
• Monitor for memory leaks in application
• Current load suggests good utilization of resources
```

---

## Advanced Features

### Configuration Management

Create `~/.config/azure-tui/config.yaml`:

```yaml
# Azure TUI Configuration
azure:
  default_subscription: "prod-subscription-id"
  resource_groups:
    favorites:
      - "prod-webapp-rg"
      - "shared-services-rg"
  
ai:
  provider: "openai"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4"
  temperature: 0.3
  
interface:
  default_mode: "tree"  # or "traditional"
  theme: "azure"
  show_icons: true
  auto_refresh: 300  # seconds
  
shortcuts:
  custom:
    "ctrl+r": "refresh_all"
    "ctrl+f": "find_resource"
    
notifications:
  cost_alerts: true
  security_warnings: true
  performance_issues: true
```

### AI Prompts Customization

```yaml
ai:
  prompts:
    cost_analysis: |
      Analyze the Azure resources and provide specific cost optimization 
      recommendations with estimated savings amounts. Focus on:
      - Right-sizing opportunities
      - Reserved instance benefits  
      - Storage tier optimization
      - Unused resource cleanup
      
    security_analysis: |
      Perform a comprehensive security assessment of this Azure resource.
      Check for compliance with industry standards (SOC 2, PCI DSS, GDPR).
      Provide actionable recommendations with risk ratings.
```

### Keyboard Shortcuts Reference

| Category | Key | Action | Description |
|----------|-----|--------|-------------|
| **Navigation** | `j` or `↓` | Move Down | Navigate down in tree/list |
| | `k` or `↑` | Move Up | Navigate up in tree/list |
| | `Space` | Expand/Collapse | Toggle tree node |
| | `Enter` | Open Resource | Open in content tab |
| **Tabs** | `Tab` | Next Tab | Switch to next content tab |
| | `Shift+Tab` | Previous Tab | Switch to previous tab |
| | `Ctrl+W` | Close Tab | Close current content tab |
| **Actions** | `a` | AI Analysis | Get AI insights |
| | `M` | Metrics | Show performance dashboard |
| | `E` | Edit | Resource configuration editor |
| | `T` | Terraform | Generate Terraform code |
| | `B` | Bicep | Generate Bicep template |
| | `O` | Optimize | Cost optimization analysis |
| | `Ctrl+D` | Delete | Safe resource deletion |
| **Interface** | `F2` | Toggle Mode | Switch tree/traditional view |
| | `?` | Help | Show shortcuts |
| | `Esc` | Close | Close dialogs/popups |
| | `q` | Quit | Exit application |

---

## Troubleshooting

### Common Issues

#### "Failed to load Azure resources"

```bash
# Check Azure CLI authentication
az account show

# If not logged in:
az login

# Verify subscription access:
az account list --output table

# Set specific subscription:
az account set --subscription "your-subscription-id"
```

#### "OpenAI API key not configured"

```bash
# Set OpenAI API key
export OPENAI_API_KEY="your-api-key"

# Or add to your shell profile
echo 'export OPENAI_API_KEY="your-api-key"' >> ~/.bashrc
source ~/.bashrc
```

#### "Tree view not displaying properly"

```bash
# Terminal compatibility issue - try different terminal
# Ensure Unicode support is enabled
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# For tmux users:
echo 'set -g utf8 on' >> ~/.tmux.conf
```

#### "Slow performance with large subscriptions"

```bash
# Use filtering to reduce load
export AZURE_RESOURCE_GROUP_FILTER="prod-*"

# Or run in focused mode
./aztui --resource-group "specific-rg-name"
```

### Debug Mode

```bash
# Enable debug logging
DEBUG=true ./aztui

# Save debug output
DEBUG=true ./aztui 2> debug.log

# Verbose Azure CLI calls
AZ_DEBUG=true ./aztui
```

---

## Best Practices

### Daily Workflow

1. **Morning Resource Check**:
   - Launch Azure TUI
   - Navigate to critical resource groups
   - Use AI analysis (`a`) on key resources
   - Check for any alerts or issues

2. **Cost Management**:
   - Weekly cost optimization analysis (`O`)
   - Review and act on right-sizing recommendations
   - Monitor for unused resources

3. **Security Reviews**:
   - Monthly security analysis on all resource groups
   - Address critical security recommendations
   - Verify compliance status

### Resource Management

1. **Infrastructure as Code**:
   - Use Terraform generation (`T`) for resource standardization
   - Maintain consistent configurations across environments
   - Version control generated templates

2. **Monitoring Setup**:
   - Enable metrics dashboard (`M`) for critical resources
   - Set up automated monitoring for key performance indicators
   - Regular performance baseline reviews

### Team Collaboration

1. **Shared Configurations**:
   - Use shared config files for team standards
   - Document custom shortcuts and workflows
   - Standardize AI prompts for consistent analysis

2. **Documentation**:
   - Export AI analysis reports for team reviews
   - Share optimization recommendations
   - Maintain infrastructure documentation from generated code

---

## Integration Examples

### CI/CD Pipeline Integration

```bash
#!/bin/bash
# Example: Cost monitoring in CI/CD

# Generate cost report
aztui --batch --cost-analysis --format json > cost-report.json

# Check for cost increases
if [ $(jq '.cost_increase_percentage' cost-report.json) -gt 20 ]; then
    echo "Warning: Cost increase >20% detected"
    exit 1
fi
```

### Monitoring Scripts

```bash
#!/bin/bash
# Example: Automated security check

# Run security analysis on all resource groups  
aztui --batch --security-scan --output security-report.txt

# Alert on critical issues
if grep -q "🚨 Critical" security-report.txt; then
    # Send alert to Slack/Teams
    curl -X POST -H 'Content-type: application/json' \
         --data '{"text":"Critical security issues detected in Azure resources"}' \
         $SLACK_WEBHOOK_URL
fi
```

This manual provides comprehensive real-world examples and practical guidance for using the Azure TUI effectively in production environments. The examples demonstrate how the tool transforms Azure resource management from a web-based experience to a powerful, keyboard-driven workflow that DevOps professionals will find intuitive and efficient.
