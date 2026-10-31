package networkflow_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
)

func publicationResourceLimits(t testing.TB, input publicationFixtureCase) (string, EffectiveLimits) {
	t.Helper()
	values := map[string]int64{"max_graph_vertices": 16, "max_graph_edges": 16, "max_example_row_refs_per_edge": 4, "max_aggregate_counter_digits": 40, "max_contributing_rows_per_graph": 16, "max_time_buckets_per_graph": 4}
	for key, value := range input.ResourceLimits {
		values[key] = value
	}
	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	configuration, finding := ApplyConfigurationOverlay(Configuration{Claimed: true, KeyRingManifestPath: fixtures.Path("network-flow", "key-rings.json")}, []string{"network_flow_activity", "resource_limits"}, string(raw))
	if finding != nil {
		t.Fatal(finding)
	}
	configuration, findings := NormalizeAndValidateConfiguration(configuration)
	if len(findings) != 0 {
		t.Fatal(findings)
	}
	return string(raw), configuration.EffectiveResourceLimits()
}

type graphFixtureQuery struct {
	Name    string          `json:"name"`
	Request json.RawMessage `json:"request"`
	Failure bool            `json:"failure,omitempty"`
}
type graphFixtureCase struct {
	publicationFixtureCase
	MissingSelector   json.RawMessage          `json:"missing_selector,omitempty"`
	AdditionalImports []publicationFixtureCase `json:"additional_imports,omitempty"`
	Queries           []graphFixtureQuery      `json:"queries"`
}

func (r *publicationFixtureRuntime) compareGraphQuery(input publicationFixtureCase, query graphFixtureQuery) (map[string]any, error) {
	path := "/api/v1/incidents/" + input.IncidentID + "/network-flow/graphs/query"
	probe := CaptureFixtureGraph(r.t, r.h.Pool, r.graphLimits, uuid.MustParse(input.IncidentID), query.Request)
	var response map[string]any
	if query.Failure {
		response = r.observedError(http.MethodPost, path, query.Request)
		if probe["result"] != nil || len(probe["projection_invocations"].([]map[string]any)) != 0 {
			return nil, fmt.Errorf("%s: failed composition produced partial adapter output", query.Name)
		}
	} else {
		response = r.request(http.MethodPost, path, query.Request, http.StatusOK)
		actual, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}
		direct, err := json.Marshal(probe["result"])
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(actual, direct) {
			return nil, fmt.Errorf("%s: route and production composition differ", query.Name)
		}
	}
	return map[string]any{"case": query.Name, "request": query.Request, "response": response, "composition": probe}, nil
}

func runGraphFixtureFamily(t *testing.T, bundle *networkflowfixtures.Bundle, after func(*publicationFixtureRuntime, graphFixtureCase, map[string]any) error) ([]map[string]any, error) {
	raw, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var inputs []graphFixtureCase
	if err = json.Unmarshal(raw, &inputs); err != nil {
		return nil, err
	}
	results := []map[string]any{}
	for _, input := range inputs {
		var result map[string]any
		var failure error
		passed := t.Run(input.Name, func(t *testing.T) {
			result, failure = runPublicationFixture(t, bundle, input.publicationFixtureCase, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
				imports := []map[string]any{}
				for _, next := range input.AdditionalImports {
					imported, err := runPublicationFixture(t, bundle, next, r, nil)
					if err != nil {
						return err
					}
					imports = append(imports, imported)
				}
				queries := []map[string]any{}
				for _, query := range input.Queries {
					compared, err := r.compareGraphQuery(input.publicationFixtureCase, query)
					if err != nil {
						return err
					}
					queries = append(queries, compared)
				}
				result["additional_imports"] = imports
				result["graph_queries"] = queries
				if after != nil {
					if err := after(r, input, result); err != nil {
						return err
					}
				}
				result["graph_durable_after"] = r.comparableState(r.durableState())
				result["graph_publication_counts"] = r.jsonQuery(`SELECT jsonb_build_object('declarations',(SELECT count(*) FROM network_flow_graph_views),'results',(SELECT count(*) FROM graph_projection_results))`)
				return nil
			})
		})
		if failure != nil {
			return nil, failure
		}
		if !passed {
			return nil, fmt.Errorf("%s: graph family adapter failed", input.Name)
		}
		results = append(results, result)
	}
	return results, nil
}

func TestNetworkFlowCrossTableGraphFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-006-cross-table-graph", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runGraphFixtureFamily(t, bundle, nil)
	})
}
func TestNetworkFlowGraphAdapterFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-015-graph-adapter-input", 4, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runGraphFixtureFamily(t, bundle, nil)
	})
}

func TestNetworkFlowAggregateBoundsFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-028-graph-aggregate-bounds", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runGraphFixtureFamily(t, bundle, nil)
	})
}
