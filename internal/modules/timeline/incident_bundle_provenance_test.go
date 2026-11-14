package timeline_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/provenancevalue"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const timelineProvenancePath = "data/timeline_source_provenance.ndjson"

func timelineProvenanceDigest(literal string) string {
	digest := sha256.Sum256([]byte(literal))
	return hex.EncodeToString(digest[:])
}
func timelineProvenanceMetadata(t testing.TB, row map[string]any, metadata map[string]any) {
	t.Helper()
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	row["source_metadata"] = metadata
	row["source_identity_sha256"] = timelineProvenanceDigest(string(raw))
}
func timelineProvenanceRow(incident uuid.UUID, ordinal int) map[string]any {
	return map[string]any{"record_id": timelineEventRow(incident, 1)["record_id"], "source_identity_sha256": timelineProvenanceDigest(`{"source_kind":"capture"}`), "source_row_ordinal": ordinal, "source_column_ordinal": 1, "source_kind": "capture", "source_metadata": map[string]any{"source_kind": "capture"}, "source_header": " raw e\u0301\r\n<>& ", "raw_value": "raw\r\n\t<>&\u200d", "cell_kind": "string", "created_at": "2026-01-02T03:04:05.123456+00:00"}
}
func timelineProvenanceBundle(t testing.TB, incident uuid.UUID, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	bundle := timelineEventBundle(t, timelineEventRow(incident, 1))
	var raw bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&raw).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	bundle[timelineProvenancePath] = raw.Bytes()
	return bundle
}

func TestTimelinePortableProvenanceAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, actor := uuid.New(), uuid.New()
	port := timeline.NewIncidentBundleSourcePort()
	input := timelineProfileInput(t, incident, actor)
	prepare := func(rows ...map[string]any) error {
		bundle := timelineProvenanceBundle(t, incident, rows...)
		before := maps.Clone(bundle)
		for path, raw := range before {
			before[path] = bytes.Clone(raw)
		}
		_, err := port.PrepareImport(ctx, bundle, input)
		if !reflect.DeepEqual(before, bundle) {
			t.Fatal("provenance admission mutated source")
		}
		return err
	}
	if err := prepare(timelineProvenanceRow(incident, 1)); err != nil {
		t.Fatal(err)
	}
	if err := prepare(); err != nil {
		t.Fatal(err)
	}
	identity := map[string]bool{"record_id": true, "source_identity_sha256": true, "source_row_ordinal": true, "source_column_ordinal": true}
	for field := range timelineProvenanceRow(incident, 1) {
		for _, variant := range []string{"missing", "type", "null"} {
			t.Run(field+"/"+variant, func(t *testing.T) {
				row := timelineProvenanceRow(incident, 1)
				switch variant {
				case "missing":
					delete(row, field)
				case "type":
					row[field] = []any{}
				case "null":
					row[field] = nil
				}
				err := prepare(row)
				if variant == "null" && (field == "source_header" || field == "cell_kind") {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				invariant := "version_shape_exact"
				if identity[field] {
					invariant = "source_identity_admitted"
				}
				requireTimelineProfileFailure(t, err, invariant)
			})
		}
	}
	for field, values := range map[string][]any{
		"record_id":              {uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", "bad"},
		"source_identity_sha256": {strings.Repeat("A", 64), strings.Repeat("g", 64), strings.Repeat("a", 63)},
		"source_row_ordinal":     {0, -1, json.Number("2147483648"), json.Number("1e2"), json.Number("1.0"), "1", true},
		"source_column_ordinal":  {0, -1, json.Number("2147483648"), json.Number("1e2"), json.Number("1.0"), "1", true},
		"source_kind":            {"", strings.Repeat("a", 65), "bad\nkind", "bad\x7fkind"},
		"cell_kind":              {"", strings.Repeat("a", 65), "bad\nkind", "bad\x7fkind"},
		"source_header":          {true, 12, json.Number("1"), map[string]any{}, "bad\x00header", strings.Repeat("a", 65535)},
		"raw_value":              {"bad\x00value", strings.Repeat("a", 65537)},
		"created_at":             {"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.100000Z", "2026-01-01T01:00:00+01:00"},
		"unknown":                {true},
	} {
		for _, value := range values {
			row := timelineProvenanceRow(incident, 1)
			row[field] = value
			invariant := "version_shape_exact"
			if identity[field] {
				invariant = "source_identity_admitted"
			}
			requireTimelineProfileFailure(t, prepare(row), invariant)
		}
	}
	for field, values := range map[string][]any{"source_row_ordinal": {2147483647}, "source_column_ordinal": {2147483647}, "source_header": {"", strings.Repeat("a", 65534)}, "raw_value": {"", strings.Repeat("a", 65536)}, "cell_kind": {strings.Repeat("界", 21) + "a"}, "created_at": {"0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999+00:00"}} {
		for _, value := range values {
			row := timelineProvenanceRow(incident, 1)
			row[field] = value
			if err := prepare(row); err != nil {
				t.Fatal(field, err)
			}
		}
	}
	for _, value := range []any{map[string]any{}, map[string]any{"source_kind": "capture", "unknown": "value"}, map[string]any{"source_kind": "capture", "locator": 12}, map[string]any{"source_kind": "capture", "locator": "bad\x00text"}} {
		row := timelineProvenanceRow(incident, 1)
		row["source_metadata"] = value
		requireTimelineProfileFailure(t, prepare(row), "version_shape_exact")
	}
	for _, metadata := range []map[string]any{{"source_kind": "other"}, {"source_kind": "capture", "locator": ""}, {"source_kind": "capture", "import_unit_id": uuid.Nil.String()}, {"source_kind": "capture", "mapping_fingerprint": "not-a-digest"}} {
		row := timelineProvenanceRow(incident, 1)
		timelineProvenanceMetadata(t, row, metadata)
		requireTimelineProfileFailure(t, prepare(row), "provenance_coherent")
	}
	wrongDigest := timelineProvenanceRow(incident, 1)
	wrongDigest["source_identity_sha256"] = strings.Repeat("0", 64)
	requireTimelineProfileFailure(t, prepare(wrongDigest), "provenance_coherent")
	// An independently written literal freezes ordering, Unicode and HTML escapes.
	row := timelineProvenanceRow(incident, 1)
	row["source_kind"] = "clipboard_paste"
	row["source_metadata"] = map[string]any{"source_kind": "clipboard_paste", "paste_client_txn_id": "txn<>&\u2028\u2029界", "mapping_fingerprint": strings.Repeat("a", 64)}
	literal := `{"mapping_fingerprint":"` + strings.Repeat("a", 64) + `","paste_client_txn_id":"txn\u003c\u003e\u0026\u2028\u2029界","source_kind":"clipboard_paste"}`
	row["source_identity_sha256"] = timelineProvenanceDigest(literal)
	if err := prepare(row); err != nil {
		t.Fatal("independent identity vector", err)
	}
	for _, key := range []string{"paste_client_txn_id", "mapping_fingerprint"} {
		bad := maps.Clone(row)
		metadata := maps.Clone(row["source_metadata"].(map[string]any))
		delete(metadata, key)
		timelineProvenanceMetadata(t, bad, metadata)
		requireTimelineProfileFailure(t, prepare(bad), "provenance_coherent")
	}
	fileLiteral := `{"import_session_id":"11111111-1111-4111-8111-111111111111","import_unit_id":"22222222-2222-4222-8222-222222222222","locator":"file","locator_kind":"csv_file","mapping_fingerprint":"` + strings.Repeat("a", 64) + `","parser_profile_id":"cartulary.import.workbook.v2","parser_version":"workbook_import_adapter_v2","source_content_sha256":"` + strings.Repeat("b", 64) + `","source_file_kind":"csv","source_kind":"file_import","source_rect_a1":"A1:B2"}`
	var fileMetadata map[string]any
	if err := json.Unmarshal([]byte(fileLiteral), &fileMetadata); err != nil {
		t.Fatal(err)
	}
	fileRow := timelineProvenanceRow(incident, 1)
	fileRow["source_kind"], fileRow["source_metadata"], fileRow["source_identity_sha256"] = "file_import", fileMetadata, timelineProvenanceDigest(fileLiteral)
	if err := prepare(fileRow); err != nil {
		t.Fatal("independent file identity vector", err)
	}
	for key := range fileMetadata {
		bad := maps.Clone(fileRow)
		metadata := maps.Clone(fileMetadata)
		delete(metadata, key)
		timelineProvenanceMetadata(t, bad, metadata)
		invariant := "provenance_coherent"
		if key == "source_kind" {
			invariant = "version_shape_exact"
		}
		requireTimelineProfileFailure(t, prepare(bad), invariant)
	}
	// The extra JSONB separator spaces matter even while compact JSON still fits.
	prefix, suffix := `{"paste_client_txn_id":"`, `","source_kind":"capture"}`
	for _, extra := range []int{0, 1} {
		value := strings.Repeat("a", 65536-len(prefix)-len(suffix)-3+extra)
		row := timelineProvenanceRow(incident, 1)
		row["source_metadata"] = map[string]any{"source_kind": "capture", "paste_client_txn_id": value}
		row["source_identity_sha256"] = timelineProvenanceDigest(prefix + value + suffix)
		err := prepare(row)
		if extra == 0 {
			if err != nil {
				t.Fatal("exact JSONB metadata bound", err)
			}
		} else {
			requireTimelineProfileFailure(t, err, "version_shape_exact")
		}
	}
	for _, count := range []int{10922, 10923} {
		row := timelineProvenanceRow(incident, 1)
		row["source_header"] = strings.Repeat("<", count)
		err := prepare(row)
		if count == 10922 {
			if err != nil {
				t.Fatal("escaped header boundary", err)
			}
		} else {
			requireTimelineProfileFailure(t, err, "version_shape_exact")
		}
	}
	for _, field := range []string{"kind", "metadata", "header", "raw", "cell"} {
		metadata := map[string]any{"source_kind": "capture"}
		var kind, header, raw, cell any = "capture", "header", "raw", "cell"
		invalid := string([]byte{0xff})
		switch field {
		case "kind":
			kind = invalid
		case "metadata":
			metadata["locator"] = invalid
		case "header":
			header = invalid
		case "raw":
			raw = invalid
		case "cell":
			cell = invalid
		}
		if _, ok := provenancevalue.ParseValues(kind, metadata, header, raw, cell); ok {
			t.Fatal("invalid Unicode repaired", field)
		}
	}
	first, second := timelineProvenanceRow(incident, 1), timelineProvenanceRow(incident, 2)
	requireTimelineProfileFailure(t, prepare(first, first), "source_identity_admitted")
	requireTimelineProfileFailure(t, prepare(second, first), "source_identity_admitted")
	otherSource := timelineProvenanceRow(incident, 1)
	otherSource["source_kind"] = "other"
	otherSource["source_metadata"] = map[string]any{"source_kind": "other"}
	otherSource["source_identity_sha256"] = timelineProvenanceDigest(`{"source_kind":"other"}`)
	if first["source_identity_sha256"].(string) > otherSource["source_identity_sha256"].(string) {
		first, otherSource = otherSource, first
	}
	if err := prepare(first, otherSource); err != nil {
		t.Fatal("equal ordinals from distinct sources", err)
	}
	requireTimelineProfileFailure(t, prepare(otherSource, first), "source_identity_admitted")
	orphan := timelineProvenanceRow(incident, 1)
	orphan["record_id"] = timelineEventRow(incident, 2)["record_id"]
	requireTimelineProfileFailure(t, prepare(orphan), "provenance_non_orphaned")
	rows := make([]map[string]any, 4096)
	for i := range rows {
		rows[i] = timelineProvenanceRow(incident, i+1)
	}
	if err := prepare(rows...); err != nil {
		t.Fatal("entry bound", err)
	}
	rows = append(rows, timelineProvenanceRow(incident, 4097))
	requireTimelineProfileFailure(t, prepare(rows...), "provenance_coherent")
	for _, raw := range []string{"null\n", "[]\n", "{} {}\n", "{\"raw_value\":\"\\ud800\"}\n", "{\"raw_value\":\"a\",\"raw_value\":\"b\"}\n"} {
		bundle := timelineProvenanceBundle(t, incident)
		bundle[timelineProvenancePath] = []byte(raw)
		_, err := port.PrepareImport(ctx, bundle, input)
		requireTimelineProfileFailure(t, err, "version_shape_exact")
	}
	bundle := timelineProvenanceBundle(t, incident)
	badProfile := timelineProfileRow(incident)
	badProfile["local_offset_minutes"] = nil
	bundle[timelineProfilePath] = timelineProfileBundle(t, badProfile)[timelineProfilePath]
	bundle[timelineProvenancePath] = []byte("null\n")
	_, err := port.PrepareImport(ctx, bundle, input)
	requireTimelineProfileFailure(t, err, "version_shape_exact")
}
