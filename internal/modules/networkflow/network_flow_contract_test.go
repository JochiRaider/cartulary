package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestImportOwnerErrorTranslationUsesClosedRegisteredUnion(t *testing.T) {
	t.Parallel()

	field := "source.ip"
	column := int64(2)
	diagnostic := RejectedRowDiagnostic{
		SourceRowNumber:     3,
		SourceColumnOrdinal: &column,
		FieldKey:            &field,
		ErrorCode:           "network_flow_invalid_ipv4",
		ReasonCode:          "invalid_ipv4",
	}
	cases := []struct {
		name       string
		err        error
		ownerCode  string
		coreReason string
	}{
		{
			name:       "no data rows",
			err:        importOwnerError("network_flow_no_data_rows", nil),
			ownerCode:  "network_flow_no_data_rows",
			coreReason: "owner_apply_validation_failed",
		},
		{
			name:       "all rows rejected",
			err:        allRowsRejectedOwnerError([]RejectedRowDiagnostic{diagnostic}, false),
			ownerCode:  "network_flow_all_rows_rejected",
			coreReason: "owner_apply_validation_failed",
		},
		{
			name: "mapping invalid",
			err: importOwnerError("network_flow_mapping_invalid", map[string]any{
				"reason_code": "variant_member_conflict",
				"field":       "source.ip",
			}),
			ownerCode:  "network_flow_mapping_invalid",
			coreReason: "owner_apply_validation_failed",
		},
		{
			name:       "source changed",
			err:        importOwnerError("network_flow_source_changed", nil),
			ownerCode:  "network_flow_source_changed",
			coreReason: "source_changed",
		},
		{
			name: "contract unavailable",
			err: importOwnerError("network_flow_target_unavailable", map[string]any{
				"reason_code": "owner_apply_contract_unavailable",
			}),
			ownerCode:  "network_flow_target_unavailable",
			coreReason: "owner_apply_contract_unavailable",
		},
		{
			name:       "internal failure",
			err:        importOwnerError("network_flow_internal_failure", nil),
			ownerCode:  "network_flow_internal_failure",
			coreReason: "owner_apply_validation_failed",
		},
	}
	facade := &importFacade{}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			translation, ok := facade.TranslateImportUnitError(testCase.err)
			if !ok ||
				translation.ErrorSchemaID != networkFlowImportOwnerErrorSchemaID ||
				translation.ErrorTranslationID != networkFlowImportErrorTranslationID ||
				translation.CoreReasonCode != testCase.coreReason ||
				translation.OwnerError.OwnerCode != testCase.ownerCode ||
				translation.OwnerError.Retryable {
				t.Fatalf("translation = %#v, ok=%t", translation, ok)
			}
			if err := facade.ValidateImportUnitError(translation.OwnerError); err != nil {
				t.Fatalf("validate translated owner error: %v", err)
			}
		})
	}

	if _, ok := facade.TranslateImportUnitError(errors.New("raw secret")); ok {
		t.Fatal("unknown owner error unexpectedly translated")
	}
	for _, invalid := range []imports.ExtensionImportOwnerError{
		{
			SchemaID:    networkFlowImportOwnerErrorSchemaID,
			OwnerCode:   "unregistered_owner_token",
			SafeDetails: map[string]any{},
		},
		{
			SchemaID:    networkFlowImportOwnerErrorSchemaID,
			OwnerCode:   "network_flow_source_changed",
			SafeDetails: map[string]any{"raw_source": "secret"},
		},
	} {
		if err := facade.ValidateImportUnitError(invalid); err == nil {
			t.Fatalf("invalid owner error accepted: %#v", invalid)
		}
	}
}

func AssertJSONAdmissionAndErrorDetails(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	_, apiErr := decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","visible_label":"Source IP"}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_request", "unknown_member")
	_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","schema_id":"cartulary.network_flow.table_query_request.v1"}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_request", "duplicate_member")
	for _, body := range []string{`[]`, `{`, `{"schema_id":null}`} {
		_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(body), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
		requireAPIError(t, apiErr, "network_flow_invalid_request", "")
	}
}

