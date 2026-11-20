package evidence_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/blobfixture"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestEvidenceBlobLockOrder_Integration covers AC-609 (REQ-02-190a,
// REQ-02-196a) with deterministic two-session interleavings. A test-held
// blocker session pins the Blob; each step waits until PostgreSQL reports the
// production session blocked (pg_blocking_pids), never on elapsed time.
func TestEvidenceBlobLockOrder_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-blob-lock-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "lock-order-incident", "incident_key": "LOCK-ORDER", "title": "Evidence Blob lock order",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	quarantine := evidence.NewBlobQuarantinerForTest(t, harness.Pool)
	order := lockOrderHarness{t: t, harness: harness, login: login, actor: actor, incidentID: incidentID}

	t.Run("patch to available behind a quarantine is refused", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("patch-behind-quarantine", false)
		blocker := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR SHARE`)
		quarantined := goAwait(func() error { return quarantine(context.Background(), actor, blob) })
		quarantinePID := order.awaitBlocked(quarantined.done, blocker.pid)
		patched := goAwait(func() httpResult {
			return order.patchLifecycle(record, 2, "patch-behind-quarantine-available", "available")
		})
		order.awaitBlocked(patched.done, blocker.pid, quarantinePID)
		blocker.release()
		requireNoDeadlock(t, "quarantine", quarantined.wait(t))
		patched.wait(t).requireError(t, http.StatusConflict, "illegal_transition")
		order.requireState(record, blob, "received", "quarantined")
	})

	t.Run("quarantine behind a patch to available quarantines the Evidence", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("quarantine-behind-patch", false)
		blocker := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR UPDATE`)
		patched := goAwait(func() httpResult {
			return order.patchLifecycle(record, 2, "quarantine-behind-patch-available", "available")
		})
		patchPID := order.awaitBlocked(patched.done, blocker.pid)
		quarantined := goAwait(func() error { return quarantine(context.Background(), actor, blob) })
		order.awaitBlocked(quarantined.done, blocker.pid, patchPID)
		blocker.release()
		patched.wait(t).requireStatus(t, http.StatusOK)
		requireNoDeadlock(t, "quarantine", quarantined.wait(t))
		order.requireState(record, blob, "quarantined", "quarantined")
	})

	t.Run("patch decides the bridge under the Blob lock", func(t *testing.T) {
		// A Blob-state writer holds the Blob row while it changes it. The
		// patch must wait for the writer and decide against its committed
		// state, not the pre-change snapshot.
		order := order.with(t)
		record, blob := order.linkedEvidence("patch-blob-writer", false)
		writer := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR UPDATE`)
		blobfixture.Move(t, writer.tx, blob, blobfixture.Quarantined)
		patched := goAwait(func() httpResult { return order.patchLifecycle(record, 2, "patch-blob-writer-available", "available") })
		waited := order.awaitBlockedOrDone(patched.done, writer.pid)
		writer.commit()
		result := patched.wait(t)
		order.requireBridge(record, blob)
		if !waited {
			t.Fatalf("patch decided the Evidence bridge without waiting for the Blob lock: %#v", result)
		}
		result.requireError(t, http.StatusConflict, "illegal_transition")
		order.requireState(record, blob, "received", "quarantined")
	})

	t.Run("attach replay racing quarantine does not deadlock", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("attach-replay", true)
		blocker := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR SHARE`)
		quarantined := goAwait(func() error { return quarantine(context.Background(), actor, blob) })
		quarantinePID := order.awaitBlocked(quarantined.done, blocker.pid)
		replayed := goAwait(func() httpResult { return order.attach(record, blob, 1, "attach-replay-attach") })
		order.awaitBlocked(replayed.done, blocker.pid, quarantinePID)
		blocker.release()
		requireNoDeadlock(t, "quarantine", quarantined.wait(t))
		// REQ-01-245: the exact replay waits only on the Evidence lock and
		// then returns its original receipt over the now-quarantined Blob.
		replayed.wait(t).requireStatus(t, http.StatusOK)
		order.requireState(record, blob, "quarantined", "quarantined")
	})

	t.Run("restore racing quarantine is record_locked then bridge-satisfying", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("restore-race", false)
		order.call(http.MethodDelete, "/api/v1/records/"+record.String(), map[string]any{"base_row_version": 2, "client_txn_id": "restore-race-delete"}).requireStatus(t, http.StatusOK)
		blocker := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR SHARE`)
		quarantined := goAwait(func() error { return quarantine(context.Background(), actor, blob) })
		order.awaitBlocked(quarantined.done, blocker.pid)
		body := map[string]any{"base_row_version": 3, "client_txn_id": "restore-race-restore"}
		order.call(http.MethodPost, "/api/v1/records/"+record.String()+"/restore", body).requireRecordLocked(t, record)
		blocker.release()
		requireNoDeadlock(t, "quarantine", quarantined.wait(t))
		order.requireState(record, blob, "received", "quarantined")
		order.call(http.MethodPost, "/api/v1/records/"+record.String()+"/restore", body).requireStatus(t, http.StatusOK)
		order.requireState(record, blob, "received", "quarantined")
		order.requireLive(record, true)
	})

	t.Run("rollback racing quarantine is record_locked then bridge-satisfying", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("rollback-race", false)
		retitled := order.call(http.MethodPatch, "/api/v1/records/"+record.String(), map[string]any{
			"view_schema_id": "cartulary.view.evidence.v1", "base_row_version": 2, "client_txn_id": "rollback-race-title",
			"changes": []map[string]any{{"field_key": "evidence.title", "value": "Retitled lock order"}},
		})
		retitled.requireStatus(t, http.StatusOK)
		changeSet := retitled.body["data"].(map[string]any)["change_set_id"]
		blocker := order.block(blob, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR SHARE`)
		quarantined := goAwait(func() error { return quarantine(context.Background(), actor, blob) })
		order.awaitBlocked(quarantined.done, blocker.pid)
		body := map[string]any{"base_row_version": 3, "client_txn_id": "rollback-race-rollback", "target": map[string]any{"kind": "change_set", "change_set_id": changeSet}}
		order.call(http.MethodPost, "/api/v1/records/"+record.String()+"/rollback", body).requireRecordLocked(t, record)
		blocker.release()
		requireNoDeadlock(t, "quarantine", quarantined.wait(t))
		order.call(http.MethodPost, "/api/v1/records/"+record.String()+"/rollback", body).requireStatus(t, http.StatusOK)
		order.requireState(record, blob, "received", "quarantined")
		var title string
		if err := harness.Pool.QueryRow(context.Background(), `SELECT title FROM evidence WHERE record_id = $1`, record).Scan(&title); err != nil || title != "Lock order rollback-race" {
			t.Fatalf("rollback title = %q (%v), want the retained title", title, err)
		}
	})

	t.Run("patch over a tuple-violating Blob fails closed", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("patch-breach", false)
		blobfixture.BreachTupleInvariant(t, harness.Pool, blob)
		before := order.stateSnapshot(record, blob)
		order.call(http.MethodPatch, "/api/v1/records/"+record.String(), map[string]any{
			"view_schema_id": "cartulary.view.evidence.v1", "base_row_version": 2, "client_txn_id": "patch-breach-title",
			"changes": []map[string]any{{"field_key": "evidence.title", "value": "Must not commit"}},
		}).requireError(t, http.StatusInternalServerError, "internal_error")
		if after := order.stateSnapshot(record, blob); after != before {
			t.Fatalf("patch over a tuple-violating Blob committed:\nbefore %s\nafter  %s", before, after)
		}
	})

	t.Run("quarantine fails closed when the association changes before every attempt", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("quarantine-churn-exhausted", false)
		before := order.stateSnapshot(record, blob)
		churn := order.associationChurn(record, blob, evidence.QuarantineAssociationAttemptsForTest)
		err := evidence.NewBlobQuarantinerForTest(t, churn)(context.Background(), actor, blob)
		if !errors.Is(err, evidence.ErrQuarantineAssociationUnstableForTest) {
			t.Fatalf("quarantine under continuous association churn = %v, want the fail-closed unstable outcome", err)
		}
		if churn.begins != evidence.QuarantineAssociationAttemptsForTest || churn.churns != evidence.QuarantineAssociationAttemptsForTest {
			t.Fatalf("quarantine ran %d attempts with %d competing changes, want %d of each", churn.begins, churn.churns, evidence.QuarantineAssociationAttemptsForTest)
		}
		if churn.churns%2 == 1 {
			// An odd number of competing changes leaves the Evidence detached.
			// Relink it so the snapshot compares only what quarantine could write.
			if err := churn.toggle(context.Background()); err != nil {
				t.Fatalf("relink churned Evidence: %v", err)
			}
		}
		if after := order.stateSnapshot(record, blob); after != before {
			t.Fatalf("exhausted quarantine committed:\nbefore %s\nafter  %s", before, after)
		}
		order.requireState(record, blob, "received", "available")
	})

	t.Run("quarantine succeeds when the association settles within the bound", func(t *testing.T) {
		order := order.with(t)
		record, blob := order.linkedEvidence("quarantine-churn-settles", false)
		churn := order.associationChurn(record, blob, evidence.QuarantineAssociationAttemptsForTest-1)
		if err := evidence.NewBlobQuarantinerForTest(t, churn)(context.Background(), actor, blob); err != nil {
			t.Fatalf("quarantine after the association settled: %v", err)
		}
		if churn.begins != evidence.QuarantineAssociationAttemptsForTest {
			t.Fatalf("quarantine ran %d attempts, want %d", churn.begins, evidence.QuarantineAssociationAttemptsForTest)
		}
		var uploadState string
		if err := harness.Pool.QueryRow(context.Background(), `SELECT upload_state FROM object_blobs WHERE object_blob_id = $1`, blob).Scan(&uploadState); err != nil || uploadState != "quarantined" {
			t.Fatalf("Blob upload_state = %q (%v), want quarantined", uploadState, err)
		}
		blobfixture.RequireCoherent(t, harness.Pool, blob)
	})
}

