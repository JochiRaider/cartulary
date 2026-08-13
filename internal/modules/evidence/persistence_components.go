package evidence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// blobSlotRepository owns pending-slot persistence inside a caller transaction.
type blobSlotRepository struct{}

func (repository blobSlotRepository) insertTx(
	ctx context.Context,
	tx pgx.Tx,
	params BlobSlotParams,
) error {
	_, err := tx.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, expected_sha256_hex,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8, $9, $10, $11, $11)
`, params.ObjectBlobID, params.IncidentID, params.ActorUserID, params.StorageKey, params.ByteSize,
		params.FilenameHint, params.ContentTypeHint, params.ExpectedSHA256Hex,
		params.TargetExpiresAt.UTC(), params.PendingExpiresAt.UTC(), params.TargetExpiresAt.Add(-60*time.Minute).UTC())
	return err
}

// uploadLeaseRepository owns the durable, one-shot application upload lease.
// The signed capability is useful only when its digest and every binding agree
// with this server-side record.
type uploadLeaseRepository struct {
	db postgres.DB
}

func (uploadLeaseRepository) insertTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	incidentID uuid.UUID,
	params UploadLeaseCreateParams,
) error {
	headers, err := json.Marshal(params.RequiredHeaders)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO evidence_object_upload_leases (
    object_blob_id, lease_id, capability_hash, incident_id,
    issuing_user_id, issuing_session_id, issued_at, expires_at,
    required_method, required_headers, accepted_contract_sha256,
    lease_state, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'issued', $7, $7)
`, objectBlobID, params.LeaseID, params.CapabilityHash, incidentID,
		params.IssuingUserID, params.IssuingSessionID, params.IssuedAt.UTC(), params.ExpiresAt.UTC(),
		params.RequiredMethod, headers, params.AcceptedContractSHA256)
	return err
}

func (repository uploadLeaseRepository) load(ctx context.Context, leaseID uuid.UUID) (UploadLeaseRecord, error) {
	return repository.scan(ctx, repository.db.QueryRow(ctx, `
SELECT lease_id, object_blob_id, incident_id, capability_hash,
       issuing_user_id, issuing_session_id, issued_at, expires_at,
       required_method, required_headers, accepted_contract_sha256, lease_state
  FROM evidence_object_upload_leases
 WHERE lease_id = $1
`, leaseID))
}

func (repository uploadLeaseRepository) loadForUpdateTx(ctx context.Context, tx pgx.Tx, leaseID uuid.UUID) (UploadLeaseRecord, error) {
	record, err := repository.scan(ctx, tx.QueryRow(ctx, `
SELECT lease_id, object_blob_id, incident_id, capability_hash,
       issuing_user_id, issuing_session_id, issued_at, expires_at,
       required_method, required_headers, accepted_contract_sha256, lease_state
  FROM evidence_object_upload_leases
 WHERE lease_id = $1
 FOR UPDATE
`, leaseID))
	if err != nil {
		return UploadLeaseRecord{}, err
	}
	blob, err := loadBlobForUpdateTx(ctx, tx, record.ObjectBlobID)
	if err != nil {
		return UploadLeaseRecord{}, err
	}
	record.Blob = blob
	return record, nil
}

type uploadLeaseRow interface {
	Scan(...any) error
}

func (repository uploadLeaseRepository) scan(ctx context.Context, row uploadLeaseRow) (UploadLeaseRecord, error) {
	var record UploadLeaseRecord
	var headers []byte
	if err := row.Scan(&record.LeaseID, &record.ObjectBlobID, &record.IncidentID, &record.CapabilityHash,
		&record.IssuingUserID, &record.IssuingSessionID, &record.IssuedAt, &record.ExpiresAt,
		&record.RequiredMethod, &headers, &record.AcceptedContractSHA256, &record.LeaseState); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UploadLeaseRecord{}, ErrUploadLeaseNotFound
		}
		return UploadLeaseRecord{}, err
	}
	if err := json.Unmarshal(headers, &record.RequiredHeaders); err != nil {
		return UploadLeaseRecord{}, err
	}
	blob, err := blobRepository(repository).load(ctx, record.ObjectBlobID)
	if err != nil {
		return UploadLeaseRecord{}, err
	}
	record.Blob = blob
	return record, nil
}

