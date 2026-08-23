package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	contractentities "github.com/JochiRaider/cartulary/internal/gen/contractentities"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const entityIdentifierCorpusPath = "contracts/entities/identifier-normalization-corpus.v1.json"

type entityIdentifierNormalizationCorpus struct {
	SchemaID     string                              `json:"schema_id"`
	NormalizerID string                              `json:"normalizer_id"`
	Cases        []entityIdentifierNormalizationCase `json:"cases"`
}

type entityIdentifierNormalizationCase struct {
	CaseID          string  `json:"case_id"`
	IdentifierType  string  `json:"identifier_type"`
	RawValue        string  `json:"raw_value"`
	Admitted        bool    `json:"admitted"`
	NormalizedValue *string `json:"normalized_value"`
}

func TestEntityIdentifierNormalizationAndClaimsMigration_Integration(t *testing.T) {
	t.Run("valid 36 to 37 upgrade projects the closed corpus and supports disposable rollback", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 36)
		fixture := seedSourceIntegrityFixture(t, migrationDB.SQL())
		ctx := context.Background()

		if err := migrationDB.ApplyThrough(ctx, 37); err != nil {
			t.Fatalf("apply Entities active-identifier-claims migration: %v", err)
		}
		assertEntityActiveIdentifierClaimObjects(t, migrationDB.SQL(), true)
		assertLegacyEntityExactMatchIndexes(t, migrationDB.SQL(), false)
		assertEntityIdentifierNormalizationParity(t, migrationDB.SQL())

		var claimCount int
		if err := migrationDB.SQL().QueryRowContext(ctx, `
SELECT count(*)
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
`, fixture.incidentID).Scan(&claimCount); err != nil {
			t.Fatalf("count backfilled entity identifier claims: %v", err)
		}
		if claimCount != 6 {
			t.Fatalf("backfilled entity identifier claims = %d, want 6", claimCount)
		}
		if !entityIdentifierClaimsValid(t, migrationDB.SQL()) {
			t.Fatal("backfilled entity identifier claims failed exact-set validation")
		}
		assertEntityActiveIdentifierClaimPrivileges(t, migrationDB.SQL())

		if err := migrationDB.RollbackThrough(ctx, 36); err != nil {
			t.Fatalf("roll back Entities active-identifier-claims migration: %v", err)
		}
		assertEntityActiveIdentifierClaimObjects(t, migrationDB.SQL(), false)
		assertLegacyEntityExactMatchIndexes(t, migrationDB.SQL(), true)
		if err := migrationDB.ApplyThrough(ctx, 37); err != nil {
			t.Fatalf("reapply Entities active-identifier-claims migration: %v", err)
		}
		assertEntityActiveIdentifierClaimObjects(t, migrationDB.SQL(), true)
		assertLegacyEntityExactMatchIndexes(t, migrationDB.SQL(), false)
		if !entityIdentifierClaimsValid(t, migrationDB.SQL()) {
			t.Fatal("reapplied entity identifier claims failed exact-set validation")
		}
	})

	t.Run("duplicate active owners block preflight without schema mutation", func(t *testing.T) {
		harness := pgtest.Start(t)
		migrationDB := harness.MigrationDatabaseThroughT(t, 36)
		fixture := seedSourceIntegrityFixture(t, migrationDB.SQL())
		entitytest.SeedHostRecord(
			t,
			migrationDB.SQL(),
			fixture.incidentID,
			fixture.actorID,
			uuid.New(),
			"Competing source host",
			"SOURCE-HOST",
			"competing-source.example.test",
			"AAD-COMPETING-SOURCE",
		)

		err := migrationDB.ApplyThrough(context.Background(), 37)
		var postgresError *pgconn.PgError
		if err == nil ||
			!strings.Contains(err.Error(), "entities_active_identifier_claims_preflight_failed") ||
			!errors.As(err, &postgresError) ||
			!strings.Contains(postgresError.Detail, "duplicate_claim_groups=1") {
			t.Fatalf("claims preflight error = %v, want one duplicate claim group", err)
		}
		assertEntityActiveIdentifierClaimObjects(t, migrationDB.SQL(), false)
		assertLegacyEntityExactMatchIndexes(t, migrationDB.SQL(), true)

		var appliedHead int64
		if err := migrationDB.SQL().QueryRowContext(context.Background(), `
SELECT max(version_id)
  FROM goose_db_version
 WHERE is_applied
`).Scan(&appliedHead); err != nil {
			t.Fatalf("inspect migration head after claims preflight: %v", err)
		}
		if appliedHead != 36 {
			t.Fatalf("migration head after rejected claims preflight = %d, want 36", appliedHead)
		}
	})
}

