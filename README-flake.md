# Azure TUI/CLI (aztui) - Nix/Flake Manual

## Nix/Flake Usage

- Build with flakes: `nix build`
- Run with flakes: `nix run`
- Development shell: `nix develop`
- See `flake.nix` for details

---

## Configuration

- User config: `~/.config/azure-tui/config.yaml`
- Set naming standards, AI model, Copilot token, agent selection, etc.
- See `internal/config/config.go` for config structure

---

## AI Integration

- **Copilot agents**: For IaC, troubleshooting, security, cost, documentation, CLI help
- **Scenario-driven agent selection**: Helper function for mapping scenario to agent
- **Configurable**: Model, API base, agent, and prompt can be set via config/env
- **Usage**: All TUI/CLI workflows can invoke AI for code generation, validation, troubleshooting, and documentation

---

## Advanced Workflows

- Use TUI or CLI for all resource actions
- AI-powered code generation, validation, troubleshooting
- In-place IaC editing and validation (coming soon)
- Streaming/multi-turn AI context (coming soon)

---

## Roadmap

- Full AI integration in resource workflows
- More config-driven customization
- Advanced resource actions (SSH, monitoring)

---

## See Also

- [README.md](./README.md)
- [project-plan.md](./project-plan.md)
