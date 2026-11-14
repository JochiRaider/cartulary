package evidence

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const portableBlobPath = "data/object_blobs.ndjson"

type portableBlobRow struct {
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

func parsePortableBlobRow(raw map[string]any) (portableBlobRow, bool) {
	fields := strings.Fields("object_blob_id incident_id created_by_user_id upload_state byte_size filename_hint content_type_hint expected_sha256_hex observed_size observed_content_type observed_sha256_hex target_expires_at pending_expires_at finalized_at terminal_reason failed_at cleanup_due_at cleaned_up_at created_at updated_at")
	if len(raw) != len(fields) {
		return portableBlobRow{}, false
	}
	for _, field := range fields {
		value, present := raw[field]
		if !present {
			return portableBlobRow{}, false
		}
		if value, ok := value.(string); ok && (!utf8.ValidString(value) || strings.ContainsRune(value, '\x00')) {
			return portableBlobRow{}, false
		}
	}
	for _, field := range []string{"object_blob_id", "incident_id", "created_by_user_id"} {
		value, ok := raw[field].(string)
		id, err := uuid.Parse(value)
		if !ok || err != nil || id == uuid.Nil || id.String() != value {
			return portableBlobRow{}, false
		}
	}
	if _, ok := raw["upload_state"].(string); !ok {
		return portableBlobRow{}, false
	}
	for _, field := range []string{"byte_size", "observed_size"} {
		if field == "observed_size" && raw[field] == nil {
			continue
		}
		if _, ok := portableBlobSize(raw[field]); !ok {
			return portableBlobRow{}, false
		}
	}
	for _, field := range []string{"expected_sha256_hex", "observed_sha256_hex"} {
		if raw[field] == nil {
			continue
		}
		if _, ok := portableBlobDigest(raw[field]); !ok {
			return portableBlobRow{}, false
		}
	}
	for _, field := range []string{"filename_hint", "content_type_hint", "observed_content_type", "terminal_reason"} {
		if raw[field] == nil {
			continue
		}
		value, ok := raw[field].(string)
		if !ok {
			return portableBlobRow{}, false
		}
		if (field == "filename_hint" || field == "content_type_hint") && (value == "" || strings.TrimSpace(value) != value) {
			return portableBlobRow{}, false
		}
	}
	for _, field := range []string{"target_expires_at", "pending_expires_at", "created_at", "updated_at", "finalized_at", "failed_at", "cleanup_due_at", "cleaned_up_at"} {
		if raw[field] == nil && (field == "finalized_at" || field == "failed_at" || field == "cleanup_due_at" || field == "cleaned_up_at") {
			continue
		}
		if _, ok := portableSourceClock(raw[field]); !ok {
			return portableBlobRow{}, false
		}
	}
	payload, err := json.Marshal(raw)
	var row portableBlobRow
	if err != nil || json.Unmarshal(payload, &row) != nil {
		return portableBlobRow{}, false
	}
	row.utc()
	return row, true
}

func (r *portableBlobRow) utc() {
	r.TargetExpiresAt, r.PendingExpiresAt = r.TargetExpiresAt.UTC(), r.PendingExpiresAt.UTC()
	r.CreatedAt, r.UpdatedAt = r.CreatedAt.UTC(), r.UpdatedAt.UTC()
	for _, clock := range []**time.Time{&r.FinalizedAt, &r.FailedAt, &r.CleanupDueAt, &r.CleanedUpAt} {
		if *clock != nil {
			value := (**clock).UTC()
			*clock = &value
		}
	}
}

func (r portableBlobRow) lifecycleLegal() bool {
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
	case "pending":
		return noObservation && noFailure
	case "failed":
		if !noObservation || r.TerminalReason == nil || r.FailedAt == nil || r.CleanupDueAt == nil || !r.CleanupDueAt.Equal(r.FailedAt.Add(45*time.Minute)) ||
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
	case "available", "quarantined":
		return noFailure && r.FinalizedAt != nil && r.FinalizedAt.Before(r.PendingExpiresAt) && r.ObservedSize != nil && *r.ObservedSize == r.ByteSize && r.ObservedSHA256Hex != nil &&
			(r.ExpectedSHA256Hex == nil || *r.ExpectedSHA256Hex == *r.ObservedSHA256Hex)
	default:
		return false
	}
}
