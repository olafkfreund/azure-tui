# Project Plan: Azure TUI/CLI (aztui)

## Overview

A modular Go TUI/CLI tool for managing Azure resources, supporting reading, listing, editing, and deploying Terraform/Bicep files and state, with AI-powered code generation, validation, and troubleshooting. The app uses standard az login/config for authentication, YAML config for naming and settings, and provides a modern, keyboard-driven TUI interface. All TUI functions are available via CLI, and AI always guides and confirms with the user.

---

## Features

- Modular Azure resource logic (internal/azure/*)
- Two-panel TUI layout with popups for logs/alarms
- IaC file scanning (Terraform/Bicep/tfstate) and navigation
- CLI flags for resource creation (`--create`) and deployment (`--deploy`), config-driven naming
- YAML config loader (`internal/config/config.go`), sample config at `~/.config/azure-tui/config.yaml`
- OpenAI/Copilot API integration with model selection, agent/role/prompt flexibility, Copilot tokens
- Multiple Copilot agents for Azure scenarios: IaC, troubleshooting, security, cost, documentation, CLI help
- Helper for scenario-driven agent selection
- All features documented for TUI and CLI usage
- Fallback/demo data for offline use
- **Multi-tab and window support in TUI**: Create new tabs/windows for resource management, connections (AKS, VM), monitoring, health checks, etc., similar to tmux/zellij. Tabs can be opened/closed dynamically, supporting nested/multiple interfaces.
- **Status line**: Persistent status bar at the bottom showing environment and connection status.
- **Popup for shortcuts**: Keyboard shortcut popup for TUI navigation and actions.
- **Resource/connection/monitoring tabs**: Actions for AKS, VM, monitoring, and health checks now open new tabs in the TUI, allowing parallel management and nested interfaces.

---

## ✅ COMPLETED FEATURES (June 2025)

### Core TUI Enhancements
- ✅ **Fixed Unicode/emoji alignment** using go-runewidth for all rendering
- ✅ **Modern styling** with lipgloss throughout the interface
- ✅ **Two-panel layout** (left: resource groups, right: resources)
- ✅ **Tabbed interface** with Azure service-specific icons
- ✅ **Resource navigation** with up/down for groups, left/right for resources
- ✅ **Tab management** with open/close/switch functionality

### AI-Powered Features
- ✅ **AI resource analysis** (`a` key) - Comprehensive resource insights
- ✅ **Terraform code generation** (`T` key) - Generate .tf files for resources
- ✅ **Bicep code generation** (`B` key) - Generate .bicep files for resources  
- ✅ **Cost optimization analysis** (`O` key) - AI-powered cost savings suggestions
- ✅ **OpenAI integration** with configurable API endpoints and keys

### Interactive Dialogs & Actions
- ✅ **Metrics dashboard** (`M` key) - Real-time resource metrics with trends
- ✅ **Resource editor** (`E` key) - Edit resource configurations safely
- ✅ **Delete confirmation** (`Ctrl+D`) - Safe resource deletion with warnings
- ✅ **Resource actions menu** - Context-aware action suggestions
- ✅ **Keyboard shortcuts help** (`F1`) - Complete shortcut reference

### Technical Improvements
- ✅ **Enhanced error handling** with graceful fallbacks to demo data
- ✅ **Responsive rendering** with proper panel sizing and alignment
- ✅ **Azure CLI integration** with SDK fallback support
- ✅ **Demo mode** for offline usage and testing
- ✅ **Configuration system** with YAML support

### Advanced Integrations
- ✅ **Multiple AI providers** (OpenAI, MCP server support)
- ✅ **Azure SDK integration** for reliable resource operations
- ✅ **Infrastructure as Code** support for Terraform and Bicep
- ✅ **Resource type detection** with service-specific icons and actions

---

## In Progress / Next Steps

- **AI-driven code generation, validation, troubleshooting**: Wire Copilot agents into resource creation and deployment workflows (TUI/CLI)
- **Streaming/multi-turn context** for AI interactions
- **Polish TUI/CLI UX**: error handling, progress, user guidance
- **Config-driven customization**: agents, prompts, user scenarios
- **In-place IaC editing and validation**
- **Expand resource types and advanced actions** (SSH, advanced monitoring)
- **Implement multi-tab/window TUI**: Add tab/window management, tabbed connections for AKS/VM, monitoring, health checks, and nested interfaces. Implement tab open/close logic and status line. Add popup for keyboard shortcuts. **Resource/connection/monitoring actions are now wired to open new tabs.**

---

## File Map

- `cmd/main.go`: TUI/CLI logic, resource loading, IaC panel, popups, CLI entry points, tab/connection/monitoring logic
- `internal/azure/tfbicep/filescan.go`, `tfbicep.go`: IaC file scanning, Terraform/Bicep helpers
- `internal/config/config.go`: YAML config loader, naming standards
- `internal/openai/openai.go`: OpenAI/Copilot integration, agent/role/prompt logic
- `internal/tui/tui.go`: TUI logic, panels, popups, tab/window management
- `~/.config/azure-tui/config.yaml`: user config for naming, AI, etc.
- `README.md`, `README-flake.md`, `project-plan.md`: user and dev documentation

---

## AI Integration

- **Copilot agents**: Defined for IaC, troubleshooting, security, cost, documentation, CLI help
- **Scenario-driven agent selection**: Helper function for mapping scenario to agent
- **Configurable**: Model, API base, agent, and prompt can be set via config/env
- **Usage**: All TUI/CLI workflows can invoke AI for code generation, validation, troubleshooting, and documentation

---

## Manual/Docs To Update

- README.md: Add AI workflow usage, agent scenarios, config-driven customization, multi-tab/window TUI, and resource/connection/monitoring tab actions
- README-flake.md: Nix/Flake usage, update for new config and AI features
- project-plan.md: This file (updated)
