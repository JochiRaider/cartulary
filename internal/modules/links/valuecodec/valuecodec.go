package valuecodec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RecordLinkIdentity struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
}

type RecordTagIdentity struct {
	RecordTagID uuid.UUID
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
}

func LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
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
`, recordLinkID).Scan(&raw); err != nil {
		return nil, err
	}
	return DecodeMutationValue(raw)
}

func LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
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
`, recordTagID).Scan(&raw); err != nil {
		return nil, err
	}
	return DecodeMutationValue(raw)
}

func ParseRecordLinkIdentity(value map[string]any) (RecordLinkIdentity, error) {
	var identity RecordLinkIdentity
	var err error
	if identity.RecordLinkID, err = UUIDFromMap(value, "record_link_id"); err != nil {
		return identity, err
	}
	if identity.IncidentID, err = UUIDFromMap(value, "incident_id"); err != nil {
		return identity, err
	}
	if identity.SrcRecordID, err = UUIDFromMap(value, "src_record_id"); err != nil {
		return identity, err
	}
	if identity.DstRecordID, err = UUIDFromMap(value, "dst_record_id"); err != nil {
		return identity, err
	}
	linkType, ok := StringFromMap(value, "link_type")
	if !ok || linkType == "" {
		return identity, fmt.Errorf("missing link_type")
	}
	identity.LinkType = linkType
	return identity, nil
}

func ParseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	var identity RecordTagIdentity
	var err error
	if identity.RecordTagID, err = UUIDFromMap(value, "record_tag_id"); err != nil {
		return identity, err
	}
	if identity.IncidentID, err = UUIDFromMap(value, "incident_id"); err != nil {
		return identity, err
	}
	if identity.RecordID, err = UUIDFromMap(value, "record_id"); err != nil {
		return identity, err
	}
	if tagName, ok := StringFromMap(value, "tag_name"); !ok || tagName == "" {
		return identity, fmt.Errorf("missing tag_name")
	}
	if normalized, ok := StringFromMap(value, "normalized_tag_name"); !ok || normalized == "" {
		return identity, fmt.Errorf("missing normalized_tag_name")
	}
	return identity, nil
}

func DecodeMutationValue(raw []byte) (map[string]any, error) {
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func UUIDFromMap(value map[string]any, key string) (uuid.UUID, error) {
	text, ok := StringFromMap(value, key)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("missing %s", key)
	}
	return uuid.Parse(text)
}

func StringFromMap(value map[string]any, key string) (string, bool) {
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
