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
