import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useReducer,
  useRef,
} from "react";
import type {
  RecordLifecycleAccepted,
  RecordRouteCommandPort,
} from "../mutations/workbookMutationCommandPorts";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorMessageFeedback,
} from "./workbookInspectorErrorModel";
import {
  buildRecordRollbackTargetFromHistoryAction,
  initialWorkbookRecordHistoryState,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryPendingAction,
  type WorkbookRecordHistorySubject,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";

type WorkbookRecordHistoryOwnerEffects = {
  readonly deleteAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
  readonly restoreAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
  readonly rollbackAccepted: (
    accepted: RecordLifecycleAccepted,
  ) => Promise<void> | void;
};

export function useWorkbookRecordHistoryController({
  canMutate,
  commands,
  ownerEffects,
  subject,
}: {
  readonly canMutate: boolean;
  readonly commands: RecordRouteCommandPort;
  readonly ownerEffects: WorkbookRecordHistoryOwnerEffects;
  readonly subject: WorkbookRecordHistorySubject | null;
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

  const load = useCallback(
    async (activeSubject: WorkbookRecordHistorySubject) => {
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
          requestId,
          subject: activeSubject,
          type: "load_rejected",
        });
        return;
      }
      dispatch({
        data: outcome.value,
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
    const pending = snapshotRef.current.pendingAction;
    if (pending === null || !canMutate) {
      dispatch({ type: "cancel" });
      return;
    }
    const operationId = workbookRecordHistoryOperationId(
      operationSequenceRef.current + 1,
    );
    operationSequenceRef.current = operationId.value;
    dispatch({ operationId, type: "submit" });
    const outcome =
      pending.kind === "rollback"
        ? await commands.rollback({
            baseRowVersion: pending.rowVersion,
            reason: `Rollback ${pending.action} from the workbook inspector`,
            recordId: pending.recordId,
            target: pending.target,
          })
        : await commands.execute({
            action: pending.operation,
            baseRowVersion: pending.rowVersion,
            reason:
              pending.operation === "delete"
                ? "Deleted from the workbook inspector"
                : "Restored from the workbook inspector",
            recordId: pending.recordId,
          });
    if (outcome.kind === "rejected") {
      dispatch({
        error: workbookInspectorErrorPresentation(outcome.failure),
        operationId,
        type: "operation_rejected",
      });
      return;
    }
    const accepted = outcome.value;
    const nextSubject: WorkbookRecordHistorySubject = {
      kind:
        pending.kind === "destructive"
          ? pending.operation === "delete"
            ? "deleted"
            : "live"
          : (snapshotRef.current.subject?.kind ?? "live"),
      recordId: accepted.recordId,
      rowVersion: accepted.rowVersion,
    };
    if (pending.kind === "rollback") {
      await ownerEffectsRef.current.rollbackAccepted(accepted);
    } else if (pending.operation === "delete") {
      await ownerEffectsRef.current.deleteAccepted(accepted);
    } else {
      await ownerEffectsRef.current.restoreAccepted(accepted);
    }
    const operationIsCurrent =
      snapshotRef.current.phase === "submitting" &&
      snapshotRef.current.operationId?.value === operationId.value &&
      snapshotRef.current.subject?.recordId === pending.recordId;
    dispatch({
      feedback: workbookInspectorMessageFeedback(
        pending.kind === "rollback"
          ? `Rolled back record ${accepted.recordId}.`
          : pending.operation === "delete"
            ? `Deleted record ${accepted.recordId}.`
            : `Restored record ${accepted.recordId}.`,
        "polite",
      ),
      operationId,
      subject: nextSubject,
      type: "operation_accepted",
    });
    if (operationIsCurrent) await load(nextSubject);
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
