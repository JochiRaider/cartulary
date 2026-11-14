package incidentbundles_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPortableFatalCompletionPreservesParent_Integration(t *testing.T) {
	t.Run("exhaustion during abort lock", testPortableFatalDuringAbortLock)
	for _, phase := range []string{"after_abort", "after_receipt", "cooperative"} {
		t.Run(phase, func(t *testing.T) { testPortableFatalAfterAbort(t, phase) })
	}
	f := referencetest.NewPortable(t)
	fatal := make(chan error, 1)
	lifetime, err := reference_data.NewVerificationLifetime(20*time.Millisecond, func(err error) { fatal <- err })
	if err != nil {
		t.Fatal(err)
	}
	service := f.Service(f.Root, lifetime)
	request := f.Request()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	release := make(chan struct{})
	releaseCallback := sync.OnceFunc(func() { close(release) })
	prepared := make(chan struct{})
	done := make(chan error, 1)
	finished := make(chan struct{})
	defer joinPortableFixture(t, cancel, finished, releaseCallback)
	go func() {
		defer close(finished)
		done <- service.RunImportExecution(ctx, request.Execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
			request.ExecutionScope = scope
			if _, err := service.PrepareImport(run, request); err != nil {
				return err
			}
			close(prepared)
			err := scope.CompleteCleanup(func(context.Context) error { <-release; return errors.New("late close failure") })
			// Even an ordinary late error cannot clear the tracked fatal selection.
			incidentbundles.CompleteFailedImportForTest(t, context.Background(), scope, request.Execution, errors.New("late ordinary result"))
			if abort := scope.AbortTx(context.Background(), nil); !errors.Is(abort, extensiondeadline.ErrGraceExhausted) {
				t.Error("fatal abort reached mutation", abort)
			}
			return err
		})
	}()
	select {
	case <-prepared:
	case err := <-done:
		t.Fatal("portable preparation failed", err)
	case <-time.After(10 * time.Second):
		t.Fatal("no preparation")
	}
	beforeState := snapshotPortableFatalState(t, f)
	requirePortableRecoveryInput(t, f, request.Execution.JobID())
	cancel()
	select {
	case <-fatal:
	case <-time.After(time.Second):
		t.Fatal("no fatal consequence")
	}
	select {
	case <-done:
		t.Fatal("operation returned while cleanup active")
	default:
	}
	releaseCallback()
	if err := awaitPortableCompletion(t, done); !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		t.Fatal("late cleanup replaced fatal", err)
	}
	requirePortableFatalState(t, f, beforeState)
	requirePortableRecoveryInput(t, f, request.Execution.JobID())
	job, err := f.Manager.Get(t.Context(), request.Execution.JobID())
	if err != nil || job.Status != "running" {
		t.Fatal("fatal completion terminated Job", job.Status, err)
	}
	entries, err := os.ReadDir(filepath.Join(f.Temporary, "reference-packs", "verification"))
	if err != nil || len(entries) != 0 {
		t.Fatal("retained scratch", len(entries), err)
	}
}

// Each fixture owns an isolated database. Preserve whole rows so updating an
// existing fact cannot hide behind unchanged table cardinality.
func snapshotPortableFatalState(t testing.TB, f *referencetest.Portable) map[string]string {
	t.Helper()
	state := map[string]string{}
	for _, table := range []string{
		"reference_pack_events", "reference_pack_portable_catalogs", "reference_pack_version_pins",
		"reference_pack_pins", "reference_pack_versions", "extension_job_commit_proofs",
		"jobs", "incident_bundle_job_payloads", "route_idempotency",
		"reference_pack_operations", "reference_pack_portable_selections",
	} {
		var rows string
		if err := f.Pool.QueryRow(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(s) ORDER BY to_jsonb(s)::text),'[]'::jsonb)::text FROM "+table+" s").Scan(&rows); err != nil {
			t.Fatal(err)
		}
		state[table] = rows
	}
	return state
}

