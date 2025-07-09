package resourceactions

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/olafkfreund/azure-tui/internal/bicep"
	"github.com/olafkfreund/azure-tui/internal/ssh"
)

// ActionResult represents the result of a resource action
type ActionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// StartVM starts a virtual machine
func StartVM(vmName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "vm", "start", "--name", vmName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start VM: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("VM '%s' started successfully", vmName),
		Output:  string(output),
	}
}

// StopVM stops a virtual machine
func StopVM(vmName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "vm", "deallocate", "--name", vmName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to stop VM: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("VM '%s' stopped successfully", vmName),
		Output:  string(output),
	}
}

// RestartVM restarts a virtual machine
func RestartVM(vmName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "vm", "restart", "--name", vmName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to restart VM: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("VM '%s' restarted successfully", vmName),
		Output:  string(output),
	}
}

// GetVMStatus gets the current status of a virtual machine
func GetVMStatus(vmName, resourceGroup string) (string, error) {
	cmd := exec.Command("az", "vm", "get-instance-view",
		"--name", vmName,
		"--resource-group", resourceGroup,
		"--query", "instanceView.statuses[1].displayStatus",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ConnectVMSSH attempts to connect to a VM via SSH
func ConnectVMSSH(vmName, resourceGroup, username string) ActionResult {
	// First get the VM's public IP
	cmd := exec.Command("az", "vm", "list-ip-addresses",
		"--name", vmName,
		"--resource-group", resourceGroup,
		"--query", "[0].virtualMachine.network.publicIpAddresses[0].ipAddress",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return ActionResult{
			Success: false,
			Message: "Failed to get VM public IP address",
			Output:  string(output),
		}
	}

	publicIP := strings.TrimSpace(string(output))
	if publicIP == "" {
		return ActionResult{
			Success: false,
			Message: "VM does not have a public IP address",
			Output:  "",
		}
	}

	// Create SSH command
	sshCmd := fmt.Sprintf("ssh %s@%s", username, publicIP)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("SSH command ready for VM '%s'", vmName),
		Output:  sshCmd,
	}
}

// ExecuteVMSSH executes SSH connection to a VM
func ExecuteVMSSH(vmName, resourceGroup, username string) ActionResult {
	// First get the VM's public IP
	cmd := exec.Command("az", "vm", "list-ip-addresses",
		"--name", vmName,
		"--resource-group", resourceGroup,
		"--query", "[0].virtualMachine.network.publicIpAddresses[0].ipAddress",
		"--output", "tsv")

	output, err := cmd.Output()
	if err != nil {
		return ActionResult{
			Success: false,
			Message: "Failed to get VM public IP address",
			Output:  string(output),
		}
	}

	publicIP := strings.TrimSpace(string(output))
	if publicIP == "" {
		return ActionResult{
			Success: false,
			Message: "VM does not have a public IP address. Consider using Azure Bastion instead.",
			Output:  "",
		}
	}

	// Check if SSH key authentication is available
	var keyAuthEnabled bool
	sshKeyCmd := exec.Command("az", "vm", "show",
		"--name", vmName,
		"--resource-group", resourceGroup,
		"--query", "osProfile.linuxConfiguration.disablePasswordAuthentication",
		"--output", "tsv")

	keyAuthOutput, err := sshKeyCmd.Output()
	if err != nil {
		// If we can't determine key auth status, assume password auth
		keyAuthEnabled = false
	} else {
		keyAuthEnabled = strings.TrimSpace(string(keyAuthOutput)) == "true"
	}

	// Prepare SSH command with appropriate options
	var sshCmd *exec.Cmd
	if keyAuthEnabled {
		// Use SSH key authentication
		sshCmd = exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", fmt.Sprintf("%s@%s", username, publicIP))
	} else {
		// Use password authentication
		sshCmd = exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "PasswordAuthentication=yes", fmt.Sprintf("%s@%s", username, publicIP))
	}

	// Return the command for execution in a terminal
	cmdStr := strings.Join(sshCmd.Args, " ")

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("SSH connection ready for VM '%s' at %s", vmName, publicIP),
		Output: fmt.Sprintf("Execute: %s\n\nConnection Details:\n- VM: %s\n- IP: %s\n- User: %s\n- Auth: %s",
			cmdStr, vmName, publicIP, username,
			map[bool]string{true: "SSH Key", false: "Password"}[keyAuthEnabled]),
	}
}

// ConnectVMBastion connects to a VM via Azure Bastion
func ConnectVMBastion(vmName, resourceGroup string) ActionResult {
	// Check if VM has Bastion available
	cmd := exec.Command("az", "network", "bastion", "list",
		"--resource-group", resourceGroup,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return ActionResult{
			Success: false,
			Message: "Failed to check Bastion availability",
			Output:  string(output),
		}
	}

	// If Bastion is available, create tunnel command
	bastionCmd := fmt.Sprintf("az network bastion tunnel --name <bastion-name> --resource-group %s --target-resource-id <vm-resource-id> --resource-port 22 --port 2222", resourceGroup)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Bastion command ready for VM '%s'", vmName),
		Output:  bastionCmd,
	}
}

// StartWebApp starts an Azure Web App
func StartWebApp(appName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "webapp", "start", "--name", appName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start Web App: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Web App '%s' started successfully", appName),
		Output:  string(output),
	}
}