func (uploadLeaseRepository) claimTx(
	ctx context.Context,
	tx pgx.Tx,
	leaseID uuid.UUID,
	capabilityHash []byte,
	now time.Time,
) error {
	tag, err := tx.Exec(ctx, `
UPDATE evidence_object_upload_leases
   SET lease_state = 'claimed', claimed_at = $3, updated_at = $3
 WHERE lease_id = $1
   AND capability_hash = $2
   AND lease_state = 'issued'
`, leaseID, capabilityHash, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrUploadLeaseUnavailable
	}
	return nil
}

func (repository uploadLeaseRepository) complete(ctx context.Context, leaseID uuid.UUID, now time.Time) error {
	tag, err := repository.db.Exec(ctx, `
UPDATE evidence_object_upload_leases
   SET lease_state = 'completed', completed_at = $2, updated_at = $2
 WHERE lease_id = $1
   AND lease_state = 'claimed'
`, leaseID, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrUploadLeaseUnavailable
	}
	return nil
}

// blobRepository owns blob reads and row locking. Lifecycle transitions remain
// with blobLifecycleRepository so storage identity and state changes are not
// conflated.
type blobRepository struct {
	db postgres.DB
}

func (repository blobRepository) load(ctx context.Context, objectBlobID uuid.UUID) (BlobRecord, error) {
	row := repository.db.QueryRow(ctx, `
SELECT b.object_blob_id, b.incident_id, b.storage_key, b.upload_state, b.byte_size,
       b.filename_hint, b.content_type_hint, b.expected_sha256_hex,
       b.observed_size, b.observed_content_type, b.observed_sha256_hex,
       b.target_expires_at, b.pending_expires_at,
       COALESCE(l.lease_state, '')
  FROM object_blobs b
  LEFT JOIN evidence_object_upload_leases l ON l.object_blob_id = b.object_blob_id
 WHERE b.object_blob_id = $1
`, objectBlobID)
	var record BlobRecord
	if err := row.Scan(&record.ObjectBlobID, &record.IncidentID, &record.StorageKey, &record.UploadState, &record.ByteSize,
		&record.FilenameHint, &record.ContentTypeHint, &record.ExpectedSHA256Hex,
		&record.ObservedSize, &record.ObservedContentType, &record.ObservedSHA256Hex,
		&record.TargetExpiresAt, &record.PendingExpiresAt, &record.UploadLeaseState); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BlobRecord{}, ErrBlobNotFound
		}
		return BlobRecord{}, err
	}
	return record, nil
}

func (blobRepository) loadForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
) (BlobRecord, error) {
	return loadBlobForUpdateTx(ctx, tx, objectBlobID)
}

type evidenceRecordRepository struct{}

func (evidenceRecordRepository) loadForUpdateTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (evidenceMeta, error) {
	return loadEvidenceMetaForUpdateTx(ctx, tx, recordID)
}

func (evidenceRecordRepository) associateBlobTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	objectBlobID uuid.UUID,
	storageRef string,
	sha256 *string,
	now time.Time,
) error {
	_, err := tx.Exec(ctx, `
UPDATE evidence
   SET object_blob_id = $2,
       upload_state = 'available',
       storage_ref = $3,
       blob_hash = $4,
       received_at = COALESCE(received_at, $5),
       updated_at = $5
 WHERE record_id = $1
`, recordID, objectBlobID, storageRef, sha256, now)
	return err
}

// blobLifecycleRepository owns blob-state transitions and cleanup selection.
// Callers retain transaction orchestration and object-store byte deletion.
type blobLifecycleRepository struct {
	db postgres.DB
}

func (blobLifecycleRepository) failTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	reason string,
	now time.Time,
) error {
	return failBlobTx(ctx, tx, objectBlobID, reason, now)
}

func (blobLifecycleRepository) recordFinalizeFailureTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	now time.Time,
) (bool, error) {
	return recordNonTerminalFinalizeFailureTx(ctx, tx, objectBlobID, now)
}

func (blobLifecycleRepository) markAvailableTx(
	ctx context.Context,
	tx pgx.Tx,
	objectBlobID uuid.UUID,
	observed *ObservedObject,
	now time.Time,
) error {
	return markBlobAvailableTx(ctx, tx, objectBlobID, observed, now)
}

func (repository blobLifecycleRepository) markExpiredPending(
	ctx context.Context,
	now time.Time,
) (int, error) {
	schedule := evidencepolicy.ScheduleFailure(now)
	tag, err := repository.db.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'failed',
       terminal_reason = 'pending_timeout',
       failed_at = $1,
       cleanup_due_at = $2,
       updated_at = $1
 WHERE upload_state = 'pending'
   AND pending_expires_at <= $1
`, schedule.FailedAt, schedule.CleanupDueAt)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