func AssertMappingApprovalBoundary(t *testing.T) {
	t.Helper()
	mapping := approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV)
	if err := validateApprovedMapping(mapping); err != nil {
		t.Fatalf("baseline mapping should validate: %v", err)
	}
	mapping.SourceProfileID = "reserved_ipfix_v1"
	err := validateApprovedMapping(mapping)
	var mappingErr *MappingValidationError
	if !errors.As(err, &mappingErr) || mappingErr.Code != "network_flow_unsupported_source_profile" {
		t.Fatalf("reserved source profile got %T %[1]v", err)
	}
	mapping = approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV)
	mapping.FieldMappings = append(mapping.FieldMappings, FieldMapping{
		MappingKind:         MappingKindSourceColumn,
		FieldKey:            FieldExporterID,
		SourceColumnOrdinal: len(mapping.SourceColumns),
		TransformID:         TransformTrimASCIISpace,
		EmptyValuePolicy:    EmptyPolicyNull,
	})
	err = validateApprovedMapping(mapping)
	if !errors.As(err, &mappingErr) || mappingErr.ReasonCode != "field_not_supported_by_profile" {
		t.Fatalf("unsupported Cisco SNA field got %T %[1]v", err)
	}
}

func AssertQueryAndTableScopeBoundary(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	_, apiErr := decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.rows_query_request.v1","table_scope":{"mode":"selected_tables","selected_table_ids":["nft_a","nft_a"]}}`), schemaRowsQueryRequest, schemaRowsQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_table_scope", "empty_resolved_scope")
	_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","filters":[{"field_key":"Source IP","op":"eq","value":"192.0.2.10"}]}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_filter", "unknown_field")
	_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","filters":[{"field_key":"network_flow.src_ip","op":"in","value":["198.51.100.200","198.51.100.200"]}]}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_filter", "duplicate_in_value")
	_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","sort":[{"field_key":"network_flow.endpoint_ip","direction":"asc"}]}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_sort", "unknown_field")
	request, apiErr := decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1"}`), schemaTableQueryRequest, schemaTableQueryContinuation, Limits{MaxQueryLimit: 50})
	if apiErr != nil || request.Limit != 50 {
		t.Fatalf("default query limit got request=%#v err=%v", request, apiErr)
	}
	acceptedEcho := string(canonicalJSON(acceptedRowsQueryEcho(nil, nil, effectiveSort(nil), []string{"nft_a"})))
	if !strings.Contains(acceptedEcho, `"filters":[]`) || !strings.Contains(acceptedEcho, `"sort":[]`) {
		t.Fatalf("accepted query echo must materialize omitted arrays: %s", acceptedEcho)
	}
	rejectedEcho := string(canonicalJSON(rejectedRowsQueryEcho(RejectedRowsQueryRequest{})))
	if !strings.Contains(rejectedEcho, `"error_codes":[]`) || !strings.Contains(rejectedEcho, `"field_keys":[]`) {
		t.Fatalf("rejected query echo must materialize omitted arrays: %s", rejectedEcho)
	}
	AssertKeysetAndCursorRuntime(t)
}

