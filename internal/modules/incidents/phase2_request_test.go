package incidents

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/contracts"
	moduleextensions "github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestPhase2_U_2_01_IncidentCreateAcceptsDeclaredMembersAndNormalizesIncidentKey(t *testing.T) {
	request, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01",
		"incident_key":"  IR-E\u0301-2026-001  ",
		"title":"  Incident E\u0301xample  ",
		"description":"  First line\r\nSecond line  ",
		"severity":"  high  ",
		"tlp":"TLP:AMBER",
		"current_phase":"  triage  ",
		"primary_external_case_ref":"  CASE-42  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident create request: %v", apiErr)
	}
	requireWritableStringNormalization(t, request.IncidentKey, "IR-\u00C9-2026-001")
	requireWritableStringNormalization(t, request.Title, "Incident \u00C9xample")
	if request.TLP == nil || *request.TLP != "TLP:AMBER" {
		t.Fatalf("expected canonical TLP token, got %#v", request.TLP)
	}
	if request.Description == nil {
		t.Fatal("expected normalized description")
	}
	requireWritableStringNormalization(t, *request.Description, "First line\nSecond line")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-memberships",
		"incident_key":"IR-2026-002",
		"title":"Example",
		"initial_memberships":[]
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "initial_memberships", "collaborator_seeding_not_supported")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-top-level",
		"incident_key":"IR-2026-003",
		"title":"Example",
		"unexpected":true
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "unexpected", "unknown_field")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-server-managed",
		"incident_key":"IR-2026-004",
		"title":"Example",
		"incident_version":1
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "incident_version", "server_managed_field")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-tlp",
		"incident_key":"IR-2026-005",
		"title":"Example",
		"tlp":"amber"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "tlp", "invalid_value")
}

func TestSupportPhase2_IncidentRequestStringContracts(t *testing.T) {
	createPayload := func(overrides map[string]any) string {
		payload := map[string]any{
			"client_txn_id": "txn-string-contract",
			"incident_key":  "IR-STRING-CONTRACT",
			"title":         "String Contract",
		}
		for key, value := range overrides {
			payload[key] = value
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal create payload: %v", err)
		}
		return string(encoded)
	}

	createCases := []struct {
		name  string
		field string
		value any
	}{
		{name: "incident key max bytes", field: "incident_key", value: strings.Repeat("\u00e9", 65)},
		{name: "title max scalar values", field: "title", value: strings.Repeat("a", 513)},
		{name: "description max scalar values", field: "description", value: strings.Repeat("a", 16385)},
		{name: "metadata max scalar values", field: "severity", value: strings.Repeat("a", 129)},
		{name: "tlp exact token", field: "tlp", value: " TLP:AMBER "},
	}
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeIncidentCreateRequest(strings.NewReader(createPayload(map[string]any{tc.field: tc.value})))
			requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", tc.field, "invalid_value")
		})
	}

	patchPayload, err := json.Marshal(map[string]any{
		"base_incident_version": 3,
		"current_phase":         " \t ",
	})
	if err != nil {
		t.Fatalf("marshal patch payload: %v", err)
	}
	patchRequest, apiErr := DecodeIncidentPatchRequest(strings.NewReader(string(patchPayload)))
	if apiErr != nil {
		t.Fatalf("decode whitespace metadata clear patch: %v", apiErr)
	}
	if !patchRequest.CurrentPhase.Present || patchRequest.CurrentPhase.Value != nil {
		t.Fatalf("whitespace metadata patch must clear to null, got %#v", patchRequest.CurrentPhase)
	}

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":3,
		"tlp":" TLP:AMBER "
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "tlp", "invalid_value")
}

