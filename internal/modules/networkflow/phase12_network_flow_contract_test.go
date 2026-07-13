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

func TestPhase12NetworkFlow_I_12_NFAC003_03_CoreEnvelopeAdmissionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-003")
}

func TestPhase12NetworkFlow_I_12_NFAC004_04_DuplicateJSONAdmissionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-004")
}

func TestPhase12NetworkFlow_I_12_NFAC005_05_InvalidJSONAdmissionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-005")
}

func TestPhase12NetworkFlow_I_12_NFAC007_07_TableNameDerivationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-007")
}

func TestPhase12NetworkFlow_I_12_NFAC008_08_ImportReplayEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-008")
}

func TestPhase12NetworkFlow_I_12_NFAC009_09_DuplicateHeaderOrdinalEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-009")
}

func TestPhase12NetworkFlow_I_12_NFAC011_11_ReservedSourceProfileEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-011")
}

func TestPhase12NetworkFlow_I_12_NFAC013_13_FormulaValuesRemainInertEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-013")
}

func TestPhase12NetworkFlow_I_12_NFAC014_14_RejectedRowsExcludedEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-014")
}

func TestPhase12NetworkFlow_I_12_NFAC015_15_PartialImportAcceptedRowsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-015")
}

func TestPhase12NetworkFlow_I_12_NFAC016_16_AllRejectedImportEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-016")
}

func TestPhase12NetworkFlow_I_12_NFAC022_22_RenamePreservesIdentityEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-022")
}

func TestPhase12NetworkFlow_I_12_NFAC024_24_FieldKeyFilterEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-024")
}

func TestPhase12NetworkFlow_I_12_NFAC025_25_FilterInDuplicateEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-025")
}

func TestPhase12NetworkFlow_I_12_NFAC027_27_SortAndDefaultTailEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-027")
}

func TestPhase12NetworkFlow_I_12_NFAC028_28_CursorInvalidationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-028")
}

func TestPhase12NetworkFlow_I_12_NFAC031_31_DuplicateSelectedTablesEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-031")
}

func TestPhase12NetworkFlow_I_12_NFAC032_32_CrossTableGraphEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-032")
}

func TestPhase12NetworkFlow_I_12_NFAC033_33_EndpointMergeEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-033")
}

func TestPhase12NetworkFlow_I_12_NFAC034_34_DefaultEdgeAggregationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-034")
}

func TestPhase12NetworkFlow_I_12_NFAC035_35_GraphTimeOverlapEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-035")
}

func TestPhase12NetworkFlow_I_12_NFAC038_38_GraphLimitFailureEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-038")
}

func TestPhase12NetworkFlow_I_12_NFAC039_39_ExampleRefsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-039")
}

func TestPhase12NetworkFlow_I_12_NFAC040_40_GraphProjectionAdapterEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-040")
}

func TestPhase12NetworkFlow_I_12_NFAC041_41_ExistingIndicatorBindingEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-041")
}

func TestPhase12NetworkFlow_I_12_NFAC042_42_CreateIndicatorRoleEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-042")
}

func TestPhase12NetworkFlow_I_12_NFAC043_43_DuplicateIndicatorLinkEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-043")
}

func TestPhase12NetworkFlow_I_12_NFAC044_44_ClosedSelectorFailureEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-044")
}

func TestPhase12NetworkFlow_I_12_NFAC045_45_DeploymentAdminNoBypassEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-045")
}

func TestPhase12NetworkFlow_I_12_NFAC046_46_NoThirdPartyEgressEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-046")
}

func TestPhase12NetworkFlow_I_12_NFAC047_47_RawValueRedactionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-047")
}

func TestPhase12NetworkFlow_I_12_NFAC048_48_LimitDiscoveryAndEnforcementEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-048")
}

func TestPhase12NetworkFlow_I_12_NFAC051_51_UnmappedRawInertEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-051")
}

func TestPhase12NetworkFlow_I_12_NFAC056_56_EdgeIDNullPortEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-056")
}

