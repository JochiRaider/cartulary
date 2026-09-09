import {
  type ChangeEvent,
  useCallback,
  useLayoutEffect,
  useSyncExternalStore,
} from "react";
import type { NetworkFlowImportController } from "./NetworkFlowImportController";
import {
  type NetworkFlowImportHandoffRequest,
  type NetworkFlowImportHandoffResult,
  networkFlowImportHasWork,
} from "./networkFlowImportState";

/** The workspace binds presentation and table selection to the session-owned operation. */
export function useNetworkFlowImportController({
  controller,
  onImported,
}: {
  readonly controller: NetworkFlowImportController;
  readonly onImported: (
    request: NetworkFlowImportHandoffRequest,
  ) => Promise<NetworkFlowImportHandoffResult>;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    controller.setWorkspaceActive(true);
    return () => controller.setWorkspaceActive(false);
  }, [controller]);
  useLayoutEffect(() => {
    controller.setHandoff(onImported);
    return () => controller.setHandoff(null);
  }, [controller, onImported]);
  const handleImportChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      const file = event.currentTarget.files?.[0];
      event.currentTarget.value = "";
      if (file) {
        if (controller.getSnapshot().stage === "finished")
          controller.startNew();
        void controller.upload(file);
      }
    },
    [controller],
  );
  return {
    state,
    handleImportChange,
    hasWork: networkFlowImportHasWork(state),
    importing:
      state.write?.disposition === "pending" ||
      Boolean((state.applyJob ?? state.discoveryJob)?.observing),
  };
}
