import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
import type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowGraphEdge,
  NetworkFlowGraphQueryRequest,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowGraphVertex,
  NetworkFlowPaging,
  NetworkFlowTable,
  NetworkFlowTableScope,
} from "./networkFlowClient";
import {
  queryNetworkFlowContributors,
  queryNetworkFlowGraph,
} from "./networkFlowClient";
import {
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowProtectedStateLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import {
  type GraphQuerySettings,
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowContributors,
  validateGraphDraft,
} from "./networkFlowQueryModel";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";

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
  onQueryResult,
}: {
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
  const [contributors, setContributors] = useState<
    readonly NetworkFlowContributor[]
  >([]);
  const [contributorPaging, setContributorPaging] =
    useState<NetworkFlowPaging | null>(null);
  const [contributorPageIndex, setContributorPageIndex] = useState(0);
  const [contributorLoadState, setContributorLoadState] =
    useState<NetworkFlowQueryLoadState>("idle");
  const [contributorError, setContributorError] =
    useState<NetworkFlowRequestError | null>(null);
  const [contributorLoadGenerationKey, setContributorLoadGenerationKey] =
    useState(0);
  const graphControllerRef = useRef<AbortController | null>(null);
  const contributorControllerRef = useRef<AbortController | null>(null);
  const contributorGenerationRef = useRef(0);
  const contributorHistoryRef = useRef<NetworkFlowContributorPageRequest[]>([]);
  const contributorPageIndexRef = useRef(0);
  const contributorPagingRef = useRef<NetworkFlowPaging | null>(null);
  const contributorSelectionKeyRef = useRef("");

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

  useEffect(() => {
    void revision;
    void graphGeneration;
    void tableScopeKey;
    void resolvedSourceKey;
    graphControllerRef.current?.abort();
    contributorControllerRef.current?.abort();
    contributorGenerationRef.current += 1;
    setSelectionState(null);
    resetContributors({
      setContributors,
      setLoadState: setContributorLoadState,
      setPageIndex: setContributorPageIndex,
      setPaging: setContributorPaging,
    });
    setContributorError(null);
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
    const controller = new AbortController();
    graphControllerRef.current = controller;
    setGraph(null);
    setGraphLoadState("loading");
    onError(null);
    void queryNetworkFlowGraph({
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
      signal: controller.signal,
    })
      .then((nextGraph) => {
        if (controller.signal.aborted) {
          return;
        }
        onQueryResult(revision, null);
        setGraph(nextGraph);
        setGraphLoadState("ready");
        onError(null);
      })
      .catch((caught: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow graph query failed.",
        );
        if (isNetworkFlowAuthorizationLoss(requestError)) {
          onIncidentAccessLost?.();
        }
        onQueryResult(revision, requestError);
        setGraph(null);
        setGraphLoadState("error");
        onError(requestError);
      });
    return () => controller.abort();
  }, [
    availability,
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

  const executeContributorRequest = useCallback(
    async (
      request: NetworkFlowContributorPageRequest,
      selectionKey: string,
    ) => {
      contributorControllerRef.current?.abort();
      const controller = new AbortController();
      contributorControllerRef.current = controller;
      contributorGenerationRef.current += 1;
      const generation = contributorGenerationRef.current;
      setContributorLoadGenerationKey(generation);
      setContributorLoadState(
        contributorPagingRef.current === null ? "loading" : "refreshing",
      );
      setContributorError(null);
      try {
        const result = await queryNetworkFlowContributors({
          availability,
          apiBase,
          incidentId,
          request,
          signal: controller.signal,
        });
        if (
          controller.signal.aborted ||
          generation !== contributorGenerationRef.current ||
          selectionKey !== contributorSelectionKeyRef.current
        ) {
          return;
        }
        const nextPaging = result.meta.paging;
        contributorPagingRef.current = nextPaging;
        setContributors((current) =>
          reconcileNetworkFlowContributors(current, result.contributors),
        );
        setContributorPaging(nextPaging);
        setContributorLoadState("ready");
        setContributorError(null);
        onError(null);
      } catch (caught) {
        if (
          controller.signal.aborted ||
          generation !== contributorGenerationRef.current
        ) {
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow contributor query failed.",
        );
        if (isNetworkFlowAuthorizationLoss(requestError)) {
          onIncidentAccessLost?.();
        }
        const clearProtectedContributors =
          isNetworkFlowProtectedStateLoss(requestError) ||
          requestError.code === "network_flow_graph_query_stale";
        if (clearProtectedContributors) {
          setGraph(null);
          setSelectionState(null);
          contributorPagingRef.current = null;
          setContributors([]);
          setContributorPaging(null);
        }
        setContributorLoadState("error");
        setContributorError(requestError);
        onError(requestError);
      }
    },
    [availability, apiBase, incidentId, onError, onIncidentAccessLost],
  );

  const selectionKey =
    graph === null || selection === null
      ? ""
      : `${graph.graph_query_digest}:${JSON.stringify(selection)}`;
  contributorSelectionKeyRef.current = selectionKey;
  useEffect(() => {
    contributorControllerRef.current?.abort();
    contributorGenerationRef.current += 1;
    resetContributors({
      setContributors,
      setLoadState: setContributorLoadState,
      setPageIndex: setContributorPageIndex,
      setPaging: setContributorPaging,
    });
    setContributorError(null);
    contributorHistoryRef.current = [];
    contributorPageIndexRef.current = 0;
    contributorPagingRef.current = null;
    if (graph === null || selection === null) {
      return;
    }
    const initialRequest: NetworkFlowContributorPageRequest = {
      schema_id: "cartulary.network_flow.graph_contributor_query_request.v2",
      graph_query: graph.semantic_query,
      graph_query_digest: graph.graph_query_digest,
      selector: selection,
      limit: 500,
    };
    contributorHistoryRef.current = [initialRequest];
    void executeContributorRequest(initialRequest, selectionKey);
    return () => contributorControllerRef.current?.abort();
  }, [executeContributorRequest, graph, selection, selectionKey]);

  const selectGraphObject = useCallback(
    (next: NetworkFlowGraphSelection | null) => {
      setSelectionState((current) =>
        graphSelectionEqual(current, next) ? current : next,
      );
    },
    [],
  );
  const nextContributorPage = useCallback(() => {
    const cursor = contributorPagingRef.current?.next_cursor_token ?? null;
    if (cursor === null || selectionKey === "") {
      return;
    }
    const request: NetworkFlowContributorPageRequest = {
      schema_id:
        "cartulary.network_flow.graph_contributor_query_continuation.v1",
      cursor_token: cursor,
    };
    const nextIndex = contributorPageIndexRef.current + 1;
    contributorHistoryRef.current = [
      ...contributorHistoryRef.current.slice(0, nextIndex),
      request,
    ];
    contributorPageIndexRef.current = nextIndex;
    setContributorPageIndex(nextIndex);
    void executeContributorRequest(request, selectionKey);
  }, [executeContributorRequest, selectionKey]);
  const previousContributorPage = useCallback(() => {
    if (contributorPageIndexRef.current === 0 || selectionKey === "") {
      return;
    }
    const previousIndex = contributorPageIndexRef.current - 1;
    const request = contributorHistoryRef.current[previousIndex];
    if (request === undefined) {
      return;
    }
    contributorPageIndexRef.current = previousIndex;
    setContributorPageIndex(previousIndex);
    void executeContributorRequest(request, selectionKey);
  }, [executeContributorRequest, selectionKey]);
  const retryContributorPage = useCallback(() => {
    if (selectionKey === "") {
      return;
    }
    const request =
      contributorHistoryRef.current[contributorPageIndexRef.current];
    if (request !== undefined) {
      void executeContributorRequest(request, selectionKey);
    }
  }, [executeContributorRequest, selectionKey]);

  const clearGraph = useCallback(() => {
    staleRef.current = false;
    setGraphStale(false);
    graphControllerRef.current?.abort();
    contributorControllerRef.current?.abort();
    contributorGenerationRef.current += 1;
    setGraph(null);
    setGraphLoadState("idle");
    setSelectionState(null);
    resetContributors({
      setContributors,
      setLoadState: setContributorLoadState,
      setPageIndex: setContributorPageIndex,
      setPaging: setContributorPaging,
    });
    setContributorError(null);
  }, []);
  const markGraphStale = useCallback(() => {
    graphControllerRef.current?.abort();
    contributorControllerRef.current?.abort();
    contributorGenerationRef.current++;
    staleRef.current = true;
    setGraphStale(true);
    setGraph(null);
    setGraphLoadState("error");
    setSelectionState(null);
    resetContributors({
      setContributors,
      setLoadState: setContributorLoadState,
      setPageIndex: setContributorPageIndex,
      setPaging: setContributorPaging,
    });
    setContributorError(null);
  }, []);

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

  return {
    graphStale,
    aggregationMode,
    bucketWidthSeconds,
    canNextContributorPage:
      contributorPaging?.next_cursor_token !== null &&
      contributorPaging !== null,
    canPreviousContributorPage: contributorPageIndex > 0,
    clearGraph,
    contributorError,
    contributorLoadState,
    contributorLoadGenerationKey,
    contributorPageNumber: contributorPageIndex + 1,
    contributors,
    firstContributor: contributors[0] ?? null,
    graph,
    graphLoadState,
    markGraphStale,
    nextContributorPage,
    previousContributorPage,
    refreshGraph: () => {
      staleRef.current = false;
      setGraphStale(false);
      setGraphGeneration((current) => current + 1);
    },
    retryContributorPage,
    scopeMode,
    selectGraphObject,
    selectedEdge,
    selectedTableIds,
    selectedVertex,
    selection,
    validationMessage,
  };
}

function resetContributors(options: {
  readonly setContributors: (value: readonly NetworkFlowContributor[]) => void;
  readonly setLoadState: (value: NetworkFlowQueryLoadState) => void;
  readonly setPageIndex: (value: number) => void;
  readonly setPaging: (value: NetworkFlowPaging | null) => void;
}) {
  options.setContributors([]);
  options.setPaging(null);
  options.setPageIndex(0);
  options.setLoadState("idle");
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
  return JSON.stringify(left) === JSON.stringify(right);
}
