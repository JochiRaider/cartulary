package imports

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
)

func assertFileMappingCompatibilityCharacterization(t testing.TB) {
	t.Helper()

	discovered := []map[string]any{
		{"source_column_ordinal": 1, "source_header_text": "Náme"},
		{"source_column_ordinal": 2, "source_header_text": "Aliases"},
		{"source_column_ordinal": 3, "source_header_text": "Location"},
		{"source_column_ordinal": 4, "source_header_text": `Unknown "<>&`},
	}
	body := `{
		"client_txn_id":"txn-file-map-界",
		"target_view_schema_id":"cartulary.view.hosts.v1",
		"header_row_ref":1,
		"data_start_row_ref":2,
		"unknown_column_policy":"preserve_raw_capture",
		"source_columns":[
			{"source_column_ordinal":1,"source_header_text":"Náme","field_key":"host.display_name","entity_binding_mode":"entity_origin","transform_id":"trim_v1","transform_options":{},"empty_value_policy":"omit_field"},
			{"source_column_ordinal":2,"source_header_text":"Aliases","field_key":"host.aliases","entity_binding_mode":"entity_origin","transform_id":"split_delimited_v1","transform_options":{"delimiter":";"},"empty_value_policy":"omit_field"},
			{"source_column_ordinal":3,"source_header_text":"Location","field_key":"host.location","entity_binding_mode":"entity_origin","transform_id":"lowercase_v1","transform_options":{},"empty_value_policy":"write_null"},
			{"source_column_ordinal":4,"source_header_text":"Unknown \"<>&","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"}
		]
	}`
	request, apiErr := DecodeMappingRequest(strings.NewReader(body), discovered)
	if apiErr != nil {
		t.Fatalf("decode file mapping characterization: %#v", apiErr)
	}
	approvedBytes := mappingJSONForReadiness(request.ApprovedMapping)
	const approvedGolden = `{"target_view_schema_id":"cartulary.view.hosts.v1","unknown_column_policy":"preserve_raw_capture","source_columns":[{"source_column_ordinal":1,"source_header_text":"Náme","field_key":"host.display_name","entity_binding_mode":"entity_origin","transform_id":"trim_v1","transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":2,"source_header_text":"Aliases","field_key":"host.aliases","entity_binding_mode":"entity_origin","transform_id":"split_delimited_v1","transform_options":{"delimiter":";"},"empty_value_policy":"omit_field"},{"source_column_ordinal":3,"source_header_text":"Location","field_key":"host.location","entity_binding_mode":"entity_origin","transform_id":"lowercase_v1","transform_options":{},"empty_value_policy":"write_null"},{"source_column_ordinal":4,"source_header_text":"Unknown \"\u003c\u003e\u0026","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"}]}`
	const normalizedGolden = `{"client_txn_id":"txn-file-map-界","mapping":{"data_start_row_ref":2,"header_row_ref":1,"source_columns":[{"source_column_ordinal":1,"source_header_text":"Náme","field_key":"host.display_name","entity_binding_mode":"entity_origin","transform_id":"trim_v1","transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":2,"source_header_text":"Aliases","field_key":"host.aliases","entity_binding_mode":"entity_origin","transform_id":"split_delimited_v1","transform_options":{"delimiter":";"},"empty_value_policy":"omit_field"},{"source_column_ordinal":3,"source_header_text":"Location","field_key":"host.location","entity_binding_mode":"entity_origin","transform_id":"lowercase_v1","transform_options":{},"empty_value_policy":"write_null"},{"source_column_ordinal":4,"source_header_text":"Unknown \"\u003c\u003e\u0026","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"}],"target_view_schema_id":"cartulary.view.hosts.v1","unknown_column_policy":"preserve_raw_capture"}}`
	const fingerprintGolden = "8b046f4cace5f6a164e3e752ceceee0d1897e31dd200a8a5d2549ab173df6234"
	if string(approvedBytes) != approvedGolden || string(request.Normalized) != normalizedGolden || request.Fingerprint != fingerprintGolden {
		t.Fatalf(
			"file mapping compatibility changed:\napproved=%s\nnormalized=%s\nfingerprint=%s",
			approvedBytes,
			request.Normalized,
			request.Fingerprint,
		)
	}
	replay, replayErr := DecodeMappingRequest(
		strings.NewReader(strings.Replace(body, "txn-file-map-界", "txn-file-map-replay", 1)),
		discovered,
	)
	if replayErr != nil || replay.Fingerprint != request.Fingerprint {
		t.Fatalf("client transaction identity changed file mapping bytes: replay=%#v error=%#v", replay, replayErr)
	}
	if request.ApprovedMapping.SourceColumns[1].TransformOptions["trim_items"] != nil ||
		request.ApprovedMapping.SourceColumns[1].TransformOptions["drop_empty_items"] != nil {
		t.Fatalf("omitted split booleans acquired a non-false representation: %#v", request.ApprovedMapping.SourceColumns[1])
	}
	transformed, err := mappingKernelTransformCompatibility(
		" alpha ; ; beta ",
		request.ApprovedMapping.SourceColumns[1],
	)
	if err != nil || transformed != " alpha ; ; beta " {
		t.Fatalf("omitted split booleans no longer behave as false: value=%q error=%v", transformed, err)
	}

	assertImportTransformCompatibility(t)
	assertCrossPathMappingKernelCompatibility(t)
	assertImportOwnerRequestCompatibility(t)
	assertImportMappingFailureCompatibility(t, discovered)
}

