package vm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// VMInfo represents virtual machine information
type VMInfo struct {
	Name              string
	ResourceGroup     string
	Location          string
	Size              string
	OSType            string
	PowerState        string
	PrivateIPAddress  string
	PublicIPAddress   string
	FQDN              string
	NetworkInterfaces []string
	AdminUsername     string
	SSHPublicKeys     []string
}

// VMManager manages Azure Virtual Machines
type VMManager struct {
	cred           *azidentity.DefaultAzureCredential
	subscriptionID string
	timeout        time.Duration
}

// NewVMManager creates a new VM manager
func NewVMManager(cred *azidentity.DefaultAzureCredential, subscriptionID string) *VMManager {
	return &VMManager{
		cred:           cred,
		subscriptionID: subscriptionID,
		timeout:        30 * time.Second,
	}
}

// GetVMInfo retrieves detailed information about a VM including IP addresses
func (vm *VMManager) GetVMInfo(ctx context.Context, resourceGroupName, vmName string) (*VMInfo, error) {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %v", err)
	}

	// Get VM details
	resp, err := client.Get(ctx, resourceGroupName, vmName, &armcompute.VirtualMachinesClientGetOptions{
		Expand: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get VM details: %v", err)
	}

	vmDetails := resp.VirtualMachine
	vmInfo := &VMInfo{
		Name:          *vmDetails.Name,
		ResourceGroup: resourceGroupName,
		Location:      *vmDetails.Location,
	}

	// Extract VM size
	if vmDetails.Properties != nil && vmDetails.Properties.HardwareProfile != nil && vmDetails.Properties.HardwareProfile.VMSize != nil {
		vmInfo.Size = string(*vmDetails.Properties.HardwareProfile.VMSize)
	}

	// Extract OS type
	if vmDetails.Properties != nil && vmDetails.Properties.StorageProfile != nil && vmDetails.Properties.StorageProfile.OSDisk != nil {
		if vmDetails.Properties.StorageProfile.OSDisk.OSType != nil {
			vmInfo.OSType = string(*vmDetails.Properties.StorageProfile.OSDisk.OSType)
		}
	}

	// Extract admin username and SSH keys
	if vmDetails.Properties != nil && vmDetails.Properties.OSProfile != nil {
		if vmDetails.Properties.OSProfile.AdminUsername != nil {
			vmInfo.AdminUsername = *vmDetails.Properties.OSProfile.AdminUsername
		}

		// Extract SSH public keys for Linux VMs
		if vmDetails.Properties.OSProfile.LinuxConfiguration != nil &&
			vmDetails.Properties.OSProfile.LinuxConfiguration.SSH != nil &&
			vmDetails.Properties.OSProfile.LinuxConfiguration.SSH.PublicKeys != nil {
			for _, key := range vmDetails.Properties.OSProfile.LinuxConfiguration.SSH.PublicKeys {
				if key.KeyData != nil {
					vmInfo.SSHPublicKeys = append(vmInfo.SSHPublicKeys, *key.KeyData)
				}
			}
		}
	}

	// Get network interfaces
	if vmDetails.Properties != nil && vmDetails.Properties.NetworkProfile != nil && vmDetails.Properties.NetworkProfile.NetworkInterfaces != nil {
		for _, nic := range vmDetails.Properties.NetworkProfile.NetworkInterfaces {
			if nic.ID != nil {
				vmInfo.NetworkInterfaces = append(vmInfo.NetworkInterfaces, *nic.ID)
			}
		}
	}

	// Get power state
	powerState, err := vm.getVMPowerState(ctx, resourceGroupName, vmName)
	if err == nil {
		vmInfo.PowerState = powerState
	}

	// Get IP addresses from network interfaces
	if len(vmInfo.NetworkInterfaces) > 0 {
		privateIP, publicIP, fqdn, err := vm.getVMIPAddresses(ctx, vmInfo.NetworkInterfaces[0])
		if err == nil {
			vmInfo.PrivateIPAddress = privateIP
			vmInfo.PublicIPAddress = publicIP
			vmInfo.FQDN = fqdn
		}
	}

	return vmInfo, nil
}

// getVMPowerState retrieves the current power state of a VM
func (vm *VMManager) getVMPowerState(ctx context.Context, resourceGroupName, vmName string) (string, error) {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.InstanceView(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return "", err
	}

	if resp.VirtualMachineInstanceView.Statuses != nil {
		for _, status := range resp.VirtualMachineInstanceView.Statuses {
			if status.Code != nil && strings.HasPrefix(*status.Code, "PowerState/") {
				return strings.TrimPrefix(*status.Code, "PowerState/"), nil
			}
		}
	}

	return "unknown", nil
}

