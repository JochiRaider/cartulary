import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import { useCallback, useRef } from "react";
import { apiPath, clientTxnID } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  parsePendingReplayPublicError,
} from "../../utils/workbookPendingQueue";
import type { PendingReplayRuntimeMeta } from "../models/timelineControllerPorts";
import {
  refreshBlocksTimelinePendingUnit,
  type TimelinePendingReplayAdmissionRequest,
  type TimelinePendingSavesRefs,
} from "../models/timelinePendingReplayModel";
import {
  type FocusFieldKey,
  materializePendingReplayPayload,
  normalizeTimelineFullRow,
  type RowValues,
  type TimelineScalarEditorSurface,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import {
  dispatchTimelinePendingReplayMutation,
  type TimelineMutationEnvelope,
} from "../services/timelineMutationRequests";

type SessionEnvelope = {
  data: {
    user_id: string;
    memberships: Array<{
      incident_id: string;
      role: string;
    }>;
  };
};

type TimelineMutableRef<T> = {
  current: T;
};

type TimelinePendingReplayControllerAdmission =
  TimelinePendingReplayAdmissionRequest<PendingReplayRuntimeMeta>;

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
  apiBase,
  applyRowMutation,
  clearSubmittedScalarEditorDraftValuesForRow,
  clearViewportContinuity,
  conflictQueueRef,
  handleMutationConflict,
  latestCommittedTimelineRow,
  pendingSavesRefsRef,
  publishPendingQueueState,
  reconcileDiscardedPendingUnit,
  recordWorkbookTiming,
  resolvePendingSocketTxn,
  rowsRef,
  scheduleSocketReconnectAfterAuthRef,
  setRefreshError,
  setRows,
  trackPendingSocketTxn,
}: {
  readonly apiBase?: string | undefined;
  readonly applyRowMutation: (
    rowKey: string,
    envelope: TimelineMutationEnvelope,
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
  readonly handleMutationConflict: (
    payload: unknown,
    rowKey: string,
    focusField: FocusFieldKey,
    surface: TimelineScalarEditorSurface,
  ) => boolean;
  readonly latestCommittedTimelineRow: (recordId: string) => WorkbookRow | null;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
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
  readonly scheduleSocketReconnectAfterAuthRef: TimelineMutableRef<
    (() => void) | null
  >;
  readonly setRefreshError: (message: string | null) => void;
  readonly setRows: (rows: WorkbookRow[]) => void;
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

  const scheduleAuthRecoveryProbe = useCallback(() => {
    if (
      pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current !== null
    ) {
      return;
    }
    pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current =
      window.setTimeout(async () => {
        pendingSavesRefsRef.current.pendingReplayAuthRetryRef.current = null;
        if (
          !pendingSavesRefsRef.current.pendingQueueRef.current.model.snapshot()
            .authPaused
        ) {
          return;
        }
        try {
          const result = await fetchWorkbookJSON<SessionEnvelope>(
            apiPath(apiBase, "/api/v1/auth/session"),
          );
          if (!result.ok) {
            scheduleAuthRecoveryProbe();
            return;
          }
          pendingSavesRefsRef.current.pendingQueueRef.current.model.resumeAfterAuthRecovery();
          publishPendingQueueState();
          schedulePendingReplay();
          scheduleSocketReconnectAfterAuthRef.current?.();
        } catch {
          scheduleAuthRecoveryProbe();
        }
      }, 1000);
  }, [
    apiBase,
    pendingSavesRefsRef,
    publishPendingQueueState,
    schedulePendingReplay,
    scheduleSocketReconnectAfterAuthRef,
  ]);
  const scheduleAuthRecoveryProbeRef = useRef(scheduleAuthRecoveryProbe);
  scheduleAuthRecoveryProbeRef.current = scheduleAuthRecoveryProbe;

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
        !refreshBlocksTimelinePendingUnit(pending, candidate.unit) &&
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
        replacementClientTxnId = clientTxnID("timeline-client");
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
    if (refreshBlocksTimelinePendingUnit(pending, unit)) {
      publishPendingQueueState();
      return;
    }
    const meta = pending.metaByUnitId.get(unit.id);
    if (meta === undefined) {
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
    const dispatchPayload = materializePendingReplayPayload(unit, currentRow);
    if (dispatchPayload === null) {
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

    let result = null as Awaited<
      ReturnType<typeof dispatchTimelinePendingReplayMutation>
    > | null;
    try {
      result = await dispatchTimelinePendingReplayMutation({
        payload: dispatchPayload,
        recordTiming: recordWorkbookTiming,
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

    if (!result.ok) {
      resolvePendingSocketTxn(dispatchedUnit.clientTxnId);
      const publicError = parsePendingReplayPublicError(result.payload);
      const settlement = pending.model.settleDispatched({
        ok: false,
        status: result.status,
        error: publicError,
      });
      if (settlement.outcome === "auth_paused") {
        setRefreshError(
          "Authentication required before queued edits can replay.",
        );
        publishPendingQueueState();
        settleCompletionCallbacks(dispatchedUnit.id, { kind: "accepted" });
        scheduleAuthRecoveryProbe();
        return;
      }

      if (settlement.outcome === "same_field_conflict") {
        const message =
          publicError.message ?? parseErrorMessage(result.payload);
        settleCompletionCallbacks(dispatchedUnit.id, {
          kind: "conflict",
          message,
        });
        if (
          handleMutationConflict(
            result.payload,
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
        setRefreshError(
          publicError.message ?? parseErrorMessage(result.payload),
        );
        publishPendingQueueState();
        return;
      }

      if (settlement.outcome === "retryable_failure") {
        publishPendingQueueState();
        schedulePendingReplayRetry();
        return;
      }

      if (settlement.outcome === "halted") {
        setRefreshError(settlement.halt.message);
      } else {
        setRefreshError(parseErrorMessage(result.payload));
      }
      const message = publicError.message ?? parseErrorMessage(result.payload);
      settleCompletionCallbacks(
        dispatchedUnit.id,
        result.status === 400
          ? { kind: "validation_error", message }
          : result.status === 404
            ? { kind: "stale_target", message }
            : result.status === 409
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
    let envelope: TimelineMutationEnvelope;
    let appliedRow: { record_id: string; row_version: number };
    try {
      envelope = readEnvelope<TimelineMutationEnvelope>(result.payload);
      const responseRow = normalizeTimelineFullRow(
        envelope.data.row,
        "pending mutation response row",
      );
      appliedRow = {
        record_id: responseRow.record_id,
        row_version: responseRow.row_version,
      };
      clearSubmittedScalarEditorDraftValuesForRow(
        dispatchedUnit.rowKey,
        meta.rowSnapshot.values,
      );
      const clearActiveCollectionFocusKey =
        meta.surface === "grid" && isCollectionDraftKey(meta.focusField)
          ? meta.focusKey
          : undefined;
      applyRowMutation(dispatchedUnit.rowKey, envelope, {
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
      ...(envelope.data.change_set_id === undefined
        ? {}
        : { change_set_id: envelope.data.change_set_id }),
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
    scheduleAuthRecoveryProbeRef,
    schedulePendingReplay,
  };
}
