package revisionprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrTargetNotFound      = errors.New("links revision provider: target not found")
	ErrStaleTarget         = errors.New("links revision provider: stale target")
	ErrTargetNotReversible = errors.New("links revision provider: target not reversible")
)

type Provider struct{}

type RecordTagIdentity struct {
	RecordTagID uuid.UUID
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
}

func NewProvider() Provider {
	return Provider{}
}

func (Provider) ValidateRecordLinkValue(value map[string]any) error {
	_, err := recordLinkIdentity(value)
	return err
}

func (Provider) ParseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	return parseRecordTagIdentity(value)
}

func (Provider) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_link_id', record_link_id::text,
    'incident_id', incident_id::text,
    'src_record_id', src_record_id::text,
    'dst_record_id', dst_record_id::text,
    'link_type', link_type,
    'field_key', field_key,
    'provenance', provenance,
    'confidence', confidence,
    'owner_user_id', owner_user_id::text,
    'created_by_user_id', created_by_user_id::text,
    'decided_at', decided_at,
    'created_at', created_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM record_links
 WHERE record_link_id = $1
`, recordLinkID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return decodeValue(raw)
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
	identity, err := recordLinkIdentity(value)
	if err != nil {
		return err
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
	var raw []byte
	err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_tag_id', record_tag_id::text,
    'tag_id', record_tag_id::text,
    'incident_id', incident_id::text,
    'record_id', record_id::text,
    'tag_name', tag_name,
    'normalized_tag_name', normalized_tag_name,
    'created_by_user_id', created_by_user_id::text,
    'created_at', created_at,
    'updated_at', updated_at,
    'deleted_at', deleted_at,
    'deleted_by_user_id', deleted_by_user_id::text
)
  FROM record_tags
 WHERE record_tag_id = $1
`, recordTagID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTargetNotFound
	}
	if err != nil {
		return nil, err
	}
	return decodeValue(raw)
}

func (Provider) RestoreRecordTagTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID, value map[string]any, now time.Time) error {
	identity, err := parseRecordTagIdentity(value)
	if err != nil {
		return err
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

type recordLinkIdentityValue struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
}

func recordLinkIdentity(value map[string]any) (recordLinkIdentityValue, error) {
	var identity recordLinkIdentityValue
	var err error
	if identity.RecordLinkID, err = uuidFromMap(value, "record_link_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if identity.IncidentID, err = uuidFromMap(value, "incident_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if identity.SrcRecordID, err = uuidFromMap(value, "src_record_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if identity.DstRecordID, err = uuidFromMap(value, "dst_record_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	linkType, ok := stringFromMap(value, "link_type")
	if !ok || linkType == "" {
		return identity, ErrTargetNotReversible
	}
	identity.LinkType = linkType
	return identity, nil
}

func parseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	var identity RecordTagIdentity
	var err error
	if identity.RecordTagID, err = uuidFromMap(value, "record_tag_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if identity.IncidentID, err = uuidFromMap(value, "incident_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if identity.RecordID, err = uuidFromMap(value, "record_id"); err != nil {
		return identity, ErrTargetNotReversible
	}
	if tagName, ok := stringFromMap(value, "tag_name"); !ok || tagName == "" {
		return identity, ErrTargetNotReversible
	}
	if normalized, ok := stringFromMap(value, "normalized_tag_name"); !ok || normalized == "" {
		return identity, ErrTargetNotReversible
	}
	return identity, nil
}

func decodeValue(raw []byte) (map[string]any, error) {
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func uuidFromMap(value map[string]any, key string) (uuid.UUID, error) {
	text, ok := stringFromMap(value, key)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("missing %s", key)
	}
	return uuid.Parse(text)
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
