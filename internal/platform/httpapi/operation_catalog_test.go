package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratedContractOperationCatalogIsClosed(t *testing.T) {
	operations := ContractOperations()
	if len(operations) == 0 {
		t.Fatal("generated operation catalog is empty")
	}
	operationIDs := make(map[string]struct{}, len(operations))
	patterns := make(map[string]struct{}, len(operations))
	enterpriseOperations := 0
	for _, operation := range operations {
		if err := validateContractOperation(operation); err != nil {
			t.Fatalf("validate %s: %v", operation.OperationID, err)
		}
		if _, duplicate := operationIDs[operation.OperationID]; duplicate {
			t.Fatalf("duplicate operation ID %q", operation.OperationID)
		}
		if _, duplicate := patterns[operation.Pattern]; duplicate {
			t.Fatalf("duplicate operation pattern %q", operation.Pattern)
		}
		operationIDs[operation.OperationID] = struct{}{}
		patterns[operation.Pattern] = struct{}{}
		if operation.Availability == "enterprise_authentication" {
			enterpriseOperations++
			if operation.OwnerID != "module.auth" {
				t.Fatalf("enterprise operation %s has owner %s", operation.OperationID, operation.OwnerID)
			}
		}
	}
	if enterpriseOperations == 0 {
		t.Fatal("generated catalog has no Enterprise Authentication availability metadata")
	}
}

func TestRouteRegistryBindsGeneratedPatternAndMetadata(t *testing.T) {
	operations := []ContractOperation{{
		OwnerID:         "module.fixture",
		Method:          http.MethodGet,
		PathTemplate:    "/api/v1/fixtures/{fixture_id}",
		Pattern:         "GET /api/v1/fixtures/{fixture_id}",
		OperationID:     "getFixture",
		Availability:    BaseOperationAvailability,
		SuccessStatuses: []int{http.StatusOK},
		Security:        [][]string{{"sessionCookie"}},
	}}
	registry, err := newRouteRegistry(operations, nil)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	mux := http.NewServeMux()
	if err := registry.Bind(mux, "module.fixture", "getFixture", func(writer http.ResponseWriter, request *http.Request) {
		metadata, ok := OperationMetadataFromContext(request.Context())
		if !ok || metadata.OwnerID != "module.fixture" || metadata.OperationID != "getFixture" {
			t.Fatalf("operation metadata = %#v, %t", metadata, ok)
		}
		writer.WriteHeader(http.StatusOK)
	}); err != nil {
		t.Fatalf("bind route: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/fixtures/example", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	wrongMethod := httptest.NewRecorder()
	mux.ServeHTTP(
		wrongMethod,
		httptest.NewRequest(http.MethodPost, "/api/v1/fixtures/example", nil),
	)
	if wrongMethod.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong-method status = %d, want %d", wrongMethod.Code, http.StatusMethodNotAllowed)
	}
	unknownDescendant := httptest.NewRecorder()
	mux.ServeHTTP(
		unknownDescendant,
		httptest.NewRequest(http.MethodGet, "/api/v1/fixtures/example/unknown", nil),
	)
	if unknownDescendant.Code != http.StatusNotFound {
		t.Fatalf("unknown-descendant status = %d, want %d", unknownDescendant.Code, http.StatusNotFound)
	}
	if err := registry.ValidateActive(); err != nil {
		t.Fatalf("validate active routes: %v", err)
	}
}

func TestRouteRegistryRejectsUnknownOwnerDuplicateAndProfileMismatch(t *testing.T) {
	operations := []ContractOperation{
		{
			OwnerID:         "module.fixture",
			Method:          http.MethodGet,
			PathTemplate:    "/api/v1/fixtures",
			Pattern:         "GET /api/v1/fixtures",
			OperationID:     "listFixtures",
			Availability:    BaseOperationAvailability,
			SuccessStatuses: []int{http.StatusOK},
		},
		{
			OwnerID:         "module.fixture",
			Method:          http.MethodPost,
			PathTemplate:    "/api/v1/fixtures/profile",
			Pattern:         "POST /api/v1/fixtures/profile",
			OperationID:     "profileFixture",
			Availability:    "fixture_profile",
			StateChanging:   true,
			SuccessStatuses: []int{http.StatusOK},
		},
	}
	registry, err := newRouteRegistry(operations, nil)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	mux := http.NewServeMux()
	handler := func(http.ResponseWriter, *http.Request) {}
	if err := registry.Bind(mux, "wrong.owner", "listFixtures", handler); err == nil || !strings.Contains(err.Error(), "owned by") {
		t.Fatalf("wrong-owner error = %v", err)
	}
	if err := registry.Bind(mux, "module.fixture", "missing", handler); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown-operation error = %v", err)
	}
	if err := registry.Bind(mux, "module.fixture", "listFixtures", handler); err != nil {
		t.Fatalf("bind base operation: %v", err)
	}
	if err := registry.Bind(mux, "module.fixture", "listFixtures", handler); err == nil || !strings.Contains(err.Error(), "already bound") {
		t.Fatalf("duplicate error = %v", err)
	}
	if err := registry.Bind(mux, "module.fixture", "profileFixture", handler); err != nil {
		t.Fatalf("bind profile operation: %v", err)
	}
	if err := registry.ValidateActive(); err == nil || !strings.Contains(err.Error(), "unexpected") {
		t.Fatalf("profile parity error = %v", err)
	}
}

func TestRouteRegistryDiagnosticsAreDeterministicAndPayloadFree(t *testing.T) {
	operations := []ContractOperation{
		{
			OwnerID:         "module.fixture",
			Method:          http.MethodGet,
			PathTemplate:    "/api/v1/fixtures",
			Pattern:         "GET /api/v1/fixtures",
			OperationID:     "listFixtures",
			Availability:    BaseOperationAvailability,
			SuccessStatuses: []int{http.StatusOK},
		},
		{
			OwnerID:         "module.fixture",
			Method:          http.MethodPost,
			PathTemplate:    "/api/v1/fixtures/profile",
			Pattern:         "POST /api/v1/fixtures/profile",
			OperationID:     "profileFixture",
			Availability:    "z_profile",
			StateChanging:   true,
			SuccessStatuses: []int{http.StatusCreated},
		},
	}
	registry, err := newRouteRegistry(operations, []ExtensionClaim{
		{ProfileID: "z_profile", Claimed: true},
		{ProfileID: "a_profile", Claimed: true},
	})
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}
	mux := http.NewServeMux()
	if err := registry.Bind(mux, "module.fixture", "listFixtures", func(http.ResponseWriter, *http.Request) {}); err != nil {
		t.Fatalf("bind base operation: %v", err)
	}

	diagnostics := registry.Diagnostics()
	if diagnostics.CanonicalSHA256 == "" || diagnostics.DocumentVersion == "" {
		t.Fatalf("missing canonical identity: %#v", diagnostics)
	}
	if diagnostics.SupportedOperationCount != 2 || diagnostics.ActiveOperationCount != 1 {
		t.Fatalf("operation counts = %#v", diagnostics)
	}
	if got, want := strings.Join(diagnostics.ClaimedProfiles, ","), "a_profile,z_profile"; got != want {
		t.Fatalf("claimed profiles = %q, want %q", got, want)
	}
}
