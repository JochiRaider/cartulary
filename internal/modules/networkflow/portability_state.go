package networkflow

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// PortabilityStateBinding is the Network Flow owner's declarative,
// read-only binding for Incident Portability. It is safe to construct when the
// profile is inactive because it executes no profile participant, worker,
// migration, or dependency probe.
type PortabilityStateBinding struct {
	database postgres.DB
}

func NewPortabilityStateBinding(database postgres.DB) (*PortabilityStateBinding, error) {
	if database == nil {
		return nil, errors.New("network flow portability state binding requires PostgreSQL")
	}
	return &PortabilityStateBinding{database: database}, nil
}

func (b *PortabilityStateBinding) RetainedAuthoritativeStatePresent(ctx context.Context, incidentID uuid.UUID, familyIDs []string) (bool, error) {
	if b == nil || b.database == nil || incidentID == uuid.Nil ||
		!slices.Equal(familyIDs, networkFlowExtensionFamilies) {
		return false, errors.New("network flow portability family scope invalid")
	}
	var present bool
	err := b.database.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM network_flow_tables
     WHERE incident_id = $1
    UNION ALL
    SELECT 1
      FROM network_flow_rows
     WHERE incident_id = $1
    UNION ALL
    SELECT 1
      FROM network_flow_rejected_row_diagnostics
     WHERE incident_id = $1
    UNION ALL
    SELECT 1
      FROM network_flow_indicator_bindings
     WHERE incident_id = $1
    UNION ALL
    SELECT 1
      FROM network_flow_graph_views
     WHERE incident_id = $1
)
`, incidentID).Scan(&present)
	return present, err
}
