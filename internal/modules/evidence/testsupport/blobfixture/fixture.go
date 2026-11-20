// Package blobfixture is the single Object Blob fixture builder for Evidence
// and cross-module tests. Seed and Move write only REQ-02-189a-coherent
// lifecycle tuples and self-check every row with the production predicate.
// BreachTupleInvariant is the only writer of a deliberately incoherent row.
package blobfixture

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
)

// DB is the narrow SQL surface the fixture uses. pgx pools, transactions and
// the Evidence test stores satisfy it.
type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

const (
	Pending     = "pending"
	Available   = "available"
	Quarantined = "quarantined"
	Failed      = "failed"
)

// Spec describes a Blob slot. Zero values select coherent defaults.
type Spec struct {
	// ObjectBlobID defaults to a new identity.
	ObjectBlobID    uuid.UUID
	IncidentID      uuid.UUID
	CreatedByUserID uuid.UUID
	// State is pending, available, quarantined or failed.
	State string
	// StorageKey defaults to a unique fixture key.
	StorageKey        string
	ByteSize          int64
	FilenameHint      *string
	ContentTypeHint   *string
	ExpectedSHA256Hex *string
	// ObservedSHA256Hex applies to finalized slots; it defaults to the
	// expected digest, else 64 zeros.
	ObservedSHA256Hex *string
	// ObservedContentType applies to finalized slots and may stay nil.
	ObservedContentType *string
	// CreatedAt defaults to now. Both expiries derive from it, so an expired
	// pending slot is created more than 24 hours in the past.
	CreatedAt time.Time
	// TerminalReason applies to failed slots and defaults to
	// declared_size_mismatch. pending_timeout fails at pending expiry.
	TerminalReason string
}

