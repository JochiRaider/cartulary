package incidents

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPhase8_WorkbookPreferencePointers_U_8_05(t *testing.T) {
	userRequest, apiErr := DecodeUserWorkbookPreferencesPutRequest(strings.NewReader(`{
		"home_sheet_ref":{"kind":"view_schema","id":"cartulary.view.timeline.v1"}
	}`))
	if apiErr != nil {
		t.Fatalf("decode home preference: %v", apiErr)
	}
	defaultRequest, apiErr := DecodeDefaultWorkbookPreferencesPutRequest(strings.NewReader(`{
		"default_sheet_ref":{"kind":"view_schema","id":"cartulary.view.hosts.v1"}
	}`))
	if apiErr != nil {
		t.Fatalf("decode default preference: %v", apiErr)
	}

	var home map[string]string
	var defaults map[string]string
	if err := json.Unmarshal(userRequest.HomeSheetRef, &home); err != nil {
		t.Fatalf("decode canonical home ref: %v", err)
	}
	if err := json.Unmarshal(defaultRequest.DefaultSheetRef, &defaults); err != nil {
		t.Fatalf("decode canonical default ref: %v", err)
	}
	if reflect.DeepEqual(home, defaults) {
		t.Fatalf("home_sheet_ref and default_sheet_ref must remain separate, got %v", home)
	}
	if home["id"] != "cartulary.view.timeline.v1" || defaults["id"] != "cartulary.view.hosts.v1" {
		t.Fatalf("unexpected refs: home=%v default=%v", home, defaults)
	}
}

func TestPhase8_WorkbookStartupFallback_I_8_02(t *testing.T) {
	explicit := StartupCandidate{SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.hosts.v1"}, Valid: true}
	home := StartupCandidate{SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.identities.v1"}, Valid: true}
	defaults := StartupCandidate{SheetRef: map[string]string{"kind": "view_schema", "id": "cartulary.view.evidence.v1"}, Valid: true}
	selected, cleared := SelectStartupSheet(explicit, home, defaults)
	if selected["id"] != "cartulary.view.hosts.v1" || len(cleared) != 0 {
		t.Fatalf("explicit sheet must win, selected=%v cleared=%v", selected, cleared)
	}

	selected, cleared = SelectStartupSheet(StartupCandidate{}, home, defaults)
	if selected["id"] != "cartulary.view.identities.v1" || len(cleared) != 0 {
		t.Fatalf("home sheet must win after omitted explicit, selected=%v cleared=%v", selected, cleared)
	}

	selected, cleared = SelectStartupSheet(
		StartupCandidate{},
		StartupCandidate{SheetRef: map[string]string{"kind": "saved_view", "id": "deleted"}, Valid: false},
		defaults,
	)
	if selected["id"] != "cartulary.view.evidence.v1" || !reflect.DeepEqual(cleared, []string{"home"}) {
		t.Fatalf("invalid home pointer must clear and fall back deterministically, selected=%v cleared=%v", selected, cleared)
	}
}
