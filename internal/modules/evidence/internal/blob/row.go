// Package blob is the single Evidence-private authority for Object Blob
// lifecycle tuples and Blob row locks. Live paths and portable bundle
// admission decide REQ-02-189a coherence with the same predicate, and every
// Blob row lock is taken here so the tuple is verified under the lock.
package blob

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
)

// Row is the complete Object Blob source row without its storage key. It is
// both the portable bundle row and the live row as decoded under a lock.
type Row struct {
	ObjectBlobID        uuid.UUID  `json:"object_blob_id"`
	IncidentID          uuid.UUID  `json:"incident_id"`
	CreatedByUserID     uuid.UUID  `json:"created_by_user_id"`
	UploadState         string     `json:"upload_state"`
	ByteSize            int64      `json:"byte_size"`
	FilenameHint        *string    `json:"filename_hint"`
	ContentTypeHint     *string    `json:"content_type_hint"`
	ExpectedSHA256Hex   *string    `json:"expected_sha256_hex"`
	ObservedSize        *int64     `json:"observed_size"`
	ObservedContentType *string    `json:"observed_content_type"`
	ObservedSHA256Hex   *string    `json:"observed_sha256_hex"`
	TargetExpiresAt     time.Time  `json:"target_expires_at"`
	PendingExpiresAt    time.Time  `json:"pending_expires_at"`
	FinalizedAt         *time.Time `json:"finalized_at"`
	TerminalReason      *string    `json:"terminal_reason"`
	FailedAt            *time.Time `json:"failed_at"`
	CleanupDueAt        *time.Time `json:"cleanup_due_at"`
	CleanedUpAt         *time.Time `json:"cleaned_up_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

var rowFields = strings.Fields("object_blob_id incident_id created_by_user_id upload_state byte_size filename_hint content_type_hint expected_sha256_hex observed_size observed_content_type observed_sha256_hex target_expires_at pending_expires_at finalized_at terminal_reason failed_at cleanup_due_at cleaned_up_at created_at updated_at")

// ParseRow admits an exact row shape. Numbers must be json.Number, clocks
// canonical UTC text, and identities canonical non-nil UUIDs.
func ParseRow(raw map[string]any) (Row, bool) {
	if len(raw) != len(rowFields) {
		return Row{}, false
	}
	for _, field := range rowFields {
		value, present := raw[field]
		if !present {
			return Row{}, false
		}
		if value, ok := value.(string); ok && (!utf8.ValidString(value) || strings.ContainsRune(value, '\x00')) {
			return Row{}, false
		}
	}
	for _, field := range []string{"object_blob_id", "incident_id", "created_by_user_id"} {
		value, ok := raw[field].(string)
		id, err := uuid.Parse(value)
		if !ok || err != nil || id == uuid.Nil || id.String() != value {
			return Row{}, false
		}
	}
	if _, ok := raw["upload_state"].(string); !ok {
		return Row{}, false
	}
	for _, field := range []string{"byte_size", "observed_size"} {
		if field == "observed_size" && raw[field] == nil {
			continue
		}
		if _, ok := ParseSize(raw[field]); !ok {
			return Row{}, false
		}
	}
	for _, field := range []string{"expected_sha256_hex", "observed_sha256_hex"} {
		if raw[field] == nil {
			continue
		}
		if _, ok := ParseDigest(raw[field]); !ok {
			return Row{}, false
		}
	}
	for _, field := range []string{"filename_hint", "content_type_hint", "observed_content_type", "terminal_reason"} {
		if raw[field] == nil {
			continue
		}
		value, ok := raw[field].(string)
		if !ok {
			return Row{}, false
		}
		if (field == "filename_hint" || field == "content_type_hint") && (value == "" || strings.TrimSpace(value) != value) {
			return Row{}, false
		}
	}
	for _, field := range []string{"target_expires_at", "pending_expires_at", "created_at", "updated_at", "finalized_at", "failed_at", "cleanup_due_at", "cleaned_up_at"} {
		if raw[field] == nil && (field == "finalized_at" || field == "failed_at" || field == "cleanup_due_at" || field == "cleaned_up_at") {
			continue
		}
		if _, ok := ParseClock(raw[field]); !ok {
			return Row{}, false
		}
	}
	payload, err := json.Marshal(raw)
	var row Row
	if err != nil || json.Unmarshal(payload, &row) != nil {
		return Row{}, false
	}
	row.UTC()
	return row, true
}

// UTC normalizes every clock so rows compare independently of location.
func (r *Row) UTC() {
	r.TargetExpiresAt, r.PendingExpiresAt = r.TargetExpiresAt.UTC(), r.PendingExpiresAt.UTC()
	r.CreatedAt, r.UpdatedAt = r.CreatedAt.UTC(), r.UpdatedAt.UTC()
	for _, clock := range []**time.Time{&r.FinalizedAt, &r.FailedAt, &r.CleanupDueAt, &r.CleanedUpAt} {
		if *clock != nil {
			value := (**clock).UTC()
			*clock = &value
		}
	}
}

// Coherent is the REQ-02-189a lifecycle-tuple predicate. It also rejects an
// upload_state outside the closed vocabulary.
func (r Row) Coherent() bool {
	if r.UpdatedAt.Before(r.CreatedAt) || !r.TargetExpiresAt.Equal(r.CreatedAt.Add(time.Hour)) || !r.PendingExpiresAt.Equal(r.CreatedAt.Add(24*time.Hour)) {
		return false
	}
	for _, clock := range []*time.Time{r.FinalizedAt, r.FailedAt, r.CleanedUpAt} {
		if clock != nil && (clock.Before(r.CreatedAt) || clock.After(r.UpdatedAt)) {
			return false
		}
	}
	noObservation := r.ObservedSize == nil && r.ObservedContentType == nil && r.ObservedSHA256Hex == nil && r.FinalizedAt == nil
	noFailure := r.TerminalReason == nil && r.FailedAt == nil && r.CleanupDueAt == nil && r.CleanedUpAt == nil
	switch r.UploadState {
	case policy.BlobPending:
		return noObservation && noFailure
	case policy.BlobFailed:
		if !noObservation || r.TerminalReason == nil || r.FailedAt == nil || r.CleanupDueAt == nil || !r.CleanupDueAt.Equal(r.FailedAt.Add(policy.CleanupDelay)) ||
			(r.CleanedUpAt != nil && r.CleanedUpAt.Before(*r.CleanupDueAt)) {
			return false
		}
		switch *r.TerminalReason {
		case "pending_timeout":
			return !r.FailedAt.Before(r.PendingExpiresAt)
		case "declared_size_mismatch", "expected_sha256_mismatch":
			return true
		default:
			return false
		}
	case policy.BlobAvailable, policy.BlobQuarantined:
		return noFailure && r.FinalizedAt != nil && r.FinalizedAt.Before(r.PendingExpiresAt) && r.ObservedSize != nil && *r.ObservedSize == r.ByteSize && r.ObservedSHA256Hex != nil &&
			(r.ExpectedSHA256Hex == nil || *r.ExpectedSHA256Hex == *r.ObservedSHA256Hex)
	default:
		return false
	}
}

// Finalized reports the states that may be associated with Evidence.
func (r Row) Finalized() bool {
	return r.UploadState == policy.BlobAvailable || r.UploadState == policy.BlobQuarantined
}

// ParseSize admits a canonical non-negative integer json.Number.
func ParseSize(raw any) (int64, bool) {
	number, ok := raw.(json.Number)
	if !ok {
		return 0, false
	}
	value, err := number.Int64()
	return value, err == nil && value >= 0 && strconv.FormatInt(value, 10) == number.String()
}

// ParseDigest admits a lowercase hexadecimal SHA-256 digest.
func ParseDigest(raw any) (string, bool) {
	value, ok := raw.(string)
	if !ok || len(value) != 64 || strings.ToLower(value) != value {
		return "", false
	}
	_, err := hex.DecodeString(value)
	return value, err == nil
}

// ParseClock admits a canonical UTC microsecond clock, written either with a
// Z suffix or the +00:00 offset PostgreSQL emits.
func ParseClock(raw any) (time.Time, bool) {
	value, ok := raw.(string)
	clock, err := time.Parse(time.RFC3339Nano, value)
	_, offset := clock.Zone()
	valid := ok && err == nil && offset == 0 && clock.Year() >= 1 && clock.Year() <= 9999 && clock.Nanosecond()%1000 == 0 &&
		(value == clock.UTC().Format(time.RFC3339Nano) || value == strings.TrimSuffix(clock.UTC().Format(time.RFC3339Nano), "Z")+"+00:00")
	return clock.UTC(), valid
}
