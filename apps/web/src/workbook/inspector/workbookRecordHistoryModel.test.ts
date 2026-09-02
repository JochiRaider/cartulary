import { describe, expect, it } from "vitest";
import {
  buildRecordRollbackTargetFromHistoryAction,
  initialWorkbookRecordHistoryState,
  type RecordHistoryData,
  type RecordHistoryItem,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";

const liveSubject = {
  kind: "live",
  recordId: "record-a",
  rowVersion: 4,
} as const;
const data: RecordHistoryData = {
  deleted: false,
  incident_id: "incident-a",
  items: [],
  record_id: liveSubject.recordId,
  row_version: liveSubject.rowVersion,
};

describe("workbookRecordHistoryModel", () => {
  it("rejects stale and duplicate loads after a retarget", () => {
    const firstRequest = workbookRecordHistoryRequestId(1);
    const secondRequest = workbookRecordHistoryRequestId(2);
    let state = initialWorkbookRecordHistoryState(liveSubject);
    state = workbookRecordHistoryReducer(state, {
      requestId: firstRequest,
      subject: liveSubject,
      type: "load_requested",
    });
    state = workbookRecordHistoryReducer(state, {
      requestId: secondRequest,
      subject: liveSubject,
      type: "load_requested",
    });
    state = workbookRecordHistoryReducer(state, {
      data,
      requestId: firstRequest,
      subject: liveSubject,
      type: "load_accepted",
    });
    expect(state.phase).toBe("loading");
    expect(state.data).toBeNull();

    const nextSubject = {
      kind: "live",
      recordId: "record-b",
      rowVersion: 1,
    } as const;
    state = workbookRecordHistoryReducer(state, {
      subject: nextSubject,
      type: "retarget",
    });
    state = workbookRecordHistoryReducer(state, {
      data,
      requestId: secondRequest,
      subject: liveSubject,
      type: "load_accepted",
    });
    expect(state).toEqual(initialWorkbookRecordHistoryState(nextSubject));
  });

  it("cancels and reopens confirmation without weakening operation identity", () => {
    const requestId = workbookRecordHistoryRequestId(1);
    let state = workbookRecordHistoryReducer(
      initialWorkbookRecordHistoryState(liveSubject),
      {
        requestId,
        subject: liveSubject,
        type: "load_requested",
      },
    );
    state = workbookRecordHistoryReducer(state, {
      data,
      requestId,
      subject: liveSubject,
      type: "load_accepted",
    });
    const pending = {
      kind: "destructive",
      operation: "delete",
      recordId: liveSubject.recordId,
      rowVersion: liveSubject.rowVersion,
    } as const;
    state = workbookRecordHistoryReducer(state, {
      pendingAction: pending,
      type: "preview",
    });
    state = workbookRecordHistoryReducer(state, { type: "cancel" });
    expect(state.pendingAction).toBeNull();
    state = workbookRecordHistoryReducer(state, {
      pendingAction: pending,
      type: "preview",
    });
    const operationId = workbookRecordHistoryOperationId(2);
    state = workbookRecordHistoryReducer(state, {
      operationId,
      type: "submit",
    });
    expect(state).toMatchObject({
      operationId,
      pendingAction: null,
      phase: "submitting",
    });

    const stale = workbookRecordHistoryReducer(state, {
      error: { primaryMessage: "stale", technicalFields: [] },
      operationId: workbookRecordHistoryOperationId(1),
      type: "operation_rejected",
    });
    expect(stale).toBe(state);
    state = workbookRecordHistoryReducer(state, {
      feedback: null,
      operationId,
      subject: {
        kind: "deleted",
        recordId: liveSubject.recordId,
        rowVersion: 5,
      },
      type: "operation_accepted",
    });
    expect(state).toMatchObject({
      phase: "idle",
      subject: {
        kind: "deleted",
        recordId: liveSubject.recordId,
        rowVersion: 5,
      },
    });
  });

  it("builds rollback targets only from the server-advertised selector", () => {
    const item: RecordHistoryItem = {
      actor_user_id: "actor-a",
      available_rollback_actions: [
        "history_entry",
        "change_set",
        "row_restore",
      ],
      change_set_id: "change-a",
      committed_at: "2026-09-01T00:00:00Z",
      diff_summary: { summary: "Changed", units: [] },
      history_entry_ref: "entry-a",
      history_item_ref: "item-a",
      operation: "patch",
      reversible: true,
      revision_no: 3,
    };
    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "history_entry"),
    ).toEqual({ history_entry_ref: "entry-a", kind: "history_entry" });
    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "change_set"),
    ).toEqual({ change_set_id: "change-a", kind: "change_set" });
    expect(
      buildRecordRollbackTargetFromHistoryAction(item, "row_restore"),
    ).toEqual({ kind: "row_restore", restore_to_revision_no: 3 });
    expect(
      buildRecordRollbackTargetFromHistoryAction(
        { ...item, available_rollback_actions: [] },
        "history_entry",
      ),
    ).toBeNull();
  });
});
