import { workbookInspectorSubjectChange } from "../inspector/workbookInspectorSubject";
import type { WorkbookRecordSubject } from "../ports/WorkbookRecordSubject";

export type WorkbookInspectorInvalidationReason =
  | "action_completed"
  | "authorization_lost"
  | "hard_refresh"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "surface_changed";

type WorkbookInspectorStateContext = {
  readonly reviewGeneration: number;
  readonly attachmentGeneration: number;
  readonly invalidationCause:
    | WorkbookInspectorInvalidationReason
    | "close"
    | "retarget"
    | "record_updated"
    | null;
  readonly lifecycleKey: string;
};

export type WorkbookInspectorState = WorkbookInspectorStateContext &
  (
    | {
        readonly phase: "closed";
        readonly subject: WorkbookRecordSubject | null;
      }
    | { readonly phase: "open_no_subject"; readonly subject: null }
    | {
        readonly phase: "open_ready";
        readonly subject: WorkbookRecordSubject;
      }
  );

export type WorkbookInspectorAction =
  | {
      readonly lifecycleKey: string;
      readonly subject: WorkbookRecordSubject | null;
      readonly type: "open";
    }
  | { readonly lifecycleKey: string; readonly type: "close" }
  | {
      readonly lifecycleKey: string;
      readonly type: "retarget";
      readonly subject: WorkbookRecordSubject | null;
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
    reviewGeneration: 0,
    attachmentGeneration: 0,
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
      reviewGeneration: state.reviewGeneration + 1,
      attachmentGeneration: state.attachmentGeneration + 1,
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
      const change = workbookInspectorSubjectChange(
        state.subject,
        action.subject,
      );
      const subject = change === null ? state.subject : action.subject;
      if (state.phase !== "closed" && change === null) return state;
      const next = {
        ...state,
        reviewGeneration: state.reviewGeneration + (change === null ? 0 : 1),
        attachmentGeneration:
          state.attachmentGeneration +
          (state.phase === "closed" || change === "retarget" ? 1 : 0),
        invalidationCause: change ?? state.invalidationCause,
      };
      return subject === null
        ? { ...next, phase: "open_no_subject", subject }
        : { ...next, phase: "open_ready", subject };
    }
    case "close":
      if (state.phase === "closed") {
        return state;
      }
      return {
        ...state,
        invalidationCause: "close",
        reviewGeneration: state.reviewGeneration + 1,
        attachmentGeneration: state.attachmentGeneration + 1,
        phase: "closed",
      };
    case "retarget": {
      const change = workbookInspectorSubjectChange(
        state.subject,
        action.subject,
      );
      if (change === null) return state;
      const next = {
        ...state,
        invalidationCause: change,
        reviewGeneration: state.reviewGeneration + 1,
        attachmentGeneration:
          state.attachmentGeneration + (change === "retarget" ? 1 : 0),
      };
      if (state.phase === "closed")
        return { ...next, phase: "closed", subject: action.subject };
      return action.subject === null
        ? {
            ...next,
            phase: "open_no_subject",
            subject: null,
          }
        : {
            ...next,
            phase: "open_ready",
            subject: action.subject,
          };
    }
    case "invalidate":
      return action.reason === "action_completed"
        ? {
            ...state,
            invalidationCause: action.reason,
            reviewGeneration: state.reviewGeneration + 1,
          }
        : {
            ...state,
            invalidationCause: action.reason,
            reviewGeneration: state.reviewGeneration + 1,
            phase: "closed",
            attachmentGeneration: state.attachmentGeneration + 1,
            subject: null,
          };
  }
}
