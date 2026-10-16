package performancefixture

import (
	"context"
	"errors"
	"maps"
	"slices"

	"github.com/google/uuid"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type EvidenceRecords interface{ RecordIDs() []uuid.UUID }

// InspectorProvider consumes a source-owner identity capability. The shared
// fixture state and retained receipts do not acquire Evidence-specific fields.
type InspectorProvider struct {
	application *EvidenceInspectorApplication
	evidence    EvidenceRecords
	descriptor  fixture.Descriptor
	attempted   bool
	target      EvidenceInspectorTarget
}

func NewInspectorProvider(application *EvidenceInspectorApplication, evidence EvidenceRecords, descriptor fixture.Descriptor) (*InspectorProvider, error) {
	if application == nil || evidence == nil || descriptor.OwnerID != "module.timeline" || len(descriptor.ExpectedCounts) != 2 ||
		descriptor.ExpectedCounts["timeline_rows"] < 2 || descriptor.ExpectedCounts["timeline_rows"] > 100000 ||
		descriptor.ExpectedCounts["attached_evidence"] < 1 || descriptor.ExpectedCounts["attached_evidence"] > 100 {
		return nil, errors.New("inspector Timeline fixture contribution is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &InspectorProvider{application: application, evidence: evidence, descriptor: descriptor}, nil
}

func (p *InspectorProvider) Descriptor() fixture.Descriptor {
	descriptor := p.descriptor
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return descriptor
}

func (p *InspectorProvider) Target() EvidenceInspectorTarget { return p.target }

func (p *InspectorProvider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if p.attempted {
		return fixture.Receipt{}, errors.New("inspector Timeline fixture contribution cannot be retried")
	}
	p.attempted = true
	if state == nil {
		return fixture.Receipt{}, errors.New("inspector Timeline fixture requires an incident")
	}
	incidentID, err := uuid.Parse(state.IncidentID)
	if err != nil {
		return fixture.Receipt{}, err
	}
	recordIDs := p.evidence.RecordIDs()
	count := p.descriptor.ExpectedCounts["attached_evidence"]
	if len(recordIDs) < count {
		return fixture.Receipt{}, errors.New("inspector Timeline fixture lacks its Evidence population")
	}
	target, err := p.application.Create(ctx, incidentID, p.descriptor.ExpectedCounts["timeline_rows"], recordIDs[:count])
	if err != nil {
		return fixture.Receipt{}, err
	}
	p.target = target
	return fixture.Receipt{ContributionID: p.descriptor.ContributionID, Version: p.descriptor.Version, OwnerID: p.descriptor.OwnerID, Counts: maps.Clone(p.descriptor.ExpectedCounts)}, nil
}
