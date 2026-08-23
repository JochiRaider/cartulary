package workbook_test

import (
	"context"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbook_AllDiscoveredBaseSurfacesQueryEmptyIncident(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-all-surfaces-empty")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-all-surfaces-empty-incident",
		"incident_key":  "IR-WORKBOOK-EMPTY",
		"title":         "Workbook empty surface query",
	})
	incidentID := incident["incident_id"].(string)

	resp := appsupport.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/view-schemas",
		nil,
		appsupport.WithCookies(adminLogin.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rawSchemas := body["data"].(map[string]any)["view_schemas"].([]any)
	registrySchemas := viewschema.ListPublicResources()
	if len(rawSchemas) != len(registrySchemas) {
		t.Fatalf("discovery returned %d view schemas, machine registry has %d", len(rawSchemas), len(registrySchemas))
	}

	gotIDs := make([]string, 0, len(rawSchemas))
	for _, rawSchema := range rawSchemas {
		viewSchemaID := rawSchema.(map[string]any)["view_schema_id"].(string)
		gotIDs = append(gotIDs, viewSchemaID)

		queryResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query",
			map[string]any{},
			appsupport.WithCookies(adminLogin.SessionCookie),
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

	wantIDs := make([]string, 0, len(registrySchemas))
	for _, resource := range registrySchemas {
		wantIDs = append(wantIDs, resource.ViewSchemaID)
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected discovered ids:\ngot  %v\nwant %v", gotIDs, wantIDs)
	}
}

func TestWorkbook_AllRegisteredSurfacesHaveCreateAdmission_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-all-surfaces-create-admission")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-all-surfaces-create-admission-incident",
		"incident_key":  "IR-WORKBOOK-CREATE-ADMISSION",
		"title":         "Workbook create admission coverage",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	resources := viewschema.ListPublicResources()
	if len(resources) == 0 {
		t.Fatal("machine view-schema registry is empty")
	}
	for index, resource := range resources {
		t.Run(resource.ViewSchemaID, func(t *testing.T) {
			response := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, resource.ViewSchemaID, uuid.Nil, map[string]any{
				"client_txn_id": "txn-workbook-create-admission-" + resource.ViewSchemaID + "-" + string(rune('a'+index)),
			})
			switch response.StatusCode {
			case http.StatusCreated:
				data := httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
				if data["view_schema_id"] != resource.ViewSchemaID {
					t.Fatalf("create admission for %s returned wrong surface: %#v", resource.ViewSchemaID, data)
				}
			case http.StatusBadRequest:
				body := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
				details := httptestx.RequireErrorDetails(t, body)
				if details["reason_code"] == "unknown_view_schema" || details["reason_code"] == "unsupported_view_schema" {
					t.Fatalf("registered create surface %s reached an unmapped dispatcher: %#v", resource.ViewSchemaID, body)
				}
			default:
				t.Fatalf("registered create surface %s returned status %d body=%#v", resource.ViewSchemaID, response.StatusCode, httptestx.ReadJSONBody(t, response))
			}
		})
	}
}

