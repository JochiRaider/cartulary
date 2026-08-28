package mutationvalue

import (
	"testing"

	"github.com/google/uuid"
)

func TestMutationRecursivelyIsolatesConstructionAndReaders(t *testing.T) {
	targetID := uuid.New().String()
	before := map[string]any{
		"nested": map[string]any{"items": []any{map[string]any{"value": "before"}}},
		"typed":  []map[string]any{{"value": "typed-before"}},
	}
	after := map[string]any{
		"nested": map[string]any{"items": []any{map[string]any{"value": "after"}}},
		"typed":  []map[string]any{{"value": "typed-after"}},
	}
	mutation, err := New(TargetRecordLink, targetID, OperationPatch, before, after)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	before["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] = "construction-corruption"
	after["typed"].([]map[string]any)[0]["value"] = "construction-corruption"

	readBefore := mutation.BeforeValue().(map[string]any)
	readAfter := mutation.AfterValue().(map[string]any)
	if readBefore["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] != "before" || readAfter["typed"].([]map[string]any)[0]["value"] != "typed-after" {
		t.Fatalf("construction inputs mutated retained value: before=%#v after=%#v", readBefore, readAfter)
	}
	readBefore["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] = "reader-corruption"
	readAfter["typed"].([]map[string]any)[0]["value"] = "reader-corruption"
	secondBefore := mutation.BeforeValue().(map[string]any)
	secondAfter := mutation.AfterValue().(map[string]any)
	if secondBefore["nested"].(map[string]any)["items"].([]any)[0].(map[string]any)["value"] != "before" || secondAfter["typed"].([]map[string]any)[0]["value"] != "typed-after" {
		t.Fatalf("reader values mutated retained value: before=%#v after=%#v", secondBefore, secondAfter)
	}

	copied := Copy([]Value{mutation})
	copied[0].BeforeValue().(map[string]any)["nested"] = "copy-corruption"
	if mutation.BeforeValue().(map[string]any)["nested"] == "copy-corruption" {
		t.Fatal("copied mutation shares mutable state")
	}
}

func TestMutationConstructorRejectsInvalidGrammar(t *testing.T) {
	linkID := uuid.New().String()
	recordID := uuid.New().String()
	tagID := uuid.New().String()
	validTagID := "record_tag:" + recordID + ":" + tagID
	validMap := map[string]any{"key": "value"}
	tests := []struct {
		name      string
		kind      string
		targetID  string
		operation string
		before    any
		after     any
	}{
		{"unknown target", "record", linkID, OperationPatch, validMap, validMap},
		{"invalid link id", TargetRecordLink, "invalid", OperationCreate, nil, validMap},
		{"invalid tag id", TargetRecordTag, "record_tag:invalid:" + tagID, OperationCreate, nil, validMap},
		{"unknown operation", TargetRecordLink, linkID, "rollback", validMap, validMap},
		{"create before", TargetRecordLink, linkID, OperationCreate, validMap, validMap},
		{"create missing after", TargetRecordLink, linkID, OperationCreate, nil, nil},
		{"patch missing before", TargetRecordLink, linkID, OperationPatch, nil, validMap},
		{"delete missing after", TargetRecordTag, validTagID, OperationDelete, validMap, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.kind, test.targetID, test.operation, test.before, test.after); err == nil {
				t.Fatal("invalid mutation grammar accepted")
			}
		})
	}
}
