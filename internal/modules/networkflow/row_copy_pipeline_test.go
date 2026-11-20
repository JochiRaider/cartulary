package networkflow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// pipelineFakeTx records COPY statements and source reads on one connection.
// It fails the test when a source read overlaps a COPY statement, which a real
// connection would reject as busy.
type pipelineFakeTx struct {
	pgx.Tx
	t          *testing.T
	mu         sync.Mutex
	copying    bool
	statements [][]int64
	reads      int
	failAfter  int // fail the COPY after this many rows in total; 0 disables
	consumed   int
	aborted    []error
}

func (tx *pipelineFakeTx) CopyFrom(_ context.Context, table pgx.Identifier, _ []string, source pgx.CopyFromSource) (int64, error) {
	tx.mu.Lock()
	if tx.copying {
		tx.mu.Unlock()
		tx.t.Error("overlapping COPY statements")
		return 0, errors.New("conn busy")
	}
	tx.copying = true
	tx.mu.Unlock()
	defer func() {
		tx.mu.Lock()
		tx.copying = false
		tx.mu.Unlock()
	}()
	if len(table) != 1 || table[0] != "network_flow_rows" {
		tx.t.Errorf("COPY target = %v", table)
	}
	var statement []int64
	for source.Next() {
		values, err := source.Values()
		if err != nil {
			return 0, err
		}
		statement = append(statement, values[0].(int64))
		tx.mu.Lock()
		tx.consumed++
		fail := tx.failAfter > 0 && tx.consumed >= tx.failAfter
		tx.mu.Unlock()
		if fail {
			return 0, errors.New("check_violation")
		}
	}
	if err := source.Err(); err != nil {
		// pgx sends CopyFail: the statement's rows never commit.
		tx.mu.Lock()
		tx.aborted = append(tx.aborted, err)
		tx.mu.Unlock()
		return 0, err
	}
	tx.mu.Lock()
	tx.statements = append(tx.statements, statement)
	tx.mu.Unlock()
	return int64(len(statement)), nil
}

type pipelineFakeRow struct{ data []byte }

func (r pipelineFakeRow) Scan(dest ...any) error {
	*(dest[0].(*[]byte)) = r.data
	return nil
}

func (tx *pipelineFakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.copying {
		tx.t.Error("source read overlapped a COPY statement")
	}
	tx.reads++
	return pipelineFakeRow{data: []byte("chunk")}
}

func copiedRows(statements [][]int64) []int64 {
	var rows []int64
	for _, statement := range statements {
		rows = append(rows, statement...)
	}
	return rows
}

