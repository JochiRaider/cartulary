package entities_test

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestEntityOriginUpsert_Integration(t *testing.T) {
	t.Run("host create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-02-host")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-02-host-incident",
			"incident_key":  "IR-I402-H",
			"title":         "Record relationships entity-resolution host create",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

		hostCreate := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":      "txn-entity_linking-i-4-02-host-create",
				"host.display_name":  "Gateway record",
				"host.hostname":      "GATEWAY-01",
				"host.aad_device_id": "AAD-DEVICE-GATEWAY-01",
				"host.fqdn":          "gateway-01.corp.example",
				"host.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "VPN Gateway"},
					},
				},
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostCreateData := appsupport.RequireSuccessData(t, hostCreate, http.StatusCreated)
		hostRow := hostCreateData["row"].(map[string]any)
		hostRecordID := appsupport.MustUUID(t, hostRow["record_id"].(string))

		var (
			hostState        string
			entityOrigin     string
			seedMentionID    sql.NullString
			displayName      string
			hostname         string
			aadDeviceID      sql.NullString
			fqdn             sql.NullString
			suggestionAliasN int
		)
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    host_state,
    entity_origin,
    seed_entity_mention_id::text,
    display_name,
    hostname,
    aad_device_id,
    fqdn,
    (
      SELECT COUNT(*)
        FROM entity_aliases
       WHERE incident_id = h.incident_id
         AND record_id = h.record_id
         AND entity_type = 'host'
         AND classification = 'suggestion_only'
         AND deleted_at IS NULL
    )
  FROM hosts h
 WHERE record_id = $1
