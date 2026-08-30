import { networkFlowContractDescriptor } from "../generated/network-flow-descriptor.js";
import { networkFlowErrorRegistry } from "../generated/network-flow-error-registry.js";
import { networkFlowMappingRegistry } from "../generated/network-flow-mapping-registry.js";
import { networkFlowPresentationRegistry } from "../generated/network-flow-presentation.js";
import type {
  GraphContributorQueryResultV2,
  GraphQueryResultV2,
  GraphViewAcceptedV3,
  GraphViewContributorQueryResultV2,
  GraphViewGetV3,
  GraphViewListV3,
  GraphViewMutationResultV3,
  GraphViewResultV3,
  ImportPreviewResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  SourceProfileListV2,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "../generated/network-flow-types.js";
import {
  validateCartularyNetworkFlowGraphContributorQueryResultV2,
  validateCartularyNetworkFlowGraphQueryResultV2,
  validateCartularyNetworkFlowGraphViewAcceptedV3,
  validateCartularyNetworkFlowGraphViewContributorQueryResultV2,
  validateCartularyNetworkFlowGraphViewGetV3,
  validateCartularyNetworkFlowGraphViewListV3,
  validateCartularyNetworkFlowGraphViewMutationResultV3,
  validateCartularyNetworkFlowGraphViewResultV3,
  validateCartularyNetworkFlowImportPreviewResultV1,
  validateCartularyNetworkFlowIndicatorLinkResultV1,
  validateCartularyNetworkFlowRejectedRowsQueryResultV1,
  validateCartularyNetworkFlowSourceProfileListV2,
  validateCartularyNetworkFlowTableListV1,
  validateCartularyNetworkFlowTableMutationResultV1,
  validateCartularyNetworkFlowTableQueryResultV1,
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
};

export const networkFlowDecoders = Object.freeze({
  tableList: createDecoder<TableList>(
    "cartulary.network_flow.table_list.v1",
    validateCartularyNetworkFlowTableListV1,
  ),
  tableMutationResult: createDecoder<TableMutationResult>(
    "cartulary.network_flow.table_mutation_result.v1",
    validateCartularyNetworkFlowTableMutationResultV1,
  ),
  tableQueryResult: createDecoder<TableQueryResult>(
    "cartulary.network_flow.table_query_result.v1",
    validateCartularyNetworkFlowTableQueryResultV1,
  ),
  rejectedRowsQueryResult: createDecoder<RejectedRowsQueryResult>(
    "cartulary.network_flow.rejected_rows_query_result.v1",
    validateCartularyNetworkFlowRejectedRowsQueryResultV1,
  ),
  sourceProfileList: createDecoder<SourceProfileListV2>(
    "cartulary.network_flow.source_profile_list.v2",
    validateCartularyNetworkFlowSourceProfileListV2,
  ),
  graphQueryResult: createDecoder<GraphQueryResultV2>(
    "cartulary.network_flow.graph_query_result.v2",
    validateCartularyNetworkFlowGraphQueryResultV2,
  ),
  graphContributorQueryResult: createDecoder<GraphContributorQueryResultV2>(
    "cartulary.network_flow.graph_contributor_query_result.v2",
    validateCartularyNetworkFlowGraphContributorQueryResultV2,
  ),
  graphViewList: createDecoder<GraphViewListV3>(
    "cartulary.network_flow.graph_view_list.v3",
    validateCartularyNetworkFlowGraphViewListV3,
  ),
  graphViewGet: createDecoder<GraphViewGetV3>(
    "cartulary.network_flow.graph_view_get.v3",
    validateCartularyNetworkFlowGraphViewGetV3,
  ),
  graphViewAccepted: createDecoder<GraphViewAcceptedV3>(
    "cartulary.network_flow.graph_view_accepted.v3",
    validateCartularyNetworkFlowGraphViewAcceptedV3,
  ),
  graphViewMutationResult: createDecoder<GraphViewMutationResultV3>(
    "cartulary.network_flow.graph_view_mutation_result.v3",
    validateCartularyNetworkFlowGraphViewMutationResultV3,
  ),
  graphViewResult: createDecoder<GraphViewResultV3>(
    "cartulary.network_flow.graph_view_result.v3",
    validateCartularyNetworkFlowGraphViewResultV3,
  ),
  graphViewContributorQueryResult:
    createDecoder<GraphViewContributorQueryResultV2>(
      "cartulary.network_flow.graph_view_contributor_query_result.v2",
      validateCartularyNetworkFlowGraphViewContributorQueryResultV2,
    ),
  indicatorLinkResult: createDecoder<IndicatorLinkResult>(
    "cartulary.network_flow.indicator_link_result.v1",
    validateCartularyNetworkFlowIndicatorLinkResultV1,
  ),
  importPreviewResult: createDecoder<ImportPreviewResult>(
    "cartulary.network_flow.import_preview_result.v1",
    validateCartularyNetworkFlowImportPreviewResultV1,
  ),
});
