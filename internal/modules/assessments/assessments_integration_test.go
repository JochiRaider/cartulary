package assessments_test

import (
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestAssessmentsCreateAndProjection(t *testing.T) {
	harness := appsupport.StartServer(t, "assessments-create")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-assessments-create-incident",
		"incident_key":  "IR-ASSESS-CREATE",
		"title":         "Assessments create",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	hostID := uuid.New()
	supportID := uuid.New()
	supportHostID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, hostID, "Assessment host", "assess-host", "", "")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, supportHostID, "Assessment support host", "assess-support-host", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, supportID)

	intentsBeforeCreate := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id = $1`, incidentID)
	body := map[string]any{
		"client_txn_id":               "txn-assessments-create",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "  Confirmed from supporting event.  ",
		"assessment.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
				{"op": "add_record_ref", "linked_record_id": supportHostID.String()},
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
			},
		},
	}
	create := postAssessment(t, harness, adminLogin, incidentID, body)
	data := appsupport.RequireSuccessData(t, create, http.StatusCreated)
	row := data["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	cells := row["cells"].(map[string]any)

	requireCellValue(t, cells, "assessment.subject_ref", hostID.String())
	requireCellValue(t, cells, "assessment.assessment_state", "confirmed")
	requireCellValue(t, cells, "assessment.confidence_band", "unset")
	requireCellValue(t, cells, "assessment.confidence_score", nil)
	requireCellValue(t, cells, "assessment.assessor", adminUserID.String())
	requireCellValue(t, cells, "assessment.supporting_link_count", float64(2))
	if got := requireCellValue(t, cells, "assessment.rationale", "Confirmed from supporting event."); got != "Confirmed from supporting event." {
		t.Fatalf("unexpected rationale normalization: %#v", got)
	}
	items := workbookscenariotest.CollectionItems(t, row, "assessment.support_refs")
	if len(items) != 2 ||
		items[0]["linked_record_id"] != supportHostID.String() ||
		items[1]["linked_record_id"] != supportID.String() {
		t.Fatalf("unexpected support refs: %#v", items)
	}

	link := linktest.LookupActiveLink(t, harness.DB, incidentID, recordID, supportID, "supported_by")
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("unexpected support link metadata: %#v", link)
	}
	hostLink := linktest.LookupActiveLink(t, harness.DB, incidentID, recordID, supportHostID, "supported_by")
	if hostLink.Provenance != "manual" || hostLink.Confidence != nil {
		t.Fatalf("unexpected heterogeneous support link metadata: %#v", hostLink)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND record_type = 'assessment'`, recordID); got != 1 {
		t.Fatalf("expected one assessment record envelope, got %d", got)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id = $1`, incidentID); got != intentsBeforeCreate+1 {
		t.Fatalf("assessment create publication intents = %d, want %d", got, intentsBeforeCreate+1)
	}
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, data["change_set_id"].(string))

	replayBody := map[string]any{
		"client_txn_id":               "txn-assessments-create",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "  host  ",
		"assessment.assessment_state": " confirmed ",
		"assessment.rationale":        "Confirmed from supporting event.",
		"assessment.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportHostID.String()},
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
			},
		},
	}
	replay := postAssessment(t, harness, adminLogin, incidentID, replayBody)
	replayData := appsupport.RequireSuccessData(t, replay, http.StatusOK)
	if got := replayData["row"].(map[string]any)["record_id"]; got != recordID.String() {
		t.Fatalf("expected idempotent replay of record %s, got %#v", recordID, replayData)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id = $1`, incidentID); got != intentsBeforeCreate+1 {
		t.Fatalf("assessment replay published an extra intent: got %d want %d", got, intentsBeforeCreate+1)
	}
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, data["change_set_id"].(string))
	conflictBody := map[string]any{
		"client_txn_id":               "txn-assessments-create",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "suspected",
		"assessment.rationale":        "Different payload.",
	}
	appsupport.RequireErrorBody(t, postAssessment(t, harness, adminLogin, incidentID, conflictBody), http.StatusConflict, "client_txn_conflict")
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id = $1`, incidentID); got != intentsBeforeCreate+1 {
		t.Fatalf("assessment divergent replay published an intent: got %d want %d", got, intentsBeforeCreate+1)
	}

	beforeInvalid := loadAssessmentDurableState(t, harness, incidentID)
	invalid := map[string]any{
		"client_txn_id":               "txn-assessments-invalid",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
	}
	appsupport.RequireErrorBody(t, postAssessment(t, harness, adminLogin, incidentID, invalid), http.StatusBadRequest, "invalid_mutation_payload")
	if afterInvalid := loadAssessmentDurableState(t, harness, incidentID); afterInvalid != beforeInvalid {
		t.Fatalf("invalid create changed durable state: before=%#v after=%#v", beforeInvalid, afterInvalid)
	}
}

func TestAssessmentsAppendOnlyFiltersAndValidation(t *testing.T) {
	harness := appsupport.StartServer(t, "assessments-filters")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-assessments-filters-incident",
		"incident_key":  "IR-ASSESS-FILTERS",
		"title":         "Assessments filters",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	hostID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, hostID, "Assessment history host", "assess-history", "", "")
	otherIncident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-assessments-filters-other-incident",
		"incident_key":  "IR-ASSESS-OTHER",
		"title":         "Assessments other incident",
	})
	otherIncidentID := appsupport.MustUUID(t, otherIncident["incident_id"].(string))
	otherHostID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, otherIncidentID, adminUserID, otherHostID, "Other assessment host", "assess-other", "", "")

	created := make(map[string]string)
	for _, tc := range []struct {
		state      string
		score      any
		assessedAt string
	}{
		{state: "unknown", score: nil, assessedAt: "2026-04-24T10:00:00Z"},
		{state: "suspected", score: 39, assessedAt: "2026-04-24T11:00:00Z"},
		{state: "confirmed", score: 40, assessedAt: "2026-04-24T12:00:00Z"},
		{state: "disproven", score: 69, assessedAt: "2026-04-24T13:00:00Z"},
		{state: "cleared", score: 70, assessedAt: "2026-04-24T14:00:00Z"},
	} {
		body := map[string]any{
			"client_txn_id":               "txn-assessments-" + tc.state,
			"assessment.subject_ref":      hostID.String(),
			"assessment.subject_type":     "host",
			"assessment.assessment_state": tc.state,
			"assessment.rationale":        "Assessment " + tc.state + ".",
			"assessment.assessed_at":      tc.assessedAt,
		}
		if tc.score != nil {
			body["assessment.confidence_score"] = tc.score
		}
		data := appsupport.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusCreated)
		created[tc.state] = data["row"].(map[string]any)["record_id"].(string)
	}

	rows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), assessments.AssessmentsViewSchemaID, adminLogin)
	if len(rows) != 5 {
		t.Fatalf("expected five append-only assessment rows, got %#v", rows)
	}
	wantOrder := []string{created["cleared"], created["disproven"], created["confirmed"], created["suspected"], created["unknown"]}
	for idx, want := range wantOrder {
		if rows[idx]["record_id"] != want {
			t.Fatalf("unexpected default sort at %d: got %#v want %s", idx, rows[idx]["record_id"], want)
		}
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1 AND subject_record_id = $2`, incidentID, hostID); got != 5 {
		t.Fatalf("expected append-only assessment history, got %d rows", got)
	}

	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.assessment_state", "cleared"), []string{created["cleared"]})
	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.assessment_state", "disproven"), []string{created["disproven"]})
	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.confidence_band", "unset"), []string{created["unknown"]})
	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.confidence_band", "low"), []string{created["suspected"]})
	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.confidence_band", "medium"), []string{created["disproven"], created["confirmed"]})
	requireFilteredRecordIDs(t, harness, adminLogin, incidentID, eqFilter("assessment.confidence_band", "high"), []string{created["cleared"]})

	for _, tc := range []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing subject",
			body: map[string]any{
				"client_txn_id":               "txn-assessments-missing-subject",
				"assessment.subject_type":     "host",
				"assessment.assessment_state": "confirmed",
				"assessment.rationale":        "Valid rationale.",
			},
		},
		{
			name: "empty rationale",
			body: withAssessmentField(validAssessmentBody("txn-assessments-empty-rationale", hostID, "host", "confirmed"), "assessment.rationale", "   "),
		},
		{
			name: "mismatched subject type",
			body: validAssessmentBody("txn-assessments-bad-type", hostID, "identity", "confirmed"),
		},
		{
			name: "cross-incident subject",
			body: validAssessmentBody("txn-assessments-cross-incident", otherHostID, "host", "confirmed"),
		},
		{
			name: "invalid state",
			body: validAssessmentBody("txn-assessments-bad-state", hostID, "host", "compromised"),
		},
		{
			name: "invalid timestamp",
			body: withAssessmentField(validAssessmentBody("txn-assessments-bad-time", hostID, "host", "confirmed"), "assessment.assessed_at", "not-a-timestamp"),
		},
		{
			name: "out-of-range confidence",
			body: withAssessmentField(validAssessmentBody("txn-assessments-bad-confidence", hostID, "host", "confirmed"), "assessment.confidence_score", 101),
		},
		{
			name: "client link confidence",
			body: withAssessmentField(validAssessmentBody("txn-assessments-bad-link-confidence", hostID, "host", "confirmed"), "assessment.support_refs", map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_record_ref", "linked_record_id": hostID.String(), "confidence": 80},
				},
			}),
		},
		{
			name: "invalid assessor",
			body: withAssessmentField(validAssessmentBody("txn-assessments-bad-assessor", hostID, "host", "confirmed"), "assessment.assessor", uuid.New().String()),
		},
	} {
		before := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incidentID)
		resp := postAssessment(t, harness, adminLogin, incidentID, tc.body)
		appsupport.RequireErrorBody(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
		if after := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incidentID); after != before {
			t.Fatalf("%s left partial assessment rows: before=%d after=%d", tc.name, before, after)
		}
	}
}