func AssertKeysetAndCursorRuntime(t *testing.T) {
	t.Helper()
	start := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	rows := []FlowRow{
		{NetworkFlowTableID: "nft_b", RowID: "nfr_2", SourceRowNumber: 2, FlowStartUTC: start, FlowEndUTC: start.Add(time.Minute), SrcIP: "2001:db8::1", DstIP: "198.51.100.2", BytesCount: "184467440737095516160", PacketsCount: "10"},
		{NetworkFlowTableID: "nft_a", RowID: "nfr_3", SourceRowNumber: 1, FlowStartUTC: start, FlowEndUTC: start.Add(time.Minute), SrcIP: "192.0.2.10", DstIP: "198.51.100.3", BytesCount: "9", PacketsCount: "2"},
		{NetworkFlowTableID: "nft_a", RowID: "nfr_1", SourceRowNumber: 1, FlowStartUTC: start, FlowEndUTC: start.Add(time.Minute), SrcIP: "192.0.2.2", DstIP: "198.51.100.1", BytesCount: "100", PacketsCount: "3"},
	}
	effective := effectiveSort([]SortSpec{{FieldKey: FieldBytesCount, Direction: "desc"}, {FieldKey: "network_flow_table_id", Direction: "asc"}})
	wantTail := []string{FieldFlowStartUTC, FieldFlowEndUTC, "source_row_number", "network_flow_row_id"}
	if len(effective) != 2+len(wantTail) {
		t.Fatalf("effective sort length got %d: %#v", len(effective), effective)
	}
	for index, field := range wantTail {
		if effective[index+2].FieldKey != field {
			t.Fatalf("effective sort tail[%d] got %q want %q", index, effective[index+2].FieldKey, field)
		}
	}

	sorted := sortRows(rows, []SortSpec{{FieldKey: FieldBytesCount, Direction: "desc"}, {FieldKey: "network_flow_table_id", Direction: "asc"}})
	seen := make([]string, 0, len(sorted))
	var position *rowCursorPosition
	for {
		page, more := pageFlowRowsAfter(sorted, position, 1)
		if len(page) == 0 {
			break
		}
		seen = append(seen, page[0].NetworkFlowTableID+"/"+page[0].RowID)
		next := newRowCursorPosition(page[0], []SortSpec{{FieldKey: FieldBytesCount, Direction: "desc"}, {FieldKey: "network_flow_table_id", Direction: "asc"}})
		encoded, err := json.Marshal(next)
		if err != nil || bytes.Contains(encoded, []byte(`"offset"`)) {
			t.Fatalf("row cursor position must be an offset-free keyset: %s err=%v", encoded, err)
		}
		position = &next
		if !more {
			break
		}
	}
	if len(seen) != len(rows) || len(mapFromStrings(seen)) != len(rows) {
		t.Fatalf("keyset pagination skipped or duplicated rows: %#v", seen)
	}
	args := []any{}
	clause, err := appendRowKeysetSQL(position.EffectiveSort, *position, &args)
	if err != nil || !strings.Contains(clause, "$1::") || strings.Contains(clause, position.NetworkFlowRowID) {
		t.Fatalf("keyset SQL must be whitelisted and parameterized: clause=%q args=%#v err=%v", clause, args, err)
	}

	tableRanks := map[string]int{"nft_a": 0, "nft_b": 1}
	contributors := append([]FlowRow(nil), rows...)
	sort.Slice(contributors, func(i, j int) bool {
		if tableRanks[contributors[i].NetworkFlowTableID] != tableRanks[contributors[j].NetworkFlowTableID] {
			return tableRanks[contributors[i].NetworkFlowTableID] < tableRanks[contributors[j].NetworkFlowTableID]
		}
		return compareRowToPosition(contributors[i], newRowCursorPosition(contributors[j], nil)) < 0
	})
	firstContributorPage, more := pageContributorRowsAfter(contributors, tableRanks, nil, 1)
	if len(firstContributorPage) != 1 || !more {
		t.Fatalf("contributor first keyset page got %#v more=%t", firstContributorPage, more)
	}
	contributorPosition := newContributorCursorPosition(firstContributorPage[0], tableRanks)
	remainingContributors, more := pageContributorRowsAfter(contributors, tableRanks, &contributorPosition, len(contributors))
	if len(remainingContributors) != len(contributors)-1 || more {
		t.Fatalf("contributor continuation got %d rows more=%t", len(remainingContributors), more)
	}

	AssertCursorCryptoRuntime(t, newRowCursorPosition(sorted[0], nil))
}

