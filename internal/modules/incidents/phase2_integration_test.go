package incidents_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	incidentsmodule "github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase2_I_2_01_IncidentCreatePersistsBootstrapStateAndRollsBackAtomically(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("persists incident membership workbook preferences and audit attribution", func(t *testing.T) {
		server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-01-persist")
		defer db.Close()

		adminLogin, adminID := provisionBootstrapAdmin(t, server)
		createResp := doPhase2JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents",
			map[string]any{
				"client_txn_id": "txn-i-2-01-create",
				"incident_key":  "  IR-I201  ",
				"title":         "  Integration Incident  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		body := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)
		data := body["data"].(map[string]any)
		incidentID := data["incident_id"].(string)
		requestID := body["meta"].(map[string]any)["request_id"].(string)
		if got := createResp.Header.Get("Location"); got != "/api/v1/incidents/"+incidentID {
			t.Fatalf("unexpected incident Location header: got %q", got)
		}
		if data["status"] != "active" || data["incident_version"] != float64(1) {
			t.Fatalf("unexpected incident create payload: %#v", data)
		}

		if got := queryCount(t, db, `SELECT COUNT(*) FROM incidents WHERE id::text = $1`, incidentID); got != 1 {
			t.Fatalf("expected one incident row, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2 AND role = 'admin'`, incidentID, adminID); got != 1 {
			t.Fatalf("expected one bootstrap admin membership, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id::text = $1`, incidentID); got != 1 {
			t.Fatalf("expected one incident workbook preferences row, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, adminID); got != 1 {
			t.Fatalf("expected one user workbook preferences row, got %d", got)
		}

		events := lookupAuditEvents(t, db, incidentID)
		if len(events) != 2 {
			t.Fatalf("expected two incident audit events, got %#v", events)
		}
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: events[0].ActorUserID,
			Source:      events[0].EventSource,
			ClientTxnID: events[0].ClientTxnID,
			RequestID:   events[0].RequestID,
			CreatedAt:   events[0].CreatedAt,
		}, adminID, "incidents", "txn-i-2-01-create")
		if events[0].EventKind != "incident_created" || events[0].RequestID != requestID {
			t.Fatalf("unexpected incident_created audit event: %#v", events[0])
		}
		if events[1].EventKind != "incident_membership_created" {
			t.Fatalf("unexpected membership bootstrap audit event: %#v", events[1])
		}
		if afterIncidentID := events[0].After["incident_id"]; afterIncidentID != incidentID {
			t.Fatalf("unexpected incident audit payload: %#v", events[0].After)
		}
		if events[1].After["role"] != "admin" || events[1].After["user_id"] != adminID {
			t.Fatalf("unexpected membership audit payload: %#v", events[1].After)
		}
	})

	t.Run("forced pre-commit failure rolls back incident create atomically", func(t *testing.T) {
		restoreHooks := incidentsmodule.SetStoreHooksForTesting(incidentsmodule.StoreHooks{
			BeforeCommit: func(routeKey string, incidentID uuid.UUID) error {
				if routeKey == "incidents.create" {
					return errors.New("forced incidents rollback")
				}
				return nil
			},
		})
		defer restoreHooks()

		server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-01-rollback")
		defer db.Close()

		adminLogin, _ := provisionBootstrapAdmin(t, server)
		createResp := doPhase2JSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/incidents",
			map[string]any{
				"client_txn_id": "txn-i-2-01-rollback",
				"incident_key":  "IR-I201R",
				"title":         "Rollback Incident",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, createResp, http.StatusInternalServerError, "internal_error")

		if got := queryCount(t, db, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, "IR-I201R"); got != 0 {
			t.Fatalf("rollback must leave no incident rows, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships`); got != 0 {
			t.Fatalf("rollback must leave no incident memberships, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_workbook_preferences`); got != 0 {
			t.Fatalf("rollback must leave no incident workbook preferences, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM user_workbook_preferences`); got != 0 {
			t.Fatalf("rollback must leave no user workbook preferences, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE incident_id IS NOT NULL OR client_txn_id = $1`, "txn-i-2-01-rollback"); got != 0 {
			t.Fatalf("rollback must leave no incident-scoped audit events, got %d", got)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'incidents.create'`); got != 0 {
			t.Fatalf("rollback must leave no create idempotency rows, got %d", got)
		}
	})
}

func TestPhase2_I_2_02_IncidentCreateReplayAndDuplicateKeyConflictUseNormalizedState(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-02")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	create := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  " IR-I202 ",
			"title":         "Replay Incident",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	first := httptestx.RequireSuccessEnvelope(t, create, http.StatusCreated)["data"].(map[string]any)
	incidentID := first["incident_id"]
	stableBefore := phase2ReplayCounts(t, db, "IR-I202")

	replay := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  "IR-I202",
			"title":         "Replay Incident",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	replayBody := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if replayBody["incident_id"] != incidentID {
		t.Fatalf("expected idempotent replay to return original incident, got %#v", replayBody)
	}

	divergent := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  "IR-I202",
			"title":         "Different title",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	duplicateKey := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-duplicate",
			"incident_key":  "  IR-I202  ",
			"title":         "Duplicate key",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, duplicateKey, http.StatusConflict, "incident_key_conflict")

	httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
		FirstStatus:     http.StatusCreated,
		ReplayStatus:    http.StatusOK,
		DivergentStatus: http.StatusConflict,
		DivergentCode:   "client_txn_conflict",
		StableBefore:    stableBefore,
		StableAfter:     phase2ReplayCounts(t, db, "IR-I202"),
	})
}

func TestPhase2_I_2_03_MembershipChangesReDeriveAuthorizationImmediately(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-03")
	defer db.Close()

	adminLogin, adminID := provisionBootstrapAdmin(t, server)
	targetID := seedLocalUserFlags(t, db, "reviewer-target@example.test", "Reviewer Target", "ReviewerTargetPass1!", false, false, true)
	deploymentOnlyID := seedLocalUserFlags(t, db, "deployment-only@example.test", "Deployment Only", "DeploymentOnly1!", false, true, true)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-03-create",
		"incident_key":  "IR-I203",
		"title":         "Membership Lifecycle",
	})
	incidentID := incident["incident_id"].(string)

	membershipCreate := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-i-2-03-membership",
			"email":         " reviewer-target@example.test ",
			"role":          "viewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	createdMembership := httptestx.RequireSuccessEnvelope(t, membershipCreate, http.StatusCreated)["data"].(map[string]any)
	if createdMembership["user_id"] != targetID || createdMembership["role"] != "viewer" {
		t.Fatalf("unexpected created membership payload: %#v", createdMembership)
	}

	targetSession, targetCSRF := loginLocalUser(t, server, "reviewer-target@example.test", "ReviewerTargetPass1!")
	targetGet := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(targetSession))
	httptestx.RequireSuccessEnvelope(t, targetGet, http.StatusOK)

	targetPatchDenied := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "amber",
		},
		withCookies(targetSession, targetCSRF),
		withHeader(authn.CSRFHeaderName, targetCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, targetPatchDenied, http.StatusForbidden, "authorization_denied")

	memberPatch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetID,
		map[string]any{
			"base_membership_version": 1,
			"role":                    "reviewer",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, memberPatch, http.StatusOK)

	targetPatchAllowed := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"current_phase":         "eradication",
		},
		withCookies(targetSession, targetCSRF),
		withHeader(authn.CSRFHeaderName, targetCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, targetPatchAllowed, http.StatusOK)
	httptestx.RequireAuthorizationReDerived(t, httptestx.AuthorizationOutcome{Status: http.StatusForbidden, Code: "authorization_denied"}, httptestx.AuthorizationOutcome{Status: http.StatusOK})

	deploymentOnlySession, _ := loginLocalUser(t, server, "deployment-only@example.test", "DeploymentOnly1!")
	deploymentOnlyDenied := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(deploymentOnlySession))
	httptestx.RequireErrorEnvelope(t, deploymentOnlyDenied, http.StatusNotFound, "incident_not_found")
	requireTimelineSocketRejected(t, server.HTTP.URL, incidentID, deploymentOnlySession.Value, http.StatusNotFound, "incident_not_found")

	deleteMembership := doPhase2JSON(
		t,
		http.MethodDelete,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetID,
		map[string]any{
			"base_membership_version": 2,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireStatus(t, deleteMembership, http.StatusNoContent)

	postDelete := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/incidents/"+incidentID, nil, withCookies(targetSession))
	httptestx.RequireErrorEnvelope(t, postDelete, http.StatusNotFound, "incident_not_found")
	requireTimelineSocketRejected(t, server.HTTP.URL, incidentID, targetSession.Value, http.StatusNotFound, "incident_not_found")

	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, adminID); got != 1 {
		t.Fatalf("expected creator membership to remain, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, deploymentOnlyID); got != 0 {
		t.Fatalf("deployment-only user must not gain implicit incident membership, got %d", got)
	}
}

func TestPhase2_I_2_04_IncidentPatchPersistsOnlyPromotedFieldsAndAdvancesOnMaterialChange(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-04")
	defer db.Close()

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-04-create",
		"incident_key":  "IR-I204",
		"title":         "Patchable Incident",
		"description":   "Description stays fixed",
		"severity":      "high",
	})
	incidentID := incident["incident_id"].(string)
	initialUpdatedAt := incident["updated_at"]

	noOpPatch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	noOpBody := httptestx.RequireSuccessEnvelope(t, noOpPatch, http.StatusOK)["data"].(map[string]any)
	if noOpBody["incident_version"] != float64(1) || noOpBody["updated_at"] != initialUpdatedAt {
		t.Fatalf("unexpected no-op patch payload: %#v", noOpBody)
	}

	time.Sleep(20 * time.Millisecond)
	patch := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "amber",
			"current_phase":             "containment",
			"primary_external_case_ref": "CASE-I204",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "amber" || patchBody["current_phase"] != "containment" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
	if patchBody["updated_at"] == initialUpdatedAt {
		t.Fatalf("expected material patch to advance updated_at: before=%v after=%v", initialUpdatedAt, patchBody["updated_at"])
	}

	stale := doPhase2JSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "green",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "incident_version_conflict")

	row := db.QueryRowContext(context.Background(), `
SELECT title, description, severity, tlp, current_phase, primary_external_case_ref, incident_version
  FROM incidents
 WHERE id::text = $1
`, incidentID)
	var (
		title                  string
		description            *string
		severity               *string
		tlp                    *string
		currentPhase           *string
		primaryExternalCaseRef *string
		incidentVersion        int64
	)
	if err := row.Scan(&title, &description, &severity, &tlp, &currentPhase, &primaryExternalCaseRef, &incidentVersion); err != nil {
		t.Fatalf("query patched incident: %v", err)
	}
	if title != "Patchable Incident" || description == nil || *description != "Description stays fixed" || severity == nil || *severity != "high" {
		t.Fatalf("patch must not persist non-promoted fields: title=%q description=%v severity=%v", title, description, severity)
	}
	if tlp == nil || *tlp != "amber" || currentPhase == nil || *currentPhase != "containment" || primaryExternalCaseRef == nil || *primaryExternalCaseRef != "CASE-I204" || incidentVersion != 2 {
		t.Fatalf("patch must persist promoted fields only: tlp=%v current_phase=%v primary_external_case_ref=%v incident_version=%d", tlp, currentPhase, primaryExternalCaseRef, incidentVersion)
	}
}

func TestPhase2_I_2_05_ExtensionDiscoveryReturnsExactZeroMembershipShapeWithoutLeaks(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-05")
	defer db.Close()

	_, _ = provisionBootstrapAdmin(t, server)
	userID := seedLocalUserFlags(t, db, "extension-user@example.test", "Extension User", "ExtensionUser1!", false, false, true)
	userSession, _ := loginLocalUser(t, server, "extension-user@example.test", "ExtensionUser1!")

	extensionsResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/extensions", nil, withCookies(userSession))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensionsResp, http.StatusOK)
	data := extensionsBody["data"].(map[string]any)
	extensions := data["extensions"].([]any)
	if len(extensions) != 5 {
		t.Fatalf("unexpected extensions payload: %#v", data)
	}
	for index, raw := range extensions {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected extension item payload: %T", raw)
		}
		if len(item) != 3 {
			t.Fatalf("extension discovery must not leak provider secrets or claim maps: %#v", item)
		}
		if _, ok := item["profile_id"]; !ok {
			t.Fatalf("missing profile_id in extension item %d: %#v", index, item)
		}
		if _, ok := item["claimed"]; !ok {
			t.Fatalf("missing claimed in extension item %d: %#v", index, item)
		}
		if _, ok := item["route_families"]; !ok {
			t.Fatalf("missing route_families in extension item %d: %#v", index, item)
		}
	}
	if got := phase2OrderedProfileIDs(t, extensions); strings.Join(got, ",") != "enterprise_authentication,import,incident_portability,reference_pack,snapshot_reporting" {
		t.Fatalf("unexpected ordered extension profile set: %v", got)
	}

	encoded, err := json.Marshal(extensionsBody)
	if err != nil {
		t.Fatalf("marshal extensions body: %v", err)
	}
	for _, forbidden := range []string{"provider_secret", "claim_map", "live_payload", "provider_claims"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("extension discovery must not leak %q: %s", forbidden, encoded)
		}
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1`, userID); got != 0 {
		t.Fatalf("zero-membership user must stay outside incident scope, got %d memberships", got)
	}
}

func TestPhase2_I_2_06_UnclaimedReservedFamiliesReturnCanonical404AndOutsidePathsDoNot(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase2Server(t, postgresHarness, s3Harness, "phase2-i-2-06")
	defer db.Close()

	_, adminID := provisionBootstrapAdmin(t, server)
	seedLocalUserFlags(t, db, "reserved-user@example.test", "Reserved User", "ReservedUser1!", false, false, true)
	userSession, _ := loginLocalUser(t, server, "reserved-user@example.test", "ReservedUser1!")

	rootReserved := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/import-sessions", nil, withCookies(userSession))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != "import" || rootDetails["route_family"] != "/api/v1/import-sessions" {
		t.Fatalf("unexpected reserved root dispatch details: %#v", rootDetails)
	}

	descendantReserved := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+adminID+"/auth-bindings/provider", nil, withCookies(userSession))
	descendantBody := httptestx.RequireErrorEnvelope(t, descendantReserved, http.StatusNotFound, "extension_profile_not_claimed")
	descendantDetails := descendantBody["error"].(map[string]any)["details"].(map[string]any)
	if descendantDetails["profile_id"] != "enterprise_authentication" || descendantDetails["route_family"] != "/api/v1/users/{user_id}/auth-bindings" {
		t.Fatalf("unexpected reserved descendant dispatch details: %#v", descendantDetails)
	}

	outsideReserved := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/outside-reserved-families", nil, withCookies(userSession))
	if outsideReserved.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected outside reserved status: got %d", outsideReserved.StatusCode)
	}
	bodyBytes, err := ioReadAll(outsideReserved.Body)
	if err != nil {
		t.Fatalf("read outside reserved body: %v", err)
	}
	if strings.Contains(string(bodyBytes), "extension_profile_not_claimed") {
		t.Fatalf("outside reserved families must keep ordinary 404 handling, got %q", string(bodyBytes))
	}
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

type auditEventRecord struct {
	EventKind   string
	ActorUserID string
	EventSource string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
	Before      map[string]any
	After       map[string]any
}

func startPhase2Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})

	bucket, err := s3Harness.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	return server, db
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, string) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	login := loginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := doPhase2JSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, sessionData["user_id"].(string)
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase2JSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return userID
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return sessionCookie, csrfCookie
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doPhase2JSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func doPhase2JSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func lookupAuditEvents(t testing.TB, db *sql.DB, incidentID string) []auditEventRecord {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT event_kind,
       actor_user_id::text,
       event_source,
       COALESCE(client_txn_id, ''),
       COALESCE(request_id, ''),
       created_at,
       before_json,
       after_json
  FROM deployment_admin_audit_events
 WHERE incident_id::text = $1
 ORDER BY created_at ASC
`, incidentID)
	if err != nil {
		t.Fatalf("query audit events: %v", err)
	}
	defer rows.Close()

	events := make([]auditEventRecord, 0, 4)
	for rows.Next() {
		var (
			record        auditEventRecord
			beforePayload []byte
			afterPayload  []byte
		)
		if err := rows.Scan(&record.EventKind, &record.ActorUserID, &record.EventSource, &record.ClientTxnID, &record.RequestID, &record.CreatedAt, &beforePayload, &afterPayload); err != nil {
			t.Fatalf("scan audit event: %v", err)
		}
		record.Before = decodeJSONMap(t, beforePayload)
		record.After = decodeJSONMap(t, afterPayload)
		events = append(events, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate audit events: %v", err)
	}
	return events
}

func decodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	if len(payload) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode JSON payload: %v", err)
	}
	return decoded
}

func phase2ReplayCounts(t testing.TB, db *sql.DB, incidentKeyCanonical string) httptestx.ReplayCounts {
	t.Helper()
	return httptestx.ReplayCounts{
		ChangeSets: queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'incidents.create' AND scope_key IS NOT NULL`),
		MutationRows: queryCount(t, db, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE incident_id IN (SELECT id FROM incidents WHERE incident_key_canonical = $1)
`, incidentKeyCanonical),
		Revisions: queryCount(t, db, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, incidentKeyCanonical),
	}
}

func phase2OrderedProfileIDs(t testing.TB, extensions []any) []string {
	t.Helper()
	ordered := make([]string, 0, len(extensions))
	for _, raw := range extensions {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected extension item: %T", raw)
		}
		profileID, ok := item["profile_id"].(string)
		if !ok {
			t.Fatalf("unexpected profile_id payload: %#v", item)
		}
		ordered = append(ordered, profileID)
	}
	return ordered
}

func requireTimelineSocketRejected(t testing.TB, serverURL string, incidentID string, sessionToken string, wantStatus int, wantCode string) {
	t.Helper()

	target, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	target.Scheme = strings.Replace(target.Scheme, "http", "ws", 1)
	target.Path = "/ws/v1/incidents/" + incidentID + "/views/" + timeline.TimelineViewSchemaID + "/changes"

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, target.String(), &websocket.DialOptions{HTTPHeader: headers})
	if err == nil {
		t.Fatal("expected websocket dial to fail")
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for websocket dial, err=%v", err)
	}
	body := httptestx.RequireErrorEnvelope(t, resp, wantStatus, wantCode)
	_ = body
}

func ioReadAll(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return io.ReadAll(body)
}
