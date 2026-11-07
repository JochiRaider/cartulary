package reference_data

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// ObjectCollector retries abandoned scratch, tombstones and unreferenced objects once a
// minute. Collection is maintenance, not part of an administrative mutation's
// outcome. Failure retains bytes and cannot reverse an acknowledged removal.
type ObjectCollector struct {
	db         postgres.DB
	storage    CollectionStorage
	workspaces WorkspaceReconciler
	ticker     func(time.Duration) (<-chan time.Time, func())
	observer   OperationObserver
	mu         sync.Mutex
	cancel     context.CancelFunc
	done       chan struct{}
	closed     bool
}

func NewObjectCollector(db postgres.DB, storage CollectionStorage, workspaces WorkspaceReconciler, observer OperationObserver) (*ObjectCollector, error) {
	if db == nil || storage == nil || workspaces == nil {
		return nil, errors.New("reference pack: incomplete collector dependencies")
	}
	return &ObjectCollector{db: db, storage: storage, workspaces: workspaces, observer: observer, ticker: func(interval time.Duration) (<-chan time.Time, func()) {
		ticker := time.NewTicker(interval)
		return ticker.C, ticker.Stop
	}}, nil
}
func (c *ObjectCollector) Sweep(ctx context.Context) (resultErr error) {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.collection")
	defer func() { end(referenceOutcome(resultErr)) }()
	scratchErr := c.workspaces.ReconcileWorkspaces(ctx)
	if err := ctx.Err(); err != nil {
		return errors.Join(scratchErr, err)
	}
	return errors.Join(scratchErr, CollectUnreferencedObjects(ctx, c.db, c.storage))
}

// Prepare installs retries without starting the cadence before serving.
func (c *ObjectCollector) Prepare(ctx context.Context, serving <-chan struct{}) error {
	if serving == nil {
		return errors.New("reference pack: collector serving gate is required")
	}
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
		select {
		case <-run.Done():
			return
		case <-serving:
		}
		if run.Err() != nil {
			return
		}
		ticks, stop := c.ticker(time.Minute)
		defer stop()
		for {
			select {
			case <-run.Done():
				return
			case <-ticks:
				if run.Err() != nil {
					return
				}
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
