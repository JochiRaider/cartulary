package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type Host struct {
	DisplayName string
	Hostname    string
}

type Identity struct {
	DisplayName string
	UPN         string
}

type Application interface {
	CreateFixtureHosts(context.Context, string, []Host) error
	CreateFixtureIdentities(context.Context, string, []Identity) error
}

type Provider struct {
	application Application
	descriptor  fixture.Descriptor
}

func New(application Application, descriptor fixture.Descriptor) (*Provider, error) {
	if application == nil {
		return nil, errors.New("entities performance fixture application is required")
	}
	if descriptor.ExpectedCounts["hosts"] < 1 || descriptor.ExpectedCounts["identities"] < 1 {
		return nil, errors.New("entities performance fixture descriptor is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, descriptor: descriptor}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("entities performance fixture requires an Incidents workspace")
	}
	const batchSize = 500
	hostCount := p.descriptor.ExpectedCounts["hosts"]
	for start := 0; start < hostCount; start += batchSize {
		end := min(start+batchSize, hostCount)
		hosts := make([]Host, end-start)
		for offset := range hosts {
			index := start + offset
			suffix := fmt.Sprintf("%04d", index)
			hosts[offset] = Host{DisplayName: "Performance Host " + suffix, Hostname: "perf-host-" + suffix}
		}
		if err := p.application.CreateFixtureHosts(ctx, state.IncidentID, hosts); err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture hosts batch %d: %w", start/batchSize, err)
		}
	}
	identityCount := p.descriptor.ExpectedCounts["identities"]
	for start := 0; start < identityCount; start += batchSize {
		end := min(start+batchSize, identityCount)
		identities := make([]Identity, end-start)
		for offset := range identities {
			index := start + offset
			suffix := fmt.Sprintf("%04d", index)
			identities[offset] = Identity{DisplayName: "Performance Identity " + suffix, UPN: "perf-identity-" + suffix + "@example.test"}
		}
		if err := p.application.CreateFixtureIdentities(ctx, state.IncidentID, identities); err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture identities batch %d: %w", start/batchSize, err)
		}
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
