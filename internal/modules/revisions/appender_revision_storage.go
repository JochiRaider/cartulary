package revisions

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (*Appender) AppendLiveRevisionTx(ctx context.Context, tx pgx.Tx, params LiveRevisionInput) error {
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
	facts, err := canonicalRevisionConflictFacts(params.ConflictFacts)
	if err != nil {
		return fmt.Errorf("validate record revision conflict facts: %w", err)
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

func (*Appender) AppendHistoricalRevisionTx(ctx context.Context, tx pgx.Tx, params HistoricalRevisionInput) error {
	beforeValue, afterValue, err := recordSnapshotPair(params.RecordID, params.BeforeSnapshot, params.AfterSnapshot)
	if err != nil {
		return err
	}
	_, err = appendRecordRevisionValuesTx(ctx, tx, appendRecordRevisionValuesParams{
		ChangeSetID: params.ChangeSetID,
		RecordID:    params.RecordID,
		RowVersion:  params.RowVersion,
		BeforeValue: beforeValue,
		AfterValue:  afterValue,
	})
	return err
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

func canonicalRevisionConflictFacts(input []RevisionConflictFact) ([]RevisionConflictFact, error) {
	facts := append([]RevisionConflictFact(nil), input...)
	slices.SortFunc(facts, func(left RevisionConflictFact, right RevisionConflictFact) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	for index, fact := range facts {
		if strings.TrimSpace(fact.FieldKey) == "" || fact.FieldKey != strings.TrimSpace(fact.FieldKey) {
			return nil, fmt.Errorf("field key %q is invalid", fact.FieldKey)
		}
		if index > 0 && facts[index-1].FieldKey == fact.FieldKey {
			return nil, fmt.Errorf("field key %q is duplicated", fact.FieldKey)
		}
	}
	return facts, nil
}
