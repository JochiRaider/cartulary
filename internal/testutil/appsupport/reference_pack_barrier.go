package appsupport

import (
	"context"
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
	entered chan<- struct{}
	release <-chan struct{}
}

func (b *ReferencePackVerificationBarrier) BlockNext(entered chan<- struct{}, release <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.next != nil {
		panic("reference pack barrier already armed")
	}
	b.next = &referencePackBlockedRead{entered: entered, release: release}
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
		db.barrier.next = nil
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
