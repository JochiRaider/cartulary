package savedviews

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func TestIncidentBundleSavedViewImportValidationNormalizesPortableRows(t *testing.T) {
	row := validPortableSavedViewRow()
	if err := validateAndNormalizeImportedSavedView(row, savedViewImportSpec()); err != nil {
		t.Fatalf("validate saved view import row: %v", err)
	}
	if row["display_name"] != "Portable saved view" {
		t.Fatalf("display_name was not normalized: %#v", row["display_name"])
	}
	query := decodeMap(t, row["query_json"])
	if _, ok := query["group_by"]; ok {
		t.Fatalf("inactive group_by must be omitted after normalization: %#v", query)
	}
	if len(query["sort"].([]any)) != 0 || len(query["filters"].([]any)) != 0 {
		t.Fatalf("query empty defaults not normalized: %#v", query)
	}
	layout := decodeMap(t, row["layout_json"])
	if layout["layout_schema_id"] != "cartulary.layout.v1" {
		t.Fatalf("layout was not normalized: %#v", layout)
	}
}

func TestIncidentBundleSavedViewImportValidationRejectsMalformedRows(t *testing.T) {
	cases := map[string]func(map[string]any){
		"private owner null": func(row map[string]any) {
			row["owner_user_id"] = nil
		},
		"system owner present": func(row map[string]any) {
			row["scope"] = "system"
		},
		"unknown view schema": func(row map[string]any) {
			row["view_schema_id"] = "cartulary.view.unknown.v1"
		},
		"group by null": func(row map[string]any) {
			row["query_json"] = map[string]any{"group_by": nil}
		},
		"missing saved view file column": func(row map[string]any) {
			delete(row, "saved_view_id")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			row := validPortableSavedViewRow()
			mutate(row)
			err := validateAndNormalizeImportedSavedView(row, savedViewImportSpec())
			var verification *incidentportability.VerificationFailure
			if !errors.As(err, &verification) || verification.ReasonCode != "malformed_manifest" {
				t.Fatalf("saved view import validation error got %T %v", err, err)
			}
		})
	}
}

func validPortableSavedViewRow() map[string]any {
	return map[string]any{
		"saved_view_id":      "00000000-0000-4000-8000-000000110401",
		"incident_id":        "00000000-0000-4000-8000-000000110402",
		"view_schema_id":     "cartulary.view.timeline.v2",
		"scope":              "private",
		"display_name":       "  Portable saved view  ",
		"query_json":         map[string]any{},
		"layout_json":        map[string]any{},
		"owner_user_id":      "00000000-0000-4000-8000-000000110403",
		"created_at":         "2026-07-02T00:00:00Z",
		"updated_at":         "2026-07-02T00:00:00Z",
		"saved_view_version": json.Number("1"),
	}
}

func decodeMap(t testing.TB, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal normalized json: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode normalized json: %v", err)
	}
	return decoded
}
