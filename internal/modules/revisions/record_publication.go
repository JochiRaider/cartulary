package revisions

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RecordPublicationEffect is Revisions' source-owned description of the
// public consequence of a destructive history operation. Application
// composition translates it to Collaboration's independently owned input.
type RecordPublicationEffect struct {
	IncidentID      uuid.UUID
	RecordID        uuid.UUID
	ChangeSetID     uuid.UUID
	ActorUserID     uuid.UUID
	RowVersion      int64
	ClientTxnID     string
	MutationOrdinal int
	CreatedAt       time.Time
	PublicFieldKeys []string
	ViewSchemaID    string
	ChangeKind      string
}

type RecordPublicationPort interface {
	AppendRecordChangedTx(context.Context, pgx.Tx, RecordPublicationEffect) error
}

func revisionFactsForFields(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) ([]RevisionConflictFact, error) {
	beforeCells, err := conflictCells(beforeRow)
	if err != nil {
		return nil, err
	}
	afterCells, err := conflictCells(afterRow)
	if err != nil {
		return nil, err
	}
	keys := append([]string(nil), fieldKeys...)
	slices.Sort(keys)
	keys = slices.Compact(keys)
	facts := make([]RevisionConflictFact, 0, len(keys))
	for _, key := range keys {
		beforeValue, beforePresent := beforeCells[key]
		afterValue, afterPresent := afterCells[key]
		facts = append(facts, RevisionConflictFact{
			FieldKey: key, BeforePresent: beforePresent, BeforeValue: beforeValue,
			AfterPresent: afterPresent, AfterValue: afterValue,
		})
	}
	return facts, nil
}

func conflictCells(row map[string]any) (map[string]any, error) {
	if row == nil || row["cells"] == nil {
		return map[string]any{}, nil
	}
	if cells, ok := row["cells"].(map[string]any); ok {
		return cells, nil
	}
	encoded, err := json.Marshal(row["cells"])
	if err != nil {
		return nil, err
	}
	var cells map[string]any
	if err := json.Unmarshal(encoded, &cells); err != nil {
		return nil, err
	}
	if cells == nil {
		cells = map[string]any{}
	}
	return cells, nil
}

func (s *commandStore) recordPublicationEffectTx(
	ctx context.Context,
	tx pgx.Tx,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	rowVersion int64,
	viewSchemaID string,
	changeKind string,
	publicFieldKeys []string,
) (RecordPublicationEffect, error) {
	envelope, err := s.envelopes.LoadEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RecordPublicationEffect{}, fmt.Errorf("load record publication envelope: %w", err)
	}
	var actorUserID uuid.UUID
	var clientTxnID *string
	var createdAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT actor_user_id, client_txn_id, created_at
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&actorUserID, &clientTxnID, &createdAt); err != nil {
		return RecordPublicationEffect{}, fmt.Errorf("load record publication identity: %w", err)
	}
	var mutationOrdinal int
	if err := tx.QueryRow(ctx, `
SELECT GREATEST(COALESCE(min(sequence_no), 1) - 1, 0)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_id = $2
`, changeSetID, recordID.String()).Scan(&mutationOrdinal); err != nil {
		return RecordPublicationEffect{}, fmt.Errorf("load record publication ordinal: %w", err)
	}
	clientTxn := ""
	if clientTxnID != nil {
		clientTxn = *clientTxnID
	}
	return RecordPublicationEffect{
		IncidentID: envelope.IncidentID, RecordID: recordID, ChangeSetID: changeSetID,
		ActorUserID: actorUserID, RowVersion: rowVersion, ClientTxnID: clientTxn,
		MutationOrdinal: mutationOrdinal, CreatedAt: createdAt.UTC(), PublicFieldKeys: slices.Clone(publicFieldKeys),
		ViewSchemaID: viewSchemaID, ChangeKind: changeKind,
	}, nil
}
