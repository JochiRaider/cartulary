import type { WorkbookInspectorSubject } from "../inspector/workbookInspectorSubject";
import { workbookInspectorSubjectsEqual } from "../inspector/workbookInspectorSubject";

export type WorkbookInspectorInvalidationReason =
  | "action_completed"
  | "authorization_lost"
  | "hard_refresh"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "surface_changed";

type WorkbookInspectorStateContext = {
  readonly invalidationGeneration: number;
  readonly invalidationCause:
    | WorkbookInspectorInvalidationReason
    | "close"
    | "retarget"
    | null;
  readonly lifecycleKey: string;
};

export type WorkbookInspectorState = WorkbookInspectorStateContext &
  (
    | {
        readonly phase: "closed";
        readonly subject: WorkbookInspectorSubject | null;
      }
    | { readonly phase: "open_no_subject"; readonly subject: null }
    | {
        readonly phase: "open_ready";
        readonly subject: WorkbookInspectorSubject;
      }
  );

export type WorkbookInspectorAction =
  | {
      readonly lifecycleKey: string;
      readonly subject: WorkbookInspectorSubject | null;
      readonly type: "open";
    }
  | { readonly lifecycleKey: string; readonly type: "close" }
  | {
      readonly lifecycleKey: string;
      readonly type: "retarget";
      readonly subject: WorkbookInspectorSubject | null;
    }
  | {
      readonly lifecycleKey: string;
      readonly type: "invalidate";
      readonly reason: WorkbookInspectorInvalidationReason;
    }
  | { readonly lifecycleKey: string; readonly type: "lifecycle_changed" };

export function initialWorkbookInspectorState({
  lifecycleKey,
}: {
  readonly lifecycleKey: string;
}): WorkbookInspectorState {
  return {
    invalidationGeneration: 0,
    invalidationCause: null,
    lifecycleKey,
    phase: "closed",
    subject: null,
  };
}

export function workbookInspectorStateIsOpen(
  state: WorkbookInspectorState,
): boolean {
  return state.phase !== "closed";
}

export function workbookInspectorReducer(
  state: WorkbookInspectorState,
  action: WorkbookInspectorAction,
): WorkbookInspectorState {
  if (action.type === "lifecycle_changed") {
    if (action.lifecycleKey === state.lifecycleKey) {
      return state;
    }
    return {
      invalidationCause: "surface_changed",
      invalidationGeneration: state.invalidationGeneration + 1,
      lifecycleKey: action.lifecycleKey,
      phase: "closed",
      subject: null,
    };
  }
  if (action.lifecycleKey !== state.lifecycleKey) {
    return state;
  }
  switch (action.type) {
    case "open": {
      const subject = workbookInspectorSubjectsEqual(
        state.subject,
        action.subject,
      )
        ? state.subject
        : action.subject;
      if (state.phase !== "closed" && subject === state.subject) {
        return state;
      }
      return subject === null
        ? { ...state, phase: "open_no_subject", subject }
        : { ...state, phase: "open_ready", subject };
    }
    case "close":
      if (state.phase === "closed") {
        return state;
      }
      return {
        ...state,
        invalidationCause: "close",
        invalidationGeneration: state.invalidationGeneration + 1,
        phase: "closed",
      };
    case "retarget":
      if (workbookInspectorSubjectsEqual(state.subject, action.subject)) {
        return state;
      }
      if (state.phase === "closed") {
        return {
          ...state,
          invalidationCause: "retarget",
          invalidationGeneration: state.invalidationGeneration + 1,
          subject: action.subject,
        };
      }
      return action.subject === null
        ? {
            ...state,
            invalidationCause: "retarget",
            invalidationGeneration: state.invalidationGeneration + 1,
            phase: "open_no_subject",
            subject: null,
          }
        : {
            ...state,
            invalidationCause: "retarget",
            invalidationGeneration: state.invalidationGeneration + 1,
            phase: "open_ready",
            subject: action.subject,
          };
    case "invalidate":
      return action.reason === "action_completed"
        ? {
            ...state,
            invalidationCause: action.reason,
            invalidationGeneration: state.invalidationGeneration + 1,
          }
        : {
            ...state,
            invalidationCause: action.reason,
            invalidationGeneration: state.invalidationGeneration + 1,
            phase: "closed",
            subject: null,
          };
  }
}
