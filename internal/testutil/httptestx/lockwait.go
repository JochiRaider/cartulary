package httptestx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
)

// A lock-wait snapshot (TH-HARNESS-REQ-864) lists a test database's blocked
// sessions and the sessions blocking them. It is bounded and carries no
// statement text, parameters, database or role names: each statement is only
// its leading keyword, mapped to a closed kind.
const (
	lockWaitDirectory    = "lock-waits"
	lockWaitSessionLimit = 64
	lockWaitQueryTimeout = 2 * time.Second
)

var (
	lockWaitSequence atomic.Int64
	// pg_stat_activity names are server vocabulary; anything else is "other".
	lockWaitName       = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _()-]{0,63}$`)
	lockStatementKinds = []string{"select", "insert", "update", "delete", "merge", "with", "copy", "lock", "begin", "start", "commit", "end", "rollback", "savepoint", "release", "prepare", "execute", "deallocate", "declare", "fetch", "close", "call", "do", "set", "reset", "show", "listen", "notify", "unlisten", "create", "alter", "drop", "truncate", "grant", "revoke", "comment", "vacuum", "analyze", "reindex", "cluster", "refresh", "checkpoint", "discard"}
)

// LockWaitSession is one blocked or blocking session.
type LockWaitSession struct {
	PID           int64   `json:"pid"`
	BackendType   string  `json:"backend_type"`
	State         string  `json:"state"`
	WaitEventType string  `json:"wait_event_type"`
	WaitEvent     string  `json:"wait_event"`
	StatementKind string  `json:"statement_kind"`
	BlockingPIDs  []int64 `json:"blocking_pids"`
	QueryAgeMS    int64   `json:"query_age_ms"`
}

// LockWaitSnapshot is one bounded observation of the current database.
type LockWaitSnapshot struct {
	CapturedAt          time.Time         `json:"captured_at"`
	Sessions            []LockWaitSession `json:"sessions"`
	OmittedSessionCount int               `json:"omitted_session_count"`
}

// Blocked counts the listed sessions that wait on another session.
func (s LockWaitSnapshot) Blocked() int {
	blocked := 0
	for _, session := range s.Sessions {
		if len(session.BlockingPIDs) > 0 {
			blocked++
		}
	}
	return blocked
}

// SnapshotLockWaits reads the blocked sessions of db's current database and
// the sessions blocking them, blocked sessions first, at most 64.
func SnapshotLockWaits(ctx context.Context, db *sql.DB) (LockWaitSnapshot, error) {
	if db == nil {
		return LockWaitSnapshot{}, errors.New("lock-wait snapshot requires a database")
	}
	rows, err := db.QueryContext(ctx, `
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
 LIMIT $1`, lockWaitSessionLimit)
	if err != nil {
		return LockWaitSnapshot{}, err
	}
	defer rows.Close()
	snapshot := LockWaitSnapshot{CapturedAt: time.Now().UTC(), Sessions: []LockWaitSession{}}
	total := 0
	for rows.Next() {
		var session LockWaitSession
		var kind, blockers string
		if err := rows.Scan(&session.PID, &session.BackendType, &session.State, &session.WaitEventType, &session.WaitEvent, &kind, &blockers, &session.QueryAgeMS, &total); err != nil {
			return LockWaitSnapshot{}, err
		}
		session.BackendType, session.State = lockWaitVocabulary(session.BackendType), lockWaitVocabulary(session.State)
		session.WaitEventType, session.WaitEvent = lockWaitVocabulary(session.WaitEventType), lockWaitVocabulary(session.WaitEvent)
		// Only the leading keyword leaves the database; an unlisted or
		// unreadable one (such as a leading comment) is "other".
		session.StatementKind = "none"
		if kind != "" {
			session.StatementKind = "other"
			if slices.Contains(lockStatementKinds, kind) {
				session.StatementKind = kind
			}
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
		return LockWaitSnapshot{}, err
	}
	snapshot.OmittedSessionCount = max(0, total-len(snapshot.Sessions))
	return snapshot, nil
}

func lockWaitVocabulary(value string) string {
	if value == "" || lockWaitName.MatchString(value) {
		return value
	}
	return "other"
}

// RecordLockWaits snapshots db's lock waits and writes the snapshot under the
// run root, resolved by the rule shared with contract evidence and fixture
// artifacts, as lock-waits/<pid>-<n>.json. It returns the run-root-relative
// path. It never creates the run root and never follows a link out of it.
func RecordLockWaits(ctx context.Context, db *sql.DB) (string, LockWaitSnapshot, error) {
	return recordLockWaitsTo(ctx, db, os.LookupEnv)
}

func recordLockWaitsTo(ctx context.Context, db *sql.DB, lookup func(string) (string, bool)) (string, LockWaitSnapshot, error) {
	snapshot, err := SnapshotLockWaits(ctx, db)
	if err != nil {
		return "", LockWaitSnapshot{}, err
	}
	runRoot, err := runevidence.RunRoot(lookup)
	if err != nil {
		return "", snapshot, err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", snapshot, err
	}
	name := fmt.Sprintf("%s/%d-%d-%d.json", lockWaitDirectory, os.Getpid(), time.Now().UnixNano(), lockWaitSequence.Add(1))
	if err := writeRunRootFile(runRoot, name, append(data, '\n')); err != nil {
		return "", snapshot, err
	}
	return name, snapshot, nil
}

func writeRunRootFile(runRoot, name string, data []byte) error {
	info, err := os.Lstat(runRoot)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !ownedByCaller(info) {
		return errors.New("unsafe lock-wait run root")
	}
	root, err := os.OpenRoot(runRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Mkdir(lockWaitDirectory, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	directory, err := root.Lstat(lockWaitDirectory)
	if err != nil {
		return err
	}
	if !directory.IsDir() || directory.Mode()&os.ModeSymlink != 0 || directory.Mode().Perm() != 0o700 || !ownedByCaller(directory) {
		return errors.New("unsafe lock-wait directory")
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return errors.Join(err, file.Close())
}

func ownedByCaller(info fs.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}

// describeLockWaits records a snapshot after a poll's deadline and returns a
// one-line summary for the failure message.
func describeLockWaits(db *sql.DB) string {
	ctx, cancel := context.WithTimeout(context.Background(), lockWaitQueryTimeout)
	defer cancel()
	path, snapshot, err := RecordLockWaits(ctx, db)
	switch {
	case err != nil && snapshot.Sessions == nil:
		return fmt.Sprintf("lock-wait snapshot unavailable: %v", err)
	case err != nil:
		return fmt.Sprintf("lock waits: %d sessions, %d blocked (not recorded: %v)", len(snapshot.Sessions), snapshot.Blocked(), err)
	default:
		return fmt.Sprintf("lock waits: %d sessions, %d blocked, recorded in %s", len(snapshot.Sessions), snapshot.Blocked(), path)
	}
}
