package timeline

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/testutil/irfixtures"
)

func TestIRCorpusClipboardBoundaries(t *testing.T) {
	root, err := filepath.Abs("../../../fixtures")
	if err != nil {
		t.Fatal(err)
	}
	corpus := irfixtures.ForTest(t, root)

	rows, err := corpus.Records("ir-incident/v1/timeline.tsv", '\t')
	if err != nil {
		t.Fatal(err)
	}
	var columns []string
	for _, source := range corpus.Manifest.Sources {
		if source.Path == "timeline.tsv" {
			columns = source.FieldKeys
		}
	}
	for _, count := range []int{499, 500, 501} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var text bytes.Buffer
			writer := csv.NewWriter(&text)
			writer.Comma = '\t'
			if err := writer.WriteAll(rows[1 : count+1]); err != nil {
				t.Fatal(err)
			}
			plan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{ViewSchemaID: TimelineViewSchemaID, ClientTxnID: fmt.Sprintf("ir-paste-%d", count), SourceKind: "clipboard_paste", Text: text.String(), Format: "tsv", StartFieldKey: columns[0], Columns: columns, RequireTargets: count})
			if count > 500 {
				if err == nil || len(plan.Rows) != 0 {
					t.Fatal("oversize paste returned a partial write plan")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Rows) != count {
				t.Fatal("clipboard records truncated")
			}
			for i, row := range plan.Rows {
				if len(row.Cells) != len(columns) {
					t.Fatalf("record %d lost cells", i+1)
				}
				for j, cell := range row.Cells {
					if cell.RawValue != rows[i+1][j] {
						t.Fatalf("record %d column %d changed", i+1, j+1)
					}
				}
			}
			if _, err := buildClipboardOwnerRows(plan); err != nil {
				t.Fatalf("owner rejected admitted clipboard: %v", err)
			}
		})
	}
	whole, err := corpus.Read("ir-incident/v1/timeline.tsv")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{ViewSchemaID: TimelineViewSchemaID, ClientTxnID: "ir-paste-whole", SourceKind: "clipboard_paste", Text: string(whole), Format: "tsv", StartFieldKey: columns[0], Columns: columns, RequireTargets: 1500})
	if err == nil || len(plan.Rows) != 0 {
		t.Fatal("complete TSV returned a partial write plan")
	}
}
