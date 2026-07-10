import {
  apiPath,
  clientTxnID,
  csrfCookieName,
  csrfHeaderName,
  extractError,
  readCookie,
} from "../services/browserApi";
import {
  fetchJSON,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";

export const networkFlowActivityProfileId = "network_flow_activity";
export const networkAnalysisWorkspaceKey = "network_analysis";

export type NetworkFlowTable = {
  readonly network_flow_table_id: string;
  readonly display_name: string;
  readonly table_version: number;
  readonly table_status: "active" | "soft_deleted";
  readonly source_filename_display: string;
  readonly mapping_fingerprint: string;
  readonly row_count_accepted: number;
  readonly row_count_rejected: number;
  readonly diagnostics_truncated: boolean;
  readonly created_at: string;
  readonly updated_at: string;
};

export type NetworkFlowRowRef = {
  readonly network_flow_table_id: string;
  readonly network_flow_row_id: string;
  readonly source_row_number: number;
  readonly mapping_fingerprint: string;
};

export type NetworkFlowRow = {
  readonly network_flow_row_id: string;
  readonly network_flow_table_id: string;
  readonly source_row_number: number;
  readonly mapping_fingerprint: string;
  readonly "network_flow.flow_start_utc": string;
  readonly "network_flow.flow_end_utc": string;
  readonly "network_flow.src_ip": string;
  readonly "network_flow.dst_ip": string;
  readonly "network_flow.src_port": number | null;
  readonly "network_flow.dst_port": number | null;
  readonly "network_flow.ip_protocol": number;
  readonly "network_flow.bytes_count": string;
  readonly "network_flow.packets_count": string;
  readonly "network_flow.exporter_id": string | null;
  readonly "network_flow.input_interface": string | null;
  readonly "network_flow.output_interface": string | null;
  readonly "network_flow.application_label": string | null;
};

export type NetworkFlowDiagnostic = {
  readonly diagnostic_id: string;
  readonly source_row_number: number;
  readonly source_column_ordinal: number | null;
  readonly field_key: string | null;
  readonly error_code: string;
  readonly reason_code: string;
  readonly safe_sample: string | null;
  readonly message: string;
};

export type NetworkFlowPaging = {
  readonly limit: number;
  readonly returned_count: number;
  readonly next_cursor_token: string | null;
};

export type NetworkFlowGraphSemanticQuery = {
  readonly schema_id: "cartulary.network_flow.graph_semantic_query.v1";
  readonly selected_table_ids: string[];
  readonly filters: unknown[];
  readonly time_range: {
    readonly start_utc: string | null;
    readonly end_utc: string | null;
  };
  readonly aggregation: {
    readonly mode: "default_flow_edge_v1";
    readonly include_example_row_refs: boolean;
  };
  readonly result_limits: {
    readonly max_vertices: number;
    readonly max_edges: number;
    readonly max_example_row_refs_per_edge: number;
    readonly max_aggregate_counter_digits: number;
  };
};

export type NetworkFlowEdgeAnnotation = {
  readonly edge_id: string;
  readonly example_row_refs: NetworkFlowRowRef[];
  readonly example_refs_truncated: boolean;
  readonly example_refs_total_count: number;
};

export type NetworkFlowGraphResult = {
  readonly schema_id: "cartulary.network_flow.graph_query_result.v1";
  readonly graph_query_digest: string;
  readonly semantic_query: NetworkFlowGraphSemanticQuery;
  readonly graph_projection_result: {
    readonly ephemeral_projection_id: string;
    readonly state: "ephemeral_available";
  };
  readonly edge_annotations: NetworkFlowEdgeAnnotation[];
  readonly source_table_refs: Array<{
    readonly network_flow_table_id: string;
    readonly table_version: number;
    readonly mapping_fingerprint: string;
    readonly row_count_accepted: number;
    readonly row_count_rejected: number;
  }>;
};

export type NetworkFlowContributor = {
  readonly row_ref: NetworkFlowRowRef;
  readonly row: NetworkFlowRow;
};

export type NetworkFlowContributorResult = {
  readonly schema_id: "cartulary.network_flow.graph_contributor_query_result.v1";
  readonly graph_query_digest: string;
  readonly selector:
    | { readonly kind: "edge"; readonly edge_id: string }
    | { readonly kind: "vertex"; readonly vertex_id: string };
  readonly contributors: NetworkFlowContributor[];
  readonly meta: { readonly paging: NetworkFlowPaging };
};

export type NetworkFlowIndicatorLinkResult = {
  readonly schema_id: "cartulary.network_flow.indicator_link_result.v1";
  readonly duplicate: boolean;
  readonly binding: {
    readonly network_flow_indicator_binding_id: string;
    readonly target_indicator_ref: {
      readonly indicator_id: string;
      readonly indicator_type: "ipv4_addr" | "ipv6_addr";
      readonly normalized_value: string;
    };
    readonly source_row_refs_total_count: number;
  };
};

type DataEnvelope<T> = {
  data: T;
};

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

type JobResource = {
  job_id: string;
  status:
    | "queued"
    | "running"
    | "cancel_requested"
    | "succeeded"
    | "failed"
    | "canceled";
  result_summary?: {
    code?: string;
    resource_refs?: unknown[];
  } | null;
};

type ImportUnit = {
  import_unit_id: string;
};

type ImportSessionUnitsEnvelope = DataEnvelope<{
  import_units: ImportUnit[];
}>;

export function networkAnalysisSheetRef() {
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
    params.get("workspace_key") === networkAnalysisWorkspaceKey
  );
}

