package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
)

type ProjectionRebuilder interface {
	RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

type ProjectionServices interface {
	ProjectionRebuilder
	Supports(viewSchemaID string) bool
	LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error)
}

func (s *commandStore) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.projections.RebuildIncidentTx(ctx, tx, incidentID)
}

func (s *commandStore) snapshotRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, provider recorddeleterestore.SourceProvider) (map[string]any, error) {
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}
	if s.projections.Supports(viewSchemaID) {
		row, err := s.projections.LoadRowTx(ctx, tx, viewSchemaID, recordID)
		if err == nil {
			return row, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	// Deleted records do not appear in published projections, and record families
	// without a public query surface have no canonical projection reader.
	return provider.SnapshotTx(ctx, tx, recordID)
}
