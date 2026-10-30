package imports

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestGenericCSVRowsAreSparse(t *testing.T) {
	projection, err := materializeGenericCSV(context.Background(), strings.NewReader("a,b,c,d\none\ntwo,\n"))
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []int{1, 2} {
		if got := len(projection.SourceRows[i]["cells"].([]map[string]any)); got != want {
			t.Fatalf("row %d has %d cells, want %d actual source fields", i, got, want)
		}
	}
}

func TestGenericCSVPreviewGeometryAndFailures(t *testing.T) {
	for _, rows := range []int{0, 1, 50, 51} {
		preview, err := analyzeGenericCSV(t.Context(), strings.NewReader("a,b\n"+strings.Repeat("one\n", rows)))
		if err != nil {
			t.Fatal(err)
		}
		if preview.InferredRowCount != rows || preview.InferredColumnCount != 2 || len(preview.PreviewRows) != min(rows, 50) {
			t.Fatalf("geometry: %#v", preview)
		}
		for i, row := range preview.PreviewRows {
			if row["source_row_ref"] != i+2 || len(row["cells"].([]map[string]any)) != 1 {
				t.Fatalf("row: %#v", row)
			}
		}
	}
	preview, err := analyzeGenericCSV(t.Context(), strings.NewReader("a\n"+strings.Repeat("x\n", 50)+"x,y,z\n"))
	if err != nil || preview.InferredColumnCount != 3 || preview.Columns[2]["source_header_text"] != nil || len(preview.PreviewRows) != 50 {
		t.Fatalf("late width: %#v %v", preview, err)
	}
	for _, source := range []string{"", "a\n" + strings.Repeat("x\n", 50) + "\"unclosed"} {
		preview, err := analyzeGenericCSV(t.Context(), strings.NewReader(source))
		if !errors.Is(err, tabularingest.ErrInvalidDelimited) || !reflect.DeepEqual(preview, genericCSVPreview{}) {
			t.Fatalf("invalid source published preview: %#v %v", preview, err)
		}
	}
	failure := errors.New("source integrity failure")
	preview, err = analyzeGenericCSV(t.Context(), io.MultiReader(strings.NewReader("a\nx\n"), genericFailingReader{failure}))
	if !errors.Is(err, failure) || !reflect.DeepEqual(preview, genericCSVPreview{}) {
		t.Fatalf("source failure: %#v %v", preview, err)
	}
}

type genericFailingReader struct{ err error }

func (r genericFailingReader) Read([]byte) (int, error) { return 0, r.err }

type genericCancelReader struct {
	io.Reader
	cancel context.CancelFunc
	read   int
}

func (r *genericCancelReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.read += n
	if r.read >= 128*1024 {
		r.cancel()
	}
	return n, err
}

func TestGenericCSVCancellation(t *testing.T) {
	for _, tail := range []bool{false, true} {
		for _, quoted := range []bool{false, true} {
			source := "a\n"
			if tail {
				source += strings.Repeat("x\n", 50)
			}
			value := strings.Repeat("x", 512*1024)
			if quoted {
				value = "\"" + value + "\""
			}
			source += value + "\n"
			ctx, cancel := context.WithCancel(t.Context())
			reader := &genericCancelReader{Reader: strings.NewReader(source), cancel: cancel}
			result, err := analyzeGenericCSV(ctx, reader)
			cancel()
			if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, genericCSVPreview{}) {
				t.Fatalf("tail=%v quoted=%v: %v", tail, quoted, err)
			}
			if reader.read > 128*1024 {
				t.Fatalf("read beyond cancellation boundary: %d", reader.read)
			}
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	result, err := scanGenericCSV(ctx, strings.NewReader("a\nx\ny\n"), func(map[string]any) { cancel() })
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, genericCSVPreview{}) {
		t.Fatalf("materialization cancellation: %#v %v", result, err)
	}
}

