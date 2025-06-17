package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Subscription struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	fmt.Println("Testing Azure CLI directly...")

	// Test subscription loading
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var subs []Subscription
	if err := json.Unmarshal(output, &subs); err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Found %d subscriptions\n", len(subs))
	for _, sub := range subs {
		fmt.Printf("- %s (%s)\n", sub.Name, sub.ID)
	}

	// Test resource groups
	cmd2 := exec.CommandContext(ctx, "az", "group", "list", "--output", "json")
	output2, err := cmd2.Output()
	if err != nil {
		fmt.Printf("RG Error: %v\n", err)
		return
	}

	var rgs []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.Unmarshal(output2, &rgs); err != nil {
		fmt.Printf("RG Parse error: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Found %d resource groups\n", len(rgs))
	for _, rg := range rgs {
		fmt.Printf("- %s (%s)\n", rg.Name, rg.Location)
	}
}
