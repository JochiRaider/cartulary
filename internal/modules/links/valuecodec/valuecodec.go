package valuecodec

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

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

type RecordLinkRestorePlan struct {
	Identity    RecordLinkIdentity
	FieldKey    *string
	Provenance  string
	Confidence  any
	OwnerUserID uuid.UUID
	DecidedAt   any
}

type RecordTagRestorePlan struct {
	Identity          RecordTagIdentity
	TagName           string
	NormalizedTagName string
}

type RecordLinkMutationInput struct {
	RecordLinkID    uuid.UUID
	IncidentID      uuid.UUID
	SrcRecordID     uuid.UUID
	DstRecordID     uuid.UUID
	LinkType        string
	FieldKey        *string
	Provenance      string
	Confidence      any
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type RecordTagMutationInput struct {
	RecordTagID       uuid.UUID
	IncidentID        uuid.UUID
	RecordID          uuid.UUID
	TagName           string
	NormalizedTagName string
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
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

func DecodeRecordLinkRestorePlan(value map[string]any, fallbackOwnerUserID uuid.UUID) (RecordLinkRestorePlan, error) {
	typedValue, err := DecodeRecordLinkMutationValue(value)
	if err != nil {
		return RecordLinkRestorePlan{}, err
	}
	var fieldKey *string
	if text, ok := StringFromMap(value, "field_key"); ok && text != "" {
		fieldKey = &text
	}
	return RecordLinkRestorePlan{
		Identity: RecordLinkIdentity{
			RecordLinkID: typedValue.RecordLinkID,
			IncidentID:   typedValue.IncidentID,
			SrcRecordID:  typedValue.SrcRecordID,
			DstRecordID:  typedValue.DstRecordID,
			LinkType:     typedValue.LinkType,
		},
		FieldKey:    fieldKey,
		Provenance:  stringDefault(value, "provenance", "rollback"),
		Confidence:  nullableAny(value, "confidence"),
		OwnerUserID: uuidDefault(value, "owner_user_id", fallbackOwnerUserID),
		DecidedAt:   nullableAny(value, "decided_at"),
	}, nil
}

func DecodeRecordTagRestorePlan(value map[string]any) (RecordTagRestorePlan, error) {
	typedValue, err := DecodeRecordTagMutationValue(value)
	if err != nil {
		return RecordTagRestorePlan{}, err
	}
	return RecordTagRestorePlan{
		Identity: RecordTagIdentity{
			RecordTagID: typedValue.RecordTagID,
			IncidentID:  typedValue.IncidentID,
			RecordID:    typedValue.RecordID,
		},
		TagName:           typedValue.TagName,
		NormalizedTagName: typedValue.NormalizedTagName,
	}, nil
}

func BuildRecordLinkMutationValue(input RecordLinkMutationInput) RecordLinkMutationValue {
	value := map[string]any{
		"record_link_id":     input.RecordLinkID.String(),
		"incident_id":        input.IncidentID.String(),
		"src_record_id":      input.SrcRecordID.String(),
		"dst_record_id":      input.DstRecordID.String(),
		"link_type":          input.LinkType,
		"field_key":          nil,
		"provenance":         input.Provenance,
		"confidence":         input.Confidence,
		"owner_user_id":      input.OwnerUserID.String(),
		"created_by_user_id": input.CreatedByUserID.String(),
		"decided_at":         input.DecidedAt.UTC().Format(time.RFC3339Nano),
		"created_at":         input.CreatedAt.UTC().Format(time.RFC3339Nano),
		"deleted_at":         formatTimestampPointer(input.DeletedAt),
		"deleted_by_user_id": formatUUIDPointer(input.DeletedByUserID),
	}
	if input.FieldKey != nil {
		value["field_key"] = *input.FieldKey
	}
	return RecordLinkMutationValue{
		RecordLinkID: input.RecordLinkID,
		IncidentID:   input.IncidentID,
		SrcRecordID:  input.SrcRecordID,
		DstRecordID:  input.DstRecordID,
		LinkType:     input.LinkType,
		fields:       value,
	}
}

func BuildRecordTagMutationValue(input RecordTagMutationInput) RecordTagMutationValue {
	value := map[string]any{
		"record_tag_id":       input.RecordTagID.String(),
		"incident_id":         input.IncidentID.String(),
		"record_id":           input.RecordID.String(),
		"tag_name":            input.TagName,
		"normalized_tag_name": input.NormalizedTagName,
		"created_by_user_id":  input.CreatedByUserID.String(),
		"created_at":          input.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":          input.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"deleted_at":          formatTimestampPointer(input.DeletedAt),
		"deleted_by_user_id":  formatUUIDPointer(input.DeletedByUserID),
	}
	return RecordTagMutationValue{
		RecordTagID:       input.RecordTagID,
		IncidentID:        input.IncidentID,
		RecordID:          input.RecordID,
		TagName:           input.TagName,
		NormalizedTagName: input.NormalizedTagName,
		fields:            value,
	}
}

func (v RecordLinkMutationValue) Map() map[string]any {
	return cloneMap(v.fields)
}

func (v RecordTagMutationValue) Map() map[string]any {
	return cloneMap(v.fields)
}

func (p RecordLinkRestorePlan) FieldKeyValue() any {
	if p.FieldKey == nil {
		return nil
	}
	return *p.FieldKey
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

func stringDefault(value map[string]any, key string, fallback string) string {
	if text, ok := StringFromMap(value, key); ok {
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

func uuidDefault(value map[string]any, key string, fallback uuid.UUID) uuid.UUID {
	if text, ok := StringFromMap(value, key); ok && text != "" {
		if parsed, err := uuid.Parse(text); err == nil {
			return parsed
		}
	}
	return fallback
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
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
