package evidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const blobVerificationJobKind = "evidence.verify_blob_v1"
const blobVerificationWorkerKind = "evidence.blob_verification_worker_v1"
const blobVerificationRequestSchema = "cartulary.evidence_blob_verification_request.v1"

var errBlobNotVerified = errors.New("evidence: blob verification is pending")

type blobVerificationRequest struct {
	SchemaID      string    `json:"schema_id"`
	ObjectBlobID  uuid.UUID `json:"object_blob_id"`
	IncidentID    uuid.UUID `json:"incident_id"`
	UploadLeaseID uuid.UUID `json:"upload_lease_id"`
}

type blobVerificationTransactions interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	CompleteSucceededTx(context.Context, pgx.Tx, jobs.Execution, jobs.SuccessCompletion, time.Time) (jobs.Resource, error)
	CompleteFailedTx(context.Context, pgx.Tx, jobs.Execution, jobs.FailureCompletion, time.Time) (jobs.Resource, error)
	CompleteTimedOutTx(context.Context, pgx.Tx, jobs.Execution, jobs.FailureCompletion, time.Time) (jobs.Resource, error)
	CompleteCanceledTx(context.Context, pgx.Tx, jobs.Execution, jobs.CancellationCompletion, time.Time) (jobs.Resource, error)
}
type blobVerificationManager interface {
	HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
}
type blobVerificationRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	Notify(uuid.UUID)
}
type blobVerificationService struct {
	db           postgres.DB
	objects      blobVerificationObjectReader
	transactions blobVerificationTransactions
	manager      blobVerificationManager
	runner       blobVerificationRunner
	now          func() time.Time
}

func rollbackBlobVerification(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}

func (s *blobVerificationService) admit(ctx context.Context, actorID, blobID uuid.UUID, clientTxnID string) (jobs.Resource, error) {
	key := authn.RouteIdempotencyKey{RouteKey: "evidence.blobs.verify", ActorUserID: actorID, ScopeKey: blobID.String(), ClientTxnID: clientTxnID}
	normalized, _ := json.Marshal(struct {
		BlobID uuid.UUID `json:"object_blob_id"`
	}{blobID})
	digest := sha256.Sum256(normalized)
	blob, err := (blobRepository{db: s.db}).load(ctx, blobID)
	if err != nil {
		return jobs.Resource{}, err
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer rollbackBlobVerification(tx)
	if err = authn.LockRouteIdempotencyTx(ctx, tx, key); err != nil {
		return jobs.Resource{}, err
	}
	if _, err = admission.NewChecker(s.db).CheckTx(ctx, tx, blob.IncidentID, actorID, admission.Requirement{AllowedRoles: admission.RolesEditorReviewerAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return jobs.Resource{}, err
	}
	previous, err := authn.GetRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(previous.RequestHash, digest[:]) {
			return jobs.Resource{}, authn.ErrClientTxnConflict
		}
		var job jobs.Resource
		if err = json.Unmarshal(previous.ResponseJSON, &job); err != nil {
			return job, err
		}
		return job, tx.Commit(ctx)
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return jobs.Resource{}, err
	}
	blob, err = loadBlobForUpdateTx(ctx, tx, blobID)
	if err != nil {
		return jobs.Resource{}, err
	}
	leaseID, err := eligibleVerificationBlob(ctx, tx, blob, s.now())
	if err != nil {
		return jobs.Resource{}, err
	}
	payload, err := json.Marshal(blobVerificationRequest{blobVerificationRequestSchema, blobID, blob.IncidentID, leaseID})
	if err != nil {
		return jobs.Resource{}, err
	}
	var total *int
	if blob.ByteSize > 0 {
		value := int(blob.ByteSize)
		total = &value
	}
	job, err := s.transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: blobVerificationJobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &blob.IncidentID}, SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Total: total}, HandlerPayload: payload}, s.now().UTC())
	if err != nil {
		return job, err
	}
	if err = authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, digest[:], http.StatusAccepted, job); err != nil {
		return job, err
	}
	if err = tx.Commit(ctx); err != nil {
		return job, err
	}
	s.runner.Notify(uuid.MustParse(job.JobID))
	return job, nil
}

func eligibleVerificationBlob(ctx context.Context, tx pgx.Tx, blob blobRecord, now time.Time) (uuid.UUID, error) {
	if blob.UploadState == "quarantined" {
		return uuid.Nil, AttachRejectedError{ReasonCode: attachReasonBlobQuarantined, Cause: errBlobNotAttachable}
	}
	if blob.UploadState != "pending" || !blob.PendingExpiresAt.After(now) {
		return uuid.Nil, AttachRejectedError{ReasonCode: attachReasonBlobFailed, Cause: errBlobNotAttachable}
	}
	associated, err := isBlobAssociatedTx(ctx, tx, blob.ObjectBlobID)
	if err != nil {
		return uuid.Nil, err
	}
	if associated {
		return uuid.Nil, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `SELECT lease_id FROM evidence_object_upload_leases WHERE object_blob_id=$1 AND lease_state='completed'`, blob.ObjectBlobID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, AttachRejectedError{ReasonCode: attachReasonBlobPending, Cause: errBlobNotAttachable}
	}
	return id, err
}

type blobObservationQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Verification facts are bound to the one-use completed lease. The caller's
// final transaction still owns attachment, availability and all record effects.
func retainedBlobObservation(ctx context.Context, q blobObservationQuerier, blobID uuid.UUID) (*observedObject, error) {
	var observed observedObject
	err := q.QueryRow(ctx, `SELECT v.observed_size,v.observed_content_type,v.observed_sha256_hex
        FROM evidence_blob_verifications v
        JOIN evidence_object_upload_leases l ON l.lease_id=v.upload_lease_id AND l.object_blob_id=v.object_blob_id
        WHERE v.object_blob_id=$1 AND l.lease_state='completed'`, blobID).Scan(&observed.Size, &observed.ContentType, &observed.SHA256Hex)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errBlobNotVerified
	}
	return &observed, err
}
