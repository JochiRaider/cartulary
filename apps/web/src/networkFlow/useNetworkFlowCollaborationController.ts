import { useCallback } from "react";
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
}) {
  const handleResourceChange = useCallback(
    async (change: NetworkFlowExtensionResourceChange) => {
      onMessage("Network Analysis data changed.");
      if (
        change.reasonCode === "authorization_lost" ||
        (change.changeKind === "remove" && change.resourceId === "*")
      ) {
        dispatchTableAction({ type: "clear_authorization" });
        clearResources();
        onMessage("Network Analysis access changed.");
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
    [clearResources, dispatchTableAction, loadTables, onMessage],
  );
  useNetworkFlowExtensionEvents({
    apiBase,
    enabled: true,
    incidentId,
    onResourceChange: handleResourceChange,
  });
}