func TestEntityActiveIdentifierClaimsAndConcurrentMatching_Integration(t *testing.T) {
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	database := runtime.PrepareServerDatabase(t, "entities-active-identifier-claims")
	harness := runtime.StartServerWithDatabase(t, "entities-active-identifier-claims", database)
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-entities-active-claims-incident",
		"incident_key":  "IR-ENT-CLAIMS",
		"title":         "Entities active identifier claims",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	hostURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() +
		"/views/" + viewtest.HostsViewSchemaID + "/rows"
	start := make(chan struct{})
	responses := make(chan *http.Response, 2)
	for index, hostname := range []string{"\u2003Concurrent.Host\u2003", "concurrent.host"} {
		go func(index int, hostname string) {
			<-start
			responses <- appsupport.DoJSON(
				t,
				http.MethodPost,
				hostURL,
				map[string]any{
					"client_txn_id":     "txn-entities-concurrent-host-" + string(rune('a'+index)),
					"host.display_name": "Concurrent host",
					"host.hostname":     hostname,
				},
				appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
			)
		}(index, hostname)
	}
	close(start)

	statuses := make([]int, 0, 2)
	recordIDs := make([]uuid.UUID, 0, 2)
	for range 2 {
		response := <-responses
		if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
			t.Fatalf("concurrent create status = %d, want 200 or 201", response.StatusCode)
		}
		statuses = append(statuses, response.StatusCode)
		data := appsupport.RequireSuccessData(t, response, response.StatusCode)
		recordIDs = append(recordIDs, appsupport.MustUUID(t, data["row"].(map[string]any)["record_id"].(string)))
	}
	slices.Sort(statuses)
	if !slices.Equal(statuses, []int{http.StatusOK, http.StatusCreated}) {
		t.Fatalf("concurrent create statuses = %#v, want one reuse and one create", statuses)
	}
	if recordIDs[0] != recordIDs[1] {
		t.Fatalf("concurrent exact matches created different records: %#v", recordIDs)
	}
	hostRecordID := recordIDs[0]
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "concurrent.host", hostRecordID)
	if got := appsupport.QueryCount(t, harness.DB, `
SELECT count(*) FROM hosts WHERE incident_id = $1 AND lower(hostname) = 'concurrent.host'
`, incidentID); got != 1 {
		t.Fatalf("concurrent matching host count = %d, want 1", got)
	}

	left := createEntityClaimTestRow(t, hostURL, adminLogin, map[string]any{
		"client_txn_id":     "txn-entities-claims-opposing-left",
		"host.display_name": "Opposing left",
		"host.hostname":     "opposing-left.example",
	}, http.StatusCreated)
	right := createEntityClaimTestRow(t, hostURL, adminLogin, map[string]any{
		"client_txn_id":     "txn-entities-claims-opposing-right",
		"host.display_name": "Opposing right",
		"host.hostname":     "opposing-right.example",
	}, http.StatusCreated)
	leftID := appsupport.MustUUID(t, left["row"].(map[string]any)["record_id"].(string))
	rightID := appsupport.MustUUID(t, right["row"].(map[string]any)["record_id"].(string))
	patchStart := make(chan struct{})
	patchResponses := make(chan *http.Response, 2)
	for index, patch := range []struct {
		recordID uuid.UUID
		value    string
	}{
		{recordID: leftID, value: "opposing-right.example"},
		{recordID: rightID, value: "opposing-left.example"},
	} {
		go func(index int, patch struct {
			recordID uuid.UUID
			value    string
		}) {
			<-patchStart
			patchResponses <- appsupport.DoJSON(
				t,
				http.MethodPatch,
				harness.Server.HTTP.URL+"/api/v1/records/"+patch.recordID.String(),
				map[string]any{
					"view_schema_id":   viewtest.HostsViewSchemaID,
					"base_row_version": 1,
					"client_txn_id":    "txn-entities-claims-opposing-patch-" + string(rune('a'+index)),
					"changes": []map[string]any{{
						"field_key": "host.hostname",
						"value":     patch.value,
					}},
				},
				appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
			)
		}(index, patch)
	}
	close(patchStart)
	for range 2 {
		response := <-patchResponses
		appsupport.RequireErrorBody(t, response, http.StatusConflict, "entity_match_conflict")
	}
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "opposing-left.example", leftID)
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "opposing-right.example", rightID)

	identityURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() +
		"/views/" + viewtest.IdentitiesViewSchemaID + "/rows"
	identityA := createEntityClaimTestRow(t, identityURL, adminLogin, map[string]any{
		"client_txn_id":             "txn-entities-claims-identity-a",
		"identity.display_name":     "Identity A",
		"identity.email":            "identity-a@example.test",
		"identity.sam_account_name": "CORP\\IDENTITY-A",
	}, http.StatusCreated)
	identityB := createEntityClaimTestRow(t, identityURL, adminLogin, map[string]any{
		"client_txn_id":             "txn-entities-claims-identity-b",
		"identity.display_name":     "Identity B",
		"identity.email":            "identity-b@example.test",
		"identity.sam_account_name": "CORP\\IDENTITY-B",
	}, http.StatusCreated)
	identityAID := appsupport.MustUUID(t, identityA["row"].(map[string]any)["record_id"].(string))
	identityBID := appsupport.MustUUID(t, identityB["row"].(map[string]any)["record_id"].(string))

	sameRecord := createEntityClaimTestRow(t, identityURL, adminLogin, map[string]any{
		"client_txn_id":             "txn-entities-claims-identity-same-record",
		"identity.display_name":     "Identity A",
		"identity.email":            "IDENTITY-A@EXAMPLE.TEST",
		"identity.sam_account_name": "corp\\identity-a",
	}, http.StatusOK)
	if got := appsupport.MustUUID(t, sameRecord["row"].(map[string]any)["record_id"].(string)); got != identityAID {
		t.Fatalf("multi-class same-record match = %s, want %s", got, identityAID)
	}

	crossRecordResponse := appsupport.DoJSON(
		t,
		http.MethodPost,
		identityURL,
		map[string]any{
			"client_txn_id":             "txn-entities-claims-identity-cross-record",
			"identity.display_name":     "Ambiguous identity",
			"identity.email":            "identity-a@example.test",
			"identity.sam_account_name": "CORP\\IDENTITY-B",
		},
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	conflictBody := appsupport.RequireErrorBody(t, crossRecordResponse, http.StatusConflict, "entity_match_conflict")
	details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "merge_required" || details["identifier_class"] != "email" {
		t.Fatalf("cross-record match conflict details = %#v", details)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM identities WHERE incident_id = $1`, incidentID); got != 2 {
		t.Fatalf("cross-record conflict identity count = %d, want 2", got)
	}
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "identity", "email", "identity-a@example.test", identityAID)
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "identity", "sam_account_name", "corp\\identity-b", identityBID)

	migrationDB, err := pgtest.OpenPurposeDatabase(database.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatalf("open migration-purpose claim probe: %v", err)
	}
	t.Cleanup(func() { _ = migrationDB.Close() })
	recoveryDB, err := pgtest.OpenPurposeDatabase(database.DSN, postgres.PurposeRecovery)
	if err != nil {
		t.Fatalf("open recovery-purpose claim probe: %v", err)
	}
	t.Cleanup(func() { _ = recoveryDB.Close() })
	runtimeRoleDB, err := pgtest.OpenPurposeDatabase(database.DSN, postgres.PurposeRuntime)
	if err != nil {
		t.Fatalf("open runtime-purpose claim probe: %v", err)
	}
	t.Cleanup(func() { _ = runtimeRoleDB.Close() })

	transitionEntityClaimLifecycle(t, migrationDB, hostRecordID, adminUserID, true)
	requireNoEntityIdentifierClaims(t, harness.DB, hostRecordID)
	transitionEntityClaimLifecycle(t, migrationDB, hostRecordID, adminUserID, false)
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "concurrent.host", hostRecordID)

	targetHost := createEntityClaimTestRow(t, hostURL, adminLogin, map[string]any{
		"client_txn_id":     "txn-entities-claims-merge-target",
		"host.display_name": "Merge target host",
		"host.hostname":     "merge-target.example",
	}, http.StatusCreated)
	targetHostID := appsupport.MustUUID(t, targetHost["row"].(map[string]any)["record_id"].(string))
	setEntityClaimMergeState(t, migrationDB, hostRecordID, targetHostID, adminUserID, true)
	requireNoEntityIdentifierClaims(t, harness.DB, hostRecordID)
	setEntityClaimMergeState(t, migrationDB, hostRecordID, targetHostID, adminUserID, false)
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "concurrent.host", hostRecordID)

	var runtimeCanWrite bool
	if err := runtimeRoleDB.QueryRowContext(ctx, `
SELECT has_table_privilege(current_user, 'entity_active_identifier_claims', 'INSERT')
    OR has_table_privilege(current_user, 'entity_active_identifier_claims', 'UPDATE')
    OR has_table_privilege(current_user, 'entity_active_identifier_claims', 'DELETE')
`).Scan(&runtimeCanWrite); err != nil {
		t.Fatalf("inspect runtime claim privileges: %v", err)
	}
	if runtimeCanWrite {
		t.Fatal("runtime role has direct claim DML")
	}
	if _, err := migrationDB.ExecContext(ctx, `DELETE FROM entity_active_identifier_claims WHERE record_id = $1`, hostRecordID); err != nil {
		t.Fatalf("corrupt derived entity claim through migration role: %v", err)
	}
	if entityIdentifierClaimsValid(t, recoveryDB) {
		t.Fatal("claim validation accepted missing derived state")
	}
	var rebuilt int64
	if err := recoveryDB.QueryRowContext(ctx, `SELECT entities_rebuild_active_identifier_claims_v1()`).Scan(&rebuilt); err != nil {
		t.Fatalf("rebuild entity identifier claims: %v", err)
	}
	if rebuilt < 1 || !entityIdentifierClaimsValid(t, recoveryDB) {
		t.Fatalf("rebuild result = %d with invalid claim set", rebuilt)
	}
	requireEntityIdentifierClaim(t, harness.DB, incidentID, "host", "hostname", "concurrent.host", hostRecordID)
	assertEntityClaimLookupUsesPrimaryKey(t, harness.DB, incidentID)
}

func assertEntityIdentifierNormalizationParity(t *testing.T, db *sql.DB) {
	t.Helper()
	artifact, ok := contractentities.Index[entityIdentifierCorpusPath]
	if !ok {
		t.Fatalf("generated Entities normalizer corpus %s is unavailable", entityIdentifierCorpusPath)
	}
	var corpus entityIdentifierNormalizationCorpus
	if err := json.Unmarshal([]byte(artifact.JSON), &corpus); err != nil {
		t.Fatalf("decode generated Entities normalizer corpus: %v", err)
	}
	if corpus.SchemaID != "cartulary.entities_identifier_normalization_corpus.v1" ||
		corpus.NormalizerID != "entities.identifier_normalization.v1" {
		t.Fatalf("unexpected Entities normalizer corpus identity: %#v", corpus)
	}
	for _, test := range corpus.Cases {
		t.Run(test.CaseID, func(t *testing.T) {
			goValue, goAdmitted := fieldnorm.NormalizeIdentifier(test.IdentifierType, test.RawValue)
			var sqlValue sql.NullString
			if err := db.QueryRowContext(context.Background(), `
SELECT entities_normalize_identifier_v1($1, $2)
`, test.IdentifierType, test.RawValue).Scan(&sqlValue); err != nil {
				t.Fatalf("normalize identifier through SQL projection: %v", err)
			}
			if goAdmitted != test.Admitted || sqlValue.Valid != test.Admitted {
				t.Fatalf("admission parity: Go=%t SQL=%t corpus=%t", goAdmitted, sqlValue.Valid, test.Admitted)
			}
			if !test.Admitted {
				return
			}
			if test.NormalizedValue == nil || goValue != *test.NormalizedValue || sqlValue.String != *test.NormalizedValue {
				t.Fatalf("normalization parity: Go=%q SQL=%q corpus=%v", goValue, sqlValue.String, test.NormalizedValue)
			}
		})
	}
}

func assertEntityActiveIdentifierClaimObjects(t testing.TB, db *sql.DB, present bool) {
	t.Helper()
	ctx := context.Background()
	for _, object := range []struct {
		kind string
		name string
	}{
		{kind: "table", name: "entity_active_identifier_claims"},
		{kind: "routine", name: "entities_normalize_identifier_v1(text,text)"},
		{kind: "routine", name: "entities_rebuild_active_identifier_claims_v1()"},
		{kind: "trigger", name: "hosts_sync_active_identifier_claims"},
		{kind: "trigger", name: "records_sync_entity_active_identifier_claims"},
	} {
		var exists bool
		var err error
		switch object.kind {
		case "table":
			err = db.QueryRowContext(ctx, `SELECT to_regclass('public.' || $1) IS NOT NULL`, object.name).Scan(&exists)
		case "routine":
			err = db.QueryRowContext(ctx, `SELECT to_regprocedure('public.' || $1) IS NOT NULL`, object.name).Scan(&exists)
		case "trigger":
			err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = $1 AND NOT tgisinternal)`, object.name).Scan(&exists)
		}
		if err != nil {
			t.Fatalf("inspect %s %s: %v", object.kind, object.name, err)
		}
		if exists != present {
			t.Fatalf("%s %s present = %t, want %t", object.kind, object.name, exists, present)
		}
	}
	if !present {
		return
	}
	var alwaysTriggerCount int
	if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM pg_trigger
 WHERE tgname IN (
    'hosts_sync_active_identifier_claims',
    'identities_sync_active_identifier_claims',
    'entity_preserved_identifiers_sync_active_claims',
    'records_sync_entity_active_identifier_claims'
 )
   AND tgenabled = 'A'