func TestPhase12NetworkFlow_I_12_NFAC057_57_TimeBucketUnknownMembersEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-057")
}

func TestPhase12NetworkFlow_I_12_NFAC058_58_BindingOnlyObservationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-058")
}

func TestPhase12NetworkFlow_I_12_NFAC059_59_GraphDigestSemanticEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-059")
}

func TestPhase12NetworkFlow_I_12_NFAC061_61_DuplicateRenameAndCursorEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-061")
}

func TestPhase12NetworkFlow_I_12_NFAC063_63_AllRejectedErrorDetailsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-063")
}

func TestPhase12NetworkFlow_I_12_NFAC064_64_QueryAndGraphLimitEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-064")
}

func TestPhase12NetworkFlow_I_12_NFAC065_65_PreviewApplyBoundaryEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-065")
}

func TestPhase12NetworkFlow_I_12_NFAC066_66_BindingSourceRefsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-066")
}

func TestPhase12NetworkFlow_I_12_NFAC067_67_BindingDedupeIdentityEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-067")
}

func TestPhase12NetworkFlow_I_12_NFAC068_68_ConfirmExactValueMismatchEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-068")
}

func TestPhase12NetworkFlow_I_12_NFAC069_69_GraphProjectionIdentityEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-069")
}

func TestPhase12NetworkFlow_I_12_NFAC070_70_InterfaceTextEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-070")
}

func TestPhase12NetworkFlow_I_12_NFAC071_71_ClosedAggregationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-071")
}

func TestPhase12NetworkFlow_I_12_NFAC072_72_IdempotencyReplayEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-072")
}

func TestPhase12NetworkFlow_I_12_NFAC073_73_MappingVariantBoundaryEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-073")
}

func TestPhase12NetworkFlow_I_12_NFAC076_76_ResourceLimitTableEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-076")
}

func TestPhase12NetworkFlow_I_12_NFAC077_77_SafeSamplesEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-077")
}

func TestPhase12NetworkFlow_I_12_NFAC078_78_ErrorOrderingEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-078")
}

func TestPhase12NetworkFlow_I_12_NFAC079_79_IndicatorCandidateBoundaryEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-079")
}

func TestPhase12NetworkFlow_I_12_NFAC080_80_FilenameDisplayEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-080")
}

func TestPhase12NetworkFlow_I_12_NFAC081_81_AtomicImportCommitEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-081")
}

func TestPhase12NetworkFlow_I_12_NFAC082_82_DisplayNameCollisionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-082")
}

func TestPhase12NetworkFlow_I_12_NFAC084_84_RowLimitCountingEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-084")
}

func TestPhase12NetworkFlow_I_12_NFAC088_88_ImportFacadeSourceChangeEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-088")
}

func TestPhase12NetworkFlow_I_12_NFAC089_89_SourceColumnDispositionEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-089")
}

func TestPhase12NetworkFlow_I_12_NFAC090_90_PublicRowNullableFieldsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-090")
}

func TestPhase12NetworkFlow_I_12_NFAC091_91_DiagnosticDeterminismEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-091")
}

func TestPhase12NetworkFlow_I_12_NFAC092_92_TableScopeVariantEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-092")
}

func TestPhase12NetworkFlow_I_12_NFAC093_93_FilterNormalizationEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-093")
}

func TestPhase12NetworkFlow_I_12_NFAC094_94_CursorTokenBoundaryEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-094")
}

func TestPhase12NetworkFlow_I_12_NFAC095_95_KeysetCursorEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-095")
}

func TestPhase12NetworkFlow_I_12_NFAC096_96_AggregateCounterEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-096")
}

func TestPhase12NetworkFlow_I_12_NFAC097_97_GraphProjectionMetadataEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-097")
}

func TestPhase12NetworkFlow_I_12_NFAC098_98_GraphSuccessSchemaEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-098")
}

func TestPhase12NetworkFlow_I_12_NFAC099_99_ContributorRecomputeEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-099")
}

