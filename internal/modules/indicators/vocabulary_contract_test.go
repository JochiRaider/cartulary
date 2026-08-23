package indicators

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
)

func TestIndicatorVocabularyMatchesPortableAndOpenAPIContracts(t *testing.T) {
	t.Parallel()
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	openAPI := readIndicatorContractJSON(t, filepath.Join(root, "contracts/openapi/cartulary.openapi.yaml"))
	indicatorSchema := readIndicatorContractJSON(t, filepath.Join(root, "contracts/incident-bundles/indicators.row.v1.schema.json"))
	observationSchema := readIndicatorContractJSON(t, filepath.Join(root, "contracts/incident-bundles/indicator_observations.row.v1.schema.json"))
	intervalSchema := readIndicatorContractJSON(t, filepath.Join(root, "contracts/incident-bundles/indicator_state_intervals.row.v1.schema.json"))

	assertVocabularyProjection(t, "bundle Indicator types", vocabulary.IndicatorTypes(), contractStrings(t, indicatorSchema, "properties", "indicator_type", "enum"))
	assertVocabularyProjection(t, "bundle parsed Indicator types", vocabulary.IndicatorTypes(), contractStrings(t, observationSchema, "properties", "parsed_indicator_type", "oneOf", "0", "enum"))
	assertVocabularyProjection(t, "OpenAPI parsed Indicator types", vocabulary.IndicatorTypes(), contractStrings(t, openAPI, "components", "schemas", "IndicatorObservationCreateRequest", "properties", "parsed_indicator_type", "enum"))
	assertVocabularyProjection(t, "bundle value kinds", vocabulary.ValueKinds(), contractStrings(t, indicatorSchema, "properties", "value_kind", "enum"))
	assertVocabularyProjection(t, "bundle observation statuses", vocabulary.ObservationStatuses(), contractStrings(t, observationSchema, "properties", "resolution_status", "enum"))
	assertVocabularyProjection(t, "OpenAPI observation statuses", vocabulary.ObservationStatuses(), contractStrings(t, openAPI, "components", "schemas", "IndicatorObservation", "properties", "resolution_status", "enum"))
	assertVocabularyProjection(t, "bundle lifecycle states", vocabulary.LifecycleStates(), contractStrings(t, intervalSchema, "properties", "lifecycle_state", "enum"))
	assertVocabularyProjection(t, "OpenAPI lifecycle request states", vocabulary.LifecycleStates(), contractStrings(t, openAPI, "components", "schemas", "IndicatorLifecycleAppendRequest", "properties", "lifecycle_state", "enum"))
	assertVocabularyProjection(t, "OpenAPI lifecycle response states", vocabulary.LifecycleStates(), contractStrings(t, openAPI, "components", "schemas", "IndicatorStateInterval", "properties", "lifecycle_state", "enum"))
	if unique, ok := contractValue(t, intervalSchema, "properties", "support_refs", "uniqueItems").(bool); !ok || !unique {
		t.Fatalf("portable lifecycle support_refs uniqueItems = %#v, want true", unique)
	}
}

