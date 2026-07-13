import { apiPath, clientTxnID } from "../services/browserApi";
import type {
  NetworkFlowContributorResult,
  NetworkFlowDiagnostic,
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowPaging,
  NetworkFlowRow,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";
import { coordinateExtensionImport } from "../shared/importCoordinator";
import type { WorkbookSheetRef } from "../shared/workbookSheetRef";

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

type TableListResponse = {
  schema_id: "cartulary.network_flow.table_list.v1";
  tables: NetworkFlowTable[];
};

type TableQueryResponse = {
  schema_id: "cartulary.network_flow.table_query_result.v1";
  rows: NetworkFlowRow[];
  meta: { paging: NetworkFlowPaging };
};

type RejectedRowsQueryResponse = {
  schema_id: "cartulary.network_flow.rejected_rows_query_result.v1";
  diagnostics: NetworkFlowDiagnostic[];
  meta: { paging: NetworkFlowPaging };
};

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
  const result = await fetchWorkbookJSON<TableListResponse>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables`,
    ),
    requestInit({}, options.signal),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<TableListResponse>(result.payload).tables ?? [];
}

export async function queryNetworkFlowTable(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly rows: NetworkFlowRow[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchWorkbookJSON<TableQueryResponse>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables/${options.tableId}/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify({
          schema_id: "cartulary.network_flow.table_query_request.v1",
          limit: 50,
        }),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const envelope = readEnvelope<TableQueryResponse>(result.payload);
  return { rows: envelope.rows ?? [], paging: envelope.meta.paging };
}

export async function queryNetworkFlowRejectedRows(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<{
  readonly diagnostics: NetworkFlowDiagnostic[];
  readonly paging: NetworkFlowPaging;
}> {
  const result = await fetchWorkbookJSON<RejectedRowsQueryResponse>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/tables/${options.tableId}/rejected-rows/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify({
          schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
          limit: 50,
        }),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const envelope = readEnvelope<RejectedRowsQueryResponse>(result.payload);
  return {
    diagnostics: envelope.diagnostics ?? [],
    paging: envelope.meta.paging,
  };
}

export async function queryNetworkFlowGraph(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableIds: readonly string[];
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowGraphResult> {
  const tableIds = [...new Set(options.tableIds)].sort();
  const result = await fetchWorkbookJSON<NetworkFlowGraphResult>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/graphs/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify({
          schema_id: "cartulary.network_flow.graph_query_request.v1",
          table_scope:
            tableIds.length === 1
              ? { mode: "active_table", active_table_id: tableIds[0] }
              : { mode: "selected_tables", selected_table_ids: tableIds },
          aggregation: {
            mode: "default_flow_edge_v1",
            include_example_row_refs: true,
          },
        }),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<NetworkFlowGraphResult>(result.payload);
}

export async function queryNetworkFlowContributors(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly graph: NetworkFlowGraphResult;
  readonly selector: { readonly kind: "edge"; readonly edge_id: string };
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowContributorResult> {
  const result = await fetchWorkbookJSON<NetworkFlowContributorResult>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/graphs/contributors/query`,
    ),
    requestInit(
      {
        method: "POST",
        body: JSON.stringify({
          schema_id:
            "cartulary.network_flow.graph_contributor_query_request.v1",
          graph_query: options.graph.semantic_query,
          graph_query_digest: options.graph.graph_query_digest,
          selector: options.selector,
          limit: 50,
        }),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<NetworkFlowContributorResult>(result.payload);
}

export async function linkNetworkFlowIndicator(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly graph: NetworkFlowGraphResult;
  readonly edgeId: string;
  readonly fieldKey: "network_flow.src_ip" | "network_flow.dst_ip";
  readonly confirmExactValue: string;
}): Promise<NetworkFlowIndicatorLinkResult> {
  const result = await fetchWorkbookJSON<NetworkFlowIndicatorLinkResult>(
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/indicator-links`,
    ),
    {
      method: "POST",
      body: JSON.stringify({
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
      }),
    },
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<NetworkFlowIndicatorLinkResult>(result.payload);
}

export async function importNetworkFlowCSV(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<void> {
  await coordinateExtensionImport({
    ...options,
    mappingPayload: networkFlowMappingPayload,
    transactionPrefix: "nf-import",
  });
}

function indicatorTypeForIP(value: string): "ipv4_addr" | "ipv6_addr" {
  return value.includes(":") ? "ipv6_addr" : "ipv4_addr";
}

function requestInit(
  init: RequestInit,
  signal: AbortSignal | undefined,
): RequestInit {
  return signal === undefined ? init : { ...init, signal };
}

function networkFlowMappingPayload(clientTxnId: string) {
  const headers = [
    "Source IP Address",
    "Destination IP Address",
    "Source Port",
    "Destination Port",
    "Protocol",
    "Bytes",
    "Packets",
    "Flow Start Time",
    "Flow End Time",
    "Input Interface",
    "Output Interface",
  ];
  const fieldKeys = [
    "network_flow.src_ip",
    "network_flow.dst_ip",
    "network_flow.src_port",
    "network_flow.dst_port",
    "network_flow.ip_protocol",
    "network_flow.bytes_count",
    "network_flow.packets_count",
    "network_flow.flow_start_utc",
    "network_flow.flow_end_utc",
    "network_flow.input_interface",
    "network_flow.output_interface",
  ];
  const transforms: Record<string, string> = {
    "network_flow.src_ip": "ip_literal_v1",
    "network_flow.dst_ip": "ip_literal_v1",
    "network_flow.src_port": "port_number_v1",
    "network_flow.dst_port": "port_number_v1",
    "network_flow.ip_protocol": "protocol_number_or_token_v1",
    "network_flow.bytes_count": "uint64_decimal_string_v1",
    "network_flow.packets_count": "uint64_decimal_string_v1",
    "network_flow.flow_start_utc": "timestamp_profile_v1",
    "network_flow.flow_end_utc": "timestamp_profile_v1",
    "network_flow.input_interface": "trim_ascii_space_v1",
    "network_flow.output_interface": "trim_ascii_space_v1",
  };
  return {
    client_txn_id: clientTxnId,
    target_kind: "network_flow_table",
    extension_profile_id: networkFlowActivityProfileId,
    owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
    owner_mapping: {
      schema_id: "cartulary.network_flow.mapping_candidate.v1",
      target_kind: "network_flow_table",
      target_table_schema_id: "cartulary.network_flow_table.v1",
      source_profile_id: "cisco_sna_netflow_csv_v1",
      parser_profile_id: "rfc4180_headered_csv_v1",
      unknown_column_policy: "preserve_unmapped_raw",
      timestamp_profile: {
        schema_id: "cartulary.network_flow.timestamp_profile.v1",
        mode: "rfc3339",
        precision: "seconds",
        timezone: null,
        timezone_ruleset_id: null,
        ambiguous_local_time_policy: "reject",
        local_time_gap_policy: "reject",
      },
      field_mappings: fieldKeys.map((fieldKey, index) => ({
        mapping_kind: "source_column",
        field_key: fieldKey,
        source_column_ordinal: index + 1,
        transform_id: transforms[fieldKey],
        empty_value_policy:
          fieldKey === "network_flow.input_interface" ||
          fieldKey === "network_flow.output_interface"
            ? "empty_string_is_null"
            : "empty_string_is_invalid",
        combinability: "single_source_only",
      })),
    },
    header_row_ref: 1,
    data_start_row_ref: 2,
    source_columns: headers.map((header, index) => ({
      source_column_ordinal: index + 1,
      source_header_text: header,
      field_key: null,
      entity_binding_mode: null,
      transform_id: null,
      transform_options: {},
      empty_value_policy: "omit_field",
    })),
  };
}
