package httpapi_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

type baseline struct {
	TargetPath      string `json:"target_path"`
	CanonicalSHA256 string `json:"canonical_sha256"`
	ByteLength      int    `json:"byte_length"`
	Metadata        struct {
		OpenAPI string `json:"openapi"`
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"metadata"`
	VersionPolicy struct {
		BreakingWireChanges              string `json:"breaking_wire_changes"`
		AdditivePublicBehavior           string `json:"additive_public_behavior"`
		NonBehavioralContractCorrections string `json:"non_behavioral_contract_corrections"`
	} `json:"version_policy"`
	Structure struct {
		PathCount                    int `json:"path_count"`
		OperationCount               int `json:"operation_count"`
		ComponentSchemaCount         int `json:"component_schema_count"`
		OperationTagCount            int `json:"operation_tag_count"`
		UniqueInternalRefCount       int `json:"unique_internal_ref_count"`
		ResponseWaiverOperationCount int `json:"response_waiver_operation_count"`
		SecurityWaiverOperationCount int `json:"security_waiver_operation_count"`
	} `json:"structure"`
	SetHashes struct {
		ComponentSchemaNamesSHA256 string `json:"component_schema_names_sha256"`
		InternalRefValuesSHA256    string `json:"internal_ref_values_sha256"`
		OperationInventorySHA256   string `json:"operation_inventory_sha256"`
		OperationTagsSHA256        string `json:"operation_tags_sha256"`
	} `json:"set_hashes"`
	GeneratedArtifacts []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"generated_artifacts"`
}

type compatibilityWaivers struct {
	BaselineSHA256                string            `json:"baseline_sha256"`
	ResponseWaivers               []operationWaiver `json:"response_waivers"`
	SecurityClassificationWaivers []operationWaiver `json:"security_classification_waivers"`
	SecuritySchemeWaivers         []namedWaiver     `json:"security_scheme_waivers"`
	ComponentWaivers              []namedWaiver     `json:"component_waivers"`
	PathParameterWaivers          []json.RawMessage `json:"path_parameter_waivers"`
}

type operationWaiver struct {
	WaiverID         string   `json:"waiver_id"`
	OwnerID          string   `json:"owner_id"`
	CorrectionID     string   `json:"correction_id"`
	Reason           string   `json:"reason"`
	RemovalCondition string   `json:"removal_condition"`
	OperationIDs     []string `json:"operation_ids"`
}

type namedWaiver struct {
	WaiverID         string `json:"waiver_id"`
	OwnerID          string `json:"owner_id"`
	CorrectionID     string `json:"correction_id"`
	Reason           string `json:"reason"`
	RemovalCondition string `json:"removal_condition"`
	Name             string `json:"name"`
}

var operationMethods = map[string]struct{}{
	"delete":  {},
	"get":     {},
	"head":    {},
	"options": {},
	"patch":   {},
	"post":    {},
	"put":     {},
	"trace":   {},
}

func TestCanonicalBaseline(t *testing.T) {
	root := repositoryRoot(t)
	var frozen baseline
	readJSON(t, filepath.Join(root, "contracts/openapi-source/baseline.v1.json"), &frozen)

	canonicalPath := filepath.Join(root, filepath.FromSlash(frozen.TargetPath))
	canonicalBytes, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("read canonical OpenAPI: %v", err)
	}
	requireEqual(t, "canonical byte length", len(canonicalBytes), frozen.ByteLength)
	requireEqual(t, "canonical sha256", hashBytes(canonicalBytes), frozen.CanonicalSHA256)

	var document map[string]any
	if err := json.Unmarshal(canonicalBytes, &document); err != nil {
		t.Fatalf("parse canonical OpenAPI: %v", err)
	}
	requireEqual(t, "OpenAPI dialect", stringValue(document["openapi"]), frozen.Metadata.OpenAPI)
	info := objectValue(t, document["info"], "info")
	requireEqual(t, "OpenAPI title", stringValue(info["title"]), frozen.Metadata.Title)
	requireEqual(t, "OpenAPI document version", stringValue(info["version"]), frozen.Metadata.Version)

	paths := objectValue(t, document["paths"], "paths")
	schemas := objectValue(t, objectValue(t, document["components"], "components")["schemas"], "components.schemas")
	schemes := objectValue(t, objectValue(t, document["components"], "components")["securitySchemes"], "components.securitySchemes")

	operationLines := make([]string, 0, frozen.Structure.OperationCount)
	operationIDs := make(map[string]struct{}, frozen.Structure.OperationCount)
	responseMissing := make([]string, 0, frozen.Structure.ResponseWaiverOperationCount)
	securityMissing := make([]string, 0, frozen.Structure.SecurityWaiverOperationCount)
	usedSecuritySchemes := make(map[string]struct{})
	tags := make(map[string]struct{})
	for path, rawPathItem := range paths {
		pathItem := objectValue(t, rawPathItem, "path item "+path)
		for method, rawOperation := range pathItem {
			if _, ok := operationMethods[method]; !ok {
				continue
			}
			operation := objectValue(t, rawOperation, method+" "+path)
			operationID := stringValue(operation["operationId"])
			if operationID == "" {
				t.Fatalf("%s %s has no operationId", method, path)
			}
			if _, duplicate := operationIDs[operationID]; duplicate {
				t.Fatalf("duplicate operationId %q", operationID)
			}
			operationIDs[operationID] = struct{}{}
			operationLines = append(operationLines, path+"\t"+method+"\t"+operationID)
			if _, ok := operation["responses"]; !ok {
				responseMissing = append(responseMissing, operationID)
			}
			if security, ok := operation["security"]; !ok {
				securityMissing = append(securityMissing, operationID)
			} else {
				for _, rawRequirement := range arrayValue(security) {
					requirement := objectValue(t, rawRequirement, operationID+" security requirement")
					for schemeName := range requirement {
						usedSecuritySchemes[schemeName] = struct{}{}
					}
				}
			}
			for _, rawTag := range arrayValue(operation["tags"]) {
				tags[stringValue(rawTag)] = struct{}{}
			}
		}
	}

	componentNames := sortedMapKeys(schemas)
	refSet := make(map[string]struct{})
	collectRefs(document, refSet)
	refs := sortedMapKeys(refSet)
	tagNames := sortedMapKeys(tags)
	sort.Strings(operationLines)
	sort.Strings(responseMissing)
	sort.Strings(securityMissing)

	requireEqual(t, "path count", len(paths), frozen.Structure.PathCount)
	requireEqual(t, "operation count", len(operationLines), frozen.Structure.OperationCount)
	requireEqual(t, "component schema count", len(componentNames), frozen.Structure.ComponentSchemaCount)
	requireEqual(t, "operation tag count", len(tagNames), frozen.Structure.OperationTagCount)
	requireEqual(t, "unique internal reference count", len(refs), frozen.Structure.UniqueInternalRefCount)
	requireEqual(t, "response waiver operation count", len(responseMissing), frozen.Structure.ResponseWaiverOperationCount)
	requireEqual(t, "security waiver operation count", len(securityMissing), frozen.Structure.SecurityWaiverOperationCount)
	requireEqual(t, "operation inventory hash", hashLines(operationLines), frozen.SetHashes.OperationInventorySHA256)
	requireEqual(t, "component schema-name hash", hashLines(componentNames), frozen.SetHashes.ComponentSchemaNamesSHA256)
	requireEqual(t, "reference-value hash", hashLines(refs), frozen.SetHashes.InternalRefValuesSHA256)
	requireEqual(t, "operation-tag hash", hashLines(tagNames), frozen.SetHashes.OperationTagsSHA256)

	for _, ref := range refs {
		if !strings.HasPrefix(ref, "#/") {
			t.Fatalf("external OpenAPI reference is not allowed: %q", ref)
		}
	}

	var waivers compatibilityWaivers
	readJSON(t, filepath.Join(root, "contracts/openapi-source/compatibility-waivers.json"), &waivers)
	requireEqual(t, "waiver baseline hash", waivers.BaselineSHA256, frozen.CanonicalSHA256)
	requireStringSetsEqual(t, "response waivers", flattenOperationWaivers(t, waivers.ResponseWaivers), responseMissing)
	requireStringSetsEqual(t, "security waivers", flattenOperationWaivers(t, waivers.SecurityClassificationWaivers), securityMissing)
	unusedSecuritySchemes := make([]string, 0)
	for _, schemeName := range sortedMapKeys(schemes) {
		if _, used := usedSecuritySchemes[schemeName]; !used {
			unusedSecuritySchemes = append(unusedSecuritySchemes, schemeName)
		}
	}
	requireStringSetsEqual(t, "security-scheme waivers", flattenNamedWaivers(t, waivers.SecuritySchemeWaivers), unusedSecuritySchemes)

	referencedSchemas := make(map[string]struct{})
	for _, ref := range refs {
		const prefix = "#/components/schemas/"
		if strings.HasPrefix(ref, prefix) {
			referencedSchemas[strings.TrimPrefix(ref, prefix)] = struct{}{}
		}
	}
	unreferencedSchemas := make([]string, 0)
	for _, name := range componentNames {
		if _, ok := referencedSchemas[name]; !ok {
			unreferencedSchemas = append(unreferencedSchemas, name)
		}
	}
	requireStringSetsEqual(t, "component waivers", flattenNamedWaivers(t, waivers.ComponentWaivers), unreferencedSchemas)

	for _, artifact := range frozen.GeneratedArtifacts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.Path)))
		if err != nil {
			t.Fatalf("read generated artifact %s: %v", artifact.Path, err)
		}
		requireEqual(t, artifact.Path+" sha256", hashBytes(content), artifact.SHA256)
	}
	canonicalGeneratedJSON, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("canonicalize OpenAPI for generated snapshot comparison: %v", err)
	}
	var generatedOpenAPI *gencontracts.Artifact
	for index := range gencontracts.OpenAPIArtifacts {
		if gencontracts.OpenAPIArtifacts[index].Path == frozen.TargetPath {
			generatedOpenAPI = &gencontracts.OpenAPIArtifacts[index]
			break
		}
	}
	if generatedOpenAPI == nil {
		t.Fatalf("generated Go OpenAPI snapshot %s is missing", frozen.TargetPath)
	}
	requireEqual(t, "generated Go OpenAPI content", generatedOpenAPI.JSON, string(canonicalGeneratedJSON))
	requireEqual(t, "generated Go OpenAPI sha256", generatedOpenAPI.SHA256, hashBytes(canonicalGeneratedJSON))
}

