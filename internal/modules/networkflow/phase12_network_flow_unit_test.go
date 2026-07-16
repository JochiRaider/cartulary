package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

type phase12FixtureManifest struct {
	SchemaID             string               `json:"schema_id"`
	ManifestVersion      int                  `json:"manifest_version"`
	FixtureID            string               `json:"fixture_id"`
	ProfileID            string               `json:"profile_id"`
	Freeze               phase12FixtureFreeze `json:"freeze"`
	SourceFiles          []phase12FixtureFile `json:"source_files"`
	ExpectedArtifacts    []phase12FixtureFile `json:"expected_artifacts"`
	TranscriptFiles      []phase12FixtureFile `json:"transcript_files"`
	AcceptanceIDs        []string             `json:"acceptance_ids"`
	ExecutionSelectors   []string             `json:"execution_selectors"`
	SourceBundleSHA256   string               `json:"source_bundle_sha256"`
	ExpectedBundleSHA256 string               `json:"expected_bundle_sha256"`
}

type phase12FixtureFreeze struct {
	Status       string `json:"status"`
	Revision     int    `json:"revision"`
	ChangePolicy string `json:"change_policy"`
}

type phase12FixtureFile struct {
	LogicalPath string `json:"logical_path"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	Role        string `json:"role"`
}

func TestPhase12NetworkFlow_U_12_NFAC010_10_CiscoSNARequiredFields(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-010")
}

func TestPhase12NetworkFlow_U_12_NFAC012_12_CSVParserEdgeOutcomes(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-012")
}

func TestPhase12NetworkFlow_U_12_NFAC017_17_DigestAndIDVectors(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-017")
}

func TestPhase12NetworkFlow_U_12_NFAC018_18_TimestampRejectsInvalidInputs(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-018")
}

func TestPhase12NetworkFlow_U_12_NFAC019_19_IPCanonicalizationVectors(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-019")
}

func TestPhase12NetworkFlow_U_12_NFAC020_20_Uint64DecimalGrammar(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-020")
}

func TestPhase12NetworkFlow_U_12_NFAC026_26_CIDRFamilyVectors(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-026")
}

func TestPhase12NetworkFlow_U_12_NFAC049_49_ErrorDetailShape(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-049")
}

func TestPhase12NetworkFlow_U_12_NFAC052_52_FixtureCorpusIsFrozen(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-052")
}

func TestPhase12NetworkFlow_U_12_NFAC053_53_MayStatementsHaveOmissionBehavior(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-053")
}

func TestGraphProjectionFailureClassification(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		reason string
	}{
		{
			name:   "invalid request remains adapter contract rejected",
			err:    graphprojection.NewLifecycleError("invalid_projection_request", "missing_required_member", map[string]any{"reason_code": "missing_required_member"}, nil),
			reason: "adapter_contract_rejected",
		},
		{
			name:   "fatal validation remains adapter contract rejected",
			err:    graphprojection.NewLifecycleError("ephemeral_projection_failed", "fatal_validation", map[string]any{"reason_code": "fatal_validation"}, nil),
			reason: "adapter_contract_rejected",
		},
		{
			name:   "projection computation failure is unavailable",
			err:    graphprojection.NewLifecycleError("ephemeral_projection_failed", "projection_computation_failed", map[string]any{"reason_code": "projection_computation_failed"}, nil),
			reason: "projection_unavailable",
		},
		{
			name:   "query unavailability is unavailable",
			err:    graphprojection.NewQueryError("projection_not_available", "refreshing", map[string]any{"reason_code": "refreshing"}, nil),
			reason: "projection_unavailable",
		},
		{
			name:   "unknown pre-outcome error is unavailable",
			err:    fmt.Errorf("graph projection dependency unavailable"),
			reason: "projection_unavailable",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireAPIError(t, graphProjectionFailedForProjectionError(test.err), "network_flow_graph_projection_failed", test.reason)
		})
	}
}

func phase12AssertGraphProjectionTimestampNormalizesToProviderPrecision(t *testing.T) {
	t.Helper()
	input := time.Date(2026, 7, 10, 12, 0, 0, 123456789, time.FixedZone("test", -4*60*60))
	if got, want := graphProjectionTimestamp(input), "2026-07-10T16:00:00.123456Z"; got != want {
		t.Fatalf("graph projection timestamp = %q, want %q", got, want)
	}
}

func phase12AssertGraphProjectionAdapterAcceptsCanonicalImportFixture(t *testing.T) {
	t.Helper()
	semanticQuery := graphSemanticQueryResource([]string{"nft_" + strings.Repeat("a", 64)}, nil, graphTimeRange{}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}, graphResultLimits{})
	if encoded := string(canonicalJSON(semanticQuery)); !strings.Contains(encoded, `"filters":[]`) {
		t.Fatalf("default-materialized semantic graph filters = %s, want empty array", encoded)
	}
	digestIncidentID := phase12IncidentID()
	digestAggregation := graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}
	digestTimeRange := graphTimeRange{Omitted: true}
	if omittedDigest, emptyDigest := graphQueryDigest(digestIncidentID, []string{"nft_" + strings.Repeat("a", 64)}, nil, digestTimeRange, digestAggregation), graphQueryDigest(digestIncidentID, []string{"nft_" + strings.Repeat("a", 64)}, []Filter{}, digestTimeRange, digestAggregation); omittedDigest != emptyDigest {
		t.Fatalf("omitted and empty graph filters produced different digests: %s != %s", omittedDigest, emptyDigest)
	}
	fixture := phase12ReadFile(t, "fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/source/cisco-sna-minimal.csv")
	parsed, err := ParseCSVApply(bytes.NewReader(fixture), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse canonical Network Flow fixture: %v", err)
	}
	mapping := phase12ApprovedMapping(SourceProfileCiscoSNANetFlowCSV)
	mapping.SourceColumns = parsed.SourceColumns
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	rows, diagnostics, _, err := ValidateRows(parsed, mapping, fingerprint, DefaultLimits())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("validate canonical Network Flow fixture: rows=%d diagnostics=%#v err=%v", len(rows), diagnostics, err)
	}
	incidentID := phase12IncidentID()
	tableID := "nft_" + strings.Repeat("a", 64)
	for index := range rows {
		rows[index].IncidentID = incidentID
		rows[index].NetworkFlowTableID = tableID
		rows[index].RowID = fmt.Sprintf("nfr_%064x", index+1)
	}
	table := TableRecord{
		IncidentID:         incidentID,
		TableID:            tableID,
		MappingFingerprint: fingerprint,
	}
	limits := DefaultLimits()
	composition := graphComposition{
		SourceTables:     []TableRecord{table},
		TableRanks:       map[string]int{tableID: 0},
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: []string{tableID},
		ResultLimits: graphResultLimits{
			MaxVertices:               int(limits.MaxGraphVertices),
			MaxEdges:                  int(limits.MaxGraphEdges),
			MaxExampleRowRefsPerEdge:  int(limits.MaxExampleRowRefsPerEdge),
			MaxAggregateCounterDigits: int(limits.MaxAggregateCounterDigits),
		},
	}
	if apiErr := composeGraphObjects(incidentID, rows, map[string]TableRecord{tableID: table}, &composition); apiErr != nil {
		t.Fatalf("compose canonical Network Flow graph: %#v", apiErr)
	}
	requestedAt := time.Date(2026, 7, 10, 12, 0, 0, 123456789, time.UTC)
	adapter := newGraphProjectionAdapter(func() time.Time { return requestedAt })
	graphViewKey := "network_flow_activity:" + incidentID.String() + ":" + strings.Repeat("b", 64)
	graphViewID, err := adapter.GraphViewID(graphViewKey)
	if err != nil {
		t.Fatalf("derive graph view ID: %v", err)
	}
	input := networkFlowProjectionInput(
		graphViewID,
		graphViewKey,
		uuid.MustParse("00000000-0000-4000-8000-000000000012"),
		"nfsnap_"+strings.Repeat("b", 64),
		composition,
		requestedAt,
	)
	result, err := adapter.ProjectEphemeral(context.Background(), canonicalJSON(input))
	if err != nil {
		var lifecycleErr *graphprojection.LifecycleError
		if errors.As(err, &lifecycleErr) {
			t.Fatalf("project canonical Network Flow graph: code=%s reason=%s field=%s details=%#v", lifecycleErr.Code, lifecycleErr.ReasonCode, lifecycleErr.Field, lifecycleErr.Details)
		}
		t.Fatalf("project canonical Network Flow graph: %v", err)
	}
	if summary, ok := result["validation_summary"].(map[string]any); !ok || summary["fatal_count"] != 0 || summary["error_count"] != 0 || summary["warning_count"] != 0 || summary["info_count"] != 0 {
		t.Fatalf("canonical Network Flow graph produced validation issues: %#v", result["validation_summary"])
	}
}

func TestPhase12NetworkFlow_U_12_NFAC054_54_NormativeAuthorityStaysInOwners(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-054")
}

func TestPhase12NetworkFlow_U_12_NFAC055_55_DocumentReferencesAndFixtureRowsResolve(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-055")
}

func TestPhase12NetworkFlow_U_12_NFAC060_60_TimestampPrecisionAndUptimeVectors(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-060")
}

func TestPhase12NetworkFlow_U_12_NFAC075_75_SuccessResourceShape(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-075")
}

func TestPhase12NetworkFlow_U_12_NFAC083_83_CSVPreviewBoundary(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-083")
}

func TestPhase12NetworkFlow_U_12_NFAC085_85_CiscoSNATargetBoundary(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-085")
}

func TestPhase12NetworkFlow_U_12_NFAC086_86_TrimASCIISpaceOnly(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-086")
}

func TestPhase12NetworkFlow_U_12_NFAC087_87_TimestampObjectBoundary(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-087")
}

func TestPhase12NetworkFlow_U_12_NFAC106_06_AdoptionPrerequisitesAreConcrete(t *testing.T) {
	phase12AssertUnitSelector(t, "NF-AC-106")
}

func phase12AssertUnitSelector(t *testing.T, acID string) {
	t.Helper()
	switch acID {
	case "NF-AC-010":
		phase12AssertCiscoSNARequiredFields(t)
	case "NF-AC-012":
		phase12AssertCSVParserEdges(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-017":
		phase12AssertDigestAlgorithms(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-018", "NF-AC-060", "NF-AC-087":
		phase12AssertTimestampRules(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-019":
		phase12AssertIPCanonicalization(t)
	case "NF-AC-020":
		phase12AssertUint64DecimalGrammar(t)
	case "NF-AC-026":
		phase12AssertCIDRFamilyBehavior(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-049":
		phase12AssertErrorDetailShape(t)
	case "NF-AC-050":
		phase12AssertLargeTimingClassification(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-052":
		phase12AssertFrozenFixtureCorpus(t)
	case "NF-AC-053":
		phase12AssertMayOmissionBehavior(t)
	case "NF-AC-054":
		phase12AssertNormativeAuthorityBoundary(t)
	case "NF-AC-055":
		phase12AssertDocumentReferencesConcrete(t)
	case "NF-AC-075":
		phase12AssertSuccessResourceShape(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-083":
		phase12AssertCSVPreviewBoundary(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-085":
		phase12AssertCiscoSNATargetBoundary(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-086":
		phase12AssertTrimASCIISpaceOnly(t)
		phase12AssertFixtureEvidence(t, acID)
	case "NF-AC-106":
		phase12AssertAdoptionPrerequisitesConcrete(t)
	default:
		t.Fatalf("unmapped Phase 12 Network Flow unit selector %s", acID)
	}
	phase12AssertFixtureRuntimeEvidenceIfPresent(t, acID)
}

func phase12AssertCiscoSNARequiredFields(t *testing.T) {
	t.Helper()
	want := []string{
		FieldFlowStartUTC,
		FieldFlowEndUTC,
		FieldSrcIP,
		FieldDstIP,
		FieldSrcPort,
		FieldDstPort,
		FieldIPProtocol,
		FieldBytesCount,
		FieldPacketsCount,
	}
	if got := requiredCiscoFields(); !sameStrings(got, want) {
		t.Fatalf("required Cisco SNA fields got %#v want %#v", got, want)
	}
	profile := sourceProfileResource()
	required, ok := profile["required_field_keys"].([]string)
	if !ok {
		t.Fatalf("source profile required_field_keys has unexpected type: %#v", profile["required_field_keys"])
	}
	if !sameStrings(required, want) {
		t.Fatalf("source profile required fields got %#v want %#v", required, want)
	}
}

func phase12AssertCSVParserEdges(t *testing.T) {
	t.Helper()
	limits := DefaultLimits()
	for name, input := range map[string]string{
		"empty":       "",
		"header_only": "a,b\n",
		"bad_quote":   "a,b\n\"unterminated,b\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCSVApply(strings.NewReader(input), "", limits); err == nil {
				t.Fatalf("expected %s CSV to fail", name)
			}
		})
	}
	if _, err := ParseCSVApply(bytes.NewReader([]byte{'a', 0x01, '\n', 'b', '\n'}), "", limits); err == nil {
		t.Fatalf("expected forbidden header control to fail")
	}
	parsed, err := ParseCSVApply(strings.NewReader("a,b\n1,2\n3\n"), "", limits)
	if err != nil {
		t.Fatalf("field-count fixture parse: %v", err)
	}
	if len(parsed.Records) != 2 || len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].ErrorCode != "network_flow_csv_field_count_mismatch" {
		t.Fatalf("unexpected field-count diagnostics: records=%#v diagnostics=%#v", parsed.Records, parsed.Diagnostics)
	}
}

func phase12AssertDigestAlgorithms(t *testing.T) {
	t.Helper()
	assertUnicode17TextCanonicalization(t)
	rowDigest := SourceRowDigest(ParserProfileRFC4180HeaderedCSV, 2, []string{"192.0.2.10", "443"})
	if !hex64(rowDigest) {
		t.Fatalf("source row digest is not sha256 hex: %q", rowDigest)
	}
	mappingFingerprint := strings.Repeat("a", 64)
	normalized := NormalizedRowDigest(mappingFingerprint, map[string]any{
		FieldFlowStartUTC:         "2026-07-10T12:00:00Z",
		FieldFlowEndUTC:           "2026-07-10T12:00:05Z",
		FieldSrcIP:                "192.0.2.10",
		FieldDstIP:                "192.0.2.20",
		FieldSrcPort:              443,
		FieldDstPort:              51515,
		FieldIPProtocol:           6,
		FieldBytesCount:           "1200",
		FieldPacketsCount:         "12",
		FieldObservationSourceRef: map[string]any{},
	}, map[string]any{})
	if !hex64(normalized) {
		t.Fatalf("normalized row digest is not sha256 hex: %q", normalized)
	}
	rowID := RowID(phase12IncidentID(), "nft_phase12", 2, rowDigest, normalized)
	if !strings.HasPrefix(rowID, "nfr_") || len(rowID) != len("nfr_")+64 {
		t.Fatalf("unexpected row ID: %q", rowID)
	}
	edgePort := int32(443)
	edgeWithPort := FlowEdgeID(phase12IncidentID(), "nfe_src", "nfe_dst", 6, &edgePort)
	edgeWithoutPort := FlowEdgeID(phase12IncidentID(), "nfe_src", "nfe_dst", 6, nil)
	if edgeWithPort == edgeWithoutPort || !strings.HasPrefix(edgeWithPort, "nff_") || !strings.HasPrefix(edgeWithoutPort, "nff_") {
		t.Fatalf("edge ID port identity not distinct: with=%q without=%q", edgeWithPort, edgeWithoutPort)
	}
}

func phase12AssertTimestampRules(t *testing.T) {
	t.Helper()
	profile := TimestampProfile{Mode: "rfc3339", Precision: "seconds"}
	for _, value := range []string{
		"2026-03-08T02:30:00",
		"2026-07-10t12:00:00z",
		"2016-12-31T23:59:60Z",
		"2026-07-10 12:00:00Z",
		"2026-07-10T12:00:00-00:00",
	} {
		if _, err := parseTimestamp(value, profile); err == nil {
			t.Fatalf("timestamp %q should be rejected", value)
		}
	}
	if got, err := parseTimestamp("2026-07-10T12:00:00Z", profile); err != nil || !got.Equal(time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("valid timestamp got %s err=%v", got, err)
	}
	assertTimestampProfileExactGrammarPrecisionAndZoneTransitions(t)
	assertNetFlowSystemUptimeTimestampAndMappingOrdinals(t)
}

func phase12AssertIPCanonicalization(t *testing.T) {
	t.Helper()
	for input, want := range map[string]string{
		"192.0.2.10":              "192.0.2.10",
		"2001:0db8:0:0::0001":     "2001:db8::1",
		"::ffff:192.0.2.10":       "::ffff:192.0.2.10",
		"2001:db8:0:0:0:0:0:0010": "2001:db8::10",
	} {
		got, err := parseIPLiteral(input)
		if err != nil || got != want {
			t.Fatalf("parse IP %q got %q err=%v want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"192.168.001.010", "2001:db8::1%eth0", "999.1.1.1"} {
		if _, err := parseIPLiteral(input); err == nil {
			t.Fatalf("IP %q should be rejected", input)
		}
	}
}

func phase12AssertUint64DecimalGrammar(t *testing.T) {
	t.Helper()
	for input, want := range map[string]string{
		"0":                    "0",
		"42":                   "42",
		"18446744073709551615": "18446744073709551615",
	} {
		got, err := parseUint64Decimal(input)
		if err != nil || got != want {
			t.Fatalf("parse uint64 %q got %q err=%v want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"01", "+1", "-1", "1.0", "1e3", "18446744073709551616"} {
		if _, err := parseUint64Decimal(input); err == nil {
			t.Fatalf("uint64 decimal %q should be rejected", input)
		}
	}
}

func phase12AssertCIDRFamilyBehavior(t *testing.T) {
	t.Helper()
	row := phase12FlowRow()
	matched, apiErr := rowMatchesFilter(row, Filter{FieldKey: FieldSrcIP, Op: "cidr_contains", Value: "192.0.2.0/24"})
	if apiErr != nil || !matched {
		t.Fatalf("expected IPv4 CIDR to match src_ip: matched=%v err=%v", matched, apiErr)
	}
	matched, apiErr = rowMatchesFilter(row, Filter{FieldKey: FieldSrcIP, Op: "cidr_contains", Value: "2001:db8::/32"})
	if apiErr != nil {
		t.Fatalf("IPv6 CIDR against IPv4 source should not error: %v", apiErr)
	}
	if matched {
		t.Fatalf("IPv6 CIDR unexpectedly matched IPv4 source")
	}
	mapped := row
	mapped.SrcIP = "::ffff:192.0.2.10"
	matched, apiErr = rowMatchesFilter(mapped, Filter{FieldKey: FieldSrcIP, Op: "cidr_contains", Value: "192.0.2.0/24"})
	if apiErr != nil {
		t.Fatalf("IPv4 CIDR against mapped IPv6 source should not error: %v", apiErr)
	}
	if matched {
		t.Fatalf("IPv4-mapped IPv6 must not be treated as IPv4")
	}
}

func phase12AssertErrorDetailShape(t *testing.T) {
	t.Helper()
	err := invalidFilter("value", "duplicate_in_value")
	if err.Status != 400 || err.Code != "network_flow_invalid_filter" {
		t.Fatalf("unexpected API error envelope core fields: %#v", err)
	}
	details := err.Details
	if details["field"] != "value" || details["reason_code"] != "duplicate_in_value" {
		t.Fatalf("unexpected API error details: %#v", details)
	}
}

func phase12AssertLargeTimingClassification(t *testing.T) {
	t.Helper()
	manifest := phase12ReadFile(t, "tools/phase12_test_map.json")
	if !bytes.Contains(manifest, []byte("Large-fixture timing evidence remains engineering-only unless Core 05")) {
		t.Fatalf("Phase 12 manifest must keep large-fixture timing evidence outside product conformance publication")
	}
}

func phase12AssertSuccessResourceShape(t *testing.T) {
	t.Helper()
	row := phase12FlowRow()
	row.UnmappedRaw = json.RawMessage(`{"12":{"source_column_ordinal":12,"raw_header_text":"Extra","decoded_value":"inert"}}`)
	resource := rowResource(row)
	for _, key := range []string{
		"network_flow_row_id",
		"network_flow_table_id",
		"source_row_number",
		FieldFlowStartUTC,
		FieldFlowEndUTC,
		FieldSrcIP,
		FieldDstIP,
		FieldSrcPort,
		FieldDstPort,
		FieldIPProtocol,
		FieldBytesCount,
		FieldPacketsCount,
		FieldExporterID,
		FieldInputInterface,
		FieldOutputInterface,
		FieldTCPFlags,
		FieldApplicationLabel,
		"unmapped_raw",
		FieldObservationSourceRef,
	} {
		if _, ok := resource[key]; !ok {
			t.Fatalf("row resource missing key %q in %#v", key, resource)
		}
	}
	if resource[FieldExporterID] != nil || resource[FieldTCPFlags] != nil || resource[FieldApplicationLabel] != nil {
		t.Fatalf("unsupported optional Cisco SNA fields must serialize as null: %#v", resource)
	}
}

func phase12AssertCSVPreviewBoundary(t *testing.T) {
	t.Helper()
	var b strings.Builder
	b.WriteString("a,b\n")
	for i := 0; i < 55; i++ {
		fmt.Fprintf(&b, "%d,%d\n", i, i)
	}
	parsed, err := ParseCSVPreview(strings.NewReader(b.String()), "", DefaultLimits())
	if err != nil {
		t.Fatalf("preview parse: %v", err)
	}
	if len(parsed.Records) != previewRecordLimit {
		t.Fatalf("preview record count got %d want %d", len(parsed.Records), previewRecordLimit)
	}
}

func phase12AssertCiscoSNATargetBoundary(t *testing.T) {
	t.Helper()
	if sourceMappableField(FieldExporterID) || sourceMappableField(FieldTCPFlags) || sourceMappableField(FieldApplicationLabel) {
		t.Fatalf("exporter, tcp_flags, and application label must not be Cisco SNA v1 source-mappable")
	}
	if !sourceMappableField(FieldInputInterface) || !sourceMappableField(FieldOutputInterface) {
		t.Fatalf("input/output interface fields must be the supported optional Cisco SNA targets")
	}
}

func phase12AssertTrimASCIISpaceOnly(t *testing.T) {
	t.Helper()
	record := CSVRecord{SourceRowNumber: 2, Fields: []string{"\t inside \t"}, FieldCountOK: true}
	mapping := ApprovedMapping{
		SourceColumns: []SourceColumnDescriptor{{SourceColumnOrdinal: 1, RawHeaderSHA256: strings.Repeat("a", 64)}},
		FieldMappings: []FieldMapping{{
			MappingKind:         MappingKindSourceColumn,
			FieldKey:            FieldInputInterface,
			SourceColumnOrdinal: 1,
			TransformID:         TransformTrimASCIISpace,
			EmptyValuePolicy:    EmptyPolicyNull,
		}},
	}
	value, diag := mappedValue(record, mapping, mapping.FieldMappings[0], FieldInputInterface)
	if diag != nil || value != "\t inside \t" {
		t.Fatalf("trim_ascii_space_v1 must not trim tab characters: value=%#v diag=%#v", value, diag)
	}
	record.Fields[0] = "  inside  "
	value, diag = mappedValue(record, mapping, mapping.FieldMappings[0], FieldInputInterface)
	if diag != nil || value != "inside" {
		t.Fatalf("trim_ascii_space_v1 must trim only ASCII spaces: value=%#v diag=%#v", value, diag)
	}
}

func phase12AssertMayOmissionBehavior(t *testing.T) {
	t.Helper()
	doc := string(phase12ReadFile(t, "docs/network-flow-activity-nlspec.md"))
	for _, paragraph := range strings.Split(doc, "\n\n") {
		if !strings.Contains(paragraph, "MAY") {
			continue
		}
		if strings.Contains(paragraph, "key words") || strings.Contains(paragraph, "| **MAY** |") || strings.Contains(paragraph, "A `MAY` statement") {
			continue
		}
		if strings.Contains(paragraph, "| ID | Criterion |") && strings.Contains(paragraph, "NF-AC-053") {
			continue
		}
		if !strings.Contains(paragraph, "Omission behavior:") {
			t.Fatalf("MAY paragraph lacks omission behavior:\n%s", paragraph)
		}
	}
}

func phase12AssertNormativeAuthorityBoundary(t *testing.T) {
	t.Helper()
	doc := string(phase12ReadFile(t, "docs/network-flow-activity-nlspec.md"))
	for _, required := range []string{
		"Research reports, UI guides, implementation guides, appendices, and external vendor documents MAY justify design choices",
		"MUST NOT become implementation-conformance authority",
		"This appendix is non-normative",
		"It does not add v1 implementation-conformance behavior",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("Network Flow NLSpec missing authority-boundary text %q", required)
		}
	}
}

func phase12AssertDocumentReferencesConcrete(t *testing.T) {
	t.Helper()
	doc := string(phase12ReadFile(t, "docs/network-flow-activity-nlspec.md"))
	if strings.Contains(doc, "TODO:") && strings.Contains(doc, "| `TODO:") {
		t.Fatalf("Network Flow NLSpec still contains concrete TODO table cells")
	}
	for _, required := range []string{
		"docs/spec/00_document_set_status_and_precedence.md",
		"docs/graph_projection_nlspec.md",
		"docs/testing-harness-nlspec.md",
		"fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/manifest.json",
		"fixtures/network-flow/NF-FIX-028-graph-aggregate-bounds/manifest.json",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("Network Flow NLSpec missing concrete locator %q", required)
		}
	}
}

