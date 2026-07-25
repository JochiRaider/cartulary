package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

var openAPIOperationMethods = map[string]struct{}{
	"delete":  {},
	"get":     {},
	"head":    {},
	"options": {},
	"patch":   {},
	"post":    {},
	"put":     {},
	"trace":   {},
}

func TestRuntimeOpenAPICharacterization(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	openAPIOperations := make(map[string]string)
	paths := openAPIObject(t, document["paths"], "paths")
	for path, rawPathItem := range paths {
		pathItem := openAPIObject(t, rawPathItem, "path item "+path)
		for method, rawOperation := range pathItem {
			if _, ok := openAPIOperationMethods[method]; !ok {
				continue
			}
			operation := openAPIObject(t, rawOperation, method+" "+path)
			key := strings.ToUpper(method) + " " + path
			openAPIOperations[key] = openAPIString(operation["operationId"])
		}
	}

	descriptors := productionPublicOperations()
	registry := httpapi.NewPublicOperationRegistry()
	if err := registry.Declare(descriptors...); err != nil {
		t.Fatalf("declare characterized public operations: %v", err)
	}
	requireOpenAPIEqual(t, "characterized runtime operation count", len(registry.Snapshot()), 110)

	characterizedOpenAPI := make(map[string]string)
	for _, operation := range descriptors {
		if err := httpapi.ValidatePublicOperation(operation); err != nil {
			t.Fatalf("validate %s: %v", operation.OperationID, err)
		}
		key := operation.Method + " " + operation.PathTemplate
		publishedOperationID, ok := openAPIOperations[key]
		if !ok {
			t.Fatalf("characterized runtime operation %s is missing from generated OpenAPI", key)
		}
		requireOpenAPIEqual(t, key+" operation ID", operation.OperationID, publishedOperationID)
		characterizedOpenAPI[key] = publishedOperationID
	}
	requireOpenAPIEqual(
		t,
		"OpenAPI runtime characterization",
		formatPublicOperationMap(characterizedOpenAPI),
		formatPublicOperationMap(openAPIOperations),
	)

	for path := range paths {
		switch {
		case strings.HasPrefix(path, "/api/v1/test/"):
			t.Fatalf("test-only route leaked into public OpenAPI: %s", path)
		case strings.HasPrefix(path, "/ws/"):
			t.Fatalf("WebSocket route leaked into public OpenAPI: %s", path)
		case strings.Contains(path, "/network-flow"):
			t.Fatalf("separate Network Flow contract leaked into core OpenAPI: %s", path)
		}
	}

	schemas := openAPIObject(t, openAPIObject(t, document["components"], "components")["schemas"], "components.schemas")
	viewFieldEntry := openAPIObject(t, schemas["ViewFieldEntry"], "ViewFieldEntry")
	gridEditable := openAPIObject(t, openAPIObject(t, viewFieldEntry["properties"], "ViewFieldEntry.properties")["grid_editable"], "ViewFieldEntry.properties.grid_editable")
	if openAPIString(gridEditable["type"]) != "boolean" {
		t.Fatalf("ViewFieldEntry grid_editable must be boolean, got %#v", gridEditable["type"])
	}
	if !slices.Contains(openAPIStringArray(viewFieldEntry["required"]), "grid_editable") {
		t.Fatalf("ViewFieldEntry required omits grid_editable: %#v", viewFieldEntry["required"])
	}
	viewSchemaResource := openAPIObject(t, schemas["ViewSchemaResource"], "ViewSchemaResource")
	inlineCreate := openAPIObject(t, openAPIObject(t, viewSchemaResource["properties"], "ViewSchemaResource.properties")["inline_create"], "ViewSchemaResource.properties.inline_create")
	if openAPIString(inlineCreate["type"]) != "object" || inlineCreate["additionalProperties"] != false {
		t.Fatalf("ViewSchemaResource inline_create must be a closed object, got %#v", inlineCreate)
	}
	if !slices.Equal(openAPIStringArray(inlineCreate["required"]), []string{"minimum_create_field_sets", "permits_zero_field_create"}) {
		t.Fatalf("inline_create required members changed: %#v", inlineCreate["required"])
	}
	if !slices.Contains(openAPIStringArray(viewSchemaResource["required"]), "inline_create") {
		t.Fatalf("ViewSchemaResource required omits inline_create: %#v", viewSchemaResource["required"])
	}
}

