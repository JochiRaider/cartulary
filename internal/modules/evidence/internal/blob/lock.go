package blob

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrTupleViolation is a REQ-02-189a storage invariant failure: a loaded
	// Blob violates the lifecycle tuple or its upload_state is outside the
	// closed vocabulary. Every route maps it to internal_error.
	ErrTupleViolation = errors.New("evidence: blob lifecycle tuple violated")
	// ErrNotFound reports that no Blob row exists for the identity.
	ErrNotFound = errors.New("evidence: blob not found")
	// ErrContended reports that a non-waiting lock found the row held. It
	// carries no Blob identity or database text.
	ErrContended = errors.New("evidence: blob row lock unavailable")
)

// LockMode selects the strength of a Blob row lock and how it treats a
// competing holder (REQ-02-190a).
type LockMode uint8

const (
	// UpdateWait takes FOR UPDATE and waits; Blob state writers use it.
	UpdateWait LockMode = iota
	// UpdateNoWait takes FOR UPDATE and fails with ErrContended at once.
	UpdateNoWait
	// ShareWait takes FOR SHARE and waits; an ordinary write that only
	// decides a bridge against the Blob (patch) uses it.
	ShareWait
	// ShareNoWait takes FOR SHARE and fails with ErrContended at once;
	// destructive bridge deciders (restore, rollback; REQ-01-104) use it.
	ShareNoWait
)

// Locked is a Blob row held under a row lock through commit, with its
// lifecycle tuple verified.
type Locked struct {
	Row
	StorageKey string
}

// LockTx locks one Blob row in mode and verifies its lifecycle tuple under the
// lock. A violation returns ErrTupleViolation and the caller must commit
// nothing. Callers that also lock Evidence take this lock last (REQ-02-190a).
func LockTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID, mode LockMode) (Locked, error) {
	var row pgx.Row
	switch mode {
	case UpdateWait:
		row = tx.QueryRow(ctx, `
SELECT (to_jsonb(b) - 'storage_key')::text, b.storage_key
  FROM object_blobs b
 WHERE b.object_blob_id = $1
 FOR UPDATE OF b
`, objectBlobID)
	case UpdateNoWait:
		row = tx.QueryRow(ctx, `
SELECT (to_jsonb(b) - 'storage_key')::text, b.storage_key
  FROM object_blobs b
 WHERE b.object_blob_id = $1
 FOR UPDATE OF b NOWAIT
`, objectBlobID)
	case ShareWait:
		row = tx.QueryRow(ctx, `
SELECT (to_jsonb(b) - 'storage_key')::text, b.storage_key
  FROM object_blobs b
 WHERE b.object_blob_id = $1
 FOR SHARE OF b
`, objectBlobID)
	case ShareNoWait:
		row = tx.QueryRow(ctx, `
SELECT (to_jsonb(b) - 'storage_key')::text, b.storage_key
  FROM object_blobs b
 WHERE b.object_blob_id = $1
 FOR SHARE OF b NOWAIT
`, objectBlobID)
	default:
		return Locked{}, fmt.Errorf("evidence: unknown blob lock mode %d", mode)
	}
	var payload []byte
	var locked Locked
	if err := row.Scan(&payload, &locked.StorageKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Locked{}, ErrNotFound
		}
		if lockUnavailable(err) {
			return Locked{}, ErrContended
		}
		return Locked{}, err
	}
	decoded, err := decodeVerified(objectBlobID, payload)
	if err != nil {
		return Locked{}, err
	}
	locked.Row = decoded
	return locked, nil
}

// VerifyPayload applies the live tuple predicate to one to_jsonb(object_blobs)
// payload without its storage key. Test fixtures use it to self-check rows.
func VerifyPayload(payload []byte) (Row, error) {
	return decodeVerified(uuid.Nil, payload)
}

// decodeVerified decodes numbers exactly as portable rows are decoded
// (json.Number), then applies the shared tuple predicate.
func decodeVerified(objectBlobID uuid.UUID, payload []byte) (Row, error) {
	var raw map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return Row{}, fmt.Errorf("%w: %s", ErrTupleViolation, objectBlobID)
	}
	row, ok := ParseRow(raw)
	if ok && objectBlobID != uuid.Nil && row.ObjectBlobID != objectBlobID {
		ok = false
	}
	if !ok || !row.Coherent() {
		if objectBlobID == uuid.Nil {
			objectBlobID = row.ObjectBlobID
		}
		return Row{}, fmt.Errorf("%w: %s", ErrTupleViolation, objectBlobID)
	}
	return row, nil
}

func lockUnavailable(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "55P03"
}

// ExpiredPending is the outcome of locking every pending Blob whose pending
// expiry has passed.
type ExpiredPending struct {
	// Coherent rows may take the pending-timeout transition.
	Coherent []Row
	// Violating rows fail closed: they stay locked and unchanged for operator
	// recovery and never block the coherent rows.
	Violating []uuid.UUID
}

