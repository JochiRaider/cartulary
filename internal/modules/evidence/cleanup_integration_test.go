package evidence_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestEvidenceDurableCleanupClaimEngine_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence-durable-cleanup-claims")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence-durable-cleanup-claims-incident",
		"incident_key":  "evidence-durable-cleanup-claims",
		"title":         "Evidence durable cleanup claims",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	cleanup, err := evidence.NewCleanupService(harness.Pool)
	if err != nil {
		t.Fatalf("compose cleanup service: %v", err)
	}
	base := time.Now().UTC().Truncate(time.Second)

	t.Run("production remains inert before dispatcher activation", func(t *testing.T) {
		blobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("inert", blobID), base)
		var count int
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT COUNT(*) FROM evidence_blob_cleanup_claims WHERE object_blob_id = $1
`, blobID).Scan(&count); err != nil {
			t.Fatalf("count inert cleanup claims: %v", err)
		}
		if count != 0 {
			t.Fatalf("server created %d cleanup claims before S09 dispatcher activation", count)
		}
		deleteCleanupFixture(t, harness, blobID)
	})

	t.Run("multi instance claim has one deletion owner", func(t *testing.T) {
		blobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("concurrent", blobID), base)
		deleter := newBlockingCleanupDeleter()
		firstResult := make(chan evidence.CleanupSweepResult, 1)
		firstErr := make(chan error, 1)
		go func() {
			result, sweepErr := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
			firstResult <- result
			firstErr <- sweepErr
		}()
		select {
		case <-deleter.entered:
		case <-time.After(10 * time.Second):
			t.Fatal("first cleanup owner did not reach deletion")
		}
		second, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
		if err != nil {
			t.Fatalf("second cleanup instance: %v", err)
		}
		if second.ClaimedBlobCount != 0 || second.CleanedBlobCount != 0 {
			t.Fatalf("second instance result = %#v, want no claimed work", second)
		}
		close(deleter.release)
		if err := <-firstErr; err != nil {
			t.Fatalf("first cleanup instance: %v", err)
		}
		first := <-firstResult
		if first.ClaimedBlobCount != 1 || first.CleanedBlobCount != 1 || deleter.callCount() != 1 {
			t.Fatalf("first instance result=%#v delete_calls=%d, want one claimed and cleaned", first, deleter.callCount())
		}
	})

	t.Run("multi instance dispatchers retain one deletion owner", func(t *testing.T) {
		blobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("dispatcher-concurrent", blobID), base)
		secondCleanup, err := evidence.NewCleanupService(harness.Pool)
		if err != nil {
			t.Fatalf("compose second cleanup service: %v", err)
		}
		deleter := newBlockingCleanupDeleter()
		firstObserver := &integrationCleanupObserver{observed: make(chan evidence.CleanupSweepObservation, 1)}
		secondObserver := &integrationCleanupObserver{observed: make(chan evidence.CleanupSweepObservation, 1)}
		firstDispatcher, err := evidence.NewCleanupDispatcher(cleanup, deleter, firstObserver, func() time.Time { return base })
		if err != nil {
			t.Fatal(err)
		}
		secondDispatcher, err := evidence.NewCleanupDispatcher(secondCleanup, deleter, secondObserver, func() time.Time { return base })
		if err != nil {
			t.Fatal(err)
		}
		if err := firstDispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-deleter.entered:
		case <-time.After(10 * time.Second):
			t.Fatal("first dispatcher did not reach object deletion")
		}
		if err := secondDispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-secondObserver.observed:
		case <-time.After(10 * time.Second):
			t.Fatal("second dispatcher did not complete its excluded sweep")
		}
		close(deleter.release)
		select {
		case <-firstObserver.observed:
		case <-time.After(10 * time.Second):
			t.Fatal("first dispatcher did not complete its owned sweep")
		}
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := firstDispatcher.Close(closeCtx); err != nil {
			t.Fatal(err)
		}
		if err := secondDispatcher.Close(closeCtx); err != nil {
			t.Fatal(err)
		}
		if deleter.callCount() != 1 {
			t.Fatalf("multi-instance dispatchers deleted %d times, want once", deleter.callCount())
		}
	})

	t.Run("claimed rows block lifecycle and association races", func(t *testing.T) {
		blobID := uuid.New()
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("race", blobID), base)
		deleter := newBlockingCleanupDeleter()
		done := make(chan error, 1)
		go func() {
			_, sweepErr := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
			done <- sweepErr
		}()
		select {
		case <-deleter.entered:
		case <-time.After(10 * time.Second):
			t.Fatal("cleanup race fixture did not reach deletion")
		}
		_, lifecycleErr := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs
   SET upload_state = 'available', terminal_reason = NULL, failed_at = NULL,
       finalized_at = now(), updated_at = now()
 WHERE object_blob_id = $1
`, blobID)
		if lifecycleErr == nil {
			t.Fatal("claimed blob admitted a finalization/quarantine/restore state mutation")
		}
		_, associationErr := harness.DB.ExecContext(context.Background(), `
UPDATE evidence SET object_blob_id = $2, updated_at = now() WHERE record_id = $1
`, recordID, blobID)
		if associationErr == nil {
			t.Fatal("claimed blob admitted an attachment/import/replacement association")
		}
		close(deleter.release)
		if err := <-done; err != nil {
			t.Fatalf("complete cleanup after rejected races: %v", err)
		}
	})

	t.Run("expired lease is restart reclaimable", func(t *testing.T) {
		blobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("restart", blobID), base)
		insertClaim(t, harness, blobID, "claimed", 1, base.Add(-10*time.Minute), base.Add(-5*time.Minute), nil, nil)
		restarted, err := evidence.NewCleanupService(harness.Pool)
		if err != nil {
			t.Fatalf("recompose cleanup after simulated restart: %v", err)
		}
		deleter := &scriptedCleanupDeleter{}
		result, err := restarted.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
		if err != nil {
			t.Fatalf("reclaim expired cleanup lease: %v", err)
		}
		if result.ClaimedBlobCount != 1 || result.CleanedBlobCount != 1 || deleter.calls != 1 {
			t.Fatalf("restart result=%#v calls=%d, want reclaimed once", result, deleter.calls)
		}
		requireCleanupAttemptCount(t, harness, blobID, 2)
	})

	t.Run("batch stays bounded and drains more than one hundred", func(t *testing.T) {
		for index := 0; index < 101; index++ {
			blobID := uuid.New()
			insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey(fmt.Sprintf("batch-%03d", index), blobID), base)
		}
		deleter := &scriptedCleanupDeleter{}
		first, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
		if err != nil {
			t.Fatalf("first bounded cleanup batch: %v", err)
		}
		second, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
		if err != nil {
			t.Fatalf("second bounded cleanup batch: %v", err)
		}
		if first.ClaimedBlobCount != 100 || !first.HasMore || second.ClaimedBlobCount != 1 || second.HasMore || deleter.calls != 101 {
			t.Fatalf("batch results first=%#v second=%#v calls=%d", first, second, deleter.calls)
		}
	})

	t.Run("retry schedule is deterministic and never terminally exhausts", func(t *testing.T) {
		blobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, blobID, cleanupKey("retry", blobID), base)
		deleter := &scriptedCleanupDeleter{errors: []error{
			errors.New("transient-1"),
			errors.New("transient-2"),
			errors.New("transient-3"),
			errors.New("transient-4"),
			nil,
		}}
		attemptTimes := []time.Time{
			base,
			base.Add(time.Minute),
			base.Add(6 * time.Minute),
			base.Add(21 * time.Minute),
			base.Add(36 * time.Minute),
		}
		wantNext := []time.Time{
			base.Add(time.Minute),
			base.Add(6 * time.Minute),
			base.Add(21 * time.Minute),
			base.Add(36 * time.Minute),
		}
		for index, at := range attemptTimes {
			result, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, at)
			if err != nil {
				t.Fatalf("cleanup retry attempt %d: %v", index+1, err)
			}
			if index < len(wantNext) {
				if result.RetryScheduledCount != 1 || result.CleanedBlobCount != 0 {
					t.Fatalf("retry attempt %d result=%#v", index+1, result)
				}
				requireCleanupNextAttempt(t, harness, blobID, wantNext[index])
			} else if result.CleanedBlobCount != 1 {
				t.Fatalf("successful retry result=%#v", result)
			}
		}
		requireCleanupAttemptCount(t, harness, blobID, 5)
	})

	t.Run("not found and deadline failures are classified safely", func(t *testing.T) {
		notFoundBlobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, notFoundBlobID, cleanupKey("not-found", notFoundBlobID), base)
		notFound := &scriptedCleanupDeleter{errors: []error{&objectstore.AdapterError{
			Code: objectstore.ErrorCodeObjectNotFound,
		}}}
		result, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), notFound, base)
		if err != nil || result.CleanedBlobCount != 1 || result.RetryScheduledCount != 0 {
			t.Fatalf("not-found cleanup result=%#v err=%v", result, err)
		}

		timeoutBlobID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, timeoutBlobID, cleanupKey("timeout", timeoutBlobID), base)
		deadline := &deadlineInspectingDeleter{}
		result, err = cleanup.SweepFailedUnattachedBlobs(context.Background(), deadline, base)
		if err != nil || result.RetryScheduledCount != 1 || result.CleanedBlobCount != 0 {
			t.Fatalf("deadline cleanup result=%#v err=%v", result, err)
		}
		if deadline.remaining < 59*time.Second || deadline.remaining > time.Minute {
			t.Fatalf("delete timeout window=%s, want one minute", deadline.remaining)
		}
		requireCleanupFailureClass(t, harness, timeoutBlobID, "delete_timeout")
		deleteCleanupFixture(t, harness, timeoutBlobID)
	})

	t.Run("pending timeout meets one hour deletion bound", func(t *testing.T) {
		blobID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, blobID, cleanupKey("pending", blobID), base)
		deleter := &scriptedCleanupDeleter{}
		first, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base)
		if err != nil {
			t.Fatalf("first pending-timeout stage: %v", err)
		}
		if first.ExpiredPendingCount != 1 || first.ClaimedBlobCount != 0 {
			t.Fatalf("pending first stage=%#v, want expired without premature deletion", first)
		}
		requireCleanupDueAt(t, harness, blobID, base.Add(45*time.Minute))
		second, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, base.Add(45*time.Minute))
		if err != nil || second.CleanedBlobCount != 1 {
			t.Fatalf("pending deadline cleanup result=%#v err=%v", second, err)
		}
	})

	t.Run("metadata survives seven days then is hard deleted", func(t *testing.T) {
		retainedID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, retainedID, cleanupKey("retained", retainedID), base)
		markCompletedCleanup(t, harness, retainedID, base.Add(-7*24*time.Hour).Add(time.Second))
		deletedID := uuid.New()
		insertFailedCleanupBlob(t, harness, incidentID, adminID, deletedID, cleanupKey("deleted", deletedID), base)
		markCompletedCleanup(t, harness, deletedID, base.Add(-7*24*time.Hour).Add(-time.Second))
		result, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), &scriptedCleanupDeleter{}, base)
		if err != nil {
			t.Fatalf("apply cleanup metadata retention: %v", err)
		}
		if result.DeletedMetadataCount != 1 {
			t.Fatalf("deleted metadata count=%d, want 1", result.DeletedMetadataCount)
		}
		requireObjectBlobExists(t, harness, retainedID, true)
		requireObjectBlobExists(t, harness, deletedID, false)
	})
}

