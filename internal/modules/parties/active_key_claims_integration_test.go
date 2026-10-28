package parties_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPartyActiveKeyClaimsBaseline_Integration(t *testing.T) {

	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	incidentID, actorID := seedPartyClaimIncident(t, db, "IR-PARTY-CLAIMS-BASELINE")
	recordID := seedPartyClaimRow(t, db, incidentID, actorID, "Claim Owner", "Owner@Example.Test", "EXT-Claim")
	assertPartyClaimObjects(t, db, true)
	requireMigrationPartyClaim(t, db, incidentID, "primary_email", "owner@example.test", recordID)
	requireMigrationPartyClaim(t, db, incidentID, "external_ref", "EXT-Claim", recordID)
	if !partyClaimsValid(t, db) {
		t.Fatal("new Party claim set is invalid")
	}
	assertPartyClaimPrivileges(t, db)

}

func TestPartyConcurrentCreateAndRecoveryClaims_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	database := runtime.PrepareIsolatedDatabase(t, "parties-active-key-claims")
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "parties-active-key-claims", Database: database, TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-parties-active-claims-incident",
		"incident_key":  "IR-PARTY-CLAIMS",
		"title":         "Party active key claims",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() +
		"/views/" + parties.ViewSchemaID + "/rows"

	start := make(chan struct{})
	responses := make(chan *http.Response, 2)
	for index, email := range []string{"Concurrent@Example.Test", " concurrent@example.test "} {
		go func(index int, email string) {
			<-start
			responses <- httptestx.DoJSON(t, http.DefaultClient,
				http.MethodPost,
				createURL,
				map[string]any{
					"client_txn_id":       "txn-party-concurrent-" + string(rune('a'+index)),
					"party.display_name":  "Concurrent Party",
					"party.party_kind":    "person",
					"party.primary_email": email,
				},
				httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
			)
		}(index, email)
	}
	close(start)
	var statuses []int
	var recordIDs []uuid.UUID
	for range 2 {
		response := <-responses
		if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
			t.Fatalf("concurrent Party create status = %d, want 200 or 201", response.StatusCode)
		}
		statuses = append(statuses, response.StatusCode)
		data := httptestx.RequireSuccessData(t, response, response.StatusCode)
		recordIDs = append(recordIDs, appsupport.MustUUID(t, data["row"].(map[string]any)["record_id"].(string)))
	}
	slices.Sort(statuses)
	if !slices.Equal(statuses, []int{http.StatusOK, http.StatusCreated}) || recordIDs[0] != recordIDs[1] {
		t.Fatalf("concurrent Party create did not converge: statuses=%v ids=%v", statuses, recordIDs)
	}
	requireRuntimePartyClaim(t, harness, incidentID, "primary_email", "concurrent@example.test", recordIDs[0])

	createParty := func(body map[string]any, wantStatus int) map[string]any {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, createURL, body,
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		return httptestx.RequireSuccessData(t, response, wantStatus)
	}
	createParty(map[string]any{
		"client_txn_id": "txn-party-cross-email", "party.display_name": "Email Owner",
		"party.party_kind": "organization", "party.primary_email": "cross@example.test",
	}, http.StatusCreated)
	createParty(map[string]any{
		"client_txn_id": "txn-party-cross-reference", "party.display_name": "Reference Owner",
		"party.party_kind": "organization", "party.external_ref": "CROSS-REF",
	}, http.StatusCreated)
	conflictResponse := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, createURL,
		map[string]any{
			"client_txn_id": "txn-party-cross-conflict", "party.display_name": "Cross Candidate",
			"party.party_kind": "organization", "party.primary_email": "cross@example.test",
			"party.external_ref": "CROSS-REF",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	conflictBody := httptestx.RequireErrorEnvelope(t, conflictResponse, http.StatusConflict, "party_match_conflict")
	conflictDetails := conflictBody["error"].(map[string]any)["details"].(map[string]any)
	if conflictDetails["reason_code"] != "cross_key_exact_match" {
		t.Fatalf("Party cross-key reason = %#v", conflictDetails)
	}
	fields, ok := conflictDetails["conflicting_field_keys"].([]any)
	if !ok || len(fields) != 2 || fields[0] != "party.external_ref" || fields[1] != "party.primary_email" {
		t.Fatalf("Party cross-key fields = %#v", conflictDetails["conflicting_field_keys"])
	}
	encodedConflict, err := json.Marshal(conflictBody)
	if err != nil {
		t.Fatalf("encode Party conflict body: %v", err)
	}
	for _, forbidden := range []string{"cross@example.test", "CROSS-REF", "party_active_key_claims", "constraint"} {
		if strings.Contains(string(encodedConflict), forbidden) {
			t.Fatalf("Party conflict exposed forbidden value %q: %s", forbidden, encodedConflict)
		}
	}

	migrationDB, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatalf("open migration-purpose Party claim probe: %v", err)
	}
	t.Cleanup(func() { _ = migrationDB.Close() })
	recoveryDB, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeRecovery)
	if err != nil {
		t.Fatalf("open recovery-purpose Party claim probe: %v", err)
	}
	t.Cleanup(func() { _ = recoveryDB.Close() })
	if _, err := migrationDB.ExecContext(context.Background(), `DELETE FROM party_active_key_claims WHERE party_record_id = $1`, recordIDs[0]); err != nil {
		t.Fatalf("corrupt derived Party claim through migration role: %v", err)
	}
	if partyClaimsValid(t, recoveryDB) {
		t.Fatal("Party claim validation accepted a missing claim")
	}
	var rebuilt int64
	if err := recoveryDB.QueryRowContext(context.Background(), `SELECT parties_rebuild_active_key_claims_v1()`).Scan(&rebuilt); err != nil {
		t.Fatalf("rebuild Party claims: %v", err)
	}
	if rebuilt < 1 || !partyClaimsValid(t, recoveryDB) {
		t.Fatalf("Party claim rebuild = %d with invalid resulting set", rebuilt)
	}
	requireRuntimePartyClaim(t, harness, incidentID, "primary_email", "concurrent@example.test", recordIDs[0])
}

