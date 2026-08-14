package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "projections.timeline_entities.v1"

type Application interface {
	ValidateFixtureProjectionSets(context.Context, string, []string) error
}

type Provider struct {
	application Application
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("projections performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.projections",
		Dependencies:   []string{"links.timeline_associations.v1"},
		ExpectedCounts: map[string]int{"projection_sets": 3},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("projections performance fixture requires an Incidents workspace")
	}
	viewSchemaIDs := []string{
		"cartulary.view.hosts.v1",
		"cartulary.view.identities.v1",
		"cartulary.view.timeline.v2",
	}
	if err := p.application.ValidateFixtureProjectionSets(ctx, state.IncidentID, viewSchemaIDs); err != nil {
		return fixture.Receipt{}, fmt.Errorf("validate fixture projection sets: %w", err)
	}
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.projections",
		Counts:         map[string]int{"projection_sets": 3},
	}, nil
}
