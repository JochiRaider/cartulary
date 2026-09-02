import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "./workbookInspectorErrorModel";

export type RecordHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export type RecordHistoryRollbackTarget =
  | {
      readonly kind: "history_entry";
      readonly history_entry_ref: string;
    }
  | {
      readonly kind: "change_set";
      readonly change_set_id: string;
    }
  | {
      readonly kind: "row_restore";
      readonly restore_to_revision_no: number;
    };

export type RecordHistoryItem = {
  readonly actor_user_id: string;
  readonly committed_at: string;
  readonly history_item_ref: string;
  readonly operation: string;
  readonly diff_summary: {
    readonly summary: string;
    readonly units: readonly Record<string, unknown>[];
  };
  readonly change_set_id: string;
  readonly reversible: boolean;
  readonly available_rollback_actions: readonly RecordHistoryRollbackAction[];
  readonly history_entry_ref?: string;
  readonly revision_no?: number;
};

export type RecordHistoryData = {
  readonly incident_id: string;
  readonly record_id: string;
  readonly row_version: number;
  readonly deleted: boolean;
  readonly items: readonly RecordHistoryItem[];
};

export type WorkbookRecordHistorySubject =
  | {
      readonly kind: "live";
      readonly recordId: string;
      readonly rowVersion: number;
    }
  | {
      readonly kind: "deleted";
      readonly recordId: string;
      readonly rowVersion: number;
    };

export type WorkbookRecordHistoryPendingAction =
  | {
      readonly kind: "rollback";
      readonly action: RecordHistoryRollbackAction;
      readonly historyItemRef: string;
      readonly recordId: string;
      readonly rowVersion: number;
      readonly target: RecordHistoryRollbackTarget;
    }
  | {
      readonly kind: "destructive";
      readonly operation: "delete" | "restore";
      readonly recordId: string;
      readonly rowVersion: number;
    };

export type WorkbookRecordHistoryRequestId = {
  readonly kind: "record_history_request";
  readonly value: number;
};

export type WorkbookRecordHistoryOperationId = {
  readonly kind: "record_history_operation";
  readonly value: number;
};

export type WorkbookRecordHistoryState = {
  readonly subject: WorkbookRecordHistorySubject | null;
  readonly phase: "idle" | "loading" | "ready" | "submitting";
  readonly data: RecordHistoryData | null;
  readonly error: WorkbookInspectorErrorPresentation | null;
  readonly feedback: WorkbookInspectorFeedback | null;
  readonly pendingAction: WorkbookRecordHistoryPendingAction | null;
  readonly requestId: WorkbookRecordHistoryRequestId | null;
  readonly operationId: WorkbookRecordHistoryOperationId | null;
};

export type WorkbookRecordHistoryEvent =
  | {
      readonly type: "retarget";
      readonly subject: WorkbookRecordHistorySubject | null;
    }
  | { readonly type: "clear" }
  | {
      readonly type: "load_requested";
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookRecordHistorySubject;
    }
  | {
      readonly type: "load_accepted";
      readonly data: RecordHistoryData;
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookRecordHistorySubject;
    }
  | {
      readonly type: "load_rejected";
      readonly error: WorkbookInspectorErrorPresentation;
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookRecordHistorySubject;
    }
  | {
      readonly type: "preview";
      readonly pendingAction: WorkbookRecordHistoryPendingAction;
    }
  | { readonly type: "cancel" }
  | {
      readonly type: "submit";
      readonly operationId: WorkbookRecordHistoryOperationId;
    }
  | {
      readonly type: "operation_accepted";
      readonly feedback: WorkbookInspectorFeedback | null;
      readonly operationId: WorkbookRecordHistoryOperationId;
      readonly subject: WorkbookRecordHistorySubject;
    }
  | {
      readonly type: "operation_rejected";
      readonly error: WorkbookInspectorErrorPresentation;
      readonly operationId: WorkbookRecordHistoryOperationId;
    }
  | { readonly type: "feedback_cleared" };

const rollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

export function initialWorkbookRecordHistoryState(
  subject: WorkbookRecordHistorySubject | null = null,
): WorkbookRecordHistoryState {
  return {
    data: null,
    error: null,
    feedback: null,
    operationId: null,
    pendingAction: null,
    phase: "idle",
    requestId: null,
    subject,
  };
}

export function workbookRecordHistoryRequestId(
  value: number,
): WorkbookRecordHistoryRequestId {
  return { kind: "record_history_request", value };
}

export function workbookRecordHistoryOperationId(
  value: number,
): WorkbookRecordHistoryOperationId {
  return { kind: "record_history_operation", value };
}

function workbookRecordHistorySubjectsEqual(
  left: WorkbookRecordHistorySubject | null,
  right: WorkbookRecordHistorySubject | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.kind === right.kind &&
      left.recordId === right.recordId &&
      left.rowVersion === right.rowVersion)
  );
}

