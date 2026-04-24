package workbook_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestWorkbook_AllDiscoveredBaseSurfacesQueryEmptyIncident(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-all-surfaces-empty")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-all-surfaces-empty-incident",
		"incident_key":  "IR-WORKBOOK-EMPTY",
		"title":         "Workbook empty surface query",
	})
	incidentID := incident["incident_id"].(string)

	resp := phase4test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/view-schemas",
		nil,
		phase4test.WithCookies(adminLogin.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rawSchemas := body["data"].(map[string]any)["view_schemas"].([]any)
	if len(rawSchemas) != 14 {
		t.Fatalf("expected fourteen discovered view schemas, got %d", len(rawSchemas))
	}

	gotIDs := make([]string, 0, len(rawSchemas))
	for _, rawSchema := range rawSchemas {
		viewSchemaID := rawSchema.(map[string]any)["view_schema_id"].(string)
		gotIDs = append(gotIDs, viewSchemaID)

		queryResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query",
			map[string]any{},
			phase4test.WithCookies(adminLogin.SessionCookie),
		)
		queryBody := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)
		data := queryBody["data"].(map[string]any)
		if data["view_schema_id"] != viewSchemaID {
			t.Fatalf("query for %s returned wrong view_schema_id: %#v", viewSchemaID, data)
		}
		if data["incident_id"] != incidentID {
			t.Fatalf("query for %s returned wrong incident_id: %#v", viewSchemaID, data)
		}
		if rows := data["rows"].([]any); len(rows) != 0 {
			t.Fatalf("expected empty %s rows, got %#v", viewSchemaID, rows)
		}
		metaQuery := queryBody["meta"].(map[string]any)["query"].(map[string]any)
		if filters := metaQuery["filters"].([]any); len(filters) != 0 {
			t.Fatalf("expected empty filters for %s, got %#v", viewSchemaID, filters)
		}
		if sort := metaQuery["sort"].([]any); len(sort) == 0 {
			t.Fatalf("expected default sort for %s, got %#v", viewSchemaID, metaQuery)
		}
	}

	wantIDs := []string{
		"cartulary.view.assessments.v1",
		"cartulary.view.comm_log.v1",
		"cartulary.view.decisions.v1",
		"cartulary.view.evidence.v1",
		"cartulary.view.handoff.v1",
		"cartulary.view.hosts.v1",
		"cartulary.view.identities.v1",
		"cartulary.view.indicators.v1",
		"cartulary.view.lesson.v1",
		"cartulary.view.notes.v1",
		"cartulary.view.parties.v1",
		"cartulary.view.status_review.v1",
		"cartulary.view.task_requests.v1",
		"cartulary.view.timeline.v1",
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected discovered ids:\ngot  %v\nwant %v", gotIDs, wantIDs)
	}
}

func TestWorkbook_NewRequiredSurfacesQuerySeededRows(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-seeded-surfaces")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-seeded-surfaces-incident",
		"incident_key":  "IR-WORKBOOK-SEEDED",
		"title":         "Workbook seeded surface query",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	seeds := []struct {
		viewSchemaID string
		recordType   string
		insertChild  func(recordID uuid.UUID)
		filter       map[string]any
		wantField    string
		wantValue    any
	}{
		{
			viewSchemaID: "cartulary.view.evidence.v1",
			recordType:   "evidence",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at)
VALUES ($1, $2, 'Endpoint triage image', 'received', 'complete', '2026-04-24T10:00:00Z')
`, recordID, incidentID)
			},
			filter:    eqFilter("evidence.lifecycle_state", "received"),
			wantField: "evidence.title",
			wantValue: "Endpoint triage image",
		},
		{
			viewSchemaID: "cartulary.view.notes.v1",
			recordType:   "artifact",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
VALUES ($1, $2, 'note', 'Analyst note', 'needle appears here', $3)
`, recordID, incidentID, adminUserID)
			},
			filter:    map[string]any{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "needle"}},
			wantField: "note.title",
			wantValue: "Analyst note",
		},
		{
			viewSchemaID: "cartulary.view.assessments.v1",
			recordType:   "assessment",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO compromise_assessments (compromise_assessment_id, incident_id, subject_id, subject_type, state, confidence, assessed_by_user_id)
VALUES ($1, $2, $3, 'host', 'compromised', 80, $4)
`, recordID, incidentID, uuid.New(), adminUserID)
			},
			filter:    prefixFilter("assessment.assessment_state", "comp"),
			wantField: "assessment.confidence_band",
			wantValue: "high",
		},
		{
			viewSchemaID: "cartulary.view.task_requests.v1",
			recordType:   "task_request",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO task_requests (record_id, incident_id, title, status, priority, updated_at)
VALUES ($1, $2, 'Collect firewall logs', 'open', 'high', '2026-04-24T11:00:00Z')
`, recordID, incidentID)
			},
			filter:    prefixFilter("task.status", "op"),
			wantField: "task.title",
			wantValue: "Collect firewall logs",
		},
		{
			viewSchemaID: "cartulary.view.decisions.v1",
			recordType:   "decision",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO decisions (record_id, incident_id, summary, status, decision_type, decided_at)
VALUES ($1, $2, 'Contain workstation', 'approved', 'containment', '2026-04-24T12:00:00Z')
`, recordID, incidentID)
			},
			filter:    prefixFilter("decision.status", "app"),
			wantField: "decision.summary",
			wantValue: "Contain workstation",
		},
		{
			viewSchemaID: "cartulary.view.parties.v1",
			recordType:   "party",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind, updated_at)
VALUES ($1, $2, 'Acme Legal', 'external', '2026-04-24T12:30:00Z')
`, recordID, incidentID)
			},
			filter:    prefixFilter("party.display_name", "acme"),
			wantField: "party.party_kind",
			wantValue: "external",
		},
		{
			viewSchemaID: "cartulary.view.comm_log.v1",
			recordType:   "artifact",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, comm_type, audience, channel_or_meeting, summary, timestamp_utc, created_by_user_id)
