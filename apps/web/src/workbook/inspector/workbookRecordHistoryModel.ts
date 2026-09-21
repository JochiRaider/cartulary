import type { WorkbookRecordHistoryPendingAction } from "../history/workbookHistoryItem";

export {
  buildRecordRollbackTargetFromHistoryAction,
  normalizeRecordHistoryData,
  type RecordHistoryRollbackAction,
  type RecordHistoryRollbackTarget,
  type WorkbookRecordHistoryPendingAction,
} from "../history/workbookHistoryItem";

import type { HistoryLookupState } from "../history/HistoryActionLookup";
import type { HistoryBrowsingState } from "../history/workbookHistoryBrowsing";
import type { RecordHistoryData } from "../history/workbookHistoryPage";
import { workbookInspectorErrorPresentation } from "./workbookInspectorErrorModel";

export type {
  RecordHistoryData,
  RecordHistoryItem,
} from "../history/workbookHistoryPage";

import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "./workbookInspectorErrorModel";
import {
  updateWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "./workbookInspectorSubject";

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
      readonly retainedData?: RecordHistoryData;
    };

export type WorkbookRecordHistoryState = {
  readonly browsing?: HistoryBrowsingState;
  readonly lookup?: HistoryLookupState | undefined;
} & (
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
    }
);

export type WorkbookRecordHistoryEvent =
  | {
      readonly type: "browsing_changed";
      readonly browsing: HistoryBrowsingState;
    }
  | {
      readonly type: "lookup_changed";
      readonly lookup: HistoryLookupState | undefined;
    }
  | { readonly type: "review_accepted"; readonly data: RecordHistoryData }
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
  if (state.browsing?.accepted) return state.browsing.accepted.data;
  switch (state.phase) {
    case "idle":
      return null;
    case "loading":
      return state.retainedData ?? null;
    case "ready":
      return state.result.kind === "loaded"
        ? state.result.data
        : (state.result.retainedData ?? null);
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
  if (state.browsing?.failure)
    return workbookInspectorErrorPresentation(state.browsing.failure.error);
  return state.phase === "ready" && state.result.kind === "load_error"
    ? state.result.error
    : null;
}

