import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "./workbookInspectorErrorModel";
import {
  updateWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "./workbookInspectorSubject";

export type RecordHistoryRollbackAction =
  | "change_set"
  | "history_entry"
  | "row_restore";

export type RecordHistoryRollbackTarget =
  | { readonly kind: "history_entry"; readonly history_entry_ref: string }
  | { readonly kind: "change_set"; readonly change_set_id: string }
  | { readonly kind: "row_restore"; readonly restore_to_revision_no: number };

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

type WorkbookRecordHistoryReadyResult =
  | { readonly kind: "loaded"; readonly data: RecordHistoryData }
  | {
      readonly kind: "load_error";
      readonly error: WorkbookInspectorErrorPresentation;
    };

export type WorkbookRecordHistoryState =
  | {
      readonly phase: "idle";
      readonly subject: WorkbookInspectorSubject | null;
      readonly feedback?: WorkbookInspectorFeedback | undefined;
    }
  | {
      readonly phase: "loading";
      readonly subject: WorkbookInspectorSubject;
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly retainedData?: RecordHistoryData | undefined;
    }
  | {
      readonly phase: "ready";
      readonly subject: WorkbookInspectorSubject;
      readonly result: WorkbookRecordHistoryReadyResult;
      readonly feedback?: WorkbookInspectorFeedback | undefined;
      readonly pendingAction?: WorkbookRecordHistoryPendingAction | undefined;
    }
  | {
      readonly phase: "submitting";
      readonly subject: WorkbookInspectorSubject;
      readonly data: RecordHistoryData;
      readonly operationId: WorkbookRecordHistoryOperationId;
      readonly operation: {
        readonly pendingAction: WorkbookRecordHistoryPendingAction;
      };
    };

export type WorkbookRecordHistoryEvent =
  | {
      readonly type: "retarget";
      readonly subject: WorkbookInspectorSubject | null;
    }
  | { readonly type: "clear" }
  | {
      readonly type: "load_requested";
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookInspectorSubject;
    }
  | {
      readonly type: "load_accepted";
      readonly data: RecordHistoryData;
      readonly feedback?: WorkbookInspectorFeedback | undefined;
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookInspectorSubject;
    }
  | {
      readonly type: "load_rejected";
      readonly error: WorkbookInspectorErrorPresentation;
      readonly feedback?: WorkbookInspectorFeedback | undefined;
      readonly requestId: WorkbookRecordHistoryRequestId;
      readonly subject: WorkbookInspectorSubject;
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
      readonly feedback?: WorkbookInspectorFeedback | undefined;
      readonly operationId: WorkbookRecordHistoryOperationId;
      readonly recordId: string;
      readonly rowVersion: number;
    }
  | {
      readonly type: "operation_rejected";
      readonly feedback: WorkbookInspectorFeedback;
      readonly operationId: WorkbookRecordHistoryOperationId;
    }
  | { readonly type: "feedback_cleared" };

const rollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

export function initialWorkbookRecordHistoryState(
  subject: WorkbookInspectorSubject | null = null,
): WorkbookRecordHistoryState {
  return { phase: "idle", subject };
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

export function workbookRecordHistoryLoadedData(
  state: WorkbookRecordHistoryState,
): RecordHistoryData | null {
  switch (state.phase) {
    case "idle":
      return null;
    case "loading":
      return state.retainedData ?? null;
    case "ready":
      return state.result.kind === "loaded" ? state.result.data : null;
    case "submitting":
      return state.data;
  }
}

export function workbookRecordHistoryPendingAction(
  state: WorkbookRecordHistoryState,
): WorkbookRecordHistoryPendingAction | null {
  return state.phase === "ready" && state.result.kind === "loaded"
    ? (state.pendingAction ?? null)
    : null;
}

export function workbookRecordHistoryFeedback(
  state: WorkbookRecordHistoryState,
): WorkbookInspectorFeedback | null {
  return state.phase === "idle" || state.phase === "ready"
    ? (state.feedback ?? null)
    : null;
}

export function workbookRecordHistoryLoadError(
  state: WorkbookRecordHistoryState,
): WorkbookInspectorErrorPresentation | null {
  return state.phase === "ready" && state.result.kind === "load_error"
    ? state.result.error
    : null;
}

export function workbookRecordHistoryReducer(
  state: WorkbookRecordHistoryState,
  event: WorkbookRecordHistoryEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "retarget":
      return workbookInspectorSubjectsEqual(state.subject, event.subject)
        ? state
        : initialWorkbookRecordHistoryState(event.subject);
    case "clear":
      return initialWorkbookRecordHistoryState();
    case "cancel":
    case "feedback_cleared":
    case "load_accepted":
    case "load_rejected":
    case "load_requested":
    case "operation_accepted":
    case "operation_rejected":
    case "preview":
    case "submit":
      return reduceHistoryPhase(state, event);
  }
}

type WorkbookRecordHistoryPhaseEvent = Exclude<
  WorkbookRecordHistoryEvent,
  { readonly type: "clear" | "retarget" }
>;

function reduceHistoryPhase(
  state: WorkbookRecordHistoryState,
  event: WorkbookRecordHistoryPhaseEvent,
): WorkbookRecordHistoryState {
  switch (state.phase) {
    case "idle":
      return reduceIdleHistory(state, event);
    case "loading":
      return reduceLoadingHistory(state, event);
    case "ready":
      return reduceReadyHistory(state, event);
    case "submitting":
      return reduceSubmittingHistory(state, event);
  }
}

function reduceIdleHistory(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "idle" }>,
  event: WorkbookRecordHistoryPhaseEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "load_requested":
      return reduceHistoryLoadRequested(state, event);
    case "feedback_cleared":
      return state.feedback === undefined
        ? state
        : { ...state, feedback: undefined };
    case "cancel":
    case "load_accepted":
    case "load_rejected":
    case "operation_accepted":
    case "operation_rejected":
    case "preview":
    case "submit":
      return state;
  }
}

