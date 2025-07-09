package vm

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewVMManager(t *testing.T) {
	// Since we can't create real Azure credentials in tests, we'll test with nil
	// In a real scenario, this would be created with proper credentials
	manager := NewVMManager(nil, "test-subscription-id")
	
	if manager == nil {
		t.Fatal("NewVMManager returned nil")
	}
	
	if manager.subscriptionID != "test-subscription-id" {
		t.Errorf("Expected subscription ID 'test-subscription-id', got '%s'", manager.subscriptionID)
	}
	
	if manager.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", manager.timeout)
	}
}

func TestVMInfo(t *testing.T) {
	// Test VMInfo struct
	vmInfo := &VMInfo{
		Name:              "test-vm",
		ResourceGroup:     "test-rg",
		Location:          "eastus",
		Size:              "Standard_B1s",
		OSType:            "Linux",
		PowerState:        "running",
		PrivateIPAddress:  "10.0.0.4",
		PublicIPAddress:   "20.1.2.3",
		FQDN:              "test-vm.eastus.cloudapp.azure.com",
		NetworkInterfaces: []string{"/subscriptions/test/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/test-vm-nic"},
		AdminUsername:     "azureuser",
		SSHPublicKeys:     []string{"ssh-rsa AAAAB3... test@test.com"},
	}
	
	if vmInfo.Name != "test-vm" {
		t.Errorf("Expected VM name 'test-vm', got '%s'", vmInfo.Name)
	}
	
	if vmInfo.OSType != "Linux" {
		t.Errorf("Expected OS type 'Linux', got '%s'", vmInfo.OSType)
	}
	
	if vmInfo.PowerState != "running" {
		t.Errorf("Expected power state 'running', got '%s'", vmInfo.PowerState)
	}
	
	if len(vmInfo.NetworkInterfaces) != 1 {
		t.Errorf("Expected 1 network interface, got %d", len(vmInfo.NetworkInterfaces))
	}
	
	if len(vmInfo.SSHPublicKeys) != 1 {
		t.Errorf("Expected 1 SSH public key, got %d", len(vmInfo.SSHPublicKeys))
	}
}

func TestSetTimeout(t *testing.T) {
	manager := NewVMManager(nil, "test-subscription-id")
	
	newTimeout := 60 * time.Second
	manager.SetTimeout(newTimeout)
	
	if manager.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, manager.timeout)
	}
}

// Test VM resource ID parsing logic
func TestVMResourceIDParsing(t *testing.T) {
	testCases := []struct {
		resourceID       string
		expectedRG       string
		expectedResource string
		shouldFail       bool
	}{
		{
			resourceID:       "/subscriptions/12345/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/test-nic",
			expectedRG:       "test-rg",
			expectedResource: "test-nic",
			shouldFail:       false,
		},
		{
			resourceID:       "/subscriptions/12345/resourceGroups/my-rg/providers/Microsoft.Network/publicIPAddresses/my-ip",
			expectedRG:       "my-rg",
			expectedResource: "my-ip",
			shouldFail:       false,
		},
		{
			resourceID: "invalid-resource-id",
			shouldFail: true,
		},
		{
			resourceID: "/subscriptions/12345/resourceGroups",
			shouldFail: true,
		},
	}
	
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("TestCase%d", i+1), func(t *testing.T) {
			parts := strings.Split(tc.resourceID, "/")
			
			if tc.shouldFail {
				if len(parts) >= 9 {
					t.Errorf("Expected parsing to fail for invalid resource ID: %s", tc.resourceID)
				}
				return
			}
			
			if len(parts) < 9 {
				t.Errorf("Expected valid resource ID, but parsing failed: %s", tc.resourceID)
				return
			}
			
			resourceGroup := parts[4]
			resourceName := parts[8]
			
			if resourceGroup != tc.expectedRG {
				t.Errorf("Expected resource group '%s', got '%s'", tc.expectedRG, resourceGroup)
			}
			
			if resourceName != tc.expectedResource {
				t.Errorf("Expected resource name '%s', got '%s'", tc.expectedResource, resourceName)
			}
		})
	}
}

// Test default username logic
func TestDefaultUsernameLogic(t *testing.T) {
	testCases := []struct {
		osType           string
		providedUsername string
		expectedUsername string
	}{
		{
			osType:           "Linux",
			providedUsername: "",
			expectedUsername: "azureuser",
		},
		{
			osType:           "Windows",
			providedUsername: "",
			expectedUsername: "azureuser",
		},
		{
			osType:           "linux",
			providedUsername: "",
			expectedUsername: "azureuser",
		},
		{
			osType:           "windows",
			providedUsername: "",
			expectedUsername: "azureuser",
		},
		{
			osType:           "Unknown",
			providedUsername: "",
			expectedUsername: "azureuser",
		},
		{
			osType:           "Linux",
			providedUsername: "customuser",
			expectedUsername: "customuser",
		},
	}
	
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("UsernameTest%d", i+1), func(t *testing.T) {
			username := tc.providedUsername
			if username == "" {
				// Simulate default username logic
				switch strings.ToLower(tc.osType) {
				case "linux", "windows":
					username = "azureuser"
				default:
					username = "azureuser"
				}
			}
			
			if username != tc.expectedUsername {
				t.Errorf("Expected username '%s', got '%s' for OS type '%s'", tc.expectedUsername, username, tc.osType)
			}
		})
	}
}

// Test VM connection priority logic (public IP > private IP > FQDN)
func TestVMConnectionPriority(t *testing.T) {
	testCases := []struct {
		name               string
		publicIP           string
		privateIP          string
		fqdn               string
		expectedHost       string
		shouldHaveHost     bool
	}{
		{
			name:           "Public IP available",
			publicIP:       "20.1.2.3",
			privateIP:      "10.0.0.4",
			fqdn:           "test.eastus.cloudapp.azure.com",
			expectedHost:   "20.1.2.3",
			shouldHaveHost: true,
		},
		{
			name:           "Only private IP",
			publicIP:       "",
			privateIP:      "10.0.0.4",
			fqdn:           "",
			expectedHost:   "10.0.0.4",
			shouldHaveHost: true,
		},
		{
			name:           "Only FQDN",
			publicIP:       "",
			privateIP:      "",
			fqdn:           "test.eastus.cloudapp.azure.com",
			expectedHost:   "test.eastus.cloudapp.azure.com",
			shouldHaveHost: true,
		},
		{
			name:           "No connection options",
			publicIP:       "",
			privateIP:      "",
			fqdn:           "",
			expectedHost:   "",
			shouldHaveHost: false,
		},
		{
			name:           "Private IP and FQDN (prefer private IP)",
			publicIP:       "",
			privateIP:      "10.0.0.4",
			fqdn:           "test.eastus.cloudapp.azure.com",
			expectedHost:   "10.0.0.4",
			shouldHaveHost: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate connection host selection logic
			host := tc.publicIP
			if host == "" {
				host = tc.privateIP
			}
			if host == "" && tc.fqdn != "" {
				host = tc.fqdn
			}
			
			if tc.shouldHaveHost {
				if host == "" {
					t.Error("Expected to have a host, but got empty string")
				} else if host != tc.expectedHost {
					t.Errorf("Expected host '%s', got '%s'", tc.expectedHost, host)
				}
			} else {
				if host != "" {
					t.Errorf("Expected no host, but got '%s'", host)
				}
			}
		})
	}
}

