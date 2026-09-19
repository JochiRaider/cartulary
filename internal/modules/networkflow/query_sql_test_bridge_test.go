package networkflow

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

// Called by the authenticated pagination integration fixture using its live DB.
func AssertLiveSQLKeysetBoundaries(t *testing.T, store *store, incident uuid.UUID, table string) {
	t.Helper()
	for _, tc := range []struct {
		field string
		want  []int64
	}{
		{fieldBytesCount, []int64{2, 3, 1}}, {fieldSrcPort, []int64{1, 2, 3}},
	} {
		var position *rowCursorPosition
		var got []int64
		sort := []sortSpec{{FieldKey: tc.field, Direction: "asc"}}
		for {
			rows, more, err := store.QueryRowsPage(context.Background(), incident, []string{table}, nil, sort, position, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 {
				t.Fatalf("expected one row: %v", rows)
			}
			got = append(got, rows[0].SourceRowNumber)
			next := newRowCursorPosition(rows[0], sort)
			position = &next
			if !more {
				break
			}
			if len(got) > 3 {
				t.Fatal("keyset did not advance")
			}
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("SQL %s ordering: %v want %v", tc.field, got, tc.want)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := store.QueryRowsPage(ctx, incident, []string{table}, nil, nil, nil, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("query ignored cancellation: %v", err)
	}
}
