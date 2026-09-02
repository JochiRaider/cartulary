import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useReducer,
  useRef,
} from "react";
import type { RecordRouteCommandPort } from "../mutations/workbookMutationCommandPorts";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorFeedback,
  type workbookInspectorMessageFeedback,
} from "./workbookInspectorErrorModel";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";
import {
  buildRecordRollbackTargetFromHistoryAction,
  initialWorkbookRecordHistoryState,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryPendingAction,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";
import {
  applyWorkbookRecordHistoryOwnerEffect,
  executeWorkbookRecordHistoryOperation,
  workbookRecordHistoryCompletionFeedback,
} from "./workbookRecordHistoryOperation";
import type { WorkbookRecordHistoryOwnerEffects } from "./workbookRecordHistoryOwnerEffects";

export function useWorkbookRecordHistoryController({
  canMutate,
  commands,
  ownerEffects,
  subject,
}: {
  readonly canMutate: boolean;
  readonly commands: RecordRouteCommandPort;
  readonly ownerEffects: WorkbookRecordHistoryOwnerEffects;
  readonly subject: WorkbookInspectorSubject | null;
}) {
  const [snapshot, reactDispatch] = useReducer(
    workbookRecordHistoryReducer,
    subject,
    initialWorkbookRecordHistoryState,
  );
  const snapshotRef = useRef(snapshot);
  const requestSequenceRef = useRef(0);
  const operationSequenceRef = useRef(0);
  const ownerEffectsRef = useRef(ownerEffects);
  const dispatchHistory = useCallback(
    (event: WorkbookRecordHistoryEvent): WorkbookRecordHistoryState => {
      const next = workbookRecordHistoryReducer(snapshotRef.current, event);
      snapshotRef.current = next;
      reactDispatch(event);
      return next;
    },
    [],
  );

  useLayoutEffect(() => {
    snapshotRef.current = snapshot;
    ownerEffectsRef.current = ownerEffects;
  });

  useEffect(() => {
    dispatchHistory({ subject, type: "retarget" });
  }, [dispatchHistory, subject]);

  useEffect(() => {
    if (!canMutate) dispatchHistory({ type: "cancel" });
  }, [canMutate, dispatchHistory]);

  const load = useCallback(
    async (
      activeSubject: WorkbookInspectorSubject,
      completionFeedback?: ReturnType<typeof workbookInspectorMessageFeedback>,
    ) => {
      const requestId = workbookRecordHistoryRequestId(
        requestSequenceRef.current + 1,
      );
      requestSequenceRef.current = requestId.value;
      dispatchHistory({
        requestId,
        subject: activeSubject,
        type: "load_requested",
      });
      const outcome = await commands.loadHistory({
        recordId: activeSubject.recordId,
      });
      if (outcome.kind === "rejected") {
        dispatchHistory({
          error: workbookInspectorErrorPresentation(outcome.failure),
          feedback: completionFeedback,
          requestId,
          subject: activeSubject,
          type: "load_rejected",
        });
        return;
      }
      dispatchHistory({
        data: outcome.value,
        feedback: completionFeedback,
        requestId,
        subject: activeSubject,
        type: "load_accepted",
      });
    },
    [commands, dispatchHistory],
  );

  const open = useCallback(() => {
    const activeSubject = snapshotRef.current.subject;
    if (activeSubject !== null) void load(activeSubject);
  }, [load]);

  const previewDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const activeSubject = snapshotRef.current.subject;
      if (activeSubject === null || !canMutate) return;
      const pendingAction: WorkbookRecordHistoryPendingAction = {
        kind: "destructive",
        operation,
        recordId: activeSubject.recordId,
        rowVersion: activeSubject.rowVersion,
      };
      dispatchHistory({ pendingAction, type: "preview" });
    },
    [canMutate, dispatchHistory],
  );

  const previewRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const activeSubject = snapshotRef.current.subject;
      if (activeSubject === null || !canMutate) return;
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) return;
      dispatchHistory({
        pendingAction: {
          action,
          historyItemRef: item.history_item_ref,
          kind: "rollback",
          recordId: activeSubject.recordId,
          rowVersion: activeSubject.rowVersion,
          target,
        },
        type: "preview",
      });
    },
    [canMutate, dispatchHistory],
  );

  const cancel = useCallback(
    () => dispatchHistory({ type: "cancel" }),
    [dispatchHistory],
  );

  const confirm = useCallback(async () => {
    const pending = workbookRecordHistoryPendingAction(snapshotRef.current);
    const operationSubject = snapshotRef.current.subject;
    if (pending === null || operationSubject === null || !canMutate) {
      dispatchHistory({ type: "cancel" });
      return;
    }
    const operationId = workbookRecordHistoryOperationId(
      operationSequenceRef.current + 1,
    );
    operationSequenceRef.current = operationId.value;
    const capturedOwnerEffects = ownerEffectsRef.current;
    dispatchHistory({ operationId, type: "submit" });
    const outcome = await executeWorkbookRecordHistoryOperation(
      commands,
      pending,
    );
    if (outcome.kind === "rejected") {
      dispatchHistory({
        feedback: {
          error: workbookInspectorErrorPresentation(outcome.failure),
          kind: "error",
        },
        operationId,
        type: "operation_rejected",
      });
      return;
    }
    const accepted = outcome.value;
    if (
      accepted.recordId !== pending.recordId ||
      !Number.isInteger(accepted.rowVersion) ||
      accepted.rowVersion <= 0
    ) {
      dispatchHistory({
        feedback: workbookInspectorLocalErrorFeedback(
          "The history operation returned an invalid record identity.",
        ),
        operationId,
        type: "operation_rejected",
      });
      return;
    }
    await applyWorkbookRecordHistoryOwnerEffect(
      capturedOwnerEffects,
      pending,
      accepted,
    );
    const completionFeedback = workbookRecordHistoryCompletionFeedback(
      pending,
      accepted,
    );
    const next = dispatchHistory({
      feedback: completionFeedback,
      operationId,
      recordId: accepted.recordId,
      rowVersion: accepted.rowVersion,
      type: "operation_accepted",
    });
    if (
      next.phase === "idle" &&
      next.subject?.recordId === accepted.recordId &&
      next.subject.rowVersion === accepted.rowVersion
    ) {
      await load(next.subject, completionFeedback);
    }
  }, [canMutate, commands, dispatchHistory, load]);

  return {
    commands: {
      cancel,
      clearFeedback: () => dispatchHistory({ type: "feedback_cleared" }),
      confirm,
      open,
      previewDeleteRestore,
      previewRollback,
    },
    snapshot,
  };
}
