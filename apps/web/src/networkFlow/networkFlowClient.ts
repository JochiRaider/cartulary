import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import { apiPath, clientTxnID, fetchJSON } from "../services/browserApi";
import type {
  NetworkFlowContributorPageRequest,
  NetworkFlowContributorResult,
  NetworkFlowDiagnostic,
  NetworkFlowFilter,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkRequest,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorSelector,
  NetworkFlowIndicatorTarget,
  NetworkFlowPaging,
  NetworkFlowRow,
  NetworkFlowSavedGraph,
  NetworkFlowSavedGraphAccepted,
  NetworkFlowSavedGraphContributorQueryRequest,
  NetworkFlowSavedGraphContributorResult,
  NetworkFlowSavedGraphCreateRequest,
  NetworkFlowSavedGraphRefreshRequest,
  NetworkFlowSavedGraphRenameRequest,
  NetworkFlowSavedGraphResult,
  NetworkFlowSavedGraphRetireRequest,
  NetworkFlowTable,
  NetworkFlowTableRenameRequest,
  NetworkFlowTableScope,
  NetworkFlowTableSoftDeleteRequest,
} from "../services/networkFlowContractAdapter";
import {
  decodeNetworkFlowContributorResult,
  decodeNetworkFlowGraphResult,
  decodeNetworkFlowIndicatorLinkResult,
  decodeNetworkFlowRejectedRowsQueryResult,
  decodeNetworkFlowSavedGraphAccepted,
  decodeNetworkFlowSavedGraphContributorResult,
  decodeNetworkFlowSavedGraphList,
  decodeNetworkFlowSavedGraphMutationResult,
  decodeNetworkFlowSavedGraphResult,
  decodeNetworkFlowSourceProfileList,
  decodeNetworkFlowTableList,
  decodeNetworkFlowTableMutationResult,
  decodeNetworkFlowTableQueryResult,
} from "../services/networkFlowContractAdapter";
import { networkFlowRequestError } from "./networkFlowErrors";
import type {
  NetworkFlowAcceptedPageRequest,
  NetworkFlowRejectedPageRequest,
} from "./networkFlowQueryModel";

export {
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../extensions/extensionWorkspaceIdentities";
export type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowDiagnostic,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphEdge,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphSemanticQuery,
  NetworkFlowGraphVertex,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorSelector,
  NetworkFlowIndicatorTarget,
  NetworkFlowPaging,
  NetworkFlowRow,
  NetworkFlowRowRef,
  NetworkFlowSavedGraph,
  NetworkFlowSavedGraphAccepted,
  NetworkFlowSavedGraphContributorResult,
  NetworkFlowSavedGraphResult,
  NetworkFlowTable,
  NetworkFlowTableScope,
} from "../services/networkFlowContractAdapter";

function networkFlowResponseData(payload: unknown): unknown {
  if (
    payload === null ||
    typeof payload !== "object" ||
    Array.isArray(payload) ||
    !("data" in payload)
  ) {
    throw new Error("invalid_network_flow_success_envelope");
  }
  return payload.data;
}