export function workbookRecordHistoryReducer(
  state: WorkbookRecordHistoryState,
  event: WorkbookRecordHistoryEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "retarget":
      return workbookRecordHistorySubjectsEqual(state.subject, event.subject)
        ? state
        : initialWorkbookRecordHistoryState(event.subject);
    case "clear":
      return initialWorkbookRecordHistoryState();
    case "load_requested":
      if (!workbookRecordHistorySubjectsEqual(state.subject, event.subject)) {
        return state;
      }
      return {
        ...state,
        error: null,
        pendingAction: null,
        phase: "loading",
        requestId: event.requestId,
      };
    case "load_accepted":
      if (
        state.phase !== "loading" ||
        !requestIdsEqual(state.requestId, event.requestId) ||
        !workbookRecordHistorySubjectsEqual(state.subject, event.subject) ||
        event.data.record_id !== event.subject.recordId ||
        !isPositiveInteger(event.data.row_version)
      ) {
        return state;
      }
      return {
        ...state,
        data: event.data,
        error: null,
        phase: "ready",
        requestId: null,
        subject: {
          kind: event.data.deleted ? "deleted" : "live",
          recordId: event.data.record_id,
          rowVersion: event.data.row_version,
        },
      };
    case "load_rejected":
      if (
        state.phase !== "loading" ||
        !requestIdsEqual(state.requestId, event.requestId) ||
        !workbookRecordHistorySubjectsEqual(state.subject, event.subject)
      ) {
        return state;
      }
      return {
        ...state,
        data: null,
        error: event.error,
        phase: "ready",
        requestId: null,
      };
    case "preview":
      if (
        state.phase !== "ready" ||
        state.subject === null ||
        event.pendingAction.recordId !== state.subject.recordId ||
        event.pendingAction.rowVersion !== state.subject.rowVersion ||
        !pendingActionMatchesSubject(event.pendingAction, state.subject)
      ) {
        return state;
      }
      return {
        ...state,
        error: null,
        feedback: null,
        pendingAction: event.pendingAction,
      };
    case "cancel":
      return state.pendingAction === null
        ? state
        : { ...state, pendingAction: null };
    case "submit":
      return state.phase === "ready" && state.pendingAction !== null
        ? {
            ...state,
            error: null,
            feedback: null,
            operationId: event.operationId,
            pendingAction: null,
            phase: "submitting",
          }
        : state;
    case "operation_accepted":
      if (
        state.phase !== "submitting" ||
        !operationIdsEqual(state.operationId, event.operationId)
      ) {
        return state;
      }
      return {
        ...state,
        data: null,
        error: null,
        feedback: event.feedback,
        operationId: null,
        pendingAction: null,
        phase: "idle",
        requestId: null,
        subject: event.subject,
      };
    case "operation_rejected":
      if (
        state.phase !== "submitting" ||
        !operationIdsEqual(state.operationId, event.operationId)
      ) {
        return state;
      }
      return {
        ...state,
        error: event.error,
        operationId: null,
        pendingAction: null,
        phase: "ready",
      };
    case "feedback_cleared":
      return state.feedback === null ? state : { ...state, feedback: null };
  }
}

export function normalizeRecordHistoryData(
  data: RecordHistoryData,
): RecordHistoryData | null {
  if (data.record_id.trim() === "" || !isPositiveInteger(data.row_version)) {
    return null;
  }
  const seen = new Set<string>();
  for (const item of data.items) {
    if (
      item.history_item_ref.trim() === "" ||
      item.change_set_id.trim() === "" ||
      seen.has(item.history_item_ref)
    ) {
      return null;
    }
    seen.add(item.history_item_ref);
    let previous = -1;
    for (const action of item.available_rollback_actions) {
      const index = rollbackActionOrder.indexOf(action);
      if (index <= previous || !validItemSelector(item, action)) return null;
      previous = index;
    }
  }
  return data;
}

export function buildRecordRollbackTargetFromHistoryAction(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): RecordHistoryRollbackTarget | null {
  if (!item.available_rollback_actions.includes(action)) return null;
  switch (action) {
    case "history_entry":
      return typeof item.history_entry_ref === "string" &&
        item.history_entry_ref.trim() !== ""
        ? {
            history_entry_ref: item.history_entry_ref,
            kind: "history_entry",
          }
        : null;
    case "change_set":
      return item.change_set_id.trim() === ""
        ? null
        : { change_set_id: item.change_set_id, kind: "change_set" };
    case "row_restore":
      return isPositiveInteger(item.revision_no)
        ? {
            kind: "row_restore",
            restore_to_revision_no: item.revision_no,
          }
        : null;
  }
}

function pendingActionMatchesSubject(
  pending: WorkbookRecordHistoryPendingAction,
  subject: WorkbookRecordHistorySubject,
): boolean {
  return pending.kind === "rollback"
    ? pending.target.kind === pending.action
    : (pending.operation === "delete" && subject.kind === "live") ||
        (pending.operation === "restore" && subject.kind === "deleted");
}

function validItemSelector(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): boolean {
  switch (action) {
    case "history_entry":
      return (
        typeof item.history_entry_ref === "string" &&
        item.history_entry_ref.trim() !== ""
      );
    case "change_set":
      return item.change_set_id.trim() !== "";
    case "row_restore":
      return isPositiveInteger(item.revision_no);
  }
}

function requestIdsEqual(
  left: WorkbookRecordHistoryRequestId | null,
  right: WorkbookRecordHistoryRequestId,
): boolean {
  return left?.value === right.value;
}

function operationIdsEqual(
  left: WorkbookRecordHistoryOperationId | null,
  right: WorkbookRecordHistoryOperationId,
): boolean {
  return left?.value === right.value;
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}
