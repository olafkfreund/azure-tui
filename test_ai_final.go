package main

import (
	"fmt"
	"os"

	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
	"github.com/olafkfreund/azure-tui/internal/openai"
)

func main() {
	fmt.Println("🧪 Testing AI functionality in Azure TUI context...")

	// Force OpenAI provider since GitHub Copilot is having permission issues
	os.Setenv("USE_GITHUB_COPILOT", "false")

	// Test AI provider initialization (same as in main.go)
	ai := openai.NewAIProviderAuto()
	if ai == nil {
		fmt.Println("❌ No AI provider configured")
		return
	}

	fmt.Printf("✅ AI Provider initialized: %s\n", ai.ProviderType)

	// Create a sample Azure resource (similar to what the TUI would have)
	testResource := AzureResource{
		ID:            "/subscriptions/test/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/test-vm",
		Name:          "test-vm",
		Type:          "Microsoft.Compute/virtualMachines",
		Location:      "eastus",
		ResourceGroup: "test-rg",
		Status:        "Running",
		Tags:          map[string]string{"Environment": "Test", "Owner": "DevTeam"},
		Properties:    map[string]interface{}{"vmSize": "Standard_B2s", "osType": "Linux"},
	}

	// Create sample resource details
	details := &resourcedetails.ResourceDetails{
		ID:       testResource.ID,
		Name:     testResource.Name,
		Type:     testResource.Type,
		Location: testResource.Location,
		Status:   testResource.Status,
		Properties: map[string]interface{}{
			"vmSize": "Standard_B2s",
			"osProfile": map[string]interface{}{
				"computerName":  "test-vm",
				"adminUsername": "azureuser",
			},
			"storageProfile": map[string]interface{}{
				"osDisk": map[string]interface{}{
					"osType":       "Linux",
					"createOption": "FromImage",
				},
			},
		},
	}

	// Test the same AI analysis that the TUI uses
	fmt.Println("\n🤖 Testing AI analysis (same logic as pressing 'a' in TUI)...")

	detailsStr := fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s\nStatus: %s",
		testResource.Name, testResource.Type, testResource.Location, testResource.Status)

	if details != nil {
		detailsStr += fmt.Sprintf("\nProperties: %v", details.Properties)
	}

	description, err := ai.DescribeResource(testResource.Type, testResource.Name, detailsStr)
	if err != nil {
		fmt.Printf("❌ AI analysis failed: %v\n", err)
		return
	}

	fmt.Println("✅ AI analysis successful!")
	fmt.Println("\n📝 AI Analysis Result:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println(description)
	fmt.Println("═══════════════════════════════════════════════════════════════")

	fmt.Println("\n🎉 Azure TUI AI functionality is working correctly!")
	fmt.Println("   You can now press 'a' in the TUI to analyze any selected resource.")
}

// Copy the AzureResource struct from main.go to make this test work
type AzureResource struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Location      string                 `json:"location"`
	ResourceGroup string                 `json:"resourceGroup"`
	Status        string                 `json:"status,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
}