`).Scan(&alwaysTriggerCount); err != nil {
		t.Fatalf("inspect always-enabled entity claim triggers: %v", err)
	}
	if alwaysTriggerCount != 4 {
		t.Fatalf("always-enabled entity claim triggers = %d, want 4", alwaysTriggerCount)
	}
}

func assertLegacyEntityExactMatchIndexes(t testing.TB, db *sql.DB, present bool) {
	t.Helper()
	for _, name := range []string{
		"entity_preserved_identifiers_exact_lookup_idx",
		"hosts_incident_aad_device_id_idx",
		"hosts_incident_fqdn_idx",
		"hosts_incident_hostname_idx",
		"identities_incident_aad_object_id_idx",
		"identities_incident_email_idx",
		"identities_incident_sam_account_name_idx",
		"identities_incident_sid_idx",
		"identities_incident_upn_idx",
	} {
		var exists bool
		if err := db.QueryRowContext(
			context.Background(),
			`SELECT to_regclass('public.' || $1) IS NOT NULL`,
			name,
		).Scan(&exists); err != nil {
			t.Fatalf("inspect legacy exact-match index %s: %v", name, err)
		}
		if exists != present {
			t.Fatalf("legacy exact-match index %s present = %t, want %t", name, exists, present)
		}
	}
}

func assertEntityActiveIdentifierClaimPrivileges(t testing.TB, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	var publicCanUseClaimType bool
	if err := db.QueryRowContext(ctx, `
