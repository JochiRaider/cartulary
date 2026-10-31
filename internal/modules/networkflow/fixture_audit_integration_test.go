package networkflow_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

type auditFixtureRequests struct {
	Case                 string          `json:"case"`
	GraphSuccess         json.RawMessage `json:"graph_success"`
	GraphMissingTable    json.RawMessage `json:"graph_missing_table"`
	ConflictingLink      json.RawMessage `json:"conflicting_link"`
	ForbiddenAuditValues []string        `json:"forbidden_audit_values"`
}

func TestNetworkFlowAuditReplayFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-026-audit-and-replay", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		raw, err := bundle.Read("source/audit-requests.json")
		if err != nil {
			return nil, err
		}
		var requests auditFixtureRequests
		if err := json.Unmarshal(raw, &requests); err != nil {
			return nil, err
		}
		return runLinkFixtureFamily(t, bundle, func(r *publicationFixtureRuntime, input linkFixtureCase, result map[string]any, aliases *linkFixtureAliases, known map[string]bool) error {
			if input.Name != requests.Case {
				return fmt.Errorf("audit request case does not match isolated import")
			}
			root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
			operations := []any{}
			for _, step := range []struct {
				name, path string
				body       json.RawMessage
				success    bool
			}{
				{"graph-success", root + "/graphs/query", requests.GraphSuccess, true},
				{"graph-missing-table", root + "/graphs/query", requests.GraphMissingTable, false},
				{"changed-replay-intent", root + "/indicator-links", requests.ConflictingLink, false},
			} {
				before := r.linkState()
				response := r.observedPost(step.path, step.body)
				after := r.linkState()
				if !step.success && !reflect.DeepEqual(before, after) {
					return fmt.Errorf("%s: rejected operation mutated state", step.name)
				}
				operations = append(operations, map[string]any{"name": step.name, "request": step.body, "response": aliases.apply(t, response), "state": r.comparableLinkState(after, aliases, known)})
			}
			state := r.linkState()
			r.transport = append(r.transport, map[string]any{"audit_final_state": state})
			audit := state["network_flow"].(map[string]any)["audit"]
			value, err := privacyJSON(audit)
			if err != nil {
				return err
			}
			if err := rejectPrivacyValues(value, requests.ForbiddenAuditValues); err != nil {
				return err
			}
			result["audit_operations"] = operations
			result["audit_final_state"] = r.comparableLinkState(state, aliases, known)
			result["forbidden_audit_values_checked"] = requests.ForbiddenAuditValues

			return nil
		})
	})
}