// getVMIPAddresses retrieves IP addresses associated with a network interface
func (vm *VMManager) getVMIPAddresses(ctx context.Context, nicID string) (privateIP, publicIP, fqdn string, err error) {
	// Parse NIC ID to extract resource group and NIC name
	// Format: /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/networkInterfaces/{nicName}
	parts := strings.Split(nicID, "/")
	if len(parts) < 9 {
		return "", "", "", fmt.Errorf("invalid network interface ID format")
	}

	resourceGroupName := parts[4]
	nicName := parts[8]

	// Get network interface details
	nicClient, err := armnetwork.NewInterfacesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return "", "", "", err
	}

	nicResp, err := nicClient.Get(ctx, resourceGroupName, nicName, nil)
	if err != nil {
		return "", "", "", err
	}

	networkInterface := nicResp.Interface

	// Extract private IP
	if networkInterface.Properties != nil && networkInterface.Properties.IPConfigurations != nil {
		for _, ipConfig := range networkInterface.Properties.IPConfigurations {
			if ipConfig.Properties != nil && ipConfig.Properties.PrivateIPAddress != nil {
				privateIP = *ipConfig.Properties.PrivateIPAddress
			}

			// Get public IP if associated
			if ipConfig.Properties != nil && ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
				publicIPAddr, publicFQDN, err := vm.getPublicIPDetails(ctx, *ipConfig.Properties.PublicIPAddress.ID)
				if err == nil {
					publicIP = publicIPAddr
					fqdn = publicFQDN
				}
			}
		}
	}

	return privateIP, publicIP, fqdn, nil
}

// getPublicIPDetails retrieves public IP address details
func (vm *VMManager) getPublicIPDetails(ctx context.Context, publicIPID string) (ipAddress, fqdn string, err error) {
	// Parse public IP ID
	parts := strings.Split(publicIPID, "/")
	if len(parts) < 9 {
		return "", "", fmt.Errorf("invalid public IP ID format")
	}

	resourceGroupName := parts[4]
	publicIPName := parts[8]

	// Get public IP details
	pipClient, err := armnetwork.NewPublicIPAddressesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return "", "", err
	}

	pipResp, err := pipClient.Get(ctx, resourceGroupName, publicIPName, nil)
	if err != nil {
		return "", "", err
	}

	publicIP := pipResp.PublicIPAddress

	if publicIP.Properties != nil {
		if publicIP.Properties.IPAddress != nil {
			ipAddress = *publicIP.Properties.IPAddress
		}
		if publicIP.Properties.DNSSettings != nil && publicIP.Properties.DNSSettings.Fqdn != nil {
			fqdn = *publicIP.Properties.DNSSettings.Fqdn
		}
	}

	return ipAddress, fqdn, nil
}

// ListVMs lists all VMs in a resource group
func (vm *VMManager) ListVMs(ctx context.Context, resourceGroupName string) ([]*VMInfo, error) {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %v", err)
	}

	var vms []*VMInfo
	pager := client.NewListPager(resourceGroupName, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get VM page: %v", err)
		}

		for _, vmResource := range page.Value {
			if vmResource.Name != nil {
				vmInfo, err := vm.GetVMInfo(ctx, resourceGroupName, *vmResource.Name)
				if err != nil {
					// Log error but continue with other VMs
					continue
				}
				vms = append(vms, vmInfo)
			}
		}
	}

	return vms, nil
}

// StartVM starts a virtual machine
func (vm *VMManager) StartVM(ctx context.Context, resourceGroupName, vmName string) error {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create VM client: %v", err)
	}

	poller, err := client.BeginStart(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to start VM: %v", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

// StopVM stops a virtual machine
func (vm *VMManager) StopVM(ctx context.Context, resourceGroupName, vmName string) error {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create VM client: %v", err)
	}

	poller, err := client.BeginPowerOff(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to stop VM: %v", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

// RestartVM restarts a virtual machine
func (vm *VMManager) RestartVM(ctx context.Context, resourceGroupName, vmName string) error {
	client, err := armcompute.NewVirtualMachinesClient(vm.subscriptionID, vm.cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create VM client: %v", err)
	}

	poller, err := client.BeginRestart(ctx, resourceGroupName, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to restart VM: %v", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

// GetVMConnectionInfo gets the information needed for SSH connection
func (vm *VMManager) GetVMConnectionInfo(ctx context.Context, resourceGroupName, vmName string) (host, username string, err error) {
	vmInfo, err := vm.GetVMInfo(ctx, resourceGroupName, vmName)
	if err != nil {
		return "", "", err
	}

	// Prefer public IP, fall back to private IP
	host = vmInfo.PublicIPAddress
	if host == "" {
		host = vmInfo.PrivateIPAddress
	}

	// Use FQDN if available and no IP addresses
	if host == "" && vmInfo.FQDN != "" {
		host = vmInfo.FQDN
	}

	if host == "" {
		return "", "", fmt.Errorf("no accessible IP address found for VM %s", vmName)
	}

	username = vmInfo.AdminUsername
	if username == "" {
		// Default usernames for common OS types
		switch strings.ToLower(vmInfo.OSType) {
		case "linux":
			username = "azureuser"
		case "windows":
			username = "azureuser"
		default:
			username = "azureuser"
		}
	}

	return host, username, nil
}

// SetTimeout sets the operation timeout
func (vm *VMManager) SetTimeout(timeout time.Duration) {
	vm.timeout = timeout
}