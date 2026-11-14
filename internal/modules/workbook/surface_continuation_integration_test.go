package workbook_test

import (
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookSurfaceContinuation_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "workbook-surface-continuation", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	// Independently authored route and default-order expectations. Do not derive
	// this matrix from the dispatcher or schema registry being exercised.
	for _, tc := range []struct {
		name      string
		field     string
		minimum   map[string]any
		sortField string
		direction string
	}{
		{"assessments", "assessment.rationale", map[string]any{"assessment.subject_type": "host", "assessment.assessment_state": "confirmed", "assessment.confidence_score": 85, "assessment.assessed_at": "2026-09-10T00:00:00Z"}, "assessment.assessed_at", "desc"},
		{"comm_log", "comm_log.summary", map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "Operations", "comm_log.channel_or_meeting": "Bridge"}, "comm_log.timestamp_utc", "desc"},
		{"decisions", "decision.summary", map[string]any{"decision.decision_type": "scope", "decision.rationale": "Continuation fixture"}, "decision.decided_at", "desc"},
		{"evidence", "evidence.title", map[string]any{"evidence.collector_party_text": "Continuation collector"}, "evidence.requested_at", "desc"},
		{"findings", "finding.statement", map[string]any{}, "finding.updated_at", "desc"},
		{"forensic_keywords", "forensic_keyword.pattern", map[string]any{"forensic_keyword.reason": "Continuation fixture"}, "forensic_keyword.created_at", "desc"},
		{"handoff", "handoff.current_state_summary", map[string]any{"handoff.incoming_owner_user_id": actorID.String()}, "handoff.timestamp_utc", "desc"},
		{"hosts", "host.display_name", map[string]any{}, "host.display_name", "asc"},
		{"identities", "identity.display_name", map[string]any{}, "identity.display_name", "asc"},
		{"indicators", "indicator.display_value", map[string]any{"indicator.indicator_type": "domain_name", "indicator.value_kind": "atomic"}, "indicator.last_observed_at", "desc"},
		{"investigative_queries", "investigative_query.purpose", map[string]any{"investigative_query.platform": "SQL", "investigative_query.query_text": "select value"}, "investigative_query.created_at", "desc"},
		{"lesson", "lesson.summary", map[string]any{}, "lesson.timestamp_utc", "desc"},
		{"notes", "note.title", map[string]any{"note.body": "Continuation fixture"}, "note.updated_at", "desc"},
		{"parties", "party.display_name", map[string]any{"party.party_kind": "person"}, "party.display_name", "asc"},
		{"status_review", "status_review.current_state_summary", map[string]any{}, "status_review.timestamp_utc", "desc"},
		{"task_requests", "task.title", map[string]any{"task.task_kind": "question"}, "task.updated_at", "desc"},
		{"timeline", "timeline.activity_synopsis_text", map[string]any{}, "timeline.activity_sort_ts", "asc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
				"client_txn_id": "continuation-incident-" + tc.name,
				"incident_key":  "CONTINUATION-" + tc.name,
				"title":         "Continuation " + tc.name,
			})
			incidentID := incident["incident_id"].(string)
			view := "cartulary.view." + tc.name + ".v1"
			if tc.name == "timeline" {
				view = "cartulary.view.timeline.v2"
			}
			creates := 0
			create := func(schema string, fields map[string]any) string {
				t.Helper()
				creates++
				body := maps.Clone(fields)
				body["client_txn_id"] = fmt.Sprintf("continuation-%s-%d", tc.name, creates)
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+schema+"/rows", body,
					httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
				data := httptestx.RequireSuccessData(t, resp, http.StatusCreated)
				row := data["row"].(map[string]any)
				id := row["record_id"].(string)
				if parsed, err := uuid.Parse(id); err != nil || parsed == uuid.Nil || row["row_version"] != float64(1) {
					t.Fatalf("new fixture must have a valid identity and version 1: %#v", row)
				}
				return id
			}
			minimum := maps.Clone(tc.minimum)
			// Auxiliary records stay in this case's incident, so they cannot
			// contaminate the Hosts or Parties cases' exact three-record set.
			if tc.name == "assessments" {
				minimum["assessment.subject_ref"] = create("cartulary.view.hosts.v1", map[string]any{"host.hostname": "continuation-subject.example.test"})
			}
			if tc.name == "evidence" {
				minimum["evidence.collector_party_id"] = create("cartulary.view.parties.v1", map[string]any{"party.party_kind": "organization", "party.display_name": "Continuation collector"})
			}
			wantValues := map[string]string{}
			for index := 1; index <= 3; index++ {
				fields := maps.Clone(minimum)
				value := fmt.Sprintf("Continuation %03d", index)
				switch tc.name {
				case "hosts":
					fields["host.hostname"] = fmt.Sprintf("continuation-%d.example.test", index)
				case "identities":
					fields["identity.email"] = fmt.Sprintf("continuation-%d@example.test", index)
				case "indicators":
					value = fmt.Sprintf("continuation-%d.example.test", index)
				}
				fields[tc.field] = value
				id := create(view, fields)
				if _, duplicate := wantValues[id]; duplicate {
					t.Fatalf("distinct fixture creates reused record %s", id)
				}
				wantValues[id] = value
			}
			wantSort := []any{map[string]any{"field_key": tc.sortField, "direction": tc.direction}}
			if tc.name == "indicators" {
				wantSort = append(wantSort, map[string]any{"field_key": "indicator.display_value", "direction": "asc"})
			}
			wantSort = append(wantSort, map[string]any{"field_key": "record_id", "direction": "asc"})
			wantQuery := map[string]any{"filters": []any{}, "sort": wantSort}
			queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/views/" + view + "/query"
			query := func(request map[string]any, count int, limit int, more bool) map[string]any {
				t.Helper()
				envelope := queryWorkbook(t, harness, login, queryURL, request)
				data := envelope["data"].(map[string]any)
				if data["incident_id"] != incidentID || data["view_schema_id"] != view {
					t.Fatalf("query returned wrong route identity: %#v", data)
				}
				if got := envelope["meta"].(map[string]any)["query"]; !reflect.DeepEqual(got, wantQuery) {
					t.Fatalf("canonical query: got %#v want %#v", got, wantQuery)
				}
				paging := responsePaging(envelope)
				if paging["limit"] != float64(limit) || paging["has_more"] != more {
					t.Fatalf("paging: got %#v want limit=%d has_more=%v", paging, limit, more)
				}
				if more {
					if cursor, ok := paging["next_cursor"].(string); !ok || cursor == "" {
						t.Fatalf("nonterminal page needs an opaque cursor: %#v", paging)
					}
				} else if cursor, present := paging["next_cursor"]; !present || cursor != nil {
					t.Fatalf("terminal page needs explicit null cursor: %#v", paging)
				}
				rows := responseRows(envelope)
				if len(rows) != count {
					t.Fatalf("got %d rows, want %d", len(rows), count)
				}
				seen := map[string]bool{}
				for _, row := range rows {
					id := row["record_id"].(string)
					want, exists := wantValues[id]
					if !exists || seen[id] || row["row_version"] != float64(1) {
						t.Fatalf("unexpected, duplicate or stale row: %#v", row)
					}
					seen[id] = true
					requireCellValue(t, row, tc.field, want)
				}
				return envelope
			}
			baseline := query(map[string]any{}, 3, 100, false)
			first := query(map[string]any{"limit": 2}, 2, 2, true)
			terminal := query(map[string]any{"cursor_token": responsePaging(first)["next_cursor"]}, 1, 2, false)
			gotIDs := append(rowIDs(responseRows(first)), rowIDs(responseRows(terminal))...)
			// No intervening writes: the full cursor chain must preserve the
			// default query's exact order and set, not assume creation order.
			if wantIDs := rowIDs(responseRows(baseline)); !reflect.DeepEqual(gotIDs, wantIDs) {
				t.Fatalf("continuation lost, repeated or reordered records: got %v want %v", gotIDs, wantIDs)
			}
		})
	}
}
