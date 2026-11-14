package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollaborationContractRejectsCompatibilityAndInvalidResultShapes(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repository root: %v", err)
	}
	load := func(relativePath string) map[string]any {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(root, "contracts", "collaboration", filepath.FromSlash(relativePath)))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		decoded, err := decodeContract(raw)
		if err != nil {
			t.Fatalf("decode %s: %v", relativePath, err)
		}
		object, err := asObject(decoded, relativePath)
		if err != nil {
			t.Fatalf("read %s as object: %v", relativePath, err)
		}
		return object
	}

	if err := validateCollaborationContractFamily(root); err != nil {
		t.Fatalf("canonical Collaboration contract family: %v", err)
	}
	if err := validateCollaborationResultFixture(load("fixtures/operator-requeue-result.v2.success.json"), true); err != nil {
		t.Fatalf("canonical success fixture: %v", err)
	}
	if err := validateCollaborationResultFixture(load("fixtures/operator-requeue-result.v2.failure.json"), false); err != nil {
		t.Fatalf("canonical failure fixture: %v", err)
	}
	for _, success := range []bool{false, true} {
		for _, tc := range []struct {
			name   string
			mutate func(map[string]any)
			valid  bool
		}{
			{"exact operation", func(value map[string]any) { value["operation"] = "other_operation" }, false},
			{"null operation", func(value map[string]any) { value["operation"] = nil }, false},
			{"null incident", func(value map[string]any) {
				value["incident_id"] = nil
				if !success {
					value["error"] = map[string]any{"code": "invalid_operator_request", "reason_code": "missing_required_flag", "message": "Incident ID is required."}
				}
			}, !success},
			{"invalid incident", func(value map[string]any) { value["incident_id"] = "invalid" }, false},
			{"equal times", func(value map[string]any) { value["completed_at"] = value["started_at"] }, true},
			{"invalid time", func(value map[string]any) { value["completed_at"] = "invalid" }, false},
			{"noncanonical UTC time", func(value map[string]any) { value["completed_at"] = "2026-08-06T17:00:01+00:00" }, false},
			{"reversed times", func(value map[string]any) {
				value["started_at"], value["completed_at"] = value["completed_at"], value["started_at"]
			}, false},
			{"fractional times", func(value map[string]any) {
				value["started_at"] = "2026-08-06T17:00:00Z"
				value["completed_at"] = "2026-08-06T17:00:00.1Z"
			}, true},
			{"reversed fractional times", func(value map[string]any) {
				value["started_at"] = "2026-08-06T17:00:00.1Z"
				value["completed_at"] = "2026-08-06T17:00:00Z"
			}, false},
		} {
			branch := "failure"
			if success {
				branch = "success"
			}
			t.Run(branch+"/"+tc.name, func(t *testing.T) {
				value := load("fixtures/operator-requeue-result.v2." + branch + ".json")
				tc.mutate(value)
				if err := validateCollaborationResultFixture(value, success); (err == nil) != tc.valid {
					t.Fatalf("result fixture admission = %v, want valid=%t", err, tc.valid)
				}
			})
		}
	}

	tests := []struct {
		name   string
		value  func() map[string]any
		verify func(map[string]any) error
	}{
		{
			name: "v1 schema has no reader",
			value: func() map[string]any {
				value := load("fixtures/operator-requeue-result.v2.success.json")
				value["schema_id"] = "cartulary.operator.collaboration_requeue_result.v1"
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationResultFixture(value, true) },
		},
		{
			name: "unknown result member",
			value: func() map[string]any {
				value := load("fixtures/operator-requeue-result.v2.success.json")
				value["legacy_result"] = true
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationResultFixture(value, true) },
		},
		{
			name: "success with null count",
			value: func() map[string]any {
				value := load("fixtures/operator-requeue-result.v2.success.json")
				value["requeued_intent_count"] = nil
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationResultFixture(value, true) },
		},
		{
			name: "failure with mismatched reason",
			value: func() map[string]any {
				value := load("fixtures/operator-requeue-result.v2.failure.json")
				value["error"].(map[string]any)["reason_code"] = "caller_cancelled"
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationResultFixture(value, false) },
		},
		{
			name: "historical reader admitted",
			value: func() map[string]any {
				value := load("index.json")
				value["historical_reader_schema_ids"] = []any{"cartulary.operator.collaboration_requeue_result.v1"}
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationIndex(value) },
		},
		{
			name: "parser alias admitted",
			value: func() map[string]any {
				value := load("index.json")
				value["compatibility_policy"].(map[string]any)["parser_aliases"] = true
				return value
			},
			verify: func(value map[string]any) error { return validateCollaborationIndex(value) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.verify(test.value()); err == nil {
				t.Fatal("invalid Collaboration contract mutation was admitted")
			}
		})
	}
}
