package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const (
	networkFlowIndexPath        = "contracts/network-flow/index.json"
	networkFlowRoutesSchemaID   = "cartulary.network_flow_route_contracts.v1"
	coreErrorRegistryPath       = "contracts/errors/index.json"
	networkFlowGateStage        = "incident_visibility_role_lifecycle"
	networkFlowUUIDFormat       = "uuid"
	networkFlowRouteErrorScope  = "route"
	networkFlowRouteOrJobScope  = "route_or_job"
	networkFlowRouteOrRowScope  = "route_or_row_diagnostic"
	networkFlowMinimumNotFound  = 400
	networkFlowMaximumNotFound  = 599
	networkFlowCSRFCookieScheme = "csrfCookie"
	networkFlowCSRFHeaderScheme = "csrfHeader"
)

// networkFlowSecuritySchemes is the closed scheme vocabulary a Network Flow
// route may declare; it names the same schemes as the OpenAPI document.
var networkFlowSecuritySchemes = map[string]struct{}{
	"bearerSession":             {},
	networkFlowCSRFCookieScheme: {},
	networkFlowCSRFHeaderScheme: {},
	"sessionCookie":             {},
}

type networkFlowContractIndex struct {
	OwnerID       string            `json:"owner_id"`
	ContractFiles map[string]string `json:"contract_files"`
}

type networkFlowRouteContract struct {
	SchemaID       string                                  `json:"schema_id"`
	ProfileID      string                                  `json:"profile_id"`
	PathParameters map[string]networkFlowPathParameterSpec `json:"path_parameters"`
	Routes         []struct {
		RouteID             string     `json:"route_id"`
		Method              string     `json:"method"`
		Path                string     `json:"path"`
		Security            [][]string `json:"security"`
		SuccessHTTPStatuses []int      `json:"success_http_statuses"`
	} `json:"routes"`
}

type networkFlowPathParameterSpec struct {
	Format         string `json:"format"`
	SchemaDef      string `json:"schema_def"`
	AdmissionStage string `json:"admission_stage"`
	NotFound       struct {
		Code       string         `json:"code"`
		HTTPStatus int            `json:"http_status"`
		Details    map[string]any `json:"details"`
	} `json:"not_found"`
}

type networkFlowOperationPolicy struct {
	AdmissionStages []string `json:"admission_stages"`
	Operations      []struct {
		RouteID      string `json:"route_id"`
		CSRFRequired bool   `json:"csrf_required"`
	} `json:"operations"`
}

type errorRegistry struct {
	Errors []struct {
		Code       string `json:"code"`
		HTTPStatus *int   `json:"http_status"`
		Scope      string `json:"scope"`
	} `json:"errors"`
}

type schemaBundle struct {
	Defs map[string]struct {
		// Type is a JSON Schema type: a name or an array of names.
		Type    any    `json:"type"`
		Pattern string `json:"pattern"`
	} `json:"$defs"`
}

// networkFlowRoute is one Network Flow operation projected for the public route
// registry. Network Flow keeps its own admission order (NF NLSpec Table 21-C,
// Core 01 REQ-01-707 family clause): its owner gate decides incident
// visibility, role and lifecycle after the parameters of that stage and before
// the resource-path parameters.
type networkFlowRoute struct {
	RouteID         string
	OwnerID         string
	Availability    string
	Method          string
	Path            string
	Pattern         string
	StateChanging   bool
	SuccessStatuses []int
	Security        [][]string
	OwnerGate       bool
	ErrorRegistry   string
	PathParameters  []networkFlowPathParameter
}

type networkFlowPathParameter struct {
	Name            string
	Format          string
	Pattern         string
	AfterOwnerGate  bool
	NotFoundCode    string
	NotFoundStatus  int
	NotFoundDetails map[string]any
}

// networkFlowSources are the contract inputs of the Network Flow projection.
type networkFlowSources struct {
	index    networkFlowContractIndex
	routes   networkFlowRouteContract
	policy   networkFlowOperationPolicy
	schemas  schemaBundle
	nfErrors errorRegistry
	core     errorRegistry
}

