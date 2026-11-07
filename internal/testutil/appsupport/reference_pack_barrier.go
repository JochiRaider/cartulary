package appsupport

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

// ReferencePackVerificationBarrier blocks the next preparation configuration
// read after durable attempt admission. It belongs to one server's dependency
// graph and never intercepts another runtime or an admission transaction.
type ReferencePackVerificationBarrier struct {
	mu   sync.Mutex
	next *referencePackBlockedRead
}
type referencePackBlockedRead struct {
	query   string
	failure error
	entered chan<- struct{}
	release <-chan struct{}
}

func (b *ReferencePackVerificationBarrier) BlockNext(entered chan<- struct{}, release <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.next != nil {
		panic("reference pack barrier already armed")
	}
	b.next = &referencePackBlockedRead{query: "SELECT configuration_sha256 FROM reference_pack_current_set WHERE singleton", entered: entered, release: release}
}
func (b *ReferencePackVerificationBarrier) WrapDatabase(db postgres.DB) postgres.DB {
	return referencePackBarrierDB{DB: db, barrier: b}
}

type referencePackBarrierDB struct {
	postgres.DB
	barrier *ReferencePackVerificationBarrier
}
type referencePackCanceledRead struct{ err error }

func (r referencePackCanceledRead) Scan(...any) error { return r.err }

func (db referencePackBarrierDB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if query == "SELECT configuration_sha256 FROM reference_pack_current_set WHERE singleton" {
		db.barrier.mu.Lock()
		next := db.barrier.next
		if next != nil && next.query == query {
			db.barrier.next = nil
		} else {
			next = nil
		}
		db.barrier.mu.Unlock()
		if next != nil {
			close(next.entered)
			select {
			case <-next.release:
			case <-ctx.Done():
				return referencePackCanceledRead{ctx.Err()}
			}
		}
	}
	return db.DB.QueryRow(ctx, query, args...)
}

// BlockNextIndex interrupts private indexing after extraction has allocated real scratch.
func (b *ReferencePackVerificationBarrier) BlockNextIndex(entered chan<- struct{}, release <-chan struct{}, failure error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.next != nil {
		panic("reference pack barrier already armed")
	}
	b.next = &referencePackBlockedRead{query: "INSERT INTO reference_pack_index_generations", entered: entered, release: release, failure: failure}
}
func (db referencePackBarrierDB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.barrier.mu.Lock()
	next := db.barrier.next
	if next != nil && strings.HasPrefix(query, next.query) {
		db.barrier.next = nil
	} else {
		next = nil
	}
	db.barrier.mu.Unlock()
	if next != nil {
		close(next.entered)
		select {
		case <-next.release:
		case <-ctx.Done():
			return pgconn.CommandTag{}, ctx.Err()
		}
		if next.failure != nil {
			return pgconn.CommandTag{}, next.failure
		}
	}
	return db.DB.Exec(ctx, query, args...)
}
