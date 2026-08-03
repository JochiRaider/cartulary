package revisions_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type revisionAllocatorHarness struct {
	db         *pgxpool.Pool
	actor      authn.UserRecord
	incidentID uuid.UUID
	recordID   uuid.UUID
}

func TestRevisionsIncidentBundleNextIDsDoNotCollide(t *testing.T) {
	t.Run("imported maximum advances the allocator", func(t *testing.T) {
		harness := newRevisionAllocatorHarness(t, "imported-maximum")
		originalNext := readRevisionSequenceNext(t, harness.db)
		importedID := originalNext + 1000

		tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin imported revision transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		lockedNext, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(context.Background(), tx)
		if err != nil {
			t.Fatalf("lock imported revision sequence: %v", err)
		}
		if lockedNext != originalNext {
			t.Fatalf("locked original next = %d, want %d", lockedNext, originalNext)
		}
		insertImportedRevisionTx(t, tx, harness, importedID)
		if err := revisions.FinishIncidentBundleImportedRevisionSequenceTx(context.Background(), tx, lockedNext); err != nil {
			t.Fatalf("repair imported revision sequence: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit imported revision transaction: %v", err)
		}

		if next := consumeRevisionSequenceNext(t, harness.db); next != importedID+1 {
			t.Fatalf("next revision id = %d, want %d", next, importedID+1)
		}
	})

	t.Run("pre-existing sequence gap is retained", func(t *testing.T) {
		harness := newRevisionAllocatorHarness(t, "preexisting-gap")
		for range 5 {
			consumeRevisionSequenceNext(t, harness.db)
		}
		originalNext := readRevisionSequenceNext(t, harness.db)

		tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin gap-preservation transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		lockedNext, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(context.Background(), tx)
		if err != nil {
			t.Fatalf("lock gapped revision sequence: %v", err)
		}
		if err := revisions.FinishIncidentBundleImportedRevisionSequenceTx(context.Background(), tx, lockedNext); err != nil {
			t.Fatalf("finish gapped revision sequence: %v", err)
		}
		if err := tx.Commit(context.Background()); err != nil {
			t.Fatalf("commit gap-preservation transaction: %v", err)
		}

		if next := consumeRevisionSequenceNext(t, harness.db); next != originalNext {
			t.Fatalf("next revision id after empty import = %d, want retained %d", next, originalNext)
		}
	})
}

func TestRevisionsIncidentBundleSequenceRepairRollsBackOnLaterFailure(t *testing.T) {
	harness := newRevisionAllocatorHarness(t, "rollback")
	originalNext := readRevisionSequenceNext(t, harness.db)
	tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin rollback transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	lockedNext, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(context.Background(), tx)
	if err != nil {
		t.Fatalf("lock revision sequence: %v", err)
	}
	insertImportedRevisionTx(t, tx, harness, originalNext+1000)
	if err := revisions.FinishIncidentBundleImportedRevisionSequenceTx(context.Background(), tx, lockedNext); err != nil {
		t.Fatalf("repair revision sequence before injected failure: %v", err)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatalf("roll back injected later failure: %v", err)
	}

	if next := readRevisionSequenceNext(t, harness.db); next != originalNext {
		t.Fatalf("rolled-back repair left next revision id %d, want %d", next, originalNext)
	}
}

func TestRevisionsIncidentBundleSequenceRepairRunsAfterValidation(t *testing.T) {
	sourcePath := filepath.Join("..", "incidentbundles", "source.go")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read Incident Bundles coordinator: %v", err)
	}
	text := string(source)
	begin := strings.Index(text, "BeginIncidentBundleImportedRevisionSequenceTx")
	apply := strings.Index(text, ".ApplyImportTx(")
	validate := strings.Index(text, ".ValidateImportTx(")
	finish := strings.Index(text, "FinishIncidentBundleImportedRevisionSequenceTx")
	if begin < 0 || apply < 0 || validate < 0 || finish < 0 || !(begin < apply && apply < validate && validate < finish) {
		t.Fatalf("sequence protocol order begin=%d apply=%d validate=%d finish=%d", begin, apply, validate, finish)
	}
	for _, runtimePath := range []string{
		"incident_bundle_portability.go",
		filepath.Join("..", "incidentbundles", "source.go"),
	} {
		payload, err := os.ReadFile(runtimePath)
		if err != nil {
			t.Fatalf("read runtime sequence caller %s: %v", runtimePath, err)
		}
		if strings.Contains(strings.ToLower(string(payload)), "setval(") {
			t.Fatalf("runtime sequence caller %s still uses setval", runtimePath)
		}
	}
}

