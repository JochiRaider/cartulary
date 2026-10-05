package reference_data

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestRequestValidationNormalizationAndClosedRegistries_Unit(t *testing.T) {
	t.Run("semantic rejection projection", func(t *testing.T) {
		for _, tc := range []struct {
			err    error
			code   string
			status int
		}{
			{&RequestRejection{Field: "reason", Reason: "missing_required_field"}, "invalid_reference_pack_request", http.StatusBadRequest},
			{&ActivationRejection{Reason: "already_active"}, "reference_pack_activation_rejected", http.StatusConflict},
			{&OperationRejection{Reason: "stale_admission_state"}, "reference_pack_operation_rejected", http.StatusConflict},
		} {
			got := coordinatorAPIError(tc.err)
			if got.Code != tc.code || got.Status != tc.status {
				t.Fatalf("semantic projection: %#v", got)
			}
		}
	})
	t.Run("staging rejection classification", testStagingRejectionClassification)
	envelope := httpapi.UploadEnvelope{
		Metadata: map[string]json.RawMessage{
			"client_txn_id": json.RawMessage(`"txn-import"`),
		},
		FileSHA256Hex: "abc123",
	}
	request, apiErr := DecodeImportMetadata(envelope)
	if apiErr != nil {
		t.Fatalf("DecodeImportMetadata: %v", apiErr)
	}
	if request.ActivationPolicy != "staged_only" {
		t.Fatalf("default activation policy = %q", request.ActivationPolicy)
	}
	explicit := envelope
	explicit.Metadata = map[string]json.RawMessage{
		"client_txn_id":     json.RawMessage(`"txn-import"`),
		"activation_policy": json.RawMessage(`"staged_only"`),
	}
	explicitRequest, apiErr := DecodeImportMetadata(explicit)
	if apiErr != nil {
		t.Fatalf("explicit staged_only: %v", apiErr)
	}
	if !bytes.Equal(request.Normalized, explicitRequest.Normalized) {
		t.Fatalf("omitted and explicit staged_only must normalize equally: %s vs %s", request.Normalized, explicitRequest.Normalized)
	}
	auto := envelope
	auto.Metadata = map[string]json.RawMessage{
		"client_txn_id":     json.RawMessage(`"txn-import"`),
		"activation_policy": json.RawMessage(`"activate"`),
	}
	if _, apiErr := DecodeImportMetadata(auto); apiErr == nil || apiErr.Details["reason_code"] != "auto_activation_not_supported" {
		t.Fatalf("auto activation rejection = %#v", apiErr)
	}

	action, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action","reason":" \r\n "}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest: %v", apiErr)
	}
	nullAction, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action","reason":null}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest null: %v", apiErr)
	}
	omittedAction, apiErr := DecodeActionRequest(bytes.NewBufferString(`{"client_txn_id":"txn-action"}`))
	if apiErr != nil {
		t.Fatalf("DecodeActionRequest omitted: %v", apiErr)
	}
	if !bytes.Equal(action.Normalized, nullAction.Normalized) || !bytes.Equal(action.Normalized, omittedAction.Normalized) {
		t.Fatalf("empty, null, and omitted reason must normalize equally: %s %s %s", action.Normalized, nullAction.Normalized, omittedAction.Normalized)
	}
	t.Run("closed administrative JSON and reason boundaries", func(t *testing.T) {
		for _, body := range []string{`{"client_txn_id":"a","client_txn_id":"b"}`, `{"client_txn_id":"a","reason":"\ud800"}`, `[]`, `null`, `{} {}`} {
			if _, err := DecodeActionRequest(strings.NewReader(body)); err == nil {
				t.Fatalf("invalid JSON admitted: %s", body)
			}
		}
		for _, tc := range []struct {
			reason string
			code   string
		}{
			{"  " + strings.Repeat("e\u0301", 4096) + "  ", ""},
			{strings.Repeat("😀", 4096), ""}, {strings.Repeat("😀", 4097), "reason_too_long"}, {"a\x00b", "invalid_reason"},
		} {
			body, _ := json.Marshal(map[string]any{"client_txn_id": "a", "reason": tc.reason})
			_, err := DecodeActionRequest(bytes.NewReader(body))
			if (err == nil) != (tc.code == "") || err != nil && err.Details["reason_code"] != tc.code {
				t.Fatalf("reason boundary: %v", err)
			}
		}
		for _, tc := range []struct{ body, code string }{
			{`{"client_txn_id":"a"}`, "missing_required_field"}, {`{"client_txn_id":"a","reason":null}`, "field_not_nullable"}, {`{"client_txn_id":"a","reason":" "}`, "invalid_reason"}, {`{"client_txn_id":"a","reason":false}`, "invalid_reason"},
		} {
			if _, err := DecodeRemovalRequest(strings.NewReader(tc.body)); err == nil || err.Details["reason_code"] != tc.code {
				t.Fatalf("removal reason: %v", err)
			}
		}
		_, err := DecodeActionRequest(strings.NewReader(`{"client_txn_id":"a","hostile-secret":1}`))
		if err == nil || err.Details["field"] != "request" {
			t.Fatalf("unknown field leaked: %v", err)
		}
		body := `{"client_txn_id":"a"}`
		if _, err := DecodeActionRequest(strings.NewReader(body + strings.Repeat(" ", MaxAdministrativeRequestBytes-len(body)))); err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeActionRequest(strings.NewReader(body + strings.Repeat(" ", MaxAdministrativeRequestBytes-len(body)+1))); err == nil || err.Details["reason_code"] != "request_too_large" {
			t.Fatalf("request byte bound: %v", err)
		}
	})

	refresh, apiErr := DecodeRefreshRequest(bytes.NewBufferString(`{"client_txn_id":"txn-refresh","pack_keys":["z","a","z"]}`))
	if apiErr != nil {
		t.Fatalf("DecodeRefreshRequest: %v", apiErr)
	}
	if got := refresh.PackKeys; len(got) != 2 || got[0] != "a" || got[1] != "z" {
		t.Fatalf("pack_keys must coalesce and sort, got %#v", got)
	}
	if _, apiErr := DecodeRefreshRequest(bytes.NewBufferString(`{"client_txn_id":"txn-refresh","pack_keys":[]}`)); apiErr == nil || apiErr.Details["reason_code"] != "empty_pack_keys" {
		t.Fatalf("empty pack_keys rejection = %#v", apiErr)
	}

}

