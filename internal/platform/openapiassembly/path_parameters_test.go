package openapiassembly

import (
	"strings"
	"testing"
)

// TestAssemblyEnforcesPathParameterDeclarations covers Core 01 REQ-01-707 and
// AC-617: every path placeholder resolves, directly or through a component, to
// a parameter that declares its format (format uuid, or an anchored pattern
// that compiles, with optional length bounds) and the not-found code a
// malformed value receives. That code is registered, is not a pre-resolution
// code, and is declared by the operation at its registered status; a non-404
// not-found response also declares its fixed details. A component parameter is
// checked against every operation that uses it.
func TestAssemblyEnforcesPathParameterDeclarations(t *testing.T) {
	const registry = `{"errors":[{"code":"credential_bootstrap_rejected","http_status":409},{"code":"csrf_verification_failed","http_status":403},` +
		`{"code":"method_not_allowed","http_status":405},{"code":"other_not_found","http_status":404},{"code":"route_not_found","http_status":404},` +
		`{"code":"session_required","http_status":401},{"code":"thing_not_found","http_status":404},{"code":"thing_rejected","http_status":409},` +
		fixtureImpliedRegistryEntries + `]}`
	const (
		uuidSchema = `"schema":{"type":"string","format":"uuid"}`
		notFound   = `"404":{"description":"nf","x-cartulary-error-codes":["thing_not_found"]}`
		rejected   = `"409":{"description":"r","x-cartulary-error-codes":["credential_bootstrap_rejected","thing_rejected"]}`
		components = `"components":{"parameters":{"SharedThing":{"name":"thing_id","in":"path","required":true,` + uuidSchema + `,"x-cartulary-not-found-code":"thing_not_found"}},` +
			`"schemas":{"ThingKey":{"type":"string","pattern":"^t_[a-z]{1,8}$"}},` +
			`"responses":{"ThingNotFound":{"description":"nf","x-cartulary-error-codes":["thing_not_found"]}}}`
		otherOperation = `,"/others/{thing_id}":{"get":{"operationId":"getOther","parameters":[{"$ref":"#/components/parameters/SharedThing"}],"security":[],` +
			`"responses":{"200":{"description":"ok"},"404":{"description":"nf","x-cartulary-error-codes":["other_not_found"]}}}}`
	)
	// Each case varies the parameter and responses of GET /things/{thing_id};
	// the first operation also carries the fixture's full security.
	document := func(parameter string, responses []string, extraPaths string) string {
		return `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{"/things/{thing_id}":{"get":{"operationId":"getThing",` +
			`"parameters":[` + parameter + `],"responses":{` + strings.Join(append([]string{`"200":{"description":"ok"}`}, responses...), ",") + `}}}` + extraPaths + `},` + components + `}`
	}
	inline := func(schema, extension string) string {
		return `{"name":"thing_id","in":"path","required":true,` + schema + extension + `}`
	}
	withCode := func(schema string) string {
		return inline(schema, `,"x-cartulary-not-found-code":"thing_not_found"`)
	}
	for _, tc := range []struct {
		name, document, wantError string
	}{
		{"uuid format", document(withCode(uuidSchema), []string{notFound}, ""), ""},
		{"pattern with length bounds", document(withCode(`"schema":{"type":"string","pattern":"^[a-z]+$","minLength":1,"maxLength":8}`), []string{notFound}, ""), ""},
		{"schema reference", document(withCode(`"schema":{"$ref":"#/components/schemas/ThingKey"}`), []string{notFound}, ""), ""},
		{"component parameter and component response", document(`{"$ref":"#/components/parameters/SharedThing"}`, []string{`"404":{"$ref":"#/components/responses/ThingNotFound"}`}, ""), ""},
		{"non-404 not-found with details", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"thing_rejected","x-cartulary-not-found-details":{"reason_code":"thing_not_visible"}`), []string{rejected}, ""), ""},
		{"404 not-found with fixed details", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"thing_not_found","x-cartulary-not-found-details":{"validation_code":"thing_not_found"}`), []string{notFound}, ""), ""},

		{"missing code", document(inline(uuidSchema, ""), []string{notFound}, ""), `GET /things/{thing_id}: path parameter "thing_id": requires x-cartulary-not-found-code`},
		{"empty code", document(inline(uuidSchema, `,"x-cartulary-not-found-code":""`), []string{notFound}, ""), "x-cartulary-not-found-code must be a non-empty string"},
		{"unregistered code", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"ghost_not_found"`), []string{notFound}, ""), `names unregistered code "ghost_not_found"`},
		{"pre-resolution code", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"route_not_found"`), []string{notFound}, ""), `names "route_not_found", which belongs to no operation`},
		{"code not declared at its status", document(withCode(uuidSchema), []string{`"404":{"description":"nf","x-cartulary-error-codes":["other_not_found"]}`}, ""), `"thing_not_found" must be declared by the operation's 404 response`},
		{"code without its status response", document(withCode(uuidSchema), nil, ""), `"thing_not_found" must be declared by the operation's 404 response`},
		{"non-404 not-found without details", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"thing_rejected"`), []string{rejected}, ""), "has status 409, so the parameter must declare x-cartulary-not-found-details"},
		{"details not an object", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"thing_not_found","x-cartulary-not-found-details":"x"`), []string{notFound}, ""), "x-cartulary-not-found-details must be a non-empty object"},
		{"details with a non-string value", document(inline(uuidSchema, `,"x-cartulary-not-found-code":"thing_not_found","x-cartulary-not-found-details":{"count":1}`), []string{notFound}, ""), `x-cartulary-not-found-details member "count" must be a non-empty string`},
		{"neither format nor pattern", document(withCode(`"schema":{"type":"string"}`), []string{notFound}, ""), "schema must declare exactly one of format and pattern"},
		{"both format and pattern", document(withCode(`"schema":{"type":"string","format":"uuid","pattern":"^[a-z]+$"}`), []string{notFound}, ""), "schema must declare exactly one of format and pattern"},
		{"unsupported format", document(withCode(`"schema":{"type":"string","format":"date"}`), []string{notFound}, ""), `schema format "date" is not supported`},
		{"unanchored pattern", document(withCode(`"schema":{"type":"string","pattern":"[a-z]+"}`), []string{notFound}, ""), "must be anchored with ^ and $"},
		{"pattern that does not compile", document(withCode(`"schema":{"type":"string","pattern":"^(a$"}`), []string{notFound}, ""), "does not compile"},
		{"non-string type", document(withCode(`"schema":{"type":"integer","format":"uuid"}`), []string{notFound}, ""), "schema type must be string"},
		{"undeclared schema member", document(withCode(`"schema":{"type":"string","format":"uuid","enum":["a"]}`), []string{notFound}, ""), `schema member "enum" is not allowed`},
		{"inverted length bounds", document(withCode(`"schema":{"type":"string","pattern":"^[a-z]+$","minLength":9,"maxLength":8}`), []string{notFound}, ""), "schema minLength exceeds maxLength"},
		{"fractional length bound", document(withCode(`"schema":{"type":"string","pattern":"^[a-z]+$","maxLength":1.5}`), []string{notFound}, ""), "schema maxLength must be a non-negative integer"},
		{"zero maximum length", document(withCode(`"schema":{"type":"string","pattern":"^[a-z]+$","maxLength":0}`), []string{notFound}, ""), "schema maxLength must be positive"},
		{"missing schema", document(inline("", `"x-cartulary-not-found-code":"thing_not_found"`), []string{notFound}, ""), "requires a schema"},
		{"shared component under another code", document(withCode(uuidSchema), []string{notFound}, otherOperation), `GET /others/{thing_id}: path parameter "thing_id": x-cartulary-not-found-code "thing_not_found" must be declared by the operation's 404 response`},
		{"placeholder without a declaration", document(`{"name":"other_id","in":"path","required":true,`+uuidSchema+`,"x-cartulary-not-found-code":"thing_not_found"}`, []string{notFound}, ""), "path placeholders without exact parameter declarations: thing_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): tc.document, errorRegistryPath: registry})
			_, _, err := Assemble(fixture.manifestPath)
			switch {
			case tc.wantError == "" && err != nil:
				t.Fatalf("assemble: %v", err)
			case tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)):
				t.Fatalf("assemble error = %v, want substring %q", err, tc.wantError)
			}
		})
	}
	t.Run("operation declaration overrides the path item", func(t *testing.T) {
		document := `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{"/things/{thing_id}":{"parameters":[` + inline(uuidSchema, "") + `],` +
			`"get":{"operationId":"getThing","parameters":[` + withCode(uuidSchema) + `],"responses":{"200":{"description":"ok"},` + notFound + `}}}},` + components + `}`
		fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): document, errorRegistryPath: registry})
		if _, _, err := Assemble(fixture.manifestPath); err != nil {
			t.Fatalf("assemble: %v", err)
		}
	})
}
