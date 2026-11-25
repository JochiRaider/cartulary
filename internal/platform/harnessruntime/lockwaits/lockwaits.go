// Package lockwaits defines the bounded lock-wait snapshot of TH-HARNESS-REQ-864:
// a database's blocked sessions and the sessions blocking them. It carries no
// statement text, parameters, database or role names: each statement is only
// its leading keyword, mapped to a closed kind. Test support and the harness
// server's stall diagnostics share this one definition.
package lockwaits

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// SessionLimit bounds a snapshot; Query takes it as its only parameter.
	SessionLimit = 64
	// QueryTimeout bounds one snapshot query.
	QueryTimeout = 2 * time.Second
)

// Query lists the current database's blocked sessions and their blockers,
// blocked sessions first. Its columns are Collect's scan order.
const Query = `
WITH blocked AS (
  SELECT pid, pg_catalog.pg_blocking_pids(pid) AS blockers
    FROM pg_catalog.pg_stat_activity
   WHERE datname = current_database() AND pid <> pg_catalog.pg_backend_pid()
     AND cardinality(pg_catalog.pg_blocking_pids(pid)) > 0
), involved AS (
  SELECT pid FROM blocked UNION SELECT unnest(blockers) FROM blocked
)
SELECT a.pid, COALESCE(a.backend_type, ''), COALESCE(a.state, ''),
       COALESCE(a.wait_event_type, ''), COALESCE(a.wait_event, ''),
       CASE WHEN COALESCE(a.query, '') = '' THEN ''
            ELSE COALESCE(lower(substring(ltrim(a.query, E' \t\r\n(') FROM '^[A-Za-z]{1,16}')), '?') END,
       COALESCE(array_to_string(b.blockers, ','), ''),
       COALESCE(GREATEST(0, LEAST(86400000, floor(extract(epoch FROM (clock_timestamp() - a.query_start)) * 1000))), 0)::bigint,
       count(*) OVER ()
  FROM involved i
  JOIN pg_catalog.pg_stat_activity a ON a.pid = i.pid
  LEFT JOIN blocked b ON b.pid = a.pid
 ORDER BY b.pid IS NULL, a.pid
 LIMIT $1`

var (
	// pg_stat_activity names are server vocabulary; anything else is "other".
	vocabulary     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _()-]{0,63}$`)
	statementKinds = []string{"select", "insert", "update", "delete", "merge", "with", "copy", "lock", "begin", "start", "commit", "end", "rollback", "savepoint", "release", "prepare", "execute", "deallocate", "declare", "fetch", "close", "call", "do", "set", "reset", "show", "listen", "notify", "unlisten", "create", "alter", "drop", "truncate", "grant", "revoke", "comment", "vacuum", "analyze", "reindex", "cluster", "refresh", "checkpoint", "discard"}
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

// Rows is the row cursor both database/sql and pgx return for Query.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

// Collect reads Query's rows into a snapshot, reducing every name to the
// closed vocabulary. The caller owns and closes rows.
func Collect(rows Rows, capturedAt time.Time) (Snapshot, error) {
	snapshot := Snapshot{CapturedAt: capturedAt.UTC(), Sessions: []Session{}}
	total := 0
	for rows.Next() {
		var session Session
		var kind, blockers string
		if err := rows.Scan(&session.PID, &session.BackendType, &session.State, &session.WaitEventType, &session.WaitEvent, &kind, &blockers, &session.QueryAgeMS, &total); err != nil {
			return Snapshot{}, err
		}
		session.BackendType, session.State = name(session.BackendType), name(session.State)
		session.WaitEventType, session.WaitEvent = name(session.WaitEventType), name(session.WaitEvent)
		session.StatementKind = statementKind(kind)
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

// Only the leading keyword leaves the database; an unlisted or unreadable one
// (such as a leading comment) is "other".
func statementKind(kind string) string {
	if kind == "" {
		return "none"
	}
	if slices.Contains(statementKinds, kind) {
		return kind
	}
	return "other"
}