// StopWebApp stops an Azure Web App
func StopWebApp(appName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "webapp", "stop", "--name", appName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to stop Web App: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Web App '%s' stopped successfully", appName),
		Output:  string(output),
	}
}

// RestartWebApp restarts an Azure Web App
func RestartWebApp(appName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "webapp", "restart", "--name", appName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to restart Web App: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Web App '%s' restarted successfully", appName),
		Output:  string(output),
	}
}

// StartAKSCluster starts an AKS cluster
func StartAKSCluster(clusterName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "aks", "start", "--name", clusterName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start AKS cluster: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("AKS cluster '%s' started successfully", clusterName),
		Output:  string(output),
	}
}

// StopAKSCluster stops an AKS cluster
func StopAKSCluster(clusterName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "aks", "stop", "--name", clusterName, "--resource-group", resourceGroup)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to stop AKS cluster: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("AKS cluster '%s' stopped successfully", clusterName),
		Output:  string(output),
	}
}

// ScaleAKSCluster scales an AKS cluster node pool
func ScaleAKSCluster(clusterName, resourceGroup string, nodeCount int) ActionResult {
	cmd := exec.Command("az", "aks", "scale",
		"--name", clusterName,
		"--resource-group", resourceGroup,
		"--node-count", fmt.Sprintf("%d", nodeCount))

	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to scale AKS cluster: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("AKS cluster '%s' scaled to %d nodes", clusterName, nodeCount),
		Output:  string(output),
	}
}

// AKS kubectl management functions

// ConnectAKSCluster gets credentials and connects to AKS cluster
func ConnectAKSCluster(clusterName, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "aks", "get-credentials",
		"--name", clusterName,
		"--resource-group", resourceGroup,
		"--overwrite-existing")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get AKS credentials: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Successfully connected to AKS cluster '%s'", clusterName),
		Output:  string(output),
	}
}

// ListAKSPods lists all pods in the AKS cluster
func ListAKSPods(clusterName, resourceGroup string) ActionResult {
	// First get credentials
	connectResult := ConnectAKSCluster(clusterName, resourceGroup)
	if !connectResult.Success {
		return connectResult
	}

	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces", "-o", "wide")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list pods: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Pods in cluster '%s'", clusterName),
		Output:  string(output),
	}
}

// ListAKSDeployments lists all deployments in the AKS cluster
func ListAKSDeployments(clusterName, resourceGroup string) ActionResult {
	// First get credentials
	connectResult := ConnectAKSCluster(clusterName, resourceGroup)
	if !connectResult.Success {
		return connectResult
	}

	cmd := exec.Command("kubectl", "get", "deployments", "--all-namespaces", "-o", "wide")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list deployments: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Deployments in cluster '%s'", clusterName),
		Output:  string(output),
	}
}

// ListAKSServices lists all services in the AKS cluster
func ListAKSServices(clusterName, resourceGroup string) ActionResult {
	// First get credentials
	connectResult := ConnectAKSCluster(clusterName, resourceGroup)
	if !connectResult.Success {
		return connectResult
	}

	cmd := exec.Command("kubectl", "get", "services", "--all-namespaces", "-o", "wide")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list services: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Services in cluster '%s'", clusterName),
		Output:  string(output),
	}
}

// ShowAKSLogs shows logs for a specific pod
func ShowAKSLogs(clusterName, resourceGroup, podName, namespace string) ActionResult {
	// First get credentials
	connectResult := ConnectAKSCluster(clusterName, resourceGroup)
	if !connectResult.Success {
		return connectResult
	}

	var cmd *exec.Cmd
	if namespace != "" {
		cmd = exec.Command("kubectl", "logs", podName, "-n", namespace, "--tail=100")
	} else {
		cmd = exec.Command("kubectl", "logs", podName, "--tail=100")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get logs for pod '%s': %v", podName, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Logs for pod '%s'", podName),
		Output:  string(output),
	}
}

// GetAKSNodes lists all nodes in the AKS cluster
func GetAKSNodes(clusterName, resourceGroup string) ActionResult {
	// First get credentials
	connectResult := ConnectAKSCluster(clusterName, resourceGroup)
	if !connectResult.Success {
		return connectResult
	}

	cmd := exec.Command("kubectl", "get", "nodes", "-o", "wide")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list nodes: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Nodes in cluster '%s'", clusterName),
		Output:  string(output),
	}
}

