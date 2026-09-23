import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import { apiPath, extractError, fetchJSON } from "../services/browserApi";
import type {
  NetworkFlowContributorPageRequest,
  NetworkFlowContributorResult,
  NetworkFlowDiagnostic,
  NetworkFlowFilter,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowRow,
  NetworkFlowSavedGraph,
  NetworkFlowSavedGraphContributorQueryRequest,
  NetworkFlowSavedGraphContributorResult,
  NetworkFlowSavedGraphResult,
  NetworkFlowTable,
  NetworkFlowTableScope,
} from "../services/networkFlowContractAdapter";
import {
  decodeNetworkFlowContributorResult,
  decodeNetworkFlowGraphResult,
  decodeNetworkFlowIndicatorLinkResult,
  decodeNetworkFlowRejectedRowsQueryResult,
  decodeNetworkFlowSavedGraphAccepted,
  decodeNetworkFlowSavedGraphContributorResult,
  decodeNetworkFlowSavedGraphGet,
  decodeNetworkFlowSavedGraphList,
  decodeNetworkFlowSavedGraphMutationResult,
  decodeNetworkFlowSavedGraphResult,
  decodeNetworkFlowSourceProfileList,
  decodeNetworkFlowTableList,
  decodeNetworkFlowTableMutationResult,
  decodeNetworkFlowTableQueryResult,
  type NetworkFlowDiagnosticPageMetadata,
  type NetworkFlowTablePageMetadata,
  validNetworkFlowErrorEnvelope,
} from "../services/networkFlowContractAdapter";
import { networkFlowRequestError } from "./networkFlowErrors";
import {
  type IndicatorLinkAttempt,
  IndicatorLinkWriteError,
  validateIndicatorLinkReceipt,
  validIndicatorLinkRejection,
} from "./networkFlowIndicatorLinkOperation";
import type {
  NetworkFlowAcceptedPageRequest,
  NetworkFlowRejectedPageRequest,
} from "./networkFlowQueryModel";
import { graphInitialRequest } from "./networkFlowQueryModel";
import {
  type TableAttempt,
  TableWriteError,
  validateTableReceipt,
} from "./networkFlowTableOperation";
import {
  type SavedGraphAttempt,
  type SavedGraphReceipt,
  SavedGraphWriteError,
  savedGraphWriteFailure,
} from "./savedGraphOperation";
import { savedGraphResponseError } from "./savedGraphReadFailure";

export {
  networkAnalysisSheetRef,
  networkAnalysisWorkspaceKey,
  networkFlowActivityProfileId,
} from "../extensions/extensionWorkspaceIdentities";
export type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowDiagnostic,
  NetworkFlowGraphEdge,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
  NetworkFlowIndicatorLinkResult,
  NetworkFlowIndicatorSelector,
  NetworkFlowRow,
  NetworkFlowRowRef,
  NetworkFlowSavedGraph,
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
    throw new SyntaxError("invalid_network_flow_success_envelope");
  }
  return payload.data;
}

function networkFlowReadError(status: number, payload: unknown) {
  if (validNetworkFlowErrorEnvelope(status, payload))
    return networkFlowRequestError(status, payload);
  // Current HTTP authority loss still withdraws exposure when its body is malformed.
  if (status === 401 || status === 403)
    return networkFlowRequestError(status, {
      error: {
        code: status === 401 ? "session_required" : "authorization_denied",
        details: { retry_action: "do_not_retry" },
      },
    });
  return new SyntaxError("invalid_network_flow_error_envelope");
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
  const tables = decodeNetworkFlowTableList(
    networkFlowResponseData(result.payload),
  ).tables;
  if (
    result.status !== 200 ||
    tables.some(
      (table) =>
        table.incident_id !== options.incidentId ||
        table.table_status !== "active" ||
        table.deleted_at !== null,
    ) ||
    new Set(tables.map((table) => table.network_flow_table_id)).size !==
      tables.length
  )
    throw new SyntaxError("invalid_network_flow_table_catalog");
  return tables;
}

