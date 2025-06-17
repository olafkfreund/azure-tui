package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	ai "github.com/olafkfreund/azure-tui/internal/openai"
	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Enhanced network resource structures

type VirtualNetwork struct {
	Name          string                 `json:"name"`
	Location      string                 `json:"location"`
	ResourceGroup string                 `json:"resourceGroup"`
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AddressSpace  []string               `json:"addressSpace"`
	Subnets       []Subnet               `json:"subnets"`
	DnsServers    []string               `json:"dnsServers"`
	Tags          map[string]interface{} `json:"tags"`
	Properties    map[string]interface{} `json:"properties"`
}

type Subnet struct {
	Name             string   `json:"name"`
	ID               string   `json:"id"`
	AddressPrefix    string   `json:"addressPrefix"`
	NSGName          string   `json:"nsgName"`
	RouteTableName   string   `json:"routeTableName"`
	Delegations      []string `json:"delegations"`
	PrivateEndpoints []string `json:"privateEndpoints"`
}

type NetworkSecurityGroup struct {
	Name              string                 `json:"name"`
	Location          string                 `json:"location"`
	ResourceGroup     string                 `json:"resourceGroup"`
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	SecurityRules     []SecurityRule         `json:"securityRules"`
	DefaultRules      []SecurityRule         `json:"defaultSecurityRules"`
	NetworkInterfaces []string               `json:"networkInterfaces"`
	Subnets           []string               `json:"subnets"`
	Tags              map[string]interface{} `json:"tags"`
}

type SecurityRule struct {
	Name                     string `json:"name"`
	Priority                 int    `json:"priority"`
	Direction                string `json:"direction"`
	Access                   string `json:"access"`
	Protocol                 string `json:"protocol"`
	SourcePortRange          string `json:"sourcePortRange"`
	DestinationPortRange     string `json:"destinationPortRange"`
	SourceAddressPrefix      string `json:"sourceAddressPrefix"`
	DestinationAddressPrefix string `json:"destinationAddressPrefix"`
	Description              string `json:"description"`
}

type RouteTable struct {
	Name          string                 `json:"name"`
	Location      string                 `json:"location"`
	ResourceGroup string                 `json:"resourceGroup"`
	ID            string                 `json:"id"`
	Routes        []Route                `json:"routes"`
	Subnets       []string               `json:"subnets"`
	Tags          map[string]interface{} `json:"tags"`
}

type Route struct {
	Name           string `json:"name"`
	AddressPrefix  string `json:"addressPrefix"`
	NextHopType    string `json:"nextHopType"`
	NextHopAddress string `json:"nextHopIpAddress"`
}

type PublicIP struct {
	Name               string                 `json:"name"`
	Location           string                 `json:"location"`
	ResourceGroup      string                 `json:"resourceGroup"`
	ID                 string                 `json:"id"`
	IPAddress          string                 `json:"ipAddress"`
	AllocationMethod   string                 `json:"publicIPAllocationMethod"`
	SKU                string                 `json:"sku"`
	Zone               []string               `json:"zones"`
	AssociatedResource string                 `json:"associatedResource"`
	Tags               map[string]interface{} `json:"tags"`
}

type NetworkInterface struct {
	Name               string                 `json:"name"`
	Location           string                 `json:"location"`
	ResourceGroup      string                 `json:"resourceGroup"`
	ID                 string                 `json:"id"`
	PrivateIPAddress   string                 `json:"privateIPAddress"`
	PublicIPAddress    string                 `json:"publicIPAddress"`
	SubnetID           string                 `json:"subnetId"`
	NSGName            string                 `json:"networkSecurityGroup"`
	VirtualMachine     string                 `json:"virtualMachine"`
	EnableIPForwarding bool                   `json:"enableIPForwarding"`
	Tags               map[string]interface{} `json:"tags"`
}

type LoadBalancer struct {
	Name               string                 `json:"name"`
	Location           string                 `json:"location"`
	ResourceGroup      string                 `json:"resourceGroup"`
	ID                 string                 `json:"id"`
	SKU                string                 `json:"sku"`
	Type               string                 `json:"type"`
	FrontendIPs        []FrontendIP           `json:"frontendIPConfigurations"`
	BackendPools       []BackendPool          `json:"backendAddressPools"`
	LoadBalancingRules []LoadBalancingRule    `json:"loadBalancingRules"`
	Probes             []Probe                `json:"probes"`
	Tags               map[string]interface{} `json:"tags"`
}

