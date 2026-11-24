package harnessplan

import (
	"encoding/json"
	"errors"
	"maps"
	"testing"
)

func TestInputResolution(t *testing.T) {
	limit := func(n int64) *int64 { return &n }
	target := SurfaceTarget{Name: "probe", CommandID: "cartulary.harness.command.probe.v1", TargetClass: "public", InputContract: &InputContract{
		Inputs: []InputDefinition{
			{Name: "MODE", Type: "enum", Values: []string{"normal", "off"}, Default: json.RawMessage(`"normal"`), AllowedSources: []string{"make_command_line", "environment"}, EmptyString: "omitted", Normalization: "trim"},
			{Name: "STRICT", Type: "enum", Values: []string{"a", "b"}, AllowedSources: []string{"make_command_line"}, EmptyString: "omitted", Normalization: "trim"},
			{Name: "JSON", Type: "exact_1_bool", Default: json.RawMessage(`false`), AllowedSources: []string{"make_command_line", "environment"}, EmptyString: "false", Normalization: "trim"},
			{Name: "WORKERS", Type: "positive_integer", Default: json.RawMessage(`4`), Min: limit(1), Max: limit(8), AllowedSources: []string{"make_command_line"}, EmptyString: "omitted", Normalization: "trim"},
			{Name: "ROWS", Type: "row_ids", AllowedSources: []string{"make_command_line"}, EmptyString: "invalid", Normalization: "trim"},
			{Name: "OWNER", Type: "owner_id", AllowedSources: []string{"make_command_line"}, EmptyString: "invalid", Normalization: "trim"},
		},
		Constraints: []InputConstraint{{Kind: "requires", Input: "ROWS", RequiredInput: "OWNER"}},
	}}
	resolve := func(environment map[string]string) (map[string]string, error) {
		resolved, err := resolveInputs(target, environment)
		values := map[string]string{}
		for name, input := range resolved {
			values[name] = input.Value
		}
		return values, err
	}

	t.Run("values follow their Make origin, defaults and normalization", func(t *testing.T) {
		for name, test := range map[string]struct {
			environment map[string]string
			want        map[string]string
		}{
			"defaults only": {map[string]string{}, map[string]string{"MODE": "normal", "JSON": "false", "WORKERS": "4"}},
			"command line": {
				map[string]string{"MODE": " off ", "WORKERS": "06", "OWNER": "harness.x", "ROWS": "harness.x.unit.b, harness.x.unit.a", makeInputSources: "MODE=cli WORKERS=cli OWNER=cli ROWS=cli"},
				map[string]string{"MODE": "off", "JSON": "false", "WORKERS": "6", "OWNER": "harness.x", "ROWS": "harness.x.unit.a,harness.x.unit.b"},
			},
			"allowed environment":                       {map[string]string{"MODE": "off", makeInputSources: "MODE=env"}, map[string]string{"MODE": "off", "JSON": "false", "WORKERS": "4"}},
			"disallowed environment is ignored":         {map[string]string{"STRICT": "a", makeInputSources: "STRICT=env"}, map[string]string{"MODE": "normal", "JSON": "false", "WORKERS": "4"}},
			"Makefile default origin uses the default":  {map[string]string{"MODE": "off", makeInputSources: "MODE=file"}, map[string]string{"MODE": "normal", "JSON": "false", "WORKERS": "4"}},
			"a value without an origin is not an input": {map[string]string{"MODE": "off"}, map[string]string{"MODE": "normal", "JSON": "false", "WORKERS": "4"}},
			"empty values take their empty-string rule": {
				map[string]string{"MODE": " ", "JSON": "", makeInputSources: "MODE=cli JSON=cli"},
				map[string]string{"MODE": "normal", "JSON": "false", "WORKERS": "4"},
			},
		} {
			got, err := resolve(test.environment)
			if err != nil || !maps.Equal(got, test.want) {
				t.Errorf("%s: resolved %v, %v; want %v", name, got, err, test.want)
			}
		}
	})

	t.Run("invalid values and contracts are rejected", func(t *testing.T) {
		for name, environment := range map[string]map[string]string{
			"enum outside its values":   {"MODE": "warm", makeInputSources: "MODE=cli"},
			"integer above its maximum": {"WORKERS": "9", makeInputSources: "WORKERS=cli"},
			"integer with a sign":       {"WORKERS": "+3", makeInputSources: "WORKERS=cli"},
			"exact bool other than 1":   {"JSON": "true", makeInputSources: "JSON=cli"},
			"duplicate rows":            {"OWNER": "harness.x", "ROWS": "harness.x.unit.a,harness.x.unit.a", makeInputSources: "OWNER=cli ROWS=cli"},
			"malformed owner":           {"OWNER": "x", makeInputSources: "OWNER=cli"},
			"empty invalid input":       {"OWNER": "", makeInputSources: "OWNER=cli"},
			"unmet requirement":         {"ROWS": "harness.x.unit.a", makeInputSources: "ROWS=cli"},
			"malformed origin token":    {makeInputSources: "MODE"},
			"repeated origin token":     {makeInputSources: "MODE=cli MODE=env"},
		} {
			if _, err := resolve(environment); !errors.Is(err, ErrSelection) {
				t.Errorf("%s: %v, want ErrSelection", name, err)
			}
		}
		unknown := target
		unknown.InputContract = &InputContract{Inputs: []InputDefinition{{Name: "X", Type: "semantic_text", AllowedSources: []string{"make_command_line"}}}}
		if _, err := resolveInputs(unknown, map[string]string{"X": "y", makeInputSources: "X=cli"}); !errors.Is(err, ErrSelection) {
			t.Errorf("an input type graph commands do not declare: %v", err)
		}
	})
}
