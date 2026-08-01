package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

func TestTaskSourcePreservesNullAndExcludesCollections(t *testing.T) {
	t.Parallel()
	value := map[string]any{"cells": map[string]any{
		"task.title":              map[string]any{"value": "Collect logs"},
		"task.decision_record_id": map[string]any{"value": nil},
		"task.linked_record_ids":  map[string]any{"value": []any{"ignored"}},
	}}
	got, ok := taskSourceForRollbackValue(value)
	if !ok {
		t.Fatal("taskSourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"title": "Collect logs", "decision_record_id": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("task source = %#v, want %#v", got, want)
	}
}

func TestProvidersRejectInvalidOwnerValues(t *testing.T) {
	t.Parallel()
	if policy.DefaultTaskStatus != "open" || policy.DefaultTaskPriority != "normal" || policy.DefaultDecisionStatus != "proposed" {
		t.Fatal("shared source defaults changed")
	}
	for _, status := range []string{"open", "in_progress", "blocked", "done", "canceled"} {
		if !policy.ValidTaskStatus(status) {
			t.Fatalf("policy rejected task status %q", status)
		}
		if err := (TaskRequestProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"status": status}}); err != nil {
			t.Fatalf("rollback rejected policy task status %q: %v", status, err)
		}
	}
	for _, status := range []string{"proposed", "approved", "rejected", "superseded", "executed"} {
		if !policy.ValidDecisionStatus(status) {
			t.Fatalf("policy rejected decision status %q", status)
		}
		if err := (DecisionProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"status": status}}); err != nil {
			t.Fatalf("rollback rejected policy decision status %q: %v", status, err)
		}
	}
	if err := (TaskRequestProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"status": "invalid"}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("task error = %v", err)
	}
	if err := (DecisionProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"decision_type": "invalid"}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("decision error = %v", err)
	}
}
