package networkflow_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

// The expectations were independently calculated from the original CSV and the
// adopted identity/aggregation algorithms, without running the product. Each
// persisted row is compared in full; raw provenance uses a sorted-key JSON hash
// to keep the reviewed expectation compact. Full durable bytes and transport
// are retained by the publication adapter.
func TestCiscoSNAFlowSearchExampleFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-029-cisco-sna-flow-search", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		data, err := bundle.Read("source/cases.json")
		if err != nil {
			return nil, err
		}
		var cases []publicationFixtureCase
		if err = json.Unmarshal(data, &cases); err != nil {
			return nil, err
		}
		if len(cases) != 1 {
			return nil, fmt.Errorf("expected one independent example")
		}
		input := cases[0]
		result, err := runPublicationFixture(t, bundle, input, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
			root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
			tableID := "nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
			result["table"] = r.request(http.MethodGet, root+"/tables/"+tableID, nil, http.StatusOK)
			first, token, payload := r.normalizedPage(result["rows"].(map[string]any), "first")
			if token == "" {
				return fmt.Errorf("example must exercise row continuation")
			}
			result["rows"], result["row_cursor"] = first, payload
			second, next, nextPayload := r.normalizedPage(r.request(http.MethodPost, root+"/tables/"+tableID+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v1", "cursor_token": token}, http.StatusOK), "second")
			if next == "" {
				return fmt.Errorf("second example page must have a continuation")
			}
			result["continuation"], result["continuation_cursor"] = second, nextPayload
			query, err := json.Marshal(map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": tableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}})
			if err != nil {
				return err
			}
			probe := CaptureFixtureGraph(t, r.h.Pool, r.graphLimits, uuid.MustParse(input.IncidentID), query)
			direct, err := json.Marshal(probe["result"])
			if err != nil {
				return err
			}
			routed, err := json.Marshal(result["graph"])
			if err != nil {
				return err
			}
			if !bytes.Equal(direct, routed) {
				return fmt.Errorf("example route and production composition differ")
			}
			// The full result is already compared at graph; retain all probe facts
			// without duplicating the large result a second time.
			delete(probe, "result")
			result["composition"] = probe
			state := r.durableState()
			if err := validateCiscoExampleOperationFacts(r, input, state); err != nil {
				return err
			}
			result["durable"] = r.comparableState(state)
			return nil
		})
		if err != nil {
			return nil, err
		}
		if result["job_status"] != "succeeded" {
			return nil, fmt.Errorf("example import failed: %v", result["job_error"])
		}
		state := result["durable"].(map[string]any)
		rows := state["rows"].([]any)
		for _, value := range rows {
			row := value.(map[string]any)
			var encoded bytes.Buffer
			encoder := json.NewEncoder(&encoded)
			encoder.SetEscapeHTML(false)
			if err := encoder.Encode(row["unmapped_raw"]); err != nil {
				return nil, err
			}
			digest := sha256.Sum256(bytes.TrimSuffix(encoded.Bytes(), []byte("\n")))
			row["unmapped_raw_sha256"] = hex.EncodeToString(digest[:])
			delete(row, "unmapped_raw")
		}
		return []map[string]any{result}, nil
	})
}

// Opaque operation facts are checked before the publication adapter's narrow
// comparison substitutions. Replay equality is checked by that adapter too.
func validateCiscoExampleOperationFacts(r *publicationFixtureRuntime, input publicationFixtureCase, state map[string]any) error {
	requestIDs := map[string]bool{}
	var terminal map[string]any
	for _, entry := range r.transport {
		var response map[string]any
		switch value := entry["response"].(type) {
		case json.RawMessage:
			if err := json.Unmarshal(value, &response); err != nil {
				return err
			}
		case map[string]any:
			response = value
		}
		if response != nil {
			if meta, ok := response["meta"].(map[string]any); ok {
				id, ok := meta["request_id"].(string)
				if !ok || id == "" || requestIDs[id] {
					return fmt.Errorf("invalid or repeated request identity")
				}
				requestIDs[id] = true
			}
		}
		if job, ok := entry["terminal_job"].(map[string]any); ok {
			terminal = job
		}
	}
	if terminal == nil {
		return fmt.Errorf("missing terminal job")
	}
	id, ok := terminal["job_id"].(string)
	if !ok {
		return fmt.Errorf("missing job ID")
	}
	if _, err := uuid.Parse(id); err != nil {
		return err
	}
	scope, ok := terminal["scope"].(map[string]any)
	if !ok || scope["kind"] != "incident" || scope["incident_id"] != input.IncidentID || terminal["submitted_by_user_id"] != input.ActorID || terminal["status_route"] != "/api/v1/jobs/"+id {
		return fmt.Errorf("job identity or attribution differs from import context")
	}
	prior := r.started
	for _, key := range []string{"submitted_at", "started_at", "finished_at", "updated_at"} {
		value, ok := terminal[key].(string)
		if !ok {
			return fmt.Errorf("missing terminal job time %s", key)
		}
		at, err := time.Parse(time.RFC3339Nano, value)
		if err != nil || at.Before(prior) || at.After(time.Now().UTC().Add(time.Second)) {
			return fmt.Errorf("invalid terminal job time %s", key)
		}
		prior = at
	}
	retained, ok := terminal["retained_until"].(string)
	if !ok {
		return fmt.Errorf("missing job retention time")
	}
	until, err := time.Parse(time.RFC3339Nano, retained)
	if err != nil || !until.After(prior) {
		return fmt.Errorf("invalid job retention interval")
	}
	auditIDs := map[string]bool{}
	for _, item := range state["audit"].([]any) {
		audit := item.(map[string]any)
		id, ok := audit["id"].(string)
		if !ok || auditIDs[id] {
			return fmt.Errorf("repeated audit identity")
		}
		auditIDs[id] = true
		if requestID := audit["request_id"]; requestID != nil {
			value, ok := requestID.(string)
			if !ok || !requestIDs[value] {
				return fmt.Errorf("audit request identity has no observed request")
			}
		}
	}
	return nil
}
