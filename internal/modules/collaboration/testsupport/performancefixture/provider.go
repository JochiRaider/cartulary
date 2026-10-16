package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

// Provider completes construction publications before the isolated database is
// sealed. It uses normal durable sequencing and retains every replay event, so
// restored servers start at the construction high-water mark rather than
// dispatching old fixture mutations as new live traffic.
type Provider struct {
	db         postgres.DB
	stream     *stream.PostgresStream
	descriptor fixture.Descriptor
}

func New(db postgres.DB, descriptor fixture.Descriptor) (*Provider, error) {
	if descriptor.OwnerID != "module.collaboration" || len(descriptor.ExpectedCounts) != 1 {
		return nil, errors.New("collaboration fixture descriptor is incompatible")
	}
	if count, ok := descriptor.ExpectedCounts["pending_intents"]; !ok || count != 0 {
		return nil, errors.New("collaboration fixture requires no pending publications")
	}
	durable, err := stream.NewPostgresStream(db)
	if err != nil {
		return nil, err
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{db: db, stream: durable, descriptor: descriptor}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if state == nil {
		return fixture.Receipt{}, errors.New("collaboration fixture requires an Incidents workspace")
	}
	incidentID, err := uuid.Parse(state.IncidentID)
	if err != nil || incidentID == uuid.Nil {
		return fixture.Receipt{}, errors.New("collaboration fixture requires a valid incident")
	}
	for {
		var pending, unrelated, quarantined int
		if err := p.db.QueryRow(ctx, `SELECT
 (SELECT COUNT(*) FROM collaboration_event_intents WHERE dispatch_state='pending' AND incident_id=$1),
 (SELECT COUNT(*) FROM collaboration_event_intents WHERE dispatch_state='pending' AND incident_id<>$1),
 (SELECT COUNT(*) FROM collaboration_incident_stream_cursors WHERE quarantined_at IS NOT NULL)`, incidentID).Scan(&pending, &unrelated, &quarantined); err != nil {
			return fixture.Receipt{}, fmt.Errorf("inspect fixture publications: %w", err)
		}
		if unrelated != 0 || quarantined != 0 {
			return fixture.Receipt{}, errors.New("fixture publications are unrelated or quarantined")
		}
		if pending == 0 {
			break
		}
		processed, err := p.stream.SequencePending(ctx, time.Now().UTC())
		if err != nil {
			return fixture.Receipt{}, fmt.Errorf("sequence fixture publications: %w", err)
		}
		// A batch that cannot advance fails construction. Do not wait for a retry,
		// bypass backoff, clear quarantine or rewrite an invalid owner payload.
		if processed == 0 {
			return fixture.Receipt{}, errors.New("fixture publication sequencing made no progress")
		}
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
