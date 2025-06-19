package main

import (
	"fmt"
	"log"

	"github.com/olafkfreund/azure-tui/internal/openai"
)

func main() {
	fmt.Println("Testing AI Provider functionality...")

	// Test auto-detection
	provider := openai.NewAIProviderAuto()
	if provider == nil {
		fmt.Println("❌ No AI provider configured")
		return
	}

	fmt.Printf("✅ AI Provider initialized successfully: %s\n", provider.ProviderType)

	// Test a simple AI request
	testResourceInfo := `{
		"name": "test-vm",
		"type": "Microsoft.Compute/virtualMachines",
		"location": "eastus",
		"properties": {
			"vmSize": "Standard_B2s",
			"osProfile": {
				"computerName": "test-vm",
				"adminUsername": "azureuser"
			},
			"storageProfile": {
				"osDisk": {
					"osType": "Linux",
					"createOption": "FromImage"
				}
			}
		}
	}`

	fmt.Println("\nTesting AI analysis with sample resource...")
	description, err := provider.DescribeResource("Virtual Machine", "test-vm", testResourceInfo)
	if err != nil {
		log.Printf("❌ AI analysis failed: %v", err)
		return
	}

	fmt.Println("✅ AI analysis successful!")
	fmt.Printf("Description: %s\n", description)
}