// Seed inserts one coherent Blob row and returns its identity.
func Seed(t testing.TB, db DB, spec Spec) uuid.UUID {
	t.Helper()
	if spec.ObjectBlobID == uuid.Nil {
		spec.ObjectBlobID = uuid.New()
	}
	if spec.StorageKey == "" {
		spec.StorageKey = "evidence-fixture/" + spec.ObjectBlobID.String()
	}
	createdAt := spec.CreatedAt.UTC().Truncate(time.Microsecond)
	if spec.CreatedAt.IsZero() {
		createdAt = time.Now().UTC().Truncate(time.Microsecond)
	}
	targetExpiresAt, pendingExpiresAt := createdAt.Add(time.Hour), createdAt.Add(24*time.Hour)
	updatedAt := createdAt
	var observedSize *int64
	var observedSHA, observedContentType, terminalReason *string
	var finalizedAt, failedAt, cleanupDueAt *time.Time
	switch spec.State {
	case Pending:
	case Available, Quarantined:
		size := spec.ByteSize
		observedSize = &size
		observedSHA = spec.ObservedSHA256Hex
		if observedSHA == nil {
			observedSHA = spec.ExpectedSHA256Hex
		}
		if observedSHA == nil {
			zero := strings.Repeat("0", 64)
			observedSHA = &zero
		}
		observedContentType = spec.ObservedContentType
		finalizedAt = &createdAt
	case Failed:
		reason := spec.TerminalReason
		if reason == "" {
			reason = "declared_size_mismatch"
		}
		terminalReason = &reason
		failed := createdAt
		if reason == "pending_timeout" {
			failed = pendingExpiresAt
		}
		due := failed.Add(45 * time.Minute)
		failedAt, cleanupDueAt, updatedAt = &failed, &due, failed
	default:
		t.Fatalf("blob fixture state %q is outside the closed vocabulary", spec.State)
	}
	if _, err := db.Exec(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, expected_sha256_hex,
    observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, cleaned_up_at,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12,
    $13, $14, $15,
    $16, $17, $18, NULL,
    $19, $20
)
`, spec.ObjectBlobID, spec.IncidentID, spec.CreatedByUserID, spec.StorageKey, spec.State,
		spec.ByteSize, spec.FilenameHint, spec.ContentTypeHint, spec.ExpectedSHA256Hex,
		observedSize, observedContentType, observedSHA,
		targetExpiresAt, pendingExpiresAt, finalizedAt,
		terminalReason, failedAt, cleanupDueAt,
		createdAt, updatedAt); err != nil {
		t.Fatalf("seed coherent object blob: %v", err)
	}
	RequireCoherent(t, db, spec.ObjectBlobID)
	return spec.ObjectBlobID
}

// Move rewrites an existing Blob into the coherent tuple of state, keeping
// its identity, incident, size, digests and clocks where the target allows.
// It is a fixture shortcut, not a legal lifecycle transition.
func Move(t testing.TB, db DB, objectBlobID uuid.UUID, state string) {
	t.Helper()
	switch state {
	case Pending, Available, Quarantined, Failed:
	default:
		t.Fatalf("blob fixture state %q is outside the closed vocabulary", state)
	}
	tag, err := db.Exec(context.Background(), `
UPDATE object_blobs
   SET upload_state = $2::text,
       observed_size = CASE WHEN $2::text IN ('available', 'quarantined') THEN byte_size END,
       observed_sha256_hex = CASE WHEN $2::text IN ('available', 'quarantined')
                                  THEN COALESCE(observed_sha256_hex, expected_sha256_hex, repeat('0', 64)) END,
       observed_content_type = CASE WHEN $2::text IN ('available', 'quarantined') THEN observed_content_type END,
       finalized_at = CASE WHEN $2::text IN ('available', 'quarantined') THEN COALESCE(finalized_at, created_at) END,
       terminal_reason = CASE WHEN $2::text = 'failed' THEN COALESCE(terminal_reason, 'declared_size_mismatch') END,
       failed_at = CASE WHEN $2::text = 'failed' THEN COALESCE(failed_at, GREATEST(updated_at, created_at)) END,
       cleanup_due_at = CASE WHEN $2::text = 'failed'
                             THEN COALESCE(failed_at, GREATEST(updated_at, created_at)) + interval '45 minutes' END,
       cleaned_up_at = NULL,
       updated_at = GREATEST(updated_at, created_at)
 WHERE object_blob_id = $1
`, objectBlobID, state)
	if err != nil {
		t.Fatalf("move object blob fixture to %s: %v", state, err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("move object blob fixture %s: no row", objectBlobID)
	}
	RequireCoherent(t, db, objectBlobID)
}

// RequireCoherent fails the test unless the live row satisfies the
// production REQ-02-189a predicate.
func RequireCoherent(t testing.TB, db DB, objectBlobID uuid.UUID) {
	t.Helper()
	if err := verify(db, objectBlobID); err != nil {
		t.Fatalf("object blob fixture %s violates the lifecycle tuple: %v", objectBlobID, err)
	}
}

// ClaimForCleanup records an unexpired durable cleanup claim on the Blob.
func ClaimForCleanup(t testing.TB, db DB, objectBlobID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO evidence_blob_cleanup_claims (
    object_blob_id, claim_token, claim_state, attempt_count,
    claimed_at, claim_expires_at, created_at, updated_at
) VALUES ($1, $2, 'claimed', 1, now(), now() + interval '5 minutes', now(), now())
`, objectBlobID, uuid.New()); err != nil {
		t.Fatalf("claim object blob fixture for cleanup: %v", err)
	}
}

// BreachTupleInvariant is the invariant-breach fixture: it moves the Blob's
// update clock before its creation clock, which no legal write produces and
// the cleanup-claim guard does not block, leaving every other field intact.
// It asserts that the production predicate now rejects the row.
func BreachTupleInvariant(t testing.TB, db DB, objectBlobID uuid.UUID) {
	t.Helper()
	tag, err := db.Exec(context.Background(), `
UPDATE object_blobs
   SET updated_at = created_at - interval '1 second'
 WHERE object_blob_id = $1
`, objectBlobID)
	if err != nil {
		t.Fatalf("breach object blob tuple invariant: %v", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("breach object blob tuple invariant %s: no row", objectBlobID)
	}
	if err := verify(db, objectBlobID); err == nil {
		t.Fatalf("object blob %s still satisfies the lifecycle tuple after the breach", objectBlobID)
	}
}

func verify(db DB, objectBlobID uuid.UUID) error {
	var payload []byte
	if err := db.QueryRow(context.Background(), `
SELECT (to_jsonb(b) - 'storage_key')::text
  FROM object_blobs b
 WHERE b.object_blob_id = $1
`, objectBlobID).Scan(&payload); err != nil {
		return err
	}
	_, err := evidenceblob.VerifyPayload(payload)
	return err
}
