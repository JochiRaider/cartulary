import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useReducer,
  useRef,
  useSyncExternalStore,
} from "react";
import { observeAsyncOperation } from "../../services/asyncObservation";
import type { HistoryActionLookup } from "../history/HistoryActionLookup";
import { useWorkbookHistoryRuntime } from "../history/WorkbookHistoryContext";
import type { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import {
  acceptHistoryPage,
  beginHistoryRead,
  type HistoryReadKind,
  initialHistoryBrowsing,
  rejectHistoryRead,
} from "../history/workbookHistoryBrowsing";
import type { HistoryBinding } from "../history/workbookHistoryOperation";
import {
  type HistoryPage,
  sameHistoryReadScope,
} from "../history/workbookHistoryPage";
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
  const readAbort = useRef<AbortController | null>(null);
  const previewAbort = useRef<AbortController | null>(null);
  const previewReview = useRef<{
    lookup: HistoryActionLookup;
    pending: WorkbookRecordHistoryPendingAction;
    token: number;
  } | null>(null);
  const scope = owner?.readScope ?? null;
  const scopeKey = scope
    ? `${scope.actorId}:${scope.incidentId}:${scope.sessionIdentity}:${scope.epoch}`
    : "unavailable";
  const priorScopeKey = useRef(scopeKey);
  const priorScope = useRef(scope);
  const previewPending = useRef<WorkbookRecordHistoryPendingAction | null>(
    null,
  );

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
      readAbort.current?.abort();
      previewAbort.current?.abort();
      previewReview.current?.lookup.cancel();
    };
  }, []);
  useLayoutEffect(() => {
    void targetIdentity;
    generation.current += 1;
    reloadRetarget.current =
      snapshotRef.current.phase !== "idle" &&
      snapshotRef.current.subject?.recordId !==
        targetSubjectRef.current?.recordId;
    if (
      snapshotRef.current.subject?.recordId !==
      targetSubjectRef.current?.recordId
    )
      readAbort.current?.abort();
    previewAbort.current?.abort();
    previewReview.current?.lookup.cancel();
    previewReview.current = null;
    dispatchHistory({ subject: targetSubjectRef.current, type: "retarget" });
  }, [dispatchHistory, targetIdentity]);
  useEffect(() => {
    if (!canMutate) dispatchHistory({ type: "cancel" });
  }, [canMutate, dispatchHistory]);

  const load = useCallback(
    async (
      activeSubject: WorkbookInspectorSubject,
      feedback?: ReturnType<typeof workbookInspectorMessageFeedback>,
      requestedKind?: HistoryReadKind,
      retry = false,
    ) => {
      const scope = owner?.readScope;
      if (!owner || !scope || !mounted.current) return null;
      const existing = snapshotRef.current.browsing;
      const browsing =
        existing &&
        existing.recordId === activeSubject.recordId &&
        sameHistoryReadScope(existing.scope, scope)
          ? existing
          : initialHistoryBrowsing(
              scope,
              activeSubject.recordId,
              activeSubject.viewSchemaId,
            );
      const kind = requestedKind ?? (browsing.accepted ? "refresh" : "initial");
      const requested = beginHistoryRead(browsing, kind, retry);
      if (requested === browsing || !requested.pending) return null;
      readAbort.current?.abort();
      const controller = new AbortController();
      readAbort.current = controller;
      const request = requested.pending;
      dispatchHistory({ type: "browsing_changed", browsing: requested });
      const outcome = await owner.load(
        activeSubject.recordId,
        controller.signal,
        request.request,
      );
      if (
        !mounted.current ||
        controller.signal.aborted ||
        !sameHistoryReadScope(scope, owner.readScope)
      )
        return null;
      const current = snapshotRef.current.browsing;
      if (!current || current.pending !== request) return null;
      const subject = snapshotRef.current.subject;
      const next =
        outcome.kind === "accepted"
          ? acceptHistoryPage(current, request, outcome.value, {
              rowVersion: Math.max(
                subject?.rowVersion ?? 0,
                owner.latestVersion(activeSubject.recordId) ?? 0,
              ),
              deleted: subject?.kind === "deleted",
            })
          : rejectHistoryRead(current, request, outcome.failure);
      void feedback;
      return dispatchHistory({ type: "browsing_changed", browsing: next });
    },
    [owner, dispatchHistory],
  );
  useLayoutEffect(() => {
    if (!owner || !subject) return;
    return owner.registerRecordPresentation(subject.recordId, async () => {
      const current = targetSubjectRef.current;
      if (
        !mounted.current ||
        !activeRef.current ||
        !current ||
        current.recordId !== subject.recordId ||
        snapshotRef.current.phase === "idle"
      )
        return;
      previewAbort.current?.abort();
      previewReview.current?.lookup.cancel();
      previewReview.current = null;
      dispatchHistory({ type: "cancel" });
      const refreshed = await load(current);
      if (
        !mounted.current ||
        !activeRef.current ||
        targetSubjectRef.current?.recordId !== current.recordId
      )
        return;
      if (
        !refreshed?.browsing?.accepted ||
        refreshed.browsing.failure ||
        refreshed.browsing.accepted.data.row_version <
          (owner.latestVersion(current.recordId) ?? 0)
      )
        throw new Error("Decision history refresh remains incomplete");
    });
  }, [owner, subject, load, dispatchHistory]);
  useLayoutEffect(() => {
    if (priorScopeKey.current === scopeKey) return;
    const wasOpen = snapshotRef.current.phase !== "idle";
    priorScopeKey.current = scopeKey;
    const nextScope = owner?.readScope ?? null;
    const previousScope = priorScope.current;
    priorScope.current = nextScope;
    const sameSession =
      nextScope &&
      previousScope &&
      nextScope.actorId === previousScope.actorId &&
      nextScope.incidentId === previousScope.incidentId &&
      nextScope.sessionIdentity === previousScope.sessionIdentity;
    generation.current += 1;
    bindingGeneration.current += 1;
    readAbort.current?.abort();
    previewAbort.current?.abort();
    previewReview.current?.lookup.cancel();
    previewReview.current = null;
    const retained = snapshotRef.current.browsing;
    if (sameSession && retained?.accepted && nextScope) {
      dispatchHistory({ type: "cancel" });
      dispatchHistory({
        type: "browsing_changed",
        browsing: { ...retained, scope: nextScope, pending: null },
      });
      return;
    }
    dispatchHistory({ type: "clear" });
    dispatchHistory({ type: "retarget", subject: targetSubjectRef.current });
    if (wasOpen && owner?.readable && targetSubjectRef.current)
      void load(targetSubjectRef.current);
  }, [scopeKey, dispatchHistory, load, owner]);
  useEffect(() => {
    void targetIdentity;
    if (reloadRetarget.current && snapshotRef.current.subject) {
      reloadRetarget.current = false;
      void load(snapshotRef.current.subject);
    }
  }, [load, targetIdentity]);
  useLayoutEffect(() => {
    if (presentationActive) return;
    generation.current += 1;
    readAbort.current?.abort();
    previewAbort.current?.abort();
    previewReview.current?.lookup.cancel();
    previewReview.current = null;
    dispatchHistory({ type: "cancel" });
    const browsing = snapshotRef.current.browsing;
    if (browsing?.pending)
      dispatchHistory({
        type: "browsing_changed",
        browsing: rejectHistoryRead(browsing, browsing.pending, {
          kind: "retryable",
          message: "History reading stopped. Retry to continue.",
        }),
      });
  }, [presentationActive, dispatchHistory]);
  const open = useCallback(() => {
    const active = snapshotRef.current.subject ?? targetSubjectRef.current;
    if (active) {
      generation.current += 1;
      previewAbort.current?.abort();
      previewReview.current?.lookup.cancel();
      previewReview.current = null;
      dispatchHistory({ type: "cancel" });
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
      if (snapshotRef.current.subject?.recordId !== active.recordId) return;
      dispatchHistory({
        type: "lookup_changed",
        lookup: {
          phase: "failed",
          pagesChecked: 0,
          page: null,
          provenance: undefined,
          failure: { kind: "stale_target", message },
        },
      });
    },
    [dispatchHistory],
  );
  const runPreview = useCallback(
    async (restart = false) => {
      const review = previewReview.current;
      if (!review || !owner) return;
      if (restart) review.lookup.restart();
      const result = review.lookup.run(previewAbort.current?.signal);
      dispatchHistory({
        type: "lookup_changed",
        lookup: review.lookup.snapshot,
      });
      const checked = await result;
      if (
        !mounted.current ||
        generation.current !== review.token ||
        previewReview.current !== review
      )
        return;
      dispatchHistory({ type: "lookup_changed", lookup: checked });
      const active = snapshotRef.current.subject;
      if (
        checked.phase !== "matched" ||
        !checked.page ||
        !active ||
        active.recordId !== review.pending.recordId
      )
        return;
      if (
        (owner.latestVersion(active.recordId) ?? 0) > checked.page.row_version
      ) {
        rejectReview(active, "The record changed. Review this action again.");
        return;
      }
      if (
        !snapshotRef.current.browsing?.accepted &&
        !checked.provenance?.request.cursorToken &&
        owner.readScope
      ) {
        const initial = beginHistoryRead(
          initialHistoryBrowsing(
            owner.readScope,
            active.recordId,
            active.viewSchemaId,
          ),
          "initial",
        );
        if (!initial.pending) return;
        dispatchHistory({
          type: "browsing_changed",
          browsing: acceptHistoryPage(initial, initial.pending, checked.page, {
            rowVersion: active.rowVersion,
            deleted: active.kind === "deleted",
          }),
        });
      }
      dispatchHistory({ type: "review_accepted", data: checked.page });
      const pending = {
        ...review.pending,
        rowVersion: checked.page.row_version,
      };
      if (
        canMutate &&
        owner.permitted(
          pending.kind === "rollback" ? "rollback" : pending.operation,
        )
      )
        dispatchHistory({ type: "preview", pendingAction: pending });
    },
    [owner, dispatchHistory, canMutate, rejectReview],
  );
  const preview = useCallback(
    async (
      build: (
        active: WorkbookInspectorSubject,
      ) => WorkbookRecordHistoryPendingAction | null,
    ) => {
      const active = snapshotRef.current.subject;
      if (!active || !canMutate || !owner) return;
      const pending = build(active);
      if (!pending) return;
      previewPending.current = pending;
      previewAbort.current?.abort();
      previewReview.current?.lookup.cancel();
      const controller = new AbortController();
      previewAbort.current = controller;
      const token = ++generation.current;
      dispatchHistory({ type: "cancel" });
      dispatchHistory({
        type: "lookup_changed",
        lookup: {
          phase: "checking",
          pagesChecked: 0,
          page: null,
          provenance: undefined,
          failure: null,
        },
      });
      const settling = observeAsyncOperation((signal) =>
        settle(active, signal),
      );
      const cancel = () => settling.cancel();
      controller.signal.addEventListener("abort", cancel, { once: true });
      const result = await settling.result;
      controller.signal.removeEventListener("abort", cancel);
      if (
        !mounted.current ||
        generation.current !== token ||
        controller.signal.aborted
      )
        return;
      if (result.kind !== "completed" || result.value === null) {
        rejectReview(
          active,
          "Earlier row changes must settle before reviewing this action.",
        );
        return;
      }
      const provenance =
        pending.kind === "rollback"
          ? snapshotRef.current.browsing?.accepted?.provenance.get(
              pending.historyItemRef,
            )
          : undefined;
      const lookup = owner.createLookup({
        subject: active,
        pending,
        ...(provenance ? { provenance } : {}),
      });
      if (!lookup) return;
      previewReview.current = { lookup, pending, token };
      await runPreview();
    },
    [canMutate, owner, settle, dispatchHistory, rejectReview, runPreview],
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
    previewAbort.current?.abort();
    previewReview.current?.lookup.cancel();
    previewReview.current = null;
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
    const provenance = previewReview.current?.lookup.snapshot.provenance;
    const intent = {
      subject: active,
      pending,
      ...(provenance ? { provenance } : {}),
    };
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
          const scope = owner.readScope;
          if (scope && "paging" in history) {
            const initial = beginHistoryRead(
              initialHistoryBrowsing(
                scope,
                nextSubject.recordId,
                nextSubject.viewSchemaId,
              ),
              "initial",
            );
            if (!initial.pending) return;
            dispatchHistory({
              type: "browsing_changed",
              browsing: acceptHistoryPage(
                initial,
                initial.pending,
                history as HistoryPage,
                {
                  rowVersion: nextSubject.rowVersion,
                  deleted: nextSubject.kind === "deleted",
                },
              ),
            });
          }
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
          "This action could not be admitted. Review current history or open Recovery for recovery.",
        ),
      });
      return;
    }
    await owner.execute(attempt);
    const entry = owner
      .getSnapshot()
      .find((entry) => entry.attempt.id === attempt.id);
    if (
      isCurrent() &&
      entry &&
      entry.phase !== "acknowledged" &&
      entry.phase !== "preparing"
    )
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
                "The outcome is unknown. Open Recovery to recover this action.",
              ),
      });
  }, [owner, canMutate, dispatchHistory, runtime]);
  useEffect(() => {
    const state = snapshotRef.current;
    if (state.phase !== "submitting") return;
    const entry = operations.find(
      (entry) => entry.attempt.subject.recordId === state.subject.recordId,
    );
    if (
      !entry ||
      entry.phase === "preparing" ||
      entry.phase === "submitting" ||
      entry.phase === "acknowledged"
    )
      return;
    dispatchHistory({
      type: "operation_rejected",
      operationId: state.operationId,
      feedback: entry.failure
        ? {
            kind: "error",
            error: workbookInspectorErrorPresentation(entry.failure),
          }
        : workbookInspectorLocalErrorFeedback(
            "The outcome is unknown. Open Recovery to recover this action.",
          ),
    });
  }, [operations, dispatchHistory]);
  return {
    commands: {
      cancel,
      clearFeedback: () => dispatchHistory({ type: "feedback_cleared" }),
      confirm,
      open,
      loadOlder: () => {
        const active = snapshotRef.current.subject;
        if (active) void load(active, undefined, "continuation");
      },
      retryRead: () => {
        const state = snapshotRef.current;
        const failure = state.browsing?.failure;
        if (state.subject && failure)
          void load(state.subject, undefined, failure.request.kind, true);
      },
      continuePreview: () => {
        const pending = previewPending.current;
        if (previewReview.current) void runPreview();
        else if (pending) void preview(() => pending);
      },
      restartPreview: () => void runPreview(true),
      previewDeleteRestore,
      previewRollback,
      load,
    },
    snapshot,
    operations,
  };
}