/** Mutation identity and payload belong to the table lifecycle owner. */
export async function submitNetworkFlowTableMutation(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly attempt: TableAttempt;
  readonly signal: AbortSignal;
  readonly authorizeDispatch: () => void;
}): Promise<NetworkFlowTable> {
  const { attempt } = options;
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    tableURL({
      apiBase: options.apiBase,
      incidentId: attempt.authority.incidentId,
      tableId: attempt.target.network_flow_table_id,
    }),
    {
      method: attempt.action === "rename" ? "PATCH" : "DELETE",
      body: attempt.body,
      signal: options.signal,
    },
    () => {
      options.signal.throwIfAborted();
      options.authorizeDispatch();
    },
  );
  if (!result.ok) {
    if (!validNetworkFlowErrorEnvelope(result.status, result.payload))
      throw new TableWriteError(
        "uncertain",
        "The server response did not establish whether the request committed. Replay the exact request.",
      );
    const error = networkFlowRequestError(result.status, result.payload);
    throw new TableWriteError(
      result.status >= 400 && result.status < 500 ? "rejected" : "uncertain",
      error.message,
      error,
    );
  }
  try {
    return validateTableReceipt(
      decodeNetworkFlowTableMutationResult(
        networkFlowResponseData(result.payload),
      ).table,
      result.status,
      attempt,
    );
  } catch (error) {
    if (error instanceof TableWriteError) throw error;
    throw new TableWriteError(
      "uncertain",
      "The acknowledgement could not be verified. Replay the exact request to recover.",
    );
  }
}

export async function queryNetworkFlowTable(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowAcceptedPageRequest;
  readonly signal?: AbortSignal | undefined;
  readonly authorizeDispatch?: () => void;
}): Promise<{
  readonly rows: NetworkFlowRow[];
  readonly paging: NetworkFlowTablePageMetadata["paging"];
  readonly metadata: NetworkFlowTablePageMetadata["query"];
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
    () => {
      options.signal?.throwIfAborted();
      options.authorizeDispatch?.();
    },
  );
  if (!result.ok) {
    throw networkFlowReadError(result.status, result.payload);
  }
  const response = decodeNetworkFlowTableQueryResult(
    networkFlowResponseData(result.payload),
    options.tableId,
    options.incidentId,
  );
  return {
    rows: response.rows,
    paging: response.meta.paging,
    metadata: response.meta.query,
  };
}

export async function queryNetworkFlowRejectedRows(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly tableId: string;
  readonly request: NetworkFlowRejectedPageRequest;
  readonly signal?: AbortSignal | undefined;
  readonly authorizeDispatch?: () => void;
}): Promise<{
  readonly diagnostics: NetworkFlowDiagnostic[];
  readonly paging: NetworkFlowDiagnosticPageMetadata["paging"];
  readonly metadata: NetworkFlowDiagnosticPageMetadata["query"];
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
    () => {
      options.signal?.throwIfAborted();
      options.authorizeDispatch?.();
    },
  );
  if (!result.ok) {
    throw networkFlowReadError(result.status, result.payload);
  }
  const response = decodeNetworkFlowRejectedRowsQueryResult(
    networkFlowResponseData(result.payload),
    options.tableId,
  );
  return {
    diagnostics: response.diagnostics,
    paging: response.meta.paging,
    metadata: response.meta.query,
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
  readonly authorizeDispatch?: () => void;
}): Promise<NetworkFlowGraphResult> {
  const request = graphInitialRequest(options);
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
    () => {
      options.signal?.throwIfAborted();
      options.authorizeDispatch?.();
    },
  );
  if (!result.ok) {
    throw networkFlowReadError(result.status, result.payload);
  }
  return decodeNetworkFlowGraphResult(networkFlowResponseData(result.payload));
}

export async function queryNetworkFlowContributors(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly request: NetworkFlowContributorPageRequest;
  readonly context?: Extract<
    NetworkFlowContributorPageRequest,
    { schema_id: "cartulary.network_flow.graph_contributor_query_request.v2" }
  >;
  readonly signal?: AbortSignal | undefined;
  readonly authorizeDispatch?: () => void;
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
    () => {
      options.signal?.throwIfAborted();
      options.authorizeDispatch?.();
    },
  );
  if (!result.ok) {
    throw networkFlowReadError(result.status, result.payload);
  }
  return decodeNetworkFlowContributorResult(
    networkFlowResponseData(result.payload),
    options.context,
    options.incidentId,
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
    throw savedGraphResponseError(result.status, result.payload);
  }
  const graphs = decodeNetworkFlowSavedGraphList(
    networkFlowResponseData(result.payload),
  ).graph_views;
  if (
    result.status !== 200 ||
    graphs.some(
      (graph) =>
        graph.incident_id !== options.incidentId || graph.state !== "active",
    ) ||
    new Set(graphs.map((graph) => graph.graph_view_id)).size !== graphs.length
  )
    throw new SyntaxError("invalid_saved_graph_list_scope");
  return graphs;
}

