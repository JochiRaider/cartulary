import { networkFlowContractDescriptor } from "../generated/network-flow-descriptor.js";
import { networkFlowMappingRegistry } from "../generated/network-flow-mapping-registry.js";
import { networkFlowPresentationRegistry } from "../generated/network-flow-presentation.js";
import type {
  GraphContributorQueryResult,
  GraphQueryResult,
  ImportPreviewResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  SourceProfileList,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "../generated/network-flow-types.js";

import { parseContractArtifact } from "./contractArtifacts.js";
import { createGeneratedDecoder } from "./runtimeValidation.js";

export type * from "../generated/network-flow-types.js";
export {
  networkFlowContractDescriptor,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
};

export type NetworkFlowErrorRetryAction =
  | "correct_request"
  | "refresh_resource"
  | "restart_query"
  | "reduce_scope_or_limits"
  | "retry_with_backoff"
  | "do_not_retry";

export type NetworkFlowErrorContract = {
  readonly code: string;
  readonly http_status: number | null;
  readonly retry_action: NetworkFlowErrorRetryAction;
  readonly scope: string;
};

export type NetworkFlowErrorRegistry = {
  readonly schema_id: "cartulary.network_flow_error_contracts.v1";
  readonly errors: readonly NetworkFlowErrorContract[];
};

export function getNetworkFlowErrorRegistry(): NetworkFlowErrorRegistry {
  return parseContractArtifact<NetworkFlowErrorRegistry>(
    "contracts/network-flow/errors.v1.json",
  );
}

export const networkFlowDecoders = Object.freeze({
  tableList: createGeneratedDecoder<TableList>(
    "cartulary.network_flow.table_list.v1",
  ),
  tableMutationResult: createGeneratedDecoder<TableMutationResult>(
    "cartulary.network_flow.table_mutation_result.v1",
  ),
  tableQueryResult: createGeneratedDecoder<TableQueryResult>(
    "cartulary.network_flow.table_query_result.v1",
  ),
  rejectedRowsQueryResult: createGeneratedDecoder<RejectedRowsQueryResult>(
    "cartulary.network_flow.rejected_rows_query_result.v1",
  ),
  sourceProfileList: createGeneratedDecoder<SourceProfileList>(
    "cartulary.network_flow.source_profile_list.v1",
  ),
  graphQueryResult: createGeneratedDecoder<GraphQueryResult>(
    "cartulary.network_flow.graph_query_result.v1",
  ),
  graphContributorQueryResult:
    createGeneratedDecoder<GraphContributorQueryResult>(
      "cartulary.network_flow.graph_contributor_query_result.v1",
    ),
  indicatorLinkResult: createGeneratedDecoder<IndicatorLinkResult>(
    "cartulary.network_flow.indicator_link_result.v1",
  ),
  importPreviewResult: createGeneratedDecoder<ImportPreviewResult>(
    "cartulary.network_flow.import_preview_result.v1",
  ),
});