func TestStableOpenAPIMetadataAndCompatibilityClosure(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	requireEqual(t, "stable OpenAPI dialect", stringValue(document["openapi"]), "3.1.0")
	info := objectValue(t, document["info"], "info")
	requireEqual(t, "stable OpenAPI title", stringValue(info["title"]), "Cartulary HTTP API")
	requireEqual(t, "stable OpenAPI document version", stringValue(info["version"]), "1.0.0")

	root := repositoryRoot(t)
	var frozen baseline
	readJSON(t, filepath.Join(root, "contracts/openapi-source/baseline.v1.json"), &frozen)
	requireEqual(t, "breaking wire change version increment", frozen.VersionPolicy.BreakingWireChanges, "major")
	requireEqual(t, "additive public behavior version increment", frozen.VersionPolicy.AdditivePublicBehavior, "minor")
	requireEqual(t, "non-behavioral contract correction version increment", frozen.VersionPolicy.NonBehavioralContractCorrections, "patch")

	var waivers compatibilityWaivers
	readJSON(t, filepath.Join(root, "contracts/openapi-source/compatibility-waivers.json"), &waivers)
	requireEqual(t, "response waiver count", len(waivers.ResponseWaivers), 0)
	requireEqual(t, "security-classification waiver count", len(waivers.SecurityClassificationWaivers), 0)
	requireEqual(t, "security-scheme waiver count", len(waivers.SecuritySchemeWaivers), 0)
	requireEqual(t, "component waiver count", len(waivers.ComponentWaivers), 0)
	requireEqual(t, "path-parameter waiver count", len(waivers.PathParameterWaivers), 0)
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../.."))
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(content, target); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func objectValue(t *testing.T, raw any, label string) map[string]any {
	t.Helper()
	object, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("%s must be an object, got %T", label, raw)
	}
	return object
}

