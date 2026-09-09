import type { NetworkFlowQueryLoadState } from "./useNetworkFlowPagedQuery";

/** The first applicable Network Flow status, independent of persistent recovery. */
export function networkFlowWorkspaceStatus(input: {
  readonly importStatus:
    | "validation_failed"
    | "mapping_required"
    | "validating"
    | null;
  readonly linkStatus: "link_pending" | "link_committed" | null;
  readonly graphState: NetworkFlowQueryLoadState;
  readonly hasGraph: boolean;
  readonly rejectedRows: number | null;
}) {
  if (input.importStatus !== null) return input.importStatus;
  if (input.linkStatus === "link_pending") return "link_pending";
  if (input.graphState === "loading" || input.graphState === "refreshing")
    return "graph_pending";
  if (input.hasGraph && input.graphState !== "ready") return "graph_stale";
  if (input.linkStatus === "link_committed") return "link_committed";
  if (input.hasGraph) return "graph_available";
  if (input.rejectedRows === null) return null;
  return input.rejectedRows > 0 ? "loaded_with_rejections" : "loaded";
}
