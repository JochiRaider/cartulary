package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type SetExpectation struct {
	ViewSchemaID string
	ExactRows    int
}

type Application interface {
	ValidateFixtureProjectionSets(context.Context, string, []SetExpectation) error
}

type Provider struct {
	application  Application
	descriptor   fixture.Descriptor
	expectations []SetExpectation
}

func New(application Application, descriptor fixture.Descriptor, expectations []SetExpectation) (*Provider, error) {
	if application == nil {
		return nil, errors.New("projections performance fixture application is required")
	}
	if descriptor.ExpectedCounts["projection_sets"] < 1 || len(expectations) != descriptor.ExpectedCounts["projection_sets"] {
		return nil, errors.New("projections performance fixture descriptor is incompatible")
	}
	for _, expectation := range expectations {
		if expectation.ViewSchemaID == "" || expectation.ExactRows < 1 {
			return nil, errors.New("projections performance fixture expectation is incomplete")
		}
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, descriptor: descriptor, expectations: slices.Clone(expectations)}, nil
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
	if err := p.application.ValidateFixtureProjectionSets(ctx, state.IncidentID, slices.Clone(p.expectations)); err != nil {
		return fixture.Receipt{}, fmt.Errorf("validate fixture projection sets: %w", err)
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
