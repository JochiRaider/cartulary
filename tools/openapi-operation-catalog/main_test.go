package main

import (
	"bytes"
	"testing"
)

func TestExplicitMutationDeclaration(t *testing.T) {
	for _, value := range []any{nil, "false", 0, []any{}, map[string]any{}, false, true} {
		input := map[string]any{"x-cartulary-state-changing": value}
		got, err := stateChangingDeclaration(input)
		want, valid := value.(bool)
		if valid {
			if err != nil || got != want {
				t.Fatalf("boolean declaration: %v %v", got, err)
			}
		} else if err == nil {
			t.Fatalf("invalid declaration admitted: %#v", value)
		}
	}
	if _, err := stateChangingDeclaration(map[string]any{}); err == nil {
		t.Fatal("missing declaration admitted")
	}
	operations := []operation{{OperationID: "preview", Method: "POST", StateChanging: false}, {OperationID: "apply", Method: "POST", StateChanging: true}}
	generated, err := render(catalogMetadata{DocumentVersion: "2.0.0", SHA256: "fixture"}, operations)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(generated, []byte(`OperationID: "preview", Availability: "", StateChanging: false`)) || !bytes.Contains(generated, []byte(`OperationID: "apply", Availability: "", StateChanging: true`)) {
		t.Fatal("render did not preserve explicit same-method semantics")
	}
}