// GetResourceActions returns available actions for a resource type
func GetResourceActions(resourceType string) []string {
	var actions []string

	switch {
	case strings.Contains(resourceType, "Microsoft.Compute/virtualMachines"):
		actions = []string{"start", "stop", "restart", "ssh", "bastion", "status"}
	case strings.Contains(resourceType, "Microsoft.Web/sites"):
		actions = []string{"start", "stop", "restart", "browse", "logs"}
	case strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters"):
		actions = []string{"start", "stop", "scale", "connect", "pods", "deployments", "services", "logs", "nodes"}
	case strings.Contains(resourceType, "Microsoft.Sql/servers"):
		actions = []string{"backup", "scale", "connect", "security"}
	case strings.Contains(resourceType, "Microsoft.Storage/storageAccounts"):
		actions = []string{"browse", "keys", "backup", "metrics"}
	case strings.Contains(resourceType, "Microsoft.Network/virtualNetworks"):
		actions = []string{"create", "delete", "peering"}
	case strings.Contains(resourceType, "Microsoft.Network/networkSecurityGroups"):
		actions = []string{"create", "delete", "rule", "associate"}
	case strings.Contains(resourceType, "Microsoft.Network/routeTables"):
		actions = []string{"create", "delete", "route"}
	case strings.Contains(resourceType, "Microsoft.Network/publicIPAddresses"):
		actions = []string{"create", "delete"}
	case strings.Contains(resourceType, "Microsoft.Network/loadBalancers"):
		actions = []string{"create", "delete"}
	case strings.Contains(resourceType, "Microsoft.Network/networkInterfaces"):
		actions = []string{"create", "delete"}
	case strings.Contains(resourceType, "Microsoft.Network/networkWatchers"):
		actions = []string{"enable", "test-connectivity"}
	default:
		actions = []string{"view", "edit", "delete"}
	}

	return actions
}

// ExecuteResourceAction executes a specific action on a resource
func ExecuteResourceAction(action, resourceType, resourceName, resourceGroup string, params map[string]interface{}) ActionResult {
	switch action {
	case "start":
		if strings.Contains(resourceType, "Microsoft.Compute/virtualMachines") {
			return StartVM(resourceName, resourceGroup)
		} else if strings.Contains(resourceType, "Microsoft.Web/sites") {
			return StartWebApp(resourceName, resourceGroup)
		} else if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return StartAKSCluster(resourceName, resourceGroup)
		}
	case "stop":
		if strings.Contains(resourceType, "Microsoft.Compute/virtualMachines") {
			return StopVM(resourceName, resourceGroup)
		} else if strings.Contains(resourceType, "Microsoft.Web/sites") {
			return StopWebApp(resourceName, resourceGroup)
		} else if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return StopAKSCluster(resourceName, resourceGroup)
		}
	case "restart":
		if strings.Contains(resourceType, "Microsoft.Compute/virtualMachines") {
			return RestartVM(resourceName, resourceGroup)
		} else if strings.Contains(resourceType, "Microsoft.Web/sites") {
			return RestartWebApp(resourceName, resourceGroup)
		}
	case "ssh":
		if strings.Contains(resourceType, "Microsoft.Compute/virtualMachines") {
			username := "azureuser" // Default username
			if u, ok := params["username"].(string); ok {
				username = u
			}
			return ExecuteVMSSH(resourceName, resourceGroup, username)
		}
	case "bastion":
		if strings.Contains(resourceType, "Microsoft.Compute/virtualMachines") {
			return ConnectVMBastion(resourceName, resourceGroup)
		}
	case "scale":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			nodeCount := 3 // Default
			if n, ok := params["nodeCount"].(int); ok {
				nodeCount = n
			}
			return ScaleAKSCluster(resourceName, resourceGroup, nodeCount)
		}
	case "connect":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return ConnectAKSCluster(resourceName, resourceGroup)
		}
	case "pods":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return ListAKSPods(resourceName, resourceGroup)
		}
	case "deployments":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return ListAKSDeployments(resourceName, resourceGroup)
		}
	case "services":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return ListAKSServices(resourceName, resourceGroup)
		}
	case "logs":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			podName := ""
			namespace := ""
			if p, ok := params["podName"].(string); ok {
				podName = p
			}
			if n, ok := params["namespace"].(string); ok {
				namespace = n
			}
			return ShowAKSLogs(resourceName, resourceGroup, podName, namespace)
		}
	case "nodes":
		if strings.Contains(resourceType, "Microsoft.ContainerService/managedClusters") {
			return GetAKSNodes(resourceName, resourceGroup)
		}
	case "create":
		if strings.Contains(resourceType, "Microsoft.Network/virtualNetworks") {
			location := params["location"].(string)
			addressPrefixes := params["addressPrefixes"].([]string)
			return CreateVirtualNetworkAction(resourceName, resourceGroup, location, addressPrefixes)
		} else if strings.Contains(resourceType, "Microsoft.Network/networkSecurityGroups") {
			location := params["location"].(string)
			return CreateNetworkSecurityGroupAction(resourceName, resourceGroup, location)
		} else if strings.Contains(resourceType, "Microsoft.Network/routeTables") {
			location := params["location"].(string)
			return CreateRouteTableAction(resourceName, resourceGroup, location)
		} else if strings.Contains(resourceType, "Microsoft.Network/publicIPAddresses") {
			location := params["location"].(string)
			allocationMethod := params["allocationMethod"].(string)
			sku := params["sku"].(string)
			return CreatePublicIPAction(resourceName, resourceGroup, location, allocationMethod, sku)
		} else if strings.Contains(resourceType, "Microsoft.Network/loadBalancers") {
			location := params["location"].(string)
			sku := params["sku"].(string)
			publicIPName := params["publicIPName"].(string)
			return CreateLoadBalancerAction(resourceName, resourceGroup, location, sku, publicIPName)
		} else if strings.Contains(resourceType, "Microsoft.Network/networkInterfaces") {
			location := params["location"].(string)
			subnetID := params["subnetID"].(string)
			publicIPName := params["publicIPName"].(string)
			nsgName := params["nsgName"].(string)
			return CreateNetworkInterfaceAction(resourceName, resourceGroup, location, subnetID, publicIPName, nsgName)
		}
	case "delete":
		if strings.Contains(resourceType, "Microsoft.Network/virtualNetworks") {
			return DeleteVirtualNetworkAction(resourceName, resourceGroup)
		}
	case "rule":
		if strings.Contains(resourceType, "Microsoft.Network/networkSecurityGroups") {
			ruleName := params["ruleName"].(string)
			priority := params["priority"].(int)
			direction := params["direction"].(string)
			access := params["access"].(string)
			protocol := params["protocol"].(string)
			sourcePort := params["sourcePort"].(string)
			destPort := params["destPort"].(string)
			sourceAddress := params["sourceAddress"].(string)
			destAddress := params["destAddress"].(string)
			return AddSecurityRuleAction(resourceName, resourceGroup, ruleName, priority, direction, access, protocol, sourcePort, destPort, sourceAddress, destAddress)
		}
	case "associate":
		if strings.Contains(resourceType, "Microsoft.Network/networkSecurityGroups") {
			subnetName := params["subnetName"].(string)
			vnetName := params["vnetName"].(string)
			nsgName := params["nsgName"].(string)
			return AssociateNSGWithSubnetAction(subnetName, vnetName, resourceGroup, nsgName)
		}
	case "route":
		if strings.Contains(resourceType, "Microsoft.Network/routeTables") {
			routeName := params["routeName"].(string)
			addressPrefix := params["addressPrefix"].(string)
			nextHopType := params["nextHopType"].(string)
			nextHopAddress := params["nextHopAddress"].(string)
			return AddRouteAction(resourceName, resourceGroup, routeName, addressPrefix, nextHopType, nextHopAddress)
		}
	case "enable":
		if strings.Contains(resourceType, "Microsoft.Network/networkWatchers") {
			location := params["location"].(string)
			return EnableNetworkWatcherAction(resourceGroup, location)
		}
	case "test-connectivity":
		if strings.Contains(resourceType, "Microsoft.Network/networkWatchers") {
			sourceResourceID := params["sourceResourceID"].(string)
			destResourceID := params["destResourceID"].(string)
			return TestNetworkConnectivityAction(sourceResourceID, destResourceID)
		}
	case "peering":
		if strings.Contains(resourceType, "Microsoft.Network/virtualNetworks") {
			localVNet := params["localVNet"].(string)
			localResourceGroup := params["localResourceGroup"].(string)
			remoteVNet := params["remoteVNet"].(string)
			remoteResourceGroup := params["remoteResourceGroup"].(string)
			return CreateVNetPeeringAction(localVNet, localResourceGroup, remoteVNet, remoteResourceGroup)
		}
	}

	return ActionResult{
		Success: false,
		Message: fmt.Sprintf("Action '%s' not supported for resource type '%s'", action, resourceType),
		Output:  "",
	}
}

