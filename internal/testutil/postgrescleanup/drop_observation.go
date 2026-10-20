package postgrescleanup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
)

const maxDiagnosticCounter int64 = 9007199254740991

// DropWaitSnapshot contains structural observations only, never SQL, process
// identifiers, database names, connection strings or extension-provided text.
type DropWaitSnapshot struct {
	WriteMS              float64 `json:"write_ms"`
	SyncMS               float64 `json:"sync_ms"`
	statsReset           time.Time
	Barrier              *BarrierWaitSnapshot `json:"barrier,omitempty"`
	Type                 string               `json:"type"`
	Event                string               `json:"event"`
	Blockers             int                  `json:"blockers"`
	CheckpointerType     string               `json:"checkpointer_type"`
	CheckpointerEvent    string               `json:"checkpointer_event"`
	CheckpointerBlockers int                  `json:"checkpointer_blockers"`
	Requested            int64                `json:"requested"`
	Done                 int64                `json:"done"`
	Buffers              int64                `json:"buffers"`
}

// BarrierWaitSnapshot describes only the server-reported participant's sampled
// state. It does not retain the transient PID, notice, SQL or database identity.
type BarrierWaitSnapshot struct {
	Operation      string `json:"operation"`
	DatabaseScope  string `json:"database_scope"`
	OperationAgeMS int64  `json:"operation_age_ms"`
	Status         string `json:"status"`
	Backend        string `json:"backend"`
	State          string `json:"state"`
	Type           string `json:"type"`
	Event          string `json:"event"`
	Blockers       int    `json:"blockers"`
}

func (b *BarrierWaitSnapshot) valid() bool {
	if b == nil {
		return true
	}
	if b.Status == "unavailable" || b.Status == "not_observed" {
		return b.Operation == "" && b.DatabaseScope == "" && b.OperationAgeMS == 0 && b.Backend == "" && b.State == "" && b.Type == "" && b.Event == "" && b.Blockers == 0
	}
	return b.Status == "sampled" && validBarrierOperation(b.Operation) && slices.Contains([]string{"none", "admin", "target", "other"}, b.DatabaseScope) && b.OperationAgeMS >= 0 && b.OperationAgeMS <= 86400000 && validBackend(b.Backend) && validBackendState(b.State) && validWaitType(b.Type) && validWaitEvent(b.Event) && b.Blockers >= 0 && b.Blockers <= 100000
}

func validBackend(value string) bool {
	return slices.Contains([]string{"other", "client backend", "autovacuum worker", "autovacuum launcher", "background writer", "checkpointer", "walwriter", "walsender", "walreceiver", "startup", "archiver", "parallel worker", "logical replication launcher", "logical replication worker", "io worker"}, value)
}
func validBackendState(value string) bool {
	return slices.Contains([]string{"", "other", "active", "idle", "idle in transaction", "idle in transaction (aborted)", "fastpath function call", "disabled"}, value)
}

func validBarrierOperation(value string) bool {
	return slices.Contains([]string{"none", "other", "create_database", "alter_database", "drop_database", "select", "with", "insert", "update", "delete", "copy", "create", "alter", "drop", "do", "begin", "commit", "rollback", "vacuum", "analyze", "checkpoint", "discard", "show", "set"}, value)
}

func barrierOperation(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return "none"
	}
	// pg_stat_activity truncates query text; an unknown prefix stays other.
	// Only closed grammar categories survive. No statement text is returned.
	words := strings.Fields(strings.ToLower(query))
	verb := strings.TrimSuffix(words[0], ";")
	if len(words) > 1 && words[1] == "database" && (verb == "create" || verb == "alter" || verb == "drop") {
		return verb + "_database"
	}
	if validBarrierOperation(verb) && verb != "none" && !strings.Contains(verb, "_") {
		return verb
	}
	return "other"
}

func barrierNoticePID(notice *pgconn.Notice) int64 {
	if notice == nil || notice.SeverityUnlocalized != "LOG" || notice.Code != "00000" || notice.Routine != "WaitForProcSignalBarrier" {
		return 0
	}
	const prefix = "still waiting for backend with PID "
	const suffix = " to accept ProcSignalBarrier"
	value, ok := strings.CutPrefix(notice.Message, prefix)
	if !ok {
		return 0
	}
	value, ok = strings.CutSuffix(value, suffix)
	if !ok || len(value) == 0 || len(value) > 10 {
		return 0
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0
		}
	}
	pid, err := strconv.ParseInt(value, 10, 32)
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

type DropObservation struct {
	ServiceGeneration *string             `json:"-"`
	Checkpoint        *CheckpointInterval `json:"checkpoint,omitempty"`
	Status            string              `json:"status"`
	First             *DropWaitSnapshot   `json:"first"`
	Last              *DropWaitSnapshot   `json:"last"`
}

func validWaitType(value string) bool {
	return slices.Contains([]string{"", "unknown", "Activity", "BufferPin", "Client", "Extension", "IO", "IPC", "Lock", "LWLock", "Timeout"}, value)
}

