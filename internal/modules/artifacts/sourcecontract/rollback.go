package sourcecontract

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
)

type StorageMapping struct {
	Table  string
	Column string
}

// WritableDirectStorageMappings is the exact reviewed identifier allowlist for
// artifact scalar mutation. Callers receive a copy.
func WritableDirectStorageMappings() map[string]StorageMapping {
	result := make(map[string]StorageMapping, 36)
	for _, surface := range contractartifacts.SourceCatalog {
		for _, field := range surface.DirectFields {
			result[field.FieldKey] = StorageMapping{Table: field.Table, Column: field.Column}
		}
	}
	return result
}

// ConflictFieldSourceKeys declares the source-owned scalar projection required
// to reconstruct optimistic-concurrency facts from canonical snapshots.
func ConflictFieldSourceKeys() map[string]string {
	result := make(map[string]string)
	for fieldKey, storage := range WritableDirectStorageMappings() {
		result[fieldKey] = storage.Column
	}
	return result
}

// ExtractRollbackSource accepts only the source member of a canonical retained
// snapshot envelope.
func ExtractRollbackSource(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	return nil, false
}

// ValidRollbackSource rejects invalid typed values and subtype invariants
// before a rollback adapter executes source SQL.
func ValidRollbackSource(source map[string]any) bool {
	for _, key := range []string{
		"title", "body", "comm_id", "comm_type", "audience", "channel_or_meeting", "summary",
		"privilege_tag", "handoff_id", "current_state_summary", "next_checks", "status_review_id",
		"active_risks_summary", "lesson_id", "closure_state", "kind", "statement", "state",
		"query_id", "platform", "purpose", "query_text", "keyword_id", "pattern", "reason", "match_mode",
	} {
		if raw, present := source[key]; present && raw != nil {
			if _, valid := raw.(string); !valid {
				return false
			}
		}
	}
	for _, key := range []string{
		"comm_id", "comm_type", "audience", "channel_or_meeting", "summary", "handoff_id",
		"current_state_summary", "status_review_id", "lesson_id", "kind", "statement", "state",
		"query_id", "platform", "purpose", "query_text", "keyword_id", "pattern", "reason", "match_mode",
	} {
		if raw, present := source[key]; present && !nonEmptyText(raw) {
			return false
		}
	}
	if raw, present := source["comm_type"]; present && !oneOfText(raw, "meeting", "notification", "approval", "briefing", "handoff") {
		return false
	}
	if raw, present := source["closure_state"]; present && !oneOfText(raw, "open", "closed") {
		return false
	}
	if raw, present := source["kind"]; present && !oneOfText(raw, "finding", "hypothesis") {
		return false
	}
	if raw, present := source["state"]; present && !oneOfText(raw, "open", "closed") {
		return false
	}
	if raw, present := source["match_mode"]; present && !oneOfText(raw, "literal", "regex") {
		return false
	}
	if raw, present := source["confidence_score"]; present && raw != nil {
		score, valid := integerValue(raw)
		if !valid || score < 0 || score > 100 {
			return false
		}
	}
	for key, kind := range map[string]string{
		"timestamp_utc": "time", "next_report_at": "time", "acknowledged_at": "time", "closed_at": "time",
		"outgoing_owner_user_id": "uuid", "incoming_owner_user_id": "uuid",
		"review_owner_user_id": "uuid", "owner_user_id": "uuid", "case_sensitive": "bool",
	} {
		if raw, present := source[key]; present && raw != nil && !validKind(raw, kind) {
			return false
		}
	}
	for _, key := range []string{"timestamp_utc", "incoming_owner_user_id", "review_owner_user_id", "owner_user_id", "case_sensitive"} {
		if raw, present := source[key]; present && raw == nil {
			return false
		}
	}
	return true
}

func validKind(value any, kind string) bool {
	switch kind {
	case "time":
		switch typed := value.(type) {
		case time.Time:
			return true
		case string:
			_, err := time.Parse(time.RFC3339Nano, typed)
			return err == nil
		}
	case "uuid":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return false
		}
		_, err := uuid.Parse(text)
		return err == nil
	case "bool":
		_, ok := value.(bool)
		return ok
	}
	return false
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func integerValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), math.Trunc(typed) == typed
	default:
		return 0, false
	}
}

func oneOfText(value any, allowed ...string) bool {
	text, valid := value.(string)
	if !valid {
		return false
	}
	for _, candidate := range allowed {
		if text == candidate {
			return true
		}
	}
	return false
}

func nonEmptyText(value any) bool {
	text, valid := value.(string)
	return valid && strings.TrimSpace(text) != ""
}
