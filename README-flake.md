# Azure TUI/CLI - Nix Flake Installation & Build Guide

This guide explains how to set up, build, and run the Azure TUI/CLI app using Nix Flakes for a fully reproducible development environment.

## Prerequisites
- [Nix](https://nixos.org/download.html) (with flakes enabled)
- Git

## Quickstart

```sh
# Clone the repository
$ git clone https://github.com/your-org/azure-tui.git
$ cd azure-tui

# Enter the Nix development shell (installs Go, Node, Azure CLI, etc.)
$ nix develop

# Build the app using Just or Go
$ just           # (recommended, runs go build)
# or
$ go build -o aztui ./cmd/main.go

# Run the app
$ ./aztui
```

## Flake Features
- **Reproducible Go/Node/Azure CLI dev environment**
- **`nix develop`**: Drops you into a shell with all dependencies
- **`nix build`**: Builds the app using Go modules
- **`just`**: Provides common build/test commands

## Example: Build and Run with Flakes
```sh
# Build the binary using Nix Flake
$ nix build
# The binary will be in ./result/bin/aztui
$ ./result/bin/aztui
```

## Troubleshooting
- If you see errors about missing flakes, ensure you have enabled flakes in your Nix config.
- For Azure authentication, run `az login` before starting the app.

---
For more details, see the main README below.
