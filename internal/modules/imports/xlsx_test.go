package imports

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"testing"
)

func TestXLSXIndexDiscoversSemanticLocatorsAndFormulaDiagnostics(t *testing.T) {
	t.Parallel()

	workbook, apiErr := indexXLSXWorkbook(
		advancedXLSXWorkbook(t, false),
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
		ArchiveLimits{
			DefaultMaxExtractedBytes: 1 << 20,
			MaxCompressionRatio:      100,
			MaxMembers:               40,
		},
	)
	if apiErr != nil {
		t.Fatalf("index workbook: %#v", apiErr)
	}
	kinds := make(map[string]int)
	for _, located := range workbook.ranges {
		kinds[located.kind]++
	}
	if kinds["xlsx_used_range"] != 2 ||
		kinds["xlsx_table"] != 1 ||
		kinds["xlsx_named_range"] != 2 {
		t.Fatalf("unexpected semantic locator inventory: %#v", kinds)
	}
	visible := workbook.sheetsByName["Visible"]
	if visible == nil || visible.usedRange == nil || sourceRectangleA1(*visible.usedRange) != "A1:C3" {
		t.Fatalf("style-only cells changed the used range: %#v", visible)
	}
	decoded, decodeErr := workbook.decodeRectangle(
		visible,
		*visible.usedRange,
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
	)
	if decodeErr != nil {
		t.Fatalf("decode visible range: %#v", decodeErr)
	}
	if decoded.rows[1][2].CellKind != "formula_cached" ||
		decoded.rows[1][2].DisplayText != "2" ||
		decoded.rows[2][2].CellKind != "blank" {
		t.Fatalf("formulas were not decoded as inert cached data: %#v", decoded.rows)
	}
	for _, warning := range []string{
		"external_links_ignored",
		"filtered_or_hidden_state_ignored",
		"formula_cached_value_missing",
		"formula_inert",
	} {
		if !slices.Contains(decoded.warningCodes, warning) {
			t.Fatalf("missing warning %q from %#v", warning, decoded.warningCodes)
		}
	}
	if !slices.Equal(decoded.blockingColumnOrdinals, []int{3}) {
		t.Fatalf("unexpected formula-cache blocking columns: %#v", decoded.blockingColumnOrdinals)
	}
}

func TestXLSXIndexRejectsDynamicNamedRangeAndArchiveAbuse(t *testing.T) {
	t.Parallel()

	_, dynamicErr := indexXLSXWorkbook(
		advancedXLSXWorkbook(t, true),
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
		ArchiveLimits{
			DefaultMaxExtractedBytes: 1 << 20,
			MaxCompressionRatio:      100,
			MaxMembers:               40,
		},
	)
	if dynamicErr == nil ||
		dynamicErr.Code != "import_source_unsupported" ||
		dynamicErr.Details["reason_code"] != "unsupported_named_range" {
		t.Fatalf("dynamic named range did not fail closed: %#v", dynamicErr)
	}

	_, memberErr := indexXLSXWorkbook(
		advancedXLSXWorkbook(t, false),
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
		ArchiveLimits{
			DefaultMaxExtractedBytes: 1 << 20,
			MaxCompressionRatio:      100,
			MaxMembers:               2,
		},
	)
	if memberErr == nil ||
		memberErr.Code != "import_source_rejected" ||
		memberErr.Details["reason_code"] != "archive_member_count_exceeded" {
		t.Fatalf("archive member limit did not fail closed: %#v", memberErr)
	}
}

func TestFormulaCacheBlockerClearsOnlyWhenAffectedColumnIsUnmapped(t *testing.T) {
	t.Parallel()

	field := "summary"
	mapped, err := json.Marshal(approvedMapping{SourceColumns: []sourceColumnMapping{
		{SourceColumnOrdinal: 3, FieldKey: &field},
	}})
	if err != nil {
		t.Fatalf("marshal mapped plan: %v", err)
	}
	state := unitSelectionState{
		status:             "mapped",
		mappingFingerprint: stringPointer("fingerprint"),
		approvedMapping:    mapped,
		blockingColumns:    []int32{3},
	}
	if state.statusAfterSelection() != "mapped" {
		t.Fatalf("mapped formula without a cache entered ready")
	}
	unmapped, err := json.Marshal(approvedMapping{SourceColumns: []sourceColumnMapping{
		{SourceColumnOrdinal: 3},
	}})
	if err != nil {
		t.Fatalf("marshal unmapped plan: %v", err)
	}
	state.approvedMapping = unmapped
	if state.statusAfterSelection() != "ready" {
		t.Fatalf("excluding the affected formula column did not clear readiness")
	}
}

func stringPointer(value string) *string {
	return &value
}

func advancedXLSXWorkbook(t testing.TB, dynamicName bool) []byte {
	t.Helper()

	definedNames := `
  <definedNames>
    <definedName name="LocalBlock" localSheetId="0">Visible!$A$1:$B$2</definedName>
    <definedName name="HiddenBlock">'Hidden Data'!$A$1:$A$2</definedName>
  </definedNames>`
	if dynamicName {
		definedNames = `
  <definedNames>
    <definedName name="DynamicBlock">OFFSET(Visible!$A$1,0,0,2,2)</definedName>
  </definedNames>`
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	advancedXLSXText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Visible" sheetId="1" r:id="rId1"/>
    <sheet name="Hidden Data" sheetId="2" state="veryHidden" r:id="rId2"/>
  </sheets>`+definedNames+`
</workbook>`)
	advancedXLSXText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
  <Relationship Id="rIdExternal" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/externalLink" Target="https://invalid.example/source.xlsx" TargetMode="External"/>
</Relationships>`)
	advancedXLSXText(t, writer, "xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <cols><col min="2" max="2" hidden="1"/></cols>
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>name</t></is></c><c r="B1" t="inlineStr"><is><t>value</t></is></c><c r="C1" t="inlineStr"><is><t>formula</t></is></c><c r="Z99" s="1"/></row>
    <row r="2" hidden="1"><c r="A2" t="inlineStr"><is><t>alpha</t></is></c><c r="B2"><v>1</v></c><c r="C2"><f>B2+1</f><v>2</v></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>beta</t></is></c><c r="B3"><v>2</v></c><c r="C3"><f>B3+1</f></c></row>
  </sheetData>
  <autoFilter ref="A1:C3"/>
  <tableParts count="1"><tablePart r:id="rTable1"/></tableParts>
</worksheet>`)
	advancedXLSXText(t, writer, "xl/worksheets/_rels/sheet1.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rTable1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/table" Target="../tables/table1.xml"/>
</Relationships>`)
	advancedXLSXText(t, writer, "xl/tables/table1.xml", `<?xml version="1.0" encoding="UTF-8"?>
<table xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" name="DataTable" displayName="DataTable" ref="A1:B3"/>`)
	advancedXLSXText(t, writer, "xl/worksheets/sheet2.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>hidden</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>semantic</t></is></c></row>
  </sheetData>
</worksheet>`)
	advancedXLSXText(t, writer, "xl/vbaProject.bin", "macro bytes are never opened")
	if err := writer.Close(); err != nil {
		t.Fatalf("close advanced XLSX: %v", err)
	}
	return buffer.Bytes()
}

func advancedXLSXText(t testing.TB, writer *zip.Writer, name string, content string) {
	t.Helper()
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create XLSX member %s: %v", name, err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		t.Fatalf("write XLSX member %s: %v", name, err)
	}
}
