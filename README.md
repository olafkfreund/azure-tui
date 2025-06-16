# Azure TUI/CLI (aztui)

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

---

## Getting Started

1. **Install Go** (>=1.21)
2. **Clone this repo**
3. **Configure Azure CLI**: `az login`
4. **Create/Edit config**: `~/.config/azure-tui/config.yaml`
5. **Run TUI**: `go run ./cmd/main.go`
6. **Use CLI**: `go run ./cmd/main.go --create --resource <type>`

---

## Configuration

- `~/.config/azure-tui/config.yaml`: Naming standards, AI model, Copilot token, agent selection, etc.
- See `internal/config/config.go` for config structure.

---

## AI Integration

- **Copilot agents**: Defined for IaC, troubleshooting, security, cost, documentation, CLI help
- **Scenario-driven agent selection**: Helper function for mapping scenario to agent
- **Configurable**: Model, API base, agent, and prompt can be set via config/env
- **Usage**: All TUI/CLI workflows can invoke AI for code generation, validation, troubleshooting, and documentation

---

## Roadmap

- AI-driven code generation, validation, troubleshooting in resource workflows
- Streaming/multi-turn context for AI
- In-place IaC editing and validation
- Advanced resource actions (SSH, monitoring)
- More config-driven customization

---

## See Also

- [project-plan.md](./project-plan.md)
- [README-flake.md](./README-flake.md)
