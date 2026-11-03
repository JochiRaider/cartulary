package crossownertransaction

import (
	"encoding/json"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/validation"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// ValidationPolicy is supplied by composition from the admitted owner registry.
// The coordinator retains its participant contract and transaction ownership.
type ValidationPolicy interface {
	AcceptDescriptor(Descriptor) bool
	AcceptInput(Descriptor, any) bool
	AcceptFinding(Descriptor, Finding) bool
	FailureFinding(string, Descriptor, int64, bool) (Finding, bool)
}

type ValidationError struct {
	ReasonCode string
	Findings   []Finding
}

func (e *ValidationError) Error() string { return e.ReasonCode }
func (e *ValidationError) Unwrap() error { return ErrValidation }

func (c *Coordinator) classifyValidation(descriptor Descriptor, result ValidationResult, invocationErr error, timeout time.Duration) error {
	normalized := make([]Finding, 0)
	disposition := validation.Classify(invocationErr, result.Findings != nil, len(result.Findings), c.limits.ValidationFindings, func() bool {
		if (result.Status != "valid" && result.Status != "invalid") || (result.Status == "valid") != (len(result.Findings) == 0) {
			return false
		}
		for index, finding := range result.Findings {
			// Details are the actual participant byte boundary. Reject duplicate
			// members and malformed Unicode before schema or canonical ordering.
			details, err := canonicaljson.DecodeStrict(finding.Details)
			if err != nil {
				return false
			}
			if _, ok := details.(map[string]any); !ok {
				return false
			}
			canonical, err := canonicaljson.MarshalExtension(details)
			if err != nil {
				return false
			}
			finding.Details = append([]byte(nil), canonical[:len(canonical)-1]...)
			if !c.validationPolicy.AcceptFinding(descriptor, finding) {
				return false
			}
			if index > 0 && compareFinding(normalized[index-1], finding) >= 0 {
				return false
			}
			normalized = append(normalized, finding)
		}
		canonical, err := canonicaljson.MarshalExtension(map[string]any{
			"schema_id": ValidationResultSchema, "participant_id": descriptor.ParticipantID,
			"status": result.Status, "findings": normalized,
		})
		_, withinLimit := boundedTotal(0, int64(len(canonical)), c.limits.ResultBytes, c.limits.ResultBytes)
		return err == nil && withinLimit
	})
	condition := "extension_validation_remaining_schema_defect_transaction"
	switch disposition {
	case validation.Success:
		return nil
	case validation.Findings:
		return &ValidationError{ReasonCode: ErrValidation.Error(), Findings: normalized}
	case validation.InvocationFailed:
		condition = "extension_participant_validation_invocation_failure"
	case validation.StructureInvalid:
		condition = "extension_validation_result_structural_invalid_transaction"
	case validation.Overflow:
		condition = "extension_validation_result_overflow_transaction"
	}
	// Coordinator cancellation/deadline outcomes stop the protocol before this
	// result classifier. An owner-returned error cannot prove budget expiry.
	finding, registered := c.validationPolicy.FailureFinding(condition, descriptor, int64(timeout/time.Second), false)
	if !registered {
		finding = Finding{Path: "$", ReasonCode: "extension_validation_result_invalid", Message: "An extension validation algorithm returned an invalid result.", Details: []byte(`{"actual":"redacted","algorithm_id":null,"phase":"transaction","profile_id":null}`)}
	}
	return &ValidationError{ReasonCode: finding.ReasonCode, Findings: []Finding{finding}}
}

// Ensure the exact four-member finding is serialized, including JSON details
// rather than a base64 representation of their bytes.
func (f Finding) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Path       string          `json:"path"`
		ReasonCode string          `json:"reason_code"`
		Message    string          `json:"message"`
		Details    json.RawMessage `json:"details"`
	}{f.Path, f.ReasonCode, f.Message, f.Details})
}

// Owner receipts travel directly to the finalizer. They are not members of the
// closed participant write-result envelope and are never serialized here.
func (c *Coordinator) validWriteResult(descriptor Descriptor, result WriteResult) bool {
	if result.Status != "written" {
		return false
	}
	canonical, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id": WriteResultSchema, "participant_id": descriptor.ParticipantID, "status": result.Status,
	})
	_, withinLimit := boundedTotal(0, int64(len(canonical)), c.limits.ResultBytes, c.limits.ResultBytes)
	return err == nil && withinLimit
}

func prepareKeyObjects(keys []OrderedSerializationKey) []any {
	rows := make([]any, len(keys))
	for i, key := range keys {
		rows[i] = map[string]any{"participant_id": key.ParticipantID, "key_kind": key.KeyKind, "key": key.Key}
	}
	return rows
}
