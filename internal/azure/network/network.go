package network

import (
	"encoding/json"
	"os/exec"
	"strings"

	ai "github.com/olafkfreund/azure-tui/internal/openai"
	"github.com/olafkfreund/azure-tui/internal/tui"
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

// Example: Show a matrix graph of VNet usage in the TUI
// (This would be called from your TUI's View or update logic)
func ExampleShowVNetMatrixGraph() string {
	vnets, err := ListVirtualNetworks()
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "VNet Error",
			Content: err.Error(),
			Level:   "error",
		})
	}
	// Build a simple matrix: Name | Location | ResourceGroup
	rows := [][]string{}
	for _, v := range vnets {
		rows = append(rows, []string{v.Name, v.Location, v.ResourceGroup})
	}
	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  "Azure Virtual Networks",
		Rows:   rows,
		Labels: []string{"Name", "Location", "ResourceGroup"},
	})
}

// Example: Show a popup for a firewall error or alarm in the TUI
func ExampleShowFirewallAlarmPopup(errMsg string) string {
	return tui.RenderPopup(tui.PopupMsg{
		Title:   "Firewall Alarm",
		Content: errMsg,
		Level:   "alarm",
	})
}

// Example: Show a matrix graph of Firewalls in the TUI
func ExampleShowFirewallMatrixGraph() string {
	fws, err := ListFirewalls()
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "Firewall Error",
			Content: err.Error(),
			Level:   "error",
		})
	}
	rows := [][]string{}
	for _, f := range fws {
		rows = append(rows, []string{f.Name, f.Location, f.ResourceGroup})
	}
	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  "Azure Firewalls",
		Rows:   rows,
		Labels: []string{"Name", "Location", "ResourceGroup"},
	})
}

// AI-powered summary for VNets
func ExampleShowVNetAISummary() string {
	vnets, err := ListVirtualNetworks()
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "VNet Error",
			Content: err.Error(),
			Level:   "error",
		})
	}
	var names []string
	for _, v := range vnets {
		names = append(names, v.Name)
	}
	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key
	summary, err := aiProvider.SummarizeResourceGroups(names)
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "AI Summary Error",
			Content: err.Error(),
			Level:   "error",
		})
	}
	return tui.RenderPopup(tui.PopupMsg{
		Title:   "AI VNet Summary",
		Content: summary,
		Level:   "info",
	})
}

// AI-powered log analysis for firewall errors
func ExampleShowFirewallAILogAnalysis(logs []string) string {
	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key
	prompt := "Analyze the following Azure Firewall logs for errors, alarms, and recommendations:\n" + strings.Join(logs, "\n")
	result, err := aiProvider.Ask(prompt, "Azure Firewall Log Analysis")
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "AI Log Analysis Error",
			Content: err.Error(),
			Level:   "error",
		})
	}
	return tui.RenderPopup(tui.PopupMsg{
		Title:   "AI Firewall Log Analysis",
		Content: result,
		Level:   "info",
	})
}
