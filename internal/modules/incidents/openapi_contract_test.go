package incidents

import (
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestIncidentOpenAPIResponsesAndSecurityMatchRuntime_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObjectAt(t, document, "paths")
	components := openAPIObjectAt(t, document, "components")

	expected := map[string]incidentOpenAPIExpectation{
		"closeIncident": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_incident_lifecycle_request"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found"},
				"409": {"client_txn_conflict", "credential_bootstrap_rejected", "incident_version_conflict", "illegal_transition"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentEnvelope"},
		},
		"createIncident": {
			statuses: map[string][]string{
				"200": nil,
				"201": nil,
				"400": {"invalid_incident_create"},
				"401": {"session_required"},
				"403": {"csrf_verification_failed"},
				"409": {"client_txn_conflict", "credential_bootstrap_rejected", "incident_key_conflict"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentEnvelope", "201": "IncidentEnvelope"},
		},
		"createIncidentMembership": {
			statuses: map[string][]string{
				"200": nil,
				"201": nil,
				"400": {"invalid_mutation_payload"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found", "user_not_found"},
				"409": {"client_txn_conflict", "credential_bootstrap_rejected", "membership_exists_use_patch", "user_inactive"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentMembershipEnvelope", "201": "IncidentMembershipEnvelope"},
		},
		"deleteIncidentMembership": {
			statuses: map[string][]string{
				"204": nil,
				"400": {"invalid_mutation_payload"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found", "membership_not_found"},
				"409": {"credential_bootstrap_rejected", "last_incident_admin", "membership_version_conflict"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"204": ""},
		},
		"getIncident": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_list_query"},
				"401": {"session_required"},
				"404": {"incident_not_found"},
				"409": {"credential_bootstrap_rejected"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentEnvelope"},
		},
		"listIncidentMembershipAuditEvents": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_list_query", "invalid_pagination_request"},
				"401": {"session_required"},
				"403": {"authorization_denied"},
				"404": {"incident_not_found"},
				"409": {"credential_bootstrap_rejected"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "AdministrativeAuditEnvelope"},
		},
		"listIncidentMemberships": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_pagination_request"},
				"401": {"session_required"},
				"404": {"incident_not_found"},
				"409": {"credential_bootstrap_rejected"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentMembershipListEnvelope"},
		},
		"listVisibleIncidents": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_list_query", "invalid_pagination_request"},
				"401": {"session_required"},
				"409": {"credential_bootstrap_rejected"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentListEnvelope"},
		},
		"patchIncident": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_incident_patch"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found"},
				"409": {"credential_bootstrap_rejected", "incident_closed", "incident_version_conflict"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentEnvelope"},
		},
		"patchIncidentMembership": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_mutation_payload"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found", "membership_not_found"},
				"409": {"credential_bootstrap_rejected", "last_incident_admin", "membership_version_conflict"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentMembershipEnvelope"},
		},
		"reopenIncident": {
			statuses: map[string][]string{
				"200": nil,
				"400": {"invalid_incident_lifecycle_request"},
				"401": {"session_required"},
				"403": {"authorization_denied", "csrf_verification_failed"},
				"404": {"incident_not_found"},
				"409": {"client_txn_conflict", "credential_bootstrap_rejected", "incident_version_conflict", "illegal_transition"},
				"500": {"internal_error"},
			},
			successSchemas: map[string]string{"200": "IncidentEnvelope"},
		},
	}

	descriptors := httpapi.ContractOperationsForOwner("module.incidents")
	if len(descriptors) != len(expected) {
		t.Fatalf("incident descriptor inventory changed: got %d want %d", len(descriptors), len(expected))
	}
	seen := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		expectation, ok := expected[descriptor.OperationID]
		if !ok {
			t.Fatalf("runtime descriptor %s has no OpenAPI expectation", descriptor.OperationID)
		}
		operation := openAPIObjectAt(
			t,
			openAPIObjectAt(t, paths, descriptor.PathTemplate),
			incidentOpenAPIMethodKey(descriptor.Method),
		)
		if operation["operationId"] != descriptor.OperationID {
			t.Fatalf("%s operationId mismatch: %#v", descriptor.OperationID, operation["operationId"])
		}
		if got := incidentOpenAPISecurity(operation["security"]); !reflect.DeepEqual(got, descriptor.Security) {
			t.Fatalf("%s security mismatch: got %#v want %#v", descriptor.OperationID, got, descriptor.Security)
		}

		responses := openAPIObjectAt(t, operation, "responses")
		gotStatuses := make([]string, 0, len(responses))
		for status := range responses {
			gotStatuses = append(gotStatuses, status)
		}
		slices.Sort(gotStatuses)
		wantStatuses := make([]string, 0, len(expectation.statuses))
		for status := range expectation.statuses {
			wantStatuses = append(wantStatuses, status)
		}
		slices.Sort(wantStatuses)
		if !slices.Equal(gotStatuses, wantStatuses) {
			t.Fatalf("%s response statuses changed: got %v want %v", descriptor.OperationID, gotStatuses, wantStatuses)
		}
		for _, successStatus := range descriptor.SuccessStatuses {
			if _, ok := responses[http.StatusText(successStatus)]; ok {
				t.Fatalf("%s used a status phrase instead of a numeric OpenAPI response key", descriptor.OperationID)
			}
		}

		for status, wantCodes := range expectation.statuses {
			response := incidentOpenAPIResolvedResponse(t, components, responses[status])
			if wantSchema, success := expectation.successSchemas[status]; success {
				incidentOpenAPIRequireSuccessSchema(t, descriptor.OperationID, status, response, wantSchema)
				continue
			}
			if gotCodes := toStrings(t, response["x-cartulary-error-codes"]); !slices.Equal(gotCodes, wantCodes) {
				t.Fatalf("%s %s error codes got %v want %v", descriptor.OperationID, status, gotCodes, wantCodes)
			}
			media := openAPIObjectAt(t, openAPIObjectAt(t, response, "content"), "application/json")
			if openAPIObjectAt(t, media, "schema")["$ref"] != "#/components/schemas/ErrorEnvelope" {
				t.Fatalf("%s %s error schema is not ErrorEnvelope", descriptor.OperationID, status)
			}
		}
		seen[descriptor.OperationID] = struct{}{}
	}
	for operationID := range expected {
		if _, ok := seen[operationID]; !ok {
			t.Fatalf("expected incident operation %s has no runtime descriptor", operationID)
		}
	}

	incidentOpenAPIRequireListChain(t, components, "IncidentListEnvelope", "IncidentListData", "incidents", "IncidentResource")
	incidentOpenAPIRequireListChain(t, components, "IncidentMembershipListEnvelope", "IncidentMembershipListData", "memberships", "IncidentMembershipResource")
	incidentOpenAPIRequireResourceEnvelope(t, components, "IncidentMembershipEnvelope", "IncidentMembershipResource")
}

type incidentOpenAPIExpectation struct {
	statuses       map[string][]string
	successSchemas map[string]string
}

func incidentOpenAPISecurity(raw any) [][]string {
	alternatives, _ := raw.([]any)
	result := make([][]string, 0, len(alternatives))
	for _, alternative := range alternatives {
		requirement, _ := alternative.(map[string]any)
		schemes := make([]string, 0, len(requirement))
		for scheme := range requirement {
			schemes = append(schemes, scheme)
		}
		slices.Sort(schemes)
		result = append(result, schemes)
	}
	slices.SortFunc(result, func(left, right []string) int {
		return strings.Compare(strings.Join(left, ","), strings.Join(right, ","))
	})
	return result
}

func incidentOpenAPIResolvedResponse(t testing.TB, components map[string]any, raw any) map[string]any {
	t.Helper()
	response := raw.(map[string]any)
	ref, referenced := response["$ref"].(string)
	if !referenced {
		return response
	}
	const prefix = "#/components/responses/"
	if !strings.HasPrefix(ref, prefix) {
		t.Fatalf("unexpected response reference %q", ref)
	}
	return openAPIObjectAt(t, openAPIObjectAt(t, components, "responses"), strings.TrimPrefix(ref, prefix))
}

func incidentOpenAPIRequireSuccessSchema(t testing.TB, operationID, status string, response map[string]any, wantSchema string) {
	t.Helper()
	if wantSchema == "" {
		if _, present := response["content"]; present {
			t.Fatalf("%s %s no-content response declares content", operationID, status)
		}
		return
	}
	media := openAPIObjectAt(t, openAPIObjectAt(t, response, "content"), "application/json")
	if want := "#/components/schemas/" + wantSchema; openAPIObjectAt(t, media, "schema")["$ref"] != want {
		t.Fatalf("%s %s success schema mismatch: got %#v want %q", operationID, status, media, want)
	}
}

func incidentOpenAPIRequireResourceEnvelope(t testing.TB, components map[string]any, envelopeName, resourceName string) {
	t.Helper()
	schemas := openAPIObjectAt(t, components, "schemas")
	envelope := openAPIObjectAt(t, schemas, envelopeName)
	data := openAPIObjectAt(t, openAPIObjectAt(t, envelope, "properties"), "data")
	if want := "#/components/schemas/" + resourceName; data["$ref"] != want {
		t.Fatalf("%s data reference mismatch: got %#v want %q", envelopeName, data["$ref"], want)
	}
}

func incidentOpenAPIRequireListChain(
	t testing.TB,
	components map[string]any,
	envelopeName string,
	dataName string,
	memberName string,
	resourceName string,
) {
	t.Helper()
	incidentOpenAPIRequireResourceEnvelope(t, components, envelopeName, dataName)
	schemas := openAPIObjectAt(t, components, "schemas")
	data := openAPIObjectAt(t, schemas, dataName)
	member := openAPIObjectAt(t, openAPIObjectAt(t, data, "properties"), memberName)
	items := openAPIObjectAt(t, member, "items")
	if want := "#/components/schemas/" + resourceName; items["$ref"] != want {
		t.Fatalf("%s.%s item reference mismatch: got %#v want %q", dataName, memberName, items["$ref"], want)
	}
}

func incidentOpenAPIMethodKey(method string) string {
	return strings.ToLower(method)
}
