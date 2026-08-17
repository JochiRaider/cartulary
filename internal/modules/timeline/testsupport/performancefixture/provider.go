package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

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
	descriptor  fixture.Descriptor
}

func New(application Application, descriptor fixture.Descriptor) (*Provider, error) {
	if application == nil {
		return nil, errors.New("timeline performance fixture application is required")
	}
	if descriptor.ExpectedCounts["timeline_rows"] < 1 {
		return nil, errors.New("timeline performance fixture descriptor is incompatible")
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
		return fixture.Receipt{}, errors.New("timeline performance fixture requires an Incidents workspace")
	}
	rowCount := p.descriptor.ExpectedCounts["timeline_rows"]
	rows := make([]Row, rowCount)
	for index := range rows {
		row := Row{Summary: fmt.Sprintf("Performance Timeline %05d", index)}
		if index%20 == 0 {
			suffix := fmt.Sprintf("%04d", index/20)
			row.Summary += " perf-host-" + suffix + " perf-identity-" + suffix + "@example.test https://fixture-" + suffix + ".example.test/trace"
			row.HostRef = "perf-host-" + suffix
			row.IdentityRef = "perf-identity-" + suffix + "@example.test"
			row.Tag = "perf-tag-" + suffix
			row.DataSource = "https://fixture-" + suffix + ".example.test/trace"
		}
		rows[index] = row
	}
	if err := p.application.CreateFixtureTimelineRows(ctx, state.IncidentID, rows); err != nil {
		return fixture.Receipt{}, fmt.Errorf("create fixture Timeline set: %w", err)
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}

func (p *Provider) PerformanceFixtureBatchDiagnostic() fixture.BatchDiagnostic {
	rowCount := p.descriptor.ExpectedCounts["timeline_rows"]
	return fixture.BatchDiagnostic{
		Strategy:            "owner_set_oriented",
		BatchCount:          1,
		ConfiguredBatchSize: rowCount,
		ItemCount:           rowCount,
	}
}
