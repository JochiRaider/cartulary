package hostidentity

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

// TestEntityRevisionWindowClassification_Unit verifies that an entity patch
// whose stale window cannot be reconstructed fails closed as
// row_version_conflict, like every other record owner (REQ-01-070), while
// every other rebase failure keeps its own classification or stays
// unclassified (REQ-01-706).
func TestEntityRevisionWindowClassification_Unit(t *testing.T) {
	recordID := uuid.MustParse("1d2c3b4a-5e6f-4a7b-8c9d-0e1f2a3b4c01")

	window := &conflicts.RevisionWindowError{RecordID: recordID, BaseRowVersion: 2, CurrentRowVersion: 5}
	var conflict *RowVersionConflictError
	if err := adaptRevisionWindowError(fmt.Errorf("wrapped: %w", window)); !errors.As(err, &conflict) {
		t.Fatalf("adapted window error = %v, want RowVersionConflictError", err)
	}
	if want := (RowVersionConflictError{RecordID: recordID, BaseRowVersion: 2, CurrentRowVersion: 5}); *conflict != want {
		t.Fatalf("row version conflict = %+v, want %+v", *conflict, want)
	}
	if err := adaptRevisionWindowError(nil); err != nil {
		t.Fatalf("adapted nil error = %v", err)
	}

	sameField := &SameFieldConflictError{Conflict: SameFieldConflict{RecordID: recordID, FieldKey: "host.display_name"}}
	for _, other := range []error{
		errors.New("injected revision window database failure"),
		pgx.ErrTxClosed,
		context.Canceled,
		sameField,
		ErrInvalidAliasReference,
	} {
		err := adaptRevisionWindowError(other)
		var unexpected *RowVersionConflictError
		if errors.As(err, &unexpected) || err != other {
			t.Fatalf("adapted %v = %v, want it unchanged", other, err)
		}
	}
}