type FrontendIP struct {
	Name             string `json:"name"`
	PrivateIPAddress string `json:"privateIPAddress"`
	PublicIPAddress  string `json:"publicIPAddress"`
	SubnetID         string `json:"subnetId"`
}

type BackendPool struct {
	Name      string   `json:"name"`
	Resources []string `json:"backendAddresses"`
}

type LoadBalancingRule struct {
	Name                 string `json:"name"`
	Protocol             string `json:"protocol"`
	FrontendPort         int    `json:"frontendPort"`
	BackendPort          int    `json:"backendPort"`
	EnableFloatingIP     bool   `json:"enableFloatingIP"`
	IdleTimeoutInMinutes int    `json:"idleTimeoutInMinutes"`
}

type Probe struct {
	Name              string `json:"name"`
	Protocol          string `json:"protocol"`
	Port              int    `json:"port"`
	Path              string `json:"requestPath"`
	IntervalInSeconds int    `json:"intervalInSeconds"`
	NumberOfProbes    int    `json:"numberOfProbes"`
}

type Firewall struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

// Network Dashboard represents a comprehensive network overview
type NetworkDashboard struct {
	VirtualNetworks       []VirtualNetwork       `json:"virtualNetworks"`
	NetworkSecurityGroups []NetworkSecurityGroup `json:"networkSecurityGroups"`
	RouteTables           []RouteTable           `json:"routeTables"`
	PublicIPs             []PublicIP             `json:"publicIPs"`
	NetworkInterfaces     []NetworkInterface     `json:"networkInterfaces"`
	LoadBalancers         []LoadBalancer         `json:"loadBalancers"`
	Firewalls             []Firewall             `json:"firewalls"`
	Summary               NetworkSummary         `json:"summary"`
	Topology              NetworkTopology        `json:"topology"`
}

type NetworkSummary struct {
	TotalVNets      int `json:"totalVNets"`
	TotalSubnets    int `json:"totalSubnets"`
	TotalNSGs       int `json:"totalNSGs"`
	TotalRoutes     int `json:"totalRoutes"`
	TotalPublicIPs  int `json:"totalPublicIPs"`
	TotalPrivateIPs int `json:"totalPrivateIPs"`
}

type NetworkTopology struct {
	VNetConnections []VNetConnection `json:"vnetConnections"`
	PeeringStatus   []PeeringStatus  `json:"peeringStatus"`
	GatewayStatus   []GatewayStatus  `json:"gatewayStatus"`
}

type VNetConnection struct {
	SourceVNet     string `json:"sourceVNet"`
	TargetVNet     string `json:"targetVNet"`
	ConnectionType string `json:"connectionType"`
	Status         string `json:"status"`
}

type PeeringStatus struct {
	VNetName          string `json:"vnetName"`
	PeerVNetName      string `json:"peerVNetName"`
	PeeringState      string `json:"peeringState"`
	ProvisioningState string `json:"provisioningState"`
}

type GatewayStatus struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	VNetName string `json:"vnetName"`
}

// =============================================================================
// COMPREHENSIVE NETWORK RESOURCE MANAGEMENT FUNCTIONS
// =============================================================================

// GetNetworkDashboard retrieves comprehensive network information for dashboard display
func GetNetworkDashboard(resourceGroup string) (*NetworkDashboard, error) {
	dashboard := &NetworkDashboard{}

	// Get all network resources in parallel
	vnets, err := ListVirtualNetworks()
	if err == nil {
		dashboard.VirtualNetworks = vnets
	}

	nsgs, err := ListNetworkSecurityGroups()
	if err == nil {
		dashboard.NetworkSecurityGroups = nsgs
	}

	routeTables, err := ListRouteTables()
	if err == nil {
		dashboard.RouteTables = routeTables
	}

	publicIPs, err := ListPublicIPs()
	if err == nil {
		dashboard.PublicIPs = publicIPs
	}

	nics, err := ListNetworkInterfaces()
	if err == nil {
		dashboard.NetworkInterfaces = nics
	}

	lbs, err := ListLoadBalancers()
	if err == nil {
		dashboard.LoadBalancers = lbs
	}

	firewalls, err := ListFirewalls()
	if err == nil {
		dashboard.Firewalls = firewalls
	}

	// Calculate summary
	dashboard.Summary = calculateNetworkSummary(dashboard)

	// Get topology information
	dashboard.Topology = getNetworkTopology(dashboard)

	return dashboard, nil
}

