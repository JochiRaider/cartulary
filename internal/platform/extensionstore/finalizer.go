package extensionstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrIndeterminateCommit = errors.New("extension job finalization commit is indeterminate")

type OwnerMutation func(context.Context, pgx.Tx) error

type JobFinalizationRequest struct {
	Execution          jobs.Execution
	Completion         jobs.SuccessCompletion
	FinalCommitID      string
	AuditCorrelationID *string
	Mutate             OwnerMutation
}

type JobFailureFinalizationRequest struct {
	Execution  jobs.Execution
	Completion jobs.FailureCompletion
	Mutate     OwnerMutation
}

type JobCancellationFinalizationRequest struct {
	Execution  jobs.Execution
	Completion jobs.CancellationCompletion
	Mutate     OwnerMutation
}

type FinalReceiptReconciliationPort interface {
	ReconcileFinalIdempotencyOutcomeTx(context.Context, pgx.Tx, jobs.RouteIdempotencyKey, []byte, jobs.Resource) (bool, error)
}

type JobFinalizationPort interface {
	ExtensionFinalizationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error)
	ExtensionCancellationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error)
	CompleteSucceededTx(context.Context, pgx.Tx, jobs.Execution, jobs.SuccessCompletion, time.Time) (jobs.Resource, error)
	CompleteFailedTx(context.Context, pgx.Tx, jobs.Execution, jobs.FailureCompletion, time.Time) (jobs.Resource, error)
	CompleteTimedOutTx(context.Context, pgx.Tx, jobs.Execution, jobs.FailureCompletion, time.Time) (jobs.Resource, error)
	CompleteCanceledTx(context.Context, pgx.Tx, jobs.Execution, jobs.CancellationCompletion, time.Time) (jobs.Resource, error)
	ReadTerminalResourceTx(context.Context, pgx.Tx, uuid.UUID) (jobs.Resource, error)
}

type OwnerFinalizer struct {
	store        *Store
	transactions JobFinalizationPort
	idempotency  FinalReceiptReconciliationPort
	now          func() time.Time
	fatalSink    func(error)
	commit       func(context.Context, pgx.Tx) error
}

func NewOwnerFinalizer(store *Store, transactions JobFinalizationPort, idempotency FinalReceiptReconciliationPort, now func() time.Time, fatalSink func(error)) (*OwnerFinalizer, error) {
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

// NewOwnerMutationFinalizer supplies the same commit-proof classification to
// Base operations that have no Job or extension claim. It cannot finalize Jobs.
func NewOwnerMutationFinalizer(fatalSink func(error)) (*OwnerFinalizer, error) {
	if fatalSink == nil {
		return nil, errors.New("owner mutation finalizer requires a fatal integrity sink")
	}
	return &OwnerFinalizer{
		fatalSink: fatalSink,
		commit:    func(ctx context.Context, tx pgx.Tx) error { return tx.Commit(ctx) },
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
	outcome, err := f.CommitSuccessTx(ctx, tx, request, resource)
	if outcome == CommitUnknown {
		f.fatalSink(err)
		return jobs.Resource{}, fmt.Errorf("%w: %v", ErrIndeterminateCommit, err)
	}
	if outcome != CommitProven {
		return jobs.Resource{}, err
	}
	return resource, nil
}

// CommitSuccessTx commits a transaction already finalized by FinalizeSuccessTx.
// It also serves cross-owner transactions, whose coordinator owns the fatal
// integrity consequence. An absent proof cannot establish rollback after an
// uncertain acknowledgement. Classification gets a bounded independent context
// so cancellation after commit cannot hide an authoritative success.
func (f *OwnerFinalizer) CommitSuccessTx(ctx context.Context, tx pgx.Tx, request JobFinalizationRequest, resource jobs.Resource) (CommitOutcome, error) {
	if f == nil || f.store == nil || f.transactions == nil || tx == nil || request.Execution.JobID().String() != resource.JobID || resource.Status != jobs.StatusSucceeded {
		return CommitUnknown, ErrInvalidTransition
	}
	return f.commitWithProof(ctx, tx, func(proofCtx context.Context) (bool, error) {
		return f.provesSuccess(proofCtx, request, resource) && f.provesTerminalResource(proofCtx, resource), nil
	})
}

// commitWithProof is the single physical final-commit classifier for owner
// mutations. Its caller decides how to report a proven absence or invoke the
// fatal integrity path; it never retries a transaction.
func (f *OwnerFinalizer) commitWithProof(ctx context.Context, tx pgx.Tx, proveCommit func(context.Context) (bool, error)) (CommitOutcome, error) {
	err := f.commit(ctx, tx)
	if err == nil {
		return CommitProven, nil
	}
	proofCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if rollbackErr := tx.Rollback(proofCtx); rollbackErr == nil || provenAbsentCommitError(err) {
		return CommitAbsent, err
	}
	committed, proofErr := proveCommit(proofCtx)
	if proofErr == nil && committed {
		return CommitProven, nil
	}
	return CommitUnknown, errors.Join(err, proofErr)
}

func provenAbsentCommitError(err error) bool {
	if errors.Is(err, pgx.ErrTxCommitRollback) {
		return true
	}
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "40001", "40P01", "23505", "23503", "23514":
			return true
		}
	}
	return false
}

