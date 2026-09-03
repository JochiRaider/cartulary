import type { GridDataState } from "./core";

export type GridDataStatePresentation = {
  readonly action?: {
    readonly label: string;
    readonly onInvoke: () => void;
  };
  readonly blocking: boolean;
  readonly live: "assertive" | "polite";
  readonly message: string;
  readonly role: "alert" | "status";
};

export function resolveGridDataStatePresentation(
  state: GridDataState,
  delayedInitialLoadingMessage?: string,
): GridDataStatePresentation | null {
  switch (state.kind) {
    case "ready":
      return null;
    case "initial_loading":
      return presentation(
        delayedInitialLoadingMessage ?? `Loading ${state.surfaceLabel}…`,
        true,
      );
    case "refreshing":
      return presentation(`Refreshing ${state.surfaceLabel}…`, false);
    case "empty":
      return presentation(state.message, false, state.action);
    case "filtered_empty":
      return presentation(
        "No rows match the current filters.",
        false,
        state.action,
      );
    case "stale_error":
      return presentation(
        `${state.message} Previously loaded rows may be stale.`,
        false,
        state.action,
        true,
      );
    case "unavailable":
      return presentation(state.message, true, state.action, true);
    case "permission_denied":
      return presentation(
        state.message ?? "You no longer have access to this workbook.",
        true,
        undefined,
        true,
      );
  }
}

function presentation(
  message: string,
  blocking: boolean,
  action?: GridDataStatePresentation["action"],
  assertive = false,
): GridDataStatePresentation {
  return {
    ...(action === undefined ? {} : { action }),
    blocking,
    live: assertive ? "assertive" : "polite",
    message,
    role: assertive ? "alert" : "status",
  };
}
