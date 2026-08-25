package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func appendEvidenceRecordChangeIntentsTx(
	ctx context.Context,
	tx pgx.Tx,
	appender collaboration.RecordChangedAppender,
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
	primaryPatch := collabprotocol.BuildViewRowPatch(primaryRow, primary.ChangedFieldKeys)
	primaryKind := "invalidate"
	if primaryPatch != nil {
		primaryKind = "patch"
	}
	changes := make([]collaboration.RecordChangeIntentInput, 0, 1+len(affected))
	changes = append(changes, collaboration.RecordChangeIntentInput{
		IncidentID:      incidentID,
		RecordID:        primary.RecordID,
		RowVersion:      primary.RowVersion,
		ChangeSetID:     changeSetID,
		ClientTxnID:     clientTxnID,
		ActorUserID:     actorUserID,
		PublicFieldKeys: primary.ChangedFieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{
			ViewSchemaID: primary.ViewSchemaID, RecordID: primary.RecordID, RowVersion: primary.RowVersion,
			ChangeKind: primaryKind, PatchCells: primaryPatch,
		}},
	})
	for ordinal, change := range affected {
		affectedViews := make([]collaboration.AffectedViewChange, 0, len(change.AffectedViews))
		for _, view := range change.AffectedViews {
			var patchCells map[string]any
			if view.Patch != nil {
				patchCells = map[string]any{
					"record_id": view.Patch.RecordID.String(), "row_version": view.Patch.RowVersion,
					"cells": view.Patch.Cells,
				}
				if len(view.Patch.GroupValues) > 0 {
					patchCells["group_values"] = view.Patch.GroupValues
				}
			}
			affectedViews = append(affectedViews, collaboration.AffectedViewChange{
				ViewSchemaID: view.ViewSchemaID,
				RecordID:     change.RecordID,
				RowVersion:   change.RowVersion,
				ChangeKind:   string(view.ChangeKind),
				PatchCells:   patchCells,
			})
		}
		changes = append(changes, collaboration.RecordChangeIntentInput{
			IncidentID:      incidentID,
			RecordID:        change.RecordID,
			RowVersion:      change.RowVersion,
			ChangeSetID:     changeSetID,
			ClientTxnID:     clientTxnID,
			ActorUserID:     actorUserID,
			MutationOrdinal: ordinal + 1,
			CreatedAt:       createdAt,
			PublicFieldKeys: change.ChangedFieldKeys,
			AffectedViews:   affectedViews,
		})
	}
	changes[0].CreatedAt = createdAt
	for _, change := range changes {
		if err := appender.AppendRecordChangedTx(ctx, tx, change); err != nil {
			return err
		}
	}
	return nil
}