type lockOrderHarness struct {
	t          *testing.T
	harness    *appsupport.ServerHarness
	login      authcookietest.AuthCookies
	actor      uuid.UUID
	incidentID uuid.UUID
}

func (h lockOrderHarness) with(t *testing.T) lockOrderHarness {
	h.t = t
	return h
}

// linkedEvidence creates received Evidence through the public route and
// attaches a coherent available Blob (row_version 2). When available is set
// the Evidence is then made available (row_version 3).
func (h lockOrderHarness) linkedEvidence(label string, available bool) (uuid.UUID, uuid.UUID) {
	h.t.Helper()
	created := h.call(http.MethodPost, "/api/v1/incidents/"+h.incidentID.String()+"/views/cartulary.view.evidence.v1/rows", map[string]any{
		"client_txn_id": label + "-create", "evidence.title": "Lock order " + label, "evidence.lifecycle_state": "received",
	})
	created.requireStatus(h.t, http.StatusCreated)
	record := uuid.MustParse(created.body["data"].(map[string]any)["row"].(map[string]any)["record_id"].(string))
	blob := blobfixture.Seed(h.t, h.harness.Pool, blobfixture.Spec{IncidentID: h.incidentID, CreatedByUserID: h.actor, State: blobfixture.Available, ByteSize: 4})
	h.attach(record, blob, 1, label+"-attach").requireStatus(h.t, http.StatusOK)
	if available {
		h.patchLifecycle(record, 2, label+"-available", "available").requireStatus(h.t, http.StatusOK)
	}
	return record, blob
}

