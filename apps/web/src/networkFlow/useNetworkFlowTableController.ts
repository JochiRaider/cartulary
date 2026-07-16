import { useCallback, useEffect, useMemo, useReducer, useState } from "react";
import { listNetworkFlowTables } from "./networkFlowClient";
import {
  initialNetworkFlowControllerState,
  networkFlowControllerReducer,
} from "./networkFlowController";
import { isNetworkFlowAuthorizationLoss } from "./networkFlowErrors";

export function useNetworkFlowTableController({
  apiBase,
  incidentId,
  onIncidentAccessLost,
}: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const [{ tables, activeTableId }, dispatch] = useReducer(
    networkFlowControllerReducer,
    initialNetworkFlowControllerState,
  );
  const [loadState, setLoadState] = useState<"loading" | "ready" | "error">(
    "loading",
  );
  const [error, setError] = useState<string | null>(null);

  const loadTables = useCallback(
    async (selectTableId?: string) => {
      setLoadState((current) => (current === "ready" ? current : "loading"));
      try {
        const nextTables = await listNetworkFlowTables({ apiBase, incidentId });
        dispatch({ type: "replace_tables", tables: nextTables });
        if (
          selectTableId !== undefined &&
          nextTables.some(
            (table) => table.network_flow_table_id === selectTableId,
          )
        ) {
          dispatch({ type: "select_table", tableId: selectTableId });
        }
        setLoadState("ready");
        setError(null);
      } catch (caught) {
        const message =
          caught instanceof Error ? caught.message : "load_failed";
        if (isNetworkFlowAuthorizationLoss(message)) {
          onIncidentAccessLost?.();
        }
        setLoadState("error");
        setError(message);
      }
    },
    [apiBase, incidentId, onIncidentAccessLost],
  );

  useEffect(() => {
    void loadTables();
  }, [loadTables]);

  return {
    activeTable: useMemo(
      () =>
        tables.find((table) => table.network_flow_table_id === activeTableId) ??
        null,
      [activeTableId, tables],
    ),
    activeTableId,
    dispatch,
    error,
    loadState,
    loadTables,
    tableIds: useMemo(
      () => tables.map((table) => table.network_flow_table_id),
      [tables],
    ),
    tables,
  };
}