func (f *OwnerFinalizer) provesTerminalResource(ctx context.Context, expected jobs.Resource) bool {
	tx, err := f.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false
	}
	defer func() { _ = tx.Rollback(ctx) }()
	id, err := uuid.Parse(expected.JobID)
	if err != nil {
		return false
	}
	retained, err := f.transactions.ReadTerminalResourceTx(ctx, tx, id)
	if err != nil {
		return false
	}
	a, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	b, err := json.Marshal(retained)
	return err == nil && bytes.Equal(a, b)
}

func (f *OwnerFinalizer) provesSuccess(ctx context.Context, request JobFinalizationRequest, resource jobs.Resource) bool {
	proof, err := f.store.JobCommitProof(ctx, request.Execution.JobID())
	if err != nil || proof == nil || resource.FinishedAt == nil || proof.FinalCommitID != request.FinalCommitID || !proof.CommittedAt.Equal(*resource.FinishedAt) {
		return false
	}
	var retained jobs.ResultSummary
	if json.Unmarshal(proof.TerminalResult, &retained) != nil || !reflect.DeepEqual(&retained, resource.ResultSummary) {
		return false
	}
	encoded, err := json.Marshal(resource.ResultSummary)
	if err != nil {
		return false
	}
	digest := sha256.Sum256(encoded)
	return proof.TerminalResultSHA256 == hex.EncodeToString(digest[:])
}

