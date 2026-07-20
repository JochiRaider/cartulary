import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import type {
  NetworkFlowContributor,
  NetworkFlowContributorPageRequest,
  NetworkFlowGraphEdge,
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
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowContributors,
} from "./networkFlowQueryModel";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";

export type NetworkFlowGraphScopeMode =
  | "active_table"
  | "selected_tables"
  | "all_active_tables";

export type NetworkFlowGraphSelection = NetworkFlowGraphSelector;

export function useNetworkFlowGraphController({
  availability,
  activeTableId,
  apiBase,
  enabled,
  incidentId,
  onError,
  onIncidentAccessLost,
  query,
  tables,
}: {
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
  const tableIds = useMemo(
    () => tables.map((table) => table.network_flow_table_id),
    [tables],
  );
  const [scopeMode, setScopeModeState] =
    useState<NetworkFlowGraphScopeMode>("active_table");
  const [selectedTableIds, setSelectedTableIds] = useState<readonly string[]>(
    () => (activeTableId === null ? [] : [activeTableId]),
  );
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
  const graphControllerRef = useRef<AbortController | null>(null);
  const contributorControllerRef = useRef<AbortController | null>(null);
  const contributorGenerationRef = useRef(0);
  const contributorHistoryRef = useRef<NetworkFlowContributorPageRequest[]>([]);
  const contributorPageIndexRef = useRef(0);
  const contributorPagingRef = useRef<NetworkFlowPaging | null>(null);
  const contributorSelectionKeyRef = useRef("");

  useEffect(() => {
    setSelectedTableIds((current) => {
      const retained = tableIds.filter((tableId) => current.includes(tableId));
      const next =
        retained.length > 0
          ? retained
          : activeTableId !== null && tableIds.includes(activeTableId)
            ? [activeTableId]
            : tableIds.slice(0, 1);
      return equalStrings(current, next) ? current : next;
    });
  }, [activeTableId, tableIds]);

  const tableScope = useMemo<NetworkFlowTableScope | null>(() => {
    if (tableIds.length === 0) {
      return null;
    }
    if (scopeMode === "all_active_tables") {
      return { mode: "all_active_tables" };
    }
    if (scopeMode === "selected_tables") {
      const ordered = tableIds.filter((tableId) =>
        selectedTableIds.includes(tableId),
      );
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
  const tableScopeKey = JSON.stringify(tableScope);

  useEffect(() => {
    void graphGeneration;
    void tableScopeKey;
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
    if (!enabled || tableScope === null) {
      setGraph(null);
      setGraphLoadState("idle");
      return;
    }
    const controller = new AbortController();
    graphControllerRef.current = controller;
    setGraph(null);
    setGraphLoadState("loading");
    onError(null);
    void queryNetworkFlowGraph({
      availability,
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
        setGraph(null);
        setGraphLoadState("error");
        onError(requestError);
      });
    return () => controller.abort();
  }, [
    availability,
    apiBase,
    enabled,
    graphGeneration,
    incidentId,
    onError,
    onIncidentAccessLost,
    query,
    tableScope,
    tableScopeKey,
  ]);

  const selectedVertex = useMemo<NetworkFlowGraphVertex | null>(() => {
    if (selection?.kind !== "vertex") {
      return null;
    }
    return (
      graph?.graph_projection_result.vertices.find(
        (vertex) => semanticVertexId(vertex) === selection.vertex_id,
      ) ?? null
    );
  }, [graph, selection]);
  const selectedEdge = useMemo<NetworkFlowGraphEdge | null>(() => {
    if (selection?.kind !== "edge") {
      return null;
    }
    return (
      graph?.graph_projection_result.edges.find(
        (edge) => semanticEdgeId(edge) === selection.edge_id,
      ) ?? null
    );
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
      setContributorLoadState(
        contributorPagingRef.current === null ? "loading" : "refreshing",
      );
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
        if (
          isNetworkFlowProtectedStateLoss(requestError) ||
          requestError.code === "network_flow_graph_query_stale"
        ) {
          setGraph(null);
          setSelectionState(null);
        }
        contributorPagingRef.current = null;
        setContributors([]);
        setContributorPaging(null);
        setContributorLoadState("error");
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
    contributorHistoryRef.current = [];
    contributorPageIndexRef.current = 0;
    contributorPagingRef.current = null;
    if (graph === null || selection === null) {
      return;
    }
    const initialRequest: NetworkFlowContributorPageRequest = {
      schema_id: "cartulary.network_flow.graph_contributor_query_request.v1",
      graph_query: graph.semantic_query,
      graph_query_digest: graph.graph_query_digest,
      selector: selection,
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
  const setScopeMode = useCallback(
    (mode: NetworkFlowGraphScopeMode) => {
      if (mode === "selected_tables" && selectedTableIds.length === 0) {
        const fallback = activeTableId ?? tableIds[0];
        if (fallback !== undefined) {
          setSelectedTableIds([fallback]);
        }
      }
      setScopeModeState(mode);
    },
    [activeTableId, selectedTableIds.length, tableIds],
  );
  const setTableSelected = useCallback(
    (tableId: string, selected: boolean) => {
      if (!tableIds.includes(tableId)) {
        return;
      }
      setSelectedTableIds((current) => {
        const nextSet = new Set(current);
        if (selected) {
          nextSet.add(tableId);
        } else if (nextSet.size > 1) {
          nextSet.delete(tableId);
        }
        return tableIds.filter((candidate) => nextSet.has(candidate));
      });
    },
    [tableIds],
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
  }, []);
  const markGraphStale = useCallback(() => {
    setGraph(null);
    setGraphLoadState("error");
    setSelectionState(null);
    resetContributors({
      setContributors,
      setLoadState: setContributorLoadState,
      setPageIndex: setContributorPageIndex,
      setPaging: setContributorPaging,
    });
  }, []);

  return {
    canNextContributorPage:
      contributorPaging?.next_cursor_token !== null &&
      contributorPaging !== null,
    canPreviousContributorPage: contributorPageIndex > 0,
    clearGraph,
    contributorLoadState,
    contributorPageNumber: contributorPageIndex + 1,
    contributors,
    firstContributor: contributors[0] ?? null,
    graph,
    graphLoadState,
    markGraphStale,
    nextContributorPage,
    previousContributorPage,
    refreshGraph: () => setGraphGeneration((current) => current + 1),
    retryContributorPage,
    scopeMode,
    selectGraphObject,
    selectedEdge,
    selectedTableIds,
    selectedVertex,
    selection,
    setScopeMode,
    setTableSelected,
  };
}

function semanticVertexId(vertex: NetworkFlowGraphVertex): string | null {
  return vertex.source_entity_ref?.source_entity_id ?? null;
}

function semanticEdgeId(edge: NetworkFlowGraphEdge): string | null {
  const propertyEdgeId = edge.properties.edge_id;
  return typeof propertyEdgeId === "string"
    ? propertyEdgeId
    : (edge.source_relationship_ref?.source_relationship_id ?? null);
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
  return left.kind === "vertex"
    ? left.vertex_id === (right as { readonly vertex_id: string }).vertex_id
    : left.edge_id === (right as { readonly edge_id: string }).edge_id;
}

function equalStrings(
  left: readonly string[],
  right: readonly string[],
): boolean {
  return (
    left.length === right.length &&
    left.every((value, index) => value === right[index])
  );
}
