import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import type {
  NetworkFlowContributor,
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
  NetworkFlowSavedGraph,
  NetworkFlowSavedGraphResult,
} from "./networkFlowClient";
import {
  createNetworkFlowSavedGraph,
  getNetworkFlowSavedGraphResult,
  listNetworkFlowSavedGraphs,
  queryNetworkFlowSavedGraphContributors,
  refreshNetworkFlowSavedGraph,
  renameNetworkFlowSavedGraph,
  retireNetworkFlowSavedGraph,
} from "./networkFlowClient";
import {
  isNetworkFlowAuthorizationLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";
import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";

const savedGraphPollMilliseconds = 1_500;

export function useNetworkFlowSavedGraphController(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const listGeneration = useRef(0);
  const resultGeneration = useRef(0);
  const contributorGeneration = useRef(0);
  const graphsRef = useRef<readonly NetworkFlowSavedGraph[]>([]);
  const [graphs, setGraphs] = useState<readonly NetworkFlowSavedGraph[]>([]);
  const [listState, setListState] = useState<NetworkFlowQueryLoadState>("idle");
  const [selectedGraphViewId, setSelectedGraphViewId] = useState<string | null>(
    null,
  );
  const [result, setResult] = useState<NetworkFlowSavedGraphResult | null>(
    null,
  );
  const [resultState, setResultState] =
    useState<NetworkFlowQueryLoadState>("idle");
  const [selection, setSelection] = useState<NetworkFlowGraphSelector | null>(
    null,
  );
  const [contributors, setContributors] = useState<
    readonly NetworkFlowContributor[]
  >([]);
  const [contributorState, setContributorState] =
    useState<NetworkFlowQueryLoadState>("idle");
  const [mutationPending, setMutationPending] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    graphsRef.current = graphs;
  }, [graphs]);

  const reportError = useCallback(
    (caught: unknown, fallback: string) => {
      const error = networkFlowErrorFromUnknown(caught, fallback);
      options.onError(error);
      if (isNetworkFlowAuthorizationLoss(error)) {
        options.onIncidentAccessLost?.();
      }
      return error;
    },
    [options.onError, options.onIncidentAccessLost],
  );

  const loadGraphs = useCallback(async () => {
    if (!options.enabled) {
      return;
    }
    listGeneration.current += 1;
    const activeGeneration = listGeneration.current;
    setListState((current) => (current === "ready" ? "refreshing" : "loading"));
    try {
      const nextGraphs = await listNetworkFlowSavedGraphs({
        availability: options.availability,
        apiBase: options.apiBase,
        incidentId: options.incidentId,
      });
      if (listGeneration.current !== activeGeneration) {
        return;
      }
      const transitionNotice = materializationTransitionNotice(
        graphsRef.current,
        nextGraphs,
      );
      const sortedGraphs = sortSavedGraphs(nextGraphs);
      graphsRef.current = sortedGraphs;
      setGraphs(sortedGraphs);
      if (transitionNotice !== null) setNotice(transitionNotice);
      setSelectedGraphViewId((current) =>
        current !== null &&
        nextGraphs.some((graph) => graph.graph_view_id === current)
          ? current
          : (nextGraphs[0]?.graph_view_id ?? null),
      );
      setListState("ready");
      options.onError(null);
    } catch (caught) {
      if (listGeneration.current !== activeGeneration) {
        return;
      }
      reportError(caught, "Saved graphs could not be loaded.");
      setListState("error");
    }
  }, [
    options.apiBase,
    options.availability,
    options.enabled,
    options.incidentId,
    options.onError,
    reportError,
  ]);

  useEffect(() => {
    listGeneration.current += 1;
    resultGeneration.current += 1;
    contributorGeneration.current += 1;
    setGraphs([]);
    graphsRef.current = [];
    setSelectedGraphViewId(null);
    setResult(null);
    setSelection(null);
    setContributors([]);
    setResultState("idle");
    setContributorState("idle");
    setNotice(null);
    if (!options.enabled) {
      setListState("idle");
      return;
    }
    void loadGraphs();
  }, [options.enabled, loadGraphs]);

  useEffect(() => {
    if (
      !options.enabled ||
      !graphs.some(
        (graph) =>
          graph.last_materialization_status === "queued" ||
          graph.last_materialization_status === "running",
      )
    ) {
      return;
    }
    const timer = window.setTimeout(
      () => void loadGraphs(),
      savedGraphPollMilliseconds,
    );
    return () => window.clearTimeout(timer);
  }, [graphs, loadGraphs, options.enabled]);

  const selectedGraph = useMemo(
    () =>
      graphs.find((graph) => graph.graph_view_id === selectedGraphViewId) ??
      null,
    [graphs, selectedGraphViewId],
  );

  const loadResult = useCallback(async () => {
    const graph = selectedGraph;
    resultGeneration.current += 1;
    const activeGeneration = resultGeneration.current;
    setSelection(null);
    setContributors([]);
    setContributorState("idle");
    if (graph?.selected_result === null || graph === null) {
      setResult(null);
      setResultState(graph === null ? "idle" : "ready");
      return;
    }
    setResultState("loading");
    try {
      const nextResult = await getNetworkFlowSavedGraphResult({
        availability: options.availability,
        apiBase: options.apiBase,
        graphViewId: graph.graph_view_id,
        incidentId: options.incidentId,
      });
      if (resultGeneration.current !== activeGeneration) {
        return;
      }
      setResult(nextResult);
      setResultState("ready");
      options.onError(null);
    } catch (caught) {
      if (resultGeneration.current !== activeGeneration) {
        return;
      }
      reportError(caught, "The saved graph result could not be loaded.");
      setResultState("error");
    }
  }, [
    options.apiBase,
    options.availability,
    options.incidentId,
    options.onError,
    reportError,
    selectedGraph,
  ]);

  useEffect(() => {
    void loadResult();
  }, [loadResult]);

  const selectObject = useCallback(
    async (nextSelection: NetworkFlowGraphSelector | null) => {
      contributorGeneration.current += 1;
      const activeGeneration = contributorGeneration.current;
      setSelection(nextSelection);
      setContributors([]);
      if (
        nextSelection === null ||
        selectedGraph === null ||
        selectedGraph.selected_result === null
      ) {
        setContributorState("idle");
        return;
      }
      setContributorState("loading");
      try {
        const response = await queryNetworkFlowSavedGraphContributors({
          availability: options.availability,
          apiBase: options.apiBase,
          graphViewId: selectedGraph.graph_view_id,
          incidentId: options.incidentId,
          projectionResultId:
            selectedGraph.selected_result.projection_result_id,
          selector: nextSelection,
        });
        if (contributorGeneration.current !== activeGeneration) {
          return;
        }
        setContributors(response.contributors);
        setContributorState("ready");
        options.onError(null);
      } catch (caught) {
        if (contributorGeneration.current !== activeGeneration) {
          return;
        }
        reportError(caught, "Saved graph contributors could not be loaded.");
        setContributorState("error");
      }
    },
    [
      options.apiBase,
      options.availability,
      options.incidentId,
      options.onError,
      reportError,
      selectedGraph,
    ],
  );

  const createGraph = useCallback(
    async (displayName: string, graph: NetworkFlowGraphResult) => {
      setMutationPending(true);
      try {
        const accepted = await createNetworkFlowSavedGraph({
          availability: options.availability,
          apiBase: options.apiBase,
          displayName,
          incidentId: options.incidentId,
          semanticQuery: graph.semantic_query,
        });
        setGraphs((current) =>
          sortSavedGraphs([...current, accepted.graph_view]),
        );
        setSelectedGraphViewId(accepted.graph_view.graph_view_id);
        setNotice(
          `${accepted.graph_view.display_name} is queued for materialization.`,
        );
        options.onError(null);
        return true;
      } catch (caught) {
        reportError(caught, "The saved graph could not be created.");
        return false;
      } finally {
        setMutationPending(false);
      }
    },
    [
      options.apiBase,
      options.availability,
      options.incidentId,
      options.onError,
      reportError,
    ],
  );

  const renameGraph = useCallback(
    async (displayName: string) => {
      if (selectedGraph === null) return false;
      setMutationPending(true);
      try {
        const renamed = await renameNetworkFlowSavedGraph({
          availability: options.availability,
          apiBase: options.apiBase,
          baseGraphViewVersion: selectedGraph.graph_view_version,
          displayName,
          graphViewId: selectedGraph.graph_view_id,
          incidentId: options.incidentId,
        });
        setGraphs((current) =>
          sortSavedGraphs(
            current.map((graph) =>
              graph.graph_view_id === renamed.graph_view_id ? renamed : graph,
            ),
          ),
        );
        setNotice(`Saved graph renamed to ${renamed.display_name}.`);
        options.onError(null);
        return true;
      } catch (caught) {
        reportError(caught, "The saved graph could not be renamed.");
        return false;
      } finally {
        setMutationPending(false);
      }
    },
    [
      options.apiBase,
      options.availability,
      options.incidentId,
      options.onError,
      reportError,
      selectedGraph,
    ],
  );

  const refreshGraph = useCallback(async () => {
    if (selectedGraph === null) return false;
    setMutationPending(true);
    try {
      const accepted = await refreshNetworkFlowSavedGraph({
        availability: options.availability,
        apiBase: options.apiBase,
        baseGraphViewVersion: selectedGraph.graph_view_version,
        graphViewId: selectedGraph.graph_view_id,
        incidentId: options.incidentId,
      });
      setGraphs((current) =>
        current.map((graph) =>
          graph.graph_view_id === accepted.graph_view.graph_view_id
            ? accepted.graph_view
            : graph,
        ),
      );
      setNotice(
        accepted.graph_view.selected_result === null
          ? `${accepted.graph_view.display_name} is queued for materialization.`
          : `${accepted.graph_view.display_name} is refreshing; its last successful result remains visible.`,
      );
      options.onError(null);
      return true;
    } catch (caught) {
      reportError(caught, "The saved graph could not be refreshed.");
      return false;
    } finally {
      setMutationPending(false);
    }
  }, [
    options.apiBase,
    options.availability,
    options.incidentId,
    options.onError,
    reportError,
    selectedGraph,
  ]);

  const retireGraph = useCallback(async () => {
    if (selectedGraph === null) return false;
    setMutationPending(true);
    try {
      await retireNetworkFlowSavedGraph({
        availability: options.availability,
        apiBase: options.apiBase,
        baseGraphViewVersion: selectedGraph.graph_view_version,
        graphViewId: selectedGraph.graph_view_id,
        incidentId: options.incidentId,
      });
      setGraphs((current) =>
        current.filter(
          (graph) => graph.graph_view_id !== selectedGraph.graph_view_id,
        ),
      );
      setSelectedGraphViewId(null);
      setResult(null);
      setNotice(`${selectedGraph.display_name} was retired.`);
      options.onError(null);
      return true;
    } catch (caught) {
      reportError(caught, "The saved graph could not be retired.");
      return false;
    } finally {
      setMutationPending(false);
    }
  }, [
    options.apiBase,
    options.availability,
    options.incidentId,
    options.onError,
    reportError,
    selectedGraph,
  ]);

  return {
    contributorState,
    contributors,
    createGraph,
    graphs,
    listState,
    loadGraphs,
    loadResult,
    mutationPending,
    notice,
    refreshGraph,
    renameGraph,
    result,
    resultState,
    retireGraph,
    selectedGraph,
    selectedGraphViewId,
    selection,
    selectGraphView: setSelectedGraphViewId,
    selectObject,
  };
}

