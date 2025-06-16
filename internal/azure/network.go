package network

import (
	"encoding/json"
	"os/exec"
)

type VirtualNetwork struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

type Firewall struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

func ListVirtualNetworks() ([]VirtualNetwork, error) {
	cmd := exec.Command("az", "network", "vnet", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var vnets []VirtualNetwork
	if err := json.Unmarshal(out, &vnets); err != nil {
		return nil, err
	}
	return vnets, nil
}

func CreateVirtualNetwork(name, group, location string) error {
	return exec.Command("az", "network", "vnet", "create", "--name", name, "--resource-group", group, "--location", location, "--address-prefix", "10.0.0.0/16").Run()
}

func DeleteVirtualNetwork(name, group string) error {
	return exec.Command("az", "network", "vnet", "delete", "--name", name, "--resource-group", group).Run()
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
