package valuecodec

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OptionalString returns the canonical JSON-facing representation of a
// nullable string.
func OptionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

// OptionalUUID returns the canonical JSON-facing representation of a
// nullable UUID.
func OptionalUUID(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

// Timestamp formats an instant in UTC with nanosecond precision when present.
func Timestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

// OptionalTimestamp returns the canonical JSON-facing representation of a
// nullable instant.
func OptionalTimestamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return Timestamp(*value)
}

// OptionalDate returns the UTC calendar date for a nullable instant.
func OptionalDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.DateOnly)
}

// Collection returns the canonical collection_value_v1 shape. Nil and empty
// item slices intentionally have the same non-null, empty-array representation.
func Collection(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

// CanonicalJSONSHA256 hashes the encoding/json canonical representation used
// by Timeline request fingerprints and deterministic local identifiers.
func CanonicalJSONSHA256(value any) []byte {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}
