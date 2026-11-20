package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	manifestSchemaID = "cartulary.openapi_source_manifest.v3"
	baseAvailability = "base"
)

type manifest struct {
	SchemaID string `json:"schema_id"`
	Units    []struct {
		OwnerID string `json:"owner_id"`
		Path    string `json:"path"`
	} `json:"units"`
}

type extensionIndex struct {
	Artifacts []struct {
		Path          string `json:"path"`
		ArtifactClass string `json:"artifact_class"`
	} `json:"artifacts"`
}

type profileContract struct {
	ProfileID string `json:"profile_id"`
}

type operation struct {
	OwnerID         string
	Method          string
	PathTemplate    string
	Pattern         string
	OperationID     string
	Availability    string
	StateChanging   bool
	SuccessStatuses []int
	ErrorCodes      map[int][]string
	Security        [][]string
	PathParameters  []pathParameter
}

// pathParameter projects one path parameter's admission declaration (Core 01
// REQ-01-707): its format or pattern with optional length bounds, and the
// not-found response a malformed value receives. A zero MaxLength is
// unbounded.
type pathParameter struct {
	Name            string
	Format          string
	Pattern         string
	MinLength       int
	MaxLength       int
	NotFoundCode    string
	NotFoundStatus  int
	NotFoundDetails map[string]string
}

type catalogMetadata struct {
	DocumentVersion string
	SHA256          string
}

func main() {
	root := flag.String("root", ".", "repository root")
	output := flag.String("output", "internal/gen/openapioperations/catalog_gen.go", "generated Go output")
	networkFlowOutput := flag.String("network-flow-output", "internal/gen/networkflowroutes/catalog_gen.go", "generated Network Flow Go output")
	check := flag.Bool("check", false, "check output instead of writing it")
	flag.Parse()

	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		fail("resolve repository root", err)
	}
	operations, err := loadOperations(absoluteRoot)
	if err != nil {
		fail("load operation catalog", err)
	}
	metadata, err := loadCatalogMetadata(absoluteRoot)
	if err != nil {
		fail("load canonical OpenAPI metadata", err)
	}
	generated, err := render(metadata, operations)
	if err != nil {
		fail("render operation catalog", err)
	}
	outputPath := filepath.Join(absoluteRoot, filepath.FromSlash(*output))
	networkFlowRoutes, err := loadNetworkFlowRoutes(absoluteRoot)
	if err != nil {
		fail("load Network Flow route catalog", err)
	}
	generatedNetworkFlow, err := renderNetworkFlow(networkFlowRoutes)
	if err != nil {
		fail("render Network Flow route catalog", err)
	}
	networkFlowOutputPath := filepath.Join(absoluteRoot, filepath.FromSlash(*networkFlowOutput))
	if *check {
		current, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			fail("read generated operation catalog", readErr)
		}
		if !bytes.Equal(current, generated) {
			fail("check generated operation catalog", errors.New("generated output is stale"))
		}
		currentNetworkFlow, readErr := os.ReadFile(networkFlowOutputPath)
		if readErr != nil {
			fail("read generated Network Flow route catalog", readErr)
		}
		if !bytes.Equal(currentNetworkFlow, generatedNetworkFlow) {
			fail("check generated Network Flow route catalog", errors.New("generated output is stale"))
		}
		return
	}
	if err := writeAtomically(outputPath, generated); err != nil {
		fail("write generated operation catalog", err)
	}
	if err := writeAtomically(networkFlowOutputPath, generatedNetworkFlow); err != nil {
		fail("write generated Network Flow route catalog", err)
	}
}

