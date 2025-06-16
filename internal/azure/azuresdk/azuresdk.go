package azuresdk

import (
	context "context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type AzureClient struct {
	Cred *azidentity.DefaultAzureCredential
}

func NewAzureClient() (*AzureClient, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	return &AzureClient{Cred: cred}, nil
}

func (c *AzureClient) ListResourceGroups(subscriptionID string) ([]*armresources.ResourceGroup, error) {
	client, err := armresources.NewResourceGroupsClient(subscriptionID, c.Cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	var result []*armresources.ResourceGroup
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