func AssertCursorCryptoRuntime(t *testing.T, position rowCursorPosition) {
	t.Helper()
	now := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	manifest := `{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"network_flow-cursor","state":"active","secret_ref":{"kind":"env","name":"network_flow-cursor"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"network_flow-safe","state":"active","secret_ref":{"kind":"env","name":"network_flow-safe"}}]}
}`
	rings, err := ParseKeyRings([]byte(manifest), map[string]string{
		"CARTULARY_SECRET_NETWORK_FLOW_CURSOR": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_NETWORK_FLOW_SAFE":   "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}, now)
	if err != nil {
		t.Fatalf("parse Network Flow key rings: %v", err)
	}
	clock := now
	codec, err := newCursorCodec(rings, func() time.Time { return clock })
	if err != nil {
		t.Fatalf("construct Network Flow cursor protector: %v", err)
	}
	binding := CursorBinding{Route: "nf.rows.query", ActorUserID: "actor", SessionID: "session", IncidentID: "incident", Scope: map[string]string{"table_ids": "nft_a"}, QueryHash: "query-hash", QueryEcho: json.RawMessage(`{"sort":[]}`), Limit: 1}
	token, err := codec.Encode(binding, "row_keyset_v1", position)
	if err != nil || !strings.HasPrefix(token, "nfc2.network_flow-cursor.") {
		t.Fatalf("encode nfc2 cursor token=%q err=%v", token, err)
	}
	payload, reason := codec.Decode(token)
	if reason != "" || payload.PositionKind != "row_keyset_v1" || bytes.Contains(payload.Position, []byte(`"offset"`)) {
		t.Fatalf("decode keyset cursor payload=%#v reason=%q", payload, reason)
	}
	changed := binding
	changed.ActorUserID = "other-actor"
	if reason := payload.Validate(changed); reason != "actor_mismatch" {
		t.Fatalf("cursor actor mismatch reason got %q", reason)
	}
	if _, reason := codec.Decode("nfc1.legacy.payload"); reason != "malformed" {
		t.Fatalf("legacy cursor must be invalid, got %q", reason)
	}
	clock = now.Add(cursorTTL)
	if _, reason := codec.Decode(token); reason != "expired" {
		t.Fatalf("cursor at expiry equality got %q", reason)
	}
	digester, err := newSafeDigester(rings, func() time.Time { return now })
	if err != nil {
		t.Fatalf("construct Network Flow safe digester: %v", err)
	}
	digest, keyID, err := digester.Digest("source_filename", "flows.csv")
	if err != nil || keyID != "network_flow-safe" || !hex64(digest) {
		t.Fatalf("configured safe digest=%q key_id=%q err=%v", digest, keyID, err)
	}
}

func AssertDiagnosticKeysetRuntime(t *testing.T) {
	t.Helper()
	column := int64(3)
	field := FieldSrcIP
	diagnostics := []RejectedRowDiagnostic{
		{DiagnosticID: "nfd_3", SourceRowNumber: 2, SourceColumnOrdinal: nil, FieldKey: nil, ErrorCode: "z", ReasonCode: "z"},
		{DiagnosticID: "nfd_2", SourceRowNumber: 2, SourceColumnOrdinal: &column, FieldKey: &field, ErrorCode: "a", ReasonCode: "b"},
		{DiagnosticID: "nfd_1", SourceRowNumber: 2, SourceColumnOrdinal: &column, FieldKey: &field, ErrorCode: "a", ReasonCode: "a"},
	}
	sort.Slice(diagnostics, func(i, j int) bool { return compareDiagnostics(diagnostics[i], diagnostics[j]) < 0 })
	first, more := pageDiagnosticsAfter(diagnostics, nil, 1)
	if len(first) != 1 || !more || first[0].DiagnosticID != "nfd_1" {
		t.Fatalf("diagnostic keyset first page=%#v more=%t", first, more)
	}
	position := newDiagnosticCursorPosition(first[0])
	rest, more := pageDiagnosticsAfter(diagnostics, &position, len(diagnostics))
	if len(rest) != 2 || more || rest[0].DiagnosticID != "nfd_2" || rest[1].DiagnosticID != "nfd_3" {
		t.Fatalf("diagnostic keyset continuation=%#v more=%t", rest, more)
	}
}

func AssertDuplicateHeaderRuntime(t *testing.T) {
	t.Helper()
	parsed, err := ParseCSVApply(strings.NewReader("Source IP,Source IP\n192.0.2.1,192.0.2.2\n"), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse duplicate-header CSV: %v", err)
	}
	if len(parsed.SourceColumns) != 2 || parsed.SourceColumns[0].SourceColumnOrdinal != 1 || parsed.SourceColumns[1].SourceColumnOrdinal != 2 || parsed.SourceColumns[0].RawHeaderText != parsed.SourceColumns[1].RawHeaderText {
		t.Fatalf("duplicate headers did not retain ordinal identity: %#v", parsed.SourceColumns)
	}
}

