package contracttest

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	contractaudit "github.com/JochiRaider/cartulary/internal/gen/contractaudit"
	contracterrors "github.com/JochiRaider/cartulary/internal/gen/contracterrors"
	contractextensions "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	contractnetworkflow "github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
	contractopenapi "github.com/JochiRaider/cartulary/internal/gen/contractopenapi"
	contractviewschemas "github.com/JochiRaider/cartulary/internal/gen/contractviewschemas"
	contractws "github.com/JochiRaider/cartulary/internal/gen/contractws"
)

const (
	OpenAPIContractPath   = "contracts/openapi/cartulary.openapi.yaml"
	ErrorRegistryPath     = "contracts/errors/index.json"
	ExtensionRegistryPath = "contracts/extensions/generated/profile-registry.json"
	WSIndexPath           = "contracts/ws/index.schema.json"
	ViewSchemaPrefix      = "contracts/view-schemas/"
)

type ErrorContract struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"http_status"`
	Summary    string `json:"summary"`
}

type ExtensionProfileContract struct {
	ProfileID     string   `json:"profile_id"`
	Claimable     bool     `json:"claimable"`
	ContractMajor *int     `json:"contract_major"`
	RouteFamilies []string `json:"route_families"`
	WorkspaceKeys []string `json:"workspace_keys"`
	CapabilityIDs []string `json:"capability_ids"`
}

type errorRegistryDocument struct {
	Errors []ErrorContract `json:"errors"`
}

type extensionRegistryDocument struct {
	Profiles []ExtensionProfileContract `json:"profiles"`
}

func ErrorContractByCode(t testing.TB, code string) ErrorContract {
	t.Helper()

	document := loadErrorRegistry(t)
	for _, contract := range document.Errors {
		if contract.Code == code {
			return contract
		}
	}
	t.Fatalf("missing error contract for code %q", code)
	return ErrorContract{}
}

func RequireErrorContract(t testing.TB, code string, wantStatus int) {
	t.Helper()

	contract := ErrorContractByCode(t, code)
	if contract.HTTPStatus != wantStatus {
		t.Fatalf("unexpected error contract status for %q: got %d want %d", code, contract.HTTPStatus, wantStatus)
	}
}

func CurrentProfileExtensions(t testing.TB) []ExtensionProfileContract {
	t.Helper()

	document := loadExtensionRegistry(t)
	profiles := append([]ExtensionProfileContract(nil), document.Profiles...)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].ProfileID < profiles[j].ProfileID
	})
	return profiles
}

func ContractArtifactJSON(t testing.TB, path string) string {
	t.Helper()

	jsonText, ok := contractArtifactJSON(path)
	if !ok {
		t.Fatalf("missing generated contract artifact: %s", path)
	}
	return jsonText
}

func ContractArtifactPaths(t testing.TB, prefix string) []string {
	t.Helper()

	paths := make([]string, 0)
	switch {
	case strings.HasPrefix(prefix, "contracts/openapi"):
		paths = matchingArtifactPaths(contractopenapi.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/ws"):
		paths = matchingArtifactPaths(contractws.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/view-schemas"):
		paths = matchingArtifactPaths(contractviewschemas.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/errors"):
		paths = matchingArtifactPaths(contracterrors.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/extensions"):
		paths = matchingArtifactPaths(contractextensions.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/network-flow"):
		paths = matchingArtifactPaths(contractnetworkflow.Index, prefix)
	case strings.HasPrefix(prefix, "contracts/audit"):
		paths = matchingArtifactPaths(contractaudit.Index, prefix)
	}
	if len(paths) == 0 {
		t.Fatalf("missing generated contract artifacts with prefix %q", prefix)
	}
	sort.Strings(paths)
	return paths
}

type artifactWithJSON interface {
	contractopenapi.Artifact | contractws.Artifact | contractviewschemas.Artifact |
		contracterrors.Artifact | contractextensions.Artifact | contractnetworkflow.Artifact |
		contractaudit.Artifact
}

func matchingArtifactPaths[T artifactWithJSON](index map[string]T, prefix string) []string {
	paths := make([]string, 0)
	for path := range index {
		if strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	return paths
}

func contractArtifactJSON(path string) (string, bool) {
	switch {
	case strings.HasPrefix(path, "contracts/openapi/"):
		artifact, ok := contractopenapi.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/ws/"):
		artifact, ok := contractws.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/view-schemas/"):
		artifact, ok := contractviewschemas.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/errors/"):
		artifact, ok := contracterrors.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/extensions/"):
		artifact, ok := contractextensions.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/network-flow/"):
		artifact, ok := contractnetworkflow.Index[path]
		return artifact.JSON, ok
	case strings.HasPrefix(path, "contracts/audit/"):
		artifact, ok := contractaudit.Index[path]
		return artifact.JSON, ok
	default:
		return "", false
	}
}

func DecodeContractArtifact(t testing.TB, path string, target any) {
	t.Helper()

	if err := json.Unmarshal([]byte(ContractArtifactJSON(t, path)), target); err != nil {
		t.Fatalf("decode generated contract artifact %s: %v", path, err)
	}
}

func ContractDocument(t testing.TB, path string) map[string]any {
	t.Helper()

	var document map[string]any
	DecodeContractArtifact(t, path, &document)
	return document
}

func OpenAPIArtifactJSON(t testing.TB) string {
	t.Helper()

	return ContractArtifactJSON(t, OpenAPIContractPath)
}

func OpenAPIDocument(t testing.TB) map[string]any {
	t.Helper()

	return ContractDocument(t, OpenAPIContractPath)
}

func DecodeOpenAPI(t testing.TB, target any) {
	t.Helper()

	DecodeContractArtifact(t, OpenAPIContractPath, target)
}

func ErrorRegistryArtifactJSON(t testing.TB) string {
	t.Helper()

	return ContractArtifactJSON(t, ErrorRegistryPath)
}

func ErrorRegistryDocument(t testing.TB) map[string]any {
	t.Helper()

	return ContractDocument(t, ErrorRegistryPath)
}

func DecodeErrorRegistry(t testing.TB, target any) {
	t.Helper()

	DecodeContractArtifact(t, ErrorRegistryPath, target)
}

func ExtensionRegistryArtifactJSON(t testing.TB) string {
	t.Helper()

	return ContractArtifactJSON(t, ExtensionRegistryPath)
}

func ExtensionRegistryDocument(t testing.TB) map[string]any {
	t.Helper()

	return ContractDocument(t, ExtensionRegistryPath)
}

func DecodeExtensionRegistry(t testing.TB, target any) {
	t.Helper()

	DecodeContractArtifact(t, ExtensionRegistryPath, target)
}

func WSIndexArtifactJSON(t testing.TB) string {
	t.Helper()

	return ContractArtifactJSON(t, WSIndexPath)
}

func WSIndexDocument(t testing.TB) map[string]any {
	t.Helper()

	return ContractDocument(t, WSIndexPath)
}

func DecodeWSIndex(t testing.TB, target any) {
	t.Helper()

	DecodeContractArtifact(t, WSIndexPath, target)
}

func ViewSchemaArtifactPaths(t testing.TB) []string {
	t.Helper()

	return ContractArtifactPaths(t, ViewSchemaPrefix)
}

func loadErrorRegistry(t testing.TB) errorRegistryDocument {
	t.Helper()

	var document errorRegistryDocument
	DecodeErrorRegistry(t, &document)
	return document
}

func loadExtensionRegistry(t testing.TB) extensionRegistryDocument {
	t.Helper()

	var document extensionRegistryDocument
	DecodeExtensionRegistry(t, &document)
	return document
}