func TestWorkbook_ProjectionBackedQueryRouteUsesCommonBoundaryBehavior(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-projection-query-boundary")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-projection-query-boundary-incident",
		"incident_key":  "IR-WORKBOOK-PROJECTION-QUERY",
		"title":         "Workbook projection query boundary",
	})
	incidentID := incident["incident_id"].(string)
	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + artifacts.NotesViewSchemaID + "/query"

	unauthenticated := appsupport.DoJSON(t, http.MethodPost, queryURL, map[string]any{})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	invalidSort := appsupport.DoJSON(
		t,
		http.MethodPost,
		queryURL,
		map[string]any{"sort": []map[string]any{{"field_key": "timeline.date_entered", "direction": "asc"}}},
		appsupport.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, invalidSort, http.StatusBadRequest, "invalid_view_query")

	valid := appsupport.DoJSON(
		t,
		http.MethodPost,
		queryURL,
		map[string]any{"limit": 1},
		appsupport.WithCookies(adminLogin.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, valid, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["view_schema_id"] != artifacts.NotesViewSchemaID || data["incident_id"] != incidentID {
		t.Fatalf("projection-backed query returned wrong route identity: %#v", data)
	}
	paging := body["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["limit"] != float64(1) || paging["has_more"] != false || paging["next_cursor"] != nil {
		t.Fatalf("projection-backed query did not use common paging metadata: %#v", paging)
	}
}

func TestWorkbook_CoordinationDefaultQueryReturnsCreatedRows(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-coordination-default-query")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-coordination-default-query-incident",
		"incident_key":  "IR-WORKBOOK-COORD-QUERY",
		"title":         "Workbook coordination default query",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	tests := []struct {
		name          string
		viewSchemaID  string
		body          map[string]any
		wantField     string
		wantValue     any
		requireFields func(testing.TB, map[string]any)
	}{
		{
			name:         "comm_log",
			viewSchemaID: artifacts.CommLogViewSchemaID,
			body: map[string]any{
				"client_txn_id":               "txn-workbook-default-query-comm",
				"comm_log.comm_type":          "briefing",
				"comm_log.audience":           "leadership",
				"comm_log.channel_or_meeting": "Bridge",
				"comm_log.summary":            "Default query comm log",
			},
			wantField: "comm_log.summary",
			wantValue: "Default query comm log",
			requireFields: func(t testing.TB, row map[string]any) {
				t.Helper()
				requireNonEmptyCellValue(t, row, "comm_log.timestamp_utc")
				requireCollectionItemCount(t, row, "comm_log.decision_ids", 0)
				requireCollectionItemCount(t, row, "comm_log.action_task_ids", 0)
				requireCollectionItemCount(t, row, "comm_log.audience_party_ids", 0)
				requireCollectionItemCount(t, row, "comm_log.attendee_party_ids", 0)
				requireCellValue(t, row, "comm_log.next_report_at", nil)
				requireCellValue(t, row, "comm_log.privilege_tag", nil)
			},
		},
		{
			name:         "handoff",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			body: map[string]any{
				"client_txn_id":                  "txn-workbook-default-query-handoff",
				"handoff.incoming_owner_user_id": adminUserID.String(),
				"handoff.current_state_summary":  "Default query handoff",
			},
			wantField: "handoff.current_state_summary",
			wantValue: "Default query handoff",
			requireFields: func(t testing.TB, row map[string]any) {
				t.Helper()
				requireNonEmptyCellValue(t, row, "handoff.timestamp_utc")
				requireCellValue(t, row, "handoff.outgoing_owner_user_id", adminUserID.String())
				requireCollectionItemCount(t, row, "handoff.open_task_ids", 0)
				requireCollectionItemCount(t, row, "handoff.open_decision_ids", 0)
				requireCollectionItemCount(t, row, "handoff.open_risk_refs", 0)
				requireCellValue(t, row, "handoff.next_checks", nil)
				requireCellValue(t, row, "handoff.acknowledged_at", nil)
			},
		},
		{
			name:         "status_review",
			viewSchemaID: artifacts.StatusReviewViewSchemaID,
			body: map[string]any{
				"client_txn_id":                       "txn-workbook-default-query-status",
				"status_review.current_state_summary": "Default query status review",
			},
			wantField: "status_review.current_state_summary",
			wantValue: "Default query status review",
			requireFields: func(t testing.TB, row map[string]any) {
				t.Helper()
				requireNonEmptyCellValue(t, row, "status_review.timestamp_utc")
				requireCellValue(t, row, "status_review.review_owner_user_id", adminUserID.String())
				requireCollectionItemCount(t, row, "status_review.blocked_task_ids", 0)
				requireCollectionItemCount(t, row, "status_review.pending_evidence_ids", 0)
				requireCollectionItemCount(t, row, "status_review.open_decision_ids", 0)
				requireCellValue(t, row, "status_review.active_risks_summary", nil)
				requireCellValue(t, row, "status_review.next_report_at", nil)
			},
		},
		{
			name:         "lesson",
			viewSchemaID: artifacts.LessonViewSchemaID,
			body: map[string]any{
				"client_txn_id":  "txn-workbook-default-query-lesson",
				"lesson.summary": "Default query lesson",
			},
			wantField: "lesson.summary",
			wantValue: "Default query lesson",
			requireFields: func(t testing.TB, row map[string]any) {
				t.Helper()
				requireNonEmptyCellValue(t, row, "lesson.timestamp_utc")
				requireCellValue(t, row, "lesson.owner_user_id", adminUserID.String())
				requireCellValue(t, row, "lesson.closure_state", "open")
				requireCollectionItemCount(t, row, "lesson.follow_up_task_ids", 0)
				requireCollectionItemCount(t, row, "lesson.evidence_refs", 0)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			created := requireWorkbookCreate(t, harness, adminLogin, incidentID, tc.viewSchemaID, tc.body)
			createdRow := created["row"].(map[string]any)
			recordID := appsupport.MustUUID(t, createdRow["record_id"].(string))
			requireCellValue(t, createdRow, tc.wantField, tc.wantValue)

			queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + tc.viewSchemaID + "/query"
			query := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{})
			data := query["data"].(map[string]any)
			if data["view_schema_id"] != tc.viewSchemaID {
				t.Fatalf("default query for %s returned wrong view_schema_id: %#v", tc.viewSchemaID, data)
			}
			row := findResponseRow(t, responseRows(query), recordID)
			requireDefaultVisibleCells(t, tc.viewSchemaID, row)
			requireCellValue(t, row, tc.wantField, tc.wantValue)
			tc.requireFields(t, row)
		})
	}
}

