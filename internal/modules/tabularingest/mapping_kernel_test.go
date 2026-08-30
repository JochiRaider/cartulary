package tabularingest_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
)

func TestMappingKernelV1TransformsPoliciesOrderingAndProvenance(t *testing.T) {
	firstField := "field.first"
	secondField := "field.second"
	transform := tabularingest.TransformSplitDelimitedV1
	request := tabularingest.MappingKernelRequestV1{
		TargetFields: []tabularingest.MappingKernelTargetFieldV1{
			{FieldKey: firstField, Order: 1, CreateWritable: true, Clearable: true},
			{FieldKey: secondField, Order: 2, Writable: true, Clearable: true},
		},
		SourceColumns: []tabularingest.MappingKernelSourceColumnV1{
			{
				SourceColumnOrdinal: 1, SourceHeaderText: `Second "<>&`, FieldKey: &secondField,
				TransformID: &transform, TransformOptions: map[string]any{"delimiter": ",", "trim_items": true, "drop_empty_items": true},
				EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1,
			},
			{
				SourceColumnOrdinal: 2, SourceHeaderText: "First", FieldKey: &firstField,
				TransformOptions: map[string]any{}, EmptyValuePolicy: tabularingest.EmptyValueWriteNullV1,
			},
			{
				SourceColumnOrdinal: 3, SourceHeaderText: "Unknown 界",
				TransformOptions: map[string]any{}, EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1,
			},
		},
		Rows: []tabularingest.MappingKernelSourceRowV1{
			{SourceRowOrdinal: 4, Cells: []tabularingest.MappingKernelScalarCellV1{
				{SourceColumnOrdinal: 1, RawValue: " alpha, ,beta ", CellKind: "string", Classification: tabularingest.MappingKernelCellScalarV1, Present: true},
				{SourceColumnOrdinal: 2, CellKind: "blank", Classification: tabularingest.MappingKernelCellEmptyV1, Present: true},
				{SourceColumnOrdinal: 3, RawValue: "raw\u2028value", CellKind: "inline_string", Classification: tabularingest.MappingKernelCellScalarV1, Present: true},
			}},
		},
		UnknownColumnPolicy: tabularingest.UnknownColumnPreserveRawCaptureV1,
	}
	plan, err := tabularingest.BuildMappingKernelPlanV1(context.Background(), request)
	if err != nil {
		t.Fatalf("build mapping kernel plan: %v", err)
	}
	if plan.ContractID != tabularingest.MappingKernelContractV1 || len(plan.Rows) != 1 || len(plan.Rows[0].Values) != 2 {
		t.Fatalf("unexpected kernel plan: %#v", plan)
	}
	values := plan.Rows[0].Values
	if values[0].FieldKey != secondField || values[0].TransformedValue != "alpha,beta" || values[0].Disposition != tabularingest.MappingKernelValueV1 {
		t.Fatalf("source-ordered split result changed: %#v", values[0])
	}
	if values[1].FieldKey != firstField || values[1].Disposition != tabularingest.MappingKernelNullV1 {
		t.Fatalf("write-null disposition changed: %#v", values[1])
	}
	if got := plan.Rows[0].UnknownValues; len(got) != 1 || got[0].SourceColumnOrdinal != 3 || got[0].SourceHeaderText != "Unknown 界" || got[0].CellKind != "inline_string" {
		t.Fatalf("unknown provenance changed: %#v", got)
	}
	if got := plan.Warnings; len(got) != 1 || got[0].Code != tabularingest.MappingKernelWarningUnmappedV1 || got[0].SourceRowOrdinal != 4 || got[0].SourceColumnOrdinal != 3 {
		t.Fatalf("warning order changed: %#v", got)
	}

	request.SourceColumns[0].TransformOptions["trim_items"] = false
	request.Rows[0].Cells[0].RawValue = "mutated"
	if values[0].TransformedValue != "alpha,beta" || values[0].RawValue != " alpha, ,beta " {
		t.Fatalf("kernel result retained mutable request state: %#v", values[0])
	}
}

