import type {
  GridDataState,
  GridDataStateAction,
  GridInteractionMode,
} from "@cartulary/grid-adapter";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { WorkbookQueryState } from "./workbookQuery";

export type WorkbookQueryLoadState =
  | { readonly kind: "initial_loading" }
  | { readonly kind: "ready" }
  | { readonly kind: "refreshing" }
  | { readonly kind: "stale_error"; readonly message: string }
  | { readonly kind: "unavailable"; readonly message: string }
  | { readonly kind: "permission_denied"; readonly message?: string };

export const initialWorkbookQueryLoadState: WorkbookQueryLoadState = {
  kind: "initial_loading",
};

export function workbookGridDataState({
  emptyAction,
  emptyMessage,
  loadState,
  onClearFilters,
  onRetry,
  queryState,
  rowCount,
  surfaceLabel,
}: {
  readonly emptyAction?: GridDataStateAction | undefined;
  readonly emptyMessage: string;
  readonly loadState: WorkbookQueryLoadState;
  readonly onClearFilters: () => void;
  readonly onRetry: () => void;
  readonly queryState: WorkbookQueryState;
  readonly rowCount: number;
  readonly surfaceLabel: string;
}): GridDataState {
  switch (loadState.kind) {
    case "initial_loading":
      return { kind: "initial_loading", surfaceLabel };
    case "refreshing":
      return { kind: "refreshing", surfaceLabel };
    case "stale_error":
      return {
        action: { label: "Retry", onInvoke: onRetry },
        kind: "stale_error",
        message: loadState.message,
      };
    case "unavailable":
      return {
        action: { label: "Retry", onInvoke: onRetry },
        kind: "unavailable",
        message: loadState.message,
      };
    case "permission_denied":
      return {
        ...(loadState.message === undefined
          ? {}
          : { message: loadState.message }),
        kind: "permission_denied",
      };
    case "ready":
      if (rowCount > 0) return { kind: "ready" };
      if (queryState.filters.length > 0) {
        return {
          action: { label: "Clear filters", onInvoke: onClearFilters },
          kind: "filtered_empty",
        };
      }
      return {
        ...(emptyAction === undefined ? {} : { action: emptyAction }),
        kind: "empty",
        message: emptyMessage,
      };
  }
}

export function workbookGridInteractionMode(
  incidentStatus: "active" | "closed" | undefined,
  role: WorkbookIncidentRole | null,
): GridInteractionMode {
  if (incidentStatus === "closed") {
    return { kind: "read_only", label: "Closed, read-only" };
  }
  if (role === "admin" || role === "editor" || role === "reviewer") {
    return { kind: "editable" };
  }
  return { kind: "read_only", label: "Read-only" };
}