func TestNetworkFlowPipelinedRowCopy(t *testing.T) {
	ctx := context.Background()
	t.Run("rows_copy_in_order_and_source_reads_end_the_statement", func(t *testing.T) {
		tx := &pipelineFakeTx{t: t}
		err := newRowCopyPipeline().run(ctx, tx, []string{"source_row_number"}, func(gate rowSourceGate) error {
			pipeline := gate.(*rowCopyPipeline)
			for row := int64(1); row <= 1000; row++ {
				if row%250 == 0 {
					// Wait until COPY consumes this range so the read must end
					// an open statement rather than precede it.
					waitForPipelineCopy(t, pipeline)
					if err := gate.withConnection(func() error {
						var chunk []byte
						return tx.QueryRow(ctx, "SELECT chunk").Scan(&chunk)
					}); err != nil {
						return err
					}
				}
				if err := pipeline.push([]any{row}, 64); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		rows := copiedRows(tx.statements)
		if len(rows) != 1000 {
			t.Fatalf("copied %d rows, want 1000", len(rows))
		}
		for i, row := range rows {
			if row != int64(i+1) {
				t.Fatalf("row %d copied at position %d", row, i)
			}
		}
		if tx.reads != 4 || len(tx.statements) < 4 {
			t.Fatalf("reads=%d statements=%d; each source read must end the open COPY statement", tx.reads, len(tx.statements))
		}
	})
	t.Run("producer_failure_aborts_the_open_statement_and_keeps_its_classification", func(t *testing.T) {
		tx := &pipelineFakeTx{t: t}
		err := newRowCopyPipeline().run(ctx, tx, []string{"source_row_number"}, func(gate rowSourceGate) error {
			pipeline := gate.(*rowCopyPipeline)
			for row := int64(1); row <= 300; row++ {
				if err := pipeline.push([]any{row}, 64); err != nil {
					return err
				}
			}
			return errSourceChanged
		})
		if !errors.Is(err, errSourceChanged) {
			t.Fatalf("error = %v, want the producer's source-change classification", err)
		}
	})
	t.Run("copy_failure_stops_the_producer_and_is_reported", func(t *testing.T) {
		tx := &pipelineFakeTx{t: t, failAfter: 10}
		produced := 0
		err := newRowCopyPipeline().run(ctx, tx, []string{"source_row_number"}, func(gate rowSourceGate) error {
			pipeline := gate.(*rowCopyPipeline)
			for row := int64(1); row <= 1_000_000; row++ {
				if err := pipeline.push([]any{row}, 1<<20); err != nil {
					return err
				}
				produced++
			}
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "insert network flow row batch: check_violation") {
			t.Fatalf("error = %v, want the wrapped COPY failure", err)
		}
		if produced >= 1_000_000 {
			t.Fatal("producer was not stopped by the COPY failure")
		}
	})
	t.Run("a_source_read_after_copy_failure_fails_without_touching_the_connection", func(t *testing.T) {
		tx := &pipelineFakeTx{t: t, failAfter: 1}
		reads := make(chan error, 1)
		err := newRowCopyPipeline().run(ctx, tx, []string{"source_row_number"}, func(gate rowSourceGate) error {
			pipeline := gate.(*rowCopyPipeline)
			if err := pipeline.push([]any{int64(1)}, 64); err != nil {
				return err
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				pipeline.mu.Lock()
				failed := pipeline.consumerErr != nil
				pipeline.mu.Unlock()
				if failed || time.Now().After(deadline) {
					break
				}
				time.Sleep(time.Millisecond)
			}
			readErr := gate.withConnection(func() error {
				var chunk []byte
				return tx.QueryRow(ctx, "SELECT chunk").Scan(&chunk)
			})
			reads <- readErr
			return readErr
		})
		if err == nil || !strings.Contains(err.Error(), "check_violation") {
			t.Fatalf("error = %v, want the COPY failure", err)
		}
		if readErr := <-reads; readErr == nil || tx.reads != 0 {
			t.Fatalf("source read after COPY failure: err=%v reads=%d", readErr, tx.reads)
		}
	})
	t.Run("cancellation_ends_without_deadlock", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(ctx)
		tx := &pipelineFakeTx{t: t}
		done := make(chan error, 1)
		go func() {
			done <- newRowCopyPipeline().run(cancelled, tx, []string{"source_row_number"}, func(gate rowSourceGate) error {
				pipeline := gate.(*rowCopyPipeline)
				for row := int64(1); ; row++ {
					if err := cancelled.Err(); err != nil {
						return err
					}
					if err := pipeline.push([]any{row}, 64); err != nil {
						return err
					}
					if row == 100 {
						cancel()
					}
				}
			})
		}()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v, want cancellation", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("pipeline did not end after cancellation")
		}
	})
	t.Run("read_ahead_preserves_bytes_then_error", func(t *testing.T) {
		pipeline := newRowCopyPipeline()
		pipeline.ctx = ctx
		payload := bytes.Repeat([]byte("0123456789"), 10_000)
		failure := errors.New("source_changed")
		reader := pipeline.readAhead(io.MultiReader(&chunkedReader{data: payload, chunk: 777}, errorReader{failure}), 4096)
		got, err := io.ReadAll(reader)
		if !errors.Is(err, failure) || !bytes.Equal(got, payload) {
			t.Fatalf("read-ahead returned %d bytes and %v, want %d bytes then the source error", len(got), err, len(payload))
		}
		if pipeline.connHeld != 0 || pipeline.wantConn {
			t.Fatal("read-ahead left the connection held")
		}
	})
}

func waitForPipelineCopy(t *testing.T, pipeline *rowCopyPipeline) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		pipeline.mu.Lock()
		copying := pipeline.copying
		pipeline.mu.Unlock()
		if copying {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("COPY did not start for queued rows")
		}
		time.Sleep(time.Millisecond)
	}
}

type chunkedReader struct {
	data  []byte
	chunk int
}

func (r *chunkedReader) Read(destination []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := min(len(destination), r.chunk, len(r.data))
	copy(destination, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
