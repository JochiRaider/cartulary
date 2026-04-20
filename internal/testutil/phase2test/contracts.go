package phase2test

import (
	"encoding/json"
	"sort"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

const (
	errorRegistryPath     = "contracts/errors/index.json"
	extensionRegistryPath = "contracts/extensions/index.json"
)

type ErrorContract struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"http_status"`
	Summary    string `json:"summary"`
}

type ExtensionProfileContract struct {
	ProfileID     string   `json:"profile_id"`
	RouteFamilies []string `json:"route_families"`
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

func loadErrorRegistry(t testing.TB) errorRegistryDocument {
	t.Helper()

	artifact, ok := gencontracts.ContractArtifactIndex[errorRegistryPath]
	if !ok {
		t.Fatalf("missing generated error registry artifact: %s", errorRegistryPath)
	}

	var document errorRegistryDocument
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode error registry artifact: %v", err)
	}
	return document
}

func loadExtensionRegistry(t testing.TB) extensionRegistryDocument {
	t.Helper()

	artifact, ok := gencontracts.ContractArtifactIndex[extensionRegistryPath]
	if !ok {
		t.Fatalf("missing generated extension registry artifact: %s", extensionRegistryPath)
	}

	var document extensionRegistryDocument
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode extension registry artifact: %v", err)
	}
	return document
}
