package tasksdecisions

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

// TestTaskDecisionRevisionWindowClassification_Unit verifies REQ-01-706 at the
// Tasks and Decisions conflict-window boundary: only the typed
// unavailable-window condition fails closed as row_version_conflict; a
// database, configuration, token or invariant failure passes through
// unclassified and reaches internal_error.
func TestTaskDecisionRevisionWindowClassification_Unit(t *testing.T) {
	recordID := uuid.MustParse("5b0f9d1e-2c3a-4e5f-8a6b-7c8d9e0f1a01")
	windowRecordID := uuid.MustParse("5b0f9d1e-2c3a-4e5f-8a6b-7c8d9e0f1a02")

	t.Run("unavailable window is a row version conflict", func(t *testing.T) {
		window := &conflicttokens.RevisionWindowError{RecordID: windowRecordID, BaseRowVersion: 2, CurrentRowVersion: 5}
		var conflict *RowVersionConflictError
		if err := adaptRevisionWindowError(recordID, 3, 6, fmt.Errorf("wrapped: %w", window)); !errors.As(err, &conflict) {
			t.Fatalf("adapted window error = %v, want RowVersionConflictError", err)
		}
		if want := (RowVersionConflictError{RecordID: windowRecordID, BaseRowVersion: 2, CurrentRowVersion: 5}); *conflict != want {
			t.Fatalf("row version conflict = %+v, want %+v", *conflict, want)
		}
		if err := adaptRevisionWindowError(recordID, 3, 6, nil); err != nil {
			t.Fatalf("adapted nil error = %v", err)
		}
	})

	t.Run("unclassified failures pass through", func(t *testing.T) {
		for _, injected := range []error{
			errors.New("injected revision window database failure"),
			pgx.ErrTxClosed,
			context.Canceled,
			&pgconn.PgError{Code: "40P01", Message: "deadlock detected"},
			conflicttokens.ErrInvalidSnapshotProjector,
		} {
			err := adaptRevisionWindowError(recordID, 3, 6, injected)
			var conflict *RowVersionConflictError
			if errors.As(err, &conflict) || !errors.Is(err, injected) {
				t.Fatalf("adapted %v = %v, want the unclassified failure", injected, err)
			}
		}
	})

	t.Run("same-field preview classifies only an unusable base", func(t *testing.T) {
		row := func(value any) map[string]any {
			return map[string]any{"cells": map[string]any{"task.related_record_ids": map[string]any{"value": value}}}
		}
		collection := map[string]any{"kind": "collection_value_v1", "ordered": false, "items": []any{}}
		linked := uuid.MustParse("5b0f9d1e-2c3a-4e5f-8a6b-7c8d9e0f1a03")
		params := func(base any, action CollectionAction) sameFieldConflictParams {
			return sameFieldConflictParams{
				RecordID: recordID, BaseRowVersion: 3, CurrentRowVersion: 6,
				Window:     conflicttokens.PatchConflictWindow{BaseRow: row(base)},
				Change:     PatchChange{FieldKey: "task.related_record_ids", Collection: &CollectionActionPayload{Actions: []CollectionAction{action}}},
				CurrentRow: row(collection),
			}
		}
		_, err := buildSameFieldConflict(params("not a collection", CollectionAction{Op: "add_record_ref", LinkedRecordID: &linked}))
		var window *conflicttokens.RevisionWindowError
		if !errors.As(err, &window) || window.RecordID != recordID || window.BaseRowVersion != 3 || window.CurrentRowVersion != 6 {
			t.Fatalf("unusable base = %v, want the record's RevisionWindowError", err)
		}
		for _, action := range []CollectionAction{{Op: "add_record_ref"}, {Op: "rename_record_ref"}} {
			_, err = buildSameFieldConflict(params(collection, action))
			if err == nil || errors.As(err, &window) {
				t.Fatalf("invariant failure for %+v = %v, want an unclassified failure", action, err)
			}
		}
	})
}
