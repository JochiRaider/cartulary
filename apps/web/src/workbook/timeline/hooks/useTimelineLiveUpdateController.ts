import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
} from "react";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import {
  isPresenceRecord,
  type PresenceRecord,
} from "../../utils/workbookPresence";
import {
  buildWorkbookPresenceUpdateMessage,
  buildWorkbookSocketSessionMessage,
  isRecordChangedMessage,
  type RecordChangedPayload,
  shouldIgnoreSelfOriginatedRecordChange,
  type TimelinePresenceDraft,
} from "../services/workbookCollaborationMessages";
import type { WorkbookSocketLifecycleEffect } from "../services/workbookSocketLifecycle";
import type { TimelineLiveUpdateRefs } from "./useTimelineLiveUpdates";
import type { PendingReplayRuntimeMeta } from "./useTimelinePendingReplayController";
import {
  ensureTimelineTabClientInstanceId,
  type TimelinePendingRefreshBlockScope,
  type TimelinePendingSavesRefs,
} from "./useTimelinePendingSaves";

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

function socketIsOpen(socket: WebSocket) {
  return socket.readyState === WebSocket.OPEN;
}

function websocketPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}${path}`;
  }

  const target = new URL(trimmedBase);
  target.protocol = target.protocol === "https:" ? "wss:" : "ws:";
  target.pathname = path;
  target.search = "";
  target.hash = "";
  return target.toString();
}

export function useTimelineLiveUpdateController({
  activeSheetRuntimeRef,
  advanceViewportContinuityRef,
  apiBase,
  applyRecordChangedPatchRef,
  beginRefreshInFlightRef,
  beginViewportContinuityRef,
  currentPresence,
  finishRefreshInFlightRef,
  incidentId,
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
  const {
    activeSocketRef,
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketEstablishedRef,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
    socketResumeTokenRef,
  } = liveUpdateRefs;

  const sendPresenceUpdate = useCallback(
    (presence: TimelinePresenceDraft) => {
      const target = activeSocketRef.current;
      if (
        target === null ||
        !socketEstablishedRef.current ||
        !socketIsOpen(target)
      ) {
        return;
      }
      target.send(
        JSON.stringify(
          buildWorkbookPresenceUpdateMessage(
            presence,
            activeSheetRuntimeRef.current,
          ),
        ),
      );
    },
    [activeSheetRuntimeRef, activeSocketRef, socketEstablishedRef],
  );

  useEffect(() => {
    currentPresenceRef.current = currentPresence;
    const target = activeSocketRef.current;
    if (
      target === null ||
      !socketEstablishedRef.current ||
      !socketIsOpen(target)
    ) {
      return;
    }
    if (presenceUpdateTimerRef.current !== null) {
      window.clearTimeout(presenceUpdateTimerRef.current);
    }
    presenceUpdateTimerRef.current = window.setTimeout(() => {
      presenceUpdateTimerRef.current = null;
      sendPresenceUpdate(currentPresenceRef.current);
    }, 150);
  }, [
    activeSocketRef,
    currentPresence,
    currentPresenceRef,
    presenceUpdateTimerRef,
    sendPresenceUpdate,
    socketEstablishedRef,
  ]);

  useEffect(() => {
    if (incidentId.trim() === "") {
      return;
    }

    let closed = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    const clientInstanceId = ensureTimelineTabClientInstanceId(
      pendingSavesRefsRef.current.socketClientInstanceIdRef,
    );
    const changeSocketURL = websocketPath(
      apiBase,
      `/ws/v1/incidents/${incidentId}`,
    );

    const scheduleReconnect = () => {
      if (
        closed ||
        reconnectTimer !== null ||
        socketLifecycleRef.current.reconnectSuppressed
      ) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 1000);
    };
    socketReconnectAfterAuthRef.current = scheduleReconnect;

    const sendSessionEstablishment = (target: WebSocket) => {
      target.send(
        JSON.stringify(
          buildWorkbookSocketSessionMessage({
            clientInstanceId,
            lastSeenStreamSeq: socketLastSeenStreamSeqRef.current,
            presence: currentPresenceRef.current,
            resumeToken: socketResumeTokenRef.current,
            sheetRef: activeSheetRuntimeRef.current,
          }),
        ),
      );
    };

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
      target: WebSocket,
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
          case "close_socket":
            target.close();
            break;
          case "request_refresh":
            if (
              effect.reason === "reset_required" ||
              effect.reason === "sequence_gap"
            ) {
              requestSocketLifecycleRefresh();
            }
            break;
          case "resume_pending_replay":
            pendingSavesRefsRef.current.pendingQueueRef.current.model.resumeAfterAuthRecovery();
            publishPendingQueueStateRef.current();
            pendingSavesRefsRef.current.schedulePendingReplayRef.current();
            break;
          case "apply_record_change":
          case "ignore_duplicate_sequence":
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
      if (!presence || typeof presence !== "object") {
        return;
      }
      const candidate = presence as Record<string, unknown>;
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
        const withoutExisting = current.filter(
          (record) => record.connection_id !== nextRecord.connection_id,
        );
        return [...withoutExisting, nextRecord];
      });
    };

    const handleMessage = (target: WebSocket, raw: unknown) => {
      if (!raw || typeof raw !== "object") {
        return;
      }
      const message = raw as {
        type?: string;
        stream_seq?: number;
        payload?: Record<string, unknown>;
      };
      if (message.type === "ping") {
        target.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (message.type === "hello_ack" || message.type === "resume_ack") {
        const effects = dispatchSocketLifecycleRef.current({
          type: "session_ack",
          messageType: message.type,
          ...(message.payload === undefined
            ? {}
            : { payload: message.payload }),
        });
        applySocketLifecycleEffects(effects, target);
        return;
      }
      if (message.type === "presence_snapshot") {
        applyPresenceSnapshot(message.payload ?? {});
        return;
      }
      if (message.type === "presence_delta") {
        applyPresenceDelta(message.payload ?? {});
        return;
      }
      if (message.type === "session_revoked") {
        const effects = dispatchSocketLifecycleRef.current({
          type: "session_revoked",
        });
        applySocketLifecycleEffects(effects, target);
        return;
      }
      if (
        shouldIgnoreSelfOriginatedRecordChange(
          raw,
          resolvePendingSocketTxnRef.current,
        )
      ) {
        return;
      }
      if (!isRecordChangedMessage(raw)) {
        return;
      }
      const recordChangedPayload = message.payload as RecordChangedPayload;
      const streamEffects = dispatchSocketLifecycleRef.current({
        type: "record_changed_received",
        message: {
          ...(typeof message.stream_seq === "number"
            ? { stream_seq: message.stream_seq }
            : {}),
          payload: recordChangedPayload,
        },
      });
      for (const effect of streamEffects) {
        if (effect.kind === "ignore_duplicate_sequence") {
          return;
        }
        if (
          effect.kind === "request_refresh" &&
          effect.reason === "sequence_gap"
        ) {
          applySocketLifecycleEffects([effect], target);
          return;
        }
      }
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
          {
            viewportContinuityToken,
          },
          { kind: "record", recordId: recordChangedPayload.record_id },
        );
      }
    };

    const connect = () => {
      if (closed) {
        return;
      }
      socket = new WebSocket(changeSocketURL);
      activeSocketRef.current = socket;
      dispatchSocketLifecycleRef.current({ type: "socket_connecting" });
      socket.onopen = () => {
        if (socket) {
          sendSessionEstablishment(socket);
        }
      };
      socket.onmessage = (event) => {
        if (!socket) {
          return;
        }
        handleMessage(socket, JSON.parse(event.data) as unknown);
      };
      socket.onclose = (event) => {
        if (
          event.code === 1008 &&
          (event.reason === "session_revoked" ||
            event.reason === "authorization_denied")
        ) {
          const effects = dispatchSocketLifecycleRef
            .current({
              type: "authorization_closed",
            })
            .filter((effect) => effect.kind !== "close_socket");
          applySocketLifecycleEffects(effects, socket as WebSocket);
        }
        scheduleReconnect();
      };
      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();

    return () => {
      closed = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      if (presenceUpdateTimerRef.current !== null) {
        window.clearTimeout(presenceUpdateTimerRef.current);
        presenceUpdateTimerRef.current = null;
      }
      activeSocketRef.current = null;
      socketReconnectAfterAuthRef.current = null;
      dispatchSocketLifecycleRef.current({ type: "socket_connecting" });
      socket?.close();
    };
  }, [
    activeSheetRuntimeRef,
    activeSocketRef,
    advanceViewportContinuityRef,
    apiBase,
    applyRecordChangedPatchRef,
    beginRefreshInFlightRef,
    beginViewportContinuityRef,
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    finishRefreshInFlightRef,
    incidentId,
    loadRowsRef,
    pendingSavesRefsRef,
    publishPendingQueueStateRef,
    resolvePendingSocketTxnRef,
    scheduleAuthRecoveryProbeRef,
    setPresenceRecords,
    setRefreshError,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
    presenceUpdateTimerRef,
    socketResumeTokenRef,
  ]);

  return {
    sendPresenceUpdate,
  };
}
