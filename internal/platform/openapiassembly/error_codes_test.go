package openapiassembly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAssemblyEnforcesDeclaredErrorCodes covers Core 01 REQ-01-702: an error
// response, direct or through a component, must declare a non-empty, ascending,
// duplicate-free list of codes the registry assigns that status, and never a
// route-resolution code, which belongs to no operation (REQ-01-707). The fixture
// operation names every security scheme, so its 409 also carries the implied
// credential_bootstrap_rejected.
func TestAssemblyEnforcesDeclaredErrorCodes(t *testing.T) {
	const registry = `{"errors":[{"code":"alpha_conflict","http_status":409},{"code":"beta_conflict","http_status":409},{"code":"credential_bootstrap_rejected","http_status":409},` +
		`{"code":"csrf_verification_failed","http_status":403},{"code":"session_required","http_status":401},{"code":"thing_not_found","http_status":404},` +
		fixtureImpliedRegistryEntries + `]}`
	// The route-resolution rejections are registered but belong to no
	// operation (REQ-01-707), so no operation may declare them.
	preResolutionRegistry := strings.Replace(registry, `{"errors":[`, `{"errors":[{"code":"method_not_allowed","http_status":405},{"code":"route_not_found","http_status":404},`, 1)
	document := func(conflict string) string {
		return `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{"/a":{"post":{"operationId":"a","responses":{"200":{"description":"ok"},"409":` + conflict + `}}}},` +
			`"components":{"responses":{"Shared":{"description":"shared","x-cartulary-error-codes":["thing_not_found"]},"Valid":{"description":"valid","x-cartulary-error-codes":["alpha_conflict","beta_conflict","credential_bootstrap_rejected"]}}}}`
	}
	for _, tc := range []struct {
		name, conflict, registry, wantError string
	}{
		{"direct declaration", `{"description":"c","x-cartulary-error-codes":["alpha_conflict","beta_conflict","credential_bootstrap_rejected"]}`, registry, ""},
		{"component declaration", `{"$ref":"#/components/responses/Valid"}`, registry, ""},
		{"missing", `{"description":"c"}`, registry, "requires x-cartulary-error-codes"},
		{"empty", `{"description":"c","x-cartulary-error-codes":[]}`, registry, "must be a non-empty array"},
		{"unsorted", `{"description":"c","x-cartulary-error-codes":["beta_conflict","alpha_conflict"]}`, registry, "ascending without duplicates"},
		{"duplicated", `{"description":"c","x-cartulary-error-codes":["alpha_conflict","alpha_conflict"]}`, registry, "ascending without duplicates"},
		{"unregistered", `{"description":"c","x-cartulary-error-codes":["gamma_conflict"]}`, registry, `unregistered code "gamma_conflict"`},
		{"status mismatch", `{"description":"c","x-cartulary-error-codes":["thing_not_found"]}`, registry, "assigns status 404"},
		{"component status mismatch", `{"$ref":"#/components/responses/Shared"}`, registry, "assigns status 404"},
		{"registry absent", `{"description":"c","x-cartulary-error-codes":["alpha_conflict"]}`, "", "read error registry"},
		{"route_not_found declared", `{"description":"c","x-cartulary-error-codes":["route_not_found"]}`, preResolutionRegistry, `names "route_not_found", which belongs to no operation`},
		{"method_not_allowed declared", `{"description":"c","x-cartulary-error-codes":["alpha_conflict","method_not_allowed"]}`, preResolutionRegistry, `names "method_not_allowed", which belongs to no operation`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): document(tc.conflict)})
			if tc.registry != "" {
				writeFixtureFile(t, fixture.root, errorRegistryPath, tc.registry)
			} else if err := os.Remove(filepath.Join(fixture.root, filepath.FromSlash(errorRegistryPath))); err != nil {
				t.Fatal(err)
			}
			_, _, err := Assemble(fixture.manifestPath)
			switch {
			case tc.wantError == "" && err != nil:
				t.Fatalf("assemble: %v", err)
			case tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)):
				t.Fatalf("assemble error = %v, want substring %q", err, tc.wantError)
			}
		})
	}
	t.Run("the canonical registry registers the route-resolution codes", func(t *testing.T) {
		repository, err := os.OpenRoot(filepath.Join("..", "..", ".."))
		if err != nil {
			t.Fatal(err)
		}
		defer repository.Close()
		canonical, err := loadErrorRegistry(repository)
		if err != nil {
			t.Fatal(err)
		}
		for code, status := range map[string]int{"route_not_found": 404, "method_not_allowed": 405} {
			if canonical[code] != status {
				t.Fatalf("registry assigns %s status %d, want %d", code, canonical[code], status)
			}
		}
	})
}

