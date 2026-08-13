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
	primary attachRecordChange,
	primaryRow map[string]any,
	affected []attachRecordChange,
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
		affectedViews := make([]collaboration.AffectedViewChange, 0, len(change.AffectedViews))
		for _, view := range change.AffectedViews {
			var patchCells map[string]any
			if view.Patch != nil {
				patchCells = view.Patch.Cells
			}
			affectedViews = append(affectedViews, collaboration.AffectedViewChange{
				ViewSchemaID: view.ViewSchemaID,
				ChangeKind:   string(view.ChangeKind),
				PatchCells:   patchCells,
			})
		}
		changes = append(changes, collaboration.RecordChange{
			IncidentID:       incidentID,
			RecordID:         change.RecordID,
			RowVersion:       change.RowVersion,
			ChangeSetID:      changeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: change.ChangedFieldKeys,
			ViewSchemaID:     change.ViewSchemaID,
			AffectedViews:    affectedViews,
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
