package parties

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

// TestPartyRevisionWindowClassification_Unit verifies REQ-01-706 at the
// Parties conflict-window boundary: only the typed unavailable-window
// condition fails closed as row_version_conflict; a database, field-catalog or
// token failure passes through unclassified and reaches internal_error.
func TestPartyRevisionWindowClassification_Unit(t *testing.T) {
	recordID := uuid.MustParse("9c4b2a1d-6e5f-4a3b-8c2d-1e0f9a8b7c01")
	windowRecordID := uuid.MustParse("9c4b2a1d-6e5f-4a3b-8c2d-1e0f9a8b7c02")

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
	for _, injected := range []error{
		errors.New("injected revision window database failure"),
		pgx.ErrTxClosed,
		context.Canceled,
		&pgconn.PgError{Code: "40P01", Message: "deadlock detected"},
		fmt.Errorf("issue conflict token: %w", errors.New("injected token failure")),
	} {
		err := adaptRevisionWindowError(recordID, 3, 6, injected)
		var unexpected *RowVersionConflictError
		if errors.As(err, &unexpected) || !errors.Is(err, injected) {
			t.Fatalf("adapted %v = %v, want the unclassified failure", injected, err)
		}
	}
}
