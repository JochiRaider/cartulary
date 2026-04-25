package workbook_test

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
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
		paging := queryBody["meta"].(map[string]any)["paging"].(map[string]any)
		if paging["limit"] != float64(100) || paging["has_more"] != false || paging["next_cursor"] != nil {
			t.Fatalf("expected terminal default paging for %s, got %#v", viewSchemaID, paging)
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

func TestWorkbook_QueryPaginationContract(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-query-pagination")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-pagination-incident",
		"incident_key":  "IR-WORKBOOK-PAGING",
		"title":         "Workbook query pagination",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	hostA := uuid.New()
	hostB := uuid.New()
	hostC := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostC, "Charlie")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	first := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit": 1,
		"sort":  sortByName,
	})
	firstRows := responseRows(first)
	if len(firstRows) != 1 || firstRows[0]["record_id"] != hostA.String() {
		t.Fatalf("expected first page host A, got %#v", firstRows)
	}
	firstPaging := responsePaging(first)
	if firstPaging["limit"] != float64(1) || firstPaging["has_more"] != true || firstPaging["next_cursor"] == nil {
		t.Fatalf("expected first page cursor, got %#v", firstPaging)
	}
	cursor := firstPaging["next_cursor"].(string)

	second := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	})
	secondPaging := responsePaging(second)
	if secondPaging["limit"] != float64(1) || secondPaging["has_more"] != true || secondPaging["next_cursor"] == nil {
		t.Fatalf("expected continuation to reuse limit 1, got %#v", secondPaging)
	}
	if rows := responseRows(second); len(rows) != 1 || rows[0]["record_id"] != hostB.String() {
		t.Fatalf("expected second page host B, got %#v", rows)
	}

	terminal := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": secondPaging["next_cursor"].(string),
		"sort":         sortByName,
	})
	terminalPaging := responsePaging(terminal)
	if terminalPaging["limit"] != float64(1) || terminalPaging["has_more"] != false || terminalPaging["next_cursor"] != nil {
		t.Fatalf("expected terminal continuation, got %#v", terminalPaging)
	}
	if rows := responseRows(terminal); len(rows) != 1 || rows[0]["record_id"] != hostC.String() {
		t.Fatalf("expected terminal page host C, got %#v", rows)
	}

	for name, body := range map[string]map[string]any{
		"changed limit":    {"cursor_token": cursor, "limit": 2, "sort": sortByName},
		"changed sort":     {"cursor_token": cursor, "sort": []map[string]any{{"field_key": "host.display_name", "direction": "desc"}}},
		"changed group_by": {"cursor_token": cursor, "sort": sortByName, "group_by": "host.host_state"},
	} {
		t.Run(name, func(t *testing.T) {
			resp := phase4test.DoJSON(t, http.MethodPost, queryURL, body, phase4test.WithCookies(adminLogin.SessionCookie))
			errBody := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_view_query")
			if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "cursor_query_mismatch" {
				t.Fatalf("expected cursor_query_mismatch, got %#v", details)
			}
		})
	}

	zero := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"filters": []map[string]any{{"field_key": "host.location", "op": "eq", "arg": map[string]any{"value": "missing-location"}}},
	})
	if rows := responseRows(zero); len(rows) != 0 {
		t.Fatalf("expected zero-match page, got %#v", rows)
	}
	zeroPaging := responsePaging(zero)
	if zeroPaging["limit"] != float64(100) || zeroPaging["has_more"] != false || zeroPaging["next_cursor"] != nil {
		t.Fatalf("expected zero-match terminal paging, got %#v", zeroPaging)
	}

	grouped := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit":    1,
		"sort":     sortByName,
		"group_by": "host.host_state",
	})
	groupedRows := responseRows(grouped)
	if len(groupedRows) != 1 {
		t.Fatalf("expected grouping page size to count rows only, got %#v", groupedRows)
	}
	if groupedRows[0]["group_values"].(map[string]any)["host.host_state"] != "canonical" {
		t.Fatalf("expected full row group_values, got %#v", groupedRows[0])
	}
}

func TestWorkbook_QueryCursorSnapshotsSurviveInterveningMutations(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-query-snapshot")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-snapshot-incident",
		"incident_key":  "IR-WORKBOOK-SNAPSHOT",
		"title":         "Workbook query snapshot",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	hostA := uuid.New()
	hostC := uuid.New()
	hostD := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostC, "Charlie")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostD, "Delta")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit": 1,
		"sort":  sortByName,
	})
	cursor := responsePaging(pageOne)["next_cursor"].(string)

	hostB := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")
	updateHostDisplayNameForPaging(t, harness, hostC, "Zulu", 2)

	pageTwo := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	})
	pageTwoRows := responseRows(pageTwo)
	if len(pageTwoRows) != 1 || pageTwoRows[0]["record_id"] != hostC.String() {
		t.Fatalf("expected snapshot continuation to preserve host C as second row, got %#v", pageTwoRows)
	}
	if pageTwoRows[0]["row_version"] != float64(1) {
		t.Fatalf("expected snapshot row_version 1, got %#v", pageTwoRows[0])
	}
	if got := pageTwoRows[0]["cells"].(map[string]any)["host.display_name"].(map[string]any)["value"]; got != "Charlie" {
		t.Fatalf("expected snapshot display name Charlie, got %#v", pageTwoRows[0])
	}

	pageThree := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": responsePaging(pageTwo)["next_cursor"].(string),
		"sort":         sortByName,
	})
	pageThreeRows := responseRows(pageThree)
	if len(pageThreeRows) != 1 || pageThreeRows[0]["record_id"] != hostD.String() {
		t.Fatalf("expected snapshot continuation to preserve host D as third row, got %#v", pageThreeRows)
	}
	if responsePaging(pageThree)["has_more"] != false {
		t.Fatalf("expected snapshot chain to terminate, got %#v", responsePaging(pageThree))
	}

	fresh := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"sort": sortByName})
	freshIDs := rowIDs(responseRows(fresh))
	wantFresh := []string{hostA.String(), hostB.String(), hostD.String(), hostC.String()}
	if !reflect.DeepEqual(freshIDs, wantFresh) {
		t.Fatalf("fresh query did not reflect live ordering:\ngot  %v\nwant %v", freshIDs, wantFresh)
	}
	freshC := findResponseRow(t, responseRows(fresh), hostC)
	if freshC["row_version"] != float64(2) {
		t.Fatalf("expected fresh host C row_version 2, got %#v", freshC)
	}
}

