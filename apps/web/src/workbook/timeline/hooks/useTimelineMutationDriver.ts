import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { useCallback, useEffect, useRef } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { TimelineMutationIdentityPort } from "../../mutations/workbookMutationCommandPorts";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationAccepted } from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationOwnerEnvelope } from "../../runtime/WorkbookMutationDriverRegistry";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
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
import type {
  FocusFieldKey,
  RowValues,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import {
  planTimelineAcceptedProjection,
  planTimelineDiscard,
  planTimelineRejectedSettlement,
  planTimelineReplayAdmission,
  type TimelinePendingReplayAdmission,
  type TimelineReplayAdmissionPlan,
} from "../models/timelineMutationDriverPlans";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import type { WorkbookRow } from "../models/timelineRowModel";

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

export function useTimelineMutationDriver({
  applyAcceptedRowMutation,
  clearSubmittedScalarEditorDraftValuesForRow,
  clearViewportContinuity,
  conflictQueueRef,
  registerMutationConflict,
  latestCommittedTimelineRow,
  loadRows,
  mutationCommands,
  mutationRuntime,
  sheetRef,
  pendingSavesRefs,
  postMutationQueryRefreshRequired,
  publishPendingQueueState,
  reconcileDiscardedPendingUnit,
  recordWorkbookTiming,
  rowsRef,
  requestAuthorizationRecovery,
  setRefreshError,
  rowStoreCommands,
}: {
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    accepted: WorkbookPendingMutationAccepted,
    options?: {
      clearActiveCollectionFocusKey?: string;
      continueOnFreshDraft?: boolean;
      detectAutoResolution?: boolean;
      promoteToCommittedRowInspect?: boolean;
      viewportContinuityToken?: number;
    },
  ) => WorkbookRow;
  readonly clearSubmittedScalarEditorDraftValuesForRow: (
    rowKey: string,
    submittedValues: RowValues,
    submittedCollections?: Partial<WorkbookRow["collectionDrafts"]>,
  ) => void;
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
  readonly rowStoreCommands: TimelineRowStoreCommands;
}) {
  const { replaceRows } = rowStoreCommands;
  const contextByUnitId = pendingSavesRefs.replayContextByUnitId;
  const completionCallbacksRef = useRef(
    new Map<string, Array<(outcome: GridEditCommitOutcome) => void>>(),
  );
  const settleCompletionCallbacks = useCallback(
    (unitId: string, outcome: GridEditCommitOutcome) => {
      const callbacks = completionCallbacksRef.current.get(unitId) ?? [];
      completionCallbacksRef.current.delete(unitId);
      for (const callback of callbacks) callback(outcome);
    },
    [],
  );
  const clearPendingSignatureForUnit = useCallback(
    (unit: {
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
        row.key === unit.rowKey &&
        row.pendingSignature === unit.mutationSignature
          ? { ...row, pendingSignature: null }
          : row,
      );
      rowsRef.current = nextRows;
      replaceRows(nextRows);
    },
    [pendingSavesRefs, replaceRows, rowsRef],
  );

  const schedulePendingReplayRetry = useCallback(() => {
    mutationRuntime.scheduleRetry(1000);
  }, [mutationRuntime]);

  const requestPendingReplay = useCallback(
    (reason: string) => {
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
    },
    [conflictQueueRef, mutationRuntime, pendingSavesRefs, recordWorkbookTiming],
  );

  const enqueuePendingReplayUnit = useCallback(
    (
      unit: TimelinePendingReplayAdmission,
      onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
    ) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      const snapshotBeforeAdmission = pending.model.snapshot();
      const admissionIsBacklogged =
        snapshotBeforeAdmission.inFlightCount > 0 ||
        snapshotBeforeAdmission.queuedCount > 0;
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
        sheetRef,
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
        presentationHint: { ...input.presentationHint, sheetRef },
      };
      const admission = pending.model.admit(pendingInput);
      if (admission.status === "duplicate") {
        if (onSettled !== undefined) {
          const callbacks =
            completionCallbacksRef.current.get(admission.unit.id) ?? [];
          callbacks.push(onSettled);
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

      if (onSettled !== undefined && !admissionIsBacklogged) {
        const callbacks =
          completionCallbacksRef.current.get(admission.unit.id) ?? [];
        callbacks.push(onSettled);
        completionCallbacksRef.current.set(admission.unit.id, callbacks);
      }
      if (admissionIsBacklogged) onSettled?.({ kind: "accepted" });

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
    },
    [
      clearPendingSignatureForUnit,
      clearViewportContinuity,
      mutationRuntime,
      sheetRef,
      contextByUnitId,
      pendingSavesRefs,
      publishPendingQueueState,
      recordWorkbookTiming,
      requestPendingReplay,
    ],
  );

  const retryBlockedEdit = useCallback(
    (unitId: string): boolean => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      let replacementClientTxnId: string;
      try {
        replacementClientTxnId = mutationCommands.createConflictRecoveryId();
      } catch (error) {
        setRefreshError(
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
      setRefreshError(null);
      publishPendingQueueState();
      requestPendingReplay("client_transaction_conflict_retried");
      return true;
    },
    [
      mutationCommands,
      pendingSavesRefs,
      publishPendingQueueState,
      requestPendingReplay,
      setRefreshError,
    ],
  );

  const discardBlockedEdit = useCallback(
    (unitId: string): boolean => {
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
      setRefreshError(null);
      publishPendingQueueState();
      requestPendingReplay("blocked_edit_discarded");
      return true;
    },
    [
      clearPendingSignatureForUnit,
      clearViewportContinuity,
      contextByUnitId,
      mutationRuntime,
      pendingSavesRefs,
      publishPendingQueueState,
      reconcileDiscardedPendingUnit,
      requestPendingReplay,
      setRefreshError,
      settleCompletionCallbacks,
    ],
  );

  const rejectMissingMetadata = useCallback(
    (pending: WorkbookPendingQueueRuntime, unit: PendingReplayUnitState) => {
      const dispatch = pending.model.markDispatched(unit.id);
      if (dispatch !== null) {
        const settlement = pending.model.settleDispatched({
          ok: false,
          status: 0,
          error: {
            code: "pending_runtime_metadata_missing",
            message: "Queued edit metadata is missing.",
          },
        });
        if (settlement.outcome === "halted") {
          setRefreshError(settlement.halt.message);
        }
      }
      publishPendingQueueState();
      settleCompletionCallbacks(unit.id, {
        kind: "rejected_mutation",
        message: "Queued edit metadata is missing.",
      });
    },
    [publishPendingQueueState, setRefreshError, settleCompletionCallbacks],
  );

  const handleNonDispatchAdmission = useCallback(
    (
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
    },
    [
      publishPendingQueueState,
      rejectMissingMetadata,
      schedulePendingReplayRetry,
    ],
  );

  const refreshTimelineConflict = useCallback(
    async (
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
    },
    [mutationCommands, mutationRuntime],
  );

  const registerTimelineConflict = useCallback(
    (
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
        () =>
          refreshTimelineConflict(settlement.unit, conflict.base_row_version),
        meta.sheetRef,
      );
      if (!registered) {
        setRefreshError(failureMessage);
        publishPendingQueueState();
        return;
      }
      contextByUnitId.delete(settlement.unit.id);
      mutationRuntime.releaseMutationUnit(settlement.unit.id);
      clearPendingSignatureForUnit(settlement.unit);
      publishPendingQueueState();
    },
    [
      clearPendingSignatureForUnit,
      clearViewportContinuity,
      mutationRuntime,
      contextByUnitId,
      publishPendingQueueState,
      refreshTimelineConflict,
      registerMutationConflict,
      setRefreshError,
      settleCompletionCallbacks,
    ],
  );

  const handleRejectedMutation = useCallback(
    (
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
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
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
        settleCompletionCallbacks(unit.id, { kind: "accepted" });
        schedulePendingReplayRetry();
        return;
      }
      setRefreshError(
        plan.kind === "halt" || plan.kind === "invalid_settlement"
          ? plan.message
          : failure.message,
      );
      settleCompletionCallbacks(unit.id, gridOutcomeForFailure(failure));
      publishPendingQueueState();
    },
    [
      mutationRuntime,
      publishPendingQueueState,
      registerTimelineConflict,
      requestAuthorizationRecovery,
      schedulePendingReplayRetry,
      setRefreshError,
      settleCompletionCallbacks,
    ],
  );

  const handleAcceptedMutation = useCallback(
    async (
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
        postMutationQueryRefreshRequired,
        responseRowVersion: accepted.row.row_version,
      });
      const appliedRow = {
        record_id: accepted.row.record_id,
        row_version: accepted.row.row_version,
      };
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
        );
        const clearActiveCollectionFocusKey =
          meta.surface === "grid" && isCollectionDraftKey(meta.focusField)
            ? meta.focusKey
            : undefined;
        applyAcceptedRowMutation(unit.rowKey, accepted, {
          ...(clearActiveCollectionFocusKey === undefined
            ? {}
            : { clearActiveCollectionFocusKey }),
          continueOnFreshDraft:
            meta.continueOnFreshDraft && meta.rowSnapshot.recordId === null,
          detectAutoResolution: meta.detectAutoResolution,
          promoteToCommittedRowInspect: meta.promoteToCommittedRowInspect,
          viewportContinuityToken: meta.viewportContinuityToken,
        });
      } catch (error) {
        recordWorkbookTiming("pending_result_apply_error", {
          clientTxnId: unit.clientTxnId,
          kind: unit.kind,
          message: error instanceof Error ? error.message : String(error),
          rowKey: unit.rowKey,
        });
        const settlement = pending.model.settleDispatched({
          ok: false,
          status: 0,
          error: {
            code: "client_apply_error",
            message: error instanceof Error ? error.message : String(error),
          },
        });
        if (settlement.outcome === "halted") {
          setRefreshError(settlement.halt.message);
        }
        settleCompletionCallbacks(unit.id, {
          kind: "rejected_mutation",
          message:
            error instanceof Error
              ? error.message
              : "The accepted edit could not be applied.",
        });
        publishPendingQueueState();
        return;
      }
      if (plan.refreshAfterApply) {
        try {
          await loadRows({ showLoading: false });
        } catch {
          setRefreshError("Timeline projection refresh failed.");
        }
      }
      recordWorkbookTiming("pending_result_apply_end", {
        clientTxnId: unit.clientTxnId,
        kind: unit.kind,
        recordId: appliedRow.record_id,
        rowKey: unit.rowKey,
        rowVersion: appliedRow.row_version,
        staleResponseProtected: plan.preserveKnownCommittedRow,
      });
      const settlement = pending.model.settleDispatched({
        ok: true,
        row: appliedRow,
        change_set_id: accepted.changeSetId,
      });
      if (settlement.outcome === "success") {
        contextByUnitId.delete(settlement.unit.id);
        mutationRuntime.releaseMutationUnit(settlement.unit.id);
        clearPendingSignatureForUnit(settlement.unit);
        settleCompletionCallbacks(settlement.unit.id, { kind: "accepted" });
      }
      publishPendingQueueState();
      requestPendingReplay("unit_completed");
    },
    [
      applyAcceptedRowMutation,
      clearPendingSignatureForUnit,
      clearSubmittedScalarEditorDraftValuesForRow,
      loadRows,
      contextByUnitId,
      mutationRuntime,
      postMutationQueryRefreshRequired,
      publishPendingQueueState,
      recordWorkbookTiming,
      requestPendingReplay,
      setRefreshError,
      settleCompletionCallbacks,
    ],
  );

  const dispatchTimelineUnit = useCallback(
    async (
      pending: WorkbookPendingQueueRuntime,
      unit: PendingReplayUnitState,
      meta: TimelineReplayContext,
      committedRowVersion: number | null,
      currentRowVersion: number | null | undefined,
    ) => {
      const dispatch = pending.model.markDispatched(unit.id);
      if (dispatch === null) {
        publishPendingQueueState();
        return;
      }
      publishPendingQueueState();
      let result: WorkbookOperationOutcome<WorkbookPendingMutationAccepted>;
      try {
        recordWorkbookTiming("pending_fetch_start", {
          clientTxnId: dispatch.unit.clientTxnId,
          kind: dispatch.unit.kind,
          rowKey: dispatch.unit.rowKey,
        });
        result = await mutationRuntime.dispatchPendingMutation({
          committedRowVersion,
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
        settleCompletionCallbacks(dispatch.unit.id, { kind: "accepted" });
        schedulePendingReplayRetry();
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
    },
    [
      handleAcceptedMutation,
      handleRejectedMutation,
      mutationRuntime,
      publishPendingQueueState,
      recordWorkbookTiming,
      schedulePendingReplayRetry,
      settleCompletionCallbacks,
    ],
  );

  const replayPendingQueue = useCallback(
    async (
      expectedUnit: PendingReplayUnitState,
      envelope: Extract<
        WorkbookMutationOwnerEnvelope,
        { readonly kind: "timeline_row" }
      >,
    ) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      const snapshot = pending.model.snapshot();
      const unit = pending.model.peekNextQueued()?.unit ?? null;
      const meta = unit === null ? undefined : contextByUnitId.get(unit.id);
      const currentRow = currentTimelineReplayRow(
        unit,
        rowsRef.current,
        latestCommittedTimelineRow,
      );
      const plan = planTimelineReplayAdmission({
        candidate: unit,
        currentRowVersion: currentRow?.rowVersion,
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
      );
    },
    [
      conflictQueueRef,
      dispatchTimelineUnit,
      handleNonDispatchAdmission,
      latestCommittedTimelineRow,
      contextByUnitId,
      pendingSavesRefs,
      rowsRef,
    ],
  );

  const replayPendingQueueRef = useRef(replayPendingQueue);
  replayPendingQueueRef.current = replayPendingQueue;
  const publishPendingQueueStateRef = useRef(publishPendingQueueState);
  publishPendingQueueStateRef.current = publishPendingQueueState;
  const setRefreshErrorRef = useRef(setRefreshError);
  setRefreshErrorRef.current = setRefreshError;

  useEffect(() => {
    const registration = mutationRuntime.registerDriver({
      kind: "timeline_row",
      drain: (unit, envelope) => replayPendingQueueRef.current(unit, envelope),
    });
    if (!registration.accepted) {
      setRefreshErrorRef.current(
        "Timeline mutation driver is already registered.",
      );
      publishPendingQueueStateRef.current();
      return;
    }
    mutationRuntime.requestDrain();
    return registration.unregister;
  }, [mutationRuntime]);

  return {
    discardBlockedEdit,
    enqueuePendingReplayUnit,
    retryBlockedEdit,
  };
}
