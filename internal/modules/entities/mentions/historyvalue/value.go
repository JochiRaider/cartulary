// Package historyvalue owns the complete retained Entity Mention value.
package historyvalue

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ErrInvalidValue = errors.New("invalid retained entity mention")

type Value struct {
	MentionID, SourceRecordID                                uuid.UUID
	EntityType, SourceFieldKey, OriginKind                   string
	OriginLocator, RawText, NormalizedText, ResolutionStatus string
	RowVersion                                               int64
	Ordinal                                                  int32
	CreatedByUserID                                          uuid.UUID
	CreatedAt                                                time.Time
	ResolvedRecordID, ResolvedByUserID                       *uuid.UUID
	ResolvedAt                                               *time.Time
	ResolutionMethod                                         *string
}

func (v Value) Map() map[string]any {
	var record, actor, at, method any
	if v.ResolvedRecordID != nil {
		record = v.ResolvedRecordID.String()
	}
	if v.ResolvedByUserID != nil {
		actor = v.ResolvedByUserID.String()
	}
	if v.ResolvedAt != nil {
		at = v.ResolvedAt.UTC().Format(time.RFC3339Nano)
	}
	if v.ResolutionMethod != nil {
		method = *v.ResolutionMethod
	}
	return map[string]any{
		"entity_mention_id": v.MentionID.String(), "source_record_id": v.SourceRecordID.String(),
		"entity_type": v.EntityType, "source_field_key": v.SourceFieldKey, "origin_kind": v.OriginKind,
		"origin_locator": v.OriginLocator, "raw_text": v.RawText, "normalized_text": v.NormalizedText,
		"resolution_status": v.ResolutionStatus, "row_version": v.RowVersion, "ordinal": v.Ordinal,
		"created_by_user_id": v.CreatedByUserID.String(), "created_at": v.CreatedAt.UTC().Format(time.RFC3339Nano),
		"resolved_record_id": record, "resolved_by_user_id": actor, "resolved_at": at, "resolution_method": method,
	}
}

func Decode(input map[string]any) (Value, error) {
	fields := []string{"entity_mention_id", "source_record_id", "entity_type", "source_field_key", "origin_kind", "origin_locator", "raw_text", "normalized_text", "resolution_status", "row_version", "ordinal", "created_by_user_id", "created_at", "resolved_record_id", "resolved_by_user_id", "resolved_at", "resolution_method"}
	if len(input) != len(fields) {
		return Value{}, ErrInvalidValue
	}
	for _, field := range fields {
		if _, present := input[field]; !present {
			return Value{}, ErrInvalidValue
		}
	}
	v := Value{}
	for key, out := range map[string]*uuid.UUID{"entity_mention_id": &v.MentionID, "source_record_id": &v.SourceRecordID, "created_by_user_id": &v.CreatedByUserID} {
		id, ok := canonicalUUID(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = id
	}
	for key, out := range map[string]*string{"entity_type": &v.EntityType, "source_field_key": &v.SourceFieldKey, "origin_kind": &v.OriginKind, "origin_locator": &v.OriginLocator, "raw_text": &v.RawText, "normalized_text": &v.NormalizedText, "resolution_status": &v.ResolutionStatus} {
		text, ok := nonblankText(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = text
	}
	if v.EntityType != "host" && v.EntityType != "identity" {
		return Value{}, ErrInvalidValue
	}
	switch v.OriginKind {
	case "manual_entry", "clipboard_paste", "csv_import", "xlsx_import", "api_import", "extraction", "system":
	default:
		return Value{}, ErrInvalidValue
	}
	var ok bool
	v.RowVersion, ok = positiveInteger(input["row_version"])
	if !ok {
		return Value{}, ErrInvalidValue
	}
	ordinal, ok := positiveInteger(input["ordinal"])
	if !ok || ordinal > math.MaxInt32 {
		return Value{}, ErrInvalidValue
	}
	v.Ordinal = int32(ordinal)
	v.CreatedAt, ok = canonicalTime(input["created_at"])
	if !ok {
		return Value{}, ErrInvalidValue
	}
	for key, out := range map[string]**uuid.UUID{"resolved_record_id": &v.ResolvedRecordID, "resolved_by_user_id": &v.ResolvedByUserID} {
		if input[key] == nil {
			continue
		}
		id, ok := canonicalUUID(input[key])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		*out = &id
	}
	if input["resolved_at"] != nil {
		at, ok := canonicalTime(input["resolved_at"])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		v.ResolvedAt = &at
	}
	if input["resolution_method"] != nil {
		method, ok := nonblankText(input["resolution_method"])
		if !ok {
			return Value{}, ErrInvalidValue
		}
		v.ResolutionMethod = &method
	}
	switch v.ResolutionStatus {
	case "resolved":
		if v.ResolvedRecordID == nil || v.ResolvedByUserID == nil || v.ResolvedAt == nil || v.ResolutionMethod == nil {
			return Value{}, ErrInvalidValue
		}
	case "unresolved", "dismissed":
		if v.ResolvedRecordID != nil || v.ResolvedByUserID != nil || v.ResolvedAt != nil || v.ResolutionMethod != nil {
			return Value{}, ErrInvalidValue
		}
	default:
		return Value{}, ErrInvalidValue
	}
	return v, nil
}

func ValidateMutation(targetID, operation string, beforeValue, afterValue map[string]any) error {
	if operation != "patch" && operation != "rollback" {
		return ErrInvalidValue
	}
	before, err := Decode(beforeValue)
	if err != nil {
		return err
	}
	after, err := Decode(afterValue)
	if err != nil {
		return err
	}
	if targetID != before.MentionID.String() || targetID != after.MentionID.String() ||
		before.SourceRecordID != after.SourceRecordID || before.EntityType != after.EntityType ||
		before.SourceFieldKey != after.SourceFieldKey || before.OriginKind != after.OriginKind ||
		before.OriginLocator != after.OriginLocator || before.RawText != after.RawText || before.NormalizedText != after.NormalizedText ||
		before.Ordinal != after.Ordinal || before.CreatedByUserID != after.CreatedByUserID || !before.CreatedAt.Equal(after.CreatedAt) ||
		before.RowVersion == math.MaxInt64 || after.RowVersion != before.RowVersion+1 {
		return ErrInvalidValue
	}
	if before.ResolutionStatus == after.ResolutionStatus && equalPointer(before.ResolvedRecordID, after.ResolvedRecordID) &&
		equalPointer(before.ResolvedByUserID, after.ResolvedByUserID) && equalPointer(before.ResolvedAt, after.ResolvedAt) && equalPointer(before.ResolutionMethod, after.ResolutionMethod) {
		return ErrInvalidValue
	}
	return nil
}

func equalPointer[T comparable](left, right *T) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func canonicalUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(text)
	return id, err == nil && id != uuid.Nil && id.String() == text
}

func nonblankText(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok && utf8.ValidString(text) && !strings.ContainsRune(text, 0) && strings.TrimSpace(text) != ""
}

func positiveInteger(value any) (int64, bool) {
	var n int64
	switch value := value.(type) {
	case int:
		n = int64(value)
	case int32:
		n = int64(value)
	case int64:
		n = value
	case json.Number:
		parsed, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil || strconv.FormatInt(parsed, 10) != value.String() {
			return 0, false
		}
		n = parsed
	default:
		return 0, false
	}
	return n, n > 0
}

func canonicalTime(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	at, err := time.Parse(time.RFC3339Nano, text)
	return at, err == nil && at.Year() >= 1 && at.Year() <= 9999 && at.Nanosecond()%1000 == 0 && at.UTC().Format(time.RFC3339Nano) == text
}
