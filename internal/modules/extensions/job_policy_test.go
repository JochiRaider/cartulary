package extensions

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func checkJobPolicyVariants(t *testing.T) {
	for _, policy := range []string{"forbidden", "none", "not_cancelable", "combined", "aggregate overflow", "retired terminal"} {
		t.Run(policy, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(row map[string]any) {
				if row["job_kind"] != "import.apply_v1" || row["proof_policy"] == nil {
					return
				}
				switch policy {
				case "forbidden":
					row["proof_policy"], row["resource_ref_contracts"] = "forbidden", []any{}
				case "none":
					row["idempotency_policy"], row["idempotency_identity_schema_id"] = "none", nil
				case "not_cancelable":
					row["cancellation_policy"] = "not_cancelable"
				case "combined":
					row["proof_policy"], row["resource_ref_contracts"] = "forbidden", []any{}
					row["idempotency_policy"], row["idempotency_identity_schema_id"] = "none", nil
					row["cancellation_policy"] = "not_cancelable"
				case "retired terminal":
					row["terminal_result_schema_id"] = "cartulary.common_job_terminal_success.v1"
				case "aggregate overflow":
					row["resource_ref_contracts"].([]any)[0].(map[string]any)["max_refs"] = 2
				}
			})
			coordinator, err := NewCoordinator(source)
			if policy == "aggregate overflow" || policy == "retired terminal" {
				if err == nil {
					t.Fatal("admitted invalid job contract", policy)
				}
				return
			}
			if err != nil {
				t.Fatalf("adopted policy rejected: %v", err)
			}
			for _, contract := range coordinator.JobKindContracts() {
				if contract.JobKind == "import.apply_v1" && contract.SHA256() != source[generatedExtensionsRoot+"job-contracts/import/import.apply_v1.json"].SHA256 {
					t.Fatal("typed contract lost policy/null identity")
				}
			}
		})
	}
}

func checkNonIdempotentInactiveJobs(t *testing.T) {
	contract := JobKindContract{ProfileID: "test_profile", JobKind: "test_profile.run_v1", OperationKind: "test_profile.run", ProofPolicy: "forbidden", IdempotencyPolicy: "none", TerminalSuccessCodes: []string{"done"}, TerminalResultSchemaID: "cartulary.common_job_terminal_success.v2", CancellationPolicy: "not_cancelable", MaxProofBytes: 4096}
	row := InactiveJob{JobID: "00000000-0000-4000-8000-000000000001", OwnerProfileID: contract.ProfileID, JobKind: contract.JobKind, SubmittedAt: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC), IdempotencyIdentity: json.RawMessage(`null`), NormalizedRequestSHA256: strings.Repeat("a", 64)}
	store := &fakeInactiveJobStore{rows: []InactiveJob{row}}
	if err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{contract}, testReconciliationOptions(t, 1, nil)); err != nil {
		t.Fatalf("non-idempotent inactive job rejected: %v", err)
	}
	if store.applyCalls != 1 || len(store.outcomes) != 1 || store.outcomes[0].EvidenceKind != "absence" {
		t.Fatal("missing atomic inactive disposition")
	}
	invalid := contract
	invalid.TerminalSuccessCodes = nil
	store = &fakeInactiveJobStore{rows: []InactiveJob{row}}
	if ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{invalid}, testReconciliationOptions(t, 1, nil)) == nil || store.applyCalls != 0 {
		t.Fatal("incomplete selected terminal policy mutated reconciliation")
	}
	for _, identity := range []json.RawMessage{nil, json.RawMessage(`{}`), testInactiveJobIdentity()} {
		row.IdempotencyIdentity = identity
		store = &fakeInactiveJobStore{rows: []InactiveJob{row}}
		if ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{contract}, testReconciliationOptions(t, 1, nil)) == nil || store.applyCalls != 0 {
			t.Fatal("invalid none identity mutated reconciliation")
		}
	}
}

