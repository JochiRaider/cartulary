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
  type WorkbookRecordHistoryPendingAction,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";
import {
  acceptedWorkbookRecordHistorySubject,
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
  const [snapshot, dispatch] = useReducer(
    workbookRecordHistoryReducer,
    subject,
    initialWorkbookRecordHistoryState,
  );
  const snapshotRef = useRef(snapshot);
  const requestSequenceRef = useRef(0);
  const operationSequenceRef = useRef(0);
  const ownerEffectsRef = useRef(ownerEffects);

  useLayoutEffect(() => {
    snapshotRef.current = snapshot;
    ownerEffectsRef.current = ownerEffects;
  });

  useEffect(() => {
    dispatch({ subject, type: "retarget" });
  }, [subject]);

  useEffect(() => {
    if (!canMutate) dispatch({ type: "cancel" });
  }, [canMutate]);

  const load = useCallback(
    async (
      activeSubject: WorkbookInspectorSubject,
      completionFeedback?: ReturnType<typeof workbookInspectorMessageFeedback>,
    ) => {
      const requestId = workbookRecordHistoryRequestId(
        requestSequenceRef.current + 1,
      );
      requestSequenceRef.current = requestId.value;
      dispatch({ requestId, subject: activeSubject, type: "load_requested" });
      const outcome = await commands.loadHistory({
        recordId: activeSubject.recordId,
      });
      if (outcome.kind === "rejected") {
        dispatch({
          error: workbookInspectorErrorPresentation(outcome.failure),
          feedback: completionFeedback,
          requestId,
          subject: activeSubject,
          type: "load_rejected",
        });
        return;
      }
      dispatch({
        data: outcome.value,
        feedback: completionFeedback,
        requestId,
        subject: activeSubject,
        type: "load_accepted",
      });
    },
    [commands],
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
      dispatch({ pendingAction, type: "preview" });
    },
    [canMutate],
  );

  const previewRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const activeSubject = snapshotRef.current.subject;
      if (activeSubject === null || !canMutate) return;
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) return;
      dispatch({
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
    [canMutate],
  );

  const cancel = useCallback(() => dispatch({ type: "cancel" }), []);

  const confirm = useCallback(async () => {
    const pending = workbookRecordHistoryPendingAction(snapshotRef.current);
    const operationSubject = snapshotRef.current.subject;
    if (pending === null || operationSubject === null || !canMutate) {
      dispatch({ type: "cancel" });
      return;
    }
    const operationId = workbookRecordHistoryOperationId(
      operationSequenceRef.current + 1,
    );
    operationSequenceRef.current = operationId.value;
    const capturedOwnerEffects = ownerEffectsRef.current;
    dispatch({ operationId, type: "submit" });
    const outcome = await executeWorkbookRecordHistoryOperation(
      commands,
      pending,
    );
    if (outcome.kind === "rejected") {
      dispatch({
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
    const nextSubject = acceptedWorkbookRecordHistorySubject(
      operationSubject,
      pending,
      accepted,
    );
    if (nextSubject === null) {
      dispatch({
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
    const operationIsCurrent =
      snapshotRef.current.phase === "submitting" &&
      snapshotRef.current.operationId.value === operationId.value &&
      snapshotRef.current.subject?.recordId === pending.recordId;
    const completionFeedback = workbookRecordHistoryCompletionFeedback(
      pending,
      accepted,
    );
    dispatch({
      feedback: completionFeedback,
      operationId,
      subject: nextSubject,
      type: "operation_accepted",
    });
    if (operationIsCurrent) await load(nextSubject, completionFeedback);
  }, [canMutate, commands, load]);

  return {
    commands: {
      cancel,
      clearFeedback: () => dispatch({ type: "feedback_cleared" }),
      confirm,
      open,
      previewDeleteRestore,
      previewRollback,
    },
    snapshot,
  };
}
