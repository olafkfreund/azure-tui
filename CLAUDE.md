# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build the application
just         # or: go build -o azure-tui ./cmd/main.go

# Run the application
just run     # or: go run ./cmd/main.go

# Run all tests
just test    # or: go test ./...

# Format code
just fmt     # or: gofmt -w .

# Tidy dependencies
just tidy    # or: go mod tidy

# Lint code (requires golangci-lint)
just lint    # or: golangci-lint run ./...

# Clean build artifacts
just clean   # removes azure-tui binary
```

## Architecture Overview

This is a Go-based Terminal User Interface for Azure resource management built with the Bubble Tea framework, following a Model-View-Update (MVU) architecture.

### Core Architecture Patterns

1. **Event-Driven Message System**: All UI updates happen through typed messages (`subscriptionsLoadedMsg`, `resourceDetailsLoadedMsg`, etc.) defined in `cmd/main.go`. Commands return `tea.Cmd` functions for async operations.

2. **Progressive Resource Loading**: Resources load on-demand in this flow:
   - Initial: Load subscriptions and resource groups
   - User expands group → Load resources in that group
   - User selects resource → Load details → Optionally load AI description

3. **Dual Azure Integration**:
   - Primary: Azure CLI commands with JSON parsing (most operations)
   - Secondary: Azure SDK for Go (specific operations requiring better performance)

4. **Panel Navigation System**:
   - Left panel (1/3 width): Tree view with vim-style navigation (j/k, h/l)
   - Right panel (2/3 width): Resource details/dashboards
   - Tab to switch panels, visual indicators show active panel

### Key Architectural Components

- **Entry Point**: `cmd/main.go` contains the main TUI logic and state management
- **State Container**: The `model` struct in `cmd/main.go` holds all application state
- **Azure Services**: `internal/azure/` modules handle resource-specific operations
- **UI Components**: `internal/tui/` contains reusable UI components (TreeView, StatusBar)
- **Network Module**: `internal/azure/network/network.go` demonstrates comprehensive resource handling patterns

### Important Implementation Details

1. **Resource Actions**: Context-sensitive actions mapped by resource type (e.g., VMs have start/stop/ssh, AKS has pods/nodes/services)

2. **View Management**: The `activeView` string in the model controls displayed content (welcome, details, dashboard, network-dashboard, etc.)

3. **Error Handling**: Graceful degradation - operations continue even if some resources fail to load

4. **AI Integration**: Optional OpenAI integration through `internal/openai/` for resource analysis

5. **Demo Mode**: Set `DEMO_MODE=true` to run without Azure credentials

### Testing Approach

Standard Go testing with test files in `/test/` directory and alongside source files (e.g., `internal/tui/tui_test.go`).

### Common Development Tasks

When adding new Azure resource types:
1. Create module in `internal/azure/<resource>/`
2. Add message types in `cmd/main.go`
3. Implement load command function
4. Add case to Update() method for handling messages
5. Map resource actions in action handling logic