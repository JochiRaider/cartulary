import { errorEnvelopeDecoder } from "@cartulary/protocol-ts/http";
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
  type GraphViewAcceptedV4,
  type GraphViewContributorQueryRequestV2,
  type GraphViewContributorQueryResultV2,
  type GraphViewCreateRequestV3,
  type GraphViewGetV4,
  type GraphViewListV4,
  type GraphViewMutationResultV4,
  type GraphViewRefreshRequest,
  type GraphViewRenameRequestV2,
  type GraphViewResultV4,
  type GraphViewRetireRequest,
  type GraphViewV4,
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
  networkFlowTimestampMetadata,
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
export type NetworkFlowSavedGraph = GraphViewV4;
export type NetworkFlowSavedGraphAccepted = GraphViewAcceptedV4;
export type NetworkFlowSavedGraphContributorQueryRequest =
  GraphViewContributorQueryRequestV2;
export type NetworkFlowSavedGraphContributorResult =
  GraphViewContributorQueryResultV2;
export type NetworkFlowSavedGraphCreateRequest = GraphViewCreateRequestV3;
export type NetworkFlowSavedGraphList = GraphViewListV4;
export type NetworkFlowSavedGraphMutationResult = GraphViewMutationResultV4;
export type NetworkFlowSavedGraphRefreshRequest = GraphViewRefreshRequest;
export type NetworkFlowSavedGraphRenameRequest = GraphViewRenameRequestV2;
export type NetworkFlowSavedGraphResult = GraphViewResultV4;
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
export { networkFlowTimestampMetadata };
export const networkFlowMappingCandidateSchemaId =
  "cartulary.network_flow.mapping_candidate.v1";
export const networkFlowErrorMetadata = networkFlowErrorRegistry;
export const networkFlowPresentationMetadata = networkFlowPresentationRegistry;

const supportedNetworkFlowContractMajors = new Set([6]);

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

/** Saved-route error certainty requires a complete envelope matching the HTTP status. */
export function validNetworkFlowErrorEnvelope(
  status: number,
  payload: unknown,
): boolean {
  const decoded = errorEnvelopeDecoder.decode(payload);
  return decoded.ok && decoded.value.error.status === status;
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
): GraphViewListV4 {
  const result = decodeOrThrow(networkFlowDecoders.graphViewList, value);
  for (const graph of result.graph_views) validateSavedGraphDeclaration(graph);
  return result;
}

export function decodeNetworkFlowSavedGraphAccepted(
  value: unknown,
): GraphViewAcceptedV4 {
  const result = decodeOrThrow(networkFlowDecoders.graphViewAccepted, value);
  validateSavedGraphDeclaration(result.graph_view);
  if (
    result.job.job_id !== result.graph_view.latest_job_id ||
    result.job.status_route !== `/api/v1/jobs/${result.job.job_id}`
  ) {
    rejectSavedGraphProjection("/job");
  }
  return result;
}

export function decodeNetworkFlowSavedGraphMutationResult(
  value: unknown,
): GraphViewMutationResultV4 {
  const result = decodeOrThrow(
    networkFlowDecoders.graphViewMutationResult,
    value,
  );
  validateSavedGraphDeclaration(result.graph_view);
  return result;
}

export function decodeNetworkFlowSavedGraphResult(
  value: unknown,
): GraphViewResultV4 {
  const result = decodeOrThrow(networkFlowDecoders.graphViewResult, value);
  validateSavedGraphDeclaration(result.graph_view);
  const binding = result.graph_view.selected_result_binding;
  const projection = result.result.graph_projection_result;
  if (
    binding === null ||
    projection.graph_view_id !== result.graph_view.graph_view_id ||
    projection.source_owner_id !== "network_flow_activity" ||
    savedGraphBindingMembers.some((key) => binding[key] !== projection[key])
  ) {
    rejectSavedGraphProjection("/result/graph_projection_result");
  }
  return result;
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

export function decodeNetworkFlowSavedGraphGet(value: unknown): GraphViewGetV4 {
  const result = decodeOrThrow(networkFlowDecoders.graphViewGet, value);
  validateSavedGraphDeclaration(result.graph_view);
  return result;
}

export const savedGraphBindingMembers = [
  "projection_result_id",
  "source_snapshot_id",
  "projection_schema_id",
  "projection_version",
  "normalized_configuration_sha256",
  "normalized_source_sha256",
  "canonical_output_sha256",
] as const;

export function savedGraphBindingIdentity(
  graph: NetworkFlowSavedGraph,
): string | null {
  const binding = graph.selected_result_binding;
  return binding === null
    ? null
    : JSON.stringify([
        graph.incident_id,
        graph.graph_view_id,
        ...savedGraphBindingMembers.map((key) => binding[key]),
      ]);
}

export function normalizeSavedGraphDisplayName(value: string):
  | { readonly ok: true; readonly name: string }
  | {
      readonly ok: false;
      readonly reason:
        | "forbidden_control"
        | "empty_display_name"
        | "display_name_too_long";
    } {
  if (
    Array.from(value).some((char) => {
      const cp = char.codePointAt(0) ?? 0;
      return cp <= 31 || (cp >= 127 && cp <= 159);
    }) ||
    /[\ud800-\udfff]/u.test(value)
  ) {
    return { ok: false, reason: "forbidden_control" };
  }
  // Exact owner whitespace: FEFF is intentionally not JavaScript trim space.
  const name = value
    .normalize("NFC")
    .replace(
      /^[\u0020\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+|[\u0020\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000]+$/gu,
      "",
    );
  if (name.length === 0) return { ok: false, reason: "empty_display_name" };
  if (new TextEncoder().encode(name).byteLength > 64)
    return { ok: false, reason: "display_name_too_long" };
  return { ok: true, name };
}

function validateSavedGraphDeclaration(graph: NetworkFlowSavedGraph): void {
  const name = normalizeSavedGraphDisplayName(graph.display_name);
  if (!name.ok || name.name !== graph.display_name)
    rejectSavedGraphProjection("/graph_view/display_name");
}

function rejectSavedGraphProjection(instancePath: string): never {
  throw new NetworkFlowContractDecodeError({
    boundary: "generated_protocol",
    schemaId: "cartulary.network_flow.graph_view.v4",
    instancePath,
    reasonCategory: "constraint_violation",
  });
}
