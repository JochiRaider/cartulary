package projections

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestGenericProjectionSurfaceMatrixCoversRegisteredViews(t *testing.T) {
	expected := map[string]bool{
		assessmentsViewSchemaID:          true,
		commLogViewSchemaID:              true,
		decisionsViewSchemaID:            true,
		evidenceViewSchemaID:             true,
		findingsViewSchemaID:             true,
		forensicKeywordsViewSchemaID:     true,
		handoffViewSchemaID:              true,
		investigativeQueriesViewSchemaID: true,
		lessonViewSchemaID:               true,
		notesViewSchemaID:                true,
		partiesViewSchemaID:              true,
		statusReviewViewSchemaID:         true,
		taskRequestsViewSchemaID:         true,
	}
	if !reflect.DeepEqual(surfaceKeySet(genericSurfaces), expected) {
		t.Fatalf("generic projection surface matrix changed:\ngot  %#v\nwant %#v", surfaceKeySet(genericSurfaces), expected)
	}
	for viewSchemaID, surface := range genericSurfaces {
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			t.Fatalf("generic projection surface %s has no registered view schema", viewSchemaID)
		}
		fields := schema.Fields()
		for _, field := range surface.fields {
			if _, ok := fields[field.key]; !ok {
				t.Fatalf("generic projection surface %s maps unknown field %s", viewSchemaID, field.key)
			}
		}
	}
}

func TestArtifactProjectionSurfacesUseContractFilters(t *testing.T) {
	tests := map[string]string{
		commLogViewSchemaID:              "comm_log",
		findingsViewSchemaID:             "finding",
		forensicKeywordsViewSchemaID:     "forensic_keyword",
		handoffViewSchemaID:              "handoff",
		investigativeQueriesViewSchemaID: "investigative_query",
		lessonViewSchemaID:               "lesson",
		notesViewSchemaID:                "note",
		statusReviewViewSchemaID:         "status_review",
	}
	for viewSchemaID, artifactType := range tests {
		t.Run(viewSchemaID, func(t *testing.T) {
			surface, ok := genericSurfaces[viewSchemaID]
			if !ok {
				t.Fatalf("missing generic projection surface %s", viewSchemaID)
			}
			if got := artifactTypeForSurface(viewSchemaID, artifactType); got != artifactType {
				t.Fatalf("%s artifact type: got %q want %q", viewSchemaID, got, artifactType)
			}
			if !strings.Contains(surface.whereSQL, "p.artifact_type = '"+artifactType+"'") {
				t.Fatalf("%s whereSQL does not use contract artifact filter: %q", viewSchemaID, surface.whereSQL)
			}
		})
	}
}

func TestGenericProjectionGroupedRowIsFullViewRow(t *testing.T) {
	recordID := uuid.MustParse("00000000-0000-0000-0000-000000000807")
	groupBy := "host.host_state"
	row, err := buildGenericRow(genericSurface{
		viewSchemaID: "cartulary.view.hosts.v1",
		recordExpr:   "h.record_id",
		fields: []genericField{
			{key: "host.display_name", kind: fieldKindText},
			{key: "host.host_state", kind: fieldKindText},
			{key: "host.edited_at", kind: fieldKindTimestamp},
		},
	}, &groupBy, []any{
		recordID,
		int64(12),
		"Host A",
		"reviewed",
		time.Date(2026, 5, 16, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("build grouped workbook row: %v", err)
	}

	if !reflect.DeepEqual(sortedMapKeys(row), []string{"cells", "group_values", "record_id", "row_version"}) {
		t.Fatalf("grouped workbook response row must remain a full row resource, got keys %#v in %#v", sortedMapKeys(row), row)
	}
	if row["record_id"] != recordID.String() || row["row_version"] != int64(12) {
		t.Fatalf("grouped workbook row must keep top-level record identity and version, got %#v", row)
	}
	cells, ok := row["cells"].(map[string]any)
	if !ok || len(cells) != 3 {
		t.Fatalf("grouped workbook row must serialize field cells, got %#v", row["cells"])
	}
	groupValues, ok := row["group_values"].(map[string]any)
	if !ok || groupValues[groupBy] != "reviewed" {
		t.Fatalf("grouped workbook row must serialize group_values without a header row, got %#v", row["group_values"])
	}

	payload, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal grouped workbook row: %v", err)
	}
	encoded := string(payload)
	for _, forbidden := range []string{
		`"row_kind"`,
		`"is_group_header"`,
		`"subtotal"`,
		`"writable"`,
		`"group_header_id"`,
		`"mutation_target"`,
		`"paste_target"`,
		`"target_record_id"`,
	} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("grouped workbook row serialized presentation-only marker %s: %s", forbidden, encoded)
		}
	}
}

func surfaceKeySet(surfaces map[string]genericSurface) map[string]bool {
	keys := make(map[string]bool, len(surfaces))
	for key := range surfaces {
		keys[key] = true
	}
	return keys
}

func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
