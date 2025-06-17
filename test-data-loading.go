package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId"`
	IsDefault bool   `json:"isDefault"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func main() {
	fmt.Println("🔍 Testing Azure TUI data loading...")

	// Test 1: Load subscriptions
	fmt.Println("\n1. Loading subscriptions...")
	subs, err := fetchSubscriptions()
	if err != nil {
		fmt.Printf("❌ Error loading subscriptions: %v\n", err)
	} else {
		fmt.Printf("✅ Loaded %d subscriptions:\n", len(subs))
		for i, sub := range subs {
			if i < 3 { // Show first 3
				fmt.Printf("   - %s (%s)\n", sub.Name, sub.ID)
			}
		}
		if len(subs) > 3 {
			fmt.Printf("   ... and %d more\n", len(subs)-3)
		}
	}

	// Test 2: Load resource groups
	fmt.Println("\n2. Loading resource groups...")
	rgs, err := fetchResourceGroups()
	if err != nil {
		fmt.Printf("❌ Error loading resource groups: %v\n", err)
	} else {
		fmt.Printf("✅ Loaded %d resource groups:\n", len(rgs))
		for i, rg := range rgs {
			if i < 5 { // Show first 5
				fmt.Printf("   - %s (%s)\n", rg.Name, rg.Location)
			}
		}
		if len(rgs) > 5 {
			fmt.Printf("   ... and %d more\n", len(rgs)-5)
		}
	}

	fmt.Println("\n🎉 Data loading test complete!")
}

func fetchSubscriptions() ([]Subscription, error) {
	cmd := exec.Command("az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute az command: %w", err)
	}

	var subscriptions []Subscription
	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to parse subscription JSON: %w", err)
	}

	return subscriptions, nil
}

func fetchResourceGroups() ([]ResourceGroup, error) {
	cmd := exec.Command("az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute az command: %w", err)
	}

	var resourceGroups []ResourceGroup
	if err := json.Unmarshal(output, &resourceGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource group JSON: %w", err)
	}

	return resourceGroups, nil
}
