package hostidentity

import (
	"bytes"
	"strings"
	"testing"
)

func TestClipboardPasteRequestDecodePlanAndHash(t *testing.T) {
	body := strings.NewReader(`{
		"view_schema_id":"cartulary.view.hosts.v1",
		"client_txn_id":"txn-host-paste",
		"clipboard_text":"Gateway One\tgateway-one\nGateway Two\tgateway-two",
		"format":"tsv",
		"start_field_key":"host.display_name",
		"columns":["host.display_name","host.hostname"],
		"targets":[{"kind":"create"},{"kind":"create"}]
	}`)

	request, apiErr := DecodeClipboardPasteRequest(body, HostsViewSchemaID)
	if apiErr != nil {
		t.Fatalf("decode clipboard paste request: %#v", apiErr)
	}
	plan, err := BuildClipboardPastePlan(request)
	if err != nil {
		t.Fatalf("build clipboard paste plan: %v", err)
	}
	if plan.ViewSchemaID != HostsViewSchemaID || plan.ClientTxnID != "txn-host-paste" || len(plan.Rows) != 2 {
		t.Fatalf("unexpected plan identity: %#v", plan)
	}
	if len(plan.Rows[0].Cells) != 2 || plan.Rows[0].Cells[0].FieldKey != "host.display_name" || plan.Rows[0].Cells[1].FieldKey != "host.hostname" {
		t.Fatalf("unexpected row cells: %#v", plan.Rows[0].Cells)
	}
	expectedHash := EntityClipboardPasteRequestHash(HostsViewSchemaID, "txn-host-paste", "Gateway One\tgateway-one\nGateway Two\tgateway-two", "tsv", "host.display_name", []string{"host.display_name", "host.hostname"})
	if !bytes.Equal(request.RequestHash(), expectedHash) {
		t.Fatalf("request hash changed")
	}
}

func TestClipboardPasteRequestDecodeRejectsWorkbookInvalidPayloads(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{
			name: "unknown field",
			body: `{
				"view_schema_id":"cartulary.view.hosts.v1",
				"client_txn_id":"txn-host-paste",
				"clipboard_text":"Gateway One",
				"start_field_key":"host.display_name",
				"columns":["host.display_name"],
				"unexpected":true
			}`,
			field:      "unexpected",
			reasonCode: "unknown_field",
		},
		{
			name: "unsupported target",
			body: `{
				"view_schema_id":"cartulary.view.hosts.v1",
				"client_txn_id":"txn-host-paste",
				"clipboard_text":"Gateway One",
				"start_field_key":"host.display_name",
				"columns":["host.display_name"],
				"targets":[{"kind":"update"}]
			}`,
			field:      "targets",
			reasonCode: "invalid_value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeClipboardPasteRequest(strings.NewReader(tc.body), HostsViewSchemaID)
			if apiErr == nil {
				t.Fatalf("expected api error")
			}
			if apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("unexpected code: %s", apiErr.Code)
			}
			if apiErr.Details["field"] != tc.field || apiErr.Details["reason_code"] != tc.reasonCode {
				t.Fatalf("unexpected details: %#v", apiErr.Details)
			}
		})
	}
}