// =============================================================================
// NETWORK RESOURCE ACTIONS
// =============================================================================

// CreateVirtualNetworkAction creates a new virtual network
func CreateVirtualNetworkAction(name, resourceGroup, location string, addressPrefixes []string) ActionResult {
	args := []string{"network", "vnet", "create", "--name", name, "--resource-group", resourceGroup, "--location", location}

	if len(addressPrefixes) > 0 {
		args = append(args, "--address-prefixes")
		args = append(args, addressPrefixes...)
	} else {
		args = append(args, "--address-prefix", "10.0.0.0/16")
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create VNet '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Virtual Network '%s' created successfully", name),
		Output:  string(output),
	}
}

// DeleteVirtualNetworkAction deletes a virtual network
func DeleteVirtualNetworkAction(name, resourceGroup string) ActionResult {
	cmd := exec.Command("az", "network", "vnet", "delete", "--name", name, "--resource-group", resourceGroup, "--yes")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to delete VNet '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Virtual Network '%s' deleted successfully", name),
		Output:  string(output),
	}
}

// CreateSubnetAction creates a new subnet in a virtual network
func CreateSubnetAction(name, vnetName, resourceGroup, addressPrefix string) ActionResult {
	cmd := exec.Command("az", "network", "vnet", "subnet", "create",
		"--name", name,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--address-prefix", addressPrefix)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create subnet '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Subnet '%s' created successfully in VNet '%s'", name, vnetName),
		Output:  string(output),
	}
}

// CreateNetworkSecurityGroupAction creates a new network security group
func CreateNetworkSecurityGroupAction(name, resourceGroup, location string) ActionResult {
	cmd := exec.Command("az", "network", "nsg", "create", "--name", name, "--resource-group", resourceGroup, "--location", location)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create NSG '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Network Security Group '%s' created successfully", name),
		Output:  string(output),
	}
}

