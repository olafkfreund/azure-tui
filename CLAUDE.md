# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Azure TUI (`azure-tui`) is a modern Terminal User Interface for managing Azure resources, built in Go using the Bubble Tea framework. The application provides NeoVim-style navigation, AI-powered resource analysis, Infrastructure as Code generation, and comprehensive Azure resource management capabilities.

## Build & Development Commands

### Essential Commands (use `just` command runner)
- **Build**: `just build` - Build for current platform
- **Run**: `just run` - Start the TUI
- **Test**: `just test` - Run all tests
- **Format**: `just fmt` - Format Go code
- **Clean**: `just clean` - Clean build artifacts

### Alternative Go Commands
- **Build**: `go build -o azure-tui ./cmd/main.go`
- **Run**: `go run ./cmd/main.go` 
- **Test**: `go test ./...`
- **Tidy**: `go mod tidy`

### Development Modes
- **Debug**: `just run-debug` or `DEBUG=true go run ./cmd/main.go`
- **Demo Mode**: `DEMO_MODE=true ./azure-tui` (no Azure credentials required)
- **OpenAI Provider**: `just run-openai` or `USE_GITHUB_COPILOT=false go run ./cmd/main.go`

### Quality Assurance
- **Security Check**: `just security` (requires gosec)
- **All QA**: `just qa` (format, tidy, test)
- **Coverage**: `just test-coverage`

## Architecture Overview

### Core Structure
```
cmd/main.go              # Application entry point
internal/
├── tui/                 # Bubble Tea TUI framework (main interface logic)
├── azure/               # Azure service integrations
│   ├── azuresdk/        # Azure SDK client wrapper
│   ├── aks/            # AKS (Kubernetes) management
│   ├── storage/        # Storage account operations  
│   ├── devops/         # Azure DevOps integration
│   └── tfbicep/        # Terraform/Bicep IaC support
├── openai/             # AI integration (OpenAI/GitHub Copilot)
├── config/             # Configuration management
├── terraform/          # Terraform integration
└── search/             # Resource search functionality
```

### Key Components

**TUI Framework (`internal/tui/`)**: 
- Bubble Tea-based terminal interface with NeoVim-style navigation
- Tab management system for multiple resource views
- Tree view for hierarchical resource display
- Popup system for interactive operations

**Azure Integration (`internal/azure/`)**:
- Dual Azure CLI + SDK support for maximum compatibility
- Service-specific modules (AKS, Storage, DevOps, etc.)
- Resource details and actions management

**AI Services (`internal/openai/`)**:
- Supports both OpenAI API and GitHub Copilot
- Resource analysis, cost optimization, security assessment
- Infrastructure as Code generation (Terraform/Bicep)

### Technology Stack
- **UI**: Bubble Tea + Lipgloss (terminal interface)
- **Azure**: Azure SDK for Go + Azure CLI
- **AI**: OpenAI GPT-4 or GitHub Copilot API
- **IaC**: Terraform and Bicep integration
- **Language**: Go 1.24+

## Key Features & Navigation

### Main Interface Modes
- **Tree View** (default): NeoVim-style interface with resource tree + content tabs
- **Traditional Mode**: `F2` to toggle classic two-panel layout

### Core Keyboard Shortcuts
- **Navigation**: `j/k` or `↑/↓` (tree navigation), `Space` (expand/collapse)
- **Resource Actions**: `a` (AI analysis), `M` (metrics), `E` (edit), `T` (Terraform), `B` (Bicep)
- **Infrastructure**: `Ctrl+T` (Terraform manager), `Ctrl+O` (DevOps manager)
- **Tabs**: `Tab/Shift+Tab` (switch tabs), `Ctrl+W` (close tab), `Enter` (open resource)
- **Operations**: `s` (start), `S` (stop), `r` (restart), `Ctrl+D` (delete with confirmation)

### AI Integration Modes
- **Manual AI** (default): Press `a` key to trigger analysis
- **Automatic AI**: Set `AZURE_TUI_AUTO_AI="true"` for automatic analysis on resource selection

## Environment Variables

### Required for Full Functionality
```bash
# Azure (automatic via az login)
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
export AZURE_TENANT_ID="your-tenant-id"

# AI Integration (choose one)
export OPENAI_API_KEY="your-openai-api-key"
# OR
export GITHUB_TOKEN="your-github-token"
export USE_GITHUB_COPILOT="true"

# Azure DevOps (optional)
export AZURE_DEVOPS_PAT="your-personal-access-token" 
export AZURE_DEVOPS_ORG="your-organization"
export AZURE_DEVOPS_PROJECT="your-project"
```

### Behavior Configuration
```bash
export AZURE_TUI_AUTO_AI="true"    # Enable automatic AI analysis
export DEMO_MODE="true"            # Run without Azure credentials
export DEBUG="true"                # Enable debug logging
```

## Development Notes

### Test Structure
- Tests are located alongside source files (`*_test.go`)
- Main test areas: `internal/openai/ai_test.go`, `internal/search/search_test.go`, `internal/tui/tui_test.go`
- Use `just test-pkg PKG` to test specific packages

### Key Integration Points
- **Azure Authentication**: Uses `azidentity.DefaultAzureCredential` with fallback to Azure CLI
- **AI Provider Selection**: Auto-detects GitHub Copilot via `GITHUB_TOKEN`, falls back to OpenAI
- **Configuration**: YAML config file at `~/.config/azure-tui/config.yaml` (optional)

### Special Terraform Integration
- Enhanced Terraform manager accessible via `Ctrl+T`
- Features: visual state management, interactive plan visualization, workspace management
- Supports multiple Terraform backends and workspace switching
- Templates available in `terraform/templates/` directory

### Cross-Platform Support
- Primary development on Linux/NixOS
- Windows builds: `just build-windows` or `GOOS=windows GOARCH=amd64 go build -o aztui.exe ./cmd/main.go`
- NixOS support via `flake.nix` for reproducible builds

## Common Development Patterns

### Adding New Azure Service Integration
1. Create service package in `internal/azure/[service]/`
2. Implement Azure SDK client wrapper
3. Add TUI integration in appropriate location
4. Update navigation and keyboard shortcuts as needed

### Adding New AI Features
1. Extend `internal/openai/ai.go` with new analysis functions
2. Add corresponding TUI actions in resource action handlers
3. Update help documentation and keyboard shortcuts

### Testing Approach
- Use demo mode for UI/UX testing: `DEMO_MODE=true ./azure-tui`
- Demo scripts available in `demo/` directory for testing specific features
- Script tests in `script-tests/` for automated testing scenarios