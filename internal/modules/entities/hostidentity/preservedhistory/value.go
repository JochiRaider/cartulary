// Package preservedhistory owns retained preserved-identifier history values.
package preservedhistory

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var ErrInvalidValue = errors.New("invalid retained preserved identifier")

type Value struct {
	ID, IncidentID, RecordID, CreatedByUserID                             uuid.UUID
	EntityType, IdentifierType, RawValue, NormalizedValue, Classification string
	CreatedAt                                                             time.Time
	DeletedAt                                                             *time.Time
}

func (v Value) TargetID() string { return "entity_preserved_identifier:" + v.ID.String() }

func (v Value) Map() map[string]any {
	var deleted any
	if v.DeletedAt != nil {
		deleted = v.DeletedAt.UTC().Format(time.RFC3339Nano)
	}
	return map[string]any{
		"entity_preserved_identifier_id": v.ID.String(), "incident_id": v.IncidentID.String(), "record_id": v.RecordID.String(),
		"entity_type": v.EntityType, "identifier_type": v.IdentifierType, "raw_value": v.RawValue, "normalized_value": v.NormalizedValue,
		"classification": v.Classification, "created_by_user_id": v.CreatedByUserID.String(), "created_at": v.CreatedAt.UTC().Format(time.RFC3339Nano), "deleted_at": deleted,
	}
}

func Decode(input map[string]any) (Value, error) {
	fields := []string{"entity_preserved_identifier_id", "incident_id", "record_id", "entity_type", "identifier_type", "raw_value", "normalized_value", "classification", "created_by_user_id", "created_at", "deleted_at"}
	if len(input) != len(fields) {
		return Value{}, ErrInvalidValue
	}
	for _, field := range fields {
		if _, ok := input[field]; !ok {
			return Value{}, ErrInvalidValue
		}
	}
	v := Value{}
	for key, out := range map[string]*uuid.UUID{"entity_preserved_identifier_id": &v.ID, "incident_id": &v.IncidentID, "record_id": &v.RecordID, "created_by_user_id": &v.CreatedByUserID} {
		text, ok := input[key].(string)
		id, err := uuid.Parse(text)
		if !ok || err != nil || id == uuid.Nil || id.String() != text {
			return Value{}, ErrInvalidValue
		}
		*out = id
	}
	for key, out := range map[string]*string{"entity_type": &v.EntityType, "identifier_type": &v.IdentifierType, "raw_value": &v.RawValue, "normalized_value": &v.NormalizedValue, "classification": &v.Classification} {
		text, ok := input[key].(string)
		if !ok || !utf8.ValidString(text) || strings.ContainsRune(text, 0) || strings.TrimSpace(text) == "" {
			return Value{}, ErrInvalidValue
		}
		*out = text
	}
	allowed := false
	switch v.EntityType {
	case "host":
		switch v.IdentifierType {
		case "aad_device_id", "fqdn", "hostname":
			allowed = true
		}
	case "identity":
		switch v.IdentifierType {
		case "aad_object_id", "sid", "upn", "email", "sam_account_name":
			allowed = true
		}
	}
	if !allowed {
		return Value{}, ErrInvalidValue
	}
	switch v.Classification {
	case "exact_match_reuse", "suggestion_only", "provenance_only":
	default:
		return Value{}, ErrInvalidValue
	}
	normalized, ok := fieldnorm.NormalizeIdentifier(v.IdentifierType, v.RawValue)
	if !ok || normalized != v.NormalizedValue {
		return Value{}, ErrInvalidValue
	}
	v.CreatedAt, ok = canonicalTime(input["created_at"])
	if !ok {
		return Value{}, ErrInvalidValue
	}
	if input["deleted_at"] != nil {
		at, ok := canonicalTime(input["deleted_at"])
		if !ok || at.Before(v.CreatedAt) {
			return Value{}, ErrInvalidValue
		}
		v.DeletedAt = &at
	}
	return v, nil
}

func ValidateMutation(target, operation string, beforeValue, afterValue map[string]any) error {
	after, err := Decode(afterValue)
	if err != nil || after.TargetID() != target {
		return ErrInvalidValue
	}
	switch operation {
	case "create":
		if beforeValue != nil || after.DeletedAt != nil {
			return ErrInvalidValue
		}
	case "rollback":
		before, err := Decode(beforeValue)
		if err != nil || before.TargetID() != target || before.DeletedAt != nil || after.DeletedAt == nil {
			return ErrInvalidValue
		}
		after.DeletedAt = nil
		if before != after {
			return ErrInvalidValue
		}
	default:
		return ErrInvalidValue
	}
	return nil
}

func canonicalTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	at, err := time.Parse(time.RFC3339Nano, text)
	return at, err == nil && at.Year() >= 1 && at.Year() <= 9999 && at.Nanosecond()%1000 == 0 && at.UTC().Format(time.RFC3339Nano) == text
}