`, hostRecordID).Scan(&hostState, &entityOrigin, &seedMentionID, &displayName, &hostname, &aadDeviceID, &fqdn, &suggestionAliasN); err != nil {
			t.Fatalf("lookup created host row: %v", err)
		}
		if hostState != "stub" || entityOrigin != "entity_sheet" || seedMentionID.Valid {
			t.Fatalf("expected entity_sheet host provenance without seed mention, got state=%q origin=%q seed=%v", hostState, entityOrigin, seedMentionID)
		}
		requireEntityOriginDefault(t, harness.DB, "hosts", "entity_sheet")
		requireEntityOriginRejected(t, harness.DB, "hosts", hostRecordID, "direct_create")
		requireEntityOriginRejected(t, harness.DB, "hosts", hostRecordID, "not_a_core02_origin")
		if displayName != "Gateway record" || hostname != "GATEWAY-01" || !aadDeviceID.Valid || aadDeviceID.String != "AAD-DEVICE-GATEWAY-01" || !fqdn.Valid || fqdn.String != "gateway-01.corp.example" || suggestionAliasN != 1 {
			t.Fatalf("unexpected created host state: display=%q hostname=%q aad=%v fqdn=%v aliases=%d", displayName, hostname, aadDeviceID, fqdn, suggestionAliasN)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("entity-origin host create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE hosts SET fqdn = NULL WHERE record_id = $1`, hostRecordID); err != nil {
			t.Fatalf("clear host fqdn to force preserved-identifier reuse: %v", err)
		}
		hostReuse := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":      "txn-entity_linking-i-4-02-host-reuse",
				"host.display_name":  "Gateway reused",
				"host.aad_device_id": "AAD-DEVICE-GATEWAY-01",
				"host.fqdn":          "gateway-01.corp.example",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReuseData := appsupport.RequireSuccessData(t, hostReuse, http.StatusOK)
		if got := appsupport.MustUUID(t, hostReuseData["row"].(map[string]any)["record_id"].(string)); got != hostRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original host record, got %#v", hostReuseData)
		}
		if state, mergedInto, rowVersion, restoredFQDN := entitytest.LookupHostState(t, harness.DB, hostRecordID); state != "stub" || mergedInto != nil || rowVersion != 2 || restoredFQDN != "gateway-01.corp.example" {
			t.Fatalf("unexpected reused host state: state=%s merged_into=%v row_version=%d fqdn=%q", state, mergedInto, rowVersion, restoredFQDN)
		}

		aliasOnly := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-entity_linking-i-4-02-host-alias-only",
				"host.display_name": "VPN Gateway",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := appsupport.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := appsupport.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == hostRecordID {
			t.Fatalf("expected alias-only create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		aliasPayloadOnly := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-entity_linking-i-4-02-host-alias-payload-only",
				"host.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "VPN Gateway Only"},
					},
				},
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, aliasPayloadOnly, http.StatusBadRequest, "invalid_mutation_payload")

		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "Conflict Host A", "COLLISION-01", "", "AAD-COLLISION-A")
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateHostRecordID, "Conflict Host B", "COLLISION-02", "", "AAD-COLLISION-B")
		conflictResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":      "txn-entity_linking-i-4-02-host-conflict",
				"host.display_name":  "Conflict Host",
				"host.aad_device_id": "AAD-COLLISION-B",
				"host.hostname":      "COLLISION-01",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := appsupport.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "host" || details["identifier_class"] != "aad_device_id" {
			t.Fatalf("unexpected host conflict details: %#v", details)
		}
		candidateIDs := details["candidate_record_ids"].([]any)
		if len(candidateIDs) != 2 {
			t.Fatalf("expected two conflict candidates, got %#v", details)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("conflicted host create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("identity create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-02-identity")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-02-identity-incident",
			"incident_key":  "IR-I402-I",
			"title":         "Record relationships entity-resolution identity create",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

		identityCreate := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":             "txn-entity_linking-i-4-02-identity-create",
				"identity.display_name":     "Alex Analyst",
				"identity.aad_object_id":    "AAD-OBJECT-ALEX-01",
				"identity.sid":              "S-1-5-21-401",
				"identity.email":            "alex.analyst@example.test",
				"identity.sam_account_name": "ALEXA",
				"identity.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "Case Owner"},
					},
				},
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityCreateData := appsupport.RequireSuccessData(t, identityCreate, http.StatusCreated)
		identityRecordID := appsupport.MustUUID(t, identityCreateData["row"].(map[string]any)["record_id"].(string))

		var (
			identityState   string
			entityOrigin    string
			seedMentionID   sql.NullString
			displayName     string
			aadObjectID     sql.NullString
			sid             sql.NullString
			email           sql.NullString
			samAccountName  sql.NullString
			suggestionCount int
		)
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    identity_state,
    entity_origin,
    seed_entity_mention_id::text,
    display_name,
    aad_object_id,
    sid,
    email::text,
    sam_account_name,
    (
      SELECT COUNT(*)
        FROM entity_aliases
       WHERE incident_id = i.incident_id
         AND record_id = i.record_id
         AND entity_type = 'identity'
         AND classification = 'suggestion_only'
         AND deleted_at IS NULL
    )
  FROM identities i
 WHERE record_id = $1
