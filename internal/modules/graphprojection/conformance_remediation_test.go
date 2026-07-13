package graphprojection

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestMalformedJSONIsNotMisclassifiedAsDuplicateMember(t *testing.T) {
	_, err := admitProjectionInput([]byte(`{"projection_schema_id":`), admitOptions{Operation: "project_ephemeral"})
	var operationError *LifecycleError
	if !errors.As(err, &operationError) || operationError.ReasonCode != "invalid_json_syntax" {
		t.Fatalf("malformed JSON error = %#v", err)
	}
	if len(operationError.Details) != 4 || operationError.Details["operation"] != "project_ephemeral" || operationError.Details["field"] != nil || operationError.Details["validation_code"] != nil {
		t.Fatalf("malformed JSON details = %#v", operationError.Details)
	}
}

func TestGraphViewIDMismatchDoesNotExposeDerivedIdentity(t *testing.T) {
	input := minimalInput(t, "mismatched-identity")
	input["graph_view_id"] = "gv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err := admitProjectionInput(mustJSON(t, input), admitOptions{Operation: "create_projection"})
	var operationError *LifecycleError
	if !errors.As(err, &operationError) || operationError.ReasonCode != "invalid_graph_view_id" {
		t.Fatalf("graph view mismatch error = %#v", err)
	}
	encoded, marshalErr := json.Marshal(operationError.Details)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if strings.Contains(string(encoded), "expected") || strings.Contains(string(encoded), "mismatched-identity") {
		t.Fatalf("graph view mismatch leaked derived/source value: %s", encoded)
	}
}

type virtualProjectionInput struct {
	length int
	read   bool
}

func (input *virtualProjectionInput) Len() int { return input.length }

func (input *virtualProjectionInput) Bytes() []byte {
	input.read = true
	return nil
}

func TestAdmissionRejectsVirtualOversizedInputBeforeRead(t *testing.T) {
	input := &virtualProjectionInput{length: graphProjectionLimits.MaxInputBytes + 1}
	_, err := admitProjectionInputSource(input, admitOptions{Operation: "project_ephemeral"})
	if input.read {
		t.Fatal("oversized input bytes were read")
	}
	var operationError *LifecycleError
	if !errors.As(err, &operationError) || operationError.Code != "invalid_projection_request" || operationError.ReasonCode != "whole_input_limit_exceeded" {
		t.Fatalf("oversized admission error = %#v", err)
	}
	if operationError.Details["operation"] != "project_ephemeral" || operationError.Details["validation_code"] != "resource_limit_exceeded" {
		t.Fatalf("oversized admission details = %#v", operationError.Details)
	}
}

func TestAdmissionRejectsNestedUnknownMemberWithClosedDetails(t *testing.T) {
	input := minimalInput(t, "nested-unknown")
	input["projection_config"].(map[string]any)["retention_policy"] = map[string]any{"retention_count": 2, "private_knob": true}
	_, err := admitProjectionInput(mustJSON(t, input), admitOptions{Operation: "refresh_projection"})
	var operationError *LifecycleError
	if !errors.As(err, &operationError) {
		t.Fatalf("nested unknown error = %T %v", err, err)
	}
	if operationError.ReasonCode != "unknown_member" || operationError.Field != "$.projection_config.retention_policy.private_knob" {
		t.Fatalf("nested unknown error = %#v", operationError)
	}
	if operationError.Details["operation"] != "refresh_projection" || operationError.Details["validation_code"] != nil {
		t.Fatalf("nested unknown details = %#v", operationError.Details)
	}
}