func AssertImportRuntime(t *testing.T) {
	t.Helper()
	header := strings.Join(append(requiredCiscoFields(), "Notes"), ",")
	valid := "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.0.2.10,198.51.100.2,443,51515,TCP,18446744073709551615,12,=1+1"
	invalid := "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.168.001.010,198.51.100.2,443,51515,TCP,1,1,invalid"
	parsed, err := ParseCSVApply(strings.NewReader(header+"\n"+valid+"\n"+invalid+"\n"), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse import CSV: %v", err)
	}
	mapping := approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV)
	mapping.SourceColumns = parsed.SourceColumns
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	accepted, diagnostics, truncated, err := ValidateRows(parsed, mapping, fingerprint, DefaultLimits())
	if err != nil || len(accepted) != 1 || len(diagnostics) != 1 || truncated {
		t.Fatalf("partial import accepted=%d diagnostics=%d truncated=%t err=%v", len(accepted), len(diagnostics), truncated, err)
	}
	if accepted[0].ExporterID != nil || accepted[0].InputInterface != nil || accepted[0].OutputInterface != nil {
		t.Fatalf("omitted nullable public fields must remain null: %#v", accepted[0])
	}
	if !bytes.Contains(accepted[0].UnmappedRaw, []byte("=1+1")) {
		t.Fatalf("formula-like unmapped value must remain inert data: %s", accepted[0].UnmappedRaw)
	}

	allRejected, err := ParseCSVApply(strings.NewReader(header+"\n"+invalid+"\n"), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse all-rejected CSV: %v", err)
	}
	mapping.SourceColumns = allRejected.SourceColumns
	accepted, diagnostics, _, err = ValidateRows(allRejected, mapping, MappingFingerprint(mapping, allRejected.SourceContentSHA256), DefaultLimits())
	if err != nil || len(accepted) != 0 || len(diagnostics) != 1 {
		t.Fatalf("all-rejected import accepted=%d diagnostics=%d err=%v", len(accepted), len(diagnostics), err)
	}

	var many strings.Builder
	many.WriteString(header + "\n")
	for range 51 {
		many.WriteString(valid + "\n")
	}
	preview, err := ParseCSVPreview(strings.NewReader(many.String()), "", DefaultLimits())
	if err != nil || len(preview.Records) != previewRecordLimit {
		t.Fatalf("preview record boundary got=%d err=%v", len(preview.Records), err)
	}
	apply, err := ParseCSVApply(strings.NewReader(many.String()), "", DefaultLimits())
	if err != nil || len(apply.Records) != 51 {
		t.Fatalf("apply record boundary got=%d err=%v", len(apply.Records), err)
	}
	limits := DefaultLimits()
	limits.MaxRowsPerCSV = 1
	if _, err := ParseCSVApply(strings.NewReader(header+"\n"+valid+"\n"+valid+"\n"), "", limits); err == nil {
		t.Fatal("apply parser accepted rows above the configured row limit")
	}
	if _, err := ParseCSVApply(strings.NewReader(header+"\n"+valid+"\n"), strings.Repeat("0", 64), DefaultLimits()); !errors.Is(err, ErrSourceChanged) {
		t.Fatalf("source hash mismatch got %T %[1]v", err)
	}
	if recovered, err := ParseCSVApply(strings.NewReader(header+"\n"+valid+"\n"), "", DefaultLimits()); err != nil || len(recovered.Records) != 1 {
		t.Fatalf("parser did not recover independently after a rejected operation: records=%d err=%v", len(recovered.Records), err)
	}
}

func AssertNameAndLifecycleRuntime(t *testing.T) {
	t.Helper()
	first, err := DeriveTableDisplayName("C:\\tmp\\flows.csv", map[string]struct{}{})
	if err != nil || first != "flows" {
		t.Fatalf("derive initial display name=%q err=%v", first, err)
	}
	second, err := DeriveTableDisplayName("flows.csv", map[string]struct{}{first: {}})
	if err != nil || second != "flows (2)" {
		t.Fatalf("derive collision display name=%q err=%v", second, err)
	}
	states := LifecycleStates()
	if len(states) != 2 || states[0] != TableStatusActive || states[1] != TableStatusSoftDeleted {
		t.Fatalf("closed lifecycle states drifted: %#v", states)
	}
	row := FlowRow{NetworkFlowTableID: "nft_stable", RowID: "nfr_stable", SourceRowNumber: 2}
	before := rowRefResource(row)
	_ = "renamed flows"
	after := rowRefResource(row)
	if !bytes.Equal(canonicalJSON(before), canonicalJSON(after)) {
		t.Fatalf("display-only rename changed row identity: before=%#v after=%#v", before, after)
	}
}

