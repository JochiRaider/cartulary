package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestReportingFrozenOutputOptions(t *testing.T) {
	for _, kind := range []string{OutputKindSlidev, OutputKindMermaid} {
		for _, scope := range supportedReleaseScopes() {
			for _, sourceOnly := range []bool{false, true} {
				if sourceOnly && scope == ReleaseScopeExternal {
					continue
				}
				raw := json.RawMessage(`{"source_only":false}`)
				if sourceOnly {
					raw = json.RawMessage(`{"source_only":true}`)
				}
				admitted, apiErr := materializeOutputOptions(raw, kind, scope)
				if apiErr != nil {
					t.Fatal(apiErr)
				}
				got, err := parseRenderOutputOptions(admitted, kind, scope)
				if err != nil || got.SourceOnly != sourceOnly {
					t.Fatal("admitted value did not survive render consumption", err)
				}
				canonical, err := canonicalJSON(got)
				if err != nil || !bytes.Equal(admitted, canonical) {
					t.Fatal("render consumer changed admitted bytes", err)
				}
			}
		}
	}
	base := fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft)
	mutate := func(key string, value any, remove bool) []byte {
		var object map[string]any
		if err := json.Unmarshal(base, &object); err != nil {
			t.Fatal(err)
		}
		if remove {
			delete(object, key)
		} else {
			object[key] = value
		}
		encoded, err := canonicalJSON(object)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	invalid := map[string][]byte{
		"absent": nil, "null": []byte(`null`), "array": []byte(`[]`),
		"unknown": mutate("future", false, false), "retired schema": mutate("schema_id", "cartulary.reporting_render_request_options.v0", false),
		"boolean string": mutate("pdf", "true", false), "number": mutate("pdf", 1, false),
		"duplicate":               []byte(strings.Replace(string(base), `"pdf":true`, `"pdf":true,"pdf":true`, 1)),
		"invalid UTF8":            append(append([]byte(nil), base[:len(base)-1]...), 0xff, '}'),
		"conflicting source only": mutate("source_only", true, false),
		"unsupported png":         mutate("png", true, false), "unsupported pptx": mutate("pptx", true, false),
	}
	for _, key := range []string{"schema_id", "source_only", "pdf", "svg", "png", "pptx", "rendered_diagrams"} {
		invalid["missing "+key] = mutate(key, nil, true)
		invalid["null "+key] = mutate(key, nil, false)
		t.Run("admission rejects null "+key, func(t *testing.T) {
			if _, apiErr := materializeOutputOptions(mutate(key, nil, false), OutputKindSlidev, ReleaseScopeInternalDraft); apiErr == nil || apiErr.Details["reason_code"] != "field_not_nullable" {
				t.Fatal("explicit null became a default boolean", apiErr)
			}
		})
	}
	for name, raw := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := parseRenderOutputOptions(raw, OutputKindSlidev, ReleaseScopeInternalDraft)
			var validation *renderValidationError
			if !errors.As(err, &validation) {
				t.Fatal("malformed frozen options were repaired or admitted", err)
			}
		})
	}
	if _, err := parseRenderOutputOptions(mutate("pdf", false, false), OutputKindSlidev, ReleaseScopeExternal); err == nil {
		t.Fatal("external required output omission was admitted")
	}
}
