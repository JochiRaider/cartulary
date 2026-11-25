package lockwaits

import (
	"errors"
	"testing"
	"time"
)

type fakeCursor struct {
	rows [][]any
	err  error
}

func (c *fakeCursor) Next() bool { return len(c.rows) > 0 }
func (c *fakeCursor) Err() error { return c.err }
func (c *fakeCursor) Scan(dest ...any) error {
	row := c.rows[0]
	c.rows = c.rows[1:]
	for i, value := range row {
		switch target := dest[i].(type) {
		case *int64:
			*target = value.(int64)
		case *string:
			*target = value.(string)
		case *int:
			*target = value.(int)
		}
	}
	return nil
}

// TestCollectKeepsOnlyServerVocabulary covers the Go side of the closed
// vocabulary: server names outside it, and any statement kind the query did not
// reduce, become "other". The query's own keyword reduction is covered against
// a real database by the stall-diagnostics and HTTP lock-wait tests.
func TestCollectKeepsOnlyServerVocabulary(t *testing.T) {
	rows := &fakeCursor{rows: [][]any{
		{int64(11), "client backend", "active", "Lock", "transactionid", "update", "12,13", int64(3100), 70},
		{int64(12), "client backend", "idle in transaction", "Client", "ClientRead", "none", "", int64(9000), 70},
		{int64(13), "client backend; DROP", "active", "Lock", "relation\nprivate", "Select private", "", int64(0), 70},
	}}
	snapshot, err := collect(rows, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Blocked() != 1 || snapshot.OmittedSessionCount != 67 || snapshot.CapturedAt.Location() != time.UTC {
		t.Fatalf("snapshot counts = blocked %d omitted %d", snapshot.Blocked(), snapshot.OmittedSessionCount)
	}
	first, second, third := snapshot.Sessions[0], snapshot.Sessions[1], snapshot.Sessions[2]
	if first.StatementKind != "update" || len(first.BlockingPIDs) != 2 || first.BlockingPIDs[1] != 13 {
		t.Fatalf("blocked session = %#v", first)
	}
	if second.StatementKind != "none" || len(second.BlockingPIDs) != 0 || second.State != "idle in transaction" {
		t.Fatalf("blocking session = %#v", second)
	}
	if third.BackendType != "other" || third.WaitEvent != "other" || third.StatementKind != "other" {
		t.Fatalf("values outside the vocabulary must be other: %#v", third)
	}
	if _, err := collect(&fakeCursor{err: errors.New("cursor failed")}, time.Now()); err == nil {
		t.Fatal("a failed cursor produced a snapshot")
	}
}
