package server

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type routeContribution struct {
	id        string
	registrar httpapi.RouteRegistrar
}

var requiredBuiltInRouteContributionIDs = []string{
	"auth",
	"incidents",
	"extensions",
	"jobs",
	"imports",
	"network_flow",
	"reporting",
	"report_composition",
	"reference_data",
	"incident_bundles",
	"saved_views",
	"view_schemas",
	"collaboration",
	"entities",
	"evidence",
	"assessments",
	"workbook",
	"timeline",
	"revisions",
}

func builtInRouteRegistrars(contributions []routeContribution) ([]httpapi.RouteRegistrar, error) {
	if len(contributions) != len(requiredBuiltInRouteContributionIDs) {
		return nil, fmt.Errorf("built-in route contribution count got %d want %d", len(contributions), len(requiredBuiltInRouteContributionIDs))
	}
	for index, requiredID := range requiredBuiltInRouteContributionIDs {
		if contributions[index].id != requiredID {
			return nil, fmt.Errorf("built-in route contribution %d got %q want %q", index, contributions[index].id, requiredID)
		}
	}
	return routeRegistrars(contributions)
}

func routeRegistrars(contributions []routeContribution) ([]httpapi.RouteRegistrar, error) {
	registrars := make([]httpapi.RouteRegistrar, 0, len(contributions))
	seen := make(map[string]struct{}, len(contributions))
	for _, contribution := range contributions {
		if contribution.id == "" {
			return nil, fmt.Errorf("route contribution id is required")
		}
		if contribution.registrar == nil {
			return nil, fmt.Errorf("route contribution %q has no registrar", contribution.id)
		}
		if _, exists := seen[contribution.id]; exists {
			return nil, fmt.Errorf("duplicate route contribution %q", contribution.id)
		}
		seen[contribution.id] = struct{}{}
		registrars = append(registrars, contribution.registrar)
	}
	return registrars, nil
}
