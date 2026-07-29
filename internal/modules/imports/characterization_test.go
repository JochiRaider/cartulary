package imports

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestCharacterizationCurrentImportSessionMemberRoutes(t *testing.T) {
	t.Parallel()

	sessionID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	unitID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	base := "/api/v1/import-sessions/" + sessionID.String()
	cases := []struct {
		path   string
		kind   string
		unitID uuid.UUID
	}{
		{path: base, kind: "session"},
		{path: base + "/units", kind: "units"},
		{path: base + "/units/" + unitID.String(), kind: "unit", unitID: unitID},
		{path: base + "/units/" + unitID.String() + "/preview", kind: "preview", unitID: unitID},
		{path: base + "/units/" + unitID.String() + "/mapping-preview", kind: "mapping_preview", unitID: unitID},
		{path: base + "/units/" + unitID.String() + "/mapping", kind: "mapping", unitID: unitID},
		{path: base + "/units/" + unitID.String() + "/select", kind: "select", unitID: unitID},
		{path: base + "/units/" + unitID.String() + "/skip", kind: "skip", unitID: unitID},
		{path: base + "/apply", kind: "apply"},
	}
	for _, testCase := range cases {
		t.Run(testCase.kind, func(t *testing.T) {
			route, ok := parseImportSessionPath(testCase.path)
			if !ok {
				t.Fatalf("expected current route %q to parse", testCase.path)
			}
			if route.Kind != testCase.kind || route.SessionID != sessionID || route.UnitID != testCase.unitID {
				t.Fatalf("unexpected route: got %#v", route)
			}
		})
	}

	for _, path := range []string{
		"/api/v1/import-sessions/",
		base + "/unknown",
		base + "/units/not-a-uuid",
		base + "/units/" + unitID.String() + "/regions",
	} {
		if route, ok := parseImportSessionPath(path); ok {
			t.Fatalf("expected current unsupported path %q to fail, got %#v", path, route)
		}
	}
}

func TestCharacterizationCurrentImportTargetInventory(t *testing.T) {
	t.Parallel()

	type expectedTarget struct {
		applyStatus string
		facade      string
	}
	expectedViews := map[string]expectedTarget{
		"cartulary.view.timeline.v2":              {applyStatus: applyStatusSupported, facade: "timeline.import_create"},
		"cartulary.view.hosts.v1":                 {applyStatus: applyStatusSupported, facade: "entities.host.import_create"},
		"cartulary.view.identities.v1":            {applyStatus: applyStatusSupported, facade: "entities.identity.import_create"},
		"cartulary.view.indicators.v1":            {applyStatus: applyStatusSupported, facade: "indicators.import_create"},
		"cartulary.view.evidence.v1":              {applyStatus: applyStatusSupported, facade: "evidence.import_create"},
		"cartulary.view.notes.v1":                 {applyStatus: applyStatusSupported, facade: "artifacts.note.import_create"},
		"cartulary.view.assessments.v1":           {applyStatus: applyStatusSupported, facade: "assessments.import_create"},
		"cartulary.view.task_requests.v1":         {applyStatus: applyStatusSupported, facade: "tasksdecisions.task_request.import_create"},
		"cartulary.view.decisions.v1":             {applyStatus: applyStatusSupported, facade: "tasksdecisions.decision.import_create"},
		"cartulary.view.parties.v1":               {applyStatus: applyStatusSupported, facade: "parties.import_create"},
		"cartulary.view.comm_log.v1":              {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.handoff.v1":               {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.status_review.v1":         {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.lesson.v1":                {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.findings.v1":              {applyStatus: applyStatusSupportedWhenAvailable},
		"cartulary.view.investigative_queries.v1": {applyStatus: applyStatusSupportedWhenAvailable},
		"cartulary.view.forensic_keywords.v1":     {applyStatus: applyStatusSupportedWhenAvailable},
	}
	if len(importTargets) != len(expectedViews) {
		t.Fatalf("current view target count changed: got %d want %d", len(importTargets), len(expectedViews))
	}
	for selector, expected := range expectedViews {
		target, ok := importTargets[selector]
		if !ok {
			t.Fatalf("missing current target %q", selector)
		}
		if target.ViewSchemaID != selector ||
			target.TargetKind != ImportTargetKindViewSchema ||
			target.ApplyStatus != expected.applyStatus ||
			target.CreateFacade != expected.facade {
			t.Fatalf("unexpected current target %q: got %#v want %#v", selector, target, expected)
		}
		if expected.applyStatus == applyStatusSupported && !target.importable(nil) {
			t.Fatalf("supported target %q is not importable", selector)
		}
		if expected.applyStatus == applyStatusSupportedWhenAvailable && target.importable(func(string) bool { return true }) {
			t.Fatalf("reserved target %q became importable", selector)
		}
	}

	key := analyticalImportTargetKey{
		TargetKind:         ImportTargetKindNetworkFlowTable,
		ExtensionProfileID: NetworkFlowExtensionProfileID,
	}
	if len(analyticalImportTargets) != 1 {
		t.Fatalf("current analytical target count changed: got %d want 1", len(analyticalImportTargets))
	}
	analytical := analyticalImportTargets[key]
	if analytical.ApplyStatus != applyStatusSupportedWhenClaimed ||
		analytical.ApplyFacade != "network_flow_import_facade_v1" ||
		analytical.importable(nil) ||
		!analytical.importable(func(profileID string) bool {
			return profileID == NetworkFlowExtensionProfileID
		}) {
		t.Fatalf("unexpected current analytical target: %#v", analytical)
	}
}

func TestCharacterizationKnownNonConformanceInternalErrorEchoesRawMessage(t *testing.T) {
	t.Parallel()

	apiErr := internalAPIError(errors.New("sensitive owner detail"))
	if apiErr.Status != http.StatusInternalServerError ||
		apiErr.Code != "internal_error" ||
		apiErr.Message != "sensitive owner detail" {
		t.Fatalf("current internal error behavior changed: %#v", apiErr)
	}
}

func TestCharacterizationKnownNonConformanceHiddenXLSXSheetIsSkipped(t *testing.T) {
	t.Parallel()

	tables, apiErr := parseXLSXTables(
		characterizationWorkbook(t),
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
		ArchiveLimits{
			DefaultMaxExtractedBytes: 1 << 20,
			MaxCompressionRatio:      100,
			MaxMembers:               20,
		},
	)
	if apiErr != nil {
		t.Fatalf("parse characterization workbook: %#v", apiErr)
	}
	if len(tables) != 1 || tables[0].SheetName != "Visible" {
		t.Fatalf("current parser must expose only the visible sheet before RS-09, got %#v", tables)
	}
}

func characterizationWorkbook(t testing.TB) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	writeCharacterizationZipText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Visible" sheetId="1" r:id="rId1"/>
    <sheet name="Hidden" sheetId="2" state="hidden" r:id="rId2"/>
  </sheets>
</workbook>`)
	writeCharacterizationZipText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`)
	for index, value := range []string{"visible", "hidden"} {
		writeCharacterizationZipText(t, writer, "xl/worksheets/sheet"+string(rune('1'+index))+".xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>header</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>`+value+`</t></is></c></row>
  </sheetData>
</worksheet>`)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close characterization workbook: %v", err)
	}
	return buffer.Bytes()
}

func writeCharacterizationZipText(t testing.TB, writer *zip.Writer, name string, content string) {
	t.Helper()

	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
