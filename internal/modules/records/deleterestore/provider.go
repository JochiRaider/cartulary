package deleterestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SourceProvider owns delete/restore behavior for one first-class record source.
// Revisions retains transaction, envelope, history, projection, and event orchestration.
type SourceProvider interface {
	SnapshotTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error
	ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error)
	ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error)
}

type TableProvider struct {
	SourceTable        string
	SourceRecordCol    string
	StaticViewSchemaID string
	SourceTombstone    bool
}

func (p TableProvider) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	query := fmt.Sprintf(`
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(s))
  FROM records r
  JOIN %s s
    ON s.%s = r.record_id
 WHERE r.record_id = $1
`, p.SourceTable, p.SourceRecordCol)
	var raw []byte
	if err := tx.QueryRow(ctx, query, recordID).Scan(&raw); err != nil {
		return nil, err
	}
	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (p TableProvider) UpdateSourceDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, deleting bool) error {
	if !p.SourceTombstone {
		return nil
	}
	if deleting {
		_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2
 WHERE %s = $1
`, p.SourceTable, p.SourceRecordCol), recordID, now.UTC(), actorUserID)
		return err
	}
	_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s
   SET deleted_at = NULL,
       deleted_by_user_id = NULL,
       updated_at = $2
 WHERE %s = $1
`, p.SourceTable, p.SourceRecordCol), recordID, now.UTC())
	return err
}

func (p TableProvider) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return p.StaticViewSchemaID, nil
}

func (p TableProvider) ValidateDeletePreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (string, bool, error) {
	return "", false, nil
}