func TestSupportPhase2_IncidentListQueryUsesCoreListQueryErrors(t *testing.T) {
	valid, apiErr := parseIncidentListScope("search=CASE+High&status=active&limit=1")
	if apiErr != nil {
		t.Fatalf("parse valid incident list query: %v", apiErr)
	}
	if valid.Scope["search"] != "case high" || valid.Scope["status"] != "active" {
		t.Fatalf("unexpected canonical incident list scope: %#v", valid.Scope)
	}

	cases := []struct {
		name       string
		query      string
		code       string
		reasonCode string
	}{
		{
			name:       "unknown query member",
			query:      "group_by=status",
			code:       "invalid_list_query",
			reasonCode: "unknown_query_member",
		},
		{
			name:       "duplicate member",
			query:      "status=active&status=closed",
			code:       "invalid_list_query",
			reasonCode: "duplicate_query_member",
		},
		{
			name:       "invalid status",
			query:      "status=archived",
			code:       "invalid_list_query",
			reasonCode: "invalid_filter_value",
		},
		{
			name:       "pagination alias",
			query:      "page=2",
			code:       "invalid_pagination_request",
			reasonCode: "invalid_limit",
		},
		{
			name:       "invalid search",
			query:      "search=---",
			code:       "invalid_list_query",
			reasonCode: "invalid_search",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := parseIncidentListScope(tc.query)
			requireAPIError(t, apiErr, http.StatusBadRequest, tc.code, "", tc.reasonCode)
		})
	}
}

func TestPhase2_U_2_05_IncidentPatchAllowsPromotedFieldsAndKeepsNoOpVersionStable(t *testing.T) {
	request, apiErr := DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"tlp":null,
		"current_phase":"  containment  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident patch request: %v", apiErr)
	}
	if !request.TLP.Present || request.TLP.Value != nil {
		t.Fatalf("expected tlp explicit null clear, got %#v", request.TLP)
	}
	if !request.CurrentPhase.Present || request.CurrentPhase.Value == nil || *request.CurrentPhase.Value != "containment" {
		t.Fatalf("unexpected current_phase patch field: %#v", request.CurrentPhase)
	}
	if request.PrimaryExternalCaseRef.Present {
		t.Fatalf("unexpected omitted field to be marked present: %#v", request.PrimaryExternalCaseRef)
	}

	current := IncidentRecord{
		ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000905"),
		TLP:                    stringRef("TLP:AMBER"),
		CurrentPhase:           stringRef("containment"),
		PrimaryExternalCaseRef: stringRef("CASE-1"),
		UpdatedAt:              timeRef(2026, 4, 17, 12, 0),
		IncidentVersion:        7,
	}
	noOp, changed := ApplyIncidentPatch(current, IncidentPatchRequest{BaseIncidentVersion: 7}, uuid.MustParse("00000000-0000-0000-0000-000000000777"), timeRef(2026, 4, 17, 13, 0))
	if changed {
		t.Fatalf("expected structurally valid no-op patch to remain version-stable: %#v", noOp)
	}
	if noOp.IncidentVersion != current.IncidentVersion || !noOp.UpdatedAt.Equal(current.UpdatedAt) {
		t.Fatalf("no-op patch must keep version and updated_at stable: before=%#v after=%#v", current, noOp)
	}

	material, changed := ApplyIncidentPatch(current, IncidentPatchRequest{
		BaseIncidentVersion:    7,
		TLP:                    OptionalNullableString{Present: true, Value: stringRef("TLP:GREEN")},
		PrimaryExternalCaseRef: OptionalNullableString{Present: true, Value: nil},
	}, uuid.MustParse("00000000-0000-0000-0000-000000000777"), timeRef(2026, 4, 17, 13, 0))
	if !changed {
		t.Fatal("expected material patch to change promoted fields")
	}
	if material.IncidentVersion != 8 || material.TLP == nil || *material.TLP != "TLP:GREEN" || material.PrimaryExternalCaseRef != nil {
		t.Fatalf("unexpected material patch projection: %#v", material)
	}

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"title":"forbidden"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "title", "forbidden_field")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"tlp":"TLP:AMBER"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "base_incident_version", "missing_required_field")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"tlp":"amber"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "tlp", "invalid_value")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"unknown":"field"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "unknown", "unknown_field")
}

