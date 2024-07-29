// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/advisor/armadvisor"
	"github.com/rs/zerolog/log"
)

// AdvisorResult - Advisor result
type AdvisorResult struct {
	SubscriptionID, SubscriptionName, Name, Type, Category, Description, PotentialBenefits, Risk, LearnMoreLink string
}

// AdvisorScanner - Advisor scanner
type AdvisorScanner struct {
	config *azqr.ScannerConfig
	client *armadvisor.RecommendationsClient
}

// Init - Initializes the Advisor Scanner
func (s *AdvisorScanner) Init(config *azqr.ScannerConfig) error {
	s.config = config
	var err error
	s.client, err = armadvisor.NewRecommendationsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	return nil
}

// ListRecommendations - Lists Azure Advisor recommendations.
func (s *AdvisorScanner) ListRecommendations() ([]AdvisorResult, error) {
	azqr.LogSubscriptionScan(s.config.SubscriptionID, "Advisor Recommendations")

	pager := s.client.NewListPager(&armadvisor.RecommendationsClientListOptions{})

	recommendations := make([]*armadvisor.ResourceRecommendationBase, 0)
	for pager.More() {
		resp, err := pager.NextPage(s.config.Ctx)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, resp.Value...)
	}

	returnRecommendations := make([]AdvisorResult, 0)
	for _, recommendation := range recommendations {
		ar := AdvisorResult{
			SubscriptionID:   s.config.SubscriptionID,
			SubscriptionName: s.config.SubscriptionName,
		}
		if recommendation.Properties.ImpactedValue != nil {
			ar.Name = *recommendation.Properties.ImpactedValue
		}
		if recommendation.Properties.Category != nil {
			ar.Category = string(*recommendation.Properties.Category)
		}
		if recommendation.Properties.ShortDescription != nil && recommendation.Properties.ShortDescription.Problem != nil {
			ar.Description = *recommendation.Properties.ShortDescription.Problem
		}
		if recommendation.Properties.ImpactedField != nil {
			ar.Type = *recommendation.Properties.ImpactedField
		}
		if recommendation.Properties.PotentialBenefits != nil {
			ar.PotentialBenefits = *recommendation.Properties.PotentialBenefits
		}
		if recommendation.Properties.Risk != nil {
			ar.Risk = string(*recommendation.Properties.Risk)
		}
		if recommendation.Properties.LearnMoreLink != nil {
			ar.LearnMoreLink = *recommendation.Properties.LearnMoreLink
		}
		returnRecommendations = append(returnRecommendations, ar)
	}

	return returnRecommendations, nil
}

func (s *AdvisorScanner) Scan(scan bool, config *azqr.ScannerConfig) []AdvisorResult {
	advisorResults := []AdvisorResult{}
	if scan {
		err := s.Init(config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Advisor Scanner")
		}

		rec, err := s.ListRecommendations()
		if err != nil {
			if azqr.ShouldSkipError(err) {
				rec = []AdvisorResult{}
			} else {
				log.Fatal().Err(err).Msg("Failed to list Advisor recommendations")
			}
		}
		advisorResults = append(advisorResults, rec...)
	}
	return advisorResults
}
