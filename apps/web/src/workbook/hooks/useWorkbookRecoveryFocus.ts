import { type RefObject, useCallback, useLayoutEffect, useRef } from "react";
import type {
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "../runtime/WorkbookMutationRuntime";

type WorkbookRecoveryFocusOptions = {
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly runtime: WorkbookMutationRuntime;
  readonly snapshot: WorkbookMutationSnapshot;
};

/** Owns deterministic focus transfer among conflict and recovery surfaces. */
export function useWorkbookRecoveryFocus({
  activeSurfaceRef,
  runtime,
  snapshot,
}: WorkbookRecoveryFocusOptions) {
  const editRecoveryPanelRef = useRef<HTMLElement | null>(null);
  const overflowNoticeRef = useRef<HTMLElement | null>(null);
  const sameFieldSummaryRef = useRef<HTMLDivElement | null>(null);
  const conflictInvokerRef = useRef<HTMLButtonElement | null>(null);
  const pendingConflictFocusRef = useRef(false);
  const recoveryFocusOwnedRef = useRef(false);
  const previousRecoveryTargetKeyRef = useRef<string | null>(null);
  const recoveryTargetKey =
    snapshot.blockedEdit !== null
      ? `blocked:${snapshot.blockedEdit.unitId}`
      : snapshot.overflowMessage !== null
        ? "overflow"
        : null;

  const onFocusWithinChange = useCallback((focused: boolean) => {
    recoveryFocusOwnedRef.current = focused;
  }, []);
  const focusSameFieldSummary = useCallback(() => {
    sameFieldSummaryRef.current?.focus({ preventScroll: true });
  }, []);
  const activateConflictStatus = useCallback(
    (invoker: HTMLButtonElement) => {
      conflictInvokerRef.current = invoker;
      if (snapshot.blockedEdit !== null) {
        recoveryFocusOwnedRef.current = true;
        editRecoveryPanelRef.current?.focus({ preventScroll: true });
        return;
      }
      if (snapshot.overflowMessage !== null) {
        recoveryFocusOwnedRef.current = true;
        overflowNoticeRef.current?.focus({ preventScroll: true });
        return;
      }
      if (snapshot.conflicts.length === 0) return;
      pendingConflictFocusRef.current = true;
      runtime.activateConflict();
      if (snapshot.conflictPanelOpen) {
        pendingConflictFocusRef.current = false;
        focusSameFieldSummary();
      }
    },
    [focusSameFieldSummary, runtime, snapshot],
  );
  const activate =
    snapshot.blockedEdit !== null ||
    snapshot.overflowMessage !== null ||
    snapshot.conflicts.length > 0
      ? activateConflictStatus
      : undefined;

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
        snapshot.blockedEdit !== null
          ? editRecoveryPanelRef.current
          : overflowNoticeRef.current;
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
    snapshot.blockedEdit,
    snapshot.conflictPanelOpen,
  ]);

  return {
    activate,
    editRecoveryPanelRef,
    focusSameFieldSummary,
    onFocusWithinChange,
    overflowNoticeRef,
    sameFieldSummaryRef,
  };
}