// AddSecurityRuleAction adds a new security rule to an NSG
func AddSecurityRuleAction(nsgName, resourceGroup, ruleName string, priority int, direction, access, protocol, sourcePort, destPort, sourceAddress, destAddress string) ActionResult {
	cmd := exec.Command("az", "network", "nsg", "rule", "create",
		"--nsg-name", nsgName,
		"--resource-group", resourceGroup,
		"--name", ruleName,
		"--priority", fmt.Sprintf("%d", priority),
		"--direction", direction,
		"--access", access,
		"--protocol", protocol,
		"--source-port-ranges", sourcePort,
		"--destination-port-ranges", destPort,
		"--source-address-prefixes", sourceAddress,
		"--destination-address-prefixes", destAddress)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to add security rule '%s': %v", ruleName, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Security rule '%s' added to NSG '%s' successfully", ruleName, nsgName),
		Output:  string(output),
	}
}

// AssociateNSGWithSubnetAction associates an NSG with a subnet
func AssociateNSGWithSubnetAction(subnetName, vnetName, resourceGroup, nsgName string) ActionResult {
	nsgID := fmt.Sprintf("/subscriptions/$(az account show --query id -o tsv)/resourceGroups/%s/providers/Microsoft.Network/networkSecurityGroups/%s", resourceGroup, nsgName)
	cmd := exec.Command("az", "network", "vnet", "subnet", "update",
		"--name", subnetName,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--network-security-group", nsgID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to associate NSG '%s' with subnet '%s': %v", nsgName, subnetName, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("NSG '%s' associated with subnet '%s' successfully", nsgName, subnetName),
		Output:  string(output),
	}
}

// CreateRouteTableAction creates a new route table
func CreateRouteTableAction(name, resourceGroup, location string) ActionResult {
	cmd := exec.Command("az", "network", "route-table", "create", "--name", name, "--resource-group", resourceGroup, "--location", location)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create route table '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Route table '%s' created successfully", name),
		Output:  string(output),
	}
}

// AddRouteAction adds a new route to a route table
func AddRouteAction(routeTableName, resourceGroup, routeName, addressPrefix, nextHopType, nextHopAddress string) ActionResult {
	args := []string{"network", "route-table", "route", "create",
		"--route-table-name", routeTableName,
		"--resource-group", resourceGroup,
		"--name", routeName,
		"--address-prefix", addressPrefix,
		"--next-hop-type", nextHopType}

	if nextHopAddress != "" {
		args = append(args, "--next-hop-ip-address", nextHopAddress)
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to add route '%s': %v", routeName, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Route '%s' added to route table '%s' successfully", routeName, routeTableName),
		Output:  string(output),
	}
}

// CreatePublicIPAction creates a new public IP address
func CreatePublicIPAction(name, resourceGroup, location, allocationMethod, sku string) ActionResult {
	cmd := exec.Command("az", "network", "public-ip", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--allocation-method", allocationMethod,
		"--sku", sku)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create public IP '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Public IP '%s' created successfully", name),
		Output:  string(output),
	}
}

// CreateLoadBalancerAction creates a new load balancer
func CreateLoadBalancerAction(name, resourceGroup, location, sku, publicIPName string) ActionResult {
	args := []string{"network", "lb", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--sku", sku}

	if publicIPName != "" {
		args = append(args, "--public-ip-address", publicIPName)
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create load balancer '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Load balancer '%s' created successfully", name),
		Output:  string(output),
	}
}

// CreateNetworkInterfaceAction creates a new network interface
func CreateNetworkInterfaceAction(name, resourceGroup, location, subnetID, publicIPName, nsgName string) ActionResult {
	args := []string{"network", "nic", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--subnet", subnetID}

	if publicIPName != "" {
		args = append(args, "--public-ip-address", publicIPName)
	}

	if nsgName != "" {
		args = append(args, "--network-security-group", nsgName)
	}

	cmd := exec.Command("az", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create network interface '%s': %v", name, err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Network interface '%s' created successfully", name),
		Output:  string(output),
	}
}

// EnableNetworkWatcherAction enables Network Watcher for monitoring
func EnableNetworkWatcherAction(resourceGroup, location string) ActionResult {
	cmd := exec.Command("az", "network", "watcher", "configure", "--resource-group", resourceGroup, "--locations", location, "--enabled", "true")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to enable Network Watcher: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: "Network Watcher enabled successfully",
		Output:  string(output),
	}
}

// TestNetworkConnectivityAction tests connectivity between network resources
func TestNetworkConnectivityAction(sourceResourceID, destResourceID string) ActionResult {
	cmd := exec.Command("az", "network", "watcher", "test-connectivity",
		"--source-resource", sourceResourceID,
		"--dest-resource", destResourceID)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to test connectivity: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: "Network connectivity test completed",
		Output:  string(output),
	}
}

// CreateVNetPeeringAction creates VNet peering between two virtual networks
func CreateVNetPeeringAction(localVNet, localResourceGroup, remoteVNet, remoteResourceGroup string) ActionResult {
	remoteVNetID := fmt.Sprintf("/subscriptions/$(az account show --query id -o tsv)/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s", remoteResourceGroup, remoteVNet)

	cmd := exec.Command("az", "network", "vnet", "peering", "create",
		"--name", fmt.Sprintf("%s-to-%s", localVNet, remoteVNet),
		"--vnet-name", localVNet,
		"--resource-group", localResourceGroup,
		"--remote-vnet", remoteVNetID,
		"--allow-vnet-access")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create VNet peering: %v", err),
			Output:  string(output),
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("VNet peering created between '%s' and '%s'", localVNet, remoteVNet),
		Output:  string(output),
	}
}