type scriptedCleanupDeleter struct {
	errors []error
	calls  int
}

func (deleter *scriptedCleanupDeleter) DeleteObject(context.Context, string) error {
	index := deleter.calls
	deleter.calls++
	if index < len(deleter.errors) {
		return deleter.errors[index]
	}
	return nil
}

type blockingCleanupDeleter struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	mu      sync.Mutex
	calls   int
}

func newBlockingCleanupDeleter() *blockingCleanupDeleter {
	return &blockingCleanupDeleter{entered: make(chan struct{}), release: make(chan struct{})}
}

func (deleter *blockingCleanupDeleter) DeleteObject(ctx context.Context, _ string) error {
	deleter.mu.Lock()
	deleter.calls++
	deleter.mu.Unlock()
	deleter.once.Do(func() { close(deleter.entered) })
	select {
	case <-deleter.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (deleter *blockingCleanupDeleter) callCount() int {
	deleter.mu.Lock()
	defer deleter.mu.Unlock()
	return deleter.calls
}

type deadlineInspectingDeleter struct {
	remaining time.Duration
}

type integrationCleanupObserver struct {
	observed chan evidence.CleanupSweepObservation
}

func (observer *integrationCleanupObserver) ObserveCleanupSweep(_ context.Context, observation evidence.CleanupSweepObservation) {
	observer.observed <- observation
}

func (deleter *deadlineInspectingDeleter) DeleteObject(ctx context.Context, _ string) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("cleanup delete context has no deadline")
	}
	deleter.remaining = time.Until(deadline)
	return context.DeadlineExceeded
}

