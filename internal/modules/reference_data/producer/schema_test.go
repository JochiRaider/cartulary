package producer

import "testing"

func testProducerSchemaCompiler(t *testing.T) {
	for _, input := range []string{
		`{`, `{"type":"string","type":"number"}`, `{"type":"not-a-type"}`,
		`{"$ref":"https://example.invalid/schema.json"}`, `{"$ref":"file:///etc/passwd"}`,
	} {
		if _, err := compileProducerSchema([]byte(input), ""); err == nil {
			t.Fatal("invalid or external schema admitted")
		}
	}
	schema, err := compileProducerSchema([]byte(`{"$defs":{"stamp":{"type":"string","format":"date-time"}}}`), "#/$defs/stamp")
	if err != nil {
		t.Fatal(err)
	}
	if schema.Validate("2026-11-06T00:00:00Z") != nil || schema.Validate("not-a-date") == nil {
		t.Fatal("format assertion changed")
	}
	if Digest([]byte("abc")) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("SHA-256 witness failed")
	}
}