SELECT has_type_privilege(
    'public',
    'public.entity_active_identifier_claims',
    'USAGE'
)
`).Scan(&publicCanUseClaimType); err != nil {
		t.Fatalf("inspect PUBLIC active claim row-type privilege: %v", err)
	}
	if publicCanUseClaimType {
		t.Fatal("PUBLIC has USAGE on active claim row type")
	}
	for _, role := range []string{"cartulary_runtime", "cartulary_recovery"} {
		var canSelect, canWrite, canTruncate bool
		if err := db.QueryRowContext(ctx, `
SELECT has_table_privilege($1, 'entity_active_identifier_claims', 'SELECT'),
       has_table_privilege($1, 'entity_active_identifier_claims', 'INSERT')
        OR has_table_privilege($1, 'entity_active_identifier_claims', 'UPDATE')
        OR has_table_privilege($1, 'entity_active_identifier_claims', 'DELETE'),
       has_table_privilege($1, 'entity_active_identifier_claims', 'TRUNCATE')
`, role).Scan(&canSelect, &canWrite, &canTruncate); err != nil {
			t.Fatalf("inspect %s claim table privileges: %v", role, err)
		}
		wantTruncate := role == "cartulary_recovery"
		if !canSelect || canWrite || canTruncate != wantTruncate {
			t.Fatalf("%s claim privileges: select=%t write=%t truncate=%t", role, canSelect, canWrite, canTruncate)
		}
	}
	for _, signature := range []string{
		"public.entities_rebuild_active_identifier_claims_v1()",
		"public.entities_active_identifier_claims_are_valid_v1()",
	} {
		var recoveryCanExecute, publicCanExecute bool
		if err := db.QueryRowContext(ctx, `
