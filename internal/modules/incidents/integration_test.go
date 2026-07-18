package incidents_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	hostroutetest "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/faulttest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	indicatorroutetest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport/routetest"
	recordroutetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/routetest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	workbookroutetest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func TestIncidentCreatePersistsBootstrapStateAndRollsBackAtomically_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)

	t.Run("persists incident membership workbook preferences and audit attribution", func(t *testing.T) {
		harness := runtime.StartServer(t, "phase2-i-2-01-persist")

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
		createResp := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents",
			map[string]any{
				"client_txn_id": "txn-i-2-01-create",
				"incident_key":  "  IR-I201  ",
				"title":         "  Integration Incident  ",
			},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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

		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE id::text = $1`, incidentID); got != 1 {
			t.Fatalf("expected one incident row, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2 AND role = 'admin'`, incidentID, adminID); got != 1 {
			t.Fatalf("expected one bootstrap admin membership, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id::text = $1`, incidentID); got != 1 {
			t.Fatalf("expected one incident workbook preferences row, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, adminID); got != 1 {
			t.Fatalf("expected one user workbook preferences row, got %d", got)
		}

		resourceEvents := mutationtest.LookupOwnerMutations(
			t, mutationtest.SQLDatabase(

				harness.DB),

			mutationtest.MutationSelector{IncidentID: incidentID},
			mutationtest.MutationOwnerIncidentResource)

		membershipEvents := mutationtest.LookupOwnerMutations(
			t, mutationtest.SQLDatabase(

				harness.DB),

			mutationtest.MutationSelector{IncidentID: incidentID},
			mutationtest.MutationOwnerIncidentMembership)

		if len(resourceEvents) != 1 || len(membershipEvents) != 1 {
			t.Fatalf("expected one incident resource mutation and one incident membership mutation, got resource=%#v membership=%#v", resourceEvents, membershipEvents)
		}
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: resourceEvents[0].ActorUserID,
			Source:      resourceEvents[0].EventSource,
			ClientTxnID: resourceEvents[0].ClientTxnID,
			RequestID:   resourceEvents[0].RequestID,
			CreatedAt:   resourceEvents[0].CreatedAt,
		}, adminID, "incidents", "txn-i-2-01-create")
		if resourceEvents[0].EventKind != "incident_created" || resourceEvents[0].RequestID != requestID {
			t.Fatalf("unexpected incident resource mutation: %#v", resourceEvents[0])
		}
		bootstrapMembershipEvent := mutationtest.RequireOwnerMutationEvent(
			t,
			membershipEvents,
			mutationtest.MutationOwnerIncidentMembership,
			"incident_membership_created",
			adminID,
			adminID,
		)
		if bootstrapMembershipEvent.RequestID != requestID {
			t.Fatalf("unexpected membership bootstrap request_id: %#v", bootstrapMembershipEvent)
		}
		if afterIncidentID := resourceEvents[0].After["incident_id"]; afterIncidentID != incidentID {
			t.Fatalf("unexpected incident resource payload: %#v", resourceEvents[0].After)
		}
		if bootstrapMembershipEvent.After["role"] != "admin" || bootstrapMembershipEvent.After["user_id"] != adminID {
			t.Fatalf("unexpected incident membership payload: %#v", bootstrapMembershipEvent.After)
		}
	})

	t.Run("forced pre-commit failure rolls back incident create atomically", func(t *testing.T) {
		harness := runtime.StartServerWithTestDependencies(t, "phase2-i-2-01-rollback", faulttest.IncidentCreateRollbackFaultDependencies())

		adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
		createResp := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents",
			map[string]any{
				"client_txn_id": "txn-i-2-01-rollback",
				"incident_key":  "IR-I201R",
				"title":         "Rollback Incident",
			},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, createResp, http.StatusInternalServerError, "internal_error")

		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, "IR-I201R"); got != 0 {
			t.Fatalf("rollback must leave no incident rows, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships`); got != 0 {
			t.Fatalf("rollback must leave no incident memberships, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_workbook_preferences`); got != 0 {
			t.Fatalf("rollback must leave no incident workbook preferences, got %d", got)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences`); got != 0 {
			t.Fatalf("rollback must leave no user workbook preferences, got %d", got)
		}
		mutationtest.RequireNoMutationArtifacts(
			t, mutationtest.SQLDatabase(

				harness.DB),

			mutationtest.MutationSelector{ClientTxnID: "txn-i-2-01-rollback"},
			mutationtest.MutationOwnerIncidentResource,
			mutationtest.MutationOwnerIncidentMembership)

		if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'incidents.create'
   AND actor_user_id::text = $1
   AND scope_key = 'actor'
   AND client_txn_id = $2
`, adminID, "txn-i-2-01-rollback"); got != 0 {
			t.Fatalf("rollback must leave no create idempotency rows, got %d", got)
		}
	})
}

