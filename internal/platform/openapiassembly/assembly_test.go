package openapiassembly

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProductionAssemblyIsByteIdentical(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "..", "contracts", "openapi-source", "manifest.json")
	output, target, err := Assemble(manifestPath)
	if err != nil {
		t.Fatalf("assemble production manifest: %v", err)
	}
	canonical, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read canonical target: %v", err)
	}
	if !bytes.Equal(output, canonical) {
		t.Fatal("assembled production document is not byte-identical")
	}
}

func TestProductionCompletenessFoundation(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "..", "contracts", "openapi-source", "manifest.json")
	output, _, err := Assemble(manifestPath)
	if err != nil {
		t.Fatalf("assemble production manifest: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(output, &document); err != nil {
		t.Fatalf("decode assembled OpenAPI: %v", err)
	}
	components := fixtureObject(t, document["components"], "components")
	schemes := fixtureObject(t, components["securitySchemes"], "components.securitySchemes")
	if strings.Join(sortedMapKeysForTest(schemes), "\x00") != strings.Join(requiredSecuritySchemeNames, "\x00") {
		t.Fatalf("security scheme inventory changed: %#v", sortedMapKeysForTest(schemes))
	}
	requireAPIKeyScheme(t, schemes, "sessionCookie", "cookie", "cartulary_session")
	requireAPIKeyScheme(t, schemes, "csrfCookie", "cookie", "cartulary_csrf")
	requireAPIKeyScheme(t, schemes, "csrfHeader", "header", "X-CSRF-Token")
	requireBearerScheme(t, schemes, "bearerSession")
	requireBearerScheme(t, schemes, "credentialBootstrapBearer")

	parameters := fixtureObject(t, components["parameters"], "components.parameters")
	// S-1089 splits the record, user and view-schema identifiers by owner,
	// because their not-found responses differ (Core 01 REQ-01-707).
	expectedParameters := []string{
		"AuthBindingIDPathParameter", "ConflictTokenPathParameter", "EnterpriseBindingUserIDPathParameter",
		"EvidenceHandleTokenPathParameter", "EvidenceRecordIDPathParameter",
		"IncidentIDPathParameter", "IndicatorIDPath", "IndicatorListCursor", "IndicatorListLimit",
		"IndicatorObservationIDPath", "MembershipUserIDPathParameter", "ProviderKeyPathParameter",
		"RecordIDPathParameter", "UserIDPathParameter", "ViewSchemaDiscoveryIDPathParameter",
		"ViewSchemaIDPathParameter",
	}
	if strings.Join(sortedMapKeysForTest(parameters), "\x00") != strings.Join(expectedParameters, "\x00") {
		t.Fatalf("shared parameter inventory changed: %#v", sortedMapKeysForTest(parameters))
	}
	for name, raw := range parameters {
		parameter := fixtureObject(t, raw, "components.parameters."+name)
		if name == "IndicatorListCursor" || name == "IndicatorListLimit" {
			if parameter["in"] != "query" || parameter["required"] != false {
				t.Fatalf("%s must be an optional query parameter, got %#v", name, parameter)
			}
			continue
		}
		if parameter["in"] != "path" || parameter["required"] != true {
			t.Fatalf("%s must be a required path parameter, got %#v", name, parameter)
		}
		if code, _ := parameter[notFoundCodeMember].(string); code == "" {
			t.Fatalf("%s must declare %s", name, notFoundCodeMember)
		}
	}

	responses := fixtureObject(t, components["responses"], "components.responses")
	expectedResponses := []string{
		"BadRequestErrorResponse",
		"ConflictErrorResponse",
		"ForbiddenErrorResponse",
		"InternalServerErrorResponse",
		"NotFoundErrorResponse",
		"ServiceUnavailableErrorResponse",
		"TooManyRequestsErrorResponse",
		"UnauthorizedErrorResponse",
	}
	// The shared responses for the declarations every operation, or every
	// operation of a reserved extension route family, implies (Core 01
	// REQ-01-702) declare only the implied code, so any such operation may
	// share them.
	for name, code := range map[string]string{
		"InternalServerErrorResponse":     "internal_error",
		"ServiceUnavailableErrorResponse": "service_unavailable",
		"NotFoundErrorResponse":           "extension_profile_not_claimed",
	} {
		response := fixtureObject(t, responses[name], "components.responses."+name)
		if codes, _ := response[errorCodesMember].([]any); len(codes) != 1 || codes[0] != code {
			t.Fatalf("%s must declare exactly %s, got %#v", name, code, response[errorCodesMember])
		}
	}
	for _, name := range expectedResponses {
		raw, ok := responses[name]
		if !ok {
			t.Fatalf("common response %s is missing from %#v", name, sortedMapKeysForTest(responses))
		}
		response := fixtureObject(t, raw, "components.responses."+name)
		content := fixtureObject(t, response["content"], name+".content")
		media := fixtureObject(t, content["application/json"], name+".content.application/json")
		schema := fixtureObject(t, media["schema"], name+".schema")
		if schema["$ref"] != "#/components/schemas/ErrorEnvelope" {
			t.Fatalf("%s must reference ErrorEnvelope, got %#v", name, schema)
		}
	}

}

