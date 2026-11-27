package imports_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtureartifacts"
)

// A job attempt record keeps the evidence a load-dependent job failure needs
// (TH-HARNESS-REQ-864): the job's durable attempt history and, while it runs,
// the destination database's open transactions and lock waits. It reads only
// the job's attempt columns and pg_stat_activity state. No statement text,
// parameters, database or role names leave the database; each statement is
// reduced to its leading keyword from a closed list.
const (
	jobAttemptSampleInterval = 100 * time.Millisecond
	jobAttemptQueryTimeout   = 2 * time.Second
	jobAttemptSegmentLimit   = 800
	jobAttemptSessionLimit   = 32
)

var (
	jobAttemptVocabulary = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _()-]{0,63}$`)
	jobAttemptStatements = []string{"select", "insert", "update", "delete", "merge", "with", "copy", "lock", "begin", "start", "commit", "end", "rollback", "savepoint", "release", "prepare", "execute", "declare", "fetch", "close", "call", "do", "set", "reset", "show", "listen", "notify", "unlisten", "create", "alter", "drop", "truncate", "analyze"}
)

type jobAttemptEvidence struct {
	JobID                string                 `json:"job_id"`
	ObservedMS           int64                  `json:"observed_ms"`
	FinalStatus          string                 `json:"final_status"`
	FinalErrorCode       string                 `json:"final_error_code,omitempty"`
	Samples              int                    `json:"samples"`
	SampleErrors         int                    `json:"sample_errors"`
	LastSampleError      string                 `json:"last_sample_error,omitempty"`
	Attempts             []jobAttemptWindow     `json:"attempts"`
	Transactions         []jobAttemptSegment    `json:"transaction_segments"`
	OmittedSegmentCount  int                    `json:"omitted_segment_count"`
	AttemptTransactions  []jobAttemptPhaseTotal `json:"attempt_phases"`
	ObservationStartedAt time.Time              `json:"observation_started_at"`
	// MemberReads counts Incident Bundle member opens by phase and store
	// ("final_transaction"/"bundle" are reopens after the final-commit
	// protocol started, which Core 01 REQ-01-641 forbids).
	MemberReads map[string]map[string]int `json:"member_reads"`
	// BundlePathsAfterBegin counts bundle member opens in the final
	// transaction by owner path (blob members collapse to their prefix).
	BundlePathsAfterBegin map[string]int `json:"bundle_paths_after_begin"`
	// PayloadPathsAfterBegin counts verified opaque payload opens in the final
	// transaction by path class. They stream sealed bundle bytes through the
	// payload reader and are not row reads.
	PayloadPathsAfterBegin map[string]int `json:"payload_reads_after_begin"`
}

// jobAttemptWindow is one handler attempt. Offsets are milliseconds from the
// start of the observation; a negative claim offset predates it.
type jobAttemptWindow struct {
	Number            int    `json:"number"`
	ClaimedOffsetMS   int64  `json:"claimed_offset_ms"`
	FirstSeenOffsetMS int64  `json:"first_seen_offset_ms"`
	EndedOffsetMS     int64  `json:"ended_offset_ms"`
	DurationMS        int64  `json:"duration_ms"`
	LeaseRenewals     int    `json:"lease_renewals"`
	MaxRenewalGapMS   int64  `json:"max_renewal_gap_ms"`
	LeaseRemainingMS  int64  `json:"lease_remaining_at_end_ms"`
	RecordedReason    string `json:"recorded_reason"`
	FailureCountAfter int    `json:"failure_count_after"`
	StatusAfter       string `json:"status_after"`
	NextAttemptInMS   int64  `json:"next_attempt_in_ms,omitempty"`
	// LeaseHistory lists each lease expiry the attempt held, as offsets: when
	// it was first observed and when it expires.
	LeaseHistory []jobAttemptLease `json:"lease_history"`
}

type jobAttemptLease struct {
	ObservedMS int64 `json:"observed_ms"`
	ExpiresMS  int64 `json:"expires_ms"`
}

// jobAttemptSegment is one contiguous run of identical observations of one
// transaction: its backend, state, wait, statement kind and blockers.
type jobAttemptSegment struct {
	PID           int64   `json:"pid"`
	XactOffsetMS  int64   `json:"xact_start_offset_ms"`
	State         string  `json:"state"`
	WaitEventType string  `json:"wait_event_type"`
	WaitEvent     string  `json:"wait_event"`
	StatementKind string  `json:"statement_kind"`
	BlockingPIDs  []int64 `json:"blocking_pids"`
	FirstMS       int64   `json:"first_ms"`
	LastMS        int64   `json:"last_ms"`
}

// jobAttemptPhaseTotal summarizes the longest transaction observed in one
// attempt: when it began after the claim, how long it was seen, how long it
// was seen committing, and how long other sessions were seen blocked by it.
type jobAttemptPhaseTotal struct {
	Attempt               int            `json:"attempt"`
	TransactionPID        int64          `json:"transaction_pid"`
	TransactionStartAfter int64          `json:"transaction_start_after_claim_ms"`
	TransactionSeenMS     int64          `json:"transaction_seen_ms"`
	CommitSeenMS          int64          `json:"commit_seen_ms"`
	StatementSeenMS       map[string]int `json:"statement_seen_ms"`
	BlockedOthersMS       int64          `json:"blocked_others_ms"`
	TransactionWaitMS     int64          `json:"transaction_lock_wait_ms"`
	// TransactionEndMS is the last observation of the transaction, and
	// LeaseExpiryMS the attempt's lease expiry then in force, as offsets.
	TransactionEndMS int64 `json:"transaction_end_ms"`
	LeaseExpiryMS    int64 `json:"lease_expiry_at_transaction_end_ms"`
	// IdleInTransactionMS and ActiveMS split TransactionSeenMS by session state.
	IdleInTransactionMS int64 `json:"idle_in_transaction_ms"`
	ActiveMS            int64 `json:"active_ms"`
}

type jobAttemptRecorder struct {
	t        *testing.T
	db       *sql.DB
	jobID    string
	started  time.Time
	stop     chan struct{}
	finished chan struct{}

	mu             sync.Mutex
	evidence       jobAttemptEvidence
	current        *jobAttemptWindow
	currentAttempt string
	lastLease      time.Time
	lastRenew      time.Time
	open           map[string]int
	memberReads    *memberReadLog
}

// observeMemberReads attaches the destination server's member-read log.
func (r *jobAttemptRecorder) observeMemberReads(log *memberReadLog) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberReads = log
}

// memberReadLog keeps one destination server's Incident Bundle member opens per
// import operation (the import Job identity). It holds counts only.
type memberReadLog struct {
	mu                sync.Mutex
	byOperation       map[string]map[string]map[string]int
	lateByPath        map[string]map[string]int
	latePayloadByPath map[string]map[string]int
}

func newMemberReadLog() *memberReadLog {
	return &memberReadLog{byOperation: map[string]map[string]map[string]int{}, lateByPath: map[string]map[string]int{}, latePayloadByPath: map[string]map[string]int{}}
}

func (l *memberReadLog) observe(operationID, store, phase, path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	phases := l.byOperation[operationID]
	if phases == nil {
		phases = map[string]map[string]int{}
		l.byOperation[operationID] = phases
	}
	if phases[phase] == nil {
		phases[phase] = map[string]int{}
	}
	phases[phase][store]++
	if store == "bundle" && phase == "final_transaction" {
		if l.lateByPath[operationID] == nil {
			l.lateByPath[operationID] = map[string]int{}
		}
		l.lateByPath[operationID][memberReadPathClass(path)]++
	}
	if store == "payload" && phase == "final_transaction" {
		if l.latePayloadByPath[operationID] == nil {
			l.latePayloadByPath[operationID] = map[string]int{}
		}
		l.latePayloadByPath[operationID][memberReadPathClass(path)]++
	}
}

// requirePayloadReadsAfterBegin requires every final-transaction payload open
// to belong to one of classes, and each class to be read (Core 01
// REQ-01-641: opaque payloads stream from their sealed bundle member).
func (l *memberReadLog) requirePayloadReadsAfterBegin(t *testing.T, classes ...string) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	read := map[string]int{}
	for _, paths := range l.latePayloadByPath {
		for path, count := range paths {
			read[path] += count
		}
	}
	for path, count := range read {
		if !slices.Contains(classes, path) {
			t.Errorf("final transaction read payload %s %d times outside the verified payload classes", path, count)
		}
	}
	for _, class := range classes {
		if read[class] == 0 {
			t.Errorf("final transaction streamed no %s payload from its sealed bundle member", class)
		}
	}
}

func (l *memberReadLog) latePayloadsForOperation(operationID string) map[string]int {
	result := map[string]int{}
	if l == nil {
		return result
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for path, count := range l.latePayloadByPath[operationID] {
		result[path] = count
	}
	return result
}

// memberReadPathClass keeps data members' logical paths and collapses other
// members, such as blobs, to their first path segment.
func memberReadPathClass(path string) string {
	if strings.HasPrefix(path, "data/") {
		return path
	}
	if first, _, ok := strings.Cut(path, "/"); ok {
		return first + "/*"
	}
	return path
}

// bundlePathsAfterBegin returns the final-transaction bundle opens of every
// operation by path class.
func (l *memberReadLog) bundlePathsAfterBegin() map[string]int {
	result := map[string]int{}
	if l == nil {
		return result
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, paths := range l.lateByPath {
		for path, count := range paths {
			result[path] += count
		}
	}
	return result
}

// migratedBundlePaths are the path classes whose final transaction must
// reopen no bundle member. Every family is migrated (Core 01 REQ-01-641), so
// the empty prefix covers every bundle member.
var migratedBundlePaths = []string{""}

// requireNoBundleReadsAfterBegin fails when any final transaction reopened a
// bundle member whose path class starts with one of prefixes (Core 01
// REQ-01-641).
func (l *memberReadLog) requireNoBundleReadsAfterBegin(t *testing.T, prefixes ...string) {
	t.Helper()
	for path, count := range l.bundlePathsAfterBegin() {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				t.Errorf("final transaction reopened bundle member %s %d times", path, count)
			}
		}
	}
}

func (l *memberReadLog) forOperation(operationID string) map[string]map[string]int {
	result := map[string]map[string]int{}
	if l == nil {
		return result
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for phase, stores := range l.byOperation[operationID] {
		result[phase] = map[string]int{}
		for store, count := range stores {
			result[phase][store] = count
		}
	}
	return result
}

func (l *memberReadLog) lateForOperation(operationID string) map[string]int {
	result := map[string]int{}
	if l == nil {
		return result
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for path, count := range l.lateByPath[operationID] {
		result[path] = count
	}
	return result
}

// recordJobAttempts samples the job's attempts in db until the returned
// recorder is finished.
func recordJobAttempts(t *testing.T, db *sql.DB, jobID string) *jobAttemptRecorder {
	t.Helper()
	r := &jobAttemptRecorder{
		t: t, db: db, jobID: jobID, started: time.Now(), stop: make(chan struct{}), finished: make(chan struct{}),
		open: map[string]int{},
	}
	r.evidence = jobAttemptEvidence{JobID: jobID, ObservationStartedAt: r.started.UTC(), Attempts: []jobAttemptWindow{}, Transactions: []jobAttemptSegment{}, AttemptTransactions: []jobAttemptPhaseTotal{}}
	go r.run()
	// A test that ends before finishing still retains what was observed.
	t.Cleanup(func() { r.finish(nil) })
	return r
}

func (r *jobAttemptRecorder) run() {
	defer close(r.finished)
	ticker := time.NewTicker(jobAttemptSampleInterval)
	defer ticker.Stop()
	for {
		r.sample()
		select {
		case <-r.stop:
			r.sample()
			return
		case <-ticker.C:
		}
	}
}

func (r *jobAttemptRecorder) offset(at time.Time) int64 { return at.Sub(r.started).Milliseconds() }

func (r *jobAttemptRecorder) sample() {
	ctx, cancel := context.WithTimeout(context.Background(), jobAttemptQueryTimeout)
	defer cancel()
	now := time.Now()
	var status, attemptID, lastError string
	var failures int
	var claimed, lease, next sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT status, COALESCE(handler_attempt_id::text, ''), handler_failure_count, COALESCE(handler_last_error, ''),
       handler_last_attempted_at, handler_lease_expires_at, handler_next_attempt_at
  FROM jobs WHERE job_id = $1::uuid`, r.jobID).Scan(&status, &attemptID, &failures, &lastError, &claimed, &lease, &next)
	sessions, sessionErr := r.sessions(ctx)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evidence.Samples++
	for _, failure := range []error{err, sessionErr} {
		if failure != nil {
			r.evidence.SampleErrors++
			r.evidence.LastSampleError = boundedJobAttemptError(failure)
		}
	}
	at := r.offset(now)
	if sessionErr == nil {
		r.recordSessions(at, sessions)
	}
	if err != nil {
		return
	}
	if r.current != nil && attemptID != r.currentAttempt {
		r.closeAttempt(at, status, lastError, failures, next, now)
	}
	if attemptID != "" && r.current == nil {
		window := jobAttemptWindow{Number: len(r.evidence.Attempts) + 1, FirstSeenOffsetMS: at, ClaimedOffsetMS: at, LeaseHistory: []jobAttemptLease{}}
		if claimed.Valid {
			window.ClaimedOffsetMS = r.offset(claimed.Time)
		}
		r.evidence.Attempts = append(r.evidence.Attempts, window)
		r.current = &r.evidence.Attempts[len(r.evidence.Attempts)-1]
		r.currentAttempt = attemptID
		r.lastLease, r.lastRenew = time.Time{}, now
		if claimed.Valid {
			r.lastRenew = claimed.Time
		}
	}
	if r.current != nil && lease.Valid {
		if !lease.Time.Equal(r.lastLease) && len(r.current.LeaseHistory) < 64 {
			r.current.LeaseHistory = append(r.current.LeaseHistory, jobAttemptLease{ObservedMS: at, ExpiresMS: r.offset(lease.Time)})
		}
		if !r.lastLease.IsZero() && lease.Time.After(r.lastLease) {
			r.current.LeaseRenewals++
			r.lastRenew = now
		}
		r.lastLease = lease.Time
		if gap := now.Sub(r.lastRenew).Milliseconds(); gap > r.current.MaxRenewalGapMS {
			r.current.MaxRenewalGapMS = gap
		}
		r.current.LeaseRemainingMS = lease.Time.Sub(now).Milliseconds()
	}
}

