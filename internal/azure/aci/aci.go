package aci

import (
	"encoding/json"
	"os/exec"
)

type ContainerInstance struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListContainerInstances() ([]ContainerInstance, error) {
	cmd := exec.Command("az", "container", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var containers []ContainerInstance
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}

func CreateContainerInstance(name, group, location, image string) error {
	return exec.Command("az", "container", "create", "--name", name, "--resource-group", group, "--location", location, "--image", image).Run()
}

func DeleteContainerInstance(name, group string) error {
	return exec.Command("az", "container", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}
