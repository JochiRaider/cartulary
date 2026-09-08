package httpapi

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestIncidentMetadataOpenAPIProjectsSparseNullableVersionedRequest_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	operation := openAPIObjectAt(t, openAPIObjectAt(t, openAPIObjectAt(t, document, "paths"), "/api/v1/incidents/{incident_id}"), "patch")
	body := openAPIObjectAt(t, operation, "requestBody")
	if body["required"] != true {
		t.Fatal("metadata body must be required")
	}
	ref := openAPIObjectAt(t, openAPIObjectAt(t, openAPIObjectAt(t, body, "content"), "application/json"), "schema")
	if ref["$ref"] != "#/components/schemas/IncidentPatchRequest" {
		t.Fatalf("unexpected request: %#v", ref)
	}
	schema := openAPIObjectAt(t, openAPIObjectAt(t, openAPIObjectAt(t, document, "components"), "schemas"), "IncidentPatchRequest")
	if schema["additionalProperties"] != false || !reflect.DeepEqual(schema["required"], []any{"base_incident_version"}) {
		t.Fatal("request must be closed with only base version required")
	}
	if _, ok := schema["minProperties"]; ok {
		t.Fatal("must not invent an effective-change requirement")
	}
	properties := openAPIObjectAt(t, schema, "properties")
	if len(properties) != 6 {
		t.Fatal("unexpected metadata members")
	}
	version := openAPIObjectAt(t, properties, "base_incident_version")
	if version["type"] != "integer" || version["minimum"] != float64(1) {
		t.Fatal("base must be positive integer")
	}
	for _, field := range []string{"description", "severity", "current_phase", "primary_external_case_ref"} {
		property := openAPIObjectAt(t, properties, field)
		contract, limit := "incident_metadata_text_v1", float64(128)
		if field == "description" {
			contract, limit = "multiline_body_v1", 16384
		}
		if !reflect.DeepEqual(property["type"], []any{"string", "null"}) || openAPIObjectAt(t, openAPIObjectAt(t, operation, "x-cartulary-string-contracts"), field)["string_contract_id"] != contract || openAPIObjectAt(t, openAPIObjectAt(t, operation, "x-cartulary-string-contracts"), field)["normalized_max_scalar_values"] != limit {
			t.Fatalf("incorrect bound field %s: %#v", field, property)
		}
		if _, ok := property["maxLength"]; ok {
			t.Fatal("raw length is not normalized scalar length")
		}
	}
	tlp := openAPIObjectAt(t, properties, "tlp")
	if !reflect.DeepEqual(tlp["enum"], []any{"TLP:CLEAR", "TLP:GREEN", "TLP:AMBER", "TLP:AMBER+STRICT", "TLP:RED", nil}) {
		t.Fatal("TLP tokens drifted")
	}
}