func assertCrossPathMappingKernelCompatibility(t testing.TB) {
	t.Helper()
	title := "note.title"
	body := "note.body"
	unit := ApplyUnitData{
		ApprovedMapping: ApprovedMapping{
			TargetViewSchemaID:  "cartulary.view.notes.v1",
			UnknownColumnPolicy: tabularingest.UnknownColumnPreserveRawCaptureV1,
			SourceColumns: []SourceColumnMapping{
				{SourceColumnOrdinal: 1, SourceHeaderText: "Title", FieldKey: &title, TransformOptions: map[string]any{}, EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1},
				{SourceColumnOrdinal: 2, SourceHeaderText: "Body", FieldKey: &body, TransformOptions: map[string]any{}, EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1},
				{SourceColumnOrdinal: 3, SourceHeaderText: "Unknown 界", TransformOptions: map[string]any{}, EmptyValuePolicy: tabularingest.EmptyValueOmitFieldV1},
			},
		},
	}
	sourceRow := map[string]any{
		"source_row_ref": 2,
		"cells": []any{
			map[string]any{"source_column_ordinal": 1, "display_text": "Title", "cell_kind": "string"},
			map[string]any{"source_column_ordinal": 2, "display_text": "Body\u2028界", "cell_kind": "inline_string"},
			map[string]any{"source_column_ordinal": 3, "display_text": "raw", "cell_kind": "string"},
		},
	}
	kernelRequest, err := importMappingKernelRequest(unit, sourceRow, 2)
	if err != nil {
		t.Fatalf("build file kernel request: %v", err)
	}
	filePlan, err := tabularingest.BuildMappingKernelPlanV1(context.Background(), kernelRequest)
	if err != nil {
		t.Fatalf("build file kernel plan: %v", err)
	}
	clipboardPlan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
		ViewSchemaID: "cartulary.view.notes.v1", ClientTxnID: "txn-clipboard-cross-path",
		SourceKind: "clipboard_paste", Text: "Title\tBody\u2028界\traw", Format: "tsv",
		StartFieldKey: "note.title", Columns: []string{"note.title", "note.body"},
		SourceHeaders: []any{"Title", "Body", "Unknown 界"}, RequireTargets: -1,
	})
	if err != nil {
		t.Fatalf("build clipboard plan: %v", err)
	}
	fileRow := filePlan.Rows[0]
	clipboardRow := clipboardPlan.Rows[0]
	if len(fileRow.Values) != len(clipboardRow.Cells) || len(fileRow.UnknownValues) != len(clipboardRow.Unmapped) {
		t.Fatalf("cross-path plan widths differ: file=%#v clipboard=%#v", fileRow, clipboardRow)
	}
	for index, value := range fileRow.Values {
		cell := clipboardRow.Cells[index]
		if value.FieldKey != cell.FieldKey || value.RawValue != cell.RawValue || value.SourceColumnOrdinal != cell.SourceColumnOrdinal || value.EntityBindingMode != cell.EntityBindingMode {
			t.Fatalf("cross-path mapped value %d differs: file=%#v clipboard=%#v", index, value, cell)
		}
	}
	fileUnknown := fileRow.UnknownValues[0]
	clipboardUnknown := clipboardRow.Unmapped[0]
	if fileUnknown.SourceColumnOrdinal != clipboardUnknown.SourceColumnOrdinal ||
		fileUnknown.SourceHeaderText != clipboardUnknown.SourceHeaderText ||
		fileUnknown.RawValue != clipboardUnknown.RawValue {
		t.Fatalf("cross-path unknown provenance differs: file=%#v clipboard=%#v", fileUnknown, clipboardUnknown)
	}
	if clipboardUnknown.SourceKind != "clipboard_paste" || clipboardUnknown.SourceClientTxnID != "txn-clipboard-cross-path" {
		t.Fatalf("clipboard lifecycle provenance was not retained: %#v", clipboardUnknown)
	}
}