func loadNetworkFlowRoutes(root string) ([]networkFlowRoute, error) {
	var sources networkFlowSources
	if err := decodeOpen(filepath.Join(root, networkFlowIndexPath), &sources.index); err != nil {
		return nil, fmt.Errorf("contract index: %w", err)
	}
	files := sources.index.ContractFiles
	for _, input := range []struct {
		key    string
		target any
	}{
		{"routes", &sources.routes},
		{"operation_policy", &sources.policy},
		{"schemas", &sources.schemas},
		{"errors", &sources.nfErrors},
	} {
		path := files[input.key]
		if path == "" {
			return nil, fmt.Errorf("contract index names no %s file", input.key)
		}
		if err := decodeOpen(filepath.Join(root, filepath.FromSlash(path)), input.target); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	if err := decodeOpen(filepath.Join(root, coreErrorRegistryPath), &sources.core); err != nil {
		return nil, fmt.Errorf("core error registry: %w", err)
	}
	return projectNetworkFlowRoutes(sources)
}

// projectNetworkFlowRoutes validates the Network Flow route contract against
// its operation policy, schema bundle and error registries, and projects each
// route's owner, availability, declared security, state-changing flag and
// path-parameter admission in template order.
func projectNetworkFlowRoutes(sources networkFlowSources) ([]networkFlowRoute, error) {
	contract := sources.routes
	if contract.SchemaID != networkFlowRoutesSchemaID {
		return nil, fmt.Errorf("unexpected schema_id %q", contract.SchemaID)
	}
	if sources.index.OwnerID == "" || contract.ProfileID == "" {
		return nil, errors.New("network flow owner and profile are required")
	}
	errorRegistryPath := sources.index.ContractFiles["errors"]
	stages := make(map[string]int, len(sources.policy.AdmissionStages))
	for index, stage := range sources.policy.AdmissionStages {
		stages[stage] = index
	}
	gate, ok := stages[networkFlowGateStage]
	if !ok {
		return nil, fmt.Errorf("operation policy has no %s admission stage", networkFlowGateStage)
	}
	csrf := make(map[string]bool, len(sources.policy.Operations))
	for _, operation := range sources.policy.Operations {
		if _, duplicate := csrf[operation.RouteID]; duplicate {
			return nil, fmt.Errorf("duplicate operation policy for %q", operation.RouteID)
		}
		csrf[operation.RouteID] = operation.CSRFRequired
	}
	registered := registeredNotFoundStatuses(sources.nfErrors, sources.core)
	parameters := make(map[string]networkFlowPathParameter, len(contract.PathParameters))
	for name, spec := range contract.PathParameters {
		parameter, err := projectNetworkFlowPathParameter(name, spec, sources.schemas, registered, stages, gate)
		if err != nil {
			return nil, err
		}
		parameters[name] = parameter
	}
	used := make(map[string]struct{}, len(parameters))
	seenIDs := make(map[string]struct{}, len(contract.Routes))
	seenPatterns := make(map[string]struct{}, len(contract.Routes))
	routes := make([]networkFlowRoute, 0, len(contract.Routes))
	for _, route := range contract.Routes {
		pattern := route.Method + " " + route.Path
		if route.RouteID == "" || route.Method == "" || route.Path == "" || len(route.SuccessHTTPStatuses) == 0 {
			return nil, errors.New("network flow route metadata is incomplete")
		}
		if _, duplicate := seenIDs[route.RouteID]; duplicate {
			return nil, fmt.Errorf("duplicate network flow route ID %q", route.RouteID)
		}
		if _, duplicate := seenPatterns[pattern]; duplicate {
			return nil, fmt.Errorf("duplicate network flow route pattern %q", pattern)
		}
		seenIDs[route.RouteID] = struct{}{}
		seenPatterns[pattern] = struct{}{}
		stateChanging, ok := csrf[route.RouteID]
		if !ok {
			return nil, fmt.Errorf("route %q has no operation policy", route.RouteID)
		}
		if err := validateNetworkFlowSecurity(route.Security, stateChanging); err != nil {
			return nil, fmt.Errorf("route %q security: %w", route.RouteID, err)
		}
		projected := networkFlowRoute{
			RouteID: route.RouteID, OwnerID: sources.index.OwnerID, Availability: contract.ProfileID,
			Method: route.Method, Path: route.Path, Pattern: pattern, StateChanging: stateChanging,
			SuccessStatuses: append([]int(nil), route.SuccessHTTPStatuses...),
			Security:        cloneStringMatrix(route.Security), ErrorRegistry: errorRegistryPath,
		}
		previousStage := -1
		for _, name := range templatePlaceholders(route.Path) {
			parameter, declared := parameters[name]
			if !declared {
				return nil, fmt.Errorf("route %q path parameter %q has no declaration", route.RouteID, name)
			}
			stage := stages[contract.PathParameters[name].AdmissionStage]
			if stage < previousStage {
				return nil, fmt.Errorf("route %q path parameter %q precedes an earlier admission stage", route.RouteID, name)
			}
			previousStage = stage
			if stage == gate {
				projected.OwnerGate = true
			}
			used[name] = struct{}{}
			projected.PathParameters = append(projected.PathParameters, cloneNetworkFlowPathParameter(parameter))
		}
		routes = append(routes, projected)
	}
	if len(csrf) != len(routes) {
		return nil, errors.New("operation policy names routes outside the route contract")
	}
	for name := range parameters {
		if _, ok := used[name]; !ok {
			return nil, fmt.Errorf("path parameter %q is declared but no route uses it", name)
		}
	}
	sort.Slice(routes, func(left, right int) bool {
		return routes[left].RouteID < routes[right].RouteID
	})
	return routes, nil
}

// registeredNotFoundStatuses maps each route-level code to its registered
// status: Network Flow codes from its own registry, Core codes from the
// canonical registry.
func registeredNotFoundStatuses(networkFlow, core errorRegistry) map[string]int {
	registered := make(map[string]int)
	for _, entry := range core.Errors {
		if entry.HTTPStatus != nil {
			registered[entry.Code] = *entry.HTTPStatus
		}
	}
	for _, entry := range networkFlow.Errors {
		switch entry.Scope {
		case networkFlowRouteErrorScope, networkFlowRouteOrJobScope, networkFlowRouteOrRowScope:
			if entry.HTTPStatus != nil {
				registered[entry.Code] = *entry.HTTPStatus
			}
		}
	}
	return registered
}

func projectNetworkFlowPathParameter(
	name string,
	spec networkFlowPathParameterSpec,
	schemas schemaBundle,
	registered map[string]int,
	stages map[string]int,
	gate int,
) (networkFlowPathParameter, error) {
	parameter := networkFlowPathParameter{Name: name}
	stage, ok := stages[spec.AdmissionStage]
	if !ok || spec.AdmissionStage == "" {
		return parameter, fmt.Errorf("path parameter %q admission stage %q is not an operation-policy stage", name, spec.AdmissionStage)
	}
	if stage < gate {
		return parameter, fmt.Errorf("path parameter %q admission stage %q precedes the %s stage", name, spec.AdmissionStage, networkFlowGateStage)
	}
	parameter.AfterOwnerGate = stage > gate
	switch {
	case spec.Format != "" && spec.SchemaDef != "":
		return parameter, fmt.Errorf("path parameter %q must declare exactly one of format and schema_def", name)
	case spec.Format == networkFlowUUIDFormat:
		parameter.Format = spec.Format
	case spec.Format != "":
		return parameter, fmt.Errorf("path parameter %q format %q is not supported", name, spec.Format)
	case spec.SchemaDef != "":
		definition, ok := schemas.Defs[spec.SchemaDef]
		if !ok || definition.Type != "string" || definition.Pattern == "" {
			return parameter, fmt.Errorf("path parameter %q schema_def %q is not a string pattern definition", name, spec.SchemaDef)
		}
		if !strings.HasPrefix(definition.Pattern, "^") || !strings.HasSuffix(definition.Pattern, "$") {
			return parameter, fmt.Errorf("path parameter %q pattern %q must be anchored", name, definition.Pattern)
		}
		if _, err := regexp.Compile(definition.Pattern); err != nil {
			return parameter, fmt.Errorf("path parameter %q pattern: %w", name, err)
		}
		parameter.Pattern = definition.Pattern
	default:
		return parameter, fmt.Errorf("path parameter %q must declare exactly one of format and schema_def", name)
	}
	notFound := spec.NotFound
	status, ok := registered[notFound.Code]
	if notFound.Code == "" || !ok {
		return parameter, fmt.Errorf("path parameter %q not-found code %q is not registered", name, notFound.Code)
	}
	if notFound.HTTPStatus != status || status < networkFlowMinimumNotFound || status > networkFlowMaximumNotFound {
		return parameter, fmt.Errorf("path parameter %q not-found code %q is registered at %d, not %d", name, notFound.Code, status, notFound.HTTPStatus)
	}
	if notFound.Details == nil {
		return parameter, fmt.Errorf("path parameter %q not-found details must be an object", name)
	}
	parameter.NotFoundCode, parameter.NotFoundStatus = notFound.Code, status
	parameter.NotFoundDetails = cloneJSONObject(notFound.Details)
	return parameter, nil
}

// validateNetworkFlowSecurity requires a canonical, non-empty, session-only
// declaration that pairs cookie CSRF exactly when the operation policy marks
// the route state-changing.
func validateNetworkFlowSecurity(security [][]string, stateChanging bool) error {
	if len(security) == 0 {
		return errors.New("must declare at least one alternative")
	}
	csrf := false
	for index, alternative := range security {
		if len(alternative) == 0 {
			return errors.New("an anonymous alternative is not admitted")
		}
		if !slices.IsSorted(alternative) || len(slices.Compact(slices.Clone(alternative))) != len(alternative) {
			return fmt.Errorf("alternative %d must be sorted and duplicate-free", index+1)
		}
		for _, scheme := range alternative {
			if _, known := networkFlowSecuritySchemes[scheme]; !known {
				return fmt.Errorf("unknown scheme %q", scheme)
			}
		}
		if slices.Contains(alternative, networkFlowCSRFCookieScheme) || slices.Contains(alternative, networkFlowCSRFHeaderScheme) {
			csrf = true
		}
		if index > 0 && strings.Join(security[index-1], "\x00") >= strings.Join(alternative, "\x00") {
			return errors.New("alternatives must be sorted and duplicate-free")
		}
	}
	if csrf != stateChanging {
		return fmt.Errorf("cookie CSRF is declared=%t but the operation policy requires %t", csrf, stateChanging)
	}
	return nil
}

func cloneStringMatrix(values [][]string) [][]string {
	result := make([][]string, len(values))
	for index := range values {
		result[index] = append([]string(nil), values[index]...)
	}
	return result
}

func cloneNetworkFlowPathParameter(parameter networkFlowPathParameter) networkFlowPathParameter {
	parameter.NotFoundDetails = cloneJSONObject(parameter.NotFoundDetails)
	return parameter
}

func cloneJSONObject(object map[string]any) map[string]any {
	result := make(map[string]any, len(object))
	for key, value := range object {
		result[key] = cloneJSONValue(value)
	}
	return result
}

func cloneJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneJSONObject(typed)
	case []any:
		result := make([]any, len(typed))
		for index := range typed {
			result[index] = cloneJSONValue(typed[index])
		}
		return result
	default:
		return typed
	}
}

