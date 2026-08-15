package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type Application interface {
	ValidateFixtureProjectionSets(context.Context, string, []string) error
}

type Provider struct {
	application   Application
	descriptor    fixture.Descriptor
	viewSchemaIDs []string
}

func New(application Application, descriptor fixture.Descriptor, viewSchemaIDs []string) (*Provider, error) {
	if application == nil {
		return nil, errors.New("projections performance fixture application is required")
	}
	if descriptor.ExpectedCounts["projection_sets"] < 1 || len(viewSchemaIDs) != descriptor.ExpectedCounts["projection_sets"] {
		return nil, errors.New("projections performance fixture descriptor is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, descriptor: descriptor, viewSchemaIDs: slices.Clone(viewSchemaIDs)}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("projections performance fixture requires an Incidents workspace")
	}
	if err := p.application.ValidateFixtureProjectionSets(ctx, state.IncidentID, slices.Clone(p.viewSchemaIDs)); err != nil {
		return fixture.Receipt{}, fmt.Errorf("validate fixture projection sets: %w", err)
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
