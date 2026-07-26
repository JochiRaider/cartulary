package mentions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateParams struct {
	SourceRecordID   uuid.UUID
	EntityType       string
	SourceFieldKey   string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	Ordinal          int
	CreatedByUserID  uuid.UUID
	CreatedAt        time.Time
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
}

func (*Store) NextOrdinalTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string) (int, error) {
	var ordinal int
	err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(ordinal), 0) + 1
  FROM entity_mentions
 WHERE source_record_id = $1
   AND source_field_key = $2
`, recordID, fieldKey).Scan(&ordinal)
	return ordinal, err
}

func (*Store) InsertTx(ctx context.Context, tx pgx.Tx, params CreateParams) error {
	_, err := tx.Exec(ctx, `
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal,
    created_by_user_id, created_at, resolved_record_id, resolved_by_user_id,
    resolved_at, resolution_method
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, $9, $10, $11, $12, $13, $14, $15)
`, params.SourceRecordID, params.EntityType, params.SourceFieldKey, params.OriginKind, params.OriginLocator, params.RawText, params.NormalizedText, params.ResolutionStatus, params.Ordinal, params.CreatedByUserID, params.CreatedAt.UTC(), params.ResolvedRecordID, params.ResolvedByUserID, params.ResolvedAt, params.ResolutionMethod)
	return err
}