func TestRepresentableScalarViolationIsAdmittedThenFails(t *testing.T) {
	input := minimalInput(t, "scalar-admitted")
	input["source_snapshot_id"] = " invalid"
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "scalar-nonce", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatalf("representable scalar admission: %v", err)
	}
	if run.ProjectionRunID == "" || run.ProjectionConfigDigest == "" || run.ProjectionSourceDigest == "" {
		t.Fatalf("admitted identity not fixed: %#v", run)
	}
	if run.State != RunStateFailed || run.GraphView != nil || len(run.ValidationSummary.Issues) != 1 {
		t.Fatalf("scalar violation result = %#v", run)
	}
	issue := run.ValidationSummary.Issues[0]
	if issue.Code != "invalid_input_shape" || issue.Field == nil || *issue.Field != "$.source_snapshot_id" || issue.Details["reason_code"] != "scalar_contract_violation" {
		t.Fatalf("scalar violation issue = %#v", issue)
	}
}

func TestDuplicateIdentifiersUseOwnerCodeAndStableOrder(t *testing.T) {
	input := incidentGraphInput(t)
	entities := input["source_entities"].([]any)
	entities = append(entities, map[string]any{"source_entity_id": "host1", "source_entity_kind": "host"})
	input["source_entities"] = entities
	run, err := project(mustJSON(t, input), projectOptions{ProjectionRunNonce: "duplicate-nonce", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatalf("duplicate identifier projection: %v", err)
	}
	if run.State != RunStateFailed || len(run.ValidationSummary.Issues) == 0 {
		t.Fatalf("duplicate result = %#v", run)
	}
	issue := run.ValidationSummary.Issues[0]
	if issue.Code != "duplicate_identifier" || issue.TargetKind != "projection_input" || issue.Details["collection"] != "$.source_entities" || issue.Details["identifier_value"] != "host1" {
		t.Fatalf("duplicate issue = %#v", issue)
	}
}

func TestProjectedSystemMetadataFieldPathRejected(t *testing.T) {
	if validFieldPath("projected.metadata.mapping_rule_id") {
		t.Fatal("system metadata field path was accepted")
	}
	if !validFieldPath("projected.metadata.owner") {
		t.Fatal("mapped metadata field path was rejected")
	}
}

func TestResourceRegistryCollectionOverflow(t *testing.T) {
	run := ProjectionRun{GraphViewID: "gv_test", ProjectionRunID: "gpr_test"}
	run.Request.SourceEntities = make([]SourceEntity, graphProjectionLimits.MaxSourceEntities+1)
	issues := admittedResourceLimitIssues(run)
	if len(issues) != 1 || issues[0].Code != "resource_limit_exceeded" || issues[0].Details["limit_key"] != "max_source_entities" {
		t.Fatalf("resource issues = %#v", issues)
	}
}

func TestIdempotencyKeyUsesUnicodeScalarContract(t *testing.T) {
	if err := validateIdempotencyKey("create_projection", strings.Repeat("é", 128)); err != nil {
		t.Fatalf("128 scalar idempotency key rejected: %v", err)
	}
	for _, value := range []string{" key", "key ", strings.Repeat("é", 129), "key\u0085value"} {
		err := validateIdempotencyKey("create_projection", value)
		wantReason := "invalid_value"
		if len([]rune(value)) > 128 {
			wantReason = "out_of_bounds"
		}
		if !IsLifecycleError(err, "invalid_operation_request", wantReason) {
			t.Fatalf("invalid key %q error = %v", value, err)
		}
	}
}

func TestIdempotencyPresenceDistinguishesOmittedNullAndEmpty(t *testing.T) {
	if value, err := resolveIdempotencyKey("create_projection", Optional[string]{}); err != nil || value != "" {
		t.Fatalf("omitted idempotency = %q, %v", value, err)
	}
	if _, err := resolveIdempotencyKey("create_projection", ExplicitNull[string]()); !IsLifecycleError(err, "invalid_operation_request", "explicit_null_not_allowed") {
		t.Fatalf("explicit null idempotency error = %v", err)
	}
	if _, err := resolveIdempotencyKey("create_projection", ValueOf("")); !IsLifecycleError(err, "invalid_operation_request", "out_of_bounds") {
		t.Fatalf("empty idempotency error = %v", err)
	}
}
