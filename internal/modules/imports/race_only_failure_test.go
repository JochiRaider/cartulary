package imports

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// TestImportRaceOnlyFailuresUseDeclaredCodes_Unit verifies the outcomes an
// Import route can reach only when state changes after its own admission and
// lookup. A unit removed before its source read is the route's
// import_unit_not_found, and a region creator who lost visibility or role
// before the transactional recheck receives the session's not-found response
// or authorization_denied. Each code is one the operation declares; any other
// failure stays internal_error (REQ-01-704, REQ-01-706).
func TestImportRaceOnlyFailuresUseDeclaredCodes_Unit(t *testing.T) {
	declared := map[string]map[int][]string{}
	for _, operation := range httpapi.ContractOperationsForOwner("module.imports") {
		declared[operation.OperationID] = operation.ErrorCodes
	}
	requireDeclared := func(t *testing.T, operationID string, apiErr *httpapi.APIError) {
		t.Helper()
		if !slices.Contains(declared[operationID][apiErr.Status], apiErr.Code) {
			t.Fatalf("%s does not declare %d %s; declared=%v", operationID, apiErr.Status, apiErr.Code, declared[operationID])
		}
	}
	injected := errors.New("injected import store failure")

	t.Run("unit removed before its source read", func(t *testing.T) {
		apiErr := unitSourceAPIError(fmt.Errorf("load unit source capability: %w", errNotFound))
		if apiErr.Status != http.StatusNotFound || apiErr.Code != "import_unit_not_found" || len(apiErr.Details) != 0 {
			t.Fatalf("missing unit = %#v, want 404 import_unit_not_found", apiErr)
		}
		for _, operationID := range []string{"inspectImportUnitExtensionSource", "previewImportUnitExtensionMapping"} {
			requireDeclared(t, operationID, apiErr)
		}
		if apiErr := unitSourceAPIError(injected); apiErr.Status != http.StatusInternalServerError || apiErr.Code != "internal_error" {
			t.Fatalf("source read failure = %#v, want internal_error", apiErr)
		}
	})

	t.Run("region creation admission recheck", func(t *testing.T) {
		notVisible := regionAdmissionAPIError(fmt.Errorf("admit region: %w", &admission.Denied{Code: admission.DenialNotVisible}))
		if notVisible == nil || notVisible.Status != http.StatusNotFound || notVisible.Code != "import_session_not_found" || len(notVisible.Details) != 0 {
			t.Fatalf("lost visibility = %#v, want 404 import_session_not_found", notVisible)
		}
		requireDeclared(t, "createImportUnitRegion", notVisible)
		insufficient := regionAdmissionAPIError(&admission.Denied{Code: admission.DenialInsufficientRole})
		if insufficient == nil || insufficient.Status != http.StatusForbidden || insufficient.Code != "authorization_denied" ||
			!reflect.DeepEqual(insufficient.Details, map[string]any{"required_role": "editor|reviewer|admin"}) {
			t.Fatalf("lost role = %#v, want 403 authorization_denied", insufficient)
		}
		requireDeclared(t, "createImportUnitRegion", insufficient)
		// incident_closed and every unclassified failure remain with the
		// shared store mapping, which reports the latter as internal_error.
		for _, err := range []error{&admission.Denied{Code: admission.DenialIncidentClosed}, injected, nil} {
			if apiErr := regionAdmissionAPIError(err); apiErr != nil {
				t.Fatalf("region recheck mapped %v to %#v", err, apiErr)
			}
		}
	})

	// Region creation re-reads a source discovery already parsed in full, so
	// only the current limits can reject it; any other source failure is an
	// invariant break (REQ-01-620e, REQ-01-706).
	t.Run("region creation base-source re-read", func(t *testing.T) {
		rejected := regionSourceAPIError(importSourceRejected("archive_member_count_exceeded", 2, 1))
		if rejected.Status != http.StatusRequestEntityTooLarge || rejected.Code != "import_source_rejected" {
			t.Fatalf("limit failure = %#v, want 413 import_source_rejected", rejected)
		}
		requireDeclared(t, "createImportUnitRegion", rejected)
		for _, apiErr := range []*httpapi.APIError{importSourceUnsupported("encrypted_or_unparseable_workbook"), nil} {
			if got := regionSourceAPIError(apiErr); got.Status != http.StatusInternalServerError || got.Code != "internal_error" {
				t.Fatalf("re-read failure %#v = %#v, want internal_error", apiErr, got)
			}
		}
	})
}
