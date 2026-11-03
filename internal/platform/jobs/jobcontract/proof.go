package jobcontract

import (
	"encoding/json"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const CommitProofSchema = "cartulary.extension_job_commit_proof.v2"

// CommitProof is the logical current proof. Selected-owner policy and retained
// evidence validation belong to the caller; its byte encoding is shared by the
// final writer and inactive-profile admission.
type CommitProof struct {
	JobID                   string          `json:"job_id"`
	OwnerProfileID          string          `json:"owner_profile_id"`
	OperationKind           string          `json:"operation_kind"`
	FinalCommitID           string          `json:"final_commit_id"`
	IdempotencyIdentity     json.RawMessage `json:"idempotency_identity"`
	NormalizedRequestSHA256 string          `json:"normalized_request_sha256"`
	TerminalResult          json.RawMessage `json:"terminal_result"`
	TerminalResultSHA256    string          `json:"terminal_result_sha256"`
	ResourceRefs            json.RawMessage `json:"resource_refs"`
	AuditCorrelationID      *string         `json:"audit_correlation_id"`
	CommittedAt             time.Time       `json:"committed_at"`
}

func CanonicalCommitProof(proof CommitProof) ([]byte, error) {
	if !canonicalIdentityUUID(&proof.JobID) || !validCommitProofIdentifier(proof.FinalCommitID) ||
		(proof.AuditCorrelationID != nil && !validCommitProofIdentifier(*proof.AuditCorrelationID)) {
		return nil, errors.New("invalid proof identity")
	}
	for _, value := range []string{proof.JobID, proof.OwnerProfileID, proof.OperationKind, proof.FinalCommitID, proof.NormalizedRequestSHA256, proof.TerminalResultSHA256} {
		if !utf8.ValidString(value) {
			return nil, errors.New("invalid proof string encoding")
		}
	}
	// encoding/json is intentionally not a byte-boundary decoder: it repairs
	// malformed Unicode and accepts duplicate members. Reject those first.
	for _, raw := range []json.RawMessage{proof.IdempotencyIdentity, proof.TerminalResult, proof.ResourceRefs} {
		if _, err := canonicaljson.DecodeStrict(raw); err != nil {
			return nil, err
		}
	}
	proof.CommittedAt = proof.CommittedAt.UTC()
	return canonicaljson.Marshal(struct {
		SchemaID string `json:"schema_id"`
		CommitProof
	}{SchemaID: CommitProofSchema, CommitProof: proof})
}

// These two Extensions-owned scalars have the same closed ASCII grammar. Job
// UUIDs retain the separate Core scalar validator; deployment uniqueness stays
// under the database transaction's unique constraint.
func validCommitProofIdentifier(value string) bool {
	if len(value) < 1 || len(value) > 160 {
		return false
	}
	for index, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' {
			continue
		}
		if index == 0 || character != '_' && character != '.' && character != ':' && character != '-' {
			return false
		}
	}
	return true
}
