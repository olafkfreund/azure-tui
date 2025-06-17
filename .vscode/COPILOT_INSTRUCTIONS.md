# Copilot Project Instructions for azure-tui

## Mandatory: Read Before Each New Chat

Welcome to the azure-tui project! Please read and follow these instructions and best practices every time you start a new Copilot chat or coding session in this repository.

---

## Project Overview
- **Purpose:** Modular, extensible Azure TUI/CLI tool for managing Azure resources (subscriptions, tenants, resource groups, AKS, Key Vault, Storage, ACR, ACI, SQL, VNet, Firewall, etc.), with usage/alarms, AI/Copilot integration, and Terraform/Bicep support.
- **Tech Stack:** Go, Bubble Tea (TUI), Azure CLI/SDK, OpenAI/Copilot, Nix (flake), Justfile, and optional Node.js for AI tools.

---

## Best Practices & Rules

### General
- **Always** keep the project modular: each Azure resource type should have its own Go package/module.
- **Never** hardcode secrets, API keys, or credentials in code or config.
- **Document** all new features and CLI/TUI commands in the README or project-plan.md.
- **Write clear, concise commit messages** and keep PRs focused.
- **Use the Justfile and flake.nix** for all builds, tests, and dev shells.
- **Prefer the Azure Go SDK** for new features; use Azure CLI only for quick prototyping or where SDK support is missing.
- **Keep the TUI/CLI and AI logic decoupled** and testable.

### TUI/CLI
- **All features must be accessible via both TUI and CLI.**
- **Use Bubble Tea idioms** for state, messages, and prompt handling.
- **Add keyboard shortcuts and help screens** for all new TUI features.
- **Show clear error messages and loading indicators** in the TUI.
- **For popups/alarms/matrix/graphs:**
  - Use Bubble Tea popups for alarms/errors.
  - Use ASCII/Unicode graphs for matrix/usage visualizations.
  - Always provide a fallback text summary for non-graph terminals.

### AI/Copilot
- **All AI features must use the OpenAI provider abstraction.**
- **Summarize, explain, and recommend actions for all resource types.**
- **Never expose sensitive data in AI prompts or completions.**
- **Document all AI prompt templates and logic.**

### Nix/Dev Environment
- **Use the provided flake.nix and Justfile** for all development.
- **Add new tools to the devShell in flake.nix as needed.**
- **Test builds and CLI commands in a clean shell before PR.**

### Testing & Quality
- **Write tests for all new modules and features.**
- **Run `just test` and `just lint` before every commit.**
- **Keep code idiomatic and formatted (`just fmt`).**

---

## Process
1. **Start every session by reading this file.**
2. **Check project-plan.md for current priorities and pending tasks.**
3. **Use the dev shell (`nix develop`) and Justfile for all commands.**
4. **Add new features in a modular way, update docs, and test.**
5. **Ask Copilot for help, but always review and adapt its suggestions.**
6. **Keep the TUI, CLI, and AI features in sync.**
7. **Update this file and project-plan.md with any new rules or processes.**

---

## Quick Reference
- **Build:** `just` or `just build`
- **Run:** `just run`
- **Test:** `just test`
- **Format:** `just fmt`
- **Dev Shell:** `nix develop`
- **Lint:** `just lint`

---

Thank you for contributing to azure-tui! Always read this file before starting a new Copilot chat or coding session.