func TestIncidentCreateReplayAndDuplicateKeyConflictUseNormalizedState_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-i-2-02")
	normalizedIncidentKey := "IR-\u00C9-202"

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	create := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  "  IR-E\u0301-202  ",
			"title":         "Replay Incident",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	first := httptestx.RequireSuccessEnvelope(t, create, http.StatusCreated)["data"].(map[string]any)
	if first["incident_key"] != normalizedIncidentKey {
		t.Fatalf("expected normalized incident key in create response, got %#v", first)
	}
	incidentID := first["incident_id"].(string)
	replaySelector := storetest.IncidentCreateReplaySelector{
		ActorUserID: uuid.MustParse(first["created_by_user_id"].(string)),
		ClientTxnID: "txn-i-2-02-create",
		IncidentID:  uuid.MustParse(incidentID),
	}
	stableBefore := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.SQLReplayDatabase(harness.DB), replaySelector)

	replay := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  normalizedIncidentKey,
			"title":         "Replay Incident",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayBody := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if replayBody["incident_id"] != incidentID {
		t.Fatalf("expected idempotent replay to return original incident, got %#v", replayBody)
	}
	if replayBody["incident_key"] != normalizedIncidentKey {
		t.Fatalf("expected replay to keep the normalized incident key, got %#v", replayBody)
	}
	if stableAfterReplay := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.SQLReplayDatabase(harness.DB), replaySelector); stableAfterReplay != stableBefore {
		t.Fatalf("incident create replay must keep durable side effects stable: before=%+v after=%+v", stableBefore, stableAfterReplay)
	}

	divergent := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-i-2-02-create",
			"incident_key":  normalizedIncidentKey,
			"title":         "Different title",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	if stableAfterConflict := storetest.SnapshotIncidentCreateReplaySideEffects(t, storetest.SQLReplayDatabase(harness.DB), replaySelector); stableAfterConflict != stableBefore {
		t.Fatalf("divergent create replay must not change durable side effects: before=%+v after=%+v", stableBefore, stableAfterConflict)
	}

	duplicateClientTxnID := "txn-i-2-02-duplicate"
	type duplicateConflictSideEffects struct {
		BootstrapMembershipRows        int
		IncidentRows                   int
		IncidentWorkbookPreferenceRows int
		OwnerMutationRows              int
		RouteIdempotencyRows           int
		UserWorkbookPreferenceRows     int
	}
	snapshotDuplicateConflictSideEffects := func() duplicateConflictSideEffects {
		return duplicateConflictSideEffects{
			BootstrapMembershipRows: dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships`),
			IncidentRows:            dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incidents`),
			IncidentWorkbookPreferenceRows: dbassert.CountSQL(
				t,
				harness.DB,
				`SELECT COUNT(*) FROM incident_workbook_preferences`,
			),
			OwnerMutationRows: mutationtest.CountMutationArtifacts(
				t, mutationtest.SQLDatabase(

					harness.DB),

				mutationtest.MutationSelector{ClientTxnID: duplicateClientTxnID},
				mutationtest.MutationOwnerIncidentResource,
				mutationtest.MutationOwnerIncidentMembership),

			RouteIdempotencyRows: dbassert.CountSQL(
				t,
				harness.DB,
				`
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'incidents.create'
   AND actor_user_id::text = $1
   AND scope_key = 'actor'
   AND client_txn_id = $2
`,
				replaySelector.ActorUserID.String(),
				duplicateClientTxnID,
			),
			UserWorkbookPreferenceRows: dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM user_workbook_preferences`),
		}
	}
	duplicateBefore := snapshotDuplicateConflictSideEffects()
	duplicateKey := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		map[string]any{
			"client_txn_id": duplicateClientTxnID,
			"incident_key":  normalizedIncidentKey,
			"title":         "Duplicate key",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	duplicateBody := httptestx.RequireErrorEnvelope(t, duplicateKey, http.StatusConflict, "incident_key_conflict")
	duplicateDetails := duplicateBody["error"].(map[string]any)["details"].(map[string]any)
	if duplicateDetails["field"] != "incident_key" {
		t.Fatalf("expected duplicate conflict field detail incident_key, got %#v", duplicateDetails)
	}
	if duplicateDetails["incident_key_canonical"] != normalizedIncidentKey {
		t.Fatalf("expected duplicate conflict canonical key %q, got %#v", normalizedIncidentKey, duplicateDetails)
	}
	duplicateAfter := snapshotDuplicateConflictSideEffects()
	if duplicateAfter != duplicateBefore {
		t.Fatalf("duplicate incident key conflict must not change durable side effects: before=%+v after=%+v", duplicateBefore, duplicateAfter)
	}
	if duplicateAfter.OwnerMutationRows != 0 {
		t.Fatalf("duplicate incident key conflict must not leave mutation artifacts for %s, got %d", duplicateClientTxnID, duplicateAfter.OwnerMutationRows)
	}
	if duplicateAfter.RouteIdempotencyRows != 0 {
		t.Fatalf("duplicate incident key conflict must not leave create idempotency rows for %s, got %d", duplicateClientTxnID, duplicateAfter.RouteIdempotencyRows)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incidents WHERE incident_key_canonical = $1`, normalizedIncidentKey); got != 1 {
		t.Fatalf("expected one canonical incident row after composed-vs-decomposed duplicate conflict, got %d", got)
	}
}

