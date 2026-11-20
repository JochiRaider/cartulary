package indicators_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// TestIndicatorRoutesReplayFirstAndConcealVisibility_Integration verifies the
// REQ-01-704 and REQ-01-705 order on the Indicator routes.
//
//   - On every route, a caller who cannot see the incident receives the route's
//     own not-found response, identical in status, code and details to the
//     response for a missing record or observation, and commits nothing.
//   - On every mutation route, an exact committed replay returns its original
//     receipt after the addressed Indicator and source record are deleted and
//     after the incident closes, and commits nothing; a fresh request on the
//     deleted Indicator is still 409 record_deleted_use_restore.
func TestIndicatorRoutesReplayFirstAndConcealVisibility_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "indicator-replay-concealment", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-indicator-replay-concealment-incident",
		"incident_key":  "IR-IND-REPLAY-CONCEALMENT",
		"title":         "Indicator replay and concealment",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	createIndicator := func(txn, domain string) uuid.UUID {
		t.Helper()
		row := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost,
			"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":            txn,
				"indicator.indicator_type": "domain_name",
				"indicator.value_kind":     "atomic",
				"indicator.display_value":  domain,
			},
		), http.StatusCreated)["data"].(map[string]any)["row"].(map[string]any)
		return appsupport.MustUUID(t, row["record_id"].(string))
	}
	indicatorID := createIndicator("txn-replay-indicator", "replay-target.example.test")
	sourceText := "replay-source.example"
	sourceID := uuid.New()
	timelinetest.SeedTimelineRecordWithSourceText(t, harness.DB, incidentID, actorID, sourceID, sourceText)

	lifecyclePath := func(id uuid.UUID) string { return "/api/v1/indicators/" + id.String() + "/state-intervals" }
	observationsByIndicator := func(id uuid.UUID) string { return "/api/v1/indicators/" + id.String() + "/observations" }
	observationsBySource := func(id uuid.UUID) string { return "/api/v1/records/" + id.String() + "/indicator-observations" }
	observationAction := func(id uuid.UUID, action string) string {
		return "/api/v1/indicator-observations/" + id.String() + "/" + action
	}
	lifecycleBody := func(txn string, base int) map[string]any {
		return map[string]any{
			"client_txn_id": txn, "base_row_version": base,
			"lifecycle_state": "false_positive", "valid_from": "2026-11-18T12:00:00Z",
			"valid_to": nil, "confidence": 60, "rationale": "replay fixture", "support_refs": []any{}, "assessor": "reviewer",
		}
	}
	captureBody := func(txn string, base int, resolved uuid.UUID) map[string]any {
		return map[string]any{
			"client_txn_id": txn, "base_row_version": base, "source_field_key": timelinetest.FieldSourceText,
			"span_start_byte": 0, "span_end_byte": len(sourceText), "resolved_indicator_record_id": resolved.String(),
		}
	}
	snapshot := func() workbookscenariotest.MutationEffects {
		t.Helper()
		return workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
	}

	// Commit one request on every Indicator mutation route. Each commit
	// advances the Indicator, and each observation commit the source too.
	type committedRequest struct {
		name, path string
		body       map[string]any
		receipt    map[string]any
	}
	commit := func(name, path string, body map[string]any, status int) committedRequest {
		t.Helper()
		receipt := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, path, body), status)["data"].(map[string]any)
		return committedRequest{name: name, path: path, body: body, receipt: receipt}
	}
	committed := []committedRequest{commit("lifecycle append", lifecyclePath(indicatorID), lifecycleBody("txn-replay-append", 1), http.StatusCreated)}
	capture := commit("observation capture", observationsBySource(sourceID), captureBody("txn-replay-capture", 1, indicatorID), http.StatusCreated)
	committed = append(committed, capture)
	observationID := appsupport.MustUUID(t, capture.receipt["observation"].(map[string]any)["observation_id"].(string))
	committed = append(committed,
		commit("observation dismiss", observationAction(observationID, "dismiss"), map[string]any{"client_txn_id": "txn-replay-dismiss", "base_row_version": 1}, http.StatusOK),
		commit("observation restore", observationAction(observationID, "restore"), map[string]any{"client_txn_id": "txn-replay-restore", "base_row_version": 2}, http.StatusOK),
		commit("observation resolve", observationAction(observationID, "resolve"), map[string]any{
			"client_txn_id": "txn-replay-resolve", "base_row_version": 3, "resolved_indicator_record_id": indicatorID.String(),
		}, http.StatusOK),
	)

	// REQ-01-704 on every Indicator route: the outsider addresses live
	// targets, the member addresses missing ones, and the two responses match.
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "indicator-concealment-outsider@example.test", "Outsider", "OutsiderPass1!", false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, outsider.Email, "OutsiderPass1!", nil)
	missingID := uuid.New()
	for _, route := range []struct {
		name, method, code string
		path               func(uuid.UUID) string
		live               uuid.UUID
		body               map[string]any
	}{
		{"listSourceRecordIndicatorObservations", http.MethodGet, "indicator_source_record_not_found", observationsBySource, sourceID, nil},
		{"createManualIndicatorObservation", http.MethodPost, "indicator_source_record_not_found", observationsBySource, sourceID, captureBody("txn-concealed-capture", 2, indicatorID)},
		{"listIndicatorObservations", http.MethodGet, "indicator_not_found", observationsByIndicator, indicatorID, nil},
		{"listIndicatorStateIntervals", http.MethodGet, "indicator_not_found", lifecyclePath, indicatorID, nil},
		{"appendIndicatorStateInterval", http.MethodPost, "indicator_not_found", lifecyclePath, indicatorID, lifecycleBody("txn-concealed-append", 6)},
		{"resolveIndicatorObservation", http.MethodPost, "indicator_observation_not_found", func(id uuid.UUID) string { return observationAction(id, "resolve") }, observationID,
			map[string]any{"client_txn_id": "txn-concealed-resolve", "base_row_version": 4, "resolved_indicator_record_id": indicatorID.String()}},
		{"dismissIndicatorObservation", http.MethodPost, "indicator_observation_not_found", func(id uuid.UUID) string { return observationAction(id, "dismiss") }, observationID,
			map[string]any{"client_txn_id": "txn-concealed-dismiss", "base_row_version": 4}},
		{"restoreIndicatorObservation", http.MethodPost, "indicator_observation_not_found", func(id uuid.UUID) string { return observationAction(id, "restore") }, observationID,
			map[string]any{"client_txn_id": "txn-concealed-restore", "base_row_version": 4}},
	} {
		request := func(caller authcookietest.AuthCookies, id uuid.UUID) map[string]any {
			t.Helper()
			before := snapshot()
			response := childRouteJSON(t, harness, caller, route.method, route.path(id), route.body)
			if response.StatusCode != http.StatusNotFound {
				t.Fatalf("%s: status %d, want 404 %s; body=%#v", route.name, response.StatusCode, route.code, httptestx.ReadJSONBody(t, response))
			}
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, route.code)
			if after := snapshot(); after != before {
				t.Fatalf("%s committed effects: %+v -> %+v", route.name, before, after)
			}
			// Everything but the per-request identifier must match.
			errorObject := body["error"].(map[string]any)
			delete(errorObject, "request_id")
			return errorObject
		}
		concealed, missing := request(outsiderLogin, route.live), request(login, missingID)
		if !reflect.DeepEqual(concealed, missing) {
			t.Fatalf("%s: non-member response %#v differs from the missing-target response %#v", route.name, concealed, missing)
		}
	}

	// REQ-01-705: an exact committed replay precedes the deleted state and the
	// incident lifecycle on every mutation route.
	requireReplays := func(stage string) {
		t.Helper()
		before := snapshot()
		for _, request := range committed {
			got := httptestx.RequireSuccessEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, request.path, request.body), http.StatusOK)["data"].(map[string]any)
			if got["replayed"] != true {
				t.Fatalf("%s %s: response is not a replay: %#v", stage, request.name, got)
			}
			original := make(map[string]any, len(request.receipt))
			for key, value := range request.receipt {
				original[key] = value
			}
			original["replayed"] = true
			if !reflect.DeepEqual(got, original) {
				t.Fatalf("%s %s: replay = %#v, want the original receipt %#v", stage, request.name, got, original)
			}
		}
		if after := snapshot(); after != before {
			t.Fatalf("%s replays committed effects: %+v -> %+v", stage, before, after)
		}
	}
	envelopetest.SoftDeleteRecord(t, harness.DB, envelopetest.Tombstone{RecordID: indicatorID, DeletedBy: actorID})
	envelopetest.SoftDeleteRecord(t, harness.DB, envelopetest.Tombstone{RecordID: sourceID, DeletedBy: actorID})
	requireReplays("after deletion")
	before := snapshot()
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, lifecyclePath(indicatorID), lifecycleBody("txn-replay-fresh-append", 6)), http.StatusConflict, "record_deleted_use_restore")
	httptestx.RequireErrorEnvelope(t, childRouteJSON(t, harness, login, http.MethodPost, observationsBySource(sourceID), captureBody("txn-replay-fresh-capture", 5, indicatorID)), http.StatusConflict, "record_deleted_use_restore")
	if after := snapshot(); after != before {
		t.Fatalf("fresh requests on deleted targets committed effects: %+v -> %+v", before, after)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}
	requireReplays("after deletion and closure")
}