VALUES ($1, $2, 'comm_log', 'briefing', 'leadership', 'Bridge', 'Initial update', '2026-04-24T13:00:00Z', $3)
`, recordID, incidentID, adminUserID)
			},
			filter:    prefixFilter("comm_log.comm_type", "brief"),
			wantField: "comm_log.summary",
			wantValue: "Initial update",
		},
		{
			viewSchemaID: "cartulary.view.handoff.v1",
			recordType:   "artifact",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, incoming_owner_user_id, current_state_summary, timestamp_utc, created_by_user_id)
VALUES ($1, $2, 'handoff', $3, 'Night shift takes containment', '2026-04-24T14:00:00Z', $3)
`, recordID, incidentID, adminUserID)
			},
			filter:    eqFilter("handoff.ack_state", "pending"),
			wantField: "handoff.current_state_summary",
			wantValue: "Night shift takes containment",
		},
		{
			viewSchemaID: "cartulary.view.status_review.v1",
			recordType:   "artifact",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, current_state_summary, timestamp_utc, review_owner_user_id, created_by_user_id)
VALUES ($1, $2, 'status_review', 'Containment in progress', '2026-04-24T15:00:00Z', $3, $3)
`, recordID, incidentID, adminUserID)
			},
			filter:    map[string]any{"field_key": "status_review.timestamp_day", "op": "eq", "arg": map[string]any{"value": "2026-04-24"}},
			wantField: "status_review.current_state_summary",
			wantValue: "Containment in progress",
		},
		{
			viewSchemaID: "cartulary.view.lesson.v1",
			recordType:   "artifact",
			insertChild: func(recordID uuid.UUID) {
				execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, summary, owner_user_id, closure_state, timestamp_utc, created_by_user_id)
VALUES ($1, $2, 'lesson', 'Preserve VPN logs earlier', $3, 'open', '2026-04-24T16:00:00Z', $3)
`, recordID, incidentID, adminUserID)
			},
			filter:    prefixFilter("lesson.closure_state", "op"),
			wantField: "lesson.summary",
			wantValue: "Preserve VPN logs earlier",
		},
	}

	for _, seed := range seeds {
		t.Run(seed.viewSchemaID, func(t *testing.T) {
			recordID := uuid.New()
			seedRecordEnvelope(t, harness, incidentID, adminUserID, recordID, seed.recordType)
			seed.insertChild(recordID)

			resp := phase4test.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+seed.viewSchemaID+"/query",
				map[string]any{"filters": []map[string]any{seed.filter}},
				phase4test.WithCookies(adminLogin.SessionCookie),
			)
			if resp.StatusCode != http.StatusOK {
				payload, _ := io.ReadAll(resp.Body)
				t.Fatalf("unexpected status for %s: %d %s", seed.viewSchemaID, resp.StatusCode, string(payload))
			}
			body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
			rows := body["data"].(map[string]any)["rows"].([]any)
			if len(rows) != 1 {
				t.Fatalf("expected one %s row, got %#v", seed.viewSchemaID, rows)
			}
			row := rows[0].(map[string]any)
			if row["record_id"] != recordID.String() {
				t.Fatalf("unexpected record_id for %s: %#v", seed.viewSchemaID, row)
			}
			cells := row["cells"].(map[string]any)
			if got := cells[seed.wantField].(map[string]any)["value"]; got != seed.wantValue {
				t.Fatalf("unexpected %s value: got %#v want %#v", seed.wantField, got, seed.wantValue)
			}
		})
	}
}

func seedRecordEnvelope(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorID)
}

func execSeed(t testing.TB, harness *phase4test.ServerHarness, sql string, args ...any) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), sql, args...); err != nil {
		t.Fatalf("seed query failed: %v", err)
	}
}

func eqFilter(fieldKey string, value any) map[string]any {
	return map[string]any{"field_key": fieldKey, "op": "eq", "arg": map[string]any{"value": value}}
}

func prefixFilter(fieldKey string, value string) map[string]any {
	return map[string]any{"field_key": fieldKey, "op": "prefix", "arg": map[string]any{"value": value}}
}