func TestWorkbook_QueryPaginationContract(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-query-pagination")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-pagination-incident",
		"incident_key":  "IR-WORKBOOK-PAGING",
		"title":         "Workbook query pagination",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	hostA := uuid.New()
	hostB := uuid.New()
	hostC := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostC, "Charlie")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + hostidentity.HostsViewSchemaID + "/query"
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
			resp := appsupport.DoJSON(t, http.MethodPost, queryURL, body, appsupport.WithCookies(adminLogin.SessionCookie))
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

	for name, sort := range map[string][]map[string]any{
		"numeric duplicate sort": {{"field_key": "host.linked_event_count", "direction": "desc"}},
		"null duplicate sort":    {{"field_key": "host.location", "direction": "asc"}},
	} {
		t.Run(name, func(t *testing.T) {
			pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"limit": 1, "sort": sort})
			pageOneRows := responseRows(pageOne)
			pageOneCursor, ok := responsePaging(pageOne)["next_cursor"].(string)
			if len(pageOneRows) != 1 || !ok {
				t.Fatalf("first duplicate-value page = %#v, paging=%#v", pageOneRows, responsePaging(pageOne))
			}
			pageTwo := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"cursor_token": pageOneCursor, "sort": sort})
			pageTwoRows := responseRows(pageTwo)
			if len(pageTwoRows) != 1 || pageTwoRows[0]["record_id"] == pageOneRows[0]["record_id"] {
				t.Fatalf("duplicate-value keyset did not advance: first=%#v second=%#v", pageOneRows, pageTwoRows)
			}
		})
	}
}

