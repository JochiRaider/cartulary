package extensions

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

type StagedCleanupStore interface {
	PrepareCleanupBatch(context.Context, time.Time, time.Time, int) ([]extensionstore.StagedObject, error)
	RecordDeletionSuccess(context.Context, string) error
	RecordDeletionFailure(context.Context, string, string, time.Time) error
}

type StagedObjectDeleter interface {
	DeleteStagedObject(context.Context, string) error
}

type StagedCleanupReadiness interface {
	StagedCleanupUnavailable(error)
	StagedCleanupAvailable()
}

type StagedObjectJanitor struct {
	store     StagedCleanupStore
	deleter   StagedObjectDeleter
	readiness StagedCleanupReadiness
	now       func() time.Time
	limit     int

	mu      sync.Mutex
	running bool
	pending bool
}

func NewStagedObjectJanitor(store StagedCleanupStore, deleter StagedObjectDeleter, readiness StagedCleanupReadiness, now func() time.Time, limit int) (*StagedObjectJanitor, error) {
	if store == nil || deleter == nil || limit < 1 {
		return nil, errors.New("extension staged-object janitor dependencies are incomplete")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &StagedObjectJanitor{store: store, deleter: deleter, readiness: readiness, now: now, limit: limit}, nil
}

// Sweep serializes cleanup. Any number of overlapping or missed triggers are
// coalesced into at most one immediately following pass.
func (j *StagedObjectJanitor) Sweep(ctx context.Context) error {
	j.mu.Lock()
	if j.running {
		j.pending = true
		j.mu.Unlock()
		return nil
	}
	j.running = true
	j.mu.Unlock()
	defer func() {
		j.mu.Lock()
		j.running = false
		j.pending = false
		j.mu.Unlock()
	}()
	for {
		if err := j.sweepOnce(ctx); err != nil {
			if j.readiness != nil {
				j.readiness.StagedCleanupUnavailable(err)
			}
			return err
		}
		if j.readiness != nil {
			j.readiness.StagedCleanupAvailable()
		}
		j.mu.Lock()
		pending := j.pending
		j.pending = false
		j.mu.Unlock()
		if !pending {
			return nil
		}
	}
}

func (j *StagedObjectJanitor) sweepOnce(ctx context.Context) error {
	now := j.now().UTC()
	objects, err := j.store.PrepareCleanupBatch(ctx, now, now, j.limit)
	if err != nil {
		return err
	}
	for _, object := range objects {
		if err := j.deleter.DeleteStagedObject(ctx, object.StorageIdentity); err != nil {
			if recordErr := j.store.RecordDeletionFailure(ctx, object.StagingID, "object_delete_failed", j.now().UTC()); recordErr != nil {
				return errors.Join(err, recordErr)
			}
			continue
		}
		if err := j.store.RecordDeletionSuccess(ctx, object.StagingID); err != nil {
			return err
		}
	}
	return nil
}

func (j *StagedObjectJanitor) Run(ctx context.Context, interval time.Duration) error {
	if j == nil || interval <= 0 {
		return errors.New("extension staged-object sweep interval is invalid")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = j.Sweep(ctx)
		}
	}
}
