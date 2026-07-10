package rollbackprovider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
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
	if err := (TaskRequestProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"status": "invalid"}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("task error = %v", err)
	}
	if err := (DecisionProvider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"decision_type": "invalid"}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("decision error = %v", err)
	}
}
