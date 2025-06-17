package azuresdk

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type NetworkClient struct {
	cred *azidentity.DefaultAzureCredential
}

func NewNetworkClient() (*NetworkClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return &NetworkClient{cred: cred}, nil
}

// =============================================================================
// VIRTUAL NETWORK OPERATIONS
// =============================================================================

func (c *NetworkClient) ListVirtualNetworks(subscriptionID, resourceGroup string) ([]*armnetwork.VirtualNetwork, error) {
	client, err := armnetwork.NewVirtualNetworksClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.VirtualNetwork
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetVirtualNetwork(subscriptionID, resourceGroup, vnetName string) (*armnetwork.VirtualNetwork, error) {
	client, err := armnetwork.NewVirtualNetworksClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, vnetName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

// =============================================================================
// NETWORK SECURITY GROUP OPERATIONS
// =============================================================================

func (c *NetworkClient) ListNetworkSecurityGroups(subscriptionID, resourceGroup string) ([]*armnetwork.SecurityGroup, error) {
	client, err := armnetwork.NewSecurityGroupsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.SecurityGroup
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetNetworkSecurityGroup(subscriptionID, resourceGroup, nsgName string) (*armnetwork.SecurityGroup, error) {
	client, err := armnetwork.NewSecurityGroupsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, nsgName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SecurityGroup, nil
}

// =============================================================================
// SUBNET OPERATIONS
// =============================================================================

func (c *NetworkClient) ListSubnets(subscriptionID, resourceGroup, vnetName string) ([]*armnetwork.Subnet, error) {
	client, err := armnetwork.NewSubnetsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, vnetName, nil)
	var result []*armnetwork.Subnet
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetSubnet(subscriptionID, resourceGroup, vnetName, subnetName string) (*armnetwork.Subnet, error) {
	client, err := armnetwork.NewSubnetsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

// =============================================================================
// ROUTE TABLE OPERATIONS
// =============================================================================

func (c *NetworkClient) ListRouteTables(subscriptionID, resourceGroup string) ([]*armnetwork.RouteTable, error) {
	client, err := armnetwork.NewRouteTablesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.RouteTable
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetRouteTable(subscriptionID, resourceGroup, routeTableName string) (*armnetwork.RouteTable, error) {
	client, err := armnetwork.NewRouteTablesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, routeTableName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.RouteTable, nil
}

// =============================================================================
// PUBLIC IP OPERATIONS
// =============================================================================

func (c *NetworkClient) ListPublicIPs(subscriptionID, resourceGroup string) ([]*armnetwork.PublicIPAddress, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.PublicIPAddress
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetPublicIP(subscriptionID, resourceGroup, publicIPName string) (*armnetwork.PublicIPAddress, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, publicIPName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.PublicIPAddress, nil
}

// =============================================================================
// NETWORK INTERFACE OPERATIONS
// =============================================================================

func (c *NetworkClient) ListNetworkInterfaces(subscriptionID, resourceGroup string) ([]*armnetwork.Interface, error) {
	client, err := armnetwork.NewInterfacesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.Interface
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetNetworkInterface(subscriptionID, resourceGroup, nicName string) (*armnetwork.Interface, error) {
	client, err := armnetwork.NewInterfacesClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, nicName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Interface, nil
}

// =============================================================================
// LOAD BALANCER OPERATIONS
// =============================================================================

func (c *NetworkClient) ListLoadBalancers(subscriptionID, resourceGroup string) ([]*armnetwork.LoadBalancer, error) {
	client, err := armnetwork.NewLoadBalancersClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.LoadBalancer
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetLoadBalancer(subscriptionID, resourceGroup, lbName string) (*armnetwork.LoadBalancer, error) {
	client, err := armnetwork.NewLoadBalancersClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, lbName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.LoadBalancer, nil
}

// =============================================================================
// FIREWALL OPERATIONS (ORIGINAL)
// =============================================================================

func (c *NetworkClient) ListFirewalls(subscriptionID, resourceGroup string) ([]*armnetwork.AzureFirewall, error) {
	client, err := armnetwork.NewAzureFirewallsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.AzureFirewall
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func (c *NetworkClient) GetFirewall(subscriptionID, resourceGroup, firewallName string) (*armnetwork.AzureFirewall, error) {
	client, err := armnetwork.NewAzureFirewallsClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	resp, err := client.Get(ctx, resourceGroup, firewallName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.AzureFirewall, nil
}

// =============================================================================
// NETWORK WATCHER OPERATIONS
// =============================================================================

func (c *NetworkClient) ListNetworkWatchers(subscriptionID, resourceGroup string) ([]*armnetwork.Watcher, error) {
	client, err := armnetwork.NewWatchersClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.Watcher
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

// =============================================================================
// VPN GATEWAY OPERATIONS
// =============================================================================

func (c *NetworkClient) ListVpnGateways(subscriptionID, resourceGroup string) ([]*armnetwork.VirtualNetworkGateway, error) {
	client, err := armnetwork.NewVirtualNetworkGatewaysClient(subscriptionID, c.cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceGroup, nil)
	var result []*armnetwork.VirtualNetworkGateway
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}