func TestWorkbook_QueryCursorContinuationUsesLiveRows(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-query-live-cursor")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-live-cursor-incident",
		"incident_key":  "IR-WORKBOOK-LIVE-CURSOR",
		"title":         "Workbook query live cursor",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	hostA := uuid.New()
	hostC := uuid.New()
	hostD := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostC, "Charlie")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostD, "Delta")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + hostidentity.HostsViewSchemaID + "/query"
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
	if len(pageTwoRows) != 1 || pageTwoRows[0]["record_id"] != hostB.String() {
		t.Fatalf("expected live continuation to include inserted host B as second row, got %#v", pageTwoRows)
	}
	if got := pageTwoRows[0]["cells"].(map[string]any)["host.display_name"].(map[string]any)["value"]; got != "Bravo" {
		t.Fatalf("expected live display name Bravo, got %#v", pageTwoRows[0])
	}

	pageThree := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": responsePaging(pageTwo)["next_cursor"].(string),
		"sort":         sortByName,
	})
	pageThreeRows := responseRows(pageThree)
	if len(pageThreeRows) != 1 || pageThreeRows[0]["record_id"] != hostD.String() {
		t.Fatalf("expected live continuation to return host D as third row, got %#v", pageThreeRows)
	}
	if responsePaging(pageThree)["has_more"] != true {
		t.Fatalf("expected cursor chain to continue to updated host C, got %#v", responsePaging(pageThree))
	}

	pageFour := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"cursor_token": responsePaging(pageThree)["next_cursor"].(string),
		"sort":         sortByName,
	})
	pageFourRows := responseRows(pageFour)
	if len(pageFourRows) != 1 || pageFourRows[0]["record_id"] != hostC.String() {
		t.Fatalf("expected live continuation to return updated host C as final row, got %#v", pageFourRows)
	}
	if responsePaging(pageFour)["has_more"] != false {
		t.Fatalf("expected cursor chain to terminate, got %#v", responsePaging(pageFour))
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

func TestWorkbook_QueryCursorRejectsTampering(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-query-cursor-tamper")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-query-cursor-tamper-incident",
		"incident_key":  "IR-WORKBOOK-CURSOR-TAMPER",
		"title":         "Workbook query cursor tamper",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	hostA := uuid.New()
	hostB := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + hostidentity.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"limit": 1,
		"sort":  sortByName,
	})
	cursor := responsePaging(pageOne)["next_cursor"].(string)

	tamperedCursor := tamperCursor(t, cursor)
	resp := appsupport.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": tamperedCursor,
		"sort":         sortByName,
	}, appsupport.WithCookies(adminLogin.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "invalid_cursor_token" {
		t.Fatalf("expected invalid_cursor_token, got %#v", details)
	}

	fresh := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"sort": sortByName})
	if rows := responseRows(fresh); len(rows) != 2 {
		t.Fatalf("expected fresh query after expiry to succeed, got %#v", rows)
	}
}

