package evidence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WorkbookFieldValue struct {
	Text      *string
	Timestamp *time.Time
	UUID      *uuid.UUID
	Number    *int64
	Bool      *bool
}

type WorkbookCreateParams struct {
	Values map[string]WorkbookFieldValue
}

type WorkbookLifecyclePatchChange struct {
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

func ValidLifecycleState(value string) bool {
	switch value {
	case "requested", "pending_receipt", "received", "available", "quarantined", "released":
		return true
	default:
		return false
	}
}

func (s *Store) InsertWorkbookRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, params WorkbookCreateParams, now time.Time) error {
	lifecycleState := nullableTextValue(params.Values, "evidence.lifecycle_state")
	if lifecycleState == nil {
		lifecycleState = "requested"
	}
	requestedAt := nullableTimestampValue(params.Values, "evidence.requested_at")
	if requestedAt == nil && lifecycleState == "requested" {
		requestedAt = now
	}
	_, err := tx.Exec(ctx, `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, requested_at, received_at,
    storage_ref, collector_party_text, collector_party_id, source_party_text,
    source_party_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $12
)
`, recordID, incidentID,
		nullableTextValue(params.Values, "evidence.title"),
		lifecycleState,
		requestedAt,
		nullableTimestampValue(params.Values, "evidence.received_at"),
		nullableTextValue(params.Values, "evidence.storage_ref"),
		nullableTextValue(params.Values, "evidence.collector_party_text"),
		nullableUUIDValue(params.Values, "evidence.collector_party_id"),
		nullableTextValue(params.Values, "evidence.source_party_text"),
		nullableUUIDValue(params.Values, "evidence.source_party_id"),
		now)
	if err != nil {
		return fmt.Errorf("insert evidence: %w", err)
	}
	return nil
}

func (s *Store) ValidateWorkbookLifecyclePatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changes []WorkbookLifecyclePatchChange) error {
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
	if to != from && !legalLifecycleTransition(from, to) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"evidence.lifecycle_state"}}
	}
	linkedBlobState := ""
	if uploadState.Valid {
		linkedBlobState = uploadState.String
	}
	if violatesBlobBridge(to, objectBlobID.Valid, linkedBlobState) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: []string{"evidence.lifecycle_state", "object_blobs.upload_state"}}
	}
	return nil
}

func (s *Store) ApplyWorkbookDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string, value WorkbookFieldValue, now time.Time) (bool, error) {
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

func (s *Store) TouchWorkbookRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `UPDATE evidence SET updated_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
		return fmt.Errorf("touch evidence row: %w", err)
	}
	return nil
}

func legalLifecycleTransition(from string, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "requested":
		return to == "pending_receipt" || to == "received" || to == "available"
	case "pending_receipt":
		return to == "requested" || to == "received" || to == "available"
	case "received":
		return to == "pending_receipt" || to == "available" || to == "quarantined"
	case "available":
		return to == "received" || to == "quarantined" || to == "released"
	case "quarantined":
		return to == "received" || to == "available"
	case "released":
		return to == "available" || to == "quarantined"
	default:
		return false
	}
}

func violatesBlobBridge(lifecycleState string, hasBlob bool, uploadState string) bool {
	switch lifecycleState {
	case "available", "released":
		return !hasBlob || uploadState != "available"
	case "quarantined":
		return hasBlob && uploadState != "quarantined"
	default:
		return false
	}
}

func directDBValue(value WorkbookFieldValue) any {
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

func nullableTextValue(values map[string]WorkbookFieldValue, field string) any {
	if value, ok := values[field]; ok && value.Text != nil {
		return *value.Text
	}
	return nil
}

func nullableUUIDValue(values map[string]WorkbookFieldValue, field string) any {
	if value, ok := values[field]; ok && value.UUID != nil {
		return *value.UUID
	}
	return nil
}

func nullableTimestampValue(values map[string]WorkbookFieldValue, field string) any {
	if value, ok := values[field]; ok && value.Timestamp != nil {
		return value.Timestamp.UTC()
	}
	return nil
}