func TestAssemblyIgnoresManifestAndMemberOrder(t *testing.T) {
	fixture := newAssemblyFixture(t, []unitEntry{
		{OwnerID: "platform.openapi", Path: "contracts/openapi-source/owners/platform.openapi/root.json", Role: unitRootRole},
		{OwnerID: "module.auth", Path: "contracts/openapi-source/owners/module.auth/auth.json", Role: unitOwnerRole},
		{OwnerID: "module.incidents", Path: "contracts/openapi-source/owners/module.incidents/incidents.json", Role: unitOwnerRole},
	}, map[string]string{
		"contracts/openapi-source/owners/platform.openapi/root.json":      `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{},"components":{"schemas":{}}}`,
		"contracts/openapi-source/owners/module.auth/auth.json":           `{"paths":{"/z":{"get":{"operationId":"z"}}},"components":{"schemas":{"Z":{"type":"string"}}}}`,
		"contracts/openapi-source/owners/module.incidents/incidents.json": `{"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false}}},"components":{"schemas":{"A":{"type":"string"}}}}`,
	})
	first, _, err := Assemble(fixture.manifestPath)
	if err != nil {
		t.Fatalf("first assembly: %v", err)
	}
	second, _, err := Assemble(fixture.manifestPath)
	if err != nil {
		t.Fatalf("second assembly: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated assembly changed bytes")
	}
	fixture.manifest.Units[1], fixture.manifest.Units[2] = fixture.manifest.Units[2], fixture.manifest.Units[1]
	fixture.writeManifest(t)
	permuted, _, err := Assemble(fixture.manifestPath)
	if err != nil {
		t.Fatalf("permuted assembly: %v", err)
	}
	if !bytes.Equal(first, permuted) {
		t.Fatal("manifest permutation changed canonical bytes")
	}
	// strings.Index returns -1 for an absent member, so require presence first.
	for _, pair := range [][2]string{{`"/a"`, `"/z"`}, {`"A"`, `"Z"`}} {
		before, after := strings.Index(string(first), pair[0]), strings.Index(string(first), pair[1])
		if before < 0 || after < 0 || before > after {
			t.Fatalf("%s and %s are not both serialized in canonical order: %d, %d", pair[0], pair[1], before, after)
		}
	}
}

