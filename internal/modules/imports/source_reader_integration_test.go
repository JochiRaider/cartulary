package imports_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func assertBoundedSourceReaders(t *testing.T, harness *appsupport.ServerHarness, ref string) {
	ctx := context.Background()
	config := harness.Pool.Config().Copy()
	config.MaxConns = 1
	config.MinConns = 0
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	port, err := imports.NewExtensionSourcePort(pool)
	if err != nil {
		t.Fatal(err)
	}
	write := func(content []byte, digest string) {
		t.Helper()
		if _, err := harness.DB.ExecContext(ctx, `UPDATE import_source_streams SET source_bytes=$2,source_byte_size=$3,source_content_sha256=$4 WHERE source_stream_ref=$1`, ref, content, len(content), digest); err != nil {
			t.Fatal(err)
		}
	}
	sum := func(content []byte) string { return fmt.Sprintf("%x", sha256.Sum256(content)) }
	read := func(reader io.ReadCloser, want []byte) {
		t.Helper()
		got, err := io.ReadAll(reader)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("content or integrity: length=%d err=%v", len(got), err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for _, size := range []int{0, 1, 65535, 65536, 65537, 524288} {
		t.Run(fmt.Sprintf("bytes_%d", size), func(t *testing.T) {
			content := bytes.Repeat([]byte{'x'}, size)
			write(content, sum(content))
			stream, err := port.OpenSourceStream(ctx, ref)
			if err != nil {
				t.Fatal(err)
			}
			read(stream.Reader, content)
			if pool.Stat().AcquiredConns() != 0 {
				t.Fatal("EOF leaked owned snapshot")
			}
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			guard := &boundedSourceTx{Tx: tx, t: t}
			stream, err = port.OpenSourceStreamTx(ctx, guard, ref)
			if err != nil {
				t.Fatal(err)
			}
			// The connection must be usable between reads and after EOF/close.
			prefix := make([]byte, min(size, 7))
			if _, err := io.ReadFull(stream.Reader, prefix); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, "SELECT 1"); err != nil {
				t.Fatal(err)
			}
			read(stream.Reader, content[len(prefix):])
			if _, err := tx.Exec(ctx, "SELECT 1"); err != nil {
				t.Fatalf("reader ended borrowed transaction: %v", err)
			}
			if guard.maxBytes > 65536 || guard.reads != (size+65535)/65536+1 {
				t.Fatalf("unbounded or unexpected reads: %d / %d", guard.maxBytes, guard.reads)
			}
		})
	}
	t.Run("consistent snapshot", func(t *testing.T) {
		before := bytes.Repeat([]byte{'a'}, 131073)
		after := bytes.Repeat([]byte{'b'}, len(before))
		write(before, sum(before))
		stream, err := port.OpenSourceStream(ctx, ref)
		if err != nil {
			t.Fatal(err)
		}
		write(after, sum(after))
		read(stream.Reader, before)
		stream, err = port.OpenSourceStream(ctx, ref)
		if err != nil {
			t.Fatal(err)
		}
		read(stream.Reader, after)
	})
	t.Run("incorrect declared size and slice failures", func(t *testing.T) {
		content := []byte("exact length")
		write(content, sum(content))
		for _, delta := range []int64{-1, 1} {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			guard := &boundedSourceTx{Tx: tx, t: t, declaredDelta: delta}
			stream, err := port.OpenSourceStreamTx(ctx, guard, ref)
			if err != nil {
				t.Fatal(err)
			}
			_, readErr := io.Copy(io.Discard, stream.Reader)
			_ = stream.Reader.Close()
			_ = tx.Rollback(ctx)
			if readErr == nil {
				t.Fatalf("incorrect byte count admitted: %d", delta)
			}
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		failure := errors.New("injected slice failure")
		stream, err := port.OpenSourceStreamTx(ctx, &boundedSourceTx{Tx: tx, t: t, readError: failure}, ref)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(io.Discard, stream.Reader); !errors.Is(err, failure) {
			t.Fatalf("slice failure hidden: %v", err)
		}
		if _, err := tx.Exec(ctx, "SELECT 1"); err != nil {
			t.Fatalf("failure ended borrowed transaction: %v", err)
		}
	})
	t.Run("database slice cost observations", func(t *testing.T) {
		for _, size := range []int{524288, 4194304} {
			content := bytes.Repeat([]byte{'c'}, size)
			write(content, sum(content))
			var planJSON []byte
			if err := harness.DB.QueryRowContext(ctx, `EXPLAIN (ANALYZE,BUFFERS,FORMAT JSON) SELECT substring(source_bytes FROM 65537 FOR 65536) FROM import_source_streams WHERE source_stream_ref=$1`, ref).Scan(&planJSON); err != nil {
				t.Fatal(err)
			}
			var plan []struct{ Plan map[string]any }
			if err := json.Unmarshal(planJSON, &plan); err != nil {
				t.Fatal(err)
			}
			// This is a database page-access bound, not a latency/RSS benchmark.
			// A full read of the larger EXTERNAL value needs over 500 pages.
			blocks := plan[0].Plan["Shared Hit Blocks"].(float64) + plan[0].Plan["Shared Read Blocks"].(float64)
			if blocks < 1 || blocks > 64 {
				t.Fatalf("slice fetched excessive database pages: source=%d blocks=%v", size, blocks)
			}
			t.Logf("source_bytes=%d slice_bytes=65536 database_shared_hit_blocks=%v database_shared_read_blocks=%v", size, plan[0].Plan["Shared Hit Blocks"], plan[0].Plan["Shared Read Blocks"])
		}
	})
	t.Run("integrity failure and early close", func(t *testing.T) {
		write([]byte("wrong digest"), strings.Repeat("0", 64))
		stream, err := port.OpenSourceStream(ctx, ref)
		if err != nil {
			t.Fatalf("metadata lookup read content: %v", err)
		}
		if _, err := io.Copy(io.Discard, stream.Reader); err == nil {
			t.Fatal("digest mismatch accepted")
		}
		if pool.Stat().AcquiredConns() != 0 {
			t.Fatal("integrity failure leaked snapshot")
		}
		stream, err = port.OpenSourceStream(ctx, ref)
		if err != nil {
			t.Fatal(err)
		}
		if err := stream.Reader.Close(); err != nil {
			t.Fatal(err)
		}
		if err := stream.Reader.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := stream.Reader.Read(make([]byte, 1)); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("early close certified source: %v", err)
		}
	})
	t.Run("idle cancellation releases single connection", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		stream, err := port.OpenSourceStream(cancelCtx, ref)
		if err != nil {
			t.Fatal(err)
		}
		cancel()
		acquireCtx, done := context.WithTimeout(ctx, 2*time.Second)
		defer done()
		conn, err := pool.Acquire(acquireCtx)
		if err != nil {
			t.Fatalf("cancel leaked snapshot: %v", err)
		}
		conn.Release()
		if _, err := stream.Reader.Read(make([]byte, 1)); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel result: %v", err)
		}
	})
	t.Run("borrowed source remains protected after reader close", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		stream, err := port.OpenSourceStreamTx(ctx, tx, ref)
		if err != nil {
			t.Fatal(err)
		}
		if err := stream.Reader.Close(); err != nil {
			t.Fatal(err)
		}
		other, err := harness.DB.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer other.Rollback()
		if _, err := other.ExecContext(ctx, "SET LOCAL lock_timeout='100ms'"); err != nil {
			t.Fatal(err)
		}
		if _, err := other.ExecContext(ctx, `UPDATE import_source_streams SET source_media_type='text/plain' WHERE source_stream_ref=$1`, ref); err == nil || !strings.Contains(err.Error(), "lock timeout") {
			t.Fatalf("source not protected: %v", err)
		}
		if _, err := tx.Exec(ctx, "SELECT 1"); err != nil {
			t.Fatal(err)
		}
	})
}

