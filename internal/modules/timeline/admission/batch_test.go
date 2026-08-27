package admission

import (
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
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
	if string(BulkMutationRequestHash(fillDown)) == string(BulkMutationRequestHash(tagAssignment)) {
		t.Fatal("distinct bulk operation discriminators must produce distinct request hashes")
	}
	if fillHash, tagHash := hex.EncodeToString(BulkMutationRequestHash(fillDown)), hex.EncodeToString(BulkMutationRequestHash(tagAssignment)); fillHash != "10245b0f39fa61c4c3edcbda0bd5cf9819acdbf6bb1af443c71ab9a56c93f8e0" ||
		tagHash != "a2868037151547591db6d78d63c629e6176d24c9cedeb4556d9ea60f4e47ca11" {
		t.Fatalf("canonical bulk request hashes changed: fill=%s tag=%s", fillHash, tagHash)
	}

	exactHeaderLabels, exactHeaderFieldKeys, err := timelineV2ExactHeaders()
	if err != nil {
		t.Fatalf("derive exact Timeline headers: %v", err)
	}
	wantHeaderLabels := []string{
		"Date Entered", "Analyst", "MITRE", "Device/Object", "IP Address",
		"Activity Date (UTC)", "Activity Date (Local Time)", "RAW Activity",
		"Activity Synopsis", "Data Source",
	}
	wantHeaderFieldKeys := []string{
		"timeline.date_entered_text", "timeline.analyst_text", "timeline.mitre_stage_text",
		"timeline.device_object_text", "timeline.ip_address_text", "timeline.activity_utc_text",
		"timeline.activity_local_text", "timeline.raw_activity_text",
		"timeline.activity_synopsis_text", "timeline.data_source_text",
	}
	if !slices.Equal(exactHeaderLabels, wantHeaderLabels) || !slices.Equal(exactHeaderFieldKeys, wantHeaderFieldKeys) {
		t.Fatalf("derived exact Timeline headers changed: labels=%#v keys=%#v", exactHeaderLabels, exactHeaderFieldKeys)
	}
	resource, ok := viewschema.LookupPublicResource(timeline.TimelineViewSchemaID)
	if !ok {
		t.Fatal("Timeline public view schema is unavailable")
	}
	for _, field := range resource.Fields {
		included := slices.Contains(exactHeaderFieldKeys, field.FieldKey)
		if included != (!field.DefaultHidden && field.GridEditable) {
			t.Fatalf("exact-header derivation admitted the wrong field: %#v", field)
		}
	}

	exactRequest := ClipboardPasteRequest{
		ViewSchemaID:  timeline.TimelineViewSchemaID,
		ClientTxnID:   "txn-exact-header",
		ClipboardText: strings.Join(exactHeaderLabels, "\t") + "\n2026-07-25",
		Format:        "tsv",
		StartFieldKey: "timeline.activity_synopsis_text",
		Columns:       []string{"timeline.activity_synopsis_text"},
		Targets:       []timeline.OwnerBatchTargetV1{{Kind: "create"}},
	}
	exactPlan, err := BuildClipboardPlan(exactRequest)
	if err != nil {
		t.Fatalf("build exact-header Timeline plan: %v", err)
	}
	if len(exactPlan.Rows) != 1 || len(exactPlan.Rows[0].Cells) != len(exactHeaderFieldKeys) {
		t.Fatalf("exact Timeline header did not retain its complete stable mapping: %#v", exactPlan.Rows)
	}
	if exactPlan.Rows[0].Cells[0].RawValue != "2026-07-25" || exactPlan.Rows[0].Cells[1].RawValue != "" {
		t.Fatalf("exact Timeline header did not preserve empty trailing cells: %#v", exactPlan.Rows[0].Cells)
	}
	nonHeaderPlan, err := BuildClipboardPlan(ClipboardPasteRequest{
		ViewSchemaID: timeline.TimelineViewSchemaID, ClientTxnID: "txn-non-header",
		ClipboardText: "2026-07-25\tanalyst", Format: "tsv",
		StartFieldKey: exactHeaderFieldKeys[0], Columns: exactHeaderFieldKeys,
		Targets: []timeline.OwnerBatchTargetV1{{Kind: "create"}},
	})
	if err != nil {
		t.Fatalf("build non-header Timeline plan: %v", err)
	}
	if len(nonHeaderPlan.Rows) != 1 || len(nonHeaderPlan.Rows[0].Cells) != 2 ||
		nonHeaderPlan.Rows[0].Cells[0].FieldKey != exactHeaderFieldKeys[0] ||
		nonHeaderPlan.Rows[0].Cells[1].FieldKey != exactHeaderFieldKeys[1] {
		t.Fatalf("non-header Timeline mapping changed: %#v", nonHeaderPlan.Rows)
	}
	if got := hex.EncodeToString(ClipboardPasteRequestHash(exactRequest)); got != "8784cbc7cbb56d3876c070fb697651a2c7b292b3795d2da326a705ee002fb651" {
		t.Fatalf("canonical clipboard request hash changed: %s", got)
	}
	if exactPlan.MappingFingerprint != "e84df225b4fa1163e147b026e1fee1e2c44bbeacdad5d19acea2f65fb0bd5587" ||
		nonHeaderPlan.MappingFingerprint != "57b54f923a21cfcc8cf84c7266b4fb01226b2abf3337f669f5481edcbe09ff8b" {
		t.Fatalf("Timeline mapping fingerprints changed: exact=%s non-header=%s", exactPlan.MappingFingerprint, nonHeaderPlan.MappingFingerprint)
	}

	assertTimelineBatchAdmissionLimits(t)
}

