# Linux VM Terraform Template

This template creates a complete Linux VM infrastructure on Azure with security best practices.

## Resources Created

- **Resource Group**: Container for all resources
- **Virtual Network**: Isolated network with 10.0.0.0/16 address space
- **Subnet**: VM subnet with 10.0.1.0/24 address space
- **Public IP**: Static public IP for external access
- **Network Security Group**: Firewall rules (SSH access on port 22)
- **Network Interface**: VM network connection
- **Linux Virtual Machine**: Ubuntu 22.04 LTS with SSH key authentication
- **SSH Key Pair**: Auto-generated RSA 4096-bit key pair
- **VM Extension**: Optional Docker installation script

## Quick Start

1. **Initialize Terraform**:
   ```bash
   terraform init
   ```

2. **Plan the deployment**:
   ```bash
   terraform plan
   ```

3. **Apply the configuration**:
   ```bash
   terraform apply
   ```

4. **Connect via SSH**:
   ```bash
   # Save the private key (from terraform output)
   terraform output -raw ssh_private_key > private_key.pem
   chmod 600 private_key.pem
   
   # Connect to VM
   ssh -i private_key.pem azureuser@$(terraform output -raw public_ip_address)
   ```

## Customization

### Variables

| Variable | Description | Default | Type |
|----------|-------------|---------|------|
| `resource_group_name` | Resource group name | `rg-linux-vm` | string |
| `location` | Azure region | `East US` | string |
| `vm_name` | Virtual machine name | `vm-linux-01` | string |
| `vm_size` | VM size | `Standard_B2s` | string |
| `admin_username` | Admin username | `azureuser` | string |
| `os_disk_type` | OS disk type | `Standard_LRS` | string |
| `install_docker` | Install Docker | `false` | bool |
| `tags` | Resource tags | See variables.tf | map |

### Example terraform.tfvars

```hcl
resource_group_name = "rg-my-linux-vm"
location           = "West Europe"
vm_name            = "vm-web-server"
vm_size            = "Standard_D2s_v3"
admin_username     = "myadmin"
os_disk_type       = "Premium_LRS"
install_docker     = true

tags = {
  Environment = "Production"
  Project     = "WebServer"
  Owner       = "DevOps Team"
}
```

## Security Features

- **SSH Key Authentication**: Password authentication disabled
- **Network Security Group**: Restricts inbound traffic to SSH only
- **Latest OS Image**: Uses Ubuntu 22.04 LTS latest image
- **Encrypted Storage**: OS disk encryption available (configure in main.tf)

## Outputs

- `public_ip_address`: VM public IP
- `private_ip_address`: VM private IP
- `ssh_connection_command`: Ready-to-use SSH command
- `ssh_private_key`: Private key for SSH (sensitive)
- All resource IDs for integration with other templates

## Extensions

The template supports optional VM extensions:

### Docker Installation
Set `install_docker = true` to automatically install:
- Docker Engine
- Docker Compose
- User added to docker group

### Custom Scripts
Modify `scripts/install-docker.sh` or add new scripts for:
- Software installation
- Configuration management
- Monitoring agents
- Security hardening

## Network Configuration

Default network configuration:
- **VNet**: 10.0.0.0/16
- **Subnet**: 10.0.1.0/24
- **NSG**: SSH (22) from any source

To restrict SSH access, modify the NSG rule in `main.tf`:
```hcl
source_address_prefix = "YOUR_IP_ADDRESS/32"
```

## Cost Optimization

For development/testing:
- Use `Standard_B1s` or `Standard_B2s` for lower costs
- Use `Standard_LRS` storage
- Consider spot instances for non-critical workloads

For production:
- Use `Standard_D2s_v3` or larger
- Use `Premium_LRS` for better performance
- Enable monitoring and backup

## Troubleshooting

### Common Issues

1. **SSH Connection Refused**:
   - Check NSG rules
   - Verify VM is running
   - Confirm public IP is assigned

2. **Terraform Apply Fails**:
   - Verify Azure CLI authentication
   - Check resource quotas
   - Ensure unique resource names

3. **VM Extension Fails**:
   - Check custom script syntax
   - Verify script permissions
   - Review VM extension logs

### Debugging

Enable Terraform debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

Check VM boot diagnostics in Azure portal for startup issues.

## Integration with Azure-TUI

This template integrates with azure-tui for:
- **AI Code Generation**: Modify templates with AI assistance
- **Resource Monitoring**: View deployed resources in TUI
- **State Management**: Track and manage Terraform state
- **Cost Analysis**: Monitor resource costs
- **Security Scanning**: Validate security configurations

Use azure-tui keyboard shortcuts:
- `T`: Generate/modify Terraform code
- `Enter`: Deploy selected template
- `Ctrl+E`: Edit in external editor
- `Ctrl+P`: Plan changes
- `Ctrl+A`: Apply changes
