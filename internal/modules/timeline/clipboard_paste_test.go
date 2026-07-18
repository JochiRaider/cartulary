package timeline

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestClipboardPasteParsingMappingRawCaptureAndBinding(t *testing.T) {
	tsv, err := ParseClipboardTable("alpha\tbravo\ncharlie\tdelta\n", "tsv")
	if err != nil {
		t.Fatalf("parse tsv clipboard: %v", err)
	}
	if len(tsv) != 2 || tsv[0][0] != "alpha" || tsv[1][1] != "delta" {
		t.Fatalf("unexpected tsv parse: %#v", tsv)
	}
	csvRows, err := ParseClipboardTable("\"alpha, one\",bravo\ncharlie,delta", "csv")
	if err != nil {
		t.Fatalf("parse csv clipboard: %v", err)
	}
	if csvRows[0][0] != "alpha, one" || csvRows[1][1] != "delta" {
		t.Fatalf("unexpected csv parse: %#v", csvRows)
	}

	request := ClipboardPasteRequest{
		ViewSchemaID:  TimelineViewSchemaID,
		ClientTxnID:   "txn-u-9-02-clipboard",
		ClipboardText: "Gateway host\tanalyst@example.test\tWorkbook inspector summary\tunmapped value",
		Format:        "tsv",
		StartFieldKey: "timeline.host_refs",
		Columns: []string{
			"timeline.host_refs",
			"timeline.identity_refs",
			"timeline.activity_synopsis_text",
		},
		Targets: []ClipboardPasteTarget{{Kind: "create"}},
	}
	plan, err := BuildClipboardPastePlan(request)
	if err != nil {
		t.Fatalf("build clipboard paste plan: %v", err)
	}
	if len(plan.Rows) != 1 {
		t.Fatalf("expected one planned row, got %#v", plan.Rows)
	}
	row := plan.Rows[0]
	if len(row.Cells) != 3 {
		t.Fatalf("expected three known field cells, got %#v", row.Cells)
	}
	if row.Cells[0].FieldKey != "timeline.host_refs" || row.Cells[0].Change.ActionPayload == nil {
		t.Fatalf("host cell did not map to stable field key/action payload: %#v", row.Cells[0])
	}
	hostAction := row.Cells[0].Change.ActionPayload.Actions[0]
	if hostAction.Op != "add_token" || hostAction.RawText != "Gateway host" || strings.TrimSpace(hostAction.NormalizedText) == "" {
		t.Fatalf("host paste did not use mention-origin binding action: %#v", hostAction)
	}
	if row.Cells[1].FieldKey != "timeline.identity_refs" || row.Cells[1].Change.ActionPayload.Actions[0].Op != "add_token" {
		t.Fatalf("identity paste did not use declared entity binding mode: %#v", row.Cells[1])
	}
	if row.Cells[2].FieldKey != "timeline.activity_synopsis_text" || row.Cells[2].Change.TextValue == nil || *row.Cells[2].Change.TextValue != "Workbook inspector summary" {
		t.Fatalf("summary did not normalize as stable field-key value: %#v", row.Cells[2])
	}
	if len(row.Unknown) != 1 {
		t.Fatalf("expected one raw-capture unknown column, got %#v", row.Unknown)
	}
	unknown := row.Unknown[0]
	if unknown.SourceKind != "clipboard_paste" ||
		unknown.PasteClientTxnID != request.ClientTxnID ||
		unknown.SourceRowOrdinal != 1 ||
		unknown.SourceColumnOrdinal != 4 ||
		unknown.RawValue != "unmapped value" {
		t.Fatalf("unexpected raw import column: %#v", unknown)
	}

	rawCapture := rawCaptureWithImportColumns(map[string]any{"kept": "value"}, row.Unknown)
	importColumns, ok := rawCapture["import_columns"].([]any)
	if !ok || len(importColumns) != 1 {
		t.Fatalf("expected structured import_columns raw capture, got %#v", rawCapture)
	}
	if rawCapture["kept"] != "value" {
		t.Fatalf("raw capture did not preserve existing structure: %#v", rawCapture)
	}
}

func TestClipboardPasteRejectsCrossIncidentRecordTarget(t *testing.T) {
	authorizedIncidentID := uuid.New()
	otherIncidentID := uuid.New()
	current := sourceRecord{
		RecordID:   uuid.New(),
		IncidentID: otherIncidentID,
	}

	if err := ensureClipboardPasteRecordIncident(current, authorizedIncidentID); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("cross-incident paste target must be hidden as not found, got %v", err)
	}
	if err := ensureClipboardPasteRecordIncident(current, otherIncidentID); err != nil {
		t.Fatalf("same-incident paste target should be accepted: %v", err)
	}
}
