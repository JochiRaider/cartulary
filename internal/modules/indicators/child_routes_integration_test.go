package indicators_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIndicatorProductionChildRoutes_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "indicator-production-child-routes")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-indicator-child-routes-incident",
		"incident_key":  "IR-IND-CHILD-ROUTES",
		"title":         "Indicator child route acceptance",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	example := indicatortest.PrimaryExample()

	indicatorResponse := childRouteJSON(t, harness, login, http.MethodPost,
		"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":              "txn-indicator-child-routes-indicator",
			"indicator.indicator_type":   example.IndicatorType,
			"indicator.value_kind":       example.ValueKind,
			"indicator.display_value":    example.DisplayValue,
			"indicator.normalized_value": example.NormalizedValue,
		},
	)
	indicatorData := httptestx.RequireSuccessEnvelope(t, indicatorResponse, http.StatusCreated)["data"].(map[string]any)
	indicatorRow := indicatorData["row"].(map[string]any)
	indicatorID := appsupport.MustUUID(t, indicatorRow["record_id"].(string))

	sourceID := uuid.New()
	sourceText := "  α.example  "
	timelinetest.SeedTimelineRecordWithSourceText(t, harness.DB, incidentID, actorID, sourceID, sourceText)
	observationPath := "/api/v1/records/" + sourceID.String() + "/indicator-observations"
	createBody := map[string]any{
		"client_txn_id":                "txn-indicator-child-routes-observation",
		"base_row_version":             1,
		"source_field_key":             timelinetest.FieldSourceText,
		"span_start_byte":              0,
		"span_end_byte":                len(sourceText),
		"parsed_indicator_type":        "text",
		"resolved_indicator_record_id": indicatorID.String(),
	}
	createdEnvelope := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationPath, createBody), http.StatusCreated)
	created := createdEnvelope["data"].(map[string]any)
	observation := created["observation"].(map[string]any)
	_ = appsupport.MustUUID(t, observation["observation_id"].(string))
	if observation["observed_text"] != sourceText || observation["origin_kind"] != "manual_entry" {
		t.Fatalf("server-derived observation provenance = %#v", observation)
	}
	requireSortedAffectedVersions(t, created["affected_records"].([]any), sourceID, indicatorID)

	replayed := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationPath, createBody), http.StatusOK)["data"].(map[string]any)
	if replayed["replayed"] != true || replayed["change_set_id"] != created["change_set_id"] {
		t.Fatalf("exact observation replay = %#v", replayed)
	}
	divergent := cloneMap(createBody)
	divergent["span_end_byte"] = len(sourceText) - 1
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationPath, divergent), http.StatusConflict, "client_txn_conflict")
	forbiddenProvenance := cloneMap(createBody)
	forbiddenProvenance["client_txn_id"] = "txn-indicator-child-routes-forbidden-provenance"
	forbiddenProvenance["origin_kind"] = "system"
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationPath, forbiddenProvenance), http.StatusBadRequest, "invalid_mutation_payload")

	secondBody := map[string]any{
		"client_txn_id":    "txn-indicator-child-routes-observation-two",
		"base_row_version": 2,
		"source_field_key": timelinetest.FieldSummary,
		"span_start_byte":  0,
		"span_end_byte":    len(sourceText),
	}
	second := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationPath, secondBody), http.StatusCreated)["data"].(map[string]any)
	secondObservation := second["observation"].(map[string]any)
	secondID := appsupport.MustUUID(t, secondObservation["observation_id"].(string))

	pageOne := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodGet, observationPath+"?limit=1", nil), http.StatusOK)
	pageOneRows := pageOne["data"].(map[string]any)["observations"].([]any)
	pageOneMeta := pageOne["meta"].(map[string]any)["paging"].(map[string]any)
	if len(pageOneRows) != 1 || pageOneMeta["has_more"] != true || pageOneMeta["next_cursor"] == nil {
		t.Fatalf("first observation keyset page = %#v", pageOne)
	}
	cursor := pageOneMeta["next_cursor"].(string)
	pageTwo := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodGet, observationPath+"?limit=1&cursor_token="+cursor, nil), http.StatusOK)
	if len(pageTwo["data"].(map[string]any)["observations"].([]any)) != 1 {
		t.Fatalf("second observation keyset page = %#v", pageTwo)
	}

	dismissPath := "/api/v1/indicator-observations/" + secondID.String() + "/dismiss"
	dismissBody := map[string]any{"client_txn_id": "txn-indicator-child-routes-dismiss", "base_row_version": 1}
	dismissed := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, dismissPath, dismissBody), http.StatusOK)["data"].(map[string]any)
	if dismissed["observation"].(map[string]any)["resolution_status"] != "dismissed" {
		t.Fatalf("dismissed observation = %#v", dismissed)
	}
	dismissReplay := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, dismissPath, dismissBody), http.StatusOK)["data"].(map[string]any)
	if dismissReplay["replayed"] != true || dismissReplay["change_set_id"] != dismissed["change_set_id"] {
		t.Fatalf("dismiss replay = %#v", dismissReplay)
	}
	dismissDivergent := cloneMap(dismissBody)
	dismissDivergent["base_row_version"] = 2
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, dismissPath, dismissDivergent), http.StatusConflict, "client_txn_conflict")
	illegalDismiss := map[string]any{"client_txn_id": "txn-indicator-child-routes-dismiss-again", "base_row_version": 2}
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, dismissPath, illegalDismiss), http.StatusConflict, "illegal_transition")

	restorePath := "/api/v1/indicator-observations/" + secondID.String() + "/restore"
	restored := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, restorePath, map[string]any{
		"client_txn_id": "txn-indicator-child-routes-restore", "base_row_version": 2,
	}), http.StatusOK)["data"].(map[string]any)
	if restored["observation"].(map[string]any)["resolution_status"] != "unresolved" {
		t.Fatalf("restored observation = %#v", restored)
	}
	restoreReplayBody := map[string]any{"client_txn_id": "txn-indicator-child-routes-restore", "base_row_version": 2}
	restoreReplay := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, restorePath, restoreReplayBody), http.StatusOK)["data"].(map[string]any)
	if restoreReplay["replayed"] != true || restoreReplay["change_set_id"] != restored["change_set_id"] {
		t.Fatalf("restore replay = %#v", restoreReplay)
	}

	resolvePath := "/api/v1/indicator-observations/" + secondID.String() + "/resolve"
	resolved := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, resolvePath, map[string]any{
		"client_txn_id": "txn-indicator-child-routes-resolve", "base_row_version": 3, "resolved_indicator_record_id": indicatorID.String(),
	}), http.StatusOK)["data"].(map[string]any)
	if resolved["observation"].(map[string]any)["resolution_status"] != "resolved" {
		t.Fatalf("resolved observation = %#v", resolved)
	}
	resolveReplayBody := map[string]any{
		"client_txn_id": "txn-indicator-child-routes-resolve", "base_row_version": 3, "resolved_indicator_record_id": indicatorID.String(),
	}
	resolveReplay := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, resolvePath, resolveReplayBody), http.StatusOK)["data"].(map[string]any)
	if resolveReplay["replayed"] != true || resolveReplay["change_set_id"] != resolved["change_set_id"] {
		t.Fatalf("resolve replay = %#v", resolveReplay)
	}
	resolveDivergent := cloneMap(resolveReplayBody)
	resolveDivergent["resolved_indicator_record_id"] = uuid.New().String()
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, resolvePath, resolveDivergent), http.StatusConflict, "client_txn_conflict")
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, resolvePath, map[string]any{
		"client_txn_id": "txn-indicator-child-routes-resolve-again", "base_row_version": 4, "resolved_indicator_record_id": indicatorID.String(),
	}), http.StatusConflict, "illegal_transition")

	indicatorObservations := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodGet,
		"/api/v1/indicators/"+indicatorID.String()+"/observations", nil), http.StatusOK)["data"].(map[string]any)["observations"].([]any)
	if len(indicatorObservations) != 2 {
		t.Fatalf("Indicator observation list = %#v", indicatorObservations)
	}

	lifecyclePath := "/api/v1/indicators/" + indicatorID.String() + "/state-intervals"
	lifecycleBody := map[string]any{
		"client_txn_id": "txn-indicator-child-routes-lifecycle", "base_row_version": 3,
		"lifecycle_state": "false_positive", "valid_from": "2026-08-03T20:00:00Z",
		"valid_to": nil, "confidence": 80, "rationale": "validated in route test", "support_refs": []any{sourceID.String()}, "assessor": "incident reviewer",
	}
	lifecycle := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, lifecyclePath, lifecycleBody), http.StatusCreated)["data"].(map[string]any)
	if lifecycle["interval"].(map[string]any)["lifecycle_state"] != "false_positive" {
		t.Fatalf("lifecycle append = %#v", lifecycle)
	}
	lifecycleReplay := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, lifecyclePath, lifecycleBody), http.StatusOK)["data"].(map[string]any)
	if lifecycleReplay["replayed"] != true || lifecycleReplay["change_set_id"] != lifecycle["change_set_id"] {
		t.Fatalf("lifecycle replay = %#v", lifecycleReplay)
	}
	lifecycleDivergent := cloneMap(lifecycleBody)
	lifecycleDivergent["confidence"] = 79
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, lifecyclePath, lifecycleDivergent), http.StatusConflict, "client_txn_conflict")
	intervals := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodGet, lifecyclePath, nil), http.StatusOK)["data"].(map[string]any)["intervals"].([]any)
	if len(intervals) != 1 {
		t.Fatalf("lifecycle list = %#v", intervals)
	}

	var observationCount int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM indicator_observations WHERE incident_id = $1 AND deleted_at IS NULL`, incidentID).Scan(&observationCount); err != nil || observationCount != 2 {
		t.Fatalf("committed observation count = %d, %v", observationCount, err)
	}
	collaborationIntentCount := collaborationsupport.CountIntentsForChangeSetSources(
		t,
		harness.DB,
		incidentID.String(),
		"indicators.observations.capture",
		"indicators.observations.resolve",
		"indicators.observations.dismiss",
		"indicators.observations.restore",
		"indicators.lifecycle.append",
	)
	if collaborationIntentCount != 8 {
		t.Fatalf("Indicator child collaboration intents = %d", collaborationIntentCount)
	}
}

func childRouteJSON(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, method string, path string, body any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(t, method, harness.Server.HTTP.URL+path, body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func requireSortedAffectedVersions(t testing.TB, rows []any, want ...uuid.UUID) {
	t.Helper()
	for index := 1; index < len(want); index++ {
		if want[index].String() < want[index-1].String() {
			want[index], want[index-1] = want[index-1], want[index]
		}
	}
	if len(rows) != len(want) {
		t.Fatalf("affected records = %#v, want %v", rows, want)
	}
	for index, expected := range want {
		row := rows[index].(map[string]any)
		if row["record_id"] != expected.String() || row["row_version"].(float64) < 2 {
			t.Fatalf("affected record %d = %#v, want %s", index, row, expected)
		}
	}
}

func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
