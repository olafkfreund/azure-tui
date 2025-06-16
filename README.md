# Azure TUI/CLI

A modern, keyboard-driven TUI and CLI for managing Azure resources, environments, and infrastructure-as-code (Terraform/Bicep), with AI-powered code generation, YAML config, and advanced log/usage popups.

---

## Features

- Azure resource management (AKS, Key Vault, Storage, VNet, Firewall, etc.)
- Interactive TUI (Bubble Tea, Lipgloss) with two-panel layout
- IaC file scanning, viewing, and (soon) editing for Terraform/Bicep/tfstate
- AI-powered code generation, log analysis, and resource recommendations
- YAML config in `~/.config/azure-tui/` for persistent settings
- Azure authentication via `az login` and standard CLI context
- Matrix/log popups, alarms, and usage dashboards
- Fallback/demo data for offline use

---

## Installation

### Option 1: Go (manual)

```sh
git clone https://github.com/your-org/azure-tui.git
cd azure-tui
go build -o aztui ./cmd/main.go
./aztui
```

### Option 2: Nix Flake (recommended, reproducible)

```sh
git clone https://github.com/your-org/azure-tui.git
cd azure-tui
nix develop      # Enter dev shell with Go, Node, Azure CLI, etc.
just             # Build using Justfile
./aztui
```

See `README-flake.md` for full Nix Flake instructions.

---

## Authentication

- Run `az login` to authenticate with Azure before starting the app.
- The TUI will use your current Azure CLI context (subscription, tenant).

---

## Usage: TUI Walkthrough

### Start the App

```sh
./aztui
```

### Keyboard Shortcuts

- `tab` / `shift+tab`: Switch subscription/tenant
- `up` / `down`: Navigate resource groups
- `right` / `left`: Navigate resources in group
- `enter`: Set active context or select
- `i`: Open/close IaC file panel (Terraform/Bicep)
- `F`: Scan for IaC files in current directory
- `n` / `p`: Next/previous IaC file
- `v`: View selected IaC file
- `esc`: Close popups or panels
- `m`: Show usage matrix
- `A`: Show alarms popup
- `q`: Quit

### Example: Scan and View Terraform/Bicep Files

1. Press `i` to open the IaC panel
2. Press `F` to scan for `.tf`, `.bicep`, `.tfstate` files
3. Use `n`/`p` to select a file
4. Press `v` to view the file contents
5. Press `esc` to close the file view or panel

### Example: List and Manage Azure Resources

- Use arrow keys to select a resource group and resource
- Press `d` to view resource details
- Use `k`/`K` to list or create AKS clusters
- Use `v`/`V` for Key Vaults, `s`/`S` for Storage, etc.

---

## Advanced: AI & Automation

- (Planned) Use AI to generate Terraform/Bicep code, recommend resources, and analyze logs
- (Planned) Edit IaC files in-place and apply changes
- (Planned) YAML config at `~/.config/azure-tui/config.yaml` for persistent settings

---

## Real-Life Example: Deploy a VM with Terraform

1. Place your `.tf` files in the project directory
2. Open the TUI, press `i` and `F` to scan
3. Select your main Terraform file, press `v` to review
4. (Planned) Press `e` to edit, or use AI to generate new code
5. (Planned) Apply changes from the TUI or CLI

---

## Troubleshooting

- If you see no Azure resources, check your `az login` status
- If the TUI shows only demo data, ensure Azure CLI is installed and authenticated
- For IaC file scanning, ensure files are in the current directory or subfolders

---

## Contributing

PRs and issues welcome! See `project-plan.md` for roadmap and architecture.

---

## License

MIT
