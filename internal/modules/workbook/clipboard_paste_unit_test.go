package workbook

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestSharedPasteAndBulkPlanningGroupsOneVisibleAction_Unit(t *testing.T) {
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
	plan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
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
	if len(plan.Rows[0].Unmapped) != 1 || plan.Rows[0].Unmapped[0].SourceColumnOrdinal != 3 || plan.Rows[0].Unmapped[0].RawValue != "unknown" {
		t.Fatalf("unmapped column was not preserved by ordinal: %#v", plan.Rows[0].Unmapped)
	}

	fillDown, apiErr := DecodeTimelineBulkMutationRequest(strings.NewReader(`{
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
	if fillDown.Kind != timeline.OwnerBatchOperationFillDownV1 || fillDown.ClientTxnID != "txn-u-9-02-fill-down" {
		t.Fatalf("fill-down did not preserve its exact operation identity: %#v", fillDown)
	}
	if fillDown.FieldKey != "timeline.raw_activity_text" || fillDown.Value != "Filled source" {
		t.Fatalf("fill-down did not preserve its stable field command: %#v", fillDown)
	}
	if len(fillDown.Targets) != 2 {
		t.Fatalf("fill-down did not preserve stable targets: %#v", fillDown)
	}
	if fillDown.Targets[0].RecordID.String() != "11111111-1111-4111-8111-111111111111" || fillDown.Targets[0].BaseRowVersion != 3 {
		t.Fatalf("fill-down first target changed: %#v", fillDown.Targets[0])
	}

	tagAssignment, apiErr := DecodeTimelineBulkMutationRequest(strings.NewReader(`{
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
	if tagAssignment.Kind != timeline.OwnerBatchOperationMultiRowTagAssignmentV1 || tagAssignment.ClientTxnID != "txn-u-9-02-tag-assignment" {
		t.Fatalf("tag assignment did not preserve its exact operation identity: %#v", tagAssignment)
	}
	if tagAssignment.TagName != "Bulk Tag" || tagAssignment.NormalizedTag != "bulk tag" {
		t.Fatalf("tag assignment did not preserve its normalized command: %#v", tagAssignment)
	}
	if len(tagAssignment.Targets) != 2 {
		t.Fatalf("tag assignment did not preserve stable targets: %#v", tagAssignment)
	}
	if tagAssignment.Targets[1].RecordID.String() != "44444444-4444-4444-8444-444444444444" || tagAssignment.Targets[1].BaseRowVersion != 6 {
		t.Fatalf("tag assignment second target changed: %#v", tagAssignment.Targets[1])
	}
	if string(TimelineBulkMutationRequestHash(fillDown)) == string(TimelineBulkMutationRequestHash(tagAssignment)) {
		t.Fatal("distinct bulk operation discriminators must produce distinct request hashes")
	}

	exactPlan, err := BuildTimelineClipboardPlan(TimelineClipboardPasteRequest{
		ViewSchemaID:  timeline.TimelineViewSchemaID,
		ClientTxnID:   "txn-exact-header",
		ClipboardText: strings.Join(timelineV2ExactHeaderLabels, "\t") + "\n2026-07-25",
		Format:        "tsv",
		StartFieldKey: "timeline.activity_synopsis_text",
		Columns:       []string{"timeline.activity_synopsis_text"},
		Targets:       []TimelineBatchTarget{{Kind: "create"}},
	})
	if err != nil {
		t.Fatalf("build exact-header Timeline plan: %v", err)
	}
	if len(exactPlan.Rows) != 1 || len(exactPlan.Rows[0].Cells) != len(timelineV2ExactHeaderFieldKeys) {
		t.Fatalf("exact Timeline header did not retain its complete stable mapping: %#v", exactPlan.Rows)
	}
	if exactPlan.Rows[0].Cells[0].RawValue != "2026-07-25" || exactPlan.Rows[0].Cells[1].RawValue != "" {
		t.Fatalf("exact Timeline header did not preserve empty trailing cells: %#v", exactPlan.Rows[0].Cells)
	}
}
