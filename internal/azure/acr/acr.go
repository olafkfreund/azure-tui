package acr

import (
	"encoding/json"
	"os/exec"
)

type ContainerRegistry struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListContainerRegistries() ([]ContainerRegistry, error) {
	cmd := exec.Command("az", "acr", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var registries []ContainerRegistry
	if err := json.Unmarshal(out, &registries); err != nil {
		return nil, err
	}
	return registries, nil
}

func CreateContainerRegistry(name, group, location string) error {
	return exec.Command("az", "acr", "create", "--name", name, "--resource-group", group, "--location", location, "--sku", "Basic").Run()
}

func DeleteContainerRegistry(name, group string) error {
	return exec.Command("az", "acr", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}
