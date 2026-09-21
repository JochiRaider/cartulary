import type { HistoryLookupState } from "../history/HistoryPageLookup";

import {
  type HistoryBrowsingState,
  observeHistoryRecord,
} from "../history/workbookHistoryBrowsing";
import type { WorkbookRecordHistoryPendingAction } from "../history/workbookHistoryItem";
import type { HistoryPage } from "../history/workbookHistoryPage";
import type { WorkbookRecordSubject } from "../ports/WorkbookRecordSubject";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
} from "./workbookInspectorErrorModel";
import {
  updateWorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "./workbookInspectorSubject";

type WorkbookRecordHistoryOperationId = {
  readonly kind: "record_history_operation";
  readonly value: number;
};

type HistoryPresentation = {
  readonly subject: WorkbookRecordSubject | null;
  readonly browsing: HistoryBrowsingState | null;
  readonly lookup?: HistoryLookupState | undefined;
  readonly pendingAction?: WorkbookRecordHistoryPendingAction | undefined;
  readonly submission?:
    | {
        readonly operationId: WorkbookRecordHistoryOperationId;
        readonly pendingAction: WorkbookRecordHistoryPendingAction;
      }
    | undefined;
  readonly feedback?: WorkbookInspectorFeedback | undefined;
};

/** Read content exists only in browsing; attachment/review never owns a copy. */
export type WorkbookRecordHistoryState = HistoryPresentation & {
  readonly phase: "idle" | "loading" | "ready" | "submitting";
};

export type WorkbookRecordHistoryEvent =
  | {
      readonly type: "browsing_changed";
      readonly browsing: HistoryBrowsingState;
    }
  | {
      readonly type: "lookup_changed";
      readonly lookup: HistoryLookupState | undefined;
    }
  | { readonly type: "review_accepted"; readonly data: HistoryPage }
  | {
      readonly type: "retarget";
      readonly subject: WorkbookRecordSubject | null;
    }
  | { readonly type: "clear" }
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

function presentation(state: HistoryPresentation): WorkbookRecordHistoryState {
  return {
    ...state,
    // A derived presentation phase, never a second read-state machine.
    phase: state.submission
      ? "submitting"
      : state.browsing?.accepted || state.browsing?.failure
        ? "ready"
        : state.browsing?.pending
          ? "loading"
          : "idle",
  };
}

export function initialWorkbookRecordHistoryState(
  subject: WorkbookRecordSubject | null = null,
): WorkbookRecordHistoryState {
  return presentation({ subject, browsing: null });
}

export function workbookRecordHistoryOperationId(
  value: number,
): WorkbookRecordHistoryOperationId {
  return { kind: "record_history_operation", value };
}

export function workbookRecordHistoryLoadedData(
  state: WorkbookRecordHistoryState,
): HistoryPage | null {
  return state.browsing?.accepted?.data ?? null;
}

export function workbookRecordHistoryPendingAction(
  state: WorkbookRecordHistoryState,
): WorkbookRecordHistoryPendingAction | null {
  return !state.submission && state.browsing?.accepted
    ? (state.pendingAction ?? null)
    : null;
}

export function workbookRecordHistoryFeedback(
  state: WorkbookRecordHistoryState,
): WorkbookInspectorFeedback | null {
  return state.feedback ?? null;
}

export function workbookRecordHistoryLoadError(
  state: WorkbookRecordHistoryState,
) {
  return state.browsing?.failure
    ? workbookInspectorErrorPresentation(state.browsing.failure.error)
    : null;
}

function updateObservation(
  state: HistoryPresentation,
  rowVersion: number,
  deleted: boolean,
): HistoryPresentation {
  const subject = state.subject;
  if (!subject || rowVersion < subject.rowVersion) return state;
  return {
    ...state,
    browsing: state.browsing
      ? observeHistoryRecord(state.browsing, {
          recordId: subject.recordId,
          rowVersion,
          deleted,
        })
      : null,
    subject:
      updateWorkbookInspectorSubject(subject, {
        recordId: subject.recordId,
        rowVersion,
        kind: deleted ? "deleted" : "live",
      }) ?? subject,
  };
}

export function workbookRecordHistoryReducer(
  state: WorkbookRecordHistoryState,
  event: WorkbookRecordHistoryEvent,
): WorkbookRecordHistoryState {
  switch (event.type) {
    case "clear":
      return initialWorkbookRecordHistoryState();
    case "retarget": {
      if (workbookInspectorSubjectsEqual(state.subject, event.subject))
        return state;
      if (
        !state.subject ||
        !event.subject ||
        state.subject.recordId !== event.subject.recordId ||
        state.subject.viewSchemaId !== event.subject.viewSchemaId
      )
        return initialWorkbookRecordHistoryState(event.subject);
      if (event.subject.rowVersion <= state.subject.rowVersion) return state;
      return presentation({
        ...updateObservation(
          state,
          event.subject.rowVersion,
          event.subject.kind === "deleted",
        ),
        pendingAction: undefined,
        submission: undefined,
      });
    }
    case "browsing_changed": {
      const browsing = event.browsing;
      if (
        !state.subject ||
        state.subject.recordId !== browsing.recordId ||
        state.subject.viewSchemaId !== browsing.viewSchemaId
      )
        return state;
      const data = browsing.accepted?.data;
      const sameVersion =
        data &&
        data.row_version === state.subject.rowVersion &&
        data.deleted === (state.subject.kind === "deleted");
      return presentation({
        ...(data
          ? updateObservation(state, data.row_version, data.deleted)
          : state),
        browsing,
        submission:
          data && data.row_version > state.subject.rowVersion
            ? undefined
            : state.submission,
        pendingAction:
          sameVersion &&
          browsing.pending?.kind !== "refresh" &&
          browsing.pending?.kind !== "initial"
            ? state.pendingAction
            : undefined,
      });
    }
    case "lookup_changed":
      return presentation({ ...state, lookup: event.lookup });
    case "review_accepted": {
      if (
        !state.subject ||
        event.data.record_id !== state.subject.recordId ||
        event.data.row_version < state.subject.rowVersion
      )
        return state;
      return presentation({
        ...updateObservation(state, event.data.row_version, event.data.deleted),
        pendingAction: undefined,
      });
    }
    case "cancel":
      return state.pendingAction || state.lookup
        ? presentation({
            ...state,
            pendingAction: undefined,
            lookup: undefined,
          })
        : state;
    case "feedback_cleared":
      return state.feedback
        ? presentation({ ...state, feedback: undefined })
        : state;
    case "preview":
      if (
        !state.subject ||
        state.submission ||
        !state.browsing?.accepted ||
        event.pendingAction.recordId !== state.subject.recordId ||
        event.pendingAction.rowVersion !== state.subject.rowVersion ||
        !pendingActionMatchesSubject(event.pendingAction, state.subject)
      )
        return state;
      return presentation({
        ...state,
        feedback: undefined,
        pendingAction: event.pendingAction,
      });
    case "submit":
      return state.pendingAction &&
        !state.submission &&
        state.browsing?.accepted
        ? presentation({
            ...state,
            submission: {
              operationId: event.operationId,
              pendingAction: state.pendingAction,
            },
            pendingAction: undefined,
          })
        : state;
    case "operation_rejected":
      return state.submission?.operationId.value === event.operationId.value
        ? presentation({
            ...state,
            submission: undefined,
            feedback: event.feedback,
          })
        : state;
    case "operation_accepted": {
      const pending = state.submission?.pendingAction;
      if (
        !pending ||
        !state.subject ||
        state.submission?.operationId.value !== event.operationId.value ||
        pending.recordId !== state.subject.recordId ||
        pending.rowVersion !== state.subject.rowVersion ||
        !pendingActionMatchesSubject(pending, state.subject) ||
        event.recordId !== pending.recordId ||
        !Number.isInteger(event.rowVersion) ||
        event.rowVersion < 1
      )
        return state;
      const deleted =
        pending.kind === "destructive"
          ? pending.operation === "delete"
          : state.subject.kind === "deleted";
      return presentation({
        ...updateObservation(state, event.rowVersion, deleted),
        submission: undefined,
        pendingAction: undefined,
        feedback: event.feedback,
      });
    }
  }
}

function pendingActionMatchesSubject(
  pending: WorkbookRecordHistoryPendingAction,
  subject: WorkbookRecordSubject,
): boolean {
  return pending.kind === "rollback"
    ? pending.target.kind === pending.action
    : pending.operation === "delete"
      ? subject.kind === "live"
      : subject.kind === "deleted";
}
