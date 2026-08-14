package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "timeline.large_grid.v1"

type Row struct {
	Summary     string
	HostRef     string
	IdentityRef string
	Tag         string
	DataSource  string
}

type Application interface {
	CreateFixtureTimelineRows(context.Context, string, []Row) error
}

type Provider struct {
	application Application
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("timeline performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.timeline",
		Dependencies:   []string{"entities.hosts_identities.v1"},
		ExpectedCounts: map[string]int{"timeline_rows": 20000},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state.IncidentID == "" {
		return fixture.Receipt{}, errors.New("timeline performance fixture requires an Incidents workspace")
	}
	const (
		rowCount  = 20000
		batchSize = 500
	)
	for start := 0; start < rowCount; start += batchSize {
		rows := make([]Row, batchSize)
		for offset := range batchSize {
			index := start + offset
			row := Row{Summary: fmt.Sprintf("Performance Timeline %05d", index)}
			if index%20 == 0 {
				suffix := fmt.Sprintf("%04d", index/20)
				row.Summary += " perf-host-" + suffix + " perf-identity-" + suffix + "@example.test https://fixture-" + suffix + ".example.test/trace"
				row.HostRef = "perf-host-" + suffix
				row.IdentityRef = "perf-identity-" + suffix + "@example.test"
				row.Tag = "perf-tag-" + suffix
				row.DataSource = "https://fixture-" + suffix + ".example.test/trace"
			}
			rows[offset] = row
		}
		if err := p.application.CreateFixtureTimelineRows(ctx, state.IncidentID, rows); err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture Timeline batch %d: %w", start/batchSize, err)
		}
	}
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.timeline",
		Counts:         map[string]int{"timeline_rows": 20000},
	}, nil
}
