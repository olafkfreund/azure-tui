package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func fetchResourceGroups() ([]ResourceGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Azure CLI timeout after 10 seconds")
		}
		return nil, fmt.Errorf("failed to get resource groups: %v", err)
	}

	var azGroups []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.Unmarshal(output, &azGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource group data: %v", err)
	}

	var result []ResourceGroup
	for _, g := range azGroups {
		result = append(result, ResourceGroup{
			Name:     g.Name,
			Location: g.Location,
		})
	}
	return result, nil
}

func main() {
	fmt.Println("Testing Azure CLI resource group loading...")

	groups, err := fetchResourceGroups()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully loaded %d resource groups:\n", len(groups))
	for _, group := range groups {
		fmt.Printf("  - %s (%s)\n", group.Name, group.Location)
	}
}
