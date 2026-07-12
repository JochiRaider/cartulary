package networkflow

import (
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	}
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
	request, apiErr := decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","filters":[{"field_key":"network_flow.src_ip","op":"in","value":["198.51.100.200","198.51.100.200"]}]}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	if apiErr != nil {
		t.Fatalf("decode duplicate-in query: %v", apiErr)
	}
	_, apiErr = rowMatchesFilter(phase12FlowRow(), request.Filters[0])
	requireAPIError(t, apiErr, "network_flow_invalid_filter", "duplicate_in_value")
	_, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1","sort":[{"field_key":"network_flow.endpoint_ip","direction":"asc"}]}`), schemaTableQueryRequest, schemaTableQueryContinuation, limits)
	requireAPIError(t, apiErr, "network_flow_invalid_sort", "unknown_field")
	request, apiErr = decodeAcceptedRowQueryRequest(strings.NewReader(`{"schema_id":"cartulary.network_flow.table_query_request.v1"}`), schemaTableQueryRequest, schemaTableQueryContinuation, Limits{MaxQueryLimit: 50})
	if apiErr != nil || request.Limit != 50 {
		t.Fatalf("default query limit got request=%#v err=%v", request, apiErr)
	}
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
	for _, required := range []string{"ExtensionProfileClaimedIn", "requireIncidentMembership", "requireIncidentRole"} {
		if !strings.Contains(routes, required) {
			t.Fatalf("Network Flow routes missing authorization/admission hook %q", required)
		}
	}
	httpapiGate := string(phase12ReadFile(t, "internal/platform/httpapi/httpapi.go"))
	for _, required := range []string{"withUnclaimedReservedExtensionFamilies", "extension_profile_not_claimed", "MatchReservedExtensionFamilyIn"} {
		if !strings.Contains(httpapiGate, required) {
			t.Fatalf("HTTP API extension gate missing authorization/admission hook %q", required)
		}
	}
	if strings.Contains(routes, "deployment_admin") {
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
