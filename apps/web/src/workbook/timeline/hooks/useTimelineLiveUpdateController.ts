import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
} from "react";
import { useIncidentCollaborationSession } from "../../../collaboration/IncidentCollaborationSession";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import {
  isPresenceRecord,
  type PresenceRecord,
} from "../../utils/workbookPresence";
import type {
  PendingReplayRuntimeMeta,
  TimelineLiveUpdateRefs,
} from "../models/timelineControllerPorts";
import type {
  TimelinePendingRefreshBlockScope,
  TimelinePendingSavesRefs,
} from "../models/timelinePendingReplayModel";
import {
  buildWorkbookPresenceInput,
  isRecordChangedMessage,
  type RecordChangedPayload,
  shouldIgnoreSelfOriginatedRecordChange,
  type TimelinePresenceDraft,
} from "../services/workbookCollaborationMessages";
import type { WorkbookSocketLifecycleEffect } from "../services/workbookSocketLifecycle";

type TimelineMutableRef<T> = {
  current: T;
};

type TimelineLiveUpdateLoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
  viewportContinuityToken?: number;
};

type TimelineLiveUpdateViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

function recordValue(value: unknown): Record<string, unknown> {
  if (value !== null && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

export function useTimelineLiveUpdateController({
  activeSheetRuntimeRef,
  advanceViewportContinuityRef,
  apiBase: _apiBase,
  applyRecordChangedPatchRef,
  beginRefreshInFlightRef,
  beginViewportContinuityRef,
  currentPresence,
  finishRefreshInFlightRef,
  incidentId: _incidentId,
  liveUpdateRefs,
  loadRowsRef,
  pendingSavesRefsRef,
  publishPendingQueueStateRef,
  resolvePendingSocketTxnRef,
  scheduleAuthRecoveryProbeRef,
  setPresenceRecords,
  setRefreshError,
}: {
  readonly activeSheetRuntimeRef: TimelineMutableRef<WorkbookSheetRef>;
  readonly advanceViewportContinuityRef: TimelineMutableRef<
    (token: number | undefined) => void
  >;
  readonly apiBase?: string | undefined;
  readonly applyRecordChangedPatchRef: TimelineMutableRef<
    (payload: RecordChangedPayload) => boolean
  >;
  readonly beginRefreshInFlightRef: TimelineMutableRef<
    (scope: TimelinePendingRefreshBlockScope) => void
  >;
  readonly beginViewportContinuityRef: TimelineMutableRef<
    (target: TimelineLiveUpdateViewportContinuityTarget) => number
  >;
  readonly currentPresence: TimelinePresenceDraft;
  readonly finishRefreshInFlightRef: TimelineMutableRef<
    (scope: TimelinePendingRefreshBlockScope) => void
  >;
  readonly incidentId: string;
  readonly liveUpdateRefs: TimelineLiveUpdateRefs;
  readonly loadRowsRef: TimelineMutableRef<
    (options: TimelineLiveUpdateLoadRowsOptions) => Promise<void>
  >;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<PendingReplayRuntimeMeta>
  >;
  readonly publishPendingQueueStateRef: TimelineMutableRef<() => void>;
  readonly resolvePendingSocketTxnRef: TimelineMutableRef<
    (clientTxnId: string | null | undefined) => boolean
  >;
  readonly scheduleAuthRecoveryProbeRef: TimelineMutableRef<() => void>;
  readonly setPresenceRecords: Dispatch<SetStateAction<PresenceRecord[]>>;
  readonly setRefreshError: (message: string | null) => void;
}) {
  const { disconnect, publishPresence, reconnect, subscribe } =
    useIncidentCollaborationSession();
  const {
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketReconnectAfterAuthRef,
  } = liveUpdateRefs;

  const sendPresenceUpdate = useCallback(
    (presence: TimelinePresenceDraft) => {
      publishPresence(
        buildWorkbookPresenceInput(presence, activeSheetRuntimeRef.current),
      );
    },
    [activeSheetRuntimeRef, publishPresence],
  );

  useEffect(() => {
    currentPresenceRef.current = currentPresence;
    if (presenceUpdateTimerRef.current !== null) {
      window.clearTimeout(presenceUpdateTimerRef.current);
    }
    presenceUpdateTimerRef.current = window.setTimeout(() => {
      presenceUpdateTimerRef.current = null;
      sendPresenceUpdate(currentPresenceRef.current);
    }, 150);
  }, [
    currentPresence,
    currentPresenceRef,
    presenceUpdateTimerRef,
    sendPresenceUpdate,
  ]);

  useEffect(() => {
    socketReconnectAfterAuthRef.current = reconnect;

    const requestSocketLifecycleRefresh = (
      options: Omit<TimelineLiveUpdateLoadRowsOptions, "showLoading"> = {},
      refreshScope: TimelinePendingRefreshBlockScope = { kind: "all" },
    ) => {
      beginRefreshInFlightRef.current(refreshScope);
      void loadRowsRef
        .current({ showLoading: false, ...options })
        .finally(() => {
          finishRefreshInFlightRef.current(refreshScope);
        });
    };

    const applySocketLifecycleEffects = (
      effects: readonly WorkbookSocketLifecycleEffect[],
    ) => {
      for (const effect of effects) {
        switch (effect.kind) {
          case "pause_for_auth_recovery":
            pendingSavesRefsRef.current.pendingQueueRef.current.model.pauseForAuthRecovery();
            setRefreshError(
              "Authentication required before queued edits can replay.",
            );
            publishPendingQueueStateRef.current();
            break;
          case "schedule_auth_recovery_probe":
            scheduleAuthRecoveryProbeRef.current();
            break;
          case "request_refresh":
            if (effect.reason === "reset_required") {
              requestSocketLifecycleRefresh();
            }
            break;
          case "resume_pending_replay":
            pendingSavesRefsRef.current.pendingQueueRef.current.model.resumeAfterAuthRecovery();
            publishPendingQueueStateRef.current();
            pendingSavesRefsRef.current.schedulePendingReplayRef.current();
            break;
          case "close_socket":
          case "apply_record_change":
          case "suppress_reconnect":
            break;
        }
      }
    };

    const applyPresenceSnapshot = (payload: Record<string, unknown>) => {
      const presences = Array.isArray(payload.presences)
        ? payload.presences
        : [];
      setPresenceRecords(
        presences.filter(isPresenceRecord).map((presence) => ({
          ...presence,
          sheet_ref: { ...presence.sheet_ref },
        })),
      );
    };

    const applyPresenceDelta = (payload: Record<string, unknown>) => {
      const deltaKind = payload.delta_kind;
      const presence = payload.presence;
      if (presence === null || typeof presence !== "object") {
        return;
      }
      const candidate = recordValue(presence);
      const connectionID = candidate.connection_id;
      if (typeof connectionID !== "string") {
        return;
      }
      setPresenceRecords((current) => {
        if (deltaKind === "remove") {
          return current.filter(
            (record) => record.connection_id !== connectionID,
          );
        }
        if (deltaKind !== "upsert" || !isPresenceRecord(candidate)) {
          return current;
        }
        const nextRecord = {
          ...candidate,
          sheet_ref: { ...candidate.sheet_ref },
        };
        return [
          ...current.filter(
            (record) => record.connection_id !== nextRecord.connection_id,
          ),
          nextRecord,
        ];
      });
    };

    const unsubscribe = subscribe((event) => {
      if (event.kind === "established") {
        const effects = dispatchSocketLifecycleRef.current({
          type: "session_ack",
          messageType: event.messageType,
          payload: { ...event.payload },
        });
        applySocketLifecycleEffects(effects);
        return;
      }
      if (event.kind === "reset_required") {
        if (event.reason === "sequence_gap") {
          requestSocketLifecycleRefresh();
        }
        return;
      }
      if (
        event.kind === "session_revoked" ||
        event.kind === "authorization_lost"
      ) {
        setPresenceRecords([]);
        const effects = dispatchSocketLifecycleRef.current({
          type:
            event.kind === "session_revoked"
              ? "session_revoked"
              : "authorization_closed",
        });
        applySocketLifecycleEffects(effects);
        return;
      }
      if (event.kind === "incident_closed") {
        setPresenceRecords([]);
        setRefreshError(
          "This incident is closed. Pending local work will not replay automatically.",
        );
        pendingSavesRefsRef.current.pendingQueueRef.current.model.pauseForTerminalLifecycle();
        publishPendingQueueStateRef.current();
        disconnect();
        return;
      }
      const message = event.message;
      if (message.type === "presence_snapshot") {
        applyPresenceSnapshot(recordValue(message.payload));
        return;
      }
      if (message.type === "presence_delta") {
        applyPresenceDelta(recordValue(message.payload));
        return;
      }
      if (
        shouldIgnoreSelfOriginatedRecordChange(
          message,
          resolvePendingSocketTxnRef.current,
        )
      ) {
        return;
      }
      if (!isRecordChangedMessage(message)) {
        return;
      }
      const recordChangedPayload = message.payload;
      dispatchSocketLifecycleRef.current({
        type: "record_changed_received",
        message: { payload: recordChangedPayload },
      });
      const viewportContinuityToken = beginViewportContinuityRef.current({
        kind: "scroll-only",
      });
      const applied = applyRecordChangedPatchRef.current(recordChangedPayload);
      const followupEffects = dispatchSocketLifecycleRef.current({
        type: "record_change_result",
        applied,
      });
      if (applied) {
        advanceViewportContinuityRef.current(viewportContinuityToken);
        return;
      }
      if (
        followupEffects.some(
          (effect) =>
            effect.kind === "request_refresh" &&
            effect.reason === "record_change_requery",
        )
      ) {
        requestSocketLifecycleRefresh(
          { viewportContinuityToken },
          { kind: "record", recordId: recordChangedPayload.record_id },
        );
      }
    });

    return () => {
      unsubscribe();
      if (presenceUpdateTimerRef.current !== null) {
        window.clearTimeout(presenceUpdateTimerRef.current);
        presenceUpdateTimerRef.current = null;
      }
      socketReconnectAfterAuthRef.current = null;
      dispatchSocketLifecycleRef.current({ type: "socket_connecting" });
    };
  }, [
    advanceViewportContinuityRef,
    applyRecordChangedPatchRef,
    beginRefreshInFlightRef,
    beginViewportContinuityRef,
    disconnect,
    dispatchSocketLifecycleRef,
    finishRefreshInFlightRef,
    loadRowsRef,
    pendingSavesRefsRef,
    presenceUpdateTimerRef,
    publishPendingQueueStateRef,
    reconnect,
    resolvePendingSocketTxnRef,
    scheduleAuthRecoveryProbeRef,
    setPresenceRecords,
    setRefreshError,
    socketReconnectAfterAuthRef,
    subscribe,
  ]);

  return { sendPresenceUpdate };
}
