import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import {
  type ExtensionAvailabilityController,
  ExtensionAvailabilityUnavailableError,
} from "../extensions/extensionAvailability";
import { boundedRead } from "../services/asyncObservation";
import {
  networkFlowContractEqual,
  validateNetworkFlowPageContinuation,
} from "../services/networkFlowContractAdapter";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
import type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowGraphEdge,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
  NetworkFlowTable,
  NetworkFlowTableScope,
} from "./networkFlowClient";
import {
  queryNetworkFlowContributors,
  queryNetworkFlowGraph,
} from "./networkFlowClient";
import {
  isNetworkFlowAuthorizationLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import {
  type GraphQuerySettings,
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowContributors,
  validateGraphDraft,
} from "./networkFlowQueryModel";
import {
  type NetworkFlowQueryLoadState,
  useNetworkFlowPagedQuery,
} from "./useNetworkFlowPagedQuery";

export type NetworkFlowGraphScopeMode =
  | "active_table"
  | "selected_tables"
  | "all_active_tables";

export type NetworkFlowGraphSelection = NetworkFlowGraphSelector;
export type NetworkFlowGraphAggregationMode =
  NetworkFlowGraphQueryRequest["aggregation"]["mode"];
export type NetworkFlowGraphBucketWidth = Extract<
  NetworkFlowGraphQueryRequest["aggregation"],
  { readonly mode: "time_bucket_v1" }
>["bucket_width_seconds"];

export function useNetworkFlowGraphController({
  availability,
  tableLifecycle,
  activeTableId,
  apiBase,
  enabled,
  incidentId,
  onError,
  onIncidentAccessLost,
  query,
  tables,
  settings,
  revision,
  applicationRevision,
  readIdentity,
  isCurrentRead,
  onProtectedStateLoss,
  onQueryResult,
}: {
  readonly readIdentity?: string | null;
  readonly isCurrentRead?: () => boolean;
  readonly onProtectedStateLoss?: (error: NetworkFlowRequestError) => void;
  readonly settings: GraphQuerySettings;
  readonly applicationRevision: number;
  readonly revision: number;
  readonly onQueryResult: (
    revision: number,
    error: NetworkFlowRequestError | null,
  ) => void;
  readonly tableLifecycle: NetworkFlowTableController;
  readonly availability: ExtensionAvailabilityController;
  readonly activeTableId: string | null;
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly query: NetworkFlowAcceptedQuery;
  readonly tables: readonly NetworkFlowTable[];
}) {
  const tableIdsKey = JSON.stringify(
    tables.map((table) => table.network_flow_table_id),
  );
  const tableIds = useMemo<readonly string[]>(
    () => JSON.parse(tableIdsKey),
    [tableIdsKey],
  );
  const { scopeMode, selectedTableIds, aggregation } = settings;
  const aggregationMode = aggregation.mode;
  const bucketWidthSeconds =
    aggregation.mode === "time_bucket_v1"
      ? aggregation.bucket_width_seconds
      : 3600;
  const staleRef = useRef(false);
  const [graphStale, setGraphStale] = useState(false);
  const [graph, setGraph] = useState<NetworkFlowGraphResult | null>(null);
  const [graphLoadState, setGraphLoadState] =
    useState<NetworkFlowQueryLoadState>("idle");
  const [selection, setSelectionState] =
    useState<NetworkFlowGraphSelection | null>(null);
  const [graphGeneration, setGraphGeneration] = useState(0);
  const graphControllerRef = useRef<AbortController | null>(null);
  const clearContributors = useRef<() => void>(() => {});

  const clearGraph = useCallback(() => {
    staleRef.current = false;
    setGraphStale(false);
    graphControllerRef.current?.abort();
    clearContributors.current();
    setGraph(null);
    setGraphLoadState("idle");
    setSelectionState(null);
  }, []);
  const markGraphStale = useCallback(() => {
    graphControllerRef.current?.abort();
    clearContributors.current();
    staleRef.current = true;
    setGraphStale(true);
    setGraph(null);
    setGraphLoadState("error");
    setSelectionState(null);
  }, []);
  const derivedTableScope = useMemo<NetworkFlowTableScope | null>(() => {
    if (tableIds.length === 0) {
      return null;
    }
    if (scopeMode === "all_active_tables") {
      return { mode: "all_active_tables" };
    }
    if (scopeMode === "selected_tables") {
      const ordered = selectedTableIds;
      const [first, ...remaining] = ordered;
      return first === undefined
        ? null
        : {
            mode: "selected_tables",
            selected_table_ids: [first, ...remaining],
          };
    }
    return activeTableId === null
      ? null
      : { mode: "active_table", active_table_id: activeTableId };
  }, [activeTableId, scopeMode, selectedTableIds, tableIds]);
  const tableScopeKey = JSON.stringify(derivedTableScope);
  const stableScope = useRef({ key: tableScopeKey, value: derivedTableScope });
  if (stableScope.current.key !== tableScopeKey)
    stableScope.current = { key: tableScopeKey, value: derivedTableScope };
  const tableScope = stableScope.current.value;
  const resolvedSourceIds =
    scopeMode === "all_active_tables"
      ? tableIds
      : scopeMode === "selected_tables"
        ? selectedTableIds
        : activeTableId
          ? [activeTableId]
          : [];
  const resolvedSourceKey = JSON.stringify(resolvedSourceIds);
  const validationMessage =
    validateGraphDraft(settings, query, activeTableId, tables)[0]?.message ??
    null;
  useLayoutEffect(() => {
    void applicationRevision;
    staleRef.current = false;
    setGraphStale(false);
  }, [applicationRevision]);

  const sourceState = useRef({
    ids: resolvedSourceIds,
    displayed: graph !== null || graphLoadState === "loading",
  });
  sourceState.current = {
    ids: resolvedSourceIds,
    displayed: graph !== null || graphLoadState === "loading",
  };
  useLayoutEffect(
    () =>
      tableLifecycle.subscribeChanges((change) => {
        if (
          change.changeKind === "remove" &&
          change.resourceKind === "network_flow_table" &&
          sourceState.current.displayed &&
          sourceState.current.ids.includes(change.resourceId)
        )
          markGraphStale();
      }),
    [tableLifecycle, markGraphStale],
  );
  const previousMembership = useRef(tableIdsKey);
  useLayoutEffect(() => {
    if (
      previousMembership.current !== tableIdsKey &&
      scopeMode === "all_active_tables" &&
      sourceState.current.displayed
    )
      markGraphStale();
    previousMembership.current = tableIdsKey;
  }, [tableIdsKey, scopeMode, markGraphStale]);

  const graphContextKey = JSON.stringify([
    apiBase,
    incidentId,
    readIdentity,
    revision,
    graphGeneration,
    tableScopeKey,
    resolvedSourceKey,
    query,
    aggregation,
    enabled,
  ]);
  const graphContext = useRef(graphContextKey);
  graphContext.current = graphContextKey;
  useLayoutEffect(() => {
    void revision;
    void graphGeneration;
    void tableScopeKey;
    void resolvedSourceKey;
    graphControllerRef.current?.abort();
    clearContributors.current();
    setSelectionState(null);
    if (!enabled) {
      staleRef.current = false;
      setGraphStale(false);
    }
    if (
      !enabled ||
      graphStale ||
      staleRef.current ||
      tableScope === null ||
      validationMessage !== null
    ) {
      setGraph(null);
      setGraphLoadState(graphStale && enabled ? "error" : "idle");
      return;
    }
    const accepts = () =>
      graphContext.current === graphContextKey && isCurrentRead?.() !== false;
    const controller = new AbortController();
    graphControllerRef.current = controller;
    const settleStoppedRead = () => {
      if (
        !controller.signal.aborted &&
        graphContext.current === graphContextKey
      ) {
        setGraph(null);
        setGraphLoadState("idle");
      }
    };
    setGraph(null);
    setGraphLoadState("loading");
    onError(null);
    void boundedRead(
      (signal) =>
        queryNetworkFlowGraph({
          availability,
          aggregation:
            aggregationMode === "time_bucket_v1"
              ? {
                  mode: "time_bucket_v1",
                  bucket_width_seconds: bucketWidthSeconds,
                  include_example_row_refs: true,
                }
              : {
                  mode: "default_flow_edge_v1",
                  include_example_row_refs: true,
                },
          apiBase,
          filters: [...query.filters],
          incidentId,
          tableScope,
          timeRange:
            query.timeWindow === null
              ? null
              : {
                  start_utc: query.timeWindow.startUTC,
                  end_utc: query.timeWindow.endUTC,
                },
          signal,
          authorizeDispatch: () => {
            if (!accepts()) throw new ExtensionAvailabilityUnavailableError();
          },
        }),
      controller.signal,
    )
      .then((nextGraph) => {
        if (controller.signal.aborted || !accepts()) {
          settleStoppedRead();
          return;
        }
        setGraph(nextGraph);
        setGraphLoadState("ready");
        onQueryResult(revision, null);
        if (accepts()) onError(null);
      })
      .catch((caught: unknown) => {
        if (controller.signal.aborted || !accepts()) {
          settleStoppedRead();
          return;
        }
        if (caught instanceof ExtensionAvailabilityUnavailableError) {
          setGraphLoadState("idle");
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow graph query failed.",
        );
        if (isGraphSourceLoss(requestError)) {
          markGraphStale();
          if (accepts()) onQueryResult(revision, requestError);
          void tableLifecycle.loadTables();
          return;
        }
        if (isNetworkFlowAuthorizationLoss(requestError)) {
          onProtectedStateLoss?.(requestError);
          if (accepts()) onIncidentAccessLost?.();
        }
        if (!accepts()) return;
        setGraph(null);
        setGraphLoadState("error");
        onQueryResult(revision, requestError);
        if (accepts()) onError(requestError);
      });
    return () => controller.abort();
  }, [
    availability,
    graphContextKey,
    isCurrentRead,
    onProtectedStateLoss,
    markGraphStale,
    tableLifecycle,
    revision,
    onQueryResult,
    aggregationMode,
    apiBase,
    enabled,
    graphGeneration,
    graphStale,
    resolvedSourceKey,
    incidentId,
    onError,
    onIncidentAccessLost,
    query,
    bucketWidthSeconds,
    tableScope,
    tableScopeKey,
    validationMessage,
  ]);

  const selectedVertex = useMemo<NetworkFlowGraphVertex | null>(() => {
    if (selection?.kind !== "vertex") {
      return null;
    }
    const binding = graph?.vertex_selectors.find((candidate) =>
      graphSelectionEqual(candidate.selector, selection),
    );
    return binding === undefined
      ? null
      : (graph?.graph_projection_result.vertices.find(
          (vertex) => vertex.vertex_id === binding.projected_vertex_id,
        ) ?? null);
  }, [graph, selection]);
  const selectedEdge = useMemo<NetworkFlowGraphEdge | null>(() => {
    if (
      selection?.kind !== "default_edge" &&
      selection?.kind !== "time_bucket_edge"
    ) {
      return null;
    }
    const annotation = graph?.edge_annotations.find((candidate) =>
      graphSelectionEqual(candidate.selector, selection),
    );
    return annotation === undefined
      ? null
      : (graph?.graph_projection_result.edges.find(
          (edge) => edge.edge_id === annotation.projected_edge_id,
        ) ?? null);
  }, [graph, selection]);

  const selectionKey =
    graph === null || selection === null
      ? ""
      : JSON.stringify([
          apiBase,
          incidentId,
          revision,
          graph.graph_query_digest,
          graph.semantic_query,
          selection,
        ]);
  const contributorContext = useRef(selectionKey);
  contributorContext.current = selectionKey;
  const contributorInitial = useMemo<Extract<
    NetworkFlowContributorPageRequest,
    { schema_id: "cartulary.network_flow.graph_contributor_query_request.v2" }
  > | null>(
    () =>
      graph && selection
        ? {
            schema_id:
              "cartulary.network_flow.graph_contributor_query_request.v2",
            graph_query: graph.semantic_query,
            graph_query_digest: graph.graph_query_digest,
            selector: selection,
            limit: 500,
          }
        : null,
    [graph, selection],
  );
  const contributorPage = useNetworkFlowPagedQuery<
    NetworkFlowContributor,
    NetworkFlowContributorPageRequest
  >({
    enabled: enabled && !graphStale && contributorInitial !== null,
    // The disabled hook never dispatches this placeholder.
    initialRequest: contributorInitial ?? {
      schema_id:
        "cartulary.network_flow.graph_contributor_query_continuation.v1",
      cursor_token: "",
    },
    readIdentity,
    queryKey: selectionKey,
    isCurrent: () =>
      contributorContext.current === selectionKey &&
      graphContext.current === graphContextKey &&
      isCurrentRead?.() !== false &&
      !staleRef.current,
    fetchPage: async (request, signal) => {
      if (contributorInitial === null)
        throw new ExtensionAvailabilityUnavailableError();
      const result = await queryNetworkFlowContributors({
        availability,
        apiBase,
        incidentId,
        request,
        signal,
        context: contributorInitial,
        authorizeDispatch: () => {
          if (
            contributorContext.current !== selectionKey ||
            graphContext.current !== graphContextKey ||
            isCurrentRead?.() === false ||
            staleRef.current
          )
            throw new ExtensionAvailabilityUnavailableError();
        },
      });
      return { items: result.contributors, paging: result.meta.paging };
    },
    isContinuation: (request) => "cursor_token" in request,
    makeContinuation: (cursor_token) => ({
      schema_id:
        "cartulary.network_flow.graph_contributor_query_continuation.v1",
      cursor_token,
    }),
    reconcile: reconcileNetworkFlowContributors,
    onError,
    onIncidentAccessLost,
    onProtectedStateLoss: (error) => {
      if (isGraphSourceLoss(error)) {
        markGraphStale();
        void tableLifecycle.loadTables();
      } else {
        clearGraph();
        onProtectedStateLoss?.(error);
      }
    },
    onGraphStale: markGraphStale,
    validatePage: (page, previous, request) => {
      if ("cursor_token" in request && previous)
        validateNetworkFlowPageContinuation(
          null,
          null,
          page.paging.limit,
          previous.paging.limit,
        );
    },
  });
  clearContributors.current = contributorPage.clear;
  const selectGraphObject = useCallback(
    (next: NetworkFlowGraphSelection | null) => {
      if (!graphSelectionEqual(selection, next)) {
        contributorContext.current = "";
        clearContributors.current();
        setSelectionState(next);
      }
    },
    [selection],
  );

  return {
    graphStale,
    aggregationMode,
    bucketWidthSeconds,
    contributorPage,
    canNextContributorPage: contributorPage.canNext,
    canPreviousContributorPage: contributorPage.canPrevious,
    clearGraph,
    contributorError: contributorPage.error,
    contributorLoadState: contributorPage.loadState,
    contributorLoadGenerationKey: contributorPage.loadGenerationKey,
    contributorPageNumber: contributorPage.pageNumber,
    contributors: contributorPage.items,
    firstContributor: contributorPage.items[0] ?? null,
    graph,
    graphLoadState,
    markGraphStale,
    nextContributorPage: contributorPage.nextPage,
    previousContributorPage: contributorPage.previousPage,
    refreshGraph: () => {
      staleRef.current = false;
      setGraphStale(false);
      setGraphGeneration((current) => current + 1);
    },
    retryContributorPage: contributorPage.retry,
    scopeMode,
    selectGraphObject,
    selectedEdge,
    selectedTableIds,
    selectedVertex,
    selection,
    validationMessage,
  };
}

function isGraphSourceLoss(error: NetworkFlowRequestError): boolean {
  return (
    error.code === "network_flow_table_not_active" ||
    error.code === "network_flow_table_not_found" ||
    error.code === "network_flow_graph_query_stale" ||
    (error.code === "network_flow_cursor_invalid" &&
      error.reasonCode === "scope_stale")
  );
}

function graphSelectionEqual(
  left: NetworkFlowGraphSelection | null,
  right: NetworkFlowGraphSelection | null,
): boolean {
  if (left === right) {
    return true;
  }
  if (left === null || right === null || left.kind !== right.kind) {
    return false;
  }
  return networkFlowContractEqual(left, right);
}
