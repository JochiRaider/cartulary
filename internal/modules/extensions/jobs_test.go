package extensions

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestReconcileInactiveExtensionJobs_Unit(t *testing.T) {
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	identity := testInactiveJobIdentity()
	terminal := json.RawMessage(`{"code":"done","message":"Done.","resource_refs":[{"kind":"thing","id":"1","route":"/things/1"}]}`)
	refs := json.RawMessage(`[{"kind":"thing","id":"1","route":"/things/1"}]`)
	digest := sha256.Sum256(terminal)
	contract := JobKindContract{
		ProfileID:                   "test_profile",
		JobKind:                     "test_profile.run_v1",
		ProgressUnitID:              "test_profile.run.attempt.v1",
		OperationKind:               "test_profile.run",
		ProofPolicy:                 "required_on_terminal_success",
		IdempotencyPolicy:           "required",
		IdempotencyIdentitySchemaID: "cartulary.route_scoped_idempotency_identity.v1",
		TerminalResultSchemaID:      "cartulary.common_job_terminal_success.v1",
		CancellationPolicy:          "precommit_observable",
		MaxProofBytes:               4096,
		ResourceRefContracts: []JobResourceRefContract{{
			ResourceRefKind: "thing",
			MaxRefs:         1,
		}},
	}
	proofJob := InactiveJob{
		JobID:                   "00000000-0000-4000-8000-000000000001",
		OwnerProfileID:          "test_profile",
		JobKind:                 contract.JobKind,
		SubmittedAt:             now,
		IdempotencyIdentity:     identity,
		NormalizedRequestSHA256: fmt.Sprintf("%064x", 1),
		Proof: &InactiveJobProof{
			JobID:                   "00000000-0000-4000-8000-000000000001",
			OwnerProfileID:          "test_profile",
			OperationKind:           contract.OperationKind,
			FinalCommitID:           "commit:1",
			IdempotencyIdentity:     identity,
			NormalizedRequestSHA256: fmt.Sprintf("%064x", 1),
			TerminalResult:          terminal,
			TerminalResultSHA256:    fmt.Sprintf("%x", digest[:]),
			ResourceRefs:            refs,
			CommittedAt:             now,
		},
		Cancellation: &InactiveJobCancellation{
			CancellationRequestID:     "cancel:1",
			JobID:                     "00000000-0000-4000-8000-000000000001",
			ObservedAt:                now.Add(-time.Second),
			ObservedBeforeFinalCommit: true,
		},
	}
	canceledJob := InactiveJob{
		JobID:                   "00000000-0000-4000-8000-000000000002",
		OwnerProfileID:          "test_profile",
		JobKind:                 contract.JobKind,
		SubmittedAt:             now.Add(time.Second),
		IdempotencyIdentity:     identity,
		NormalizedRequestSHA256: fmt.Sprintf("%064x", 2),
		Cancellation: &InactiveJobCancellation{
			CancellationRequestID:     "cancel:2",
			JobID:                     "00000000-0000-4000-8000-000000000002",
			ObservedAt:                now,
			ObservedBeforeFinalCommit: true,
		},
	}
	failedJob := InactiveJob{
		JobID:                   "00000000-0000-4000-8000-000000000003",
		OwnerProfileID:          "test_profile",
		JobKind:                 contract.JobKind,
		SubmittedAt:             now.Add(2 * time.Second),
		IdempotencyIdentity:     identity,
		NormalizedRequestSHA256: fmt.Sprintf("%064x", 3),
	}
	store := &fakeInactiveJobStore{rows: []InactiveJob{proofJob, canceledJob, failedJob}}
	if err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{"test_profile"}, []JobKindContract{contract}, 3, nil); err != nil {
		t.Fatalf("reconcile inactive jobs: %v", err)
	}
	if store.applyCalls != 1 || len(store.outcomes) != 3 {
		t.Fatalf("unexpected atomic application: calls=%d outcomes=%#v", store.applyCalls, store.outcomes)
	}
	if store.outcomes[0].Status != "succeeded" || string(store.outcomes[0].TerminalResult) != string(terminal) {
		t.Fatalf("proof did not take precedence with exact terminal result: %#v", store.outcomes[0])
	}
	if store.outcomes[1].Status != "canceled" || store.outcomes[2].Status != "failed" {
		t.Fatalf("unexpected cancellation/unclaimed outcomes: %#v", store.outcomes)
	}
}