func requirePortableRecoveryInput(t testing.TB, f *referencetest.Portable, jobID uuid.UUID) {
	t.Helper()
	rows, err := f.Pool.Query(t.Context(), `SELECT s.pack_key,s.pack_version,s.expected_container_sha256,s.expected_container_bytes,o.sha256,o.size_bytes,o.available,o.storage_ref FROM reference_pack_portable_selections s LEFT JOIN reference_pack_objects o ON o.object_id=s.input_object_id WHERE s.operation_id IN (SELECT operation_id FROM reference_pack_operations WHERE job_id=$1) AND s.input_object_id IS NOT NULL ORDER BY s.ordinal`, jobID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	digest := sha256.Sum256(f.Pack.Container)
	wantDigest := hex.EncodeToString(digest[:])
	wantSize := int64(len(f.Pack.Container))
	count := 0
	for rows.Next() {
		count++
		var key, version, selectedDigest, objectDigest, raw string
		var selectedSize, objectSize int64
		var available bool
		if err := rows.Scan(&key, &version, &selectedDigest, &selectedSize, &objectDigest, &objectSize, &available, &raw); err != nil {
			t.Fatal(err)
		}
		if key != f.Pack.Key || version != f.Pack.Version || selectedDigest != wantDigest || objectDigest != wantDigest || selectedSize != wantSize || objectSize != wantSize || !available {
			t.Fatal("recovery input identity, size, digest or availability differs from fixture")
		}
		ref, err := reference_data.ParseStorageRef(raw)
		if err != nil {
			t.Fatal(err)
		}
		reader, size, err := f.Root.OpenPublished(t.Context(), ref)
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(io.NewSectionReader(reader, 0, wantSize+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || size != wantSize || !bytes.Equal(data, f.Pack.Container) {
			t.Fatalf("recovery input bytes differ: size %d, read %v, close %v", size, readErr, closeErr)
		}
	}
	if rows.Err() != nil || count != 1 {
		t.Fatalf("retained recovery input count = %d, want 1: %v", count, rows.Err())
	}
}

func requirePortableFatalState(t testing.TB, f *referencetest.Portable, before map[string]string) {
	t.Helper()
	for table, after := range snapshotPortableFatalState(t, f) {
		if before[table] != after {
			t.Fatalf("fatal completion changed durable %s rows", table)
		}
	}
}

// Release barriers and join before NewPortable tears down its database and roots.
func joinPortableFixture(t testing.TB, cancel context.CancelFunc, finished <-chan struct{}, releases ...func()) {
	t.Helper()
	cancel()
	for _, release := range releases {
		release()
	}
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Error("portable fixture did not finish after cancellation and barrier release")
	}
}

func awaitPortableCompletion(t testing.TB, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("portable operation did not complete after barrier release")
		return nil
	}
}

// Hold the real query result at the driver boundary while the operation exhausts
// grace. This exercises a late lock result without sleeps or database polling.
type portableAbortLockTx struct {
	pgx.Tx
	entered, release chan struct{}
}

func (tx *portableAbortLockTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	row := tx.Tx.QueryRow(ctx, query, args...)
	if tx.entered == nil {
		return row
	}
	entered := tx.entered
	tx.entered = nil
	return portableAbortLockRow{row, entered, tx.release}
}

type portableAbortLockRow struct {
	pgx.Row
	entered, release chan struct{}
}

func (row portableAbortLockRow) Scan(dest ...any) error {
	close(row.entered)
	<-row.release
	return row.Row.Scan(dest...)
}

func testPortableFatalDuringAbortLock(t *testing.T) {
	f := referencetest.NewPortable(t)
	fatal := make(chan error, 1)
	lifetime, err := reference_data.NewVerificationLifetime(20*time.Millisecond, func(err error) { fatal <- err })
	if err != nil {
		t.Fatal(err)
	}
	service := f.Service(f.Root, lifetime)
	request := f.Request()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	releaseCallback := sync.OnceFunc(func() { close(release) })
	done := make(chan error, 1)
	finished := make(chan struct{})
	defer joinPortableFixture(t, cancel, finished, releaseCallback)
	go func() {
		defer close(finished)
		done <- service.RunImportExecution(ctx, request.Execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
			request.ExecutionScope = scope
			if _, err := service.PrepareImport(run, request); err != nil {
				return err
			}
			outcome, _, _, err := scope.ClassifyAbsent(context.Background(), errors.New("ordinary failure before cancellation"))
			if err != nil || outcome != "execution_failed" {
				t.Error("classification failed", outcome, err)
				return err
			}
			tx, err := f.Pool.Begin(context.Background())
			if err != nil {
				return err
			}
			defer tx.Rollback(context.Background())
			err = scope.AbortTx(context.Background(), &portableAbortLockTx{tx, entered, release})
			if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Error("late lock authorized abort", err)
			}
			// Commit even the empty test transaction so a missing guard cannot be
			// hidden by a fixture rollback. No owner abort may have been written.
			if commitErr := tx.Commit(context.Background()); commitErr != nil {
				t.Error(commitErr)
			}
			return err
		})
	}()
	select {
	case <-entered:
	case err := <-done:
		t.Fatal("abort lock not reached", err)
	case <-time.After(10 * time.Second):
		t.Fatal("abort lock not reached")
	}
	beforeState := snapshotPortableFatalState(t, f)
	requirePortableRecoveryInput(t, f, request.Execution.JobID())
	cancel()
	select {
	case <-fatal:
	case <-time.After(time.Second):
		t.Error("abort wait escaped grace")
	}
	releaseCallback()
	if err := awaitPortableCompletion(t, done); !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		t.Fatal(err)
	}
	requirePortableFatalState(t, f, beforeState)
	requirePortableRecoveryInput(t, f, request.Execution.JobID())
	var terminal int
	if err := f.Pool.QueryRow(t.Context(), `SELECT count(*) FROM reference_pack_operations WHERE job_id=$1 AND terminal_at IS NOT NULL`, request.Execution.JobID()).Scan(&terminal); err != nil || terminal != 0 {
		t.Fatal("late abort terminated operation", terminal, err)
	}
	job, err := f.Manager.Get(t.Context(), request.Execution.JobID())
	if err != nil || job.Status != "running" {
		t.Fatal("late abort terminated parent", job.Status, err)
	}
}