func TestGenericCSVRaggedResourceBudget(t *testing.T) {
	source := strings.Repeat("h,", 9999) + "h\n" + strings.Repeat("x\n", 10000)
	measure := func(name string, run func()) {
		t.Helper()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		run()
		runtime.ReadMemStats(&after)
		allocated := after.TotalAlloc - before.TotalAlloc
		t.Logf("%s allocated %d bytes", name, allocated)
		if allocated > 64*1024*1024 {
			t.Fatalf("%s exceeded 64 MiB isolated allocation budget: %d", name, allocated)
		}
	}
	measure("preview", func() {
		preview, err := analyzeGenericCSV(t.Context(), strings.NewReader(source))
		if err != nil {
			t.Fatal(err)
		}
		if preview.InferredRowCount != 10000 || preview.InferredColumnCount != 10000 || len(preview.PreviewRows) != 50 {
			t.Fatal("wrong preview dimensions")
		}
		for _, row := range preview.PreviewRows {
			if len(row["cells"].([]map[string]any)) != 1 {
				t.Fatal("preview padded cells")
			}
		}
	})
	measure("approval", func() {
		projection, err := materializeGenericCSV(t.Context(), strings.NewReader(source))
		if err != nil {
			t.Fatal(err)
		}
		if len(projection.SourceRows) != 10000 {
			t.Fatal("wrong source count")
		}
		for _, row := range projection.SourceRows {
			if len(row["cells"].([]map[string]any)) != 1 {
				t.Fatal("projection padded cells")
			}
		}
	})
}

func TestGenericCSVPreviewRetentionBudget(t *testing.T) {
	// Keep these measurements sequential: TotalAlloc includes the whole process.
	// Build fixtures/readers and check results outside the measured calls.
	const budget = 1024 * 1024
	for _, test := range []struct {
		name string
		rows int
		rect string
	}{{"10000_rows", 10000, "A1:A10001"}, {"100000_rows", 100000, "A1:A100001"}} {
		t.Run(test.name, func(t *testing.T) {
			source := genericCSVNarrowSource(test.rows)
			reader, ctx := strings.NewReader(source), t.Context()
			var preview genericCSVPreview
			var err error
			allocated := measureGenericCSVAllocations(t, "preview", func() { preview, err = analyzeGenericCSV(ctx, reader) })
			if err != nil {
				t.Fatal(err)
			}
			if allocated > budget {
				t.Fatalf("preview exceeded 1 MiB retention budget: %d", allocated)
			}
			if preview.InferredRowCount != test.rows || preview.InferredColumnCount != 1 || preview.SourceRectA1 != test.rect || preview.HeaderRowRef != 1 || preview.DataStartRowRef != 2 {
				t.Fatalf("wrong source geometry: %#v", preview)
			}
			if len(preview.Columns) != 1 || preview.Columns[0]["source_header_text"] != "h" || len(preview.PreviewRows) != 50 {
				t.Fatalf("wrong retained preview: %#v", preview)
			}
			for i, row := range preview.PreviewRows {
				cells := row["cells"].([]map[string]any)
				if row["source_row_ref"] != i+2 || len(cells) != 1 || cells[0]["source_column_ordinal"] != 1 || cells[0]["display_text"] != "preview" || cells[0]["cell_kind"] != "string" {
					t.Fatalf("wrong preview row %d: %#v", i, row)
				}
			}
			if test.rows == 10000 {
				reader := strings.NewReader(source)
				var projection discoveredUnit
				allocated := measureGenericCSVAllocations(t, "full materialization control", func() { projection, err = materializeGenericCSV(ctx, reader) })
				if err != nil || len(projection.SourceRows) != test.rows {
					t.Fatalf("invalid full materialization control: rows=%d err=%v", len(projection.SourceRows), err)
				}
				if allocated <= budget {
					t.Fatalf("1 MiB bound would admit full materialization: %d", allocated)
				}
			}
		})
	}
}

