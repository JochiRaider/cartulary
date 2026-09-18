package viewschema

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestNormalizeLayoutAppendsMissingDefaultHiddenReadOnlyField(t *testing.T) {
	layout := defaultLayoutMap(t, "cartulary.view.hosts.v1")
	layout["column_order"] = removeStringValue(t, layout["column_order"], "host.reusable_identifiers")
	layout["hidden_field_keys"] = removeStringValue(t, layout["hidden_field_keys"], "host.reusable_identifiers")
	layout["frozen_through_field_key"] = "host.display_name"

	normalized := normalizeLayoutMap(t, "cartulary.view.hosts.v1", layout)
	columnOrder := stringSliceFromJSON(t, normalized["column_order"])
	hiddenFieldKeys := stringSliceFromJSON(t, normalized["hidden_field_keys"])
	if normalized["frozen_through_field_key"] != "host.display_name" {
		t.Fatalf("field evolution changed the frozen boundary: %v", normalized)
	}

	if !slices.Contains(columnOrder, "host.reusable_identifiers") {
		t.Fatalf("column_order must include evolved field, got %v", columnOrder)
	}
	if !slices.Contains(hiddenFieldKeys, "host.reusable_identifiers") {
		t.Fatalf("hidden_field_keys must hide evolved field, got %v", hiddenFieldKeys)
	}
	if !slices.IsSorted(hiddenFieldKeys) {
		t.Fatalf("hidden_field_keys must remain sorted, got %v", hiddenFieldKeys)
	}
}

func TestNormalizeLayoutRejectsMissingVisibleField(t *testing.T) {
	layout := defaultLayoutMap(t, "cartulary.view.hosts.v1")
	layout["column_order"] = removeStringValue(t, layout["column_order"], "host.display_name")

	_, layoutErr := normalizeLayout(t, "cartulary.view.hosts.v1", layout)
	if layoutErr == nil || layoutErr.Field != "layout_json.column_order" || layoutErr.ReasonCode != "invalid_column_order" {
		t.Fatalf("NormalizeLayout must reject missing visible field, got %+v", layoutErr)
	}
}

func TestNormalizeLayoutRejectsMissingWritableHiddenField(t *testing.T) {
	layout := defaultLayoutMap(t, "cartulary.view.hosts.v1")
	layout["column_order"] = removeStringValue(t, layout["column_order"], "host.fqdn")
	layout["hidden_field_keys"] = removeStringValue(t, layout["hidden_field_keys"], "host.fqdn")

	_, layoutErr := normalizeLayout(t, "cartulary.view.hosts.v1", layout)
	if layoutErr == nil || layoutErr.Field != "layout_json.column_order" || layoutErr.ReasonCode != "invalid_column_order" {
		t.Fatalf("NormalizeLayout must reject missing writable hidden field, got %+v", layoutErr)
	}
}

func defaultLayoutMap(t testing.TB, viewSchemaID string) map[string]any {
	t.Helper()
	raw, layoutErr := DefaultLayout(viewSchemaID)
	if layoutErr != nil {
		t.Fatalf("DefaultLayout(%q): %+v", viewSchemaID, layoutErr)
	}
	var layout map[string]any
	if err := json.Unmarshal(raw, &layout); err != nil {
		t.Fatalf("decode default layout: %v", err)
	}
	return layout
}

func normalizeLayoutMap(t testing.TB, viewSchemaID string, layout map[string]any) map[string]any {
	t.Helper()
	raw, layoutErr := normalizeLayout(t, viewSchemaID, layout)
	if layoutErr != nil {
		t.Fatalf("NormalizeLayout(%q): %+v", viewSchemaID, layoutErr)
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		t.Fatalf("decode normalized layout: %v", err)
	}
	return normalized
}

func normalizeLayout(t testing.TB, viewSchemaID string, layout map[string]any) (json.RawMessage, *LayoutError) {
	t.Helper()
	raw, err := json.Marshal(layout)
	if err != nil {
		t.Fatalf("encode layout: %v", err)
	}
	return NormalizeLayout(raw, viewSchemaID)
}