func validWaitEvent(value string) bool {
	return slices.Contains([]string{"", "other", "ProcSignalBarrier", "DataFileRead", "DataFileWrite", "DataFileSync", "WALWrite", "WALSync", "SyncRep", "CheckpointDone", "CheckpointStart", "CheckpointWriteDelay", "CheckpointDelayStart", "CheckpointDelayComplete", "CheckpointerMain", "BufferIO", "relation", "transactionid", "virtualxid", "object", "advisory", "extend", "frozenid", "ClientRead", "ClientWrite"}, value)
}

func (s DropWaitSnapshot) valid() bool {
	return validMilliseconds(s.WriteMS) && validMilliseconds(s.SyncMS) && s.Barrier.valid() && validWaitType(s.Type) && validWaitEvent(s.Event) && validWaitType(s.CheckpointerType) && validWaitEvent(s.CheckpointerEvent) &&
		s.Blockers >= 0 && s.Blockers <= 100000 && s.CheckpointerBlockers >= 0 && s.CheckpointerBlockers <= 100000 &&
		s.Requested >= 0 && s.Requested <= maxDiagnosticCounter && s.Done >= 0 && s.Done <= maxDiagnosticCounter && s.Buffers >= 0 && s.Buffers <= maxDiagnosticCounter
}

func (o DropObservation) Validate() error {
	if o.Checkpoint != nil && !o.Checkpoint.valid() {
		return errors.New("invalid checkpoint interval")
	}
	if o.Status == "sampled" && o.First != nil && o.Last != nil && o.First.valid() && o.Last.valid() {
		return nil
	}
	if (o.Status == "unavailable" || o.Status == "not_observed") && o.First == nil && o.Last == nil {
		return nil
	}
	return errors.New("invalid PostgreSQL drop observation")
}

func (o DropObservation) String() string {
	if o.Validate() != nil {
		return "observer_unavailable"
	}
	if o.Status != "sampled" {
		return o.Status
	}
	describe := func(s *DropWaitSnapshot) string {
		description := DropWaitDescription(s.Type, s.Event, s.Blockers) + "; checkpointer={" +
			DropWaitDescription(s.CheckpointerType, s.CheckpointerEvent, s.CheckpointerBlockers) + " " +
			CheckpointProgressDescription(s.Requested, s.Done, s.Buffers) + "}"
		if b := s.Barrier; b != nil {
			description += fmt.Sprintf("; barrier={status=%s backend=%s state=%s operation=%s scope=%s age_ms=%d %s}", b.Status, b.Backend, b.State, b.Operation, b.DatabaseScope, b.OperationAgeMS, DropWaitDescription(b.Type, b.Event, b.Blockers))
		}
		return description
	}
	if !reflect.DeepEqual(o.First, o.Last) {
		return "first={" + describe(o.First) + "}; last={" + describe(o.Last) + "}"
	}
	return describe(o.Last)
}

type observedDropError struct {
	err         error
	observation DropObservation
}

func (e *observedDropError) Error() string {
	return fmt.Sprintf("%v (drop wait: %s)", e.err, e.observation.String())
}
func (e *observedDropError) Unwrap() error { return e.err }

func executeObservedDatabaseDrop(ctx context.Context, admin *sql.Conn, dsn, name string, participant func() int64) error {
	var backend int64
	// Read the already-admitted physical connection identity without issuing SQL.
	// It is transient and never included in error text or retained diagnostics.
	_ = admin.Raw(func(driver any) error {
		if connection, ok := driver.(*stdlib.Conn); ok {
			backend = int64(connection.Conn().PgConn().PID())
		}
		return nil
	})
	observationCtx, stop := context.WithCancel(ctx)
	observed := make(chan DropObservation, 1)
	go func() { observed <- observeDatabaseDrop(observationCtx, dsn, name, backend, participant) }()
	err := executeDatabaseDrop(ctx, admin, name)
	stop()
	observation := <-observed
	if err != nil {
		return &observedDropError{err: err, observation: observation}
	}
	return nil
}