func phase12AssertAdoptionPrerequisitesConcrete(t *testing.T) {
	t.Helper()
	phase12AssertFrozenFixtureCorpus(t)
	phase12AssertDocumentReferencesConcrete(t)
	doc := string(phase12ReadFile(t, "docs/network-flow-activity-nlspec.md"))
	statusCount := strings.Count(doc, "\nstatus: draft\n") + strings.Count(doc, "\nstatus: adopted/current\n")
	if statusCount != 1 {
		t.Fatalf("Network Flow NLSpec front matter must contain exactly one concrete status marker")
	}
}

func phase12AssertFixtureEvidence(t *testing.T, acID string) {
	t.Helper()
	byAC := phase12FixtureManifestsByAC(t)
	manifests := byAC[acID]
	if len(manifests) == 0 {
		t.Fatalf("%s has no frozen Network Flow fixture manifest coverage", acID)
	}
	for _, manifest := range manifests {
		if len(manifest.ExpectedArtifacts) == 0 {
			t.Fatalf("%s fixture %s has no expected artifact", acID, manifest.FixtureID)
		}
		if len(manifest.TranscriptFiles) == 0 {
			t.Fatalf("%s fixture %s has no transcript", acID, manifest.FixtureID)
		}
	}
}

// phase12AssertFixtureRuntimeEvidenceIfPresent makes frozen fixtures
// supplemental to a concrete product path. It deliberately executes authored
// CSV and mapping bytes through admission, canonicalization, and row
// validation; manifest existence and digests alone cannot satisfy the row.
func phase12AssertFixtureRuntimeEvidenceIfPresent(t *testing.T, acID string) {
	t.Helper()
	manifests := phase12FixtureManifestsByAC(t)[acID]
	for _, manifest := range manifests {
		root := filepath.Join(phase12RepoRoot(t), "fixtures", "network-flow", manifest.FixtureID)
		var parsed []*ParsedCSV
		var mappings []ApprovedMapping
		executed := 0
		for _, file := range manifest.SourceFiles {
			path := filepath.Join(root, filepath.FromSlash(file.LogicalPath))
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read runtime fixture %s/%s: %v", manifest.FixtureID, file.LogicalPath, err)
			}
			switch {
			case file.Role == "input" && strings.HasSuffix(file.LogicalPath, ".csv"):
				executed++
				value, parseErr := ParseCSVApply(bytes.NewReader(content), file.SHA256, DefaultLimits())
				if parseErr == nil {
					parsed = append(parsed, &value)
				}
				_ = SanitizeSourceFilenameDisplay(file.LogicalPath)
			case file.Role == "input" && strings.HasSuffix(file.LogicalPath, ".jsonl"):
				for _, line := range bytes.Split(content, []byte{'\n'}) {
					if len(bytes.TrimSpace(line)) == 0 {
						continue
					}
					executed++
					_, _ = decodeAcceptedRowQueryRequest(bytes.NewReader(line), schemaTableQueryRequest, schemaTableQueryContinuation, DefaultLimits())
				}
			case file.Role == "mapping" && strings.HasSuffix(file.LogicalPath, ".json"):
				executed++
				mapping, mappingErr := DecodeApprovedMapping(content)
				if mappingErr == nil {
					mappings = append(mappings, mapping)
				}
			}
		}
		for _, csv := range parsed {
			for _, mapping := range mappings {
				if !sourceColumnsMatch(mapping.SourceColumns, csv.SourceColumns) {
					continue
				}
				executed++
				fingerprint := MappingFingerprint(mapping, csv.SourceContentSHA256)
				_, _, _, _ = ValidateRows(*csv, mapping, fingerprint, DefaultLimits())
			}
		}
		if executed == 0 {
			t.Fatalf("%s fixture %s has no executable product-runtime input", acID, manifest.FixtureID)
		}
	}
}