func TestSupportReferencePackListQueryUsesSharedListQueryScope(t *testing.T) {
	valid, apiErr := parseReferencePackListScope("search=host+registry&pack_version_state=staged&active=false&limit=50")
	if apiErr != nil {
		t.Fatalf("parse valid reference pack query: %v", apiErr)
	}
	if valid.Scope["search"] != "host registry" ||
		valid.Scope["pack_version_state"] != "staged" ||
		valid.Scope["active"] != "false" {
		t.Fatalf("unexpected reference pack scope: %#v", valid.Scope)
	}

	cases := []struct {
		name       string
		query      string
		code       string
		reasonCode string
	}{
		{
			name:       "unknown member",
			query:      "sort=pack_key",
			code:       "invalid_list_query",
			reasonCode: "unknown_query_member",
		},
		{
			name:       "duplicate member",
			query:      "active=true&active=false",
			code:       "invalid_list_query",
			reasonCode: "duplicate_query_member",
		},
		{
			name:       "invalid state",
			query:      "pack_version_state=available",
			code:       "invalid_list_query",
			reasonCode: "invalid_filter_value",
		},
		{
			name:       "pagination alias",
			query:      "page_size=10",
			code:       "invalid_pagination_request",
			reasonCode: "invalid_limit",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := parseReferencePackListScope(tc.query)
			if apiErr == nil || apiErr.Code != tc.code || apiErr.Details["reason_code"] != tc.reasonCode {
				t.Fatalf("unexpected error for %s: %#v", tc.name, apiErr)
			}
		})
	}
}

func TestSupportReferencePackListFilterAppliesSearchAndExactFilters(t *testing.T) {
	source := "https://offline.invalid/reference-packs/host"
	records := []AdministrativeVersion{
		{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "2026.04", Condition: "staged", Health: "staged", SourceIdentifier: &source},
		{PackKey: "type_registry.indicator", PackKind: "type_registry", PackVersion: "1", Condition: "verified_available", Health: "verified_available", Active: true},
	}
	scope, err := parseReferencePackListScope("search=type_registry.host+2026.04&pack_version_state=staged&active=false")
	if err != nil {
		t.Fatal(err)
	}
	filtered := filterAdministrativeVersions(records, scope.Scope)
	if len(filtered) != 1 || filtered[0].PackKey != "type_registry.host" {
		t.Fatalf("unexpected filtered records: %#v", filtered)
	}
}
