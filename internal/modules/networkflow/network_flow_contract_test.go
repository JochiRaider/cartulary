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

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestNetworkFlow_CoreEnvelopeAdmissionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-003")
}

func TestNetworkFlow_DuplicateJSONAdmissionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-004")
}

func TestNetworkFlow_InvalidJSONAdmissionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-005")
}

func TestNetworkFlow_TableNameDerivationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-007")
}

func TestNetworkFlow_ImportReplayEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-008")
}

func TestNetworkFlow_DuplicateHeaderOrdinalEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-009")
}

func TestNetworkFlow_ReservedSourceProfileEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-011")
}

func TestNetworkFlow_FormulaValuesRemainInertEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-013")
}

func TestNetworkFlow_RejectedRowsExcludedEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-014")
}

func TestNetworkFlow_PartialImportAcceptedRowsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-015")
}

func TestNetworkFlow_AllRejectedImportEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-016")
}

func TestNetworkFlow_RenamePreservesIdentityEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-022")
}

func TestNetworkFlow_FieldKeyFilterEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-024")
}

func TestNetworkFlow_FilterInDuplicateEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-025")
}

func TestNetworkFlow_SortAndDefaultTailEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-027")
}

func TestNetworkFlow_CursorInvalidationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-028")
}

func TestNetworkFlow_DuplicateSelectedTablesEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-031")
}

func TestNetworkFlow_CrossTableGraphEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-032")
}

func TestNetworkFlow_EndpointMergeEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-033")
}

func TestNetworkFlow_DefaultEdgeAggregationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-034")
}

func TestNetworkFlow_GraphTimeOverlapEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-035")
}

func TestNetworkFlow_GraphLimitFailureEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-038")
}

func TestNetworkFlow_ExampleRefsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-039")
}

func TestNetworkFlow_GraphProjectionAdapterEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-040")
	AssertGraphProjectionTimestampNormalizesToProviderPrecision(t)
	AssertGraphProjectionAdapterAcceptsCanonicalImportFixture(t)
}

func TestNetworkFlow_ExistingIndicatorBindingEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-041")
}

func TestNetworkFlow_CreateIndicatorRoleEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-042")
}

func TestNetworkFlow_DuplicateIndicatorLinkEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-043")
}

func TestNetworkFlow_ClosedSelectorFailureEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-044")
}

func TestNetworkFlow_DeploymentAdminNoBypassEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-045")
}

func TestNetworkFlow_NoThirdPartyEgressEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-046")
}

func TestNetworkFlow_RawValueRedactionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-047")
}

func TestNetworkFlow_LimitDiscoveryAndEnforcementEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-048")
}

func TestNetworkFlow_UnmappedRawInertEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-051")
}

func TestNetworkFlow_EdgeIDNullPortEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-056")
}

func TestNetworkFlow_TimeBucketUnknownMembersEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-057")
}

func TestNetworkFlow_BindingOnlyObservationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-058")
}

func TestNetworkFlow_GraphDigestSemanticEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-059")
}

func TestNetworkFlow_DuplicateRenameAndCursorEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-061")
}

func TestNetworkFlow_AllRejectedErrorDetailsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-063")
}

func TestNetworkFlow_QueryAndGraphLimitEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-064")
}

func TestNetworkFlow_PreviewApplyBoundaryEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-065")
}

func TestNetworkFlow_BindingSourceRefsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-066")
}

func TestNetworkFlow_BindingDedupeIdentityEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-067")
}

func TestNetworkFlow_ConfirmExactValueMismatchEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-068")
}

func TestNetworkFlow_GraphProjectionIdentityEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-069")
}

func TestNetworkFlow_InterfaceTextEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-070")
}

func TestNetworkFlow_ClosedAggregationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-071")
}

func TestNetworkFlow_IdempotencyReplayEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-072")
}

func TestNetworkFlow_MappingVariantBoundaryEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-073")
}

func TestNetworkFlow_ResourceLimitTableEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-076")
}

func TestNetworkFlow_SafeSamplesEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-077")
}

