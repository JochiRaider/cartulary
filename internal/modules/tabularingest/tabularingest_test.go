package tabularingest_test

import (
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
)

func TestSharedTabularIngestParsesMapsAndGroupsBatch(t *testing.T) {
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
	plan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
		ViewSchemaID:   "cartulary.view.timeline.v2",
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
	if plan.Kind != tabularingest.TabularRowPlanV1Kind ||
		plan.SourceKind != "clipboard_paste" ||
		plan.SourceFormat != tabularingest.SourceFormatTSV ||
		plan.ClientTxnID != "txn-u-9-02-shared-ingest" ||
		plan.MappingFingerprint == "" ||
		len(plan.Rows) != 2 {
		t.Fatalf("unexpected batch identity/grouping: %#v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("validate tabular row plan: %v", err)
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
	if len(plan.Warnings) != 2 || plan.Warnings[0].Code != tabularingest.WarningUnmappedValueV1 {
		t.Fatalf("unmapped values did not produce closed warnings: %#v", plan.Warnings)
	}

	tampered := plan
	tampered.MappingFingerprint = "tampered"
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered mapping fingerprint must fail validation")
	}

	assertMappingCompatibilityGoldens(t)
}

func assertMappingCompatibilityGoldens(t testing.TB) {
	t.Helper()
	type mappingCase struct {
		viewSchemaID  string
		startFieldKey string
		columns       []string
		text          string
		headers       []any
	}
	cases := []mappingCase{
		{
			viewSchemaID: "cartulary.view.timeline.v2", startFieldKey: "timeline.host_refs",
			columns: []string{"timeline.host_refs", "timeline.activity_synopsis_text"},
			text:    "  Gateway One  \tSummary 界\tunknown\nGateway Two\tSecond\textra",
			headers: []any{"Höst", "Summary", `Unknown "<>&`},
		},
		{
			viewSchemaID: "cartulary.view.hosts.v1", startFieldKey: "host.display_name",
			columns: []string{"host.display_name", "host.hostname"}, text: "Gateway\tgateway\textra",
			headers: []any{"Name", "Hostname", "Unknown"},
		},
		{
			viewSchemaID: "cartulary.view.identities.v1", startFieldKey: "identity.display_name",
			columns: []string{"identity.display_name", "identity.email"}, text: "Analyst\tanalyst@example.test\textra",
			headers: []any{"Name", "Email", "Unknown"},
		},
		{
			viewSchemaID: "cartulary.view.notes.v1", startFieldKey: "note.title",
			columns: []string{"note.title", "note.body"}, text: "Title\tBody\textra",
			headers: []any{"Title", "Body", "Unknown"},
		},
	}
	wantFingerprints := map[string]string{
		"cartulary.view.hosts.v1":      "afef824e68445aba7cf61f16821393cde96222a5ab3fa4b8c48580f88218a280",
		"cartulary.view.identities.v1": "7f5c9616a40424075c1fece83bd3f0b960b1f0351a07391cce253fcdeaf2dc88",
		"cartulary.view.notes.v1":      "a799631444bbd61cfd281b43014a07a01d49165e275142803f7d29a7c976cb1d",
		"cartulary.view.timeline.v2":   "379f1470b18ab9d8c41be53588fec00db4867f7cc92ac9047e2c2ac451f9a1a8",
	}
	gotFingerprints := make(map[string]string, len(cases))
	var timelinePlan tabularingest.TabularRowPlanV1
	for _, testCase := range cases {
		plan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
			ViewSchemaID: testCase.viewSchemaID, ClientTxnID: "txn-clipboard-界",
			SourceKind: "clipboard_paste", Text: testCase.text, Format: "tsv",
			StartFieldKey: testCase.startFieldKey, Columns: testCase.columns,
			SourceHeaders: testCase.headers, RequireTargets: -1,
		})
		if err != nil {
			t.Fatalf("build %s mapping compatibility plan: %v", testCase.viewSchemaID, err)
		}
		gotFingerprints[testCase.viewSchemaID] = plan.MappingFingerprint
		if testCase.viewSchemaID == "cartulary.view.timeline.v2" {
			timelinePlan = plan
		}
	}
	if encoded, _ := json.Marshal(gotFingerprints); string(encoded) != mustJSON(t, wantFingerprints) {
		t.Fatalf("clipboard mapping fingerprints changed: %s", encoded)
	}
	encoded, err := json.Marshal(timelinePlan)
	if err != nil {
		t.Fatalf("encode Timeline mapping compatibility plan: %v", err)
	}
	const timelinePlanGolden = `{"Kind":"tabular_row_plan_v1","ViewSchemaID":"cartulary.view.timeline.v2","ClientTxnID":"txn-clipboard-界","SourceKind":"clipboard_paste","SourceFormat":"tsv","SourceColumns":[{"Ordinal":1,"HeaderText":"Höst"},{"Ordinal":2,"HeaderText":"Summary"},{"Ordinal":3,"HeaderText":"Unknown \"\u003c\u003e\u0026"}],"FieldMappings":[{"SourceColumnOrdinal":1,"FieldKey":"timeline.host_refs","EntityBindingMode":"mention_origin"},{"SourceColumnOrdinal":2,"FieldKey":"timeline.activity_synopsis_text","EntityBindingMode":null}],"Rows":[{"RowOrdinal":1,"Cells":[{"FieldKey":"timeline.host_refs","RawValue":"  Gateway One  ","SourceColumnOrdinal":1,"EntityBindingMode":"mention_origin"},{"FieldKey":"timeline.activity_synopsis_text","RawValue":"Summary 界","SourceColumnOrdinal":2,"EntityBindingMode":null}],"Unmapped":[{"SourceKind":"clipboard_paste","SourceClientTxnID":"txn-clipboard-界","SourceRowOrdinal":1,"SourceColumnOrdinal":3,"SourceHeaderText":"Unknown \"\u003c\u003e\u0026","RawValue":"unknown"}]},{"RowOrdinal":2,"Cells":[{"FieldKey":"timeline.host_refs","RawValue":"Gateway Two","SourceColumnOrdinal":1,"EntityBindingMode":"mention_origin"},{"FieldKey":"timeline.activity_synopsis_text","RawValue":"Second","SourceColumnOrdinal":2,"EntityBindingMode":null}],"Unmapped":[{"SourceKind":"clipboard_paste","SourceClientTxnID":"txn-clipboard-界","SourceRowOrdinal":2,"SourceColumnOrdinal":3,"SourceHeaderText":"Unknown \"\u003c\u003e\u0026","RawValue":"extra"}]}],"Warnings":[{"Code":"unmapped_source_value","SourceRowOrdinal":1,"SourceColumnOrdinal":3},{"Code":"unmapped_source_value","SourceRowOrdinal":2,"SourceColumnOrdinal":3}],"MappingFingerprint":"379f1470b18ab9d8c41be53588fec00db4867f7cc92ac9047e2c2ac451f9a1a8"}`
	if string(encoded) != timelinePlanGolden {
		t.Fatalf("Timeline row-plan compatibility changed:\n%s", encoded)
	}

	headerPlan, err := tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
		ViewSchemaID: "cartulary.view.notes.v1", ClientTxnID: "txn-header",
		SourceKind: "clipboard_paste", Text: "Title,Body\nT,\"line one\nline two\"", Format: "csv",
		StartFieldKey: "note.body", Columns: []string{"note.body"},
		ExactHeaderLabels:    []string{"Title", "Body"},
		ExactHeaderFieldKeys: []string{"note.title", "note.body"}, RequireTargets: 1,
	})
	if err != nil || len(headerPlan.Rows) != 1 || len(headerPlan.Rows[0].Cells) != 2 ||
		headerPlan.Rows[0].Cells[1].RawValue != "line one\nline two" {
		t.Fatalf("exact-header/embedded-line compatibility changed: plan=%#v error=%v", headerPlan, err)
	}

	failures := []struct {
		name    string
		request tabularingest.MappingRequest
		message string
	}{
		{name: "empty", request: tabularingest.MappingRequest{SourceKind: "clipboard_paste", ViewSchemaID: "cartulary.view.notes.v1"}, message: "empty tabular payload"},
		{name: "unknown schema", request: tabularingest.MappingRequest{SourceKind: "clipboard_paste", ViewSchemaID: "unknown", Text: "x"}, message: "unknown view schema"},
		{name: "target mismatch", request: tabularingest.MappingRequest{SourceKind: "clipboard_paste", ViewSchemaID: "cartulary.view.notes.v1", Text: "x\ny", Columns: []string{"note.title"}, StartFieldKey: "note.title", RequireTargets: 1}, message: "target count must equal tabular row count"},
		{name: "unsupported field", request: tabularingest.MappingRequest{SourceKind: "clipboard_paste", ViewSchemaID: "cartulary.view.notes.v1", Text: "x", Columns: []string{"note.unknown"}, StartFieldKey: "note.unknown", RequireTargets: -1}, message: "unsupported field key note.unknown"},
	}
	for _, failure := range failures {
		_, err := tabularingest.BuildTabularRowPlanV1(failure.request)
		if err == nil || err.Error() != failure.message {
			t.Fatalf("%s failure changed: %v", failure.name, err)
		}
	}
}

func mustJSON(t testing.TB, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode compatibility value: %v", err)
	}
	return string(encoded)
}
