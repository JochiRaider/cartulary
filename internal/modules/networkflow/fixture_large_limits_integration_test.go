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

type largeLimitFixtureCase struct {
	publicationFixtureCase
	AdditionalImport *publicationFixtureCase `json:"additional_import,omitempty"`
	DeleteBeforeNext bool                    `json:"delete_before_next,omitempty"`
	Queries          []graphFixtureQuery     `json:"queries"`
}

func TestNetworkFlowLargeLimitsFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-008-large-limits", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		raw, err := bundle.Read("source/cases.json")
		if err != nil {
			return nil, err
		}
		var inputs []largeLimitFixtureCase
		if err := json.Unmarshal(raw, &inputs); err != nil {
			return nil, err
		}
		outcomes := []map[string]any{}
		for _, input := range inputs {
			var result map[string]any
			var failure error
			passed := t.Run(input.Name, func(t *testing.T) {
				result, failure = runPublicationFixture(t, bundle, input.publicationFixtureCase, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
					var quota any
					if input.AdditionalImport != nil {
						effects := map[string]any{"deletion": nil, "durable_after_deletion": nil}
						if input.DeleteBeforeNext {
							path := "/api/v1/incidents/" + input.IncidentID + "/network-flow/tables/nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
							body := map[string]any{"client_txn_id": "fixture-limit-delete", "base_table_version": 1}
							deletion := r.request(http.MethodDelete, path, body, http.StatusOK)
							state := r.durableState()
							if replay := r.request(http.MethodDelete, path, body, http.StatusOK); !reflect.DeepEqual(deletion, replay) || !reflect.DeepEqual(state, r.durableState()) {
								return fmt.Errorf("quota setup deletion replay changed retained state")
							}
							effects["deletion"], effects["durable_after_deletion"] = deletion, r.comparableState(state)
						}
						attempt, err := runPublicationFixture(t, bundle, *input.AdditionalImport, r, nil)
						if err != nil {
							return err
						}
						effects["attempt"] = attempt
						quota = effects
					}
					queries := []map[string]any{}
					for _, query := range input.Queries {
						value, err := r.compareGraphQuery(input.publicationFixtureCase, query)
						if err != nil {
							return err
						}
						queries = append(queries, value)
					}
					result["table_quota"], result["graph_queries"] = quota, queries
					result["final_state"] = r.comparableState(r.durableState())
					result["final_counts"] = r.jsonQuery(`SELECT jsonb_build_object('active_tables',(SELECT count(*) FROM network_flow_tables WHERE table_status='active'),'retained_tables',(SELECT count(*) FROM network_flow_tables),'graph_declarations',(SELECT count(*) FROM network_flow_graph_views),'graph_results',(SELECT count(*) FROM graph_projection_results),'delete_receipts',(SELECT count(*) FROM route_idempotency WHERE route_key='nf.tables.delete'))`)
					return nil
				})
			})
			if failure != nil {
				return nil, failure
			}
			if !passed {
				return nil, fmt.Errorf("%s: large limit adapter failed", input.Name)
			}
			outcomes = append(outcomes, result)
		}
		return outcomes, nil
	})
}