func phase12AssertFrozenFixtureCorpus(t *testing.T) {
	t.Helper()
	root := phase12RepoRoot(t)
	fixtureRoot := filepath.Join(root, "fixtures", "network-flow")
	entries, err := os.ReadDir(fixtureRoot)
	if err != nil {
		t.Fatalf("read fixture root: %v", err)
	}
	dirs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	if len(dirs) != 28 {
		t.Fatalf("Network Flow fixture directory count got %d want 28: %#v", len(dirs), dirs)
	}
	seen := map[string]struct{}{}
	for _, dir := range dirs {
		manifest := phase12ReadManifest(t, filepath.Join(fixtureRoot, dir, "manifest.json"))
		if manifest.SchemaID != "cartulary.network_flow_fixture_manifest.v1" ||
			manifest.ManifestVersion != 1 ||
			manifest.ProfileID != ProfileID ||
			manifest.Freeze.Status != "frozen" ||
			manifest.Freeze.Revision != 1 ||
			manifest.Freeze.ChangePolicy != "new_fixture_revision_required" {
			t.Fatalf("fixture %s has invalid freeze metadata: %#v", dir, manifest)
		}
		if _, exists := seen[manifest.FixtureID]; exists {
			t.Fatalf("duplicate fixture id %q", manifest.FixtureID)
		}
		seen[manifest.FixtureID] = struct{}{}
		if len(manifest.SourceFiles) == 0 || len(manifest.ExpectedArtifacts) == 0 || len(manifest.TranscriptFiles) == 0 {
			t.Fatalf("fixture %s must include source, expected, and transcript files", manifest.FixtureID)
		}
		if len(manifest.AcceptanceIDs) == 0 || len(manifest.ExecutionSelectors) == 0 {
			t.Fatalf("fixture %s must include acceptance ids and execution selectors", manifest.FixtureID)
		}
		if !hex64(manifest.SourceBundleSHA256) || !hex64(manifest.ExpectedBundleSHA256) {
			t.Fatalf("fixture %s has invalid bundle hashes", manifest.FixtureID)
		}
		for _, file := range append(append([]phase12FixtureFile{}, manifest.SourceFiles...), append(manifest.ExpectedArtifacts, manifest.TranscriptFiles...)...) {
			phase12AssertFixtureFile(t, filepath.Join(fixtureRoot, dir), file)
		}
	}
}

