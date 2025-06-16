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