func TestPhase2_U_2_06_MembershipCreateUsesLookupOnlyForUserOrEmailTargets(t *testing.T) {
	request, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-email",
		"email":"  Analyst@Example.Test  ",
		"role":"reviewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership create request: %v", apiErr)
	}
	if request.Email == nil {
		t.Fatal("expected normalized email target")
	}
	requireWritableStringNormalization(t, *request.Email, "Analyst@Example.Test")
	emailLookup := &membershipTargetLookupStub{
		getByEmail: func(_ context.Context, email string) (authn.UserRecord, error) {
			if email != "Analyst@Example.Test" {
				t.Fatalf("unexpected email lookup target: %q", email)
			}
			return authn.UserRecord{
				ID:          uuid.MustParse("00000000-0000-0000-0000-000000000603"),
				Email:       email,
				DisplayName: "Analyst",
				IsActive:    true,
			}, nil
		},
	}
	target, apiErr := resolveMembershipTarget(context.Background(), emailLookup, request)
	if apiErr != nil {
		t.Fatalf("resolve normalized email target: %v", apiErr)
	}
	if target.Email != "Analyst@Example.Test" || emailLookup.emailCalls != 1 || emailLookup.idCalls != 0 {
		t.Fatalf("membership email target must use lookup-only resolution: target=%#v lookup=%#v", target, emailLookup)
	}

	userOnly, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-user",
		"user_id":"00000000-0000-0000-0000-000000000601",
		"role":"viewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode user-target membership create request: %v", apiErr)
	}
	if userOnly.UserID == nil || userOnly.Email != nil {
		t.Fatalf("expected user_id-only membership target, got %#v", userOnly)
	}
	userLookup := &membershipTargetLookupStub{
		getByID: func(_ context.Context, userID uuid.UUID) (authn.UserRecord, error) {
			if userOnly.UserID == nil || userID != *userOnly.UserID {
				t.Fatalf("unexpected user_id lookup target: %s", userID)
			}
			return authn.UserRecord{
				ID:          userID,
				Email:       "lookup-user@example.test",
				DisplayName: "Lookup User",
				IsActive:    true,
			}, nil
		},
	}
	target, apiErr = resolveMembershipTarget(context.Background(), userLookup, userOnly)
	if apiErr != nil {
		t.Fatalf("resolve user_id target: %v", apiErr)
	}
	if userOnly.UserID == nil || target.ID != *userOnly.UserID || userLookup.idCalls != 1 || userLookup.emailCalls != 0 {
		t.Fatalf("membership user_id target must use lookup-only resolution: target=%#v lookup=%#v", target, userLookup)
	}

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-missing-target",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-dual-target",
		"user_id":"00000000-0000-0000-0000-000000000602",
		"email":"dual@example.test",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-invalid-role",
		"email":"solo@example.test",
		"role":"owner"
	}`))
	requireClosedVocabularyRejected(t, apiErr, "role", "invalid_role")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-no-invite",
		"email":"solo@example.test",
		"role":"viewer",
		"invitation_email":"new@example.test"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "invitation_email", "unknown_field")

	notFoundRequest, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-not-found",
		"email":"missing@example.test",
		"role":"viewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode not-found membership create request: %v", apiErr)
	}
	_, apiErr = resolveMembershipTarget(context.Background(), &membershipTargetLookupStub{
		getByEmail: func(context.Context, string) (authn.UserRecord, error) {
			return authn.UserRecord{}, authn.ErrNotFound
		},
	}, notFoundRequest)
	requireAPIError(t, apiErr, http.StatusNotFound, "user_not_found", "", "")

	inactiveRequest, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-inactive",
		"email":"inactive@example.test",
		"role":"viewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode inactive membership create request: %v", apiErr)
	}
	_, apiErr = resolveMembershipTarget(context.Background(), &membershipTargetLookupStub{
		getByEmail: func(context.Context, string) (authn.UserRecord, error) {
			return authn.UserRecord{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000604"),
				Email:    "inactive@example.test",
				IsActive: false,
			}, nil
		},
	}, inactiveRequest)
	requireAPIError(t, apiErr, http.StatusConflict, "user_inactive", "", "")
}

func TestSupportPhase2_OpenAPIWorkbookPreferencesExposeGetAndPutContracts(t *testing.T) {
	artifact, ok := contracts.ContractArtifactIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("missing generated OpenAPI contract artifact")
	}

	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI contract artifact: %v", err)
	}

	defaultPath := openAPIObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/workbook-preferences/default")
	requireOpenAPIOperation(t, defaultPath, "get", "getIncidentDefaultWorkbookPreferences")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, defaultPath, "get"), "DefaultWorkbookPreferencesEnvelope")
	requireOpenAPIOperation(t, defaultPath, "put", "putIncidentDefaultWorkbookPreferences")
	requireOpenAPIRequestSchemaRef(t, openAPIObjectAt(t, defaultPath, "put"), "DefaultWorkbookPreferencesPutRequest")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, defaultPath, "put"), "DefaultWorkbookPreferencesEnvelope")

	userPath := openAPIObjectAt(t, document, "paths", "/api/v1/incidents/{incident_id}/workbook-preferences/me")
	requireOpenAPIOperation(t, userPath, "get", "getCurrentUserWorkbookPreferences")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, userPath, "get"), "UserWorkbookPreferencesEnvelope")
	requireOpenAPIOperation(t, userPath, "put", "putCurrentUserWorkbookPreferences")
	requireOpenAPIRequestSchemaRef(t, openAPIObjectAt(t, userPath, "put"), "UserWorkbookPreferencesPutRequest")
	requireOpenAPIResponseSchemaRef(t, openAPIObjectAt(t, userPath, "put"), "UserWorkbookPreferencesEnvelope")

	schemas := openAPIObjectAt(t, document, "components", "schemas")
	requireOpenAPISheetRefSchema(t, openAPIObjectAt(t, schemas, "SheetRef"))
	requireOpenAPIWorkbookPreferencesPutSchema(t, openAPIObjectAt(t, schemas, "DefaultWorkbookPreferencesPutRequest"), "default_sheet_ref")
	requireOpenAPIWorkbookPreferencesPutSchema(t, openAPIObjectAt(t, schemas, "UserWorkbookPreferencesPutRequest"), "home_sheet_ref")
	requireOpenAPIEnvelopeSchema(t, openAPIObjectAt(t, schemas, "DefaultWorkbookPreferencesEnvelope"), "DefaultWorkbookPreferencesResource")
	requireOpenAPIEnvelopeSchema(t, openAPIObjectAt(t, schemas, "UserWorkbookPreferencesEnvelope"), "UserWorkbookPreferencesResource")
}

func TestSupportPhase2_OpenAPIExtensionDiscoveryExposesClosedContract(t *testing.T) {
	artifact, ok := contracts.ContractArtifactIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("missing generated OpenAPI contract artifact")
	}

	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI contract artifact: %v", err)
	}

	extensionsPath := openAPIObjectAt(t, document, "paths", "/api/v1/extensions")
	requireOpenAPIOperation(t, extensionsPath, "get", "listDeploymentExtensions")
	operation := openAPIObjectAt(t, extensionsPath, "get")
	requireOpenAPIResponseSchemaRef(t, operation, "ExtensionDiscoveryEnvelope")
	requireOpenAPIStatusResponseSchemaRef(t, operation, "400", "ErrorEnvelope")
	requireOpenAPIStatusResponseSchemaRef(t, operation, "401", "ErrorEnvelope")

	schemas := openAPIObjectAt(t, document, "components", "schemas")
	requireOpenAPIEnvelopeSchema(t, openAPIObjectAt(t, schemas, "ExtensionDiscoveryEnvelope"), "ExtensionDiscoveryData")

	data := openAPIObjectAt(t, schemas, "ExtensionDiscoveryData")
	if data["type"] != "object" || data["additionalProperties"] != false {
		t.Fatalf("extension discovery data must be a closed object schema: %#v", data)
	}
	if required := toStrings(t, data["required"]); !equalStringSlices(required, []string{"extensions"}) {
		t.Fatalf("unexpected extension discovery data required fields: %v", required)
	}
	extensionsItems := openAPIObjectAt(t, data, "properties", "extensions", "items")
	if extensionsItems["$ref"] != "#/components/schemas/ExtensionProfileResource" {
		t.Fatalf("unexpected extension discovery items ref: %#v", extensionsItems)
	}

	resource := openAPIObjectAt(t, schemas, "ExtensionProfileResource")
	if resource["type"] != "object" || resource["additionalProperties"] != false {
		t.Fatalf("extension profile resource must be a closed object schema: %#v", resource)
	}
	if required := toStrings(t, resource["required"]); !equalStringSlices(required, []string{"profile_id", "claimed", "route_families"}) {
		t.Fatalf("unexpected extension profile required fields: %v", required)
	}
	properties := openAPIObjectAt(t, resource, "properties")
	if len(properties) != 3 {
		t.Fatalf("extension profile resource must expose only profile_id, claimed, and route_families: %#v", properties)
	}
	if profileID := openAPIObjectAt(t, properties, "profile_id"); profileID["$ref"] != "#/components/schemas/ExtensionProfileID" {
		t.Fatalf("unexpected extension profile_id schema: %#v", profileID)
	}
	if claimed := openAPIObjectAt(t, properties, "claimed"); claimed["type"] != "boolean" {
		t.Fatalf("unexpected extension claimed schema: %#v", claimed)
	}
	routeFamilies := openAPIObjectAt(t, properties, "route_families", "items")
	if routeFamilies["$ref"] != "#/components/schemas/ExtensionRouteFamily" {
		t.Fatalf("unexpected extension route_families item schema: %#v", routeFamilies)
	}

	profileIDSchema := openAPIObjectAt(t, schemas, "ExtensionProfileID")
	if profileIDSchema["type"] != "string" {
		t.Fatalf("extension profile id schema must be string: %#v", profileIDSchema)
	}
	if enum := toStrings(t, profileIDSchema["enum"]); !equalStringSlices(enum, []string{"enterprise_authentication", "import", "incident_portability", "network_flow_activity", "reference_pack", "snapshot_reporting"}) {
		t.Fatalf("unexpected extension profile id enum: %v", enum)
	}

	routeFamilySchema := openAPIObjectAt(t, schemas, "ExtensionRouteFamily")
	if routeFamilySchema["type"] != "string" {
		t.Fatalf("extension route family schema must be string: %#v", routeFamilySchema)
	}
	if enum := toStrings(t, routeFamilySchema["enum"]); !equalStringSlices(enum, []string{
		"/api/v1/auth/oidc",
		"/api/v1/auth/providers",
		"/api/v1/auth/saml",
		"/api/v1/import-sessions",
		"/api/v1/incidents/{incident_id}/network-flow",
		"/api/v1/incidents/{incident_id}/report-compositions",
		"/api/v1/incident-bundles",
		"/api/v1/reference-packs",
		"/api/v1/releases",
		"/api/v1/snapshots",
		"/api/v1/users/{user_id}/auth-bindings",
	}) {
		t.Fatalf("unexpected extension route family enum: %v", enum)
	}
}

func TestPhase2_U_2_09_ExtensionDiscoveryReturnsExactSingletonProfileShape(t *testing.T) {
	query := url.Values{"cursor_token": []string{"opaque"}}
	apiErr := httpapi.ValidateSingletonReadQuery(query)
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_pagination_request", "", "pagination_not_supported")

	data := moduleextensions.BuildResponseData(httpapi.CurrentExtensionProfiles())
	extensions, ok := data["extensions"].([]map[string]any)
	if !ok {
		t.Fatalf("unexpected extensions response data: %#v", data)
	}
	wantProfiles := currentProfileExtensions(t)
	if len(extensions) != len(wantProfiles) {
		t.Fatalf("unexpected extension profile count: got %d want %d", len(extensions), len(wantProfiles))
	}

	for index, want := range wantProfiles {
		got := extensions[index]
		if len(got) != 3 {
			t.Fatalf("extension resource must expose only profile_id, claimed, and route_families: %#v", got)
		}
		if got["profile_id"] != want.ProfileID || got["claimed"] != httpapi.ExtensionProfileClaimed(want.ProfileID) {
			t.Fatalf("unexpected extension resource at index %d: %#v", index, got)
		}
		if families := toStrings(t, got["route_families"]); !equalStringSlices(families, want.RouteFamilies) {
			t.Fatalf("unexpected route_families at index %d: got %v want %v", index, families, want.RouteFamilies)
		}
	}
}

func TestPhase2_U_2_10_ReservedExtensionDispatchHonorsBaseRoutesClaimedFamiliesAndOutsideFallback(t *testing.T) {
	importProfile := extensionContract(t, "import")
	enterpriseProfile := extensionContract(t, "enterprise_authentication")

	baseHandler, err := httpapi.NewHandler(httpapi.Options{})
	if err != nil {
		t.Fatalf("build base handler: %v", err)
	}
	ready := performRequest(t, baseHandler, http.MethodGet, "/readyz", nil)
	if ready.Code != http.StatusOK || !strings.Contains(ready.Body.String(), "ready") {
		t.Fatalf("expected base route to retain precedence, got status=%d body=%q", ready.Code, ready.Body.String())
	}

	restoreClaimed := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       true,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
	})
	claimedHandler, err := httpapi.NewHandler(httpapi.Options{
		AdditionalRoutes: []httpapi.RouteRegistrar{
			func(mux *http.ServeMux, deps httpapi.DependencySet) error {
				mux.HandleFunc(importProfile.RouteFamilies[0], func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				mux.HandleFunc(importProfile.RouteFamilies[0]+"/children", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				})
				return nil
			},
		},
	})
	restoreClaimed()
	if err != nil {
		t.Fatalf("build claimed handler: %v", err)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil); got.Code != http.StatusNoContent {
		t.Fatalf("claimed family root must dispatch to the registered route, got %d", got.Code)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0]+"/children", nil); got.Code != http.StatusAccepted {
		t.Fatalf("claimed family descendant must dispatch to the registered route, got %d", got.Code)
	}

	restoreUnclaimed := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
		{
			ProfileID:     enterpriseProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), enterpriseProfile.RouteFamilies...),
		},
	})
	defer restoreUnclaimed()
	unclaimedHandler, err := httpapi.NewHandler(httpapi.Options{})
	if err != nil {
		t.Fatalf("build unclaimed handler: %v", err)
	}

	rootReserved := performRequest(t, unclaimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil)
	rootBody := readJSONMap(t, rootReserved.Body.String())
	errorBody := rootBody["error"].(map[string]any)
	requireErrorContract(t, "extension_profile_not_claimed", http.StatusNotFound)
	if rootReserved.Code != http.StatusNotFound || errorBody["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved root response: status=%d body=%#v", rootReserved.Code, rootBody)
	}
	rootDetails := errorBody["details"].(map[string]any)
	if rootDetails["profile_id"] != importProfile.ProfileID || rootDetails["route_family"] != importProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved root details: %#v", rootDetails)
	}

	descendantReserved := performRequest(t, unclaimedHandler, http.MethodGet, strings.Replace(enterpriseProfile.RouteFamilies[0], "{user_id}", uuid.NewString(), 1)+"/provider", nil)
	descendantBody := readJSONMap(t, descendantReserved.Body.String())
	descendantError := descendantBody["error"].(map[string]any)
	if descendantReserved.Code != http.StatusNotFound || descendantError["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved descendant response: status=%d body=%#v", descendantReserved.Code, descendantBody)
	}
	descendantDetails := descendantError["details"].(map[string]any)
	if descendantDetails["profile_id"] != enterpriseProfile.ProfileID || descendantDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved descendant details: %#v", descendantDetails)
	}

	outsideReserved := performRequest(t, unclaimedHandler, http.MethodGet, "/api/v1/outside-reserved-families", nil)
	if outsideReserved.Code != http.StatusNotFound || strings.Contains(outsideReserved.Body.String(), "extension_profile_not_claimed") {
		t.Fatalf("ordinary unknown paths must keep ordinary not-found handling: status=%d body=%q", outsideReserved.Code, outsideReserved.Body.String())
	}
}

type membershipTargetLookupStub struct {
	emailCalls int
	getByEmail func(context.Context, string) (authn.UserRecord, error)
	getByID    func(context.Context, uuid.UUID) (authn.UserRecord, error)
	idCalls    int
}

func (s *membershipTargetLookupStub) GetUserByID(ctx context.Context, userID uuid.UUID) (authn.UserRecord, error) {
	s.idCalls++
	if s.getByID == nil {
		return authn.UserRecord{}, nil
	}
	return s.getByID(ctx, userID)
}

func (s *membershipTargetLookupStub) GetUserByNormalizedEmail(ctx context.Context, email string) (authn.UserRecord, error) {
	s.emailCalls++
	if s.getByEmail == nil {
		return authn.UserRecord{}, nil
	}
	return s.getByEmail(ctx, email)
}

func requireAPIError(t testing.TB, apiErr *httpapi.APIError, wantStatus int, wantCode string, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
		return
	}
	requireErrorContract(t, wantCode, wantStatus)
	if apiErr.Status != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", apiErr.Status, wantStatus)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("unexpected code: got %q want %q", apiErr.Code, wantCode)
	}
	if wantField != "" {
		if got := apiErr.Details["field"]; got != wantField {
			t.Fatalf("unexpected field detail: got %v want %s", got, wantField)
		}
	}
	if wantReasonCode != "" {
		if got := apiErr.Details["reason_code"]; got != wantReasonCode {
			t.Fatalf("unexpected reason_code detail: got %v want %s", got, wantReasonCode)
		}
	}
}

func requireClosedVocabularyRejected(t testing.TB, apiErr *httpapi.APIError, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
		return
	}
	if apiErr.Code != "invalid_mutation_payload" && apiErr.Code != "invalid_view_query" {
		t.Fatalf("unexpected closed-vocabulary rejection code: %q", apiErr.Code)
	}
	if apiErr.Details == nil {
		t.Fatal("expected closed-vocabulary rejection details")
	}
	if apiErr.Details["field"] != wantField {
		t.Fatalf("unexpected closed-vocabulary field: got %v want %q", apiErr.Details["field"], wantField)
	}
	if apiErr.Details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected closed-vocabulary reason_code: got %v want %q", apiErr.Details["reason_code"], wantReasonCode)
	}
}

func requireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func extensionContract(t testing.TB, profileID string) extensionProfileContract {
	t.Helper()

	for _, profile := range currentProfileExtensions(t) {
		if profile.ProfileID == profileID {
			return profile
		}
	}
	t.Fatalf("missing extension contract for %q", profileID)
	return extensionProfileContract{}
}

func stringRef(value string) *string {
	return &value
}

func performRequest(t testing.TB, handler http.Handler, method string, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func readJSONMap(t testing.TB, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode JSON body: %v", err)
	}
	return payload
}

func toStrings(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]string)
	if ok {
		return append([]string(nil), items...)
	}
	interfaces, ok := raw.([]any)
	if !ok {
		t.Fatalf("unexpected string slice payload: %T", raw)
	}
	values := make([]string, 0, len(interfaces))
	for _, item := range interfaces {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("unexpected string item payload: %T", item)
		}
		values = append(values, value)
	}
	return values
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func openAPIObjectAt(t testing.TB, root any, path ...string) map[string]any {
	t.Helper()

	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected OpenAPI object before %q, got %T", segment, current)
		}
		var found bool
		current, found = object[segment]
		if !found {
			t.Fatalf("missing OpenAPI path segment %q in %#v", segment, object)
		}
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI object at %v, got %T", path, current)
	}
	return object
}

func requireOpenAPIOperation(t testing.TB, path map[string]any, method string, wantOperationID string) {
	t.Helper()

	operation := openAPIObjectAt(t, path, method)
	if operation["operationId"] != wantOperationID {
		t.Fatalf("unexpected %s operationId: got %v want %q", method, operation["operationId"], wantOperationID)
	}
}

func requireOpenAPIRequestSchemaRef(t testing.TB, operation map[string]any, wantSchemaName string) {
	t.Helper()

	requestBody := openAPIObjectAt(t, operation, "requestBody")
	if requestBody["required"] != true {
		t.Fatalf("OpenAPI requestBody must be required: %#v", requestBody)
	}
	schema := openAPIObjectAt(t, requestBody, "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchemaName
	if schema["$ref"] != wantRef {
		t.Fatalf("unexpected request schema ref: got %v want %q", schema["$ref"], wantRef)
	}
}

func requireOpenAPIResponseSchemaRef(t testing.TB, operation map[string]any, wantSchemaName string) {
	t.Helper()

	requireOpenAPIStatusResponseSchemaRef(t, operation, "200", wantSchemaName)
}

func requireOpenAPIStatusResponseSchemaRef(t testing.TB, operation map[string]any, status string, wantSchemaName string) {
	t.Helper()

	schema := openAPIObjectAt(t, operation, "responses", status, "content", "application/json", "schema")
	wantRef := "#/components/schemas/" + wantSchemaName
	if schema["$ref"] != wantRef {
		t.Fatalf("unexpected response %s schema ref: got %v want %q", status, schema["$ref"], wantRef)
	}
}

func requireOpenAPISheetRefSchema(t testing.TB, schema map[string]any) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("SheetRef must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, []string{"kind", "id"}) {
		t.Fatalf("unexpected SheetRef required fields: %v", required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	if len(properties) != 2 {
		t.Fatalf("SheetRef must expose exactly kind and id properties: %#v", properties)
	}
	kind := openAPIObjectAt(t, properties, "kind")
	if enum := toStrings(t, kind["enum"]); !equalStringSlices(enum, []string{"view_schema", "saved_view"}) {
		t.Fatalf("unexpected SheetRef kind enum: %v", enum)
	}
	id := openAPIObjectAt(t, properties, "id")
	if id["type"] != "string" || id["minLength"] != float64(1) {
		t.Fatalf("unexpected SheetRef id schema: %#v", id)
	}
}

func requireOpenAPIWorkbookPreferencesPutSchema(t testing.TB, schema map[string]any, field string) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("workbook preferences PUT request must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, []string{field}) {
		t.Fatalf("unexpected workbook preferences PUT required fields: %v", required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	if len(properties) != 1 {
		t.Fatalf("workbook preferences PUT request must expose exactly one property: %#v", properties)
	}
	fieldSchema := openAPIObjectAt(t, properties, field)
	oneOf, ok := fieldSchema["oneOf"].([]any)
	if !ok || len(oneOf) != 2 {
		t.Fatalf("workbook preferences PUT field must be SheetRef or null: %#v", fieldSchema)
	}
	sheetRef, ok := oneOf[0].(map[string]any)
	if !ok || sheetRef["$ref"] != "#/components/schemas/SheetRef" {
		t.Fatalf("workbook preferences PUT field must reference SheetRef first: %#v", oneOf[0])
	}
	nullRef, ok := oneOf[1].(map[string]any)
	if !ok || nullRef["type"] != "null" {
		t.Fatalf("workbook preferences PUT field must allow explicit null: %#v", oneOf[1])
	}
}

func requireOpenAPIEnvelopeSchema(t testing.TB, schema map[string]any, wantDataSchemaName string) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("workbook preferences envelope must be a closed object schema: %#v", schema)
	}
	if required := toStrings(t, schema["required"]); !equalStringSlices(required, []string{"data", "meta"}) {
		t.Fatalf("unexpected workbook preferences envelope required fields: %v", required)
	}
	properties := openAPIObjectAt(t, schema, "properties")
	data := openAPIObjectAt(t, properties, "data")
	if data["$ref"] != "#/components/schemas/"+wantDataSchemaName {
		t.Fatalf("unexpected workbook preferences envelope data ref: %#v", data)
	}
	meta := openAPIObjectAt(t, properties, "meta")
	if meta["$ref"] != "#/components/schemas/EnvelopeMeta" {
		t.Fatalf("unexpected workbook preferences envelope meta ref: %#v", meta)
	}
}

func timeRef(year int, month int, day int, hour int, minute int) time.Time {
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
}

type errorContract struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"http_status"`
}

