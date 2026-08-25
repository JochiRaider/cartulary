package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/intenttest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestCompactExpiredJobs_Integration(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "jobs-compact-expired")
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	cutoff := time.Date(2026, 8, 8, 17, 0, 0, 0, time.UTC)
	manager := &Manager{pool: pool, now: func() time.Time { return cutoff }, policy: ProductionRuntimePolicy()}
	actorID := uuid.New()
	incidentID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Expiry Actor', 'hash', false, true, true)
`, actorID, actorID.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $2, 'Expiry Incident', 'active', $3, $3)
`, incidentID, "expiry-"+incidentID.String(), actorID); err != nil {
		t.Fatal(err)
	}
	specialID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, started_at,
    finished_at, retained_until, result_summary_json, message, auth_policy,
    handler_name, handler_payload_json, handler_failure_count,
    handler_last_attempted_at, handler_last_error, job_kind, progress_unit_id,
    extension_owner_profile_id, extension_idempotency_identity,
    extension_idempotency_route_key, extension_idempotency_scope_key,
    extension_normalized_request_sha256
) VALUES (
    $1, 'incident', $2, 'succeeded', false, $3,
    $4::timestamptz - interval '8 days', $4::timestamptz - interval '8 days', 1, 1, $4::timestamptz - interval '8 days',
    $4::timestamptz - interval '7 days', $4::timestamptz - interval '2 hours',
    '{"code":"private_result","message":"private result"}', 'private message',
    'incident_membership', 'expiry.worker_v1', '{"secret":"payload"}', 2,
    $4::timestamptz - interval '7 days', 'private_diagnostic',
    'expiry.run_v1', 'expiry.run.attempt.v1', 'expiry',
    '{"secret":"idempotency"}', 'expiry.run', 'incident:private', repeat('a', 64)
)
`, specialID, incidentID, actorID, cutoff); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO extension_job_commit_proofs (
    job_id, owner_profile_id, operation_kind, final_commit_id,
    idempotency_identity, normalized_request_sha256, terminal_result,
    terminal_result_sha256, resource_refs, committed_at
) VALUES (
    $1, 'expiry', 'expiry.run', $2, '{"proof":"retained"}', repeat('a', 64),
    '{"result":"retained"}', repeat('b', 64), '[]', $3::timestamptz - interval '7 days'
)
`, specialID, "expiry-proof-"+specialID.String(), cutoff); err != nil {
		t.Fatal(err)
	}
	intenttest.InsertPersistedJobProgressIntentFixture(
		t,
		pool,
		"expiry-intent-"+specialID.String(),
		incidentID,
		[]byte(`{"status":"succeeded"}`),
		"job:"+specialID.String(),
		cutoff.Add(-7*24*time.Hour),
	)
	if _, err := pool.Exec(ctx, `
INSERT INTO jobs (
    scope_kind, status, cancelable, submitted_by_user_id, submitted_at, updated_at,
    progress_completed, finished_at, retained_until, result_summary_json,
    auth_policy, handler_name, job_kind, progress_unit_id
)
SELECT 'deployment', 'succeeded', false, $1,
       $2::timestamptz - interval '8 days', $2::timestamptz - interval '7 days', 1,
       $2::timestamptz - interval '7 days', $2::timestamptz - interval '1 hour',
       '{"code":"done","message":"Done."}', 'deployment_admin',
       'expiry.worker_v1', 'expiry.run_v1', 'expiry.run.attempt.v1'
  FROM generate_series(1, 1000)
`, actorID, cutoff); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO jobs (
    scope_kind, status, cancelable, submitted_by_user_id, submitted_at, updated_at,
    progress_completed, finished_at, retained_until, result_summary_json,
    auth_policy, handler_name, job_kind, progress_unit_id
) VALUES (
    'deployment', 'succeeded', false, $1, $2::timestamptz, $2::timestamptz, 1,
    $2::timestamptz, $2::timestamptz + interval '1 hour',
    '{"code":"future","message":"Future."}', 'deployment_admin',
    'expiry.worker_v1', 'expiry.run_v1', 'expiry.run.attempt.v1'
)
`, actorID, cutoff); err != nil {
		t.Fatal(err)
	}

	blocker, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var locked bool
	if err := blocker.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, expirySweepLockKey).Scan(&locked); err != nil || !locked {
		_ = blocker.Rollback(ctx)
		t.Fatalf("hold concurrent expiry sweep lock = %t, %v", locked, err)
	}
	compacted, err := manager.compactExpiredJobs(ctx, 1000)
	if err != nil || compacted != 0 {
		_ = blocker.Rollback(ctx)
		t.Fatalf("concurrent compaction = %d, %v; want neutral 0", compacted, err)
	}
	if err := blocker.Rollback(ctx); err != nil {
		t.Fatal(err)
	}

	compacted, err = manager.compactExpiredJobs(ctx, 1000)
	if err != nil || compacted != 1000 {
		t.Fatalf("first compaction = %d, %v; want 1000", compacted, err)
	}
	var expiredCount, expiredCandidateCount, eventCount, proofCount int
	if err := pool.QueryRow(ctx, `
