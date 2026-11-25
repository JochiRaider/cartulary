package harnessruntime

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/lockwaits"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	stallDiagnosticsSchemaID = "cartulary.test.runtime_stall_diagnostics.v1"
	stallDiagnosticsPath     = "/api/v1/test/runtime/stall-diagnostics"
	// A goroutine profile beyond this is cut at a line boundary.
	goroutineProfileLimit = 256 << 10
)

// StallDiagnostics is the harness server's own view of a stalled test
// (TH-HARNESS-REQ-864): its database pool counters, the bounded lock-wait
// snapshot of its database, and its aggregated goroutine stacks. Pool
// counters separate pool exhaustion from a database lock wait; the goroutine
// profile names the code a stalled operation is waiting in. The profile is
// function names and source lines only, without argument values or labels.
type StallDiagnostics struct {
	SchemaID     string               `json:"schema_id"`
	CapturedAt   time.Time            `json:"captured_at"`
	DatabasePool *PoolCounters        `json:"database_pool"`
	LockWaits    LockWaitObservation  `json:"lock_waits"`
	Goroutines   GoroutineObservation `json:"goroutines"`
}

type PoolCounters struct {
	MaxConns               int32 `json:"max_conns"`
	TotalConns             int32 `json:"total_conns"`
	AcquiredConns          int32 `json:"acquired_conns"`
	IdleConns              int32 `json:"idle_conns"`
	ConstructingConns      int32 `json:"constructing_conns"`
	AcquireCount           int64 `json:"acquire_count"`
	EmptyAcquireCount      int64 `json:"empty_acquire_count"`
	CanceledAcquireCount   int64 `json:"canceled_acquire_count"`
	EmptyAcquireWaitTimeMS int64 `json:"empty_acquire_wait_time_ms"`
}

// LockWaitObservation is "observed" with a snapshot, or "unavailable" with a
// closed reason: no database, a pool that yielded no connection in time, or a
// failed query.
type LockWaitObservation struct {
	Status   string              `json:"status"`
	Reason   *string             `json:"reason"`
	Snapshot *lockwaits.Snapshot `json:"snapshot"`
}

type GoroutineObservation struct {
	Count     int    `json:"count"`
	Profile   string `json:"profile"`
	Truncated bool   `json:"truncated"`
}

// RegisterStallDiagnosticsRoute exposes the server's stall diagnostics to test
// support. Like every test route it exists only with test routes enabled.
func RegisterStallDiagnosticsRoute() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		pool := deps.Postgres
		return httpapi.BindTestRoute(mux, deps, guard, http.MethodGet, stallDiagnosticsPath, func(w http.ResponseWriter, r *http.Request) {
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, ObserveStall(r.Context(), pool))
		})
	}
}

// ObserveStall collects the diagnostics. It never fails: an unavailable
// observation is itself evidence, such as an exhausted pool.
func ObserveStall(ctx context.Context, pool *pgxpool.Pool) StallDiagnostics {
	diagnostics := StallDiagnostics{SchemaID: stallDiagnosticsSchemaID, CapturedAt: time.Now().UTC()}
	if pool != nil {
		stat := pool.Stat()
		diagnostics.DatabasePool = &PoolCounters{
			MaxConns: stat.MaxConns(), TotalConns: stat.TotalConns(), AcquiredConns: stat.AcquiredConns(),
			IdleConns: stat.IdleConns(), ConstructingConns: stat.ConstructingConns(),
			AcquireCount: stat.AcquireCount(), EmptyAcquireCount: stat.EmptyAcquireCount(),
			CanceledAcquireCount: stat.CanceledAcquireCount(), EmptyAcquireWaitTimeMS: stat.EmptyAcquireWaitTime().Milliseconds(),
		}
	}
	diagnostics.LockWaits = observeLockWaits(ctx, pool)
	diagnostics.Goroutines = observeGoroutines()
	return diagnostics
}

func observeLockWaits(ctx context.Context, pool *pgxpool.Pool) LockWaitObservation {
	unavailable := func(reason string) LockWaitObservation {
		return LockWaitObservation{Status: "unavailable", Reason: &reason}
	}
	if pool == nil {
		return unavailable("no_database")
	}
	ctx, cancel := context.WithTimeout(ctx, lockwaits.QueryTimeout)
	defer cancel()
	connection, err := pool.Acquire(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return unavailable("pool_acquire_timeout")
		}
		return unavailable("pool_acquire_failed")
	}
	defer connection.Release()
	rows, err := connection.Query(ctx, lockwaits.Query, lockwaits.SessionLimit)
	if err != nil {
		return unavailable("query_failed")
	}
	defer rows.Close()
	snapshot, err := lockwaits.Collect(rows, time.Now())
	if err != nil {
		return unavailable("query_failed")
	}
	return LockWaitObservation{Status: "observed", Snapshot: &snapshot}
}

// The debug=1 profile aggregates identical stacks with a count. It carries
// program counters, function names and source lines; no goroutine values.
func observeGoroutines() GoroutineObservation {
	observation := GoroutineObservation{Count: runtime.NumGoroutine()}
	var profile bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&profile, 1); err != nil {
		return observation
	}
	// Labels are caller-chosen values; only stacks leave the process.
	var data []byte
	for line := range bytes.Lines(profile.Bytes()) {
		if !bytes.HasPrefix(bytes.TrimLeft(line, "\t "), []byte("# labels:")) {
			data = append(data, line...)
		}
	}
	if len(data) > goroutineProfileLimit {
		data = data[:goroutineProfileLimit]
		if end := bytes.LastIndexByte(data, '\n'); end >= 0 {
			data = data[:end+1]
		}
		observation.Truncated = true
	}
	observation.Profile = string(data)
	return observation
}
