package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "entities.hosts_identities.v1"

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
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("entities performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.entities",
		Dependencies:   []string{"incidents.workspace.v1"},
		ExpectedCounts: map[string]int{"hosts": 1000, "identities": 1000},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("entities performance fixture requires an Incidents workspace")
	}
	const batchSize = 500
	for start := 0; start < 1000; start += batchSize {
		hosts := make([]Host, batchSize)
		identities := make([]Identity, batchSize)
		for offset := range batchSize {
			index := start + offset
			suffix := fmt.Sprintf("%04d", index)
			hosts[offset] = Host{DisplayName: "Performance Host " + suffix, Hostname: "perf-host-" + suffix}
			identities[offset] = Identity{DisplayName: "Performance Identity " + suffix, UPN: "perf-identity-" + suffix + "@example.test"}
		}
		if err := p.application.CreateFixtureHosts(ctx, state.IncidentID, hosts); err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture hosts batch %d: %w", start/batchSize, err)
		}
		if err := p.application.CreateFixtureIdentities(ctx, state.IncidentID, identities); err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture identities batch %d: %w", start/batchSize, err)
		}
	}
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.entities",
		Counts:         map[string]int{"hosts": 1000, "identities": 1000},
	}, nil
}
