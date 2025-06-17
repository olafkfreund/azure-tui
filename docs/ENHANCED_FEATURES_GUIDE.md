# Azure TUI Enhanced Features Guide

## 🎯 Overview

This guide covers the new enhanced features added to Azure TUI, including table-formatted properties, SSH connectivity for VMs, and comprehensive AKS management.

## 📊 Table-Formatted Properties

### What's New
- Properties are now displayed in organized tables instead of simple lists
- Automatic formatting of property names (camelCase → Title Case)
- Intelligent value formatting for different data types
- Sorted display for consistent user experience

### How to Use
1. Navigate to any Azure resource in the left panel
2. Select the resource to view its details
3. Properties will automatically display in a formatted table with:
   - **Property** column showing formatted property names
   - **Value** column showing formatted values
   - Proper handling of complex objects, arrays, and null values

### Example Output
```
⚙️  Configuration Properties
Property                    │ Value
────────────────────────────┼─────────────────────
Admin Username              │ azureuser
Computer Name               │ myvm-001
OS Type                     │ Linux
Provisioning State          │ Succeeded
VM Size                     │ Standard_B2s
```

## 🔐 Enhanced SSH Functionality for VMs

### What's New
- Intelligent IP detection (public IP preferred, private IP fallback)
- Authentication method detection and display
- Enhanced error handling and user feedback
- Support for both direct SSH and Azure Bastion connections

### Keyboard Shortcuts
- **`[c]`** - SSH Connect: Direct SSH connection to the VM
- **`[b]`** - Bastion Connect: Connect via Azure Bastion

### How to Use
1. Navigate to a Virtual Machine in the resource tree
2. Select the VM to view its details
3. Use the keyboard shortcuts in the "Available Actions" section
4. The system will:
   - Automatically detect the VM's IP addresses
   - Choose the best connection method (public IP preferred)
   - Display connection details including IP, username, and auth method
   - Handle errors gracefully (e.g., VM without public IP)

### Example Output
```
🎮 Available Actions
[s] Start VM
[S] Stop VM
[r] Restart VM
[c] SSH Connect
[b] Bastion Connect

✅ SSH connection initiated to 40.112.123.45 (azureuser, SSH key)
```

## 🚢 Comprehensive AKS Management

### What's New
- Full kubectl integration with automatic credential management
- Real-time cluster information and management
- Pod, deployment, service, and node management
- Automatic authentication via `az aks get-credentials`

### Keyboard Shortcuts
- **`[s]`** - Start Cluster: Start the AKS cluster
- **`[S]`** - Stop Cluster: Stop the AKS cluster  
- **`[p]`** - List Pods: Show all pods across namespaces
- **`[D]`** - List Deployments: Show all deployments
- **`[n]`** - List Nodes: Show cluster nodes
- **`[v]`** - List Services: Show all services

### Features

#### Automatic Credential Management
```bash
# Automatically executed when connecting to AKS cluster
az aks get-credentials --resource-group <rg> --name <cluster> --overwrite-existing
```

#### Pod Management
- Lists all pods across all namespaces
- Shows pod status, namespace, and basic information
- Color-coded status indicators

#### Deployment Management
- Shows deployment status and replicas
- Indicates ready vs desired replica counts
- Namespace organization

#### Service Management
- Lists all services with types (ClusterIP, LoadBalancer, etc.)
- Shows external IPs for LoadBalancer services
- Port information display

#### Node Management
- Shows cluster nodes with status
- Node capacity and allocatable resources
- Kubernetes version information

### How to Use
1. Navigate to an AKS cluster in the resource tree
2. Select the cluster to view its details
3. Use the keyboard shortcuts in the "AKS Management Actions" section
4. The system will:
   - Automatically retrieve cluster credentials
   - Execute kubectl commands with proper context
   - Display results in a formatted, readable way
   - Handle errors and provide feedback

### Example Output
```
🚢 AKS Management Actions
[s] Start Cluster
[S] Stop Cluster
[p] List Pods
[D] List Deployments
[n] List Nodes
[v] List Services

✅ Retrieved credentials for cluster 'my-aks-cluster'
✅ Found 12 pods across 4 namespaces
```

## 🎮 Navigation and Usage

### General Navigation
- **`Tab`** - Switch between left (resource tree) and right (details) panels
- **`↑/↓`** - Navigate through resources
- **`Space/Enter`** - Expand resource groups or select resources
- **`[d]`** - Switch to dashboard view for selected resource
- **`[e]`** - Expand complex properties (like AKS agent pools)

### Panel Indicators
- **Left Panel**: Blue border when active, shows resource tree
- **Right Panel**: Green border when active, shows resource details
- **Active Panel Markers**: 🔍 for left panel, 📊 for right panel

### Status Feedback
- **⏳ Action in progress...** - Shown during long-running operations
- **✅ Success messages** - Green text for successful operations
- **❌ Error messages** - Red text for failed operations

## 🔧 Prerequisites

### Required Tools
- **Azure CLI** (`az`) - For Azure resource management
- **kubectl** - For AKS cluster management (optional but recommended)
- **SSH client** - For VM SSH connections

### Authentication
```bash
# Login to Azure
az login

# Verify authentication
az account show
```

### Permissions
Ensure your Azure account has appropriate permissions:
- **Reader** role minimum for viewing resources
- **Contributor** role for start/stop operations
- **Virtual Machine Contributor** for VM SSH operations
- **Azure Kubernetes Service Cluster User** role for AKS operations

## 🐛 Troubleshooting

### Common Issues

#### SSH Connection Failed
- **Cause**: VM doesn't have a public IP or NSG blocks SSH
- **Solution**: Use Bastion connection `[b]` or configure network access

#### kubectl Commands Fail
- **Cause**: kubectl not installed or cluster credentials not configured
- **Solution**: Install kubectl and ensure `az aks get-credentials` works

#### Permission Denied
- **Cause**: Insufficient Azure RBAC permissions
- **Solution**: Contact Azure administrator to grant appropriate roles

#### Resource Not Found
- **Cause**: Resource may have been deleted or moved
- **Solution**: Refresh with `[R]` or re-authenticate with Azure

## 💡 Tips and Best Practices

1. **Use Tab Navigation**: Efficiently switch between panels to browse and view details
2. **Monitor Status**: Watch for action progress indicators and results
3. **Resource Expansion**: Use `[e]` to expand complex properties for detailed information
4. **Dashboard View**: Use `[d]` for metric-focused views of resources
5. **Refresh Data**: Use `[R]` to reload resource information if it seems outdated

## 🚀 Advanced Usage

### Batch Operations
While individual resource operations are supported, for batch operations consider:
- Using Azure CLI scripts for multiple resources
- Leveraging Azure Resource Manager templates
- Using the TUI for monitoring and verification

### Integration with Other Tools
- Export resource configurations for Terraform/Bicep
- Use SSH connections for configuration management
- Combine with Azure Monitor for comprehensive monitoring

## 📝 Feedback and Contributions

The Azure TUI is designed to be extensible. Future enhancements might include:
- Interactive SSH sessions within the TUI
- Real-time log streaming for AKS pods
- Resource creation and modification capabilities
- Integration with Azure Monitor metrics
- Support for additional Azure resource types
