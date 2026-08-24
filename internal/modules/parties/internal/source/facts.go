package source

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type RecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func LoadRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RecordMeta, error) {
	var meta RecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return RecordMeta{}, err
	}
	if deletedAt.Valid {
		return RecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}
