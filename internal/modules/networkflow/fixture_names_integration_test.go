package networkflow_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

func TestNetworkFlowNamesFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-013-filename-display", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		raw, err := bundle.Read("source/cases.json")
		if err != nil {
			return nil, err
		}
		var inputs []publicationFixtureCase
		if err := json.Unmarshal(raw, &inputs); err != nil {
			return nil, err
		}
		outcomes := []map[string]any{}
		for _, input := range inputs {
			var result map[string]any
			var failure error
			passed := t.Run(input.Name, func(t *testing.T) {
				if strings.HasPrefix(input.Name, "invalid-override-") {
					source, err := bundle.Read(input.Source)
					if err != nil {
						failure = err
						return
					}
					r := startPublicationFixture(t, input, source)
					path := "/api/v1/import-sessions/" + input.SessionID + "/units/" + input.UnitID
					inspection := r.request(http.MethodPost, path+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}, http.StatusOK)
					before := r.durableState()
					body := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1", "owner_mapping": input.Mapping}
					invalid := r.errorRequest(http.MethodPost, path+"/mapping-preview", body, http.StatusBadRequest, "invalid_import_request")
					repeated := r.errorRequest(http.MethodPost, path+"/mapping-preview", body, http.StatusBadRequest, "invalid_import_request")
					if !reflect.DeepEqual(invalid, repeated) || !reflect.DeepEqual(before, r.durableState()) {
						t.Fatal("rejected override changed state or result")
					}
					result = map[string]any{"case": input.Name, "request": input, "inspection": inspection, "error": invalid, "durable": r.comparableState(before), "repeated_unchanged": true}
					encoded, err := retainedPublicationTransport(r.transport)
					if err != nil {
						failure = err
						return
					}
					failure = bundle.Record(input.Name+"-transport.json", encoded)
					return
				}
				result, failure = runPublicationFixture(t, bundle, input, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
					if input.Name != "duplicate-allocation" {
						return nil
					}
					raw, err := bundle.Read("source/naming-sequence.json")
					if err != nil {
						return err
					}
					var sequence []publicationFixtureCase
					if err := json.Unmarshal(raw, &sequence); err != nil {
						return err
					}
					if len(sequence) != 3 {
						return fmt.Errorf("naming sequence must contain all three independent import attempts")
					}
					imports := []map[string]any{}
					lifecycle := map[string]any{}
					for index, next := range sequence {
						if index == 2 {
							path := "/api/v1/incidents/" + input.IncidentID + "/network-flow/tables/nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
							body := map[string]any{"client_txn_id": "fixture-name-delete", "base_table_version": 1}
							deletion := r.request(http.MethodDelete, path, body, http.StatusOK)
							state := r.durableState()
							if replay := r.request(http.MethodDelete, path, body, http.StatusOK); !reflect.DeepEqual(deletion, replay) || !reflect.DeepEqual(state, r.durableState()) {
								return fmt.Errorf("delete replay changed name or retained effects")
							}
							lifecycle["deletion"] = deletion
							lifecycle["durable_after_deletion"] = r.comparableState(state)
						}
						outcome, err := runPublicationFixture(t, bundle, next, r, nil)
						if err != nil {
							return err
						}
						imports = append(imports, outcome)
					}
					lifecycle["imports"] = imports
					lifecycle["final_mutation_receipts"] = r.jsonQuery(`SELECT to_jsonb(count(*)) FROM route_idempotency WHERE route_key LIKE 'nf.tables.%'`)
					result["name_lifecycle"] = lifecycle
					return nil
				})
			})
			if failure != nil {
				return nil, failure
			}
			if !passed {
				return nil, fmt.Errorf("%s: naming adapter failed", input.Name)
			}
			outcomes = append(outcomes, result)
		}
		return outcomes, nil
	})
}