func requireRuntimePartyClaim(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, keyKind, normalizedValue string, recordID uuid.UUID) {
	t.Helper()
	var ownerID uuid.UUID
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1 AND key_kind = $2 AND normalized_value = $3
`, incidentID, keyKind, normalizedValue).Scan(&ownerID); err != nil {
		t.Fatalf("load runtime Party claim %s: %v", keyKind, err)
	}
	if ownerID != recordID {
		t.Fatalf("runtime Party claim %s owner = %s, want %s", keyKind, ownerID, recordID)
	}
}

func seedPartyClaimIncident(t testing.TB, db *sql.DB, incidentKey string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	// Only the historical attribution identity is relevant to these upgrades.
	actorID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, is_active)
VALUES ($1, $2, 'Party Claim Actor', 'historical-migration-fixture-no-login', false)
`, actorID, strings.ToLower(incidentKey)+"@example.test"); err != nil {
		t.Fatalf("seed historical Party claim actor: %v", err)
	}
	incidentID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, lower($2), $3, 'active', $4, $4)
`, incidentID, incidentKey, "Party claim "+incidentKey, actorID); err != nil {
		t.Fatalf("seed Party claim incident: %v", err)
	}
	return incidentID, actorID
}

func seedPartyClaimRow(t testing.TB, db *sql.DB, incidentID, actorID uuid.UUID, displayName, email, externalRef string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'party', $3, $3)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed Party claim envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, primary_email, external_ref
) VALUES ($1, $2, $3, 'organization', $4, $5)
`, recordID, incidentID, displayName, email, externalRef); err != nil {
		t.Fatalf("seed Party claim source: %v", err)
	}
	return recordID
}

func assertPartyClaimObjects(t testing.TB, db *sql.DB, present bool) {
	t.Helper()
	for _, object := range []struct {
		kind string
		name string
	}{
		{kind: "table", name: "party_active_key_claims"},
		{kind: "routine", name: "parties_normalize_active_key_v1(text,text)"},
		{kind: "routine", name: "parties_rebuild_active_key_claims_v1()"},
		{kind: "trigger", name: "parties_sync_active_key_claims"},
		{kind: "trigger", name: "records_sync_party_active_key_claims"},
	} {
		var exists bool
		var err error
		switch object.kind {
		case "table":
			err = db.QueryRowContext(context.Background(), `SELECT to_regclass('public.' || $1) IS NOT NULL`, object.name).Scan(&exists)
		case "routine":
			err = db.QueryRowContext(context.Background(), `SELECT to_regprocedure('public.' || $1) IS NOT NULL`, object.name).Scan(&exists)
		case "trigger":
			err = db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = $1 AND NOT tgisinternal)`, object.name).Scan(&exists)
		}
		if err != nil {
			t.Fatalf("inspect %s %s: %v", object.kind, object.name, err)
		}
		if exists != present {
			t.Fatalf("%s %s present = %t, want %t", object.kind, object.name, exists, present)
		}
	}
}

func requireMigrationPartyClaim(t testing.TB, db *sql.DB, incidentID uuid.UUID, keyKind, normalizedValue string, recordID uuid.UUID) {
	t.Helper()
	var ownerID uuid.UUID
	if err := db.QueryRowContext(context.Background(), `
SELECT party_record_id
  FROM party_active_key_claims
 WHERE incident_id = $1 AND key_kind = $2 AND normalized_value = $3
`, incidentID, keyKind, normalizedValue).Scan(&ownerID); err != nil {
		t.Fatalf("load Party claim %s: %v", keyKind, err)
	}
	if ownerID != recordID {
		t.Fatalf("Party claim %s owner = %s, want %s", keyKind, ownerID, recordID)
	}
}

func partyClaimsValid(t testing.TB, db *sql.DB) bool {
	t.Helper()
	var valid bool
	if err := db.QueryRowContext(context.Background(), `SELECT parties_active_key_claims_are_valid_v1()`).Scan(&valid); err != nil {
		t.Fatalf("validate Party claims: %v", err)
	}
	return valid
}

func assertPartyClaimPrivileges(t testing.TB, db *sql.DB) {
	t.Helper()
	for _, role := range []string{"cartulary_runtime", "cartulary_recovery"} {
		var canSelect, canWrite, canTruncate bool
		if err := db.QueryRowContext(context.Background(), `
SELECT has_table_privilege($1, 'party_active_key_claims', 'SELECT'),
       has_table_privilege($1, 'party_active_key_claims', 'INSERT')
        OR has_table_privilege($1, 'party_active_key_claims', 'UPDATE')
        OR has_table_privilege($1, 'party_active_key_claims', 'DELETE'),
       has_table_privilege($1, 'party_active_key_claims', 'TRUNCATE')
`, role).Scan(&canSelect, &canWrite, &canTruncate); err != nil {
			t.Fatalf("inspect %s Party claim privileges: %v", role, err)
		}
		wantTruncate := role == "cartulary_recovery"
		if !canSelect || canWrite || canTruncate != wantTruncate {
			t.Fatalf("%s Party claim privileges: select=%t write=%t truncate=%t", role, canSelect, canWrite, canTruncate)
		}
	}
}