func phase12FixtureManifestsByAC(t *testing.T) map[string][]phase12FixtureManifest {
	t.Helper()
	phase12AssertFrozenFixtureCorpus(t)
	root := filepath.Join(phase12RepoRoot(t), "fixtures", "network-flow")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read fixture root: %v", err)
	}
	byAC := map[string][]phase12FixtureManifest{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest := phase12ReadManifest(t, filepath.Join(root, entry.Name(), "manifest.json"))
		for _, acID := range manifest.AcceptanceIDs {
			byAC[acID] = append(byAC[acID], manifest)
		}
	}
	return byAC
}

func phase12AssertFixtureFile(t *testing.T, fixtureDir string, file phase12FixtureFile) {
	t.Helper()
	if file.LogicalPath == "" || file.SHA256 == "" {
		t.Fatalf("invalid fixture file metadata: %#v", file)
	}
	path := filepath.Clean(filepath.Join(fixtureDir, file.LogicalPath))
	if !strings.HasPrefix(path, filepath.Clean(fixtureDir)+string(os.PathSeparator)) {
		t.Fatalf("fixture logical path escapes fixture directory: %q", file.LogicalPath)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture file %s: %v", path, err)
	}
	if int64(len(data)) != file.SizeBytes {
		t.Fatalf("fixture file %s size got %d want %d", path, len(data), file.SizeBytes)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != file.SHA256 {
		t.Fatalf("fixture file %s sha256 got %s want %s", path, got, file.SHA256)
	}
}

func phase12ReadManifest(t *testing.T, path string) phase12FixtureManifest {
	t.Helper()
	var manifest phase12FixtureManifest
	if err := json.Unmarshal(phase12ReadAbsoluteFile(t, path), &manifest); err != nil {
		t.Fatalf("decode manifest %s: %v", path, err)
	}
	return manifest
}

func phase12ReadFile(t *testing.T, relative string) []byte {
	t.Helper()
	return phase12ReadAbsoluteFile(t, filepath.Join(phase12RepoRoot(t), filepath.FromSlash(relative)))
}

func phase12ReadAbsoluteFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func phase12RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root from %s", dir)
		}
		dir = parent
	}
}

