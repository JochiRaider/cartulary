import { describe, expect, it } from "vitest";
import {
  buildRecordRollbackTargetFromHistoryAction,
  initialWorkbookRecordHistoryState,
  type RecordHistoryData,
  type RecordHistoryItem,
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "./workbookRecordHistoryModel";

const liveSubject = {
  kind: "live",
  label: "Record A",
  recordId: "record-a",
  rowVersion: 4,
  surfaceLabel: "Test records",
  viewSchemaId: "cartulary.view.test.v1",
} as const;
const data: RecordHistoryData = {
  deleted: false,
  incident_id: "incident-a",
  items: [],
  record_id: liveSubject.recordId,
  row_version: liveSubject.rowVersion,
};

describe("workbookRecordHistoryModel", () => {
  it("preserves an exactly equal subject reference and resets every changed identity", () => {
    const initial = initialWorkbookRecordHistoryState(liveSubject);
    const equal = workbookRecordHistoryReducer(initial, {
      subject: { ...liveSubject },
      type: "retarget",
    });
    expect(equal).toBe(initial);

    for (const subject of [
      {
        ...liveSubject,
        kind: "deleted" as const,
        stateLabel: "Deleted",
      },
      { ...liveSubject, recordId: "record-b" },
      { ...liveSubject, rowVersion: liveSubject.rowVersion + 1 },
      null,
    ]) {
      expect(
        workbookRecordHistoryReducer(initial, { subject, type: "retarget" }),
      ).toEqual(initialWorkbookRecordHistoryState(subject));
    }
  });

  it("makes events from the wrong phase, request, operation, or subject explicit no-ops", () => {
    const initial = initialWorkbookRecordHistoryState(liveSubject);
    const requestId = workbookRecordHistoryRequestId(1);
    const operationId = workbookRecordHistoryOperationId(1);
    const wrongSubject = { ...liveSubject, recordId: "record-b" };
    const events = [
      {
        data,
        requestId,
        subject: liveSubject,
        type: "load_accepted" as const,
      },
      {
        error: { primaryMessage: "late", technicalFields: [] },
        requestId,
        subject: liveSubject,
        type: "load_rejected" as const,
      },
      {
        pendingAction: {
          kind: "destructive" as const,
          operation: "delete" as const,
          recordId: liveSubject.recordId,
          rowVersion: liveSubject.rowVersion,
        },
        type: "preview" as const,
      },
      { operationId, type: "submit" as const },
      {
        operationId,
        subject: liveSubject,
        type: "operation_accepted" as const,
      },
      {
        feedback: {
          error: { primaryMessage: "late", technicalFields: [] },
          kind: "error" as const,
        },
        operationId,
        type: "operation_rejected" as const,
      },
      {
        requestId,
        subject: wrongSubject,
        type: "load_requested" as const,
      },
    ];

    for (const event of events) {
      expect(workbookRecordHistoryReducer(initial, event)).toBe(initial);
    }
  });

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
    expect(workbookRecordHistoryLoadedData(state)).toBeNull();

    const nextSubject = {
      ...liveSubject,
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

  it("retains only same-subject loaded data while loading and makes load failure exclusive", () => {
    const firstRequest = workbookRecordHistoryRequestId(1);
    let state = workbookRecordHistoryReducer(
      initialWorkbookRecordHistoryState(liveSubject),
      {
        requestId: firstRequest,
        subject: liveSubject,
        type: "load_requested",
      },
    );
    state = workbookRecordHistoryReducer(state, {
      data,
      requestId: firstRequest,
      subject: liveSubject,
      type: "load_accepted",
    });
    const secondRequest = workbookRecordHistoryRequestId(2);
    state = workbookRecordHistoryReducer(state, {
      requestId: secondRequest,
      subject: liveSubject,
      type: "load_requested",
    });
    expect(state).toMatchObject({
      phase: "loading",
      retainedData: data,
      requestId: secondRequest,
    });
    state = workbookRecordHistoryReducer(state, {
      error: { primaryMessage: "Load failed", technicalFields: [] },
      requestId: secondRequest,
      subject: liveSubject,
      type: "load_rejected",
    });
    expect(state).toEqual({
      phase: "ready",
      result: {
        error: { primaryMessage: "Load failed", technicalFields: [] },
        kind: "load_error",
      },
      subject: liveSubject,
    });
    expect(workbookRecordHistoryLoadedData(state)).toBeNull();
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
    expect(workbookRecordHistoryPendingAction(state)).toBeNull();
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
      operation: { pendingAction: pending },
      phase: "submitting",
    });

    const stale = workbookRecordHistoryReducer(state, {
      feedback: {
        error: { primaryMessage: "stale", technicalFields: [] },
        kind: "error",
      },
      operationId: workbookRecordHistoryOperationId(1),
      type: "operation_rejected",
    });
    expect(stale).toBe(state);
    state = workbookRecordHistoryReducer(state, {
      operationId,
      subject: {
        ...liveSubject,
        kind: "deleted",
        rowVersion: 5,
        stateLabel: "Deleted",
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
