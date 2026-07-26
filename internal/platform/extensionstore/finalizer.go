package extensionstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrIndeterminateCommit = errors.New("extension job finalization commit is indeterminate")

type OwnerMutation func(context.Context, pgx.Tx) error

type JobFinalizationRequest struct {
	Transition         jobs.TransitionParams
	FinalCommitID      string
	AuditCorrelationID *string
	Mutate             OwnerMutation
}

type OwnerFinalizer struct {
	store     *Store
	manager   *jobs.Manager
	now       func() time.Time
	fatalSink func(error)
	commit    func(context.Context, pgx.Tx) error
}

func NewOwnerFinalizer(store *Store, manager *jobs.Manager, now func() time.Time, fatalSink func(error)) (*OwnerFinalizer, error) {
	if store == nil || store.pool == nil || manager == nil {
		return nil, errors.New("extension owner finalizer requires store and job manager")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if fatalSink == nil {
		fatalSink = func(error) {}
	}
	return &OwnerFinalizer{
		store: store, manager: manager, now: now, fatalSink: fatalSink,
		commit: func(ctx context.Context, tx pgx.Tx) error { return tx.Commit(ctx) },
	}, nil
}

func (f *OwnerFinalizer) FinalizeSuccess(ctx context.Context, request JobFinalizationRequest) (jobs.Resource, error) {
	if f == nil || f.store == nil || f.store.pool == nil {
		return jobs.Resource{}, ErrInvalidTransition
	}
	tx, err := f.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resource, err := f.FinalizeSuccessTx(ctx, tx, request, f.now().UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := f.commit(ctx, tx); err != nil {
		f.fatalSink(err)
		return jobs.Resource{}, fmt.Errorf("%w: %v", ErrIndeterminateCommit, err)
	}
	return resource, nil
}

func (f *OwnerFinalizer) FinalizeSuccessTx(ctx context.Context, tx pgx.Tx, request JobFinalizationRequest, committedAt time.Time) (jobs.Resource, error) {
	if f == nil || f.manager == nil || tx == nil || request.Transition.JobID == uuid.Nil ||
		request.Transition.ResultSummary == nil || request.FinalCommitID == "" || committedAt.IsZero() {
		return jobs.Resource{}, ErrInvalidTransition
	}
	metadata, err := extensionJobMetadataForUpdate(ctx, tx, request.Transition.JobID)
	if err != nil {
		return jobs.Resource{}, err
	}
	contract, present := f.manager.ExtensionContract(metadata.JobKind)
	if !present || !contract.ProofRequired ||
		contract.OwnerProfileID != metadata.OwnerProfileID ||
		contract.WorkerKind != metadata.WorkerKind {
		return jobs.Resource{}, ErrIntegrity
	}
	normalizedSummary, terminalJSON, resourceRefsJSON, terminalDigest, err :=
		jobs.CanonicalExtensionTerminalSuccess(contract, request.Transition.ResultSummary)
	if err != nil {
		return jobs.Resource{}, err
	}
	request.Transition.ResultSummary = normalizedSummary
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	resource, err := jobs.CompleteSucceededTx(ctx, tx, request.Transition, committedAt.UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	proof := JobCommitProof{
		JobID:                   request.Transition.JobID,
		OwnerProfileID:          metadata.OwnerProfileID,
		OperationKind:           contract.OperationKind,
		FinalCommitID:           request.FinalCommitID,
		IdempotencyIdentity:     metadata.IdempotencyIdentity,
		NormalizedRequestSHA256: metadata.NormalizedRequestSHA256,
		TerminalResult:          terminalJSON,
		TerminalResultSHA256:    terminalDigest,
		ResourceRefs:            resourceRefsJSON,
		AuditCorrelationID:      request.AuditCorrelationID,
		CommittedAt:             committedAt.UTC(),
	}
	if err := validateProofSize(proof, contract.MaxProofBytes); err != nil {
		return jobs.Resource{}, err
	}
	if err := InsertJobCommitProof(ctx, tx, proof); err != nil {
		return jobs.Resource{}, err
	}
	if err := updateFinalIdempotencyOutcome(ctx, tx, metadata, resource); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}

type extensionJobMetadata struct {
	OwnerProfileID          string
	JobKind                 string
	WorkerKind              string
	IdempotencyIdentity     json.RawMessage
	IdempotencyRouteKey     string
	IdempotencyScopeKey     string
	NormalizedRequestSHA256 string
	ActorUserID             uuid.UUID
	ClientTxnID             string
}

func extensionJobMetadataForUpdate(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (extensionJobMetadata, error) {
	var metadata extensionJobMetadata
	err := tx.QueryRow(ctx, `
SELECT extension_owner_profile_id, extension_job_kind, handler_name,
       extension_idempotency_identity, extension_idempotency_route_key,
       extension_idempotency_scope_key, extension_normalized_request_sha256,
       submitted_by_user_id
  FROM jobs
 WHERE job_id = $1
   AND status IN ('queued', 'running', 'cancel_requested')
 FOR UPDATE
`, jobID).Scan(
		&metadata.OwnerProfileID,
		&metadata.JobKind,
		&metadata.WorkerKind,
		&metadata.IdempotencyIdentity,
		&metadata.IdempotencyRouteKey,
		&metadata.IdempotencyScopeKey,
		&metadata.NormalizedRequestSHA256,
		&metadata.ActorUserID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return extensionJobMetadata{}, jobs.ErrInvalidTransition
	}
	if err != nil {
		return extensionJobMetadata{}, err
	}
	var identity jobs.RouteScopedIdempotencyIdentity
	if err := json.Unmarshal(metadata.IdempotencyIdentity, &identity); err != nil ||
		identity.SchemaID != "cartulary.route_scoped_idempotency_identity.v1" ||
		identity.ActorUserID != metadata.ActorUserID.String() ||
		identity.RouteIdentity != metadata.IdempotencyRouteKey+":"+metadata.IdempotencyScopeKey {
		return extensionJobMetadata{}, ErrIntegrity
	}
	metadata.ClientTxnID = identity.ClientTxnID
	return metadata, nil
}

func validateProofSize(proof JobCommitProof, maxBytes int) error {
	payload := map[string]any{
		"schema_id":                 "cartulary.extension_job_commit_proof.v1",
		"job_id":                    proof.JobID.String(),
		"owner_profile_id":          proof.OwnerProfileID,
		"operation_kind":            proof.OperationKind,
		"final_commit_id":           proof.FinalCommitID,
		"idempotency_identity":      json.RawMessage(proof.IdempotencyIdentity),
		"normalized_request_sha256": proof.NormalizedRequestSHA256,
		"terminal_result":           json.RawMessage(proof.TerminalResult),
		"terminal_result_sha256":    proof.TerminalResultSHA256,
		"resource_refs":             json.RawMessage(proof.ResourceRefs),
		"audit_correlation_id":      proof.AuditCorrelationID,
		"committed_at":              proof.CommittedAt.UTC().Format(time.RFC3339Nano),
	}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if len(canonical) > maxBytes {
		return ErrInvalidTransition
	}
	return nil
}

func updateFinalIdempotencyOutcome(ctx context.Context, tx pgx.Tx, metadata extensionJobMetadata, resource jobs.Resource) error {
	requestDigest, err := hex.DecodeString(metadata.NormalizedRequestSHA256)
	if err != nil {
		return ErrIntegrity
	}
	payload, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
UPDATE route_idempotency
   SET response_json = $1
 WHERE route_key = $2
   AND actor_user_id = $3
   AND scope_key = $4
   AND client_txn_id = $5
   AND request_hash = $6
`, payload, metadata.IdempotencyRouteKey, metadata.ActorUserID,
		metadata.IdempotencyScopeKey, metadata.ClientTxnID, requestDigest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIntegrity
	}
	return nil
}
