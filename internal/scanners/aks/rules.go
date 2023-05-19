// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aks

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/cmendible/azqr/internal/scanners"
)

// GetRules - Returns the rules for the AKSScanner
func (a *AKSScanner) GetRules() map[string]scanners.AzureRule {
	return map[string]scanners.AzureRule{
		"DiagnosticSettings": {
			Id:          "aks-001",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityDiagnosticLogs,
			Description: "AKS Cluster should have diagnostic settings enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				service := target.(*armcontainerservice.ManagedCluster)
				hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*service.ID)
				if err != nil {
					log.Fatalf("Error checking diagnostic settings for service %s: %s", *service.Name, err)
				}

				return !hasDiagnostics, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs",
		},
		"AvailabilityZones": {
			Id:          "aks-002",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityAvailabilityZones,
			Description: "AKS Cluster should have availability zones enabled",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				cluster := target.(*armcontainerservice.ManagedCluster)
				zones := true
				for _, profile := range cluster.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
					}
				}
				return !zones, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/availability-zones",
		},
		"SLA": {
			Id:          "aks-003",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySLA,
			Description: "AKS Cluster should have an SLA",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				zones := true
				for _, profile := range c.Properties.AgentPoolProfiles {
					if profile.AvailabilityZones == nil || (profile.AvailabilityZones != nil && len(profile.AvailabilityZones) <= 1) {
						zones = false
					}
				}

				sku := "Free"
				if c.SKU != nil && c.SKU.Tier != nil {
					sku = string(*c.SKU.Tier)
				}
				sla := "None"
				if !strings.Contains(sku, "Free") {
					sla = "99.9%"
					if zones {
						sla = "99.95%"
					}
				}
				return sla == "None", sla
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers#uptime-sla-terms-and-conditions",
		},
		"Private": {
			Id:          "aks-004",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityPrivateEndpoint,
			Description: "AKS Cluster should be private",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				pe := c.Properties.APIServerAccessProfile != nil && c.Properties.APIServerAccessProfile.EnablePrivateCluster != nil && *c.Properties.APIServerAccessProfile.EnablePrivateCluster
				return !pe, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/private-clusters",
		},
		"SKU": {
			Id:          "aks-005",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilitySKU,
			Description: "AKS Production Cluster should use Standard SKU",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				sku := "Free"
				if c.SKU != nil && c.SKU.Tier != nil {
					sku = string(*c.SKU.Tier)
				}
				return sku == "Free", sku
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers",
		},
		"CAF": {
			Id:          "aks-006",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceCAF,
			Description: "AKS Name should comply with naming conventions",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				caf := strings.HasPrefix(*c.Name, "aks")
				return !caf, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations",
		},
		"aks-007": {
			Id:          "aks-007",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "AKS should integrate authentication with AAD (Managed)",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				aad := c.Properties.AADProfile != nil && c.Properties.AADProfile.Managed != nil && *c.Properties.AADProfile.Managed
				return !aad, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/aks/managed-azure-ad",
		},
		"aks-008": {
			Id:          "aks-008",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "AKS should be RBAC enabled.",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				rbac := *c.Properties.EnableRBAC
				return !rbac, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/manage-azure-rbac",
		},
		"aks-009": {
			Id:          "aks-009",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityIdentity,
			Description: "AKS should have local accounts disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)

				if c.Properties.DisableLocalAccounts != nil && *c.Properties.DisableLocalAccounts {
					return false, ""
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/managed-aad#disable-local-accounts",
		},
		"aks-010": {
			Id:          "aks-010",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: "Best Practices",
			Description: "AKS should have httpApplicationRouting disabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["httpApplicationRouting"]
				broken := exists && *p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/http-application-routing",
		},
		"aks-011": {
			Id:          "aks-011",
			Category:    scanners.RulesCategoryReliability,
			Subcategory: scanners.RulesSubcategoryReliabilityMonitoring,
			Description: "AKS should have Container Insights enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				p, exists := c.Properties.AddonProfiles["omsagent"]
				broken := !exists || !*p.Enabled
				return broken, ""
			},
			Url: "https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview",
		},
		"aks-012": {
			Id:          "aks-012",
			Category:    scanners.RulesCategorySecurity,
			Subcategory: scanners.RulesSubcategorySecurityNetworking,
			Description: "AKS should have outbound type set to user defined routing",
			Severity:    scanners.SeverityHigh,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				out := *c.Properties.NetworkProfile.OutboundType == armcontainerservice.OutboundTypeUserDefinedRouting
				return !out, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/limit-egress-traffic",
		},
		"aks-013": {
			Id:          "aks-013",
			Category:    scanners.RulesCategoryPerformanceEfficienccy,
			Subcategory: scanners.RulesSubcategoryPerformanceEfficienccyNetworking,
			Description: "AKS should avoid using kubenet network plugin",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				out := *c.Properties.NetworkProfile.NetworkPlugin == armcontainerservice.NetworkPluginKubenet
				return out, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/operator-best-practices-network",
		},
		"aks-014": {
			Id:          "aks-014",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryReliabilityScaling,
			Description: "AKS should have autoscaler enabled",
			Severity:    scanners.SeverityMedium,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				if c.Properties.AgentPoolProfiles != nil {
					for _, p := range c.Properties.AgentPoolProfiles {
						if p.EnableAutoScaling != nil {
							return !*p.EnableAutoScaling, ""
						} else {
							return true, ""
						}
					}
				}
				return true, ""
			},
			Url: "https://learn.microsoft.com/azure/aks/concepts-scale",
		},
		"aks-015": {
			Id:          "aks-015",
			Category:    scanners.RulesCategoryOperationalExcellence,
			Subcategory: scanners.RulesSubcategoryOperationalExcellenceTags,
			Description: "AKS should have tags",
			Severity:    scanners.SeverityLow,
			Eval: func(target interface{}, scanContext *scanners.ScanContext) (bool, string) {
				c := target.(*armcontainerservice.ManagedCluster)
				return c.Tags == nil || len(c.Tags) == 0, ""
			},
			Url: "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json",
		},
	}
}