func mapFromStrings(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func AssertGraphContractBoundary(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	_, apiErr := decodeGraphQueryRequest(httptest.NewRequest("POST", "/graphs/query", strings.NewReader(`{"schema_id":"cartulary.network_flow.graph_query_request.v1","table_scope":{"mode":"selected_tables","selected_table_ids":["nft_a","nft_a"]}}`)), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_table_scope", "empty_resolved_scope")
	_, apiErr = decodeGraphQueryRequest(httptest.NewRequest("POST", "/graphs/query", strings.NewReader(`{"schema_id":"cartulary.network_flow.graph_query_request.v1","table_scope":{"mode":"all_active_tables"},"time_range":{"bucket":"hour"}}`)), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_request", "unknown_member")
	_, apiErr = decodeGraphQueryRequest(httptest.NewRequest("POST", "/graphs/query", strings.NewReader(`{"schema_id":"cartulary.network_flow.graph_query_request.v1","table_scope":{"mode":"all_active_tables"},"aggregation":{"mode":"time_bucket_v1"}}`)), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_request", "invalid_value")
	_, apiErr = decodeGraphQueryRequest(httptest.NewRequest("POST", "/graphs/query", strings.NewReader(`{"schema_id":"cartulary.network_flow.graph_query_request.v1","table_scope":{"mode":"all_active_tables"},"limit_overrides":{"max_vertices":999999}}`)), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_limit_override", "above_maximum")
	first := graphQueryDigest(IncidentID(), []string{"nft_b", "nft_a"}, nil, graphTimeRange{Omitted: true}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true})
	second := graphQueryDigest(IncidentID(), []string{"nft_a", "nft_b"}, nil, graphTimeRange{Omitted: true}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true})
	if first != second || !hex64(first) {
		t.Fatalf("graph query digest must be stable across table order: first=%q second=%q", first, second)
	}
}

func AssertIndicatorLinkContractBoundary(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	base := `{"schema_id":"cartulary.network_flow.indicator_link_request.v1","client_txn_id":"txn","selector":{"kind":"row_field_value","network_flow_table_id":"nft","network_flow_row_id":"nfr","field_key":"network_flow.src_ip"},"target":{"mode":"create_indicator","indicator_type":"ipv4_addr"},"observation_mode":"binding_only","confirm_exact_value":"192.0.2.10"}`
	request, apiErr := decodeIndicatorLinkRequest(httptest.NewRequest("POST", "/indicator-links", strings.NewReader(base)), limits)
	if apiErr != nil || request.Selector.Kind != "row_field_value" || request.Target.Mode != "create_indicator" {
		t.Fatalf("baseline indicator link decode request=%#v err=%v", request, apiErr)
	}
	_, apiErr = decodeIndicatorLinkRequest(httptest.NewRequest("POST", "/indicator-links", strings.NewReader(strings.Replace(base, `"binding_only"`, `"create_observation"`, 1))), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_request", "invalid_value")
	_, apiErr = decodeIndicatorLinkRequest(httptest.NewRequest("POST", "/indicator-links", strings.NewReader(strings.Replace(base, `"network_flow.src_ip"`, `"network_flow.bytes_count"`, 1))), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_indicator_selector", "field_not_linkable")
	rowRefs := `{"schema_id":"cartulary.network_flow.indicator_link_request.v1","client_txn_id":"txn","selector":{"kind":"row_refs","field_key":"network_flow.src_ip","row_refs":[{"network_flow_table_id":"nft","network_flow_row_id":"nfr","source_row_number":2,"mapping_fingerprint":"` + strings.Repeat("a", 64) + `"},{"network_flow_table_id":"nft","network_flow_row_id":"nfr","source_row_number":2,"mapping_fingerprint":"` + strings.Repeat("a", 64) + `"}]},"target":{"mode":"create_indicator","indicator_type":"ipv4_addr"},"observation_mode":"binding_only","confirm_exact_value":"192.0.2.10"}`
	_, apiErr = decodeIndicatorLinkRequest(httptest.NewRequest("POST", "/indicator-links", strings.NewReader(rowRefs)), limits)
	requireAPIError(t, apiErr, "network_flow_invalid_indicator_selector", "duplicate_row_ref")
	if !canonicalIPLiteral("192.0.2.10") || canonicalIPLiteral("192.168.001.010") {
		t.Fatalf("confirm_exact_value canonical IP predicate drifted")
	}
}

func AssertAuthorizationBoundary(t *testing.T) {
	t.Helper()
	routes := string(ReadFile(t, "internal/modules/networkflow/routes.go"))
	module := string(ReadFile(t, "internal/modules/networkflow/module.go"))
	boundary := module + "\n" + routes
	for _, required := range []string{"requireIncidentMembership", "requireIncidentRole"} {
		if !strings.Contains(boundary, required) {
			t.Fatalf("Network Flow routes missing authorization hook %q", required)
		}
	}
	assembly := string(ReadFile(t, "internal/app/server/runtime.go")) + "\n" + string(ReadFile(t, "internal/app/server/runtime_routes.go"))
	for _, required := range []string{"networkflow.RouteContributionID", "applicationRouteRegistrars"} {
		if !strings.Contains(assembly, required) {
			t.Fatalf("Network Flow application admission missing exact catalog hook %q", required)
		}
	}
	identities := string(ReadFile(t, "internal/modules/networkflow/api.go"))
	if !strings.Contains(identities, `RouteContributionID         = "network_flow_activity.route_family"`) {
		t.Fatal("Network Flow owner-local route contribution identity is missing")
	}
	httpapiGate := string(ReadFile(t, "internal/platform/httpapi/httpapi.go"))
	for _, required := range []string{"withUnclaimedReservedExtensionFamilies", "extension_profile_not_claimed", "MatchReservedExtensionRouteIn"} {
		if !strings.Contains(httpapiGate, required) {
			t.Fatalf("HTTP API extension gate missing authorization/admission hook %q", required)
		}
	}
	if strings.Contains(boundary, "deployment_admin") {
		t.Fatalf("Network Flow routes must not special-case deployment_admin incident-data access")
	}
}

func AssertNoThirdPartyEgress(t *testing.T) {
	t.Helper()
	for _, file := range NetworkFlowGoFiles(t) {
		content := string(ReadAbsoluteFile(t, file))
		for _, forbidden := range []string{"http.Get(", "http.Post(", "http.DefaultClient", "net.Dial(", "grpc.Dial("} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("Network Flow module contains forbidden egress primitive %q in %s", forbidden, file)
			}
		}
	}
}

func AssertRedactionAuditAndSafeDigestBoundary(t *testing.T) {
	t.Helper()
	textSample := sampleForValue("192.0.2.10")
	if textSample.SafeSample != nil || textSample.RawValueSHA256 == nil {
		t.Fatalf("IP-like raw value should be digest-only safe sample: %#v", textSample)
	}
	numericSample := sampleForValue("12345")
	if numericSample.SafeSample == nil || *numericSample.SafeSample != "12345" || numericSample.RawValueSHA256 == nil {
		t.Fatalf("bounded numeric sample should expose safe sample plus digest: %#v", numericSample)
	}
	digest, keyID := SafeDigest("network_flow-key", []byte("network_flow-secret"), "candidate", "192.0.2.10")
	if keyID != "network_flow-key" || !hex64(digest) {
		t.Fatalf("safe digest got digest=%q key_id=%q", digest, keyID)
	}
	AssertCursorCryptoRuntime(t, rowCursorPosition{
		EffectiveSort:      effectiveSort(nil),
		Values:             []any{"2026-07-13T12:00:00Z", "2026-07-13T12:01:00Z", int64(2), "nfr_network_flow"},
		NetworkFlowTableID: "nft_network_flow",
		NetworkFlowRowID:   "nfr_network_flow",
	})
}

func AssertResourceLimitBoundary(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	resource := effectiveLimitsResource(limits)
	for _, key := range []string{
		"network_flow.max_active_tables_per_incident",
		"network_flow.max_retained_tables_per_incident",
		"network_flow.max_selected_tables_per_query",
		"network_flow.max_query_limit",
		"network_flow.max_graph_vertices",
		"network_flow.max_graph_edges",
		"network_flow.max_aggregate_counter_digits",
	} {
		if resource[key] == nil {
			t.Fatalf("effective limit resource missing %q in %#v", key, resource)
		}
	}
	_, apiErr := decodeLowerableGraphLimit([]byte("0"), "max_vertices", 1, int(limits.MaxGraphVertices))
	requireAPIError(t, apiErr, "network_flow_invalid_limit_override", "below_minimum")
}

func AssertUnmappedRawInert(t *testing.T) {
	t.Helper()
	if isFilterField("unmapped_raw") || isSortField("unmapped_raw") || networkFlowLinkableIPField("unmapped_raw") {
		t.Fatalf("unmapped_raw must not be filterable, sortable, or indicator-linkable")
	}
}

func AssertInterfaceTextBoundary(t *testing.T) {
	t.Helper()
	if got, err := parseBoundedText256("00123"); err != nil || got != "00123" {
		t.Fatalf("interface identifiers must remain text: got=%q err=%v", got, err)
	}
	if _, err := parseBoundedText256(strings.Repeat("x", 257)); err == nil {
		t.Fatalf("interface text beyond bounded length should fail")
	}
}

func AssertFilenameDisplayBoundary(t *testing.T) {
	t.Helper()
	for input, want := range map[string]string{
		`C:\tmp\flows.csv`: "flows.csv",
		"/tmp/.csv":        ".csv",
		"file.":            "file.",
	} {
		if got := SanitizeSourceFilenameDisplay(input); got != want {
			t.Fatalf("sanitize filename %q got %q want %q", input, got, want)
		}
	}
}

func AssertImportFacadeBoundary(t *testing.T) {
	t.Helper()
	source := string(ReadFile(t, "internal/modules/networkflow/import_facade.go"))
	for _, required := range []string{"ErrSourceChanged", "ParseCSVPreview", "ParseCSVApply", "CreateTable"} {
		if !strings.Contains(source, required) {
			t.Fatalf("Network Flow import facade missing closed-boundary hook %q", required)
		}
	}
	AssertImportRuntime(t)
}

func requireAPIError(t *testing.T, apiErr *httpapi.APIError, code string, reason string) {
	t.Helper()
	if apiErr == nil {
		t.Fatalf("expected API error %s/%s", code, reason)
	}
	if apiErr.Code != code {
		t.Fatalf("API error code got %q want %q", apiErr.Code, code)
	}
	if reason == "" {
		return
	}
	got, _ := apiErr.Details["reason_code"].(string)
	if got != reason {
		t.Fatalf("API error reason got %q want %q", got, reason)
	}
}

func approvedMappingFixture(sourceProfileID string) ApprovedMapping {
	sourceColumns := make([]SourceColumnDescriptor, 0, len(requiredCiscoFields()))
	fieldMappings := make([]FieldMapping, 0, len(requiredCiscoFields())+1)
	for index, fieldKey := range requiredCiscoFields() {
		ordinal := index + 1
		sourceColumns = append(sourceColumns, SourceColumnDescriptor{
			SourceColumnOrdinal:           ordinal,
			RawHeaderText:                 fieldKey,
			NormalizedHeaderForSuggestion: SourceAliasMatchKey(fieldKey),
			RawHeaderSHA256:               strings.Repeat("a", 64),
		})
		fieldMappings = append(fieldMappings, FieldMapping{
			MappingKind:         MappingKindSourceColumn,
			FieldKey:            fieldKey,
			SourceColumnOrdinal: ordinal,
			TransformID:         defaultTransformForField(fieldKey),
			EmptyValuePolicy:    defaultEmptyPolicyForField(fieldKey),
		})
	}
	fieldMappings = append(fieldMappings, FieldMapping{
		MappingKind:   MappingKindSystemDerivation,
		FieldKey:      FieldObservationSourceRef,
		DerivationID:  "network_flow.observation_source_ref.v1",
		Combinability: "single_source_only",
	})
	return ApprovedMapping{
		TargetKind:          TargetKindNetworkFlowTable,
		TargetTableSchemaID: TargetTableSchemaID,
		SourceProfileID:     sourceProfileID,
		ParserProfileID:     ParserProfileRFC4180HeaderedCSV,
		UnknownColumnPolicy: UnknownColumnPolicyPreserve,
		TimestampProfile:    materializeTimestampProfile(TimestampProfile{}),
		SourceColumns:       sourceColumns,
		FieldMappings:       fieldMappings,
	}
}

func NetworkFlowGoFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Join(RepoRoot(t), "internal", "modules", "networkflow")
	files := []string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk networkflow module: %v", err)
	}
	return files
}
