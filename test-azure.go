package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("Testing Azure CLI integration...")

	// Test basic command
	cmd := exec.Command("az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var groups []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}

	if err := json.Unmarshal(output, &groups); err != nil {
		fmt.Printf("JSON Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d resource groups:\n", len(groups))
	for i, g := range groups {
		fmt.Printf("%d. %s (%s)\n", i+1, g.Name, g.Location)
	}
}
