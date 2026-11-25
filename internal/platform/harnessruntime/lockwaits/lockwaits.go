// Package lockwaits owns the bounded lock-wait snapshot of TH-HARNESS-REQ-864:
// a database's blocked sessions and the sessions blocking them. It carries no
// statement text, parameters, database or role names. The query reduces each
// statement to its leading keyword within a closed kind before any text leaves
// the database. Test support and the harness server's stall diagnostics both
// read it here, so the query and its vocabulary have one owner.
package lockwaits

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	// SessionLimit bounds a snapshot.
	SessionLimit = 64
	// QueryTimeout bounds one snapshot query.
	QueryTimeout = 2 * time.Second
)

// query lists the current database's blocked sessions and their blockers,
// blocked sessions first. statement_kind is "none" without a statement, the
// leading keyword when it is in the closed list, and "other" otherwise (such
// as a leading comment).
const query = `
WITH blocked AS (
  SELECT pid, pg_catalog.pg_blocking_pids(pid) AS blockers
    FROM pg_catalog.pg_stat_activity
   WHERE datname = current_database() AND pid <> pg_catalog.pg_backend_pid()
     AND cardinality(pg_catalog.pg_blocking_pids(pid)) > 0
), involved AS (
  SELECT pid FROM blocked UNION SELECT unnest(blockers) FROM blocked
), described AS (
  SELECT a.pid, a.backend_type, a.state, a.wait_event_type, a.wait_event, a.query_start,
         COALESCE(a.query, '') = '' AS idle_statement,
         lower(substring(ltrim(COALESCE(a.query, ''), E' \t\r\n(') FROM '^[A-Za-z]{1,16}')) AS keyword
    FROM involved i
    JOIN pg_catalog.pg_stat_activity a ON a.pid = i.pid
)
SELECT d.pid, COALESCE(d.backend_type, ''), COALESCE(d.state, ''),
       COALESCE(d.wait_event_type, ''), COALESCE(d.wait_event, ''),
       CASE WHEN d.idle_statement THEN 'none'
            WHEN d.keyword = ANY (ARRAY['select','insert','update','delete','merge','with','copy','lock','begin','start','commit','end','rollback','savepoint','release','prepare','execute','deallocate','declare','fetch','close','call','do','set','reset','show','listen','notify','unlisten','create','alter','drop','truncate','grant','revoke','comment','vacuum','analyze','reindex','cluster','refresh','checkpoint','discard']) THEN d.keyword
            ELSE 'other' END,
       COALESCE(array_to_string(b.blockers, ','), ''),
       COALESCE(GREATEST(0, LEAST(86400000, floor(extract(epoch FROM (clock_timestamp() - d.query_start)) * 1000))), 0)::bigint,
       count(*) OVER ()
  FROM described d
  LEFT JOIN blocked b ON b.pid = d.pid
 ORDER BY b.pid IS NULL, d.pid
 LIMIT $1`

// pg_stat_activity names are server vocabulary; anything else is "other".
var (
	vocabulary = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _()-]{0,63}$`)
	kind       = regexp.MustCompile(`^[a-z]{1,16}$`)
)

// Session is one blocked or blocking session.
type Session struct {
	PID           int64   `json:"pid"`
	BackendType   string  `json:"backend_type"`
	State         string  `json:"state"`
	WaitEventType string  `json:"wait_event_type"`
	WaitEvent     string  `json:"wait_event"`
	StatementKind string  `json:"statement_kind"`
	BlockingPIDs  []int64 `json:"blocking_pids"`
	QueryAgeMS    int64   `json:"query_age_ms"`
}

// Snapshot is one bounded observation of a database.
type Snapshot struct {
	CapturedAt          time.Time `json:"captured_at"`
	Sessions            []Session `json:"sessions"`
	OmittedSessionCount int       `json:"omitted_session_count"`
}

// Blocked counts the listed sessions that wait on another session.
func (s Snapshot) Blocked() int {
	blocked := 0
	for _, session := range s.Sessions {
		if len(session.BlockingPIDs) > 0 {
			blocked++
		}
	}
	return blocked
}

// SnapshotDB reads db's current database through database/sql.
func SnapshotDB(ctx context.Context, db *sql.DB) (Snapshot, error) {
	if db == nil {
		return Snapshot{}, errors.New("lock-wait snapshot requires a database")
	}
	rows, err := db.QueryContext(ctx, query, SessionLimit)
	if err != nil {
		return Snapshot{}, err
	}
	defer rows.Close()
	return collect(rows, time.Now())
}

// Querier is a pgx connection, pool or transaction.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// SnapshotConn reads the connection's current database through pgx.
func SnapshotConn(ctx context.Context, connection Querier) (Snapshot, error) {
	if connection == nil {
		return Snapshot{}, errors.New("lock-wait snapshot requires a database")
	}
	rows, err := connection.Query(ctx, query, SessionLimit)
	if err != nil {
		return Snapshot{}, err
	}
	defer rows.Close()
	return collect(rows, time.Now())
}

type cursor interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func collect(rows cursor, capturedAt time.Time) (Snapshot, error) {
	snapshot := Snapshot{CapturedAt: capturedAt.UTC(), Sessions: []Session{}}
	total := 0
	for rows.Next() {
		var session Session
		var blockers string
		if err := rows.Scan(&session.PID, &session.BackendType, &session.State, &session.WaitEventType, &session.WaitEvent, &session.StatementKind, &blockers, &session.QueryAgeMS, &total); err != nil {
			return Snapshot{}, err
		}
		session.BackendType, session.State = name(session.BackendType), name(session.State)
		session.WaitEventType, session.WaitEvent = name(session.WaitEventType), name(session.WaitEvent)
		if !kind.MatchString(session.StatementKind) {
			session.StatementKind = "other"
		}
		session.BlockingPIDs = []int64{}
		for _, field := range strings.Split(blockers, ",") {
			if pid, err := strconv.ParseInt(field, 10, 64); err == nil {
				session.BlockingPIDs = append(session.BlockingPIDs, pid)
			}
		}
		snapshot.Sessions = append(snapshot.Sessions, session)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, err
	}
	snapshot.OmittedSessionCount = max(0, total-len(snapshot.Sessions))
	return snapshot, nil
}

func name(value string) string {
	if value == "" || vocabulary.MatchString(value) {
		return value
	}
	return "other"
}