func loadCatalogMetadata(root string) (catalogMetadata, error) {
	content, err := os.ReadFile(filepath.Join(root, "contracts/openapi/cartulary.openapi.yaml"))
	if err != nil {
		return catalogMetadata{}, err
	}
	var document struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		return catalogMetadata{}, fmt.Errorf("decode canonical document: %w", err)
	}
	if strings.TrimSpace(document.Info.Version) == "" {
		return catalogMetadata{}, errors.New("canonical document info.version is required")
	}
	digest := sha256.Sum256(content)
	return catalogMetadata{
		DocumentVersion: document.Info.Version,
		SHA256:          fmt.Sprintf("%x", digest[:]),
	}, nil
}

func loadOperations(root string) ([]operation, error) {
	var sourceManifest manifest
	if err := decodeOpen(filepath.Join(root, "contracts/openapi-source/manifest.json"), &sourceManifest); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	if sourceManifest.SchemaID != manifestSchemaID {
		return nil, fmt.Errorf("manifest schema_id must be %q", manifestSchemaID)
	}
	profiles, err := loadProfiles(root)
	if err != nil {
		return nil, err
	}
	operationIDs := make(map[string]string)
	routeKeys := make(map[string]string)
	var operations []operation
	documents := make([]map[string]any, len(sourceManifest.Units))
	// Components are merged across units: an operation may reference a
	// response, parameter or schema owned by another unit, such as the root
	// unit's shared errors and path parameters.
	responseComponents := make(map[string]map[string]any)
	parameterComponents := make(map[string]map[string]any)
	schemaComponents := make(map[string]map[string]any)
	for index, unit := range sourceManifest.Units {
		if err := decodeClosed(filepath.Join(root, filepath.FromSlash(unit.Path)), &documents[index]); err != nil {
			return nil, fmt.Errorf("owner unit %s: %w", unit.OwnerID, err)
		}
		components, _ := documents[index]["components"].(map[string]any)
		for category, target := range map[string]map[string]map[string]any{
			"responses":  responseComponents,
			"parameters": parameterComponents,
			"schemas":    schemaComponents,
		} {
			entries, _ := components[category].(map[string]any)
			for name, raw := range entries {
				if entry, ok := raw.(map[string]any); ok {
					target[name] = entry
				}
			}
		}
	}
	for index, unit := range sourceManifest.Units {
		document := documents[index]
		paths, _ := document["paths"].(map[string]any)
		for pathTemplate, rawPathItem := range paths {
			pathItem, ok := rawPathItem.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s path %s must be an object", unit.OwnerID, pathTemplate)
			}
			for rawMethod, rawOperation := range pathItem {
				method := strings.ToUpper(rawMethod)
				if !isOperationMethod(method) {
					continue
				}
				operationObject, ok := rawOperation.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("%s %s must be an object", method, pathTemplate)
				}
				operationID, ok := operationObject["operationId"].(string)
				if !ok || operationID == "" {
					return nil, fmt.Errorf("%s %s has no operationId", method, pathTemplate)
				}
				routeKey := method + " " + pathTemplate
				if previous, duplicate := operationIDs[operationID]; duplicate {
					return nil, fmt.Errorf("duplicate operationId %q at %s and %s", operationID, previous, routeKey)
				}
				if previous, duplicate := routeKeys[routeKey]; duplicate {
					return nil, fmt.Errorf("duplicate operation route %s owned by %s and %s", routeKey, previous, unit.OwnerID)
				}
				availability := baseAvailability
				if rawAvailability, exists := operationObject["x-cartulary-availability"]; exists {
					profileID, ok := rawAvailability.(string)
					if !ok || profileID == "" {
						return nil, fmt.Errorf("%s x-cartulary-availability must be a non-empty string", operationID)
					}
					if _, known := profiles[profileID]; !known {
						return nil, fmt.Errorf("%s references unknown availability profile %q", operationID, profileID)
					}
					availability = profileID
				}
				successStatuses, err := successfulStatuses(operationObject)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", operationID, err)
				}
				errorCodes, err := declaredErrorCodes(operationObject, responseComponents)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", operationID, err)
				}
				security, err := securityDeclaration(operationObject)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", operationID, err)
				}
				stateChanging, err := stateChangingDeclaration(operationObject)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", operationID, err)
				}
				pathParameters, err := projectPathParameters(pathTemplate, pathItem, operationObject, errorCodes, parameterComponents, schemaComponents)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", operationID, err)
				}
				operations = append(operations, operation{
					OwnerID:         unit.OwnerID,
					Method:          method,
					PathTemplate:    pathTemplate,
					Pattern:         routeKey,
					OperationID:     operationID,
					Availability:    availability,
					StateChanging:   stateChanging,
					SuccessStatuses: successStatuses,
					ErrorCodes:      errorCodes,
					Security:        security,
					PathParameters:  pathParameters,
				})
				operationIDs[operationID] = routeKey
				routeKeys[routeKey] = unit.OwnerID
			}
		}
	}
	sort.Slice(operations, func(left, right int) bool {
		if operations[left].PathTemplate != operations[right].PathTemplate {
			return operations[left].PathTemplate < operations[right].PathTemplate
		}
		if operations[left].Method != operations[right].Method {
			return operations[left].Method < operations[right].Method
		}
		return operations[left].OperationID < operations[right].OperationID
	})
	return operations, nil
}