func (r *jobAttemptRecorder) closeAttempt(at int64, status, lastError string, failures int, next sql.NullTime, now time.Time) {
	r.current.EndedOffsetMS = at
	r.current.DurationMS = at - r.current.ClaimedOffsetMS
	r.current.RecordedReason = lastError
	r.current.FailureCountAfter = failures
	r.current.StatusAfter = status
	if next.Valid {
		r.current.NextAttemptInMS = next.Time.Sub(now).Milliseconds()
	}
	r.current, r.currentAttempt = nil, ""
}

type jobAttemptSession struct {
	pid                                  int64
	xactMS                               int64
	state, waitType, wait, kind, blocked string
	blockers                             []int64
}

func (r *jobAttemptRecorder) sessions(ctx context.Context) ([]jobAttemptSession, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT a.pid, COALESCE(a.state, ''), COALESCE(a.wait_event_type, ''), COALESCE(a.wait_event, ''),
       CASE WHEN COALESCE(a.query, '') = '' THEN ''
            ELSE COALESCE(lower(substring(ltrim(a.query, E' \t\r\n(') FROM '^[A-Za-z]{1,16}')), '?') END,
       COALESCE(array_to_string(pg_catalog.pg_blocking_pids(a.pid), ','), ''),
       floor(extract(epoch FROM a.xact_start) * 1000)::bigint
  FROM pg_catalog.pg_stat_activity a
 WHERE a.datname = current_database() AND a.pid <> pg_catalog.pg_backend_pid()
   AND a.xact_start IS NOT NULL
   AND (clock_timestamp() - a.xact_start > interval '500 milliseconds'
        OR cardinality(pg_catalog.pg_blocking_pids(a.pid)) > 0)
 ORDER BY a.pid
 LIMIT $1`, jobAttemptSessionLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []jobAttemptSession
	for rows.Next() {
		var s jobAttemptSession
		var blockers string
		if err := rows.Scan(&s.pid, &s.state, &s.waitType, &s.wait, &s.kind, &blockers, &s.xactMS); err != nil {
			return nil, err
		}
		s.state, s.waitType, s.wait = jobAttemptName(s.state), jobAttemptName(s.waitType), jobAttemptName(s.wait)
		switch {
		case s.kind == "":
			s.kind = "none"
		case slices.Contains(jobAttemptStatements, s.kind):
		default:
			s.kind = "other"
		}
		s.blockers = []int64{}
		for _, field := range strings.Split(blockers, ",") {
			if pid, err := strconv.ParseInt(field, 10, 64); err == nil {
				s.blockers = append(s.blockers, pid)
			}
		}
		s.blocked = blockers
		result = append(result, s)
	}
	return result, rows.Err()
}

func jobAttemptName(value string) string {
	if value == "" || jobAttemptVocabulary.MatchString(value) {
		return value
	}
	return "other"
}

func (r *jobAttemptRecorder) recordSessions(at int64, sessions []jobAttemptSession) {
	seen := map[string]int{}
	for _, s := range sessions {
		xact := s.xactMS - r.started.UnixMilli()
		key := fmt.Sprintf("%d/%d/%s/%s/%s/%s/%s", s.pid, xact, s.state, s.waitType, s.wait, s.kind, s.blocked)
		if index, ok := r.open[key]; ok {
			r.evidence.Transactions[index].LastMS = at
			seen[key] = index
			continue
		}
		if len(r.evidence.Transactions) >= jobAttemptSegmentLimit {
			r.evidence.OmittedSegmentCount++
			continue
		}
		r.evidence.Transactions = append(r.evidence.Transactions, jobAttemptSegment{
			PID: s.pid, XactOffsetMS: xact, State: s.state, WaitEventType: s.waitType, WaitEvent: s.wait,
			StatementKind: s.kind, BlockingPIDs: s.blockers, FirstMS: at, LastMS: at,
		})
		seen[key] = len(r.evidence.Transactions) - 1
	}
	r.open = seen
}

// finish stops sampling, writes the record into the run root and, unless the
// job succeeded, logs a per-attempt summary with the test's failure output.
func (r *jobAttemptRecorder) finish(job map[string]any) {
	select {
	case <-r.stop:
		return
	default:
		close(r.stop)
	}
	<-r.finished
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.current != nil {
		r.current.EndedOffsetMS = r.offset(time.Now())
		r.current.DurationMS = r.current.EndedOffsetMS - r.current.ClaimedOffsetMS
		r.current = nil
	}
	r.evidence.ObservedMS = r.offset(time.Now())
	if job != nil {
		r.evidence.FinalStatus, _ = job["status"].(string)
		if summary, ok := job["error_summary"].(map[string]any); ok {
			r.evidence.FinalErrorCode, _ = summary["code"].(string)
		}
	}
	r.evidence.AttemptTransactions = r.phases()
	r.evidence.MemberReads = r.memberReads.forOperation(r.jobID)
	r.evidence.BundlePathsAfterBegin = r.memberReads.lateForOperation(r.jobID)
	r.evidence.PayloadPathsAfterBegin = r.memberReads.latePayloadsForOperation(r.jobID)
	name := "job-attempts-" + r.jobID + ".json"
	location := "not recorded"
	if root, err := fixtureartifacts.SelectedUnitRoot(); err != nil {
		location = "not recorded: " + boundedJobAttemptError(err)
	} else if data, err := json.MarshalIndent(r.evidence, "", "  "); err != nil {
		location = "not recorded: " + boundedJobAttemptError(err)
	} else if err := os.WriteFile(filepath.Join(root, name), append(data, '\n'), 0o600); err != nil {
		location = "not recorded: " + boundedJobAttemptError(err)
	} else {
		location = name
	}
	if r.evidence.FinalStatus == "succeeded" {
		return
	}
	r.t.Logf("job %s attempts (%s): %s", r.jobID, location, r.summary())
}

func (r *jobAttemptRecorder) phases() []jobAttemptPhaseTotal {
	totals := []jobAttemptPhaseTotal{}
	for _, attempt := range r.evidence.Attempts {
		// The attempt's transaction is the longest-lived one seen in its window.
		type xact struct {
			pid    int64
			start  int64
			last   int64
			commit int64
			waited int64
			kinds  map[string]int
		}
		byXact := map[string]*xact{}
		for _, segment := range r.evidence.Transactions {
			if segment.LastMS < attempt.ClaimedOffsetMS || segment.FirstMS > attempt.EndedOffsetMS {
				continue
			}
			key := fmt.Sprintf("%d/%d", segment.PID, segment.XactOffsetMS)
			entry := byXact[key]
			if entry == nil {
				entry = &xact{pid: segment.PID, start: segment.XactOffsetMS, last: segment.LastMS, kinds: map[string]int{}}
				byXact[key] = entry
			}
			entry.last = max(entry.last, segment.LastMS)
			seen := segment.LastMS - segment.FirstMS + jobAttemptSampleInterval.Milliseconds()
			entry.kinds[segment.StatementKind+"/"+segment.State] += int(seen)
			if segment.StatementKind == "commit" && segment.State == "active" {
				entry.commit += seen
			}
			if len(segment.BlockingPIDs) > 0 {
				entry.waited += seen
			}
		}
		var longest *xact
		for _, entry := range byXact {
			if longest == nil || entry.last-entry.start > longest.last-longest.start {
				longest = entry
			}
		}
		total := jobAttemptPhaseTotal{Attempt: attempt.Number, StatementSeenMS: map[string]int{}}
		if longest != nil {
			total.TransactionPID = longest.pid
			total.TransactionStartAfter = longest.start - attempt.ClaimedOffsetMS
			total.TransactionSeenMS = longest.last - longest.start
			total.CommitSeenMS = longest.commit
			total.StatementSeenMS = longest.kinds
			total.TransactionWaitMS = longest.waited
			total.TransactionEndMS = longest.last
			for kind, seen := range longest.kinds {
				if _, state, ok := strings.Cut(kind, "/"); ok && strings.HasPrefix(state, "idle in transaction") {
					total.IdleInTransactionMS += int64(seen)
				} else if ok && state == "active" {
					total.ActiveMS += int64(seen)
				}
			}
			for _, lease := range attempt.LeaseHistory {
				if lease.ObservedMS <= longest.last {
					total.LeaseExpiryMS = lease.ExpiresMS
				}
			}
			for _, segment := range r.evidence.Transactions {
				if segment.LastMS < attempt.ClaimedOffsetMS || segment.FirstMS > attempt.EndedOffsetMS {
					continue
				}
				if slices.Contains(segment.BlockingPIDs, longest.pid) {
					total.BlockedOthersMS += segment.LastMS - segment.FirstMS + jobAttemptSampleInterval.Milliseconds()
				}
			}
		}
		totals = append(totals, total)
	}
	return totals
}

func (r *jobAttemptRecorder) summary() string {
	parts := []string{fmt.Sprintf("final=%s/%s samples=%d sample_errors=%d bundle_reads_after_begin=%d", r.evidence.FinalStatus, r.evidence.FinalErrorCode, r.evidence.Samples, r.evidence.SampleErrors, r.evidence.MemberReads["final_transaction"]["bundle"])}
	for index, attempt := range r.evidence.Attempts {
		line := fmt.Sprintf("attempt %d: claimed=%dms ended=%dms duration=%dms reason=%q failures=%d renewals=%d max_renewal_gap=%dms lease_left=%dms",
			attempt.Number, attempt.ClaimedOffsetMS, attempt.EndedOffsetMS, attempt.DurationMS, attempt.RecordedReason, attempt.FailureCountAfter,
			attempt.LeaseRenewals, attempt.MaxRenewalGapMS, attempt.LeaseRemainingMS)
		if index < len(r.evidence.AttemptTransactions) {
			phase := r.evidence.AttemptTransactions[index]
			line += fmt.Sprintf(" tx_start_after_claim=%dms tx_seen=%dms idle_in_tx=%dms tx_end=%dms lease_expiry=%dms commit_seen=%dms tx_lock_wait=%dms blocked_others=%dms",
				phase.TransactionStartAfter, phase.TransactionSeenMS, phase.IdleInTransactionMS, phase.TransactionEndMS, phase.LeaseExpiryMS, phase.CommitSeenMS, phase.TransactionWaitMS, phase.BlockedOthersMS)
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, "; ")
}

func boundedJobAttemptError(err error) string {
	text := err.Error()
	if len(text) > 160 {
		text = text[:160]
	}
	return text
}
