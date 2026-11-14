package assessments_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestAssessmentsCreateAndProjection(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "assessments-create", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-assessments-create-incident",
		"incident_key":  "IR-ASSESS-CREATE",
		"title":         "Assessments create",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	hostID := uuid.New()
	supportID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	supportHostID := uuid.MustParse("f0000000-0000-4000-8000-000000000002")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, hostID, "Assessment host", "assess-host", "", "")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, supportHostID, "Assessment support host", "assess-support-host", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, supportID)

	intentsBeforeCreate := collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()})
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
	data := httptestx.RequireSuccessData(t, create, http.StatusCreated)
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
	wantSupport := map[string]any{
		"kind": "collection_value_v1", "ordered": false,
		"items": []any{
			map[string]any{"item_ref": "record_ref:" + supportID.String(), "item_kind": "record_ref", "display_text": "timeline_event:" + supportID.String(), "linked_record_id": supportID.String()},
			map[string]any{"item_ref": "record_ref:" + supportHostID.String(), "item_kind": "record_ref", "display_text": "host:" + supportHostID.String(), "linked_record_id": supportHostID.String()},
		},
	}
	requireSupport := func(observation string, observedRow map[string]any) {
		t.Helper()
		value := observedRow["cells"].(map[string]any)["assessment.support_refs"].(map[string]any)["value"]
		if !reflect.DeepEqual(value, wantSupport) {
			t.Fatalf("%s support collection = %#v, want %#v", observation, value, wantSupport)
		}
	}
	requireSupport("create", row)
	queried := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), assessments.AssessmentsViewSchemaID, adminLogin)
	if len(queried) != 1 || queried[0]["record_id"] != recordID.String() || queried[0]["row_version"] != row["row_version"] {
		t.Fatalf("unexpected Assessment query identity/version: %#v", queried)
	}
	requireSupport("query", queried[0])

	link := linktest.LookupActiveLink(t, harness.DB, incidentID, recordID, supportID, "supported_by")
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("unexpected support link metadata: %#v", link)
	}
	hostLink := linktest.LookupActiveLink(t, harness.DB, incidentID, recordID, supportHostID, "supported_by")
	if hostLink.Provenance != "manual" || hostLink.Confidence != nil {
		t.Fatalf("unexpected heterogeneous support link metadata: %#v", hostLink)
	}
	if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_kind = 'record_link'
