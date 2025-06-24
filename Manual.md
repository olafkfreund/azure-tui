# Azure TUI Manual 📚

## Table of Contents

1. [Getting Started](#getting-started)
2. [Interface Overview](#interface-overview)  
3. [AI Configuration](#ai-configuration)
4. [Storage Management](#storage-management)
5. [Real-World Examples](#real-world-examples)
6. [Advanced Features](#advanced-features)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)
9. [Integration Examples](#integration-examples)

**New in this version**: Enhanced navigation system with panel switching, property expansion, AI configuration control, and comprehensive storage management.

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

**Panel Navigation** *(NEW)*:

- `h` or `←` - Move to left panel (Tree View)
- `l` or `→` - Move to right panel (Details View)
- `Tab` - Cycle between panels (Tree → Details → Tree)

**Property Management** *(NEW)*:

- `e` - Expand/collapse complex properties (AKS Agent Pools, etc.)
- Context-sensitive scrolling in Details panel

**Tab Management**:

- `Tab` - Switch to next content tab (when in tab content)
- `Shift+Tab` - Switch to previous content tab
- `Ctrl+W` - Close current content tab

**General**:

- `F2` - Toggle between tree view and traditional mode
- `?` - Show keyboard shortcuts
- `Esc` - Close dialogs/popups
- `q` - Quit application

---

## Interface Overview

### Tree View Mode (Default) - Enhanced Navigation

```
┌─ Azure Resources ───────────────┐ ┌─ Resource Details ──────────────────┐
│ ☁️  Azure Resources             │ │ 🖥️ my-production-vm               │
│ ├─ 🗂️  prod-webapp-rg  [ACTIVE]│ │                                    │
│ │  ├─ 🌐 webapp-frontend        │ │ Name: my-production-vm             │
│ │  ├─ 🗄️  webapp-database       │ │ Type: Microsoft.Compute/VM         │
│ │  └─ 🔑 webapp-secrets         │ │ Location: West Europe              │
│ ├─ 🗂️  dev-environment-rg       │ │ Resource Group: prod-webapp-rg     │
│ │  ├─ 🖥️  dev-jumpbox           │ │ Status: Running                    │
│ │  └─ 🚢 dev-k8s-cluster        │ │                                    │
│ └─ 🗂️  monitoring-rg            │ │ Agent Pool Profiles: 2 Agent Pool(s) [e to expand]
│    ├─ 📊 central-logs           │ │                                    │
│    └─ 🚨 critical-alerts        │ │ Actions:                           │
└─[🔍 Tree] ─────────────────────┘ │ • Press 'a' for AI analysis        │
                                     │ • Press 'M' for metrics            │
                                     │ • Press 'E' to edit                │
                                     │ • Press 'T' for Terraform          │
                                     └─[📊 Details] ─────────────────────┘
┌─ Status: Tree View │ h/l or ←/→ to switch panels │ e to expand properties ─┐
└─────────────────────────────────────────────────────────────────────────────┘
```

**Panel Navigation Features**:
- **Blue Border**: Active Tree panel - use j/k to navigate resources
- **Green Border**: Active Details panel - use j/k to scroll content
- **Property Expansion**: Press `e` to expand complex AKS properties
- **Visual Indicators**: [ACTIVE] markers and colored borders show current focus

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

## AI Configuration

### AI Analysis Behavior

Azure TUI now provides **manual AI analysis by default** for better control and reduced API usage. This represents a significant change from previous versions.

#### Default Behavior (Manual AI)

By default, AI analysis requires **manual trigger**:

- **Trigger**: Press `a` key when any resource is selected
- **Scope**: Analyzes the currently selected resource only
- **Benefits**: 
  - Reduces unnecessary API calls
  - Gives users full control over when AI analysis occurs
  - Minimizes costs for OpenAI API usage
  - Prevents automatic analysis on resource navigation

#### Automatic AI Analysis (Optional)

To enable automatic AI analysis (previous behavior), set the environment variable:

```bash
export AZURE_TUI_AUTO_AI="true"
```

With automatic mode enabled:
- AI analysis triggers automatically when resources are selected
- Analysis occurs in the background during navigation
- Higher API usage and potential costs
- Continuous insights without manual intervention

### Environment Variables for AI

| Variable | Default | Description |
|----------|---------|-------------|
| `AZURE_TUI_AUTO_AI` | `false` | Enable automatic AI analysis on resource selection |
| `OPENAI_API_KEY` | - | OpenAI API key for AI features |
| `OPENAI_MODEL` | `gpt-4` | AI model to use for analysis |
| `GITHUB_TOKEN` | - | GitHub token for Copilot integration |
| `USE_GITHUB_COPILOT` | auto-detect | Use GitHub Copilot instead of OpenAI |

### AI Configuration Examples

#### Basic Manual AI Setup
```bash
# Set OpenAI API key (required)
export OPENAI_API_KEY="sk-your-openai-api-key"

# Launch Azure TUI (manual AI mode by default)
./aztui

# Use: Navigate to any resource and press 'a' for AI analysis
```

#### Automatic AI Setup
```bash
# Enable automatic AI analysis
export AZURE_TUI_AUTO_AI="true"
export OPENAI_API_KEY="sk-your-openai-api-key"

# Launch Azure TUI
./aztui

# Use: AI analysis will trigger automatically when selecting resources
```

#### GitHub Copilot Setup
```bash
# Use GitHub Copilot for AI analysis
export GITHUB_TOKEN="your-github-token"
export USE_GITHUB_COPILOT="true"

# Optional: Enable automatic mode with Copilot
export AZURE_TUI_AUTO_AI="true"

./aztui
```

### AI Features Available

When AI is configured (either manual or automatic), you can access:

- **Resource Analysis** (`a`): Comprehensive resource insights and recommendations
- **Cost Optimization** (`O`): AI-driven cost savings analysis
- **Infrastructure as Code** (`T`/`B`): Generate Terraform and Bicep templates
- **Security Assessment**: Security posture evaluation and recommendations
- **Performance Insights**: Resource utilization and optimization suggestions

### Best Practices for AI Configuration

1. **Start with Manual Mode**: Use default manual AI mode to understand analysis patterns
2. **Monitor API Usage**: Track OpenAI API usage, especially with automatic mode
3. **Use Specific Analysis**: Press `a` only on resources you need detailed insights for
4. **Consider GitHub Copilot**: Often provides better Azure-specific recommendations
5. **Environment-Specific Setup**: Use automatic mode in development, manual in production

---

## Storage Management

Azure TUI provides comprehensive storage account management with intuitive keyboard shortcuts and real-time progress tracking.

### Storage Operations Overview

When a **Storage Account** is selected, Azure TUI provides dedicated storage management capabilities:

| Operation | Key | Description | Progress Tracking |
|-----------|-----|-------------|-------------------|
| List Containers | `T` | Show all blob containers | ✅ With progress bar |
| Create Container | `Shift+T` | Create new blob container | ✅ With feedback |
| List Blobs | `B` | Show blobs in container | ✅ With progress bar |
| Upload Blob | `U` | Upload file to container | ✅ With progress |
| Delete Items | `Ctrl+X` | Delete containers/blobs | ✅ With confirmation |

### Storage Workflow

#### 1. Container Management

**Navigation Flow**:
```
Storage Account → [T] → Container List → [B] → Blob List → [Enter] → Blob Details
```

**Container Operations**:
- **View Containers**: Press `T` to list all containers with metadata
- **Create Container**: Press `Shift+T` to create a new blob container
- **Container Details**: Shows last modified, public access, metadata, and lease status

**Example Container View**:
```
🗄️  Storage Containers in 'webappstorageacct'
═══════════════════════════════════════════════════════════════

📋 Container Inventory:
• web-assets (🟢 Available)
  Last Modified: 2024-01-15T10:30:00Z
  Public Access: blob

• backup-data (🔒 Leased)
  Last Modified: 2024-01-14T08:15:00Z
  Metadata: environment=production, backup=daily

Available Actions:
• Press 'B' to list blobs in a container
• Press 'Shift+T' to create a new container
• Press 'Ctrl+X' to delete a container
```

#### 2. Blob Management

**Blob Operations**:
- **View Blobs**: Press `B` from container list to show all blobs
- **Upload Blob**: Press `U` to upload files to the current container
- **Delete Blob**: Press `Ctrl+X` to delete selected blobs
- **Blob Details**: Shows size, content type, access tier, tags, and metadata

**Example Blob View**:
```
📁 Blobs in Container 'web-assets' (Account: webappstorageacct)
═══════════════════════════════════════════════════════════════

📋 Blob Inventory:
🧱 index.html (2.5 KB)
   Type: text/html
   Modified: 2024-01-15T10:30:00Z
   Access Tier: Hot

📄 styles.css (15.7 KB)
   Type: text/css
   Modified: 2024-01-15T09:45:00Z
   Access Tier: Hot

🖼️ logo.png (45.2 KB)
   Type: image/png
   Modified: 2024-01-14T16:20:00Z
   Access Tier: Hot

Available Actions:
• Press 'U' to upload a new blob
• Press 'Ctrl+X` to delete a blob
• Press 'Esc' to go back to containers
```

### Progress Tracking System

Azure TUI implements comprehensive progress tracking for all storage operations:

#### Loading Indicators
- **Container Loading**: Visual progress bar when fetching containers
- **Blob Loading**: Progress tracking during blob enumeration
- **Operation Status**: Real-time feedback for create/delete operations
- **Error Handling**: Clear error messages with troubleshooting guidance

#### Progress Flow
```
Operation Start → Progress Updates → Completion/Error → Result Display
      ↓               ↓                    ↓              ↓
   Loading...     [████████░░] 80%     Success!      Show Results
```

### Enhanced User Feedback

#### Empty State Handling

**No Containers Found**:
```
📭 No containers found in this storage account.

This might happen because:
   • Storage account is newly created
   • Containers were deleted or moved
   • Access permissions are insufficient
   • Check Azure portal for container visibility
   • Verify storage account permissions
   • Refresh the view with 'R'

Available Actions:
• Press 'Shift+T' to create a new container
• Press 'R' to refresh the container list
• Press 'Esc' to go back
```

**No Blobs Found**:
```
📭 No blobs found in container 'web-assets'.

This might happen because:
   • Container is empty or newly created
   • Blobs were deleted or moved to another container
   • Prefix/filter settings exclude visible blobs
   • Press 'U' to upload a blob to this container
   • Check other containers for your files
   • Verify blob naming and paths
   • Use Azure Storage Explorer for detailed view

Available Actions:
• Press 'U' to upload a blob
• Press 'R' to refresh the blob list
• Press 'Esc' to go back to containers
```

### Storage Management Best Practices

1. **Navigation**: Use `Esc` key to navigate back through storage views
2. **Refresh**: Press `R` to refresh container or blob lists when needed
3. **Progress**: Wait for progress completion before initiating new operations
4. **Error Handling**: Review error messages for troubleshooting guidance
5. **Permissions**: Ensure proper storage account access permissions

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

### Example 7: Enhanced Navigation and Property Management *(NEW)*

**Scenario**: You need to explore an AKS cluster with complex configurations and navigate efficiently between different resources.

```bash
# Navigation Steps:
1. Launch Azure TUI
2. Navigate to AKS resource group using 'j/k'
3. Press 'Space' to expand the resource group
4. Navigate to AKS cluster using 'j/k'
5. Press 'Enter' to open cluster details
6. Use new navigation features to explore properties
```

**Enhanced Navigation Features**:

```
Current Panel: Tree View [🔍 Blue Border]
├─ 🗂️  aks-production-rg  [ACTIVE]
│  ├─ 🚢 my-aks-cluster
│  ├─ 🔒 aks-identity
│  └─ 🌐 aks-vnet

Navigation Options:
• h/← - Stay in Tree panel (current)
• l/→ - Move to Details panel  
• Tab - Cycle between panels
• j/k - Navigate resources in tree
```

**Property Expansion Example**:

```
Details Panel: AKS Cluster [📊 Green Border]

Name: my-aks-cluster
Type: Microsoft.ContainerService/managedClusters
Location: West Europe
Resource Group: aks-production-rg
Status: Running

Agent Pool Profiles: 2 Agent Pool(s) [Press 'e' to expand]

Actions:
• Press 'e' to expand Agent Pools
• Press 'a' for AI analysis
• Press 'j/k' to scroll content
```

**After pressing 'e' to expand**:

```
Agent Pool Profiles: [Expanded]
  Pool 1: nodepool1
    ├─ VM Size: Standard_D4s_v3
    ├─ Count: 3 nodes
    ├─ OS Type: Linux
    ├─ Availability Zones: [1, 2, 3]
    └─ Auto Scaling: Enabled (min: 1, max: 10)
  
  Pool 2: userpool
    ├─ VM Size: Standard_D8s_v3  
    ├─ Count: 2 nodes
    ├─ OS Type: Linux
    ├─ Availability Zones: [1, 2]
    └─ Auto Scaling: Disabled

Network Profile:
  ├─ Network Plugin: kubenet
  ├─ Service CIDR: 10.0.0.0/16
  └─ DNS Service IP: 10.0.0.10

[Press 'e' again to collapse]
```

**Key Navigation Benefits**:
- **Immediate Visual Feedback**: Colored borders show active panel
- **Context-Sensitive Controls**: j/k behavior adapts to current panel
- **Property Management**: Complex properties become readable and navigable
- **Efficient Exploration**: Quickly switch between tree navigation and detail review

---

## Terraform Integration 🏗️

Azure TUI includes a comprehensive Terraform integration that allows you to manage Infrastructure as Code directly from the TUI interface. Access it with `Ctrl+T`.

### Key Features

- **Project Discovery**: Automatically finds Terraform projects in your workspace
- **Code Analysis**: Validates and analyzes Terraform configurations
- **Operations**: Execute terraform init, plan, apply, validate, format, destroy
- **Template Management**: Create new projects from predefined templates
- **Editor Integration**: Open projects in your preferred editor (VS Code, vim, nvim)

### Available Templates

Azure TUI includes production-ready Terraform templates for:

- **Linux VMs**: Complete virtual machine setup with networking, security groups, and SSH access
- **Azure Kubernetes Service (AKS)**: Managed Kubernetes clusters with monitoring and logging
- **Azure SQL Server**: SQL Server instances with databases, security, and Key Vault integration
- **Container Instances (ACI)**: Both single and multi-container deployments
- **Multi-Container Apps**: Complex containerized applications with load balancing

### Real-World Examples

#### Example 1: Analyzing an Existing Terraform Project

**Scenario**: You have a Terraform project for a web application infrastructure and want to validate it.

1. **Open Terraform Manager**: Press `Ctrl+T`
2. **Select "Analyze Code"**: Navigate with ↑/↓, press Enter
3. **Choose Project**: Select your web-app terraform folder
4. **View Analysis**: See validation results and file structure

**What you'll see**:
```
📁 Terraform Project Analysis: ./web-app-infrastructure

✅ main.tf found
✅ variables.tf found  
✅ outputs.tf found
❌ terraform.tf missing

🔍 Use Terraform operations to validate and manage this project.
```

#### Example 2: Creating a New AKS Cluster

**Scenario**: You need to deploy a new Kubernetes cluster for a staging environment.

1. **Open Terraform Manager**: Press `Ctrl+T`
2. **Create from Template**: Select "Create from Template"
3. **Choose AKS Template**: Select the basic-aks template
4. **Customize Variables**: Edit variables for your staging environment

**Template includes**:
- AKS cluster with system-assigned managed identity
- Default node pool with configurable VM sizes
- Azure Container Registry integration
- Log Analytics workspace for monitoring
- Network security groups and subnets

#### Example 3: Validating Multi-Container Application

**Scenario**: You're deploying a microservices application using Azure Container Instances.

1. **Navigate to Project**: Use file explorer or `Ctrl+T` → "Browse Folders"
2. **Select multi-container project**: Choose your microservices folder
3. **Run Validation**: Select "Terraform Operations" → Validate
4. **Check Results**: Review validation output for errors

**Multi-container template features**:
- Web frontend (nginx) on port 80
- API backend (custom app) on port 8080
- Health checks and readiness probes
- Environment variable configuration
- Log Analytics integration

#### Example 4: Infrastructure Code Review Workflow

**Scenario**: Team code review process for infrastructure changes.

1. **Analyze Code**: `Ctrl+T` → "Analyze Code" → Select project
2. **Validate Syntax**: `Ctrl+T` → "Terraform Operations" → Select project → Validate
3. **Format Code**: Run terraform format to ensure consistent styling
4. **Open in Editor**: `Ctrl+T` → "Open External Editor" → Make changes
5. **Re-validate**: Repeat validation after changes

### Terraform Operations Walkthrough

#### 1. Project Initialization
```bash
# What happens when you select "Terraform Operations" → Init
terraform init
# Downloads providers, initializes backend, prepares workspace
```

#### 2. Planning Changes
```bash
# What happens when you select "Terraform Operations" → Plan  
terraform plan
# Shows what resources will be created, modified, or destroyed
```

#### 3. Applying Infrastructure
```bash
# What happens when you select "Terraform Operations" → Apply
terraform apply -auto-approve
# Creates/updates infrastructure based on your configuration
```

#### 4. Validation and Formatting
```bash
# Validation
terraform validate
# Checks syntax and configuration validity

# Formatting  
terraform fmt
# Ensures consistent code formatting
```

### Template Structure Examples

#### Linux VM Template Structure
```
terraform/templates/vm/linux-vm/
├── main.tf           # VM, networking, security group
├── variables.tf      # Customizable parameters
├── outputs.tf        # IP addresses, SSH commands
└── install.sh        # Post-deployment scripts
```

**Key Resources**:
- Resource Group
- Virtual Network and Subnet
- Network Security Group (SSH + HTTP)
- Public IP Address
- Network Interface
- Linux Virtual Machine
- Custom Script Extension

#### AKS Template Structure  
```
terraform/templates/aks/basic-aks/
├── main.tf           # AKS cluster, node pools
├── variables.tf      # Cluster configuration
└── outputs.tf        # Kubeconfig, cluster info
```

**Key Resources**:
- Resource Group
- AKS Cluster with managed identity
- Default node pool (configurable size)
- Log Analytics workspace
- Container Registry (optional)

#### SQL Server Template Structure
```
terraform/templates/sql/sql-server/
├── main.tf           # SQL Server, database, security
├── variables.tf      # Server and DB configuration  
└── outputs.tf        # Connection strings, endpoints
```

**Key Resources**:
- Resource Group
- SQL Server with managed identity
- SQL Database with configurable tier
- Key Vault for secrets
- Firewall rules and virtual network rules

### Integration Workflow Examples

#### DevOps Pipeline Integration

**Scenario**: Using Azure TUI in a CI/CD pipeline for infrastructure validation.

1. **Pre-commit Hooks**:
   - Use `Ctrl+T` → "Terraform Operations" → Validate before commits
   - Format code with terraform fmt

2. **Pull Request Reviews**:
   - Analyze code structure with "Analyze Code"
   - Validate syntax and configuration

3. **Deployment Preparation**:
   - Use "Browse Folders" to review all terraform projects
   - Plan deployments with terraform plan

#### Development Workflow

**Scenario**: Daily development workflow for infrastructure changes.

**Morning Routine**:
1. Open Azure TUI: `./azure-tui`
2. Check infrastructure status in main interface
3. Review terraform projects: `Ctrl+T` → "Browse Folders"

**Making Changes**:
1. Open project in editor: `Ctrl+T` → "Open External Editor"
2. Make infrastructure changes in VS Code/vim
3. Return to Azure TUI
4. Validate changes: `Ctrl+T` → "Terraform Operations" → Validate
5. Plan deployment: Run terraform plan
6. Apply if satisfied: Run terraform apply

**Code Review**:
1. Analyze modified projects: `Ctrl+T` → "Analyze Code"
2. Check for best practices and missing files
3. Format code: Run terraform fmt
4. Final validation before commit

### Keyboard Shortcuts Reference

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl+T` | Open Terraform Manager | Main entry point for all Terraform operations |
| `↑/↓` | Navigate Menu | Move between options in Terraform popup |
| `Enter` | Select Option | Choose current menu item or folder |
| `Esc` | Go Back/Close | Return to previous menu or close popup |

### Tips and Best Practices

1. **Project Organization**: Keep related terraform files in dedicated folders
2. **Regular Validation**: Use `Ctrl+T` → "Terraform Operations" → Validate frequently
3. **Code Formatting**: Always format code before commits using terraform fmt
4. **Template Usage**: Start new projects from templates for best practices
5. **Analysis First**: Run "Analyze Code" on new projects to ensure completeness

### Error Handling

**Common Issues and Solutions**:

- **"No Terraform projects found"**: Ensure you have .tf files in your directories
- **Validation errors**: Use "Analyze Code" to identify missing files or syntax issues  
- **Editor not opening**: Terraform integration tries VS Code, then vim, nvim, nano in order
- **Permission errors**: Ensure terraform binary is installed and in PATH

### Integration with Azure Resources

The Terraform integration works seamlessly with the main Azure TUI interface:

1. **View Live Resources**: Use main interface to see current Azure resources
2. **Plan Infrastructure**: Use Terraform integration to plan changes
3. **Monitor Deployment**: Return to main interface to see deployment results
4. **Troubleshoot Issues**: Use both views for comprehensive infrastructure management

---

## Azure DevOps Integration 🔄

Azure TUI includes a comprehensive DevOps integration module for managing Azure DevOps organizations, projects, and pipelines through a popup-based interface.

### Key Features

- **Organization Management**: List and switch between Azure DevOps organizations
- **Project Navigation**: Browse projects within organizations  
- **Pipeline Discovery**: List all build and release pipelines
- **Pipeline Operations**: View pipeline details, runs, and history
- **Real-time Status**: Monitor pipeline execution status
- **Tree-based Interface**: Hierarchical view of DevOps resources

### Configuration

#### Environment Variables
```bash
# Required: Personal Access Token
export AZURE_DEVOPS_PAT="your-personal-access-token"

# Optional: Default organization and project
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_PROJECT="your-project"
```

#### Config File Setup
Add to `~/.config/azure-tui/config.yaml`:
```yaml
devops:
  organization: "your-organization"
  project: "your-project"  
  base_url: "https://dev.azure.com"
```

### Personal Access Token Setup

1. **Navigate to Azure DevOps**: Go to your organization settings
2. **Create Token**: User Settings → Personal Access Tokens → New Token
3. **Set Permissions**:
   - **Build**: Read & execute
   - **Release**: Read, write & execute
   - **Project and Team**: Read
   - **Identity**: Read
4. **Configure**: Copy token and set `AZURE_DEVOPS_PAT` environment variable

### Accessing DevOps Manager

The Azure DevOps integration is now fully integrated into the main TUI interface:

**Keyboard Access**: Press `Ctrl+O` from anywhere in the main interface to open the Azure DevOps Manager popup.

**Navigation**: 
- Use `↑/↓` or `j/k` to navigate through DevOps menu options
- Press `Enter` to select and execute operations
- Use `Esc` to go back to previous menu or close the popup
- Press `?` for help (includes DevOps shortcuts)

### DevOps Module Features

#### Organization & Project Management
- List available Azure DevOps organizations
- Browse projects within selected organization
- Switch between different organizations/projects
- Display project metadata and status

#### Pipeline Management
- Discover all build and release pipelines
- Filter pipelines by name, status, or recent activity
- Display pipeline information (name, repository, last run)
- Monitor pipeline execution status

#### Pipeline Operations
- View detailed pipeline configuration
- Check recent run history and results
- Monitor active pipeline executions
- Access pipeline logs and artifacts

### Usage Examples

#### Daily DevOps Workflow
1. **Morning Stand-up**: Check pipeline status across projects
2. **Build Monitoring**: Monitor active deployments and releases
3. **Project Review**: Review pipeline health across teams
4. **Release Management**: Track release pipeline execution

#### Pipeline Management
1. **Pipeline Discovery**: Find pipelines across multiple projects
2. **Status Monitoring**: Real-time pipeline execution tracking
3. **History Review**: Check recent runs and failure patterns
4. **Cross-Project View**: Monitor pipelines across organizations

### Integration Architecture

The DevOps integration follows the same pattern as Terraform integration:
- **Popup-based Interface**: Clean, borderless popup overlay
- **Hierarchical Navigation**: Tree-based organization → project → pipeline structure
- **Keyboard-driven**: Arrow keys for navigation, Enter to select, Esc to exit
- **Real-time Data**: Live status updates and pipeline information

### Future Enhancements

The DevOps integration is designed for future expansion:
- **Pipeline Triggering**: Start builds and releases from TUI
- **Approval Management**: Handle deployment approvals
- **Work Item Integration**: Link builds to work items and PRs
- **Dashboard Integration**: DevOps metrics in main TUI dashboard

**Note**: DevOps integration is now fully integrated with direct keyboard shortcut access (`Ctrl+O`) from the main TUI interface.

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
  
devops:
  organization: "your-organization"
  project: "your-project"
  base_url: "https://dev.azure.com"
  
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
| | `h` or `←` | Left Panel | Move to Tree View panel |
| | `l` or `→` | Right Panel | Move to Details View panel |
| | `Space` | Expand/Collapse | Toggle tree node |
| | `Enter` | Open Resource | Open in content tab |
| **Property Management** | `e` | Expand/Collapse | Toggle complex property expansion |
| **Tabs** | `Tab` | Panel/Tab Cycle | Switch panels or content tabs |
| | `Shift+Tab` | Previous Tab | Switch to previous tab |
| | `Ctrl+W` | Close Tab | Close current content tab |
| **Actions** | `a` | AI Analysis | Get AI insights (manual trigger by default) |
| | `Ctrl+T` | Terraform Manager | Open Terraform integration |
| | `Ctrl+O` | DevOps Manager | Open Azure DevOps integration |
| | `M` | Metrics | Show performance dashboard |
| | `E` | Edit | Resource configuration editor |
| | `T` | List Containers | List storage containers (Storage Accounts) |
| | `Shift+T` | Create Container | Create storage container (Storage Accounts) |
| | `B` | List Blobs | List blobs in container (Storage) / Generate Bicep template (other resources) |
| | `U` | Upload Blob | Upload blob to container (Storage Accounts) |
| | `O` | Optimize | Cost optimization analysis |
| | `Ctrl+X` | Delete Storage | Delete containers/blobs (Storage Accounts) |
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

#### "AI analysis not working / No AI response"

```bash
# Check if AI is in manual mode (default)
# Press 'a' key to manually trigger AI analysis

# For automatic AI analysis, set environment variable:
export AZURE_TUI_AUTO_AI="true"

# Verify API key is set correctly
echo $OPENAI_API_KEY

# Test API connectivity
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models
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