// =============================================================================
// VIRTUAL NETWORK MANAGEMENT
// =============================================================================

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

func GetVirtualNetworkDetails(name, resourceGroup string) (*VirtualNetwork, error) {
	cmd := exec.Command("az", "network", "vnet", "show", "--name", name, "--resource-group", resourceGroup, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var vnet VirtualNetwork
	if err := json.Unmarshal(out, &vnet); err != nil {
		return nil, err
	}
	return &vnet, nil
}

func CreateVirtualNetwork(name, group, location string) error {
	return exec.Command("az", "network", "vnet", "create", "--name", name, "--resource-group", group, "--location", location, "--address-prefix", "10.0.0.0/16").Run()
}

func CreateVirtualNetworkAdvanced(name, group, location string, addressPrefixes []string, dnsServers []string) error {
	args := []string{"network", "vnet", "create", "--name", name, "--resource-group", group, "--location", location}

	if len(addressPrefixes) > 0 {
		args = append(args, "--address-prefixes")
		args = append(args, addressPrefixes...)
	} else {
		args = append(args, "--address-prefix", "10.0.0.0/16")
	}

	if len(dnsServers) > 0 {
		args = append(args, "--dns-servers")
		args = append(args, dnsServers...)
	}

	return exec.Command("az", args...).Run()
}

func DeleteVirtualNetwork(name, group string) error {
	return exec.Command("az", "network", "vnet", "delete", "--name", name, "--resource-group", group, "--yes").Run()
}

// =============================================================================
// SUBNET MANAGEMENT
// =============================================================================

func ListSubnets(vnetName, resourceGroup string) ([]Subnet, error) {
	cmd := exec.Command("az", "network", "vnet", "subnet", "list", "--vnet-name", vnetName, "--resource-group", resourceGroup, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var subnets []Subnet
	if err := json.Unmarshal(out, &subnets); err != nil {
		return nil, err
	}
	return subnets, nil
}

func CreateSubnet(name, vnetName, resourceGroup, addressPrefix string) error {
	return exec.Command("az", "network", "vnet", "subnet", "create",
		"--name", name,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--address-prefix", addressPrefix).Run()
}

func AssociateSubnetWithNSG(subnetName, vnetName, resourceGroup, nsgName string) error {
	nsgID := fmt.Sprintf("/subscriptions/$(az account show --query id -o tsv)/resourceGroups/%s/providers/Microsoft.Network/networkSecurityGroups/%s", resourceGroup, nsgName)
	return exec.Command("az", "network", "vnet", "subnet", "update",
		"--name", subnetName,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--network-security-group", nsgID).Run()
}

func AssociateSubnetWithRouteTable(subnetName, vnetName, resourceGroup, routeTableName string) error {
	routeTableID := fmt.Sprintf("/subscriptions/$(az account show --query id -o tsv)/resourceGroups/%s/providers/Microsoft.Network/routeTables/%s", resourceGroup, routeTableName)
	return exec.Command("az", "network", "vnet", "subnet", "update",
		"--name", subnetName,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--route-table", routeTableID).Run()
}

func DeleteSubnet(name, vnetName, resourceGroup string) error {
	return exec.Command("az", "network", "vnet", "subnet", "delete",
		"--name", name,
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup).Run()
}

// =============================================================================
// NETWORK SECURITY GROUP MANAGEMENT
// =============================================================================

func ListNetworkSecurityGroups() ([]NetworkSecurityGroup, error) {
	cmd := exec.Command("az", "network", "nsg", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var nsgs []NetworkSecurityGroup
	if err := json.Unmarshal(out, &nsgs); err != nil {
		return nil, err
	}
	return nsgs, nil
}

func GetNetworkSecurityGroupDetails(name, resourceGroup string) (*NetworkSecurityGroup, error) {
	cmd := exec.Command("az", "network", "nsg", "show", "--name", name, "--resource-group", resourceGroup, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var nsg NetworkSecurityGroup
	if err := json.Unmarshal(out, &nsg); err != nil {
		return nil, err
	}
	return &nsg, nil
}

func CreateNetworkSecurityGroup(name, resourceGroup, location string) error {
	return exec.Command("az", "network", "nsg", "create", "--name", name, "--resource-group", resourceGroup, "--location", location).Run()
}

func CreateSecurityRule(nsgName, resourceGroup, ruleName string, priority int, direction, access, protocol, sourcePort, destPort, sourceAddress, destAddress string) error {
	return exec.Command("az", "network", "nsg", "rule", "create",
		"--nsg-name", nsgName,
		"--resource-group", resourceGroup,
		"--name", ruleName,
		"--priority", fmt.Sprintf("%d", priority),
		"--direction", direction,
		"--access", access,
		"--protocol", protocol,
		"--source-port-ranges", sourcePort,
		"--destination-port-ranges", destPort,
		"--source-address-prefixes", sourceAddress,
		"--destination-address-prefixes", destAddress).Run()
}

func DeleteSecurityRule(nsgName, resourceGroup, ruleName string) error {
	return exec.Command("az", "network", "nsg", "rule", "delete",
		"--nsg-name", nsgName,
		"--resource-group", resourceGroup,
		"--name", ruleName).Run()
}

func DeleteNetworkSecurityGroup(name, resourceGroup string) error {
	return exec.Command("az", "network", "nsg", "delete", "--name", name, "--resource-group", resourceGroup).Run()
}

// =============================================================================
// ROUTE TABLE MANAGEMENT
// =============================================================================

func ListRouteTables() ([]RouteTable, error) {
	cmd := exec.Command("az", "network", "route-table", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var routeTables []RouteTable
	if err := json.Unmarshal(out, &routeTables); err != nil {
		return nil, err
	}
	return routeTables, nil
}

func CreateRouteTable(name, resourceGroup, location string) error {
	return exec.Command("az", "network", "route-table", "create", "--name", name, "--resource-group", resourceGroup, "--location", location).Run()
}

func CreateRoute(routeTableName, resourceGroup, routeName, addressPrefix, nextHopType, nextHopAddress string) error {
	args := []string{"network", "route-table", "route", "create",
		"--route-table-name", routeTableName,
		"--resource-group", resourceGroup,
		"--name", routeName,
		"--address-prefix", addressPrefix,
		"--next-hop-type", nextHopType}

	if nextHopAddress != "" {
		args = append(args, "--next-hop-ip-address", nextHopAddress)
	}

	return exec.Command("az", args...).Run()
}

func DeleteRoute(routeTableName, resourceGroup, routeName string) error {
	return exec.Command("az", "network", "route-table", "route", "delete",
		"--route-table-name", routeTableName,
		"--resource-group", resourceGroup,
		"--name", routeName).Run()
}

func DeleteRouteTable(name, resourceGroup string) error {
	return exec.Command("az", "network", "route-table", "delete", "--name", name, "--resource-group", resourceGroup).Run()
}

// =============================================================================
// PUBLIC IP MANAGEMENT
// =============================================================================

func ListPublicIPs() ([]PublicIP, error) {
	cmd := exec.Command("az", "network", "public-ip", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var publicIPs []PublicIP
	if err := json.Unmarshal(out, &publicIPs); err != nil {
		return nil, err
	}
	return publicIPs, nil
}

func CreatePublicIP(name, resourceGroup, location, allocationMethod, sku string) error {
	return exec.Command("az", "network", "public-ip", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--allocation-method", allocationMethod,
		"--sku", sku).Run()
}

func DeletePublicIP(name, resourceGroup string) error {
	return exec.Command("az", "network", "public-ip", "delete", "--name", name, "--resource-group", resourceGroup).Run()
}

// =============================================================================
// NETWORK INTERFACE MANAGEMENT
// =============================================================================

func ListNetworkInterfaces() ([]NetworkInterface, error) {
	cmd := exec.Command("az", "network", "nic", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var nics []NetworkInterface
	if err := json.Unmarshal(out, &nics); err != nil {
		return nil, err
	}
	return nics, nil
}

func CreateNetworkInterface(name, resourceGroup, location, subnetID, publicIPName, nsgName string) error {
	args := []string{"network", "nic", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--subnet", subnetID}

	if publicIPName != "" {
		args = append(args, "--public-ip-address", publicIPName)
	}

	if nsgName != "" {
		args = append(args, "--network-security-group", nsgName)
	}

	return exec.Command("az", args...).Run()
}

func DeleteNetworkInterface(name, resourceGroup string) error {
	return exec.Command("az", "network", "nic", "delete", "--name", name, "--resource-group", resourceGroup).Run()
}

// =============================================================================
// LOAD BALANCER MANAGEMENT
// =============================================================================

func ListLoadBalancers() ([]LoadBalancer, error) {
	cmd := exec.Command("az", "network", "lb", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var lbs []LoadBalancer
	if err := json.Unmarshal(out, &lbs); err != nil {
		return nil, err
	}
	return lbs, nil
}

func CreateLoadBalancer(name, resourceGroup, location, sku, publicIPName string) error {
	args := []string{"network", "lb", "create",
		"--name", name,
		"--resource-group", resourceGroup,
		"--location", location,
		"--sku", sku}

	if publicIPName != "" {
		args = append(args, "--public-ip-address", publicIPName)
	}

	return exec.Command("az", args...).Run()
}

func DeleteLoadBalancer(name, resourceGroup string) error {
	return exec.Command("az", "network", "lb", "delete", "--name", name, "--resource-group", resourceGroup).Run()
}

// =============================================================================
// FIREWALL MANAGEMENT (Original functions)
// =============================================================================

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

// =============================================================================
// NETWORK TOPOLOGY AND ANALYSIS
// =============================================================================

func calculateNetworkSummary(dashboard *NetworkDashboard) NetworkSummary {
	summary := NetworkSummary{
		TotalVNets:     len(dashboard.VirtualNetworks),
		TotalNSGs:      len(dashboard.NetworkSecurityGroups),
		TotalRoutes:    0,
		TotalPublicIPs: len(dashboard.PublicIPs),
	}

	// Count subnets and routes
	for _, vnet := range dashboard.VirtualNetworks {
		summary.TotalSubnets += len(vnet.Subnets)
	}

	for _, rt := range dashboard.RouteTables {
		summary.TotalRoutes += len(rt.Routes)
	}

	// Count private IPs from network interfaces
	for _, nic := range dashboard.NetworkInterfaces {
		if nic.PrivateIPAddress != "" {
			summary.TotalPrivateIPs++
		}
	}

	return summary
}

func getNetworkTopology(dashboard *NetworkDashboard) NetworkTopology {
	topology := NetworkTopology{
		VNetConnections: []VNetConnection{},
		PeeringStatus:   []PeeringStatus{},
		GatewayStatus:   []GatewayStatus{},
	}

	// Get VNet peering information
	for _, vnet := range dashboard.VirtualNetworks {
		peeringsOut, err := getVNetPeerings(vnet.Name, vnet.ResourceGroup)
		if err == nil {
			topology.PeeringStatus = append(topology.PeeringStatus, peeringsOut...)
		}
	}

	return topology
}

func getVNetPeerings(vnetName, resourceGroup string) ([]PeeringStatus, error) {
	cmd := exec.Command("az", "network", "vnet", "peering", "list",
		"--vnet-name", vnetName,
		"--resource-group", resourceGroup,
		"--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var peerings []PeeringStatus
	if err := json.Unmarshal(out, &peerings); err != nil {
		return nil, err
	}

	return peerings, nil
}

// =============================================================================
// TERRAFORM/BICEP CODE GENERATION
// =============================================================================

func GenerateVNetTerraform(vnet VirtualNetwork) string {
	template := `
resource "azurerm_virtual_network" "%s" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"
  address_space       = [%s]

  tags = {
    Environment = "Production"
    ManagedBy   = "Terraform"
  }
}
`
	addressSpace := `"` + strings.Join(vnet.AddressSpace, `", "`) + `"`
	return fmt.Sprintf(template,
		strings.ReplaceAll(vnet.Name, "-", "_"),
		vnet.Name,
		vnet.Location,
		vnet.ResourceGroup,
		addressSpace)
}

func GenerateNSGTerraform(nsg NetworkSecurityGroup) string {
	template := `
resource "azurerm_network_security_group" "%s" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"

  tags = {
    Environment = "Production" 
    ManagedBy   = "Terraform"
  }
}
`
	return fmt.Sprintf(template,
		strings.ReplaceAll(nsg.Name, "-", "_"),
		nsg.Name,
		nsg.Location,
		nsg.ResourceGroup)
}

func GenerateVNetBicep(vnet VirtualNetwork) string {
	template := `
param location string = '%s'
param vnetName string = '%s'

resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-09-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [%s]
    }
  }
  tags: {
    Environment: 'Production'
    ManagedBy: 'Bicep'
  }
}

output vnetId string = virtualNetwork.id
`
	addressPrefixes := "'" + strings.Join(vnet.AddressSpace, "', '") + "'"
	return fmt.Sprintf(template, vnet.Location, vnet.Name, addressPrefixes)
}

// =============================================================================
// TUI RENDERING FUNCTIONS FOR NETWORK DASHBOARD
// =============================================================================

// RenderNetworkDashboard renders a comprehensive network resource dashboard
func RenderNetworkDashboard() string {
	dashboard, err := GetNetworkDashboard("")
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "Network Dashboard Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	// Build comprehensive network matrix
	rows := [][]string{}

	// Header row
	rows = append(rows, []string{"Resource Type", "Name", "Location", "Resource Group", "Status", "Associated Resources"})

	// Virtual Networks with subnet details
	for _, vnet := range dashboard.VirtualNetworks {
		associatedResources := fmt.Sprintf("%d subnets", len(vnet.Subnets))
		if len(vnet.DnsServers) > 0 {
			associatedResources += fmt.Sprintf(", %d DNS servers", len(vnet.DnsServers))
		}
		rows = append(rows, []string{"🌐 VNet", vnet.Name, vnet.Location, vnet.ResourceGroup, "Active", associatedResources})

		// Add subnet details
		for _, subnet := range vnet.Subnets {
			subnetDetails := subnet.AddressPrefix
			if subnet.NSGName != "" {
				subnetDetails += fmt.Sprintf(" (NSG: %s)", subnet.NSGName)
			}
			if subnet.RouteTableName != "" {
				subnetDetails += fmt.Sprintf(" (RT: %s)", subnet.RouteTableName)
			}
			rows = append(rows, []string{"  ┗━ 🏠 Subnet", subnet.Name, "-", "-", "Active", subnetDetails})
		}
	}

	// Network Security Groups with rule count
	for _, nsg := range dashboard.NetworkSecurityGroups {
		ruleCount := fmt.Sprintf("%d rules", len(nsg.SecurityRules))
		associatedCount := fmt.Sprintf("%d subnets, %d NICs", len(nsg.Subnets), len(nsg.NetworkInterfaces))
		rows = append(rows, []string{"🔒 NSG", nsg.Name, nsg.Location, nsg.ResourceGroup, "Active", ruleCount + ", " + associatedCount})
	}

	// Route Tables
	for _, rt := range dashboard.RouteTables {
		routeCount := fmt.Sprintf("%d routes", len(rt.Routes))
		subnetCount := fmt.Sprintf("%d subnets", len(rt.Subnets))
		rows = append(rows, []string{"🗺️ Route Table", rt.Name, rt.Location, rt.ResourceGroup, "Active", routeCount + ", " + subnetCount})
	}

	// Public IPs
	for _, pip := range dashboard.PublicIPs {
		details := pip.AllocationMethod
		if pip.IPAddress != "" {
			details += fmt.Sprintf(" (%s)", pip.IPAddress)
		}
		if pip.AssociatedResource != "" {
			details += fmt.Sprintf(" → %s", pip.AssociatedResource)
		}
		rows = append(rows, []string{"🌍 Public IP", pip.Name, pip.Location, pip.ResourceGroup, "Active", details})
	}

	// Network Interfaces
	for _, nic := range dashboard.NetworkInterfaces {
		details := nic.PrivateIPAddress
		if nic.PublicIPAddress != "" {
			details += fmt.Sprintf(" / %s", nic.PublicIPAddress)
		}
		if nic.VirtualMachine != "" {
			details += fmt.Sprintf(" → %s", nic.VirtualMachine)
		}
		rows = append(rows, []string{"🔗 Network Interface", nic.Name, nic.Location, nic.ResourceGroup, "Active", details})
	}

	// Load Balancers
	for _, lb := range dashboard.LoadBalancers {
		details := fmt.Sprintf("%s (%d frontends, %d backends)", lb.SKU, len(lb.FrontendIPs), len(lb.BackendPools))
		rows = append(rows, []string{"⚖️ Load Balancer", lb.Name, lb.Location, lb.ResourceGroup, "Active", details})
	}

	// Firewalls
	for _, fw := range dashboard.Firewalls {
		rows = append(rows, []string{"🔥 Firewall", fw.Name, fw.Location, fw.ResourceGroup, "Active", "Azure Firewall"})
	}

	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title: fmt.Sprintf("🌐 Azure Network Dashboard - %d VNets, %d NSGs, %d Routes",
			dashboard.Summary.TotalVNets,
			dashboard.Summary.TotalNSGs,
			dashboard.Summary.TotalRoutes),
		Rows:   rows,
		Labels: []string{"Type", "Name", "Location", "Resource Group", "Status", "Details"},
	})
}

// RenderVNetDetails renders detailed information for a specific VNet
func RenderVNetDetails(vnetName, resourceGroup string) string {
	vnet, err := GetVirtualNetworkDetails(vnetName, resourceGroup)
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "VNet Details Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	rows := [][]string{}

	// Basic information
	rows = append(rows, []string{"Property", "Value"})
	rows = append(rows, []string{"Name", vnet.Name})
	rows = append(rows, []string{"Resource Group", vnet.ResourceGroup})
	rows = append(rows, []string{"Location", vnet.Location})
	rows = append(rows, []string{"Resource ID", vnet.ID})
	rows = append(rows, []string{"Address Space", strings.Join(vnet.AddressSpace, ", ")})

	if len(vnet.DnsServers) > 0 {
		rows = append(rows, []string{"DNS Servers", strings.Join(vnet.DnsServers, ", ")})
	}

	// Subnet details
	if len(vnet.Subnets) > 0 {
		rows = append(rows, []string{"", ""}) // Spacer
		rows = append(rows, []string{"SUBNETS", ""})
		for _, subnet := range vnet.Subnets {
			subnetInfo := subnet.AddressPrefix
			if subnet.NSGName != "" {
				subnetInfo += fmt.Sprintf(" (NSG: %s)", subnet.NSGName)
			}
			rows = append(rows, []string{fmt.Sprintf("└─ %s", subnet.Name), subnetInfo})
		}
	}

	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  fmt.Sprintf("🌐 Virtual Network: %s", vnetName),
		Rows:   rows,
		Labels: []string{"Property", "Value"},
	})
}

// RenderNSGDetails renders detailed information for a specific Network Security Group
func RenderNSGDetails(nsgName, resourceGroup string) string {
	nsg, err := GetNetworkSecurityGroupDetails(nsgName, resourceGroup)
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "NSG Details Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	rows := [][]string{}

	// Headers
	rows = append(rows, []string{"Rule Name", "Priority", "Direction", "Access", "Protocol", "Source", "Destination", "Ports"})

	// Security rules
	for _, rule := range nsg.SecurityRules {
		sourceInfo := rule.SourceAddressPrefix
		if rule.SourcePortRange != "*" {
			sourceInfo += fmt.Sprintf(":%s", rule.SourcePortRange)
		}

		destInfo := rule.DestinationAddressPrefix
		if rule.DestinationPortRange != "*" {
			destInfo += fmt.Sprintf(":%s", rule.DestinationPortRange)
		}

		rows = append(rows, []string{
			rule.Name,
			fmt.Sprintf("%d", rule.Priority),
			rule.Direction,
			rule.Access,
			rule.Protocol,
			sourceInfo,
			destInfo,
			rule.DestinationPortRange,
		})
	}

	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  fmt.Sprintf("🔒 Network Security Group: %s (%d rules)", nsgName, len(nsg.SecurityRules)),
		Rows:   rows,
		Labels: []string{"Rule", "Priority", "Direction", "Access", "Protocol", "Source", "Destination", "Ports"},
	})
}

// RenderNetworkTopology renders network topology and connections
func RenderNetworkTopology() string {
	dashboard, err := GetNetworkDashboard("")
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "Network Topology Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	rows := [][]string{}
	rows = append(rows, []string{"Connection Type", "Source", "Target", "Status", "Details"})

	// VNet Peerings
	for _, peering := range dashboard.Topology.PeeringStatus {
		rows = append(rows, []string{
			"🔗 VNet Peering",
			peering.VNetName,
			peering.PeerVNetName,
			peering.PeeringState,
			peering.ProvisioningState,
		})
	}

	// Gateway connections
	for _, gateway := range dashboard.Topology.GatewayStatus {
		rows = append(rows, []string{
			fmt.Sprintf("🚪 %s Gateway", gateway.Type),
			gateway.VNetName,
			gateway.Name,
			gateway.Status,
			"Gateway Connection",
		})
	}

	// Add subnet to NSG associations
	for _, vnet := range dashboard.VirtualNetworks {
		for _, subnet := range vnet.Subnets {
			if subnet.NSGName != "" {
				rows = append(rows, []string{
					"🔒 NSG Association",
					fmt.Sprintf("%s/%s", vnet.Name, subnet.Name),
					subnet.NSGName,
					"Active",
					"Subnet Protection",
				})
			}
		}
	}

	return tui.RenderMatrixGraph(tui.MatrixGraphMsg{
		Title:  "🗺️ Network Topology & Connections",
		Rows:   rows,
		Labels: []string{"Type", "Source", "Target", "Status", "Details"},
	})
}

// =============================================================================
// AI-POWERED NETWORK ANALYSIS
// =============================================================================

// RenderNetworkAIAnalysis provides AI-powered insights for network configuration
func RenderNetworkAIAnalysis() string {
	dashboard, err := GetNetworkDashboard("")
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "Network Analysis Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	// Prepare network summary for AI analysis
	summary := fmt.Sprintf(`
Azure Network Environment Analysis:
- Virtual Networks: %d
- Subnets: %d
- Network Security Groups: %d
- Route Tables: %d routes
- Public IPs: %d
- Private IPs: %d
- Load Balancers: %d

Network Security Analysis:
`, dashboard.Summary.TotalVNets, dashboard.Summary.TotalSubnets, dashboard.Summary.TotalNSGs,
		dashboard.Summary.TotalRoutes, dashboard.Summary.TotalPublicIPs, dashboard.Summary.TotalPrivateIPs, len(dashboard.LoadBalancers))

	// Add NSG rule analysis
	for _, nsg := range dashboard.NetworkSecurityGroups {
		summary += fmt.Sprintf("NSG '%s': %d custom rules\n", nsg.Name, len(nsg.SecurityRules))
	}

	// Add topology information
	summary += fmt.Sprintf("\nNetwork Topology:\n")
	for _, peering := range dashboard.Topology.PeeringStatus {
		summary += fmt.Sprintf("Peering: %s <-> %s (%s)\n", peering.VNetName, peering.PeerVNetName, peering.PeeringState)
	}

	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key
	prompt := "Analyze this Azure network configuration and provide recommendations for security, optimization, and best practices:\n" + summary
	result, err := aiProvider.Ask(prompt, "Azure Network Analysis")
	if err != nil {
		return tui.RenderPopup(tui.PopupMsg{
			Title:   "AI Analysis Error",
			Content: err.Error(),
			Level:   "error",
		})
	}

	return tui.RenderPopup(tui.PopupMsg{
		Title:   "🤖 AI Network Analysis & Recommendations",
		Content: result,
		Level:   "info",
	})
}

// =============================================================================
// ORIGINAL FUNCTIONS (PRESERVED FOR COMPATIBILITY)
// =============================================================================

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
