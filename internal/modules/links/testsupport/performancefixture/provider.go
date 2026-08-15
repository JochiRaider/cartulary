package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

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
	descriptor  fixture.Descriptor
	stride      int
}

func New(application Application, descriptor fixture.Descriptor, stride int) (*Provider, error) {
	if application == nil {
		return nil, errors.New("links performance fixture application is required")
	}
	if descriptor.ExpectedCounts["links"] < 1 || descriptor.ExpectedCounts["mentions"] < 1 || descriptor.ExpectedCounts["tags"] < 1 || stride < 1 {
		return nil, errors.New("links performance fixture descriptor is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, descriptor: descriptor, stride: stride}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("links performance fixture requires an Incidents workspace")
	}
	expected := Expectations{
		Links: p.descriptor.ExpectedCounts["links"], Mentions: p.descriptor.ExpectedCounts["mentions"],
		Tags: p.descriptor.ExpectedCounts["tags"], Stride: p.stride,
	}
	if err := p.application.ValidateFixtureAssociations(ctx, state.IncidentID, expected); err != nil {
		return fixture.Receipt{}, fmt.Errorf("validate fixture associations: %w", err)
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
