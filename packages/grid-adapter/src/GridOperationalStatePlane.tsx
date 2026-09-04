import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import {
  type MouseEvent as ReactMouseEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";

import {
  type GridDataState,
  type GridDataStateAction,
  type GridInteractionMode,
  type GridSurfaceIdentity,
  gridSurfaceIdentityKey,
} from "./core";
import {
  type GridInitialLoadingPhase,
  resolveGridDataStatePresentation,
  resolveGridInteractionModePresentation,
} from "./semanticDataState";

export type GridOperationalStatePlaneProps = {
  readonly accessibleLabel?: string | undefined;
  readonly dataState: GridDataState;
  readonly focusRoot: () => boolean;
  readonly interactionMode: GridInteractionMode;
  readonly surface: GridSurfaceIdentity;
};

export function GridOperationalStatePlane({
  accessibleLabel,
  dataState,
  focusRoot,
  interactionMode,
  surface,
}: GridOperationalStatePlaneProps) {
  const loadingPhase = useInitialLoadingPhase(dataState, surface);
  const dataPresentation = resolveGridDataStatePresentation(
    dataState,
    loadingPhase,
  );
  const interactionPresentation =
    resolveGridInteractionModePresentation(interactionMode);
  const suppressInteraction =
    cartularyDesignPresentation.gridStateComposition.suppressInteractionForDataStates.some(
      (state) => state === dataState.kind,
    );
  const interactionMessage =
    interactionPresentation.visible && !suppressInteraction
      ? interactionPresentation.message
      : null;
  const visible =
    dataPresentation.message !== null || interactionMessage !== null;
  const live =
    dataPresentation.live !== "none"
      ? dataPresentation.live
      : interactionPresentation.live;
  const role =
    dataPresentation.role !== "none"
      ? dataPresentation.role
      : interactionPresentation.role;
  const [actionPending, setActionPending] = useState(false);
  const actionOrigin = useRef<HTMLButtonElement | null>(null);
  const activatedAction = useRef<GridDataStateAction | null>(null);
  const action = dataPresentation.action;

  useEffect(() => {
    const previous = activatedAction.current;
    if (previous === null || previous === action) return;
    const origin = actionOrigin.current;
    activatedAction.current = null;
    actionOrigin.current = null;
    setActionPending(false);
    const activeElement = document.activeElement;
    if (
      origin !== null &&
      (activeElement === origin ||
        activeElement === document.body ||
        activeElement === null ||
        !activeElement.isConnected)
    ) {
      focusRoot();
    }
  }, [action, focusRoot]);

  const invokeAction = useCallback(
    (event: ReactMouseEvent<HTMLButtonElement>) => {
      if (action === undefined || actionPending) return;
      actionOrigin.current = event.currentTarget;
      activatedAction.current = action;
      setActionPending(true);
      try {
        action.onInvoke();
      } catch (error) {
        activatedAction.current = null;
        actionOrigin.current = null;
        setActionPending(false);
        throw error;
      }
      window.setTimeout(() => {
        if (document.activeElement !== actionOrigin.current) {
          activatedAction.current = null;
          actionOrigin.current = null;
          setActionPending(false);
        }
      }, 0);
    },
    [action, actionPending],
  );

  if (!visible) return null;

  return (
    <section
      aria-label={`${accessibleLabel ?? "Grid"} operational state`}
      className="cartulary-grid-operational-state-plane"
      data-grid-blocking={dataPresentation.blocking ? "true" : "false"}
      data-grid-data-state={dataState.kind}
      data-grid-placement={dataPresentation.placement}
      data-grid-posture={dataPresentation.posture}
    >
      <div className="cartulary-grid-operational-state-content">
        {role === "none" ? null : (
          <div
            aria-atomic="true"
            aria-live={live === "none" ? "off" : live}
            role={role}
          >
            {dataPresentation.message === null ? null : (
              <span>{dataPresentation.message}</span>
            )}
            {interactionMessage === null ? null : (
              <span
                className="cartulary-grid-interaction-state"
                data-grid-interaction-mode="read_only"
              >
                {interactionMessage}
              </span>
            )}
          </div>
        )}
        {action === undefined ? null : (
          <button disabled={actionPending} type="button" onClick={invokeAction}>
            {action.label}
          </button>
        )}
      </div>
    </section>
  );
}

function useInitialLoadingPhase(
  state: GridDataState,
  surface: GridSurfaceIdentity,
): GridInitialLoadingPhase {
  const loadingIdentity =
    state.kind === "initial_loading"
      ? `${gridSurfaceIdentityKey(surface)}\u0000${String(state.generationKey)}`
      : null;
  const [delayedIdentity, setDelayedIdentity] = useState<string | null>(null);

  useEffect(() => {
    setDelayedIdentity(null);
    if (loadingIdentity === null) return;
    const startedAt = performance.now();
    let timeout = 0;
    const markDelayed = () => {
      const remaining =
        cartularyDesignPresentation.initialLoading.delayMs -
        (performance.now() - startedAt);
      if (remaining > 0) {
        timeout = window.setTimeout(markDelayed, remaining);
        return;
      }
      setDelayedIdentity(loadingIdentity);
    };
    timeout = window.setTimeout(
      markDelayed,
      cartularyDesignPresentation.initialLoading.delayMs,
    );
    return () => window.clearTimeout(timeout);
  }, [loadingIdentity]);

  return loadingIdentity !== null && delayedIdentity === loadingIdentity
    ? "delayed"
    : "immediate";
}