type errorRegistry struct {
	Errors []errorContract `json:"errors"`
}

type extensionProfileContract struct {
	ProfileID     string   `json:"profile_id"`
	RouteFamilies []string `json:"route_families"`
}

type extensionRegistry struct {
	Profiles []extensionProfileContract `json:"profiles"`
}

func requireErrorContract(t testing.TB, code string, httpStatus int) {
	t.Helper()

	artifact, ok := contracts.ContractArtifactIndex["contracts/errors/index.json"]
	if !ok {
		t.Fatal("missing generated error contract registry")
	}

	var registry errorRegistry
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode generated error contract registry: %v", err)
	}

	for _, candidate := range registry.Errors {
		if candidate.Code == code {
			if candidate.HTTPStatus != httpStatus {
				t.Fatalf("unexpected contract status for %q: got %d want %d", code, candidate.HTTPStatus, httpStatus)
			}
			return
		}
	}
	t.Fatalf("missing generated error contract for %q", code)
}

func currentProfileExtensions(t testing.TB) []extensionProfileContract {
	t.Helper()

	artifact, ok := contracts.ContractArtifactIndex["contracts/extensions/index.json"]
	if !ok {
		t.Fatal("missing generated extension contract registry")
	}

	var registry extensionRegistry
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode generated extension contract registry: %v", err)
	}
	return append([]extensionProfileContract(nil), registry.Profiles...)
}
