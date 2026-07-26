import { useCallback } from "react";
import type { NetworkFlowTable } from "./networkFlowClient";
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
  tables,
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
  readonly tables: readonly NetworkFlowTable[];
}) {
  const handleResourceChange = useCallback(
    async (change: NetworkFlowExtensionResourceChange) => {
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
        const removedTable = tables.find(
          (table) => table.network_flow_table_id === change.resourceId,
        );
        dispatchTableAction({
          type: "remove_table",
          tableId: change.resourceId,
        });
        clearResources();
        if (removedTable !== undefined) {
          onMessage(`${removedTable.display_name} was deleted.`);
        }
        return;
      }
      onMessage("Network Analysis data changed.");
      await loadTables();
    },
    [
      clearResources,
      dispatchTableAction,
      loadTables,
      onMessage,
      onProtectedStateLoss,
      tables,
    ],
  );
  useNetworkFlowExtensionEvents({
    apiBase,
    enabled: true,
    incidentId,
    onResourceChange: handleResourceChange,
  });
}
