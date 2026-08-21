package admission

import (
	"bytes"
	"encoding/hex"
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
	leftHash := CreateRequestHash(left)
	if !bytes.Equal(leftHash, CreateRequestHash(right)) {
		t.Fatal("member order changed Indicator request hash")
	}
	wantHash, err := hex.DecodeString("49dd4b43356f985be78b671d6b57cfe912dcfc2782573acc9fb6c2cda8b5e6a6")
	if err != nil {
		t.Fatalf("decode expected hash: %v", err)
	}
	if !bytes.Equal(leftHash, wantHash) {
		t.Fatalf("request hash = %x, want %x", leftHash, wantHash)
	}

	differentTxn := right
	differentTxn.ClientTxnID = "txn-indicator-other"
	if bytes.Equal(leftHash, CreateRequestHash(differentTxn)) {
		t.Fatal("client transaction ID must remain part of the Indicator replay hash")
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
		{name: "not object", body: `[1]`, reasonCode: "request_not_object"},
		{name: "missing transaction", body: `{}`, field: "client_txn_id", reasonCode: "missing_required_field"},
		{name: "unknown field", body: `{"client_txn_id":"txn","unknown":"value"}`, field: "unknown", reasonCode: "unknown_field"},
		{name: "null required value", body: `{"client_txn_id":"txn","indicator.indicator_type":null}`, field: "indicator.indicator_type", reasonCode: "field_not_nullable"},
		{name: "missing identity", body: `{"client_txn_id":"txn"}`, field: "indicator.indicator_type", reasonCode: "missing_required_field"},
		{name: "invalid type alias", body: `{"client_txn_id":"txn","indicator.indicator_type":"domain","indicator.value_kind":"atomic","indicator.display_value":"example.test"}`, field: "indicator.indicator_type", reasonCode: "invalid_value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, apiErr := DecodeCreateRequest(strings.NewReader(test.body))
			if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("API error = %#v", apiErr)
			}
			if got, _ := apiErr.Details["field"].(string); got != test.field {
				t.Fatalf("field = %q, want %q", got, test.field)
			}
			if got, _ := apiErr.Details["reason_code"].(string); got != test.reasonCode {
				t.Fatalf("reason = %q, want %q", got, test.reasonCode)
			}
		})
	}
}
