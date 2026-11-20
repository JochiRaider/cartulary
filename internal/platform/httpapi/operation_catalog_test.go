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
	// Route resolution answers through the registry's handler (Core 01
	// REQ-01-707): an undeclared method, including HEAD on the GET pattern,
	// and an unknown descendant receive registered envelopes.
	handler := withRequestID(registry.Handler(mux), nil)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/fixtures/example", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	for _, method := range []string{http.MethodPost, http.MethodHead} {
		wrongMethod := httptest.NewRecorder()
		handler.ServeHTTP(wrongMethod, httptest.NewRequest(method, "/api/v1/fixtures/example", nil))
		if wrongMethod.Code != http.StatusMethodNotAllowed || wrongMethod.Header().Get("Allow") != http.MethodGet {
			t.Fatalf("%s status = %d Allow=%q, want %d GET", method, wrongMethod.Code, wrongMethod.Header().Get("Allow"), http.StatusMethodNotAllowed)
		}
		if method == http.MethodPost {
			requireMethodNotAllowedEnvelope(t, wrongMethod, http.MethodGet)
		} else if wrongMethod.Body.Len() != 0 {
			t.Fatalf("HEAD rejection carries a body: %q", wrongMethod.Body)
		}
	}
	unknownDescendant := httptest.NewRecorder()
	handler.ServeHTTP(unknownDescendant, httptest.NewRequest(http.MethodGet, "/api/v1/fixtures/example/unknown", nil))
	requireAdmissionError(t, unknownDescendant, http.StatusNotFound, "route_not_found", map[string]any{})
	if err := registry.ValidateActive(); err != nil {
		t.Fatalf("validate active routes: %v", err)
	}
}

func TestRouteRegistryRejectsUnknownOwnerDuplicateAndProfileMismatch(t *testing.T) {
	t.Run("owner family registration and dispatch", testOwnerFamilyRegistration)
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

func TestExplicitGeneratedMutationSemantics(t *testing.T) {
	expected := map[string]bool{
		"previewImportUnitExtensionMapping": false, "inspectImportUnitExtensionSource": false,
		"queryWorkbookView": false, "locateWorkbookViewRecord": false, "validateReportComposition": false,
		"queryEvidenceMetadata": false, "createImportSession": true, "putImportUnitMapping": true,
		"completeEnterpriseOIDC": true, "finishEnterpriseSAML": true, "redeemEvidenceHandle": true,
	}
	for _, operation := range ContractOperations() {
		want, ok := expected[operation.OperationID]
		if !ok {
			continue
		}
		if operation.StateChanging != want {
			t.Fatalf("%s mutation semantics = %v", operation.OperationID, operation.StateChanging)
		}
		if operation.OperationID == "queryEvidenceMetadata" {
			csrf := false
			for _, alternative := range operation.Security {
				for _, scheme := range alternative {
					if scheme == "csrfHeader" {
						csrf = true
					}
				}
			}
			if !csrf {
				t.Fatal("read-only Evidence metadata lost its independent CSRF requirement")
			}
		}
		delete(expected, operation.OperationID)
	}
	if len(expected) != 0 {
		t.Fatalf("missing operations: %v", expected)
	}
}
