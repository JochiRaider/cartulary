import {
  type RefObject,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type {
  WorkbookMutationRuntime,
  WorkbookStatusPresentation,
} from "../runtime/WorkbookMutationRuntime";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";

type WorkbookRecoveryFocusOptions = {
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly runtime: WorkbookMutationRuntime;
  readonly onSessionRecovery: () => Promise<void>;
  readonly snapshot: WorkbookStatusPresentation;
};

/** Owns deterministic focus transfer among conflict and recovery surfaces. */
export function useWorkbookRecoveryFocus({
  activeSurfaceRef,
  runtime,
  snapshot,
  onSessionRecovery,
}: WorkbookRecoveryFocusOptions) {
  const [resolverActivation, setResolverActivation] = useState<{
    conflictKey: string;
    sequence: number;
  } | null>(null);
  const editRecoveryPanelRef = useRef<HTMLElement | null>(null);
  const overflowNoticeRef = useRef<HTMLElement | null>(null);
  const sameFieldSummaryRef = useRef<HTMLDivElement | null>(null);
  const conflictInvokerRef = useRef<HTMLButtonElement | null>(null);
  const pendingConflictFocusRef = useRef(false);
  const recoveryFocusOwnedRef = useRef(false);
  const previousRecoveryTargetKeyRef = useRef<string | null>(null);
  const recoveryTargetKey =
    snapshot.action?.kind === "transaction_recovery" ||
    snapshot.action?.kind === "terminal_failure"
      ? `blocked:${snapshot.action.unitId}`
      : snapshot.action?.kind === "overflow"
        ? "overflow"
        : snapshot.action?.kind === "same_field_resolver" &&
            snapshot.conflictPanelOpen
          ? `resolver:${snapshot.action.conflictKey}`
          : null;

  const onFocusWithinChange = useCallback((focused: boolean) => {
    recoveryFocusOwnedRef.current = focused;
  }, []);
  const focusSameFieldSummary = useCallback(() => {
    sameFieldSummaryRef.current?.focus({ preventScroll: true });
  }, []);
  const activateConflictStatus = useCallback(
    (invoker: HTMLButtonElement, action: WorkbookStatusAction) => {
      conflictInvokerRef.current = invoker;
      if (action.kind === "session_recovery") {
        void onSessionRecovery();
        return;
      }
      if (
        action.kind === "transaction_recovery" ||
        action.kind === "terminal_failure"
      ) {
        recoveryFocusOwnedRef.current = true;
        editRecoveryPanelRef.current?.focus({ preventScroll: true });
        return;
      }
      if (action.kind === "overflow") {
        recoveryFocusOwnedRef.current = true;
        overflowNoticeRef.current?.focus({ preventScroll: true });
        return;
      }
      if (
        action.kind !== "same_field_resolver" ||
        !snapshot.conflicts.some((entry) => entry.key === action.conflictKey)
      )
        return;
      setResolverActivation((current) => ({
        conflictKey: action.conflictKey,
        sequence: (current?.sequence ?? 0) + 1,
      }));
      pendingConflictFocusRef.current = true;
      recoveryFocusOwnedRef.current = true;
      runtime.activateConflict();
      if (snapshot.conflictPanelOpen) {
        pendingConflictFocusRef.current = false;
        focusSameFieldSummary();
      }
    },
    [focusSameFieldSummary, onSessionRecovery, runtime, snapshot],
  );
  const activate =
    snapshot.action === null ? undefined : activateConflictStatus;

  useLayoutEffect(() => {
    if (!pendingConflictFocusRef.current || !snapshot.conflictPanelOpen) {
      return;
    }
    pendingConflictFocusRef.current = false;
    focusSameFieldSummary();
  }, [focusSameFieldSummary, snapshot.conflictPanelOpen]);

  useLayoutEffect(() => {
    const previousTarget = previousRecoveryTargetKeyRef.current;
    previousRecoveryTargetKeyRef.current = recoveryTargetKey;
    if (
      previousTarget === null ||
      previousTarget === recoveryTargetKey ||
      !recoveryFocusOwnedRef.current
    ) {
      return;
    }
    if (recoveryTargetKey !== null) {
      const nextTarget =
        snapshot.action?.kind === "same_field_resolver"
          ? sameFieldSummaryRef.current
          : snapshot.action?.kind === "overflow"
            ? overflowNoticeRef.current
            : editRecoveryPanelRef.current;
      nextTarget?.focus({ preventScroll: true });
      return;
    }
    recoveryFocusOwnedRef.current = false;
    if (snapshot.conflictPanelOpen) {
      focusSameFieldSummary();
      return;
    }
    const invoker = conflictInvokerRef.current;
    if (invoker?.isConnected && !invoker.disabled) {
      invoker.focus({ preventScroll: true });
      return;
    }
    activeSurfaceRef.current?.focus({ preventScroll: true });
  }, [
    activeSurfaceRef,
    focusSameFieldSummary,
    recoveryTargetKey,
    snapshot.action,
    snapshot.conflictPanelOpen,
  ]);

  return {
    activate,
    resolverActivation,
    editRecoveryPanelRef,
    focusSameFieldSummary,
    onFocusWithinChange,
    overflowNoticeRef,
    sameFieldSummaryRef,
  };
}