// renderJSONLiteral renders a decoded JSON value as a Go literal of type any.
func renderJSONLiteral(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "nil", nil
	case bool:
		return strconv.FormatBool(typed), nil
	case string:
		return strconv.Quote(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'g', -1, 64), nil
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			rendered, err := renderJSONLiteral(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, rendered)
		}
		return "[]any{" + strings.Join(parts, ", ") + "}", nil
	case map[string]any:
		return renderJSONObjectLiteral(typed)
	default:
		return "", fmt.Errorf("unsupported JSON value %T", value)
	}
}

func renderJSONObjectLiteral(object map[string]any) (string, error) {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	members := make([]string, 0, len(keys))
	for _, key := range keys {
		rendered, err := renderJSONLiteral(object[key])
		if err != nil {
			return "", err
		}
		members = append(members, strconv.Quote(key)+": "+rendered)
	}
	return "map[string]any{" + strings.Join(members, ", ") + "}", nil
}

func renderNetworkFlowPathParameters(parameters []networkFlowPathParameter) (string, error) {
	if len(parameters) == 0 {
		return "nil", nil
	}
	parts := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		details, err := renderJSONObjectLiteral(parameter.NotFoundDetails)
		if err != nil {
			return "", fmt.Errorf("path parameter %q details: %w", parameter.Name, err)
		}
		parts = append(parts, fmt.Sprintf(
			"{Name: %q, Format: %q, Pattern: %q, AfterOwnerGate: %t, NotFoundCode: %q, NotFoundStatus: %d, NotFoundDetails: %s}",
			parameter.Name, parameter.Format, parameter.Pattern, parameter.AfterOwnerGate,
			parameter.NotFoundCode, parameter.NotFoundStatus, details,
		))
	}
	return "[]PathParameter{" + strings.Join(parts, ", ") + "}", nil
}

