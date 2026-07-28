export type WorkbookLifecycleState = {
  readonly isInitialLoading: boolean;
  readonly isRefreshing: boolean;
  readonly loadError: string | null;
  readonly refreshError: string | null;
};

export type WorkbookLifecycleAction =
  | { readonly type: "initial_loading"; readonly value: boolean }
  | { readonly type: "refreshing"; readonly value: boolean }
  | { readonly type: "load_error"; readonly value: string | null }
  | { readonly type: "refresh_error"; readonly value: string | null };

export const initialWorkbookLifecycleState: WorkbookLifecycleState = {
  isInitialLoading: true,
  isRefreshing: false,
  loadError: null,
  refreshError: null,
};

export function reduceWorkbookLifecycle(
  state: WorkbookLifecycleState,
  action: WorkbookLifecycleAction,
): WorkbookLifecycleState {
  switch (action.type) {
    case "initial_loading":
      return { ...state, isInitialLoading: action.value };
    case "refreshing":
      return { ...state, isRefreshing: action.value };
    case "load_error":
      return { ...state, loadError: action.value };
    case "refresh_error":
      return { ...state, refreshError: action.value };
  }
}
