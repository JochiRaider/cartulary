import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { useCallback, useRef } from "react";
import type { TimelineMutationIdentityPort } from "../../mutations/workbookMutationCommandPorts";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import type {
  WorkbookPendingMutationAccepted,
  WorkbookPendingMutationPort,
} from "../../ports/WorkbookPendingMutationPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { workbookPendingMutationFailureResult } from "../../runtime/workbookPendingMutationSettlement";
import {
  refreshBlocksWorkbookPendingUnit,
  type WorkbookPendingReplayAdmissionRequest,
  type WorkbookPendingSavesRefs,
} from "../../runtime/workbookPendingReplayRuntime";
import type {
  PendingReplayUnitInput,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";
import type { PendingReplayRuntimeMeta } from "../models/timelineControllerPorts";
import type {
  FocusFieldKey,
  RowValues,
  TimelineScalarEditorSurface,
  WorkbookRow,
} from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

type TimelinePendingReplayControllerAdmission =
  WorkbookPendingReplayAdmissionRequest<PendingReplayRuntimeMeta>;

function isCollectionDraftKey(
  focusField: FocusFieldKey,
): focusField is "hostRefs" | "identityRefs" | "tags" {
  return (
    focusField === "hostRefs" ||
    focusField === "identityRefs" ||
    focusField === "tags"
  );
}

export function useTimelinePendingReplayController({
  applyAcceptedRowMutation,
  clearSubmittedScalarEditorDraftValuesForRow,
  clearViewportContinuity,
  conflictQueueRef,
  registerMutationConflict,
  latestCommittedTimelineRow,
  loadRowsRef,
  mutationCommands,
  mutationRuntime,
  pendingSavesRefsRef,
  postMutationQueryRefreshRequired,
  publishPendingQueueState,
  reconcileDiscardedPendingUnit,
  recordWorkbookTiming,
  resolvePendingSocketTxn,
  rowsRef,
  requestAuthorizationRecovery,
  setRefreshError,
  setRows,
  pendingMutationPort,
  trackPendingSocketTxn,
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
  ) => boolean;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly loadRowsRef: TimelineMutableRef<
    (options: { readonly showLoading: boolean }) => Promise<void>
  >;
  readonly mutationCommands: TimelineMutationIdentityPort;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    WorkbookPendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly postMutationQueryRefreshRequired: boolean;
  readonly publishPendingQueueState: () => void;
  readonly reconcileDiscardedPendingUnit: (
    discardedUnit: PendingReplayUnitState,
    remainingUnits: readonly PendingReplayUnitState[],
  ) => void;
  readonly recordWorkbookTiming: (
    name: string,
    details?: Record<string, unknown>,
  ) => void;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: TimelineMutableRef<WorkbookRow[]>;
  readonly requestAuthorizationRecovery: () => void;
  readonly setRefreshError: (message: string | null) => void;
  readonly setRows: (rows: WorkbookRow[]) => void;
  readonly pendingMutationPort: WorkbookPendingMutationPort;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
}) {
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
        pendingSavesRefsRef.current.pendingSignaturesRef.current.get(
          unit.rowKey,
        ) === unit.mutationSignature
      ) {
        pendingSavesRefsRef.current.pendingSignaturesRef.current.delete(
          unit.rowKey,
        );
      }
      const nextRows = rowsRef.current.map((row) =>
        row.key === unit.rowKey &&
        row.pendingSignature === unit.mutationSignature
          ? { ...row, pendingSignature: null }
          : row,
      );
      rowsRef.current = nextRows;
      setRows(nextRows);
    },
    [pendingSavesRefsRef, rowsRef, setRows],
  );

  const replayPendingQueueRef = useRef<() => Promise<void>>(async () => {
    return undefined;
  });

  const schedulePendingReplayAfter = useCallback(
    (delayMs: number) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      if (pending.replayScheduled) {
        return;
      }
      pending.replayScheduled = true;
      pendingSavesRefsRef.current.pendingReplayTimerRef.current =
        window.setTimeout(() => {
          pendingSavesRefsRef.current.pendingReplayTimerRef.current = null;
          void replayPendingQueueRef.current();
        }, delayMs);
    },
    [pendingSavesRefsRef],
  );

  const schedulePendingReplay = useCallback(() => {
    schedulePendingReplayAfter(0);
  }, [schedulePendingReplayAfter]);
  pendingSavesRefsRef.current.schedulePendingReplayRef.current =
    schedulePendingReplay;

  const schedulePendingReplayRetry = useCallback(() => {
    schedulePendingReplayAfter(1000);
  }, [schedulePendingReplayAfter]);

  const requestPendingReplay = useCallback(
    (reason: string) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
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
      if (!readyForImmediateDrain) {
        schedulePendingReplay();
        return;
      }
      if (pendingSavesRefsRef.current.pendingReplayTimerRef.current !== null) {
        window.clearTimeout(
          pendingSavesRefsRef.current.pendingReplayTimerRef.current,
        );
        pendingSavesRefsRef.current.pendingReplayTimerRef.current = null;
      }
      pending.replayScheduled = false;
      recordWorkbookTiming("pending_replay_drain_immediate", { reason });
      void replayPendingQueueRef.current();
    },
    [
      conflictQueueRef,
      pendingSavesRefsRef,
      recordWorkbookTiming,
      schedulePendingReplay,
    ],
  );

  const enqueuePendingReplayUnit = useCallback(
    (
      unit: TimelinePendingReplayControllerAdmission,
      onSettled?: ((outcome: GridEditCommitOutcome) => void) | undefined,
    ) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
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
      const meta: PendingReplayRuntimeMeta = {
        focusField,
        focusKey,
        surface,
        rowSnapshot,
        continueOnFreshDraft,
        detectAutoResolution,
        promoteToCommittedRowInspect,
        viewportContinuityToken,
      };
      const admission = pending.model.admit(input as PendingReplayUnitInput);
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

      pending.metaByUnitId.set(admission.unit.id, meta);
      pendingSavesRefsRef.current.pendingSignaturesRef.current.set(
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
      pendingSavesRefsRef,
      publishPendingQueueState,
      recordWorkbookTiming,
      requestPendingReplay,
    ],
  );

  const retryBlockedEdit = useCallback(
    (unitId: string): boolean => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
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
      pendingSavesRefsRef,
      publishPendingQueueState,
      requestPendingReplay,
      setRefreshError,
    ],
  );

  const discardBlockedEdit = useCallback(
    (unitId: string): boolean => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      const recovery = pending.model.discardHaltedUnit(unitId);
      if (!recovery.recovered) {
        publishPendingQueueState();
        return false;
      }
      const meta = pending.metaByUnitId.get(recovery.unit.id);
      reconcileDiscardedPendingUnit(recovery.unit, recovery.snapshot.units);
      pending.metaByUnitId.delete(recovery.unit.id);
      clearPendingSignatureForUnit(recovery.unit);
      if (meta !== undefined) {
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
      pendingSavesRefsRef,
      publishPendingQueueState,
      reconcileDiscardedPendingUnit,
      requestPendingReplay,
      setRefreshError,
      settleCompletionCallbacks,
    ],
  );

  replayPendingQueueRef.current = async () => {
    const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
    pending.replayScheduled = false;
    const snapshot = pending.model.snapshot();
    if (
      snapshot.authPaused ||
      snapshot.halted !== null ||
      snapshot.sameFieldConflicts.length > 0 ||
      Object.keys(conflictQueueRef.current).length > 0
    ) {
      publishPendingQueueState();
      return;
    }
    const candidate = pending.model.peekNextQueued();
    if (candidate === null) {
      publishPendingQueueState();
      return;
    }
    const unit = candidate.unit;
    if (refreshBlocksWorkbookPendingUnit(pending, unit)) {
      publishPendingQueueState();
      return;
    }
    const meta = pending.metaByUnitId.get(unit.id);
    if (meta === undefined) {
      if (mutationRuntime.ownsManagedUnit(unit.id)) {
        mutationRuntime.requestDrain();
        publishPendingQueueState();
        return;
      }
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
      return;
    }

    const currentRow =
      unit.recordId === null
        ? rowsRef.current.find((row) => row.key === unit.rowKey)
        : (latestCommittedTimelineRow(unit.recordId) ??
          rowsRef.current.find((row) => row.recordId === unit.recordId));
    if (
      unit.kind === "patch" &&
      (currentRow?.rowVersion === null || currentRow?.rowVersion === undefined)
    ) {
      publishPendingQueueState();
      schedulePendingReplayRetry();
      return;
    }

    const dispatch = pending.model.markDispatched(unit.id);
    if (dispatch === null) {
      publishPendingQueueState();
      return;
    }
    const dispatchedUnit = dispatch.unit;
    publishPendingQueueState();
    trackPendingSocketTxn(dispatchedUnit.clientTxnId);

    let result: WorkbookOperationOutcome<WorkbookPendingMutationAccepted>;
    try {
      recordWorkbookTiming("pending_fetch_start", {
        clientTxnId: dispatchedUnit.clientTxnId,
        kind: dispatchedUnit.kind,
        rowKey: dispatchedUnit.rowKey,
      });
      result = await pendingMutationPort.execute({
        committedRowVersion: currentRow?.rowVersion ?? null,
        unit: dispatchedUnit,
      });
    } catch {
      resolvePendingSocketTxn(dispatchedUnit.clientTxnId);
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
      settleCompletionCallbacks(dispatchedUnit.id, { kind: "accepted" });
      schedulePendingReplayRetry();
      return;
    }

    if (result.kind === "rejected") {
      resolvePendingSocketTxn(dispatchedUnit.clientTxnId);
      const publicFailure = workbookPendingMutationFailureResult(
        result.failure,
      );
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: publicFailure.status,
        error: publicFailure.error,
      });
      if (settlement.outcome === "auth_paused") {
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
        publishPendingQueueState();
        settleCompletionCallbacks(dispatchedUnit.id, { kind: "accepted" });
        requestAuthorizationRecovery();
        return;
      }

      if (settlement.outcome === "same_field_conflict") {
        clearViewportContinuity(meta.viewportContinuityToken);
        const message = result.failure.message;
        settleCompletionCallbacks(dispatchedUnit.id, {
          kind: "conflict",
          message,
        });
        if (
          result.failure.kind === "same_field_conflict" &&
          registerMutationConflict(
            result.failure.conflict,
            settlement.unit.rowKey,
            meta.focusField,
            meta.surface,
          )
        ) {
          pending.metaByUnitId.delete(settlement.unit.id);
          clearPendingSignatureForUnit(settlement.unit);
          publishPendingQueueState();
          return;
        }
        setRefreshError(result.failure.message);
        publishPendingQueueState();
        return;
      }

      if (settlement.outcome === "retryable_failure") {
        publishPendingQueueState();
        // Admission to the durable in-memory queue completes the editor
        // commit. Keep replay ownership here without pinning the grid editor
        // open while transport recovery is pending.
        settleCompletionCallbacks(dispatchedUnit.id, { kind: "accepted" });
        schedulePendingReplayRetry();
        return;
      }

      if (settlement.outcome === "halted") {
        setRefreshError(settlement.halt.message);
      } else {
        setRefreshError(result.failure.message);
      }
      const message = result.failure.message;
      settleCompletionCallbacks(
        dispatchedUnit.id,
        result.failure.kind === "validation"
          ? { kind: "validation_error", message }
          : result.failure.kind === "stale_target"
            ? { kind: "stale_target", message }
            : result.failure.kind === "same_field_conflict" ||
                result.failure.kind === "client_txn_conflict"
              ? { kind: "conflict", message }
              : { kind: "rejected_mutation", message },
      );
      publishPendingQueueState();
      return;
    }

    recordWorkbookTiming("pending_result_apply_start", {
      clientTxnId: dispatchedUnit.clientTxnId,
      kind: dispatchedUnit.kind,
      rowKey: dispatchedUnit.rowKey,
    });
    let appliedRow: { record_id: string; row_version: number };
    try {
      appliedRow = {
        record_id: result.value.row.record_id,
        row_version: result.value.row.row_version,
      };
      clearSubmittedScalarEditorDraftValuesForRow(
        dispatchedUnit.rowKey,
        meta.rowSnapshot.values,
      );
      const clearActiveCollectionFocusKey =
        meta.surface === "grid" && isCollectionDraftKey(meta.focusField)
          ? meta.focusKey
          : undefined;
      applyAcceptedRowMutation(dispatchedUnit.rowKey, result.value, {
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
        clientTxnId: dispatchedUnit.clientTxnId,
        kind: dispatchedUnit.kind,
        message: error instanceof Error ? error.message : String(error),
        rowKey: dispatchedUnit.rowKey,
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
      settleCompletionCallbacks(dispatchedUnit.id, {
        kind: "rejected_mutation",
        message:
          error instanceof Error
            ? error.message
            : "The accepted edit could not be applied.",
      });
      publishPendingQueueState();
      return;
    }
    if (postMutationQueryRefreshRequired) {
      try {
        await loadRowsRef.current({ showLoading: false });
      } catch {
        setRefreshError("Timeline projection refresh failed.");
      }
    }
    recordWorkbookTiming("pending_result_apply_end", {
      clientTxnId: dispatchedUnit.clientTxnId,
      kind: dispatchedUnit.kind,
      recordId: appliedRow.record_id,
      rowKey: dispatchedUnit.rowKey,
      rowVersion: appliedRow.row_version,
    });
    const successResult = {
      ok: true,
      row: appliedRow,
      change_set_id: result.value.changeSetId,
    } as const;
    const settlement = pending.model.settleDispatched(successResult);
    if (settlement.outcome === "success") {
      pending.metaByUnitId.delete(settlement.unit.id);
      clearPendingSignatureForUnit(settlement.unit);
      settleCompletionCallbacks(settlement.unit.id, { kind: "accepted" });
    }
    publishPendingQueueState();
    requestPendingReplay("unit_completed");
  };

  return {
    discardBlockedEdit,
    enqueuePendingReplayUnit,
    retryBlockedEdit,
    schedulePendingReplay,
  };
}