`, identityRecordID).Scan(&identityState, &entityOrigin, &seedMentionID, &displayName, &aadObjectID, &sid, &email, &samAccountName, &suggestionCount); err != nil {
			t.Fatalf("lookup created identity row: %v", err)
		}
		if identityState != "stub" || entityOrigin != "entity_sheet" || seedMentionID.Valid {
			t.Fatalf("expected entity_sheet identity provenance without seed mention, got state=%q origin=%q seed=%v", identityState, entityOrigin, seedMentionID)
		}
		requireEntityOriginDefault(t, harness.DB, "identities", "entity_sheet")
		requireEntityOriginRejected(t, harness.DB, "identities", identityRecordID, "direct_create")
		requireEntityOriginRejected(t, harness.DB, "identities", identityRecordID, "not_a_core02_origin")
		if displayName != "Alex Analyst" || !aadObjectID.Valid || aadObjectID.String != "AAD-OBJECT-ALEX-01" || !sid.Valid || sid.String != "S-1-5-21-401" || !email.Valid || email.String != "alex.analyst@example.test" || !samAccountName.Valid || samAccountName.String != "ALEXA" || suggestionCount != 1 {
			t.Fatalf("unexpected created identity state: display=%q aad=%v sid=%v email=%v sam=%v aliases=%d", displayName, aadObjectID, sid, email, samAccountName, suggestionCount)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("entity-origin identity create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE identities SET email = NULL WHERE record_id = $1`, identityRecordID); err != nil {
			t.Fatalf("clear identity email to force preserved-identifier reuse: %v", err)
		}
		identityReuse := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":          "txn-entity_linking-i-4-02-identity-reuse",
				"identity.display_name":  "Alex Analyst Reused",
				"identity.aad_object_id": "AAD-OBJECT-ALEX-01",
				"identity.sid":           "S-1-5-21-401",
				"identity.email":         "alex.analyst@example.test",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityReuseData := appsupport.RequireSuccessData(t, identityReuse, http.StatusOK)
		if got := appsupport.MustUUID(t, identityReuseData["row"].(map[string]any)["record_id"].(string)); got != identityRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original identity record, got %#v", identityReuseData)
		}
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM identities
 WHERE incident_id = $1
   AND record_id = $2
   AND row_version = 2
   AND email = 'alex.analyst@example.test'
