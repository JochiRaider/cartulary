package imports_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func currentOutcomeBytes(t testing.TB, db *sql.DB, unit string) string {
	t.Helper()
	var raw string
	if err := db.QueryRowContext(t.Context(), `SELECT COALESCE(jsonb_agg(to_jsonb(o)),'[]'::jsonb)::text FROM import_unit_apply_outcomes o WHERE import_unit_id=$1`, unit).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func installCurrentWriteWitness(t testing.TB, db *sql.DB, family, scenario string) func() {
	t.Helper()
	table := "records"
	if family == "analytical" {
		table = "network_flow_tables"
	}
	// Cancellation blocks a row insert after the owner has staged its parent;
	// other cases witness the first owner write, including failed duplicates.
	if scenario == "precommit_cancel" {
		table = "timeline_events"
		if family == "analytical" {
			table = "network_flow_rows"
		}
	}
	body := "PERFORM nextval('current_owner_write_attempts');"
	var blocker *sql.Conn
	if scenario == "precommit_cancel" {
		var err error
		blocker, err = db.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := blocker.ExecContext(t.Context(), `SELECT pg_advisory_lock(177,299)`); err != nil {
			t.Fatal(err)
		}
		body += " PERFORM pg_advisory_xact_lock(177,299);"
	}
	currentExec(t, db, `CREATE SEQUENCE current_owner_write_attempts; GRANT USAGE,SELECT ON SEQUENCE current_owner_write_attempts TO cartulary_runtime`)
	currentExec(t, db, `CREATE FUNCTION current_owner_write_witness() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN `+body+` RETURN NEW; END $$; CREATE TRIGGER current_owner_write_witness BEFORE INSERT ON `+table+` FOR EACH ROW EXECUTE FUNCTION current_owner_write_witness()`)
	return func() {
		if blocker != nil {
			_, err := blocker.ExecContext(context.Background(), `SELECT pg_advisory_unlock(177,299)`)
			if err != nil {
				t.Error(err)
			}
			if err := blocker.Close(); err != nil {
				t.Error(err)
			}
			blocker = nil
		}
	}
}

func installCurrentCommitFault(t testing.TB, db *sql.DB, scenario string) {
	t.Helper()
	table, condition := "import_unit_apply_outcomes", "true"
	if scenario == "after_commit_fault" {
		table, condition = "jobs", "NEW.status IN ('succeeded','failed','canceled')"
	}
	currentExec(t, db, `CREATE FUNCTION current_commit_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF `+condition+` THEN RAISE EXCEPTION 'injected current workflow boundary'; END IF; RETURN NEW; END $$; CREATE TRIGGER current_commit_fault BEFORE INSERT OR UPDATE ON `+table+` FOR EACH ROW EXECUTE FUNCTION current_commit_fault()`)
}

func recoverCurrentCommitFault(t testing.TB, db *sql.DB, runtime *appsupport.Runtime, first *appsupport.ServerHarness, options appsupport.ServerOptions, ids currentImportIDs, family, scenario string) *appsupport.ServerHarness {
	t.Helper()
	if got := waitForSQLCount(t, db, 20*time.Second, 1, `SELECT count(*) FROM jobs WHERE job_id::text=$1 AND handler_failure_count=1 AND handler_attempt_id IS NULL AND handler_next_attempt_at IS NOT NULL`, ids.job); got != 1 {
		t.Fatal("worker did not reach interrupted commit boundary")
	}
	wantEffects := 0
	table := "import_unit_apply_outcomes"
	if scenario == "after_commit_fault" {
		wantEffects = 1
		table = "jobs"
	}
	if got := currentEffectCount(t, db, family, ids.incident); got != wantEffects {
		t.Fatalf("interrupted effects %d, want %d", got, wantEffects)
	}
	if scenario == "after_commit_fault" {
		assertCurrentOutcome(t, db, ids, family, 1)
	} else if currentOutcomeBytes(t, db, ids.unit) != "[]" {
		t.Fatal("failed commit retained an outcome")
	}
	first.Server.Close()
	currentExec(t, db, `DROP TRIGGER current_commit_fault ON `+table+`; DROP FUNCTION current_commit_fault()`)
	options.ObjectStore = first.ObjectStore
	return runtime.StartServer(t, options)
}

func cancelCurrentBlockedUnit(t testing.TB, db *sql.DB, first *appsupport.ServerHarness, ids currentImportIDs) {
	t.Helper()
	if got := waitForSQLCount(t, db, 20*time.Second, 1, `SELECT count(*) FROM pg_locks WHERE locktype='advisory' AND classid=177 AND objid=299 AND NOT granted`); got != 1 {
		t.Fatal("current worker did not reach uncommitted owner write")
	}
	login := flowtest.LoginLocalUser(t, http.DefaultClient, first.Server.HTTP.URL, "current-import@example.test", "Current-import-test-42!", nil)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, first.Server.HTTP.URL+"/api/v1/jobs/"+ids.job+"/cancel", bytes.NewBufferString(`{"client_txn_id":"txn-current-cancel"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Log(databaseLockGraph(db))
		t.Fatalf("cancellation could not interrupt the blocked current unit: %v", err)
	}
	httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
}

func assertCurrentWriteAttempts(t testing.TB, db *sql.DB, scenario string, wantEffects int) {
	t.Helper()
	want := wantEffects
	switch scenario {
	case "before_commit_fault":
		want = 2
	case "precommit_cancel":
		want = 1
	}
	var got int
	if err := db.QueryRowContext(t.Context(), `SELECT CASE WHEN is_called THEN last_value ELSE 0 END FROM current_owner_write_attempts`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("owner write attempts %d, want %d for %s", got, want, scenario)
	}
}