func TestGenericCSVPreviewTailValidation(t *testing.T) {
	source := genericCSVNarrowSource(100000)
	t.Run("late width", func(t *testing.T) {
		preview, err := analyzeGenericCSV(t.Context(), strings.NewReader(source+"last,late\n"))
		if err != nil {
			t.Fatal(err)
		}
		if preview.InferredRowCount != 100001 || preview.InferredColumnCount != 2 || preview.SourceRectA1 != "A1:B100002" || len(preview.Columns) != 2 || preview.Columns[1]["source_header_text"] != nil || len(preview.PreviewRows) != 50 {
			t.Fatalf("discarded tail width was lost: %#v", preview)
		}
		for _, row := range preview.PreviewRows {
			if len(row["cells"].([]map[string]any)) != 1 {
				t.Fatal("late width padded preview")
			}
		}
	})
	failure := errors.New("terminal source integrity failure")
	for _, test := range []struct {
		name   string
		reader io.Reader
		want   error
	}{
		{"malformed final quote", strings.NewReader(source + "\"unclosed"), tabularingest.ErrInvalidDelimited},
		{"terminal integrity failure", io.MultiReader(strings.NewReader(source), genericFailingReader{failure}), failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			preview, err := analyzeGenericCSV(t.Context(), test.reader)
			if !errors.Is(err, test.want) || !reflect.DeepEqual(preview, genericCSVPreview{}) {
				t.Fatalf("discarded tail failure published preview: %#v %v", preview, err)
			}
		})
	}
	t.Run("cancellation in discarded rows", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		reader := &genericCancelReader{Reader: strings.NewReader(source), cancel: cancel}
		preview, err := analyzeGenericCSV(ctx, reader)
		if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(preview, genericCSVPreview{}) {
			t.Fatalf("canceled tail published preview: %#v %v", preview, err)
		}
		if reader.read != 128*1024 {
			t.Fatalf("wrong cancellation read boundary: %d", reader.read)
		}
	})
}

func genericCSVNarrowSource(rows int) string {
	return "h\n" + strings.Repeat("preview\n", 50) + strings.Repeat("tail\n", rows-50)
}

func measureGenericCSVAllocations(t *testing.T, label string, run func()) uint64 {
	t.Helper()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	run()
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("%s allocated %d bytes", label, allocated)
	return allocated
}

func TestGenericCSVAbsentMapping(t *testing.T) {
	title, body, mode := "host.display_name", "host.location", "entity_origin"
	unit := applyUnitData{SourceFileKind: sourceFileKindCSV, approvedMapping: approvedMapping{
		TargetViewSchemaID: "cartulary.view.hosts.v1", UnknownColumnPolicy: "preserve_raw_capture",
		SourceColumns: []sourceColumnMapping{
			{SourceColumnOrdinal: 1, SourceHeaderText: "title", FieldKey: &title, EntityBindingMode: &mode, TransformOptions: map[string]any{}, EmptyValuePolicy: "omit_field"},
			{SourceColumnOrdinal: 2, SourceHeaderText: "body", FieldKey: &body, EntityBindingMode: &mode, TransformOptions: map[string]any{}, EmptyValuePolicy: "write_null"},
			{SourceColumnOrdinal: 3, SourceHeaderText: "unknown", TransformOptions: map[string]any{}, EmptyValuePolicy: "omit_field"},
		},
	}}
	for _, explicit := range []bool{false, true} {
		source := "title,body,unknown\none"
		if explicit {
			source += ",,"
		}
		projection, err := materializeGenericCSV(t.Context(), strings.NewReader(source))
		if err != nil {
			t.Fatal(err)
		}
		request, err := importMappingKernelRequest(unit, projection.SourceRows[0], 2)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := tabularingest.BuildMappingKernelPlanV1(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		row := plan.Rows[0]
		if explicit {
			if len(row.Values) != 2 || row.Values[1].Disposition != "null" || len(row.UnknownValues) != 1 || len(plan.Warnings) != 1 {
				t.Fatalf("explicit empties: %#v", plan)
			}
		} else if len(row.Values) != 1 || len(row.UnknownValues) != 0 || len(plan.Warnings) != 0 {
			t.Fatalf("absent fields: %#v", plan)
		}
	}
}