func removeStringValue(t testing.TB, value any, remove string) []any {
	t.Helper()
	values := stringSliceFromJSON(t, value)
	result := make([]any, 0, len(values))
	for _, item := range values {
		if item != remove {
			result = append(result, item)
		}
	}
	if len(result) == len(values) {
		t.Fatalf("value %q was not present in %v", remove, values)
	}
	return result
}

func stringSliceFromJSON(t testing.TB, value any) []string {
	t.Helper()
	values, ok := value.([]any)
	if !ok {
		t.Fatalf("value is %T, want []any", value)
	}
	result := make([]string, 0, len(values))
	for _, item := range values {
		fieldKey, ok := item.(string)
		if !ok {
			t.Fatalf("item is %T, want string", item)
		}
		result = append(result, fieldKey)
	}
	return result
}

func TestLayoutVersionCompatibilityAndFrozenBoundary_Unit(t *testing.T) {
	const schema = "cartulary.view.hosts.v1"
	t.Run("current round trip preserves hidden boundary and sparse widths", func(t *testing.T) {
		layout := defaultLayoutMap(t, schema)
		layout["frozen_through_field_key"] = "host.fqdn"
		layout["column_widths"] = []any{map[string]any{"field_key": "host.fqdn", "width_px": 4096}}
		first, err := normalizeLayout(t, schema, layout)
		if err != nil {
			t.Fatalf("current: %+v", err)
		}
		second, err := NormalizeLayout(first, schema)
		if err != nil || string(first) != string(second) {
			t.Fatalf("round trip changed: %s / %s / %+v", first, second, err)
		}
	})
	t.Run("legacy conversion and original portable canonicality", func(t *testing.T) {
		layout := defaultLayoutMap(t, schema)
		delete(layout, "frozen_through_field_key")
		layout["layout_schema_id"] = LegacyLayoutSchemaID
		raw, _ := json.Marshal(layout)
		legacy, err := NormalizeLegacyLayout(raw, schema)
		if err != nil {
			t.Fatalf("legacy: %+v", err)
		}
		normalized := normalizeLayoutMap(t, schema, layout)
		if normalized["layout_schema_id"] != LayoutSchemaID || normalized["frozen_through_field_key"] != nil {
			t.Fatal(normalized)
		}
		var original map[string]any
		if json.Unmarshal(legacy, &original) != nil || original["layout_schema_id"] != LegacyLayoutSchemaID || len(original) != 4 {
			t.Fatal(string(legacy))
		}
	})
	for name, change := range map[string]func(map[string]any){
		"unknown version":     func(m map[string]any) { m["layout_schema_id"] = "cartulary.layout.v99" },
		"missing boundary":    func(m map[string]any) { delete(m, "frozen_through_field_key") },
		"unknown boundary":    func(m map[string]any) { m["frozen_through_field_key"] = "host.unknown" },
		"technical boundary":  func(m map[string]any) { m["frozen_through_field_key"] = "record_id" },
		"non-string boundary": func(m map[string]any) { m["frozen_through_field_key"] = 1 },
		"unknown member":      func(m map[string]any) { m["frozen_count"] = 1 },
		"legacy extension":    func(m map[string]any) { m["layout_schema_id"] = LegacyLayoutSchemaID },
		"duplicate order":     func(m map[string]any) { m["column_order"] = append(m["column_order"].([]any), "host.fqdn") },
		"invalid width": func(m map[string]any) {
			m["column_widths"] = []any{map[string]any{"field_key": "host.fqdn", "width_px": 4097}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			layout := defaultLayoutMap(t, schema)
			change(layout)
			if _, err := normalizeLayout(t, schema, layout); err == nil {
				t.Fatal("invalid layout accepted")
			}
		})
	}
	t.Run("duplicate JSON members", func(t *testing.T) {
		raw, _ := DefaultLayout(schema)
		duplicate := append([]byte(`{"frozen_through_field_key":null,`), raw[1:]...)
		if _, err := NormalizeLayout(duplicate, schema); err == nil {
			t.Fatal("duplicate JSON member accepted")
		}
	})
}