func loadProfiles(root string) (map[string]struct{}, error) {
	var index extensionIndex
	indexPath := filepath.Join(root, "contracts/extensions/index.json")
	if err := decodeOpen(indexPath, &index); err != nil {
		return nil, fmt.Errorf("extension index: %w", err)
	}
	profiles := make(map[string]struct{})
	for _, artifact := range index.Artifacts {
		if artifact.ArtifactClass != "profile_contract" {
			continue
		}
		var contract profileContract
		path := filepath.Join(root, "contracts/extensions", filepath.FromSlash(artifact.Path))
		if err := decodeOpen(path, &contract); err != nil {
			return nil, fmt.Errorf("extension profile %s: %w", artifact.Path, err)
		}
		if contract.ProfileID == "" {
			continue
		}
		profiles[contract.ProfileID] = struct{}{}
	}
	return profiles, nil
}

func successfulStatuses(operationObject map[string]any) ([]int, error) {
	responses, ok := operationObject["responses"].(map[string]any)
	if !ok {
		return nil, errors.New("responses must be an object")
	}
	var statuses []int
	for rawStatus := range responses {
		status, err := strconv.Atoi(rawStatus)
		if err == nil && status >= 200 && status < 400 {
			statuses = append(statuses, status)
		}
	}
	sort.Ints(statuses)
	if len(statuses) == 0 {
		return nil, errors.New("operation has no successful response status")
	}
	return statuses, nil
}

// declaredErrorCodes projects each error response's x-cartulary-error-codes
// (Core 01 REQ-01-702), resolving component response references.
func declaredErrorCodes(operationObject map[string]any, components map[string]map[string]any) (map[int][]string, error) {
	responses, _ := operationObject["responses"].(map[string]any)
	declared := make(map[int][]string)
	for rawStatus, rawResponse := range responses {
		status, err := strconv.Atoi(rawStatus)
		if err != nil || status < 400 {
			continue
		}
		response, _ := rawResponse.(map[string]any)
		if ref, ok := response["$ref"].(string); ok {
			name := strings.TrimPrefix(ref, "#/components/responses/")
			resolved, known := components[name]
			if name == ref || !known {
				return nil, fmt.Errorf("response %d references unknown component %q", status, ref)
			}
			response = resolved
		}
		rawCodes, present := response["x-cartulary-error-codes"]
		if !present {
			continue
		}
		list, ok := rawCodes.([]any)
		if !ok {
			return nil, fmt.Errorf("response %d x-cartulary-error-codes must be an array", status)
		}
		codes := make([]string, 0, len(list))
		for _, rawCode := range list {
			code, ok := rawCode.(string)
			if !ok || code == "" {
				return nil, fmt.Errorf("response %d declares a non-string error code", status)
			}
			codes = append(codes, code)
		}
		declared[status] = codes
	}
	return declared, nil
}