export async function listNetworkFlowTables(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowTable[]> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables`,
    ),
    requestInit({}, options.signal),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowTableList(networkFlowResponseData(result.payload))
    .tables;
}

export async function renameNetworkFlowTable(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly baseTableVersion: number;
  readonly displayName: string;
  readonly incidentId: string;
  readonly tableId: string;
}): Promise<NetworkFlowTable> {
  const request: NetworkFlowTableRenameRequest = {
    client_txn_id: clientTxnID("nf-table-rename"),
    base_table_version: options.baseTableVersion,
    display_name: options.displayName,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    tableURL(options),
    requestInit({ method: "PATCH", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowTableMutationResult(
    networkFlowResponseData(result.payload),
  ).table;
}

export async function softDeleteNetworkFlowTable(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly baseTableVersion: number;
  readonly incidentId: string;
  readonly tableId: string;
}): Promise<NetworkFlowTable> {
  const request: NetworkFlowTableSoftDeleteRequest = {
    client_txn_id: clientTxnID("nf-table-delete"),
    base_table_version: options.baseTableVersion,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    tableURL(options),
    requestInit({ method: "DELETE", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowTableMutationResult(
    networkFlowResponseData(result.payload),
  ).table;
}

export async function queryNetworkFlowTable(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowAcceptedPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly rows: NetworkFlowRow[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables/${options.tableId}/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify(options.request),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  const response = decodeNetworkFlowTableQueryResult(
    networkFlowResponseData(result.payload),
  );
  return { rows: response.rows, paging: response.meta.paging };
}

export async function queryNetworkFlowRejectedRows(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowRejectedPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly diagnostics: NetworkFlowDiagnostic[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables/${options.tableId}/rejected-rows/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify(options.request),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  const response = decodeNetworkFlowRejectedRowsQueryResult(
    networkFlowResponseData(result.payload),
  );
  return {
    diagnostics: response.diagnostics,
    paging: response.meta.paging,
  };
}

export async function queryNetworkFlowGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly filters: readonly NetworkFlowFilter[];
  readonly incidentId: string;
  readonly aggregation: NetworkFlowGraphQueryRequest["aggregation"];
  readonly tableScope: NetworkFlowTableScope;
  readonly timeRange: NonNullable<
    NetworkFlowGraphQueryRequest["time_range"]
  > | null;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowGraphResult> {
  if (
    options.aggregation.mode === "time_bucket_v1" &&
    (options.timeRange?.start_utc == null || options.timeRange.end_utc == null)
  ) {
    throw new Error("network_flow_complete_graph_time_range_required");
  }
  const request = {
    schema_id: "cartulary.network_flow.graph_query_request.v2",
    table_scope: options.tableScope,
    ...(options.filters.length === 0 ? {} : { filters: [...options.filters] }),
    ...(options.timeRange === null ? {} : { time_range: options.timeRange }),
    aggregation: options.aggregation,
  } as NetworkFlowGraphQueryRequest;
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/graphs/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify(request),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowGraphResult(networkFlowResponseData(result.payload));
}

export async function queryNetworkFlowContributors(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly request: NetworkFlowContributorPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowContributorResult> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/graphs/contributors/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify(options.request),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowContributorResult(
    networkFlowResponseData(result.payload),
  );
}

export async function listNetworkFlowSavedGraphs(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowSavedGraph[]> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    graphViewsURL(options),
    requestInit({ method: "GET" }, options.signal),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphList(
    networkFlowResponseData(result.payload),
  ).graph_views;
}

export async function createNetworkFlowSavedGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly displayName: string;
  readonly incidentId: string;
  readonly semanticQuery: NetworkFlowGraphResult["semantic_query"];
}): Promise<NetworkFlowSavedGraphAccepted> {
  const request: NetworkFlowSavedGraphCreateRequest = {
    schema_id: "cartulary.network_flow.graph_view_create_request.v2",
    client_txn_id: clientTxnID("nf-graph-view-create"),
    display_name: options.displayName,
    semantic_query: options.semanticQuery,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    graphViewsURL(options),
    requestInit({ method: "POST", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphAccepted(
    networkFlowResponseData(result.payload),
  );
}

export async function renameNetworkFlowSavedGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly baseGraphViewVersion: number;
  readonly displayName: string;
  readonly graphViewId: string;
  readonly incidentId: string;
}): Promise<NetworkFlowSavedGraph> {
  const request: NetworkFlowSavedGraphRenameRequest = {
    schema_id: "cartulary.network_flow.graph_view_rename_request.v1",
    client_txn_id: clientTxnID("nf-graph-view-rename"),
    base_graph_view_version: options.baseGraphViewVersion,
    display_name: options.displayName,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    graphViewURL(options),
    requestInit({ method: "PATCH", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphMutationResult(
    networkFlowResponseData(result.payload),
  ).graph_view;
}

export async function refreshNetworkFlowSavedGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly baseGraphViewVersion: number;
  readonly graphViewId: string;
  readonly incidentId: string;
}): Promise<NetworkFlowSavedGraphAccepted> {
  const request: NetworkFlowSavedGraphRefreshRequest = {
    schema_id: "cartulary.network_flow.graph_view_refresh_request.v1",
    client_txn_id: clientTxnID("nf-graph-view-refresh"),
    base_graph_view_version: options.baseGraphViewVersion,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    `${graphViewURL(options)}/refresh`,
    requestInit({ method: "POST", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphAccepted(
    networkFlowResponseData(result.payload),
  );
}

export async function retireNetworkFlowSavedGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly baseGraphViewVersion: number;
  readonly graphViewId: string;
  readonly incidentId: string;
}): Promise<NetworkFlowSavedGraph> {
  const request: NetworkFlowSavedGraphRetireRequest = {
    schema_id: "cartulary.network_flow.graph_view_retire_request.v1",
    client_txn_id: clientTxnID("nf-graph-view-retire"),
    base_graph_view_version: options.baseGraphViewVersion,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    graphViewURL(options),
    requestInit({ method: "DELETE", body: JSON.stringify(request) }, undefined),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphMutationResult(
    networkFlowResponseData(result.payload),
  ).graph_view;
}

export async function getNetworkFlowSavedGraphResult(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly graphViewId: string;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowSavedGraphResult> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    `${graphViewURL(options)}/result`,
    requestInit({ method: "GET" }, options.signal),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphResult(
    networkFlowResponseData(result.payload),
  );
}

export async function queryNetworkFlowSavedGraphContributors(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly graphViewId: string;
  readonly incidentId: string;
  readonly projectionResultId: string;
  readonly selector: NetworkFlowSavedGraphContributorQueryRequest["selector"];
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowSavedGraphContributorResult> {
  const request: NetworkFlowSavedGraphContributorQueryRequest = {
    schema_id: "cartulary.network_flow.graph_view_contributor_query_request.v2",
    projection_result_id: options.projectionResultId,
    selector: options.selector,
    limit: 100,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    `${graphViewURL(options)}/contributors/query`,
    requestInit(
      { method: "POST", body: JSON.stringify(request) },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowSavedGraphContributorResult(
    networkFlowResponseData(result.payload),
  );
}

export async function linkNetworkFlowIndicator(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly selector: NetworkFlowIndicatorSelector;
  readonly target: NetworkFlowIndicatorTarget;
  readonly confirmExactValue: string;
}): Promise<NetworkFlowIndicatorLinkResult> {
  const request: NetworkFlowIndicatorLinkRequest = {
    schema_id: "cartulary.network_flow.indicator_link_request.v1",
    client_txn_id: clientTxnID("nf-indicator-link"),
    selector: options.selector,
    target: options.target,
    observation_mode: "binding_only",
    confirm_exact_value: options.confirmExactValue,
  };
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/indicator-links`,
    ),
    {
      method: "POST",
      body: JSON.stringify(request),
    },
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  return decodeNetworkFlowIndicatorLinkResult(
    networkFlowResponseData(result.payload),
  );
}