func TestNetworkFlow_ErrorOrderingEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-078")
}

func TestNetworkFlow_IndicatorCandidateBoundaryEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-079")
}

func TestNetworkFlow_FilenameDisplayEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-080")
}

func TestNetworkFlow_AtomicImportCommitEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-081")
}

func TestNetworkFlow_DisplayNameCollisionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-082")
}

func TestNetworkFlow_RowLimitCountingEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-084")
}

func TestNetworkFlow_ImportFacadeSourceChangeEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-088")
}

func TestNetworkFlow_SourceColumnDispositionEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-089")
}

func TestNetworkFlow_PublicRowNullableFieldsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-090")
}

func TestNetworkFlow_DiagnosticDeterminismEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-091")
}

func TestNetworkFlow_TableScopeVariantEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-092")
}

func TestNetworkFlow_FilterNormalizationEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-093")
}

func TestNetworkFlow_CursorTokenBoundaryEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-094")
}

func TestNetworkFlow_KeysetCursorEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-095")
}

func TestNetworkFlow_AggregateCounterEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-096")
}

func TestNetworkFlow_GraphProjectionMetadataEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-097")
}

func TestNetworkFlow_GraphSuccessSchemaEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-098")
}

func TestNetworkFlow_ContributorRecomputeEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-099")
}

func TestNetworkFlow_IndicatorClosedVariantsEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-100")
}

func TestNetworkFlow_IndicatorAtomicCommitEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-101")
}

func TestNetworkFlow_SafeDigestKeyEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-102")
}

func TestNetworkFlow_AuditOccurrenceEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-103")
}

func TestNetworkFlow_RetentionSoftDeleteEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-104")
}

func TestNetworkFlow_RouteStatusCatalogEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-105")
}

func TestNetworkFlow_CancellationRecoveryEvidence_Integration(t *testing.T) {
	AssertIntegrationSelector(t, "NF-AC-107")
}