// projectPathParameters projects the path template's parameters in template
// order (Core 01 REQ-01-707). An operation-level declaration overrides a
// path-item declaration of the same name. The not-found status is the status
// whose declared error codes name the parameter's not-found code.
func projectPathParameters(
	pathTemplate string,
	pathItem, operationObject map[string]any,
	errorCodes map[int][]string,
	parameterComponents, schemaComponents map[string]map[string]any,
) ([]pathParameter, error) {
	declared := make(map[string]map[string]any)
	for _, container := range []map[string]any{pathItem, operationObject} {
		rawParameters, _ := container["parameters"].([]any)
		for _, raw := range rawParameters {
			parameter, ok := raw.(map[string]any)
			if !ok {
				return nil, errors.New("parameter must be an object")
			}
			if ref, referenced := parameter["$ref"].(string); referenced {
				name := strings.TrimPrefix(ref, "#/components/parameters/")
				resolved, known := parameterComponents[name]
				if name == ref || !known {
					return nil, fmt.Errorf("parameter references unknown component %q", ref)
				}
				parameter = resolved
			}
			if parameter["in"] == "path" {
				name, _ := parameter["name"].(string)
				declared[name] = parameter
			}
		}
	}
	placeholders := templatePlaceholders(pathTemplate)
	if len(placeholders) == 0 {
		return nil, nil
	}
	projected := make([]pathParameter, 0, len(placeholders))
	for _, name := range placeholders {
		parameter, ok := declared[name]
		if !ok {
			return nil, fmt.Errorf("path parameter %q has no declaration", name)
		}
		schema, _ := parameter["schema"].(map[string]any)
		if ref, referenced := schema["$ref"].(string); referenced {
			componentName := strings.TrimPrefix(ref, "#/components/schemas/")
			resolved, known := schemaComponents[componentName]
			if componentName == ref || !known {
				return nil, fmt.Errorf("path parameter %q references unknown schema %q", name, ref)
			}
			schema = resolved
		}
		projection := pathParameter{Name: name}
		projection.Format, _ = schema["format"].(string)
		projection.Pattern, _ = schema["pattern"].(string)
		if (projection.Format == "") == (projection.Pattern == "") {
			return nil, fmt.Errorf("path parameter %q must declare exactly one of format and pattern", name)
		}
		for key, target := range map[string]*int{"minLength": &projection.MinLength, "maxLength": &projection.MaxLength} {
			if raw, present := schema[key]; present {
				bound, ok := raw.(float64)
				if !ok || bound < 0 || bound != float64(int(bound)) {
					return nil, fmt.Errorf("path parameter %q %s must be a non-negative integer", name, key)
				}
				*target = int(bound)
			}
		}
		projection.NotFoundCode, _ = parameter["x-cartulary-not-found-code"].(string)
		if projection.NotFoundCode == "" {
			return nil, fmt.Errorf("path parameter %q requires x-cartulary-not-found-code", name)
		}
		for status, codes := range errorCodes {
			for _, code := range codes {
				if code == projection.NotFoundCode {
					projection.NotFoundStatus = status
				}
			}
		}
		if projection.NotFoundStatus == 0 {
			return nil, fmt.Errorf("path parameter %q not-found code %q is not declared by the operation", name, projection.NotFoundCode)
		}
		if rawDetails, present := parameter["x-cartulary-not-found-details"]; present {
			details, ok := rawDetails.(map[string]any)
			if !ok || len(details) == 0 {
				return nil, fmt.Errorf("path parameter %q x-cartulary-not-found-details must be a non-empty object", name)
			}
			projection.NotFoundDetails = make(map[string]string, len(details))
			for key, raw := range details {
				value, ok := raw.(string)
				if !ok || value == "" {
					return nil, fmt.Errorf("path parameter %q x-cartulary-not-found-details member %q must be a non-empty string", name, key)
				}
				projection.NotFoundDetails[key] = value
			}
		}
		projected = append(projected, projection)
	}
	return projected, nil
}