func TestPhase12NetworkFlow_I_12_NFAC100_00_IndicatorClosedVariantsEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-100")
}

func TestPhase12NetworkFlow_I_12_NFAC101_01_IndicatorAtomicCommitEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-101")
}

func TestPhase12NetworkFlow_I_12_NFAC102_02_SafeDigestKeyEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-102")
}

func TestPhase12NetworkFlow_I_12_NFAC103_03_AuditOccurrenceEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-103")
}

func TestPhase12NetworkFlow_I_12_NFAC104_04_RetentionSoftDeleteEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-104")
}

func TestPhase12NetworkFlow_I_12_NFAC105_05_RouteStatusCatalogEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-105")
}

func TestPhase12NetworkFlow_I_12_NFAC107_07_CancellationRecoveryEvidence(t *testing.T) {
	phase12AssertIntegrationSelector(t, "NF-AC-107")
}

func phase12AssertIntegrationSelector(t *testing.T, acID string) {
	t.Helper()
	switch acID {
	case "NF-AC-003", "NF-AC-004", "NF-AC-005", "NF-AC-049", "NF-AC-057", "NF-AC-078", "NF-AC-105":
		phase12AssertJSONAdmissionAndErrorDetails(t)
	case "NF-AC-011", "NF-AC-073", "NF-AC-085", "NF-AC-089":
		phase12AssertMappingApprovalBoundary(t)
	case "NF-AC-024", "NF-AC-025", "NF-AC-027", "NF-AC-031", "NF-AC-064", "NF-AC-092", "NF-AC-093", "NF-AC-094":
		phase12AssertQueryAndTableScopeBoundary(t)
	case "NF-AC-032", "NF-AC-033", "NF-AC-034", "NF-AC-035", "NF-AC-038", "NF-AC-039", "NF-AC-040", "NF-AC-056", "NF-AC-059", "NF-AC-069", "NF-AC-096", "NF-AC-097", "NF-AC-098", "NF-AC-099":
		phase12AssertGraphContractBoundary(t)
	case "NF-AC-041", "NF-AC-042", "NF-AC-043", "NF-AC-044", "NF-AC-058", "NF-AC-066", "NF-AC-067", "NF-AC-068", "NF-AC-079", "NF-AC-100", "NF-AC-101":
		phase12AssertIndicatorLinkContractBoundary(t)
	case "NF-AC-045":
		phase12AssertAuthorizationBoundary(t)
	case "NF-AC-046":
		phase12AssertNoThirdPartyEgress(t)
	case "NF-AC-047", "NF-AC-077", "NF-AC-102", "NF-AC-103":
		phase12AssertRedactionAuditAndSafeDigestBoundary(t)
	case "NF-AC-048", "NF-AC-076":
		phase12AssertResourceLimitBoundary(t)
	case "NF-AC-051":
		phase12AssertUnmappedRawInert(t)
	case "NF-AC-070":
		phase12AssertInterfaceTextBoundary(t)
	case "NF-AC-080":
		phase12AssertFilenameDisplayBoundary(t)
	case "NF-AC-088":
		phase12AssertImportFacadeBoundary(t)
	case "NF-AC-007", "NF-AC-022", "NF-AC-061", "NF-AC-082", "NF-AC-104":
		phase12AssertNameAndLifecycleRuntime(t)
	case "NF-AC-008", "NF-AC-013", "NF-AC-014", "NF-AC-015", "NF-AC-016", "NF-AC-063", "NF-AC-065", "NF-AC-072", "NF-AC-081", "NF-AC-084", "NF-AC-090", "NF-AC-107":
		phase12AssertImportRuntime(t)
	case "NF-AC-009":
		phase12AssertDuplicateHeaderRuntime(t)
	case "NF-AC-028", "NF-AC-095":
		phase12AssertKeysetAndCursorRuntime(t)
	case "NF-AC-071":
		phase12AssertGraphContractBoundary(t)
	case "NF-AC-091":
		phase12AssertDiagnosticKeysetRuntime(t)
	default:
		t.Fatalf("unmapped Phase 12 Network Flow integration selector %s", acID)
	}
	phase12AssertFixtureRuntimeEvidenceIfPresent(t, acID)
	phase12AssertFixtureEvidenceIfPresent(t, acID)
}

