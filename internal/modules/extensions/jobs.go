package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"github.com/google/uuid"
	"sort"
	"time"
)

var (
	ErrReconciliationLimitExceeded = errors.New("extension_reconciliation_limit_exceeded")
	ErrUnclaimReconciliationFailed = errors.New("extension_unclaim_reconciliation_failed")
)

type InactiveJobProof struct {
	JobID                   string
	OwnerProfileID          string
	OperationKind           string
	FinalCommitID           string
	IdempotencyIdentity     json.RawMessage
	NormalizedRequestSHA256 string
	TerminalResult          json.RawMessage
	TerminalResultSHA256    string
	ResourceRefs            json.RawMessage
	AuditCorrelationID      *string
	CommittedAt             time.Time
}

type InactiveJobCancellation struct {
	CancellationRequestID     string
	JobID                     string
	ObservedAt                time.Time
	ObservedBeforeFinalCommit bool
}

type InactiveJob struct {
	JobID                   string
	OwnerProfileID          string
	JobKind                 string
	SubmittedAt             time.Time
	IdempotencyIdentity     json.RawMessage
	NormalizedRequestSHA256 string
	Proof                   *InactiveJobProof
	Cancellation            *InactiveJobCancellation
}

type InactiveJobTerminalOutcome struct {
	JobID                     string
	SubmittedAt               time.Time
	Status                    string
	TerminalResult            json.RawMessage
	EvidenceKind              string
	ProofTerminalResultSHA256 string
	CancellationRequestID     string
}

type ReconciliationCommitOutcome string

const (
	ReconciliationCommitted     ReconciliationCommitOutcome = "committed"
	ReconciliationCommitAbsent  ReconciliationCommitOutcome = "absent"
	ReconciliationIndeterminate ReconciliationCommitOutcome = "indeterminate"
)

type InactiveJobStore interface {
	LoadInactiveJobs(context.Context, string, int) ([]InactiveJob, error)
	ApplyInactiveJobOutcomes(context.Context, string, []InactiveJobTerminalOutcome) (ReconciliationCommitOutcome, error)
}

type ReconciliationOptions struct {
	Coordinator        *Coordinator
	MaxNonterminalJobs int
	Timeout            time.Duration
	FatalIntegritySink func(error)
}