// templatePlaceholders returns the template's {name} segments in order.
func templatePlaceholders(pathTemplate string) []string {
	var placeholders []string
	for _, segment := range strings.Split(pathTemplate, "/") {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") && len(segment) > 2 {
			placeholders = append(placeholders, segment[1:len(segment)-1])
		}
	}
	return placeholders
}

func renderPathParameters(parameters []pathParameter) string {
	if len(parameters) == 0 {
		return "nil"
	}
	parts := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		details := "nil"
		if len(parameter.NotFoundDetails) > 0 {
			keys := make([]string, 0, len(parameter.NotFoundDetails))
			for key := range parameter.NotFoundDetails {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			members := make([]string, 0, len(keys))
			for _, key := range keys {
				members = append(members, fmt.Sprintf("%q: %q", key, parameter.NotFoundDetails[key]))
			}
			details = "map[string]string{" + strings.Join(members, ", ") + "}"
		}
		parts = append(parts, fmt.Sprintf(
			"{Name: %q, Format: %q, Pattern: %q, MinLength: %d, MaxLength: %d, NotFoundCode: %q, NotFoundStatus: %d, NotFoundDetails: %s}",
			parameter.Name, parameter.Format, parameter.Pattern, parameter.MinLength, parameter.MaxLength,
			parameter.NotFoundCode, parameter.NotFoundStatus, details,
		))
	}
	return "[]PathParameter{" + strings.Join(parts, ", ") + "}"
}

func renderErrorCodes(declared map[int][]string) string {
	statuses := make([]int, 0, len(declared))
	for status := range declared {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	parts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		parts = append(parts, fmt.Sprintf("%d: %#v", status, declared[status]))
	}
	return "map[int][]string{" + strings.Join(parts, ", ") + "}"
}

func securityDeclaration(operationObject map[string]any) ([][]string, error) {
	rawSecurity, exists := operationObject["security"]
	if !exists {
		return nil, errors.New("operation has no explicit security declaration")
	}
	alternatives, ok := rawSecurity.([]any)
	if !ok {
		return nil, errors.New("security must be an array")
	}
	result := make([][]string, 0, len(alternatives))
	for _, rawAlternative := range alternatives {
		alternative, ok := rawAlternative.(map[string]any)
		if !ok {
			return nil, errors.New("security alternative must be an object")
		}
		names := make([]string, 0, len(alternative))
		for name := range alternative {
			names = append(names, name)
		}
		sort.Strings(names)
		result = append(result, names)
	}
	sort.Slice(result, func(left, right int) bool {
		return strings.Join(result[left], "\x00") < strings.Join(result[right], "\x00")
	})
	return result, nil
}

