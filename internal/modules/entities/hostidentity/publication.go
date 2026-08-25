package hostidentity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func entityRevisionFacts(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	facts := make([]revisions.RevisionConflictFact, 0, len(fieldKeys))
	for _, fieldKey := range fieldKeys {
		beforeValue, beforePresent := beforeCells[fieldKey]
		afterValue, afterPresent := afterCells[fieldKey]
		facts = append(facts, revisions.RevisionConflictFact{
			FieldKey: fieldKey, BeforePresent: beforePresent, BeforeValue: beforeValue,
			AfterPresent: afterPresent, AfterValue: afterValue,
		})
	}
	return facts
}

// HostRevisionConflictFacts derives Revisions-private facts from the Entities
// owner's canonical host projection for the merge transaction.
func (*MergeCapability) HostRevisionConflictFacts(before HostRecord, after HostRecord, fieldKeys []string) []revisions.RevisionConflictFact {
	return entityRevisionFacts(buildHostRow(before), buildHostRow(after), fieldKeys)
}

// IdentityRevisionConflictFacts is the identity counterpart of
// HostRevisionConflictFacts.
func (*MergeCapability) IdentityRevisionConflictFacts(before IdentityRecord, after IdentityRecord, fieldKeys []string) []revisions.RevisionConflictFact {
	return entityRevisionFacts(buildIdentityRow(before), buildIdentityRow(after), fieldKeys)
}

func (core *mutationCore) appendRecordChangedTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	clientTxnID string,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	mutationOrdinal int,
	createdAt time.Time,
	viewSchemaID string,
	afterRow map[string]any,
	fieldKeys []string,
) error {
	patch := collabprotocol.BuildViewRowPatch(afterRow, fieldKeys)
	changeKind := "invalidate"
	if patch != nil {
		changeKind = "patch"
	}
	return core.publications.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID:      incidentID,
		RecordID:        recordID,
		ChangeSetID:     changeSetID,
		ActorUserID:     actorUserID,
		RowVersion:      rowVersion,
		ClientTxnID:     clientTxnID,
		MutationOrdinal: mutationOrdinal,
		CreatedAt:       createdAt.UTC(),
		PublicFieldKeys: fieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{
			ViewSchemaID: viewSchemaID,
			RecordID:     recordID,
			RowVersion:   rowVersion,
			ChangeKind:   changeKind,
			PatchCells:   patch,
		}},
	})
}
