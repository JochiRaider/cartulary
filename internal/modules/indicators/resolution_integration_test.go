package indicators_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// indicator-resolution / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestIndicatorsRoute_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-i-4-07-indicators")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-entity_linking-i-4-07-incident",
		"incident_key":  "IR-I407",
		"title":         "Record relationships indicator route",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	example := indicatortest.PrimaryExample()
	payload := map[string]any{
		"client_txn_id":              "txn-entity_linking-i-4-07-create",
		"indicator.indicator_type":   example.IndicatorType,
		"indicator.value_kind":       example.ValueKind,
		"indicator.display_value":    example.DisplayValue,
		"indicator.normalized_value": example.NormalizedValue,
	}
	response := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
		payload,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	data := appsupport.RequireSuccessData(t, response, http.StatusCreated)
	row := data["row"].(map[string]any)
	if row["record_id"] == "" || row["row_version"] == nil {
		t.Fatalf("indicator create row is incomplete: %#v", row)
	}
	replayResponse := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
		payload,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayData := appsupport.RequireSuccessData(t, replayResponse, http.StatusOK)
	if !reflect.DeepEqual(replayData, data) {
		t.Fatalf("exact Indicator replay changed its committed result: first=%#v replay=%#v", data, replayData)
	}
	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + viewtest.IndicatorsViewSchemaID + "/rows"
	post := func(body map[string]any, status int) map[string]any {
		t.Helper()
		return appsupport.RequireSuccessData(t, appsupport.DoJSON(t, http.MethodPost, createURL, body,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), status)
	}
	normalizedPayload := map[string]any{}
	for key, value := range payload {
		normalizedPayload[key] = value
	}
	normalizedPayload["indicator.display_value"] = "203[.]0[.]113[.]24"
	delete(normalizedPayload, "indicator.normalized_value")
	if got := post(normalizedPayload, http.StatusOK); !reflect.DeepEqual(got, data) {
		t.Fatalf("normalized replay differs: %#v", got)
	}
	normalizedPayload["client_txn_id"] = "txn-fresh-canonical-match"
	normalizedPayload["indicator.stix_pattern"] = "[ipv4-addr:value = '203.0.113.24']"
	normalizedPayload["indicator.defanged_value"] = "203[.]0[.]113[.]24"
	reuseData := post(normalizedPayload, http.StatusCreated)
	if !reflect.DeepEqual(reuseData["row"], row) || reuseData["change_set_id"] == data["change_set_id"] {
		t.Fatalf("reuse changed metadata/version or lost receipt: %#v", reuseData)
	}
	if got := post(normalizedPayload, http.StatusOK); !reflect.DeepEqual(got, reuseData) {
		t.Fatalf("reuse replay differs: %#v", got)
	}
	var unchangedHistory, liveRevisions int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1 AND before_value = after_value AND before_version_id = after_version_id`, reuseData["change_set_id"]).Scan(&unchangedHistory); err != nil {
		t.Fatal(err)
	}
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM record_revisions WHERE record_id = $1`, row["record_id"]).Scan(&liveRevisions); err != nil {
		t.Fatal(err)
	}
	publications := collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{SourceRecordID: row["record_id"].(string)})
	if unchangedHistory != 1 || liveRevisions != 1 || publications != 1 {
		t.Fatalf("reuse history/revision/publication counts: %d/%d/%d", unchangedHistory, liveRevisions, publications)
	}
	divergentPayload := map[string]any{}
	for key, value := range payload {
		divergentPayload[key] = value
	}
	divergentPayload["indicator.display_value"] = "203.0.113.89"
	divergentPayload["indicator.normalized_value"] = "203.0.113.89"
	divergentResponse := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
		divergentPayload,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergentResponse, http.StatusConflict, "client_txn_conflict")
	changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), data["change_set_id"].(string))
	if changeSet.ActorUserID != adminUserID.String() || changeSet.Source != "indicators.rows.create" {
		t.Fatalf("indicator mutation attribution mismatch: %#v", changeSet)
	}
	login := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
	rows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login)
	queried := workbookscenariotest.FindRow(t, rows, row["record_id"].(string))
	if queried["record_id"] != row["record_id"] {
		t.Fatalf("indicator route readback mismatch: %#v", queried)
	}
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	application := newIndicatorTestApplication(t, harness.Pool, harness.Revisions.Appender())
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.SiblingRecordID)
	for index, sourceRecordID := range []struct {
		id    string
		field string
	}{
		{id: timelinetest.RecordID.String(), field: timelinetest.FieldSourceText},
		{id: timelinetest.SiblingRecordID.String(), field: timelinetest.FieldSummary},
	} {
		sourceID := appsupport.MustUUID(t, sourceRecordID.id)
		if _, err := application.CreateIndicatorObservation(context.Background(), adminUserID, manualObservationParams(
			incidentID, sourceID, sourceRecordID.field, &recordID,
			"txn-entity-linking-route-observation-"+string(rune('1'+index)),
		)); err != nil {
			t.Fatalf("create observation %d: %v", index, err)
		}
	}
	if _, err := application.AppendIndicatorLifecycleInterval(context.Background(), adminUserID, lifecycleAppendParams(
		incidentID, recordID, 3, indicatortest.PastTime(), "txn-entity-linking-route-lifecycle",
	)); err != nil {
		t.Fatalf("append lifecycle: %v", err)
	}
	rowsBeforeRebuild := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login)
	rowBeforeRebuild := workbookscenariotest.FindRow(t, rowsBeforeRebuild, recordID.String())
	cells := rowBeforeRebuild["cells"].(map[string]any)
	if cells["indicator.observation_count"].(map[string]any)["value"] != float64(2) || cells["indicator.lifecycle_summary"].(map[string]any)["value"] != "active" {
		t.Fatalf("indicator projection readback mismatch: %#v", cells)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		t.Fatalf("clear indicator projections: %v", err)
	}
	if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild indicator projections: %v", err)
	}
	rowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login), recordID.String())
	if !reflect.DeepEqual(rowBeforeRebuild["cells"], rowAfterRebuild["cells"]) {
		t.Fatalf("indicator projection rebuild drifted: before=%#v after=%#v", rowBeforeRebuild, rowAfterRebuild)
	}
	// Current transport authorization precedes replay, including after closure.
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, createURL, payload), http.StatusUnauthorized, "session_required")
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, createURL, payload, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incidentID); err != nil {
		t.Fatal(err)
	}
	if got := post(payload, http.StatusOK); !reflect.DeepEqual(got, data) {
		t.Fatalf("closed-incident replay replaced the historical row: %#v", got)
	}
	normalizedPayload["client_txn_id"] = "txn-closed-fresh-create"
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, createURL, normalizedPayload, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusConflict, "incident_closed")

	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role = 'viewer', membership_version = membership_version + 1, updated_at = now(), updated_by_user_id = $2 WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, createURL, payload, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusForbidden, "authorization_denied")
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, createURL, payload, appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "incident_not_found")

}
