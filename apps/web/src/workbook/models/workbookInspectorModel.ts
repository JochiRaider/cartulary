export type WorkbookInspectorSubject = {
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly rowVersion: number;
};

export type WorkbookInspectorStatus = "closed" | "no_row_selected" | "ready";

export type WorkbookInspectorInvalidationReason =
  | "action_completed"
  | "authorization_lost"
  | "hard_refresh"
  | "incident_closed"
  | "record_deleted"
  | "record_merged"
  | "surface_changed";

export type WorkbookInspectorState = {
  readonly invalidationGeneration: number;
  readonly invalidationCause:
    | WorkbookInspectorInvalidationReason
    | "close"
    | "retarget"
    | null;
  readonly isOpen: boolean;
  readonly status: WorkbookInspectorStatus;
  readonly subject: WorkbookInspectorSubject | null;
};

export type WorkbookInspectorAction =
  | { readonly type: "open" }
  | { readonly type: "close" }
  | {
      readonly type: "retarget";
      readonly subject: WorkbookInspectorSubject | null;
    }
  | {
      readonly type: "invalidate";
      readonly reason: WorkbookInspectorInvalidationReason;
    };

export function initialWorkbookInspectorState(): WorkbookInspectorState {
  return {
    invalidationGeneration: 0,
    invalidationCause: null,
    isOpen: false,
    status: "closed",
    subject: null,
  };
}

export function workbookInspectorReducer(
  state: WorkbookInspectorState,
  action: WorkbookInspectorAction,
): WorkbookInspectorState {
  switch (action.type) {
    case "open":
      return {
        ...state,
        isOpen: true,
        status: state.subject === null ? "no_row_selected" : "ready",
      };
    case "close":
      return {
        ...state,
        invalidationCause: "close",
        invalidationGeneration: state.invalidationGeneration + 1,
        isOpen: false,
        status: "closed",
      };
    case "retarget":
      if (workbookInspectorSubjectsEqual(state.subject, action.subject)) {
        return state;
      }
      return {
        ...state,
        invalidationCause: "retarget",
        invalidationGeneration: state.invalidationGeneration + 1,
        status: state.isOpen
          ? action.subject === null
            ? "no_row_selected"
            : "ready"
          : "closed",
        subject: action.subject,
      };
    case "invalidate":
      return {
        ...state,
        invalidationCause: action.reason,
        invalidationGeneration: state.invalidationGeneration + 1,
        isOpen: action.reason === "action_completed" ? state.isOpen : false,
        status:
          action.reason === "action_completed"
            ? state.isOpen
              ? state.subject === null
                ? "no_row_selected"
                : "ready"
              : "closed"
            : "closed",
        subject: action.reason === "action_completed" ? state.subject : null,
      };
  }
}

export function workbookInspectorSubjectsEqual(
  left: WorkbookInspectorSubject | null,
  right: WorkbookInspectorSubject | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.viewSchemaId === right.viewSchemaId &&
      left.recordId === right.recordId &&
      left.rowVersion === right.rowVersion)
  );
}
