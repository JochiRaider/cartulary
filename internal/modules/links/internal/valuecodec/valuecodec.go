package valuecodec

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RecordLinkMutationValue struct {
	RecordLinkID    uuid.UUID
	IncidentID      uuid.UUID
	SrcRecordID     uuid.UUID
	DstRecordID     uuid.UUID
	LinkType        string
	FieldKey        *string
	Provenance      string
	Confidence      *int
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
	fields          map[string]any
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
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
	fields            map[string]any
}

type RecordTagIdentity struct {
	RecordTagID uuid.UUID
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
}

type RecordLinkRestorePlan struct {
	Identity        RecordLinkIdentity
	FieldKey        *string
	Provenance      string
	Confidence      *int
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type RecordTagRestorePlan struct {
	Identity          RecordTagIdentity
	TagName           string
	NormalizedTagName string
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
}

type RecordLinkMutationInput struct {
	RecordLinkID    uuid.UUID
	IncidentID      uuid.UUID
	SrcRecordID     uuid.UUID
	DstRecordID     uuid.UUID
	LinkType        string
	FieldKey        *string
	Provenance      string
	Confidence      *int
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

var recordLinkMembers = []string{
	"record_link_id", "incident_id", "src_record_id", "dst_record_id",
	"link_type", "field_key", "provenance", "confidence", "owner_user_id",
	"created_by_user_id", "decided_at", "created_at", "deleted_at",
	"deleted_by_user_id",
}

var recordTagMembers = []string{
	"record_tag_id", "incident_id", "record_id", "tag_name",
	"normalized_tag_name", "created_by_user_id", "created_at", "updated_at",
	"deleted_at", "deleted_by_user_id",
}

func ValidateHistoryMutation(targetKind string, targetID string, operationKind string, beforeValue map[string]any, afterValue map[string]any) error {
	switch targetKind {
	case "record_link":
		return validateRecordLinkHistoryMutation(targetID, operationKind, beforeValue, afterValue)
	case "record_tag":
		return validateRecordTagHistoryMutation(targetID, operationKind, beforeValue, afterValue)
	default:
		return fmt.Errorf("unsupported target kind")
	}
}

func validateRecordLinkHistoryMutation(targetID string, operationKind string, beforeValue map[string]any, afterValue map[string]any) error {
	if operationKind == "create" {
		if beforeValue != nil || afterValue == nil {
			return fmt.Errorf("invalid create sides")
		}
		after, err := DecodeRecordLinkMutationValue(afterValue)
		if err != nil || after.DeletedAt != nil || targetID != after.RecordLinkID.String() {
			return fmt.Errorf("invalid create value")
		}
		return nil
	}
	if beforeValue == nil || afterValue == nil {
		return fmt.Errorf("missing retained side")
	}
	before, err := DecodeRecordLinkMutationValue(beforeValue)
	if err != nil {
		return err
	}
	after, err := DecodeRecordLinkMutationValue(afterValue)
	if err != nil {
		return err
	}
	if targetID != before.RecordLinkID.String() || targetID != after.RecordLinkID.String() ||
		!sameRecordLinkIdentity(before, after) || reflect.DeepEqual(before.Map(), after.Map()) {
		return fmt.Errorf("unstable link identity")
	}
	switch operationKind {
	case "patch":
		if before.DeletedAt != nil || after.DeletedAt != nil ||
			before.CreatedByUserID != after.CreatedByUserID || !before.CreatedAt.Equal(after.CreatedAt) {
			return fmt.Errorf("invalid patch transition")
		}
	case "delete":
		if before.DeletedAt != nil || after.DeletedAt == nil ||
			!equalRecordLinkExceptDeletion(before, after) {
			return fmt.Errorf("invalid delete transition")
		}
	case "rollback":
		if !legalRecordLinkRollbackTransition(before, after) {
			return fmt.Errorf("invalid rollback transition")
		}
	default:
		return fmt.Errorf("unsupported operation")
	}
	return nil
}

func validateRecordTagHistoryMutation(targetID string, operationKind string, beforeValue map[string]any, afterValue map[string]any) error {
	if operationKind == "create" {
		if beforeValue != nil || afterValue == nil {
			return fmt.Errorf("invalid create sides")
		}
		after, err := DecodeRecordTagMutationValue(afterValue)
		if err != nil || after.DeletedAt != nil || targetID != recordTagTargetID(after.RecordID, after.RecordTagID) {
			return fmt.Errorf("invalid create value")
		}
		return nil
	}
	if beforeValue == nil || afterValue == nil {
		return fmt.Errorf("missing retained side")
	}
	before, err := DecodeRecordTagMutationValue(beforeValue)
	if err != nil {
		return err
	}
	after, err := DecodeRecordTagMutationValue(afterValue)
	if err != nil {
		return err
	}
	if before.RecordTagID != after.RecordTagID || before.IncidentID != after.IncidentID ||
		before.CreatedByUserID != after.CreatedByUserID || !before.CreatedAt.Equal(after.CreatedAt) ||
		before.TagName != after.TagName || before.NormalizedTagName != after.NormalizedTagName ||
		reflect.DeepEqual(before.Map(), after.Map()) {
		return fmt.Errorf("unstable tag identity")
	}
	switch operationKind {
	case "patch":
		if targetID != recordTagTargetID(before.RecordID, before.RecordTagID) ||
			before.DeletedAt != nil || after.DeletedAt != nil || !after.UpdatedAt.After(before.UpdatedAt) {
			return fmt.Errorf("invalid patch transition")
		}
	case "delete":
		if targetID != recordTagTargetID(before.RecordID, before.RecordTagID) ||
			before.RecordID != after.RecordID || before.DeletedAt != nil || after.DeletedAt == nil ||
			!after.UpdatedAt.Equal(*after.DeletedAt) || !after.UpdatedAt.After(before.UpdatedAt) {
			return fmt.Errorf("invalid delete transition")
		}
	case "rollback":
		if targetID != recordTagTargetID(after.RecordID, after.RecordTagID) ||
			!legalRecordTagRollbackTransition(before, after) {
			return fmt.Errorf("invalid rollback transition")
		}
	default:
		return fmt.Errorf("unsupported operation")
	}
	return nil
}

func LoadRecordLinkMutationValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (RecordLinkMutationValue, error) {
	var (
		input           RecordLinkMutationInput
		fieldKey        pgtype.Text
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := tx.QueryRow(ctx, `
SELECT
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
  FROM record_links
 WHERE record_link_id = $1
`, recordLinkID).Scan(
		&input.RecordLinkID,
		&input.IncidentID,
		&input.SrcRecordID,
		&input.DstRecordID,
		&input.LinkType,
		&fieldKey,
		&input.Provenance,
		&confidence,
		&input.OwnerUserID,
		&input.CreatedByUserID,
		&input.DecidedAt,
		&input.CreatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return RecordLinkMutationValue{}, err
	}
	if fieldKey.Valid {
		input.FieldKey = &fieldKey.String
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		input.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		input.DeletedAt = &value
	}
	if deletedByUserID.Valid {
		value := uuid.UUID(deletedByUserID.Bytes)
		input.DeletedByUserID = &value
	}
	return BuildRecordLinkMutationValue(input), nil
}

func LoadRecordTagMutationValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (RecordTagMutationValue, error) {
	var (
		input           RecordTagMutationInput
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := tx.QueryRow(ctx, `
SELECT
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at, deleted_by_user_id
  FROM record_tags
 WHERE record_tag_id = $1
`, recordTagID).Scan(
		&input.RecordTagID,
		&input.IncidentID,
		&input.RecordID,
		&input.TagName,
		&input.NormalizedTagName,
		&input.CreatedByUserID,
		&input.CreatedAt,
		&input.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return RecordTagMutationValue{}, err
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		input.DeletedAt = &value
	}
	if deletedByUserID.Valid {
		value := uuid.UUID(deletedByUserID.Bytes)
		input.DeletedByUserID = &value
	}
	return BuildRecordTagMutationValue(input), nil
}

func DecodeRecordLinkMutationValue(value map[string]any) (RecordLinkMutationValue, error) {
	if err := exactMembers(value, recordLinkMembers); err != nil {
		return RecordLinkMutationValue{}, err
	}
	identity, err := ParseRecordLinkIdentity(value)
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	fieldKey, err := nullableNonemptyString(value, "field_key")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	provenance, err := requiredString(value, "provenance")
	if err != nil || !isKnownProvenance(provenance) {
		return RecordLinkMutationValue{}, fmt.Errorf("invalid provenance")
	}
	confidence, err := nullableInteger(value, "confidence")
	if err != nil || (confidence != nil && (*confidence < 0 || *confidence > 100)) {
		return RecordLinkMutationValue{}, fmt.Errorf("invalid confidence")
	}
	if provenance == "manual" && confidence != nil {
		return RecordLinkMutationValue{}, fmt.Errorf("manual confidence must be null")
	}
	if provenance == "auto_match" &&
		(identity.LinkType != "observed_on_host" && identity.LinkType != "observed_as_identity" || confidence == nil || *confidence != 100) {
		return RecordLinkMutationValue{}, fmt.Errorf("invalid auto_match metadata")
	}
	ownerUserID, err := canonicalUUID(value, "owner_user_id")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	createdByUserID, err := canonicalUUID(value, "created_by_user_id")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	decidedAt, err := canonicalTimestamp(value, "decided_at")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	createdAt, err := canonicalTimestamp(value, "created_at")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	deletedAt, err := nullableCanonicalTimestamp(value, "deleted_at")
	if err != nil {
		return RecordLinkMutationValue{}, err
	}
	deletedByUserID, err := nullableCanonicalUUID(value, "deleted_by_user_id")
	if err != nil || (deletedAt == nil) != (deletedByUserID == nil) {
		return RecordLinkMutationValue{}, fmt.Errorf("invalid deletion tuple")
	}
	if deletedAt != nil && deletedAt.Before(createdAt) {
		return RecordLinkMutationValue{}, fmt.Errorf("invalid deletion timestamp")
	}
	return RecordLinkMutationValue{
		RecordLinkID:    identity.RecordLinkID,
		IncidentID:      identity.IncidentID,
		SrcRecordID:     identity.SrcRecordID,
		DstRecordID:     identity.DstRecordID,
		LinkType:        identity.LinkType,
		FieldKey:        fieldKey,
		Provenance:      provenance,
		Confidence:      confidence,
		OwnerUserID:     ownerUserID,
		CreatedByUserID: createdByUserID,
		DecidedAt:       decidedAt,
		CreatedAt:       createdAt,
		DeletedAt:       deletedAt,
		DeletedByUserID: deletedByUserID,
		fields:          canonicalRecordLinkMap(value, confidence),
	}, nil
}

func DecodeRecordTagMutationValue(value map[string]any) (RecordTagMutationValue, error) {
	if err := exactMembers(value, recordTagMembers); err != nil {
		return RecordTagMutationValue{}, err
	}
	identity, err := ParseRecordTagIdentity(value)
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	tagName, err := requiredString(value, "tag_name")
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	normalized, err := requiredString(value, "normalized_tag_name")
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	createdByUserID, err := canonicalUUID(value, "created_by_user_id")
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	createdAt, err := canonicalTimestamp(value, "created_at")
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	updatedAt, err := canonicalTimestamp(value, "updated_at")
	if err != nil || updatedAt.Before(createdAt) {
		return RecordTagMutationValue{}, fmt.Errorf("invalid updated_at")
	}
	deletedAt, err := nullableCanonicalTimestamp(value, "deleted_at")
	if err != nil {
		return RecordTagMutationValue{}, err
	}
	deletedByUserID, err := nullableCanonicalUUID(value, "deleted_by_user_id")
	if err != nil || (deletedAt == nil) != (deletedByUserID == nil) {
		return RecordTagMutationValue{}, fmt.Errorf("invalid deletion tuple")
	}
	if deletedAt != nil && !deletedAt.Equal(updatedAt) {
		return RecordTagMutationValue{}, fmt.Errorf("deleted_at must equal updated_at")
	}
	return RecordTagMutationValue{
		RecordTagID:       identity.RecordTagID,
		IncidentID:        identity.IncidentID,
		RecordID:          identity.RecordID,
		TagName:           tagName,
		NormalizedTagName: normalized,
		CreatedByUserID:   createdByUserID,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		DeletedAt:         deletedAt,
		DeletedByUserID:   deletedByUserID,
		fields:            cloneMap(value),
	}, nil
}

func DecodeRecordLinkRestorePlan(value map[string]any) (RecordLinkRestorePlan, error) {
	typedValue, err := DecodeRecordLinkMutationValue(value)
	if err != nil {
		return RecordLinkRestorePlan{}, err
	}
	return RecordLinkRestorePlan{
		Identity: RecordLinkIdentity{
			RecordLinkID: typedValue.RecordLinkID,
			IncidentID:   typedValue.IncidentID,
			SrcRecordID:  typedValue.SrcRecordID,
			DstRecordID:  typedValue.DstRecordID,
			LinkType:     typedValue.LinkType,
		},
		FieldKey:        copyStringPointer(typedValue.FieldKey),
		Provenance:      typedValue.Provenance,
		Confidence:      copyIntPointer(typedValue.Confidence),
		OwnerUserID:     typedValue.OwnerUserID,
		CreatedByUserID: typedValue.CreatedByUserID,
		DecidedAt:       typedValue.DecidedAt,
		CreatedAt:       typedValue.CreatedAt,
		DeletedAt:       copyTimePointer(typedValue.DeletedAt),
		DeletedByUserID: copyUUIDPointer(typedValue.DeletedByUserID),
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
		CreatedByUserID:   typedValue.CreatedByUserID,
		CreatedAt:         typedValue.CreatedAt,
		UpdatedAt:         typedValue.UpdatedAt,
		DeletedAt:         copyTimePointer(typedValue.DeletedAt),
		DeletedByUserID:   copyUUIDPointer(typedValue.DeletedByUserID),
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
		"confidence":         integerPointerValue(input.Confidence),
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
		RecordLinkID:    input.RecordLinkID,
		IncidentID:      input.IncidentID,
		SrcRecordID:     input.SrcRecordID,
		DstRecordID:     input.DstRecordID,
		LinkType:        input.LinkType,
		FieldKey:        copyStringPointer(input.FieldKey),
		Provenance:      input.Provenance,
		Confidence:      copyIntPointer(input.Confidence),
		OwnerUserID:     input.OwnerUserID,
		CreatedByUserID: input.CreatedByUserID,
		DecidedAt:       input.DecidedAt.UTC(),
		CreatedAt:       input.CreatedAt.UTC(),
		DeletedAt:       copyTimePointer(input.DeletedAt),
		DeletedByUserID: copyUUIDPointer(input.DeletedByUserID),
		fields:          value,
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
		CreatedByUserID:   input.CreatedByUserID,
		CreatedAt:         input.CreatedAt.UTC(),
		UpdatedAt:         input.UpdatedAt.UTC(),
		DeletedAt:         copyTimePointer(input.DeletedAt),
		DeletedByUserID:   copyUUIDPointer(input.DeletedByUserID),
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
	if identity.RecordLinkID, err = canonicalUUID(value, "record_link_id"); err != nil {
		return identity, err
	}
	if identity.IncidentID, err = canonicalUUID(value, "incident_id"); err != nil {
		return identity, err
	}
	if identity.SrcRecordID, err = canonicalUUID(value, "src_record_id"); err != nil {
		return identity, err
	}
	if identity.DstRecordID, err = canonicalUUID(value, "dst_record_id"); err != nil {
		return identity, err
	}
	if identity.SrcRecordID == identity.DstRecordID {
		return identity, fmt.Errorf("identical endpoints")
	}
	linkType, err := requiredString(value, "link_type")
	if err != nil || !isKnownLinkType(linkType) {
		return identity, fmt.Errorf("invalid link_type")
	}
	identity.LinkType = linkType
	return identity, nil
}

func ParseRecordTagIdentity(value map[string]any) (RecordTagIdentity, error) {
	var identity RecordTagIdentity
	var err error
	if identity.RecordTagID, err = canonicalUUID(value, "record_tag_id"); err != nil {
		return identity, err
	}
	if identity.IncidentID, err = canonicalUUID(value, "incident_id"); err != nil {
		return identity, err
	}
	if identity.RecordID, err = canonicalUUID(value, "record_id"); err != nil {
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
	return canonicalUUID(value, key)
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

func exactMembers(value map[string]any, members []string) error {
	if value == nil || len(value) != len(members) {
		return fmt.Errorf("invalid member set")
	}
	for _, member := range members {
		if _, ok := value[member]; !ok {
			return fmt.Errorf("missing %s", member)
		}
	}
	return nil
}

func requiredString(value map[string]any, key string) (string, error) {
	text, ok := StringFromMap(value, key)
	if !ok || text == "" {
		return "", fmt.Errorf("invalid %s", key)
	}
	return text, nil
}

func nullableNonemptyString(value map[string]any, key string) (*string, error) {
	raw, ok := value[key]
	if !ok {
		return nil, fmt.Errorf("missing %s", key)
	}
	if raw == nil {
		return nil, nil
	}
	text, ok := raw.(string)
	if !ok || text == "" || strings.TrimSpace(text) != text {
		return nil, fmt.Errorf("invalid %s", key)
	}
	return &text, nil
}

func canonicalUUID(value map[string]any, key string) (uuid.UUID, error) {
	text, err := requiredString(value, key)
	if err != nil {
		return uuid.Nil, err
	}
	parsed, err := uuid.Parse(text)
	if err != nil || parsed == uuid.Nil || parsed.String() != text {
		return uuid.Nil, fmt.Errorf("invalid %s", key)
	}
	return parsed, nil
}

func nullableCanonicalUUID(value map[string]any, key string) (*uuid.UUID, error) {
	raw, ok := value[key]
	if !ok {
		return nil, fmt.Errorf("missing %s", key)
	}
	if raw == nil {
		return nil, nil
	}
	parsed, err := canonicalUUID(value, key)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func canonicalTimestamp(value map[string]any, key string) (time.Time, error) {
	text, err := requiredString(value, key)
	if err != nil {
		return time.Time{}, err
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != text {
		return time.Time{}, fmt.Errorf("invalid %s", key)
	}
	return parsed.UTC(), nil
}

func nullableCanonicalTimestamp(value map[string]any, key string) (*time.Time, error) {
	raw, ok := value[key]
	if !ok {
		return nil, fmt.Errorf("missing %s", key)
	}
	if raw == nil {
		return nil, nil
	}
	parsed, err := canonicalTimestamp(value, key)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func nullableInteger(value map[string]any, key string) (*int, error) {
	raw, ok := value[key]
	if !ok {
		return nil, fmt.Errorf("missing %s", key)
	}
	if raw == nil {
		return nil, nil
	}
	var parsed int64
	switch typed := raw.(type) {
	case int:
		parsed = int64(typed)
	case int32:
		parsed = int64(typed)
	case int64:
		parsed = typed
	case float64:
		parsed = int64(typed)
		if float64(parsed) != typed {
			return nil, fmt.Errorf("invalid %s", key)
		}
	case json.Number:
		value, err := strconv.ParseInt(typed.String(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s", key)
		}
		parsed = value
	default:
		return nil, fmt.Errorf("invalid %s", key)
	}
	if int64(int(parsed)) != parsed {
		return nil, fmt.Errorf("invalid %s", key)
	}
	result := int(parsed)
	return &result, nil
}

func isKnownLinkType(value string) bool {
	switch value {
	case "observed_on_host", "observed_as_identity", "references_indicator",
		"attached_evidence", "references_artifact", "derived_from", "merged_into",
		"supported_by", "references_record", "supersedes":
		return true
	default:
		return false
	}
}

func isKnownProvenance(value string) bool {
	switch value {
	case "manual", "auto_match", "import", "rollback", "system":
		return true
	default:
		return false
	}
}

func recordTagTargetID(recordID uuid.UUID, recordTagID uuid.UUID) string {
	return "record_tag:" + recordID.String() + ":" + recordTagID.String()
}

func sameRecordLinkIdentity(left RecordLinkMutationValue, right RecordLinkMutationValue) bool {
	return left.RecordLinkID == right.RecordLinkID && left.IncidentID == right.IncidentID &&
		left.SrcRecordID == right.SrcRecordID && left.DstRecordID == right.DstRecordID &&
		left.LinkType == right.LinkType
}

func equalRecordLinkExceptDeletion(left RecordLinkMutationValue, right RecordLinkMutationValue) bool {
	return sameRecordLinkIdentity(left, right) && equalStringPointers(left.FieldKey, right.FieldKey) && left.Provenance == right.Provenance &&
		equalIntPointers(left.Confidence, right.Confidence) && left.OwnerUserID == right.OwnerUserID &&
		left.CreatedByUserID == right.CreatedByUserID && left.DecidedAt.Equal(right.DecidedAt) &&
		left.CreatedAt.Equal(right.CreatedAt)
}

func legalRecordLinkRollbackTransition(before RecordLinkMutationValue, after RecordLinkMutationValue) bool {
	if before.CreatedByUserID != after.CreatedByUserID || !before.CreatedAt.Equal(after.CreatedAt) {
		return false
	}
	switch {
	case before.DeletedAt == nil && after.DeletedAt != nil:
		return equalRecordLinkExceptDeletion(before, after)
	case before.DeletedAt != nil && after.DeletedAt == nil:
		return equalRecordLinkExceptDeletion(before, after)
	case before.DeletedAt == nil && after.DeletedAt == nil:
		return true
	default:
		return false
	}
}

func legalRecordTagRollbackTransition(before RecordTagMutationValue, after RecordTagMutationValue) bool {
	switch {
	case before.DeletedAt == nil && after.DeletedAt != nil:
		return before.RecordID == after.RecordID
	case before.DeletedAt != nil && after.DeletedAt == nil:
		return before.RecordID == after.RecordID
	case before.DeletedAt == nil && after.DeletedAt == nil:
		return true
	default:
		return false
	}
}

func canonicalRecordLinkMap(value map[string]any, confidence *int) map[string]any {
	canonical := cloneMap(value)
	canonical["confidence"] = integerPointerValue(confidence)
	return canonical
}

func equalStringPointers(left *string, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func equalIntPointers(left *int, right *int) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func copyStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}

func copyUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
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

func integerPointerValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = cloneJSONValue(value)
	}
	return cloned
}

func cloneJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneJSONValue(item)
		}
		return cloned
	default:
		return value
	}
}