// This is the production parent failure path, with real owner/Job/receipt SQL
// and the common finalizer. Only the scheduling barriers are test substitutes.
func testPortableFatalAfterAbort(t *testing.T, phase string) {
	f := referencetest.NewPortable(t)
	definitions := appsupport.RecognizedJobDefinitions(t)
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	f.Transactions = collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, reference_data.JobTerminalEffects{}, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	f.Manager, err = jobs.NewManager(jobs.ManagerOptions{Postgres: f.Pool, Transactions: f.Transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(f.Pack.Container)
	parent := incidentbundles.AdmitPortableFailureForTest(t, f.Pool, f.Transactions, f.Actor, hex.EncodeToString(digest[:]))
	execution, claimed, err := f.Manager.Claim(t.Context(), uuid.MustParse(parent.JobID))
	if err != nil || !claimed {
		t.Fatal("claim parent", claimed, err)
	}
	request := f.RequestFor(execution)
	entered := make(chan context.Context, 1)
	release := make(chan struct{})
	resume := make(chan struct{})
	releaseCallback := sync.OnceFunc(func() { close(release) })
	resumeCallback := sync.OnceFunc(func() { close(resume) })
	fatal := make(chan error, 1)
	grace := 50 * time.Millisecond
	if phase == "cooperative" {
		grace = time.Second
	}
	lifetime, err := reference_data.NewVerificationLifetime(grace, func(err error) { fatal <- err })
	if err != nil {
		t.Fatal(err)
	}
	service := f.Service(f.Root, lifetime)
	commonStore, err := extensionstore.New(f.Pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	barrier := func(ctx context.Context, tx pgx.Tx) {
		var terminal bool
		if err := tx.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE job_id=$1`, execution.JobID()).Scan(&terminal); err != nil || !terminal {
			t.Error("barrier did not follow AbortTx", terminal, err)
		}
		if phase == "after_receipt" {
			var status, receipt string
			if err := tx.QueryRow(ctx, `SELECT status FROM jobs WHERE job_id=$1`, execution.JobID()).Scan(&status); err != nil || status != jobs.StatusFailed {
				t.Error("Job write not reached", status, err)
			}
			if err := tx.QueryRow(ctx, `SELECT response_json->>'status' FROM route_idempotency WHERE client_txn_id='portable-fatal'`).Scan(&receipt); err != nil || receipt != jobs.StatusFailed {
				t.Error("receipt write not reached", receipt, err)
			}
		}
		entered <- ctx
		<-release
	}
	var receipts extensionstore.FinalReceiptReconciliationPort = collaborationsupport.NewJobOwnerTransactionAdapters()
	if phase == "after_receipt" {
		receipts = portableReceiptBarrier{receipts, barrier}
	}
	common, err := extensionstore.NewOwnerFinalizer(commonStore, f.Transactions, receipts, time.Now, func(err error) { t.Errorf("unexpected indeterminate commit: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	finalizer := extensionassembly.NewIncidentBundleJobSuccessFinalizer(common, time.Now)
	if phase != "after_receipt" {
		finalizer = portableMutationBarrier{finalizer, barrier}
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	prepared := make(chan struct{})
	finished := make(chan struct{})
	defer joinPortableFixture(t, cancel, finished, releaseCallback, resumeCallback)
	go func() {
		defer close(finished)
		done <- service.RunImportExecution(ctx, execution, request.ExecutionStarted, func(run context.Context, scope *reference_data.IncidentReferenceExecution) error {
			request.ExecutionScope = scope
			if _, err := service.PrepareImport(run, request); err != nil {
				return err
			}
			close(prepared)
			<-resume
			return incidentbundles.FinalizePortableFailureForTest(context.Background(), scope, execution, f.Pool, finalizer, errors.New("ordinary preparation failure"))
		})
	}()
	select {
	case <-prepared:
	case err := <-done:
		t.Fatal("preparation", err)
	case <-time.After(10 * time.Second):
		t.Fatal("preparation stalled")
	}
	beforeState := snapshotPortableFatalState(t, f)
	requirePortableRecoveryInput(t, f, execution.JobID())
	resumeCallback()
	var publication context.Context
	select {
	case publication = <-entered:
	case err := <-done:
		t.Fatal("post-abort barrier not reached", err)
	case <-time.After(10 * time.Second):
		t.Fatal("post-abort barrier stalled")
	}
	cancel()
	if phase != "cooperative" {
		// Delay actual process termination. The late database callback must be safe
		// even while the application has only recorded its fatal process signal.
		select {
		case err := <-fatal:
			if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
				t.Error(err)
			}
		case <-time.After(time.Second):
			t.Error("missing fatal consequence")
		}
		if !errors.Is(context.Cause(publication), extensiondeadline.ErrGraceExhausted) {
			t.Error("publication survived fatal selection", context.Cause(publication))
		}
	}
	select {
	case err := <-done:
		t.Fatal("returned with terminal transaction still active", err)
	default:
	}
	releaseCallback()
	err = awaitPortableCompletion(t, done)
	if phase == "cooperative" {
		if err != nil {
			t.Fatal("cooperative terminal publication failed", err)
		}
		job, err := f.Manager.Get(t.Context(), execution.JobID())
		if err != nil || job.Status != jobs.StatusFailed {
			t.Fatal("cooperative failure not published", job.Status, err)
		}
		select {
		case err := <-fatal:
			t.Fatal("normal completion manufactured exhaustion", err)
		default:
		}
		return
	}
	if !errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		t.Fatal("late completion replaced fatal", err)
	}
	requirePortableFatalState(t, f, beforeState)
	requirePortableRecoveryInput(t, f, execution.JobID())
}

type portableMutationBarrier struct {
	incidentbundles.JobSuccessFinalizer
	barrier func(context.Context, pgx.Tx)
}

func (f portableMutationBarrier) FinalizeIncidentBundleJobFailure(ctx context.Context, request incidentbundles.JobFailureFinalization) (jobs.Resource, error) {
	mutate := request.Mutate
	request.Mutate = func(ctx context.Context, tx pgx.Tx) error {
		return mutate(ctx, &portablePostAbortTx{Tx: tx, barrier: f.barrier})
	}
	return f.JobSuccessFinalizer.FinalizeIncidentBundleJobFailure(ctx, request)
}

type portablePostAbortTx struct {
	pgx.Tx
	barrier func(context.Context, pgx.Tx)
}

func (tx *portablePostAbortTx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(query, "UPDATE incident_bundle_job_payloads") {
		tx.barrier(ctx, tx.Tx)
	}
	return tx.Tx.Exec(ctx, query, args...)
}

type portableReceiptBarrier struct {
	extensionstore.FinalReceiptReconciliationPort
	barrier func(context.Context, pgx.Tx)
}

func (p portableReceiptBarrier) ReconcileFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, request []byte, resource jobs.Resource) (bool, error) {
	present, err := p.FinalReceiptReconciliationPort.ReconcileFinalIdempotencyOutcomeTx(ctx, tx, key, request, resource)
	if err == nil {
		p.barrier(ctx, tx)
	}
	return present, err
}
