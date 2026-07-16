import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";

export type NetworkFlowControllerState = {
  readonly tables: readonly NetworkFlowTable[];
  readonly activeTableId: string | null;
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
};

export function networkFlowControllerReducer(
  state: NetworkFlowControllerState,
  action: NetworkFlowControllerAction,
): NetworkFlowControllerState {
  switch (action.type) {
    case "replace_tables": {
      const tables = [...action.tables];
      const activeTableId = tables.some(
        (table) => table.network_flow_table_id === state.activeTableId,
      )
        ? state.activeTableId
        : (tables[0]?.network_flow_table_id ?? null);
      return { tables, activeTableId };
    }
    case "select_table":
      return state.tables.some(
        (table) => table.network_flow_table_id === action.tableId,
      )
        ? { ...state, activeTableId: action.tableId }
        : state;
    case "replace_table":
      return state.tables.some(
        (table) =>
          table.network_flow_table_id === action.table.network_flow_table_id,
      )
        ? {
            ...state,
            tables: state.tables.map((table) =>
              table.network_flow_table_id === action.table.network_flow_table_id
                ? action.table
                : table,
            ),
          }
        : state;
    case "remove_table": {
      const removedIndex = state.tables.findIndex(
        (table) => table.network_flow_table_id === action.tableId,
      );
      const tables = state.tables.filter(
        (table) => table.network_flow_table_id !== action.tableId,
      );
      return {
        tables,
        activeTableId:
          state.activeTableId === action.tableId
            ? (tables[removedIndex]?.network_flow_table_id ??
              tables[removedIndex - 1]?.network_flow_table_id ??
              null)
            : state.activeTableId,
      };
    }
    case "clear_authorization":
      return initialNetworkFlowControllerState;
  }
}
