import { type RefObject, useCallback, useRef, useState } from "react";
import {
  type WorkbookRecoveryNavigation,
  workbookConflictRecoveryKey,
  workbookRecoveryKey,
} from "../../shared/workbookRecoveryNavigation";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";

/** Semantic routing and explicit focus requests; feature execution stays with its owner. */
export function useWorkbookRecoveryFocus({
  runtime,
  onSessionRecovery,
  navigation,
  invokerRef,
}: {
  readonly activeSurfaceRef: RefObject<HTMLElement | null>;
  readonly runtime: WorkbookMutationRuntime;
  readonly onSessionRecovery: () => Promise<void>;
  readonly navigation: WorkbookRecoveryNavigation;
  readonly invokerRef: RefObject<HTMLElement | null>;
}) {
  const [resolverActivation, setResolverActivation] = useState<{
    conflictKey: string;
    sequence: number;
  } | null>(null);
  const sameFieldSummaryRef = useRef<HTMLDivElement | null>(null);
  const activate = useCallback(
    (invoker: HTMLButtonElement, action: WorkbookStatusAction) => {
      invokerRef.current = invoker;
      if (
        action.kind === "session_recovery" ||
        runtime.getSnapshot().authPaused
      ) {
        void onSessionRecovery();
        return;
      }
      if (
        action.kind === "transaction_recovery" ||
        action.kind === "terminal_failure"
      ) {
        navigation.activate(
          workbookRecoveryKey("core", `fifo:${action.unitId}`),
        );
        return;
      }
      if (action.kind === "overflow") {
        navigation.activate(workbookRecoveryKey("core", "overflow"));
        return;
      }
      if (action.kind === "recovery_list") {
        navigation.openList();
        return;
      }
      if (action.kind === "surface_refresh") {
        const represented = navigation
          .getSnapshot()
          .entries.find((entry) =>
            entry.refreshViews?.includes(action.viewSchemaId),
          );
        navigation.activate(
          represented?.key ??
            workbookRecoveryKey("surface-refresh", action.viewSchemaId),
        );
        return;
      }
      if (action.kind !== "same_field_resolver") return;
      const conflict = runtime
        .getSnapshot()
        .conflicts.find((entry) => entry.key === action.conflictKey);
      setResolverActivation((current) => ({
        conflictKey: action.conflictKey,
        sequence: (current?.sequence ?? 0) + 1,
      }));
      navigation.activate(
        workbookConflictRecoveryKey(
          navigation.getSnapshot().entries,
          conflict ?? { key: action.conflictKey },
        ),
      );
    },
    [invokerRef, navigation, onSessionRecovery, runtime],
  );
  return {
    activate,
    resolverActivation,
    sameFieldSummaryRef,
    focusSameFieldSummary: () =>
      sameFieldSummaryRef.current?.focus({ preventScroll: true }),
  };
}