func cleanupKey(label string, blobID uuid.UUID) string {
	return "evidence_cleanup_claims/" + label + "/" + blobID.String()
}

func deleteCleanupFixture(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM object_blobs WHERE object_blob_id = $1`, blobID); err != nil {
		t.Fatalf("delete cleanup fixture: %v", err)
	}
}

func insertClaim(
	t testing.TB,
	harness *appsupport.ServerHarness,
	blobID uuid.UUID,
	state string,
	attempt int,
	claimedAt time.Time,
	expiresAt any,
	nextAttempt any,
	completedAt any,
) {
	t.Helper()
	lastAttempt := any(nil)
	failureClass := any(nil)
	if state == "retry_wait" {
		lastAttempt = claimedAt
		failureClass = "delete_failed"
	}
	if state == "completed" {
		lastAttempt = completedAt
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO evidence_blob_cleanup_claims (
    object_blob_id, claim_token, claim_state, attempt_count, claimed_at,
    claim_expires_at, next_attempt_at, last_attempt_at, completed_at,
    last_failure_class, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5::timestamptz,
    $6::timestamptz, $7::timestamptz, $8::timestamptz, $9::timestamptz,
    $10::text, $5::timestamptz,
    COALESCE($9::timestamptz, $8::timestamptz, $5::timestamptz)
)
`, blobID, uuid.New(), state, attempt, claimedAt.UTC(), expiresAt, nextAttempt, lastAttempt, completedAt, failureClass); err != nil {
		t.Fatalf("insert cleanup claim: %v", err)
	}
}

