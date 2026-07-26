package timeline

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *store) appendRecordChangeIntentTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	changeSetID uuid.UUID,
	clientTxnID string,
	actorUserID uuid.UUID,
	changedFieldKeys []string,
	row map[string]any,
	mutationOrdinal int,
	createdAt time.Time,
) error {
	if s == nil || s.collaboration == nil {
		return errors.New("timeline collaboration intent port is not configured")
	}
	return s.collaboration.AppendRecordChangeIntentTx(ctx, tx, RecordChangeIntentParams{
		IncidentID:       incidentID,
		RecordID:         recordID,
		RowVersion:       rowVersion,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: changedFieldKeys,
		ViewSchemaID:     TimelineViewSchemaID,
		Row:              row,
		MutationOrdinal:  mutationOrdinal,
		CreatedAt:        createdAt.UTC(),
	})
}
