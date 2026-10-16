package evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type blobVerificationOutcome struct {
	request     blobVerificationRequest
	blob        blobRecord
	observation *observedObject
	err         error
}

func (s *blobVerificationService) prepare(ctx context.Context, execution jobs.Execution, scope jobs.Scope) blobVerificationOutcome {
	var result blobVerificationOutcome
	raw, err := s.manager.HandlerPayload(ctx, execution)
	if err != nil {
		result.err = err
		return result
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var trailing any
	if decoder.Decode(&result.request) != nil || decoder.Decode(&trailing) != io.EOF || result.request.SchemaID != blobVerificationRequestSchema || result.request.ObjectBlobID == uuid.Nil || result.request.IncidentID == uuid.Nil || result.request.UploadLeaseID == uuid.Nil || scope.Kind != jobs.ScopeKindIncident || scope.IncidentID == nil || *scope.IncidentID != result.request.IncidentID {
		result.err = errBlobNotAttachable
		return result
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		result.err = err
		return result
	}
	defer rollbackBlobVerification(tx)
	if err = s.transactions.ValidateExecutionTx(ctx, tx, execution); err != nil {
		result.err = err
		return result
	}
	result.blob, err = loadBlobForUpdateTx(ctx, tx, result.request.ObjectBlobID)
	if err == nil && result.blob.IncidentID != result.request.IncidentID {
		err = ErrIncidentMismatch
	}
	if err == nil {
		var leaseID uuid.UUID
		leaseID, err = eligibleVerificationBlob(ctx, tx, result.blob, s.now())
		if err == nil && leaseID != result.request.UploadLeaseID {
			err = errBlobNotAttachable
		}
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		result.err = err
		return result
	}
	// No transition or source-row lock remains while consuming object bytes.
	result.observation, result.err = s.objects.observeUploadedObject(ctx, result.blob)
	return result
}

func (s *blobVerificationService) handle(parent context.Context, execution jobs.Execution) error {
	ctx, cancel := context.WithTimeout(parent, 120*time.Second)
	defer cancel()
	resource, err := s.manager.ObserveExecution(ctx, execution)
	if err != nil {
		return err
	}
	if resource.Status == jobs.StatusCancelRequested {
		return s.finishTerminal(parent, execution, blobVerificationOutcome{}, false, true)
	}
	done := make(chan blobVerificationOutcome, 1)
	go func() { done <- s.prepare(ctx, execution, resource.Scope) }()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var result blobVerificationOutcome
	canceled, timedOut := false, false
	for {
		select {
		case result = <-done:
			goto drained
		case <-ctx.Done():
			timedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
			cancel()
			result = <-done
			goto drained
		case <-ticker.C:
			observed, observeErr := s.manager.ObserveExecution(ctx, execution)
			if observeErr != nil || observed.Status == jobs.StatusCancelRequested {
				canceled = observed.Status == jobs.StatusCancelRequested
				timedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
				cancel()
				result = <-done
				if observeErr != nil && !timedOut && parent.Err() == nil {
					return observeErr
				}
				goto drained
			}
		}
	}
drained:
	timedOut = timedOut || errors.Is(ctx.Err(), context.DeadlineExceeded)
	if parent.Err() != nil {
		return parent.Err()
	}
	if timedOut || canceled {
		return s.finishTerminal(parent, execution, result, timedOut, canceled)
	}
	// Publication shares the attempt deadline. If acknowledgement is uncertain,
	// the terminal transaction must acquire the same fence before deciding: a
	// previously committed success cannot be replaced by a timeout.
	err = s.finish(ctx, execution, result, false, false)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return s.finishTerminal(parent, execution, result, true, false)
	}
	return err
}

func (s *blobVerificationService) finishTerminal(parent context.Context, execution jobs.Execution, result blobVerificationOutcome, timedOut, canceled bool) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	return s.finish(ctx, execution, result, timedOut, canceled)
}

