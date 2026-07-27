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

type networkFlowRouteContract struct {
	SchemaID string `json:"schema_id"`
	Routes   []struct {
		RouteID             string `json:"route_id"`
		Method              string `json:"method"`
		Path                string `json:"path"`
		SuccessHTTPStatuses []int  `json:"success_http_statuses"`
	} `json:"routes"`
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
	Security        [][]string
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
	for _, unit := range sourceManifest.Units {
		var document map[string]any
		if err := decodeClosed(filepath.Join(root, filepath.FromSlash(unit.Path)), &document); err != nil {
			return nil, fmt.Errorf("owner unit %s: %w", unit.OwnerID, err)
		}
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
				security, err := securityDeclaration(operationObject)
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
					StateChanging:   method != "GET" && method != "HEAD" && method != "OPTIONS",
					SuccessStatuses: successStatuses,
					Security:        security,
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

func loadNetworkFlowRoutes(root string) ([]operation, error) {
	var contract networkFlowRouteContract
	path := filepath.Join(root, "contracts/network-flow/routes.v1.json")
	if err := decodeOpen(path, &contract); err != nil {
		return nil, err
	}
	if contract.SchemaID != "cartulary.network_flow_route_contracts.v1" {
		return nil, fmt.Errorf("unexpected schema_id %q", contract.SchemaID)
	}
	seenIDs := make(map[string]struct{}, len(contract.Routes))
	seenPatterns := make(map[string]struct{}, len(contract.Routes))
	routes := make([]operation, 0, len(contract.Routes))
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
		routes = append(routes, operation{
			Method:          route.Method,
			PathTemplate:    route.Path,
			Pattern:         pattern,
			OperationID:     route.RouteID,
			SuccessStatuses: append([]int(nil), route.SuccessHTTPStatuses...),
		})
	}
	sort.Slice(routes, func(left, right int) bool {
		return routes[left].OperationID < routes[right].OperationID
	})
	return routes, nil
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
	output.WriteString("\tSuccessStatuses []int\n\tSecurity [][]string\n}\n\n")
	output.WriteString("var catalog = []Operation{\n")
	for _, operation := range operations {
		fmt.Fprintf(&output, "\t{OwnerID: %q, Method: %q, PathTemplate: %q, Pattern: %q, OperationID: %q, Availability: %q, StateChanging: %t, SuccessStatuses: %#v, Security: %#v},\n",
			operation.OwnerID,
			operation.Method,
			operation.PathTemplate,
			operation.Pattern,
			operation.OperationID,
			operation.Availability,
			operation.StateChanging,
			operation.SuccessStatuses,
			operation.Security,
		)
	}
	output.WriteString("}\n\n")
	output.WriteString("func All() []Operation {\n")
	output.WriteString("\tresult := make([]Operation, len(catalog))\n\tcopy(result, catalog)\n")
	output.WriteString("\tfor index := range result {\n")
	output.WriteString("\t\tresult[index].SuccessStatuses = append([]int(nil), result[index].SuccessStatuses...)\n")
	output.WriteString("\t\tresult[index].Security = make([][]string, len(result[index].Security))\n")
	output.WriteString("\t\tfor alternative := range result[index].Security {\n")
	output.WriteString("\t\t\tresult[index].Security[alternative] = append([]string(nil), catalog[index].Security[alternative]...)\n")
	output.WriteString("\t\t}\n\t}\n\treturn result\n}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go: %w", err)
	}
	return formatted, nil
}

func renderNetworkFlow(routes []operation) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("// Code generated by tools/openapi-operation-catalog; DO NOT EDIT.\n\n")
	output.WriteString("package networkflowroutes\n\n")
	output.WriteString("type Route struct {\n\tRouteID string\n\tMethod string\n\tPath string\n\tPattern string\n\tSuccessStatuses []int\n}\n\n")
	output.WriteString("var catalog = []Route{\n")
	for _, route := range routes {
		fmt.Fprintf(
			&output,
			"\t{RouteID: %q, Method: %q, Path: %q, Pattern: %q, SuccessStatuses: %#v},\n",
			route.OperationID,
			route.Method,
			route.PathTemplate,
			route.Pattern,
			route.SuccessStatuses,
		)
	}
	output.WriteString("}\n\n")
	output.WriteString("func All() []Route {\n\tresult := make([]Route, len(catalog))\n\tcopy(result, catalog)\n")
	output.WriteString("\tfor index := range result {\n\t\tresult[index].SuccessStatuses = append([]int(nil), result[index].SuccessStatuses...)\n\t}\n")
	output.WriteString("\treturn result\n}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Network Flow Go: %w", err)
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