func arrayValue(raw any) []any {
	array, _ := raw.([]any)
	return array
}

func stringArray(raw any) []string {
	values := arrayValue(raw)
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func stringValue(raw any) string {
	value, _ := raw.(string)
	return value
}

func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func collectRefs(value any, refs map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		if ref, ok := typed["$ref"].(string); ok {
			refs[ref] = struct{}{}
		}
		for _, child := range typed {
			collectRefs(child, refs)
		}
	case []any:
		for _, child := range typed {
			collectRefs(child, refs)
		}
	}
}

func flattenOperationWaivers(t *testing.T, waivers []operationWaiver) []string {
	t.Helper()
	seenWaivers := make(map[string]struct{}, len(waivers))
	operationIDs := make([]string, 0)
	seenOperations := make(map[string]struct{})
	for _, waiver := range waivers {
		validateWaiverFields(t, waiver.WaiverID, waiver.OwnerID, waiver.CorrectionID, waiver.Reason, waiver.RemovalCondition)
		if _, duplicate := seenWaivers[waiver.WaiverID]; duplicate {
			t.Fatalf("duplicate waiver_id %q", waiver.WaiverID)
		}
		seenWaivers[waiver.WaiverID] = struct{}{}
		for _, operationID := range waiver.OperationIDs {
			if _, duplicate := seenOperations[operationID]; duplicate {
				t.Fatalf("operation %q has multiple waivers in the same class", operationID)
			}
			seenOperations[operationID] = struct{}{}
			operationIDs = append(operationIDs, operationID)
		}
	}
	sort.Strings(operationIDs)
	return operationIDs
}

