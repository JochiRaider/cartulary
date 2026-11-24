package performancefixture

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

// Provider retains record identities only within private application assembly.
// Its public receipt contains population counts, never object keys or identities.
type Provider struct {
	owner       *evidence.PerformanceFixtureContribution
	actorID     uuid.UUID
	descriptor  fixture.Descriptor
	segments    []evidence.PerformanceFixtureSegment
	titlePrefix string
	attempted   bool
	recordIDs   []uuid.UUID
}

// Segment is a run of consecutive records whose objects are ObjectBytes long.
type Segment struct {
	Count       int
	ObjectBytes int
}

// New creates the records in segment order, so the first segment's records
// are the first in construction order.
func New(owner *evidence.PerformanceFixtureContribution, actorID uuid.UUID, descriptor fixture.Descriptor, segments []Segment, titlePrefix string) (*Provider, error) {
	count := descriptor.ExpectedCounts["evidence_records"]
	created, bytes := 0, int64(0)
	payloads := make([]evidence.PerformanceFixtureSegment, 0, len(segments))
	for _, segment := range segments {
		payload, err := ImagePayload(segment.ObjectBytes)
		if err != nil || segment.Count < 1 {
			return nil, errors.New("evidence fixture contribution is incompatible")
		}
		created += segment.Count
		bytes += int64(segment.Count) * int64(segment.ObjectBytes)
		payloads = append(payloads, evidence.PerformanceFixtureSegment{Count: segment.Count, Payload: payload})
	}
	if owner == nil || actorID == uuid.Nil || descriptor.OwnerID != "module.evidence" ||
		len(descriptor.ExpectedCounts) != 3 || count < 1 || count > 10000 || created != count ||
		descriptor.ExpectedCounts["object_blobs"] != count ||
		int64(descriptor.ExpectedCounts["total_object_bytes"]) != bytes ||
		strings.TrimSpace(titlePrefix) == "" {
		return nil, errors.New("evidence fixture contribution is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{owner: owner, actorID: actorID, descriptor: descriptor, segments: payloads, titlePrefix: titlePrefix}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	descriptor := p.descriptor
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return descriptor
}

func (p *Provider) RecordIDs() []uuid.UUID { return slices.Clone(p.recordIDs) }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if p.attempted {
		return fixture.Receipt{}, errors.New("evidence fixture contribution cannot be retried")
	}
	p.attempted = true
	if state == nil {
		return fixture.Receipt{}, errors.New("evidence fixture requires an incident")
	}
	incidentID, err := uuid.Parse(state.IncidentID)
	if err != nil {
		return fixture.Receipt{}, err
	}
	result, err := p.owner.CreatePerformanceFixtureRecords(ctx, evidence.PerformanceFixtureCommand{
		ActorUserID: p.actorID, IncidentID: incidentID, Segments: p.segments,
		TitlePrefix: p.titlePrefix, ContentType: "image/png", Now: time.Now().UTC(),
	})
	if err != nil {
		return fixture.Receipt{}, err
	}
	counts := map[string]int{"evidence_records": len(result.RecordIDs), "object_blobs": result.ObjectCount, "total_object_bytes": int(result.TotalBytes)}
	if !maps.Equal(counts, p.descriptor.ExpectedCounts) {
		return fixture.Receipt{}, errors.New("evidence fixture receipt population diverged")
	}
	p.recordIDs = slices.Clone(result.RecordIDs)
	return fixture.Receipt{ContributionID: p.descriptor.ContributionID, Version: p.descriptor.Version, OwnerID: p.descriptor.OwnerID, Counts: counts}, nil
}
