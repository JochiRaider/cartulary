package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

func FixtureEffectiveLimits(limits EffectiveLimits) map[string]any {
	return effectiveLimitsResource(limits)
}

type resourceParserFixtureCase struct {
	Name           string          `json:"name"`
	Source         string          `json:"source"`
	Preview        bool            `json:"preview"`
	ExpectedSHA256 string          `json:"expected_sha256"`
	Limits         json.RawMessage `json:"limits"`
}

func CaptureFixtureResourceParsers(t testing.TB, bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
	t.Helper()
	raw, err := bundle.Read("source/parser-cases.json")
	if err != nil {
		return nil, err
	}
	var inputs []resourceParserFixtureCase
	if err := json.Unmarshal(raw, &inputs); err != nil {
		return nil, err
	}
	outcomes := []map[string]any{}
	for _, input := range inputs {
		configuration, finding := ApplyConfigurationOverlay(Configuration{Claimed: true, KeyRingManifestPath: fixtures.Path("network-flow", "key-rings.json")}, []string{"network_flow_activity", "resource_limits"}, string(input.Limits))
		if finding != nil {
			t.Fatal(finding)
		}
		configuration, findings := NormalizeAndValidateConfiguration(configuration)
		if len(findings) != 0 {
			t.Fatal(findings)
		}
		source, err := bundle.Read(input.Source)
		if err != nil {
			return nil, err
		}
		var parsed parsedCSV
		if input.Preview {
			parsed, err = parseCSVPreview(bytes.NewReader(source), input.ExpectedSHA256, configuration.EffectiveResourceLimits())
		} else {
			parsed, err = parseCSVApply(bytes.NewReader(source), input.ExpectedSHA256, configuration.EffectiveResourceLimits())
		}
		var result any
		if err != nil {
			var invalid *sourceValidationError
			if !errors.As(err, &invalid) {
				return nil, err
			}
			result = map[string]any{"status": "rejected", "error": invalid}
		} else {
			result = map[string]any{"status": "accepted", "source_sha256": parsed.SourceContentSHA256, "columns": parsed.SourceColumns, "records": parsed.Records}
		}
		outcomes = append(outcomes, map[string]any{"kind": "parser", "case": input.Name, "request": input, "result": result})
	}
	return outcomes, nil
}
