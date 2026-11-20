package server

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// Every OpenAPI 3.1 path-item operation method.
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

func TestGeneratedRuntimeOpenAPIContractParity(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObject(t, document["paths"], "paths")
	components := openAPIObject(t, document["components"], "components")
	componentResponses := openAPIObject(t, components["responses"], "component responses")
	operations := httpapi.ContractOperations()
	seen := make(map[string]struct{}, len(operations))
	families := reservedRouteFamilies(t)

	for _, descriptor := range operations {
		pathItem := openAPIObject(t, paths[descriptor.PathTemplate], "path item "+descriptor.PathTemplate)
		operation := openAPIObject(t, pathItem[strings.ToLower(descriptor.Method)], descriptor.Pattern)
		if got := openAPIString(operation["operationId"]); got != descriptor.OperationID {
			t.Fatalf("%s operation ID = %q, want %q", descriptor.Pattern, got, descriptor.OperationID)
		}
		if got, ok := operation["x-cartulary-state-changing"].(bool); !ok || got != descriptor.StateChanging {
			t.Fatalf("%s state-changing = %#v, want %t", descriptor.OperationID, operation["x-cartulary-state-changing"], descriptor.StateChanging)
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
		requireDeclaredErrorCodes(t, descriptor, responses, componentResponses)
		requireImpliedDeclarations(t, descriptor, families)
		requirePathParameters(t, descriptor, pathItem, operation, responses, components)
		key := descriptor.Method + " " + descriptor.PathTemplate
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("runtime catalog repeats %s", key)
		}
		seen[key] = struct{}{}
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

// reservedRouteFamilies maps each route family the generated extension
// registry reserves to its profile: the families the runtime answers with
// extension_profile_not_claimed while their profile is unclaimed.
func reservedRouteFamilies(t *testing.T) map[string]string {
	t.Helper()
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	families := map[string]string{}
	for _, descriptor := range coordinator.Descriptors() {
		for _, family := range descriptor.RouteFamilies {
			families[family] = descriptor.ProfileID
		}
	}
	if len(families) == 0 {
		t.Fatal("the generated extension registry reserves no route family")
	}
	return families
}

// requireImpliedDeclarations covers the declarations every operation implies
// (REQ-01-702, AC-608): 500 internal_error and 503 service_unavailable, and
// 404 extension_profile_not_claimed exactly when the operation's path lies in
// a route family reserved for an extension profile. Assembly identifies those
// operations by x-cartulary-availability, so an operation's availability must
// name exactly the profile whose reserved family holds its path.
func requireImpliedDeclarations(t *testing.T, descriptor httpapi.ContractOperation, families map[string]string) {
	t.Helper()
	for status, code := range map[int]string{500: "internal_error", 503: "service_unavailable"} {
		if !slices.Contains(descriptor.ErrorCodes[status], code) {
			t.Fatalf("%s %d declares %v; every operation implies %s", descriptor.OperationID, status, descriptor.ErrorCodes[status], code)
		}
	}
	profile := httpapi.BaseOperationAvailability
	for family, profileID := range families {
		if descriptor.PathTemplate == family || strings.HasPrefix(descriptor.PathTemplate, family+"/") {
			profile = profileID
		}
	}
	if descriptor.Availability != profile {
		t.Fatalf("%s availability = %q, but its path %s belongs to the reserved family of %q", descriptor.OperationID, descriptor.Availability, descriptor.PathTemplate, profile)
	}
	reserved := profile != httpapi.BaseOperationAvailability
	if declared := slices.Contains(descriptor.ErrorCodes[404], "extension_profile_not_claimed"); declared != reserved {
		t.Fatalf("%s 404 declares %v; extension_profile_not_claimed must be declared exactly in a reserved route family (reserved=%t)", descriptor.OperationID, descriptor.ErrorCodes[404], reserved)
	}
}

// requireDeclaredErrorCodes covers REQ-01-702: every error response declares a
// sorted, duplicate-free list of codes registered with that status, and the
// runtime catalog projects exactly those lists.
func requireDeclaredErrorCodes(t *testing.T, descriptor httpapi.ContractOperation, responses, componentResponses map[string]any) {
	t.Helper()
	declared := map[int][]string{}
	for key, raw := range responses {
		status, err := strconv.Atoi(key)
		if err != nil || status < 400 {
			continue
		}
		response := openAPIObject(t, raw, descriptor.OperationID+" "+key)
		if ref := openAPIString(response["$ref"]); ref != "" {
			response = openAPIObject(t, componentResponses[strings.TrimPrefix(ref, "#/components/responses/")], ref)
		}
		rawCodes, _ := response["x-cartulary-error-codes"].([]any)
		codes := make([]string, 0, len(rawCodes))
		for _, rawCode := range rawCodes {
			codes = append(codes, openAPIString(rawCode))
		}
		if len(codes) == 0 || !slices.IsSorted(codes) || len(slices.Compact(slices.Clone(codes))) != len(codes) {
			t.Fatalf("%s %d declares %v; want a non-empty, sorted, duplicate-free list", descriptor.OperationID, status, codes)
		}
		for _, code := range codes {
			contracttest.RequireErrorContract(t, code, status)
		}
		declared[status] = codes
	}
	if !reflect.DeepEqual(declared, descriptor.ErrorCodes) {
		t.Fatalf("%s catalog error codes = %v, OpenAPI declares %v", descriptor.OperationID, descriptor.ErrorCodes, declared)
	}
}

// requirePathParameters covers REQ-01-707: the runtime catalog projects, in
// template order, each path placeholder's declared format or pattern, length
// bounds and not-found response, exactly as the OpenAPI document declares
// them. The not-found code is registered and declared at its status.
func requirePathParameters(t *testing.T, descriptor httpapi.ContractOperation, pathItem, operation, responses, components map[string]any) {
	t.Helper()
	componentParameters, _ := components["parameters"].(map[string]any)
	componentSchemas, _ := components["schemas"].(map[string]any)
	componentResponses, _ := components["responses"].(map[string]any)
	declared := map[string]map[string]any{}
	for _, container := range []map[string]any{pathItem, operation} {
		rawParameters, _ := container["parameters"].([]any)
		for _, raw := range rawParameters {
			parameter := openAPIObject(t, raw, descriptor.OperationID+" parameter")
			if ref := openAPIString(parameter["$ref"]); ref != "" {
				parameter = openAPIObject(t, componentParameters[strings.TrimPrefix(ref, "#/components/parameters/")], ref)
			}
			if parameter["in"] == "path" {
				declared[openAPIString(parameter["name"])] = parameter
			}
		}
	}
	var want []httpapi.PathParameter
	for _, segment := range strings.Split(descriptor.PathTemplate, "/") {
		if !strings.HasPrefix(segment, "{") || !strings.HasSuffix(segment, "}") {
			continue
		}
		name := segment[1 : len(segment)-1]
		parameter, ok := declared[name]
		if !ok {
			t.Fatalf("%s path parameter %s has no declaration", descriptor.OperationID, name)
		}
		schema := openAPIObject(t, parameter["schema"], descriptor.OperationID+" "+name+" schema")
		if ref := openAPIString(schema["$ref"]); ref != "" {
			schema = openAPIObject(t, componentSchemas[strings.TrimPrefix(ref, "#/components/schemas/")], ref)
		}
		projection := httpapi.PathParameter{
			Name:         name,
			Format:       openAPIString(schema["format"]),
			Pattern:      openAPIString(schema["pattern"]),
			NotFoundCode: openAPIString(parameter["x-cartulary-not-found-code"]),
		}
		if bound, ok := schema["minLength"].(float64); ok {
			projection.MinLength = int(bound)
		}
		if bound, ok := schema["maxLength"].(float64); ok {
			projection.MaxLength = int(bound)
		}
		if (projection.Format == "") == (projection.Pattern == "") || projection.NotFoundCode == "" {
			t.Fatalf("%s path parameter %s must declare one format or pattern and a not-found code", descriptor.OperationID, name)
		}
		for key, raw := range responses {
			status, err := strconv.Atoi(key)
			if err != nil || status < 400 {
				continue
			}
			response := openAPIObject(t, raw, descriptor.OperationID+" "+key)
			if ref := openAPIString(response["$ref"]); ref != "" {
				response = openAPIObject(t, componentResponses[strings.TrimPrefix(ref, "#/components/responses/")], ref)
			}
			codes, _ := response["x-cartulary-error-codes"].([]any)
			if slices.Contains(codes, any(projection.NotFoundCode)) {
				projection.NotFoundStatus = status
			}
		}
		if projection.NotFoundStatus == 0 {
			t.Fatalf("%s path parameter %s not-found code %s is not declared", descriptor.OperationID, name, projection.NotFoundCode)
		}
		contracttest.RequireErrorContract(t, projection.NotFoundCode, projection.NotFoundStatus)
		if rawDetails, present := parameter["x-cartulary-not-found-details"]; present {
			projection.NotFoundDetails = map[string]any{}
			for key, value := range openAPIObject(t, rawDetails, descriptor.OperationID+" "+name+" details") {
				projection.NotFoundDetails[key] = openAPIString(value)
			}
		}
		want = append(want, projection)
	}
	if !reflect.DeepEqual(want, descriptor.PathParameters) {
		t.Fatalf("%s catalog path parameters = %#v, OpenAPI declares %#v", descriptor.OperationID, descriptor.PathParameters, want)
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
