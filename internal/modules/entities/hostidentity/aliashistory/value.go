// Package aliashistory owns complete retained Entity Alias values.
package aliashistory

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var ErrInvalidValue = errors.New("invalid retained entity alias")

type Value struct {
	ID, IncidentID, RecordID, CreatedByUserID           uuid.UUID
	EntityType, RawText, NormalizedText, Classification string
	CreatedAt                                           time.Time
	DeletedAt                                           *time.Time
}

func (v Value) TargetID() string { return "entity_alias:" + v.ID.String() }

func (v Value) Map() map[string]any {
	var deleted any
	if v.DeletedAt != nil {
		deleted = v.DeletedAt.UTC().Format(time.RFC3339Nano)
	}
	return map[string]any{
		"entity_alias_id": v.ID.String(), "incident_id": v.IncidentID.String(), "record_id": v.RecordID.String(),
		"entity_type": v.EntityType, "raw_text": v.RawText, "normalized_text": v.NormalizedText, "classification": v.Classification,
		"created_by_user_id": v.CreatedByUserID.String(), "created_at": v.CreatedAt.UTC().Format(time.RFC3339Nano), "deleted_at": deleted,
	}
}

func Decode(input map[string]any) (Value, error) {
	fields := []string{"entity_alias_id", "incident_id", "record_id", "entity_type", "raw_text", "normalized_text", "classification", "created_by_user_id", "created_at", "deleted_at"}
	if len(input) != len(fields) {
		return Value{}, ErrInvalidValue
	}
	for _, field := range fields {
		if _, present := input[field]; !present {
			return Value{}, ErrInvalidValue
		}
	}
	v := Value{}
	for key, out := range map[string]*uuid.UUID{"entity_alias_id": &v.ID, "incident_id": &v.IncidentID, "record_id": &v.RecordID, "created_by_user_id": &v.CreatedByUserID} {
		text, ok := input[key].(string)
		id, err := uuid.Parse(text)
		if !ok || err != nil || id == uuid.Nil || id.String() != text {
			return Value{}, ErrInvalidValue
		}
		*out = id
	}
	for key, out := range map[string]*string{"entity_type": &v.EntityType, "raw_text": &v.RawText, "normalized_text": &v.NormalizedText, "classification": &v.Classification} {
		text, ok := input[key].(string)
		if !ok || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
			return Value{}, ErrInvalidValue
		}
		*out = text
	}
	if (v.EntityType != "host" && v.EntityType != "identity") || v.Classification != "suggestion_only" {
		return Value{}, ErrInvalidValue
	}
	canonical, ok := fieldnorm.NormalizeEntityAliasText(v.RawText)
	if !ok || canonical != v.RawText || canonical != v.NormalizedText {
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
	if operation == "create" {
		if beforeValue != nil || after.DeletedAt != nil {
			return ErrInvalidValue
		}
		return nil
	}
	before, err := Decode(beforeValue)
	if err != nil || before.TargetID() != target {
		return ErrInvalidValue
	}
	switch operation {
	case "delete":
		if before.DeletedAt != nil || after.DeletedAt == nil {
			return ErrInvalidValue
		}
	case "rollback":
		if (before.DeletedAt == nil) == (after.DeletedAt == nil) {
			return ErrInvalidValue
		}
	default:
		return ErrInvalidValue
	}
	before.DeletedAt, after.DeletedAt = nil, nil
	if before != after {
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
