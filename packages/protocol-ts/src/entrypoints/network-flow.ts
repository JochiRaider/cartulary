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
  Filter,
  GraphContributorQueryResultV2,
  GraphQueryRequestV2,
  GraphQueryResultV2,
  GraphViewAcceptedV4,
  GraphViewContributorQueryResultV2,
  GraphViewGetV4,
  GraphViewListV4,
  GraphViewMutationResultV4,
  GraphViewResultV4,
  ImportPreviewResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  SourceProfileListV2,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "../generated/network-flow-types.js";
import {
  validateCartularyNetworkFlowFilterV1,
  validateCartularyNetworkFlowGraphContributorQueryResultV2,
  validateCartularyNetworkFlowGraphQueryRequestV2,
  validateCartularyNetworkFlowGraphQueryResultV2,
  validateCartularyNetworkFlowGraphViewAcceptedV4,
  validateCartularyNetworkFlowGraphViewContributorQueryResultV2,
  validateCartularyNetworkFlowGraphViewGetV4,
  validateCartularyNetworkFlowGraphViewListV4,
  validateCartularyNetworkFlowGraphViewMutationResultV4,
  validateCartularyNetworkFlowGraphViewResultV4,
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
  networkFlowQueryMetadata,
  networkFlowTimestampMetadata,
};

export const networkFlowDecoders = Object.freeze({
  graphQueryRequest: createDecoder<GraphQueryRequestV2>(
    "cartulary.network_flow.graph_query_request.v2",
    validateCartularyNetworkFlowGraphQueryRequestV2,
  ),
  filter: createDecoder<Filter>(
    "cartulary.network_flow.filter.v1",
    validateCartularyNetworkFlowFilterV1,
  ),
  tableList: createDecoder<TableList>(
    "cartulary.network_flow_table_list.v1",
    validateCartularyNetworkFlowTableListV1,
  ),
  tableMutationResult: createDecoder<TableMutationResult>(
    "cartulary.network_flow_table_mutation_result.v1",
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
  graphViewList: createDecoder<GraphViewListV4>(
    "cartulary.network_flow.graph_view_list.v4",
    validateCartularyNetworkFlowGraphViewListV4,
  ),
  graphViewGet: createDecoder<GraphViewGetV4>(
    "cartulary.network_flow.graph_view_get.v4",
    validateCartularyNetworkFlowGraphViewGetV4,
  ),
  graphViewAccepted: createDecoder<GraphViewAcceptedV4>(
    "cartulary.network_flow.graph_view_accepted.v4",
    validateCartularyNetworkFlowGraphViewAcceptedV4,
  ),
  graphViewMutationResult: createDecoder<GraphViewMutationResultV4>(
    "cartulary.network_flow.graph_view_mutation_result.v4",
    validateCartularyNetworkFlowGraphViewMutationResultV4,
  ),
  graphViewResult: createDecoder<GraphViewResultV4>(
    "cartulary.network_flow.graph_view_result.v4",
    validateCartularyNetworkFlowGraphViewResultV4,
  ),
  graphViewContributorQueryResult:
    createDecoder<GraphViewContributorQueryResultV2>(
      "cartulary.network_flow.graph_view_contributor_query_result.v2",
      validateCartularyNetworkFlowGraphViewContributorQueryResultV2,
    ),
  indicatorLinkResult: createDecoder<IndicatorLinkResult>(
    "cartulary.network_flow_indicator_link_result.v1",
    validateCartularyNetworkFlowIndicatorLinkResultV1,
  ),
  importPreviewResult: createDecoder<ImportPreviewResult>(
    "cartulary.network_flow.import_preview_result.v1",
    validateCartularyNetworkFlowImportPreviewResultV1,
  ),
});