func TestCoordinationProjectionQueries_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "workbook-seeded-surfaces")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-seeded-surfaces-incident",
		"incident_key":  "IR-WORKBOOK-SEEDED",
		"title":         "Workbook seeded surface query",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	seeds := []struct {
		viewSchemaID string
		recordType   string
		createBody   func() map[string]any
		insertChild  func(recordID uuid.UUID)
		filter       map[string]any
		sortField    string
		groupBy      string
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
				execSeed(t, harness, `
INSERT INTO evidence_grid_projection (
    record_id,
    incident_id,
    row_version,
    title,
    lifecycle_state,
    requested_at,
    received_at,
    storage_ref,
    blob_hash,
    collector_party_text,
    collector_party_id,
    source_party_text,
    source_party_id,
    upload_state,
    linked_record_count,
    edited_at
)
SELECT
    e.record_id,
    e.incident_id,
    r.row_version,
    e.title,
    e.lifecycle_state,
    e.requested_at,
    e.received_at,
    e.storage_ref,
    e.blob_hash,
    e.collector_party_text,
    e.collector_party_id,
    e.source_party_text,
    e.source_party_id,
    e.upload_state,
    0,
    e.updated_at
  FROM evidence e
  JOIN records r
    ON r.incident_id = e.incident_id
   AND r.record_id = e.record_id
   AND r.deleted_at IS NULL
 WHERE e.record_id = $1
`, recordID)
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
				execSeed(t, harness, `
INSERT INTO artifact_grid_projection (
    record_id,
    incident_id,
    row_version,
    artifact_type,
    title,
    body,
    updated_at,
    created_at,
    created_by_user_id,
    timestamp_day,
    next_report_day,
    ack_state,
    linked_record_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.artifact_type,
    a.title,
    a.body,
    a.updated_at,
    a.created_at,
    a.created_by_user_id,
    a.timestamp_utc::date,
    a.next_report_at::date,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END,
    0
  FROM artifacts a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
 WHERE a.record_id = $1
`, recordID)
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
				envelopetest.SeedRecordEnvelope(t, harness.DB, incidentID, adminUserID, subjectID, "host")
				execSeed(t, harness, `
INSERT INTO hosts (
    record_id, incident_id, display_name, host_state, row_version,
    created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, 'Assessment subject', 'canonical', r.row_version,
       r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, subjectID)
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
				seedTaskRequestProjection(t, harness, recordID)
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
				seedDecisionProjection(t, harness, recordID)
			},
			filter:    prefixFilter("decision.status", "app"),
			wantField: "decision.summary",
			wantValue: "Contain workstation",
		},
		{
			viewSchemaID: "cartulary.view.parties.v1",
			recordType:   "party",
			createBody: func() map[string]any {
				return map[string]any{
					"client_txn_id":      "txn-workbook-query-parties-create",
					"party.display_name": "Acme Legal",
					"party.party_kind":   "organization",
				}
			},
			filter:    prefixFilter("party.display_name", "acme"),
			wantField: "party.party_kind",
			wantValue: "organization",
		},
		{
			viewSchemaID: "cartulary.view.comm_log.v1",
			recordType:   "artifact",
			createBody: func() map[string]any {
				return map[string]any{
					"client_txn_id":               "txn-workbook-query-comm-create",
					"comm_log.comm_type":          "briefing",
					"comm_log.audience":           "leadership",
					"comm_log.channel_or_meeting": "Bridge",
					"comm_log.summary":            "Initial update",
				}
			},
			filter:    prefixFilter("comm_log.comm_type", "brief"),
			sortField: "comm_log.timestamp_day",
			groupBy:   "comm_log.comm_type",
			wantField: "comm_log.summary",
			wantValue: "Initial update",
		},
		{
			viewSchemaID: "cartulary.view.handoff.v1",
			recordType:   "artifact",
			createBody: func() map[string]any {
				return map[string]any{
					"client_txn_id":                  "txn-workbook-query-handoff-create",
					"handoff.incoming_owner_user_id": adminUserID.String(),
					"handoff.current_state_summary":  "Night shift takes containment",
				}
			},
			filter:    eqFilter("handoff.ack_state", "pending"),
			sortField: "handoff.timestamp_day",
			groupBy:   "handoff.incoming_owner_user_id",
			wantField: "handoff.current_state_summary",
			wantValue: "Night shift takes containment",
		},
		{
			viewSchemaID: "cartulary.view.status_review.v1",
			recordType:   "artifact",
			createBody: func() map[string]any {
				return map[string]any{
					"client_txn_id":                       "txn-workbook-query-status-create",
					"status_review.current_state_summary": "Containment in progress",
					"status_review.timestamp_utc":         "2026-04-24T15:00:00Z",
				}
			},
			filter:    map[string]any{"field_key": "status_review.timestamp_day", "op": "eq", "arg": map[string]any{"value": "2026-04-24"}},
			sortField: "status_review.next_report_day",
			groupBy:   "status_review.review_owner_user_id",
			wantField: "status_review.current_state_summary",
			wantValue: "Containment in progress",
		},
		{
			viewSchemaID: "cartulary.view.lesson.v1",
			recordType:   "artifact",
			createBody: func() map[string]any {
				return map[string]any{
					"client_txn_id":  "txn-workbook-query-lesson-create",
					"lesson.summary": "Preserve VPN logs earlier",
				}
			},
			filter:    prefixFilter("lesson.closure_state", "op"),
			sortField: "lesson.timestamp_day",
			groupBy:   "lesson.closure_state",
			wantField: "lesson.summary",
			wantValue: "Preserve VPN logs earlier",
		},
	}

	for _, seed := range seeds {
		t.Run(seed.viewSchemaID, func(t *testing.T) {
			recordID := uuid.New()
			if seed.createBody == nil {
				seedRecordEnvelope(t, harness, incidentID, adminUserID, recordID, seed.recordType)
				seed.insertChild(recordID)
			} else {
				resp := appsupport.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+seed.viewSchemaID+"/rows",
					seed.createBody(),
					appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
					appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
				)
				data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
				recordID = appsupport.MustUUID(t, data["row"].(map[string]any)["record_id"].(string))
			}

			queryBody := map[string]any{"filters": []map[string]any{seed.filter}}
			if seed.sortField != "" {
				queryBody["sort"] = []map[string]any{{"field_key": seed.sortField, "direction": "asc"}}
			}
			if seed.groupBy != "" {
				queryBody["group_by"] = seed.groupBy
			}
			resp := appsupport.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+seed.viewSchemaID+"/query",
				queryBody,
				appsupport.WithCookies(adminLogin.SessionCookie),
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
			if seed.groupBy != "" {
				groupValues := row["group_values"].(map[string]any)
				if _, ok := groupValues[seed.groupBy]; !ok {
					t.Fatalf("expected group_values to include %s, got %#v", seed.groupBy, groupValues)
				}
			}
		})
	}
}

func seedRecordEnvelope(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorID)
}

func execSeed(t testing.TB, harness *appsupport.ServerHarness, sql string, args ...any) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), sql, args...); err != nil {
		t.Fatalf("seed query failed: %v", err)
	}
}

func seedTaskRequestProjection(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO task_request_grid_projection (
    record_id,
    incident_id,
    row_version,
    title,
    status,
    owner_user_id,
    priority,
    task_kind,
    workstream,
    due_at,
    requester_party_text,
    requester_party_id,
    blocked_reason,
    completed_at,
    external_ticket_ref,
    closure_summary,
    decision_record_id,
    linked_record_count,
    updated_at,
    no_owner
)
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.title,
    t.status,
    t.owner_user_id,
    t.priority,
    t.task_kind,
    t.workstream,
    t.due_at,
    t.requester_party_text,
    t.requester_party_id,
    t.blocked_reason,
    t.completed_at,
    t.external_ticket_ref,
    t.closure_summary,
    t.decision_record_id,
    0,
    t.updated_at,
    t.owner_user_id IS NULL
  FROM task_requests t
  JOIN records r
    ON r.incident_id = t.incident_id
   AND r.record_id = t.record_id
   AND r.deleted_at IS NULL
 WHERE t.record_id = $1
ON CONFLICT (record_id) DO UPDATE
SET row_version = EXCLUDED.row_version,
    title = EXCLUDED.title,
    status = EXCLUDED.status,
    owner_user_id = EXCLUDED.owner_user_id,
    priority = EXCLUDED.priority,
    task_kind = EXCLUDED.task_kind,
    workstream = EXCLUDED.workstream,
    due_at = EXCLUDED.due_at,
    requester_party_text = EXCLUDED.requester_party_text,
    requester_party_id = EXCLUDED.requester_party_id,
    blocked_reason = EXCLUDED.blocked_reason,
    completed_at = EXCLUDED.completed_at,
    external_ticket_ref = EXCLUDED.external_ticket_ref,
    closure_summary = EXCLUDED.closure_summary,
    decision_record_id = EXCLUDED.decision_record_id,
    linked_record_count = EXCLUDED.linked_record_count,
    updated_at = EXCLUDED.updated_at,
    no_owner = EXCLUDED.no_owner
`, recordID)
}

