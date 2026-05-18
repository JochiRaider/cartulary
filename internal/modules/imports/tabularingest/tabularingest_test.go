package tabularingest_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestSupportPhase9_SharedTabularIngestParsesMapsAndGroupsBatch(t *testing.T) {
	rows, err := tabularingest.ParseTable("\"alpha, one\",bravo\ncharlie,delta", "csv")
	if err != nil {
		t.Fatalf("parse quoted csv: %v", err)
	}
	if rows[0][0] != "alpha, one" || rows[1][1] != "delta" {
		t.Fatalf("unexpected quoted csv rows: %#v", rows)
	}
	dimensions, err := tabularingest.DimensionsForText("\"alpha, one\",bravo\ncharlie,delta", "csv")
	if err != nil {
		t.Fatalf("dimensions for quoted csv: %v", err)
	}
	if dimensions.RowCount != 2 || dimensions.ColumnCount != 2 {
		t.Fatalf("unexpected dimensions: %#v", dimensions)
	}

	hostMode := "mention_origin"
	plan, err := tabularingest.BuildBatchPlan(tabularingest.MappingRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		ClientTxnID:    "txn-u-9-02-shared-ingest",
		SourceKind:     "clipboard_paste",
		Text:           "gateway\tSummary\tunknown\nworkstation\tSecond\tother",
		Format:         "tsv",
		StartFieldKey:  "timeline.host_refs",
		Columns:        []string{"timeline.host_refs", "timeline.summary"},
		RequireTargets: 2,
	})
	if err != nil {
		t.Fatalf("build shared ingest plan: %v", err)
	}
	if plan.SourceKind != "clipboard_paste" || plan.ClientTxnID != "txn-u-9-02-shared-ingest" || len(plan.Rows) != 2 {
		t.Fatalf("unexpected batch identity/grouping: %#v", plan)
	}
	if len(plan.Rows[0].Cells) != 2 || plan.Rows[0].Cells[0].FieldKey != "timeline.host_refs" || plan.Rows[0].Cells[1].FieldKey != "timeline.summary" {
		t.Fatalf("unexpected mapped cells: %#v", plan.Rows[0].Cells)
	}
	if plan.Rows[0].Cells[0].EntityBindingMode == nil || *plan.Rows[0].Cells[0].EntityBindingMode != hostMode {
		t.Fatalf("shared ingest did not carry entity_binding_mode: %#v", plan.Rows[0].Cells[0])
	}
	if len(plan.Rows[0].Unknown) != 1 || plan.Rows[0].Unknown[0].SourceColumnOrdinal != 3 || plan.Rows[0].Unknown[0].RawValue != "unknown" {
		t.Fatalf("unknown column was not preserved by ordinal: %#v", plan.Rows[0].Unknown)
	}
}
