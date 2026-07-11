package testsupport

import (
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func RequireViewSchema(t testing.TB, testID string, viewSchemaIDs ...string) {
	t.Helper()

	missing := make([]string, 0)
	for _, viewSchemaID := range viewSchemaIDs {
		if _, ok := viewschema.Lookup(viewSchemaID); !ok {
			missing = append(missing, viewSchemaID)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s missing view schemas: %s", testID, strings.Join(missing, ", "))
	}
}

func AllowedFieldKeys(t testing.TB, testID string, viewSchemaID string) []string {
	t.Helper()

	resource, ok := viewschema.LookupPublicResource(viewSchemaID)
	if !ok {
		t.Fatalf("%s missing view schema %s", testID, viewSchemaID)
	}

	allowed := make([]string, 0, len(resource.TechnicalFields)+len(resource.Fields))
	allowed = append(allowed, resource.TechnicalFields...)
	for _, field := range resource.Fields {
		allowed = append(allowed, field.FieldKey)
	}
	slices.Sort(allowed)
	return allowed
}

func RequireFieldBindingMode(t testing.TB, testID string, viewSchemaID string, fieldKey string, wantBindingMode string) {
	t.Helper()

	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok {
		t.Fatalf("%s missing field %s in view schema %s", testID, fieldKey, viewSchemaID)
	}
	if field.EntityBindingMode == nil {
		t.Fatalf("%s expected %s %s entity_binding_mode=%s, got nil", testID, viewSchemaID, fieldKey, wantBindingMode)
	}
	if *field.EntityBindingMode != wantBindingMode {
		t.Fatalf("%s expected %s %s entity_binding_mode=%s, got %s", testID, viewSchemaID, fieldKey, wantBindingMode, *field.EntityBindingMode)
	}
}
