package evidence

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *blobLifecycleService) CreateBlobSlot(ctx context.Context, params blobSlotParams) (blobSlotResult, error) {
	key := LifecycleIdempotencyKey{
		OperationID: LifecycleOperationBlobCreate, ActorUserID: params.ActorUserID,
		ScopeKey: params.IncidentID.String(), ClientTxnID: params.ClientTxnID,
	}
	if payload, found, err := s.idempotency.Get(ctx, key, params.RequestHash); err != nil {
		return blobSlotResult{}, err
	} else if found {
		return blobSlotResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true}, nil
	}

	payload := map[string]any{
		"incident_id":        params.IncidentID.String(),
		"object_blob_id":     params.ObjectBlobID.String(),
		"upload_state":       "pending",
		"target_expires_at":  formatHTTPTime(params.TargetExpiresAt),
		"pending_expires_at": formatHTTPTime(params.PendingExpiresAt),
		"upload_target":      params.UploadTarget,
		"accepted_contract":  params.AcceptedContract,
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return blobSlotResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureIncidentVisibleTx(ctx, tx, params.IncidentID); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, params.IncidentID); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.blobSlots.insertTx(ctx, tx, params); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.uploadLeases.insertTx(ctx, tx, params.ObjectBlobID, params.IncidentID, params.UploadLease); err != nil {
		return blobSlotResult{}, err
	}
	if err := s.idempotency.PutTx(ctx, tx, key, params.RequestHash, payload); err != nil {
		return blobSlotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blobSlotResult{}, err
	}
	return blobSlotResult{Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *blobLifecycleService) GetBlob(ctx context.Context, objectBlobID uuid.UUID) (blobRecord, error) {
	return s.blobs.load(ctx, objectBlobID)
}

func (s *blobLifecycleService) GetUploadLease(ctx context.Context, leaseID uuid.UUID) (uploadLeaseRecord, error) {
	return s.uploadLeases.load(ctx, leaseID)
}

func (s *blobLifecycleService) ClaimUploadLease(ctx context.Context, leaseID uuid.UUID, capabilityHash []byte, now time.Time) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	lease, err := s.uploadLeases.loadForUpdateTx(ctx, tx, leaseID)
	if err != nil {
		return err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, lease.IncidentID); err != nil {
		return err
	}
	if !bytes.Equal(lease.CapabilityHash, capabilityHash) || lease.LeaseState != "issued" || !lease.ExpiresAt.After(now) ||
		lease.Blob.UploadState != "pending" || !lease.Blob.TargetExpiresAt.After(now) || !lease.Blob.PendingExpiresAt.After(now) {
		return errUploadLeaseUnavailable
	}
	if err := s.uploadLeases.claimTx(ctx, tx, leaseID, capabilityHash, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *blobLifecycleService) CompleteUploadLease(ctx context.Context, leaseID uuid.UUID, now time.Time) error {
	return s.uploadLeases.complete(ctx, leaseID, now)
}
