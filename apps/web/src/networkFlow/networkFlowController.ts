import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";

/** Knowledge retained only for the authorized workbook lifetime. No global revision exists. */
export type NetworkFlowControllerState = {
  readonly tables: readonly NetworkFlowTable[];
  readonly activeTableId: string | null;
  readonly removedTableIds: readonly string[];
};
export type NetworkFlowControllerAction =
  | {
      readonly type: "replace_tables";
      readonly tables: readonly NetworkFlowTable[];
    }
  | { readonly type: "select_table"; readonly tableId: string }
  | { readonly type: "replace_table"; readonly table: NetworkFlowTable }
  | { readonly type: "remove_table"; readonly tableId: string }
  | { readonly type: "clear_authorization" };
export const initialNetworkFlowControllerState: NetworkFlowControllerState = {
  tables: [],
  activeTableId: null,
  removedTableIds: [],
};
function selection(
  state: NetworkFlowControllerState,
  tables: readonly NetworkFlowTable[],
): string | null {
  const has = (id: string | null) =>
    tables.some((table) => table.network_flow_table_id === id);
  if (has(state.activeTableId)) return state.activeTableId;
  const index = state.tables.findIndex(
    (table) => table.network_flow_table_id === state.activeTableId,
  );
  if (index >= 0) {
    const next = state.tables
      .slice(index + 1)
      .find((table) => has(table.network_flow_table_id));
    const previous = state.tables
      .slice(0, index)
      .reverse()
      .find((table) => has(table.network_flow_table_id));
    if (next || previous)
      return (next ?? previous)?.network_flow_table_id ?? null;
  }
  return tables[0]?.network_flow_table_id ?? null;
}
export function networkFlowControllerReducer(
  state: NetworkFlowControllerState,
  action: NetworkFlowControllerAction,
): NetworkFlowControllerState {
  switch (action.type) {
    case "replace_tables": {
      const removed = new Set(state.removedTableIds);
      const incoming = new Map(
        action.tables.map((table) => [table.network_flow_table_id, table]),
      );
      for (const table of state.tables)
        if (!incoming.has(table.network_flow_table_id))
          removed.add(table.network_flow_table_id);
      const known = new Map(
        state.tables.map((table) => [table.network_flow_table_id, table]),
      );
      const tables = action.tables
        .filter(
          (table) =>
            !removed.has(table.network_flow_table_id) &&
            table.table_status === "active",
        )
        .map((table) => {
          const previous = known.get(table.network_flow_table_id);
          return previous && previous.table_version >= table.table_version
            ? previous
            : table;
        });
      return {
        tables,
        activeTableId: selection(state, tables),
        removedTableIds: [...removed],
      };
    }
    case "select_table":
      return state.tables.some(
        (table) => table.network_flow_table_id === action.tableId,
      )
        ? { ...state, activeTableId: action.tableId }
        : state;
    case "replace_table":
      if (state.removedTableIds.includes(action.table.network_flow_table_id))
        return state;
      if (action.table.table_status !== "active")
        return networkFlowControllerReducer(state, {
          type: "remove_table",
          tableId: action.table.network_flow_table_id,
        });
      return {
        ...state,
        tables: state.tables.map((table) =>
          table.network_flow_table_id === action.table.network_flow_table_id &&
          table.table_version < action.table.table_version
            ? action.table
            : table,
        ),
      };
    case "remove_table": {
      const tables = state.tables.filter(
        (table) => table.network_flow_table_id !== action.tableId,
      );
      return {
        tables,
        activeTableId: selection(state, tables),
        removedTableIds: [
          ...new Set([...state.removedTableIds, action.tableId]),
        ],
      };
    }
    case "clear_authorization":
      return initialNetworkFlowControllerState;
  }
}
