package revisionprovider

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/valuecodec"
)

var (
	ErrTargetNotFound      = errors.New("links revision provider: target not found")
	ErrStaleTarget         = errors.New("links revision provider: stale target")
	ErrTargetNotReversible = errors.New("links revision provider: target not reversible")
)

type Provider struct{}

type RecordTagIdentity = valuecodec.RecordTagIdentity

func NewProvider() Provider {
	return Provider{}
}

func (Provider) ValidateRecordLinkValue(value map[string]any) error {
	_, err := valuecodec.DecodeRecordLinkMutationValue(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	return nil
}

func (Provider) ParseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	parsed, err := valuecodec.DecodeRecordTagMutationValue(value)
	if err != nil {
		return RecordTagIdentity{}, ErrTargetNotReversible
	}
	return RecordTagIdentity{RecordTagID: parsed.RecordTagID, IncidentID: parsed.IncidentID, RecordID: parsed.RecordID}, nil
}

func (Provider) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) TombstoneRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $3,
       deleted_by_user_id = $4
 WHERE record_link_id = $1
   AND incident_id = $2
   AND deleted_at IS NULL
`, recordLinkID, incidentID, now.UTC(), actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleTarget
	}
	return nil
}

func (Provider) RestoreRecordLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordLinkID uuid.UUID, value map[string]any, actorUserID uuid.UUID, now time.Time) error {
	typedValue, err := valuecodec.DecodeRecordLinkMutationValue(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	identity := valuecodec.RecordLinkIdentity{
		RecordLinkID: typedValue.RecordLinkID,
		IncidentID:   typedValue.IncidentID,
		SrcRecordID:  typedValue.SrcRecordID,
		DstRecordID:  typedValue.DstRecordID,
		LinkType:     typedValue.LinkType,
	}
	if identity.IncidentID != incidentID || identity.RecordLinkID != recordLinkID {
		return ErrTargetNotFound
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET src_record_id = $3,
       dst_record_id = $4,
       link_type = $5,
       field_key = $6,
       provenance = $7,
       confidence = $8,
       owner_user_id = $9,
       decided_at = COALESCE($10, decided_at),
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_link_id = $1
   AND incident_id = $2
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, nullableStringAny(value, "field_key"), stringDefault(value, "provenance", "rollback"), nullableAny(value, "confidence"), uuidAnyDefault(value, "owner_user_id", actorUserID), nullableAny(value, "decided_at"))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	_, err = tx.Exec(ctx, `
INSERT INTO record_links (
    record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, COALESCE($10, $11), $11)
`, recordLinkID, incidentID, identity.SrcRecordID, identity.DstRecordID, identity.LinkType, nullableStringAny(value, "field_key"), stringDefault(value, "provenance", "rollback"), nullableAny(value, "confidence"), uuidAnyDefault(value, "owner_user_id", actorUserID), nullableAny(value, "decided_at"), now.UTC())
	return err
}

func (Provider) LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func (Provider) RestoreRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, value map[string]any, now time.Time) error {
	typedValue, err := valuecodec.DecodeRecordTagMutationValue(value)
	if err != nil {
		return ErrTargetNotReversible
	}
	identity := valuecodec.RecordTagIdentity{
		RecordTagID: typedValue.RecordTagID,
		IncidentID:  typedValue.IncidentID,
		RecordID:    typedValue.RecordID,
	}
	if identity.RecordTagID != recordTagID {
		return ErrTargetNotFound
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET record_id = $2,
       tag_name = $3,
       normalized_tag_name = $4,
       updated_at = $5,
       deleted_at = NULL,
       deleted_by_user_id = NULL
 WHERE record_tag_id = $1
   AND incident_id = $6
`, recordTagID, identity.RecordID, stringDefault(value, "tag_name", ""), stringDefault(value, "normalized_tag_name", ""), now.UTC(), identity.IncidentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrTargetNotFound
	}
	return nil
}

func (Provider) TombstoneRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, actorUserID uuid.UUID, now time.Time) error {
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2
 WHERE record_tag_id = $1
   AND deleted_at IS NULL
`, recordTagID, now.UTC(), actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrStaleTarget
	}
	return nil
}

func stringFromMap(value map[string]any, key string) (string, bool) {
	if value == nil {
		return "", false
	}
	raw, ok := value[key]
	if !ok || raw == nil {
		return "", false
	}
	text, ok := raw.(string)
	return text, ok
}

func stringDefault(value map[string]any, key string, fallback string) string {
	if text, ok := stringFromMap(value, key); ok {
		return text
	}
	return fallback
}

func nullableAny(value map[string]any, key string) any {
	if value == nil {
		return nil
	}
	raw, ok := value[key]
	if !ok {
		return nil
	}
	return raw
}

func nullableStringAny(value map[string]any, key string) any {
	if text, ok := stringFromMap(value, key); ok && text != "" {
		return text
	}
	return nil
}

func uuidAnyDefault(value map[string]any, key string, fallback uuid.UUID) any {
	if text, ok := stringFromMap(value, key); ok && text != "" {
		if parsed, err := uuid.Parse(text); err == nil {
			return parsed
		}
	}
	return fallback
}
