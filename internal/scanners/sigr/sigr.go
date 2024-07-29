// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

// SignalRScanner - Scanner for SignalR
type SignalRScanner struct {
	config        *azqr.ScannerConfig
	signalrClient *armsignalr.Client
}

// Init - Initializes the SignalRScanner
func (c *SignalRScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.signalrClient, err = armsignalr.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all SignalR in a Resource Group
func (c *SignalRScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	signalr, err := c.listSignalR(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, signalr := range signalr {
		rr := engine.EvaluateRecommendations(rules, signalr, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *signalr.Name,
			Type:             *signalr.Type,
			Location:         *signalr.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *SignalRScanner) listSignalR(resourceGroupName string) ([]*armsignalr.ResourceInfo, error) {
	pager := c.signalrClient.NewListByResourceGroupPager(resourceGroupName, nil)

	signalrs := make([]*armsignalr.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		signalrs = append(signalrs, resp.Value...)
	}
	return signalrs, nil
}

func (a *SignalRScanner) ResourceTypes() []string {
	return []string{"Microsoft.SignalRService/SignalR"}
}
