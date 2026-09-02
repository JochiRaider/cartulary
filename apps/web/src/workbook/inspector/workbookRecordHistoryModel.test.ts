import { describe, expect, it } from "vitest";
import {
  buildRecordRollbackTargetFromHistoryAction,
  initialWorkbookRecordHistoryState,
  type RecordHistoryData,
  type RecordHistoryItem,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryPendingAction,
  type WorkbookRecordHistoryState,
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
        recordId: liveSubject.recordId,
        rowVersion: liveSubject.rowVersion + 1,
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

  it("routes every non-global event through only its legal phase transition", () => {
    const requestId = workbookRecordHistoryRequestId(11);
    const operationId = workbookRecordHistoryOperationId(12);
    const pendingAction: WorkbookRecordHistoryPendingAction = {
      kind: "destructive",
      operation: "delete",
      recordId: liveSubject.recordId,
      rowVersion: liveSubject.rowVersion,
    };
    const feedback = {
      announcement: "polite" as const,
      kind: "message" as const,
      message: "Completed.",
    };
    const loading: WorkbookRecordHistoryState = {
      phase: "loading",
      requestId,
      subject: liveSubject,
    };
    const ready: WorkbookRecordHistoryState = {
      feedback,
      pendingAction,
      phase: "ready",
      result: { data, kind: "loaded" },
      subject: liveSubject,
    };
    const submitting: WorkbookRecordHistoryState = {
      data,
      operation: { pendingAction },
      operationId,
      phase: "submitting",
      subject: liveSubject,
    };
    const states = {
      idle: { feedback, phase: "idle", subject: liveSubject },
      loading,
      ready,
      submitting,
    } as const satisfies Record<
      WorkbookRecordHistoryState["phase"],
      WorkbookRecordHistoryState
    >;
    const events = {
      cancel: { type: "cancel" },
      feedback_cleared: { type: "feedback_cleared" },
      load_accepted: {
        data,
        requestId,
        subject: liveSubject,
        type: "load_accepted",
      },
      load_rejected: {
        error: { primaryMessage: "Rejected", technicalFields: [] },
        requestId,
        subject: liveSubject,
        type: "load_rejected",
      },
      load_requested: {
        requestId,
        subject: liveSubject,
        type: "load_requested",
      },
      operation_accepted: {
        operationId,
        recordId: liveSubject.recordId,
        rowVersion: liveSubject.rowVersion + 1,
        type: "operation_accepted",
      },
      operation_rejected: {
        feedback,
        operationId,
        type: "operation_rejected",
      },
      preview: { pendingAction, type: "preview" },
      submit: { operationId, type: "submit" },
    } as const satisfies Record<
      Exclude<WorkbookRecordHistoryEvent["type"], "clear" | "retarget">,
      WorkbookRecordHistoryEvent
    >;
    const allowed: Record<
      WorkbookRecordHistoryState["phase"],
      ReadonlySet<string>
    > = {
      idle: new Set(["feedback_cleared", "load_requested"]),
      loading: new Set(["load_accepted", "load_rejected", "load_requested"]),
      ready: new Set([
        "cancel",
        "feedback_cleared",
        "load_requested",
        "preview",
        "submit",
      ]),
      submitting: new Set([
        "load_requested",
        "operation_accepted",
        "operation_rejected",
      ]),
    } as const;

    for (const [phase, state] of Object.entries(states)) {
      for (const [eventType, event] of Object.entries(events)) {
        const next = workbookRecordHistoryReducer(state, event);
        if (
          allowed[phase as WorkbookRecordHistoryState["phase"]].has(eventType)
        ) {
          expect(next, `${phase} accepts ${eventType}`).not.toBe(state);
        } else {
          expect(next, `${phase} rejects ${eventType}`).toBe(state);
        }
      }
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
    for (const event of [
      {
        operationId: workbookRecordHistoryOperationId(1),
        recordId: liveSubject.recordId,
        rowVersion: 5,
        type: "operation_accepted" as const,
      },
      {
        operationId,
        recordId: "record-b",
        rowVersion: 5,
        type: "operation_accepted" as const,
      },
      {
        operationId,
        recordId: liveSubject.recordId,
        rowVersion: 0,
        type: "operation_accepted" as const,
      },
      {
        operationId,
        recordId: liveSubject.recordId,
        rowVersion: 5.5,
        type: "operation_accepted" as const,
      },
    ]) {
      expect(workbookRecordHistoryReducer(state, event)).toBe(state);
    }
    state = workbookRecordHistoryReducer(state, {
      operationId,
      recordId: liveSubject.recordId,
      rowVersion: 5,
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

  it("derives delete, restore, and rollback subject kinds from captured operations", () => {
    const operationId = workbookRecordHistoryOperationId(21);
    const deletedSubject = {
      ...liveSubject,
      kind: "deleted" as const,
      stateLabel: "Deleted",
    };
    const cases = [
      {
        expectedKind: "deleted",
        pendingAction: {
          kind: "destructive",
          operation: "delete",
          recordId: liveSubject.recordId,
          rowVersion: liveSubject.rowVersion,
        },
        subject: liveSubject,
      },
      {
        expectedKind: "live",
        pendingAction: {
          kind: "destructive",
          operation: "restore",
          recordId: deletedSubject.recordId,
          rowVersion: deletedSubject.rowVersion,
        },
        subject: deletedSubject,
      },
      {
        expectedKind: "deleted",
        pendingAction: {
          action: "history_entry",
          historyItemRef: "item-a",
          kind: "rollback",
          recordId: deletedSubject.recordId,
          rowVersion: deletedSubject.rowVersion,
          target: {
            history_entry_ref: "entry-a",
            kind: "history_entry",
          },
        },
        subject: deletedSubject,
      },
    ] as const;

    for (const testCase of cases) {
      const submitting: WorkbookRecordHistoryState = {
        data: { ...data, deleted: testCase.subject.kind === "deleted" },
        operation: { pendingAction: testCase.pendingAction },
        operationId,
        phase: "submitting",
        subject: testCase.subject,
      };
      const accepted = workbookRecordHistoryReducer(submitting, {
        operationId,
        recordId: testCase.subject.recordId,
        rowVersion: testCase.subject.rowVersion + 1,
        type: "operation_accepted",
      });
      expect(accepted).toMatchObject({
        phase: "idle",
        subject: {
          kind: testCase.expectedKind,
          recordId: testCase.subject.recordId,
          rowVersion: testCase.subject.rowVersion + 1,
        },
      });
    }

    const malformed: WorkbookRecordHistoryState = {
      data,
      operation: {
        pendingAction: {
          kind: "destructive",
          operation: "delete",
          recordId: "record-b",
          rowVersion: liveSubject.rowVersion,
        },
      },
      operationId,
      phase: "submitting",
      subject: liveSubject,
    };
    expect(
      workbookRecordHistoryReducer(malformed, {
        operationId,
        recordId: "record-b",
        rowVersion: 5,
        type: "operation_accepted",
      }),
    ).toBe(malformed);
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
