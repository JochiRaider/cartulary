import type {
  InspectorConfig,
  InspectorPanelId,
} from "@cartulary/view-contracts";
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
  readonly clearLocalForm?: (() => void) | undefined;
  readonly clearLifecycleState?: (() => void) | undefined;
  readonly clearMergePlan?: (() => void) | undefined;
  readonly clearPendingConfirmation?: (() => void) | undefined;
  readonly clearPreview?: (() => void) | undefined;
  readonly clearSelection?: (() => void) | undefined;
  readonly clearWorkflowForm?: (() => void) | undefined;
  readonly restoreFocus?: (() => void) | undefined;
};

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
    config,
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

  const clearFeatureState = useCallback(() => {
    const ports = actionPortsRef.current;
    ports.clearPendingConfirmation?.();
    ports.clearPreview?.();
    ports.clearMergePlan?.();
    ports.clearWorkflowForm?.();
    ports.clearLocalForm?.();
  }, []);

  useLayoutEffect(() => {
    if (configRef.current.viewSchemaId === config.viewSchemaId) {
      return;
    }
    configRef.current = config;
    subjectRef.current = null;
    blockedSubjectRef.current = subject;
    clearFeatureState();
    actionPortsRef.current.clearLifecycleState?.();
    actionPortsRef.current.clearSelection?.();
    dispatch({ type: "reset_config", config });
  }, [clearFeatureState, config, subject]);

  useLayoutEffect(() => {
    if (lifecycleKeyRef.current === lifecycleKey) {
      return;
    }
    lifecycleKeyRef.current = lifecycleKey;
    subjectRef.current = null;
    blockedSubjectRef.current = subject;
    clearFeatureState();
    actionPortsRef.current.clearLifecycleState?.();
    actionPortsRef.current.clearSelection?.();
    dispatch({ type: "invalidate", reason: "surface_changed" });
  }, [clearFeatureState, lifecycleKey, subject]);

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
    clearFeatureState();
    dispatch({ type: "retarget", subject });
  }, [clearFeatureState, subject]);

  const open = useCallback((panelId?: InspectorPanelId) => {
    dispatch({ type: "open", panelId });
  }, []);
  const close = useCallback(
    ({ restoreFocus = false }: { readonly restoreFocus?: boolean } = {}) => {
      clearFeatureState();
      dispatch({ type: "close" });
      if (restoreFocus) {
        actionPortsRef.current.restoreFocus?.();
      }
    },
    [clearFeatureState],
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
  const selectPanel = useCallback((panelId: InspectorPanelId) => {
    dispatch({ type: "select_panel", panelId });
  }, []);
  const invalidate = useCallback(
    (reason: WorkbookInspectorInvalidationReason) => {
      clearFeatureState();
      if (reason !== "action_completed") {
        subjectRef.current = null;
        actionPortsRef.current.clearLifecycleState?.();
        actionPortsRef.current.clearSelection?.();
      }
      dispatch({ type: "invalidate", reason });
    },
    [clearFeatureState],
  );
  const completeAction = useCallback(() => {
    invalidate("action_completed");
    actionPortsRef.current.restoreFocus?.();
  }, [invalidate]);

  return {
    commands: {
      close,
      completeAction,
      invalidate,
      open,
      selectPanel,
      setOpen,
    },
    snapshot,
  };
}
