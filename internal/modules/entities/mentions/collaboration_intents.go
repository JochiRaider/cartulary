package mentions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
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

	timelineChangedFieldKeys := mentionRowChangedFieldKeys(timelineResult.BeforeRow, timelineResult.AfterRow)
	changes := make([]collaboration.RecordChangeIntentInput, 0, 1+len(entityInvalidations))
	timelinePatch := collabprotocol.BuildViewRowPatch(timelineResult.AfterRow, timelineChangedFieldKeys)
	timelineChangeKind := "invalidate"
	if timelinePatch != nil {
		timelineChangeKind = "patch"
	}
	changes = append(changes, collaboration.RecordChangeIntentInput{
		IncidentID:      incidentID,
		RecordID:        timelineResult.SourceRecordID,
		RowVersion:      timelineResult.RowVersion,
		ChangeSetID:     changeSetID,
		ClientTxnID:     clientTxnID,
		ActorUserID:     actorUserID,
		PublicFieldKeys: timelineChangedFieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{
			ViewSchemaID: timelineViewSchemaID, RecordID: timelineResult.SourceRecordID,
			RowVersion: timelineResult.RowVersion, ChangeKind: timelineChangeKind, PatchCells: timelinePatch,
		}},
	})
	for ordinal, invalidation := range entityInvalidations {
		changes = append(changes, collaboration.RecordChangeIntentInput{
			IncidentID:      incidentID,
			RecordID:        invalidation.RecordID,
			RowVersion:      invalidation.RowVersion,
			ChangeSetID:     changeSetID,
			ClientTxnID:     clientTxnID,
			ActorUserID:     actorUserID,
			MutationOrdinal: ordinal + 1,
			CreatedAt:       createdAt,
			PublicFieldKeys: invalidation.ChangedFieldKeys,
			AffectedViews: []collaboration.AffectedViewChange{{
				ViewSchemaID: invalidation.ViewSchemaID, RecordID: invalidation.RecordID,
				RowVersion: invalidation.RowVersion, ChangeKind: "invalidate",
			}},
		})
	}
	changes[0].CreatedAt = createdAt
	for _, change := range changes {
		if err := s.ports.collaboration.AppendRecordChangedTx(ctx, tx, change); err != nil {
			return err
		}
	}
	return nil
}

func mentionRowChangedFieldKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	before := mentionCanonicalCells(beforeRow)
	after := mentionCanonicalCells(afterRow)
	candidates := make(map[string]struct{}, len(before)+len(after))
	for key := range before {
		candidates[key] = struct{}{}
	}
	for key := range after {
		candidates[key] = struct{}{}
	}
	changed := make([]string, 0, len(candidates))
	for key := range candidates {
		left, leftOK := before[key]
		right, rightOK := after[key]
		if leftOK != rightOK || !bytes.Equal(left, right) {
			changed = append(changed, key)
		}
	}
	slices.Sort(changed)
	return changed
}

func mentionCanonicalCells(row map[string]any) map[string]json.RawMessage {
	if row == nil || row["cells"] == nil {
		return map[string]json.RawMessage{}
	}
	payload, _ := json.Marshal(row["cells"])
	var cells map[string]json.RawMessage
	_ = json.Unmarshal(payload, &cells)
	if cells == nil {
		cells = map[string]json.RawMessage{}
	}
	return cells
}

func mentionRevisionFacts(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	before, _ := beforeRow["cells"].(map[string]any)
	after, _ := afterRow["cells"].(map[string]any)
	facts := make([]revisions.RevisionConflictFact, 0, len(fieldKeys))
	for _, key := range fieldKeys {
		beforeValue, beforePresent := before[key]
		afterValue, afterPresent := after[key]
		facts = append(facts, revisions.RevisionConflictFact{FieldKey: key, BeforePresent: beforePresent, BeforeValue: beforeValue, AfterPresent: afterPresent, AfterValue: afterValue})
	}
	return facts
}