func (s *blobVerificationService) finish(ctx context.Context, execution jobs.Execution, result blobVerificationOutcome, timedOut, canceled bool) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackBlobVerification(tx)
	now := s.now().UTC()
	resource, err := s.manager.ObserveExecution(ctx, execution)
	if err != nil {
		return err
	}
	progress := resource.Progress
	// Cancellation may have won after the final byte. Check under the same
	// transition lock used by success; no proof is visible before that decision.
	err = s.transactions.ValidateExecutionTx(ctx, tx, execution)
	if errors.Is(err, jobs.ErrCancellationRequested) {
		canceled = true
	} else if err != nil {
		return err
	}
	if canceled && !timedOut {
		if _, err = s.transactions.CompleteCanceledTx(ctx, tx, execution, jobs.CancellationCompletion{Progress: progress}, now); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	reason := ""
	if timedOut {
		reason = "verification_timeout"
	} else if result.err != nil {
		reason = "object_store_unavailable"
		var rejected AttachRejectedError
		if errors.As(result.err, &rejected) {
			reason = "blob_failed"
			if rejected.ReasonCode == attachReasonBlobPending {
				reason = "blob_pending"
			}
		}
		if errors.Is(result.err, errBlobNotAttachable) || errors.Is(result.err, ErrBlobNotFound) || errors.Is(result.err, ErrIncidentMismatch) {
			reason = "blob_failed"
		}
		if objectstore.IsObjectNotFound(result.err) {
			reason = "blob_pending"
		}
		if errors.Is(result.err, errBlobVerificationMismatch) {
			reason = "accepted_contract_mismatch"
		}
	}
	if reason == "" {
		blob, loadErr := loadBlobForUpdateTx(ctx, tx, result.request.ObjectBlobID)
		if loadErr != nil {
			return loadErr
		}
		lease, eligibleErr := eligibleVerificationBlob(ctx, tx, blob, now)
		if eligibleErr != nil || blob.IncidentID != result.request.IncidentID || lease != result.request.UploadLeaseID {
			reason = "blob_failed"
		} else if result.observation == nil || result.observation.Size != blob.ByteSize || (blob.ExpectedSHA256Hex != nil && *blob.ExpectedSHA256Hex != result.observation.SHA256Hex) {
			reason = "accepted_contract_mismatch"
		} else {
			observation := result.observation
			tag, writeErr := tx.Exec(ctx, `INSERT INTO evidence_blob_verifications(object_blob_id,upload_lease_id,job_id,verified_at,observed_size,observed_content_type,observed_sha256_hex)
                VALUES($1,$2,$3,$4,$5,$6,$7)
                ON CONFLICT(object_blob_id) DO UPDATE SET job_id=EXCLUDED.job_id,verified_at=EXCLUDED.verified_at
                WHERE evidence_blob_verifications.upload_lease_id=EXCLUDED.upload_lease_id
                  AND evidence_blob_verifications.observed_size=EXCLUDED.observed_size
                  AND evidence_blob_verifications.observed_content_type=EXCLUDED.observed_content_type
                  AND evidence_blob_verifications.observed_sha256_hex=EXCLUDED.observed_sha256_hex`, blob.ObjectBlobID, lease, execution.JobID(), now, observation.Size, observation.ContentType, observation.SHA256Hex)
			if writeErr != nil {
				return writeErr
			}
			if tag.RowsAffected() != 1 {
				return errors.New("evidence: conflicting retained verification observation")
			}
			progress.Completed = int(blob.ByteSize)
			if _, err = s.transactions.CompleteSucceededTx(ctx, tx, execution, jobs.SuccessCompletion{Progress: progress, ResultSummary: jobs.ResultSummary{Code: "evidence_blob_verified", Message: "Uploaded evidence verified."}}, now); err != nil {
				return err
			}
			return tx.Commit(ctx)
		}
	}
	if reason == "accepted_contract_mismatch" {
		blob, loadErr := loadBlobForUpdateTx(ctx, tx, result.request.ObjectBlobID)
		if loadErr != nil {
			return loadErr
		}
		lease, eligibleErr := eligibleVerificationBlob(ctx, tx, blob, now)
		if eligibleErr == nil && blob.IncidentID == result.request.IncidentID && lease == result.request.UploadLeaseID {
			mismatch := "declared_size_mismatch"
			if result.observation != nil && result.observation.Size == blob.ByteSize {
				mismatch = "expected_sha256_mismatch"
			}
			if err = failBlobTx(ctx, tx, blob.ObjectBlobID, mismatch, now); err != nil {
				return err
			}
		}
	}
	failure := jobs.FailureCompletion{Progress: progress, ErrorSummary: jobs.ErrorSummary{Code: "evidence_blob_verification_failed", Message: "Uploaded evidence verification failed.", Retryable: reason == "object_store_unavailable" || reason == "blob_pending", Details: map[string]any{"reason_code": reason}}}
	if timedOut {
		_, err = s.transactions.CompleteTimedOutTx(ctx, tx, execution, failure, now)
	} else {
		_, err = s.transactions.CompleteFailedTx(ctx, tx, execution, failure, now)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
