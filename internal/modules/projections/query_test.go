package projections

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestGenericProjectionPageSQLIsKeysetBounded(t *testing.T) {
	surface := querySurfacesForTest()[assessmentsViewSchemaID]
	positionID := "00000000-0000-0000-0000-000000000901"
	sqlText, args, err := buildGenericQueryPageSQL(
		uuid.MustParse("00000000-0000-0000-0000-000000000900"),
		surface,
		viewschema.QueryMeta{Sort: []viewschema.SortEntry{{FieldKey: "record_id", Direction: "asc"}}},
		querypage.Window{Limit: 25, Position: map[string]string{"record_id": `"` + positionID + `"`}},
	)
	if err != nil {
		t.Fatalf("build generic page SQL: %v", err)
	}
	if !strings.Contains(sqlText, "record_id >") || !strings.Contains(sqlText, " LIMIT $") || strings.Contains(strings.ToUpper(sqlText), "OFFSET") {
		t.Fatalf("generic page SQL is not bounded keyset retrieval: %s", sqlText)
	}
	if got := args[len(args)-1]; got != 26 {
		t.Fatalf("generic page SQL limit argument = %#v, want 26", got)
	}
}

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
	surfaces := querySurfacesForTest()
	if !reflect.DeepEqual(surfaceKeySet(surfaces), expected) {
		t.Fatalf("generic projection surface matrix changed:\ngot  %#v\nwant %#v", surfaceKeySet(surfaces), expected)
	}
	for viewSchemaID, surface := range surfaces {
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
		gotFieldKeys := genericSurfaceFieldKeys(surface)
		wantFieldKeys := schemaFieldKeys(fields)
		if !reflect.DeepEqual(gotFieldKeys, wantFieldKeys) {
			t.Fatalf("%s generic projection fields drifted from schema registry:\ngot  %#v\nwant %#v", viewSchemaID, gotFieldKeys, wantFieldKeys)
		}
	}
}

func TestGenericProjectionContractQueryFieldsAreMapped(t *testing.T) {
	for viewSchemaID, surface := range querySurfacesForTest() {
		t.Run(viewSchemaID, func(t *testing.T) {
			schema, ok := viewschema.Lookup(viewSchemaID)
			if !ok {
				t.Fatalf("generic projection surface %s has no registered view schema", viewSchemaID)
			}
			for _, entry := range schema.DefaultSort() {
				if _, ok := surface.field(entry.FieldKey); !ok {
					t.Fatalf("%s default sort field %s is not mapped", viewSchemaID, entry.FieldKey)
				}
			}
			for _, fieldKey := range schema.SortFields() {
				if _, ok := surface.field(fieldKey); !ok {
					t.Fatalf("%s sort field %s is not mapped", viewSchemaID, fieldKey)
				}
			}
			for _, fieldKey := range schema.FilterFields() {
				if fieldKey == "note.full_text" {
					continue
				}
				if _, ok := surface.field(fieldKey); !ok {
					t.Fatalf("%s filter field %s is not mapped", viewSchemaID, fieldKey)
				}
			}
			for _, fieldKey := range schema.GroupingFields() {
				if _, ok := surface.field(fieldKey); !ok {
					t.Fatalf("%s grouping field %s is not mapped", viewSchemaID, fieldKey)
				}
			}
		})
	}
}

func TestGenericProjectionRowShapeForEverySurface(t *testing.T) {
	recordID := uuid.MustParse("00000000-0000-0000-0000-000000000901")
	for viewSchemaID, surface := range querySurfacesForTest() {
		t.Run(viewSchemaID, func(t *testing.T) {
			schema, schemaOK := viewschema.Lookup(viewSchemaID)
			if !schemaOK {
				t.Fatalf("missing registered view schema %s", viewSchemaID)
			}
			values := make([]any, 0, len(surface.fields)+2)
			values = append(values, recordID, int64(7))
			for _, field := range surface.fields {
				values = append(values, sampleProjectionValue(field))
			}
			row, err := buildGenericRow(surface, values)
			if err != nil {
				t.Fatalf("build generic row: %v", err)
			}
			wantRowKeys := []string{"cells", "record_id", "row_version"}
			if len(schema.GroupingFields()) > 0 {
				wantRowKeys = []string{"cells", "group_values", "record_id", "row_version"}
			}
			if !reflect.DeepEqual(sortedMapKeys(row), wantRowKeys) {
				t.Fatalf("%s row keys changed: %#v", viewSchemaID, sortedMapKeys(row))
			}
			if row["record_id"] != recordID.String() || row["row_version"] != int64(7) {
				t.Fatalf("%s row identity/version changed: %#v", viewSchemaID, row)
			}
			cells, cellsOK := row["cells"].(map[string]any)
			if !cellsOK || len(cells) != len(schema.Fields()) {
				t.Fatalf("%s cells changed: %#v", viewSchemaID, row["cells"])
			}
			for _, field := range surface.fields {
				cell, ok := cells[field.key].(map[string]any)
				if !ok {
					t.Fatalf("%s field %s cell shape changed: %#v", viewSchemaID, field.key, cells[field.key])
				}
				if _, ok := cell["value"]; !ok {
					t.Fatalf("%s field %s missing value wrapper: %#v", viewSchemaID, field.key, cell)
				}
			}
			groupingFields := schema.GroupingFields()
			if len(groupingFields) == 0 {
				if _, exists := row["group_values"]; exists {
					t.Fatalf("%s must omit group_values: %#v", viewSchemaID, row["group_values"])
				}
			} else {
				groupValues, ok := row["group_values"].(map[string]any)
				if !ok || len(groupValues) != len(groupingFields) {
					t.Fatalf("%s group_values changed: %#v", viewSchemaID, row["group_values"])
				}
				for _, fieldKey := range groupingFields {
					if _, exists := groupValues[fieldKey]; !exists {
						t.Fatalf("%s group_values missing %s: %#v", viewSchemaID, fieldKey, groupValues)
					}
				}
			}
		})
	}
}

