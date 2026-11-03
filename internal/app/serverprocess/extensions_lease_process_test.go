package serverprocess

import (
	"database/sql"
	"errors"
	"net/http"
	"os/exec"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestExtensionsLeaseLifecycle_Process(t *testing.T) {
	pg, objects := sharedProcessHarnesses(t)
	testDB := pg.PrepareIsolatedDatabaseT(t, "extensions-lease-process")
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeSQL(t, db) })
	bucket := bucketName("extensions-lease-process")
	t.Cleanup(func() { cleanupBucket(t, objects, bucket) })
	env := newProcessEnv(t, processEnvOptions{
		Database: testDB.Env(), ObjectStore: objects.Env(bucket),
		ConfigPath:    writeConfig(t, string(fixtures.MustRead("config", "valid.toml"))),
		BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json"),
		Overrides: map[string]string{
			"CARTULARY_ENABLE_TEST_ROUTES":                                          "1",
			"CARTULARY_TEST_RUNTIME_MARKER":                                         "harness-owned",
			"CARTULARY_TEST_ROUTE_TOKEN":                                            httptestx.TestRouteToken,
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                             "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH":              fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                             "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":                        "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
			"CARTULARY__TIMEOUTS__EXTENSIONS__PROCESS_LEASE_ACQUIRE_SECONDS":        "1",
			"CARTULARY__TIMEOUTS__EXTENSIONS__PROCESS_LEASE_LOSS_DETECTION_SECONDS": "1",
		},
	})
	start := func() *processtest.Server {
		child := processtest.StartServer(t, processtest.ServerOptions{Env: env})
		t.Cleanup(func() { child.Stop(t) })
		return child
	}
	requireExit := func(child *processtest.Server, code int) {
		t.Helper()
		var exit *exec.ExitError
		if err := child.WaitForExit(t); !errors.As(err, &exit) || exit.ExitCode() != code {
			t.Fatalf("server exit got %v, want %d", err, code)
		}
		for _, path := range []string{"/healthz", "/readyz", "/api/v1/extensions"} {
			child.RequireConnectionRefused(t, path)
		}
		child.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
	}
	stateIdentity := func() string {
		t.Helper()
		var identity string
		if err := db.QueryRowContext(t.Context(), `
SELECT COALESCE(jsonb_agg(to_jsonb(s) ORDER BY profile_id), '[]'::jsonb)::text
FROM extension_state_metadata s`).Scan(&identity); err != nil {
			t.Fatal(err)
		}
		return identity
	}

	first := start()
	first.WaitForReady(t)
	before := stateIdentity()
	if before == "[]" {
		t.Fatal("claimed runtime did not initialize extension state")
	}
	login, _ := flowtest.ProvisionBootstrapAdmin(t, first.Client, first.BaseURL)
	options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	mutation := map[string]any{"client_txn_id": "extensions-lease-restart", "incident_key": "EXT-LEASE-RESTART", "title": "Retained across lease loss"}
	create := func(child *processtest.Server, expectedStatus int) string {
		t.Helper()
		response := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, child.Client, http.MethodPost, child.BaseURL+"/api/v1/incidents", mutation, options...), expectedStatus)
		return response["data"].(map[string]any)["incident_id"].(string)
	}
	incidentID := create(first, http.StatusCreated)
	overlap := start()
	requireExit(overlap, 2)
	overlap.RequireDiagnosticsField(t, "$", "extension_application_process_active")
	if stateIdentity() != before {
		t.Fatal("rejected overlapping process changed claimed state")
	}
	first.WaitForReady(t)

	// This fixture owns an otherwise idle isolated database. Require exactly
	// one exclusive runtime session; never infer an owner from a global PID or
	// duplicate the platform's private advisory-key value.
	var sessions, leasePID int
	if err := db.QueryRowContext(t.Context(), `
SELECT count(DISTINCT l.pid), COALESCE(min(l.pid),0)
FROM pg_catalog.pg_locks l
JOIN pg_catalog.pg_stat_activity a ON a.pid=l.pid
WHERE l.locktype='advisory' AND l.mode='ExclusiveLock' AND l.granted
  AND l.objsubid=1 AND pg_has_role(a.usesysid,'cartulary_runtime','MEMBER')
  AND l.database=(SELECT oid FROM pg_catalog.pg_database WHERE datname=current_database())
`).Scan(&sessions, &leasePID); err != nil || sessions != 1 {
		t.Fatalf("ambiguous owned application lease: sessions=%d error=%v", sessions, err)
	}
	var terminated bool
	if err := db.QueryRowContext(t.Context(), `SELECT pg_terminate_backend($1)`, leasePID).Scan(&terminated); err != nil || !terminated {
		t.Fatalf("terminate owned lease session: terminated=%v error=%v", terminated, err)
	}
	requireExit(first, 70)
	if stateIdentity() != before {
		t.Fatal("lease loss changed committed extension state")
	}

	replacement := start()
	replacement.WaitForReady(t)
	if stateIdentity() != before || create(replacement, http.StatusOK) != incidentID {
		t.Fatal("replacement changed claimed state or the committed mutation identity")
	}
	var incidents int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM incidents WHERE incident_key=$1`, mutation["incident_key"]).Scan(&incidents); err != nil || incidents != 1 {
		t.Fatalf("restart duplicated or lost the committed mutation: count=%d error=%v", incidents, err)
	}
}