func TestControlBoundaryIncidentCoreReDerivesAuthorizationImmediately_Integration(t *testing.T) {
	requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t, routetest.ControlIncidentCore())
}

func TestControlBoundaryMembershipAdminReDerivesAuthorizationImmediately_Integration(t *testing.T) {
	requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t, routetest.ControlMembershipAdmin())
}

func TestControlBoundaryWorkbookPreferencesReDerivesAuthorizationImmediately_Integration(t *testing.T) {
	requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t, workbookroutetest.ControlPreferences())
}

func TestControlBoundaryWorkbookQueriesReDerivesAuthorizationImmediately_Integration(t *testing.T) {
	routes := append(timelineroutetest.ControlQuery(), hostroutetest.ControlQueries()...)
	routes = append(routes, indicatorroutetest.ControlQuery()...)
	requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t, routes)
}

func TestControlBoundaryTimelineRecordAndLiveReDerivesAuthorizationImmediately_Integration(t *testing.T) {
	routes := append(timelineroutetest.ControlCreateAndLive(), recordroutetest.ControlMutations()...)
	requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t, routes)
}

func requireControlBoundaryInventoryReDerivesAuthorizationImmediately(t *testing.T, routes []routeinventory.Entry) {
	t.Helper()

	for _, route := range routes {
		route := route
		t.Run(route.Name, func(t *testing.T) {
			fixtureCtx := newRouteFixture(t, "phase2-i-2-03-"+route.Name)
			progressionSlug := FixtureSlug("phase2-i-2-03-" + route.Name + "-progression")
			progressionUserID := flowtest.SeedLocalUserFlags(
				t,
				fixtureCtx.harness.DB,
				progressionSlug+"@example.test",
				"Progression User "+progressionSlug,
				"ProgressionUser1!",
				false,
				false,
				true,
			)
			progressionSession, progressionCSRF := flowtest.LoginLocalUser(
				t,
				fixtureCtx.harness.Server.HTTP.URL,

				progressionSlug+"@example.test",
				"ProgressionUser1!", nil)

			requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name+"-no-membership"),
				progressionSession,
				progressionCSRF,
				controlStageNoMembership,
			)

			viewerMembership := scenariotest.CreateMembership(
				t,
				fixtureCtx.harness.Server,
				fixtureCtx.adminLogin,
				fixtureCtx.fixture.IncidentID,
				map[string]any{
					"client_txn_id": "txn-" + progressionSlug + "-viewer",
					"user_id":       progressionUserID,
					"role":          "viewer",
				},
			)
			ensureUserWorkbookPreferences(t, fixtureCtx, progressionUserID)
			requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name+"-viewer"),
				progressionSession,
				progressionCSRF,
				controlStageViewer,
			)

			reviewerMembership := scenariotest.PatchMembership(
				t,
				fixtureCtx.harness.Server,
				fixtureCtx.adminLogin,
				fixtureCtx.fixture.IncidentID,
				progressionUserID,
				map[string]any{
					"base_membership_version": viewerMembership["membership_version"],
					"role":                    "reviewer",
				},
			)
			reviewerData := requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name+"-reviewer"),
				progressionSession,
				progressionCSRF,
				controlStageReviewer,
			)
			updateRouteFixtureAfterSuccess(t, fixtureCtx, route, reviewerData, controlStageReviewer)

			adminMembership := scenariotest.PatchMembership(
				t,
				fixtureCtx.harness.Server,
				fixtureCtx.adminLogin,
				fixtureCtx.fixture.IncidentID,
				progressionUserID,
				map[string]any{
					"base_membership_version": reviewerMembership["membership_version"],
					"role":                    "admin",
				},
			)
			adminData := requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name+"-admin"),
				progressionSession,
				progressionCSRF,
				controlStageAdmin,
			)
			updateRouteFixtureAfterSuccess(t, fixtureCtx, route, adminData, controlStageAdmin)

			scenariotest.DeleteMembership(
				t,
				fixtureCtx.harness.Server,
				fixtureCtx.adminLogin,
				fixtureCtx.fixture.IncidentID,
				progressionUserID,
				map[string]any{
					"base_membership_version": adminMembership["membership_version"],
				},
			)
			requireControlRouteOutcome(
				t,
				fixtureCtx.harness.Server.HTTP.URL,
				route,
				fixtureCtx.routeFixture(route.Name+"-removed"),
				progressionSession,
				progressionCSRF,
				controlStageRemoved,
			)
		})
	}
}