function reduceLoadingHistory(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "loading" }>,
  event: WorkbookRecordHistoryPhaseEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "load_requested":
      return reduceHistoryLoadRequested(state, event);
    case "load_accepted":
      return reduceHistoryLoadAccepted(state, event);
    case "load_rejected":
      return reduceHistoryLoadRejected(state, event);
    case "cancel":
    case "feedback_cleared":
    case "operation_accepted":
    case "operation_rejected":
    case "preview":
    case "submit":
      return state;
  }
}

function reduceReadyHistory(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "ready" }>,
  event: WorkbookRecordHistoryPhaseEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "load_requested":
      return reduceHistoryLoadRequested(state, event);
    case "preview":
      return reduceHistoryPreview(state, event);
    case "cancel":
      return state.pendingAction === undefined
        ? state
        : { ...state, pendingAction: undefined };
    case "submit":
      return reduceHistorySubmit(state, event);
    case "feedback_cleared":
      return state.feedback === undefined
        ? state
        : { ...state, feedback: undefined };
    case "load_accepted":
    case "load_rejected":
    case "operation_accepted":
    case "operation_rejected":
      return state;
  }
}

function reduceSubmittingHistory(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "submitting" }>,
  event: WorkbookRecordHistoryPhaseEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "load_requested":
      return reduceHistoryLoadRequested(state, event);
    case "operation_accepted":
      return reduceHistoryOperationAccepted(state, event);
    case "operation_rejected":
      return state.operationId.value === event.operationId.value
        ? {
            feedback: event.feedback,
            phase: "ready",
            result: { data: state.data, kind: "loaded" },
            subject: state.subject,
          }
        : state;
    case "cancel":
    case "feedback_cleared":
    case "load_accepted":
    case "load_rejected":
    case "preview":
    case "submit":
      return state;
  }
}

function reduceHistoryLoadRequested(
  state: WorkbookRecordHistoryState,
  event: Extract<
    WorkbookRecordHistoryPhaseEvent,
    { readonly type: "load_requested" }
  >,
): WorkbookRecordHistoryState {
  if (!workbookInspectorSubjectsEqual(state.subject, event.subject)) {
    return state;
  }
  const retainedData = workbookRecordHistoryLoadedData(state);
  return retainedData === null
    ? {
        phase: "loading",
        requestId: event.requestId,
        subject: event.subject,
      }
    : {
        phase: "loading",
        requestId: event.requestId,
        retainedData,
        subject: event.subject,
      };
}