// =============================================================================
// ENHANCED SSH ACTIONS
// =============================================================================

// SSHManager instance for enhanced SSH operations
var sshManager *ssh.SSHManager

// BicepManager instance for Bicep operations
var bicepManager *bicep.BicepManager

// InitializeSSHManager initializes the SSH manager with Azure credentials
func InitializeSSHManager(cred *azidentity.DefaultAzureCredential, subscriptionID string) {
	sshManager = ssh.NewSSHManagerWithVM(cred, subscriptionID)
}

// InitializeBicepManager initializes the Bicep manager
func InitializeBicepManager() {
	bicepManager = bicep.NewBicepManager()
}

// ConnectVMEnhancedSSH connects to a VM using enhanced SSH with automatic IP resolution
func ConnectVMEnhancedSSH(vmName, resourceGroup string) ActionResult {
	if sshManager == nil {
		return ActionResult{
			Success: false,
			Message: "SSH manager not initialized. Call InitializeSSHManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get VM info for connection details
	vmInfo, err := sshManager.GetVMInfo(ctx, resourceGroup, vmName)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get VM information: %v", err),
			Output:  "",
		}
	}

	// Build connection info string
	connectionInfo := fmt.Sprintf("VM Connection Information:\n")
	connectionInfo += fmt.Sprintf("- Name: %s\n", vmInfo.Name)
	connectionInfo += fmt.Sprintf("- Resource Group: %s\n", vmInfo.ResourceGroup)
	connectionInfo += fmt.Sprintf("- OS Type: %s\n", vmInfo.OSType)
	connectionInfo += fmt.Sprintf("- Power State: %s\n", vmInfo.PowerState)
	connectionInfo += fmt.Sprintf("- Admin Username: %s\n", vmInfo.AdminUsername)

	if vmInfo.PublicIPAddress != "" {
		connectionInfo += fmt.Sprintf("- Public IP: %s\n", vmInfo.PublicIPAddress)
	}
	if vmInfo.PrivateIPAddress != "" {
		connectionInfo += fmt.Sprintf("- Private IP: %s\n", vmInfo.PrivateIPAddress)
	}
	if vmInfo.FQDN != "" {
		connectionInfo += fmt.Sprintf("- FQDN: %s\n", vmInfo.FQDN)
	}

	availableKeys := sshManager.GetAvailableKeys()
	if len(availableKeys) > 0 {
		connectionInfo += fmt.Sprintf("- Available SSH Keys: %d found\n", len(availableKeys))
		for i, key := range availableKeys {
			if i < 3 { // Show first 3 keys
				connectionInfo += fmt.Sprintf("  - %s\n", key)
			}
		}
		if len(availableKeys) > 3 {
			connectionInfo += fmt.Sprintf("  - ... and %d more\n", len(availableKeys)-3)
		}
	} else {
		connectionInfo += "- No SSH keys found in standard locations\n"
	}

	// Check if VM can be reached
	if vmInfo.PowerState != "running" {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("VM '%s' is not running (current state: %s). Start the VM first.", vmName, vmInfo.PowerState),
			Output:  connectionInfo,
		}
	}

	if vmInfo.PublicIPAddress == "" && vmInfo.PrivateIPAddress == "" && vmInfo.FQDN == "" {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("VM '%s' has no accessible IP addresses. Check network configuration or use Azure Bastion.", vmName),
			Output:  connectionInfo,
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Ready to connect to VM '%s'", vmName),
		Output:  connectionInfo,
	}
}

// ExecuteVMEnhancedSSH executes an SSH command on a VM using enhanced SSH
func ExecuteVMEnhancedSSH(vmName, resourceGroup, command string) ActionResult {
	if sshManager == nil {
		return ActionResult{
			Success: false,
			Message: "SSH manager not initialized. Call InitializeSSHManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	output, err := sshManager.ExecuteVMCommand(ctx, resourceGroup, vmName, command)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to execute SSH command on VM '%s': %v", vmName, err),
			Output:  output,
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Command executed successfully on VM '%s'", vmName),
		Output:  output,
	}
}