func markCompletedCleanup(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, completedAt time.Time) {
	t.Helper()
	insertClaim(t, harness, blobID, "completed", 1, completedAt.Add(-time.Minute), nil, nil, completedAt)
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs SET cleaned_up_at = $2, updated_at = $2 WHERE object_blob_id = $1
`, blobID, completedAt.UTC()); err != nil {
		t.Fatalf("mark completed cleanup metadata: %v", err)
	}
}

func requireCleanupAttemptCount(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, want int) {
	t.Helper()
	var got int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT attempt_count FROM evidence_blob_cleanup_claims WHERE object_blob_id = $1
`, blobID).Scan(&got); err != nil {
		t.Fatalf("load cleanup attempt count: %v", err)
	}
	if got != want {
		t.Fatalf("cleanup attempt count=%d, want %d", got, want)
	}
}

func requireCleanupNextAttempt(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, want time.Time) {
	t.Helper()
	var got time.Time
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT next_attempt_at FROM evidence_blob_cleanup_claims WHERE object_blob_id = $1
`, blobID).Scan(&got); err != nil {
		t.Fatalf("load cleanup next attempt: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("cleanup next attempt=%s, want %s", got, want)
	}
}

func requireCleanupFailureClass(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, want string) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT last_failure_class FROM evidence_blob_cleanup_claims WHERE object_blob_id = $1
`, blobID).Scan(&got); err != nil {
		t.Fatalf("load cleanup failure class: %v", err)
	}
	if got != want {
		t.Fatalf("cleanup failure class=%q, want %q", got, want)
	}
}

func requireCleanupDueAt(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, want time.Time) {
	t.Helper()
	var got time.Time
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT cleanup_due_at FROM object_blobs WHERE object_blob_id = $1
`, blobID).Scan(&got); err != nil {
		t.Fatalf("load cleanup due time: %v", err)
	}
	if !got.Equal(want) {
		t.Fatalf("cleanup due at=%s, want %s", got, want)
	}
}

func requireObjectBlobExists(t testing.TB, harness *appsupport.ServerHarness, blobID uuid.UUID, want bool) {
	t.Helper()
	var got bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT EXISTS (SELECT 1 FROM object_blobs WHERE object_blob_id = $1)
`, blobID).Scan(&got); err != nil {
		t.Fatalf("load object blob existence: %v", err)
	}
	if got != want {
		t.Fatalf("object blob exists=%v, want %v", got, want)
	}
}