func assertImportTransformCompatibility(t testing.TB) {
	t.Helper()
	value := func(input string, transformID *string, options map[string]any) string {
		t.Helper()
		result, err := mappingKernelTransformCompatibility(input, SourceColumnMapping{
			TransformID:      transformID,
			TransformOptions: options,
		})
		if err != nil {
			t.Fatalf("transform %v: %v", transformID, err)
		}
		return result
	}
	transform := func(id string) *string { return &id }
	if value("  A  ", nil, map[string]any{}) != "  A  " ||
		value("  A  ", transform("trim_v1"), map[string]any{}) != "A" ||
		value(" A\t B\nC ", transform("collapse_whitespace_v1"), map[string]any{}) != "A B C" ||
		value("ÄBC界", transform("lowercase_v1"), map[string]any{}) != "äbc界" {
		t.Fatal("scalar transform compatibility changed")
	}
	for _, delimiter := range []string{",", ";", "|", "\n", "\t"} {
		input := " alpha " + delimiter + " " + delimiter + " beta "
		if got := value(input, transform("split_delimited_v1"), map[string]any{
			"delimiter": delimiter, "trim_items": true, "drop_empty_items": true,
		}); got != "alpha"+delimiter+"beta" {
			t.Fatalf("split delimiter %q changed: %q", delimiter, got)
		}
	}
}

func mappingKernelTransformCompatibility(input string, column SourceColumnMapping) (string, error) {
	fieldKey := "compatibility.field"
	column.FieldKey = &fieldKey
	column.SourceColumnOrdinal = 1
	column.EmptyValuePolicy = tabularingest.EmptyValueOmitFieldV1
	plan, err := tabularingest.BuildMappingKernelPlanV1(context.Background(), tabularingest.MappingKernelRequestV1{
		TargetFields: []tabularingest.MappingKernelTargetFieldV1{{
			FieldKey: fieldKey, Order: 1, Writable: true, CreateWritable: true, Clearable: true,
		}},
		SourceColumns: []tabularingest.MappingKernelSourceColumnV1{{
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			FieldKey:            column.FieldKey,
			TransformID:         column.TransformID,
			TransformOptions:    column.TransformOptions,
			EmptyValuePolicy:    column.EmptyValuePolicy,
		}},
		Rows: []tabularingest.MappingKernelSourceRowV1{{
			SourceRowOrdinal: 1,
			Cells: []tabularingest.MappingKernelScalarCellV1{{
				SourceColumnOrdinal: 1,
				RawValue:            input,
				CellKind:            "string",
				Classification:      tabularingest.MappingKernelCellScalarV1,
				Present:             true,
			}},
		}},
		UnknownColumnPolicy: tabularingest.UnknownColumnPreserveRawCaptureV1,
	})
	if err != nil {
		return "", err
	}
	return plan.Rows[0].Values[0].TransformedValue, nil
}