// StartInteractiveVMSSH starts an interactive SSH session with a VM
func StartInteractiveVMSSH(vmName, resourceGroup string) ActionResult {
	if sshManager == nil {
		return ActionResult{
			Success: false,
			Message: "SSH manager not initialized. Call InitializeSSHManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes for interactive session
	defer cancel()

	err := sshManager.StartInteractiveVMSession(ctx, resourceGroup, vmName)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start interactive SSH session with VM '%s': %v", vmName, err),
			Output:  "",
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Interactive SSH session with VM '%s' completed", vmName),
		Output:  "",
	}
}

// TestVMSSHConnection tests SSH connectivity to a VM without executing commands
func TestVMSSHConnection(vmName, resourceGroup string) ActionResult {
	if sshManager == nil {
		return ActionResult{
			Success: false,
			Message: "SSH manager not initialized. Call InitializeSSHManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to establish connection
	conn, err := sshManager.ConnectToAzureVM(ctx, resourceGroup, vmName)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("SSH connection test failed for VM '%s': %v", vmName, err),
			Output:  "",
		}
	}

	// Test with a simple command
	output, err := sshManager.ExecuteVMCommand(ctx, resourceGroup, vmName, "echo 'SSH connection successful'")
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("SSH connection established but command execution failed for VM '%s': %v", vmName, err),
			Output:  output,
		}
	}

	connectionInfo := fmt.Sprintf("SSH Connection Test Results:\n")
	connectionInfo += fmt.Sprintf("- VM: %s\n", vmName)
	connectionInfo += fmt.Sprintf("- Host: %s\n", conn.Host)
	connectionInfo += fmt.Sprintf("- User: %s\n", conn.User)
	connectionInfo += fmt.Sprintf("- Key Path: %s\n", conn.KeyPath)
	connectionInfo += fmt.Sprintf("- Connected At: %s\n", conn.ConnTime.Format("2006-01-02 15:04:05"))
	connectionInfo += fmt.Sprintf("- Test Command Output: %s", strings.TrimSpace(output))

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("SSH connection test successful for VM '%s'", vmName),
		Output:  connectionInfo,
	}
}

// GetVMSSHConnectionInfo retrieves SSH connection information for a VM
func GetVMSSHConnectionInfo(vmName, resourceGroup string) ActionResult {
	if sshManager == nil {
		return ActionResult{
			Success: false,
			Message: "SSH manager not initialized. Call InitializeSSHManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vmInfo, err := sshManager.GetVMInfo(ctx, resourceGroup, vmName)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get VM SSH connection info: %v", err),
			Output:  "",
		}
	}

	connectionInfo := fmt.Sprintf("SSH Connection Guide for VM '%s':\n\n", vmName)
	
	// Connection options
	if vmInfo.PublicIPAddress != "" {
		connectionInfo += fmt.Sprintf("Option 1 - Direct SSH (Public IP):\n")
		connectionInfo += fmt.Sprintf("  ssh %s@%s\n\n", vmInfo.AdminUsername, vmInfo.PublicIPAddress)
	}
	
	if vmInfo.FQDN != "" {
		connectionInfo += fmt.Sprintf("Option 2 - SSH via FQDN:\n")
		connectionInfo += fmt.Sprintf("  ssh %s@%s\n\n", vmInfo.AdminUsername, vmInfo.FQDN)
	}
	
	if vmInfo.PrivateIPAddress != "" {
		connectionInfo += fmt.Sprintf("Option 3 - SSH via Private IP (requires VPN/Bastion):\n")
		connectionInfo += fmt.Sprintf("  ssh %s@%s\n\n", vmInfo.AdminUsername, vmInfo.PrivateIPAddress)
	}
	
	connectionInfo += fmt.Sprintf("Azure Bastion Option:\n")
	connectionInfo += fmt.Sprintf("  az network bastion ssh --resource-group %s --name <bastion-name> --target-resource-id <vm-resource-id> --auth-type ssh-key --username %s\n\n", resourceGroup, vmInfo.AdminUsername)
	
	// SSH key information
	availableKeys := sshManager.GetAvailableKeys()
	if len(availableKeys) > 0 {
		connectionInfo += fmt.Sprintf("Available SSH Keys:\n")
		for _, key := range availableKeys {
			connectionInfo += fmt.Sprintf("  - %s\n", key)
		}
	} else {
		connectionInfo += fmt.Sprintf("No SSH keys found. Generate one with:\n")
		connectionInfo += fmt.Sprintf("  ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa\n")
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("SSH connection information for VM '%s'", vmName),
		Output:  connectionInfo,
	}
}

// =============================================================================
// BICEP ACTIONS
// =============================================================================

// GenerateBicepFromResource generates a Bicep template from an existing Azure resource
func GenerateBicepFromResource(resourceID string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	template, err := bicepManager.GenerateBicepFromResource(ctx, resourceID)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to generate Bicep template: %v", err),
			Output:  "",
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Bicep template generated successfully for resource"),
		Output:  template.Content,
	}
}

// GenerateStorageAccountBicep generates a Bicep template for a Storage Account
func GenerateStorageAccountBicep(name, location, sku string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := bicepManager.GenerateStorageAccountTemplate(name, location, sku)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Storage Account Bicep template generated for '%s'", name),
		Output:  template.Content,
	}
}

// GenerateVMBicep generates a Bicep template for a Virtual Machine
func GenerateVMBicep(vmName, location, vmSize, adminUsername string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := bicepManager.GenerateVirtualMachineTemplate(vmName, location, vmSize, adminUsername)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Virtual Machine Bicep template generated for '%s'", vmName),
		Output:  template.Content,
	}
}

// GenerateAKSBicep generates a Bicep template for an AKS cluster
func GenerateAKSBicep(clusterName, location string, nodeCount int) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := bicepManager.GenerateAKSTemplate(clusterName, location, nodeCount)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("AKS Bicep template generated for '%s'", clusterName),
		Output:  template.Content,
	}
}

// GenerateKeyVaultBicep generates a Bicep template for a Key Vault
func GenerateKeyVaultBicep(vaultName, location, tenantId string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := bicepManager.GenerateKeyVaultTemplate(vaultName, location, tenantId)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Key Vault Bicep template generated for '%s'", vaultName),
		Output:  template.Content,
	}
}

