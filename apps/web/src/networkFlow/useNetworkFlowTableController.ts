import { useLayoutEffect, useMemo, useSyncExternalStore } from "react";
import type { NetworkFlowTableController } from "./NetworkFlowTableController";
export type NetworkFlowTableLoadState =
  | "loading"
  | "refreshing"
  | "ready"
  | "error";

/** Presentation subscribes; the workbook owns authority and operation lifetime. */
export function useNetworkFlowTableController({
  controller,
  enabled,
}: {
  readonly controller: NetworkFlowTableController;
  readonly enabled: boolean;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    controller.setActive(enabled);
    return () => controller.setActive(false);
  }, [controller, enabled]);
  return {
    ...state,
    controller,
    activeTable: useMemo(
      () =>
        state.tables.find(
          (table) => table.network_flow_table_id === state.activeTableId,
        ) ?? null,
      [state.tables, state.activeTableId],
    ),
    tableIds: useMemo(
      () => state.tables.map((table) => table.network_flow_table_id),
      [state.tables],
    ),
    clearAuthorization: controller.clearAuthorization,
    selectTable: controller.selectTable,
    loadTables: controller.loadTables,
    handoffImportedTable: controller.handoffImportedTable,
  };
}