func TestRuntimeOpenAPIContractParity(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObject(t, document["paths"], "paths")
	descriptors := productionPublicOperations()
	requireOpenAPIEqual(t, "runtime contract parity operation count", len(descriptors), 110)

	for _, descriptor := range descriptors {
		pathItem := openAPIObject(t, paths[descriptor.PathTemplate], "path item "+descriptor.PathTemplate)
		method := strings.ToLower(descriptor.Method)
		operation := openAPIObject(t, pathItem[method], method+" "+descriptor.PathTemplate)
		requireOpenAPIEqual(t, descriptor.OperationID+" operation ID", openAPIString(operation["operationId"]), descriptor.OperationID)

		wantSecurity := openAPISecurityForPublicOperation(descriptor)
		if !reflect.DeepEqual(operation["security"], wantSecurity) {
			t.Fatalf("%s security mismatch: got %#v want %#v", descriptor.OperationID, operation["security"], wantSecurity)
		}
		responses := openAPIObject(t, operation["responses"], descriptor.OperationID+" responses")
		if len(responses) == 0 {
			t.Fatalf("%s has an empty response contract", descriptor.OperationID)
		}
		primaryStatus := strconv.Itoa(descriptor.SuccessStatus)
		if _, ok := responses[primaryStatus]; !ok {
			t.Fatalf("%s omits runtime primary success status %s", descriptor.OperationID, primaryStatus)
		}
	}
}

func TestProductionPublicRoutesUseRegistrationBoundary(t *testing.T) {
	root := serverRepositoryRoot(t)
	internalRoot := filepath.Join(root, "internal")
	err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "harnesscontrol", "harnessruntime", "testsupport", "testutil":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || (selector.Sel.Name != "Handle" && selector.Sel.Name != "HandleFunc") {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			pattern, err := strconv.Unquote(literal.Value)
			if err != nil || !strings.Contains(pattern, "/api/v1/") || strings.Contains(pattern, "/api/v1/test/") {
				return true
			}
			position := fileSet.Position(call.Pos())
			t.Errorf(
				"%s:%d registers canonical public API route %q outside httpapi.HandlePublicRoute or an explicit exclusion",
				filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))),
				position.Line,
				pattern,
			)
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("inspect production public route registrations: %v", err)
	}
}

func openAPISecurityForPublicOperation(operation httpapi.PublicOperation) []any {
	if operation.Authentication == httpapi.PublicAuthenticationPublic {
		return []any{}
	}
	requirements := make([]any, 0, 3)
	if operation.StateChanging {
		requirements = append(
			requirements,
			map[string]any{"sessionCookie": []any{}, "csrfCookie": []any{}, "csrfHeader": []any{}},
		)
	} else {
		requirements = append(requirements, map[string]any{"sessionCookie": []any{}})
	}
	requirements = append(requirements, map[string]any{"bearerSession": []any{}})
	if operation.Authentication == httpapi.PublicAuthenticationSessionOrBootstrap {
		requirements = append(requirements, map[string]any{"credentialBootstrapBearer": []any{}})
	}
	return requirements
}

func openAPIObject(t *testing.T, raw any, label string) map[string]any {
	t.Helper()
	object, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("%s must be an object, got %T", label, raw)
	}
	return object
}

func openAPIString(raw any) string {
	value, _ := raw.(string)
	return value
}

func openAPIStringArray(raw any) []string {
	values, _ := raw.([]any)
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func formatPublicOperationMap(operations map[string]string) string {
	lines := make([]string, 0, len(operations))
	for key, operationID := range operations {
		lines = append(lines, key+"\t"+operationID)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func serverRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../.."))
}

func requireOpenAPIEqual[T comparable](t *testing.T, label string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}