// LockExpiredPendingTx locks, waiting, every pending Blob expired at now and
// verifies each lifecycle tuple under the lock.
func LockExpiredPendingTx(ctx context.Context, tx pgx.Tx, now time.Time) (ExpiredPending, error) {
	rows, err := tx.Query(ctx, `
SELECT b.object_blob_id, (to_jsonb(b) - 'storage_key')::text
  FROM object_blobs b
 WHERE b.upload_state = 'pending'
   AND b.pending_expires_at <= $1
 ORDER BY b.object_blob_id
 FOR UPDATE OF b
`, now)
	if err != nil {
		return ExpiredPending{}, err
	}
	defer rows.Close()
	var result ExpiredPending
	for rows.Next() {
		var objectBlobID uuid.UUID
		var payload []byte
		if err := rows.Scan(&objectBlobID, &payload); err != nil {
			return ExpiredPending{}, err
		}
		row, err := decodeVerified(objectBlobID, payload)
		if errors.Is(err, ErrTupleViolation) {
			result.Violating = append(result.Violating, objectBlobID)
			continue
		}
		if err != nil {
			return ExpiredPending{}, err
		}
		result.Coherent = append(result.Coherent, row)
	}
	if err := rows.Err(); err != nil {
		return ExpiredPending{}, err
	}
	return result, nil
}

// CleanupKey is the (cleanup_due_at, object_blob_id) cleanup scan position.
type CleanupKey struct {
	CleanupDueAt time.Time
	ObjectBlobID uuid.UUID
}

// CleanupCandidate is a due failed, unattached Blob locked for a cleanup
// claim. Coherent is false when the lifecycle tuple is violated; such a Blob
// must stay unclaimed for operator recovery.
type CleanupCandidate struct {
	ObjectBlobID uuid.UUID
	StorageKey   string
	AttemptCount int
	Coherent     bool
}

// LockCleanupCandidatesTx locks with SKIP LOCKED up to limit due cleanup
// candidates after the scan position and verifies each tuple. It returns the
// last key only when the page was full, so the caller stops once the due set
// is exhausted and continues past violating rows otherwise.
func LockCleanupCandidatesTx(ctx context.Context, tx pgx.Tx, now time.Time, after *CleanupKey, limit int) ([]CleanupCandidate, *CleanupKey, error) {
	var afterDue time.Time
	var afterID uuid.UUID
	if after != nil {
		afterDue, afterID = after.CleanupDueAt, after.ObjectBlobID
	}
	rows, err := tx.Query(ctx, `
SELECT b.object_blob_id, b.storage_key, COALESCE(c.attempt_count, 0), b.cleanup_due_at,
       (to_jsonb(b) - 'storage_key')::text
  FROM object_blobs b
  LEFT JOIN evidence_blob_cleanup_claims c
    ON c.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at IS NULL
   AND b.cleanup_due_at <= $1
   AND (NOT $3::boolean OR (b.cleanup_due_at, b.object_blob_id) > ($4::timestamptz, $5::uuid))
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
   AND (
       c.object_blob_id IS NULL
       OR (c.claim_state = 'retry_wait' AND c.next_attempt_at <= $1)
       OR (c.claim_state = 'claimed' AND c.claim_expires_at <= $1)
   )
 ORDER BY b.cleanup_due_at, b.object_blob_id
 FOR UPDATE OF b SKIP LOCKED
 LIMIT $2
`, now, limit, after != nil, afterDue, afterID)
	if err != nil {
		return nil, nil, fmt.Errorf("select Evidence cleanup claims: %w", err)
	}
	defer rows.Close()
	page := make([]CleanupCandidate, 0, limit)
	var last CleanupKey
	for rows.Next() {
		var candidate CleanupCandidate
		var payload []byte
		if err := rows.Scan(&candidate.ObjectBlobID, &candidate.StorageKey, &candidate.AttemptCount, &last.CleanupDueAt, &payload); err != nil {
			return nil, nil, fmt.Errorf("scan Evidence cleanup claim candidate: %w", err)
		}
		last.ObjectBlobID = candidate.ObjectBlobID
		_, verifyErr := decodeVerified(candidate.ObjectBlobID, payload)
		if verifyErr != nil && !errors.Is(verifyErr, ErrTupleViolation) {
			return nil, nil, fmt.Errorf("verify Evidence cleanup candidate: %w", verifyErr)
		}
		candidate.Coherent = verifyErr == nil
		page = append(page, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate Evidence cleanup claim candidates: %w", err)
	}
	if len(page) < limit {
		return page, nil, nil
	}
	return page, &last, nil
}

// LockExpiredCleanupMetadataTx locks with SKIP LOCKED up to limit cleaned,
// unattached failed Blobs whose metadata retention ended at cutoff.
func LockExpiredCleanupMetadataTx(ctx context.Context, tx pgx.Tx, cutoff time.Time, limit int) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
SELECT b.object_blob_id
  FROM object_blobs b
  JOIN evidence_blob_cleanup_claims c
    ON c.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'failed'
   AND b.cleaned_up_at <= $1
   AND c.claim_state = 'completed'
   AND c.completed_at <= $1
   AND NOT EXISTS (
       SELECT 1
         FROM evidence e
        WHERE e.object_blob_id = b.object_blob_id
   )
 ORDER BY b.cleaned_up_at, b.object_blob_id
 FOR UPDATE OF b SKIP LOCKED
 LIMIT $2
`, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("select expired Evidence cleanup metadata: %w", err)
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, limit)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan expired Evidence cleanup metadata: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired Evidence cleanup metadata: %w", err)
	}
	return ids, nil
}
