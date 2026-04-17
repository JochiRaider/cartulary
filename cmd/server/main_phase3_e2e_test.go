package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"example.com/todo/cartulary/internal/modules/timeline"
	"example.com/todo/cartulary/internal/platform/authn"
	platformws "example.com/todo/cartulary/internal/platform/ws"
	"example.com/todo/cartulary/internal/testutil/httptestx"
)

func TestPhase3_TimelineLifecycle_E_3_03(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase3-e-3-03")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	incident := phase3CreateIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-e-3-03-incident",
		"incident_key":  "IR-E303",
		"title":         "Phase 3 lifecycle",
	})
	incidentID := incident["incident_id"].(string)

	replacement := phase3CreateTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":    "txn-e-3-03-replacement",
		"timeline.summary": "Replacement row",
	})
	replacementID := replacement["row"].(map[string]any)["record_id"].(string)
	primary := phase3CreateTimelineRow(t, server, incidentID, adminLogin, map[string]any{
		"client_txn_id":    "txn-e-3-03-primary",
		"timeline.summary": "Primary row",
	})
	recordID := primary["row"].(map[string]any)["record_id"].(string)

	reviewed := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/records/"+recordID+"/mark-reviewed",
		map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-e-3-03-reviewed-1",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	reviewedData := httptestx.RequireSuccessEnvelope(t, reviewed, http.StatusOK)["data"].(map[string]any)
	if reviewedData["capture_state"] != "reviewed" {
		t.Fatalf("expected reviewed state, got %#v", reviewedData)
	}

	demoted := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 2,
			"client_txn_id":    "txn-e-3-03-demote",
			"changes": []map[string]any{
				{"field_key": "timeline.details", "value": "Material edit after review"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	demotedRow := httptestx.RequireSuccessEnvelope(t, demoted, http.StatusOK)["data"].(map[string]any)["row"].(map[string]any)
	if got := demotedRow["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("expected reviewed row to demote back to enriched, got %#v", demotedRow)
	}

	superseded := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/records/"+recordID+"/supersede",
		map[string]any{
			"base_row_version":      3,
			"client_txn_id":         "txn-e-3-03-supersede",
			"reason":                "Superseded by replacement",
			"replacement_record_id": replacementID,
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	supersededData := httptestx.RequireSuccessEnvelope(t, superseded, http.StatusOK)["data"].(map[string]any)
	if supersededData["capture_state"] != "superseded" || supersededData["replacement_record_id"] != replacementID {
		t.Fatalf("unexpected supersede payload: %#v", supersededData)
	}

	query := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
		map[string]any{},
		withCookies(adminLogin.sessionCookie),
	)
	rows := httptestx.RequireSuccessEnvelope(t, query, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	found := false
	for _, row := range rows {
		current := row.(map[string]any)
		if current["record_id"] == recordID {
			found = true
			cells := current["cells"].(map[string]any)
			if cells["timeline.capture_state"].(map[string]any)["value"] != "superseded" {
				t.Fatalf("expected queried row to be superseded, got %#v", current)
			}
			if cells["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID {
				t.Fatalf("expected queried row replacement id, got %#v", current)
			}
		}
	}
	if !found {
		t.Fatalf("expected superseded row in query result, got %#v", rows)
	}
}

func TestPhase3_RecordChangeReplayDedupe_E_3_04(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase3-e-3-04")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)

	incident := phase3CreateIncident(t, server, adminLogin, map[string]any{
		"client_txn_id": "txn-e-3-04-incident",
		"incident_key":  "IR-E304",
		"title":         "Phase 3 replay",
	})
	incidentID := incident["incident_id"].(string)

	createResp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":    "txn-e-3-04-create",
			"timeline.summary": "Socket row",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	recordID := createData["row"].(map[string]any)["record_id"].(string)
	createEvents := phase3RecordChangeSnapshot(t, server)
	if len(createEvents) != 1 || createEvents[0]["record_id"] != recordID {
		t.Fatalf("unexpected create snapshot payload: %#v", createEvents)
	}

	createReplay := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":    "txn-e-3-04-create",
			"timeline.summary": "Socket row",
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, createReplay, http.StatusOK)
	if replayEvents := phase3RecordChangeSnapshot(t, server); len(replayEvents) != 1 {
		t.Fatalf("expected no duplicate create emission, got %#v", replayEvents)
	}

	patchResp := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-e-3-04-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "Patched socket row"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)
	patchEvents := phase3RecordChangeSnapshot(t, server)
	if len(patchEvents) != 2 || patchEvents[1]["record_id"] != recordID || patchEvents[1]["row_version"] != float64(2) {
		t.Fatalf("unexpected patch snapshot payload: %#v", patchEvents)
	}

	patchReplay := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/records/"+recordID,
		map[string]any{
			"view_schema_id":   timeline.TimelineViewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-e-3-04-patch",
			"changes": []map[string]any{
				{"field_key": "timeline.summary", "value": "Patched socket row"},
			},
		},
		withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchReplay, http.StatusOK)
	if replayEvents := phase3RecordChangeSnapshot(t, server); len(replayEvents) != 2 {
		t.Fatalf("expected no duplicate patch emission, got %#v", replayEvents)
	}
}

func phase3CreateIncident(t testing.TB, server *phase0ServerProcess, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func phase3CreateTimelineRow(t testing.TB, server *phase0ServerProcess, incidentID string, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func phase3RecordChangeSnapshot(t testing.TB, server *phase0ServerProcess) []map[string]any {
	t.Helper()

	resp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/test/timeline/record-changes", nil)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	rawItems := body["record_changes"].([]any)
	items := make([]map[string]any, 0, len(rawItems))
	for _, item := range rawItems {
		items = append(items, item.(map[string]any))
	}
	return items
}

func phase3ConnectRecordChangeSocket(t testing.TB, server *phase0ServerProcess) *websocket.Conn {
	t.Helper()

	target, err := url.Parse(phase1ServerURL(server))
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	target.Scheme = strings.Replace(target.Scheme, "http", "ws", 1)
	target.Path = "/ws/v1/test/record-changes"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, target.String(), nil)
	if err != nil {
		t.Fatalf("dial record-change websocket: %v", err)
	}

	var connected platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &connected); err != nil {
		t.Fatalf("read record-change connected message: %v", err)
	}
	if connected.Type != "connected" {
		t.Fatalf("unexpected record-change socket message: %#v", connected)
	}
	return conn
}

func phase3ExpectRecordChanged(t testing.TB, conn *websocket.Conn) map[string]any {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &message); err != nil {
		t.Fatalf("read record_changed: %v", err)
	}
	if message.Type != "record_changed" {
		t.Fatalf("unexpected websocket message: %#v", message)
	}
	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("decode record_changed payload: %v", err)
	}
	return payload
}

func phase3ExpectNoRecordChanged(t testing.TB, conn *websocket.Conn) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()

	var message platformws.Message
	err := platformws.ReadJSON(ctx, conn, &message)
	if err == nil {
		t.Fatalf("expected no websocket message, got %#v", message)
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected read timeout for no replay emission, got %v", err)
	}
}