`, appsupport.MustUUID(t, data["change_set_id"].(string))); got != 2 {
		t.Fatalf("assessment support link mutations = %d, want 2", got)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE record_id = $1 AND record_type = 'assessment'`, recordID); got != 1 {
		t.Fatalf("expected one assessment record envelope, got %d", got)
	}
	if got := collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}); got != intentsBeforeCreate+1 {
		t.Fatalf("assessment create publication intents = %d, want %d", got, intentsBeforeCreate+1)
	}
	revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, data["change_set_id"].(string))
	var committedAt time.Time
	if err := harness.DB.QueryRowContext(t.Context(), `SELECT created_at FROM assessments WHERE record_id = $1`, recordID).Scan(&committedAt); err != nil {
		t.Fatal(err)
	}
	requireCellValue(t, cells, "assessment.assessed_at", committedAt.UTC().Format(time.RFC3339Nano))
	beforeReplay := loadAssessmentDurableState(t, harness, incidentID)

	replayBody := map[string]any{
		"client_txn_id":               "txn-assessments-create",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "  host  ",
		"assessment.assessment_state": " confirmed ",
		"assessment.rationale":        "Confirmed from supporting event.",
		"assessment.assessor":         adminUserID.String(),
		"assessment.assessed_at":      committedAt.UTC().Format(time.RFC3339Nano),
		"assessment.confidence_score": nil,
		"assessment.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportHostID.String()},
				{"op": "add_record_ref", "linked_record_id": supportID.String()},
			},
		},
	}
	replay := postAssessment(t, harness, adminLogin, incidentID, replayBody)
	replayData := httptestx.RequireSuccessData(t, replay, http.StatusOK)
	requireSupport("replay", replayData["row"].(map[string]any))
	if !reflect.DeepEqual(replayData, data) {
		t.Fatalf("Assessment replay changed committed result: got %#v want %#v", replayData, data)
	}
	if afterReplay := loadAssessmentDurableState(t, harness, incidentID); afterReplay != beforeReplay {
		t.Fatalf("normalized replay changed durable state: before=%#v after=%#v", beforeReplay, afterReplay)
	}
	if got := replayData["row"].(map[string]any)["record_id"]; got != recordID.String() {
		t.Fatalf("expected idempotent replay of record %s, got %#v", recordID, replayData)
	}
	if got := collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}); got != intentsBeforeCreate+1 {
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
	httptestx.RequireErrorEnvelope(t, postAssessment(t, harness, adminLogin, incidentID, conflictBody), http.StatusConflict, "client_txn_conflict")
	if got := collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}); got != intentsBeforeCreate+1 {
		t.Fatalf("assessment divergent replay published an intent: got %d want %d", got, intentsBeforeCreate+1)
	}

	beforeInvalid := loadAssessmentDurableState(t, harness, incidentID)
	invalid := map[string]any{
		"client_txn_id":               "txn-assessments-invalid",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
	}
	httptestx.RequireErrorEnvelope(t, postAssessment(t, harness, adminLogin, incidentID, invalid), http.StatusBadRequest, "invalid_mutation_payload")
	if afterInvalid := loadAssessmentDurableState(t, harness, incidentID); afterInvalid != beforeInvalid {
		t.Fatalf("invalid create changed durable state: before=%#v after=%#v", beforeInvalid, afterInvalid)
	}

	t.Run("explicit time is distinct from original default context", func(t *testing.T) {
		body := validAssessmentBody("txn-explicit-assessment-time", hostID, "host", "confirmed")
		body["assessment.assessed_at"] = "2026-04-24T12:00:00.123456789Z"
		body["assessment.confidence_score"] = 0
		created := httptestx.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusCreated)
		createdCells := created["row"].(map[string]any)["cells"].(map[string]any)
		requireCellValue(t, createdCells, "assessment.assessed_at", "2026-04-24T12:00:00.123456Z")
		before := loadAssessmentDurableState(t, harness, incidentID)
		body["assessment.assessed_at"] = "2026-04-24T13:00:00.123456+01:00"
		replayed := httptestx.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusOK)
		if !reflect.DeepEqual(replayed, created) {
			t.Fatalf("persisted timestamp replay changed original result: %#v", replayed)
		}
		delete(body, "assessment.assessed_at")
		httptestx.RequireErrorEnvelope(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusConflict, "client_txn_conflict")
		body["assessment.assessed_at"] = "2026-04-24T12:00:00.123456Z"
		body["assessment.confidence_score"] = nil
		httptestx.RequireErrorEnvelope(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusConflict, "client_txn_conflict")
		if after := loadAssessmentDurableState(t, harness, incidentID); after != before {
			t.Fatalf("timestamp/default replay changed durable effects: %#v -> %#v", before, after)
		}
	})

	t.Run("legacy receipt retains exact comparison without rewrite", func(t *testing.T) {
		body := validAssessmentBody("txn-legacy-assessment-replay", hostID, "host", "confirmed")
		created := httptestx.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusCreated)
		// Independently construct the historical admitted-request preimage. The
		// fixture represents a receipt written before normalized comparison.
		preimage, err := json.Marshal(map[string]any{
			"client_txn_id":               "txn-legacy-assessment-replay",
			"view_schema_id":              "cartulary.view.assessments.v1",
			"assessment.subject_ref":      hostID.String(),
			"assessment.subject_type":     "host",
			"assessment.assessment_state": "confirmed",
			"assessment.rationale":        "Valid rationale.",
		})
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(preimage)
		scope := incidentID.String() + ":" + assessments.AssessmentsViewSchemaID
		if _, err := harness.DB.ExecContext(t.Context(), `UPDATE route_idempotency
SET request_hash = $1,
    response_json = (response_json - 'default_assessed_at') || '{"schema_id":"cartulary.assessments.create_result.v1"}'::jsonb
WHERE route_key = 'assessments.rows.create' AND actor_user_id = $2 AND scope_key = $3 AND client_txn_id = $4`, hash[:], adminUserID, scope, body["client_txn_id"]); err != nil {
			t.Fatal(err)
		}
		readReceipt := func() ([]byte, []byte) {
			t.Helper()
			var storedHash, storedResponse []byte
			if err := harness.DB.QueryRowContext(t.Context(), `SELECT request_hash, response_json FROM route_idempotency WHERE route_key = 'assessments.rows.create' AND actor_user_id = $1 AND scope_key = $2 AND client_txn_id = $3`, adminUserID, scope, body["client_txn_id"]).Scan(&storedHash, &storedResponse); err != nil {
				t.Fatal(err)
			}
			return storedHash, storedResponse
		}
		beforeHash, beforeResponse := readReceipt()
		before := loadAssessmentDurableState(t, harness, incidentID)
		replayed := httptestx.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusOK)
		if !reflect.DeepEqual(replayed, created) {
			t.Fatalf("legacy replay changed original result: %#v", replayed)
		}
		body["assessment.assessor"] = adminUserID.String()
		httptestx.RequireErrorEnvelope(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusConflict, "client_txn_conflict")
		afterHash, afterResponse := readReceipt()
		if !bytes.Equal(beforeHash, afterHash) || !bytes.Equal(beforeResponse, afterResponse) {
			t.Fatal("legacy replay rewrote its retained receipt")
		}
		if after := loadAssessmentDurableState(t, harness, incidentID); after != before {
			t.Fatalf("legacy replay changed durable effects: %#v -> %#v", before, after)
		}
	})
	for _, divergent := range []bool{false, true} {
		name := "identical concurrent requests"
		key := "txn-assessment-concurrent-identical"
		if divergent {
			name, key = "divergent concurrent requests", "txn-assessment-concurrent-divergent"
		}
		t.Run(name, func(t *testing.T) {
			first := validAssessmentBody(key, hostID, "host", "confirmed")
			second := validAssessmentBody(key, hostID, "host", "confirmed")
			if divergent {
				second["assessment.rationale"] = "Different concurrent rationale."
			}
			requireConcurrentAssessmentReceipt(t, harness, adminLogin, incidentID, first, second, divergent)
		})
	}
	t.Run("current subject admission", func(t *testing.T) {
		exerciseAssessmentSubjectAdmission(t, harness, adminLogin, incidentID, adminUserID)
	})
	t.Run("overlapping subject and support admission", func(t *testing.T) {
		exerciseAssessmentReferenceOverlap(t, harness, adminLogin, incidentID, adminUserID)
	})
}

