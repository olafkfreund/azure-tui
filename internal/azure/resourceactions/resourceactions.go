package resourceactions

import (
	"fmt"
	"os/exec"
	"strings"
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
	sshKeyCmd := exec.Command("az", "vm", "show",
		"--name", vmName,
		"--resource-group", resourceGroup,
		"--query", "osProfile.linuxConfiguration.disablePasswordAuthentication",
		"--output", "tsv")

	keyAuthOutput, err := sshKeyCmd.Output()
	keyAuthEnabled := strings.TrimSpace(string(keyAuthOutput)) == "true"

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
	}

	return ActionResult{
		Success: false,
		Message: fmt.Sprintf("Action '%s' not supported for resource type '%s'", action, resourceType),
		Output:  "",
	}
}