// TestAssemblyEnforcesSecurityImpliedErrorCodes covers the security-implied
// declarations of Core 01 REQ-01-702 (AC-608) in both directions: the 401
// declares session_required exactly when the operation declares a security
// requirement, the 403 declares csrf_verification_failed exactly when an
// alternative includes csrfHeader, and the 409 declares
// credential_bootstrap_rejected exactly when an alternative includes
// bearerSession or credentialBootstrapBearer, directly or through a component.
func TestAssemblyEnforcesSecurityImpliedErrorCodes(t *testing.T) {
	const registry = `{"errors":[{"code":"alpha_conflict","http_status":409},{"code":"authorization_denied","http_status":403},` +
		`{"code":"credential_bootstrap_rejected","http_status":409},{"code":"csrf_verification_failed","http_status":403},` +
		`{"code":"invalid_credentials","http_status":401},{"code":"session_required","http_status":401},` + fixtureImpliedRegistryEntries + `]}`
	const (
		session   = `"401":{"description":"s","x-cartulary-error-codes":["session_required"]}`
		csrf      = `"403":{"description":"c","x-cartulary-error-codes":["csrf_verification_failed"]}`
		bootstrap = `"409":{"description":"b","x-cartulary-error-codes":["credential_bootstrap_rejected"]}`

		cookieCSRF = `{"csrfCookie":[],"csrfHeader":[],"sessionCookie":[]}`
		bearer     = `{"bearerSession":[]}`
	)
	// The anchor operation uses every scheme, as the assembled document must,
	// and is itself correctly declared; the case varies operation /a only.
	document := func(security string, responses ...string) string {
		return `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{` +
			`"/anchor":{"post":{"operationId":"anchor","x-cartulary-state-changing":true,"security":[{"bearerSession":[],"credentialBootstrapBearer":[],"csrfCookie":[],"csrfHeader":[],"sessionCookie":[]}],` +
			`"responses":{"200":{"description":"ok"},` + session + `,` + csrf + `,` + bootstrap + `}}},` +
			`"/a":{"post":{"operationId":"a","x-cartulary-state-changing":true,"security":` + security + `,` +
			`"responses":{` + strings.Join(append([]string{`"200":{"description":"ok"}`}, responses...), ",") + `}}}},` +
			`"components":{"responses":{` +
			`"Session":{"description":"s","x-cartulary-error-codes":["session_required"]},` +
			`"Credentials":{"description":"c","x-cartulary-error-codes":["invalid_credentials"]},` +
			`"Conflict":{"description":"c","x-cartulary-error-codes":["alpha_conflict"]}}}}`
	}
	for _, tc := range []struct {
		name, document, wantError string
	}{
		{"cookie with CSRF or bearer", document(`[`+cookieCSRF+`,`+bearer+`]`, session, csrf, bootstrap), ""},
		{"cookie without CSRF", document(`[{"sessionCookie":[]}]`, session), ""},
		{"credential bootstrap bearer", document(`[{"credentialBootstrapBearer":[]}]`, session, bootstrap), ""},
		{"public operation with other codes", document(`[]`,
			`"401":{"$ref":"#/components/responses/Credentials"}`,
			`"403":{"description":"d","x-cartulary-error-codes":["authorization_denied"]}`,
			`"409":{"$ref":"#/components/responses/Conflict"}`), ""},
		{"implied through a component", document(`[{"sessionCookie":[]}]`, `"401":{"$ref":"#/components/responses/Session"}`), ""},
		{"secured without 401", document(`[{"sessionCookie":[]}]`), "POST /a 401: session_required must be declared because the operation declares a security requirement"},
		{"secured 401 without session_required", document(`[{"sessionCookie":[]}]`, `"401":{"$ref":"#/components/responses/Credentials"}`), "POST /a 401: session_required must be declared"},
		{"public with session_required", document(`[]`, session), "POST /a 401: session_required must not be declared because the operation declares no security requirement"},
		{"public with session_required through a component", document(`[]`, `"401":{"$ref":"#/components/responses/Session"}`), "session_required must not be declared"},
		{"csrfHeader without 403", document(`[`+cookieCSRF+`]`, session), "POST /a 403: csrf_verification_failed must be declared because an alternative includes csrfHeader"},
		{"csrf declared without csrfHeader", document(`[{"sessionCookie":[]},`+bearer+`]`, session, csrf, bootstrap), "POST /a 403: csrf_verification_failed must not be declared because no alternative includes csrfHeader"},
		{"bearer without 409", document(`[`+bearer+`]`, session), "POST /a 409: credential_bootstrap_rejected must be declared because an alternative includes bearerSession or credentialBootstrapBearer"},
		{"credential bootstrap bearer without credential_bootstrap_rejected", document(`[{"credentialBootstrapBearer":[]}]`, session, `"409":{"$ref":"#/components/responses/Conflict"}`), "credential_bootstrap_rejected must be declared"},
		{"credential_bootstrap_rejected without bearer", document(`[`+cookieCSRF+`]`, session, csrf, bootstrap), "POST /a 409: credential_bootstrap_rejected must not be declared because no alternative includes bearerSession or credentialBootstrapBearer"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): tc.document, errorRegistryPath: registry})
			_, _, err := Assemble(fixture.manifestPath)
			switch {
			case tc.wantError == "" && err != nil:
				t.Fatalf("assemble: %v", err)
			case tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)):
				t.Fatalf("assemble error = %v, want substring %q", err, tc.wantError)
			}
		})
	}
}