func requireConcurrentAssessmentReceipt(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, first, second map[string]any, divergent bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	requests := make([]*http.Request, 0, 2)
	for _, body := range []map[string]any{first, second} {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/rows", bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
		request.AddCookie(login.Session)
		request.AddCookie(login.CSRF)
		requests = append(requests, request)
	}
	before := loadAssessmentDurableState(t, harness, incidentID)
	barrier, err := harness.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 2*time.Second)
		defer stop()
		_ = barrier.Rollback(cleanup)
	}()
	if _, err := barrier.Exec(ctx, `LOCK TABLE assessments IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	pid := int32(barrier.Conn().PgConn().PID())
	type outcome struct {
		response *http.Response
		err      error
	}
	results := make(chan outcome, 2)
	settled := 0
	defer func() {
		cancel()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for settled < 2 {
			select {
			case result := <-results:
				settled++
				if result.response != nil {
					_ = result.response.Body.Close()
				}
			case <-deadline.C:
				t.Error("concurrent Assessment requests did not settle after cancellation")
				return
			}
		}
	}()
	for _, request := range requests {
		go func() { response, err := http.DefaultClient.Do(request); results <- outcome{response, err} }()
	}
	for {
		var waiting int
		if err := harness.Pool.QueryRow(ctx, `WITH RECURSIVE blocked(pid) AS (
SELECT pid FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid))
UNION SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid=ANY(pg_blocking_pids(a.pid))
) SELECT count(DISTINCT pid) FROM blocked`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting >= 2 {
			break
		}
		select {
		case early := <-results:
			settled++
			if early.response != nil {
				_ = early.response.Body.Close()
			}
			t.Fatalf("Assessment request finished before database barrier: %v", early.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := barrier.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	responses := make([]*http.Response, 0, 2)
	for settled < 2 {
		select {
		case result := <-results:
			settled++
			if result.response != nil {
				defer result.response.Body.Close()
			}
			if result.err != nil {
				t.Fatal(result.err)
			}
			responses = append(responses, result.response)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	statuses := []int{responses[0].StatusCode, responses[1].StatusCode}
	slices.Sort(statuses)
	wantStatuses := []int{http.StatusOK, http.StatusCreated}
	if divergent {
		wantStatuses = []int{http.StatusCreated, http.StatusConflict}
	}
	if !reflect.DeepEqual(statuses, wantStatuses) {
		t.Fatalf("concurrent status = %v, want %v", statuses, wantStatuses)
	}
	var committed map[string]any
	for _, response := range responses {
		if response.StatusCode == http.StatusConflict {
			envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
			if envelope["error"].(map[string]any)["details"].(map[string]any)["client_txn_id"] != first["client_txn_id"] {
				t.Fatal("conflict lost scoped transaction identity")
			}
			continue
		}
		data := httptestx.RequireSuccessData(t, response, response.StatusCode)
		if committed == nil {
			committed = data
		} else if !reflect.DeepEqual(data, committed) {
			t.Fatal("concurrent replay changed original committed result")
		}
	}
	want := before
	want.records++
	want.sourceRows++
	want.projections++
	want.changeSets++
	want.mutations++
	want.revisions++
	want.idempotency++
	want.publications++
	if after := loadAssessmentDurableState(t, harness, incidentID); after != want {
		t.Fatalf("concurrent effects = %#v, want %#v", after, want)
	}
}

func TestAssessmentsAppendOnlyFiltersAndValidation(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "assessments-filters", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
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
		data := httptestx.RequireSuccessData(t, postAssessment(t, harness, adminLogin, incidentID, body), http.StatusCreated)
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
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
		if after := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incidentID); after != before {
			t.Fatalf("%s left partial assessment rows: before=%d after=%d", tc.name, before, after)
		}
	}
}

func postAssessment(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/rows",
		body,
		httptestx.WithCookies(login.Session, login.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value),
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

func requireFilteredRecordIDs(t testing.TB, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, filter map[string]any, want []string) {
	t.Helper()
	resp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+assessments.AssessmentsViewSchemaID+"/query",
		map[string]any{"filters": []map[string]any{filter}},
		httptestx.WithCookies(login.Session),
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
		publications: collaborationsupport.CountIntents(t, harness.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}),
	}
}
