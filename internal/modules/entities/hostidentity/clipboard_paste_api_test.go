package hostidentity

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
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

	request, apiErr := DecodeClipboardPasteRequest(body, entitycontract.HostsViewSchemaID)
	if apiErr != nil {
		t.Fatalf("decode clipboard paste request: %#v", apiErr)
	}
	plan, err := BuildClipboardPastePlan(request)
	if err != nil {
		t.Fatalf("build clipboard paste plan: %v", err)
	}
	if plan.ViewSchemaID != entitycontract.HostsViewSchemaID || plan.ClientTxnID != "txn-host-paste" || len(plan.Rows) != 2 {
		t.Fatalf("unexpected plan identity: %#v", plan)
	}
	if len(plan.Rows[0].Cells) != 2 || plan.Rows[0].Cells[0].FieldKey != "host.display_name" || plan.Rows[0].Cells[1].FieldKey != "host.hostname" {
		t.Fatalf("unexpected row cells: %#v", plan.Rows[0].Cells)
	}
	expectedHash := entityClipboardPasteRequestHash(entitycontract.HostsViewSchemaID, "txn-host-paste", "Gateway One\tgateway-one\nGateway Two\tgateway-two", "tsv", "host.display_name", []string{"host.display_name", "host.hostname"})
	if !bytes.Equal(request.RequestHash(), expectedHash) {
		t.Fatalf("request hash changed")
	}
	const hostRequestHashGolden = "ad66b01d86e8daa70fc8a032ff28498e37e6f83f54245acfaa2adb714e19bcfb"
	if got := hex.EncodeToString(request.RequestHash()); got != hostRequestHashGolden {
		t.Fatalf("Host clipboard request hash changed: %s", got)
	}

	identityRequest, failure := DecodeClipboardPasteRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.identities.v1",
		"client_txn_id":"txn-identity-paste",
		"clipboard_text":"Analyst One\tanalyst@example.test",
		"format":"tsv",
		"start_field_key":"identity.display_name",
		"columns":["identity.display_name","identity.email"],
		"targets":[{"kind":"create"}]
	}`), entitycontract.IdentitiesViewSchemaID)
	if failure != nil {
		t.Fatalf("decode Identity clipboard paste: %#v", failure)
	}
	identityPlan, err := BuildClipboardPastePlan(identityRequest)
	if err != nil || len(identityPlan.Rows) != 1 ||
		identityPlan.Rows[0].Cells[0].EntityBindingMode == nil ||
		*identityPlan.Rows[0].Cells[0].EntityBindingMode != "entity_origin" {
		t.Fatalf("Identity clipboard plan changed: plan=%#v error=%v", identityPlan, err)
	}
	const identityRequestHashGolden = "e3670f89347776b7aafa179b8dcda94afa499bca62a75654729191ff96d6a7a0"
	if got := hex.EncodeToString(identityRequest.RequestHash()); got != identityRequestHashGolden {
		t.Fatalf("Identity clipboard request hash changed: %s", got)
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
			_, failure := DecodeClipboardPasteRequest(strings.NewReader(tc.body), entitycontract.HostsViewSchemaID)
			if failure == nil {
				t.Fatalf("expected admission failure")
			}
			failureField, _ := failure.Field()
			if failureField != tc.field || string(failure.ReasonCode()) != tc.reasonCode {
				t.Fatalf("unexpected failure: %#v", failure)
			}
		})
	}
}
