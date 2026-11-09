package native

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"testing"
)

func compileInlineShape(value map[string]any, _ map[string]bool) *schema.Shape {
	b, err := canonicaljson.Marshal(value)
	if err != nil {
		panic(err)
	}
	return schema.Compile(map[string][]byte{"inline.schema.json": b}, map[string]int{"max_profile_array": 64}, "inline.schema.json")
}
func TestSchemaUnionDispatchAndLimitBindings_Unit(t *testing.T) {
	row := nativeTestObject("tool", 1)
	sourceShape := contentShapes["framework.attack/object"].Property("source")
	for _, value := range []any{map[string]any{}, map[string]any{"type": "SECRET_TAG"}, map[string]any{"type": []any{"tool"}}} {
		findings := []Finding{}
		err := sourceShape.Walk(value, "$.source", func(f Finding) error { findings = append(findings, f); return nil })
		if err != nil || len(findings) != 1 || findings[0].Path != "$.source" {
			t.Fatal("union selection", err, findings)
		}
	}
	relationship := contentShapes["framework.attack/relationship"].Property("source")
	if !relationship.Matches(nil) || relationship.Matches(row["source"]) {
		t.Fatal("nullable typed source")
	}
	overlapping := compileInlineShape(map[string]any{"oneOf": []any{map[string]any{"type": "string"}, map[string]any{"type": "string", "minLength": float64(2)}}}, map[string]bool{})
	if !overlapping.Matches("x") || overlapping.Matches("xx") {
		t.Fatal("oneOf exact cardinality")
	}
	numeric := compileInlineShape(map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "string"}}}, map[string]bool{})
	if !numeric.Matches(json.Number("2")) || numeric.Matches(json.Number("2.5")) {
		t.Fatal("numeric kind dispatch")
	}
	numericOverlap := compileInlineShape(map[string]any{"oneOf": []any{map[string]any{"type": "integer"}, map[string]any{"type": "integer", "minimum": float64(2)}}}, map[string]bool{})
	if numericOverlap.Matches(json.Number("2")) || !numericOverlap.Matches(json.Number("1")) {
		t.Fatal("overlapping numeric union")
	}
	for _, schema := range []map[string]any{{"type": "array", "maxItems": float64(64), "x-limitId": "unknown"}, {"type": "array", "maxItems": float64(65), "x-limitId": "max_profile_array"}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid annotation accepted")
				}
			}()
			compileInlineShape(schema, map[string]bool{})
		}()
	}
}