func TestWorkbook_QueryCursorSnapshotExpiry(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-query-snapshot-expiry")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-snapshot-expiry-incident",
		"incident_key":  "IR-WORKBOOK-SNAPSHOT-EXPIRY",
		"title":         "Workbook query snapshot expiry",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	hostA := uuid.New()
	hostB := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit": 1,
		"sort":  sortByName,
	})
	cursor := responsePaging(pageOne)["next_cursor"].(string)

	harness.Server.Clock.SetOffset(9 * time.Minute)
	beforeExpiry := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	})
	if rows := responseRows(beforeExpiry); len(rows) != 1 || rows[0]["record_id"] != hostB.String() {
		t.Fatalf("expected continuation before expiry to succeed, got %#v", rows)
	}

	pageOne = queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit": 1,
		"sort":  sortByName,
	})
	expiringCursor := responsePaging(pageOne)["next_cursor"].(string)
	harness.Server.Clock.SetOffset(20 * time.Minute)
	resp := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": expiringCursor,
		"sort":         sortByName,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "cursor_snapshot_unavailable" {
		t.Fatalf("expected cursor_snapshot_unavailable, got %#v", details)
	}

	fresh := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"sort": sortByName})
	if rows := responseRows(fresh); len(rows) != 2 {
		t.Fatalf("expected fresh query after expiry to succeed, got %#v", rows)
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
				subjectID := uuid.New()
				phase4test.SeedRecordEnvelope(t, harness.DB, incidentID, adminUserID, subjectID, "host")
				execSeed(t, harness, `
INSERT INTO hosts (record_id, incident_id, display_name, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'Assessment subject', 'canonical', $3, $3)
`, subjectID, incidentID, adminUserID)
				execSeed(t, harness, `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, confidence_score, rationale, assessor_user_id)
VALUES ($1, $2, $3, 'host', 'confirmed', 80, 'Seeded test assessment rationale.', $4)
`, recordID, incidentID, subjectID, adminUserID)
				execSeed(t, harness, `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_score,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT a.record_id, a.incident_id, r.row_version, a.subject_record_id, a.subject_type, a.assessment_state, a.confidence_score, 'high', a.rationale, a.assessor_user_id, a.assessed_at, 0
  FROM assessments a
  JOIN records r ON r.record_id = a.record_id
 WHERE a.record_id = $1
`, recordID)
			},
			filter:    prefixFilter("assessment.assessment_state", "conf"),
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

func seedHostForPaging(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'host', $3, $3)
`, recordID, incidentID, actorID)
	execSeed(t, harness, `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, lower($3), 'canonical', $4, $4)
`, recordID, incidentID, displayName, actorID)
	execSeed(t, harness, `
INSERT INTO host_grid_projection (record_id, incident_id, row_version, display_name, hostname, host_state, edited_at)
VALUES ($1, $2, 1, $3, lower($3), 'canonical', now())
`, recordID, incidentID, displayName)
}

func updateHostDisplayNameForPaging(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, displayName string, rowVersion int64) {
	t.Helper()
	execSeed(t, harness, `
UPDATE records
   SET row_version = $2,
       updated_at = now()
 WHERE record_id = $1
`, recordID, rowVersion)
	execSeed(t, harness, `
UPDATE hosts
   SET display_name = $2,
       row_version = $3,
       updated_at = now()
 WHERE record_id = $1
`, recordID, displayName, rowVersion)
	execSeed(t, harness, `
UPDATE host_grid_projection
   SET display_name = $2,
       row_version = $3,
       edited_at = now()
 WHERE record_id = $1
`, recordID, displayName, rowVersion)
}

func queryWorkbook(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, queryURL string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, queryURL, body, phase4test.WithCookies(login.SessionCookie))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func responseRows(envelope map[string]any) []map[string]any {
	rawRows := envelope["data"].(map[string]any)["rows"].([]any)
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		rows = append(rows, rawRow.(map[string]any))
	}
	return rows
}

func responsePaging(envelope map[string]any) map[string]any {
	return envelope["meta"].(map[string]any)["paging"].(map[string]any)
}

func rowIDs(rows []map[string]any) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row["record_id"].(string))
	}
	return ids
}

func findResponseRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("row %s not found in %#v", recordID, rows)
	return nil
}