func seedDecisionProjection(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO decision_grid_projection (
    record_id,
    incident_id,
    row_version,
    summary,
    status,
    owner_user_id,
    decision_type,
    decided_at,
    rationale,
    affected_record_count,
    supersedes_record_id,
    updated_at,
    is_superseded
)
SELECT
    d.record_id,
    d.incident_id,
    r.row_version,
    d.summary,
    d.status,
    d.owner_user_id,
    d.decision_type,
    d.decided_at,
    d.rationale,
    0,
    supersedes.supersedes_record_id,
    d.updated_at,
    COALESCE(incoming.is_superseded, false)
  FROM decisions d
  JOIN records r
    ON r.incident_id = d.incident_id
   AND r.record_id = d.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN LATERAL (
        SELECT rl.dst_record_id AS supersedes_record_id
          FROM record_links rl
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.record_type = 'decision'
           AND dst.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.src_record_id = d.record_id
           AND rl.link_type = 'supersedes'
           AND rl.deleted_at IS NULL
         ORDER BY rl.created_at DESC, rl.record_link_id DESC
         LIMIT 1
  ) supersedes ON true
  LEFT JOIN LATERAL (
        SELECT true AS is_superseded
          FROM record_links rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.record_type = 'decision'
           AND src.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.dst_record_id = d.record_id
           AND rl.link_type = 'supersedes'
           AND rl.deleted_at IS NULL
         LIMIT 1
  ) incoming ON true
 WHERE d.record_id = $1
ON CONFLICT (record_id) DO UPDATE
SET row_version = EXCLUDED.row_version,
    summary = EXCLUDED.summary,
    status = EXCLUDED.status,
    owner_user_id = EXCLUDED.owner_user_id,
    decision_type = EXCLUDED.decision_type,
    decided_at = EXCLUDED.decided_at,
    rationale = EXCLUDED.rationale,
    affected_record_count = EXCLUDED.affected_record_count,
    supersedes_record_id = EXCLUDED.supersedes_record_id,
    updated_at = EXCLUDED.updated_at,
    is_superseded = EXCLUDED.is_superseded
`, recordID)
}

func eqFilter(fieldKey string, value any) map[string]any {
	return map[string]any{"field_key": fieldKey, "op": "eq", "arg": map[string]any{"value": value}}
}

func prefixFilter(fieldKey string, value string) map[string]any {
	return map[string]any{"field_key": fieldKey, "op": "prefix", "arg": map[string]any{"value": value}}
}

func seedHostForPaging(t testing.TB, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string) {
	t.Helper()
	execSeed(t, harness, `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'host', $3, $3)
