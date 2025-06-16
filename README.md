# azure-tui
TUI for Azure onboarding and management using copilot. Managed different Azure profiles and environments for one place.

## Getting Started

### Prerequisites
- Go 1.21+
- Azure CLI (for authentication)

### Setup
```sh
# Clone the repo
# cd into the directory
cd azure-tui

# Build
go build -o azure-tui ./cmd/main.go

# Run
./azure-tui
```

## Roadmap
- [x] Project initialization
- [ ] Azure profile management
- [ ] Environment dashboard
- [ ] VM login and management
- [ ] Terraform/Bicep integration
- [ ] OpenAI Copilot integration
- [ ] UI/UX polish
