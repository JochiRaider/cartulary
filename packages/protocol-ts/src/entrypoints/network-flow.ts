import { networkFlowContractDescriptor } from "../generated/network-flow-descriptor.js";
import { networkFlowErrorRegistry } from "../generated/network-flow-error-registry.js";
import {
  networkFlowMappingRegistry,
  networkFlowTimestampMetadata,
} from "../generated/network-flow-mapping-registry.js";
import {
  networkFlowPresentationRegistry,
  networkFlowQueryMetadata,
} from "../generated/network-flow-presentation.js";
import type {
  ApprovedMapping,
  Filter,
  GraphContributorQueryResultV3,
  GraphQueryRequestV3,
  GraphQueryResultV3,
  GraphViewAcceptedV5,
  GraphViewContributorQueryResultV3,
  GraphViewGetV5,
  GraphViewListV5,
  GraphViewMutationResultV5,
  GraphViewResultV5,
  ImportPreviewResult,
  ImportSourceInspectionResult,
  IndicatorLinkResult,
  MappingCandidate,
  RejectedRowsQueryResult,
  SourceProfileListV3,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "../generated/network-flow-types.js";
import {
  validateCartularyNetworkFlowApprovedMappingV2,
  validateCartularyNetworkFlowFilterV2,
  validateCartularyNetworkFlowGraphContributorQueryResultV3,
  validateCartularyNetworkFlowGraphQueryRequestV3,
  validateCartularyNetworkFlowGraphQueryResultV3,
  validateCartularyNetworkFlowGraphViewAcceptedV5,
  validateCartularyNetworkFlowGraphViewContributorQueryResultV3,
  validateCartularyNetworkFlowGraphViewGetV5,
  validateCartularyNetworkFlowGraphViewListV5,
  validateCartularyNetworkFlowGraphViewMutationResultV5,
  validateCartularyNetworkFlowGraphViewResultV5,
  validateCartularyNetworkFlowImportPreviewResultV2,
  validateCartularyNetworkFlowImportSourceInspectionResultV2,
  validateCartularyNetworkFlowIndicatorLinkResultV2,
  validateCartularyNetworkFlowMappingCandidateV2,
  validateCartularyNetworkFlowRejectedRowsQueryResultV2,
  validateCartularyNetworkFlowSourceProfileListV3,
  validateCartularyNetworkFlowTableListV2,
  validateCartularyNetworkFlowTableMutationResultV2,
  validateCartularyNetworkFlowTableQueryResultV2,
} from "../generated/network-flow-validators.js";
import { createDecoder } from "../internal/decoder.js";

export type * from "../generated/network-flow-types.js";
export type {
  DecodeFailure,
  Decoder,
} from "../internal/decoder.js";
export {
  networkFlowContractDescriptor,
  networkFlowErrorRegistry,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
  networkFlowQueryMetadata,
  networkFlowTimestampMetadata,
};

export const networkFlowDecoders = Object.freeze({
  mappingCandidate: createDecoder<MappingCandidate>(
    "cartulary.network_flow.mapping_candidate.v2",
    validateCartularyNetworkFlowMappingCandidateV2,
  ),
  approvedMapping: createDecoder<ApprovedMapping>(
    "cartulary.network_flow.approved_mapping.v2",
    validateCartularyNetworkFlowApprovedMappingV2,
  ),
  graphQueryRequest: createDecoder<GraphQueryRequestV3>(
    "cartulary.network_flow.graph_query_request.v3",
    validateCartularyNetworkFlowGraphQueryRequestV3,
  ),
  filter: createDecoder<Filter>(
    "cartulary.network_flow.filter.v2",
    validateCartularyNetworkFlowFilterV2,
  ),
  tableList: createDecoder<TableList>(
    "cartulary.network_flow_table_list.v2",
    validateCartularyNetworkFlowTableListV2,
  ),
  tableMutationResult: createDecoder<TableMutationResult>(
    "cartulary.network_flow_table_mutation_result.v2",
    validateCartularyNetworkFlowTableMutationResultV2,
  ),
  tableQueryResult: createDecoder<TableQueryResult>(
    "cartulary.network_flow.table_query_result.v2",
    validateCartularyNetworkFlowTableQueryResultV2,
  ),
  rejectedRowsQueryResult: createDecoder<RejectedRowsQueryResult>(
    "cartulary.network_flow.rejected_rows_query_result.v2",
    validateCartularyNetworkFlowRejectedRowsQueryResultV2,
  ),
  sourceProfileList: createDecoder<SourceProfileListV3>(
    "cartulary.network_flow.source_profile_list.v3",
    validateCartularyNetworkFlowSourceProfileListV3,
  ),
  graphQueryResult: createDecoder<GraphQueryResultV3>(
    "cartulary.network_flow.graph_query_result.v3",
    validateCartularyNetworkFlowGraphQueryResultV3,
  ),
  graphContributorQueryResult: createDecoder<GraphContributorQueryResultV3>(
    "cartulary.network_flow.graph_contributor_query_result.v3",
    validateCartularyNetworkFlowGraphContributorQueryResultV3,
  ),
  graphViewList: createDecoder<GraphViewListV5>(
    "cartulary.network_flow.graph_view_list.v5",
    validateCartularyNetworkFlowGraphViewListV5,
  ),
  graphViewGet: createDecoder<GraphViewGetV5>(
    "cartulary.network_flow.graph_view_get.v5",
    validateCartularyNetworkFlowGraphViewGetV5,
  ),
  graphViewAccepted: createDecoder<GraphViewAcceptedV5>(
    "cartulary.network_flow.graph_view_accepted.v5",
    validateCartularyNetworkFlowGraphViewAcceptedV5,
  ),
  graphViewMutationResult: createDecoder<GraphViewMutationResultV5>(
    "cartulary.network_flow.graph_view_mutation_result.v5",
    validateCartularyNetworkFlowGraphViewMutationResultV5,
  ),
  graphViewResult: createDecoder<GraphViewResultV5>(
    "cartulary.network_flow.graph_view_result.v5",
    validateCartularyNetworkFlowGraphViewResultV5,
  ),
  graphViewContributorQueryResult:
    createDecoder<GraphViewContributorQueryResultV3>(
      "cartulary.network_flow.graph_view_contributor_query_result.v3",
      validateCartularyNetworkFlowGraphViewContributorQueryResultV3,
    ),
  indicatorLinkResult: createDecoder<IndicatorLinkResult>(
    "cartulary.network_flow_indicator_link_result.v2",
    validateCartularyNetworkFlowIndicatorLinkResultV2,
  ),
  importSourceInspectionResult: createDecoder<ImportSourceInspectionResult>(
    "cartulary.network_flow.import_source_inspection_result.v2",
    validateCartularyNetworkFlowImportSourceInspectionResultV2,
  ),
  importPreviewResult: createDecoder<ImportPreviewResult>(
    "cartulary.network_flow.import_preview_result.v2",
    validateCartularyNetworkFlowImportPreviewResultV2,
  ),
});
