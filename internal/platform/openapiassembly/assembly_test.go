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
	if len(parameters) != 8 {
		t.Fatalf("expected eight shared path parameters, got %d", len(parameters))
	}
	for name, raw := range parameters {
		parameter := fixtureObject(t, raw, "components.parameters."+name)
		if parameter["in"] != "path" || parameter["required"] != true {
			t.Fatalf("%s must be a required path parameter, got %#v", name, parameter)
		}
	}

	responses := fixtureObject(t, components["responses"], "components.responses")
	expectedResponses := []string{
		"BadRequestErrorResponse",
		"ConflictErrorResponse",
		"ForbiddenErrorResponse",
		"InternalServerErrorResponse",
		"NotFoundErrorResponse",
		"TooManyRequestsErrorResponse",
		"UnauthorizedErrorResponse",
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

	waiverPath := filepath.Join("..", "..", "..", "contracts", "openapi-source", "compatibility-waivers.json")
	waiverContent, err := os.ReadFile(waiverPath)
	if err != nil {
		t.Fatalf("read compatibility waivers: %v", err)
	}
	var registry compatibilityWaiverRegistry
	if err := json.Unmarshal(waiverContent, &registry); err != nil {
		t.Fatalf("decode compatibility waivers: %v", err)
	}
	if len(registry.PathParameterWaivers) != 0 {
		t.Fatalf("resolved path-parameter waivers remain: %#v", registry.PathParameterWaivers)
	}
	if countOperationWaivers(registry.ResponseWaivers) != 0 {
		t.Fatalf("response waiver inventory must remain empty")
	}
	if countOperationWaivers(registry.SecurityClassificationWaivers) != 0 {
		t.Fatalf("security waiver inventory must remain empty")
	}
	schemeWaiverNames := make([]string, 0, len(registry.SecuritySchemeWaivers))
	for _, waiver := range registry.SecuritySchemeWaivers {
		schemeWaiverNames = append(schemeWaiverNames, waiver.Name)
	}
	sort.Strings(schemeWaiverNames)
	if len(schemeWaiverNames) != 0 {
		t.Fatalf("security-scheme waivers remain after credential-bootstrap classification: %#v", schemeWaiverNames)
	}
	if len(registry.ComponentWaivers) != 0 {
		t.Fatalf("component waivers remain after component closure: %#v", registry.ComponentWaivers)
	}
}

