# Azure TUI (aztui) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Azure](https://img.shields.io/badge/Azure-CLI%20%2B%20SDK-0078D4?style=flat&logo=microsoft-azure)](https://azure.microsoft.com/)
[![AI Powered](https://img.shields.io/badge/AI-OpenAI%20GPT--4-412991?style=flat&logo=openai)](https://openai.com/)

A modern, **NeoVim-style Terminal User Interface** for managing Azure resources with AI-powered analysis, Infrastructure as Code generation, and comprehensive resource management capabilities.

## ✨ Features

### 🎯 **NeoVim-Style Interface**
- **Tree View Navigation**: Hierarchical display of Azure resources with expand/collapse
- **Powerline Statusbar**: Modern statusbar with subscription, tenant, and resource context
- **Content Tabs**: Right-side tabbed interface for opened resources
- **Vim-like Navigation**: `j/k` keys, space for expand/collapse, familiar shortcuts
- **Mouse Support**: Scroll wheel navigation and click interactions

### 🤖 **AI-Powered Resource Management**
- **Intelligent Analysis** (`a`): AI-powered resource insights and recommendations
- **Code Generation**: Generate Terraform (`T`) and Bicep (`B`) templates
- **Cost Optimization** (`O`): AI-driven cost savings suggestions
- **Security Analysis**: Automated security posture assessment

### 📊 **Interactive Dashboards**
- **Real-time Metrics** (`M`): CPU, memory, network, and disk usage with trends
- **Resource Editor** (`E`): Safe configuration editing with validation
- **Delete Protection** (`Ctrl+D`): Confirmation dialogs prevent accidental deletions

### 🔧 **Infrastructure as Code**
- **File Scanning**: Detect and analyze Terraform/Bicep files
- **Template Generation**: AI-assisted IaC code creation
- **Deployment Support**: Deploy infrastructure directly from the TUI

### 🌐 **Azure Integration**
- **Azure CLI + SDK**: Dual integration for maximum compatibility
- **Demo Mode**: Full functionality without Azure credentials
- **Multiple Subscriptions**: Easy switching between Azure contexts

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
- **AI Analysis**: `a` - Get AI insights for selected resource
- **Metrics Dashboard**: `M` - View real-time resource metrics
- **Edit Configuration**: `E` - Safely modify resource settings
- **Generate Terraform**: `T` - Create Terraform code
- **Generate Bicep**: `B` - Create Bicep templates
- **Cost Analysis**: `O` - Get optimization suggestions
- **Delete Resource**: `Ctrl+D` - Safe deletion with confirmation

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

# AI Integration
export OPENAI_API_KEY="your-openai-api-key"
export AZURE_MCP_ENDPOINT="http://localhost:5030/v1"  # optional

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
```

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