`, recordID, incidentID, actorID)
	execSeed(t, harness, `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state, row_version,
    created_at, updated_at, created_by_user_id, updated_by_user_id
)
SELECT r.record_id, r.incident_id, $2, lower($2), 'canonical', r.row_version,
       r.created_at, r.updated_at, r.created_by_user_id, r.updated_by_user_id
  FROM records r
 WHERE r.record_id = $1
`, recordID, displayName)
	execSeed(t, harness, `
INSERT INTO host_grid_projection (record_id, incident_id, row_version, display_name, hostname, host_state, edited_at)
VALUES ($1, $2, 1, $3, lower($3), 'canonical', now())
`, recordID, incidentID, displayName)
}

func updateHostDisplayNameForPaging(t testing.TB, harness *appsupport.ServerHarness, recordID uuid.UUID, displayName string, rowVersion int64) {
	t.Helper()
	ctx := context.Background()
	tx, err := harness.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin host paging update: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, `
UPDATE records
   SET row_version = $2,
       updated_at = now()
 WHERE record_id = $1
`, recordID, rowVersion); err != nil {
		t.Fatalf("advance host paging envelope: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE hosts h
   SET display_name = $2,
       row_version = r.row_version,
       updated_at = r.updated_at,
       updated_by_user_id = r.updated_by_user_id
  FROM records r
 WHERE h.record_id = $1
   AND r.record_id = h.record_id
`, recordID, displayName); err != nil {
		t.Fatalf("update host paging source: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit host paging update: %v", err)
	}
	execSeed(t, harness, `
UPDATE host_grid_projection
   SET display_name = $2,
       row_version = $3,
       edited_at = now()
 WHERE record_id = $1
`, recordID, displayName, rowVersion)
}

func queryWorkbook(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, queryURL string, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, queryURL, body, appsupport.WithCookies(login.SessionCookie))
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

func requireDefaultVisibleCells(t testing.TB, viewSchemaID string, row map[string]any) {
	t.Helper()
	resource, ok := viewschema.LookupPublicResource(viewSchemaID)
	if !ok {
		t.Fatalf("missing public view schema resource %s", viewSchemaID)
	}
	cells := row["cells"].(map[string]any)
	for _, field := range resource.Fields {
		if field.DefaultHidden {
			continue
		}
		if _, ok := cells[field.FieldKey]; !ok {
			t.Fatalf("%s default-visible field %s missing from row %#v", viewSchemaID, field.FieldKey, row)
		}
	}
}
