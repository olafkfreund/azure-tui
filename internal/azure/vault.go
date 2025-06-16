package vault

import (
	"encoding/json"
	"os/exec"
)

type KeyVault struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListKeyVaults() ([]KeyVault, error) {
	cmd := exec.Command("az", "keyvault", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var vaults []KeyVault
	if err := json.Unmarshal(out, &vaults); err != nil {
		return nil, err
	}
	return vaults, nil
}

func CreateKeyVault(name, group, location string) error {
	return exec.Command("az", "keyvault", "create", "--name", name, "--resource-group", group, "--location", location).Run()
}

func DeleteKeyVault(name, group string) error {
	return exec.Command("az", "keyvault", "delete", "--name", name, "--resource-group", group).Run()
}
