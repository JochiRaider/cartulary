package revisions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func (s *commandStore) appendRollbackRecordChangeIntentsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changes []RollbackRecordChange,
	createdAt time.Time,
) error {
	if s == nil || s.collaboration == nil {
		return errors.New("revisions collaboration intent port is not configured")
	}
	for ordinal, change := range changes {
		intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
			IncidentID:       incidentID,
			RecordID:         change.RecordID,
			RowVersion:       change.RowVersion,
			ChangeSetID:      change.ChangeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: change.ChangedFieldKeys,
			ViewSchemaID:     change.ViewSchemaID,
			ChangeKind:       "invalidate",
		}, ordinal, createdAt)
		if err != nil {
			return err
		}
		if err := s.collaboration.AppendIntentTx(ctx, tx, intent); err != nil {
			return err
		}
	}
	return nil
}

func (s *commandStore) appendDeleteRestoreRecordChangeIntentTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	viewSchemaID string,
	changeKind string,
	createdAt time.Time,
) error {
	if s == nil || s.collaboration == nil {
		return errors.New("revisions collaboration intent port is not configured")
	}
	intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
		IncidentID:   incidentID,
		RecordID:     recordID,
		RowVersion:   rowVersion,
		ChangeSetID:  changeSetID,
		ClientTxnID:  clientTxnID,
		ActorUserID:  actorUserID,
		ViewSchemaID: viewSchemaID,
		ChangeKind:   changeKind,
	}, 0, createdAt)
	if err != nil {
		return err
	}
	return s.collaboration.AppendIntentTx(ctx, tx, intent)
}