func TestRevisionsIncidentBundleSequenceLockExcludesConcurrentWriter(t *testing.T) {
	harness := newRevisionAllocatorHarness(t, "concurrent-writer")
	tx, err := harness.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin sequence lock transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	originalNext, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(context.Background(), tx)
	if err != nil {
		t.Fatalf("acquire sequence restart lock: %v", err)
	}

	started := make(chan struct{})
	result := make(chan struct {
		value int64
		err   error
	}, 1)
	go func() {
		close(started)
		var value int64
		err := harness.db.QueryRow(
			context.Background(),
			`SELECT nextval('public.record_revisions_revision_id_seq')`,
		).Scan(&value)
		result <- struct {
			value int64
			err   error
		}{value: value, err: err}
	}()
	<-started
	select {
	case outcome := <-result:
		t.Fatalf("concurrent nextval did not block: value=%d err=%v", outcome.value, outcome.err)
	case <-time.After(200 * time.Millisecond):
	}

	if err := revisions.FinishIncidentBundleImportedRevisionSequenceTx(context.Background(), tx, originalNext); err != nil {
		t.Fatalf("finish no-op sequence repair: %v", err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit sequence lock transaction: %v", err)
	}
	select {
	case outcome := <-result:
		if outcome.err != nil {
			t.Fatalf("concurrent nextval after lock release: %v", outcome.err)
		}
		if outcome.value != originalNext {
			t.Fatalf("concurrent nextval = %d, want %d", outcome.value, originalNext)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent nextval remained blocked after import commit")
	}
}

func TestRevisionsIncidentBundleSequenceFunctionsAreHardened(t *testing.T) {
	harness := newRevisionAllocatorHarness(t, "function-privileges")
	rows, err := harness.db.Query(context.Background(), `
SELECT procedure.proname,
       procedure.prosecdef,
       procedure.proconfig,
       has_function_privilege('public', procedure.oid, 'EXECUTE'),
       has_function_privilege(current_user, procedure.oid, 'EXECUTE')
  FROM pg_catalog.pg_proc AS procedure
  JOIN pg_catalog.pg_namespace AS namespace
    ON namespace.oid = procedure.pronamespace
 WHERE namespace.nspname = 'public'
   AND procedure.proname IN (
       'revisions_incident_bundle_sequence_begin_v1',
       'revisions_incident_bundle_sequence_finish_v1'
   )
 ORDER BY procedure.proname
`)
	if err != nil {
		t.Fatalf("query sequence function privileges: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var name string
		var securityDefiner bool
		var settings []string
		var publicExecute bool
		var applicationExecute bool
		if err := rows.Scan(&name, &securityDefiner, &settings, &publicExecute, &applicationExecute); err != nil {
			t.Fatalf("scan sequence function privilege: %v", err)
		}
		count++
		if !securityDefiner || publicExecute || !applicationExecute {
			t.Fatalf("function %s privilege posture: security_definer=%t public_execute=%t application_execute=%t", name, securityDefiner, publicExecute, applicationExecute)
		}
		if len(settings) != 1 || settings[0] != "search_path=pg_catalog, public" {
			t.Fatalf("function %s search_path = %#v", name, settings)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate sequence function privileges: %v", err)
	}
	if count != 2 {
		t.Fatalf("hardened sequence function count = %d, want 2", count)
	}
}

func newRevisionAllocatorHarness(t testing.TB, suffix string) revisionAllocatorHarness {
	t.Helper()
	postgresHarness := pgtest.Start(t)
	testDatabase := postgresHarness.PrepareIsolatedDatabaseT(t, "revisions-allocator-"+suffix)
	db, err := pgxpool.New(context.Background(), testDatabase.DSN)
	if err != nil {
		t.Fatalf("open allocator database: %v", err)
	}
	t.Cleanup(db.Close)
	actor := authstoretest.SeedLocalUserRecord(
		t,
		db,
		"revisions-allocator-"+uuid.NewString()+"@example.test",
		"Revisions Allocator",
		"RevisionsAllocator1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		db,
		actor,
		"txn-revisions-allocator-"+uuid.NewString(),
		"IR-RA-"+uuid.NewString()[:8],
		"Revisions allocator "+suffix,
	)
	recordID := uuid.New()
	if _, err := db.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id,
    updated_by_user_id, row_version
) VALUES ($1, $2, 'host', $3, $3, 1)
`, recordID, incident.ID, actor.ID); err != nil {
		t.Fatalf("seed allocator record: %v", err)
	}
	return revisionAllocatorHarness{
		db: db, actor: actor, incidentID: incident.ID, recordID: recordID,
	}
}

func insertImportedRevisionTx(
	t testing.TB,
	tx pgx.Tx,
	harness revisionAllocatorHarness,
	revisionID int64,
) {
	t.Helper()
	changeSetID := uuid.New()
	if _, err := tx.Exec(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source)
VALUES ($1, $2, $3, 'incident_bundle.import.test')
`, changeSetID, harness.incidentID, harness.actor.ID); err != nil {
		t.Fatalf("seed imported change set: %v", err)
	}
	if _, err := tx.Exec(context.Background(), `
INSERT INTO record_revisions (
    revision_id, change_set_id, record_id, row_version, before_json, after_json
) VALUES ($1, $2, $3, 1, NULL, '{}'::jsonb)
`, revisionID, changeSetID, harness.recordID); err != nil {
		t.Fatalf("seed imported revision: %v", err)
	}
}

func readRevisionSequenceNext(t testing.TB, db *pgxpool.Pool) int64 {
	t.Helper()
	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin sequence observation transaction: %v", err)
	}
	next, err := revisions.BeginIncidentBundleImportedRevisionSequenceTx(context.Background(), tx)
	if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
		t.Fatalf("roll back sequence observation: %v", rollbackErr)
	}
	if err != nil {
		t.Fatalf("observe sequence next value: %v", err)
	}
	return next
}

func consumeRevisionSequenceNext(t testing.TB, db *pgxpool.Pool) int64 {
	t.Helper()
	var next int64
	if err := db.QueryRow(
		context.Background(),
		`SELECT nextval('public.record_revisions_revision_id_seq')`,
	).Scan(&next); err != nil {
		t.Fatalf("consume revision sequence value: %v", err)
	}
	return next
}
