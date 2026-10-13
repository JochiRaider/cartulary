package recoverybrowsertest

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"
)

// Slow-drop observations are bounded structural diagnostics. The exact SQL is
// only a lookup key; neither it nor unrelated activity crosses this boundary.
func observeDatabaseDrop(ctx context.Context, dsn, name string) string {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var db *sql.DB
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()
	first, last := "", ""
	for {
		select {
		case <-ctx.Done():
			if first != "" && first != last {
				return "first={" + first + "}; last={" + last + "}"
			}
			return last
		case <-ticker.C:
		}
		if db == nil {
			var err error
			db, err = sql.Open("pgx", dsn)
			if err != nil {
				return "observer_unavailable"
			}
			db.SetMaxOpenConns(1)
		}
		probe, cancel := context.WithTimeout(ctx, time.Second)
		var kind, event, checkpointKind, checkpointEvent string
		var blockers, checkpointBlockers int
		var requested, done, buffers int64
		err := db.QueryRowContext(probe, `
SELECT COALESCE(d.wait_event_type, ''), COALESCE(d.wait_event, ''), cardinality(pg_blocking_pids(d.pid)),
       COALESCE(c.wait_event_type, ''), COALESCE(c.wait_event, ''), cardinality(pg_blocking_pids(c.pid)),
       s.num_requested, s.num_done, s.buffers_written
FROM pg_catalog.pg_stat_activity d
CROSS JOIN pg_catalog.pg_stat_checkpointer s
CROSS JOIN LATERAL (SELECT pid, wait_event_type, wait_event FROM pg_catalog.pg_stat_activity WHERE backend_type = 'checkpointer' LIMIT 1) c
WHERE d.pid <> pg_backend_pid() AND d.state = 'active' AND d.query = $1
LIMIT 1`, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)).Scan(
			&kind, &event, &blockers, &checkpointKind, &checkpointEvent, &checkpointBlockers,
			&requested, &done, &buffers,
		)
		cancel()
		if err == nil {
			last = dropWaitDescription(kind, event, blockers) + "; checkpointer={" +
				dropWaitDescription(checkpointKind, checkpointEvent, checkpointBlockers) + " " +
				checkpointProgressDescription(requested, done, buffers) + "}"
			if first == "" {
				first = last
			}
		} else if ctx.Err() == nil && err != sql.ErrNoRows {
			last = "observer_unavailable"
		}
	}
}

func dropWaitDescription(kind, event string, blockers int) string {
	// Closed vocabulary: even a syntactically valid extension-provided event can
	// carry private data. Unknown values never cross this boundary.
	if !slices.Contains([]string{"", "Activity", "BufferPin", "Client", "Extension", "IO", "IPC", "Lock", "LWLock", "Timeout"}, kind) || blockers < 0 || blockers > 100000 {
		return "unknown_wait"
	}
	if !slices.Contains([]string{"", "ProcSignalBarrier", "DataFileRead", "DataFileWrite", "DataFileSync", "WALWrite", "WALSync", "SyncRep", "CheckpointDone", "CheckpointStart", "CheckpointWriteDelay", "CheckpointDelayStart", "CheckpointDelayComplete", "CheckpointerMain", "BufferIO", "relation", "transactionid", "virtualxid", "object", "advisory", "extend", "frozenid", "ClientRead", "ClientWrite"}, event) {
		event = "other"
	}
	return fmt.Sprintf("type=%s event=%s blockers=%d", kind, event, blockers)
}

func checkpointProgressDescription(requested, done, buffers int64) string {
	if requested < 0 || done < 0 || buffers < 0 {
		return "progress_unavailable"
	}
	// Cumulative server counters can lag; these snapshots do not claim the
	// identity or completion of the particular checkpoint requested by DROP.
	return fmt.Sprintf("requested=%d done=%d buffers=%d", requested, done, buffers)
}
