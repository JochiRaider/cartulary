package workbook_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookTimelineMutationAdmission_Integration(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "timeline-mutation-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	type admissionCase struct {
		name        string
		createField string
		createValue any
		changes     []map[string]any
		baseVersion int
		wantStatus  int
		wantCode    string
		wantReason  string
	}
	var cases []admissionCase
	for _, field := range []struct {
		key   string
		value any
	}{
		{"record_id", nil}, // Bound to an existing record below.
		{"row_version", 1},
		{"timeline.evidence_count", 4},
		{"timeline.capture_state", "reviewed"},
		{"timeline.edited_at", "2026-04-10T10:15:00Z"},
	} {
		cases = append(cases, admissionCase{
			name: "create/" + field.key, createField: field.key, createValue: field.value,
			wantStatus: http.StatusBadRequest, wantCode: "invalid_mutation_payload", wantReason: "unknown_field",
		})
	}
	oversized := make([]map[string]any, 33)
	for index := range oversized {
		oversized[index] = map[string]any{"field_key": "timeline.activity_synopsis_text", "value": fmt.Sprintf("Over limit %d", index)}
	}
	cases = append(cases,
		admissionCase{name: "patch/empty", changes: []map[string]any{}, baseVersion: 1, wantStatus: http.StatusBadRequest, wantCode: "invalid_mutation_payload", wantReason: "empty_changes"},
		admissionCase{name: "patch/duplicate", changes: []map[string]any{
			{"field_key": "timeline.activity_synopsis_text", "value": "First"},
			{"field_key": "timeline.activity_synopsis_text", "value": "Second"},
		}, baseVersion: 1, wantStatus: http.StatusBadRequest, wantCode: "invalid_mutation_payload", wantReason: "duplicate_field_key"},
		admissionCase{name: "patch/over-limit", changes: oversized, baseVersion: 1, wantStatus: http.StatusBadRequest, wantCode: "invalid_mutation_payload", wantReason: "change_count_exceeded"},
		admissionCase{name: "patch/future-version", changes: []map[string]any{
			{"field_key": "timeline.activity_synopsis_text", "value": "Future version"},
		}, baseVersion: 101, wantStatus: http.StatusConflict, wantCode: "row_version_conflict"},
	)
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each rejection has its own incident, seed and transaction identity.
			incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{
				"client_txn_id": fmt.Sprintf("admission-incident-%d", index),
				"incident_key":  fmt.Sprintf("IR-TIMELINE-ADMISSION-%02d", index),
				"title":         "Timeline mutation admission " + tc.name,
			})
			incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
			seed := requireWorkbookCreate(t, h, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{
				"client_txn_id":                   fmt.Sprintf("admission-seed-%d", index),
				"timeline.activity_synopsis_text": "Unchanged seed",
			})
			recordID := appsupport.MustUUID(t, seed["row"].(map[string]any)["record_id"].(string))
			body := map[string]any{"client_txn_id": fmt.Sprintf("admission-rejected-%d", index)}
			method, expectedField := http.MethodPatch, "changes"
			if tc.createField != "" {
				method, expectedField = http.MethodPost, tc.createField
				body["timeline.activity_synopsis_text"] = "Must not create"
				body[tc.createField] = tc.createValue
				if tc.createField == "record_id" {
					body[tc.createField] = recordID.String()
				}
			} else {
				body["view_schema_id"] = timeline.TimelineViewSchemaID
				body["base_row_version"] = tc.baseVersion
				body["changes"] = tc.changes
			}
			before := snapshotWorkbookState(t, h, login, incidentID, timeline.TimelineViewSchemaID)
			response := doWorkbookJSON(t, h, login, method, incidentID, timeline.TimelineViewSchemaID, recordID, body)
			envelope := requireWorkbookMutationError(t, response, tc.wantStatus, tc.wantCode)
			if tc.wantReason != "" {
				details := httptestx.RequireErrorDetails(t, envelope)
				if details["field"] != expectedField || details["reason_code"] != tc.wantReason {
					t.Fatalf("unexpected admission details: %#v", details)
				}
				if tc.wantReason == "change_count_exceeded" && (details["requested_count"] != float64(33) || details["max_count"] != float64(32)) {
					t.Fatalf("expected raw change count 33 and limit 32: %#v", details)
				}
			}
			requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incidentID, timeline.TimelineViewSchemaID))
			requireTimelineSummaryAndVersion(t, h, recordID, "Unchanged seed", 1)
		})
	}
}

func requireWorkbookMutationError(t testing.TB, response *http.Response, status int, code string) map[string]any {
	t.Helper()
	envelope := httptestx.RequireErrorEnvelope(t, response, status, code)
	errorValue := envelope["error"].(map[string]any)
	for _, field := range []string{"stack", "trace", "internal"} {
		if _, present := errorValue[field]; present {
			t.Fatalf("public mutation error exposes %s", field)
		}
	}
	return envelope
}
