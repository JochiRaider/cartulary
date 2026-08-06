package server

import (
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
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
	"saved_views",
	"view_schemas",
	"collaboration",
	"entities",
	"evidence",
	"workbook",
	"timeline",
	"revisions",
	"indicators",
}

type extensionRouteBinding struct {
	id              string
	contributionIDs []string
	registrar       httpapi.RouteRegistrar
	baseRegistrarID string
}

type applicationRouteCatalog struct {
	publication extensionassembly.PublicationCatalog
}

func newApplicationRouteCatalog(publication extensionassembly.PublicationCatalog) applicationRouteCatalog {
	return applicationRouteCatalog{publication: publication}
}

func (catalog applicationRouteCatalog) Bind(
	contributions []routeContribution,
	extensionBindings []extensionRouteBinding,
) ([]httpapi.RouteRegistrar, error) {
	if len(contributions) != len(requiredBuiltInRouteContributionIDs) {
		return nil, fmt.Errorf("built-in route contribution count got %d want %d", len(contributions), len(requiredBuiltInRouteContributionIDs))
	}
	baseIDs := make(map[string]struct{}, len(contributions))
	for index, requiredID := range requiredBuiltInRouteContributionIDs {
		if contributions[index].id != requiredID {
			return nil, fmt.Errorf("built-in route contribution %d got %q want %q", index, contributions[index].id, requiredID)
		}
		baseIDs[requiredID] = struct{}{}
	}
	registrars, err := routeRegistrars(contributions)
	if err != nil {
		return nil, err
	}
	claimedRoutes := catalog.publication.ContributionIDs("http_route_family")
	claimed := make(map[string]struct{}, len(claimedRoutes))
	for _, contributionID := range claimedRoutes {
		claimed[contributionID] = struct{}{}
	}
	consumed := make(map[string]string, len(claimedRoutes))
	for _, binding := range extensionBindings {
		if binding.id == "" || len(binding.contributionIDs) == 0 {
			return nil, fmt.Errorf("extension route binding identity is incomplete")
		}
		admitted := true
		for _, contributionID := range binding.contributionIDs {
			if _, present := claimed[contributionID]; !present {
				admitted = false
			}
		}
		if !admitted {
			for _, contributionID := range binding.contributionIDs {
				if _, present := claimed[contributionID]; present {
					return nil, fmt.Errorf("extension route binding %q is partially claimed", binding.id)
				}
			}
			continue
		}
		if binding.registrar == nil {
			if _, present := baseIDs[binding.baseRegistrarID]; !present {
				return nil, fmt.Errorf("extension route binding %q has no registrar", binding.id)
			}
		} else {
			registrars = append(registrars, binding.registrar)
		}
		for _, contributionID := range binding.contributionIDs {
			if prior := consumed[contributionID]; prior != "" {
				return nil, fmt.Errorf("extension route contribution %q is bound by %q and %q", contributionID, prior, binding.id)
			}
			consumed[contributionID] = binding.id
		}
	}
	if len(consumed) != len(claimed) {
		missing := make([]string, 0)
		for contributionID := range claimed {
			if consumed[contributionID] == "" {
				missing = append(missing, contributionID)
			}
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("claimed extension route contributions are unbound: %v", missing)
	}
	return registrars, nil
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
