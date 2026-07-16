import { apiPath, clientTxnID } from "../services/browserApi";
import type {
  NetworkFlowContributorQueryRequest,
  NetworkFlowContributorResult,
  NetworkFlowDiagnostic,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkRequest,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowPaging,
  NetworkFlowRow,
  NetworkFlowTable,
  NetworkFlowTableRenameRequest,
  NetworkFlowTableSoftDeleteRequest,
} from "../services/networkFlowContractAdapter";
import {
  decodeNetworkFlowContributorResult,
  decodeNetworkFlowGraphResult,
  decodeNetworkFlowIndicatorLinkResult,
  decodeNetworkFlowRejectedRowsQueryResult,
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
  NetworkFlowDiagnostic,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphResult,
  NetworkFlowGraphSemanticQuery,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowPaging,
  NetworkFlowRow,
  NetworkFlowRowRef,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";

export const networkFlowActivityProfileId = "network_flow_activity";
export const networkAnalysisWorkspaceKey = "network_analysis";

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
  return decodeNetworkFlowTableList(result.payload).tables;
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
  return decodeNetworkFlowTableMutationResult(result.payload).table;
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
  return decodeNetworkFlowTableMutationResult(result.payload).table;
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
  const response = decodeNetworkFlowTableQueryResult(result.payload);
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
  const response = decodeNetworkFlowRejectedRowsQueryResult(result.payload);
  return {
    diagnostics: response.diagnostics,
    paging: response.meta.paging,
  };
}

export async function queryNetworkFlowGraph(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableIds: readonly string[];
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowGraphResult> {
  const tableIds = [...new Set(options.tableIds)].sort();
  const [firstTableId, ...remainingTableIds] = tableIds;
  if (firstTableId === undefined) {
    throw new Error("Network Flow graph queries require at least one table.");
  }
  const request: NetworkFlowGraphQueryRequest = {
    schema_id: "cartulary.network_flow.graph_query_request.v1",
    table_scope:
      remainingTableIds.length === 0
        ? { mode: "active_table", active_table_id: firstTableId }
        : {
            mode: "selected_tables",
            selected_table_ids: [firstTableId, ...remainingTableIds],
          },
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
  return decodeNetworkFlowGraphResult(result.payload);
}

export async function queryNetworkFlowContributors(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly graph: NetworkFlowGraphResult;
  readonly selector: { readonly kind: "edge"; readonly edge_id: string };
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowContributorResult> {
  const request: NetworkFlowContributorQueryRequest = {
    schema_id: "cartulary.network_flow.graph_contributor_query_request.v1",
    graph_query: options.graph.semantic_query,
    graph_query_digest: options.graph.graph_query_digest,
    selector: options.selector,
    limit: 50,
  };
  const result = await fetchWorkbookJSON<unknown>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/graphs/contributors/query`,
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
  return decodeNetworkFlowContributorResult(result.payload);
}

export async function linkNetworkFlowIndicator(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly graph: NetworkFlowGraphResult;
  readonly edgeId: string;
  readonly fieldKey: "network_flow.src_ip" | "network_flow.dst_ip";
  readonly confirmExactValue: string;
}): Promise<NetworkFlowIndicatorLinkResult> {
  const request: NetworkFlowIndicatorLinkRequest = {
    schema_id: "cartulary.network_flow.indicator_link_request.v1",
    client_txn_id: clientTxnID("nf-indicator-link"),
    selector: {
      kind: "graph_edge",
      graph_query: options.graph.semantic_query,
      graph_query_digest: options.graph.graph_query_digest,
      edge_id: options.edgeId,
      field_key: options.fieldKey,
    },
    target: {
      mode: "create_indicator",
      indicator_type: indicatorTypeForIP(options.confirmExactValue),
    },
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
  return decodeNetworkFlowIndicatorLinkResult(result.payload);
}

function indicatorTypeForIP(value: string): "ipv4_addr" | "ipv6_addr" {
  return value.includes(":") ? "ipv6_addr" : "ipv4_addr";
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
