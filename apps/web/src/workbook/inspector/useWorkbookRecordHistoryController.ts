import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useReducer,
  useRef,
  useSyncExternalStore,
} from "react";
import { observeAsyncOperation } from "../../services/asyncObservation";
import { useWorkbookHistoryRuntime } from "../history/WorkbookHistoryContext";
import type { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import {
  type HistoryBinding,
  historyTargetEqual,
} from "../history/workbookHistoryOperation";
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
  type RecordHistoryData,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryPendingAction,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";
import { workbookRecordHistoryCompletionFeedback } from "./workbookRecordHistoryOperation";
import type { WorkbookRecordHistoryOwnerEffects } from "./workbookRecordHistoryOwnerEffects";

const empty = [] as const;
const emptySnapshot = () => empty;
const emptySubscription = () => () => {};
export function useWorkbookRecordHistoryController({
  canMutate,
  ownerEffects,
  subject,
  owner: providedOwner,
  coordinate,
  presentation,
  presentationActive = true,
  initialHistory,
}: {
  readonly initialHistory?: RecordHistoryData;
  readonly presentationActive?: boolean;
  readonly beginMutation?: () => () => void;
  readonly commands?: RecordRouteCommandPort;
  readonly canMutate: boolean;
  readonly ownerEffects: WorkbookRecordHistoryOwnerEffects;
  readonly subject: WorkbookInspectorSubject | null;
  readonly owner?: WorkbookRecordHistoryOwner;
  readonly coordinate?: HistoryBinding["coordinate"];
  readonly presentation?: {
    readonly snapshot: WorkbookRecordHistoryState;
    readonly dispatch: (
      event: WorkbookRecordHistoryEvent,
    ) => WorkbookRecordHistoryState;
  };
}) {
  const runtime = useWorkbookHistoryRuntime();
  const owner = providedOwner ?? runtime?.history;
  const operations = useSyncExternalStore(
    owner?.subscribe ?? emptySubscription,
    owner?.getSnapshot ?? emptySnapshot,
  );
  const [localSnapshot, reactDispatch] = useReducer(
    workbookRecordHistoryReducer,
    subject,
    (activeSubject): WorkbookRecordHistoryState =>
      initialHistory && activeSubject
        ? {
            phase: "ready",
            subject: activeSubject,
            result: { kind: "loaded", data: initialHistory },
          }
        : initialWorkbookRecordHistoryState(activeSubject),
  );
  const snapshot = presentation?.snapshot ?? localSnapshot;
  const snapshotRef = useRef(snapshot);
  const presentationRef = useRef(presentation);
  const effectsRef = useRef(ownerEffects);
  const coordinateRef = useRef(coordinate);
  const targetSubjectRef = useRef(subject);
  targetSubjectRef.current = subject;
  const targetIdentity =
    subject === null
      ? "none"
      : `${subject.viewSchemaId}:${subject.recordId}:${subject.rowVersion}:${subject.kind}`;
  const mounted = useRef(true);
  const activeRef = useRef(presentationActive);
  activeRef.current = presentationActive;
  const bindingGeneration = useRef(0);
  const bindingIdentity = `${subject?.viewSchemaId}:${subject?.recordId}:${presentationActive}`;
  const generation = useRef(0);
  useLayoutEffect(() => {
    void bindingIdentity;
    bindingGeneration.current += 1;
  }, [bindingIdentity]);
  const reloadRetarget = useRef(false);
  const requestSequence = useRef(0);
  const operationSequence = useRef(0);
  const dispatchHistory = useCallback((event: WorkbookRecordHistoryEvent) => {
    const next = presentationRef.current
      ? presentationRef.current.dispatch(event)
      : workbookRecordHistoryReducer(snapshotRef.current, event);
    snapshotRef.current = next;
    if (!presentationRef.current) reactDispatch(event);
    return next;
  }, []);
  useLayoutEffect(() => {
    snapshotRef.current = snapshot;
    presentationRef.current = presentation;
    effectsRef.current = ownerEffects;
    coordinateRef.current = coordinate;
  });
  useLayoutEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      generation.current += 1;
    };
  }, []);
  useEffect(() => {
    void targetIdentity;
    generation.current += 1;
    reloadRetarget.current =
      presentationRef.current !== undefined &&
      snapshotRef.current.phase !== "idle" &&
      snapshotRef.current.subject?.recordId !==
        targetSubjectRef.current?.recordId;
    dispatchHistory({ subject: targetSubjectRef.current, type: "retarget" });
  }, [dispatchHistory, targetIdentity]);
  useEffect(() => {
    if (!canMutate) dispatchHistory({ type: "cancel" });
  }, [canMutate, dispatchHistory]);

  const load = useCallback(
    async (
      activeSubject: WorkbookInspectorSubject,
      feedback?: ReturnType<typeof workbookInspectorMessageFeedback>,
    ) => {
      if (!owner || !mounted.current) return null;
      const requestId = workbookRecordHistoryRequestId(
        ++requestSequence.current,
      );
      dispatchHistory({
        requestId,
        subject: activeSubject,
        type: "load_requested",
      });
      const outcome = await owner.load(activeSubject.recordId);
      if (!mounted.current) return null;
      if (outcome.kind === "rejected") {
        dispatchHistory({
          error: workbookInspectorErrorPresentation(outcome.failure),
          requestId,
          subject: activeSubject,
          type: "load_rejected",
        });
        return null;
      }
      const next = dispatchHistory({
        data: outcome.value,
        feedback,
        requestId,
        subject: activeSubject,
        type: "load_accepted",
      });
      return next.phase === "ready" ? next : null;
    },
    [owner, dispatchHistory],
  );
  useEffect(() => {
    void targetIdentity;
    if (reloadRetarget.current && snapshotRef.current.subject) {
      reloadRetarget.current = false;
      void load(snapshotRef.current.subject);
    }
  }, [load, targetIdentity]);
  const open = useCallback(() => {
    const active = snapshotRef.current.subject ?? targetSubjectRef.current;
    if (active) {
      dispatchHistory({ type: "retarget", subject: active });
      void load(active);
    }
  }, [load, dispatchHistory]);
  const settle = useCallback(
    async (active: WorkbookInspectorSubject, signal: AbortSignal) => {
      owner?.acceptVersion(active.recordId, active.rowVersion);
      return coordinateRef.current
        ? coordinateRef.current(active.recordId, signal)
        : runtime
          ? runtime.coordinateHistory(active.recordId, signal)
          : active.rowVersion;
    },
    [owner, runtime],
  );
  const rejectReview = useCallback(
    (active: WorkbookInspectorSubject, message: string) => {
      const requestId = workbookRecordHistoryRequestId(
        ++requestSequence.current,
      );
      dispatchHistory({ type: "load_requested", subject: active, requestId });
      dispatchHistory({
        type: "load_rejected",
        subject: active,
        requestId,
        error: { primaryMessage: message, technicalFields: [] },
      });
    },
    [dispatchHistory],
  );
  const preview = useCallback(
    async (
      build: (
        active: WorkbookInspectorSubject,
      ) => WorkbookRecordHistoryPendingAction | null,
    ) => {
      const active = snapshotRef.current.subject;
      if (!active || !canMutate || !owner) return;
      const token = ++generation.current;
      const result = await observeAsyncOperation((signal) =>
        settle(active, signal),
      ).result;
      if (!mounted.current || generation.current !== token) return;
      if (result.kind !== "completed" || result.value === null) {
        rejectReview(
          active,
          "Earlier row changes must settle before reviewing this action.",
        );
        return;
      }
      const next = await load(active);
      if (!mounted.current || generation.current !== token || !next?.subject)
        return;
      const pending = build(next.subject);
      if (pending?.kind === "rollback") {
        const item = workbookRecordHistoryLoadedData(next)?.items.find(
          (item) => item.history_item_ref === pending.historyItemRef,
        );
        const advertised =
          item &&
          buildRecordRollbackTargetFromHistoryAction(item, pending.action);
        if (!advertised || !historyTargetEqual(advertised, pending.target)) {
          rejectReview(
            next.subject,
            "This action is no longer available. Review current history.",
          );
          return;
        }
      }
      if (
        pending &&
        owner.permitted(
          pending.kind === "rollback" ? "rollback" : pending.operation,
        )
      )
        dispatchHistory({ pendingAction: pending, type: "preview" });
    },
    [canMutate, owner, settle, load, dispatchHistory, rejectReview],
  );
  const previewDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      void preview((active) => ({
        kind: "destructive",
        operation,
        recordId: active.recordId,
        rowVersion: active.rowVersion,
      }));
    },
    [preview],
  );
  const previewRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      void preview((active) => {
        const target = buildRecordRollbackTargetFromHistoryAction(item, action);
        return target
          ? {
              action,
              historyItemRef: item.history_item_ref,
              kind: "rollback",
              recordId: active.recordId,
              rowVersion: active.rowVersion,
              target,
            }
          : null;
      });
    },
    [preview],
  );
  const cancel = useCallback(() => {
    generation.current += 1;
    dispatchHistory({ type: "cancel" });
  }, [dispatchHistory]);
  const confirm = useCallback(async () => {
    const pending = workbookRecordHistoryPendingAction(snapshotRef.current);
    const active = snapshotRef.current.subject;
    if (!owner || !pending || !active || !canMutate) return;
    const operationId = workbookRecordHistoryOperationId(
      ++operationSequence.current,
    );
    const next = dispatchHistory({ operationId, type: "submit" });
    if (next.phase !== "submitting" || next.operationId !== operationId) return;
    const effects = effectsRef.current;
    const admittedCoordinate = coordinateRef.current;
    const intent = { subject: active, pending };
    const admittedGeneration = bindingGeneration.current;
    const isCurrent = () =>
      bindingGeneration.current === admittedGeneration &&
      mounted.current &&
      activeRef.current &&
      snapshotRef.current.subject?.recordId === active.recordId &&
      snapshotRef.current.subject.viewSchemaId === active.viewSchemaId;
    const binding: HistoryBinding = {
      isCurrent,
      coordinate: (recordId, signal) =>
        admittedCoordinate
          ? admittedCoordinate(recordId, signal)
          : runtime
            ? runtime.coordinateHistory(recordId, signal)
            : Promise.resolve(active.rowVersion),
      acknowledged: (receipt) => {
        dispatchHistory({
          feedback: workbookRecordHistoryCompletionFeedback(intent),
          operationId,
          recordId: receipt.recordId,
          rowVersion: receipt.rowVersion,
          type: "operation_accepted",
        });
        if (receipt.kind === "delete") effects.deleteAccepted(receipt);
      },
      reconcile: async (receipt, current, history) => {
        const nextSubject = snapshotRef.current.subject;
        if (nextSubject && current()) {
          const requestId = workbookRecordHistoryRequestId(
            ++requestSequence.current,
          );
          dispatchHistory({
            requestId,
            subject: nextSubject,
            type: "load_requested",
          });
          dispatchHistory({
            data: history,
            feedback: workbookRecordHistoryCompletionFeedback(intent),
            requestId,
            subject: nextSubject,
            type: "load_accepted",
          });
        }
        await effects.refresh();
        if (
          !current() ||
          history.row_version > receipt.rowVersion ||
          (owner.latestVersion(receipt.recordId) ?? 0) > receipt.rowVersion
        )
          return;
        if (receipt.kind === "restore") effects.restoreAccepted(receipt);
        if (receipt.kind === "rollback") effects.rollbackAccepted(receipt);
      },
    };
    const attempt = owner.admit(intent, binding);
    if (!attempt) {
      dispatchHistory({
        operationId,
        type: "operation_rejected",
        feedback: workbookInspectorLocalErrorFeedback(
          "This action could not be admitted. Review current history or open History actions for recovery.",
        ),
      });
      return;
    }
    await owner.execute(attempt);
    const entry = owner
      .getSnapshot()
      .find((entry) => entry.attempt.id === attempt.id);
    if (isCurrent() && entry && entry.phase !== "acknowledged")
      dispatchHistory({
        operationId,
        type: "operation_rejected",
        feedback:
          entry.phase === "rejected" && entry.failure
            ? {
                kind: "error",
                error: workbookInspectorErrorPresentation(entry.failure),
              }
            : workbookInspectorLocalErrorFeedback(
                "The outcome is unknown. Open History actions to recover this action.",
              ),
      });
  }, [owner, canMutate, dispatchHistory, runtime]);
  return {
    commands: {
      cancel,
      clearFeedback: () => dispatchHistory({ type: "feedback_cleared" }),
      confirm,
      open,
      previewDeleteRestore,
      previewRollback,
      load,
    },
    snapshot,
    operations,
  };
}