func assertImportOwnerRequestCompatibility(t testing.TB) {
	t.Helper()
	type ownerCase struct {
		viewSchemaID string
		fieldKey     string
		bindingMode  *string
		raw          string
	}
	mentionOrigin := "mention_origin"
	entityOrigin := "entity_origin"
	cases := []ownerCase{
		{viewSchemaID: "cartulary.view.timeline.v2", fieldKey: "timeline.host_refs", bindingMode: &mentionOrigin, raw: "  Gateway   One  "},
		{viewSchemaID: "cartulary.view.hosts.v1", fieldKey: "host.display_name", bindingMode: &entityOrigin, raw: " Gateway One "},
		{viewSchemaID: "cartulary.view.identities.v1", fieldKey: "identity.display_name", bindingMode: &entityOrigin, raw: " Analyst One "},
		{viewSchemaID: "cartulary.view.notes.v1", fieldKey: "note.title", raw: " Investigation note "},
	}
	snapshots := make([]map[string]any, 0, len(cases))
	for _, testCase := range cases {
		normalize := func(fieldKey string, raw string, policy string) (ownerfacade.ImportScalarValue, bool, error) {
			if testCase.viewSchemaID == "cartulary.view.timeline.v2" {
				return ownerfacade.NewCollectionTokenImportScalar(ownerfacade.ImportCollectionToken{
					RawText: raw, NormalizedText: strings.Join(strings.Fields(raw), " "),
				}), true, nil
			}
			return ownerfacade.NormalizeImportScalar(testCase.viewSchemaID, fieldKey, raw, policy)
		}
		facade, err := ownerfacade.NewImportOwnerCreateFacadeWithNormalizer(
			ownerfacade.ImportOwnerCreateBinding{TargetViewSchemaID: testCase.viewSchemaID, FacadeID: "characterization"},
			normalize,
			func(context.Context, pgx.Tx, ownerfacade.ImportOwnerCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
				return ownerfacade.ImportOwnerCreateResponse{}, nil
			},
		)
		if err != nil {
			t.Fatalf("build characterization facade: %v", err)
		}
		fieldKey := testCase.fieldKey
		request, err := importOwnerCreateRequest(
			context.Background(),
			ApplyStartResult{
				IncidentID:      uuid.MustParse("11111111-1111-4111-8111-111111111111"),
				ImportSessionID: uuid.MustParse("22222222-2222-4222-8222-222222222222"),
				ClientTxnID:     "txn-apply",
			},
			ApplyUnitData{
				UnitID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
				ApprovedMapping: ApprovedMapping{
					TargetViewSchemaID:  testCase.viewSchemaID,
					UnknownColumnPolicy: "preserve_raw_capture",
					SourceColumns: []SourceColumnMapping{
						{SourceColumnOrdinal: 1, SourceHeaderText: "Known", FieldKey: &fieldKey, EntityBindingMode: testCase.bindingMode, TransformOptions: map[string]any{}, EmptyValuePolicy: "omit_field"},
						{SourceColumnOrdinal: 2, SourceHeaderText: "Unknown", TransformOptions: map[string]any{}, EmptyValuePolicy: "omit_field"},
					},
				},
				MappingFingerprint: strings.Repeat("a", 64),
				SourceFileKind:     "csv", SourceContentSHA256: strings.Repeat("b", 64),
				ParserProfileID: "cartulary.import.phase2_workbook_import.v1", ParserVersion: "phase11_import_adapter_v1",
				LocatorKind: "csv_file", Locator: "file", SourceRectA1: "A1:B2",
			},
			uuid.MustParse("44444444-4444-4444-8444-444444444444"),
			map[string]any{
				"source_row_ref": 2,
				"cells": []any{
					map[string]any{"source_column_ordinal": 1, "display_text": testCase.raw, "cell_kind": "inline_string"},
					map[string]any{"source_column_ordinal": 2, "display_text": "raw\u2028value", "cell_kind": "inline_string"},
				},
			},
			2,
			"import:row:2",
			facade,
		)
		if err != nil {
			t.Fatalf("build owner request for %s: %v", testCase.viewSchemaID, err)
		}
		field := request.FieldValues[0]
		snapshots = append(snapshots, map[string]any{
			"view_schema_id":      request.TargetViewSchemaID,
			"field_key":           field.FieldKey,
			"binding_mode":        field.EntityBindingMode,
			"raw_value":           field.RawValue,
			"normalized":          importScalarCompatibilityValue(field.NormalizedValue),
			"unknown":             request.UnknownValues,
			"source_row_ref":      request.SourceRowProvenance.SourceRowRef,
			"mapping_fingerprint": request.MappingFingerprint,
			"locator":             request.Locator,
		})
	}
	encoded, err := json.Marshal(snapshots)
	if err != nil {
		t.Fatalf("encode owner request characterization: %v", err)
	}
	const ownerRequestGolden = `[{"binding_mode":"mention_origin","field_key":"timeline.host_refs","locator":"file","mapping_fingerprint":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","normalized":{"kind":"collection_token","normalized_text":"Gateway One","raw_text":"  Gateway   One  "},"raw_value":"  Gateway   One  ","source_row_ref":2,"unknown":[{"SourceColumnOrdinal":2,"SourceHeaderText":"Unknown","RawValue":"raw\u2028value","CellKind":"inline_string"}],"view_schema_id":"cartulary.view.timeline.v2"},{"binding_mode":"entity_origin","field_key":"host.display_name","locator":"file","mapping_fingerprint":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","normalized":{"kind":"text","text":"Gateway One"},"raw_value":" Gateway One ","source_row_ref":2,"unknown":[{"SourceColumnOrdinal":2,"SourceHeaderText":"Unknown","RawValue":"raw\u2028value","CellKind":"inline_string"}],"view_schema_id":"cartulary.view.hosts.v1"},{"binding_mode":"entity_origin","field_key":"identity.display_name","locator":"file","mapping_fingerprint":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","normalized":{"kind":"text","text":"Analyst One"},"raw_value":" Analyst One ","source_row_ref":2,"unknown":[{"SourceColumnOrdinal":2,"SourceHeaderText":"Unknown","RawValue":"raw\u2028value","CellKind":"inline_string"}],"view_schema_id":"cartulary.view.identities.v1"},{"binding_mode":null,"field_key":"note.title","locator":"file","mapping_fingerprint":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","normalized":{"kind":"text","text":"Investigation note"},"raw_value":" Investigation note ","source_row_ref":2,"unknown":[{"SourceColumnOrdinal":2,"SourceHeaderText":"Unknown","RawValue":"raw\u2028value","CellKind":"inline_string"}],"view_schema_id":"cartulary.view.notes.v1"}]`
	if string(encoded) != ownerRequestGolden {
		t.Fatalf("owner request compatibility changed:\n%s", encoded)
	}
}

