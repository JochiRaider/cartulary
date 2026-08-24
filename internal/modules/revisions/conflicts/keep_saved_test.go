package conflicts

import (
	"strings"
	"testing"
)

func TestDecodeStoredTargetRequiresExactPositiveIntegerRowVersion_Unit(t *testing.T) {
	for _, test := range []struct {
		name       string
		rowVersion string
	}{
		{name: "missing", rowVersion: ""},
		{name: "fractional", rowVersion: `,"row_version":1.5`},
		{name: "decimal integer", rowVersion: `,"row_version":1.0`},
		{name: "zero", rowVersion: `,"row_version":0`},
		{name: "string", rowVersion: `,"row_version":"1"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := `{"view_schema_id":"cartulary.view.parties.v1","row":{"record_id":"00000000-0000-0000-0000-000000000001"` + test.rowVersion + `}}`
			if _, err := DecodeStoredTarget([]byte(payload)); err == nil || !strings.Contains(err.Error(), "row version") {
				t.Fatalf("stored row version %q error = %v", test.rowVersion, err)
			}
		})
	}
	target, err := DecodeStoredTarget([]byte(`{"view_schema_id":"cartulary.view.parties.v1","row":{"record_id":"00000000-0000-0000-0000-000000000001","row_version":7}}`))
	if err != nil {
		t.Fatalf("decode exact integer row version: %v", err)
	}
	if value, ok := target.Row["row_version"].(int64); !ok || value != 7 {
		t.Fatalf("decoded row version = %#v", target.Row["row_version"])
	}
}