func flattenNamedWaivers(t *testing.T, waivers []namedWaiver) []string {
	t.Helper()
	seenWaivers := make(map[string]struct{}, len(waivers))
	names := make([]string, 0, len(waivers))
	seenNames := make(map[string]struct{}, len(waivers))
	for _, waiver := range waivers {
		validateWaiverFields(t, waiver.WaiverID, waiver.OwnerID, waiver.CorrectionID, waiver.Reason, waiver.RemovalCondition)
		if _, duplicate := seenWaivers[waiver.WaiverID]; duplicate {
			t.Fatalf("duplicate waiver_id %q", waiver.WaiverID)
		}
		seenWaivers[waiver.WaiverID] = struct{}{}
		if _, duplicate := seenNames[waiver.Name]; duplicate {
			t.Fatalf("name %q has multiple waivers in the same class", waiver.Name)
		}
		seenNames[waiver.Name] = struct{}{}
		names = append(names, waiver.Name)
	}
	sort.Strings(names)
	return names
}

func validateWaiverFields(t *testing.T, waiverID, ownerID, correctionID, reason, removalCondition string) {
	t.Helper()
	if waiverID == "" || ownerID == "" || correctionID == "" || reason == "" || removalCondition == "" {
		t.Fatalf("waiver %q has an empty owner, correction, reason, or removal condition", waiverID)
	}
}

func requireStringSetsEqual(t *testing.T, label string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	requireEqual(t, label, strings.Join(got, "\n"), strings.Join(want, "\n"))
}

func hashLines(lines []string) string {
	return hashBytes([]byte(strings.Join(lines, "\n") + "\n"))
}

func hashBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func requireEqual[T comparable](t *testing.T, label string, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("%s mismatch: got %v want %v", label, got, want)
	}
}
