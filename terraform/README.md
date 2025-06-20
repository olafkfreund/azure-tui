# Azure TUI Terraform Templates

This directory contains pre-built Terraform templates for common Azure resources that can be managed through azure-tui.

## Directory Structure

```
terraform/
├── templates/          # Pre-built Terraform templates
│   ├── vm/            # Virtual Machine templates
│   ├── sql/           # Azure SQL templates  
│   ├── aks/           # Azure Kubernetes Service templates
│   ├── aci/           # Azure Container Instances templates
│   └── modules/       # Reusable Terraform modules
├── workspaces/        # User Terraform workspaces
├── state/             # Terraform state files (local)
└── examples/          # Example configurations
```

## Template Categories

### Virtual Machines
- **linux-vm**: Standard Linux VM with SSH access
- **windows-vm**: Windows VM with RDP access
- **vm-with-loadbalancer**: VM behind Azure Load Balancer

### Azure SQL
- **sql-server**: Basic SQL Server deployment
- **sql-database**: SQL Database with security configurations
- **sql-elastic-pool**: Elastic pool for multiple databases

### Azure Kubernetes Service (AKS)
- **basic-aks**: Simple AKS cluster
- **aks-with-acr**: AKS with Azure Container Registry
- **aks-production**: Production-ready AKS with monitoring

### Azure Container Instances
- **single-container**: Simple container deployment
- **multi-container**: Container group with multiple containers
- **container-with-storage**: Container with persistent storage

## Usage

1. **Browse Templates**: Use azure-tui to browse available templates
2. **Customize**: Modify templates using AI assistance or external editor
3. **Deploy**: Plan and apply infrastructure changes
4. **Manage**: Monitor state and perform lifecycle operations

## AI Integration

All templates support AI-powered:
- Code generation and modification
- Best practice recommendations
- Security configuration suggestions
- Cost optimization advice

## Editor Integration

Templates can be edited using:
- Built-in TUI editor
- External editors (vim, neovim, vscode)
- AI-assisted code completion
