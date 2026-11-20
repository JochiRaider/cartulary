package main

import (
	"bytes"
	"reflect"
	"strings"
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

// TestPathParameterProjection covers the S-1089 projection of Core 01
// REQ-01-707 path-parameter declarations: parameters are projected in template
// order from path-item and operation declarations (the operation's wins),
// through parameter and schema components, with the not-found code's declared
// status and fixed details; an incomplete declaration fails generation.
func TestPathParameterProjection(t *testing.T) {
	parameterComponents := map[string]map[string]any{
		"SharedIncident": {"name": "incident_id", "in": "path", "required": true, "schema": map[string]any{"type": "string", "format": "uuid"}, "x-cartulary-not-found-code": "incident_not_found"},
	}
	schemaComponents := map[string]map[string]any{
		"Version": {"type": "string", "pattern": "^v[1-9][0-9]*$"},
	}
	errorCodes := map[int][]string{404: {"incident_not_found", "not_found"}, 409: {"thing_rejected"}}
	pathItem := map[string]any{"parameters": []any{
		map[string]any{"$ref": "#/components/parameters/SharedIncident"},
		map[string]any{"name": "version", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
	}}
	operationObject := map[string]any{"parameters": []any{
		map[string]any{"name": "version", "in": "path", "required": true, "schema": map[string]any{"$ref": "#/components/schemas/Version"},
			"x-cartulary-not-found-code": "not_found", "x-cartulary-not-found-details": map[string]any{"validation_code": "version_not_found"}},
		map[string]any{"name": "token", "in": "path", "required": true, "schema": map[string]any{"type": "string", "pattern": "^[A-Za-z0-9._~-]+$", "minLength": float64(1), "maxLength": float64(4096)},
			"x-cartulary-not-found-code": "thing_rejected", "x-cartulary-not-found-details": map[string]any{"reason_code": "thing_not_visible"}},
		map[string]any{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer"}},
	}}
	got, err := projectPathParameters("/api/v1/incidents/{incident_id}/things/{token}/versions/{version}", pathItem, operationObject, errorCodes, parameterComponents, schemaComponents)
	if err != nil {
		t.Fatal(err)
	}
	want := []pathParameter{
		{Name: "incident_id", Format: "uuid", NotFoundCode: "incident_not_found", NotFoundStatus: 404},
		{Name: "token", Pattern: "^[A-Za-z0-9._~-]+$", MinLength: 1, MaxLength: 4096, NotFoundCode: "thing_rejected", NotFoundStatus: 409, NotFoundDetails: map[string]string{"reason_code": "thing_not_visible"}},
		{Name: "version", Pattern: "^v[1-9][0-9]*$", NotFoundCode: "not_found", NotFoundStatus: 404, NotFoundDetails: map[string]string{"validation_code": "version_not_found"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projection = %#v\nwant %#v", got, want)
	}
	generated, err := render(catalogMetadata{DocumentVersion: "2.0.0", SHA256: "fixture"}, []operation{{OperationID: "readThing", Method: "GET", PathParameters: want}, {OperationID: "listThings", Method: "GET"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		`PathParameters: []PathParameter{{Name: "incident_id", Format: "uuid", Pattern: "", MinLength: 0, MaxLength: 0, NotFoundCode: "incident_not_found", NotFoundStatus: 404, NotFoundDetails: nil}`,
		`{Name: "token", Format: "", Pattern: "^[A-Za-z0-9._~-]+$", MinLength: 1, MaxLength: 4096, NotFoundCode: "thing_rejected", NotFoundStatus: 409, NotFoundDetails: map[string]string{"reason_code": "thing_not_visible"}}`,
		`OperationID: "listThings", Availability: "", StateChanging: false, SuccessStatuses: []int(nil), ErrorCodes: map[int][]string{}, Security: [][]string(nil), PathParameters: nil}`,
		"type PathParameter struct",
	} {
		if !bytes.Contains(generated, []byte(fragment)) {
			t.Fatalf("generated catalog lacks %q", fragment)
		}
	}
	for _, tc := range []struct {
		name      string
		parameter map[string]any
		wantError string
	}{
		{"missing not-found code", map[string]any{"name": "token", "in": "path", "schema": map[string]any{"type": "string", "format": "uuid"}}, "requires x-cartulary-not-found-code"},
		{"undeclared not-found code", map[string]any{"name": "token", "in": "path", "schema": map[string]any{"type": "string", "format": "uuid"}, "x-cartulary-not-found-code": "ghost_not_found"}, "is not declared by the operation"},
		{"neither format nor pattern", map[string]any{"name": "token", "in": "path", "schema": map[string]any{"type": "string"}, "x-cartulary-not-found-code": "incident_not_found"}, "exactly one of format and pattern"},
		{"empty details", map[string]any{"name": "token", "in": "path", "schema": map[string]any{"type": "string", "format": "uuid"}, "x-cartulary-not-found-code": "incident_not_found", "x-cartulary-not-found-details": map[string]any{}}, "must be a non-empty object"},
		{"fractional bound", map[string]any{"name": "token", "in": "path", "schema": map[string]any{"type": "string", "pattern": "^a$", "maxLength": 1.5}, "x-cartulary-not-found-code": "incident_not_found"}, "must be a non-negative integer"},
		{"missing declaration", map[string]any{"name": "other", "in": "path", "schema": map[string]any{"type": "string", "format": "uuid"}, "x-cartulary-not-found-code": "incident_not_found"}, `"token" has no declaration`},
	} {
		_, err := projectPathParameters("/api/v1/things/{token}", nil, map[string]any{"parameters": []any{tc.parameter}}, errorCodes, parameterComponents, schemaComponents)
		if err == nil || !strings.Contains(err.Error(), tc.wantError) {
			t.Errorf("%s: error = %v, want substring %q", tc.name, err, tc.wantError)
		}
	}
	if _, err := projectPathParameters("/api/v1/incidents/{incident_id}", map[string]any{"parameters": []any{map[string]any{"$ref": "#/components/parameters/Missing"}}}, nil, errorCodes, parameterComponents, schemaComponents); err == nil || !strings.Contains(err.Error(), "unknown component") {
		t.Errorf("unknown parameter component: error = %v", err)
	}
}
