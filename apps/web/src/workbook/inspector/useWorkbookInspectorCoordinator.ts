import type { InspectorConfig } from "@cartulary/view-contracts";
import {
  type SetStateAction,
  useCallback,
  useLayoutEffect,
  useReducer,
  useRef,
} from "react";
import {
  initialWorkbookInspectorState,
  type WorkbookInspectorInvalidationReason,
  workbookInspectorReducer,
  workbookInspectorStateIsOpen,
} from "../models/workbookInspectorModel";
import {
  type WorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "./workbookInspectorSubject";

type WorkbookInspectorOwnerActionPorts = {
  readonly resetOwnerState: (event: WorkbookInspectorResetEvent) => void;
  readonly restoreFocus: () => void;
};

type WorkbookInspectorResetCause =
  | "close"
  | "retarget"
  | "action_completed"
  | "surface_changed"
  | "authorization_lost"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "hard_refresh";

type WorkbookInspectorResetScope = "row_local" | "surface";

type WorkbookInspectorResetEvent = {
  readonly cause: WorkbookInspectorResetCause;
  readonly scope: WorkbookInspectorResetScope;
};

function workbookInspectorResetScope(
  cause: WorkbookInspectorResetCause,
): WorkbookInspectorResetScope {
  switch (cause) {
    case "close":
    case "retarget":
    case "action_completed":
      return "row_local";
    case "surface_changed":
    case "authorization_lost":
    case "incident_closed":
    case "record_deleted":
    case "record_merged":
    case "hard_refresh":
      return "surface";
  }
}

export function useWorkbookInspectorCoordinator({
  actionPorts,
  config,
  lifecycleKey,
  subject,
}: {
  readonly actionPorts: WorkbookInspectorOwnerActionPorts;
  readonly config: InspectorConfig;
  readonly lifecycleKey: string;
  readonly subject: WorkbookInspectorSubject | null;
}) {
  const effectiveLifecycleKey = `${config.viewSchemaId.length}:${config.viewSchemaId}${lifecycleKey}`;
  const [snapshot, dispatch] = useReducer(
    workbookInspectorReducer,
    { lifecycleKey: effectiveLifecycleKey },
    initialWorkbookInspectorState,
  );
  const actionPortsRef = useRef(actionPorts);
  const observedInputsRef = useRef<{
    lifecycleKey: string;
    subject: WorkbookInspectorSubject | null;
  }>({ lifecycleKey: effectiveLifecycleKey, subject: null });
  const currentSubjectRef = useRef(subject);
  const snapshotRef = useRef(snapshot);

  useLayoutEffect(() => {
    actionPortsRef.current = actionPorts;
    snapshotRef.current = snapshot;
  });

  const resetOwnerState = useCallback((cause: WorkbookInspectorResetCause) => {
    actionPortsRef.current.resetOwnerState({
      cause,
      scope: workbookInspectorResetScope(cause),
    });
  }, []);

  useLayoutEffect(() => {
    currentSubjectRef.current = subject;
    const observed = observedInputsRef.current;
    if (observed.lifecycleKey !== effectiveLifecycleKey) {
      observedInputsRef.current = {
        lifecycleKey: effectiveLifecycleKey,
        subject,
      };
      resetOwnerState("surface_changed");
      dispatch({
        lifecycleKey: effectiveLifecycleKey,
        type: "lifecycle_changed",
      });
      return;
    }
    observedInputsRef.current = {
      lifecycleKey: effectiveLifecycleKey,
      subject,
    };
    if (workbookInspectorSubjectsEqual(observed.subject, subject)) {
      return;
    }
    resetOwnerState("retarget");
    dispatch({
      lifecycleKey: effectiveLifecycleKey,
      type: "retarget",
      subject,
    });
  }, [effectiveLifecycleKey, resetOwnerState, subject]);

  const commandIsCurrent = useCallback(
    () => snapshotRef.current.lifecycleKey === effectiveLifecycleKey,
    [effectiveLifecycleKey],
  );

  const open = useCallback(() => {
    if (!commandIsCurrent()) {
      return;
    }
    dispatch({
      lifecycleKey: effectiveLifecycleKey,
      subject: currentSubjectRef.current,
      type: "open",
    });
  }, [commandIsCurrent, effectiveLifecycleKey]);
  const close = useCallback(
    ({ restoreFocus = false }: { readonly restoreFocus?: boolean } = {}) => {
      if (!commandIsCurrent()) {
        return;
      }
      resetOwnerState("close");
      dispatch({ lifecycleKey: effectiveLifecycleKey, type: "close" });
      if (restoreFocus) {
        actionPortsRef.current.restoreFocus();
      }
    },
    [commandIsCurrent, effectiveLifecycleKey, resetOwnerState],
  );
  const setOpen = useCallback(
    (next: SetStateAction<boolean>) => {
      const isOpen =
        typeof next === "function"
          ? next(workbookInspectorStateIsOpen(snapshotRef.current))
          : next;
      if (isOpen) {
        open();
      } else {
        close();
      }
    },
    [close, open],
  );
  const invalidate = useCallback(
    (reason: WorkbookInspectorInvalidationReason) => {
      if (!commandIsCurrent()) {
        return;
      }
      resetOwnerState(reason);
      dispatch({
        lifecycleKey: effectiveLifecycleKey,
        type: "invalidate",
        reason,
      });
    },
    [commandIsCurrent, effectiveLifecycleKey, resetOwnerState],
  );
  const completeAction = useCallback(() => {
    if (!commandIsCurrent()) {
      return;
    }
    invalidate("action_completed");
    actionPortsRef.current.restoreFocus();
  }, [commandIsCurrent, invalidate]);

  return {
    commands: {
      close,
      completeAction,
      invalidate,
      open,
      setOpen,
    },
    snapshot,
  };
}