// GenerateCompleteInfrastructureBicep generates a comprehensive infrastructure template
func GenerateCompleteInfrastructureBicep(projectName, location string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := bicepManager.GenerateCompleteInfrastructureTemplate(projectName, location)

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Complete infrastructure Bicep template generated for project '%s'", projectName),
		Output:  template.Content,
	}
}

// ValidateBicepFile validates a Bicep file
func ValidateBicepFile(filePath string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := bicepManager.ValidateBicep(ctx, filePath)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to validate Bicep file: %v", err),
			Output:  "",
		}
	}

	output := fmt.Sprintf("Bicep Validation Results:\n")
	output += fmt.Sprintf("- File: %s\n", filePath)
	output += fmt.Sprintf("- Valid: %t\n", result.Valid)

	if len(result.Errors) > 0 {
		output += fmt.Sprintf("- Errors:\n")
		for _, error := range result.Errors {
			output += fmt.Sprintf("  - %s\n", error)
		}
	}

	if len(result.Warnings) > 0 {
		output += fmt.Sprintf("- Warnings:\n")
		for _, warning := range result.Warnings {
			output += fmt.Sprintf("  - %s\n", warning)
		}
	}

	return ActionResult{
		Success: result.Valid,
		Message: fmt.Sprintf("Bicep validation completed for '%s'", filePath),
		Output:  output,
	}
}

// CompileBicepFile compiles a Bicep file to ARM template
func CompileBicepFile(filePath string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	outputPath, err := bicepManager.CompileBicep(ctx, filePath)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to compile Bicep file: %v", err),
			Output:  "",
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Bicep file compiled successfully"),
		Output:  fmt.Sprintf("Compiled ARM template saved to: %s", outputPath),
	}
}

// DeployBicepTemplate deploys a Bicep template to Azure
func DeployBicepTemplate(resourceGroup, deploymentName, filePath string, parameters map[string]string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes for deployment
	defer cancel()

	result, err := bicepManager.DeployBicep(ctx, resourceGroup, deploymentName, filePath, parameters)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Bicep deployment failed: %v", err),
			Output:  result.Error,
		}
	}

	output := fmt.Sprintf("Bicep Deployment Results:\n")
	output += fmt.Sprintf("- Deployment ID: %s\n", result.DeploymentID)
	output += fmt.Sprintf("- Duration: %v\n", result.Duration)
	output += fmt.Sprintf("- Success: %t\n", result.Success)
	output += fmt.Sprintf("- Message: %s\n", result.Message)

	if len(result.Output) > 0 {
		output += fmt.Sprintf("- Outputs:\n")
		for key, value := range result.Output {
			output += fmt.Sprintf("  - %s: %s\n", key, value)
		}
	}

	return ActionResult{
		Success: result.Success,
		Message: result.Message,
		Output:  output,
	}
}

// ParseBicepFile parses a Bicep file and returns its structure
func ParseBicepFile(filePath string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template, err := bicepManager.ParseBicepFile(filePath)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to parse Bicep file: %v", err),
			Output:  "",
		}
	}

	output := fmt.Sprintf("Bicep File Analysis:\n")
	output += fmt.Sprintf("- File: %s\n", template.FilePath)
	output += fmt.Sprintf("- Parameters: %d\n", len(template.Parameters))
	output += fmt.Sprintf("- Variables: %d\n", len(template.Variables))
	output += fmt.Sprintf("- Resources: %d\n", len(template.Resources))
	output += fmt.Sprintf("- Outputs: %d\n", len(template.Outputs))

	if len(template.Parameters) > 0 {
		output += fmt.Sprintf("\nParameters:\n")
		for name, param := range template.Parameters {
			output += fmt.Sprintf("  - %s (%s)\n", name, param.Type)
		}
	}

	if len(template.Resources) > 0 {
		output += fmt.Sprintf("\nResources:\n")
		for _, resource := range template.Resources {
			output += fmt.Sprintf("  - %s (%s)\n", resource.Name, resource.Type)
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Bicep file parsed successfully"),
		Output:  output,
	}
}

// SaveBicepTemplate saves a Bicep template to a file
func SaveBicepTemplate(content, filePath string) ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	template := &bicep.BicepTemplate{
		Content:  content,
		FilePath: filePath,
	}

	err := bicepManager.SaveBicepTemplate(template, filePath)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save Bicep template: %v", err),
			Output:  "",
		}
	}

	return ActionResult{
		Success: true,
		Message: fmt.Sprintf("Bicep template saved successfully"),
		Output:  fmt.Sprintf("Template saved to: %s", filePath),
	}
}

// GetBicepVersion gets the Bicep CLI version
func GetBicepVersion() ActionResult {
	if bicepManager == nil {
		return ActionResult{
			Success: false,
			Message: "Bicep manager not initialized. Call InitializeBicepManager first.",
			Output:  "",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	version, err := bicepManager.GetBicepVersion(ctx)
	if err != nil {
		return ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get Bicep version: %v", err),
			Output:  "",
		}
	}

	return ActionResult{
		Success: true,
		Message: "Bicep version retrieved successfully",
		Output:  fmt.Sprintf("Bicep CLI Version: %s", version),
	}
}
