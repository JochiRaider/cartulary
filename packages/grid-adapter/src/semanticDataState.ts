import {
  type CartularyGridDataStatePresentation,
  type CartularyGridInteractionModePresentation,
  cartularyDesignPresentation,
  cartularyGridDataStatePresentation,
  cartularyGridInteractionModePresentation,
} from "@cartulary/ui-contracts";

import type {
  GridDataState,
  GridDataStateAction,
  GridInteractionMode,
} from "./core";

export type GridInitialLoadingPhase = "immediate" | "delayed";

export type GridDataStatePresentation = {
  readonly action?: GridDataStateAction | undefined;
  readonly actionRule: CartularyGridDataStatePresentation["actionRule"];
  readonly blocking: boolean;
  readonly draftRetention: CartularyGridDataStatePresentation["draftRetention"];
  readonly focusEffect: CartularyGridDataStatePresentation["focusEffect"];
  readonly live: CartularyGridDataStatePresentation["live"];
  readonly message: string | null;
  readonly placement: CartularyGridDataStatePresentation["placement"];
  readonly posture: CartularyGridDataStatePresentation["posture"];
  readonly role: CartularyGridDataStatePresentation["role"];
  readonly rowRetention: CartularyGridDataStatePresentation["rowRetention"];
  readonly semanticStateId: GridDataState["kind"];
};

export type GridInteractionModePresentation = {
  readonly focusEffect: CartularyGridInteractionModePresentation["focusEffect"];
  readonly live: CartularyGridInteractionModePresentation["live"];
  readonly message: string | null;
  readonly posture: CartularyGridInteractionModePresentation["posture"];
  readonly role: CartularyGridInteractionModePresentation["role"];
  readonly semanticStateId: GridInteractionMode["kind"];
  readonly visible: boolean;
};

export function resolveGridDataStatePresentation(
  state: GridDataState,
  initialLoadingPhase: GridInitialLoadingPhase = "immediate",
): GridDataStatePresentation {
  const policy = cartularyGridDataStatePresentation(state.kind);
  const message = resolveMessage(state, policy, initialLoadingPhase);
  const action = resolveAction(state, policy);
  return {
    ...(action === undefined ? {} : { action }),
    actionRule: policy.actionRule,
    blocking: policy.blocking,
    draftRetention: policy.draftRetention,
    focusEffect: policy.focusEffect,
    live: policy.live,
    message,
    placement: policy.placement,
    posture: policy.posture,
    role: policy.role,
    rowRetention: policy.rowRetention,
    semanticStateId: state.kind,
  };
}

export function resolveGridInteractionModePresentation(
  mode: GridInteractionMode,
): GridInteractionModePresentation {
  const policy = cartularyGridInteractionModePresentation(mode.kind);
  return {
    focusEffect: policy.focusEffect,
    live: policy.live,
    message:
      policy.messageStrategy === "owner_label" ? readOnlyLabel(mode) : null,
    posture: policy.posture,
    role: policy.role,
    semanticStateId: mode.kind,
    visible: policy.visible,
  };
}

export function gridDataStatePresentsAuthorizedRows(
  state: GridDataState,
): boolean {
  return (
    cartularyGridDataStatePresentation(state.kind).rowRetention !== "show_none"
  );
}

export function gridDataStateBlocksInteraction(state: GridDataState): boolean {
  return cartularyGridDataStatePresentation(state.kind).blocking;
}

export function gridDataStatePresentsDraft(state: GridDataState): boolean {
  return (
    cartularyGridDataStatePresentation(state.kind).draftRetention ===
    "show_owner_draft"
  );
}

function resolveMessage(
  state: GridDataState,
  policy: CartularyGridDataStatePresentation,
  initialLoadingPhase: GridInitialLoadingPhase,
): string | null {
  switch (policy.messageStrategy) {
    case "none":
      return null;
    case "loading_surface":
      if (state.kind !== "initial_loading") {
        return impossiblePolicy(policy.messageStrategy, state.kind);
      }
      return initialLoadingPhase === "delayed"
        ? cartularyDesignPresentation.initialLoading.message
        : interpolateSurfaceLabel(policy.message, state.surfaceLabel);
    case "refreshing_surface":
      if (state.kind !== "refreshing") {
        return impossiblePolicy(policy.messageStrategy, state.kind);
      }
      return interpolateSurfaceLabel(policy.message, state.surfaceLabel);
    case "owner_message":
      if (state.kind !== "empty" && state.kind !== "unavailable") {
        return impossiblePolicy(policy.messageStrategy, state.kind);
      }
      return state.message;
    case "design_message":
      return requireDesignMessage(policy);
    case "owner_message_with_design_suffix":
      if (state.kind !== "stale_error") {
        return impossiblePolicy(policy.messageStrategy, state.kind);
      }
      return `${state.message} ${requireDesignMessage(policy)}`;
    case "owner_message_or_design_fallback":
      if (state.kind !== "permission_denied") {
        return impossiblePolicy(policy.messageStrategy, state.kind);
      }
      return state.message ?? requireDesignMessage(policy);
  }
}

function resolveAction(
  state: GridDataState,
  policy: CartularyGridDataStatePresentation,
): GridDataStateAction | undefined {
  switch (policy.actionRule) {
    case "none":
      return undefined;
    case "owner_optional_create":
      if (state.kind !== "empty") {
        return impossiblePolicy(policy.actionRule, state.kind);
      }
      return state.action;
    case "clear_filters":
      if (state.kind !== "filtered_empty") {
        return impossiblePolicy(policy.actionRule, state.kind);
      }
      return state.action;
    case "owner_optional_retry":
      if (state.kind !== "stale_error" && state.kind !== "unavailable") {
        return impossiblePolicy(policy.actionRule, state.kind);
      }
      return state.action;
  }
}

function interpolateSurfaceLabel(
  message: string | null,
  surfaceLabel: string,
): string {
  return requireDesignMessageValue(message).replace(
    "{surface_label}",
    surfaceLabel,
  );
}

function requireDesignMessage(
  policy: CartularyGridDataStatePresentation,
): string {
  return requireDesignMessageValue(policy.message);
}

function requireDesignMessageValue(message: string | null): string {
  if (message === null) {
    throw new Error("Grid presentation policy requires design-owned copy");
  }
  return message;
}

function readOnlyLabel(mode: GridInteractionMode): string {
  if (mode.kind !== "read_only") {
    throw new Error("Grid interaction policy requires an owner label");
  }
  return mode.label;
}

function impossiblePolicy(rule: string, state: GridDataState["kind"]): never {
  throw new Error(`Grid presentation policy ${rule} is invalid for ${state}`);
}
