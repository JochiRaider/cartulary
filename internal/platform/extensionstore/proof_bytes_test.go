package extensionstore

import (
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"github.com/google/uuid"
)

func checkProofCanonicalByteLimit(t *testing.T) {
	const terminal = `{"code":"done","message":"<&>"}`
	proof := JobCommitProof{
		JobID:          uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		OwnerProfileID: "test_profile", OperationKind: "test_profile.run", FinalCommitID: "commit:one",
		IdempotencyIdentity: []byte(`null`), NormalizedRequestSHA256: strings.Repeat("a", 64),
		TerminalResult: []byte(terminal), TerminalResultSHA256: strings.Repeat("b", 64), ResourceRefs: []byte(`[]`),
		CommittedAt: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC),
	}
	// An independent literal fixes the entire twelve-member encoding, including
	// explicit nulls and no final LF. This helper checks byte accounting only.
	canonical := `{"audit_correlation_id":null,"committed_at":"2026-11-02T00:00:00Z","final_commit_id":"commit:one","idempotency_identity":null,"job_id":"00000000-0000-4000-8000-000000000001","normalized_request_sha256":"` + strings.Repeat("a", 64) + `","operation_kind":"test_profile.run","owner_profile_id":"test_profile","resource_refs":[],"schema_id":"cartulary.extension_job_commit_proof.v2","terminal_result":` + terminal + `,"terminal_result_sha256":"` + strings.Repeat("b", 64) + `"}`
	encoded, err := jobcontract.CanonicalCommitProof(jobcontract.CommitProof{
		JobID: proof.JobID.String(), OwnerProfileID: proof.OwnerProfileID, OperationKind: proof.OperationKind,
		FinalCommitID: proof.FinalCommitID, IdempotencyIdentity: proof.IdempotencyIdentity,
		NormalizedRequestSHA256: proof.NormalizedRequestSHA256, TerminalResult: proof.TerminalResult,
		TerminalResultSHA256: proof.TerminalResultSHA256, ResourceRefs: proof.ResourceRefs,
		AuditCorrelationID: proof.AuditCorrelationID, CommittedAt: proof.CommittedAt,
	})
	if err != nil || string(encoded) != canonical {
		t.Fatalf("complete proof byte vector mismatch: %q / %v", encoded, err)
	}

	if err := ValidateJobCommitProofSize(proof, len(canonical)); err != nil {
		t.Errorf("exact canonical byte bound rejected: %v", err)
	}
	if err := ValidateJobCommitProofSize(proof, len(canonical)-1); err == nil {
		t.Error("one byte over canonical bound admitted")
	}
	for _, name := range []string{"duplicate members", "invalid raw Unicode", "invalid scalar Unicode", "unsafe final commit identity", "unsafe audit identity"} {
		t.Run(name, func(t *testing.T) {
			invalid := proof
			switch name {
			case "duplicate members":
				invalid.TerminalResult = []byte(`{"code":"done","code":"done","message":"<&>"}`)
			case "invalid raw Unicode":
				invalid.TerminalResult = []byte(`{"code":"done","message":"\ud800"}`)
			case "unsafe final commit identity":
				invalid.FinalCommitID = "commit:bad/path"
			case "unsafe audit identity":
				value := "audit:bad/path"
				invalid.AuditCorrelationID = &value
			case "invalid scalar Unicode":
				invalid.FinalCommitID = "commit:" + string([]byte{0xff})
			}
			if ValidateJobCommitProofSize(invalid, 4096) == nil {
				t.Fatal("malformed proof repaired during canonical encoding")
			}
		})
	}
}