SELECT count(*) FILTER (WHERE expired_at IS NOT NULL),
       count(*) FILTER (WHERE retained_until <= $1 AND expired_at IS NULL)
  FROM jobs
`, cutoff).Scan(&expiredCount, &expiredCandidateCount); err != nil {
		t.Fatal(err)
	}
	if expiredCount != 1000 || expiredCandidateCount != 1 {
		t.Fatalf("bounded batch state = expired %d remaining %d; want 1000/1", expiredCount, expiredCandidateCount)
	}
	var retainedJobKind, retainedUnit, retainedHandler, retainedOwner, retainedStatus string
	var retainedIncident, retainedSubmitter uuid.UUID
	var expiredAt, retainedUntil, finishedAt *time.Time
	var payload, result, failure, message, attemptID, lease, nextAttempt, lastAttempt, lastError any
	var failureCount int
	var identity, routeKey, scopeKey, digest any
	if err := pool.QueryRow(ctx, `
SELECT job_kind, progress_unit_id, handler_name, extension_owner_profile_id, status,
       incident_id, submitted_by_user_id, expired_at, retained_until, finished_at,
       handler_payload_json, result_summary_json, error_summary_json, message,
       handler_attempt_id, handler_lease_expires_at, handler_next_attempt_at,
       handler_last_attempted_at, handler_last_error, handler_failure_count,
       extension_idempotency_identity, extension_idempotency_route_key,
       extension_idempotency_scope_key, extension_normalized_request_sha256
  FROM jobs WHERE job_id = $1
`, specialID).Scan(
		&retainedJobKind, &retainedUnit, &retainedHandler, &retainedOwner, &retainedStatus,
		&retainedIncident, &retainedSubmitter, &expiredAt, &retainedUntil, &finishedAt,
		&payload, &result, &failure, &message, &attemptID, &lease, &nextAttempt,
		&lastAttempt, &lastError, &failureCount, &identity, &routeKey, &scopeKey, &digest,
	); err != nil {
		t.Fatal(err)
	}
	if retainedJobKind != "expiry.run_v1" || retainedUnit != "expiry.run.attempt.v1" ||
		retainedHandler != "expiry.worker_v1" || retainedOwner != "expiry" ||
		retainedStatus != StatusSucceeded || retainedIncident != incidentID || retainedSubmitter != actorID ||
		expiredAt == nil || !expiredAt.Equal(cutoff) || retainedUntil == nil || finishedAt == nil {
		t.Fatalf("tombstone lost retained identity/provenance")
	}
	if payload != nil || result != nil || failure != nil || message != nil || attemptID != nil ||
		lease != nil || nextAttempt != nil || lastAttempt != nil || lastError != nil || failureCount != 0 ||
		identity != nil || routeKey != nil || scopeKey != nil || digest != nil {
		t.Fatalf("tombstone retained cleared material")
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, specialID).Scan(&proofCount); err != nil {
		t.Fatal(err)
	}
	eventCount = intenttest.CountBySourceIdentity(t, pool, "job:"+specialID.String())
	if proofCount != 1 || eventCount != 1 {
		t.Fatalf("compaction changed proof/event rows: proof=%d event=%d", proofCount, eventCount)
	}

	restarted := &Manager{pool: pool, now: func() time.Time { return cutoff }, policy: ProductionRuntimePolicy()}
	compacted, err = restarted.compactExpiredJobs(ctx, 1000)
	if err != nil || compacted != 1 {
		t.Fatalf("restart compaction = %d, %v; want 1", compacted, err)
	}
	compacted, err = restarted.compactExpiredJobs(ctx, 1000)
	if err != nil || compacted != 0 {
		t.Fatalf("idempotent compaction = %d, %v; want 0", compacted, err)
	}
}
