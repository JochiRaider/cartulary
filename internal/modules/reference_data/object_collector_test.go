package reference_data

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Test-only access to the scheduler seam; production always uses one minute.
func NewTestObjectCollector(t testing.TB, db postgres.DB, storage CollectionStorage, workspaces WorkspaceReconciler, observer OperationObserver, ticks <-chan time.Time) *ObjectCollector {
	t.Helper()
	c, err := NewObjectCollector(db, storage, workspaces, observer)
	if err != nil {
		t.Fatal(err)
	}
	c.ticker = func(interval time.Duration) (<-chan time.Time, func()) {
		if interval != time.Minute {
			t.Error("production maintenance cadence changed", interval)
		}
		return ticks, func() {}
	}
	return c
}

type collectorScratchFunc func(context.Context) error

func (f collectorScratchFunc) ReconcileWorkspaces(ctx context.Context) error { return f(ctx) }

type collectorPublishedFunc struct {
	ArtifactStorage
	run func(context.Context) error
}

func (s collectorPublishedFunc) WithPublishedCollection(ctx context.Context, _ func(PublishedObjects) error) error {
	return s.run(ctx)
}

func TestObjectCollectorLifecycle_Unit(t *testing.T) {
	t.Run("separate phases aggregate failures within one budget", func(t *testing.T) {
		scratchErr, publishedErr := errors.New("scratch"), errors.New("published")
		order := ""
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		deadline, _ := ctx.Deadline()
		c, err := NewObjectCollector(new(pgxpool.Pool), collectorPublishedFunc{run: func(ctx context.Context) error {
			order += "published"
			if got, _ := ctx.Deadline(); got != deadline {
				t.Error("phase budget reset")
			}
			return publishedErr
		}}, collectorScratchFunc(func(context.Context) error { order += "scratch,"; return scratchErr }), nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Sweep(ctx); !errors.Is(err, scratchErr) || !errors.Is(err, publishedErr) || order != "scratch,published" {
			t.Fatal(order, err)
		}
		order = ""
		c.workspaces = collectorScratchFunc(func(context.Context) error { cancel(); return scratchErr })
		if err := c.Sweep(ctx); !errors.Is(err, context.Canceled) || !errors.Is(err, scratchErr) || order != "" {
			t.Fatal("published phase ran after budget", order, err)
		}
	})
	t.Run("serving gate serial work and joined shutdown", func(t *testing.T) {
		ticks := make(chan time.Time, 2)
		entered, interrupted, release := make(chan struct{}, 2), make(chan struct{}), make(chan struct{})
		c := NewTestObjectCollector(t, new(pgxpool.Pool), collectorPublishedFunc{run: func(context.Context) error { t.Error("published collection after shutdown"); return nil }}, collectorScratchFunc(func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 30*time.Second {
				t.Error("missing shared sweep budget")
			}
			entered <- struct{}{}
			<-ctx.Done()
			close(interrupted)
			<-release
			return ctx.Err()
		}), nil, ticks)
		serving := make(chan struct{})
		if err := c.Prepare(t.Context(), serving); err != nil {
			t.Fatal(err)
		}
		ticks <- time.Now()
		select {
		case <-entered:
			t.Fatal("collection before serving")
		case <-time.After(20 * time.Millisecond):
		}
		close(serving)
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("maintenance did not start")
		}
		ticks <- time.Now()
		select {
		case <-entered:
			t.Fatal("overlapping sweep")
		case <-time.After(20 * time.Millisecond):
		}
		closed := make(chan error, 1)
		go func() { closed <- c.Close(context.Background()) }()
		<-interrupted
		select {
		case <-closed:
			t.Fatal("shutdown returned with owned work")
		default:
		}
		close(release)
		if err := <-closed; err != nil {
			t.Fatal(err)
		}
		if err := c.Prepare(t.Context(), serving); err == nil {
			t.Fatal("closed collector restarted")
		}
	})
}
