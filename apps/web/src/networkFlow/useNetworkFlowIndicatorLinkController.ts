import type {
  NetworkFlowContributor,
  NetworkFlowEdgeAnnotation,
  NetworkFlowGraphResult,
} from "./networkFlowClient";
import { linkNetworkFlowIndicator } from "./networkFlowClient";

export function useNetworkFlowIndicatorLinkController({
  apiBase,
  firstContributor,
  graph,
  incidentId,
  onError,
  onMessage,
  selectedEdge,
}: {
  readonly apiBase: string | undefined;
  readonly firstContributor: NetworkFlowContributor | null;
  readonly graph: NetworkFlowGraphResult | null;
  readonly incidentId: string;
  readonly onError: (message: string | null) => void;
  readonly onMessage: (message: string) => void;
  readonly selectedEdge: NetworkFlowEdgeAnnotation | null;
}) {
  const linkEdge = async (
    fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
  ) => {
    if (graph === null || selectedEdge === null || firstContributor === null) {
      return;
    }
    const confirmExactValue = firstContributor.row[fieldKey];
    if (confirmExactValue.trim() === "") {
      onError("indicator_candidate_unavailable");
      return;
    }
    try {
      const result = await linkNetworkFlowIndicator({
        apiBase,
        incidentId,
        graph,
        edgeId: selectedEdge.edge_id,
        fieldKey,
        confirmExactValue,
      });
      onMessage(
        result.duplicate
          ? "Indicator link already exists."
          : "Indicator link created.",
      );
      onError(null);
    } catch (caught) {
      onError(
        caught instanceof Error ? caught.message : "indicator_link_failed",
      );
    }
  };
  return { linkEdge };
}