function sortSavedGraphs(
  graphs: readonly NetworkFlowSavedGraph[],
): NetworkFlowSavedGraph[] {
  return [...graphs].sort(
    (left, right) =>
      left.normalized_display_name.localeCompare(
        right.normalized_display_name,
      ) || left.graph_view_id.localeCompare(right.graph_view_id),
  );
}

function materializationTransitionNotice(
  previous: readonly NetworkFlowSavedGraph[],
  next: readonly NetworkFlowSavedGraph[],
): string | null {
  for (const graph of next) {
    const prior = previous.find(
      (candidate) => candidate.graph_view_id === graph.graph_view_id,
    );
    if (
      prior === undefined ||
      (prior.last_materialization_status !== "queued" &&
        prior.last_materialization_status !== "running") ||
      (graph.last_materialization_status !== "succeeded" &&
        graph.last_materialization_status !== "failed" &&
        graph.last_materialization_status !== "cancelled")
    ) {
      continue;
    }
    if (graph.last_materialization_status === "succeeded") {
      return `${graph.display_name} materialization succeeded.`;
    }
    if (graph.last_materialization_status === "failed") {
      return graph.selected_result === null
        ? `${graph.display_name} materialization failed.`
        : `${graph.display_name} refresh failed; its last successful result remains visible.`;
    }
    return graph.selected_result === null
      ? `${graph.display_name} materialization was cancelled.`
      : `${graph.display_name} refresh was cancelled; its last successful result remains visible.`;
  }
  return null;
}
