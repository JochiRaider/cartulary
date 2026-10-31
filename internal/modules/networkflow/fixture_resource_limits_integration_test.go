package networkflow_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

type resourceConfigurationFixtureCase struct {
	Name    string          `json:"name"`
	Claimed bool            `json:"claimed"`
	Limits  json.RawMessage `json:"limits,omitempty"`
}

func TestNetworkFlowResourceLimitsFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-018-resource-limits", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		outcomes := []map[string]any{}
		base, err := bundle.InputPath("source/deployment.toml")
		if err != nil {
			return nil, err
		}
		raw, err := bundle.Read("source/configuration-cases.json")
		if err != nil {
			return nil, err
		}
		var configurations []resourceConfigurationFixtureCase
		if err := json.Unmarshal(raw, &configurations); err != nil {
			return nil, err
		}
		for _, input := range configurations {
			env := map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": strconv.FormatBool(input.Claimed)}
			if input.Claimed {
				env["CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH"] = fixtures.Path("network-flow", "key-rings.json")
			}
			if input.Limits != nil {
				env["CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS"] = string(input.Limits)
			}
			loaded, err := configassembly.Load(configassembly.LoadOptions{Path: base, Env: env})
			var result any
			if err != nil {
				result = map[string]any{"status": "rejected", "error": json.RawMessage(configtest.DiagnosticsJSON(t, err))}
			} else {
				result = map[string]any{"status": "accepted", "effective_limits": FixtureEffectiveLimits(loaded.Deployment().NetworkFlowActivity.EffectiveResourceLimits())}
			}
			outcomes = append(outcomes, map[string]any{"kind": "configuration", "case": input.Name, "request": input, "result": result})
		}
		parsers, err := CaptureFixtureResourceParsers(t, bundle)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, parsers...)
		raw, err = bundle.Read("source/import-cases.json")
		if err != nil {
			return nil, err
		}
		var imports []publicationFixtureCase
		if err := json.Unmarshal(raw, &imports); err != nil {
			return nil, err
		}
		for _, input := range imports {
			var result map[string]any
			var failure error
			passed := t.Run(input.Name, func(t *testing.T) { result, failure = runPublicationFixture(t, bundle, input, nil, nil) })
			if failure != nil {
				return nil, failure
			}
			if !passed {
				return nil, fmt.Errorf("%s: resource import failed", input.Name)
			}
			outcomes = append(outcomes, map[string]any{"kind": "import", "result": result})
		}
		graphs, err := runGraphFixtureFamily(t, bundle, nil)
		if err != nil {
			return nil, err
		}
		for _, result := range graphs {
			outcomes = append(outcomes, map[string]any{"kind": "graph", "result": result})
		}
		return outcomes, nil
	})
}