func TestIndicatorPortableVocabularyAndSupportReferencesAreExact(t *testing.T) {
	t.Parallel()
	canonical, err := identity.Canonicalize(identity.Input{IndicatorType: "domain_name", ValueKind: "atomic", DisplayValue: "example.test"})
	if err != nil {
		t.Fatalf("canonical fixture: %v", err)
	}
	baseIndicator := portableIdentityRow("00000000-0000-4000-8000-000000000001", canonical)
	port := mustIndicatorSourcePort(t)
	for _, mutation := range []struct {
		name string
		edit func(indicatorTestBundle)
	}{
		{name: "Indicator type case", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "indicator_type", "DOMAIN_NAME")
		}},
		{name: "Indicator type spacing", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "indicator_type", " domain_name")
		}},
		{name: "Indicator type alias", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "indicator_type", "domain")
		}},
		{name: "Indicator type unknown", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "indicator_type", "unknown")
		}},
		{name: "value-kind case", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "value_kind", "ATOMIC")
		}},
		{name: "value-kind spacing", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "value_kind", "atomic ")
		}},
		{name: "value-kind alias", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "value_kind", "literal")
		}},
		{name: "value-kind unknown", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicators.ndjson", "value_kind", "unknown")
		}},
		{name: "parsed Indicator type case", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "parsed_indicator_type", "DOMAIN_NAME")
		}},
		{name: "parsed Indicator type spacing", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "parsed_indicator_type", "domain_name ")
		}},
		{name: "parsed Indicator type alias", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "parsed_indicator_type", "domain")
		}},
		{name: "parsed Indicator type unknown", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "parsed_indicator_type", "unknown")
		}},
		{name: "observation status case", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "resolution_status", "UNRESOLVED")
		}},
		{name: "observation status spacing", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "resolution_status", "unresolved ")
		}},
		{name: "observation status alias", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "resolution_status", "open")
		}},
		{name: "observation status unknown", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_observations.ndjson", "resolution_status", "unknown")
		}},
		{name: "lifecycle case", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_state_intervals.ndjson", "lifecycle_state", "ACTIVE")
		}},
		{name: "lifecycle spacing", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_state_intervals.ndjson", "lifecycle_state", " active")
		}},
		{name: "lifecycle alias", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_state_intervals.ndjson", "lifecycle_state", "inactive")
		}},
		{name: "lifecycle unknown", edit: func(bundle indicatorTestBundle) {
			replacePortableField(t, bundle, "data/indicator_state_intervals.ndjson", "lifecycle_state", "unknown")
		}},
		{name: "duplicate support reference", edit: func(bundle indicatorTestBundle) {
			const reference = "00000000-0000-4000-8000-000000000102"
			replacePortableField(t, bundle, "data/indicator_state_intervals.ndjson", "support_refs", []any{reference, reference})
		}},
	} {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			bundle := portableIdentityBundle(t, cloneAnyMap(baseIndicator))
			mutation.edit(bundle)
			_, err := port.PrepareImport(context.Background(), bundle, portableImportContext(t, "vocabulary-"+mutation.name))
			assertIndicatorInvariantFailure(t, err, "indicators.representation_legal")
		})
	}
}

func readIndicatorContractJSON(t testing.TB, path string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var result map[string]any
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return result
}

func contractStrings(t testing.TB, root map[string]any, path ...string) []string {
	t.Helper()
	raw, ok := contractValue(t, root, path...).([]any)
	if !ok {
		t.Fatalf("contract path %v is not an array", path)
	}
	result := make([]string, 0, len(raw))
	for _, value := range raw {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("contract path %v contains non-string %#v", path, value)
		}
		result = append(result, text)
	}
	return result
}

func contractValue(t testing.TB, root any, path ...string) any {
	t.Helper()
	current := root
	for _, segment := range path {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[segment]
		case []any:
			var index int
			if _, err := fmt.Sscanf(segment, "%d", &index); err != nil || index < 0 || index >= len(typed) {
				t.Fatalf("invalid array path segment %q in %v", segment, path)
			}
			current = typed[index]
		default:
			t.Fatalf("contract path %v stopped at %q (%T)", path, segment, current)
		}
	}
	return current
}

func assertVocabularyProjection(t testing.TB, name string, want, got []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v, want %#v", name, got, want)
	}
}

func replacePortableField(t testing.TB, bundle indicatorTestBundle, path, field string, value any) {
	t.Helper()
	var row map[string]any
	if err := json.Unmarshal(bundle[path][:len(bundle[path])-1], &row); err != nil {
		t.Fatalf("decode portable row %s: %v", path, err)
	}
	row[field] = value
	bundle[path] = marshalNDJSONRows(t, []map[string]any{row})
}