func renderNetworkFlow(routes []networkFlowRoute) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("// Code generated by tools/openapi-operation-catalog; DO NOT EDIT.\n\n")
	output.WriteString("package networkflowroutes\n\n")
	output.WriteString("// Route is one Network Flow operation projected from its route contract for\n")
	output.WriteString("// the public route registry: owner, availability, declared security and\n")
	output.WriteString("// path-parameter admission in template order. OwnerGate marks Network Flow's\n")
	output.WriteString("// own admission order (NF NLSpec Table 21-C): the owner gate runs after the\n")
	output.WriteString("// parameters before it and before those marked AfterOwnerGate. Error codes are\n")
	output.WriteString("// governed by ErrorRegistry, not by OpenAPI declarations.\n")
	output.WriteString("type Route struct {\n\tRouteID string\n\tOwnerID string\n\tAvailability string\n\tMethod string\n\tPath string\n\tPattern string\n")
	output.WriteString("\tStateChanging bool\n\tSuccessStatuses []int\n\tSecurity [][]string\n\tOwnerGate bool\n\tErrorRegistry string\n\tPathParameters []PathParameter\n}\n\n")
	output.WriteString("// PathParameter is one path parameter's declared format or pattern, its\n")
	output.WriteString("// position relative to the owner gate, and the not-found response a malformed\n")
	output.WriteString("// value receives.\n")
	output.WriteString("type PathParameter struct {\n\tName string\n\tFormat string\n\tPattern string\n\tAfterOwnerGate bool\n")
	output.WriteString("\tNotFoundCode string\n\tNotFoundStatus int\n\tNotFoundDetails map[string]any\n}\n\n")
	output.WriteString("var catalog = []Route{\n")
	for _, route := range routes {
		parameters, err := renderNetworkFlowPathParameters(route.PathParameters)
		if err != nil {
			return nil, fmt.Errorf("route %q: %w", route.RouteID, err)
		}
		fmt.Fprintf(
			&output,
			"\t{RouteID: %q, OwnerID: %q, Availability: %q, Method: %q, Path: %q, Pattern: %q, StateChanging: %t, SuccessStatuses: %#v, Security: %#v, OwnerGate: %t, ErrorRegistry: %q, PathParameters: %s},\n",
			route.RouteID, route.OwnerID, route.Availability, route.Method, route.Path, route.Pattern,
			route.StateChanging, route.SuccessStatuses, route.Security, route.OwnerGate, route.ErrorRegistry, parameters,
		)
	}
	output.WriteString("}\n\n")
	output.WriteString("// All returns a deep copy of the catalog.\n")
	output.WriteString("func All() []Route {\n\tresult := make([]Route, len(catalog))\n\tcopy(result, catalog)\n")
	output.WriteString("\tfor index := range result {\n")
	output.WriteString("\t\tresult[index].SuccessStatuses = append([]int(nil), catalog[index].SuccessStatuses...)\n")
	output.WriteString("\t\tresult[index].Security = make([][]string, len(catalog[index].Security))\n")
	output.WriteString("\t\tfor alternative := range catalog[index].Security {\n")
	output.WriteString("\t\t\tresult[index].Security[alternative] = append([]string(nil), catalog[index].Security[alternative]...)\n\t\t}\n")
	output.WriteString("\t\tresult[index].PathParameters = append([]PathParameter(nil), catalog[index].PathParameters...)\n")
	output.WriteString("\t\tfor parameter := range result[index].PathParameters {\n")
	output.WriteString("\t\t\tresult[index].PathParameters[parameter].NotFoundDetails = cloneObject(catalog[index].PathParameters[parameter].NotFoundDetails)\n")
	output.WriteString("\t\t}\n\t}\n\treturn result\n}\n\n")
	output.WriteString("func cloneObject(object map[string]any) map[string]any {\n")
	output.WriteString("\tif object == nil {\n\t\treturn nil\n\t}\n")
	output.WriteString("\tresult := make(map[string]any, len(object))\n")
	output.WriteString("\tfor key, value := range object {\n\t\tresult[key] = cloneValue(value)\n\t}\n\treturn result\n}\n\n")
	output.WriteString("func cloneValue(value any) any {\n\tswitch typed := value.(type) {\n")
	output.WriteString("\tcase map[string]any:\n\t\treturn cloneObject(typed)\n")
	output.WriteString("\tcase []any:\n\t\tresult := make([]any, len(typed))\n\t\tfor index := range typed {\n\t\t\tresult[index] = cloneValue(typed[index])\n\t\t}\n\t\treturn result\n")
	output.WriteString("\tdefault:\n\t\treturn typed\n\t}\n}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Network Flow Go: %w", err)
	}
	return formatted, nil
}