func ReconcileInactiveExtensionJobs(ctx context.Context, store InactiveJobStore, inactiveProfileIDs []string, contracts []JobKindContract, options ReconciliationOptions) error {
	if store == nil || options.Coordinator == nil || options.MaxNonterminalJobs < 1 || options.Timeout <= 0 {
		return startupFailure(ErrUnclaimReconciliationFailed, registryArtifactFinding("job_reconciliation"))
	}
	if options.FatalIntegritySink == nil {
		options.FatalIntegritySink = func(error) {}
	}
	profiles := append([]string(nil), inactiveProfileIDs...)
	sort.Strings(profiles)
	for index, profileID := range profiles {
		if profileID == "" || (index > 0 && profiles[index-1] == profileID) {
			return startupFailure(ErrUnclaimReconciliationFailed, registryArtifactFinding("job_reconciliation"))
		}
		profileCtx, cancel := context.WithTimeout(ctx, options.Timeout)
		err := reconcileInactiveExtensionProfile(profileCtx, store, profileID, contracts, options)
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func reconcileInactiveExtensionProfile(ctx context.Context, store InactiveJobStore, profileID string, contracts []JobKindContract, options ReconciliationOptions) (resultErr error) {
	limit := options.MaxNonterminalJobs
	var jobID any
	actual := 0
	defer func() {
		var fatal *FatalStateIntegrityError
		if resultErr == nil || errors.As(resultErr, &fatal) {
			return
		}
		code, message := "extension_unclaim_reconciliation_failed", "Extension-owned nonterminal work could not be reconciled safely."
		details := map[string]any{"profile_id": profileID, "job_id": jobID}
		if errors.Is(resultErr, ErrReconciliationLimitExceeded) {
			code, message = "extension_reconciliation_limit_exceeded", "Inactive extension job reconciliation exceeds the supported limit."
			details = map[string]any{"profile_id": profileID, "limit": limit, "actual": actual}
		} else if errors.Is(resultErr, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code, message = "extension_reconciliation_timeout", "Inactive extension job reconciliation exceeded its timeout."
			details = map[string]any{"profile_id": profileID, "timeout_seconds": durationSeconds(options.Timeout)}
		}
		resultErr = startupFailure(resultErr, options.Coordinator.registeredResultFinding(code, StartupFinding{Path: "$", ReasonCode: code, Message: message, Details: details}))
	}()
	rows, err := store.LoadInactiveJobs(ctx, profileID, limit+1)
	if err != nil {
		return err
	}
	if len(rows) > limit {
		actual = limit + 1
		return ErrReconciliationLimitExceeded
	}
	contractIndex := make(map[string]JobKindContract, len(contracts))
	for _, contract := range contracts {
		if contract.ProfileID == profileID {
			contractIndex[contract.JobKind] = contract
		}
	}
	outcomes := make([]InactiveJobTerminalOutcome, 0, len(rows))
	var previous InactiveJob
	for index, row := range rows {
		jobID = nil
		if parsed, err := uuid.Parse(row.JobID); err == nil && parsed.String() == row.JobID {
			jobID = row.JobID
		}
		if row.OwnerProfileID != profileID || row.JobID == "" || row.JobKind == "" ||
			len(row.IdempotencyIdentity) == 0 || len(row.NormalizedRequestSHA256) != 64 {
			return ErrUnclaimReconciliationFailed
		}
		if index > 0 && (row.SubmittedAt.Before(previous.SubmittedAt) ||
			(row.SubmittedAt.Equal(previous.SubmittedAt) && row.JobID <= previous.JobID)) {
			return ErrUnclaimReconciliationFailed
		}
		contract, present := contractIndex[row.JobKind]
		if !present {
			return ErrUnclaimReconciliationFailed
		}
		if err := validateInactiveJobIdentity(row, contract); err != nil {
			return ErrUnclaimReconciliationFailed
		}
		outcome, err := classifyInactiveJob(row, contract)
		if err != nil {
			return ErrUnclaimReconciliationFailed
		}
		outcomes = append(outcomes, outcome)
		previous = row
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	jobID = nil
	commitOutcome, err := store.ApplyInactiveJobOutcomes(ctx, profileID, outcomes)
	if commitOutcome == ReconciliationIndeterminate {
		cause := errors.New("indeterminate_database_commit")
		options.FatalIntegritySink(cause)
		return &FatalStateIntegrityError{Cause: cause}
	}
	if commitOutcome == ReconciliationCommitted {
		return nil
	}
	if err != nil {
		return err
	}
	return ErrUnclaimReconciliationFailed
}

func validJobIdentitySchema(schema string) bool { return jobcontract.HasIdentitySchema(schema) }

func validJobIdentityPolicy(contract JobKindContract) bool {
	return (contract.IdempotencyPolicy == "required" && validJobIdentitySchema(contract.IdempotencyIdentitySchemaID)) || (contract.IdempotencyPolicy == "none" && contract.IdempotencyIdentitySchemaID == "")
}

func (contract JobKindContract) resourceContracts() []jobcontract.ResourceContract {
	result := make([]jobcontract.ResourceContract, len(contract.ResourceRefContracts))
	for i, ref := range contract.ResourceRefContracts {
		result[i] = jobcontract.ResourceContract{Kind: ref.ResourceRefKind, SchemaID: ref.ResourceIDSchemaID, MinRefs: ref.MinRefs, MaxRefs: ref.MaxRefs}
	}
	return result
}

func (contract JobKindContract) terminalContract() jobcontract.TerminalContract {
	return jobcontract.TerminalContract{SchemaID: contract.TerminalResultSchemaID, SuccessCodes: contract.TerminalSuccessCodes, Resources: contract.resourceContracts()}
}

func validateInactiveJobIdentity(job InactiveJob, contract JobKindContract) error {
	if !validJobIdentityPolicy(contract) || !lowerHexDigest(job.NormalizedRequestSHA256) {
		return ErrUnclaimReconciliationFailed
	}
	if jobcontract.ValidateTerminalContract(contract.terminalContract()) != nil {
		return ErrUnclaimReconciliationFailed
	}
	if contract.IdempotencyPolicy == "none" {
		value, err := canonicaljson.DecodeStrict(job.IdempotencyIdentity)
		if err != nil || value != nil {
			return ErrUnclaimReconciliationFailed
		}
		return nil
	}
	if _, err := jobcontract.DecodeIdentity(job.IdempotencyIdentity, contract.IdempotencyIdentitySchemaID); err != nil {
		return ErrUnclaimReconciliationFailed
	}
	return nil
}

func classifyInactiveJob(job InactiveJob, contract JobKindContract) (InactiveJobTerminalOutcome, error) {
	if job.Cancellation != nil {
		if contract.CancellationPolicy != "precommit_observable" ||
			job.Cancellation.JobID != job.JobID ||
			job.Cancellation.CancellationRequestID == "" ||
			job.Cancellation.ObservedAt.IsZero() ||
			!job.Cancellation.ObservedBeforeFinalCommit {
			return InactiveJobTerminalOutcome{}, ErrUnclaimReconciliationFailed
		}
	}
	if job.Proof != nil {
		if contract.ProofPolicy != "required_on_terminal_success" {
			return InactiveJobTerminalOutcome{}, ErrUnclaimReconciliationFailed
		}
		terminalResult, err := validateInactiveProof(job, contract, *job.Proof)
		if err != nil {
			return InactiveJobTerminalOutcome{}, err
		}
		return InactiveJobTerminalOutcome{
			JobID:                     job.JobID,
			SubmittedAt:               job.SubmittedAt,
			Status:                    "succeeded",
			TerminalResult:            terminalResult,
			EvidenceKind:              "proof",
			ProofTerminalResultSHA256: job.Proof.TerminalResultSHA256,
		}, nil
	}
	if job.Cancellation != nil {

		result := json.RawMessage(`{"code":"job_canceled","message":"Job canceled."}`)
		return InactiveJobTerminalOutcome{
			JobID:                 job.JobID,
			SubmittedAt:           job.SubmittedAt,
			Status:                "canceled",
			TerminalResult:        result,
			EvidenceKind:          "cancellation",
			CancellationRequestID: job.Cancellation.CancellationRequestID,
		}, nil
	}
	failure := json.RawMessage(`{"code":"extension_profile_unclaimed","message":"Extension profile is not claimed.","retryable":false,"details":{}}`)
	return InactiveJobTerminalOutcome{
		JobID: job.JobID, SubmittedAt: job.SubmittedAt, Status: "failed",
		TerminalResult: failure, EvidenceKind: "absence",
	}, nil
}

func validateInactiveProof(job InactiveJob, contract JobKindContract, proof InactiveJobProof) (json.RawMessage, error) {
	if proof.JobID != job.JobID || proof.OwnerProfileID != job.OwnerProfileID ||
		proof.OperationKind != contract.OperationKind || proof.FinalCommitID == "" ||
		proof.NormalizedRequestSHA256 != job.NormalizedRequestSHA256 ||
		!canonicalRawEqual(proof.IdempotencyIdentity, job.IdempotencyIdentity) ||
		proof.CommittedAt.IsZero() {
		return nil, ErrUnclaimReconciliationFailed
	}
	decode, ok := jobcontract.ResolveTerminal(contract.TerminalResultSchemaID)
	if !ok {
		return nil, ErrUnclaimReconciliationFailed
	}
	terminal, err := decode(proof.TerminalResult)
	if err != nil {
		return nil, ErrUnclaimReconciliationFailed
	}
	_, canonicalTerminal, refs, digest, err := jobcontract.CanonicalSuccess(contract.terminalContract(), terminal)
	if err != nil || digest != proof.TerminalResultSHA256 || !canonicalRawEqual(refs, proof.ResourceRefs) || !canonicalRawEqual(canonicalTerminal, proof.TerminalResult) {
		return nil, ErrUnclaimReconciliationFailed
	}

	proofBytes, err := jobcontract.CanonicalCommitProof(jobcontract.CommitProof{
		JobID: proof.JobID, OwnerProfileID: proof.OwnerProfileID, OperationKind: proof.OperationKind,
		FinalCommitID: proof.FinalCommitID, IdempotencyIdentity: proof.IdempotencyIdentity,
		NormalizedRequestSHA256: proof.NormalizedRequestSHA256, TerminalResult: proof.TerminalResult,
		TerminalResultSHA256: proof.TerminalResultSHA256, ResourceRefs: proof.ResourceRefs,
		AuditCorrelationID: proof.AuditCorrelationID, CommittedAt: proof.CommittedAt,
	})
	if err != nil || len(proofBytes) > contract.MaxProofBytes ||
		jsonNestingDepth(proof.IdempotencyIdentity) > 31 || jsonNestingDepth(proof.TerminalResult) > 31 ||
		jsonNestingDepth(proof.ResourceRefs) > 31 || !lowerHexDigest(proof.TerminalResultSHA256) {
		return nil, ErrUnclaimReconciliationFailed
	}

	return canonicalTerminal, nil
}

func lowerHexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func jsonNestingDepth(raw []byte) int {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return 33
	}
	var depth func(any) int
	depth = func(current any) int {
		maxChild := 0
		switch typed := current.(type) {
		case map[string]any:
			for _, child := range typed {
				if childDepth := depth(child); childDepth > maxChild {
					maxChild = childDepth
				}
			}
			return maxChild + 1
		case []any:
			for _, child := range typed {
				if childDepth := depth(child); childDepth > maxChild {
					maxChild = childDepth
				}
			}
			return maxChild + 1
		default:
			return 1
		}
	}
	return depth(value)
}

func canonicalRawEqual(left []byte, right []byte) bool {
	leftCanonical, leftErr := canonicaljson.Canonicalize(left)
	rightCanonical, rightErr := canonicaljson.Canonicalize(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftCanonical, rightCanonical)
}