func checkInactivePolicyEvidence(t *testing.T) {
	contract := JobKindContract{ProfileID: "test_profile", JobKind: "test_profile.run_v1", OperationKind: "test_profile.run", ProofPolicy: "required_on_terminal_success", IdempotencyPolicy: "none", TerminalSuccessCodes: []string{"done"}, TerminalResultSchemaID: "cartulary.common_job_terminal_success.v2", CancellationPolicy: "not_cancelable", MaxProofBytes: 4096}
	now := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	row := InactiveJob{JobID: "00000000-0000-4000-8000-000000000001", OwnerProfileID: contract.ProfileID, JobKind: contract.JobKind, SubmittedAt: now, IdempotencyIdentity: json.RawMessage(`null`), NormalizedRequestSHA256: strings.Repeat("a", 64)}
	terminal := json.RawMessage(`{"code":"done","message":"Done."}`)
	digest := sha256.Sum256(terminal)
	row.Proof = &InactiveJobProof{JobID: row.JobID, OwnerProfileID: row.OwnerProfileID, OperationKind: contract.OperationKind, FinalCommitID: "00000000-0000-4000-8000-000000000002", IdempotencyIdentity: row.IdempotencyIdentity, NormalizedRequestSHA256: row.NormalizedRequestSHA256, TerminalResult: terminal, TerminalResultSHA256: fmt.Sprintf("%x", digest[:]), ResourceRefs: json.RawMessage(`[]`), CommittedAt: now}
	store := &fakeInactiveJobStore{rows: []InactiveJob{row}}
	if err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{contract}, testReconciliationOptions(t, 1, nil)); err != nil || store.outcomes[0].Status != "succeeded" {
		t.Fatal("null-identity proof rejected", err)
	}
	t.Run("canonical proof bound after JSONB formatting", func(t *testing.T) {
		selected, job, proof := contract, row, *row.Proof
		proof.TerminalResult = json.RawMessage(`{ "message": "<&>", "code": "done" }`)
		terminal := `{"code":"done","message":"<&>"}`
		digest := sha256.Sum256([]byte(terminal))
		proof.TerminalResultSHA256 = fmt.Sprintf("%x", digest[:])
		job.Proof = &proof
		canonical := `{"audit_correlation_id":null,"committed_at":"2026-10-31T00:00:00Z","final_commit_id":"00000000-0000-4000-8000-000000000002","idempotency_identity":null,"job_id":"00000000-0000-4000-8000-000000000001","normalized_request_sha256":"` + strings.Repeat("a", 64) + `","operation_kind":"test_profile.run","owner_profile_id":"test_profile","resource_refs":[],"schema_id":"cartulary.extension_job_commit_proof.v2","terminal_result":` + terminal + `,"terminal_result_sha256":"` + proof.TerminalResultSHA256 + `"}`
		selected.MaxProofBytes = len(canonical)
		store := &fakeInactiveJobStore{rows: []InactiveJob{job}}
		if err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{selected}, testReconciliationOptions(t, 1, nil)); err != nil || len(store.outcomes) != 1 || string(store.outcomes[0].TerminalResult) != terminal {
			t.Fatal("canonical exact proof bound failed", err)
		}
		selected.MaxProofBytes--
		store = &fakeInactiveJobStore{rows: []InactiveJob{job}}
		if err := ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{selected}, testReconciliationOptions(t, 1, nil)); err == nil || store.applyCalls != 0 {
			t.Fatal("one-byte proof overflow mutated reconciliation")
		}
	})

	for _, name := range []string{"unknown code", "cancellation as success", "missing required resource", "noncanonical empty references"} {
		t.Run(name, func(t *testing.T) {
			selected, job, proof := contract, row, *row.Proof
			job.Proof = &proof
			switch name {
			case "unknown code", "cancellation as success":
				code := "unregistered"
				if name == "cancellation as success" {
					code = "job_canceled"
				}
				proof.TerminalResult = json.RawMessage(`{"code":"` + code + `","message":"Done."}`)
				digest := sha256.Sum256(proof.TerminalResult)
				proof.TerminalResultSHA256 = fmt.Sprintf("%x", digest[:])
			case "noncanonical empty references":
				proof.TerminalResult = json.RawMessage(`{"code":"done","message":"Done.","resource_refs":[]}`)
			case "missing required resource":
				selected.ResourceRefContracts = []JobResourceRefContract{{ResourceRefKind: "incident", ResourceIDSchemaID: "cartulary.common_job_resource_ref_id.v1", MinRefs: 1, MaxRefs: 1}}
			}
			store := &fakeInactiveJobStore{rows: []InactiveJob{job}}
			if ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{selected}, testReconciliationOptions(t, 1, nil)) == nil || store.applyCalls != 0 {
				t.Fatal("invalid selected terminal proof mutated reconciliation")
			}
		})
	}
	for _, name := range []string{"forbidden proof", "forbidden cancellation with proof", "forbidden cancellation without proof"} {
		t.Run(name, func(t *testing.T) {
			selected, job := contract, row
			if name == "forbidden proof" {
				selected.ProofPolicy = "forbidden"
			} else {
				job.Cancellation = &InactiveJobCancellation{JobID: job.JobID, CancellationRequestID: "00000000-0000-4000-8000-000000000003", ObservedAt: now, ObservedBeforeFinalCommit: true}
				if name == "forbidden cancellation without proof" {
					job.Proof = nil
				}
			}
			store := &fakeInactiveJobStore{rows: []InactiveJob{job}}
			if ReconcileInactiveExtensionJobs(context.Background(), store, []string{contract.ProfileID}, []JobKindContract{selected}, testReconciliationOptions(t, 1, nil)) == nil || store.applyCalls != 0 {
				t.Fatal("contradictory policy evidence mutated job")
			}
		})
	}
}
