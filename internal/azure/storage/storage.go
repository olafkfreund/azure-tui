package storage

import (
	"encoding/json"
	"os/exec"
)

type StorageAccount struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListStorageAccounts() ([]StorageAccount, error) {
	cmd := exec.Command("az", "storage", "account", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var accounts []StorageAccount
	if err := json.Unmarshal(out, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func CreateStorageAccount(name, group, location string) error {
	return exec.Command("az", "storage", "account", "create", "--name", name, "--resource-group", group, "--location", location, "--sku", "Standard_LRS").Run()
}

func DeleteStorageAccount(name, group string) error {
	return exec.Command("az", "storage", "account", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}
