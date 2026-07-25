package incidents

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestIncidentLifecycleRequestValidationUsesExactErrorFamilyAndReasons_Unit(t *testing.T) {
	tooLongReason := strings.Repeat("x", reasonNoteMaxRunes+1)
	cases := []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{name: "non-object", body: `[]`, reasonCode: "request_not_object"},
		{name: "duplicate member", body: `{"base_incident_version":1,"base_incident_version":2,"client_txn_id":"txn","reason":"reason"}`, reasonCode: "request_not_object"},
		{name: "missing version", body: `{"client_txn_id":"txn","reason":"reason"}`, field: "base_incident_version", reasonCode: "missing_required_field"},
		{name: "null version", body: `{"base_incident_version":null,"client_txn_id":"txn","reason":"reason"}`, field: "base_incident_version", reasonCode: "field_not_nullable"},
		{name: "invalid version", body: `{"base_incident_version":0,"client_txn_id":"txn","reason":"reason"}`, field: "base_incident_version", reasonCode: "invalid_base_incident_version"},
		{name: "missing transaction", body: `{"base_incident_version":1,"reason":"reason"}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "null transaction", body: `{"base_incident_version":1,"client_txn_id":null,"reason":"reason"}`, field: "client_txn_id", reasonCode: "field_not_nullable"},
		{name: "invalid transaction", body: `{"base_incident_version":1,"client_txn_id":"  ","reason":"reason"}`, field: "client_txn_id", reasonCode: "invalid_client_txn_id"},
		{name: "missing reason", body: `{"base_incident_version":1,"client_txn_id":"txn"}`, field: "reason", reasonCode: "missing_required_field"},
		{name: "null reason", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":null}`, field: "reason", reasonCode: "field_not_nullable"},
		{name: "invalid reason type", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":false}`, field: "reason", reasonCode: "invalid_reason"},
		{name: "empty reason", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":" \r\n\t "}`, field: "reason", reasonCode: "reason_empty_after_normalization"},
		{name: "long reason", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":"` + tooLongReason + `"}`, field: "reason", reasonCode: "reason_too_long"},
		{name: "control character", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":"unsafe\u0001reason"}`, field: "reason", reasonCode: "control_character_not_allowed"},
		{name: "unknown member", body: `{"base_incident_version":1,"client_txn_id":"txn","reason":"reason","archive":true}`, field: "archive", reasonCode: "unknown_field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeIncidentLifecycleRequest(strings.NewReader(tc.body))
			if apiErr == nil {
				t.Fatal("expected lifecycle request error")
			}
			if apiErr.Code != "invalid_incident_lifecycle_request" {
				t.Fatalf("error code got %q", apiErr.Code)
			}
			if apiErr.Details["reason_code"] != tc.reasonCode {
				t.Fatalf("reason_code got %#v want %q", apiErr.Details["reason_code"], tc.reasonCode)
			}
			if tc.field == "" {
				if _, present := apiErr.Details["field"]; present {
					t.Fatalf("unexpected field detail: %#v", apiErr.Details)
				}
			} else if apiErr.Details["field"] != tc.field {
				t.Fatalf("field got %#v want %q", apiErr.Details["field"], tc.field)
			}
		})
	}

	request, apiErr := DecodeIncidentLifecycleRequest(strings.NewReader(
		`{"base_incident_version":7,"client_txn_id":"txn-lifecycle","reason":"  cafe\u0301\r\nline\t "}`,
	))
	if apiErr != nil {
		t.Fatalf("decode valid lifecycle request: %v", apiErr)
	}
	if request.BaseIncidentVersion != 7 || request.ClientTxnID != "txn-lifecycle" || request.Reason != "café\nline" {
		t.Fatalf("unexpected normalized lifecycle request: %#v", request)
	}
	closeHash := IncidentLifecycleRequestHash("close", request)
	reopenHash := IncidentLifecycleRequestHash("reopen", request)
	if reflect.DeepEqual(closeHash, reopenHash) {
		t.Fatal("action route must participate in lifecycle idempotency comparison")
	}
	otherKey := request
	otherKey.ClientTxnID = "txn-other-key"
	if !reflect.DeepEqual(closeHash, IncidentLifecycleRequestHash("close", otherKey)) {
		t.Fatal("client_txn_id belongs to the idempotency key, not the normalized comparison payload")
	}
}

func TestOpenAPIIncidentLifecycleOperationsAreCompleteAndExact_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	paths := openAPIObjectAt(t, document, "paths")
	schemas := openAPIObjectAt(t, openAPIObjectAt(t, document, "components"), "schemas")

	requestSchema := openAPIObjectAt(t, schemas, "IncidentLifecycleRequest")
	if requestSchema["type"] != "object" || requestSchema["additionalProperties"] != false {
		t.Fatalf("lifecycle request must be closed: %#v", requestSchema)
	}
	if got := toStrings(t, requestSchema["required"]); !equalStringSlices(got, []string{"base_incident_version", "client_txn_id", "reason"}) {
		t.Fatalf("unexpected lifecycle request required members: %v", got)
	}
	requestProperties := openAPIObjectAt(t, requestSchema, "properties")
	if len(requestProperties) != 3 {
		t.Fatalf("lifecycle request properties must be exact: %#v", requestProperties)
	}

	resource := openAPIObjectAt(t, schemas, "IncidentResource")
	if resource["type"] != "object" || resource["additionalProperties"] != false {
		t.Fatalf("incident resource must be closed: %#v", resource)
	}
	requiredResourceMembers := []string{
		"incident_id",
		"incident_key",
		"title",
		"description",
		"status",
		"severity",
		"tlp",
		"current_phase",
		"primary_external_case_ref",
		"created_by_user_id",
		"created_at",
		"updated_at",
		"updated_by_user_id",
		"incident_version",
		"closed_at",
	}
	if got := toStrings(t, resource["required"]); !equalStringSlices(got, requiredResourceMembers) {
		t.Fatalf("unexpected incident resource members: %v", got)
	}
	if properties := openAPIObjectAt(t, resource, "properties"); len(properties) != len(requiredResourceMembers) {
		t.Fatalf("incident resource properties must be exact: %#v", properties)
	}
	envelope := openAPIObjectAt(t, schemas, "IncidentEnvelope")
	if openAPIObjectAt(t, openAPIObjectAt(t, envelope, "properties"), "data")["$ref"] != "#/components/schemas/IncidentResource" {
		t.Fatalf("incident envelope must carry IncidentResource: %#v", envelope)
	}

	wantSecurity := []any{
		map[string]any{"sessionCookie": []any{}, "csrfCookie": []any{}, "csrfHeader": []any{}},
		map[string]any{"bearerSession": []any{}},
	}
	wantResponses := []string{"200", "400", "401", "403", "404", "409", "500"}
	wantErrorCodes := map[string][]string{
		"400": {"invalid_incident_lifecycle_request"},
		"401": {"session_required"},
		"403": {"authorization_denied", "csrf_verification_failed"},
		"404": {"incident_not_found"},
		"409": {"client_txn_conflict", "credential_bootstrap_rejected", "incident_version_conflict", "illegal_transition"},
		"500": {"internal_error"},
	}
	for _, action := range []string{"close", "reopen"} {
		operation := openAPIObjectAt(
			t,
			openAPIObjectAt(t, paths, "/api/v1/incidents/{incident_id}/"+action),
			"post",
		)
		if operation["operationId"] != action+"Incident" {
			t.Fatalf("%s operationId got %#v", action, operation["operationId"])
		}
		if !reflect.DeepEqual(operation["security"], wantSecurity) {
			t.Fatalf("%s security got %#v", action, operation["security"])
		}
		requestBody := openAPIObjectAt(t, operation, "requestBody")
		if requestBody["required"] != true {
			t.Fatalf("%s requestBody must be required", action)
		}
		requestMedia := openAPIObjectAt(t, openAPIObjectAt(t, requestBody, "content"), "application/json")
		if openAPIObjectAt(t, requestMedia, "schema")["$ref"] != "#/components/schemas/IncidentLifecycleRequest" {
			t.Fatalf("%s request schema is not IncidentLifecycleRequest", action)
		}
		responses := openAPIObjectAt(t, operation, "responses")
		responseKeys := make([]string, 0, len(responses))
		for status := range responses {
			responseKeys = append(responseKeys, status)
		}
		sort.Strings(responseKeys)
		if !equalStringSlices(responseKeys, wantResponses) {
			t.Fatalf("%s response statuses got %v", action, responseKeys)
		}
		success := openAPIObjectAt(t, responses, "200")
		successMedia := openAPIObjectAt(t, openAPIObjectAt(t, success, "content"), "application/json")
		if openAPIObjectAt(t, successMedia, "schema")["$ref"] != "#/components/schemas/IncidentEnvelope" {
			t.Fatalf("%s success schema is not IncidentEnvelope", action)
		}
		for status, codes := range wantErrorCodes {
			response := openAPIObjectAt(t, responses, status)
			if got := toStrings(t, response["x-cartulary-error-codes"]); !equalStringSlices(got, codes) {
				t.Fatalf("%s %s error codes got %v want %v", action, status, got, codes)
			}
			media := openAPIObjectAt(t, openAPIObjectAt(t, response, "content"), "application/json")
			if openAPIObjectAt(t, media, "schema")["$ref"] != "#/components/schemas/ErrorEnvelope" {
				t.Fatalf("%s %s error schema is not ErrorEnvelope", action, status)
			}
		}
	}

	contracttest.RequireErrorContract(t, "invalid_incident_lifecycle_request", 400)
	contracttest.RequireErrorContract(t, "incident_version_conflict", 409)
	contracttest.RequireErrorContract(t, "illegal_transition", 409)
}
