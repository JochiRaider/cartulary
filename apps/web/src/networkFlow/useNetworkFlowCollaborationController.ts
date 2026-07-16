import { useCallback } from "react";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  type NetworkFlowExtensionResourceChange,
  useNetworkFlowExtensionEvents,
} from "./useNetworkFlowExtensionEvents";

export function useNetworkFlowCollaborationController({
  apiBase,
  clearResources,
  dispatchTableAction,
  incidentId,
  loadTables,
  onMessage,
  onProtectedStateLoss,
}: {
  readonly apiBase: string | undefined;
  readonly clearResources: () => void;
  readonly dispatchTableAction: (
    action:
      | {
          readonly type: "clear_authorization";
        }
      | {
          readonly type: "remove_table";
          readonly tableId: string;
        },
  ) => void;
  readonly incidentId: string;
  readonly loadTables: () => Promise<void>;
  readonly onMessage: (message: string) => void;
  readonly onProtectedStateLoss: (error: NetworkFlowRequestError) => void;
}) {
  const handleResourceChange = useCallback(
    async (change: NetworkFlowExtensionResourceChange) => {
      onMessage("Network Analysis data changed.");
      if (
        change.reasonCode === "authorization_lost" ||
        (change.changeKind === "remove" && change.resourceId === "*")
      ) {
        const incidentClosed = change.reasonCode === "incident_closed";
        onProtectedStateLoss(
          new NetworkFlowRequestError({
            code: incidentClosed ? "incident_closed" : "authorization_denied",
            retryAction: "do_not_retry",
            retryable: false,
            safeMessage: incidentClosed
              ? "This incident is closed. Network Analysis is unavailable."
              : "You no longer have access to Network Analysis for this incident.",
            status: incidentClosed ? 409 : 403,
          }),
        );
        dispatchTableAction({ type: "clear_authorization" });
        clearResources();
        onMessage(
          incidentClosed
            ? "This incident is closed. Network Analysis data was cleared."
            : "Network Analysis access changed. Protected data was cleared.",
        );
        return;
      }
      if (change.changeKind === "remove") {
        dispatchTableAction({
          type: "remove_table",
          tableId: change.resourceId,
        });
        clearResources();
        return;
      }
      await loadTables();
    },
    [
      clearResources,
      dispatchTableAction,
      loadTables,
      onMessage,
      onProtectedStateLoss,
    ],
  );
  useNetworkFlowExtensionEvents({
    apiBase,
    enabled: true,
    incidentId,
    onResourceChange: handleResourceChange,
  });
}