func (f *OwnerFinalizer) FinalizeSuccessTx(ctx context.Context, tx pgx.Tx, request JobFinalizationRequest, committedAt time.Time) (jobs.Resource, error) {
	if f == nil || f.transactions == nil || tx == nil || request.Execution.JobID() == uuid.Nil ||
		request.Completion.ResultSummary.Code == "" || request.FinalCommitID == "" || committedAt.IsZero() {
		return jobs.Resource{}, ErrInvalidTransition
	}
	metadata, err := f.transactions.ExtensionFinalizationContextTx(ctx, tx, request.Execution)
	if err != nil {
		return jobs.Resource{}, err
	}
	contract := metadata.Definition
	if contract.Extension == nil {
		return jobs.Resource{}, ErrInvalidTransition
	}
	normalizedSummary, terminalJSON, resourceRefsJSON, terminalDigest, err :=
		jobs.CanonicalExtensionTerminalSuccess(contract, &request.Completion.ResultSummary)
	if err != nil {
		return jobs.Resource{}, err
	}
	request.Completion.ResultSummary = *normalizedSummary
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	resource, err := f.transactions.CompleteSucceededTx(ctx, tx, request.Execution, request.Completion, committedAt.UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	proof := JobCommitProof{
		JobID:                   request.Execution.JobID(),
		OwnerProfileID:          contract.Extension.OwnerProfileID,
		OperationKind:           contract.Extension.OperationKind,
		FinalCommitID:           request.FinalCommitID,
		IdempotencyIdentity:     metadata.IdempotencyIdentity,
		NormalizedRequestSHA256: metadata.NormalizedRequestSHA256,
		TerminalResult:          terminalJSON,
		TerminalResultSHA256:    terminalDigest,
		ResourceRefs:            resourceRefsJSON,
		AuditCorrelationID:      request.AuditCorrelationID,
		CommittedAt:             committedAt.UTC(),
	}
	if err := ValidateJobCommitProofSize(proof, contract.Extension.MaxProofBytes); err != nil {
		return jobs.Resource{}, err
	}
	if err := InsertJobCommitProof(ctx, tx, proof); err != nil {
		return jobs.Resource{}, err
	}
	if err := f.reconcileFinalIdempotencyOutcome(ctx, tx, metadata, resource); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}

func (f *OwnerFinalizer) FinalizeFailure(ctx context.Context, request JobFailureFinalizationRequest) (jobs.Resource, error) {
	return f.finalizeFailure(ctx, request, false)
}

// FinalizeTimeout records an operational abort classified under the shared
// deadline policy. Mutate records abort evidence without prepared content.
func (f *OwnerFinalizer) FinalizeTimeout(ctx context.Context, request JobFailureFinalizationRequest) (jobs.Resource, error) {
	return f.finalizeFailure(ctx, request, true)
}

func (f *OwnerFinalizer) finalizeFailure(ctx context.Context, request JobFailureFinalizationRequest, timedOut bool) (jobs.Resource, error) {
	if f == nil || f.store == nil || f.store.pool == nil {
		return jobs.Resource{}, ErrInvalidTransition
	}
	tx, err := f.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var metadata jobs.ExtensionFinalizationContext
	if timedOut {
		metadata, err = f.transactions.ExtensionCancellationContextTx(ctx, tx, request.Execution)
	} else {
		metadata, err = f.transactions.ExtensionFinalizationContextTx(ctx, tx, request.Execution)
	}
	if err != nil {
		return jobs.Resource{}, err
	}
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	var resource jobs.Resource
	if timedOut {
		resource, err = f.transactions.CompleteTimedOutTx(ctx, tx, request.Execution, request.Completion, f.now().UTC())
	} else {
		resource, err = f.transactions.CompleteFailedTx(ctx, tx, request.Execution, request.Completion, f.now().UTC())
	}
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := f.reconcileFinalIdempotencyOutcome(ctx, tx, metadata, resource); err != nil {
		return jobs.Resource{}, err
	}
	return f.commitTerminalResource(ctx, tx, resource)
}

// FinalizeCancellation gives owner attempt evidence and cancellation the same
// atomic boundary and idempotent receipt behavior as success and failure.
func (f *OwnerFinalizer) FinalizeCancellation(ctx context.Context, request JobCancellationFinalizationRequest) (jobs.Resource, error) {
	if f == nil || f.store == nil || f.store.pool == nil {
		return jobs.Resource{}, ErrInvalidTransition
	}
	tx, err := f.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	metadata, err := f.transactions.ExtensionCancellationContextTx(ctx, tx, request.Execution)
	if err != nil {
		return jobs.Resource{}, err
	}
	if request.Mutate != nil {
		if err := request.Mutate(ctx, tx); err != nil {
			return jobs.Resource{}, err
		}
	}
	resource, err := f.transactions.CompleteCanceledTx(ctx, tx, request.Execution, request.Completion, f.now().UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := f.reconcileFinalIdempotencyOutcome(ctx, tx, metadata, resource); err != nil {
		return jobs.Resource{}, err
	}
	return f.commitTerminalResource(ctx, tx, resource)
}

func (f *OwnerFinalizer) commitTerminalResource(ctx context.Context, tx pgx.Tx, resource jobs.Resource) (jobs.Resource, error) {
	outcome, err := f.commitWithProof(ctx, tx, func(proofCtx context.Context) (bool, error) {
		return f.provesTerminalResource(proofCtx, resource), nil
	})
	if outcome == CommitProven {
		return resource, nil
	}
	if outcome == CommitUnknown {
		f.fatalSink(err)
		return jobs.Resource{}, fmt.Errorf("%w: %v", ErrIndeterminateCommit, err)
	}
	return jobs.Resource{}, err
}

// ValidateJobCommitProofSize applies the same canonical bound at final commit and read-only admission.
func ValidateJobCommitProofSize(proof JobCommitProof, maxBytes int) error {
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

func (f *OwnerFinalizer) reconcileFinalIdempotencyOutcome(ctx context.Context, tx pgx.Tx, metadata jobs.ExtensionFinalizationContext, resource jobs.Resource) error {
	if metadata.OperatorOperationID != uuid.Nil {
		// Jobs has validated the declared v2 actor and lease under this
		// transaction's row lock. The Job and its proof are the local
		// invocation's durable outcome; there is no human route receipt.
		if metadata.ActorUserID != uuid.Nil || resource.SubmittedByUserID != nil || metadata.Definition.Extension == nil || metadata.Definition.Extension.IdentitySchemaID != jobs.AttributedRouteIdentitySchema {
			return ErrIntegrity
		}
		return nil
	}

	requestDigest, err := hex.DecodeString(metadata.NormalizedRequestSHA256)
	if err != nil {
		return ErrIntegrity
	}
	reconciled, err := f.idempotency.ReconcileFinalIdempotencyOutcomeTx(ctx, tx, jobs.RouteIdempotencyKey{
		RouteKey: metadata.IdempotencyRouteKey, ActorUserID: metadata.ActorUserID,
		ScopeKey: metadata.IdempotencyScopeKey, ClientTxnID: metadata.ClientTxnID,
	}, requestDigest, resource)
	if err != nil {
		return err
	}
	if !reconciled {
		return ErrIntegrity
	}
	return nil
}

// CommitOwnerMutation classifies a synchronous owner's transaction using the
// same commit rules as Jobs. The proof must identify the exact immutable owner
// outcome and idempotency receipt written by this transaction. Absence alone
// never proves rollback after an uncertain commit acknowledgement.
func (f *OwnerFinalizer) CommitOwnerMutation(ctx context.Context, tx pgx.Tx, proveCommit func(context.Context) (bool, error)) error {
	if f == nil || tx == nil || proveCommit == nil {
		return ErrInvalidTransition
	}
	outcome, err := f.commitWithProof(ctx, tx, proveCommit)
	if outcome == CommitUnknown {
		f.fatalSink(err)
		return ErrIndeterminateCommit
	}
	return err
}
