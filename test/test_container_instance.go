package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/olafkfreund/azure-tui/internal/azure/aci"
)

func main() {
	fmt.Println("🐳 Testing Container Instance Functionality")
	fmt.Println("==========================================")

	// Test 1: List container instances
	fmt.Println("\n1. Testing ListContainerInstances()...")
	containers, err := aci.ListContainerInstances()
	if err != nil {
		log.Printf("Error listing containers: %v", err)
	} else {
		fmt.Printf("   Found %d container instances\n", len(containers))
		for _, container := range containers {
			fmt.Printf("   - %s (Resource Group: %s, State: %s)\n", 
				container.Name, container.ResourceGroup, container.ProvisioningState)
		}
	}

	// Test 2: Get detailed container instance information
	fmt.Println("\n2. Testing GetContainerInstanceDetails()...")
	containerName := "cadmin"
	resourceGroup := "con_demo_01"
	
	container, err := aci.GetContainerInstanceDetails(containerName, resourceGroup)
	if err != nil {
		log.Printf("Error getting container details: %v", err)
	} else {
		fmt.Printf("   Container: %s\n", container.Name)
		fmt.Printf("   Location: %s\n", container.Location)
		fmt.Printf("   OS Type: %s\n", container.OSType)
		fmt.Printf("   Provisioning State: %s\n", container.ProvisioningState)
		if container.IPAddress != nil {
			fmt.Printf("   Public IP: %s\n", container.IPAddress.IP)
			fmt.Printf("   FQDN: %s\n", container.IPAddress.FQDN)
		}
		fmt.Printf("   Number of containers: %d\n", len(container.Containers))
		for i, cont := range container.Containers {
			fmt.Printf("     Container %d: %s (Image: %s)\n", i+1, cont.Name, cont.Image)
			if cont.Resources.Requests != nil {
				fmt.Printf("       CPU: %.1f cores, Memory: %.1f GB\n", 
					cont.Resources.Requests.CPU, cont.Resources.Requests.MemoryInGB)
			}
		}
	}

	// Test 3: Test rendering function
	fmt.Println("\n3. Testing RenderContainerInstanceDetails()...")
	rendered := aci.RenderContainerInstanceDetails(containerName, resourceGroup)
	fmt.Printf("   Rendered content length: %d characters\n", len(rendered))
	
	// Show first few lines of rendered content
	lines := strings.Split(rendered, "\n")
	maxLines := 10
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	fmt.Println("   First 10 lines of rendered output:")
	for i, line := range lines {
		fmt.Printf("   %2d: %s\n", i+1, line)
	}

	// Test 4: Test container logs
	fmt.Println("\n4. Testing GetContainerLogs()...")
	logs, err := aci.GetContainerLogs(containerName, resourceGroup, "", 10)
	if err != nil {
		log.Printf("Error getting container logs: %v", err)
	} else {
		fmt.Printf("   Retrieved logs (%d characters)\n", len(logs))
		logLines := strings.Split(logs, "\n")
		maxLogLines := 5
		if len(logLines) > maxLogLines {
			logLines = logLines[:maxLogLines]
		}
		fmt.Println("   First 5 lines of logs:")
		for i, line := range logLines {
			fmt.Printf("   %2d: %s\n", i+1, line)
		}
	}

	fmt.Println("\n✅ Container Instance functionality test completed!")
}
	"log"

	"github.com/olafkfreund/azure-tui/internal/azure/aci"
)

func main() {
	fmt.Println("🐳 Testing Container Instance Functionality")
	fmt.Println("==========================================")

	// Test 1: List container instances
	fmt.Println("\n1. Testing ListContainerInstances()...")
	containers, err := aci.ListContainerInstances()
	if err != nil {
		log.Printf("Error listing containers: %v", err)
	} else {
		fmt.Printf("   Found %d container instances\n", len(containers))
		for _, container := range containers {
			fmt.Printf("   - %s (Resource Group: %s, State: %s)\n", 
				container.Name, container.ResourceGroup, container.ProvisioningState)
		}
	}

	// Test 2: Get detailed container instance information
	fmt.Println("\n2. Testing GetContainerInstanceDetails()...")
	containerName := "cadmin"
	resourceGroup := "con_demo_01"
	
	container, err := aci.GetContainerInstanceDetails(containerName, resourceGroup)
	if err != nil {
		log.Printf("Error getting container details: %v", err)
	} else {
		fmt.Printf("   Container: %s\n", container.Name)
		fmt.Printf("   Location: %s\n", container.Location)
		fmt.Printf("   OS Type: %s\n", container.OSType)
		fmt.Printf("   Provisioning State: %s\n", container.ProvisioningState)
		if container.IPAddress != nil {
			fmt.Printf("   Public IP: %s\n", container.IPAddress.IP)
			fmt.Printf("   FQDN: %s\n", container.IPAddress.FQDN)
		}
		fmt.Printf("   Number of containers: %d\n", len(container.Containers))
		for i, cont := range container.Containers {
			fmt.Printf("     Container %d: %s (Image: %s)\n", i+1, cont.Name, cont.Image)
			if cont.Resources.Requests != nil {
				fmt.Printf("       CPU: %.1f cores, Memory: %.1f GB\n", 
					cont.Resources.Requests.CPU, cont.Resources.Requests.MemoryInGB)
			}
		}
	}

	// Test 3: Test rendering function
	fmt.Println("\n3. Testing RenderContainerInstanceDetails()...")
	rendered := aci.RenderContainerInstanceDetails(containerName, resourceGroup)
	fmt.Printf("   Rendered content length: %d characters\n", len(rendered))
	
	// Show first few lines of rendered content
	lines := splitLines(rendered, 10)
	fmt.Println("   First 10 lines of rendered output:")
	for i, line := range lines {
		fmt.Printf("   %2d: %s\n", i+1, line)
	}

	// Test 4: Test container logs
	fmt.Println("\n4. Testing GetContainerLogs()...")
	logs, err := aci.GetContainerLogs(containerName, resourceGroup, "", 10)
	if err != nil {
		log.Printf("Error getting container logs: %v", err)
	} else {
		fmt.Printf("   Retrieved logs (%d characters)\n", len(logs))
		logLines := splitLines(logs, 5)
		fmt.Println("   First 5 lines of logs:")
		for i, line := range logLines {
			fmt.Printf("   %2d: %s\n", i+1, line)
		}
	}

	fmt.Println("\n✅ Container Instance functionality test completed!")
}

// Helper function to split string into lines and limit the number
func splitLines(content string, maxLines int) []string {
	lines := []string{}
	currentLine := ""
	
	for _, char := range content {
		if char == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
			if len(lines) >= maxLines {
				break
			}
		} else {
			currentLine += string(char)
		}
	}
	
	if currentLine != "" && len(lines) < maxLines {
		lines = append(lines, currentLine)
	}
	
	return lines
}