function reduceHistoryLoadAccepted(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "loading" }>,
  event: Extract<
    WorkbookRecordHistoryPhaseEvent,
    { readonly type: "load_accepted" }
  >,
): WorkbookRecordHistoryState {
  if (
    state.requestId.value !== event.requestId.value ||
    !workbookInspectorSubjectsEqual(state.subject, event.subject) ||
    event.data.record_id !== event.subject.recordId ||
    !isPositiveInteger(event.data.row_version)
  ) {
    return state;
  }
  const subject = updateWorkbookInspectorSubject(state.subject, {
    kind: event.data.deleted ? "deleted" : "live",
    recordId: event.data.record_id,
    rowVersion: event.data.row_version,
  });
  if (subject === null) return state;
  const ready = {
    phase: "ready" as const,
    result: { data: event.data, kind: "loaded" as const },
    subject,
  };
  return event.feedback === undefined
    ? ready
    : { ...ready, feedback: event.feedback };
}

function reduceHistoryLoadRejected(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "loading" }>,
  event: Extract<
    WorkbookRecordHistoryPhaseEvent,
    { readonly type: "load_rejected" }
  >,
): WorkbookRecordHistoryState {
  if (
    state.requestId.value !== event.requestId.value ||
    !workbookInspectorSubjectsEqual(state.subject, event.subject)
  ) {
    return state;
  }
  const ready = {
    phase: "ready" as const,
    result: { error: event.error, kind: "load_error" as const },
    subject: state.subject,
  };
  return event.feedback === undefined
    ? ready
    : { ...ready, feedback: event.feedback };
}

function reduceHistoryPreview(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "ready" }>,
  event: Extract<WorkbookRecordHistoryPhaseEvent, { readonly type: "preview" }>,
): WorkbookRecordHistoryState {
  if (
    state.result.kind !== "loaded" ||
    event.pendingAction.recordId !== state.subject.recordId ||
    event.pendingAction.rowVersion !== state.subject.rowVersion ||
    !pendingActionMatchesSubject(event.pendingAction, state.subject)
  ) {
    return state;
  }
  return {
    ...state,
    feedback: undefined,
    pendingAction: event.pendingAction,
  };
}

function reduceHistorySubmit(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "ready" }>,
  event: Extract<WorkbookRecordHistoryPhaseEvent, { readonly type: "submit" }>,
): WorkbookRecordHistoryState {
  return state.result.kind === "loaded" && state.pendingAction !== undefined
    ? {
        data: state.result.data,
        operation: { pendingAction: state.pendingAction },
        operationId: event.operationId,
        phase: "submitting",
        subject: state.subject,
      }
    : state;
}

function reduceHistoryOperationAccepted(
  state: Extract<WorkbookRecordHistoryState, { readonly phase: "submitting" }>,
  event: Extract<
    WorkbookRecordHistoryPhaseEvent,
    { readonly type: "operation_accepted" }
  >,
): WorkbookRecordHistoryState {
  const pending = state.operation.pendingAction;
  if (
    state.operationId.value !== event.operationId.value ||
    pending.recordId !== state.subject.recordId ||
    pending.rowVersion !== state.subject.rowVersion ||
    !pendingActionMatchesSubject(pending, state.subject) ||
    event.recordId !== pending.recordId ||
    !isPositiveInteger(event.rowVersion)
  ) {
    return state;
  }
  const subject = updateWorkbookInspectorSubject(state.subject, {
    kind:
      pending.kind === "destructive"
        ? pending.operation === "delete"
          ? "deleted"
          : "live"
        : state.subject.kind,
    recordId: event.recordId,
    rowVersion: event.rowVersion,
  });
  if (subject === null) return state;
  return event.feedback === undefined
    ? { phase: "idle", subject }
    : { feedback: event.feedback, phase: "idle", subject };
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
        ? { history_entry_ref: item.history_entry_ref, kind: "history_entry" }
        : null;
    case "change_set":
      return item.change_set_id.trim() === ""
        ? null
        : { change_set_id: item.change_set_id, kind: "change_set" };
    case "row_restore":
      return isPositiveInteger(item.revision_no)
        ? { kind: "row_restore", restore_to_revision_no: item.revision_no }
        : null;
  }
}

function pendingActionMatchesSubject(
  pending: WorkbookRecordHistoryPendingAction,
  subject: WorkbookInspectorSubject,
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

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}
