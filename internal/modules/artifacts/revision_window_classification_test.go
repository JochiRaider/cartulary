package artifacts

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

// TestArtifactRevisionWindowClassification_Unit verifies REQ-01-706 at the
// Artifacts conflict-window boundary: only the typed unavailable-window
// condition fails closed as row_version_conflict; a database, token or
// invariant failure passes through unclassified and reaches internal_error.
func TestArtifactRevisionWindowClassification_Unit(t *testing.T) {
	recordID := uuid.MustParse("7a1c3f52-4d0b-4b0e-9a55-1f6f2a9b3c01")
	windowRecordID := uuid.MustParse("7a1c3f52-4d0b-4b0e-9a55-1f6f2a9b3c02")

	t.Run("unavailable window is a row version conflict", func(t *testing.T) {
		window := &conflicts.RevisionWindowError{RecordID: windowRecordID, BaseRowVersion: 2, CurrentRowVersion: 5}
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
			fmt.Errorf("issue conflict token: %w", errors.New("injected token failure")),
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
			return map[string]any{"cells": map[string]any{"note.tags": map[string]any{"value": value}}}
		}
		collection := map[string]any{"kind": "collection_value_v1", "ordered": false, "items": []any{}}
		params := func(base any, op string) artifactSameFieldConflictParams {
			return artifactSameFieldConflictParams{
				RecordID: recordID, BaseRowVersion: 3, CurrentRowVersion: 6,
				Window:     conflicts.PatchConflictWindow{BaseRow: row(base)},
				Change:     patchChange{FieldKey: "note.tags", Collection: &collectionActionPayload{Actions: []collectionAction{{Op: op, NormalizedText: "phishing"}}}},
				CurrentRow: row(collection),
			}
		}
		_, err := buildArtifactSameFieldConflict(params("not a collection", "add_tag"))
		var window *conflicts.RevisionWindowError
		if !errors.As(err, &window) || window.RecordID != recordID || window.BaseRowVersion != 3 || window.CurrentRowVersion != 6 {
			t.Fatalf("unusable base = %v, want the record's RevisionWindowError", err)
		}
		_, err = buildArtifactSameFieldConflict(params(collection, "rename_tag"))
		if err == nil || errors.As(err, &window) {
			t.Fatalf("action outside the admission contract = %v, want an unclassified failure", err)
		}
	})
}
