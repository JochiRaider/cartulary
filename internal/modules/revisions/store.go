package revisions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store struct{}

type ChangeSetParams struct {
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type MutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params ChangeSetParams) (uuid.UUID, error) {
	var changeSetID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO change_sets (
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING change_set_id
`, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()).Scan(&changeSetID); err != nil {
		return uuid.UUID{}, fmt.Errorf("insert change set: %w", err)
	}
	return changeSetID, nil
}

func (s *Store) InsertMutationTx(ctx context.Context, tx pgx.Tx, params MutationParams) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO change_set_mutations (
    change_set_id,
    sequence_no,
    target_kind,
    target_id,
    operation_kind,
    before_version_id,
    after_version_id,
    before_value,
    after_value
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, params.ChangeSetID, params.SequenceNo, params.TargetKind, params.TargetID, params.OperationKind, params.BeforeVersionID, params.AfterVersionID, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("insert change-set mutation: %w", err)
	}
	return nil
}

func (s *Store) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params RecordRevisionParams) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO record_revisions (
    change_set_id,
    record_id,
    row_version,
    before_json,
    after_json
)
VALUES ($1, $2, $3, $4, $5)
`, params.ChangeSetID, params.RecordID, params.RowVersion, jsonOrNil(params.BeforeValue), jsonOrNil(params.AfterValue)); err != nil {
		return fmt.Errorf("insert record revision: %w", err)
	}
	return nil
}

func jsonOrNil(value any) any {
	if value == nil {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}
