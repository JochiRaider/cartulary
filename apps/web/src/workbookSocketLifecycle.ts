import type { RecordChangedPayload } from "./workbookShellPhase4";

export type WorkbookSocketLifecycleState = {
  readonly appliedStreamSeqs: ReadonlySet<number>;
  readonly authPaused: boolean;
  readonly connectionId: string | null;
  readonly established: boolean;
  readonly lastSeenStreamSeq: number;
  readonly reconnectSuppressed: boolean;
  readonly resumeToken: string | null;
};

export type WorkbookSocketAckMessageType = "hello_ack" | "resume_ack";

export type WorkbookSocketRefreshReason =
  | "record_change_requery"
  | "reset_required"
  | "sequence_gap";

export type WorkbookSocketLifecycleAction =
  | { readonly type: "socket_connecting" }
  | { readonly type: "auth_recovered" }
  | {
      readonly type: "session_ack";
      readonly messageType: WorkbookSocketAckMessageType;
      readonly payload?: Record<string, unknown>;
    }
  | { readonly type: "session_revoked" }
  | { readonly type: "authorization_closed" }
  | {
      readonly type: "record_changed_received";
      readonly message: {
        readonly stream_seq?: number;
        readonly payload: RecordChangedPayload;
      };
    }
  | {
      readonly type: "record_change_result";
      readonly applied: boolean;
    };

export type WorkbookSocketLifecycleEffect =
  | {
      readonly kind: "apply_record_change";
      readonly payload: RecordChangedPayload;
    }
  | { readonly kind: "close_socket" }
  | { readonly kind: "ignore_duplicate_sequence"; readonly streamSeq: number }
  | { readonly kind: "pause_for_auth_recovery" }
  | {
      readonly kind: "request_refresh";
      readonly reason: WorkbookSocketRefreshReason;
      readonly streamSeq?: number;
    }
  | { readonly kind: "resume_pending_replay" }
  | { readonly kind: "schedule_auth_recovery_probe" }
  | { readonly kind: "suppress_reconnect" };

export type WorkbookSocketLifecycleReduction = {
  readonly state: WorkbookSocketLifecycleState;
  readonly effects: WorkbookSocketLifecycleEffect[];
};

export function createWorkbookSocketLifecycleState(
  overrides: Partial<WorkbookSocketLifecycleState> = {},
): WorkbookSocketLifecycleState {
  return {
    appliedStreamSeqs: new Set<number>(),
    authPaused: false,
    connectionId: null,
    established: false,
    lastSeenStreamSeq: 0,
    reconnectSuppressed: false,
    resumeToken: null,
    ...overrides,
  };
}

export function reduceWorkbookSocketLifecycle(
  state: WorkbookSocketLifecycleState,
  action: WorkbookSocketLifecycleAction,
): WorkbookSocketLifecycleReduction {
  switch (action.type) {
    case "socket_connecting":
      return {
        state: {
          ...state,
          established: false,
        },
        effects: [],
      };

    case "auth_recovered":
      return {
        state: {
          ...state,
          authPaused: false,
          reconnectSuppressed: false,
        },
        effects: [],
      };

    case "session_ack":
      return reduceSessionAck(state, action);

    case "session_revoked":
    case "authorization_closed":
      return {
        state: {
          ...state,
          authPaused: true,
          established: false,
          reconnectSuppressed: true,
          resumeToken: null,
        },
        effects: [
          { kind: "pause_for_auth_recovery" },
          { kind: "suppress_reconnect" },
          { kind: "schedule_auth_recovery_probe" },
          { kind: "close_socket" },
        ],
      };

    case "record_changed_received":
      return reduceRecordChangedReceived(state, action.message);

    case "record_change_result":
      if (action.applied) {
        return { state, effects: [] };
      }
      return {
        state,
        effects: [{ kind: "request_refresh", reason: "record_change_requery" }],
      };
  }
}

function reduceSessionAck(
  state: WorkbookSocketLifecycleState,
  action: Extract<WorkbookSocketLifecycleAction, { type: "session_ack" }>,
): WorkbookSocketLifecycleReduction {
  const connectionId = stringPayloadValue(action.payload, "connection_id");
  const resumeToken = stringPayloadValue(action.payload, "resume_token");
  const resetRequired =
    action.messageType === "resume_ack" &&
    action.payload?.status === "reset_required";

  const nextState = {
    ...state,
    authPaused: false,
    connectionId: connectionId ?? state.connectionId,
    established: true,
    reconnectSuppressed: false,
    resumeToken: resumeToken ?? state.resumeToken,
  };

  if (resetRequired) {
    return {
      state: nextState,
      effects: [{ kind: "request_refresh", reason: "reset_required" }],
    };
  }

  return {
    state: nextState,
    effects: [{ kind: "resume_pending_replay" }],
  };
}

function reduceRecordChangedReceived(
  state: WorkbookSocketLifecycleState,
  message: Extract<
    WorkbookSocketLifecycleAction,
    { type: "record_changed_received" }
  >["message"],
): WorkbookSocketLifecycleReduction {
  const streamSeq = message.stream_seq;
  if (typeof streamSeq !== "number") {
    return {
      state,
      effects: [{ kind: "apply_record_change", payload: message.payload }],
    };
  }

  if (state.appliedStreamSeqs.has(streamSeq)) {
    return {
      state,
      effects: [{ kind: "ignore_duplicate_sequence", streamSeq }],
    };
  }

  const appliedStreamSeqs = new Set(state.appliedStreamSeqs);
  appliedStreamSeqs.add(streamSeq);
  const nextState = {
    ...state,
    appliedStreamSeqs,
    lastSeenStreamSeq: Math.max(state.lastSeenStreamSeq, streamSeq),
  };

  if (state.lastSeenStreamSeq > 0 && streamSeq > state.lastSeenStreamSeq + 1) {
    return {
      state: nextState,
      effects: [{ kind: "request_refresh", reason: "sequence_gap", streamSeq }],
    };
  }

  return {
    state: nextState,
    effects: [{ kind: "apply_record_change", payload: message.payload }],
  };
}

function stringPayloadValue(
  payload: Record<string, unknown> | undefined,
  key: string,
) {
  const value = payload?.[key];
  return typeof value === "string" ? value : null;
}
