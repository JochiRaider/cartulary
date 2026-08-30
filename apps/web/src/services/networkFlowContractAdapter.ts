import {
  type Contributor,
  type DecodeFailure,
  type Decoder,
  type EdgeAnnotationV2,
  type Filter,
  type GraphContributorQueryContinuation,
  type GraphContributorQueryRequestV2,
  type GraphContributorQueryResultV2,
  type GraphProjectionEdge,
  type GraphProjectionVertex,
  type GraphQueryRequestV2,
  type GraphQueryResultV2,
  type GraphSelectorV2,
  type GraphSemanticQueryV2,
  type GraphViewAcceptedV3,
  type GraphViewContributorQueryRequestV2,
  type GraphViewContributorQueryResultV2,
  type GraphViewCreateRequestV2,
  type GraphViewListV3,
  type GraphViewMutationResultV3,
  type GraphViewRefreshRequest,
  type GraphViewRenameRequest,
  type GraphViewResultV3,
  type GraphViewRetireRequest,
  type GraphViewV3,
  type ImportPreviewResult,
  type IndicatorLinkRequest,
  type IndicatorLinkResult,
  type IndicatorSelector,
  type IndicatorTarget,
  type MappingCandidate,
  type NetworkFlowRow,
  type NetworkFlowRowRef,
  type NetworkFlowTable,
  networkFlowContractDescriptor,
  networkFlowDecoders,
  networkFlowErrorRegistry,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
  type PagingMeta,
  type RejectedRowDiagnostic,
  type RejectedRowsQueryContinuation,
  type RejectedRowsQueryRequest,
  type RejectedRowsQueryResult,
  type Sort,
  type SourceProfileListV2,
  type TableList,
  type TableMutationResult,
  type TableQueryContinuation,
  type TableQueryRequest,
  type TableQueryResult,
  type TableRenameRequest,
  type TableScope,
  type TableSoftDeleteRequest,
} from "@cartulary/protocol-ts/network-flow";

export type { NetworkFlowRow, NetworkFlowRowRef, NetworkFlowTable };

export type NetworkFlowContributor = Contributor;
export type NetworkFlowContributorResult = GraphContributorQueryResultV2;
export type NetworkFlowDiagnostic = RejectedRowDiagnostic;
export type NetworkFlowEdgeAnnotation = EdgeAnnotationV2;
export type NetworkFlowGraphResult = GraphQueryResultV2;
export type NetworkFlowGraphEdge = GraphProjectionEdge;
export type NetworkFlowGraphVertex = GraphProjectionVertex;
export type NetworkFlowGraphSelector = GraphSelectorV2;
export type NetworkFlowGraphSemanticQuery = GraphSemanticQueryV2;
export type NetworkFlowGraphQueryRequest = GraphQueryRequestV2;
export type NetworkFlowSavedGraph = GraphViewV3;
export type NetworkFlowSavedGraphAccepted = GraphViewAcceptedV3;
export type NetworkFlowSavedGraphContributorQueryRequest =
  GraphViewContributorQueryRequestV2;
export type NetworkFlowSavedGraphContributorResult =
  GraphViewContributorQueryResultV2;
export type NetworkFlowSavedGraphCreateRequest = GraphViewCreateRequestV2;
export type NetworkFlowSavedGraphList = GraphViewListV3;
export type NetworkFlowSavedGraphMutationResult = GraphViewMutationResultV3;
export type NetworkFlowSavedGraphRefreshRequest = GraphViewRefreshRequest;
export type NetworkFlowSavedGraphRenameRequest = GraphViewRenameRequest;
export type NetworkFlowSavedGraphResult = GraphViewResultV3;
export type NetworkFlowSavedGraphRetireRequest = GraphViewRetireRequest;
export type NetworkFlowContributorQueryRequest = GraphContributorQueryRequestV2;
export type NetworkFlowContributorQueryContinuation =
  GraphContributorQueryContinuation;
export type NetworkFlowContributorPageRequest =
  | GraphContributorQueryRequestV2
  | GraphContributorQueryContinuation;
export type NetworkFlowIndicatorLinkResult = IndicatorLinkResult;
export type NetworkFlowIndicatorLinkRequest = IndicatorLinkRequest;
export type NetworkFlowIndicatorSelector = IndicatorSelector;
export type NetworkFlowIndicatorTarget = IndicatorTarget;
export type NetworkFlowImportPreviewResult = ImportPreviewResult;
export type NetworkFlowMappingCandidate = MappingCandidate;
export type NetworkFlowFilter = Filter;
export type NetworkFlowPaging = PagingMeta;
export type NetworkFlowRejectedRowsQueryContinuation =
  RejectedRowsQueryContinuation;
