import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { SheetRef } from "../../../shared/sheetRef";
import type { TimelineMutationIdentityPort } from "../../mutations/workbookMutationCommandPorts";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationOwnerEnvelope } from "../../runtime/WorkbookMutationDriverRegistry";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type {
  WorkbookBatchAttempt,
  WorkbookBatchReceipt,
} from "../../runtime/workbookBatchOperation";
import { workbookPendingMutationFailureResult } from "../../runtime/workbookPendingMutationSettlement";
import {
  refreshBlocksWorkbookPendingUnit,
  type WorkbookPendingQueueRuntime,
} from "../../runtime/workbookPendingReplayRuntime";
import type {
  PendingReplaySettlement,
  PendingReplayUnitInput,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";
import type {
  TimelineReplayContext,
  TimelineRowStoreCommands,
} from "../models/timelineControllerPorts";
import {
  type FocusFieldKey,
  inputFocusKey,
  type RowValues,
  type TimelineScalarEditorSurface,
  timelineScalarBindingForField,
  timelineScalarBindings,
} from "../models/timelineFieldRegistry";
import {
  planTimelineAcceptedProjection,
  planTimelineDiscard,
  planTimelineRejectedSettlement,
  planTimelineReplayAdmission,
  type TimelinePendingReplayAdmission,
  type TimelineReplayAdmissionPlan,
} from "../models/timelineMutationDriverPlans";
import { buildFollowOnCapturePatch } from "../models/timelineMutationIntents";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "../models/timelineRowModel";

type TimelineMutableRef<T> = {
  current: T;
};

type TimelineConflictSettlement = Extract<
  PendingReplaySettlement,
  { readonly outcome: "same_field_conflict" }
>;

function isCollectionDraftKey(
  focusField: FocusFieldKey,
): focusField is "hostRefs" | "identityRefs" | "tags" {
  return (
    focusField === "hostRefs" ||
    focusField === "identityRefs" ||
    focusField === "tags"
  );
}

function gridOutcomeForFailure(
  failure: WorkbookOperationFailure,
): GridEditCommitOutcome {
  const message = failure.message;
  if (failure.kind === "validation") {
    return { kind: "validation_error", message };
  }
  if (failure.kind === "stale_target") {
    return { kind: "stale_target", message };
  }
  if (
    failure.kind === "same_field_conflict" ||
    failure.kind === "client_txn_conflict"
  ) {
    return { kind: "conflict", message };
  }
  return { kind: "rejected_mutation", message };
}

function currentTimelineReplayRow(
  unit: PendingReplayUnitState | null,
  rows: readonly WorkbookRow[],
  latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null,
): WorkbookRow | undefined {
  if (unit === null) return undefined;
  if (unit.recordId === null) {
    return rows.find((row) => row.key === unit.rowKey);
  }
  return (
    latestCommittedTimelineRow(unit.recordId) ??
    rows.find((row) => row.recordId === unit.recordId)
  );
}

export type TimelineMutationDriverPorts = {
  readonly acceptEditorPredecessor: (
    row: WorkbookRow,
    fields: readonly string[],
    previousValues?: RowValues,
  ) => void;
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    accepted: WorkbookPendingMutationAccepted,
    options?: {
      continueOnFreshDraft?: boolean;
      detectAutoResolution?: boolean;
      previousRow?: WorkbookRow;
      operationId?: string;
      promoteToCommittedRowInspect?: boolean;
      viewportContinuityToken?: number;
    },
  ) => WorkbookRow;
  readonly clearSubmittedScalarEditorDraftValuesForRow: (
    rowKey: string,
    submittedValues: RowValues,
    submittedCollections?: Partial<WorkbookRow["collectionDrafts"]>,
    revisions?: ReadonlyMap<string, number>,
  ) => void;
  readonly captureEditorDrafts: (
    rowKey: string,
    surface: TimelineScalarEditorSurface,
    fields?: ReadonlySet<string>,
  ) => ReadonlyMap<string, number>;
  readonly clearViewportContinuity: (token: number) => void;
  readonly conflictQueueRef: TimelineMutableRef<Record<string, unknown>>;
  readonly registerMutationConflict: (
    conflict: Extract<
      WorkbookOperationFailure,
      { readonly kind: "same_field_conflict" }
    >["conflict"],
    rowKey: string,
    focusField: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
    refresh: () => Promise<WorkbookOperationOutcome<unknown>>,
    originSheetRef: SheetRef,
    draftRevisions?: ReadonlyMap<string, number>,
  ) => boolean;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
  readonly mutationCommands: TimelineMutationIdentityPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly sheetRef: SheetRef;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly postMutationQueryRefreshRequired: boolean;
  readonly publishPendingQueueState: () => void;
  readonly reconcileDiscardedPendingUnit: (
    discardedUnit: PendingReplayUnitState,
    remainingUnits: readonly PendingReplayUnitState[],
    contextByUnitId: ReadonlyMap<string, TimelineReplayContext>,
  ) => void;
  readonly recordWorkbookTiming: (
    name: string,
    details?: Record<string, unknown>,
  ) => void;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly requestAuthorizationRecovery: () => void;
  readonly setRefreshError: (message: string | null) => void;
  readonly setMutationError: (message: string | null) => void;
  readonly rowStoreCommands: TimelineRowStoreCommands;
  readonly readCurrentRow?: (
    unit: PendingReplayUnitState,
  ) => Promise<WorkbookRow | null>;
  readonly beginDispatch?: () => () => void;
};