export function writeNetworkAnalysisURL(incidentId: string): void {
  const next = new URLSearchParams(window.location.search);
  next.set("incident_id", incidentId);
  next.set("sheet_ref_kind", "extension_workspace");
  next.set("extension_profile_id", networkFlowActivityProfileId);
  next.set("workspace_key", networkAnalysisWorkspaceKey);
  next.delete("view_schema_id");
  next.delete("sheet_ref_id");
  next.delete("surface");
  window.history.replaceState({}, "", `/?${next.toString()}`);
}

export async function listNetworkFlowTables(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowTable[]> {
  const result = await fetchJSON<TableListResponse>(
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
  const result = await fetchJSON<TableQueryResponse>(
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
  const result = await fetchJSON<RejectedRowsQueryResponse>(
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
  const result = await fetchJSON<NetworkFlowGraphResult>(
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
  const result = await fetchJSON<NetworkFlowContributorResult>(
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
  const result = await fetchJSON<NetworkFlowIndicatorLinkResult>(
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
  options.onProgress?.("Uploading import.");
  const uploadJob = await uploadImportSession(options);
  const discovered = await pollJob(options.apiBase, uploadJob.job_id);
  const sessionId = importSessionIdFromJob(discovered);
  if (sessionId === null) {
    throw new Error("import_session_not_returned");
  }

  options.onProgress?.("Preparing mapping.");
  const unitId = await firstImportUnitID(options.apiBase, sessionId);
  await putNetworkFlowMapping(options.apiBase, sessionId, unitId);
  await postImportJSON(
    options.apiBase,
    `/api/v1/import-sessions/${sessionId}/units/${unitId}/select`,
    {
      client_txn_id: clientTxnID("nf-import-select"),
    },
  );

  options.onProgress?.("Applying import.");
  const applyEnvelope = await postImportJSON<{ data: JobResource }>(
    options.apiBase,
    `/api/v1/import-sessions/${sessionId}/apply`,
    { client_txn_id: clientTxnID("nf-import-apply") },
  );
  await pollJob(options.apiBase, applyEnvelope.data.job_id);
}

async function uploadImportSession(options: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
}): Promise<JobResource> {
  const form = new FormData();
  form.append(
    "metadata",
    new Blob(
      [
        JSON.stringify({
          incident_id: options.incidentId,
          client_txn_id: clientTxnID("nf-import-upload"),
        }),
      ],
      { type: "application/json" },
    ),
  );
  form.append("file", options.file, options.file.name);
  const result = await fetchUploadJSON<{ data: JobResource }>(
    options.apiBase,
    "/api/v1/import-sessions",
    form,
  );
  return result.data;
}

async function firstImportUnitID(
  apiBase: string | undefined,
  sessionId: string,
): Promise<string> {
  const result = await fetchJSON<ImportSessionUnitsEnvelope>(
    apiPath(apiBase, `/api/v1/import-sessions/${sessionId}/units`),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const units =
    readEnvelope<ImportSessionUnitsEnvelope>(result.payload).data
      .import_units ?? [];
  const unitId = units[0]?.import_unit_id ?? "";
  if (unitId === "") {
    throw new Error("import_unit_not_returned");
  }
  return unitId;
}

async function putNetworkFlowMapping(
  apiBase: string | undefined,
  sessionId: string,
  unitId: string,
): Promise<void> {
  await postImportJSON(
    apiBase,
    `/api/v1/import-sessions/${sessionId}/units/${unitId}/mapping`,
    networkFlowMappingPayload(clientTxnID("nf-import-mapping")),
    "PUT",
  );
}

async function pollJob(
  apiBase: string | undefined,
  jobId: string,
): Promise<JobResource> {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const result = await fetchJSON<{ data: JobResource }>(
      apiPath(apiBase, `/api/v1/jobs/${jobId}`),
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
    const job = readEnvelope<{ data: JobResource }>(result.payload).data;
    if (job.status === "succeeded") {
      return job;
    }
    if (job.status === "failed" || job.status === "canceled") {
      throw new Error(`job_${job.status}`);
    }
    await new Promise((resolve) => window.setTimeout(resolve, 750));
  }
  throw new Error("job_poll_timeout");
}

async function postImportJSON<T extends object = Record<string, unknown>>(
  apiBase: string | undefined,
  path: string,
  body: Record<string, unknown>,
  method = "POST",
): Promise<T> {
  const result = await fetchJSON<T>(apiPath(apiBase, path), {
    method,
    body: JSON.stringify(body),
  });
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<T>(result.payload);
}

async function fetchUploadJSON<T>(
  apiBase: string | undefined,
  path: string,
  body: FormData,
): Promise<T> {
  const headers = new Headers();
  const csrfToken = readCookie(csrfCookieName);
  if (csrfToken !== null && csrfToken !== "") {
    headers.set(csrfHeaderName, csrfToken);
  }
  const response = await fetch(apiPath(apiBase, path), {
    method: "POST",
    credentials: "include",
    headers,
    body,
  });
  const payload = (await response.json()) as T | { error?: unknown };
  if (!response.ok) {
    const error = extractError(payload);
    throw new Error(error?.code ?? "upload_failed");
  }
  return payload as T;
}

function importSessionIdFromJob(job: JobResource): string | null {
  const refs = job.result_summary?.resource_refs;
  if (!Array.isArray(refs)) {
    return null;
  }
  for (const ref of refs) {
    if (!ref || typeof ref !== "object") {
      continue;
    }
    const candidate = ref as Record<string, unknown>;
    if (
      candidate.kind === "import_session" &&
      typeof candidate.id === "string" &&
      candidate.id.trim() !== ""
    ) {
      return candidate.id;
    }
  }
  return null;
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
