package auth

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestSupportPhase11_EnterpriseAuthErrorRegistriesClosed(t *testing.T) {
	errorsDoc, err := os.ReadFile("../../../contracts/errors/index.json")
	if err != nil {
		t.Fatalf("read errors registry: %v", err)
	}
	var registry struct {
		ReasonRegistries []struct {
			ErrorCode   string `json:"error_code"`
			ReasonCodes []struct {
				Code string `json:"code"`
			} `json:"reason_codes"`
		} `json:"reason_registries"`
	}
	if err := json.Unmarshal(errorsDoc, &registry); err != nil {
		t.Fatalf("decode errors registry: %v", err)
	}

	reasonsByError := map[string][]string{}
	for _, entry := range registry.ReasonRegistries {
		for _, reason := range entry.ReasonCodes {
			reasonsByError[entry.ErrorCode] = append(reasonsByError[entry.ErrorCode], reason.Code)
		}
		slices.Sort(reasonsByError[entry.ErrorCode])
	}

	want := map[string][]string{
		"invalid_enterprise_auth_request": {
			"field_not_nullable",
			"request_not_object",
			"return_to_not_allowed",
			"unknown_field",
		},
		"enterprise_auth_transaction_rejected": {
			"already_used",
			"browser_binding_mismatch",
			"completion_mismatch",
			"expired",
			"not_found",
			"provider_mismatch",
		},
		"provider_response_rejected": {
			"assertion_expired",
			"audience_mismatch",
			"code_exchange_failed",
			"issuer_mismatch",
			"missing_required_field",
			"nonce_mismatch",
			"relay_state_mismatch",
			"signature_invalid",
			"state_mismatch",
		},
		"provider_identity_rejected": {
			"ambiguous_link",
			"inactive_user",
			"no_linked_user",
			"subject_missing",
		},
		"auth_binding_conflict": {
			"binding_not_active",
			"provider_already_linked_for_user",
			"provider_subject_in_use",
		},
	}
	for errorCode, wantReasons := range want {
		slices.Sort(wantReasons)
		if got := reasonsByError[errorCode]; !slices.Equal(got, wantReasons) {
			t.Fatalf("%s reason registry mismatch: got %#v want %#v", errorCode, got, wantReasons)
		}
	}
}