export function workbookRecordHistoryReducer(
  state: WorkbookRecordHistoryState,
  event: WorkbookRecordHistoryEvent,
): WorkbookRecordHistoryState {
  if (event.type === "lookup_changed")
    return { ...state, lookup: event.lookup };
  if (event.type === "browsing_changed") {
    const browsing = event.browsing;
    if (
      !state.subject ||
      state.subject.recordId !== browsing.recordId ||
      state.subject.viewSchemaId !== browsing.viewSchemaId
    )
      return state;
    const data = browsing.accepted?.data;
    if (data) {
      const subject =
        updateWorkbookInspectorSubject(state.subject, {
          recordId: data.record_id,
          rowVersion: Math.max(state.subject.rowVersion, data.row_version),
          kind:
            data.row_version >= state.subject.rowVersion
              ? data.deleted
                ? "deleted"
                : "live"
              : state.subject.kind,
        }) ?? state.subject;
      const sameVersion =
        subject.rowVersion === state.subject.rowVersion &&
        subject.kind === state.subject.kind;
      return {
        ...state,
        phase: "ready",
        subject,
        browsing,
        result: { kind: "loaded", data },
        ...(state.phase === "ready" &&
        sameVersion &&
        browsing.pending?.kind !== "refresh" &&
        browsing.pending?.kind !== "initial"
          ? { pendingAction: state.pendingAction }
          : { pendingAction: undefined }),
      };
    }
    if (browsing.pending)
      return {
        phase: "loading",
        subject: state.subject,
        requestId: workbookRecordHistoryRequestId(browsing.pending.generation),
        browsing,
      };
    if (browsing.failure)
      return {
        phase: "ready",
        subject: state.subject,
        browsing,
        result: {
          kind: "load_error",
          error: workbookInspectorErrorPresentation(browsing.failure.error),
        },
      };
    return { phase: "idle", subject: state.subject, browsing };
  }
  if (event.type === "review_accepted") {
    if (
      !state.subject ||
      state.subject.recordId !== event.data.record_id ||
      event.data.row_version < state.subject.rowVersion
    )
      return state;
    const subject = updateWorkbookInspectorSubject(state.subject, {
      recordId: event.data.record_id,
      rowVersion: event.data.row_version,
      kind: event.data.deleted ? "deleted" : "live",
    });
    if (!subject) return state;
    const data = state.browsing?.accepted
      ? {
          ...state.browsing.accepted.data,
          row_version: event.data.row_version,
          deleted: event.data.deleted,
        }
      : event.data;
    const browsing = state.browsing?.accepted
      ? {
          ...state.browsing,
          accepted: {
            ...state.browsing.accepted,
            data: {
              ...state.browsing.accepted.data,
              row_version: event.data.row_version,
              deleted: event.data.deleted,
            },
          },
        }
      : state.browsing;
    return {
      ...state,
      phase: "ready",
      subject,
      result: { kind: "loaded", data },
      ...(browsing ? { browsing } : {}),
      pendingAction: undefined,
    };
  }
  if (
    event.type === "retarget" &&
    state.browsing &&
    state.subject &&
    event.subject &&
    state.subject.recordId === event.subject.recordId &&
    state.subject.viewSchemaId === event.subject.viewSchemaId
  ) {
    if (event.subject.rowVersion <= state.subject.rowVersion) return state;
    const loaded = workbookRecordHistoryLoadedData(state);
    if (!loaded)
      return state.phase === "ready"
        ? { ...state, subject: event.subject, pendingAction: undefined }
        : { ...state, subject: event.subject };
    return workbookRecordHistoryReducer(
      { ...state, subject: event.subject },
      {
        type: "review_accepted",
        data: {
          ...loaded,
          record_id: event.subject.recordId,
          row_version: event.subject.rowVersion,
          deleted: event.subject.kind === "deleted",
        },
      },
    );
  }
  const next = legacyHistoryReducer(state, event);
  if (next === state || event.type === "clear" || event.type === "retarget")
    return next;
  const browsing = state.browsing;
  const acknowledged =
    event.type === "operation_accepted" &&
    next.subject &&
    browsing?.accepted &&
    next.subject.rowVersion >= browsing.accepted.data.row_version
      ? {
          ...browsing,
          accepted: {
            ...browsing.accepted,
            data: {
              ...browsing.accepted.data,
              row_version: next.subject.rowVersion,
              deleted: next.subject.kind === "deleted",
            },
          },
        }
      : browsing;
  return {
    ...next,
    ...(acknowledged ? { browsing: acknowledged } : {}),
    ...(event.type === "cancel"
      ? { lookup: undefined }
      : state.lookup
        ? { lookup: state.lookup }
        : {}),
  };
}

function legacyHistoryReducer(
  state: WorkbookRecordHistoryState,
  event: Exclude<
    WorkbookRecordHistoryEvent,
    { readonly type: "browsing_changed" | "lookup_changed" | "review_accepted" }
  >,
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
  {
    readonly type:
      | "clear"
      | "retarget"
      | "browsing_changed"
      | "lookup_changed"
      | "review_accepted";
  }
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
    result: {
      error: event.error,
      kind: "load_error" as const,
      ...(state.retainedData ? { retainedData: state.retainedData } : {}),
    },
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

function pendingActionMatchesSubject(
  pending: WorkbookRecordHistoryPendingAction,
  subject: WorkbookInspectorSubject,
): boolean {
  return pending.kind === "rollback"
    ? pending.target.kind === pending.action
    : (pending.operation === "delete" && subject.kind === "live") ||
        (pending.operation === "restore" && subject.kind === "deleted");
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}
