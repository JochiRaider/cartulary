import { apiPath, clientTxnID } from "../services/browserApi";
import type {
  NetworkFlowContributorPageRequest,
  NetworkFlowContributorResult,
  NetworkFlowDiagnostic,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkRequest,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorSelector,
  NetworkFlowIndicatorTarget,
  NetworkFlowPaging,
  NetworkFlowRow,
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
  decodeNetworkFlowSourceProfileList,
  decodeNetworkFlowTableList,
  decodeNetworkFlowTableMutationResult,
  decodeNetworkFlowTableQueryResult,
} from "../services/networkFlowContractAdapter";
import { fetchWorkbookJSON } from "../services/workbookApi";
import type { WorkbookSheetRef } from "../shared/workbookSheetRef";
import { networkFlowRequestError } from "./networkFlowErrors";
import type {
  NetworkFlowAcceptedPageRequest,
  NetworkFlowRejectedPageRequest,
} from "./networkFlowQueryModel";

export type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowDiagnostic,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphEdge,
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
  NetworkFlowTable,
  NetworkFlowTableScope,
} from "../services/networkFlowContractAdapter";

export const networkFlowActivityProfileId = "network_flow_activity";
export const networkAnalysisWorkspaceKey = "network_analysis";

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

export function networkAnalysisSheetRef(): Extract<
  WorkbookSheetRef,
  { kind: "extension_workspace" }
> {
  return {
    kind: "extension_workspace",
    extension_profile_id: networkFlowActivityProfileId,
    workspace_key: networkAnalysisWorkspaceKey,
  } as const;
}

export function networkAnalysisURLSelected(params: URLSearchParams): boolean {
  return (
    params.get("sheet_ref_kind") === "extension_workspace" &&
    params.get("extension_profile_id") === networkFlowActivityProfileId &&
    params.get("sheet_ref_id") === networkAnalysisWorkspaceKey
  );
}

export function writeNetworkAnalysisURL(incidentId: string): void {
  const next = new URLSearchParams(window.location.search);
  next.set("incident_id", incidentId);
  next.set("sheet_ref_kind", "extension_workspace");
  next.set("extension_profile_id", networkFlowActivityProfileId);
  next.set("sheet_ref_id", networkAnalysisWorkspaceKey);
  next.delete("view_schema_id");
  next.delete("workspace_key");
  next.delete("surface");
  window.history.replaceState({}, "", `/?${next.toString()}`);
}

export async function listNetworkFlowTables(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowTable[]> {
  const result = await fetchWorkbookJSON<unknown>(
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
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly baseTableVersion: number;
  readonly incidentId: string;
  readonly tableId: string;
}): Promise<NetworkFlowTable> {
  const request: NetworkFlowTableSoftDeleteRequest = {
    client_txn_id: clientTxnID("nf-table-delete"),
    base_table_version: options.baseTableVersion,
  };
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowAcceptedPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly rows: NetworkFlowRow[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowRejectedPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly diagnostics: NetworkFlowDiagnostic[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly filters: NonNullable<NetworkFlowGraphQueryRequest["filters"]>;
  readonly incidentId: string;
  readonly tableScope: NetworkFlowTableScope;
  readonly timeRange: NonNullable<
    NetworkFlowGraphQueryRequest["time_range"]
  > | null;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowGraphResult> {
  const request: NetworkFlowGraphQueryRequest = {
    schema_id: "cartulary.network_flow.graph_query_request.v1",
    table_scope: options.tableScope,
    ...(options.filters.length === 0 ? {} : { filters: [...options.filters] }),
    ...(options.timeRange === null ? {} : { time_range: options.timeRange }),
    aggregation: {
      mode: "default_flow_edge_v1",
      include_example_row_refs: true,
    },
  };
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly request: NetworkFlowContributorPageRequest;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowContributorResult> {
  const result = await fetchWorkbookJSON<unknown>(
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

export async function linkNetworkFlowIndicator(options: {
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
  const result = await fetchWorkbookJSON<unknown>(
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
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<number> {
  const result = await fetchWorkbookJSON<unknown>(
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

function requestInit(
  init: RequestInit,
  signal: AbortSignal | undefined,
): RequestInit {
  return signal === undefined ? init : { ...init, signal };
}
