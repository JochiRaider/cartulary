package runtime

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"slices"
)

// RefreshLinkedRecordsTx refreshes source-owner contributions for both sides
// of a relationship, including removed endpoints. Locking envelopes before
// reading source facts serializes simultaneous updates to shared counts.
func (s *Store) RefreshLinkedRecordsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, ids []uuid.UUID) error {
	registry, err := s.providerRegistry()
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT record_id, record_type FROM records WHERE incident_id=$1 AND record_id=ANY($2::uuid[]) ORDER BY record_id FOR NO KEY UPDATE`, incidentID, ids)
	if err != nil {
		return err
	}
	type endpoint struct {
		id   uuid.UUID
		kind string
	}
	var endpoints []endpoint
	for rows.Next() {
		var value endpoint
		if err := rows.Scan(&value.id, &value.kind); err != nil {
			rows.Close()
			return err
		}
		endpoints = append(endpoints, value)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(endpoints) != len(ids) {
		return fmt.Errorf("link projections: missing incident endpoint")
	}
	for _, endpoint := range endpoints {
		for _, provider := range registry.providers {
			if !slices.Contains(provider.descriptor.SourceRecordTypes, endpoint.kind) {
				continue
			}
			if provider.descriptor.Capabilities.RefreshRow && provider.refreshRowTx != nil {
				if err := provider.refreshRowTx(ctx, s, tx, endpoint.id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
