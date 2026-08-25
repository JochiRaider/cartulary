package tasksdecisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func taskDecisionRevisionFacts(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
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

func (f *MutationFacade) appendTaskDecisionRecordChangedTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, actorUserID uuid.UUID, clientTxnID string, changeSetID uuid.UUID, recordID uuid.UUID, rowVersion int64, ordinal int, createdAt time.Time, viewSchemaID string, row map[string]any, fieldKeys []string) error {
	patch := collabprotocol.BuildViewRowPatch(row, fieldKeys)
	changeKind := "invalidate"
	if patch != nil {
		changeKind = "patch"
	}
	return f.publications.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID: incidentID, RecordID: recordID, ChangeSetID: changeSetID, ActorUserID: actorUserID,
		RowVersion: rowVersion, ClientTxnID: clientTxnID, MutationOrdinal: ordinal, CreatedAt: createdAt.UTC(), PublicFieldKeys: fieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{ViewSchemaID: viewSchemaID, RecordID: recordID, RowVersion: rowVersion, ChangeKind: changeKind, PatchCells: patch}},
	})
}
