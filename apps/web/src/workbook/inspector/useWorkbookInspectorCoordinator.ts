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
  type WorkbookInspectorSubject,
  workbookInspectorReducer,
  workbookInspectorSubjectsEqual,
} from "../models/workbookInspectorModel";

export type WorkbookInspectorOwnerActionPorts = {
  readonly resetOwnerState: (event: WorkbookInspectorResetEvent) => void;
  readonly restoreFocus: () => void;
};

export type WorkbookInspectorResetCause =
  | "close"
  | "retarget"
  | "action_completed"
  | "surface_changed"
  | "authorization_lost"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "hard_refresh";

export type WorkbookInspectorResetScope = "row_local" | "surface";

export type WorkbookInspectorResetEvent = {
  readonly cause: WorkbookInspectorResetCause;
  readonly scope: WorkbookInspectorResetScope;
};

export function workbookInspectorResetScope(
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
  const [snapshot, dispatch] = useReducer(
    workbookInspectorReducer,
    undefined,
    initialWorkbookInspectorState,
  );
  const actionPortsRef = useRef(actionPorts);
  const configRef = useRef(config);
  const lifecycleKeyRef = useRef(lifecycleKey);
  const blockedSubjectRef = useRef<WorkbookInspectorSubject | null>(null);
  const subjectRef = useRef<WorkbookInspectorSubject | null>(null);
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
    if (configRef.current.viewSchemaId === config.viewSchemaId) {
      return;
    }
    configRef.current = config;
    subjectRef.current = null;
    blockedSubjectRef.current = subject;
    resetOwnerState("surface_changed");
    dispatch({ type: "invalidate", reason: "surface_changed" });
  }, [config, resetOwnerState, subject]);

  useLayoutEffect(() => {
    if (lifecycleKeyRef.current === lifecycleKey) {
      return;
    }
    lifecycleKeyRef.current = lifecycleKey;
    subjectRef.current = null;
    blockedSubjectRef.current = subject;
    resetOwnerState("surface_changed");
    dispatch({ type: "invalidate", reason: "surface_changed" });
  }, [lifecycleKey, resetOwnerState, subject]);

  useLayoutEffect(() => {
    if (
      blockedSubjectRef.current !== null &&
      workbookInspectorSubjectsEqual(blockedSubjectRef.current, subject)
    ) {
      return;
    }
    blockedSubjectRef.current = null;
    if (workbookInspectorSubjectsEqual(subjectRef.current, subject)) {
      return;
    }
    subjectRef.current = subject;
    resetOwnerState("retarget");
    dispatch({ type: "retarget", subject });
  }, [resetOwnerState, subject]);

  const open = useCallback(() => {
    dispatch({ type: "open" });
  }, []);
  const close = useCallback(
    ({ restoreFocus = false }: { readonly restoreFocus?: boolean } = {}) => {
      resetOwnerState("close");
      dispatch({ type: "close" });
      if (restoreFocus) {
        actionPortsRef.current.restoreFocus();
      }
    },
    [resetOwnerState],
  );
  const setOpen = useCallback(
    (next: SetStateAction<boolean>) => {
      const isOpen =
        typeof next === "function" ? next(snapshotRef.current.isOpen) : next;
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
      resetOwnerState(reason);
      if (reason !== "action_completed") {
        subjectRef.current = null;
      }
      dispatch({ type: "invalidate", reason });
    },
    [resetOwnerState],
  );
  const completeAction = useCallback(() => {
    invalidate("action_completed");
    actionPortsRef.current.restoreFocus();
  }, [invalidate]);

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
