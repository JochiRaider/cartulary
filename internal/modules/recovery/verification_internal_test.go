package recovery

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type restoreIncidentSelectionQuery struct {
	calls int
	query string
	row   pgx.Row
}

func (query *restoreIncidentSelectionQuery) QueryRow(_ context.Context, statement string, _ ...any) pgx.Row {
	query.calls++
	query.query = strings.Join(strings.Fields(statement), " ")
	return query.row
}

type restoreIncidentSelectionRow struct {
	incidentID string
	err        error
}

func (row restoreIncidentSelectionRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return errors.New("unexpected selection destination count")
	}
	destination, ok := destinations[0].(*string)
	if !ok {
		return errors.New("unexpected selection destination type")
	}
	*destination = row.incidentID
	return nil
}

func TestRestoreVerificationIncidentSelectionIsLexicographicAndSingleShot(t *testing.T) {
	query := &restoreIncidentSelectionQuery{row: restoreIncidentSelectionRow{incidentID: "00000000-0000-0000-0000-000000000401"}}
	selected, err := selectRestoreVerificationIncidentID(context.Background(), query)
	if err != nil {
		t.Fatalf("select restore verification incident: %v", err)
	}
	if selected == nil || *selected != "00000000-0000-0000-0000-000000000401" || query.calls != 1 {
		t.Fatalf("incident selection got selected=%v calls=%d", selected, query.calls)
	}
	wantQuery := "SELECT id::text FROM incidents ORDER BY id::text ASC LIMIT 1"
	if query.query != wantQuery {
		t.Fatalf("incident selection query got %q want %q", query.query, wantQuery)
	}

	t.Run("no incidents returns no selection", func(t *testing.T) {
		empty := &restoreIncidentSelectionQuery{row: restoreIncidentSelectionRow{err: pgx.ErrNoRows}}
		selected, emptyErr := selectRestoreVerificationIncidentID(context.Background(), empty)
		if emptyErr != nil || selected != nil || empty.calls != 1 {
			t.Fatalf("empty incident selection got selected=%v calls=%d error=%v", selected, empty.calls, emptyErr)
		}
	})

	t.Run("query failure remains distinct from no incidents", func(t *testing.T) {
		queryFailure := errors.New("restored database unavailable")
		failing := &restoreIncidentSelectionQuery{row: restoreIncidentSelectionRow{err: queryFailure}}
		selected, failingErr := selectRestoreVerificationIncidentID(context.Background(), failing)
		if selected != nil || !errors.Is(failingErr, queryFailure) || failing.calls != 1 {
			t.Fatalf("failed incident selection got selected=%v calls=%d error=%v", selected, failing.calls, failingErr)
		}
	})
}
