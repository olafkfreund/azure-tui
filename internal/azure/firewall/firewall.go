package firewall

import (
	"encoding/json"
	"os/exec"
)

type Firewall struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListFirewalls() ([]Firewall, error) {
	cmd := exec.Command("az", "network", "firewall", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var fws []Firewall
	if err := json.Unmarshal(out, &fws); err != nil {
		return nil, err
	}
	return fws, nil
}

func CreateFirewall(name, group, location string) error {
	return exec.Command("az", "network", "firewall", "create", "--name", name, "--resource-group", group, "--location", location).Run()
}

func DeleteFirewall(name, group string) error {
	return exec.Command("az", "network", "firewall", "delete", "--name", name, "--resource-group", group).Run()
}