func (h lockOrderHarness) attach(record, blob uuid.UUID, baseRowVersion int, txn string) httpResult {
	return h.call(http.MethodPost, "/api/v1/evidence-records/"+record.String()+"/attach-blob", map[string]any{
		"object_blob_id": blob.String(), "base_row_version": baseRowVersion, "client_txn_id": txn,
	})
}

func (h lockOrderHarness) patchLifecycle(record uuid.UUID, baseRowVersion int, txn, lifecycle string) httpResult {
	return h.call(http.MethodPatch, "/api/v1/records/"+record.String(), map[string]any{
		"view_schema_id": "cartulary.view.evidence.v1", "base_row_version": baseRowVersion, "client_txn_id": txn,
		"changes": []map[string]any{{"field_key": "evidence.lifecycle_state", "value": lifecycle}},
	})
}

type httpResult struct {
	status int
	body   map[string]any
	err    error
}

// call performs one authenticated JSON request without touching testing.T, so
// it is safe inside the racing goroutines.
func (h lockOrderHarness) call(method, path string, body any) httpResult {
	payload, err := json.Marshal(body)
	if err != nil {
		return httpResult{err: err}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, method, h.harness.Server.HTTP.URL+path, bytes.NewReader(payload))
	if err != nil {
		return httpResult{err: err}
	}
	request.Header.Set("Content-Type", "application/json")
	for _, option := range authOptions(h.login) {
		option(request)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return httpResult{err: err}
	}
	defer response.Body.Close()
	result := httpResult{status: response.StatusCode}
	if err := json.NewDecoder(response.Body).Decode(&result.body); err != nil {
		result.err = fmt.Errorf("decode %s %s response: %w", method, path, err)
	}
	return result
}

func (r httpResult) requireStatus(t testing.TB, status int) {
	t.Helper()
	if r.err != nil || r.status != status {
		t.Fatalf("response status = %d (%v) body %#v, want %d", r.status, r.err, r.body, status)
	}
}

