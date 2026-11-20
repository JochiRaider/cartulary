package timeline

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

type incidentResolvingRecords struct {
	RecordPort
	incidentID uuid.UUID
	err        error
}

func (r incidentResolvingRecords) ResolveIncident(context.Context, uuid.UUID) (uuid.UUID, error) {
	return r.incidentID, r.err
}

type revisionWindowRows struct {
	RevisionPort
	rows []conflicttokens.RevisionWindowRow
	err  error
}

func (r revisionWindowRows) ListRecordRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicttokens.RevisionWindowRow, error) {
	return r.rows, r.err
}

// TestTimelineUnclassifiedFailuresStayInternal_Unit verifies REQ-01-706 at
// Timeline's record-incident, revision-window and client-preview boundaries:
// only a missing record is not found and only an unusable window or base is
// row_version_conflict; a database or invariant failure stays unclassified
// and reaches internal_error.
func TestTimelineUnclassifiedFailuresStayInternal_Unit(t *testing.T) {
	ctx := context.Background()
	recordID := uuid.MustParse("6f5e4d3c-2b1a-4c9d-8e7f-0a1b2c3d4e01")
	incidentID := uuid.MustParse("6f5e4d3c-2b1a-4c9d-8e7f-0a1b2c3d4e02")
	injected := errors.New("injected database failure")

	t.Run("record incident", func(t *testing.T) {
		resolved, err := (&store{recordStore: incidentResolvingRecords{incidentID: incidentID}}).getRecordIncident(ctx, recordID)
		if err != nil || resolved != incidentID {
			t.Fatalf("resolved incident = %s, %v; want %s", resolved, err, incidentID)
		}
		if _, err := (&store{recordStore: incidentResolvingRecords{err: fmt.Errorf("classify: %w", records.ErrTargetNotFound)}}).getRecordIncident(ctx, recordID); !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("missing record = %v, want ErrRecordNotFound", err)
		}
		for _, failure := range []error{injected, pgx.ErrTxClosed, context.DeadlineExceeded} {
			_, err := (&store{recordStore: incidentResolvingRecords{err: fmt.Errorf("classify record target: %w", failure)}}).getRecordIncident(ctx, recordID)
			if errors.Is(err, ErrRecordNotFound) || !errors.Is(err, failure) {
				t.Fatalf("lookup failure %v = %v, want the unclassified failure", failure, err)
			}
		}
	})

	t.Run("patch conflict window", func(t *testing.T) {
		s := &store{conflictSnapshots: newTimelineConflictSnapshotProjector()}
		s.revisionsStore = revisionWindowRows{}
		_, err := s.loadPatchConflictWindowTx(ctx, nil, recordID, 2, 5)
		var conflict *RowVersionConflictError
		if !errors.As(err, &conflict) || *conflict != (RowVersionConflictError{RecordID: recordID, BaseRowVersion: 2, CurrentRowVersion: 5}) {
			t.Fatalf("missing history = %v, want the record's row version conflict", err)
		}
		s.revisionsStore = revisionWindowRows{err: injected}
		if _, err := s.loadPatchConflictWindowTx(ctx, nil, recordID, 2, 5); errors.As(err, &conflict) || !errors.Is(err, injected) {
			t.Fatalf("window read failure = %v, want the unclassified failure", err)
		}
		if err := adaptPatchConflictWindowError(injected); errors.As(err, &conflict) || !errors.Is(err, injected) {
			t.Fatalf("window build failure = %v, want the unclassified failure", err)
		}
	})

	t.Run("client conflict preview", func(t *testing.T) {
		var conflict *RowVersionConflictError
		for _, unusable := range []error{
			fmt.Errorf("%w for timeline.tags", errInvalidCollectionConflictBase),
			errInvalidMentionConflictBase,
		} {
			err := adaptClientConflictValueError(recordID, 2, 5, unusable)
			if !errors.As(err, &conflict) || *conflict != (RowVersionConflictError{RecordID: recordID, BaseRowVersion: 2, CurrentRowVersion: 5}) {
				t.Fatalf("unusable base %v = %v, want the record's row version conflict", unusable, err)
			}
		}
		invalid := &InvalidCollectionActionError{FieldKey: "timeline.tags"}
		if err := adaptClientConflictValueError(recordID, 2, 5, invalid); err != invalid {
			t.Fatalf("invalid action = %v, want it unchanged", err)
		}
		if err := adaptClientConflictValueError(recordID, 2, 5, injected); errors.As(err, &conflict) || !errors.Is(err, injected) {
			t.Fatalf("preview failure = %v, want the unclassified failure", err)
		}

		change := func(op string) PatchChange {
			return PatchChange{FieldKey: "timeline.tags", ActionPayload: &CollectionActionPayload{Actions: []CollectionAction{{Op: op, NormalizedText: "phishing"}}}}
		}
		if _, err := patchClientConflictValue(recordID, change("add_tag"), "not a collection", nil); !errors.Is(err, errInvalidCollectionConflictBase) {
			t.Fatalf("unusable base preview = %v, want errInvalidCollectionConflictBase", err)
		}
		collection := map[string]any{"kind": "collection_value_v1", "ordered": false, "items": []any{}}
		_, err := patchClientConflictValue(recordID, change("rename_tag"), collection, nil)
		if err == nil || errors.Is(err, errInvalidCollectionConflictBase) {
			t.Fatalf("action outside the admission contract = %v, want an invariant failure", err)
		}
		if err := adaptClientConflictValueError(recordID, 2, 5, err); errors.As(err, &conflict) {
			t.Fatalf("invariant failure classified as %v", err)
		}
	})
}
