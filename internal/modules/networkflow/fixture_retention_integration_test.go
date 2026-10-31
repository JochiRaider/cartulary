package networkflow_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
)

// The frozen comparison owns the expected status/code/details. This loader
// retains unexpected error variants instead of aborting before their comparison.
func (r *publicationFixtureRuntime) observedError(method, path string, body any) map[string]any {
	response := httptestx.DoJSON(r.t, http.DefaultClient, method, r.h.Server.HTTP.URL+path, body, r.requestOptions...)
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		r.t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode < 400 {
		r.t.Fatalf("expected rejected fixture operation, got %d", response.StatusCode)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		r.t.Fatal(err)
	}
	response.Body = io.NopCloser(bytes.NewReader(raw))
	value := httptestx.RequireErrorEnvelope(r.t, response, response.StatusCode, envelope.Error.Code)
	r.transport = append(r.transport, map[string]any{"method": method, "path": path, "body": body, "status": response.StatusCode, "response": fixtureResponseEvidence(raw)})
	failure := value["error"].(map[string]any)
	failure["request_id"] = "$request-id"
	return failure
}
func (r *publicationFixtureRuntime) retentionState() map[string]any {
	owner := r.durableState()
	bindings := r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY network_flow_indicator_binding_id),'[]') FROM network_flow_indicator_bindings b`)
	coreAudit := r.jsonQuery(`SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY created_at,id),'[]') FROM deployment_admin_audit_events a WHERE event_source='incidents' AND event_kind='incident_close'`)
	r.transport = append(r.transport, map[string]any{"retained_owner_state": owner, "retained_bindings": bindings, "retained_core_audit": coreAudit})
	return map[string]any{"network_flow": r.comparableState(owner), "bindings": bindings, "core_audit": r.comparableState(map[string]any{"audit": coreAudit})["audit"], "counts": r.jsonQuery(`SELECT jsonb_build_object('active_tables',(SELECT count(*) FROM network_flow_tables WHERE table_status='active'),'retained_tables',(SELECT count(*) FROM network_flow_tables),'rows',(SELECT count(*) FROM network_flow_rows),'diagnostics',(SELECT count(*) FROM network_flow_rejected_row_diagnostics),'bindings',(SELECT count(*) FROM network_flow_indicator_bindings),'link_receipts',(SELECT count(*) FROM route_idempotency WHERE route_key='nf.indicator_links.create'),'delete_receipts',(SELECT count(*) FROM route_idempotency WHERE route_key='nf.tables.delete'),'close_receipts',(SELECT count(*) FROM route_idempotency WHERE client_txn_id='fixture-close'))`)}
}

// Random binding IDs and Core wall-clock times are retained verbatim in transport,
// validated independently, and aliased only for the frozen semantic comparison.
func retentionAliases(t *testing.T, value any, bindingID string, closedAt time.Time) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	var visit func(any) any
	visit = func(v any) any {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				x[k] = visit(v)
			}
		case []any:
			for i, v := range x {
				x[i] = visit(v)
			}
		case string:
			if x == bindingID {
				return "$retained-binding-id"
			}
			if !closedAt.IsZero() {
				if ts, err := time.Parse(time.RFC3339Nano, x); err == nil && ts.Equal(closedAt) {
					return "$incident-close-time"
				}
			}
		}
		return v
	}
	return visit(result).(map[string]any)
}
func TestNetworkFlowRetentionFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-027-retention-soft-delete", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runPublicationFamily(t, bundle, func(r *publicationFixtureRuntime, input publicationFixtureCase, result map[string]any) error {
			root := "/api/v1/incidents/" + input.IncidentID
			nfroot := root + "/network-flow"
			tableID := "nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
			path := nfroot + "/tables/" + tableID
			rows := result["rows"].(map[string]any)["rows"].([]any)
			rowID := rows[0].(map[string]any)["network_flow_row_id"].(string)
			targetID := uuid.MustParse("00000000-0000-4000-8000-000000000090")
			indicatortest.SeedRecord(r.t, r.h.DB, uuid.MustParse(input.IncidentID), uuid.MustParse(input.ActorID), targetID, "ipv4_addr", "atomic", "192.0.2.10")
			linkBody := map[string]any{"schema_id": "cartulary.network_flow.indicator_link_request.v1", "client_txn_id": "fixture-retained-link", "selector": map[string]any{"kind": "row_field_value", "network_flow_table_id": tableID, "network_flow_row_id": rowID, "field_key": "network_flow.src_ip"}, "target": map[string]any{"mode": "existing_indicator", "indicator_id": targetID.String()}, "observation_mode": "binding_only", "confirm_exact_value": "192.0.2.10"}
			link := r.request(http.MethodPost, nfroot+"/indicator-links", linkBody, http.StatusCreated)
			bindingID := link["binding"].(map[string]any)["network_flow_indicator_binding_id"].(string)
			if !strings.HasPrefix(bindingID, "nfb_") || len(bindingID) != 36 {
				return fmt.Errorf("invalid binding identity")
			}
			if _, err := hex.DecodeString(bindingID[4:]); err != nil {
				return err
			}
			if replay := r.request(http.MethodPost, nfroot+"/indicator-links", linkBody, http.StatusCreated); !reflect.DeepEqual(link, replay) {
				return fmt.Errorf("link replay changed exact response")
			}
			rowPage, rowToken, rowPayload := r.normalizedPage(r.request(http.MethodPost, path+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1", "limit": 1}, http.StatusOK), "retention-row")
			diagPage, diagToken, diagPayload := r.normalizedPage(r.request(http.MethodPost, path+"/rejected-rows/query", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_request.v1", "limit": 1}, http.StatusOK), "retention-diagnostic")
			if rowToken == "" || diagToken == "" {
				return fmt.Errorf("retention requires both real cursor variants")
			}
			state := r.retentionState()
			restore := r.alternateActor(input.IncidentID)
			denied := map[string]any{"rename": r.observedError(http.MethodPatch, path, map[string]any{"client_txn_id": "denied-rename", "base_table_version": 1, "display_name": "Forbidden"}), "delete": r.observedError(http.MethodDelete, path, map[string]any{"client_txn_id": "denied-delete", "base_table_version": 1})}
			restore()
			if !reflect.DeepEqual(state, r.retentionState()) {
				return fmt.Errorf("unauthorized lifecycle operations mutated owner state")
			}
			retention := map[string]any{"link": link, "row_page": rowPage, "row_cursor": rowPayload, "diagnostic_page": diagPage, "diagnostic_cursor": diagPayload, "before": state, "denied": denied}
			var closedAt time.Time
			if input.Name == "soft-delete" {
				body := map[string]any{"client_txn_id": "fixture-delete", "base_table_version": 1}
				deletion := r.request(http.MethodDelete, path, body, http.StatusOK)
				retention["mutation"] = deletion
				state = r.retentionState()
				if replay := r.request(http.MethodDelete, path, body, http.StatusOK); !reflect.DeepEqual(deletion, replay) || !reflect.DeepEqual(state, r.retentionState()) {
					return fmt.Errorf("delete replay changed retained state")
				}
				if replay := r.request(http.MethodPost, nfroot+"/indicator-links", linkBody, http.StatusCreated); !reflect.DeepEqual(link, replay) {
					return fmt.Errorf("retained link replay changed")
				}
				retention["active_list"] = r.request(http.MethodGet, nfroot+"/tables", nil, http.StatusOK)
			} else if input.Name == "incident-close" {
				before := time.Now().UTC()
				body := map[string]any{"base_incident_version": 1, "client_txn_id": "fixture-close", "reason": "Fixture retention review complete."}
				closed := r.request(http.MethodPost, root+"/close", body, http.StatusOK)
				var err error
				closedAt, err = time.Parse(time.RFC3339Nano, closed["closed_at"].(string))
				if err != nil || closedAt.Before(before) || closedAt.After(time.Now().UTC().Add(time.Second)) {
					return fmt.Errorf("Core closure timestamp outside operation interval")
				}
				if closed["closed_at"] != closed["updated_at"] {
					return fmt.Errorf("Core closure timestamps differ")
				}
				retention["mutation"] = closed
				state = r.retentionState()
				if replay := r.request(http.MethodPost, root+"/close", body, http.StatusOK); !reflect.DeepEqual(closed, replay) || !reflect.DeepEqual(state, r.retentionState()) {
					return fmt.Errorf("Core close replay changed retained state")
				}
				retention["active_list"] = r.observedError(http.MethodGet, nfroot+"/tables", nil)
			} else {
				return fmt.Errorf("unknown retention case %s", input.Name)
			}
			freshLink := map[string]any{}
			for k, v := range linkBody {
				freshLink[k] = v
			}
			freshLink["client_txn_id"] = "fixture-link-after-transition"
			retention["rejected_operations"] = map[string]any{
				"table":                   r.observedError(http.MethodGet, path, nil),
				"rows":                    r.observedError(http.MethodPost, path+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1"}),
				"diagnostics":             r.observedError(http.MethodPost, path+"/rejected-rows/query", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_request.v1"}),
				"row_continuation":        r.observedError(http.MethodPost, path+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v1", "cursor_token": rowToken}),
				"diagnostic_continuation": r.observedError(http.MethodPost, path+"/rejected-rows/query", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_continuation.v1", "cursor_token": diagToken}),
				"link":                    r.observedError(http.MethodPost, nfroot+"/indicator-links", freshLink),
				"rename":                  r.observedError(http.MethodPatch, path, map[string]any{"client_txn_id": "fixture-rename-after-transition", "base_table_version": 2, "display_name": "Impossible"}),
			}
			final := r.retentionState()
			if !reflect.DeepEqual(state, final) {
				return fmt.Errorf("rejected operations changed retained owner effects")
			}
			retention["after"] = final
			retention["replay_unchanged"] = true
			result["retention"] = retentionAliases(r.t, retention, bindingID, closedAt)
			return nil
		})
	})
}
