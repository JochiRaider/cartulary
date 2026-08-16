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

type FixtureManifest struct {
	SchemaID             string        `json:"schema_id"`
	ManifestVersion      int           `json:"manifest_version"`
	FixtureID            string        `json:"fixture_id"`
	ProfileID            string        `json:"profile_id"`
	Freeze               FixtureFreeze `json:"freeze"`
	SourceFiles          []FixtureFile `json:"source_files"`
	ExpectedArtifacts    []FixtureFile `json:"expected_artifacts"`
	TranscriptFiles      []FixtureFile `json:"transcript_files"`
	SourceBundleSHA256   string        `json:"source_bundle_sha256"`
	ExpectedBundleSHA256 string        `json:"expected_bundle_sha256"`
}

type FixtureFreeze struct {
	Status       string `json:"status"`
	Revision     int    `json:"revision"`
	ChangePolicy string `json:"change_policy"`
}

type FixtureFile struct {
	LogicalPath string `json:"logical_path"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	Role        string `json:"role"`
}

func TestNetworkFlow_CiscoSNARequiredFields_Unit(t *testing.T) {
	AssertCiscoSNARequiredFields(t)
}

func TestNetworkFlow_CSVParserEdgeOutcomes_Unit(t *testing.T) {
	AssertCSVParserEdges(t)
}

func TestNetworkFlow_DigestAndIDVectors_Unit(t *testing.T) {
	AssertDigestAlgorithms(t)
}

func TestNetworkFlow_TimestampRejectsInvalidInputs_Unit(t *testing.T) {
	AssertTimestampRules(t)
}

func TestNetworkFlow_IPCanonicalizationVectors_Unit(t *testing.T) {
	AssertIPCanonicalization(t)
}

func TestNetworkFlow_Uint64DecimalGrammar_Unit(t *testing.T) {
	AssertUint64DecimalGrammar(t)
}

func TestNetworkFlow_CIDRFamilyVectors_Unit(t *testing.T) {
	AssertCIDRFamilyBehavior(t)
}

func TestNetworkFlow_ErrorDetailShape_Unit(t *testing.T) {
	AssertErrorDetailShape(t)
}

func TestNetworkFlow_FixtureCorpusIsFrozen_Unit(t *testing.T) {
	AssertFrozenFixtureCorpus(t)
	AssertAllFixtureRuntimeBehavior(t)
}

func TestGraphProjectionFailureClassification(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		reason string
	}{
		{
			name:   "invalid request remains adapter contract rejected",
			err:    &graphprojection.ProjectionErrorV2{Code: "invalid_projection_request", ReasonCode: "missing_required_member", RetryAction: "do_not_retry"},
			reason: "adapter_contract_rejected",
		},
		{
			name:   "fatal validation remains adapter contract rejected",
			err:    &graphprojection.ProjectionErrorV2{Code: "projection_validation_failed", ReasonCode: "invalid_projection_config", RetryAction: "do_not_retry"},
			reason: "adapter_contract_rejected",
		},
		{
			name:   "projection computation failure is unavailable",
			err:    &graphprojection.ProjectionErrorV2{Code: "projection_computation_failed", ReasonCode: "dependency_unavailable", RetryAction: "retry_with_backoff"},
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

func AssertGraphProjectionSemanticInputExcludesOperationalFields(t *testing.T) {
	t.Helper()
	incidentID := IncidentID()
	input := networkFlowProjectionInput("nfsnap_"+strings.Repeat("b", 64), graphComposition{
		SourceTables: []TableRecord{{IncidentID: incidentID}},
		Vertices:     map[string]*graphVertex{},
		Edges:        map[string]*graphEdge{},
	})
	for _, removed := range []string{"graph_view_id", "source_owner_id", "requested_at", "requested_by", "relationship_definitions"} {
		if _, present := input[removed]; present {
			t.Fatalf("Graph Projection v2 semantic input retained %q", removed)
		}
	}
	config := input["projection_config"].(map[string]any)
	for _, removed := range []string{"graph_view_key", "retention_policy", "custom_config"} {
		if _, present := config[removed]; present {
			t.Fatalf("Graph Projection v2 configuration retained %q", removed)
		}
	}
}

func AssertGraphProjectionAdapterAcceptsCanonicalImportFixture(t *testing.T) {
	t.Helper()
	semanticQuery := graphSemanticQueryResource([]string{"nft_" + strings.Repeat("a", 64)}, nil, graphTimeRange{}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}, graphResultLimits{})
	if encoded := string(canonicalJSON(semanticQuery)); !strings.Contains(encoded, `"filters":[]`) {
		t.Fatalf("default-materialized semantic graph filters = %s, want empty array", encoded)
	}
	digestIncidentID := IncidentID()
	digestAggregation := graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}
	digestTimeRange := graphTimeRange{Omitted: true}
	if omittedDigest, emptyDigest := graphQueryDigest(digestIncidentID, []string{"nft_" + strings.Repeat("a", 64)}, nil, digestTimeRange, digestAggregation), graphQueryDigest(digestIncidentID, []string{"nft_" + strings.Repeat("a", 64)}, []Filter{}, digestTimeRange, digestAggregation); omittedDigest != emptyDigest {
		t.Fatalf("omitted and empty graph filters produced different digests: %s != %s", omittedDigest, emptyDigest)
	}
	fixture := ReadFile(t, "fixtures/network-flow/NF-FIX-001-cisco-sna-minimal/source/cisco-sna-minimal.csv")
	parsed, err := ParseCSVApply(bytes.NewReader(fixture), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse canonical Network Flow fixture: %v", err)
	}
	mapping := approvedMappingFixture(SourceProfileCiscoSNANetFlowCSV)
	mapping.SourceColumns = parsed.SourceColumns
	fingerprint := MappingFingerprint(mapping, parsed.SourceContentSHA256)
	rows, diagnostics, _, err := ValidateRows(parsed, mapping, fingerprint, DefaultLimits())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("validate canonical Network Flow fixture: rows=%d diagnostics=%#v err=%v", len(rows), diagnostics, err)
	}
	incidentID := IncidentID()
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
	adapter := newGraphProjectionAdapter()
	graphViewKey := "network_flow_activity:" + incidentID.String() + ":" + strings.Repeat("b", 64)
	graphViewID, err := adapter.GraphViewID(graphViewKey)
	if err != nil {
		t.Fatalf("derive graph view ID: %v", err)
	}
	input := networkFlowProjectionInput("nfsnap_"+strings.Repeat("b", 64), composition)
	result, err := adapter.ProjectEphemeral(context.Background(), graphViewID, canonicalJSON(input))
	if err != nil {
		var projectionErr *graphprojection.ProjectionErrorV2
		if errors.As(err, &projectionErr) {
			t.Fatalf("project canonical Network Flow graph: code=%s reason=%s details=%#v", projectionErr.Code, projectionErr.ReasonCode, projectionErr.Details)
		}
		t.Fatalf("project canonical Network Flow graph: %v", err)
	}
	if summary, ok := result["validation_summary"].(map[string]any); !ok || summary["fatal_count"] != 0 || summary["error_count"] != 0 || summary["warning_count"] != 0 || summary["info_count"] != 0 {
		t.Fatalf("canonical Network Flow graph produced validation issues: %#v", result["validation_summary"])
	}
	if result["projection_schema_id"] != graphprojection.ProjectionSchemaIDV2 || result["source_owner_id"] != graphSourceOwnerID || result["projection_result_id"] == "" {
		t.Fatalf("canonical Network Flow graph did not return a v2 exact result: %#v", result)
	}
	for _, removed := range []string{"ephemeral_projection_id", "projection_run_id", "generated_at", "state", "metadata"} {
		if _, present := result[removed]; present {
			t.Fatalf("Network Flow retained legacy Graph result member %q", removed)
		}
	}
}

func TestNetworkFlow_SuccessResourceShape_Unit(t *testing.T) {
	AssertSuccessResourceShape(t)
}

func TestNetworkFlow_CSVPreviewBoundary_Unit(t *testing.T) {
	AssertCSVPreviewBoundary(t)
}

func TestNetworkFlow_CiscoSNATargetBoundary_Unit(t *testing.T) {
	AssertCiscoSNATargetBoundary(t)
}

func TestNetworkFlow_TrimASCIISpaceOnly_Unit(t *testing.T) {
	AssertTrimASCIISpaceOnly(t)
}

func AssertCiscoSNARequiredFields(t *testing.T) {
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

func AssertCSVParserEdges(t *testing.T) {
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

func AssertDigestAlgorithms(t *testing.T) {
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
	rowID := RowID(IncidentID(), "nft_network_flow", 2, rowDigest, normalized)
	if !strings.HasPrefix(rowID, "nfr_") || len(rowID) != len("nfr_")+64 {
		t.Fatalf("unexpected row ID: %q", rowID)
	}
	edgePort := int32(443)
	edgeWithPort := FlowEdgeID(IncidentID(), "nfe_src", "nfe_dst", 6, &edgePort)
	edgeWithoutPort := FlowEdgeID(IncidentID(), "nfe_src", "nfe_dst", 6, nil)
	if edgeWithPort == edgeWithoutPort || !strings.HasPrefix(edgeWithPort, "nff_") || !strings.HasPrefix(edgeWithoutPort, "nff_") {
		t.Fatalf("edge ID port identity not distinct: with=%q without=%q", edgeWithPort, edgeWithoutPort)
	}
}

func AssertTimestampRules(t *testing.T) {
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

func AssertIPCanonicalization(t *testing.T) {
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

func AssertUint64DecimalGrammar(t *testing.T) {
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

func AssertCIDRFamilyBehavior(t *testing.T) {
	t.Helper()
	row := flowRowFixture()
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

func AssertErrorDetailShape(t *testing.T) {
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

func AssertSuccessResourceShape(t *testing.T) {
	t.Helper()
	row := flowRowFixture()
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

func AssertCSVPreviewBoundary(t *testing.T) {
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

func AssertCiscoSNATargetBoundary(t *testing.T) {
	t.Helper()
	if sourceMappableField(FieldExporterID) || sourceMappableField(FieldTCPFlags) || sourceMappableField(FieldApplicationLabel) {
		t.Fatalf("exporter, tcp_flags, and application label must not be Cisco SNA v1 source-mappable")
	}
	if !sourceMappableField(FieldInputInterface) || !sourceMappableField(FieldOutputInterface) {
		t.Fatalf("input/output interface fields must be the supported optional Cisco SNA targets")
	}
}

func AssertTrimASCIISpaceOnly(t *testing.T) {
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

// AssertAllFixtureRuntimeBehavior executes every frozen fixture against the
// production parsing, mapping, canonicalization, validation, and request
// admission functions represented by its source bytes. Expected artifacts and
// transcripts are decoded and matched to the same fixture revision.
func AssertAllFixtureRuntimeBehavior(t *testing.T) {
	t.Helper()
	manifests := ReadFixtureManifests(t)
	for _, manifest := range manifests {
		manifest := manifest
		t.Run(manifest.FixtureID, func(t *testing.T) {
			root := filepath.Join(RepoRoot(t), "fixtures", "network-flow", manifest.FixtureID)
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
				t.Fatalf("fixture %s has no executable product-runtime input", manifest.FixtureID)
			}
			for _, file := range append(
				append([]FixtureFile{}, manifest.ExpectedArtifacts...),
				manifest.TranscriptFiles...,
			) {
				var artifact struct {
					FixtureID      string `json:"fixture_id"`
					FreezeRevision int    `json:"freeze_revision"`
				}
				content := ReadAbsoluteFile(
					t,
					filepath.Join(root, filepath.FromSlash(file.LogicalPath)),
				)
				if err := json.Unmarshal(content, &artifact); err != nil {
					t.Fatalf("decode fixture artifact %s: %v", file.LogicalPath, err)
				}
				if artifact.FixtureID != manifest.FixtureID ||
					artifact.FreezeRevision != manifest.Freeze.Revision {
					t.Fatalf(
						"fixture artifact %s identity/revision = %s/%d, want %s/%d",
						file.LogicalPath,
						artifact.FixtureID,
						artifact.FreezeRevision,
						manifest.FixtureID,
						manifest.Freeze.Revision,
					)
				}
			}
		})
	}
}

func AssertFrozenFixtureCorpus(t *testing.T) {
	t.Helper()
	root := RepoRoot(t)
	fixtureRoot := filepath.Join(root, "fixtures", "network-flow")
	manifests := ReadFixtureManifests(t)
	seen := map[string]struct{}{}
	for _, manifest := range manifests {
		if manifest.SchemaID != "cartulary.network_flow_fixture_manifest.v2" ||
			manifest.ManifestVersion != 2 ||
			manifest.ProfileID != ProfileID ||
			manifest.Freeze.Status != "frozen" ||
			manifest.Freeze.Revision != 2 ||
			manifest.Freeze.ChangePolicy != "new_fixture_revision_required" {
			t.Fatalf("fixture %s has invalid freeze metadata: %#v", manifest.FixtureID, manifest)
		}
		if _, exists := seen[manifest.FixtureID]; exists {
			t.Fatalf("duplicate fixture id %q", manifest.FixtureID)
		}
		seen[manifest.FixtureID] = struct{}{}
		if len(manifest.SourceFiles) == 0 || len(manifest.ExpectedArtifacts) == 0 || len(manifest.TranscriptFiles) == 0 {
			t.Fatalf("fixture %s must include source, expected, and transcript files", manifest.FixtureID)
		}
		if !hex64(manifest.SourceBundleSHA256) || !hex64(manifest.ExpectedBundleSHA256) {
			t.Fatalf("fixture %s has invalid bundle hashes", manifest.FixtureID)
		}
		for _, file := range append(append([]FixtureFile{}, manifest.SourceFiles...), append(manifest.ExpectedArtifacts, manifest.TranscriptFiles...)...) {
			AssertFixtureFile(t, filepath.Join(fixtureRoot, manifest.FixtureID), file)
		}
	}
}

func ReadFixtureManifests(t *testing.T) []FixtureManifest {
	t.Helper()
	root := filepath.Join(RepoRoot(t), "fixtures", "network-flow")
	entries, err := os.ReadDir(root)
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
	manifests := make([]FixtureManifest, 0, len(dirs))
	for _, dir := range dirs {
		manifests = append(
			manifests,
			ReadManifest(t, filepath.Join(root, dir, "manifest.json")),
		)
	}
	return manifests
}

func AssertFixtureFile(t *testing.T, fixtureDir string, file FixtureFile) {
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

func ReadManifest(t *testing.T, path string) FixtureManifest {
	t.Helper()
	var manifest FixtureManifest
	if err := json.Unmarshal(ReadAbsoluteFile(t, path), &manifest); err != nil {
		t.Fatalf("decode manifest %s: %v", path, err)
	}
	return manifest
}

func ReadFile(t *testing.T, relative string) []byte {
	t.Helper()
	return ReadAbsoluteFile(t, filepath.Join(RepoRoot(t), filepath.FromSlash(relative)))
}

func ReadAbsoluteFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func RepoRoot(t *testing.T) string {
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

func IncidentID() uuid.UUID {
	return uuid.UUID{0x12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12}
}

func flowRowFixture() FlowRow {
	srcPort := int32(443)
	dstPort := int32(51515)
	return FlowRow{
		RowID:                     "nfr_" + strings.Repeat("1", 64),
		NetworkFlowTableID:        "nft_network_flow",
		IncidentID:                IncidentID(),
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
