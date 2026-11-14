package rollback

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type provider struct{}

var _ rollbackcontract.RowSourceProvider = provider{}

func NewProvider() rollbackcontract.RowSourceProvider { return provider{} }

func (provider) ValidateRollbackValue(value map[string]any) error {
	if _, ok := sourceForRollbackValue(value); !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

func (provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := sourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	values := make([]any, 0, 13)
	for _, key := range []string{
		"title", "lifecycle_state", "requested_at", "received_at", "storage_ref", "blob_hash",
		"collector_party_text", "collector_party_id", "source_party_text", "source_party_id",
		"upload_state", "object_blob_id", "owner_user_id",
	} {
		value := source[key]
		if key == "collector_party_id" || key == "source_party_id" || key == "object_blob_id" || key == "owner_user_id" {
			parsed, _, err := nullableUUID(source, key)
			if err != nil {
				return rollbackcontract.ErrTargetNotReversible
			}
			value = parsed
		}
		values = append(values, value)
	}
	_, err := tx.Exec(ctx, `
UPDATE evidence
   SET title = $2::text,
       lifecycle_state = $3::text,
       requested_at = $4::timestamptz,
       received_at = $5::timestamptz,
       storage_ref = $6::text,
       blob_hash = $7::text,
       collector_party_text = $8::text,
       collector_party_id = $9::uuid,
       source_party_text = $10::text,
       source_party_id = $11::uuid,
       upload_state = $12::text,
       object_blob_id = $13::uuid,
       owner_user_id = $14::uuid,
       updated_at = $15
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	source, sourceOK := objectMap(value, "source")
	record, recordOK := objectMap(value, "record")
	if !sourceOK || !recordOK || len(value) != 3 || value["snapshot_schema_id"] != "cartulary.revisions.snapshot.evidence.v2" ||
		sourcevalue.ValidateSnapshotSource(record, source) != nil {
		return nil, false
	}
	return source, true
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func nullableUUID(value map[string]any, key string) (any, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	if parsed == uuid.Nil {
		return nil, true, rollbackcontract.ErrTargetNotReversible
	}
	return parsed, true, nil
}
