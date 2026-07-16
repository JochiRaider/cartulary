import {
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  useState,
} from "react";
import {
  listNetworkFlowTables,
  renameNetworkFlowTable,
  softDeleteNetworkFlowTable,
} from "./networkFlowClient";
import {
  initialNetworkFlowControllerState,
  networkFlowControllerReducer,
} from "./networkFlowController";
import {
  isNetworkFlowAuthorizationLoss,
  isNetworkFlowProtectedStateLoss,
  type NetworkFlowRequestError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";

export type NetworkFlowTableLoadState =
  | "loading"
  | "refreshing"
  | "ready"
  | "error";

export type NetworkFlowTableMutationState =
  | { readonly kind: "idle" }
  | { readonly kind: "renaming"; readonly tableId: string }
  | { readonly kind: "deleting"; readonly tableId: string };

export function useNetworkFlowTableController({
  apiBase,
  enabled,
  incidentId,
  onIncidentAccessLost,
}: {
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
}) {
  const [{ tables, activeTableId }, dispatch] = useReducer(
    networkFlowControllerReducer,
    initialNetworkFlowControllerState,
  );
  const listGeneration = useRef(0);
  const listController = useRef<AbortController | null>(null);
  const mutationGeneration = useRef(0);
  const [loadState, setLoadState] =
    useState<NetworkFlowTableLoadState>("loading");
  const [mutationState, setMutationState] =
    useState<NetworkFlowTableMutationState>({ kind: "idle" });
  const [error, setError] = useState<NetworkFlowRequestError | null>(null);

  const clearAuthorization = useCallback(() => {
    listController.current?.abort();
    listGeneration.current += 1;
    mutationGeneration.current += 1;
    dispatch({ type: "clear_authorization" });
    setMutationState({ kind: "idle" });
  }, []);

  const handleProtectedFailure = useCallback(
    (requestError: NetworkFlowRequestError) => {
      if (!isNetworkFlowProtectedStateLoss(requestError)) {
        return false;
      }
      clearAuthorization();
      if (isNetworkFlowAuthorizationLoss(requestError)) {
        onIncidentAccessLost?.();
      }
      return true;
    },
    [clearAuthorization, onIncidentAccessLost],
  );

  const loadTables = useCallback(
    async (selectTableId?: string) => {
      listController.current?.abort();
      listGeneration.current += 1;
      const generation = listGeneration.current;
      if (!enabled) {
        dispatch({ type: "clear_authorization" });
        setLoadState("ready");
        setError(null);
        return;
      }
      const controller = new AbortController();
      listController.current = controller;
      setLoadState((current) =>
        current === "ready" || current === "error" ? "refreshing" : "loading",
      );
      try {
        const nextTables = await listNetworkFlowTables({
          apiBase,
          incidentId,
          signal: controller.signal,
        });
        if (
          controller.signal.aborted ||
          generation !== listGeneration.current
        ) {
          return;
        }
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
        if (
          controller.signal.aborted ||
          generation !== listGeneration.current
        ) {
          return;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow tables could not be loaded.",
        );
        handleProtectedFailure(requestError);
        setLoadState("error");
        setError(requestError);
      }
    },
    [apiBase, enabled, handleProtectedFailure, incidentId],
  );

  useEffect(() => {
    void loadTables();
    return () => listController.current?.abort();
  }, [loadTables]);

  const renameTable = useCallback(
    async (options: {
      readonly baseTableVersion: number;
      readonly displayName: string;
      readonly tableId: string;
    }): Promise<boolean> => {
      if (!enabled || mutationState.kind !== "idle") {
        return false;
      }
      mutationGeneration.current += 1;
      const generation = mutationGeneration.current;
      setMutationState({ kind: "renaming", tableId: options.tableId });
      setError(null);
      try {
        const table = await renameNetworkFlowTable({
          apiBase,
          baseTableVersion: options.baseTableVersion,
          displayName: options.displayName,
          incidentId,
          tableId: options.tableId,
        });
        if (generation !== mutationGeneration.current) {
          return false;
        }
        dispatch({ type: "replace_table", table });
        setMutationState({ kind: "idle" });
        return true;
      } catch (caught) {
        if (generation !== mutationGeneration.current) {
          return false;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow table could not be renamed.",
        );
        const protectedFailure = handleProtectedFailure(requestError);
        if (
          !protectedFailure &&
          requestError.code === "network_flow_table_version_conflict"
        ) {
          await loadTables(options.tableId);
        }
        if (!protectedFailure) {
          setMutationState({ kind: "idle" });
        }
        setError(requestError);
        return false;
      }
    },
    [
      apiBase,
      enabled,
      handleProtectedFailure,
      incidentId,
      loadTables,
      mutationState.kind,
    ],
  );

  const softDeleteTable = useCallback(
    async (options: {
      readonly baseTableVersion: number;
      readonly tableId: string;
    }): Promise<boolean> => {
      if (!enabled || mutationState.kind !== "idle") {
        return false;
      }
      mutationGeneration.current += 1;
      const generation = mutationGeneration.current;
      setMutationState({ kind: "deleting", tableId: options.tableId });
      setError(null);
      try {
        await softDeleteNetworkFlowTable({
          apiBase,
          baseTableVersion: options.baseTableVersion,
          incidentId,
          tableId: options.tableId,
        });
        if (generation !== mutationGeneration.current) {
          return false;
        }
        dispatch({ type: "remove_table", tableId: options.tableId });
        setMutationState({ kind: "idle" });
        return true;
      } catch (caught) {
        if (generation !== mutationGeneration.current) {
          return false;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow table could not be deleted.",
        );
        const protectedFailure = handleProtectedFailure(requestError);
        if (
          !protectedFailure &&
          requestError.code === "network_flow_table_version_conflict"
        ) {
          await loadTables(options.tableId);
        }
        if (!protectedFailure) {
          setMutationState({ kind: "idle" });
        }
        setError(requestError);
        return false;
      }
    },
    [
      apiBase,
      enabled,
      handleProtectedFailure,
      incidentId,
      loadTables,
      mutationState.kind,
    ],
  );

  return {
    activeTable: useMemo(
      () =>
        tables.find((table) => table.network_flow_table_id === activeTableId) ??
        null,
      [activeTableId, tables],
    ),
    activeTableId,
    clearAuthorization,
    dispatch,
    error,
    loadState,
    loadTables,
    mutationState,
    renameTable,
    softDeleteTable,
    tableIds: useMemo(
      () => tables.map((table) => table.network_flow_table_id),
      [tables],
    ),
    tables,
  };
}