SELECT has_function_privilege('cartulary_recovery', $1, 'EXECUTE'),
       has_function_privilege('public', $1, 'EXECUTE')
`, signature).Scan(&recoveryCanExecute, &publicCanExecute); err != nil {
			t.Fatalf("inspect claim recovery routine privileges: %v", err)
		}
		if !recoveryCanExecute || publicCanExecute {
			t.Fatalf("claim recovery routine %s: recovery=%t public=%t", signature, recoveryCanExecute, publicCanExecute)
		}
	}
	var runtimeCanRelease, runtimeCanRefresh, publicCanRelease, publicCanRefresh bool
	if err := db.QueryRowContext(ctx, `
SELECT has_function_privilege('cartulary_runtime', 'public.entities_release_active_identifier_claims_v1(uuid)', 'EXECUTE'),
	   has_function_privilege('cartulary_runtime', 'public.entities_refresh_active_identifier_claims_v1(uuid)', 'EXECUTE'),
	   has_function_privilege('public', 'public.entities_release_active_identifier_claims_v1(uuid)', 'EXECUTE'),
	   has_function_privilege('public', 'public.entities_refresh_active_identifier_claims_v1(uuid)', 'EXECUTE')
`).Scan(&runtimeCanRelease, &runtimeCanRefresh, &publicCanRelease, &publicCanRefresh); err != nil {
		t.Fatalf("inspect claim handoff routine privileges: %v", err)
	}
	if !runtimeCanRelease || !runtimeCanRefresh || publicCanRelease || publicCanRefresh {
		t.Fatalf(
			"claim handoff routines: runtime release=%t refresh=%t public release=%t refresh=%t",
			runtimeCanRelease,
			runtimeCanRefresh,
			publicCanRelease,
			publicCanRefresh,
		)
	}
}

func createEntityClaimTestRow(t testing.TB, url string, login appsupport.LoginResult, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	response := appsupport.DoJSON(
		t,
		http.MethodPost,
		url,
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	return appsupport.RequireSuccessData(t, response, wantStatus)
}

func transitionEntityClaimLifecycle(t testing.TB, db *sql.DB, recordID uuid.UUID, actorID uuid.UUID, deleted bool) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin entity claim lifecycle transition: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if deleted {
		if _, err := tx.ExecContext(ctx, `
