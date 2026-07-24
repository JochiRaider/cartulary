package extensions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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

func ReconcileInactiveExtensionJobs(ctx context.Context, store InactiveJobStore, inactiveProfileIDs []string, contracts []JobKindContract, limit int, fatalSink func(error)) error {
	if store == nil || limit < 1 {
		return ErrUnclaimReconciliationFailed
	}
	if fatalSink == nil {
		fatalSink = func(error) {}
	}
	profiles := append([]string(nil), inactiveProfileIDs...)
	sort.Strings(profiles)
	for index, profileID := range profiles {
		if profileID == "" || (index > 0 && profiles[index-1] == profileID) {
			return ErrUnclaimReconciliationFailed
		}
		if err := reconcileInactiveExtensionProfile(ctx, store, profileID, contracts, limit, fatalSink); err != nil {
			return err
		}
	}
	return nil
}

func reconcileInactiveExtensionProfile(ctx context.Context, store InactiveJobStore, profileID string, contracts []JobKindContract, limit int, fatalSink func(error)) error {
	rows, err := store.LoadInactiveJobs(ctx, profileID, limit+1)
	if err != nil {
		return err
	}
	if len(rows) > limit {
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
	commitOutcome, err := store.ApplyInactiveJobOutcomes(ctx, profileID, outcomes)
	if commitOutcome == ReconciliationIndeterminate {
		cause := errors.New("indeterminate_database_commit")
		fatalSink(cause)
		return cause
	}
	if err != nil || commitOutcome != ReconciliationCommitted {
		if err != nil {
			return err
		}
		return ErrUnclaimReconciliationFailed
	}
	return nil
}

func validateInactiveJobIdentity(job InactiveJob, contract JobKindContract) error {
	if contract.IdempotencyPolicy != "required" ||
		contract.IdempotencyIdentitySchemaID != "cartulary.route_scoped_idempotency_identity.v1" ||
		contract.TerminalResultSchemaID != "cartulary.common_job_terminal_success.v1" ||
		!lowerHexDigest(job.NormalizedRequestSHA256) {
		return ErrUnclaimReconciliationFailed
	}
	var identity struct {
		SchemaID      string  `json:"schema_id"`
		ActorUserID   string  `json:"actor_user_id"`
		RouteIdentity string  `json:"route_identity"`
		ScopeKind     string  `json:"scope_kind"`
		ScopeID       *string `json:"scope_id"`
		ClientTxnID   string  `json:"client_txn_id"`
	}
	if err := decodeClosedObject(
		job.IdempotencyIdentity,
		[]string{"schema_id", "actor_user_id", "route_identity", "scope_kind", "scope_id", "client_txn_id"},
		&identity,
	); err != nil ||
		identity.SchemaID != contract.IdempotencyIdentitySchemaID ||
		identity.ActorUserID == "" || identity.RouteIdentity == "" || identity.ClientTxnID == "" {
		return ErrUnclaimReconciliationFailed
	}
	switch identity.ScopeKind {
	case "deployment":
		if identity.ScopeID != nil {
			return ErrUnclaimReconciliationFailed
		}
	case "incident":
		if identity.ScopeID == nil || *identity.ScopeID == "" {
			return ErrUnclaimReconciliationFailed
		}
	default:
		return ErrUnclaimReconciliationFailed
	}
	return nil
}

func classifyInactiveJob(job InactiveJob, contract JobKindContract) (InactiveJobTerminalOutcome, error) {
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
		if contract.CancellationPolicy != "precommit_observable" ||
			job.Cancellation.JobID != job.JobID ||
			job.Cancellation.CancellationRequestID == "" ||
			job.Cancellation.ObservedAt.IsZero() ||
			!job.Cancellation.ObservedBeforeFinalCommit {
			return InactiveJobTerminalOutcome{}, ErrUnclaimReconciliationFailed
		}
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
	proofBytes, err := json.Marshal(map[string]any{
		"schema_id":                 "cartulary.extension_job_commit_proof.v1",
		"job_id":                    proof.JobID,
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
	})
	if err != nil || len(proofBytes) > contract.MaxProofBytes ||
		jsonNestingDepth(proof.IdempotencyIdentity) > 31 ||
		jsonNestingDepth(proof.TerminalResult) > 31 ||
		jsonNestingDepth(proof.ResourceRefs) > 31 ||
		!lowerHexDigest(proof.TerminalResultSHA256) {
		return nil, ErrUnclaimReconciliationFailed
	}
	var terminal struct {
		Code         string `json:"code"`
		Message      string `json:"message"`
		ResourceRefs []struct {
			Kind  string `json:"kind"`
			ID    string `json:"id"`
			Route string `json:"route,omitempty"`
		} `json:"resource_refs,omitempty"`
	}
	if err := decodeClosedObject(proof.TerminalResult, []string{"code", "message", "resource_refs"}, &terminal); err != nil ||
		terminal.Code == "" || terminal.Message == "" {
		return nil, ErrUnclaimReconciliationFailed
	}
	canonicalTerminal, err := json.Marshal(terminal)
	if err != nil {
		return nil, ErrUnclaimReconciliationFailed
	}
	digest := sha256.Sum256(canonicalTerminal)
	if fmt.Sprintf("%x", digest[:]) != proof.TerminalResultSHA256 {
		return nil, ErrUnclaimReconciliationFailed
	}
	refsJSON, err := json.Marshal(append([]struct {
		Kind  string `json:"kind"`
		ID    string `json:"id"`
		Route string `json:"route,omitempty"`
	}{}, terminal.ResourceRefs...))
	if err != nil || !canonicalRawEqual(refsJSON, proof.ResourceRefs) {
		return nil, ErrUnclaimReconciliationFailed
	}
	limits := make(map[string]int, len(contract.ResourceRefContracts))
	for _, resourceContract := range contract.ResourceRefContracts {
		limits[resourceContract.ResourceRefKind] = resourceContract.MaxRefs
	}
	counts := map[string]int{}
	previousKind, previousID := "", ""
	for index, resourceRef := range terminal.ResourceRefs {
		limit, present := limits[resourceRef.Kind]
		if !present || resourceRef.ID == "" || (resourceRef.Route != "" && resourceRef.Route[0] != '/') {
			return nil, ErrUnclaimReconciliationFailed
		}
		counts[resourceRef.Kind]++
		if counts[resourceRef.Kind] > limit {
			return nil, ErrUnclaimReconciliationFailed
		}
		if index > 0 && (resourceRef.Kind < previousKind ||
			(resourceRef.Kind == previousKind && resourceRef.ID <= previousID)) {
			return nil, ErrUnclaimReconciliationFailed
		}
		previousKind, previousID = resourceRef.Kind, resourceRef.ID
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

func decodeClosedObject(raw []byte, allowed []string, target any) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return err
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range object {
		if _, present := allowedSet[key]; !present {
			return ErrUnclaimReconciliationFailed
		}
	}
	return json.Unmarshal(raw, target)
}

func canonicalRawEqual(left []byte, right []byte) bool {
	var leftValue any
	var rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	leftCanonical, leftErr := json.Marshal(leftValue)
	rightCanonical, rightErr := json.Marshal(rightValue)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftCanonical, rightCanonical)
}