/** Owns queue admission, immutable dispatch and settlement independently of presentation. */
export function createTimelineMutationDriver(
  ports: TimelineMutationDriverPorts,
) {
  const {
    applyAcceptedRowMutation,
    clearSubmittedScalarEditorDraftValuesForRow,
    clearViewportContinuity: clearMountedViewportContinuity,
    conflictQueueRef,
    registerMutationConflict,
    latestCommittedTimelineRow,
    mutationCommands,
    mutationRuntime,
    pendingSavesRefs,
    publishPendingQueueState,
    reconcileDiscardedPendingUnit,
    recordWorkbookTiming,
    rowsRef,
    requestAuthorizationRecovery,
    setRefreshError,
    setMutationError,
    rowStoreCommands,
  } = ports;
  const clearViewportContinuity = (token: number | undefined) => {
    if (token !== undefined) clearMountedViewportContinuity(token);
  };
  const { replaceRows } = rowStoreCommands;
  const contextByUnitId = pendingSavesRefs.replayContextByUnitId;
  const completionCallbacksRef = {
    current: new Map<
      string,
      Array<{
        run: (outcome: GridEditCommitOutcome) => void;
        values: Partial<RowValues>;
        superseded: boolean;
      }>
    >(),
  };
  const settleCompletionCallbacks = (
    unitId: string,
    outcome: GridEditCommitOutcome,
  ) => {
    const callbacks = completionCallbacksRef.current.get(unitId) ?? [];
    completionCallbacksRef.current.delete(unitId);
    for (const callback of callbacks)
      callback.run(
        outcome.kind === "accepted" && callback.superseded
          ? {
              kind: "superseded",
              message:
                "A newer admitted value replaced this edit before dispatch.",
            }
          : outcome,
      );
  };
  const clearPendingSignatureForUnit = (unit: {
    readonly rowKey: string;
    readonly mutationSignature?: string;
  }) => {
    if (unit.mutationSignature === undefined) {
      return;
    }
    if (
      pendingSavesRefs.pendingSignaturesRef.current.get(unit.rowKey) ===
      unit.mutationSignature
    ) {
      pendingSavesRefs.pendingSignaturesRef.current.delete(unit.rowKey);
    }
    const nextRows = rowsRef.current.map((row) =>
      row.key === unit.rowKey && row.pendingSignature === unit.mutationSignature
        ? { ...row, pendingSignature: null }
        : row,
    );
    rowsRef.current = nextRows;
    replaceRows(nextRows);
  };

  const schedulePendingReplayRetry = () => {
    mutationRuntime.scheduleRetry(1000);
  };

  const requestPendingReplay = (reason: string) => {
    const pending = pendingSavesRefs.pendingQueueRef.current;
    const snapshot = pending.model.snapshot();
    const candidate = pending.model.peekNextQueued();
    const readyForImmediateDrain =
      !snapshot.authPaused &&
      snapshot.halted === null &&
      snapshot.sameFieldConflicts.length === 0 &&
      candidate !== null &&
      !refreshBlocksWorkbookPendingUnit(pending, candidate.unit) &&
      Object.keys(conflictQueueRef.current).length === 0 &&
      snapshot.inFlightCount === 0 &&
      snapshot.queuedCount > 0;
    if (readyForImmediateDrain) {
      recordWorkbookTiming("pending_replay_drain_immediate", { reason });
    }
    mutationRuntime.requestDrain();
  };

  const enqueuePendingReplayUnit = (
    unit: TimelinePendingReplayAdmission,
    onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
  ) => {
    const pending = pendingSavesRefs.pendingQueueRef.current;
    const {
      focusField,
      focusKey,
      surface,
      rowSnapshot,
      continueOnFreshDraft,
      detectAutoResolution,
      promoteToCommittedRowInspect,
      viewportContinuityToken,
      ...input
    } = unit;
    const meta: TimelineReplayContext = {
      draftRevisions: ports.captureEditorDrafts(
        input.rowKey,
        surface,
        input.kind === "create"
          ? undefined
          : new Set(
              (
                (input.payloadIntent.changes ?? []) as readonly {
                  field_key: string;
                }[]
              ).map((change) => change.field_key),
            ),
      ),
      sheetRef: ports.sheetRef,
      focusField,
      focusKey,
      surface,
      rowSnapshot,
      continueOnFreshDraft,
      detectAutoResolution,
      promoteToCommittedRowInspect,
      viewportContinuityToken,
    };
    const pendingInput: PendingReplayUnitInput = {
      ...input,
      presentationHint: { ...input.presentationHint, sheetRef: ports.sheetRef },
    };
    const admission = pending.model.admit(pendingInput);
    if (admission.status === "duplicate") {
      if (onSettled !== undefined) {
        const callbacks =
          completionCallbacksRef.current.get(admission.unit.id) ?? [];
        callbacks.push({ run: onSettled, values: {}, superseded: false });
        completionCallbacksRef.current.set(admission.unit.id, callbacks);
      }
      clearViewportContinuity(unit.viewportContinuityToken);
      publishPendingQueueState();
      return;
    }
    if (admission.status === "refused") {
      onSettled?.({
        kind: "rejected_mutation",
        message: "The pending edit queue rejected this mutation.",
      });
      clearPendingSignatureForUnit(unit);
      clearViewportContinuity(unit.viewportContinuityToken);
      publishPendingQueueState();
      return;
    }

    const previous = contextByUnitId.get(admission.unit.id);
    const callbacks =
      completionCallbacksRef.current.get(admission.unit.id) ?? [];
    if (admission.status === "coalesced" && previous) {
      const values = { ...meta.rowSnapshot.values };
      const baselines = { ...meta.rowSnapshot.committedValues };
      for (const binding of timelineScalarBindings) {
        const changed =
          rowSnapshot.values[binding.key] !==
          rowSnapshot.committedValues[binding.key];
        if (!changed) {
          values[binding.key] = previous.rowSnapshot.values[binding.key];
          baselines[binding.key] =
            previous.rowSnapshot.committedValues[binding.key];
        } else {
          for (const callback of callbacks)
            if (
              callback.values[binding.key] !== undefined &&
              callback.values[binding.key] !== rowSnapshot.values[binding.key]
            )
              callback.superseded = true;
        }
      }
      meta.rowSnapshot = {
        ...meta.rowSnapshot,
        values,
        committedValues: baselines,
      };
      meta.draftRevisions = new Map([
        ...(previous.draftRevisions ?? []),
        ...(meta.draftRevisions ?? []),
      ]);
    }
    if (onSettled !== undefined) {
      const values: Partial<RowValues> = {};
      for (const binding of timelineScalarBindings)
        if (
          rowSnapshot.values[binding.key] !==
          rowSnapshot.committedValues[binding.key]
        )
          values[binding.key] = rowSnapshot.values[binding.key];
      callbacks.push({ run: onSettled, values, superseded: false });
      completionCallbacksRef.current.set(admission.unit.id, callbacks);
    }

    contextByUnitId.set(admission.unit.id, meta);
    mutationRuntime.claimMutationUnit(admission.unit.id, {
      kind: "timeline_row",
      viewSchemaId: admission.unit.viewSchemaId,
    });
    pendingSavesRefs.pendingSignaturesRef.current.set(
      admission.unit.rowKey,
      admission.unit.mutationSignature,
    );
    recordWorkbookTiming("pending_unit_admitted", {
      clientTxnId: admission.unit.clientTxnId,
      kind: admission.unit.kind,
      rowKey: admission.unit.rowKey,
    });
    publishPendingQueueState();
    requestPendingReplay(
      admission.status === "coalesced" ? "coalesced_unit" : "admitted_unit",
    );
  };

  const retryBlockedEdit = (unitId: string): boolean => {
    const pending = pendingSavesRefs.pendingQueueRef.current;
    let replacementClientTxnId: string;
    try {
      replacementClientTxnId = mutationCommands.createConflictRecoveryId();
    } catch (error) {
      setMutationError(
        error instanceof Error
          ? error.message
          : "A secure request identifier could not be created.",
      );
      publishPendingQueueState();
      return false;
    }
    const recovery = pending.model.retryHaltedWithNewClientTxnId(
      unitId,
      replacementClientTxnId,
    );
    if (!recovery.recovered) {
      publishPendingQueueState();
      return false;
    }
    setMutationError(null);
    publishPendingQueueState();
    requestPendingReplay("client_transaction_conflict_retried");
    return true;
  };

  const discardBlockedEdit = (unitId: string): boolean => {
    const pending = pendingSavesRefs.pendingQueueRef.current;
    const recovery = pending.model.discardHaltedUnit(unitId);
    const meta = recovery.recovered
      ? contextByUnitId.get(recovery.unit.id)
      : undefined;
    const plan = planTimelineDiscard({
      hasMetadata: meta !== undefined,
      recovered: recovery.recovered,
    });
    if (plan.kind === "refused" || !recovery.recovered) {
      publishPendingQueueState();
      return false;
    }
    reconcileDiscardedPendingUnit(
      recovery.unit,
      recovery.snapshot.units,
      contextByUnitId,
    );
    contextByUnitId.delete(recovery.unit.id);
    mutationRuntime.releaseMutationUnit(recovery.unit.id);
    clearPendingSignatureForUnit(recovery.unit);
    if (plan.clearViewportContinuity && meta !== undefined) {
      clearViewportContinuity(meta.viewportContinuityToken);
    }
    settleCompletionCallbacks(recovery.unit.id, {
      kind: "rejected_mutation",
      message: "The blocked edit was discarded.",
    });
    setMutationError(null);
    publishPendingQueueState();
    requestPendingReplay("blocked_edit_discarded");
    return true;
  };

  const rejectMissingMetadata = (
    pending: WorkbookPendingQueueRuntime,
    unit: PendingReplayUnitState,
  ) => {
    const dispatch = pending.model.markDispatched(unit.id);
    if (dispatch !== null) {
      pending.model.settleDispatched({
        ok: false,
        status: 0,
        error: {
          code: "pending_runtime_metadata_missing",
          message: "Queued edit metadata is missing.",
        },
      });
    }
    publishPendingQueueState();
    settleCompletionCallbacks(unit.id, {
      kind: "rejected_mutation",
      message: "Queued edit metadata is missing.",
    });
  };

  const handleNonDispatchAdmission = (
    plan: Exclude<TimelineReplayAdmissionPlan, { readonly kind: "dispatch" }>,
    pending: WorkbookPendingQueueRuntime,
    unit: PendingReplayUnitState | null,
  ) => {
    if (plan.kind === "idle" || plan.kind === "pause") {
      publishPendingQueueState();
      return;
    }
    if (plan.kind === "reject_missing_metadata") {
      if (unit !== null) rejectMissingMetadata(pending, unit);
      return;
    }
    publishPendingQueueState();
    schedulePendingReplayRetry();
  };

  const refreshTimelineConflict = async (
    unit: PendingReplayUnitState,
    committedRowVersion: number,
  ): Promise<WorkbookOperationOutcome<WorkbookPendingMutationAccepted>> => {
    let clientTxnId: string;
    try {
      clientTxnId = mutationCommands.createConflictRecoveryId();
    } catch (error) {
      return {
        kind: "rejected",
        failure: {
          kind: "validation",
          message:
            error instanceof Error
              ? error.message
              : "A secure request identifier could not be created.",
        },
      };
    }
    if (unit.identity.kind !== "patch") {
      return {
        kind: "rejected",
        failure: {
          kind: "validation",
          message: "The original conflict mutation is unavailable.",
        },
      };
    }
    const refreshedUnit: PendingReplayUnitState = {
      ...unit,
      id: `${clientTxnId}:patch`,
      clientTxnId,
      status: "in_flight",
      identity: { ...unit.identity, client_txn_id: clientTxnId },
    };
    try {
      const refreshed = await mutationRuntime.dispatchPendingMutation({
        committedRowVersion,
        unit: refreshedUnit,
      });
      if (refreshed.kind === "rejected") {
        mutationRuntime.resolveSocketClientTxn(clientTxnId);
      }
      return refreshed;
    } catch {
      mutationRuntime.resolveSocketClientTxn(clientTxnId);
      return {
        kind: "rejected",
        failure: {
          kind: "retryable",
          message: "The conflict could not be refreshed.",
        },
      };
    }
  };

  const registerTimelineConflict = (
    settlement: TimelineConflictSettlement,
    conflict: Extract<
      WorkbookOperationFailure,
      { readonly kind: "same_field_conflict" }
    >["conflict"],
    meta: TimelineReplayContext,
    failureMessage: string,
  ) => {
    clearViewportContinuity(meta.viewportContinuityToken);
    settleCompletionCallbacks(settlement.unit.id, {
      kind: "conflict",
      message: failureMessage,
    });
    const registered = registerMutationConflict(
      conflict,
      settlement.unit.rowKey,
      meta.focusField,
      meta.surface,
      () => refreshTimelineConflict(settlement.unit, conflict.base_row_version),
      meta.sheetRef,
      new Map(
        [...(meta.draftRevisions ?? [])].filter(([key]) => {
          const binding = timelineScalarBindingForField(conflict.field_key);
          return (
            binding !== null &&
            (key ===
              inputFocusKey(settlement.unit.rowKey, binding.key, "grid") ||
              key ===
                inputFocusKey(settlement.unit.rowKey, binding.key, "inspector"))
          );
        }),
      ),
    );
    if (!registered) {
      setMutationError(failureMessage);
      publishPendingQueueState();
      return;
    }
    contextByUnitId.delete(settlement.unit.id);
    mutationRuntime.releaseMutationUnit(settlement.unit.id);
    clearPendingSignatureForUnit(settlement.unit);
    publishPendingQueueState();
  };

  const handleRejectedMutation = (
    pending: WorkbookPendingQueueRuntime,
    unit: PendingReplayUnitState,
    meta: TimelineReplayContext,
    failure: WorkbookOperationFailure,
  ) => {
    mutationRuntime.resolveSocketClientTxn(unit.clientTxnId);
    const publicFailure = workbookPendingMutationFailureResult(failure);
    const settlement = pending.model.settleDispatched({
      ok: false,
      status: publicFailure.status,
      error: publicFailure.error,
    });
    const plan = planTimelineRejectedSettlement(settlement, failure);
    if (plan.kind === "request_authorization") {
      publishPendingQueueState();
      settleCompletionCallbacks(unit.id, { kind: "accepted" });
      requestAuthorizationRecovery();
      return;
    }
    if (
      plan.kind === "register_conflict" &&
      settlement.outcome === "same_field_conflict"
    ) {
      registerTimelineConflict(
        settlement,
        plan.conflict,
        meta,
        failure.message,
      );
      return;
    }
    if (plan.kind === "retry") {
      publishPendingQueueState();
      schedulePendingReplayRetry();
      return;
    }
    if (plan.kind !== "halt")
      setMutationError(
        plan.kind === "invalid_settlement" ? plan.message : failure.message,
      );
    settleCompletionCallbacks(unit.id, gridOutcomeForFailure(failure));
    publishPendingQueueState();
  };

  const handleAcceptedMutation = async (
    pending: WorkbookPendingQueueRuntime,
    unit: PendingReplayUnitState,
    meta: TimelineReplayContext,
    currentRowVersion: number | null | undefined,
    accepted: WorkbookPendingMutationAccepted,
  ) => {
    recordWorkbookTiming("pending_result_apply_start", {
      clientTxnId: unit.clientTxnId,
      kind: unit.kind,
      rowKey: unit.rowKey,
    });
    const plan = planTimelineAcceptedProjection({
      currentRowVersion,
      postMutationQueryRefreshRequired: ports.postMutationQueryRefreshRequired,
      responseRowVersion: accepted.row.row_version,
    });
    const appliedRow = {
      record_id: accepted.row.record_id,
      row_version: accepted.row.row_version,
    };
    ports.acceptEditorPredecessor(
      rowFromApi(accepted.row),
      unit.identity.kind === "patch"
        ? unit.identity.changes.map((change) => change.field_key)
        : timelineScalarBindings.map((binding) => binding.fieldKey),
      meta.rowSnapshot.committedValues,
    );
    if (unit.identity.kind === "patch") {
      const committed = rowFromApi(accepted.row);
      const fields = new Set(
        unit.identity.changes.map((change) => change.field_key),
      );
      for (const [id, context] of contextByUnitId) {
        if (
          id === unit.id ||
          context.rowSnapshot.recordId !== accepted.row.record_id
        )
          continue;
        const values = { ...context.rowSnapshot.committedValues };
        for (const binding of timelineScalarBindings)
          if (
            fields.has(binding.fieldKey) &&
            values[binding.key] ===
              meta.rowSnapshot.committedValues[binding.key]
          )
            values[binding.key] = committed.committedValues[binding.key];
        contextByUnitId.set(id, {
          ...context,
          rowSnapshot: { ...context.rowSnapshot, committedValues: values },
        });
      }
    }
    try {
      clearSubmittedScalarEditorDraftValuesForRow(
        unit.rowKey,
        meta.rowSnapshot.values,
        meta.rowSnapshot.recordId === null
          ? meta.rowSnapshot.collectionDrafts
          : isCollectionDraftKey(meta.focusField)
            ? {
                [meta.focusField]:
                  meta.rowSnapshot.collectionDrafts[meta.focusField],
              }
            : undefined,
        meta.draftRevisions,
      );
      applyAcceptedRowMutation(unit.rowKey, accepted, {
        continueOnFreshDraft:
          meta.continueOnFreshDraft && meta.rowSnapshot.recordId === null,
        detectAutoResolution: meta.detectAutoResolution,
        previousRow: meta.rowSnapshot,
        operationId: unit.clientTxnId,
        promoteToCommittedRowInspect: meta.promoteToCommittedRowInspect,
        ...(meta.viewportContinuityToken === undefined
          ? {}
          : { viewportContinuityToken: meta.viewportContinuityToken }),
      });
    } catch (error) {
      recordWorkbookTiming("pending_result_apply_error", {
        clientTxnId: unit.clientTxnId,
        kind: unit.kind,
        message: error instanceof Error ? error.message : String(error),
        rowKey: unit.rowKey,
      });
      mutationRuntime.retainSurfaceRefreshDebt("cartulary.view.timeline.v2");
      setRefreshError(
        "The edit was accepted. Timeline presentation needs a refresh.",
      );
    }
    recordWorkbookTiming("pending_result_apply_end", {
      clientTxnId: unit.clientTxnId,
      kind: unit.kind,
      recordId: appliedRow.record_id,
      rowKey: unit.rowKey,
      rowVersion: appliedRow.row_version,
      staleResponseProtected: plan.preserveKnownCommittedRow,
    });
    const followOnCreatePatches: Record<
      string,
      Record<string, unknown> | null
    > = {};
    if (unit.kind === "create") {
      const committed = rowFromApi(accepted.row);
      for (const queued of pending.model.snapshot().units) {
        if (
          queued.id === unit.id ||
          queued.rowKey !== unit.rowKey ||
          queued.kind !== "create"
        )
          continue;
        const context = contextByUnitId.get(queued.id);
        if (context === undefined) continue;
        const payload = buildFollowOnCapturePatch(
          committed,
          meta.rowSnapshot,
          context.rowSnapshot,
          queued.clientTxnId,
        );
        followOnCreatePatches[queued.id] = payload;
        contextByUnitId.set(queued.id, {
          ...context,
          rowSnapshot: { ...committed, values: context.rowSnapshot.values },
          continueOnFreshDraft: false,
        });
      }
    }
    const settlement = pending.model.settleDispatched({
      ok: true,
      row: appliedRow,
      change_set_id: accepted.changeSetId,
      followOnCreatePatches,
    });
    if (settlement.outcome === "success") {
      if (unit.kind === "create")
        pendingSavesRefs.pendingSignaturesRef.current.delete(unit.rowKey);
      for (const completed of [
        settlement.unit,
        ...(settlement.acknowledgedFollowOnUnits ?? []),
      ]) {
        contextByUnitId.delete(completed.id);
        mutationRuntime.releaseMutationUnit(completed.id);
        clearPendingSignatureForUnit(completed);
        settleCompletionCallbacks(completed.id, { kind: "accepted" });
      }
    }
    publishPendingQueueState();
    requestPendingReplay("unit_completed");
    if (plan.refreshAfterApply) {
      mutationRuntime.retainSurfaceRefreshDebt("cartulary.view.timeline.v2");
      void mutationRuntime.refreshSurface("cartulary.view.timeline.v2");
    }
  };

  const dispatchTimelineUnit = async (
    pending: WorkbookPendingQueueRuntime,
    unit: PendingReplayUnitState,
    meta: TimelineReplayContext,
    committedRowVersion: number | null,
    currentRowVersion: number | null | undefined,
    currentRow: WorkbookRow | null,
  ) => {
    const dispatch = pending.model.markDispatched(unit.id, (candidate) => {
      const changes =
        candidate.identity.kind === "patch" ? candidate.identity.changes : [];
      const fields = new Set(changes.map((change) => change.field_key));
      const stale =
        currentRow &&
        timelineScalarBindings.some(
          (binding) =>
            fields.has(binding.fieldKey) &&
            meta.rowSnapshot.committedValues[binding.key] !==
              currentRow.committedValues[binding.key],
        );
      if (stale) {
        const message =
          "Review the changed saved value before submitting this retained draft. Discard the queued attempt to edit it locally.";
        pending.model.haltBeforeDispatch(unit.id, message);
        settleCompletionCallbacks(unit.id, { kind: "stale_target", message });
        return null;
      }
      return committedRowVersion;
    });
    if (dispatch === null) {
      publishPendingQueueState();
      return;
    }
    publishPendingQueueState();
    const epoch = mutationRuntime.authorizationEpoch;
    const finishPresentation = ports.beginDispatch?.();
    try {
      let result: WorkbookOperationOutcome<WorkbookPendingMutationAccepted>;
      try {
        recordWorkbookTiming("pending_fetch_start", {
          clientTxnId: dispatch.unit.clientTxnId,
          kind: dispatch.unit.kind,
          rowKey: dispatch.unit.rowKey,
        });
        result = await mutationRuntime.dispatchPendingMutation({
          committedRowVersion:
            dispatch.identity.kind === "patch"
              ? dispatch.identity.base_row_version
              : null,
          unit: dispatch.unit,
        });
      } catch {
        mutationRuntime.resolveSocketClientTxn(dispatch.unit.clientTxnId);
        pending.model.settleDispatched({
          ok: false,
          status: 0,
          error: {
            code: "transport_failure",
            message: "Transport failure",
            retryable: true,
          },
        });
        publishPendingQueueState();
        // Uncertain delivery is still pending. Only an authoritative acknowledgement
        // may close the editor or perform its queued navigation.
        schedulePendingReplayRetry();
        return;
      }
      if (mutationRuntime.retired) return;
      if (
        result.kind === "rejected" &&
        epoch !== mutationRuntime.authorizationEpoch
      ) {
        const paused = pending.model.snapshot().authPaused;
        const failure = workbookPendingMutationFailureResult(result.failure);
        pending.model.settleDispatched({
          ok: false,
          status: failure.status,
          error: failure.error,
        });
        if (!paused) pending.model.resumeAfterAuthRecovery();
        publishPendingQueueState();
        mutationRuntime.requestDrain();
        return;
      }
      if (result.kind === "rejected") {
        handleRejectedMutation(pending, dispatch.unit, meta, result.failure);
        return;
      }
      await handleAcceptedMutation(
        pending,
        dispatch.unit,
        meta,
        currentRowVersion,
        result.value,
      );
    } finally {
      finishPresentation?.();
    }
  };

  const replayPendingQueue = async (
    expectedUnit: PendingReplayUnitState,
    envelope: Extract<
      WorkbookMutationOwnerEnvelope,
      { readonly kind: "timeline_row" }
    >,
  ) => {
    const pending = pendingSavesRefs.pendingQueueRef.current;
    if (pending.model.snapshot().authPaused || mutationRuntime.retired) return;
    const unit = pending.model.peekNextQueued()?.unit ?? null;
    const meta = unit === null ? undefined : contextByUnitId.get(unit.id);
    const epoch = mutationRuntime.authorizationEpoch;
    const captured = unit !== null && pending.model.wasDispatched(unit.id);
    let currentRow: ReturnType<typeof currentTimelineReplayRow> | null;
    try {
      currentRow =
        !captured && unit && ports.readCurrentRow
          ? await ports.readCurrentRow(unit)
          : currentTimelineReplayRow(
              unit,
              rowsRef.current,
              latestCommittedTimelineRow,
            );
    } catch {
      if (
        epoch !== mutationRuntime.authorizationEpoch ||
        mutationRuntime.retired
      )
        return;
      setRefreshError(
        "Current Timeline records could not be verified. Queued edits are retained.",
      );
      requestAuthorizationRecovery();
      return;
    }
    if (epoch !== mutationRuntime.authorizationEpoch || mutationRuntime.retired)
      return;
    if (pending.model.peekNextQueued()?.unit.id !== unit?.id) return;
    const snapshot = pending.model.snapshot();
    const plan = planTimelineReplayAdmission({
      candidate: unit,
      currentRowVersion:
        captured && unit?.identity.kind === "patch"
          ? unit.identity.base_row_version
          : currentRow?.rowVersion,
      envelopeViewSchemaId: envelope.viewSchemaId,
      expectedUnitId: expectedUnit.id,
      hasLocalConflict: Object.keys(conflictQueueRef.current).length > 0,
      hasMetadata: meta !== undefined,
      refreshBlocked:
        unit !== null && refreshBlocksWorkbookPendingUnit(pending, unit),
      snapshot,
    });
    if (plan.kind !== "dispatch") {
      handleNonDispatchAdmission(plan, pending, unit);
      return;
    }
    if (unit === null || meta === undefined) return;
    await dispatchTimelineUnit(
      pending,
      unit,
      meta,
      plan.committedRowVersion,
      currentRow?.rowVersion,
      currentRow ?? null,
    );
  };

  return {
    acceptBatchPredecessor: (
      receipt: WorkbookBatchReceipt,
      attempt: WorkbookBatchAttempt,
    ) => {
      const { plan } = attempt;
      if (
        receipt.viewSchemaId !== "cartulary.view.timeline.v2" ||
        plan.operation !== "applyWorkbookBulkMutation" ||
        plan.request.kind !== "clear_cells_v1"
      )
        return;
      const pending = pendingSavesRefs.pendingQueueRef.current;
      for (const accepted of receipt.rows) {
        const committed = rowFromApi(
          normalizeTimelineFullRow(accepted, "clear predecessor"),
        );
        const fields = (plan.request.field_keys ?? []).filter(
          (field) =>
            !receipt.conflicts.some(
              (conflict) =>
                conflict.record_id === accepted.record_id &&
                conflict.field_key === field,
            ),
        );
        ports.acceptEditorPredecessor(committed, fields);
        for (const unit of pending.model.snapshot().units) {
          if (
            unit.recordId !== accepted.record_id ||
            pending.model.wasDispatched(unit.id)
          )
            continue;
          const context = contextByUnitId.get(unit.id);
          if (!context) continue;
          const values = { ...context.rowSnapshot.committedValues };
          for (const binding of timelineScalarBindings)
            if (fields.includes(binding.fieldKey))
              values[binding.key] = committed.committedValues[binding.key];
          contextByUnitId.set(unit.id, {
            ...context,
            rowSnapshot: { ...context.rowSnapshot, committedValues: values },
          });
        }
      }
    },
    drain: replayPendingQueue,
    detachPresentation: () => completionCallbacksRef.current.clear(),
    discardBlockedEdit,
    enqueuePendingReplayUnit,
    retryBlockedEdit,
  };
}