func TestMappingKernelV1ClosedTransformRegistry(t *testing.T) {
	transforms := []struct {
		name    string
		id      string
		options map[string]any
		raw     string
		want    string
	}{
		{name: "trim", id: tabularingest.TransformTrimV1, options: map[string]any{}, raw: "  Alpha  ", want: "Alpha"},
		{name: "collapse", id: tabularingest.TransformCollapseWhitespaceV1, options: map[string]any{}, raw: " Alpha\t 界\nBeta ", want: "Alpha 界 Beta"},
		{name: "lowercase", id: tabularingest.TransformLowercaseV1, options: map[string]any{}, raw: "ÄLPHA", want: "älpha"},
		{name: "comma defaults false", id: tabularingest.TransformSplitDelimitedV1, options: map[string]any{"delimiter": ","}, raw: " a, ,b ", want: " a, ,b "},
		{name: "semicolon", id: tabularingest.TransformSplitDelimitedV1, options: map[string]any{"delimiter": ";"}, raw: "a;b", want: "a;b"},
		{name: "pipe", id: tabularingest.TransformSplitDelimitedV1, options: map[string]any{"delimiter": "|"}, raw: "a|b", want: "a|b"},
		{name: "newline", id: tabularingest.TransformSplitDelimitedV1, options: map[string]any{"delimiter": "\n"}, raw: "a\nb", want: "a\nb"},
		{name: "tab", id: tabularingest.TransformSplitDelimitedV1, options: map[string]any{"delimiter": "\t"}, raw: "a\tb", want: "a\tb"},
	}
	for _, testCase := range transforms {
		t.Run(testCase.name, func(t *testing.T) {
			request := oneColumnKernelRequest(testCase.id, testCase.options, testCase.raw)
			plan, err := tabularingest.BuildMappingKernelPlanV1(context.Background(), request)
			if err != nil {
				t.Fatalf("build transform plan: %v", err)
			}
			if got := plan.Rows[0].Values[0].TransformedValue; got != testCase.want {
				t.Fatalf("transformed value = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestMappingKernelV1FailsClosed(t *testing.T) {
	base := oneColumnKernelRequest(tabularingest.TransformTrimV1, map[string]any{}, " value ")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tabularingest.BuildMappingKernelPlanV1(canceled, base); !errors.Is(err, tabularingest.ErrMappingKernelCanceled) {
		t.Fatalf("cancellation error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*tabularingest.MappingKernelRequestV1)
		match  string
	}{
		{name: "unknown policy", mutate: func(request *tabularingest.MappingKernelRequestV1) { request.UnknownColumnPolicy = "future" }, match: "invalid unknown-column policy"},
		{name: "row width", mutate: func(request *tabularingest.MappingKernelRequestV1) { request.Rows[0].Cells = nil }, match: "has 0 cells"},
		{name: "cell classification", mutate: func(request *tabularingest.MappingKernelRequestV1) {
			request.Rows[0].Cells[0].Classification = "parser_object"
		}, match: "invalid classification"},
		{name: "noncontiguous column", mutate: func(request *tabularingest.MappingKernelRequestV1) { request.SourceColumns[0].SourceColumnOrdinal = 2 }, match: "not contiguous"},
		{name: "invalid transform options", mutate: func(request *tabularingest.MappingKernelRequestV1) {
			request.SourceColumns[0].TransformOptions["delimiter"] = ","
		}, match: "does not accept options"},
		{name: "not clearable", mutate: func(request *tabularingest.MappingKernelRequestV1) {
			request.SourceColumns[0].EmptyValuePolicy = tabularingest.EmptyValueWriteNullV1
			request.TargetFields[0].Clearable = false
		}, match: "not clearable"},
		{name: "reject unmapped", mutate: func(request *tabularingest.MappingKernelRequestV1) {
			request.SourceColumns[0].FieldKey = nil
			request.SourceColumns[0].TransformID = nil
			request.SourceColumns[0].TransformOptions = map[string]any{}
			request.UnknownColumnPolicy = tabularingest.UnknownColumnRejectIfUnmappedV1
		}, match: "rejects unmapped"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			request := oneColumnKernelRequest(tabularingest.TransformTrimV1, map[string]any{}, " value ")
			testCase.mutate(&request)
			plan, err := tabularingest.BuildMappingKernelPlanV1(context.Background(), request)
			if err == nil || !strings.Contains(err.Error(), testCase.match) {
				t.Fatalf("error = %v, plan = %#v, want match %q", err, plan, testCase.match)
			}
			if !reflect.DeepEqual(plan, tabularingest.MappingKernelPlanV1{}) {
				t.Fatalf("failure returned partial plan: %#v", plan)
			}
		})
	}
}

func TestMappingKernelV1HasNoLifecycleOrInfrastructureDependencies(t *testing.T) {
	body, err := os.ReadFile("mapping_kernel.go")
	if err != nil {
		t.Fatalf("read mapping kernel: %v", err)
	}
	content := string(body)
	for _, forbidden := range []string{
		"internal/modules/imports",
		"internal/modules/revisions",
		"internal/modules/projections",
		"internal/modules/collaboration",
		"internal/platform/authn",
		"internal/platform/httpapi",
		"internal/platform/viewschema",
		"pgx.",
		"uuid.UUID",
		"http.Request",
		"csv.Reader",
		"xlsx",
		"workbook",
		"network_flow",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("mapping kernel contains forbidden dependency %q", forbidden)
		}
	}
}

func oneColumnKernelRequest(transformID string, options map[string]any, raw string) tabularingest.MappingKernelRequestV1 {
	fieldKey := "field.value"
	return tabularingest.MappingKernelRequestV1{
		TargetFields: []tabularingest.MappingKernelTargetFieldV1{{FieldKey: fieldKey, Order: 1, Writable: true, CreateWritable: true, Clearable: true}},
		SourceColumns: []tabularingest.MappingKernelSourceColumnV1{{
			SourceColumnOrdinal: 1, SourceHeaderText: "Value", FieldKey: &fieldKey,
			TransformID: &transformID, TransformOptions: options, EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1,
		}},
		Rows: []tabularingest.MappingKernelSourceRowV1{{SourceRowOrdinal: 1, Cells: []tabularingest.MappingKernelScalarCellV1{{
			SourceColumnOrdinal: 1, RawValue: raw, CellKind: "string", Classification: tabularingest.MappingKernelCellScalarV1, Present: true,
		}}}},
		UnknownColumnPolicy: tabularingest.UnknownColumnPreserveRawCaptureV1,
	}
}
