# Azure TUI - Complete User Guide 📚

## Table of Contents
1. [Getting Started](#getting-started)
2. [Interface Overview](#interface-overview)
3. [Navigation Guide](#navigation-guide)
4. [Feature Walkthroughs](#feature-walkthroughs)
5. [Keyboard Shortcuts](#keyboard-shortcuts)
6. [Troubleshooting](#troubleshooting)

---

## Getting Started

### Prerequisites
- **Azure CLI**: Installed and authenticated (`az login`)
- **Go 1.21+**: For building from source
- **Terminal**: Modern terminal with Unicode support

### Installation
```bash
# Clone the repository
git clone https://github.com/olafkfreund/azure-tui.git
cd azure-tui

# Build the application
go build -o aztui cmd/main.go

# Run the application
./aztui
```

### First Launch
1. **Demo Mode**: If Azure CLI is not configured, the app launches with demo data
2. **Azure Mode**: If authenticated, real Azure resources load automatically
3. **Loading Time**: Initial load may take 5-8 seconds for real Azure data

---

## Interface Overview

### Main Components

#### 🌳 Tree View (Left Panel)
- **Hierarchical Display**: Subscriptions → Resource Groups → Resources
- **Visual Indicators**: 
  - `📁` Folders (expandable containers)
  - `📊` Dashboards and views
  - `🔧` Individual resources
  - `✅⚠️❌❔` Health status indicators

#### 📑 Content Tabs (Right Panel)
- **Resource Details**: Selected resource information
- **AI Analysis**: GPT-powered insights and recommendations
- **Terraform/Bicep**: Generated Infrastructure as Code
- **Metrics**: Real-time performance data

#### 📊 Status Bar (Bottom)
- **Subscription Info**: Current Azure subscription and tenant
- **Resource Counts**: Total resources and health status
- **Last Updated**: Timestamp of last data refresh
- **Mode Indicators**: Auto-refresh status and monitoring info

---

## Navigation Guide

### Basic Navigation
```
j/k          ↑/↓ Move up/down in tree
h/l          ←/→ Collapse/expand or navigate tabs
Space        Toggle expand/collapse on folders
Enter        Open selected resource in new tab
Tab          Switch between tree view and content panels
```

### Tree View Operations
```
/            Search resources (fuzzy search)
n            Next search result
N            Previous search result
gg           Go to top
G            Go to bottom
```

### Tab Management
```
Ctrl+T       Open new tab
Ctrl+W       Close current tab
1-9          Switch to tab by number
Ctrl+←/→     Navigate between tabs
```

---

## Feature Walkthroughs

### 1. Resource Analysis with AI
1. **Select Resource**: Navigate to any Azure resource
2. **Trigger Analysis**: Press `a` key
3. **Review Insights**: AI provides:
   - Resource health assessment
   - Cost optimization suggestions
   - Security recommendations
   - Best practices compliance

**Example Output**:
```
🤖 AI Analysis for VM 'webapp-vm-01'

💡 Key Insights:
• Resource is healthy but underutilized (12% avg CPU)
• Consider downsizing from Standard_D4s_v3 to Standard_D2s_v3
• Estimated savings: $89/month (31% reduction)

🔐 Security Recommendations:
• Enable Azure Disk Encryption
• Configure Network Security Group rules
• Update VM agent and extensions

📊 Performance Notes:
• Memory utilization: 45% average
• Network throughput: Low (< 1 Mbps)
• Disk IOPS: Within normal range
```

### 2. Infrastructure as Code Generation
#### Terraform Generation (`T` key)
1. **Select Resource**: Choose any Azure resource
2. **Generate Code**: Press `T`
3. **Review Output**: Complete Terraform configuration with:
   - Resource definition
   - Required providers
   - Variable declarations
   - Output values

#### Bicep Generation (`B` key)
1. **Select Resource**: Choose any Azure resource  
2. **Generate Template**: Press `B`
3. **Review ARM Template**: Azure Resource Manager Bicep template

### 3. Resource Health Monitoring
The application includes real-time health monitoring:

- **Auto-refresh**: Every 30 seconds (configurable)
- **Manual Refresh**: Press `Ctrl+R`
- **Toggle Auto-refresh**: Press `h`
- **Health Indicators**:
  - `✅` Healthy
  - `⚠️` Warning
  - `❌` Critical
  - `❔` Unknown

### 4. Metrics Dashboard
1. **Open Dashboard**: Press `M` on any resource
2. **View Metrics**: Real-time performance data
3. **ASCII Graphs**: Visual trend representation
4. **Key Metrics**:
   - CPU utilization
   - Memory usage
   - Network throughput
   - Storage IOPS

---

## Keyboard Shortcuts

### Core Navigation
| Key | Action |
|-----|--------|
| `j/k` | Move up/down |
| `h/l` | Collapse/expand or navigate |
| `Space` | Toggle expand/collapse |
| `Enter` | Open resource |
| `Tab` | Switch panels |
| `F2` | Toggle interface mode |

### Resource Operations
| Key | Action |
|-----|--------|
| `a` | AI Analysis |
| `T` | Generate Terraform |
| `B` | Generate Bicep |
| `M` | Metrics Dashboard |
| `E` | Edit Resource |
| `O` | Cost Optimization |
| `Ctrl+D` | Delete Resource |

### View Controls
| Key | Action |
|-----|--------|
| `r` | Refresh Resource Groups |
| `Ctrl+R` | Refresh Health Status |
| `h` | Toggle Auto-refresh |
| `/` | Search |
| `?` | Show Help |
| `q` | Quit Application |

### Tab Management
| Key | Action |
|-----|--------|
| `Ctrl+T` | New Tab |
| `Ctrl+W` | Close Tab |
| `1-9` | Switch to Tab |
| `Ctrl+←/→` | Navigate Tabs |

---

## Advanced Features

### Configuration File
Create `~/.config/aztui/config.yaml`:
```yaml
# Azure TUI Configuration
ui:
  theme: "azure"
  auto_refresh: true
  refresh_interval: 30s
  
ai:
  provider: "openai"
  api_key: "your-api-key"
  model: "gpt-4"
  
azure:
  default_subscription: "your-subscription-id"
  timeout: 30s
```

### Environment Variables
```bash
export AZURE_TUI_AI_API_KEY="your-openai-key"
export AZURE_TUI_THEME="dark"
export AZURE_TUI_DEBUG="true"
```

### Custom Themes
The application supports custom color themes via lipgloss styling.

---

## Real-World Scenarios

### Scenario 1: Cost Optimization Review
1. Navigate to expensive resource groups
2. Use AI analysis (`a`) on high-cost resources
3. Review optimization suggestions
4. Generate Terraform (`T`) for rightsized resources
5. Plan migration strategy

### Scenario 2: Security Audit
1. Browse all resource groups
2. Check health indicators for security alerts
3. Use AI analysis for security recommendations
4. Review compliance status across resources

### Scenario 3: Infrastructure Documentation
1. Select critical resources
2. Generate Terraform/Bicep templates
3. Use AI analysis for documentation
4. Export configurations for version control

---

## Performance Tips

### Large Environments
- Use search (`/`) to quickly find resources
- Enable auto-refresh for active monitoring
- Use health indicators to spot issues quickly

### Slow Connections
- Disable auto-refresh if connection is slow
- Use demo mode for testing interfaces
- Increase timeout values in configuration

---

## Integration Guide

### CI/CD Integration
The application can be integrated into CI/CD pipelines:
```bash
# Generate Terraform for deployment
aztui --resource-id="/subscriptions/.../resourceGroups/prod" --output-terraform > prod.tf

# Check resource health
aztui --health-check --subscription="prod-sub" --format=json
```

### Automation Scripts
```bash
#!/bin/bash
# Daily cost optimization report
aztui --cost-analysis --subscription="$AZURE_SUBSCRIPTION" | \
  mail -s "Daily Azure Cost Report" admin@company.com
```

---

This user guide provides comprehensive coverage of all Azure TUI features with practical examples and real-world scenarios. For additional help, press `?` within the application or check the troubleshooting section.