func (r httpResult) requireError(t testing.TB, status int, code string) map[string]any {
	t.Helper()
	r.requireStatus(t, status)
	errorObject, _ := r.body["error"].(map[string]any)
	if errorObject == nil || errorObject["code"] != code {
		t.Fatalf("error = %#v, want %s", r.body, code)
	}
	return errorObject
}

func (r httpResult) requireRecordLocked(t testing.TB, record uuid.UUID) {
	t.Helper()
	errorObject := r.requireError(t, http.StatusConflict, "record_locked")
	details, _ := errorObject["details"].(map[string]any)
	if errorObject["retryable"] != true || len(details) != 1 || details["record_id"] != record.String() {
		t.Fatalf("record_locked error = %#v, want retryable with only the Evidence record", errorObject)
	}
}

type lockBlocker struct {
	t   *testing.T
	tx  pgx.Tx
	pid int32
}

// block opens the test-held blocker session and takes lockSQL on the Blob.
func (h lockOrderHarness) block(blob uuid.UUID, lockSQL string) lockBlocker {
	h.t.Helper()
	tx, err := h.harness.Pool.Begin(context.Background())
	if err != nil {
		h.t.Fatalf("begin blocker: %v", err)
	}
	h.t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	var pid int32
	if err := tx.QueryRow(context.Background(), `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
		h.t.Fatalf("blocker pid: %v", err)
	}
	if _, err := tx.Exec(context.Background(), lockSQL, blob); err != nil {
		h.t.Fatalf("blocker lock: %v", err)
	}
	return lockBlocker{t: h.t, tx: tx, pid: pid}
}

func (b lockBlocker) release() {
	b.t.Helper()
	if err := b.tx.Rollback(context.Background()); err != nil {
		b.t.Fatalf("release blocker: %v", err)
	}
}

func (b lockBlocker) commit() {
	b.t.Helper()
	if err := b.tx.Commit(context.Background()); err != nil {
		b.t.Fatalf("commit blocker: %v", err)
	}
}

// awaitBlocked returns the pid of a session blocked by one of blockers and not
// itself one of them. It fails if the watched operation finishes first.
func (h lockOrderHarness) awaitBlocked(done <-chan struct{}, blockers ...int32) int32 {
	h.t.Helper()
	pid, ok := h.pollBlocked(done, blockers)
	if !ok {
		h.t.Fatal("operation finished before it waited on the lock order")
	}
	return pid
}

// awaitBlockedOrDone reports whether the operation was observed waiting on one
// of blockers before it finished.
func (h lockOrderHarness) awaitBlockedOrDone(done <-chan struct{}, blockers ...int32) bool {
	h.t.Helper()
	_, ok := h.pollBlocked(done, blockers)
	return ok
}

func (h lockOrderHarness) pollBlocked(done <-chan struct{}, blockers []int32) (int32, bool) {
	h.t.Helper()
	deadline := time.NewTimer(60 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(5 * time.Millisecond)
	defer poll.Stop()
	for {
		var pid int32
		err := h.harness.Pool.QueryRow(context.Background(), `
SELECT pid
  FROM pg_stat_activity
 WHERE datname = current_database()
   AND pid <> ALL($1::integer[])
   AND pg_blocking_pids(pid) && $1::integer[]
 LIMIT 1
`, blockers).Scan(&pid)
		if err == nil {
			return pid, true
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			h.t.Fatalf("observe blocked sessions: %v", err)
		}
		select {
		case <-done:
			return 0, false
		case <-deadline.C:
			h.t.Fatal("operation never waited on the expected lock")
		case <-poll.C:
		}
	}
}

func (h lockOrderHarness) requireState(record, blob uuid.UUID, lifecycle, uploadState string) {
	h.t.Helper()
	var gotLifecycle, gotUpload string
	if err := h.harness.Pool.QueryRow(context.Background(), `
SELECT e.lifecycle_state, b.upload_state
  FROM evidence e
  JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1 AND b.object_blob_id = $2
`, record, blob).Scan(&gotLifecycle, &gotUpload); err != nil {
		h.t.Fatalf("load Evidence/Blob state: %v", err)
	}
	if gotLifecycle != lifecycle || gotUpload != uploadState {
		h.t.Fatalf("Evidence/Blob = %s/%s, want %s/%s", gotLifecycle, gotUpload, lifecycle, uploadState)
	}
	blobfixture.RequireCoherent(h.t, h.harness.Pool, blob)
}

// requireBridge fails when available or released Evidence is linked to a Blob
// that is not available (REQ-02-190).
func (h lockOrderHarness) requireBridge(record, blob uuid.UUID) {
	h.t.Helper()
	var lifecycle, uploadState string
	if err := h.harness.Pool.QueryRow(context.Background(), `
SELECT e.lifecycle_state, b.upload_state
  FROM evidence e
  JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1 AND b.object_blob_id = $2
`, record, blob).Scan(&lifecycle, &uploadState); err != nil {
		h.t.Fatalf("load Evidence/Blob bridge: %v", err)
	}
	if (lifecycle == "available" || lifecycle == "released") && uploadState != "available" {
		h.t.Fatalf("Evidence committed %s over a %s Blob", lifecycle, uploadState)
	}
}

func (h lockOrderHarness) requireLive(record uuid.UUID, live bool) {
	h.t.Helper()
	var deleted bool
	if err := h.harness.Pool.QueryRow(context.Background(), `SELECT deleted_at IS NOT NULL FROM records WHERE record_id = $1`, record).Scan(&deleted); err != nil || deleted == live {
		h.t.Fatalf("Evidence deleted=%t (%v), want live=%t", deleted, err, live)
	}
}

func (h lockOrderHarness) stateSnapshot(record, blob uuid.UUID) string {
	h.t.Helper()
	var snapshot string
	if err := h.harness.Pool.QueryRow(context.Background(), `
SELECT jsonb_build_object(
           'record', (SELECT to_jsonb(r) FROM records r WHERE r.record_id = $1),
           'evidence', (SELECT to_jsonb(e) FROM evidence e WHERE e.record_id = $1),
           'blob', (SELECT to_jsonb(b) FROM object_blobs b WHERE b.object_blob_id = $2),
           'change_sets', (SELECT count(*) FROM change_sets WHERE incident_id = $3)
       )::text
`, record, blob, h.incidentID).Scan(&snapshot); err != nil {
		h.t.Fatalf("snapshot Evidence state: %v", err)
	}
	return snapshot
}

type awaited[T any] struct {
	done  chan struct{}
	value T
}

// goAwait runs operation on its own goroutine; done closes when it returns.
func goAwait[T any](operation func() T) *awaited[T] {
	result := &awaited[T]{done: make(chan struct{})}
	go func() {
		defer close(result.done)
		result.value = operation()
	}()
	return result
}

func (a *awaited[T]) wait(t testing.TB) T {
	t.Helper()
	select {
	case <-a.done:
		return a.value
	case <-time.After(60 * time.Second):
		t.Fatal("racing operation did not finish")
		var zero T
		return zero
	}
}

func requireNoDeadlock(t testing.TB, label string, err error) {
	t.Helper()
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "40P01" {
		t.Fatalf("%s deadlocked: %v", label, err)
	}
	if err != nil {
		t.Fatalf("%s failed: %v", label, err)
	}
}

// associationChurnDB commits one competing association change on its own
// session right after each quarantine attempt's unlocked association lookup:
// the REQ-02-190a window between that lookup and the Blob lock.
type associationChurnDB struct {
	dbport.DB
	toggle         func(context.Context) error
	remaining      int
	begins, churns int
}

// associationChurn toggles record's link to blob, at most changes times. The
// competing writer follows the production order: Evidence envelope, Blob, then
// the association change.
func (h lockOrderHarness) associationChurn(record, blob uuid.UUID, changes int) *associationChurnDB {
	pool := h.harness.Pool
	return &associationChurnDB{DB: pool, remaining: changes, toggle: func(ctx context.Context) error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if _, err := tx.Exec(ctx, `SELECT 1 FROM records WHERE record_id = $1 FOR UPDATE`, record); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM object_blobs WHERE object_blob_id = $1 FOR UPDATE`, blob); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE evidence SET object_blob_id = CASE WHEN object_blob_id IS NULL THEN $2::uuid END WHERE record_id = $1`, record, blob); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}}
}

func (db *associationChurnDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	db.begins++
	return &associationChurnTx{Tx: tx, db: db}, nil
}

type associationChurnTx struct {
	pgx.Tx
	db     *associationChurnDB
	looked bool
}

// QueryRow lets the attempt's first statement, the unlocked association
// lookup, finish before the competing change commits.
func (tx *associationChurnTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := tx.Tx.QueryRow(ctx, sql, args...)
	if tx.looked {
		return row
	}
	tx.looked = true
	return churnAfterLookup{Row: row, ctx: ctx, db: tx.db}
}

type churnAfterLookup struct {
	pgx.Row
	ctx context.Context
	db  *associationChurnDB
}

func (row churnAfterLookup) Scan(dest ...any) error {
	if err := row.Row.Scan(dest...); err != nil || row.db.remaining == 0 {
		return err
	}
	row.db.remaining--
	row.db.churns++
	return row.db.toggle(row.ctx)
}