`, incidentID, identityRecordID); got != 1 {
			t.Fatalf("expected preserved-identifier reuse to restore the canonical email and increment row_version, got %d rows", got)
		}

		aliasOnly := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-entity_linking-i-4-02-identity-alias-only",
				"identity.display_name": "Case Owner",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := appsupport.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := appsupport.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == identityRecordID {
			t.Fatalf("expected alias-only identity create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		aliasPayloadOnly := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-entity_linking-i-4-02-identity-alias-payload-only",
				"identity.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "Case Owner Only"},
					},
				},
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, aliasPayloadOnly, http.StatusBadRequest, "invalid_mutation_payload")

		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalIdentityRecordID, "Conflict Identity A", "collision-a@example.test", "collision-a@example.test", "COLLISION-A")
		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, entitytest.DuplicateIdentityRecordID, "Conflict Identity B", "collision-b@example.test", "collision-b@example.test", "COLLISION-B")
		conflictResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":             "txn-entity_linking-i-4-02-identity-conflict",
				"identity.display_name":     "Conflict Identity",
				"identity.email":            "collision-a@example.test",
				"identity.sam_account_name": "COLLISION-B",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := appsupport.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "identity" || details["identifier_class"] != "email" {
			t.Fatalf("unexpected identity conflict details: %#v", details)
		}
		if got := len(details["candidate_record_ids"].([]any)); got != 2 {
			t.Fatalf("expected two identity conflict candidates, got %#v", details)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("conflicted identity create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("create routes emit history, round-trip current-state query reads, and re-derive live authorization", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-02-query-auth")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-02-query-auth-incident",
			"incident_key":  "IR-I402-Q",
			"title":         "Record relationships entity-resolution query and auth",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		viewLogin := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}

		hostPayload := map[string]any{
			"client_txn_id":     "txn-entity_linking-i-4-02-query-host",
			"host.display_name": "Gateway query host",
			"host.hostname":     "GATEWAY-Q-01",
			"host.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "Gateway Query Alias"},
				},
			},
		}
		hostResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			hostPayload,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostData := appsupport.RequireSuccessData(t, hostResp, http.StatusCreated)
		hostRecordID := appsupport.MustUUID(t, hostData["row"].(map[string]any)["record_id"].(string))
		hostChangeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), hostData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: hostChangeSet.ActorUserID,
			Source:      hostChangeSet.Source,
			ClientTxnID: hostChangeSet.ClientTxnID,
			RequestID:   hostChangeSet.RequestID,
			CreatedAt:   hostChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.hosts.rows.create", "txn-entity_linking-i-4-02-query-host")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), hostData["change_set_id"].(string)); got != 2 {
			t.Fatalf("expected host and alias create mutation rows, got %d", got)
		}

		identityPayload := map[string]any{
			"client_txn_id":             "txn-entity_linking-i-4-02-query-identity",
			"identity.display_name":     "Alex Query",
			"identity.email":            "alex.query@example.test",
			"identity.sam_account_name": "ALEXQ",
			"identity.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "Query Owner"},
				},
			},
		}
		identityResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IdentitiesViewSchemaID+"/rows",
			identityPayload,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityData := appsupport.RequireSuccessData(t, identityResp, http.StatusCreated)
		identityRecordID := appsupport.MustUUID(t, identityData["row"].(map[string]any)["record_id"].(string))
		identityChangeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), identityData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: identityChangeSet.ActorUserID,
			Source:      identityChangeSet.Source,
			ClientTxnID: identityChangeSet.ClientTxnID,
			RequestID:   identityChangeSet.RequestID,
			CreatedAt:   identityChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.identities.rows.create", "txn-entity_linking-i-4-02-query-identity")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), identityData["change_set_id"].(string)); got != 2 {
			t.Fatalf("expected identity and alias create mutation rows, got %d", got)
		}

		hostEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, hostEnvelope, viewtest.HostsViewSchemaID)
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), hostRecordID.String())
		requireViewRowFieldSurface(t, "entity-resolution", hostRow, viewtest.HostsViewSchemaID)
		hostAlias := workbookscenariotest.RequireSingleCollectionItem(t, hostRow, "host.aliases")
		if hostAlias["item_kind"] != "alias" || hostAlias["alias_text"] != "Gateway Query Alias" || !strings.HasPrefix(hostAlias["item_ref"].(string), "entity_alias:") {
			t.Fatalf("unexpected host alias readback: %#v", hostAlias)
		}

		identityEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IdentitiesViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, identityEnvelope, viewtest.IdentitiesViewSchemaID)
		identityRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requireViewRowFieldSurface(t, "entity-resolution", identityRow, viewtest.IdentitiesViewSchemaID)
		identityAlias := workbookscenariotest.RequireSingleCollectionItem(t, identityRow, "identity.aliases")
		if identityAlias["item_kind"] != "alias" || identityAlias["alias_text"] != "Query Owner" || !strings.HasPrefix(identityAlias["item_ref"].(string), "entity_alias:") {
			t.Fatalf("unexpected identity alias readback: %#v", identityAlias)
		}

		hostProjectionBefore := lookupHostProjectionSnapshot(t, harness.DB, hostRecordID)
		identityProjectionBefore := lookupIdentityProjectionSnapshot(t, harness.DB, identityRecordID)
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM host_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
			t.Fatalf("clear host projection rows: %v", err)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM identity_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
			t.Fatalf("clear identity projection rows: %v", err)
		}
		projectionRebuild := harness.Projections
		if err := projectionRebuild.RebuildHosts(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild host projections: %v", err)
		}
		if err := projectionRebuild.RebuildIdentities(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild identity projections: %v", err)
		}
		hostProjectionAfter := lookupHostProjectionSnapshot(t, harness.DB, hostRecordID)
		identityProjectionAfter := lookupIdentityProjectionSnapshot(t, harness.DB, identityRecordID)
		contractassert.RequireProjectionDeterminism(t, hostProjectionBefore, hostProjectionAfter)
		contractassert.RequireProjectionDeterminism(t, identityProjectionBefore, identityProjectionAfter)

		hostRowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), hostRecordID.String())
		requireViewRowFieldSurface(t, "entity-resolution", hostRowAfterRebuild, viewtest.HostsViewSchemaID)
		contractassert.RequireProjectionDeterminism(t, hostRow["cells"], hostRowAfterRebuild["cells"])
		if rebuiltHostAlias := workbookscenariotest.RequireSingleCollectionItem(t, hostRowAfterRebuild, "host.aliases"); rebuiltHostAlias["alias_text"] != "Gateway Query Alias" {
			t.Fatalf("unexpected rebuilt host alias readback: %#v", rebuiltHostAlias)
		}

		identityRowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requireViewRowFieldSurface(t, "entity-resolution", identityRowAfterRebuild, viewtest.IdentitiesViewSchemaID)
		contractassert.RequireProjectionDeterminism(t, identityRow["cells"], identityRowAfterRebuild["cells"])
		if rebuiltIdentityAlias := workbookscenariotest.RequireSingleCollectionItem(t, identityRowAfterRebuild, "identity.aliases"); rebuiltIdentityAlias["alias_text"] != "Query Owner" {
			t.Fatalf("unexpected rebuilt identity alias readback: %#v", rebuiltIdentityAlias)
		}

		replayStableBefore := contractassert.ReplayCounts{
			ChangeSets: appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		}
		hostReplay := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			hostPayload,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReplayData := appsupport.RequireSuccessData(t, hostReplay, http.StatusOK)
		if hostReplayData["change_set_id"] != hostData["change_set_id"] {
			t.Fatalf("expected host replay to reuse the original payload, got %#v %#v", hostData, hostReplayData)
		}
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore:    replayStableBefore,
			StableAfter: contractassert.ReplayCounts{
				ChangeSets: appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
				MutationRows: appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
			},
		})

		hostDivergent := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-entity_linking-i-4-02-query-host",
				"host.display_name": "Gateway query host divergent",
				"host.hostname":     "GATEWAY-Q-01",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostDivergentBody := appsupport.RequireErrorBody(t, hostDivergent, http.StatusConflict, "client_txn_conflict")
		contractassert.RequireDivergentReplayRejected(t, hostDivergent.StatusCode, hostDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote entity create actor membership: %v", err)
		}
		deniedResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-entity_linking-i-4-02-query-host-denied",
				"host.display_name": "Denied host",
				"host.hostname":     "DENIED-HOST",
			},
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		deniedBody := appsupport.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")
		contractassert.RequireAuthorizationReDerived(
			t,
			contractassert.AuthorizationOutcome{Status: http.StatusCreated},
			contractassert.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})

	t.Run("ordinary patch partitions remain complete while create-only identifiers reject without effects", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-02-create-only-cutover")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity-linking-create-only-incident",
			"incident_key":  "IR-CREATE-ONLY-CUTOVER",
			"title":         "Entities create-only contract cutover",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		login := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}

		type entityCase struct {
			name             string
			viewSchemaID     string
			sourceTable      string
			projectionTable  string
			createPayload    map[string]any
			createOnlyValues map[string]string
			patchChanges     []map[string]any
		}
		cases := []entityCase{
			{
				name: "Host", viewSchemaID: viewtest.HostsViewSchemaID, sourceTable: "hosts", projectionTable: "host_grid_projection",
				createPayload: map[string]any{
					"client_txn_id":      "txn-entity-linking-create-only-host",
					"host.display_name":  "Create-only Host",
					"host.hostname":      "CREATE-ONLY-HOST",
					"host.aad_device_id": "AAD-DEVICE-CREATE-ONLY",
					"host.fqdn":          "create-only-host.example.test",
				},
				createOnlyValues: map[string]string{
					"host.aad_device_id": "AAD-DEVICE-CREATE-ONLY",
					"host.fqdn":          "create-only-host.example.test",
				},
				patchChanges: []map[string]any{
					{"field_key": "host.display_name", "value": "Patched Host"},
					{"field_key": "host.hostname", "value": "PATCHED-HOST"},
					{"field_key": "host.aliases", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_alias", "alias_text": "Patched Host Alias"}}}},
					{"field_key": "host.location", "value": "HQ"},
					{"field_key": "host.os_platform", "value": "Linux"},
					{"field_key": "host.business_owner", "value": "Security"},
					{"field_key": "host.criticality", "value": "high"},
					{"field_key": "host.containment_status", "value": "contained"},
				},
			},
			{
				name: "Identity", viewSchemaID: viewtest.IdentitiesViewSchemaID, sourceTable: "identities", projectionTable: "identity_grid_projection",
				createPayload: map[string]any{
					"client_txn_id":          "txn-entity-linking-create-only-identity",
					"identity.display_name":  "Create-only Identity",
					"identity.upn":           "create-only@example.test",
					"identity.aad_object_id": "AAD-OBJECT-CREATE-ONLY",
					"identity.sid":           "S-1-5-21-300",
				},
				createOnlyValues: map[string]string{
					"identity.aad_object_id": "AAD-OBJECT-CREATE-ONLY",
					"identity.sid":           "S-1-5-21-300",
				},
				patchChanges: []map[string]any{
					{"field_key": "identity.display_name", "value": "Patched Identity"},
					{"field_key": "identity.upn", "value": "patched@example.test"},
					{"field_key": "identity.email", "value": "patched.email@example.test"},
					{"field_key": "identity.sam_account_name", "value": "PATCHED"},
					{"field_key": "identity.aliases", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_alias", "alias_text": "Patched Identity Alias"}}}},
					{"field_key": "identity.privilege_level", "value": "admin"},
					{"field_key": "identity.mfa_state", "value": "enabled"},
					{"field_key": "identity.reset_status", "value": "complete"},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				created := createEntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchemaID, login, tc.createPayload, http.StatusCreated)
				row := created["row"].(map[string]any)
				recordID := appsupport.MustUUID(t, row["record_id"].(string))
				cells := row["cells"].(map[string]any)
				if len(cells) != 15 || len(row) != 4 {
					t.Fatalf("%s create row must expose exactly 15 cells and root technical members, got %#v", tc.name, row)
				}
				for fieldKey, want := range tc.createOnlyValues {
					if got := cells[fieldKey].(map[string]any)["value"]; got != want {
						t.Fatalf("%s create-only field %s not visible after create: got %#v want %q", tc.name, fieldKey, got, want)
					}
				}

				type patchSnapshot struct {
					recordVersion     int
					sourceVersion     int
					projectionVersion int
					revisions         int
					mutations         int
					collaboration     int
				}
				snapshot := func() patchSnapshot {
					return patchSnapshot{
						recordVersion:     appsupport.QueryCount(t, harness.DB, `SELECT row_version FROM records WHERE record_id = $1`, recordID),
						sourceVersion:     appsupport.QueryCount(t, harness.DB, `SELECT row_version FROM `+tc.sourceTable+` WHERE record_id = $1`, recordID),
						projectionVersion: appsupport.QueryCount(t, harness.DB, `SELECT row_version FROM `+tc.projectionTable+` WHERE record_id = $1`, recordID),
						revisions:         appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID),
						mutations:         appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE target_id = $1`, recordID.String()),
						collaboration:     appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE source_record_id = $1`, recordID),
					}
				}

				socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), tc.viewSchemaID, adminLogin.SessionCookie.Value)
				defer socket.Close(1000, "test_complete")
				beforePatch := snapshot()
				patchTxnID := "txn-entity-linking-exact-patch-" + strings.ToLower(tc.name)
				patchResp := appsupport.DoJSON(
					t,
					http.MethodPatch,
					harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(),
					map[string]any{
						"view_schema_id":   tc.viewSchemaID,
						"base_row_version": 1,
						"client_txn_id":    patchTxnID,
						"changes":          tc.patchChanges,
					},
					appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
					appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
				)
				patchData := appsupport.RequireSuccessData(t, patchResp, http.StatusOK)
				if got := int64(patchData["row"].(map[string]any)["row_version"].(float64)); got != 2 {
					t.Fatalf("%s exact eight-field patch row_version = %d, want 2", tc.name, got)
				}
				incidentwstest.RequireRecordChanged(t, socket, recordID.String(), 2)
				afterPatch := snapshot()
				if afterPatch.recordVersion != 2 || afterPatch.sourceVersion != 2 || afterPatch.projectionVersion != 2 ||
					afterPatch.revisions != beforePatch.revisions+1 || afterPatch.mutations <= beforePatch.mutations || afterPatch.collaboration != beforePatch.collaboration+1 {
					t.Fatalf("%s exact patch effects mismatch: before=%#v after=%#v", tc.name, beforePatch, afterPatch)
				}
				if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'workbook.records.patch'
   AND actor_user_id = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, adminUserID, recordID.String(), patchTxnID); got != 1 {
					t.Fatalf("%s accepted patch idempotency count = %d, want 1", tc.name, got)
				}

				for fieldKey, originalValue := range tc.createOnlyValues {
					t.Run("reject "+fieldKey, func(t *testing.T) {
						beforeReject := snapshot()
						clientTxnID := "txn-entity-linking-reject-" + strings.NewReplacer(".", "-").Replace(fieldKey)
						resp := appsupport.DoJSON(
							t,
							http.MethodPatch,
							harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(),
							map[string]any{
								"view_schema_id":   tc.viewSchemaID,
								"base_row_version": 2,
								"client_txn_id":    clientTxnID,
								"changes":          []map[string]any{{"field_key": fieldKey, "value": map[string]any{"malformed": true}}},
							},
							appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
							appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
						)
						body := appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
						details := body["error"].(map[string]any)["details"].(map[string]any)
						if details["field"] != "field_key" || details["reason_code"] != "unsupported_field_key" {
							t.Fatalf("%s must reject before value interpretation with canonical details, got %#v", fieldKey, details)
						}
						if afterReject := snapshot(); afterReject != beforeReject {
							t.Fatalf("%s rejection must have zero durable effects: before=%#v after=%#v", fieldKey, beforeReject, afterReject)
						}
						if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'workbook.records.patch' AND actor_user_id = $1 AND scope_key = $2 AND client_txn_id = $3`, adminUserID, recordID.String(), clientTxnID); got != 0 {
							t.Fatalf("%s rejection must not persist idempotency, got %d", fieldKey, got)
						}
						current := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchemaID, login), recordID.String())
						if got := current["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]; got != originalValue {
							t.Fatalf("%s rejection changed the create-only value: got %#v want %q", fieldKey, got, originalValue)
						}
					})
				}
				incidentwstest.ExpectNoSocketMessage(t, socket)
			})
		}
	})
}

