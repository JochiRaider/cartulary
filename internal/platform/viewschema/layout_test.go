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

	normalized := normalizeLayoutMap(t, "cartulary.view.hosts.v1", layout)
	columnOrder := stringSliceFromJSON(t, normalized["column_order"])
	hiddenFieldKeys := stringSliceFromJSON(t, normalized["hidden_field_keys"])

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
