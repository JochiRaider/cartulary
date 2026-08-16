import { networkFlowContractDescriptor } from "../generated/network-flow-descriptor.js";
import { networkFlowErrorRegistry } from "../generated/network-flow-error-registry.js";
import { networkFlowMappingRegistry } from "../generated/network-flow-mapping-registry.js";
import { networkFlowPresentationRegistry } from "../generated/network-flow-presentation.js";
import type {
  GraphContributorQueryResult,
  GraphQueryResult,
  GraphViewAccepted,
  GraphViewContributorQueryResult,
  GraphViewGet,
  GraphViewList,
  GraphViewMutationResult,
  GraphViewResult,
  ImportPreviewResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  SourceProfileList,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "../generated/network-flow-types.js";
import {
  validateCartularyNetworkFlowGraphContributorQueryResultV1,
  validateCartularyNetworkFlowGraphQueryResultV1,
  validateCartularyNetworkFlowGraphViewAcceptedV1,
  validateCartularyNetworkFlowGraphViewContributorQueryResultV1,
  validateCartularyNetworkFlowGraphViewGetV1,
  validateCartularyNetworkFlowGraphViewListV1,
  validateCartularyNetworkFlowGraphViewMutationResultV1,
  validateCartularyNetworkFlowGraphViewResultV1,
  validateCartularyNetworkFlowImportPreviewResultV1,
  validateCartularyNetworkFlowIndicatorLinkResultV1,
  validateCartularyNetworkFlowRejectedRowsQueryResultV1,
  validateCartularyNetworkFlowSourceProfileListV1,
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
  sourceProfileList: createDecoder<SourceProfileList>(
    "cartulary.network_flow.source_profile_list.v1",
    validateCartularyNetworkFlowSourceProfileListV1,
  ),
  graphQueryResult: createDecoder<GraphQueryResult>(
    "cartulary.network_flow.graph_query_result.v1",
    validateCartularyNetworkFlowGraphQueryResultV1,
  ),
  graphContributorQueryResult: createDecoder<GraphContributorQueryResult>(
    "cartulary.network_flow.graph_contributor_query_result.v1",
    validateCartularyNetworkFlowGraphContributorQueryResultV1,
  ),
  graphViewList: createDecoder<GraphViewList>(
    "cartulary.network_flow.graph_view_list.v1",
    validateCartularyNetworkFlowGraphViewListV1,
  ),
  graphViewGet: createDecoder<GraphViewGet>(
    "cartulary.network_flow.graph_view_get.v1",
    validateCartularyNetworkFlowGraphViewGetV1,
  ),
  graphViewAccepted: createDecoder<GraphViewAccepted>(
    "cartulary.network_flow.graph_view_accepted.v1",
    validateCartularyNetworkFlowGraphViewAcceptedV1,
  ),
  graphViewMutationResult: createDecoder<GraphViewMutationResult>(
    "cartulary.network_flow.graph_view_mutation_result.v1",
    validateCartularyNetworkFlowGraphViewMutationResultV1,
  ),
  graphViewResult: createDecoder<GraphViewResult>(
    "cartulary.network_flow.graph_view_result.v1",
    validateCartularyNetworkFlowGraphViewResultV1,
  ),
  graphViewContributorQueryResult:
    createDecoder<GraphViewContributorQueryResult>(
      "cartulary.network_flow.graph_view_contributor_query_result.v1",
      validateCartularyNetworkFlowGraphViewContributorQueryResultV1,
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
