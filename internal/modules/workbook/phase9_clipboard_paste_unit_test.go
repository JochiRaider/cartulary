package workbook

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestPhase9_U_9_02_SharedPasteAndBulkPlanningGroupsOneVisibleAction(t *testing.T) {
	dimensions, err := tabularingest.DimensionsForText("\"alpha, one\",bravo\ncharlie,delta", "csv")
	if err != nil {
		t.Fatalf("dimensions for quoted csv: %v", err)
	}
	if dimensions.RowCount != 2 || dimensions.ColumnCount != 2 {
		t.Fatalf("unexpected quoted csv dimensions: %#v", dimensions)
	}

	singleRowDimensions, err := tabularingest.DimensionsForText("alpha,beta", "csv")
	if err != nil {
		t.Fatalf("dimensions for single-row csv: %v", err)
	}
	if singleRowDimensions.RowCount != 1 || singleRowDimensions.ColumnCount != 2 {
		t.Fatalf("unexpected single-row csv dimensions: %#v", singleRowDimensions)
	}

	hostMode := "mention_origin"
	plan, err := tabularingest.BuildBatchPlan(tabularingest.MappingRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		ClientTxnID:    "txn-u-9-02-shared-ingest",
		SourceKind:     "clipboard_paste",
		Text:           "gateway\tSummary\tunknown\nworkstation\tSecond\tother",
		Format:         "tsv",
		StartFieldKey:  "timeline.host_refs",
		Columns:        []string{"timeline.host_refs", "timeline.activity_synopsis_text"},
		RequireTargets: 2,
	})
	if err != nil {
		t.Fatalf("build shared ingest plan: %v", err)
	}
	if plan.SourceKind != "clipboard_paste" || plan.ClientTxnID != "txn-u-9-02-shared-ingest" || len(plan.Rows) != 2 {
		t.Fatalf("unexpected batch identity/grouping: %#v", plan)
	}
	if len(plan.Rows[0].Cells) != 2 || plan.Rows[0].Cells[0].FieldKey != "timeline.host_refs" || plan.Rows[0].Cells[1].FieldKey != "timeline.activity_synopsis_text" {
		t.Fatalf("unexpected mapped cells: %#v", plan.Rows[0].Cells)
	}
	if plan.Rows[0].Cells[0].EntityBindingMode == nil || *plan.Rows[0].Cells[0].EntityBindingMode != hostMode {
		t.Fatalf("shared ingest did not carry entity_binding_mode: %#v", plan.Rows[0].Cells[0])
	}
	if len(plan.Rows[0].Unknown) != 1 || plan.Rows[0].Unknown[0].SourceColumnOrdinal != 3 || plan.Rows[0].Unknown[0].RawValue != "unknown" {
		t.Fatalf("unknown column was not preserved by ordinal: %#v", plan.Rows[0].Unknown)
	}

	fillDown, apiErr := DecodeBulkMutationRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.timeline.v2",
		"client_txn_id":"txn-u-9-02-fill-down",
		"kind":"fill_down_v1",
		"field_key":"timeline.raw_activity_text",
		"value":"Filled source",
		"targets":[
			{"record_id":"11111111-1111-4111-8111-111111111111","base_row_version":3},
			{"record_id":"22222222-2222-4222-8222-222222222222","base_row_version":4}
		]
	}`), timeline.TimelineViewSchemaID)
	if apiErr != nil {
		t.Fatalf("decode fill-down bulk request: %#v", apiErr)
	}
	fillPaste := timelineBulkClipboardRequest(fillDown)
	if fillPaste.SourceKind != "bulk_edit" || fillPaste.RouteKey != workbookBulkMutationRouteKey || fillPaste.ClientTxnID != "txn-u-9-02-fill-down" {
		t.Fatalf("fill-down did not preserve one visible bulk action identity: %#v", fillPaste)
	}
	if fillPaste.StartFieldKey != "timeline.raw_activity_text" || len(fillPaste.Columns) != 1 || fillPaste.Columns[0] != "timeline.raw_activity_text" {
		t.Fatalf("fill-down did not target one stable field column: %#v", fillPaste)
	}
	if fillPaste.ClipboardText != "Filled source\nFilled source" || len(fillPaste.Targets) != 2 {
		t.Fatalf("fill-down did not expand one value over stable targets: %#v", fillPaste)
	}
	if fillPaste.Targets[0].RecordID.String() != "11111111-1111-4111-8111-111111111111" || fillPaste.Targets[0].BaseRowVersion != 3 {
		t.Fatalf("fill-down first target changed: %#v", fillPaste.Targets[0])
	}

	tagAssignment, apiErr := DecodeBulkMutationRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.timeline.v2",
		"client_txn_id":"txn-u-9-02-tag-assignment",
		"kind":"multi_row_tag_assignment_v1",
		"tag_name":"Bulk Tag",
		"targets":[
			{"record_id":"33333333-3333-4333-8333-333333333333","base_row_version":5},
			{"record_id":"44444444-4444-4444-8444-444444444444","base_row_version":6}
		]
	}`), timeline.TimelineViewSchemaID)
	if apiErr != nil {
		t.Fatalf("decode multi-row tag assignment request: %#v", apiErr)
	}
	tagPaste := timelineBulkClipboardRequest(tagAssignment)
	if tagPaste.SourceKind != "bulk_edit" || tagPaste.RouteKey != workbookBulkMutationRouteKey || tagPaste.ClientTxnID != "txn-u-9-02-tag-assignment" {
		t.Fatalf("tag assignment did not preserve one visible bulk action identity: %#v", tagPaste)
	}
	if tagPaste.StartFieldKey != "timeline.tags" || len(tagPaste.Columns) != 1 || tagPaste.Columns[0] != "timeline.tags" {
		t.Fatalf("tag assignment did not target the tags collection field: %#v", tagPaste)
	}
	if tagPaste.ClipboardText != "Bulk Tag\nBulk Tag" || len(tagPaste.Targets) != 2 {
		t.Fatalf("tag assignment did not expand one tag over stable targets: %#v", tagPaste)
	}
	if tagPaste.Targets[1].RecordID.String() != "44444444-4444-4444-8444-444444444444" || tagPaste.Targets[1].BaseRowVersion != 6 {
		t.Fatalf("tag assignment second target changed: %#v", tagPaste.Targets[1])
	}
}