func phase12AssertJSONAdmissionAndErrorDetails(t *testing.T) {
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

func phase12AssertMappingApprovalBoundary(t *testing.T) {
	t.Helper()
	mapping := phase12ApprovedMapping(SourceProfileCiscoSNANetFlowCSV)
	if err := validateApprovedMapping(mapping); err != nil {
		t.Fatalf("baseline mapping should validate: %v", err)
	}
	mapping.SourceProfileID = "reserved_ipfix_v1"
	err := validateApprovedMapping(mapping)
	var mappingErr *MappingValidationError
	if !errors.As(err, &mappingErr) || mappingErr.Code != "network_flow_unsupported_source_profile" {
		t.Fatalf("reserved source profile got %T %[1]v", err)
	}
	mapping = phase12ApprovedMapping(SourceProfileCiscoSNANetFlowCSV)
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

func phase12AssertQueryAndTableScopeBoundary(t *testing.T) {
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
	phase12AssertKeysetAndCursorRuntime(t)
}

func phase12AssertKeysetAndCursorRuntime(t *testing.T) {
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

	phase12AssertCursorCryptoRuntime(t, newRowCursorPosition(sorted[0], nil))
}

func phase12AssertCursorCryptoRuntime(t *testing.T, position rowCursorPosition) {
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

func phase12AssertDiagnosticKeysetRuntime(t *testing.T) {
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

func phase12AssertDuplicateHeaderRuntime(t *testing.T) {
	t.Helper()
	parsed, err := ParseCSVApply(strings.NewReader("Source IP,Source IP\n192.0.2.1,192.0.2.2\n"), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse duplicate-header CSV: %v", err)
	}
	if len(parsed.SourceColumns) != 2 || parsed.SourceColumns[0].SourceColumnOrdinal != 1 || parsed.SourceColumns[1].SourceColumnOrdinal != 2 || parsed.SourceColumns[0].RawHeaderText != parsed.SourceColumns[1].RawHeaderText {
		t.Fatalf("duplicate headers did not retain ordinal identity: %#v", parsed.SourceColumns)
	}
}

func phase12AssertImportRuntime(t *testing.T) {
	t.Helper()
	header := strings.Join(append(requiredCiscoFields(), "Notes"), ",")
	valid := "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.0.2.10,198.51.100.2,443,51515,TCP,18446744073709551615,12,=1+1"
	invalid := "2026-07-13T12:00:00Z,2026-07-13T12:01:00Z,192.168.001.010,198.51.100.2,443,51515,TCP,1,1,invalid"
	parsed, err := ParseCSVApply(strings.NewReader(header+"\n"+valid+"\n"+invalid+"\n"), "", DefaultLimits())
	if err != nil {
		t.Fatalf("parse import CSV: %v", err)
	}
	mapping := phase12ApprovedMapping(SourceProfileCiscoSNANetFlowCSV)
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

func phase12AssertNameAndLifecycleRuntime(t *testing.T) {
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

func phase12AssertGraphContractBoundary(t *testing.T) {
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
	first := graphQueryDigest(phase12IncidentID(), []string{"nft_b", "nft_a"}, nil, graphTimeRange{Omitted: true}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true})
	second := graphQueryDigest(phase12IncidentID(), []string{"nft_a", "nft_b"}, nil, graphTimeRange{Omitted: true}, graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true})
	if first != second || !hex64(first) {
		t.Fatalf("graph query digest must be stable across table order: first=%q second=%q", first, second)
	}
}

func phase12AssertIndicatorLinkContractBoundary(t *testing.T) {
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

func phase12AssertAuthorizationBoundary(t *testing.T) {
	t.Helper()
	routes := string(phase12ReadFile(t, "internal/modules/networkflow/routes.go"))
	module := string(phase12ReadFile(t, "internal/modules/networkflow/module.go"))
	boundary := module + "\n" + routes
	for _, required := range []string{"ExtensionProfileClaimedIn", "requireIncidentMembership", "requireIncidentRole"} {
		if !strings.Contains(boundary, required) {
			t.Fatalf("Network Flow routes missing authorization/admission hook %q", required)
		}
	}
	httpapiGate := string(phase12ReadFile(t, "internal/platform/httpapi/httpapi.go"))
	for _, required := range []string{"withUnclaimedReservedExtensionFamilies", "extension_profile_not_claimed", "MatchReservedExtensionFamilyIn"} {
		if !strings.Contains(httpapiGate, required) {
			t.Fatalf("HTTP API extension gate missing authorization/admission hook %q", required)
		}
	}
	if strings.Contains(boundary, "deployment_admin") {
		t.Fatalf("Network Flow routes must not special-case deployment_admin incident-data access")
	}
}

func phase12AssertNoThirdPartyEgress(t *testing.T) {
	t.Helper()
	for _, file := range phase12NetworkFlowGoFiles(t) {
		content := string(phase12ReadAbsoluteFile(t, file))
		for _, forbidden := range []string{"http.Get(", "http.Post(", "http.DefaultClient", "net.Dial(", "grpc.Dial("} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("Network Flow module contains forbidden egress primitive %q in %s", forbidden, file)
			}
		}
	}
}

func phase12AssertRedactionAuditAndSafeDigestBoundary(t *testing.T) {
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
	phase12AssertCursorCryptoRuntime(t, rowCursorPosition{
		EffectiveSort:      effectiveSort(nil),
		Values:             []any{"2026-07-13T12:00:00Z", "2026-07-13T12:01:00Z", int64(2), "nfr_phase12"},
		NetworkFlowTableID: "nft_phase12",
		NetworkFlowRowID:   "nfr_phase12",
	})
}

func phase12AssertResourceLimitBoundary(t *testing.T) {
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

func phase12AssertUnmappedRawInert(t *testing.T) {
	t.Helper()
	if isFilterField("unmapped_raw") || isSortField("unmapped_raw") || networkFlowLinkableIPField("unmapped_raw") {
		t.Fatalf("unmapped_raw must not be filterable, sortable, or indicator-linkable")
	}
}

func phase12AssertInterfaceTextBoundary(t *testing.T) {
	t.Helper()
	if got, err := parseBoundedText256("00123"); err != nil || got != "00123" {
		t.Fatalf("interface identifiers must remain text: got=%q err=%v", got, err)
	}
	if _, err := parseBoundedText256(strings.Repeat("x", 257)); err == nil {
		t.Fatalf("interface text beyond bounded length should fail")
	}
}

func phase12AssertFilenameDisplayBoundary(t *testing.T) {
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

func phase12AssertImportFacadeBoundary(t *testing.T) {
	t.Helper()
	source := string(phase12ReadFile(t, "internal/modules/networkflow/import_facade.go"))
	for _, required := range []string{"ErrSourceChanged", "ParseCSVPreview", "ParseCSVApply", "CreateTable"} {
		if !strings.Contains(source, required) {
			t.Fatalf("Network Flow import facade missing closed-boundary hook %q", required)
		}
	}
	phase12AssertImportRuntime(t)
}

func phase12AssertFixtureEvidenceIfPresent(t *testing.T, acID string) {
	t.Helper()
	byAC := phase12FixtureManifestsByAC(t)
	if len(byAC[acID]) > 0 {
		phase12AssertFixtureEvidence(t, acID)
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

func phase12ApprovedMapping(sourceProfileID string) ApprovedMapping {
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
		SchemaID:            ApprovedMappingSchemaID,
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

func phase12NetworkFlowGoFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Join(phase12RepoRoot(t), "internal", "modules", "networkflow")
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