export async function getNetworkFlowBindingSourceRowLimit(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<number> {
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/source-profiles`,
    ),
    requestInit({ method: "GET" }, options.signal),
  );
  if (!result.ok) {
    throw networkFlowRequestError(result.status, result.payload);
  }
  const response = decodeNetworkFlowSourceProfileList(
    networkFlowResponseData(result.payload),
  );
  return response.effective_limits["network_flow.max_binding_source_row_refs"];
}

function tableURL(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
}): string {
  return apiPath(
    options.apiBase,
    `/api/v1/incidents/${options.incidentId}/network-flow/tables/${options.tableId}`,
  );
}

function graphViewsURL(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
}): string {
  return apiPath(
    options.apiBase,
    `/api/v1/incidents/${options.incidentId}/network-flow/graph-views`,
  );
}

function graphViewURL(options: {
  readonly apiBase?: string | undefined;
  readonly graphViewId: string;
  readonly incidentId: string;
}): string {
  return `${graphViewsURL(options)}/${options.graphViewId}`;
}

function requestInit(
  init: RequestInit,
  signal: AbortSignal | undefined,
): RequestInit {
  return signal === undefined ? init : { ...init, signal };
}

function fetchNetworkFlowJSON<T>(
  availability: ExtensionAvailabilityController,
  input: RequestInfo | URL,
  init?: RequestInit,
) {
  return availability.runProfileRequest(
    networkFlowActivityProfileId,
    networkFlowRouteFamily,
    () => fetchJSON<T>(input, init),
  );
}