func TestReconcileInactiveExtensionJobsFailsBeforeMutation_Unit(t *testing.T) {
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	contract := JobKindContract{
		ProfileID: "test_profile", JobKind: "test_profile.run_v1",
		ProgressUnitID: "test_profile.run.attempt.v1",
		OperationKind:  "test_profile.run", ProofPolicy: "required_on_terminal_success",
		IdempotencyPolicy:           "required",
		IdempotencyIdentitySchemaID: "cartulary.route_scoped_idempotency_identity.v1",
		TerminalResultSchemaID:      "cartulary.common_job_terminal_success.v1",
		CancellationPolicy:          "precommit_observable", MaxProofBytes: 4096,
	}
	base := InactiveJob{
		JobID: "00000000-0000-4000-8000-000000000001", OwnerProfileID: "test_profile",
		JobKind: contract.JobKind, SubmittedAt: now,
		IdempotencyIdentity: testInactiveJobIdentity(), NormalizedRequestSHA256: fmt.Sprintf("%064x", 1),
	}
	t.Run("overflow", func(t *testing.T) {
		store := &fakeInactiveJobStore{rows: []InactiveJob{base, {
			JobID: "00000000-0000-4000-8000-000000000002", OwnerProfileID: "test_profile",
			JobKind: contract.JobKind, SubmittedAt: now.Add(time.Second),
			IdempotencyIdentity: testInactiveJobIdentity(), NormalizedRequestSHA256: fmt.Sprintf("%064x", 2),
		}}}
		err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{"test_profile"}, []JobKindContract{contract}, 1, nil)
		if !errors.Is(err, ErrReconciliationLimitExceeded) || store.applyCalls != 0 {
			t.Fatalf("overflow err=%v apply_calls=%d", err, store.applyCalls)
		}
	})
	t.Run("contradictory proof", func(t *testing.T) {
		row := base
		row.Proof = &InactiveJobProof{
			JobID: row.JobID, OwnerProfileID: row.OwnerProfileID,
			OperationKind: "wrong.operation", FinalCommitID: "commit",
			IdempotencyIdentity:     row.IdempotencyIdentity,
			NormalizedRequestSHA256: row.NormalizedRequestSHA256,
			TerminalResult:          json.RawMessage(`{"code":"done","message":"Done."}`),
			TerminalResultSHA256:    fmt.Sprintf("%064x", 1),
			ResourceRefs:            json.RawMessage(`[]`), CommittedAt: now,
		}
		store := &fakeInactiveJobStore{rows: []InactiveJob{row}}
		err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{"test_profile"}, []JobKindContract{contract}, 1, nil)
		if !errors.Is(err, ErrUnclaimReconciliationFailed) || store.applyCalls != 0 {
			t.Fatalf("contradiction err=%v apply_calls=%d", err, store.applyCalls)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		store := &fakeInactiveJobStore{rows: []InactiveJob{base}}
		err := ReconcileInactiveExtensionJobs(ctx, store, []string{"test_profile"}, []JobKindContract{contract}, 1, nil)
		if !errors.Is(err, context.Canceled) || store.applyCalls != 0 {
			t.Fatalf("timeout err=%v apply_calls=%d", err, store.applyCalls)
		}
	})
	t.Run("indeterminate commit", func(t *testing.T) {
		store := &fakeInactiveJobStore{rows: []InactiveJob{base}, commit: ReconciliationIndeterminate}
		fatal := 0
		err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{"test_profile"}, []JobKindContract{contract}, 1, func(error) { fatal++ })
		if err == nil || fatal != 1 || store.applyCalls != 1 {
			t.Fatalf("indeterminate err=%v fatal=%d apply_calls=%d", err, fatal, store.applyCalls)
		}
	})
}

func testInactiveJobIdentity() json.RawMessage {
	return json.RawMessage(`{"schema_id":"cartulary.route_scoped_idempotency_identity.v1","actor_user_id":"00000000-0000-4000-8000-000000000001","route_identity":"test.profile.run:deployment","scope_kind":"deployment","scope_id":null,"client_txn_id":"txn-1"}`)
}

type fakeInactiveJobStore struct {
	rows       []InactiveJob
	outcomes   []InactiveJobTerminalOutcome
	applyCalls int
	commit     ReconciliationCommitOutcome
}

func (s *fakeInactiveJobStore) LoadInactiveJobs(context.Context, string, int) ([]InactiveJob, error) {
	return append([]InactiveJob(nil), s.rows...), nil
}

func (s *fakeInactiveJobStore) ApplyInactiveJobOutcomes(_ context.Context, _ string, outcomes []InactiveJobTerminalOutcome) (ReconciliationCommitOutcome, error) {
	s.applyCalls++
	s.outcomes = append([]InactiveJobTerminalOutcome(nil), outcomes...)
	if s.commit == "" {
		return ReconciliationCommitted, nil
	}
	return s.commit, nil
}
