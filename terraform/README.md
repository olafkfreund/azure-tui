# 🏗️ Terraform Integration

Azure TUI now includes comprehensive Terraform support, allowing you to manage Infrastructure as Code directly from the TUI with AI assistance.

## 📁 Project Structure

```
terraform/
├── main.tf           # Provider configuration
├── variables.tf      # Input variables
├── outputs.tf        # Output values
├── vm.tf            # Virtual Machine resources
├── aks.tf           # AKS cluster resources
├── containers.tf    # Container instances
├── network.tf       # Networking resources
└── config-example.yaml  # Configuration example
```

## 🚀 Quick Start

### 1. Prerequisites

- Terraform CLI installed (`brew install terraform` or download from [terraform.io](https://terraform.io))
- Azure CLI authenticated (`az login`)
- SSH key pair for VM access (`ssh-keygen -t rsa`)

### 2. Initialize and Deploy

```bash
# Navigate to terraform directory
cd terraform/

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy infrastructure
terraform apply

# When done, destroy resources
terraform destroy
```

### 3. Access Your Resources

After deployment, you'll get outputs like:

```bash
# SSH to VM
ssh azureuser@<vm_public_ip>

# Access container apps
curl http://<container_fqdn>

# Connect to AKS
az aks get-credentials --resource-group rg-azure-tui --name azure-tui-aks
kubectl get nodes
```

## ⚙️ Configuration

### Terraform Settings in Azure TUI

The Azure TUI configuration supports Terraform integration:

```yaml
# ~/.config/azure-tui/config.yaml
terraform:
  source_folder: "./terraform"
  default_location: "uksouth"
  ai_provider: "openai"
  auto_format: true
  validate_on_save: true
  state_backend: "local"
  auto_init: true
  confirm_destroy: true
```

### Resource Customization

Edit `terraform/variables.tf` to customize resources:

```hcl
variable "location" {
  description = "Azure region"
  default     = "uksouth"
}

variable "vm_size" {
  description = "VM size"
  default     = "Standard_B1s"
}

variable "aks_node_count" {
  description = "AKS node count"
  default     = 1
}
```

## 🤖 AI-Powered Features

### Generate Terraform Code

```go
// Example using AI to generate Terraform code
req := openai.TerraformRequest{
    ResourceType: "Azure Virtual Machine",
    Description:  "Web server with load balancer",
    Requirements: []string{
        "Ubuntu 22.04",
        "Standard_B2s size",
        "Public IP",
        "SSH access",
    },
    Location:    "uksouth",
    Environment: "dev",
    Tags: map[string]string{
        "Project": "azure-tui",
        "Environment": "dev",
    },
}

response, err := ai.GenerateTerraformCodeAdvanced(req)
```

### Code Optimization

```go
// Optimize existing Terraform code
optimized, err := ai.OptimizeTerraformCode(terraformCode, "security and cost")
```

### Code Explanation

```go
// Get explanation of Terraform code
explanation, err := ai.ExplainTerraformCode(terraformCode)
```

## 📋 Available Resources

### Virtual Machine (`vm.tf`)
- **Ubuntu 22.04 LTS** virtual machine
- **Public IP** with SSH access
- **Network Security Group** with standard rules
- **Premium SSD** storage
- **Docker** pre-installed

### AKS Cluster (`aks.tf`)
- **Small Kubernetes cluster** (1-3 nodes)
- **Auto-scaling** enabled
- **Azure CNI** networking
- **Log Analytics** integration
- **Azure Policy** enabled

### Container Instances (`containers.tf`)
- **Two hello-world containers**
- **Public IP addresses**
- **Custom DNS names**
- **HTTP endpoints**

### Networking (`network.tf`)
- **Virtual Network** with multiple subnets
- **Network Security Groups**
- **Public IP addresses**
- **Load balancers** (when needed)

## 🔧 Terraform Management

### File Operations

```go
// Create new Terraform file
tm, _ := terraform.NewTerraformManager()
err := tm.CreateFile("storage.tf", storageConfig)

// Update existing file
err = tm.UpdateFile("vm.tf", newVMConfig)

// Delete file
err = tm.DeleteFile("unused.tf")

// List all files
files, err := tm.ListFiles()
```

### Deployment Operations

```go
// Initialize Terraform
err := tm.Init()

// Generate plan
plan, err := tm.Plan()
fmt.Printf("Changes: +%d ~%d -%d\n", plan.Add, plan.Change, plan.Destroy)

// Apply changes
err = tm.Apply()

// Destroy infrastructure
err = tm.Destroy()
```

### State Management

```go
// Get current state
state, err := tm.GetState()

// View resources
for _, resource := range state.Resources {
    fmt.Printf("Resource: %s.%s\n", resource.Type, resource.Name)
}
```

## 🎯 Use Cases

### 1. Development Environment

Quick setup of development infrastructure:

```bash
# Deploy dev environment
terraform apply -var="environment=dev" -var="vm_size=Standard_B1s"

# Scale up for testing
terraform apply -var="aks_node_count=3"

# Clean up when done
terraform destroy
```

### 2. Learning Platform

Use AI assistance to learn Terraform:

1. **Generate code** for new resource types
2. **Explain existing** Terraform configurations
3. **Optimize code** for best practices
4. **Troubleshoot** deployment issues

### 3. Infrastructure Templates

Create reusable templates:

```bash
# Generate VM template
azure-tui terraform generate --type vm --name web-server

# Generate AKS template
azure-tui terraform generate --type aks --name dev-cluster

# Generate complete environment
azure-tui terraform generate --type full --name dev-env
```

## 🔐 Security Best Practices

### 1. SSH Key Management

```bash
# Generate SSH key if needed
ssh-keygen -t rsa -b 4096 -f ~/.ssh/azure_tui_key

# Update vm.tf to use your key
public_key = file("~/.ssh/azure_tui_key.pub")
```

### 2. Network Security

- **NSG rules** are restrictive by default
- **SSH access** limited to port 22
- **HTTP/HTTPS** enabled for web services
- **Private subnets** for AKS nodes

### 3. Resource Tagging

All resources include:

```hcl
tags = {
  Environment = var.environment
  Project     = var.project_name
  ManagedBy   = "terraform"
  CreatedBy   = "azure-tui-app"
}
```

## 📊 Cost Management

### Resource Sizing

| Resource | Size | Monthly Cost (approx.) |
|----------|------|----------------------|
| VM (B1s) | 1 vCPU, 1GB RAM | £8-12 |
| AKS (B2s) | 2 vCPU, 4GB RAM | £25-35 |
| Container Instances | 0.5 vCPU, 1.5GB | £10-15 |
| **Total** | | **£43-62** |

### Cost Optimization

```bash
# Use smaller VM for dev
terraform apply -var="vm_size=Standard_B1s"

# Reduce AKS nodes
terraform apply -var="aks_node_count=1"

# Enable auto-shutdown for VMs
# (Add to vm.tf custom data script)
```

## 🔄 CI/CD Integration

### GitHub Actions

```yaml
name: Terraform
on: [push]
jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: hashicorp/setup-terraform@v1
    - run: terraform init
    - run: terraform plan
    - run: terraform apply -auto-approve
      if: github.ref == 'refs/heads/main'
```

### Azure DevOps

```yaml
stages:
- stage: Plan
  jobs:
  - job: TerraformPlan
    steps:
    - task: TerraformInstaller@0
    - task: TerraformTaskV3@3
      inputs:
        command: init
    - task: TerraformTaskV3@3
      inputs:
        command: plan
```

## 🛠️ Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   ```bash
   # Check VM status
   az vm show --name azure-tui-vm --resource-group rg-azure-tui
   
   # Verify NSG rules
   az network nsg show --name azure-tui-vm-nsg --resource-group rg-azure-tui
   ```

2. **AKS Connection Failed**
   ```bash
   # Get credentials
   az aks get-credentials --name azure-tui-aks --resource-group rg-azure-tui
   
   # Check nodes
   kubectl get nodes
   ```

3. **Container Not Accessible**
   ```bash
   # Check container status
   az container show --name azure-tui-container-1 --resource-group rg-azure-tui
   ```

### AI-Powered Troubleshooting

Use Azure TUI's AI features to diagnose issues:

```go
// Analyze error messages
solution, err := ai.TroubleshootError(errorMessage, "Terraform deployment")

// Get optimization suggestions
optimization, err := ai.OptimizeTerraformCode(terraformCode, "debugging")
```

## 📚 Additional Resources

- [Terraform Azure Provider Documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs)
- [Azure Virtual Machine Sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes)
- [AKS Best Practices](https://docs.microsoft.com/en-us/azure/aks/best-practices)
- [Azure Container Instances](https://docs.microsoft.com/en-us/azure/container-instances/)

## 🎉 What's Next?

1. **Enhanced Templates**: More resource types and configurations
2. **Remote State**: Azure Storage backend integration
3. **Module Support**: Terraform module management
4. **Cost Estimation**: Real-time cost analysis
5. **Team Collaboration**: Multi-user workspace support

---

**Happy Infrastructure as Code with Azure TUI! 🚀**
