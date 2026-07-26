package mentions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
)

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func (s *Store) appendMentionActionIntentsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changeSetID uuid.UUID,
	timelineResult mentioneffects.ActionResult,
	entityInvalidations []MentionEntityInvalidation,
	createdAt time.Time,
) error {
	if s == nil || s.ports.collaboration == nil {
		return errors.New("entity mention collaboration intent port is not configured")
	}

	timelineChangedFieldKeys, err := collaboration.ChangedCellKeys(
		timelineResult.BeforeRow,
		timelineResult.AfterRow,
	)
	if err != nil {
		return err
	}
	changes := make([]collaboration.RecordChange, 0, 1+len(entityInvalidations))
	changes = append(changes, collaboration.RecordChange{
		IncidentID:       incidentID,
		RecordID:         timelineResult.SourceRecordID,
		RowVersion:       timelineResult.RowVersion,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: timelineChangedFieldKeys,
		ViewSchemaID:     timelineViewSchemaID,
		Row:              timelineResult.AfterRow,
	})
	for _, invalidation := range entityInvalidations {
		changes = append(changes, collaboration.RecordChange{
			IncidentID:       incidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      changeSetID,
			ClientTxnID:      clientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
			ViewSchemaID:     invalidation.ViewSchemaID,
		})
	}
	for ordinal, change := range changes {
		intent, err := collaboration.NewRecordChangeIntent(change, ordinal, createdAt)
		if err != nil {
			return err
		}
		if err := s.ports.collaboration.AppendIntentTx(ctx, tx, intent); err != nil {
			return err
		}
	}
	return nil
}
