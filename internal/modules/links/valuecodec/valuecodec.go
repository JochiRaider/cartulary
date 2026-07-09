package valuecodec

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RecordLinkMutationValue struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
	fields       map[string]any
}

type RecordLinkIdentity struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
}

type RecordTagMutationValue struct {
	RecordTagID       uuid.UUID
	IncidentID        uuid.UUID
	RecordID          uuid.UUID
	TagName           string
	NormalizedTagName string
	fields            map[string]any
}

type RecordTagIdentity struct {
	RecordTagID uuid.UUID
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
}

func LoadRecordLinkMutationValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (RecordLinkMutationValue, error) {
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
		return RecordLinkMutationValue{}, err
	}
	value, err := DecodeMutationValue(raw)
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	return DecodeRecordLinkMutationValue(value)
}

func LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := LoadRecordLinkMutationValueTx(ctx, tx, recordLinkID)
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func LoadRecordTagMutationValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (RecordTagMutationValue, error) {
	var raw []byte
	if err := tx.QueryRow(ctx, `
SELECT jsonb_build_object(
    'record_tag_id', record_tag_id::text,
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
		return RecordTagMutationValue{}, err
	}
	value, err := DecodeMutationValue(raw)
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	return DecodeRecordTagMutationValue(value)
}

func LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := LoadRecordTagMutationValueTx(ctx, tx, recordTagID)
	if err != nil {
		return nil, err
	}
	return value.Map(), nil
}

func DecodeRecordLinkMutationValue(value map[string]any) (RecordLinkMutationValue, error) {
	identity, err := ParseRecordLinkIdentity(value)
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	return RecordLinkMutationValue{
		RecordLinkID: identity.RecordLinkID,
		IncidentID:   identity.IncidentID,
		SrcRecordID:  identity.SrcRecordID,
		DstRecordID:  identity.DstRecordID,
		LinkType:     identity.LinkType,
		fields:       cloneMap(value),
	}, nil
}

func DecodeRecordTagMutationValue(value map[string]any) (RecordTagMutationValue, error) {
	identity, err := ParseRecordTagIdentity(value)
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	tagName, _ := StringFromMap(value, "tag_name")
	normalized, _ := StringFromMap(value, "normalized_tag_name")
	canonical := cloneMap(value)
	canonical["record_tag_id"] = identity.RecordTagID.String()
	delete(canonical, "tag_id")
	return RecordTagMutationValue{
		RecordTagID:       identity.RecordTagID,
		IncidentID:        identity.IncidentID,
		RecordID:          identity.RecordID,
		TagName:           tagName,
		NormalizedTagName: normalized,
		fields:            canonical,
	}, nil
}

func (v RecordLinkMutationValue) Map() map[string]any {
	return cloneMap(v.fields)
}

func (v RecordTagMutationValue) Map() map[string]any {
	return cloneMap(v.fields)
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
		if identity.RecordTagID, err = UUIDFromMap(value, "tag_id"); err != nil {
			return identity, err
		}
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

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		if nested, ok := value.([]any); ok {
			cloned[key] = slices.Clone(nested)
			continue
		}
		cloned[key] = value
	}
	return cloned
}
