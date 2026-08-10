package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type ProjectionRebuilder interface {
	RebuildIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
}

// LiveRecordReader supplies disposable projection material for Collaboration
// consequences. Values returned here never become retained revision history.
type LiveRecordReader interface {
	Supports(viewSchemaID string) bool
	LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error)
}

func (s *commandStore) rebuildProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.projections.RebuildIncidentTx(ctx, tx, incidentID)
}

func (s *commandStore) loadLiveRecordTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, source deleterestorecontract.DeleteRestoreSource) (map[string]any, error) {
	if s.liveRecords.Supports(viewSchemaID) {
		row, err := s.liveRecords.LoadRowTx(ctx, tx, viewSchemaID, recordID)
		if err == nil {
			return row, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	// Some public event surfaces intentionally have no queryable projection.
	// Their source owner supplies live event material explicitly here. This
	// value is never passed to retained history persistence.
	return source.SnapshotTx(ctx, tx, recordID)
}
