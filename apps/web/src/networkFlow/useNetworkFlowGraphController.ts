import { useEffect, useMemo, useState } from "react";
import {
  type NetworkFlowContributor,
  type NetworkFlowGraphResult,
  queryNetworkFlowContributors,
  queryNetworkFlowGraph,
} from "./networkFlowClient";
import {
  isNetworkFlowAuthorizationLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";

export function useNetworkFlowGraphController({
  apiBase,
  enabled,
  incidentId,
  onError,
  onIncidentAccessLost,
  tableIds,
}: {
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowRequestError | null) => void;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly tableIds: readonly string[];
}) {
  const [graph, setGraph] = useState<NetworkFlowGraphResult | null>(null);
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null);
  const [contributors, setContributors] = useState<NetworkFlowContributor[]>(
    [],
  );

  useEffect(() => {
    if (!enabled || tableIds.length === 0) {
      setGraph(null);
      setSelectedEdgeId(null);
      setContributors([]);
      return;
    }
    const controller = new AbortController();
    void queryNetworkFlowGraph({
      apiBase,
      incidentId,
      tableIds,
      signal: controller.signal,
    })
      .then((nextGraph) => {
        if (controller.signal.aborted) {
          return;
        }
        setGraph(nextGraph);
        setSelectedEdgeId((current) =>
          current !== null &&
          nextGraph.edge_annotations.some((edge) => edge.edge_id === current)
            ? current
            : (nextGraph.edge_annotations[0]?.edge_id ?? null),
        );
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
        onError(requestError);
      });
    return () => controller.abort();
  }, [apiBase, enabled, incidentId, onError, onIncidentAccessLost, tableIds]);

  useEffect(() => {
    if (graph === null || selectedEdgeId === null) {
      setContributors([]);
      return;
    }
    const controller = new AbortController();
    void queryNetworkFlowContributors({
      apiBase,
      incidentId,
      graph,
      selector: { kind: "edge", edge_id: selectedEdgeId },
      signal: controller.signal,
    })
      .then((result) => {
        if (!controller.signal.aborted) {
          setContributors(result.contributors);
          onError(null);
        }
      })
      .catch((caught: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow contributor query failed.",
        );
        if (isNetworkFlowAuthorizationLoss(requestError)) {
          onIncidentAccessLost?.();
        }
        onError(requestError);
        setContributors([]);
      });
    return () => controller.abort();
  }, [
    apiBase,
    graph,
    incidentId,
    onError,
    onIncidentAccessLost,
    selectedEdgeId,
  ]);

  const clearGraph = () => {
    setGraph(null);
    setSelectedEdgeId(null);
    setContributors([]);
  };

  return {
    clearGraph,
    contributors,
    firstContributor: contributors[0] ?? null,
    graph,
    selectedEdge: useMemo(
      () =>
        graph?.edge_annotations.find(
          (edge) => edge.edge_id === selectedEdgeId,
        ) ?? null,
      [graph, selectedEdgeId],
    ),
    selectedEdgeId,
    setSelectedEdgeId,
  };
}