UPDATE records
   SET deleted_at = now(), deleted_by_user_id = $2,
       updated_at = now(), updated_by_user_id = $2,
       row_version = row_version + 1
 WHERE record_id = $1
`, recordID, actorID); err != nil {
			t.Fatalf("delete entity Records envelope: %v", err)
		}
	} else if _, err := tx.ExecContext(ctx, `
UPDATE records
   SET deleted_at = NULL, deleted_by_user_id = NULL,
       updated_at = now(), updated_by_user_id = $2,
       row_version = row_version + 1
 WHERE record_id = $1
`, recordID, actorID); err != nil {
		t.Fatalf("restore entity Records envelope: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE hosts h
   SET row_version = r.row_version,
       updated_at = r.updated_at,
       updated_by_user_id = r.updated_by_user_id
  FROM records r
 WHERE h.record_id = $1
   AND r.record_id = h.record_id
`, recordID); err != nil {
		t.Fatalf("synchronize host lifecycle mirror: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit entity claim lifecycle transition: %v", err)
	}
}

func setEntityClaimMergeState(t testing.TB, db *sql.DB, recordID uuid.UUID, targetID uuid.UUID, actorID uuid.UUID, merged bool) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin entity claim merge transition: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
UPDATE records
   SET updated_at = now(), updated_by_user_id = $2,
       row_version = row_version + 1
 WHERE record_id = $1
