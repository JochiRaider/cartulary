import {
  type Contributor,
  type DecodeFailure,
  type Decoder,
  type EdgeAnnotation,
  type Filter,
  type GraphContributorQueryContinuation,
  type GraphContributorQueryRequest,
  type GraphContributorQueryResult,
  type GraphProjectionEdge,
  type GraphProjectionVertex,
  type GraphQueryRequest,
  type GraphQueryResult,
  type GraphSelector,
  type GraphSemanticQuery,
  getNetworkFlowErrorRegistry,
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
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
  type PagingMeta,
  type RejectedRowDiagnostic,
  type RejectedRowsQueryContinuation,
  type RejectedRowsQueryRequest,
  type RejectedRowsQueryResult,
  type Sort,
  type SourceProfileList,
  type TableList,
  type TableMutationResult,
  type TableQueryContinuation,
  type TableQueryRequest,
  type TableQueryResult,
  type TableRenameRequest,
  type TableScope,
  type TableSoftDeleteRequest,
} from "@cartulary/protocol-ts";

export type { NetworkFlowRow, NetworkFlowRowRef, NetworkFlowTable };

export type NetworkFlowContributor = Contributor;
export type NetworkFlowContributorResult = GraphContributorQueryResult;
export type NetworkFlowDiagnostic = RejectedRowDiagnostic;
export type NetworkFlowEdgeAnnotation = EdgeAnnotation;
export type NetworkFlowGraphResult = GraphQueryResult;
export type NetworkFlowGraphEdge = GraphProjectionEdge;
export type NetworkFlowGraphVertex = GraphProjectionVertex;
export type NetworkFlowGraphSelector = GraphSelector;
export type NetworkFlowGraphSemanticQuery = GraphSemanticQuery;
export type NetworkFlowGraphQueryRequest = GraphQueryRequest;
export type NetworkFlowContributorQueryRequest = GraphContributorQueryRequest;
export type NetworkFlowContributorQueryContinuation =
  GraphContributorQueryContinuation;
export type NetworkFlowContributorPageRequest =
  | GraphContributorQueryRequest
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
export type NetworkFlowSourceProfileList = SourceProfileList;
export type NetworkFlowTableScope = TableScope;
export type NetworkFlowTableQueryContinuation = TableQueryContinuation;
export type NetworkFlowTableQueryRequest = TableQueryRequest;
export type NetworkFlowTableMutationResult = TableMutationResult;
export type NetworkFlowTableRenameRequest = TableRenameRequest;
export type NetworkFlowTableSoftDeleteRequest = TableSoftDeleteRequest;

export { networkFlowContractDescriptor };
export const networkFlowMappingMetadata = networkFlowMappingRegistry;
export const networkFlowErrorMetadata = getNetworkFlowErrorRegistry();
export const networkFlowPresentationMetadata = networkFlowPresentationRegistry;

const supportedNetworkFlowContractMajors = new Set([2]);

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

export function decodeNetworkFlowGraphResult(value: unknown): GraphQueryResult {
  return decodeOrThrow(networkFlowDecoders.graphQueryResult, value);
}

export function decodeNetworkFlowContributorResult(
  value: unknown,
): GraphContributorQueryResult {
  return decodeOrThrow(networkFlowDecoders.graphContributorQueryResult, value);
}

export function decodeNetworkFlowIndicatorLinkResult(
  value: unknown,
): IndicatorLinkResult {
  return decodeOrThrow(networkFlowDecoders.indicatorLinkResult, value);
}

export function decodeNetworkFlowSourceProfileList(
  value: unknown,
): SourceProfileList {
  return decodeOrThrow(networkFlowDecoders.sourceProfileList, value);
}

export function decodeNetworkFlowImportPreviewResult(
  value: unknown,
): ImportPreviewResult {
  return decodeOrThrow(networkFlowDecoders.importPreviewResult, value);
}