// TestAssemblyEnforcesOperationImpliedErrorCodes covers the declarations
// every operation implies under Core 01 REQ-01-702 (AC-608), each in both
// directions: the 500 response declares internal_error and the 503 response
// declares service_unavailable on every operation, and the 404 response
// declares extension_profile_not_claimed exactly when the operation belongs to
// a reserved extension route family, which it states through
// x-cartulary-availability. An implied code may share a component or join a
// response's other codes in sorted order. Either code declared at another
// status is a registry status mismatch.
func TestAssemblyEnforcesOperationImpliedErrorCodes(t *testing.T) {
	const registry = `{"errors":[{"code":"credential_bootstrap_rejected","http_status":409},{"code":"csrf_verification_failed","http_status":403},` +
		`{"code":"extension_profile_not_claimed","http_status":404},{"code":"object_store_invalid_request","http_status":500},` +
		`{"code":"object_store_unavailable","http_status":503},{"code":"session_required","http_status":401},{"code":"thing_conflict","http_status":409},` +
		`{"code":"thing_not_found","http_status":404},` + fixtureImpliedRegistryEntries + `]}`
	const (
		internal    = `"500":{"$ref":"#/components/responses/Internal"}`
		unavailable = `"503":{"$ref":"#/components/responses/Unavailable"}`
		reserved    = `,"x-cartulary-availability":"fixture_profile"`
	)
	// The anchor operation uses every scheme, as the assembled document must,
	// and is itself correctly declared; each case varies operation /a only,
	// which declares no security.
	document := func(availability string, responses ...string) string {
		return `{"openapi":"3.1.0","info":{"title":"fixture","version":"1"},"paths":{` +
			`"/anchor":{"post":{"operationId":"anchor","x-cartulary-state-changing":true,"security":[{"bearerSession":[],"credentialBootstrapBearer":[],"csrfCookie":[],"csrfHeader":[],"sessionCookie":[]}],` +
			`"responses":{"200":{"description":"ok"},"401":{"description":"s","x-cartulary-error-codes":["session_required"]},` +
			`"403":{"description":"c","x-cartulary-error-codes":["csrf_verification_failed"]},"409":{"description":"b","x-cartulary-error-codes":["credential_bootstrap_rejected"]},` +
			internal + `,` + unavailable + `}}},` +
			`"/a":{"post":{"operationId":"a","x-cartulary-state-changing":true,"security":[]` + availability + `,` +
			`"responses":{` + strings.Join(append([]string{`"200":{"description":"ok"}`}, responses...), ",") + `}}}},` +
			`"components":{"responses":{` +
			`"Internal":{"description":"i","x-cartulary-error-codes":["internal_error"]},` +
			`"Unavailable":{"description":"u","x-cartulary-error-codes":["service_unavailable"]},` +
			`"NotClaimed":{"description":"n","x-cartulary-error-codes":["extension_profile_not_claimed"]},` +
			`"ThingNotFound":{"description":"t","x-cartulary-error-codes":["thing_not_found"]}}}}`
	}
	for _, tc := range []struct {
		name, document, wantError string
	}{
		{"shared implied responses", document("", internal, unavailable), ""},
		{"inline implied responses", document("", `"500":{"description":"i","x-cartulary-error-codes":["internal_error"]}`, `"503":{"description":"u","x-cartulary-error-codes":["service_unavailable"]}`), ""},
		{"implied codes among other codes", document("", `"500":{"description":"i","x-cartulary-error-codes":["internal_error","object_store_invalid_request"]}`,
			`"503":{"description":"u","x-cartulary-error-codes":["object_store_unavailable","service_unavailable"]}`), ""},
		{"reserved family through a shared 404", document(reserved, `"404":{"$ref":"#/components/responses/NotClaimed"}`, internal, unavailable), ""},
		{"reserved family among other 404 codes", document(reserved, `"404":{"description":"n","x-cartulary-error-codes":["extension_profile_not_claimed","thing_not_found"]}`, internal, unavailable), ""},
		{"base operation with another 404", document("", `"404":{"$ref":"#/components/responses/ThingNotFound"}`, internal, unavailable), ""},

		{"without 500", document("", unavailable), "POST /a 500: internal_error must be declared because every operation may fail internally (REQ-01-229)"},
		{"500 without internal_error", document("", `"500":{"description":"i","x-cartulary-error-codes":["object_store_invalid_request"]}`, unavailable), "POST /a 500: internal_error must be declared"},
		{"without 503", document("", internal), "POST /a 503: service_unavailable must be declared because the admission gate may refuse every operation"},
		{"503 without service_unavailable", document("", internal, `"503":{"description":"u","x-cartulary-error-codes":["object_store_unavailable"]}`), "POST /a 503: service_unavailable must be declared"},
		{"service_unavailable under 500", document("", `"500":{"description":"i","x-cartulary-error-codes":["internal_error","service_unavailable"]}`, unavailable), `POST /a 500: x-cartulary-error-codes names "service_unavailable", which the registry assigns status 503`},
		{"internal_error under 503", document("", internal, `"503":{"description":"u","x-cartulary-error-codes":["internal_error","service_unavailable"]}`), `POST /a 503: x-cartulary-error-codes names "internal_error", which the registry assigns status 500`},
		{"reserved family without 404", document(reserved, internal, unavailable), "POST /a 404: extension_profile_not_claimed must be declared because the operation declares x-cartulary-availability, so it belongs to a reserved extension route family"},
		{"reserved family 404 without the code", document(reserved, `"404":{"$ref":"#/components/responses/ThingNotFound"}`, internal, unavailable), "POST /a 404: extension_profile_not_claimed must be declared"},
		{"base operation with the code", document("", `"404":{"description":"n","x-cartulary-error-codes":["extension_profile_not_claimed","thing_not_found"]}`, internal, unavailable), "POST /a 404: extension_profile_not_claimed must not be declared because the operation declares no x-cartulary-availability, so it belongs to no reserved extension route family"},
		{"base operation sharing the reserved 404", document("", `"404":{"$ref":"#/components/responses/NotClaimed"}`, internal, unavailable), "extension_profile_not_claimed must not be declared"},
		{"extension_profile_not_claimed under 409", document(reserved, `"404":{"$ref":"#/components/responses/NotClaimed"}`, `"409":{"description":"c","x-cartulary-error-codes":["extension_profile_not_claimed","thing_conflict"]}`, internal, unavailable), `POST /a 409: x-cartulary-error-codes names "extension_profile_not_claimed", which the registry assigns status 404`},
		{"empty availability", document(`,"x-cartulary-availability":""`, internal, unavailable), "POST /a x-cartulary-availability must be a non-empty string"},
		{"non-string availability", document(`,"x-cartulary-availability":true`, internal, unavailable), "POST /a x-cartulary-availability must be a non-empty string"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newAssemblyFixture(t, rootOnlyUnits(), map[string]string{rootUnitPath(): tc.document, errorRegistryPath: registry})
			// Write the case exactly as given: fixture completion would add the
			// implied responses the case omits.
			writeFixtureFile(t, fixture.root, rootUnitPath(), fixtureDocument(t, tc.document, true, false))
			_, _, err := Assemble(fixture.manifestPath)
			switch {
			case tc.wantError == "" && err != nil:
				t.Fatalf("assemble: %v", err)
			case tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)):
				t.Fatalf("assemble error = %v, want substring %q", err, tc.wantError)
			}
		})
	}
	t.Run("the canonical registry registers the implied codes", func(t *testing.T) {
		repository, err := os.OpenRoot(filepath.Join("..", "..", ".."))
		if err != nil {
			t.Fatal(err)
		}
		defer repository.Close()
		canonical, err := loadErrorRegistry(repository)
		if err != nil {
			t.Fatal(err)
		}
		for code, status := range map[string]int{"internal_error": 500, "service_unavailable": 503, "extension_profile_not_claimed": 404} {
			if canonical[code] != status {
				t.Fatalf("registry assigns %s status %d, want %d", code, canonical[code], status)
			}
		}
	})
}
