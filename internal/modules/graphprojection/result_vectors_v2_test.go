package graphprojection

import (
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// Captured through ProjectV2 and CompletedResult before the encoding and proof changes.
// The two null-reference vectors were individually mutated and fully rebound.
//
//go:embed postgresresult/testdata/result-vectors-v2.json
var resultVectorsV2 []byte

func checkResultVectorsV2(t *testing.T) {
	var vectors map[string]CompletedResultV2
	if err := json.Unmarshal(resultVectorsV2, &vectors); err != nil {
		t.Fatal(err)
	}
	for name, input := range map[string]map[string]any{
		"empty":  graphProjectionV2GoldenFixture(t)["input"].(map[string]any),
		"direct": fieldPathInputV2(), "reverse": reverseInputV2(), "aggregate": mixedAggregationInputV2(false),
	} {
		t.Run(name, func(t *testing.T) {
			result, _, err := projectAggregationInputV2(t, input)
			if err != nil {
				t.Fatal(err)
			}
			got, err := result.CompletedResult()
			if err != nil || !reflect.DeepEqual(got, vectors[name]) {
				t.Fatalf("binding, canonical envelope or ordered child rows changed: %v", err)
			}
			if err := got.Validate(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
	}
	// Fixed independently using byte lengths, including multibyte fields.
	for _, tc := range []struct{ mapping, source, digest, id, key string }{
		{"map-a", "entity-a", "508704d0aeae547d54b8753987c2b0e1801cf2c3afbff6884983895a3c60d7ee", "vx_9c4e4d9612a63ec17a81250e411c24aa75708474e1a10e1a5925aa5ec1668961", "sk_18cb653c30edf388e8955cb105ae4f96a772fe038c4cd63537e1728ad73cbf84"},
		{"map-é", "雪", "a723ad5733834bafe058dddf555900c57a0b636f13d297f6792ee00c07fad554", "vx_d187c4c255d282166aabaffa8586838ef0119a72f4164d15ec16354e86f2afd1", "sk_cf857dd71c5e11865f248d65f5fe5b1a7bbc48ee86e87697fa6edf0c08b368cc"},
	} {
		m := entityMappingDigestV2(tc.mapping, "kind-a", "vertex")
		id, key := directVertexIdentityV2("mixed-graph", "vertex", "kind-a", tc.source, m)
		if m != tc.digest || id != tc.id || key != tc.key {
			t.Errorf("identity framing for %s: %s %s %s", tc.source, m, id, key)
		}
	}
	input := fieldPathInputV2()
	input["source_entities"].([]any)[0].(map[string]any)["source_entity_id"] = "other-entity"
	input["source_relationships"].([]any)[0].(map[string]any)["src_source_entity_id"] = "other-entity"
	replaced, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, vertex := range replaced.Vertices {
		if vertex.SourceEntityRef.SourceEntityID == "other-entity" {
			found = vertex.VertexID == "vx_032622fa83cbffce5f6b0f43e82b9686e9bdec4e27887d18d37cc66bbb977cf3"
		}
	}
	if !found {
		t.Fatal("substituted source did not derive its fixed vertex ID")
	}
	for _, name := range []string{"null-entity", "null-relationship"} {
		t.Run(name, func(t *testing.T) {
			if err := vectors[name].Validate(t.Context()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("rebound direct object without provenance: %v", err)
			}
		})
	}
}

func checkOptionalCanonicalInputV2(t *testing.T) {
	for _, withNull := range []bool{false, true} {
		input := reverseInputV2()
		config := input["projection_config"].(map[string]any)
		config["entity_mappings"].([]any)[0].(map[string]any)["inclusion_predicate"] = filterItemV2()
		def := aggregationPropertyV2("observable", "vertex", "vertex", "properties.group", "observable", "string", "single_value")
		def["null_output_policy"] = "emit_null"
		if withNull {
			def["default_value"] = nil
		}
		appendDefinitionV2(input, def, false)
		result, _, err := projectAggregationInputV2(t, input)
		if err != nil {
			t.Fatal(err)
		}
		want := "1cc07a4a7324b88922d7f012b4242b574cc60d00355e066a75278f1a79e64f02"
		if withNull {
			want = "ad89f78dec6399080254882626df0008cefd5e8d7b7ccb1d84816b9227c922f6"
		}
		if result.NormalizedConfigurationSHA256 != want {
			t.Errorf("explicit null %t: configuration digest %s, want %s", withNull, result.NormalizedConfigurationSHA256, want)
		}
	}
}