/** Sends the immutable bytes captured by the saved-graph operation owner. */
export async function submitNetworkFlowSavedGraphMutation(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly attempt: SavedGraphAttempt;
  readonly signal: AbortSignal;
  readonly authorizeDispatch: () => void;
}): Promise<SavedGraphReceipt> {
  const { attempt } = options;
  const { intent } = attempt;
  const target = intent.target;
  const route = `/api/v1/incidents/${intent.authority.incidentId}/network-flow/graph-views${target === null ? "" : `/${target.graph_view_id}`}${intent.kind === "refresh" ? "/refresh" : ""}`;
  let result: Awaited<ReturnType<typeof fetchNetworkFlowJSON<unknown>>>;
  try {
    result = await fetchNetworkFlowJSON<unknown>(
      options.availability,
      apiPath(options.apiBase, route),
      {
        method:
          intent.kind === "rename"
            ? "PATCH"
            : intent.kind === "retire"
              ? "DELETE"
              : "POST",
        body: attempt.body,
        signal: options.signal,
      },
      () => {
        options.signal.throwIfAborted();
        options.authorizeDispatch();
      },
    );
  } catch (caught) {
    throw savedGraphWriteFailure(caught);
  }
  if (!result.ok) {
    if (!validNetworkFlowErrorEnvelope(result.status, result.payload))
      throw new SavedGraphWriteError(
        "invalid_response",
        "uncertain",
        "The rejection envelope is invalid. The original write may have committed; replay this exact attempt to recover.",
      );
    throw savedGraphWriteFailure(
      networkFlowRequestError(result.status, result.payload),
    );
  }
  try {
    if (intent.kind === "retire") {
      if (result.status !== 204 || result.payload !== "")
        throw new Error("nonempty_retirement_receipt");
      return { kind: "retired" };
    }
    if (result.status !== (intent.kind === "rename" ? 200 : 202))
      throw new Error("unexpected_receipt_status");
    const data = networkFlowResponseData(result.payload);
    const receipt: SavedGraphReceipt =
      intent.kind === "rename"
        ? {
            kind: "renamed",
            value: decodeNetworkFlowSavedGraphMutationResult(data).graph_view,
          }
        : {
            kind: "accepted",
            value: decodeNetworkFlowSavedGraphAccepted(data),
          };
    const graph =
      receipt.kind === "renamed" ? receipt.value : receipt.value.graph_view;
    if (
      graph.incident_id !== intent.authority.incidentId ||
      graph.state !== "active" ||
      (target !== null && graph.graph_view_id !== target.graph_view_id) ||
      (attempt.name !== null && graph.display_name !== attempt.name) ||
      (intent.kind === "create" &&
        (graph.created_by !== intent.authority.actorId ||
          graph.graph_view_version !== 1 ||
          graph.materialization_generation !== 1 ||
          !equalJSON(graph.semantic_query, intent.query))) ||
      (target !== null &&
        (graph.created_by !== target.created_by ||
          graph.created_at !== target.created_at ||
          graph.semantic_query_sha256 !== target.semantic_query_sha256 ||
          !equalJSON(graph.semantic_query, target.semantic_query))) ||
      (intent.kind === "rename" &&
        target !== null &&
        (graph.graph_view_version !==
          target.graph_view_version +
            (attempt.name === target.display_name ? 0 : 1) ||
          graph.materialization_generation !==
            target.materialization_generation ||
          graph.latest_job_id !== target.latest_job_id)) ||
      (intent.kind === "refresh" &&
        target !== null &&
        (graph.graph_view_version !== target.graph_view_version + 1 ||
          graph.materialization_generation !==
            target.materialization_generation + 1 ||
          graph.latest_job_id === target.latest_job_id))
    )
      throw new Error("receipt_target_mismatch");
    return receipt;
  } catch {
    throw new SavedGraphWriteError(
      "invalid_response",
      "uncertain",
      "The server returned an invalid acknowledgement. The request may have committed. Replay the exact attempt to recover.",
    );
  }
}

export async function getNetworkFlowSavedGraph(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly graphViewId: string;
  readonly incidentId: string;
  readonly signal?: AbortSignal | undefined;
}): Promise<NetworkFlowSavedGraph> {
  const response = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    graphViewURL(options),
    requestInit({ method: "GET" }, options.signal),
  );
  if (!response.ok)
    throw savedGraphResponseError(response.status, response.payload);
  const graph = decodeNetworkFlowSavedGraphGet(
    networkFlowResponseData(response.payload),
  ).graph_view;
  if (
    response.status !== 200 ||
    graph.incident_id !== options.incidentId ||
    graph.graph_view_id !== options.graphViewId ||
    graph.state !== "active"
  )
    throw new SyntaxError("invalid_saved_graph_target");
  return graph;
}

