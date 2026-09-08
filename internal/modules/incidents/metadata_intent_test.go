package incidents

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIncidentMetadataNormalizationAndFieldIntent_Unit(t *testing.T) {
	actor := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	current := IncidentRecord{ID: actor, Description: stringPointerForTest("body"), Severity: stringPointerForTest("high"), TLP: stringPointerForTest("TLP:AMBER"), CurrentPhase: stringPointerForTest("triage"), PrimaryExternalCaseRef: stringPointerForTest("CASE-1"), IncidentVersion: 7, UpdatedAt: time.Unix(1, 0), UpdatedByUserID: &actor}
	values := func(r IncidentRecord) map[string]*string {
		return map[string]*string{"description": r.Description, "severity": r.Severity, "tlp": r.TLP, "current_phase": r.CurrentPhase, "primary_external_case_ref": r.PrimaryExternalCaseRef}
	}
	admit := func(field string, value any) (IncidentPatchAdmission, *AdmissionError) {
		t.Helper()
		payload := map[string]any{"base_incident_version": 7}
		if field != "" {
			payload[field] = value
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		return AdmitIncidentPatchJSON(strings.NewReader(string(raw)))
	}
	for _, field := range []string{"description", "severity", "tlp", "current_phase", "primary_external_case_ref"} {
		t.Run(field, func(t *testing.T) {
			for _, value := range []any{nil, ""} {
				request, err := admit(field, value)
				if field == "tlp" && value == "" {
					if err == nil {
						t.Fatal("TLP empty is not clear")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				next, changed := applyIncidentPatch(current, request, actor, time.Unix(2, 0))
				if !changed || values(next)[field] != nil || next.IncidentVersion != 8 {
					t.Fatal("explicit clear did not apply")
				}
				for other, before := range values(current) {
					if other != field && !stringPointersEqual(before, values(next)[other]) {
						t.Fatalf("unrelated field changed: %s", other)
					}
				}
			}
			if field == "tlp" {
				for _, value := range []any{"TLP:CLEAR", "TLP:GREEN", "TLP:AMBER", "TLP:AMBER+STRICT", "TLP:RED"} {
					if _, err := admit(field, value); err != nil {
						t.Fatal(err)
					}
				}
				for _, value := range []any{" TLP:RED ", "red", "TLP:WHITE", " \t ", 1, true} {
					if _, err := admit(field, value); err == nil {
						t.Fatalf("accepted invalid TLP %#v", value)
					}
				}
				return
			}
			request, err := admit(field, " \u0085\u2003 ")
			if err != nil {
				t.Fatal(err)
			}
			next, _ := applyIncidentPatch(current, request, actor, time.Unix(2, 0))
			if values(next)[field] != nil {
				t.Fatal("normalized empty must clear")
			}
			limit := 128
			raw, want := "  e\u0301  interior  ", "é  interior"
			if field == "description" {
				limit = 16384
				raw, want = "  e\u0301\r\n\tline\rnext  ", "é\n\tline\nnext"
			}
			request, err = admit(field, raw)
			if err != nil {
				t.Fatal(err)
			}
			next, _ = applyIncidentPatch(current, request, actor, time.Unix(2, 0))
			if values(next)[field] == nil || *values(next)[field] != want {
				t.Fatalf("normalization drift: %#v", values(next)[field])
			}
			for _, value := range []string{strings.Repeat("😀", limit), "  " + strings.Repeat("e\u0301", limit) + "  "} {
				if _, err := admit(field, value); err != nil {
					t.Fatalf("normalized boundary rejected: %v", err)
				}
			}
			for _, value := range []any{strings.Repeat("😀", limit+1), "a\x00b", "a\u009fb", 1, true, []string{}} {
				if _, err := admit(field, value); err == nil {
					t.Fatalf("invalid text accepted: %#v", value)
				}
			}
			if field != "description" {
				for _, value := range []string{"a\nb", "a\tb"} {
					if _, err := admit(field, value); err == nil {
						t.Fatal("single-line interior control accepted")
					}
				}
			}
		})
	}
	for _, payload := range []string{`{"base_incident_version":7}`, `{"base_incident_version":7,"severity":" high "}`} {
		request, err := AdmitIncidentPatchJSON(strings.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		next, changed := applyIncidentPatch(current, request, actor, time.Unix(2, 0))
		if changed || next.IncidentVersion != 7 || !next.UpdatedAt.Equal(current.UpdatedAt) || next.UpdatedByUserID != current.UpdatedByUserID {
			t.Fatal("no-op changed attribution or version")
		}
	}
	for _, field := range []string{"incident_id", "incident_key", "title", "status", "created_by_user_id", "created_at", "updated_at", "updated_by_user_id", "incident_version", "closed_at", "memberships", "saved_views", "workbook_preferences", "client_txn_id", "view_schema_id", "field_key", "record", "unknown"} {
		if _, err := admit(field, "forbidden"); err == nil {
			t.Fatalf("undeclared member accepted: %s", field)
		}
	}
	for _, payload := range []string{`null`, `[]`, `{}`, `{"base_incident_version":0}`, `{"base_incident_version":-1}`, `{"base_incident_version":1.1}`, `{"base_incident_version":"1"}`, `{"base_incident_version":null}`} {
		if _, err := AdmitIncidentPatchJSON(strings.NewReader(payload)); err == nil {
			t.Fatalf("invalid base accepted: %s", payload)
		}
	}
}