func TestIncidentPatchPersistsOnlyPromotedFieldsAndAdvancesOnMaterialChange_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-i-2-04")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-04-create",
		"incident_key":  "IR-I204",
		"title":         "Patchable Incident",
		"description":   "Description stays fixed",
		"severity":      "high",
	})
	incidentID := incident["incident_id"].(string)
	initialUpdatedAt := incident["updated_at"]

	noOpPatch := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	noOpBody := httptestx.RequireSuccessEnvelope(t, noOpPatch, http.StatusOK)["data"].(map[string]any)
	if noOpBody["incident_version"] != float64(1) || noOpBody["updated_at"] != initialUpdatedAt {
		t.Fatalf("unexpected no-op patch payload: %#v", noOpBody)
	}

	time.Sleep(20 * time.Millisecond)
	patch := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version":     1,
			"tlp":                       "TLP:AMBER",
			"current_phase":             "containment",
			"primary_external_case_ref": "CASE-I204",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patch, http.StatusOK)["data"].(map[string]any)
	if patchBody["incident_version"] != float64(2) || patchBody["tlp"] != "TLP:AMBER" || patchBody["current_phase"] != "containment" {
		t.Fatalf("unexpected incident patch payload: %#v", patchBody)
	}
	if patchBody["updated_at"] == initialUpdatedAt {
		t.Fatalf("expected material patch to advance updated_at: before=%v after=%v", initialUpdatedAt, patchBody["updated_at"])
	}

	stale := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:GREEN",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	staleBody := httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "incident_version_conflict")
	staleDetails := staleBody["error"].(map[string]any)["details"].(map[string]any)
	if staleDetails["incident_id"] != incidentID ||
		staleDetails["base_incident_version"] != float64(1) ||
		staleDetails["current_incident_version"] != float64(2) {
		t.Fatalf("unexpected incident version conflict details: %#v", staleDetails)
	}

	row := harness.DB.QueryRow(`
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
	if tlp == nil || *tlp != "TLP:AMBER" || currentPhase == nil || *currentPhase != "containment" || primaryExternalCaseRef == nil || *primaryExternalCaseRef != "CASE-I204" || incidentVersion != 2 {
		t.Fatalf("patch must persist promoted fields only: tlp=%v current_phase=%v primary_external_case_ref=%v incident_version=%d", tlp, currentPhase, primaryExternalCaseRef, incidentVersion)
	}
}

func TestMembershipPatchSameRoleReturnsOKWithoutVersionOrMutationArtifact_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-i-2-07")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	targetUserID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase2-i207-target@example.test", "Phase 2 I207 Target", "Phase2I207TargetPass!", false, false, true)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-i-2-07-create",
		"incident_key":  "IR-I207",
		"title":         "Membership Same Role",
	})
	incidentID := incident["incident_id"].(string)
	membership := scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-i-2-07-membership",
		"user_id":       targetUserID,
		"role":          "viewer",
	})
	baseVersion := membership["membership_version"]
	baseVersionFloat, ok := baseVersion.(float64)
	if !ok {
		t.Fatalf("unexpected membership_version type in create payload: %#v", membership)
	}
	beforeMutationArtifacts := mutationtest.CountMutationArtifacts(
		t, mutationtest.SQLDatabase(

			harness.DB),

		mutationtest.MutationSelector{IncidentID: incidentID},
		mutationtest.MutationOwnerIncidentMembership)

	noOp := scenariotest.PatchMembership(t, harness.Server, adminLogin, incidentID, targetUserID, map[string]any{
		"base_membership_version": baseVersion,
		"role":                    "viewer",
	})
	if noOp["role"] != "viewer" || noOp["membership_version"] != baseVersion {
		t.Fatalf("same-role membership patch must return unchanged role and version: before=%#v after=%#v", membership, noOp)
	}

	var (
		durableRole    string
		durableVersion int64
	)
	if err := harness.DB.QueryRow(`
SELECT role, membership_version
  FROM incident_memberships
 WHERE incident_id::text = $1
   AND user_id::text = $2
`, incidentID, targetUserID).Scan(&durableRole, &durableVersion); err != nil {
		t.Fatalf("query membership after same-role patch: %v", err)
	}
	if durableRole != "viewer" || float64(durableVersion) != baseVersionFloat {
		t.Fatalf("same-role membership patch must keep durable role and version stable: role=%q version=%d before=%#v", durableRole, durableVersion, membership)
	}

	afterMutationArtifacts := mutationtest.CountMutationArtifacts(
		t, mutationtest.SQLDatabase(

			harness.DB),

		mutationtest.MutationSelector{IncidentID: incidentID},
		mutationtest.MutationOwnerIncidentMembership)

	if afterMutationArtifacts != beforeMutationArtifacts {
		t.Fatalf("same-role membership patch must not write membership mutation artifacts: before=%d after=%d", beforeMutationArtifacts, afterMutationArtifacts)
	}
}

func TestExtensionDiscoveryReturnsExactZeroMembershipShapeWithoutLeaks_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-i-2-05")

	_, _ = flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	userID := flowtest.SeedLocalUserFlags(t, harness.DB, "extension-user@example.test", "Extension User", "ExtensionUser1!", false, false, true)
	userSession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "extension-user@example.test", "ExtensionUser1!", nil)

	extensionsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/extensions", nil, httptestx.WithCookies(userSession))
	extensionsBody := httptestx.RequireSuccessEnvelope(t, extensionsResp, http.StatusOK)
	data := extensionsBody["data"].(map[string]any)
	extensions := data["extensions"].([]any)
	wantProfiles := contracttest.CurrentProfileExtensions(t)
	if len(extensions) != len(wantProfiles) {
		t.Fatalf("unexpected extensions payload: got %d want %d", len(extensions), len(wantProfiles))
	}
	for index, want := range wantProfiles {
		raw := extensions[index]
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
		wantClaimed := want.ProfileID == "import" || want.ProfileID == "incident_portability" || want.ProfileID == "reference_pack" || want.ProfileID == "snapshot_reporting"
		if item["profile_id"] != want.ProfileID || item["claimed"] != wantClaimed {
			t.Fatalf("unexpected extension item %d: %#v", index, item)
		}
		if gotFamilies := OrderedRouteFamilies(t, item["route_families"]); strings.Join(gotFamilies, ",") != strings.Join(want.RouteFamilies, ",") {
			t.Fatalf("unexpected route families for %s: got %v want %v", want.ProfileID, gotFamilies, want.RouteFamilies)
		}
	}
	if got := OrderedProfileIDs(t, extensions); strings.Join(got, ",") != strings.Join(ContractProfileIDs(wantProfiles), ",") {
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
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE user_id::text = $1`, userID); got != 0 {
		t.Fatalf("zero-membership user must stay outside incident scope, got %d memberships", got)
	}
}

