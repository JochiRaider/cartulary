package evidence

// Source mutation operations implement Evidence-owned row changes.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
)

type FieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type createParams struct {
	Values                 map[string]FieldValue
	InitialBlob            *initialBlobAssociation
	InitialBlobFinalized   bool
	InitialBlobWasSupplied bool
}

type initialBlobAssociation struct {
	ObjectBlobID uuid.UUID
	StorageRef   string
	SHA256Hex    *string
}

type lifecyclePatchChange struct {
	FieldKey string
	Text     *string
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return "evidence: invalid mutation request"
}

type LifecycleValidationError struct {
	FromStatus     string
	ToStatus       string
	ViolatedGuards []string
	ReasonCode     string
}

func (e *LifecycleValidationError) Error() string {
	return "evidence: illegal transition"
}

func (s *sourceMutationService) insertRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, params createParams, now time.Time) error {
	lifecycleState := "requested"
	if lifecycleValue, present := params.Values["evidence.lifecycle_state"]; present && lifecycleValue.Text != nil {
		lifecycleState = *lifecycleValue.Text
	}
	var requestedAt any
	if _, requestedPresent := params.Values["evidence.requested_at"]; requestedPresent {
		requestedAt = nullableTimestampValue(params.Values, "evidence.requested_at")
	} else if lifecycleState == "requested" {
		requestedAt = now
	}
	var objectBlobID any
	uploadState := "pending"
	var blobHash any
	storageRef := nullableTextValue(params.Values, "evidence.storage_ref")
	if params.InitialBlob != nil {
		objectBlobID = params.InitialBlob.ObjectBlobID
		uploadState = "available"
		blobHash = params.InitialBlob.SHA256Hex
		storageRef = params.InitialBlob.StorageRef
	}
	_, err := tx.Exec(ctx, `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, requested_at, received_at,
    storage_ref, collector_party_text, collector_party_id, source_party_text,
    source_party_id, object_blob_id, upload_state, blob_hash, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $15
)
`, recordID, incidentID,
		nullableTextValue(params.Values, "evidence.title"),
		lifecycleState,
		requestedAt,
		nullableTimestampValue(params.Values, "evidence.received_at"),
		storageRef,
		nullableTextValue(params.Values, "evidence.collector_party_text"),
		nullableUUIDValue(params.Values, "evidence.collector_party_id"),
		nullableTextValue(params.Values, "evidence.source_party_text"),
		nullableUUIDValue(params.Values, "evidence.source_party_id"),
		objectBlobID,
		uploadState,
		blobHash,
		now)
	if err != nil {
		return fmt.Errorf("insert evidence: %w", err)
	}
	return nil
}

func (s *sourceMutationService) validateLifecyclePatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changes []lifecyclePatchChange) error {
	var from string
	var objectBlobID sql.NullString
	var uploadState sql.NullString
	if err := tx.QueryRow(ctx, `
SELECT e.lifecycle_state, e.object_blob_id::text, b.upload_state
  FROM evidence e
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID).Scan(&from, &objectBlobID, &uploadState); err != nil {
		return err
	}
	to := from
	for _, change := range changes {
		if change.FieldKey == "evidence.lifecycle_state" && change.Text != nil {
			to = *change.Text
		}
	}
	if !evidencepolicy.LegalEvidenceTransition(from, to) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"evidence.lifecycle_state"}}
	}
	linkedBlobState := ""
	if uploadState.Valid {
		linkedBlobState = uploadState.String
	}
	if evidencepolicy.ViolatesEvidenceBlobBridge(to, objectBlobID.Valid, linkedBlobState) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: []string{"evidence.lifecycle_state", "object_blobs.upload_state"}}
	}
	return nil
}

func (s *sourceMutationService) applyDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string, value FieldValue, now time.Time) (bool, error) {
	if !strings.HasPrefix(fieldKey, "evidence.") {
		return false, &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	column := strings.TrimPrefix(fieldKey, "evidence.")
	dbValue := directDBValue(value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE evidence SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, column, column), recordID, dbValue, now)
	if err != nil {
		return false, fmt.Errorf("apply evidence direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *sourceMutationService) touchRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE evidence SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch evidence row: %w", err)
	}
	return nil
}

func directDBValue(value FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC()
	case value.UUID != nil:
		return *value.UUID
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func nullableTextValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return nil
}

func nullableUUIDValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return nil
}

func nullableTimestampValue(values map[string]FieldValue, field string) any {
	if value, ok := values[field]; ok && value.Timestamp != nil {
		return value.Timestamp.UTC()
	}
	return nil
}
