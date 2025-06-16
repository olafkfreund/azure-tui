package aks

import (
	"encoding/json"
	"os/exec"
)

type AKSCluster struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListAKSClusters() ([]AKSCluster, error) {
	cmd := exec.Command("az", "aks", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var clusters []AKSCluster
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

func CreateAKSCluster(name, group, location string) error {
	return exec.Command("az", "aks", "create", "--name", name, "--resource-group", group, "--location", location, "--node-count", "1", "--generate-ssh-keys").Run()
}

func DeleteAKSCluster(name, group string) error {
	return exec.Command("az", "aks", "delete", "--name", name, "--resource-group", group, "--yes", "--no-wait").Run()
}

func AKSGetCredentials(name, group string) error {
	return exec.Command("az", "aks", "get-credentials", "--name", name, "--resource-group", group, "--overwrite-existing").Run()
}
