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

type FinalIdempotencyPort interface {
	UpdateFinalIdempotencyOutcomeTx(context.Context, pgx.Tx, jobs.RouteIdempotencyKey, []byte, jobs.Resource) (bool, error)
}

type JobFinalizationPort interface {
	ExtensionFinalizationContextTx(context.Context, pgx.Tx, uuid.UUID) (jobs.ExtensionFinalizationContext, error)
	CompleteSucceededTx(context.Context, pgx.Tx, jobs.TransitionParams, time.Time) (jobs.Resource, error)
}

type OwnerFinalizer struct {
	store        *Store
	transactions JobFinalizationPort
	idempotency  FinalIdempotencyPort
	now          func() time.Time
	fatalSink    func(error)
	commit       func(context.Context, pgx.Tx) error
}

func NewOwnerFinalizer(store *Store, transactions JobFinalizationPort, idempotency FinalIdempotencyPort, now func() time.Time, fatalSink func(error)) (*OwnerFinalizer, error) {
	if store == nil || store.pool == nil || transactions == nil || idempotency == nil {
		return nil, errors.New("extension owner finalizer requires store, job manager, job transaction service, and Auth idempotency port")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if fatalSink == nil {
		fatalSink = func(error) {}
	}
	return &OwnerFinalizer{
		store: store, transactions: transactions, idempotency: idempotency, now: now, fatalSink: fatalSink,
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
	if f == nil || f.transactions == nil || tx == nil || request.Transition.JobID == uuid.Nil ||
		request.Transition.ResultSummary == nil || request.FinalCommitID == "" || committedAt.IsZero() {
		return jobs.Resource{}, ErrInvalidTransition
	}
	metadata, err := f.transactions.ExtensionFinalizationContextTx(ctx, tx, request.Transition.JobID)
	if err != nil {
		return jobs.Resource{}, err
	}
	contract := metadata.Definition
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
	resource, err := f.transactions.CompleteSucceededTx(ctx, tx, request.Transition, committedAt.UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	proof := JobCommitProof{
		JobID:                   request.Transition.JobID,
		OwnerProfileID:          contract.OwnerProfileID,
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
	if err := f.updateFinalIdempotencyOutcome(ctx, tx, metadata, resource); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
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

func (f *OwnerFinalizer) updateFinalIdempotencyOutcome(ctx context.Context, tx pgx.Tx, metadata jobs.ExtensionFinalizationContext, resource jobs.Resource) error {
	requestDigest, err := hex.DecodeString(metadata.NormalizedRequestSHA256)
	if err != nil {
		return ErrIntegrity
	}
	updated, err := f.idempotency.UpdateFinalIdempotencyOutcomeTx(ctx, tx, jobs.RouteIdempotencyKey{
		RouteKey: metadata.IdempotencyRouteKey, ActorUserID: metadata.ActorUserID,
		ScopeKey: metadata.IdempotencyScopeKey, ClientTxnID: metadata.ClientTxnID,
	}, requestDigest, resource)
	if err != nil {
		return err
	}
	if !updated {
		return ErrIntegrity
	}
	return nil
}