func TestUnclaimedReservedFamiliesReturnCanonical404AndOutsidePathsDoNot_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-i-2-06")

	_, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	flowtest.SeedLocalUserFlags(t, harness.DB, "reserved-user@example.test", "Reserved User", "ReservedUser1!", false, false, true)
	userSession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "reserved-user@example.test", "ReservedUser1!", nil)
	enterpriseProfile := ExtensionContract(t, "enterprise_authentication")
	networkFlowProfile := ExtensionContract(t, "network_flow_activity")

	rootReserved := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+enterpriseProfile.RouteFamilies[0], nil, httptestx.WithCookies(userSession))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != enterpriseProfile.ProfileID || rootDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved root dispatch details: %#v", rootDetails)
	}

	networkFlowReserved := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+strings.Replace(networkFlowProfile.RouteFamilies[0], "{incident_id}", adminID, 1),
		nil,
		httptestx.WithCookies(userSession),
	)
	networkFlowBody := httptestx.RequireErrorEnvelope(t, networkFlowReserved, http.StatusNotFound, "extension_profile_not_claimed")
	networkFlowDetails := networkFlowBody["error"].(map[string]any)["details"].(map[string]any)
	if networkFlowDetails["profile_id"] != networkFlowProfile.ProfileID || networkFlowDetails["route_family"] != networkFlowProfile.RouteFamilies[0] {
		t.Fatalf("unexpected Network Flow reserved root dispatch details: %#v", networkFlowDetails)
	}

	descendantReserved := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+strings.Replace(enterpriseProfile.RouteFamilies[0], "{user_id}", adminID, 1)+"/provider",
		nil,
		httptestx.WithCookies(userSession),
	)
	descendantBody := httptestx.RequireErrorEnvelope(t, descendantReserved, http.StatusNotFound, "extension_profile_not_claimed")
	descendantDetails := descendantBody["error"].(map[string]any)["details"].(map[string]any)
	if descendantDetails["profile_id"] != enterpriseProfile.ProfileID || descendantDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved descendant dispatch details: %#v", descendantDetails)
	}

	outsideReserved := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/outside-reserved-families", nil, httptestx.WithCookies(userSession))
	if outsideReserved.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected outside reserved status: got %d", outsideReserved.StatusCode)
	}
	bodyText := httptestx.ReadBodyString(t, outsideReserved.Body)
	if strings.Contains(bodyText, "extension_profile_not_claimed") {
		t.Fatalf("outside reserved families must keep ordinary 404 handling, got %q", bodyText)
	}
}

func OrderedProfileIDs(t testing.TB, extensions []any) []string {
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

func OrderedRouteFamilies(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("unexpected route_families payload: %T", raw)
	}
	families := make([]string, 0, len(items))
	for _, item := range items {
		family, ok := item.(string)
		if !ok {
			t.Fatalf("unexpected route_family item: %T", item)
		}
		families = append(families, family)
	}
	return families
}

func ContractProfileIDs(profiles []contracttest.ExtensionProfileContract) []string {
	ordered := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		ordered = append(ordered, profile.ProfileID)
	}
	return ordered
}

func ExtensionContract(t testing.TB, profileID string) contracttest.ExtensionProfileContract {
	t.Helper()
	for _, profile := range contracttest.CurrentProfileExtensions(t) {
		if profile.ProfileID == profileID {
			return profile
		}
	}
	t.Fatalf("missing extension contract for %q", profileID)
	return contracttest.ExtensionProfileContract{}
}