func AssertIntegrationSelector(t *testing.T, acID string) {
	t.Helper()
	switch acID {
	case "NF-AC-003", "NF-AC-004", "NF-AC-005", "NF-AC-049", "NF-AC-057", "NF-AC-078", "NF-AC-105":
		AssertJSONAdmissionAndErrorDetails(t)
	case "NF-AC-011", "NF-AC-073", "NF-AC-085", "NF-AC-089":
		AssertMappingApprovalBoundary(t)
	case "NF-AC-024", "NF-AC-025", "NF-AC-027", "NF-AC-031", "NF-AC-064", "NF-AC-092", "NF-AC-093", "NF-AC-094":
		AssertQueryAndTableScopeBoundary(t)
	case "NF-AC-032", "NF-AC-033", "NF-AC-034", "NF-AC-035", "NF-AC-038", "NF-AC-039", "NF-AC-040", "NF-AC-056", "NF-AC-059", "NF-AC-069", "NF-AC-096", "NF-AC-097", "NF-AC-098", "NF-AC-099":
		AssertGraphContractBoundary(t)
	case "NF-AC-041", "NF-AC-042", "NF-AC-043", "NF-AC-044", "NF-AC-058", "NF-AC-066", "NF-AC-067", "NF-AC-068", "NF-AC-079", "NF-AC-100", "NF-AC-101":
		AssertIndicatorLinkContractBoundary(t)
	case "NF-AC-045":
		AssertAuthorizationBoundary(t)
	case "NF-AC-046":
		AssertNoThirdPartyEgress(t)
	case "NF-AC-047", "NF-AC-077", "NF-AC-102", "NF-AC-103":
		AssertRedactionAuditAndSafeDigestBoundary(t)
	case "NF-AC-048", "NF-AC-076":
		AssertResourceLimitBoundary(t)
	case "NF-AC-051":
		AssertUnmappedRawInert(t)
	case "NF-AC-070":
		AssertInterfaceTextBoundary(t)
	case "NF-AC-080":
		AssertFilenameDisplayBoundary(t)
	case "NF-AC-088":
		AssertImportFacadeBoundary(t)
	case "NF-AC-007", "NF-AC-022", "NF-AC-061", "NF-AC-082", "NF-AC-104":
		AssertNameAndLifecycleRuntime(t)
	case "NF-AC-008", "NF-AC-013", "NF-AC-014", "NF-AC-015", "NF-AC-016", "NF-AC-063", "NF-AC-065", "NF-AC-072", "NF-AC-081", "NF-AC-084", "NF-AC-090", "NF-AC-107":
		AssertImportRuntime(t)
	case "NF-AC-009":
		AssertDuplicateHeaderRuntime(t)
	case "NF-AC-028", "NF-AC-095":
		AssertKeysetAndCursorRuntime(t)
	case "NF-AC-071":
		AssertGraphContractBoundary(t)
	case "NF-AC-091":
		AssertDiagnosticKeysetRuntime(t)
	default:
		t.Fatalf("unmapped Phase 12 Network Flow integration selector %s", acID)
	}
	AssertFixtureRuntimeEvidenceIfPresent(t, acID)
	AssertFixtureEvidenceIfPresent(t, acID)
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
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"phase12-cursor","state":"active","secret_ref":{"kind":"env","name":"phase12-cursor"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"phase12-safe","state":"active","secret_ref":{"kind":"env","name":"phase12-safe"}}]}
}`
	rings, err := ParseKeyRings([]byte(manifest), map[string]string{
		"CARTULARY_SECRET_PHASE12_CURSOR": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_PHASE12_SAFE":   "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}, now)
	if err != nil {
		t.Fatalf("parse Phase 12 key rings: %v", err)
	}
	clock := now
	codec, err := newCursorCodec(rings, func() time.Time { return clock })
	if err != nil {
		t.Fatalf("construct Phase 12 cursor protector: %v", err)
	}
	binding := CursorBinding{Route: "nf.rows.query", ActorUserID: "actor", SessionID: "session", IncidentID: "incident", Scope: map[string]string{"table_ids": "nft_a"}, QueryHash: "query-hash", QueryEcho: json.RawMessage(`{"sort":[]}`), Limit: 1}
	token, err := codec.Encode(binding, "row_keyset_v1", position)
	if err != nil || !strings.HasPrefix(token, "nfc2.phase12-cursor.") {
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
		t.Fatalf("construct Phase 12 safe digester: %v", err)
	}
	digest, keyID, err := digester.Digest("source_filename", "flows.csv")
	if err != nil || keyID != "phase12-safe" || !hex64(digest) {
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
	for _, required := range []string{"ExtensionProfileClaimedIn", "requireIncidentMembership", "requireIncidentRole"} {
		if !strings.Contains(boundary, required) {
			t.Fatalf("Network Flow routes missing authorization/admission hook %q", required)
		}
	}
	httpapiGate := string(ReadFile(t, "internal/platform/httpapi/httpapi.go"))
	for _, required := range []string{"withUnclaimedReservedExtensionFamilies", "extension_profile_not_claimed", "MatchReservedExtensionFamilyIn"} {
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
	digest, keyID := SafeDigest("phase12-key", []byte("phase12-secret"), "candidate", "192.0.2.10")
	if keyID != "phase12-key" || !hex64(digest) {
		t.Fatalf("safe digest got digest=%q key_id=%q", digest, keyID)
	}
	AssertCursorCryptoRuntime(t, rowCursorPosition{
		EffectiveSort:      effectiveSort(nil),
		Values:             []any{"2026-07-13T12:00:00Z", "2026-07-13T12:01:00Z", int64(2), "nfr_phase12"},
		NetworkFlowTableID: "nft_phase12",
		NetworkFlowRowID:   "nfr_phase12",
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

func AssertFixtureEvidenceIfPresent(t *testing.T, acID string) {
	t.Helper()
	byAC := FixtureManifestsByAC(t)
	if len(byAC[acID]) > 0 {
		AssertFixtureEvidence(t, acID)
	}
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