func TestAssemblyPreservesManifestAndMemberOrder(t *testing.T) {
	fixture := newAssemblyFixture(t, []fragmentEntry{
		{OwnerID: "platform.openapi", Path: "contracts/openapi-source/owners/platform.openapi/root.json", Role: fragmentRootRole},
		{OwnerID: "module.auth", Path: "contracts/openapi-source/owners/module.auth/auth.json", Role: fragmentOwnerRole},
		{OwnerID: "module.incidents", Path: "contracts/openapi-source/owners/module.incidents/incidents.json", Role: fragmentOwnerRole},
	}, map[string]string{
		"contracts/openapi-source/owners/platform.openapi/root.json":      `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{},"components":{"schemas":{}}}`,
		"contracts/openapi-source/owners/module.auth/auth.json":           `{"paths":{"/z":{"get":{"operationId":"z"}}},"components":{"schemas":{"Z":{"type":"string"}}}}`,
		"contracts/openapi-source/owners/module.incidents/incidents.json": `{"paths":{"/a":{"get":{"operationId":"a"}}},"components":{"schemas":{"A":{"type":"string"}}}}`,
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
	if strings.Index(string(first), `"/z"`) > strings.Index(string(first), `"/a"`) {
		t.Fatal("path order does not follow manifest order")
	}
	if strings.Index(string(first), `"Z"`) > strings.Index(string(first), `"A"`) {
		t.Fatal("component order does not follow manifest order")
	}
}

func TestAssemblyRejectsUnsafeOrAmbiguousInputs(t *testing.T) {
	tests := []struct {
		name      string
		fragments []fragmentEntry
		contents  map[string]string
		mutate    func(*testing.T, *assemblyFixture)
		wantError string
	}{
		{
			name:      "duplicate JSON key",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			wantError: "duplicate object key",
		},
		{
			name:      "duplicate operation ID",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"same"}},"/b":{"get":{"operationId":"same"}}}}`,
			},
			wantError: "duplicate operationId",
		},
		{
			name: "path method collision",
			fragments: []fragmentEntry{
				{OwnerID: "platform.openapi", Path: rootFragmentPath(), Role: fragmentRootRole},
				{OwnerID: "module.auth", Path: authFragmentPath(), Role: fragmentOwnerRole},
			},
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"root"}}}}`,
				authFragmentPath(): `{"paths":{"/a":{"get":{"operationId":"owner"}}}}`,
			},
			wantError: "path/method collision",
		},
		{
			name:      "unresolved reference",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"$ref":"#/components/responses/Missing"}}}}}}`,
			},
			wantError: "unresolved reference",
		},
		{
			name:      "unwaived placeholder",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a/{id}":{"get":{"operationId":"a"}}}}`,
			},
			wantError: "without exact parameter declarations",
		},
		{
			name: "unknown owner",
			fragments: []fragmentEntry{
				{OwnerID: "platform.openapi", Path: rootFragmentPath(), Role: fragmentRootRole},
				{OwnerID: "module.unknown", Path: authFragmentPath(), Role: fragmentOwnerRole},
			},
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authFragmentPath(): `{"paths":{"/a":{"get":{"operationId":"a"}}}}`,
			},
			wantError: "unknown active owner",
		},
		{
			name: "owner directory mismatch",
			fragments: []fragmentEntry{
				{OwnerID: "platform.openapi", Path: rootFragmentPath(), Role: fragmentRootRole},
				{OwnerID: "module.incidents", Path: authFragmentPath(), Role: fragmentOwnerRole},
			},
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authFragmentPath(): `{"paths":{"/a":{"get":{"operationId":"a"}}}}`,
			},
			wantError: "declared owner directory",
		},
		{
			name: "retired bootstrap role",
			fragments: []fragmentEntry{
				{OwnerID: "platform.openapi", Path: rootFragmentPath(), Role: fragmentRootRole},
				{OwnerID: "module.auth", Path: authFragmentPath(), Role: "bootstrap"},
			},
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
				authFragmentPath(): `{"paths":{"/a":{"get":{"operationId":"a"}}}}`,
			},
			wantError: "invalid role",
		},
		{
			name:      "orphan fragment",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				writeFixtureFile(t, fixture.root, "contracts/openapi-source/owners/module.auth/orphan.json", `{}`)
			},
			wantError: "orphan fragment",
		},
		{
			name:      "absolute fragment path",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				fixture.manifest.Fragments[0].Path = filepath.Join(fixture.root, rootFragmentPath())
				fixture.writeManifest(t)
			},
			wantError: "relative slash path",
		},
		{
			name:      "traversal fragment path",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				fixture.manifest.Fragments[0].Path = "../root.json"
				fixture.writeManifest(t)
			},
			wantError: "escapes its base directory",
		},
		{
			name:      "symlink fragment",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
			},
			mutate: func(t *testing.T, fixture *assemblyFixture) {
				t.Helper()
				path := filepath.Join(fixture.root, filepath.FromSlash(rootFragmentPath()))
				target := filepath.Join(fixture.root, "outside.json")
				if err := os.WriteFile(target, []byte(`{}`), 0o644); err != nil {
					t.Fatalf("write symlink target: %v", err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatalf("remove fragment: %v", err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatalf("create symlink: %v", err)
				}
			},
			wantError: "symlink",
		},
		{
			name:      "depth limit",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): deeplyNestedRootFragment(130),
			},
			wantError: "JSON depth exceeds 128",
		},
		{
			name:      "fragment byte limit",
			fragments: rootOnlyFragments(),
			contents: map[string]string{
				rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"` + strings.Repeat("x", 2*1024*1024) + `","version":"1"},"paths":{}}`,
			},
			wantError: "exceeds max_fragment_bytes",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, test.fragments, test.contents)
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

func TestCompletenessFoundationRejectsUnwaivedOrMalformedOperations(t *testing.T) {
	tests := []struct {
		name                string
		document            string
		responseWaivers     []string
		securityWaivers     []string
		securitySchemeNames []string
		wantError           string
	}{
		{
			name:      "new operation without responses",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","security":[]}}}}`,
			wantError: "operation requires responses",
		},
		{
			name:      "empty responses",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{},"security":[]}}}}`,
			wantError: "responses must be a nonempty object",
		},
		{
			name:      "new operation without security",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}}}}}}`,
			wantError: "operation requires explicit security",
		},
		{
			name:      "unknown security scheme",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}},"security":[{"unknown":[]}]}}}}`,
			wantError: "unknown scheme",
		},
		{
			name:      "nonempty security scopes",
			document:  `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}},"security":[{"bearerSession":["invalid"]}]}}}}`,
			wantError: "scopes must be an empty array",
		},
		{
			name:            "stale response waiver",
			document:        `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}},"security":[]}}}}`,
			responseWaivers: []string{"a"},
			wantError:       "stale response waiver",
		},
		{
			name:                "stale used scheme waiver",
			document:            `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}},"security":[{"bearerSession":[]}]}}}}`,
			securitySchemeNames: append([]string(nil), requiredSecuritySchemeNames...),
			wantError:           "stale security-scheme waiver for used scheme",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyFragments(), map[string]string{
				rootFragmentPath(): test.document,
			})
			schemeNames := test.securitySchemeNames
			if schemeNames == nil {
				schemeNames = requiredSecuritySchemeNames
			}
			writeFixtureFile(
				t,
				fixture.root,
				"contracts/openapi-source/compatibility-waivers.json",
				fixtureWaiverRegistryJSON(
					t,
					test.responseWaivers,
					test.securityWaivers,
					schemeNames,
				),
			)
			_, _, err := Assemble(fixture.manifestPath)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("assemble error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestAssemblyFailureDoesNotTouchTarget(t *testing.T) {
	fixture := newAssemblyFixture(t, rootOnlyFragments(), map[string]string{
		rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a/{id}":{"get":{"operationId":"a"}}}}`,
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
	fixture := newAssemblyFixture(t, rootOnlyFragments(), map[string]string{
		rootFragmentPath(): `{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{}}`,
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
		[]byte(`{"openapi":"3.1.0","info":{"title":"x","version":"1"},"paths":{"/a":{"get":{"operationId":"a","responses":{"200":{"description":"ok"}},"security":[]}}},"components":{"schemas":{"A":{}},"securitySchemes":{"bearerSession":{},"credentialBootstrapBearer":{},"csrfCookie":{},"csrfHeader":{},"sessionCookie":{}}}}`),
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
		{name: "components", limits: resourceLimits{MaxPaths: 10, MaxOperations: 10, MaxNamedComponents: 0}, wantError: "max_named_components"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAggregate(cloneValue(document), test.limits, compatibilityWaivers{
				securitySchemes: map[string]struct{}{
					"bearerSession":             {},
					"credentialBootstrapBearer": {},
					"csrfCookie":                {},
					"csrfHeader":                {},
					"sessionCookie":             {},
				},
			})
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
	fragments []fragmentEntry,
	contents map[string]string,
) map[string]string {
	t.Helper()
	prepared := make(map[string]string, len(contents))
	for path, content := range contents {
		prepared[path] = content
	}
	for _, fragment := range fragments {
		if fragment.Role != fragmentRootRole {
			continue
		}
		content, ok := prepared[fragment.Path]
		if !ok {
			continue
		}
		document, err := parseOrderedJSON([]byte(content), 128)
		if err != nil || document.kind != objectKind {
			continue
		}
		components, ok := objectMember(document, "components")
		if !ok {
			components = &orderedValue{kind: objectKind}
			document.object = append(document.object, orderedMember{name: "components", value: components})
		}
		if components.kind != objectKind {
			continue
		}
		if _, ok := objectMember(components, "securitySchemes"); !ok {
			schemes, err := parseOrderedJSON(
				[]byte(`{"bearerSession":{},"credentialBootstrapBearer":{},"csrfCookie":{},"csrfHeader":{},"sessionCookie":{}}`),
				128,
			)
			if err != nil {
				t.Fatalf("parse fixture security schemes: %v", err)
			}
			components.object = append(components.object, orderedMember{name: "securitySchemes", value: schemes})
		}
		var output bytes.Buffer
		writeOrderedJSON(&output, document, "  ", 0)
		prepared[fragment.Path] = output.String()
	}
	return prepared
}

func fixtureCompatibilityWaiverJSON(t *testing.T, contents map[string]string) string {
	t.Helper()
	responseOperationIDSet := make(map[string]struct{})
	securityOperationIDSet := make(map[string]struct{})
	for _, content := range contents {
		document, err := parseOrderedJSON([]byte(content), 128)
		if err != nil {
			continue
		}
		paths, ok := objectMember(document, "paths")
		if !ok || paths.kind != objectKind {
			continue
		}
		for _, path := range paths.object {
			if path.value.kind != objectKind {
				continue
			}
			for _, method := range path.value.object {
				if _, ok := operationKeys[method.name]; !ok || method.value.kind != objectKind {
					continue
				}
				operationID, ok := objectMember(method.value, "operationId")
				if !ok || operationID.kind != stringKind {
					continue
				}
				id := operationID.scalar.(string)
				if _, ok := objectMember(method.value, "responses"); !ok {
					responseOperationIDSet[id] = struct{}{}
				}
				if _, ok := objectMember(method.value, "security"); !ok {
					securityOperationIDSet[id] = struct{}{}
				}
			}
		}
	}
	responseOperationIDs := sortedStringSet(responseOperationIDSet)
	securityOperationIDs := sortedStringSet(securityOperationIDSet)
	return fixtureWaiverRegistryJSON(
		t,
		responseOperationIDs,
		securityOperationIDs,
		requiredSecuritySchemeNames,
	)
}

func fixtureWaiverRegistryJSON(
	t *testing.T,
	responseOperationIDs, securityOperationIDs, securitySchemeNames []string,
) string {
	t.Helper()
	securitySchemeWaivers := make([]namedWaiver, 0, len(securitySchemeNames))
	for _, name := range securitySchemeNames {
		securitySchemeWaivers = append(
			securitySchemeWaivers,
			fixtureNamedWaiver("scheme."+strings.ToLower(name), name),
		)
	}
	registry := compatibilityWaiverRegistry{
		ResponseWaivers:               fixtureOperationWaiverGroups("response.fixture", responseOperationIDs),
		SecurityClassificationWaivers: fixtureOperationWaiverGroups("security.fixture", securityOperationIDs),
		SecuritySchemeWaivers:         securitySchemeWaivers,
		ComponentWaivers:              []namedWaiver{},
		PathParameterWaivers:          []pathParameterWaiverGroup{},
	}
	content, err := json.Marshal(registry)
	if err != nil {
		t.Fatalf("marshal fixture compatibility waivers: %v", err)
	}
	return string(content)
}

func sortedStringSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
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

func countOperationWaivers(groups []operationWaiverGroup) int {
	total := 0
	for _, group := range groups {
		total += len(group.OperationIDs)
	}
	return total
}

func fixtureOperationWaiverGroups(waiverID string, operationIDs []string) []operationWaiverGroup {
	if len(operationIDs) == 0 {
		return []operationWaiverGroup{}
	}
	return []operationWaiverGroup{{
		WaiverID:         waiverID,
		OwnerID:          "platform.openapi",
		CorrectionID:     "OAPI-CORR-06A",
		Reason:           "Test fixture compatibility waiver.",
		RemovalCondition: "Remove with the test fixture gap.",
		OperationIDs:     operationIDs,
	}}
}

func fixtureNamedWaiver(waiverID, name string) namedWaiver {
	return namedWaiver{
		WaiverID:         waiverID,
		OwnerID:          "platform.openapi",
		CorrectionID:     "OAPI-CORR-06A",
		Reason:           "Test fixture security-scheme waiver.",
		RemovalCondition: "Remove when the fixture uses the scheme.",
		Name:             name,
	}
}

func newAssemblyFixture(
	t *testing.T,
	fragments []fragmentEntry,
	contents map[string]string,
) *assemblyFixture {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module fixture\n\ngo 1.25\n")
	writeFixtureFile(
		t,
		root,
		"contracts/requirements/registry.json",
		`{"owners":[{"owner_id":"platform.openapi","status":"active"},{"owner_id":"module.auth","status":"active"},{"owner_id":"module.incidents","status":"active"}]}`,
	)
	writeFixtureFile(
		t,
		root,
		"contracts/openapi-source/compatibility-waivers.json",
		fixtureCompatibilityWaiverJSON(t, contents),
	)
	preparedContents := prepareFixtureContents(t, fragments, contents)
	for path, content := range preparedContents {
		writeFixtureFile(t, root, path, content)
	}
	writeFixtureFile(t, root, "contracts/openapi/cartulary.openapi.yaml", "target sentinel\n")
	fixture := &assemblyFixture{
		root:         root,
		manifestPath: filepath.Join(root, "contracts", "openapi-source", "manifest.json"),
		manifest: manifest{
			SchemaID:             manifestSchemaID,
			Target:               "contracts/openapi/cartulary.openapi.yaml",
			RequirementsRegistry: "contracts/requirements/registry.json",
			CompatibilityWaivers: "contracts/openapi-source/compatibility-waivers.json",
			FragmentRoot:         "contracts/openapi-source/owners",
			Indent:               "  ",
			TrailingNewline:      true,
			Limits: resourceLimits{
				MaxFragments:       256,
				MaxFragmentBytes:   2 * 1024 * 1024,
				MaxTotalInputBytes: 16 * 1024 * 1024,
				MaxJSONDepth:       128,
				MaxPaths:           2048,
				MaxOperations:      4096,
				MaxNamedComponents: 16384,
			},
			Fragments: fragments,
		},
	}
	fixture.writeManifest(t)
	return fixture
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

func rootOnlyFragments() []fragmentEntry {
	return []fragmentEntry{
		{OwnerID: "platform.openapi", Path: rootFragmentPath(), Role: fragmentRootRole},
	}
}

func rootFragmentPath() string {
	return "contracts/openapi-source/owners/platform.openapi/root.json"
}

func authFragmentPath() string {
	return "contracts/openapi-source/owners/module.auth/auth.json"
}

func deeplyNestedRootFragment(depth int) string {
	value := `"leaf"`
	for index := 0; index < depth; index++ {
		value = `{"nested":` + value + `}`
	}
	return `{"openapi":"3.1.0","info":` + value + `,"paths":{}}`
}
