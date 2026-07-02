package startup_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestSupportPhase2_WorkbookPreferencesPutDecodersCanonicalizeSheetRefs(t *testing.T) {
	userRequest, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(strings.NewReader(`{
		"home_sheet_ref":{"id":"cartulary.view.timeline.v2","kind":"view_schema"}
	}`))
	if apiErr != nil {
		t.Fatalf("decode user workbook preferences PUT: %v", apiErr)
	}
	if string(userRequest.HomeSheetRef) != `{"kind":"view_schema","id":"cartulary.view.timeline.v2"}` {
		t.Fatalf("unexpected canonical user sheet_ref: %s", userRequest.HomeSheetRef)
	}

	savedViewRequest, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(strings.NewReader(`{
		"home_sheet_ref":{"id":"00000000-0000-0000-0000-000000000153","kind":"saved_view"}
	}`))
	if apiErr != nil {
		t.Fatalf("decode saved-view workbook preferences PUT: %v", apiErr)
	}
	if string(savedViewRequest.HomeSheetRef) != `{"kind":"saved_view","id":"00000000-0000-0000-0000-000000000153"}` {
		t.Fatalf("unexpected canonical saved-view sheet_ref: %s", savedViewRequest.HomeSheetRef)
	}

	defaultRequest, apiErr := workbookstartup.DecodeDefaultPreferencesPutRequest(strings.NewReader(`{
		"default_sheet_ref":null
	}`))
	if apiErr != nil {
		t.Fatalf("decode default workbook preferences clear: %v", apiErr)
	}
	if defaultRequest.DefaultSheetRef != nil {
		t.Fatalf("expected null default sheet_ref to clear preference, got %s", defaultRequest.DefaultSheetRef)
	}
}

func TestSupportPhase2_WorkbookPreferencesPutDecodersRejectInvalidPayloads(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{
			name:       "missing required field",
			body:       `{}`,
			field:      "home_sheet_ref",
			reasonCode: "missing_required_field",
		},
		{
			name:       "unknown top-level field",
			body:       `{"home_sheet_ref":null,"unexpected":true}`,
			field:      "unexpected",
			reasonCode: "unknown_field",
		},
		{
			name:       "invalid kind",
			body:       `{"home_sheet_ref":{"kind":"workspace","id":"cartulary.view.timeline.v2"}}`,
			field:      "home_sheet_ref.kind",
			reasonCode: "unsupported_sheet_ref_kind",
		},
		{
			name:       "unknown view schema",
			body:       `{"home_sheet_ref":{"kind":"view_schema","id":"cartulary.view.unknown.v1"}}`,
			field:      "home_sheet_ref.id",
			reasonCode: "unknown_view_schema",
		},
		{
			name:       "extra sheet-ref member",
			body:       `{"home_sheet_ref":{"kind":"view_schema","id":"cartulary.view.timeline.v2","label":"Timeline"}}`,
			field:      "home_sheet_ref.label",
			reasonCode: "unknown_field",
		},
		{
			name:       "invalid saved view id",
			body:       `{"home_sheet_ref":{"kind":"saved_view","id":"svw_1"}}`,
			field:      "home_sheet_ref.id",
			reasonCode: "invalid_saved_view_id",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(strings.NewReader(tc.body))
			requireStartupAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", tc.field, tc.reasonCode)
		})
	}
}

func TestSupportPhase8_WorkbookStartupExplicitSheetRefParser(t *testing.T) {
	viewSchemaRef, apiErr := workbookstartup.ParseExplicitSheetRef(url.Values{"view_schema_id": []string{" cartulary.view.timeline.v2 "}})
	if apiErr != nil {
		t.Fatalf("parse view_schema_id selector: %v", apiErr)
	}
	if string(viewSchemaRef) != `{"kind":"view_schema","id":"cartulary.view.timeline.v2"}` {
		t.Fatalf("unexpected explicit view schema ref: %s", viewSchemaRef)
	}

	savedViewRef, apiErr := workbookstartup.ParseExplicitSheetRef(url.Values{
		"sheet_ref_kind": []string{"saved_view"},
		"sheet_ref_id":   []string{"00000000-0000-0000-0000-000000000153"},
	})
	if apiErr != nil {
		t.Fatalf("parse saved_view selector: %v", apiErr)
	}
	if string(savedViewRef) != `{"kind":"saved_view","id":"00000000-0000-0000-0000-000000000153"}` {
		t.Fatalf("unexpected explicit saved view ref: %s", savedViewRef)
	}

	_, apiErr = workbookstartup.ParseExplicitSheetRef(url.Values{
		"view_schema_id": []string{"cartulary.view.timeline.v2"},
		"sheet_ref_kind": []string{"view_schema"},
		"sheet_ref_id":   []string{"cartulary.view.hosts.v1"},
	})
	requireStartupAPIError(t, apiErr, http.StatusBadRequest, "invalid_startup_request", "sheet_ref", "ambiguous_explicit_sheet_ref")
}

func requireStartupAPIError(t testing.TB, apiErr *httpapi.APIError, wantStatus int, wantCode string, wantField string, wantReason string) {
	t.Helper()
	if apiErr == nil {
		t.Fatalf("expected API error %s", wantCode)
		return
	}
	if apiErr.Status != wantStatus || apiErr.Code != wantCode {
		t.Fatalf("unexpected API error: got status=%d code=%q details=%#v", apiErr.Status, apiErr.Code, apiErr.Details)
	}
	if apiErr.Details["field"] != wantField || apiErr.Details["reason_code"] != wantReason {
		t.Fatalf("unexpected API error details: got %#v want field=%q reason=%q", apiErr.Details, wantField, wantReason)
	}
}