func phase12IncidentID() uuid.UUID {
	return uuid.UUID{0x12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12}
}

func phase12FlowRow() FlowRow {
	srcPort := int32(443)
	dstPort := int32(51515)
	return FlowRow{
		RowID:                     "nfr_" + strings.Repeat("1", 64),
		NetworkFlowTableID:        "nft_phase12",
		IncidentID:                phase12IncidentID(),
		SourceRowNumber:           2,
		SourceRowDigestSHA256:     strings.Repeat("2", 64),
		NormalizedRowDigestSHA256: strings.Repeat("3", 64),
		MappingFingerprint:        strings.Repeat("4", 64),
		FlowStartUTC:              time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
		FlowEndUTC:                time.Date(2026, 7, 10, 12, 0, 5, 0, time.UTC),
		SrcIP:                     "192.0.2.10",
		DstIP:                     "2001:db8::20",
		SrcPort:                   &srcPort,
		DstPort:                   &dstPort,
		IPProtocol:                6,
		BytesCount:                "1200",
		PacketsCount:              "12",
		ObservationSourceRef:      json.RawMessage(`{"source_row_number":2}`),
		CreatedAt:                 time.Date(2026, 7, 10, 12, 1, 0, 0, time.UTC),
	}
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func hex64(value string) bool {
	if len(value) != 64 {
		return false
	}
	matched, _ := regexp.MatchString(`^[0-9a-f]{64}$`, value)
	return matched
}