type boundedSourceTx struct {
	pgx.Tx
	t               *testing.T
	reads, maxBytes int
	declaredDelta   int64
	readError       error
}

func (tx *boundedSourceTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "substring(source_bytes") {
		if tx.readError != nil {
			return sourceTestRow{err: tx.readError}
		}
		tx.reads++
		length := int(args[2].(int64))
		if length < 1 || length > 65536 {
			tx.t.Fatalf("unbounded SQL read: %d", length)
		}
		tx.maxBytes = max(tx.maxBytes, length)
	} else if strings.Contains(sql, "streams.source_bytes") {
		tx.t.Fatal("metadata query materialized source content")
	}
	return sourceTestRow{Row: tx.Tx.QueryRow(ctx, sql, args...), declaredDelta: tx.declaredDelta}
}

type sourceTestRow struct {
	pgx.Row
	declaredDelta int64
	err           error
}

func (r sourceTestRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if err := r.Row.Scan(dest...); err != nil {
		return err
	}
	if len(dest) == 7 {
		*dest[5].(*int64) += r.declaredDelta
	}
	return nil
}

func (tx *boundedSourceTx) Commit(context.Context) error {
	tx.t.Fatal("reader committed borrowed transaction")
	return nil
}
func (tx *boundedSourceTx) Rollback(context.Context) error {
	tx.t.Fatal("reader rolled back borrowed transaction")
	return nil
}
