package reference_data

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// ObjectCollector retries durable tombstones and unreferenced objects once a
// minute. Collection is maintenance, not part of an administrative mutation's
// outcome. Failure retains bytes and cannot reverse an acknowledged removal.
type ObjectCollector struct {
	db       postgres.DB
	storage  CollectionStorage
	observer OperationObserver
	mu       sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
	closed   bool
}

func NewObjectCollector(db postgres.DB, storage CollectionStorage, observer OperationObserver) (*ObjectCollector, error) {
	if db == nil || storage == nil {
		return nil, errors.New("reference pack: incomplete collector dependencies")
	}
	return &ObjectCollector{db: db, storage: storage, observer: observer}, nil
}
func (c *ObjectCollector) Sweep(ctx context.Context) (resultErr error) {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.collection")
	defer func() { end(referenceOutcome(resultErr)) }()
	return CollectUnreferencedObjects(ctx, c.db, c.storage)
}
func (c *ObjectCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("reference pack: collector closed")
	}
	if c.done != nil {
		return nil
	}
	run, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})
	go func() {
		defer close(c.done)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-run.Done():
				return
			case <-ticker.C:
				attempt, cancel := context.WithTimeout(run, 30*time.Second)
				_ = c.Sweep(attempt)
				cancel()
			}
		}
	}()
	return nil
}
func (c *ObjectCollector) Close(ctx context.Context) error {
	c.mu.Lock()
	c.closed = true
	done := c.done
	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
