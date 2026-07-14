import type { RecordChangedPayload } from "./workbookCollaborationMessages";

export type TimelineCollaborationState = {
  readonly authPaused: boolean;
};

export type TimelineCollaborationAction =
  | { readonly type: "session_established" }
  | { readonly type: "authorization_lost" }
  | {
      readonly type: "record_changed_received";
      readonly payload: RecordChangedPayload;
    }
  | {
      readonly type: "record_change_result";
      readonly applied: boolean;
    };

export type TimelineCollaborationEffect =
  | {
      readonly kind: "apply_record_change";
      readonly payload: RecordChangedPayload;
    }
  | { readonly kind: "pause_for_auth_recovery" }
  | { readonly kind: "request_record_refresh" }
  | { readonly kind: "resume_pending_replay" }
  | { readonly kind: "schedule_auth_recovery_probe" };

export type TimelineCollaborationReduction = {
  readonly state: TimelineCollaborationState;
  readonly effects: readonly TimelineCollaborationEffect[];
};

export function createTimelineCollaborationState(
  overrides: Partial<TimelineCollaborationState> = {},
): TimelineCollaborationState {
  return {
    authPaused: false,
    ...overrides,
  };
}

export function reduceTimelineCollaboration(
  state: TimelineCollaborationState,
  action: TimelineCollaborationAction,
): TimelineCollaborationReduction {
  switch (action.type) {
    case "session_established":
      return {
        state: { authPaused: false },
        effects: [{ kind: "resume_pending_replay" }],
      };

    case "authorization_lost":
      return {
        state: { authPaused: true },
        effects: [
          { kind: "pause_for_auth_recovery" },
          { kind: "schedule_auth_recovery_probe" },
        ],
      };

    case "record_changed_received":
      return {
        state,
        effects: [{ kind: "apply_record_change", payload: action.payload }],
      };

    case "record_change_result":
      return action.applied
        ? { state, effects: [] }
        : { state, effects: [{ kind: "request_record_refresh" }] };
  }
}