func postAssessment(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/rows",
		body,
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func requireCellValue(t testing.TB, cells map[string]any, fieldKey string, want any) any {
	t.Helper()
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
	return got
}

func requireFilteredRecordIDs(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, filter map[string]any, want []string) {
	t.Helper()
	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/query",
		map[string]any{"filters": []map[string]any{filter}},
		appsupport.WithCookies(login.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rawRows := body["data"].(map[string]any)["rows"].([]any)
	got := make([]string, 0, len(rawRows))
	for _, raw := range rawRows {
		got = append(got, raw.(map[string]any)["record_id"].(string))
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected filtered rows for %#v: got %#v want %#v", filter, got, want)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected filtered rows for %#v: got %#v want %#v", filter, got, want)
		}
	}
}

func validAssessmentBody(txnID string, subjectID uuid.UUID, subjectType string, state string) map[string]any {
	return map[string]any{
		"client_txn_id":               txnID,
		"assessment.subject_ref":      subjectID.String(),
		"assessment.subject_type":     subjectType,
		"assessment.assessment_state": state,
		"assessment.rationale":        "Valid rationale.",
	}
}

func withAssessmentField(body map[string]any, key string, value any) map[string]any {
	body[key] = value
	return body
}

func eqFilter(fieldKey string, value any) map[string]any {
	return map[string]any{"field_key": fieldKey, "op": "eq", "arg": map[string]any{"value": value}}
}

type assessmentDurableState struct {
	records      int
	sourceRows   int
	links        int
	projections  int
	changeSets   int
	mutations    int
	revisions    int
	idempotency  int
	publications int
}

func loadAssessmentDurableState(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID) assessmentDurableState {
	t.Helper()
	return assessmentDurableState{
		records:      appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE incident_id = $1 AND record_type = 'assessment'`, incidentID),
		sourceRows:   appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incidentID),
		links:        appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND src_record_id IN (SELECT record_id FROM records WHERE incident_id = $1 AND record_type = 'assessment')`, incidentID),
		projections:  appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessment_grid_projection WHERE incident_id = $1`, incidentID),
		changeSets:   appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1 AND source = 'assessments.rows.create'`, incidentID),
		mutations:    appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1 AND c.source = 'assessments.rows.create'`, incidentID),
		revisions:    appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions rr JOIN records r ON r.record_id = rr.record_id WHERE r.incident_id = $1 AND r.record_type = 'assessment'`, incidentID),
		idempotency:  appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = 'assessments.rows.create' AND scope_key = $1`, incidentID.String()+":"+assessments.AssessmentsViewSchemaID),
		publications: appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id = $1`, incidentID),
	}
}
