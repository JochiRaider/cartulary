package reference_data

import (
	"context"
	"crypto/sha256"
	"errors"
	"hash"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// These boundary cases complement persisted-generation and atomic publication
// integration assertions without requiring a database fixture.
func TestIndexBatchBoundaries_Unit(t *testing.T) {
	ctx := context.Background()
	row := packformat.ContentRow{ID: "entry", Shape: "entry", Canonical: []byte(`{"id":"entry"}`), Keys: []packformat.LookupKey{{Kind: "id", Value: "entry", Order: "entry"}}}
	var flushed [][]indexBatchEntry
	b := &indexBuilder{id: uuid.New(), hashes: [3]hash.Hash{sha256.New(), sha256.New(), sha256.New()}, writeBatch: func(_ context.Context, entries []indexBatchEntry) error {
		flushed = append(flushed, entries)
		return nil
	}}
	for i := 0; i < 65; i++ {
		if err := b.Append(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	if len(flushed) != 1 || len(flushed[0]) != 64 || b.batchRows != 2 {
		t.Fatalf("row batch boundary: %d %d", len(flushed), b.batchRows)
	}
	row.Canonical[0] = '!'
	row.Keys[0].Value = "mutated"
	if flushed[0][0].row.Canonical[0] != '{' || flushed[0][0].row.Keys[0].Value != "entry" {
		t.Fatal("retained entry aliases caller memory")
	}
	if err := b.flush(ctx); err != nil {
		t.Fatal(err)
	}
	if len(flushed) != 2 || len(flushed[1]) != 1 || b.batchRows != 0 || b.bytes != 0 {
		t.Fatal("final flush lost or retained entries")
	}

	tx := &indexBatchTestTx{}
	entry := flushed[0][0]
	if err := executeIndexBatch(ctx, tx, []indexBatchEntry{entry}); err != nil {
		t.Fatal(err)
	}
	if tx.batch.Len() != 2 || tx.commits != 0 || tx.rollbacks != 0 {
		t.Fatal("borrowed batch changed transaction ownership")
	}
	if !reflect.DeepEqual(tx.batch.QueuedQueries[0].Arguments, []any{entry.indexID, "entry", "entry", entry.row.Canonical, entry.keys}) {
		t.Fatal("index parameters changed")
	}
	if !reflect.DeepEqual(tx.batch.QueuedQueries[1].Arguments, []any{entry.indexID, "entry", "entry", "id", []byte("entry"), []byte("entry")}) {
		t.Fatal("lookup parameters changed")
	}
	if err := writeIndexBatch(ctx, indexBatchTestDB{tx: tx}, []indexBatchEntry{entry}); err != nil {
		t.Fatal(err)
	}
	if tx.commits != 1 || tx.rollbacks != 1 {
		t.Fatal("owned batch did not finish its transaction")
	}
	sentinel := errors.New("batch failed")
	tx.closeErr = sentinel
	if err := writeIndexBatch(ctx, indexBatchTestDB{tx: tx}, []indexBatchEntry{entry}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	if tx.commits != 1 || tx.rollbacks != 2 {
		t.Fatal("failed batch committed or skipped rollback")
	}

	large := packformat.ContentRow{ID: "large", Shape: "object", Canonical: make([]byte, 1048577)}
	if err := b.Append(ctx, large); err != nil {
		t.Fatal(err)
	}
	if err := b.Append(ctx, packformat.ContentRow{ID: "next", Shape: "object", Canonical: []byte(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if len(flushed) != 3 || len(flushed[2]) != 1 || b.batchRows != 1 {
		t.Fatal("byte limit failed to flush the previous admitted row")
	}
	b.writeBatch = func(context.Context, []indexBatchEntry) error { return sentinel }
	if err := b.Append(ctx, large); !errors.Is(err, sentinel) {
		t.Fatal("flush error lost", err)
	}
}

type indexBatchTestTx struct {
	pgx.Tx
	batch              *pgx.Batch
	closeErr           error
	commits, rollbacks int
}

func (tx *indexBatchTestTx) SendBatch(_ context.Context, batch *pgx.Batch) pgx.BatchResults {
	tx.batch = batch
	return indexBatchTestResults{err: tx.closeErr}
}
func (tx *indexBatchTestTx) Commit(context.Context) error   { tx.commits++; return nil }
func (tx *indexBatchTestTx) Rollback(context.Context) error { tx.rollbacks++; return nil }

type indexBatchTestResults struct {
	pgx.BatchResults
	err error
}

func (r indexBatchTestResults) Close() error { return r.err }

type indexBatchTestDB struct {
	dbport.DB
	tx *indexBatchTestTx
}

func (db indexBatchTestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) { return db.tx, nil }
