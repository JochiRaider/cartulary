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
	if s.lockEndpoints == nil {
		return fmt.Errorf("record endpoint locker is required")
	}
	endpoints, err := s.lockEndpoints(ctx, tx, incidentID, ids)
	if err != nil {
		return err
	}
	if len(endpoints) != len(ids) {
		return fmt.Errorf("link projections: missing incident endpoint")
	}
	for _, endpoint := range endpoints {
		for _, provider := range registry.providers {
			if !slices.Contains(provider.descriptor.SourceRecordTypes, endpoint.RecordType) {
				continue
			}
			if provider.descriptor.Capabilities.RefreshRow && provider.refreshRowTx != nil {
				if err := provider.refreshRowTx(ctx, s, tx, endpoint.RecordID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