func render(metadata catalogMetadata, operations []operation) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("// Code generated by tools/openapi-operation-catalog; DO NOT EDIT.\n\n")
	output.WriteString("package openapioperations\n\n")
	fmt.Fprintf(
		&output,
		"const CanonicalSHA256 = %q\nconst DocumentVersion = %q\n\n",
		metadata.SHA256,
		metadata.DocumentVersion,
	)
	output.WriteString("type Operation struct {\n")
	output.WriteString("\tOwnerID string\n\tMethod string\n\tPathTemplate string\n\tPattern string\n")
	output.WriteString("\tOperationID string\n\tAvailability string\n\tStateChanging bool\n")
	output.WriteString("\tSuccessStatuses []int\n\tErrorCodes map[int][]string\n\tSecurity [][]string\n\tPathParameters []PathParameter\n}\n\n")
	output.WriteString("// PathParameter is one path parameter's admission declaration: its format or\n")
	output.WriteString("// pattern, optional length bounds (a zero MaxLength is unbounded), and the\n")
	output.WriteString("// not-found response a malformed value receives.\n")
	output.WriteString("type PathParameter struct {\n\tName string\n\tFormat string\n\tPattern string\n\tMinLength int\n\tMaxLength int\n")
	output.WriteString("\tNotFoundCode string\n\tNotFoundStatus int\n\tNotFoundDetails map[string]string\n}\n\n")
	output.WriteString("var catalog = []Operation{\n")
	for _, operation := range operations {
		fmt.Fprintf(&output, "\t{OwnerID: %q, Method: %q, PathTemplate: %q, Pattern: %q, OperationID: %q, Availability: %q, StateChanging: %t, SuccessStatuses: %#v, ErrorCodes: %s, Security: %#v, PathParameters: %s},\n",
			operation.OwnerID,
			operation.Method,
			operation.PathTemplate,
			operation.Pattern,
			operation.OperationID,
			operation.Availability,
			operation.StateChanging,
			operation.SuccessStatuses,
			renderErrorCodes(operation.ErrorCodes),
			operation.Security,
			renderPathParameters(operation.PathParameters),
		)
	}
	output.WriteString("}\n\n")
	output.WriteString("func All() []Operation {\n")
	output.WriteString("\tresult := make([]Operation, len(catalog))\n\tcopy(result, catalog)\n")
	output.WriteString("\tfor index := range result {\n")
	output.WriteString("\t\tresult[index].SuccessStatuses = append([]int(nil), result[index].SuccessStatuses...)\n")
	output.WriteString("\t\tresult[index].ErrorCodes = make(map[int][]string, len(catalog[index].ErrorCodes))\n")
	output.WriteString("\t\tfor status, codes := range catalog[index].ErrorCodes {\n\t\t\tresult[index].ErrorCodes[status] = append([]string(nil), codes...)\n\t\t}\n")
	output.WriteString("\t\tresult[index].Security = make([][]string, len(result[index].Security))\n")
	output.WriteString("\t\tfor alternative := range result[index].Security {\n")
	output.WriteString("\t\t\tresult[index].Security[alternative] = append([]string(nil), catalog[index].Security[alternative]...)\n")
	output.WriteString("\t\t}\n")
	output.WriteString("\t\tresult[index].PathParameters = append([]PathParameter(nil), catalog[index].PathParameters...)\n")
	output.WriteString("\t\tfor parameter := range result[index].PathParameters {\n")
	output.WriteString("\t\t\tif details := catalog[index].PathParameters[parameter].NotFoundDetails; details != nil {\n")
	output.WriteString("\t\t\t\tresult[index].PathParameters[parameter].NotFoundDetails = make(map[string]string, len(details))\n")
	output.WriteString("\t\t\t\tfor key, value := range details {\n\t\t\t\t\tresult[index].PathParameters[parameter].NotFoundDetails[key] = value\n\t\t\t\t}\n")
	output.WriteString("\t\t\t}\n\t\t}\n\t}\n\treturn result\n}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go: %w", err)
	}
	return formatted, nil
}

func isOperationMethod(method string) bool {
	switch method {
	case "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE":
		return true
	default:
		return false
	}
}

func decodeClosed(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func decodeOpen(path string, target any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, target)
}

func writeAtomically(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()
	if _, err := temporary.Write(content); err != nil {
		return err
	}
	if err := temporary.Chmod(0o644); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return err
	}
	committed = true
	return nil
}

func fail(context string, err error) {
	fmt.Fprintf(os.Stderr, "openapi operation catalog: %s: %v\n", context, err)
	os.Exit(1)
}

func stateChangingDeclaration(operationObject map[string]any) (bool, error) {
	value, ok := operationObject["x-cartulary-state-changing"].(bool)
	if !ok {
		return false, errors.New("operation requires explicit boolean x-cartulary-state-changing")
	}
	return value, nil
}
