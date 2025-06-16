# Azure TUI Project Plan

## Vision
A modern, user-friendly TUI and CLI for managing Azure environments, subscriptions, tenants, and resources, with advanced features like VM login, configuration management, usage/alarms, and AI-powered insights (OpenAI Copilot compatible).

---

## Core Features

### 1. Azure Authentication & Context

- [x] List all Azure subscriptions and tenants using Azure CLI
- [x] Switch between subscriptions and tenants in the TUI
- [x] Set active subscription/tenant (Enter key)
- [x] Show current active context in the UI

### 2. Profile & Environment Management

- [ ] Support multiple Azure CLI profiles
- [ ] Save/load custom environment sets

### 3. Resource Dashboard (TUI & CLI)

- [x] List resource groups and resources for selected subscription/tenant
- [x] Show resource status and details
- [x] Modular resource management (AKS, Key Vault, VNet, Firewall, Storage, ACR, ACI, SQL, etc.)
- [ ] Add, delete, and update resources interactively
- [ ] Print usage matrix and alarms for all resources
- [ ] Proactive health checks and recommendations

### 4. VM & Resource Actions

- [ ] SSH login to VMs from TUI/CLI
- [ ] Start/stop/reboot VMs
- [x] AKS: List, create (with prompt), delete, authenticate (get-credentials)
- [ ] Key Vault: List, create, delete
- [ ] Storage: List, create, delete
- [ ] ACR: List, create, delete
- [ ] ACI: List, create, delete
- [ ] SQL: List, create, delete
- [ ] VNet/Firewall: List, create, delete

### 5. Usage, Alarms, and Monitoring

- [x] Modular usage/alarms logic (internal/azure/usage.go)
- [ ] Show usage matrix for all resources
- [ ] List and manage alarms
- [ ] Proactive recommendations and alerts

### 6. Copilot/AI Integration

- [x] Modular OpenAI/Copilot provider (internal/openai/ai.go)
- [ ] Ask questions about environment, resources, and logs
- [ ] Summarize, explain, and recommend actions
- [ ] Integrate with TUI and CLI for context-aware Q&A

### 7. Infrastructure as Code (Terraform/Bicep)

- [x] Modular helpers for Terraform and Bicep (internal/azure/tfbicep.go)
- [ ] Read, change, create, and delete resources via Terraform/Bicep
- [ ] Generate code from TUI/CLI and AI suggestions

### 8. CLI Interface

- [ ] Expose all features via well-formed CLI commands
- [ ] Support scripting and automation

### 9. UI/UX Polish

- [x] Modern, styled TUI with Bubble Tea & Lipgloss
- [ ] Keyboard shortcuts and help screens
- [ ] Error handling and loading indicators

---

## Stretch Goals

- [ ] Multi-user collaboration
- [ ] Plugin system for custom actions
- [ ] Export/share environment configs
- [ ] Advanced log analytics and visualization

---

## Change Log

- 2025-06-16: Modular resource management, usage/alarms, and AI provider structure implemented. Interactive AKS, resource, and context management in TUI. CLI/TUI modularization started.

---

## Next Steps

- Integrate more resource types into TUI/CLI (Key Vault, Storage, ACR, ACI, SQL, VNet, Firewall)
- Add usage matrix and alarms to dashboard
- Expose all features via CLI
- Integrate AI provider for Q&A, log parsing, and code generation
- Add Terraform/Bicep code generation and editing
