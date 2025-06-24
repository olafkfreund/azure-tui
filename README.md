# Azure TUI (aztui) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Azure](https://img.shields.io/badge/Azure-CLI%20%2B%20SDK-0078D4?style=flat&logo=microsoft-azure)](https://azure.microsoft.com/)
[![AI Powered](https://img.shields.io/badge/AI-OpenAI%20GPT--4-412991?style=flat&logo=openai)](https://openai.com/)

A modern, **NeoVim-style Terminal User Interface** for managing Azure resources with AI-powered analysis, Infrastructure as Code generation, and comprehensive resource management capabilities.

## ✨ Features

### ✨ **Enhanced Features (Latest)**

- **📊 Table-Formatted Properties**: Resource properties displayed in organized tables with intelligent formatting
- **🔐 Enhanced SSH for VMs**: Direct SSH (`c`) and Bastion (`b`) connections with automatic IP detection
- **🚢 Comprehensive AKS Management**: Full kubectl integration with pod (`p`), deployment (`D`), node (`n`), and service (`v`) management
- **🔄 Azure DevOps Integration**: Complete pipeline management module with build/release monitoring, run triggering, and organization navigation (configuration available)
- **⚡ Real-time Actions**: Start (`s`), stop (`S`), restart (`r`) operations with visual feedback
- **🎮 Intuitive Controls**: Enhanced keyboard shortcuts and visual indicators for all operations

### 🎯 **NeoVim-Style Interface**
- **Tree View Navigation**: Hierarchical display of Azure resources with expand/collapse
- **Powerline Statusbar**: Modern statusbar with subscription, tenant, and resource context
- **Content Tabs**: Right-side tabbed interface for opened resources
- **Vim-like Navigation**: `j/k` keys, space for expand/collapse, familiar shortcuts
- **Mouse Support**: Scroll wheel navigation and click interactions

### 🤖 **AI-Powered Resource Management**
- **Manual AI Analysis** (`a`): AI-powered resource insights and recommendations (manual trigger by default)
- **Automatic AI Mode**: Set `AZURE_TUI_AUTO_AI="true"` to enable automatic analysis on resource selection
- **Code Generation**: Generate Terraform (`T`) and Bicep (`B`) templates
- **Cost Optimization** (`O`): AI-driven cost savings suggestions
- **Security Analysis**: Automated security posture assessment

### 📊 **Interactive Dashboards**
- **Real-time Metrics** (`M`): CPU, memory, network, and disk usage with trends
- **Resource Editor** (`E`): Safe configuration editing with validation
- **Delete Protection** (`Ctrl+D`): Confirmation dialogs prevent accidental deletions

### 🔧 **Infrastructure as Code**
- **Enhanced Terraform Integration**: Complete Terraform management suite
  - **Visual State Management**: Interactive browse and inspect Terraform state resources
  - **Interactive Plan Visualization**: View plan changes with smart filtering (all/create/update/delete)
  - **Enhanced Workspace Management**: Manage workspaces with status indicators and easy switching
  - **Dependency Viewer**: Visualize resource dependencies and relationships
  - **Target Operations**: Apply changes to specific resources with precision
  - **Approval Workflows**: Toggle approval mode for safer operations
- **File Scanning**: Detect and analyze Terraform/Bicep files
- **Template Generation**: AI-assisted IaC code creation
- **Deployment Support**: Deploy infrastructure directly from the TUI

### 🌐 **Azure Integration**
- **Azure CLI + SDK**: Dual integration for maximum compatibility
- **Demo Mode**: Full functionality without Azure credentials
- **Multiple Subscriptions**: Easy switching between Azure contexts

### ✨ **Enhanced Features (Latest)**
- **📊 Table-Formatted Properties**: Resource properties displayed in organized tables with intelligent formatting
- **🔐 Enhanced SSH for VMs**: Direct SSH (`c`) and Bastion (`b`) connections with automatic IP detection
- **🚢 Comprehensive AKS Management**: Full kubectl integration with pod (`p`), deployment (`D`), node (`n`), and service (`v`) management
- **💾 Storage Account Management**: Complete container and blob management with upload (`U`), list (`T`/`Shift+T`), and delete (`Ctrl+X`) operations
- **🤖 Manual AI Analysis**: AI analysis now requires manual trigger (`a` key) by default - set `AZURE_TUI_AUTO_AI="true"` for automatic analysis
- **📊 Progress Tracking**: Visual progress bars for storage operations and resource loading
- **⚡ Real-time Actions**: Start (`s`), stop (`S`), restart (`r`) operations with visual feedback
- **🎮 Intuitive Controls**: Enhanced keyboard shortcuts and visual indicators for all operations

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** installed
- **Azure CLI** configured (`az login`) - *optional for demo mode*
- **OpenAI API Key** for AI features (*optional*)

### Installation
```bash
# Clone the repository
git clone https://github.com/olafkfreund/azure-tui
cd azure-tui

# Build the application
go build -o aztui ./cmd

# Run with your Azure subscription
./aztui

# Or run in demo mode (no Azure setup required)
DEMO_MODE=true ./aztui
```

### First Run
1. **Configure Azure** (optional): `az login`
2. **Set OpenAI Key** (optional): `export OPENAI_API_KEY="your-key"`
3. **Launch TUI**: `./aztui`
4. **Navigate**: Use `j/k` or arrow keys to navigate
5. **Get Help**: Press `?` for keyboard shortcuts

---

## 🎮 Usage

### Navigation
- **Tree Navigation**: `j/k` or `↑/↓` to navigate resources
- **Expand/Collapse**: `Space` to toggle tree nodes
- **Open Resource**: `Enter` to open in content tab
- **Switch Tabs**: `Tab/Shift+Tab` between content tabs
- **Close Tab**: `Ctrl+W` to close active tab

### Resource Actions
- **AI Analysis**: `a` - Get AI insights for selected resource (manual trigger by default)
- **Metrics Dashboard**: `M` - View real-time resource metrics
- **Edit Configuration**: `E` - Safely modify resource settings
- **Generate Terraform**: `T` - Create Terraform code
- **Generate Bicep**: `B` - Create Bicep templates
- **Cost Analysis**: `O` - Get optimization suggestions
- **Delete Resource**: `Ctrl+D` - Safe deletion with confirmation

### Infrastructure Management
- **Terraform Manager**: `Ctrl+T` - Open enhanced Terraform integration popup
  - **Visual State Management**: `s` - Browse and inspect Terraform state resources
  - **Interactive Plan Visualization**: `p` - View plan changes with filtering options
  - **Enhanced Workspace Management**: `w` - Manage Terraform workspaces with status indicators
  - **Dependency Viewer**: `d` - Show resource dependencies and relationships
  - **Target Operations**: `t` - Apply changes to specific resources
  - **Plan Filtering**: `f` - Toggle between all/create/update/delete plan views
- **DevOps Manager**: `Ctrl+O` - Open Azure DevOps integration popup

### Storage Management (when Storage Account selected)
- **List Containers**: `T` - Show all containers with progress tracking
- **Create Container**: `Shift+T` - Create a new blob container
- **List Blobs**: `B` - Show blobs in selected container
- **Upload Blob**: `U` - Upload file to container
- **Delete Storage Items**: `Ctrl+X` - Delete containers or blobs

### Interface Modes
- **Tree View** (default): NeoVim-style interface with tree + content tabs
- **Traditional Mode**: `F2` to toggle classic two-panel layout
- **Help**: `?` to show all keyboard shortcuts
- **Quit**: `q` to exit application

---

## ⚙️ Configuration

### Environment Variables
```bash
# Azure Configuration (automatic via az login)
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
export AZURE_TENANT_ID="your-tenant-id"

# Azure DevOps Integration (optional)
export AZURE_DEVOPS_PAT="your-personal-access-token"
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_PROJECT="your-project"

# AI Integration (Choose one)
# Option 1: OpenAI API
export OPENAI_API_KEY="your-openai-api-key"

# Option 2: GitHub Copilot (Recommended)
export GITHUB_TOKEN="your-github-token"
export USE_GITHUB_COPILOT="true"  # optional, auto-detected if GITHUB_TOKEN is set

# AI Behavior Configuration
export AZURE_TUI_AUTO_AI="true"  # Enable automatic AI analysis (default: manual-only)

# Application Settings
export DEMO_MODE="true"  # Run without Azure credentials
```

### Config File (Optional)
Create `~/.config/azure-tui/config.yaml`:
```yaml
naming:
  pattern: "{{env}}-{{service}}-{{name}}"
  environments: ["dev", "staging", "prod"]

ai:
  provider: "openai"
  model: "gpt-4"
  endpoint: "https://api.openai.com/v1"

interface:
  default_mode: "tree"  # or "traditional"
  theme: "azure"

# Azure DevOps Integration (optional)
devops:
  organization: "your-organization"
  project: "your-project"
  base_url: "https://dev.azure.com"
```

---

## 🔄 Azure DevOps Integration

Azure TUI includes a comprehensive DevOps integration module for managing Azure DevOps organizations, projects, and pipelines.

### Setup
```bash
# Required: Personal Access Token with appropriate permissions
export AZURE_DEVOPS_PAT="your-personal-access-token"

# Optional: Default organization and project
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_PROJECT="your-project"
```

### Features
- **Organization Management**: List and switch between Azure DevOps organizations
- **Project Navigation**: Browse projects within organizations
- **Pipeline Discovery**: List all build and release pipelines
- **Pipeline Operations**: View pipeline details, runs, and history
- **Real-time Status**: Monitor pipeline execution status
- **Tree-based Interface**: Hierarchical view of DevOps resources

### Personal Access Token Setup
1. Go to **Azure DevOps** → **User Settings** → **Personal Access Tokens**
2. Create a new token with the following permissions:
   - **Build**: Read & execute
   - **Release**: Read, write & execute  
   - **Project and Team**: Read
   - **Identity**: Read
3. Copy the token and set the `AZURE_DEVOPS_PAT` environment variable

### Usage
The DevOps integration provides a borderless, popup-based interface similar to the Terraform integration, allowing you to:
- Navigate organizations and projects with arrow keys
- View pipeline information and recent runs
- Monitor build and release status
- Filter pipelines by name or status

**Access**: Press `Ctrl+O` from anywhere in the main TUI interface to open the Azure DevOps Manager popup.

### Navigation
- **Open DevOps Manager**: `Ctrl+O` - Access DevOps popup from main interface
- **Navigate Menu**: `↑/↓` or `j/k` - Move through DevOps options
- **Select Operation**: `Enter` - Execute selected DevOps operation
- **View Results**: `j/k` - Scroll through operation results
- **Go Back**: `Esc` - Return to previous menu or close popup
- **Help**: `?` - Show all shortcuts (includes DevOps section)

---

## 🏗️ Enhanced Terraform Integration

Azure TUI now includes a **comprehensive Terraform management suite** accessible via `Ctrl+T`, providing visual state management, interactive plan visualization, and enhanced workspace operations.

### ✨ **New Enhanced Features**

#### 🔍 **Visual State Management** (`s`)
- Browse Terraform state resources interactively
- View detailed resource properties and metadata
- Filter and search through state entries
- Examine resource dependencies and relationships

#### 📊 **Interactive Plan Visualization** (`p`) 
- View plan changes with smart filtering capabilities
- Toggle between views: All changes, Creates only, Updates only, Deletes only
- Color-coded change indicators (🟢 Create, 🟡 Update, 🔴 Delete)
- Detailed diff view for resource modifications

#### 🌐 **Enhanced Workspace Management** (`w`)**
- List all available Terraform workspaces
- Switch between workspaces with visual status indicators
- Current workspace highlighting and selection
- Workspace creation and management operations

#### 🎯 **Advanced Operations**
- **Dependency Viewer** (`d`): Visualize resource dependencies and relationships  
- **Target Operations** (`t`): Apply changes to specific resources only
- **Plan Filtering** (`f`): Cycle through filtered plan views  
- **Approval Mode** (`a`): Toggle approval workflows for safer operations

### 🚀 **Usage**

```bash
# Access enhanced Terraform integration
# Press Ctrl+T from main interface

# Navigate enhanced features:
s  - Visual State Management    
p  - Interactive Plan Visualization
w  - Enhanced Workspace Management
d  - Show Dependencies
f  - Filter Toggle (in plan view)
a  - Approval Mode Toggle  
t  - Target Resource Operations
```

### 🎨 **UI Design**
- **Frameless Design**: Consistent with Azure TUI aesthetic
- **Real-time Updates**: Live status indicators and progress feedback
- **Keyboard Navigation**: Full vim-style navigation support
- **Color Coding**: Intuitive visual indicators for different operation types

---

## 🎯 Real-World Examples

### Scenario 1: Resource Discovery & Analysis
```bash
# Launch TUI and explore your Azure resources
./aztui

# Navigate to a resource group, press 'a' for AI analysis
# Example output: "This VM is oversized for its workload. Consider downsizing to save 40% costs."
```

### Scenario 2: Infrastructure as Code Generation
```bash
# Select a VM resource, press 'T' for Terraform
# Generates complete .tf file with dependencies
# Press 'B' for Bicep equivalent
```

### Scenario 3: Cost Optimization
```bash
# Select a resource group, press 'O'
# AI analyzes all resources and suggests:
# - Right-sizing recommendations
# - Reserved instance opportunities  
# - Unused resource cleanup
```

### Scenario 4: Multi-Resource Management
```bash
# Open multiple resources in tabs (Enter key)
# Switch between tabs with Tab/Shift+Tab
# Compare configurations side-by-side
# Apply changes across multiple resources
```

---

## 🏗️ Architecture

The application follows a clean, modular architecture:

```
├── cmd/                    # Application entry point
├── internal/
│   ├── azure/             # Azure service integrations
│   │   ├── azuresdk/      # Azure SDK client
│   │   ├── aks/           # AKS management
│   │   ├── storage/       # Storage operations
│   │   └── ...            # Other Azure services
│   ├── tui/               # Terminal UI components
│   ├── openai/            # AI integration
│   ├── config/            # Configuration management
│   └── terraform/         # IaC support
```

### Key Components
- **TUI Layer**: Bubble Tea + Lipgloss for modern terminal interface
- **Azure Integration**: Dual Azure CLI + SDK support
- **AI Services**: OpenAI GPT-4 with extensible provider support
- **IaC Support**: Terraform and Bicep analysis/generation

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Setup
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build for development
go build -o aztui-dev ./cmd
```

---

### Windows Build Instructions
To build a Windows executable (`.exe`) from Linux, use Go's cross-compilation:

```bash
GOOS=windows GOARCH=amd64 go build -o aztui.exe ./cmd
```

This will create a `aztui.exe` file that you can run on Windows. You can transfer this file to a Windows machine and double-click to run, or execute from the command prompt:

```cmd
aztui.exe
```

- All features work the same as on Linux/NixOS.
- For demo mode on Windows, use:

```cmd
set DEMO_MODE=true
aztui.exe
```

#### NixOS and Linux
- For NixOS, use the provided `flake.nix` for reproducible builds (see [README-flake.md](./README-flake.md)).
- For Linux, use the standard Go build instructions above.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🆘 Support

- **Documentation**: See [Manual.md](Manual.md) for detailed usage guide
- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Join community discussions for help and ideas

---

**Transform your Azure management experience with a modern, AI-powered TUI! 🚀**

---

## See Also

- [project-plan.md](./project-plan.md)
- [README-flake.md](./README-flake.md)