func observeDatabaseDrop(ctx context.Context, dsn, name string, backend int64, participant func() int64) DropObservation {
	result := DropObservation{Status: "not_observed"}
	if validateDatabaseIdentifier(name) != nil || backend <= 0 {
		return DropObservation{Status: "unavailable"}
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var db *sql.DB
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return result
		case <-ticker.C:
		}
		if db == nil {
			var err error
			db, err = sql.Open("pgx", dsn)
			if err != nil {
				return DropObservation{Status: "unavailable"}
			}
			db.SetMaxOpenConns(1)
		}
		probe, cancel := context.WithTimeout(ctx, time.Second)
		var snapshot DropWaitSnapshot
		var generation string
		err := db.QueryRowContext(probe, `
SELECT COALESCE(d.wait_event_type, ''), COALESCE(d.wait_event, ''), cardinality(pg_blocking_pids(d.pid)),
       COALESCE(c.wait_event_type, ''), COALESCE(c.wait_event, ''), cardinality(pg_blocking_pids(c.pid)),
       s.num_requested, s.num_done, s.buffers_written, s.write_time, s.sync_time, s.stats_reset, (SELECT system_identifier::text FROM pg_catalog.pg_control_system())
FROM pg_catalog.pg_stat_activity d
CROSS JOIN pg_catalog.pg_stat_checkpointer s
CROSS JOIN LATERAL (SELECT pid, wait_event_type, wait_event FROM pg_catalog.pg_stat_activity WHERE backend_type = 'checkpointer' LIMIT 1) c
WHERE d.pid = $1 AND d.pid <> pg_backend_pid() AND d.state = 'active'
LIMIT 1`, backend).Scan(&snapshot.Type, &snapshot.Event, &snapshot.Blockers, &snapshot.CheckpointerType, &snapshot.CheckpointerEvent, &snapshot.CheckpointerBlockers, &snapshot.Requested, &snapshot.Done, &snapshot.Buffers, &snapshot.WriteMS, &snapshot.SyncMS, &snapshot.statsReset, &generation)
		if err == nil {
			snapshot.Barrier = sampleBarrierParticipant(probe, db, name, participant)
		}
		cancel()
		if err == nil {
			if !serviceGenerationPattern.MatchString(generation) || result.ServiceGeneration != nil && *result.ServiceGeneration != generation {
				return DropObservation{Status: "unavailable"}
			}
			result.ServiceGeneration = &generation
			if !validWaitType(snapshot.Type) {
				snapshot.Type = "unknown"
			}
			if !validWaitEvent(snapshot.Event) {
				snapshot.Event = "other"
			}
			if !validWaitType(snapshot.CheckpointerType) {
				snapshot.CheckpointerType = "unknown"
			}
			if !validWaitEvent(snapshot.CheckpointerEvent) {
				snapshot.CheckpointerEvent = "other"
			}
			if snapshot.valid() {
				if result.First == nil {
					first := snapshot
					result.First = &first
				}
				result.Last = &snapshot
				result.Status = "sampled"
				interval := checkpointInterval(result.First, result.Last)
				result.Checkpoint = &interval
			} else if result.First == nil {
				result.Status = "unavailable"
			}
		} else if ctx.Err() == nil && err != sql.ErrNoRows && result.First == nil {
			result.Status = "unavailable"
		}
	}
}

func DropWaitDescription(kind, event string, blockers int) string {
	if !validWaitType(kind) || blockers < 0 || blockers > 100000 {
		return "unknown_wait"
	}
	if !validWaitEvent(event) {
		event = "other"
	}
	return fmt.Sprintf("type=%s event=%s blockers=%d", kind, event, blockers)
}

func CheckpointProgressDescription(requested, done, buffers int64) string {
	if requested < 0 || done < 0 || buffers < 0 || requested > maxDiagnosticCounter || done > maxDiagnosticCounter || buffers > maxDiagnosticCounter {
		return "progress_unavailable"
	}
	// Cumulative counters do not identify the particular checkpoint requested by DROP.
	return fmt.Sprintf("requested=%d done=%d buffers=%d", requested, done, buffers)
}

func sampleBarrierParticipant(ctx context.Context, db *sql.DB, target string, participant func() int64) *BarrierWaitSnapshot {
	if participant == nil {
		return &BarrierWaitSnapshot{Status: "not_observed"}
	}
	pid := participant()
	if pid <= 0 {
		return &BarrierWaitSnapshot{Status: "not_observed"}
	}
	b := &BarrierWaitSnapshot{Status: "sampled"}
	var query string
	err := db.QueryRowContext(ctx, `SELECT backend_type, COALESCE(state, ''), COALESCE(wait_event_type, ''), COALESCE(wait_event, ''), cardinality(pg_blocking_pids(pid)), COALESCE(query, ''), CASE WHEN datname IS NULL THEN 'none' WHEN datname = current_database() THEN 'admin' WHEN datname = $2 THEN 'target' ELSE 'other' END, COALESCE(GREATEST(0, LEAST(86400000, floor(extract(epoch FROM (clock_timestamp() - query_start)) * 1000))), 0)::bigint FROM pg_catalog.pg_stat_activity WHERE pid = $1`, pid, target).Scan(&b.Backend, &b.State, &b.Type, &b.Event, &b.Blockers, &query, &b.DatabaseScope, &b.OperationAgeMS)
	if err != nil {
		return &BarrierWaitSnapshot{Status: "unavailable"}
	}
	b.Operation = barrierOperation(query)
	if !validBackend(b.Backend) {
		b.Backend = "other"
	}
	if !validBackendState(b.State) {
		b.State = "other"
	}
	if !validWaitType(b.Type) {
		b.Type = "unknown"
	}
	if !validWaitEvent(b.Event) {
		b.Event = "other"
	}
	if !b.valid() {
		return &BarrierWaitSnapshot{Status: "unavailable"}
	}
	return b
}
