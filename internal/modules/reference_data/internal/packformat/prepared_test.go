package packformat

import "testing"

func TestPreparedBoundAndSchemaReferenceConfinement_Unit(t *testing.T) {
	if !admittedHintSize(16384) || admittedHintSize(16385) || admittedHintSize(0) {
		t.Fatal("bundle hint isolated byte guard")
	}
	if !admittedSuccessfulEnvelopeSize(67108864) || admittedSuccessfulEnvelopeSize(67108865) || admittedSuccessfulEnvelopeSize(0) {
		t.Fatal("successful envelope isolated byte guard")
	}
	// A complete valid object cannot reach this defensive ceiling under its
	// nested limits. Exercise the isolated guard without weakening those limits.
	if !admittedPreparedSize(71303168) || admittedPreparedSize(71303169) || admittedPreparedSize(0) {
		t.Fatal("prepared byte guard boundary")
	}
	for _, ref := range []any{"https://example.invalid/schema.json", "../manifest.v1.schema.json", "#/properties/id", true} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("unconfined projection reference accepted")
				}
			}()
			compileShape(map[string]any{"$ref": ref}, map[string]bool{})
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("cyclic projection reference accepted")
			}
		}()
		compileProjectionChain("prepared.v1.schema.json", map[string]bool{"prepared.v1.schema.json": true})
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("reference with silently ignored siblings accepted")
			}
		}()
		compileShape(map[string]any{"$ref": "successful_envelope.v1.schema.json", "type": "string"}, map[string]bool{})
	}()
}
