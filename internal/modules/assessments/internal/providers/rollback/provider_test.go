package rollback

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValueExcludesDerivedConfidenceBand(t *testing.T) {
	t.Parallel()
	value := map[string]any{
		"source": map[string]any{
			"confidence_score": nil,
			"rationale":        "Retained",
		},
		"cells": map[string]any{
			"assessment.confidence_band":  map[string]any{"value": "high"},
			"assessment.confidence_score": map[string]any{"value": 99},
		},
	}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"confidence_score": nil, "rationale": "Retained"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("assessment source = %#v, want %#v", got, want)
	}
	if _, ok := sourceForRollbackValue(map[string]any{"cells": map[string]any{"assessment.rationale": map[string]any{"value": "legacy"}}}); ok {
		t.Fatal("schema-less projection row was accepted")
	}
}

func TestProviderRejectsInvalidOwnerInvariants(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	for _, score := range []json.Number{"0", "25", "100"} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{"confidence_score": score}}); err != nil {
			t.Fatalf("exact valid score %s: %v", score, err)
		}
	}
	for _, score := range []json.Number{"-1", "101", "0.5", "9223372036854775807"} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": map[string]any{"confidence_score": score}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("invalid exact score %s: %v", score, err)
		}
	}
	t.Run("precise score uses an integer SQL argument", func(t *testing.T) {
		tx := &assessmentRestoreTx{incidentID: uuid.New(), subjectID: uuid.New()}
		err := provider.RestoreTx(t.Context(), tx, rollbackcontract.RestoreRequest{
			RecordID: uuid.New(), RetainedValue: map[string]any{"source": map[string]any{"confidence_score": json.Number("25")}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if tx.queries != 2 || tx.writes != 1 || len(tx.args) != 16 || tx.args[7] != true || tx.args[8] != 25 {
			t.Fatalf("restoration SQL arguments: queries %d, writes %d, args %#v", tx.queries, tx.writes, tx.args)
		}
	})
	for _, value := range []map[string]any{
		{"source": map[string]any{"subject_type": "artifact"}},
		{"source": map[string]any{"subject_record_id": nil}},
		{"source": map[string]any{"confidence_score": float64(101)}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
}

// This double witnesses the driver argument boundary, not database execution.
type assessmentRestoreTx struct {
	pgx.Tx
	incidentID, subjectID uuid.UUID
	queries, writes       int
	args                  []any
}

type assessmentRestoreRow func(...any) error

func (row assessmentRestoreRow) Scan(dest ...any) error { return row(dest...) }

func (tx *assessmentRestoreTx) QueryRow(context.Context, string, ...any) pgx.Row {
	tx.queries++
	if tx.queries == 1 {
		return assessmentRestoreRow(func(dest ...any) error {
			*(dest[0].(*uuid.UUID)) = tx.incidentID
			*(dest[1].(*uuid.UUID)) = tx.subjectID
			*(dest[2].(*string)) = "host"
			return nil
		})
	}
	return assessmentRestoreRow(func(dest ...any) error { *(dest[0].(*bool)) = true; return nil })
}

func (tx *assessmentRestoreTx) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	tx.writes++
	tx.args = append([]any(nil), args...)
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