func importScalarCompatibilityValue(value ownerfacade.ImportScalarValue) map[string]any {
	result := map[string]any{"kind": value.Kind()}
	if text, ok := value.Text(); ok {
		result["text"] = text
	}
	if token, ok := value.CollectionToken(); ok {
		result["raw_text"] = token.RawText
		result["normalized_text"] = token.NormalizedText
	}
	return result
}

func assertImportMappingFailureCompatibility(t testing.TB, discovered []map[string]any) {
	t.Helper()
	base := `{"client_txn_id":"txn","target_view_schema_id":"cartulary.view.notes.v1","header_row_ref":1,"data_start_row_ref":2,"unknown_column_policy":"reject_if_unmapped","source_columns":[{"source_column_ordinal":1,"source_header_text":"Náme","field_key":"note.title","entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":2,"source_header_text":"Aliases","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":3,"source_header_text":"Location","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"},{"source_column_ordinal":4,"source_header_text":"Unknown \"<>&","field_key":null,"entity_binding_mode":null,"transform_id":null,"transform_options":{},"empty_value_policy":"omit_field"}]}`
	cases := []struct {
		name   string
		body   string
		field  string
		reason string
	}{
		{name: "noncontiguous", body: strings.Replace(base, `"source_column_ordinal":1`, `"source_column_ordinal":2`, 1), field: "source_columns", reason: "invalid_source_columns"},
		{name: "header drift", body: strings.Replace(base, `"source_header_text":"Náme"`, `"source_header_text":"Name"`, 1), field: "source_columns", reason: "invalid_source_columns"},
		{name: "invalid transform options", body: strings.Replace(base, `"transform_id":null,"transform_options":{}`, `"transform_id":"trim_v1","transform_options":{"delimiter":";"}`, 1), field: "transform_options", reason: "invalid_transform"},
		{name: "legacy null policy", body: strings.Replace(base, `"empty_value_policy":"omit_field"`, `"empty_value_policy":"use_null"`, 1), field: "empty_value_policy", reason: "invalid_empty_value_policy"},
		{name: "duplicate json member", body: strings.Replace(base, `"client_txn_id":"txn"`, `"client_txn_id":"txn","client_txn_id":"other"`, 1), field: "", reason: "request_not_object"},
	}
	for _, testCase := range cases {
		_, apiErr := DecodeMappingRequest(strings.NewReader(testCase.body), discovered)
		field, hasField := "", false
		if apiErr != nil {
			field, hasField = apiErr.Details["field"].(string)
		}
		if apiErr == nil || apiErr.Code != "invalid_import_request" ||
			(testCase.field == "" && hasField) || (testCase.field != "" && field != testCase.field) ||
			apiErr.Details["reason_code"] != testCase.reason {
			t.Fatalf("%s failure changed: %#v", testCase.name, apiErr)
		}
	}
}