`, recordID, actorID); err != nil {
		t.Fatalf("advance merge source envelope: %v", err)
	}
	state := "stub"
	var mergedInto any
	if merged {
		state = "merged"
		mergedInto = targetID
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE hosts h
   SET host_state = $2,
       merged_into_record_id = $3,
       row_version = r.row_version,
       updated_at = r.updated_at,
       updated_by_user_id = r.updated_by_user_id
  FROM records r
 WHERE h.record_id = $1
   AND r.record_id = h.record_id
`, recordID, state, mergedInto); err != nil {
		t.Fatalf("update host merge state: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit entity claim merge transition: %v", err)
	}
}

func requireEntityIdentifierClaim(t testing.TB, db *sql.DB, incidentID uuid.UUID, entityType string, identifierType string, normalizedValue string, recordID uuid.UUID) {
	t.Helper()
	var claimedRecordID uuid.UUID
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, incidentID, entityType, identifierType, normalizedValue).Scan(&claimedRecordID); err != nil {
		t.Fatalf("load active entity identifier claim: %v", err)
	}
	if claimedRecordID != recordID {
		t.Fatalf("active entity identifier claim = %s, want %s", claimedRecordID, recordID)
	}
}

func requireNoEntityIdentifierClaims(t testing.TB, db *sql.DB, recordID uuid.UUID) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*) FROM entity_active_identifier_claims WHERE record_id = $1
`, recordID).Scan(&count); err != nil {
		t.Fatalf("count active entity identifier claims: %v", err)
	}
	if count != 0 {
		t.Fatalf("active entity identifier claims = %d, want 0", count)
	}
}

func entityIdentifierClaimsValid(t testing.TB, db *sql.DB) bool {
	t.Helper()
	var valid bool
	if err := db.QueryRowContext(context.Background(), `
SELECT entities_active_identifier_claims_are_valid_v1()
`).Scan(&valid); err != nil {
		t.Fatalf("validate active entity identifier claims: %v", err)
	}
	return valid
}

func assertEntityClaimLookupUsesPrimaryKey(t testing.TB, db *sql.DB, incidentID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin entity claim plan probe: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL enable_seqscan = off`); err != nil {
		t.Fatalf("disable sequential scan for claim plan probe: %v", err)
	}
	rows, err := tx.QueryContext(ctx, `
EXPLAIN (COSTS OFF)
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = 'host'
   AND identifier_type = 'hostname'
   AND normalized_value = 'concurrent.host'
`, incidentID)
	if err != nil {
		t.Fatalf("explain entity claim lookup: %v", err)
	}
	defer rows.Close()
	planLines := make([]string, 0)
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan entity claim lookup plan: %v", err)
		}
		planLines = append(planLines, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate entity claim lookup plan: %v", err)
	}
	if plan := strings.Join(planLines, "\n"); !strings.Contains(plan, "entity_active_identifier_claims_pkey") {
		t.Fatalf("entity claim lookup plan does not use the primary key:\n%s", plan)
	}
}
