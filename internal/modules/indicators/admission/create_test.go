package admission

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestIndicatorCreateAdmissionAndHashCompatibility(t *testing.T) {
	t.Parallel()
	left, apiErr := DecodeCreateRequest(strings.NewReader(`{
        "client_txn_id":"txn-indicator",
        "indicator.indicator_type":" ipv4_addr ",
        "indicator.value_kind":"atomic",
        "indicator.display_value":"203[.]0[.]113[.]7"
    }`))
	if apiErr != nil {
		t.Fatalf("decode indicator create: %v", apiErr)
	}
	if left.ClientTxnID != "txn-indicator" || left.IndicatorType != "ipv4_addr" || left.DisplayValue != "203[.]0[.]113[.]7" {
		t.Fatalf("normalized command = %#v", left)
	}
	right, apiErr := DecodeCreateRequest(strings.NewReader(`{"indicator.display_value":"203[.]0[.]113[.]7","indicator.value_kind":"atomic","client_txn_id":"txn-indicator","indicator.indicator_type":"ipv4_addr"}`))
	if apiErr != nil {
		t.Fatalf("decode reordered indicator create: %v", apiErr)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("reordered wire members changed admitted command: left=%#v right=%#v", left, right)
	}
}

func TestIndicatorCreateAdmissionRejectsBeforeOwnerExecution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{name: "empty", body: ``, reasonCode: "request_not_object"},
		{name: "malformed", body: `{`, reasonCode: "request_not_object"},
		{name: "not object", body: `[1]`, reasonCode: "request_not_object"},
		{name: "null", body: `null`, reasonCode: "request_not_object"},
		{name: "scalar", body: `"value"`, reasonCode: "request_not_object"},
		{name: "duplicate member", body: `{"client_txn_id":"one","client_txn_id":"two"}`, reasonCode: "request_not_object"},
		{name: "duplicate nested member", body: `{"client_txn_id":"txn","unknown":{"value":1,"value":2}}`, reasonCode: "request_not_object"},
		{name: "trailing object", body: `{} {}`, reasonCode: "request_not_object"},
		{name: "missing transaction", body: `{}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "null transaction", body: `{"client_txn_id":null}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "numeric transaction", body: `{"client_txn_id":7}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "blank transaction", body: `{"client_txn_id":"  "}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "unknown field", body: `{"client_txn_id":"txn","unknown":"value"}`, field: "unknown", reasonCode: "unknown_field"},
		{name: "readonly field", body: `{"client_txn_id":"txn","indicator.observation_count":1}`, field: "indicator.observation_count", reasonCode: "unknown_field"},
		{name: "null required value", body: `{"client_txn_id":"txn","indicator.indicator_type":null}`, field: "indicator.indicator_type", reasonCode: "field_not_nullable"},
		{name: "nonstring value", body: `{"client_txn_id":"txn","indicator.indicator_type":7}`, field: "indicator.indicator_type", reasonCode: "invalid_value"},
		{name: "blank value", body: `{"client_txn_id":"txn","indicator.indicator_type":"  "}`, field: "indicator.indicator_type", reasonCode: "invalid_value"},
		{name: "missing identity", body: `{"client_txn_id":"txn"}`, field: "indicator.indicator_type", reasonCode: "missing_required_field"},
		{name: "invalid type alias", body: `{"client_txn_id":"txn","indicator.indicator_type":"domain","indicator.value_kind":"atomic","indicator.display_value":"example.test"}`, field: "indicator.indicator_type", reasonCode: "invalid_value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, validation := DecodeCreateRequest(strings.NewReader(test.body))
			if validation == nil {
				t.Fatal("validation error is nil")
			}
			if validation.Field != test.field {
				t.Fatalf("field = %q, want %q", validation.Field, test.field)
			}
			if validation.ReasonCode != test.reasonCode {
				t.Fatalf("reason = %q, want %q", validation.ReasonCode, test.reasonCode)
			}
		})
	}
	t.Run("transport imports", func(t *testing.T) {
		body, err := os.ReadFile("create.go")
		if err != nil {
			t.Fatalf("read admission source: %v", err)
		}
		for _, forbidden := range []string{
			"internal/platform/httpapi",
			"DecodeStrictJSONObject",
			"APIError",
		} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("Indicator admission retains transport dependency %q", forbidden)
			}
		}
	})
}