func TestAssemblyRejectsUnsafeOrAmbiguousInputs(t *testing.T) {
	tests := []struct {
		name      string
		units     []unitEntry
		contents  map[string]string
		mutate    func(*testing.T, *assemblyFixture)
		wantError string
	}{
		{
			name:  "duplicate JSON key",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			wantError: "duplicate object key",
		},
		{
			name:  "duplicate operation ID",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"same"}},"/b":{"get":{"operationId":"same"}}}}`,
			},
			wantError: "duplicate operationId",
		},
		{
			name: "path method collision",
			units: []unitEntry{
				{OwnerID: "platform.openapi", Path: rootUnitPath(), Role: unitRootRole},
				{OwnerID: "module.auth", Path: authUnitPath(), Role: unitOwnerRole},
			},
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"root"}}}}`,
				authUnitPath(): `{"paths":{"/a":{"get":{"operationId":"owner"}}}}`,
			},
			wantError: "path/method collision",
		},
		{
			name:  "unresolved reference",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{"200":{"$ref":"#/components/responses/Missing"}}}}}}`,
			},
			wantError: "unresolved reference",
		},
		{
			name:  "unwaived placeholder",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a/{id}":{"get":{"operationId":"a","x-cartulary-state-changing":false}}}}`,
			},
			wantError: "without exact parameter declarations",
		},
		{
			name: "invalid owner identifier",
			units: []unitEntry{
				{OwnerID: "platform.openapi", Path: rootUnitPath(), Role: unitRootRole},
				{OwnerID: "module.Invalid", Path: authUnitPath(), Role: unitOwnerRole},
			},
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authUnitPath(): `{"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false}}}}`,
			},
			wantError: "invalid owner_id",
		},
		{
			name: "owner directory mismatch",
			units: []unitEntry{
				{OwnerID: "platform.openapi", Path: rootUnitPath(), Role: unitRootRole},
				{OwnerID: "module.incidents", Path: authUnitPath(), Role: unitOwnerRole},
			},
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authUnitPath(): `{"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false}}}}`,
			},
			wantError: "declared owner directory",
		},
		{
			name: "retired bootstrap role",
			units: []unitEntry{
				{OwnerID: "platform.openapi", Path: rootUnitPath(), Role: unitRootRole},
				{OwnerID: "module.auth", Path: authUnitPath(), Role: "bootstrap"},
			},
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authUnitPath(): `{"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false}}}}`,
			},
			wantError: "invalid role",
		},
		{
			name:  "orphan unit",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				writeFixtureFile(t, fixture.root, "contracts/openapi-source/owners/module.auth/orphan.json", `{}`)
			},
			wantError: "orphan unit",
		},
		{
			name:  "absolute unit path",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				fixture.manifest.Units[0].Path = filepath.Join(fixture.root, rootUnitPath())
				fixture.writeManifest(t)
			},
			wantError: "relative slash path",
		},
		{
			name:  "traversal unit path",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				fixture.manifest.Units[0].Path = "../root.json"
				fixture.writeManifest(t)
			},
			wantError: "escapes its base directory",
		},
		{
			name:  "symlink unit",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				path := filepath.Join(fixture.root, filepath.FromSlash(rootUnitPath()))
				target := filepath.Join(fixture.root, "outside.json")
				if err := os.WriteFile(target, []byte(`{}`), 0o644); err != nil {
					t.Fatalf("write symlink target: %v", err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatalf("remove unit: %v", err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatalf("create symlink: %v", err)
				}
			},
			wantError: "symlink",
		},
		{
			name:  "depth limit",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): deeplyNestedRootUnit(130),
			},
			wantError: "JSON depth exceeds 128",
		},
		{
			name:  "unit byte limit",
			units: rootOnlyUnits(),
			contents: map[string]string{
				rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"` + strings.Repeat("x", 2*1024*1024) + `","version":"1"},"paths":{}}`,
			},
			wantError: "exceeds max_unit_bytes",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, test.units, test.contents)
			if test.mutate != nil {
				test.mutate(t, fixture)
			}
			_, _, err := Assemble(fixture.manifestPath)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("assemble error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestCompletenessFoundationRejectsIncompleteOrMalformedOperations(t *testing.T) {
	tests := []struct {
		name      string
		document  string
		wantError string
	}{
		{
			name:      "new operation without responses",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"security":[]}}}}`,
			wantError: "operation requires responses",
		},
		{
			name:      "empty responses",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{},"security":[]}}}}`,
			wantError: "responses must be a nonempty object",
		},
		{
			name:      "new operation without security",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{"200":{"description":"ok"}}}}}}`,
			wantError: "operation requires explicit security",
		},
		{
			name:      "unknown security scheme",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{"200":{"description":"ok"}},"security":[{"unknown":[]}]}}}}`,
			wantError: "unknown scheme",
		},
		{
			name:      "nonempty security scopes",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{"200":{"description":"ok"}},"security":[{"bearerSession":["invalid"]}]}}}}`,
			wantError: "scopes must be an empty array",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{
				rootUnitPath(): test.document,
			})
			writeFixtureFile(
				t,
				fixture.root,
				rootUnitPath(),
				fixtureDocument(t, test.document, true, false),
			)
			_, _, err := Assemble(fixture.manifestPath)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("assemble error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestAssemblyFailureDoesNotTouchTarget(t *testing.T) {
	fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{
		rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a/{id}":{"get":{"operationId":"a","x-cartulary-state-changing":false}}}}`,
	})
	target := filepath.Join(fixture.root, "contracts", "openapi", "cartulary.openapi.yaml")
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target before failed assembly: %v", err)
	}
	if _, _, err := Assemble(fixture.manifestPath); err == nil {
		t.Fatal("invalid assembly unexpectedly passed")
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target after failed assembly: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("failed assembly changed the target")
	}
}