func TestGenericProjectionNullAndCollectionCellShape(t *testing.T) {
	recordID := uuid.MustParse("00000000-0000-0000-0000-000000000902")
	for viewSchemaID, surface := range querySurfacesForTest() {
		t.Run(viewSchemaID, func(t *testing.T) {
			for fieldIndex, field := range surface.fields {
				values := make([]any, 0, len(surface.fields)+2)
				values = append(values, recordID, int64(9))
				for index, candidate := range surface.fields {
					if index == fieldIndex {
						values = append(values, nil)
						continue
					}
					values = append(values, sampleProjectionValue(candidate))
				}
				row, err := buildGenericRow(surface, values)
				if err != nil {
					t.Fatalf("build generic row with null %s: %v", field.key, err)
				}
				cells := row["cells"].(map[string]any)
				cell := cells[field.key].(map[string]any)
				got := cell["value"]
				if field.kind == fieldKindCollection {
					want := map[string]any{
						"kind":    "collection_value_v1",
						"ordered": false,
						"items":   []map[string]any{},
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("%s collection null shape for %s changed:\ngot  %#v\nwant %#v", viewSchemaID, field.key, got, want)
					}
					continue
				}
				if got != nil {
					t.Fatalf("%s scalar null shape for %s changed: %#v", viewSchemaID, field.key, got)
				}
			}
		})
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
	surfaces := querySurfacesForTest()
	for viewSchemaID, artifactType := range tests {
		t.Run(viewSchemaID, func(t *testing.T) {
			surface, ok := surfaces[viewSchemaID]
			if !ok {
				t.Fatalf("missing generic projection surface %s", viewSchemaID)
			}
			schema, ok := viewschema.Lookup(viewSchemaID)
			if !ok {
				t.Fatalf("%s missing view schema", viewSchemaID)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value != artifactType {
				t.Fatalf("%s canonical artifact filter got %#v want artifact_type=%q", viewSchemaID, filter, artifactType)
			}
			if !strings.Contains(surface.whereSQL, "p.artifact_type = '"+artifactType+"'") {
				t.Fatalf("%s whereSQL does not use contract artifact filter: %q", viewSchemaID, surface.whereSQL)
			}
		})
	}
}

func sampleProjectionValue(field genericField) any {
	switch field.kind {
	case fieldKindTimestamp, fieldKindDate:
		return time.Date(2026, 7, 1, 10, 15, 0, 123, time.UTC)
	case fieldKindBool:
		return true
	case fieldKindNumber:
		return int64(42)
	case fieldKindCollection:
		items := []map[string]any{{
			"item_ref":         "record_ref:00000000-0000-0000-0000-000000000903",
			"item_kind":        "record_ref",
			"display_text":     "record:00000000-0000-0000-0000-000000000903",
			"linked_record_id": "00000000-0000-0000-0000-000000000903",
		}}
		payload, err := json.Marshal(items)
		if err != nil {
			panic(err)
		}
		return string(payload)
	default:
		return "sample " + field.key
	}
}

func TestGenericProjectionGroupedRowIsFullViewRow(t *testing.T) {
	recordID := uuid.MustParse("00000000-0000-0000-0000-000000000807")
	groupBy := "host.host_state"
	row, err := buildGenericRow(genericSurface{
		viewSchemaID:   "cartulary.view.hosts.v1",
		recordExpr:     "h.record_id",
		groupingFields: []string{"host.host_state", "host.criticality"},
		fields: []genericField{
			{key: "host.display_name", kind: fieldKindText},
			{key: "host.host_state", kind: fieldKindText},
			{key: "host.criticality", kind: fieldKindText},
			{key: "host.edited_at", kind: fieldKindTimestamp},
		},
	}, []any{
		recordID,
		int64(12),
		"Host A",
		"reviewed",
		"critical",
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
	if !ok || len(cells) != 4 {
		t.Fatalf("grouped workbook row must serialize field cells, got %#v", row["cells"])
	}
	groupValues, ok := row["group_values"].(map[string]any)
	if !ok || len(groupValues) != 2 || groupValues[groupBy] != "reviewed" || groupValues["host.criticality"] != "critical" {
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

func querySurfacesForTest() map[string]genericSurface {
	return defaultProviderRegistry().querySurfaces
}

func genericSurfaceFieldKeys(surface genericSurface) []string {
	keys := make([]string, 0, len(surface.fields))
	seen := map[string]struct{}{}
	for _, field := range surface.fields {
		if _, ok := seen[field.key]; ok {
			keys = append(keys, "DUPLICATE:"+field.key)
			continue
		}
		seen[field.key] = struct{}{}
		keys = append(keys, field.key)
	}
	slices.Sort(keys)
	return keys
}

func schemaFieldKeys(fields map[string]viewschema.Field) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	slices.Sort(keys)
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
