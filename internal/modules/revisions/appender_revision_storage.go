package revisions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func (*Appender) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params AppendRecordRevisionParams) error {
	beforeValue, afterValue, err := recordSnapshotPair(params.RecordID, params.BeforeSnapshot, params.AfterSnapshot)
	if err != nil {
		return err
	}
	revisionID, err := appendRecordRevisionValuesTx(ctx, tx, appendRecordRevisionValuesParams{
		ChangeSetID: params.ChangeSetID,
		RecordID:    params.RecordID,
		RowVersion:  params.RowVersion,
		BeforeValue: beforeValue,
		AfterValue:  afterValue,
	})
	if err != nil {
		return err
	}
	facts, err := recordRevisionConflictFacts(params.LiveChange)
	if err != nil {
		return fmt.Errorf("derive record revision conflict facts: %w", err)
	}
	for _, fact := range facts {
		beforeValue, err := revisionConflictFactValue(fact.BeforeValue, fact.BeforePresent)
		if err != nil {
			return fmt.Errorf("encode record revision conflict fact before value: %w", err)
		}
		afterValue, err := revisionConflictFactValue(fact.AfterValue, fact.AfterPresent)
		if err != nil {
			return fmt.Errorf("encode record revision conflict fact after value: %w", err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO record_revision_conflict_facts (
    revision_id,
    field_key,
    before_present,
    before_value,
    after_present,
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6)
`, revisionID, fact.FieldKey, fact.BeforePresent, beforeValue, fact.AfterPresent, afterValue); err != nil {
			return fmt.Errorf("append record revision conflict fact: %w", err)
		}
	}
	return nil
}

type appendRecordRevisionValuesParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue map[string]any
	AfterValue  map[string]any
}

func appendRecordRevisionValuesTx(ctx context.Context, tx pgx.Tx, params appendRecordRevisionValuesParams) (int64, error) {
	var revisionID int64
	if err := tx.QueryRow(ctx, `
INSERT INTO record_revisions (
    change_set_id,
    record_id,
    row_version,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5)
RETURNING revision_id
`, params.ChangeSetID, params.RecordID, params.RowVersion, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)).Scan(&revisionID); err != nil {
		return 0, fmt.Errorf("append record revision: %w", err)
	}
	return revisionID, nil
}

type recordRevisionConflictFact struct {
	FieldKey      string
	BeforePresent bool
	BeforeValue   any
	AfterPresent  bool
	AfterValue    any
}

func revisionConflictFactValue(value any, present bool) (any, error) {
	if !present {
		return nil, nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func recordRevisionConflictFacts(change LiveRecordChange) ([]recordRevisionConflictFact, error) {
	beforeRow, err := collaborationRow(change.BeforeValue)
	if err != nil {
		return nil, err
	}
	afterRow, err := collaborationRow(change.AfterValue)
	if err != nil {
		return nil, err
	}
	changedFieldKeys, err := collaboration.ChangedCellKeys(beforeRow, afterRow)
	if err != nil {
		return nil, err
	}
	beforeCells, err := revisionConflictCells(beforeRow)
	if err != nil {
		return nil, err
	}
	afterCells, err := revisionConflictCells(afterRow)
	if err != nil {
		return nil, err
	}
	facts := make([]recordRevisionConflictFact, 0, len(changedFieldKeys))
	for _, fieldKey := range changedFieldKeys {
		beforeValue, beforePresent := beforeCells[fieldKey]
		afterValue, afterPresent := afterCells[fieldKey]
		facts = append(facts, recordRevisionConflictFact{
			FieldKey:      fieldKey,
			BeforePresent: beforePresent,
			BeforeValue:   beforeValue,
			AfterPresent:  afterPresent,
			AfterValue:    afterValue,
		})
	}
	return facts, nil
}

func revisionConflictCells(row map[string]any) (map[string]any, error) {
	if row == nil || row["cells"] == nil {
		return map[string]any{}, nil
	}
	cells, ok := row["cells"].(map[string]any)
	if ok {
		return cells, nil
	}
	encoded, err := json.Marshal(row["cells"])
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	return decoded, nil
}

func collaborationRow(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	if row, ok := value.(map[string]any); ok {
		return row, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var row map[string]any
	if err := json.Unmarshal(encoded, &row); err != nil {
		return nil, err
	}
	return row, nil
}