func assertTimelineBatchAdmissionLimits(t *testing.T) {
	t.Helper()
	columns := make([]string, tabularingest.MaxClipboardCols)
	for index := range columns {
		columns[index] = "timeline.activity_synopsis_text"
	}
	clipboardTargets := make([]map[string]any, mutationpolicy.MaxOwnerBatchTargets)
	for index := range clipboardTargets {
		clipboardTargets[index] = map[string]any{"kind": "create"}
	}

	clipboardPayload := map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "txn-limit",
		"clipboard_text": "value", "format": "tsv",
		"start_field_key": "timeline.activity_synopsis_text", "columns": columns,
		"targets": clipboardTargets,
	}
	if _, apiErr := decodeClipboardPayload(t, clipboardPayload); apiErr != nil {
		t.Fatalf("64 columns and 500 targets must be admitted: %#v", apiErr)
	}
	clipboardPayload["columns"] = append(columns, "timeline.activity_synopsis_text")
	if _, apiErr := decodeClipboardPayload(t, clipboardPayload); apiErr == nil {
		t.Fatal("65 clipboard columns unexpectedly admitted")
	}
	clipboardPayload["columns"] = columns
	clipboardPayload["targets"] = append(clipboardTargets, map[string]any{"kind": "create"})
	if _, apiErr := decodeClipboardPayload(t, clipboardPayload); apiErr == nil {
		t.Fatal("501 clipboard targets unexpectedly admitted")
	}

	recordTarget := map[string]any{
		"record_id": "11111111-1111-4111-8111-111111111111", "base_row_version": 3,
	}
	bulkTargets := make([]map[string]any, mutationpolicy.MaxOwnerBatchTargets)
	for index := range bulkTargets {
		bulkTargets[index] = recordTarget
	}
	bulkPayload := map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "txn-bulk-limit",
		"kind":      timeline.OwnerBatchOperationFillDownV1,
		"field_key": "timeline.activity_synopsis_text", "value": "filled", "targets": bulkTargets,
	}
	if _, apiErr := decodeBulkPayload(t, bulkPayload); apiErr != nil {
		t.Fatalf("500 bulk targets must be admitted: %#v", apiErr)
	}
	bulkPayload["targets"] = append(bulkTargets, recordTarget)
	if _, apiErr := decodeBulkPayload(t, bulkPayload); apiErr == nil {
		t.Fatal("501 bulk targets unexpectedly admitted")
	}
}

func decodeClipboardPayload(t *testing.T, payload map[string]any) (ClipboardPasteRequest, *httpapi.APIError) {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode clipboard payload: %v", err)
	}
	request, apiErr := DecodeClipboardPasteRequest(strings.NewReader(string(encoded)))
	return request, apiErr
}

func decodeBulkPayload(t *testing.T, payload map[string]any) (BulkMutationRequest, *httpapi.APIError) {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode bulk payload: %v", err)
	}
	request, apiErr := DecodeBulkMutationRequest(strings.NewReader(string(encoded)), timeline.TimelineViewSchemaID)
	return request, apiErr
}
