package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func appendEvidenceRecordChangeIntentsTx(
	ctx context.Context,
	tx pgx.Tx,
	appender collaboration.IntentAppender,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changeSetID uuid.UUID,
	primary AttachRecordChange,
	primaryRow map[string]any,
	affected []AttachRecordChange,
	createdAt time.Time,
) error {
	if appender == nil {
		return errors.New("evidence collaboration intent port is not configured")
	}
	changes := make([]collaboration.RecordChange, 0, 1+len(affected))
	changes = append(changes, collaboration.RecordChange{
		IncidentID:       incidentID,
		RecordID:         primary.RecordID,
		RowVersion:       primary.RowVersion,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: primary.ChangedFieldKeys,
		ViewSchemaID:     primary.ViewSchemaID,
		Row:              primaryRow,
	})
	for _, change := range affected {
		changes = append(changes, collaboration.RecordChange{
			IncidentID:       incidentID,
			RecordID:         change.RecordID,
			RowVersion:       change.RowVersion,
			ChangeSetID:      changeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: change.ChangedFieldKeys,
			ViewSchemaID:     change.ViewSchemaID,
		})
	}
	for ordinal, change := range changes {
		intent, err := collaboration.NewRecordChangeIntent(change, ordinal, createdAt)
		if err != nil {
			return err
		}
		if err := appender.AppendIntentTx(ctx, tx, intent); err != nil {
			return err
		}
	}
	return nil
}