function equalJSON(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (
    a === null ||
    b === null ||
    typeof a !== "object" ||
    typeof b !== "object" ||
    Array.isArray(a) !== Array.isArray(b)
  )
    return false;
  return (
    Object.keys(a).length === Object.keys(b).length &&
    Object.entries(a).every(
      ([key, value]) =>
        Object.hasOwn(b, key) &&
        equalJSON(value, (b as Record<string, unknown>)[key]),
    )
  );
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
    throw savedGraphResponseError(result.status, result.payload);
  }
  if (result.status !== 200)
    throw new SyntaxError("Unexpected saved-result status.");
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
  readonly cursorToken?: string | undefined;
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
      {
        method: "POST",
        body: JSON.stringify(
          options.cursorToken === undefined
            ? request
            : {
                schema_id:
                  "cartulary.network_flow.graph_contributor_query_continuation.v1",
                cursor_token: options.cursorToken,
              },
        ),
      },
      options.signal,
    ),
  );
  if (!result.ok) {
    throw savedGraphResponseError(result.status, result.payload);
  }
  const page = decodeNetworkFlowSavedGraphContributorResult(
    networkFlowResponseData(result.payload),
  );
  if (
    result.status !== 200 ||
    page.graph_view_id !== options.graphViewId ||
    page.projection_result_id !== options.projectionResultId ||
    !equalJSON(page.selector, options.selector) ||
    page.contributors.length > 100
  )
    throw new SyntaxError("invalid_saved_graph_contributor_target");
  return page;
}

/** Dispatches only the immutable body owned by the captured link attempt. */
export async function linkNetworkFlowIndicator(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly attempt: IndicatorLinkAttempt;
  readonly signal: AbortSignal;
  readonly authorizeDispatch: () => void;
}): Promise<NetworkFlowIndicatorLinkResult> {
  const { attempt } = options;
  const result = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${attempt.authority.incidentId}/network-flow/indicator-links`,
    ),
    { method: "POST", body: attempt.body, signal: options.signal },
    () => {
      options.signal.throwIfAborted();
      options.authorizeDispatch();
    },
  );
  if (!result.ok) {
    const error = networkFlowRequestError(result.status, result.payload);
    const wire = extractError(result.payload);
    const details = wire?.details;
    const provenTimeout =
      result.status === 503 &&
      wire?.code === "service_unavailable" &&
      wire.retryable === true &&
      details?.reason_code === "extension_transaction_timeout" &&
      details.operation_id ===
        `network-flow-indicator-link:${attempt.authority.incidentId}:${attempt.request.client_txn_id}` &&
      typeof details.timeout_seconds === "number" &&
      Number.isFinite(details.timeout_seconds) &&
      details.timeout_seconds > 0;
    const rejected =
      validIndicatorLinkRejection(result.status, result.payload) ||
      (validNetworkFlowErrorEnvelope(result.status, result.payload) &&
        provenTimeout);
    const targetError =
      error.code.includes("indicator_target") ||
      error.code === "network_flow_indicator_link_forbidden" ||
      error.field?.startsWith("target") === true;
    throw new IndicatorLinkWriteError(rejected ? "rejected" : "uncertain", {
      kind:
        result.status === 401 || result.status === 403
          ? "denied"
          : error.retryAction === "refresh_resource"
            ? "stale"
            : rejected
              ? "validation"
              : "recovery",
      field: targetError
        ? "target"
        : error.code === "network_flow_indicator_link_ambiguous" ||
            error.field === "confirm_exact_value"
          ? "confirmation"
          : null,
      message: targetError
        ? "The selected indicator is unavailable or incompatible. Choose a currently visible atomic IP indicator with the same canonical value."
        : rejected
          ? error.message
          : "The server response did not establish whether linking committed. Recover this exact request before starting another link.",
    });
  }
  try {
    return validateIndicatorLinkReceipt(
      decodeNetworkFlowIndicatorLinkResult(
        networkFlowResponseData(result.payload),
      ),
      result.status,
      attempt,
    );
  } catch {
    throw new IndicatorLinkWriteError("uncertain", {
      kind: "read_failed",
      field: null,
      message:
        "The link response could not be verified. The binding may exist. Recover this exact request before starting another link.",
    });
  }
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

export async function getSavedGraphObservationLimit(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly signal: AbortSignal;
}): Promise<number> {
  const response = await fetchNetworkFlowJSON<unknown>(
    options.availability,
    apiPath(
      options.apiBase,
      `/api/v1/incidents/${options.incidentId}/network-flow/source-profiles`,
    ),
    { method: "GET", signal: options.signal },
  );
  if (!response.ok)
    throw networkFlowRequestError(response.status, response.payload);
  return decodeNetworkFlowSourceProfileList(
    networkFlowResponseData(response.payload),
  ).effective_limits["network_flow.max_nonterminal_graph_jobs_per_incident"];
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
  authorizeDispatch?: () => void,
) {
  return availability.runProfileRequest(
    networkFlowActivityProfileId,
    networkFlowRouteFamily,
    () => {
      authorizeDispatch?.();
      return fetchJSON<T>(input, init);
    },
  );
}
