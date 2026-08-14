package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "links.timeline_associations.v1"

type Expectations struct {
	Links    int
	Mentions int
	Tags     int
	Stride   int
}

type Application interface {
	ValidateFixtureAssociations(context.Context, string, Expectations) error
}

type Provider struct {
	application Application
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("links performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.links",
		Dependencies:   []string{"timeline.large_grid.v1"},
		ExpectedCounts: map[string]int{"links": 1000, "mentions": 1000, "tags": 1000},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("links performance fixture requires an Incidents workspace")
	}
	expected := Expectations{Links: 1000, Mentions: 1000, Tags: 1000, Stride: 20}
	if err := p.application.ValidateFixtureAssociations(ctx, state.IncidentID, expected); err != nil {
		return fixture.Receipt{}, fmt.Errorf("validate fixture associations: %w", err)
	}
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.links",
		Counts:         map[string]int{"links": 1000, "mentions": 1000, "tags": 1000},
	}, nil
}
