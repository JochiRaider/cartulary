package reference_data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// indexBuilder writes a private generation in bounded batches. Only a complete
// generation can be attached to a successful version by publication. A failed
// attempt leaves no index visible to consumers, even when earlier batches have
// committed. The operation reference protects unfinished generations from GC.
type indexBuilder struct {
	pool       indexDatabase
	writeBatch func(context.Context, *pgx.Batch) error
	id         uuid.UUID
	manifest   packformat.Manifest
	batch      pgx.Batch
	bytes      int
	counts     [3]int64
	hashes     [3]hash.Hash
	finished   bool
}

type indexDatabase interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func newIndexBuilder(ctx context.Context, db postgres.DB, operationID uuid.UUID, m packformat.Manifest, manifestSHA, payloadSHA string) (*indexBuilder, error) {
	return newIndexBuilderWithPersistence(ctx, db, func(ctx context.Context, batch *pgx.Batch) error { return writeIndexBatch(ctx, db, batch) }, operationID, m, manifestSHA, payloadSHA)
}
func newIndexBuilderTx(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, m packformat.Manifest, manifestSHA, payloadSHA string) (*indexBuilder, error) {
	return newIndexBuilderWithPersistence(ctx, tx, func(ctx context.Context, batch *pgx.Batch) error { return tx.SendBatch(ctx, batch).Close() }, operationID, m, manifestSHA, payloadSHA)
}
func newIndexBuilderWithPersistence(ctx context.Context, pool indexDatabase, writeBatch func(context.Context, *pgx.Batch) error, operationID uuid.UUID, m packformat.Manifest, manifestSHA, payloadSHA string) (*indexBuilder, error) {
	id := uuid.New()
	_, err := pool.Exec(ctx, `INSERT INTO reference_pack_index_generations(index_id,operation_id,pack_key,pack_version,manifest_sha256,payload_sha256) VALUES($1,$2,$3,$4,$5,$6)`, id, operationID, m.Key, m.Version, manifestSHA, payloadSHA)
	if err != nil {
		return nil, err
	}
	return &indexBuilder{pool: pool, writeBatch: writeBatch, id: id, manifest: m, hashes: [3]hash.Hash{sha256.New(), sha256.New(), sha256.New()}}, nil
}

func (b *indexBuilder) Append(ctx context.Context, row packformat.ContentRow) error {
	if b.finished {
		return errors.New("reference pack: closed index builder")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	kind := -1
	switch row.Shape {
	case "entry":
		kind = 0
	case "object":
		kind = 1
	case "relationship":
		kind = 2
	}
	if kind < 0 {
		return errors.New("reference pack: invalid index row kind")
	}
	keys, err := canonicaljson.Marshal(row.Keys)
	if err != nil {
		return err
	}
	// Flush before retaining another row; at most one admitted row may exceed
	// the one-MiB batch target. Row and alias bounds belong to profile admission.
	if b.batch.Len() > 0 && (b.batch.Len()+len(row.Keys)+1 > 128 || b.bytes+len(keys)+len(row.Canonical) > 1048576) {
		if err := b.flush(ctx); err != nil {
			return err
		}
	}
	b.batch.Queue(`INSERT INTO reference_pack_indexes(index_id,entry_kind,entry_id,canonical_item,lookup_keys) VALUES($1,$2,$3,$4,$5)`, b.id, row.Shape, row.ID, row.Canonical, keys)
	for _, key := range row.Keys {
		b.batch.Queue(`INSERT INTO reference_pack_lookup_keys(index_id,entry_kind,entry_id,lookup_kind,lookup_value,sort_key) VALUES($1,$2,$3,$4,$5,$6)`, b.id, row.Shape, row.ID, key.Kind, []byte(key.Value), []byte(key.Order))
	}
	b.bytes += len(keys) + len(row.Canonical)
	b.counts[kind]++
	_, _ = b.hashes[kind].Write(row.Canonical)
	_, _ = b.hashes[kind].Write([]byte{'\n'})
	return nil
}

func (b *indexBuilder) flush(ctx context.Context) error {
	if b.batch.Len() == 0 {
		return nil
	}
	err := b.writeBatch(ctx, &b.batch)
	b.batch = pgx.Batch{}
	b.bytes = 0
	return err
}

func (b *indexBuilder) Complete(ctx context.Context) (uuid.UUID, error) {
	if b.finished {
		return uuid.Nil, errors.New("reference pack: index already completed")
	}
	b.finished = true
	expected := [3]*int64{b.manifest.Summary.Entries, b.manifest.Summary.Objects, b.manifest.Summary.Relationships}
	for i, n := range expected {
		if n == nil && b.counts[i] != 0 || n != nil && *n != b.counts[i] {
			return uuid.Nil, errors.New("reference pack: incomplete index")
		}
		if n != nil {
			path := []string{"payload/entries.ndjson", "payload/objects.ndjson", "payload/relationships.ndjson"}[i]
			matched := false
			for _, file := range b.manifest.Files {
				if file.Path == path && file.SHA256 == hex.EncodeToString(b.hashes[i].Sum(nil)) {
					matched = true
					break
				}
			}
			if !matched {
				return uuid.Nil, errors.New("reference pack: index differs from verified payload")
			}
		}
	}
	if err := b.flush(ctx); err != nil {
		return uuid.Nil, err
	}
	_, err := b.pool.Exec(ctx, `UPDATE reference_pack_index_generations SET entry_count=$2,object_count=$3,relationship_count=$4,complete=true WHERE index_id=$1 AND NOT complete`, b.id, b.counts[0], b.counts[1], b.counts[2])
	return b.id, err
}

// A retained rebuild already owns a transaction. Ordinary preparation owns only
// a narrow DB port and commits each bounded, still-private index batch. This
// preserves SendBatch's implicit-transaction behavior without requiring pools.
func writeIndexBatch(ctx context.Context, database postgres.DB, batch *pgx.Batch) error {
	tx, err := database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
