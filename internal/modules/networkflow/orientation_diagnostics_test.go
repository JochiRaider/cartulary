package networkflow

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func TestNetworkFlowOrientationDiagnosticDependencies(t *testing.T) {
	_, preview, baseline := approvedVariantFixture(t, "cisco_sna_flow_search_csv_v2")
	mapping, err := decodeApprovedMapping(preview.OwnerMapping, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	header := strings.Split(baseline, "\n")[0]
	base := strings.Split(strings.Split(baseline, "\n")[1], ",")
	for _, tc := range []struct {
		name     string
		mutate   func([]string)
		ordinals []int64
	}{
		{"protocol", func(c []string) { c[6], c[9], c[10] = "bogus", "client", "client" }, []int64{7, 10, 11}},
		{"endpoint", func(c []string) { c[2] = "not-an-ip" }, []int64{3, 10, 11}},
		{"counter", func(c []string) { c[7] = "-1" }, []int64{8, 10, 11}},
		{"timestamp order", func(c []string) { c[0], c[1] = c[1], c[0] }, []int64{2, 10, 11}},
		{"oversized unrelated", func(c []string) { c[11] = strings.Repeat("x", 257) }, []int64{10, 11, 12}},
		{"invalid role", func(c []string) { c[9] = "responder"; c[6] = "bogus" }, []int64{7, 10}},
		{"oversized role", func(c []string) { c[9] = strings.Repeat("x", 257); c[6] = "bogus" }, []int64{7, 10}},
		{"blank role", func(c []string) { c[9] = ""; c[6] = "bogus" }, []int64{7}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cells := append([]string(nil), base...)
			cells[9], cells[10] = "server", "server"
			tc.mutate(cells)
			source := header + "\n" + strings.Join(cells, ",") + "\n"
			limits := defaultLimits()
			limits.MaxRawCellScalarLength = 256
			parsed, err := parseCSVPreview(strings.NewReader(source), "", limits)
			if err != nil {
				t.Fatal(err)
			}
			m := mapping
			m.SourceColumns = parsed.SourceColumns
			fingerprint := mustMappingFingerprint(m, parsed.SourceContentSHA256)
			row, diagnostics, err := validateCSVRecord(parsed.Records[0], parsed.SourceColumns, m, fingerprint, parsed.SourceContentSHA256, limits, sourceFieldMappings(m))
			if err != nil || !reflect.DeepEqual(row, normalizedRow{}) || len(diagnostics) != len(tc.ordinals) {
				t.Fatalf("rejected normalization: %v %#v %#v", err, row, diagnostics)
			}
			for i, d := range diagnostics {
				ordinal := tc.ordinals[i]
				if d.SourceColumnOrdinal == nil || *d.SourceColumnOrdinal != ordinal || d.RawHeaderSHA256 == nil || *d.RawHeaderSHA256 != parsed.SourceColumns[ordinal-1].RawHeaderSHA256 || d.RawValueSHA256 == nil || *d.RawValueSHA256 != sha256Hex([]byte(cells[ordinal-1])) || d.SafeSample != nil {
					t.Fatalf("diagnostic attribution/privacy: %#v", d)
				}
				if err := validateDiagnostic(d, parsed.SourceContentSHA256, fingerprint); err != nil {
					t.Fatal(err)
				}
			}
			counts, err := validatePreview(parsed, m, fingerprint, limits)
			if err != nil || counts.AcceptedRowCount != 0 || counts.RejectedRowCount != 1 || !reflect.DeepEqual(counts.Diagnostics, diagnostics) {
				t.Fatalf("preview diverged: %v %#v", err, counts)
			}
			for _, cap := range []int64{1, 2, 10} {
				limits.MaxRejectedRowDiagnostics = cap
				capability := imports.ImportSourceCapability{SourceStreamRef: "impsrc_" + strings.Repeat("1", 32), SourceContentSHA256: parsed.SourceContentSHA256, SourceMediaType: "text/csv", SourceByteSize: int64(len(source))}
				stream := &applySourceFixture{source: source, capability: capability}
				f := &importFacade{sourceStore: stream, limits: limits}
				request := imports.ExtensionImportApplyRequest{SourceCapability: capability, ExpectedSourceContentSHA256: parsed.SourceContentSHA256, MappingFingerprint: fingerprint}
				var previous importValidationSummary
				for pass := 0; pass < 2; pass++ {
					summary, err := f.validateImportPass(context.Background(), nil, request, m, pass == 0, func(normalizedRow) error { t.Fatal("rejected observation reached accepted-row visitor"); return nil })
					if err != nil || summary.Accepted != 0 || summary.Rejected != 1 || summary.DiagnosticCount != int64(len(diagnostics)) {
						t.Fatalf("apply: %v %#v", err, summary)
					}
					if pass == 0 && !reflect.DeepEqual(summary.Diagnostics, diagnostics[:min(int(cap), len(diagnostics))]) {
						t.Fatal("retained diagnostic prefix changed")
					}
					if pass == 1 && !summary.agrees(previous) {
						t.Fatal("apply validation transcripts differ")
					}
					previous = summary
				}
				if stream.opens != 2 || stream.closes != 2 {
					t.Fatal("source stream leaked")
				}
			}
		})
	}
}
