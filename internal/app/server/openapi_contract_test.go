package server

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

var openAPIOperationMethods = map[string]struct{}{
	"delete": {},
	"get":    {},
	"patch":  {},
	"post":   {},
	"put":    {},
}

func TestGeneratedRuntimeOpenAPIContractParity(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObject(t, document["paths"], "paths")
	operations := httpapi.ContractOperations()
	seen := make(map[string]struct{}, len(operations))

	for _, descriptor := range operations {
		pathItem := openAPIObject(t, paths[descriptor.PathTemplate], "path item "+descriptor.PathTemplate)
		operation := openAPIObject(t, pathItem[strings.ToLower(descriptor.Method)], descriptor.Pattern)
		if got := openAPIString(operation["operationId"]); got != descriptor.OperationID {
			t.Fatalf("%s operation ID = %q, want %q", descriptor.Pattern, got, descriptor.OperationID)
		}
		availability := openAPIString(operation["x-cartulary-availability"])
		if availability == "" {
			availability = httpapi.BaseOperationAvailability
		}
		if availability != descriptor.Availability {
			t.Fatalf("%s availability = %q, want %q", descriptor.OperationID, availability, descriptor.Availability)
		}
		if got := openAPISecurity(operation["security"]); !reflect.DeepEqual(got, descriptor.Security) {
			t.Fatalf("%s security = %#v, want %#v", descriptor.OperationID, got, descriptor.Security)
		}
		responses := openAPIObject(t, operation["responses"], descriptor.OperationID+" responses")
		for _, status := range descriptor.SuccessStatuses {
			if _, ok := responses[strconv.Itoa(status)]; !ok {
				t.Fatalf("%s omits successful status %d", descriptor.OperationID, status)
			}
		}
		seen[descriptor.Method+" "+descriptor.PathTemplate] = struct{}{}
	}

	for path, rawPathItem := range paths {
		switch {
		case strings.HasPrefix(path, "/api/v1/test/"):
			t.Fatalf("test-only route leaked into public OpenAPI: %s", path)
		case strings.HasPrefix(path, "/ws/"):
			t.Fatalf("WebSocket route leaked into public OpenAPI: %s", path)
		case strings.Contains(path, "/network-flow"):
			t.Fatalf("separate Network Flow contract leaked into core OpenAPI: %s", path)
		}
		for method := range openAPIObject(t, rawPathItem, "path item "+path) {
			if _, ok := openAPIOperationMethods[method]; !ok {
				continue
			}
			key := strings.ToUpper(method) + " " + path
			if _, ok := seen[key]; !ok {
				t.Fatalf("OpenAPI operation %s is absent from the generated runtime catalog", key)
			}
		}
	}
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

func openAPISecurity(raw any) [][]string {
	alternatives, _ := raw.([]any)
	result := make([][]string, 0, len(alternatives))
	for _, alternative := range alternatives {
		requirement, _ := alternative.(map[string]any)
		schemes := make([]string, 0, len(requirement))
		for scheme := range requirement {
			schemes = append(schemes, scheme)
		}
		slices.Sort(schemes)
		result = append(result, schemes)
	}
	slices.SortFunc(result, func(left, right []string) int {
		return strings.Compare(strings.Join(left, ","), strings.Join(right, ","))
	})
	return result
}