export type NetworkFlowRejectedRowsQueryRequest = RejectedRowsQueryRequest;
export type NetworkFlowSort = Sort;
export type NetworkFlowSourceProfileList = SourceProfileListV2;
export type NetworkFlowTableScope = TableScope;
export type NetworkFlowTableQueryContinuation = TableQueryContinuation;
export type NetworkFlowTableQueryRequest = TableQueryRequest;
export type NetworkFlowTableMutationResult = TableMutationResult;
export type NetworkFlowTableRenameRequest = TableRenameRequest;
export type NetworkFlowTableSoftDeleteRequest = TableSoftDeleteRequest;

export { networkFlowContractDescriptor };
export const networkFlowMappingMetadata = networkFlowMappingRegistry;
export const networkFlowMappingCandidateSchemaId =
  "cartulary.network_flow.mapping_candidate.v1";
export const networkFlowErrorMetadata = networkFlowErrorRegistry;
export const networkFlowPresentationMetadata = networkFlowPresentationRegistry;

const supportedNetworkFlowContractMajors = new Set([5]);

export function isSupportedNetworkFlowContract(
  descriptor: {
    readonly contract_major: number;
  } = networkFlowContractDescriptor,
): boolean {
  return supportedNetworkFlowContractMajors.has(descriptor.contract_major);
}

export function isNetworkFlowClaimed(
  profiles: readonly {
    readonly claimed: boolean;
    readonly profile_id: string;
  }[],
): boolean {
  return (
    isSupportedNetworkFlowContract() &&
    profiles.some(
      (profile) =>
        profile.profile_id === networkFlowContractDescriptor.profile_id &&
        profile.claimed,
    )
  );
}

export class NetworkFlowContractDecodeError extends Error {
  readonly failure: DecodeFailure;

  constructor(failure: DecodeFailure) {
    super(
      `Network Flow response failed ${failure.schemaId} validation at ${
        failure.instancePath || "/"
      } (${failure.reasonCategory})`,
    );
    this.name = "NetworkFlowContractDecodeError";
    this.failure = failure;
  }
}

function decodeOrThrow<T>(decoder: Decoder<T>, value: unknown): T {
  const result = decoder.decode(value);
  if (!result.ok) {
    throw new NetworkFlowContractDecodeError(result.error);
  }
  return result.value;
}

export function decodeNetworkFlowTableList(value: unknown): TableList {
  return decodeOrThrow(networkFlowDecoders.tableList, value);
}

export function decodeNetworkFlowTableMutationResult(
  value: unknown,
): TableMutationResult {
  return decodeOrThrow(networkFlowDecoders.tableMutationResult, value);
}

export function decodeNetworkFlowTableQueryResult(
  value: unknown,
): TableQueryResult {
  return decodeOrThrow(networkFlowDecoders.tableQueryResult, value);
}

export function decodeNetworkFlowRejectedRowsQueryResult(
  value: unknown,
): RejectedRowsQueryResult {
  return decodeOrThrow(networkFlowDecoders.rejectedRowsQueryResult, value);
}

export function decodeNetworkFlowGraphResult(
  value: unknown,
): GraphQueryResultV2 {
  return decodeOrThrow(networkFlowDecoders.graphQueryResult, value);
}

export function decodeNetworkFlowContributorResult(
  value: unknown,
): GraphContributorQueryResultV2 {
  return decodeOrThrow(networkFlowDecoders.graphContributorQueryResult, value);
}

export function decodeNetworkFlowSavedGraphList(
  value: unknown,
): GraphViewListV3 {
  return decodeOrThrow(networkFlowDecoders.graphViewList, value);
}

export function decodeNetworkFlowSavedGraphAccepted(
  value: unknown,
): GraphViewAcceptedV3 {
  return decodeOrThrow(networkFlowDecoders.graphViewAccepted, value);
}

export function decodeNetworkFlowSavedGraphMutationResult(
  value: unknown,
): GraphViewMutationResultV3 {
  return decodeOrThrow(networkFlowDecoders.graphViewMutationResult, value);
}

export function decodeNetworkFlowSavedGraphResult(
  value: unknown,
): GraphViewResultV3 {
  return decodeOrThrow(networkFlowDecoders.graphViewResult, value);
}

export function decodeNetworkFlowSavedGraphContributorResult(
  value: unknown,
): GraphViewContributorQueryResultV2 {
  return decodeOrThrow(
    networkFlowDecoders.graphViewContributorQueryResult,
    value,
  );
}

export function decodeNetworkFlowIndicatorLinkResult(
  value: unknown,
): IndicatorLinkResult {
  return decodeOrThrow(networkFlowDecoders.indicatorLinkResult, value);
}

export function decodeNetworkFlowSourceProfileList(
  value: unknown,
): SourceProfileListV2 {
  return decodeOrThrow(networkFlowDecoders.sourceProfileList, value);
}

export function decodeNetworkFlowImportPreviewResult(
  value: unknown,
): ImportPreviewResult {
  return decodeOrThrow(networkFlowDecoders.importPreviewResult, value);
}