func requireEntityOriginDefault(t *testing.T, db *sql.DB, tableName string, want string) {
	t.Helper()

	if tableName != "hosts" && tableName != "identities" {
		t.Fatalf("unsupported entity_origin table %q", tableName)
	}

	var defaultExpression string
	if err := db.QueryRowContext(context.Background(), `
SELECT column_default
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = 'entity_origin'
`, tableName).Scan(&defaultExpression); err != nil {
		t.Fatalf("lookup %s.entity_origin default: %v", tableName, err)
	}
	if !strings.Contains(defaultExpression, "'"+want+"'") {
		t.Fatalf("expected %s.entity_origin default %q, got %q", tableName, want, defaultExpression)
	}
}

func requireEntityOriginRejected(t *testing.T, db *sql.DB, tableName string, recordID uuid.UUID, origin string) {
	t.Helper()

	var query string
	switch tableName {
	case "hosts":
		query = `UPDATE hosts SET entity_origin = $1 WHERE record_id = $2`
	case "identities":
		query = `UPDATE identities SET entity_origin = $1 WHERE record_id = $2`
	default:
		t.Fatalf("unsupported entity_origin table %q", tableName)
	}

	if _, err := db.ExecContext(context.Background(), query, origin, recordID); err == nil {
		t.Fatalf("expected %s.entity_origin to reject %q", tableName, origin)
	}
}