func TestCheckAndAtomicWriteModes(t *testing.T) {
	fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{
		rootUnitPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/health":{"get":{"operationId":"fixtureHealth"}}}}`,
	})
	output, target, err := Assemble(fixture.manifestPath)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if err := CheckTarget(target, output); err == nil {
		t.Fatal("check mode did not detect target drift")
	}
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target after check: %v", err)
	}
	if string(before) != "target sentinel\n" {
		t.Fatal("check mode wrote the target")
	}
	if err := WriteTargetAtomically(target, output); err != nil {
		t.Fatalf("atomic write: %v", err)
	}
	if err := CheckTarget(target, output); err != nil {
		t.Fatalf("check generated target: %v", err)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(filepath.Dir(target), "."+filepath.Base(target)+".tmp-*"))
	if err != nil {
		t.Fatalf("glob temporary targets: %v", err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("atomic write left temporary targets: %v", temporaryFiles)
	}
}

func TestAggregateResourceLimits(t *testing.T) {
	document, err := parseOrderedJSON(
		[]byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","x-cartulary-state-changing":false,"responses":{"200":{"description":"ok"}},"security":[]}}},"components":{"schemas":{"A":{}},"securitySchemes":{"bearerSession":{},"credentialBootstrapBearer":{},"csrfCookie":{},"csrfHeader":{},"sessionCookie":{}}}}`),
		128,
	)
	if err != nil {
		t.Fatalf("parse document: %v", err)
	}
	tests := []struct {
		name      string
		limits    resourceLimits
		wantError string
	}{
		{name: "paths", limits: resourceLimits{MaxPaths: 0, MaxOperations: 10, MaxNamedComponents: 10}, wantError: "max_paths"},
		{name: "operations", limits: resourceLimits{MaxPaths: 10, MaxOperations: 0, MaxNamedComponents: 10}, wantError: "max_operations"},
		{name: "components", limits: resourceLimits{MaxPaths: 10, MaxOperations: 10, MaxNamedComponents: 5}, wantError: "max_named_components"},
	}
	// The fixture holds exactly 1 path, 1 operation and 6 named components;
	// each rejection above is one below its count. At the exact counts the
	// limit checks pass and validation reaches the next check, the fixture's
	// unused security schemes.
	if err := validateAggregate(cloneValue(document), resourceLimits{MaxPaths: 1, MaxOperations: 1, MaxNamedComponents: 6}); err == nil || !strings.HasPrefix(err.Error(), "unused security scheme ") {
		t.Fatalf("exact aggregate counts must pass the limit checks, got %v", err)
	}
	nested := func(depth int) []byte {
		return []byte(strings.Repeat(`{"a":`, depth-1) + `{}` + strings.Repeat(`}`, depth-1))
	}
	if _, err := parseOrderedJSON(nested(128), 128); err != nil {
		t.Fatalf("depth 128 must be accepted: %v", err)
	}
	if _, err := parseOrderedJSON(nested(129), 128); err == nil || err.Error() != "JSON depth exceeds 128" {
		t.Fatalf("depth 129 error = %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAggregate(cloneValue(document), test.limits)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("validate error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

type assemblyFixture struct {
	root         string
	manifestPath string
	manifest     manifest
}

func prepareFixtureContents(
	t *testing.T,
	units []unitEntry,
	contents map[string]string,
) map[string]string {
	t.Helper()
	prepared := make(map[string]string, len(contents))
	for path, content := range contents {
		prepared[path] = content
	}
	for _, unit := range units {
		content, ok := prepared[unit.Path]
		if !ok {
			continue
		}
		prepared[unit.Path] = fixtureDocument(t, content, unit.Role == unitRootRole, true)
	}
	return prepared
}

func fixtureDocument(t *testing.T, content string, addSecuritySchemes, completeOperations bool) string {
	t.Helper()
	document, err := parseOrderedJSON([]byte(content), 128)
	if err != nil || document.kind != objectKind {
		return content
	}
	if addSecuritySchemes {
		components, ok := objectMember(document, "components")
		if !ok {
			components = &orderedValue{kind: objectKind}
			document.object = append(document.object, orderedMember{name: "components", value: components})
		}
		if components.kind != objectKind {
			return content
		}
		if _, ok := objectMember(components, "securitySchemes"); !ok {
			schemes, parseErr := parseOrderedJSON(
				[]byte(`{"bearerSession":{},"credentialBootstrapBearer":{},"csrfCookie":{},"csrfHeader":{},"sessionCookie":{}}`),
				128,
			)
			if parseErr != nil {
				t.Fatalf("parse fixture security schemes: %v", parseErr)
			}
			components.object = append(components.object, orderedMember{name: "securitySchemes", value: schemes})
		}
	}
	if completeOperations {
		completeFixtureOperations(t, document)
	}
	var output bytes.Buffer
	writeOrderedJSON(&output, document, "  ", 0)
	return output.String()
}

func completeFixtureOperations(t *testing.T, document *orderedValue) {
	t.Helper()
	paths, ok := objectMember(document, "paths")
	if !ok || paths.kind != objectKind {
		return
	}
	firstOperation := true
	for _, path := range paths.object {
		if path.value.kind != objectKind {
			continue
		}
		for _, method := range path.value.object {
			if _, ok := operationKeys[method.name]; !ok || method.value.kind != objectKind {
				continue
			}
			if _, ok := objectMember(method.value, "x-cartulary-state-changing"); !ok {
				method.value.object = append(method.value.object, orderedMember{name: "x-cartulary-state-changing", value: &orderedValue{kind: boolKind, scalar: false}})
			}
			if _, ok := objectMember(method.value, "responses"); !ok {
				responses, err := parseOrderedJSON([]byte(`{"200":{"description":"ok"}}`), 128)
				if err != nil {
					t.Fatalf("parse fixture responses: %v", err)
				}
				method.value.object = append(method.value.object, orderedMember{name: "responses", value: responses})
			}
			completeFixtureOperationImpliedResponses(t, method.value)
			if _, ok := objectMember(method.value, "security"); !ok {
				rawSecurity := `[]`
				if firstOperation {
					rawSecurity = `[{"bearerSession":[],"credentialBootstrapBearer":[],"csrfCookie":[],"csrfHeader":[],"sessionCookie":[]}]`
					completeFixtureImpliedResponses(t, method.value)
				}
				security, err := parseOrderedJSON([]byte(rawSecurity), 128)
				if err != nil {
					t.Fatalf("parse fixture security: %v", err)
				}
				method.value.object = append(method.value.object, orderedMember{name: "security", value: security})
			}
			firstOperation = false
		}
	}
}

// fixtureImpliedResponses are the security-implied error responses (Core 01
// REQ-01-702) of the default fixture security, which names every scheme.
const fixtureImpliedResponses = `{"401":{"description":"session","x-cartulary-error-codes":["session_required"]},` +
	`"403":{"description":"csrf","x-cartulary-error-codes":["csrf_verification_failed"]},` +
	`"409":{"description":"bootstrap","x-cartulary-error-codes":["credential_bootstrap_rejected"]}}`

// fixtureOperationImpliedResponses are the error responses every operation
// implies (Core 01 REQ-01-702): 500 internal_error and 503 service_unavailable.
const fixtureOperationImpliedResponses = `{"500":{"description":"internal","x-cartulary-error-codes":["internal_error"]},` +
	`"503":{"description":"unavailable","x-cartulary-error-codes":["service_unavailable"]}}`

// fixtureImpliedRegistryEntries registers the codes every operation implies;
// a fixture that writes its own registry includes them.
const fixtureImpliedRegistryEntries = `{"code":"internal_error","http_status":500},{"code":"service_unavailable","http_status":503}`

// fixtureErrorRegistry registers the security-implied and operation-implied
// codes; a fixture that declares other codes writes its own registry.
const fixtureErrorRegistry = `{"errors":[{"code":"credential_bootstrap_rejected","http_status":409},` +
	`{"code":"csrf_verification_failed","http_status":403},{"code":"session_required","http_status":401},` + fixtureImpliedRegistryEntries + `]}`

// completeFixtureOperationImpliedResponses adds each operation-implied error
// response the operation does not already declare for that status.
func completeFixtureOperationImpliedResponses(t *testing.T, operation *orderedValue) {
	t.Helper()
	responses, ok := objectMember(operation, "responses")
	if !ok || responses.kind != objectKind {
		return
	}
	implied, err := parseOrderedJSON([]byte(fixtureOperationImpliedResponses), 128)
	if err != nil {
		t.Fatalf("parse fixture operation-implied responses: %v", err)
	}
	for _, response := range implied.object {
		if _, declared := objectMember(responses, response.name); !declared {
			responses.object = append(responses.object, response)
		}
	}
}

// completeFixtureImpliedResponses adds each security-implied error response
// the operation does not already declare for that status.
func completeFixtureImpliedResponses(t *testing.T, operation *orderedValue) {
	t.Helper()
	responses, ok := objectMember(operation, "responses")
	if !ok || responses.kind != objectKind {
		return
	}
	implied, err := parseOrderedJSON([]byte(fixtureImpliedResponses), 128)
	if err != nil {
		t.Fatalf("parse fixture implied responses: %v", err)
	}
	for _, response := range implied.object {
		if _, declared := objectMember(responses, response.name); !declared {
			responses.object = append(responses.object, response)
		}
	}
}

func sortedMapKeysForTest(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func fixtureObject(t *testing.T, value any, label string) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s must be an object, got %T", label, value)
	}
	return object
}

func requireAPIKeyScheme(t *testing.T, schemes map[string]any, name, location, wireName string) {
	t.Helper()
	scheme := fixtureObject(t, schemes[name], "security scheme "+name)
	if scheme["type"] != "apiKey" || scheme["in"] != location || scheme["name"] != wireName {
		t.Fatalf("%s has unexpected contract: %#v", name, scheme)
	}
}

func requireBearerScheme(t *testing.T, schemes map[string]any, name string) {
	t.Helper()
	scheme := fixtureObject(t, schemes[name], "security scheme "+name)
	if scheme["type"] != "http" || scheme["scheme"] != "bearer" {
		t.Fatalf("%s has unexpected contract: %#v", name, scheme)
	}
}

func newAssemblyFixture(
	t *testing.T,
	units []unitEntry,
	contents map[string]string,
) *assemblyFixture {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module fixture\n\ngo 1.25\n")
	writeFixtureFile(t, root, "contracts/openapi-source/assembly-policy.json", fixturePolicyJSON(t))
	preparedContents := prepareFixtureContents(t, units, contents)
	for path, content := range preparedContents {
		writeFixtureFile(t, root, path, content)
	}
	if _, provided := preparedContents[errorRegistryPath]; !provided {
		writeFixtureFile(t, root, errorRegistryPath, fixtureErrorRegistry)
	}
	writeFixtureFile(t, root, "contracts/openapi/cartulary.openapi.yaml", "target sentinel\n")
	fixture := &assemblyFixture{
		root:         root,
		manifestPath: filepath.Join(root, "contracts", "openapi-source", "manifest.json"),
		manifest: manifest{
			SchemaID: manifestSchemaID,
			Target:   "contracts/openapi/cartulary.openapi.yaml",
			Policy:   "contracts/openapi-source/assembly-policy.json",
			UnitRoot: "contracts/openapi-source/owners",
			Units:    units,
		},
	}
	fixture.writeManifest(t)
	return fixture
}

func fixturePolicyJSON(t *testing.T) string {
	t.Helper()
	content, err := json.MarshalIndent(assemblyPolicy{
		SchemaID:        policySchemaID,
		Indent:          "  ",
		TrailingNewline: true,
		Limits: resourceLimits{
			MaxUnits:           256,
			MaxUnitBytes:       2 * 1024 * 1024,
			MaxTotalInputBytes: 16 * 1024 * 1024,
			MaxJSONDepth:       128,
			MaxPaths:           2048,
			MaxOperations:      4096,
			MaxNamedComponents: 16384,
		},
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture policy: %v", err)
	}
	return string(append(content, '\n'))
}

func (fixture *assemblyFixture) writeManifest(t *testing.T) {
	t.Helper()
	content, err := json.MarshalIndent(fixture.manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	content = append(content, '\n')
	if err := os.WriteFile(fixture.manifestPath, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func writeFixtureFile(t *testing.T, root, relativePath, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", relativePath, err)
	}
}

func rootOnlyUnits() []unitEntry {
	return []unitEntry{
		{OwnerID: "platform.openapi", Path: rootUnitPath(), Role: unitRootRole},
	}
}

func rootUnitPath() string {
	return "contracts/openapi-source/owners/platform.openapi/root.json"
}

func authUnitPath() string {
	return "contracts/openapi-source/owners/module.auth/auth.json"
}

func deeplyNestedRootUnit(depth int) string {
	value := `"leaf"`
	for index := 0; index < depth; index++ {
		value = `{"nested":` + value + `}`
	}
	return `{"openapi":"3.1.0","info":` + value + `,"paths":{}}`
}

func TestExplicitMutationClassificationAdmission(t *testing.T) {
	for _, value := range []string{"missing", "null", `"false"`, "0", "{}", "[]", "true", "false"} {
		t.Run(value, func(t *testing.T) {
			declaration := ""
			if value != "missing" {
				declaration = `,"x-cartulary-state-changing":` + value
			}
			responses := strings.Replace(fixtureImpliedResponses, `{`, `{"200":{"description":"ok"},`, 1)
			responses = strings.TrimSuffix(responses, `}`) + `,` + strings.TrimPrefix(fixtureOperationImpliedResponses, `{`)
			document := `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"post":{"operationId":"a","responses":` + responses + `,"security":[{"bearerSession":[],"credentialBootstrapBearer":[],"csrfCookie":[],"csrfHeader":[],"sessionCookie":[]}]` + declaration + `}}}}`
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): document})
			writeFixtureFile(t, fixture.root, rootUnitPath(), fixtureDocument(t, document, true, false))
			target := filepath.Join(fixture.root, "contracts", "openapi", "cartulary.openapi.yaml")
			before, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = Assemble(fixture.manifestPath)
			if value == "true" || value == "false" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "explicit boolean x-cartulary-state-changing") {
				t.Fatalf("invalid mutation declaration admitted: %v", err)
			}
			after, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("assembly mutated published target", err)
			}
		})
	}
}
