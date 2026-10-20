import { describe, expect, it } from "vitest";
import {
  historyDiffFixture,
  historyPresentationFixture,
} from "../../testing/workbookHistoryTestSupport";
import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import {
  acceptHistoryPage,
  beginHistoryRead,
  initialHistoryBrowsing,
  rejectHistoryRead,
} from "../history/workbookHistoryBrowsing";
import { buildRecordRollbackTargetFromHistoryAction } from "../history/workbookHistoryItem";
import type { HistoryPage } from "../history/workbookHistoryPage";
import {
  initialWorkbookRecordHistoryState,
  workbookRecordHistoryLoadedData,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
} from "./workbookRecordHistoryModel";

const subject = {
  kind: "live",
  label: "Record A",
  recordId: "record-a",
  rowVersion: 4,
  surfaceLabel: "Test records",
  viewSchemaId: "cartulary.view.test.v1",
} as const;
const page: HistoryPage = {
  deleted: false,
  incident_id: "incident-a",
  record_id: "record-a",
  row_version: 4,
  representation_generation: "cartulary.history.1",
  items: [],
  paging: { limit: 100, has_more: false, next_cursor: null },
};
const pending = {
  kind: "destructive",
  operation: "delete",
  recordId: subject.recordId,
  rowVersion: 4,
} as const;
const operationId = workbookRecordHistoryOperationId(1);

describe("record History presentation", () => {
  it("requires accepted browsing before review and rejects mismatched subjects", () => {
    const initial = initialWorkbookRecordHistoryState(subject);
    expect(
      workbookRecordHistoryReducer(initial, {
        type: "preview",
        pendingAction: pending,
      }),
    ).toBe(initial);
    expect(
      workbookRecordHistoryReducer(initial, { type: "submit", operationId }),
    ).toBe(initial);
    const other = historyPresentationFixture(
      { ...subject, recordId: "other" },
      { ...page, record_id: "other" },
    );
    expect(
      workbookRecordHistoryReducer(initial, {
        type: "browsing_changed",
        browsing: required(other.browsing),
      }),
    ).toBe(initial);
  });

  it("rejects stale read completion after retarget without keeping a second accepted value", () => {
    const ready = historyPresentationFixture(subject, page);
    const next = workbookRecordHistoryReducer(ready, {
      type: "retarget",
      subject: { ...subject, recordId: "other" },
    });
    const stale = workbookRecordHistoryReducer(next, {
      type: "browsing_changed",
      browsing: required(ready.browsing),
    });
    expect(stale).toBe(next);
    expect(workbookRecordHistoryLoadedData(stale)).toBeNull();
    const browsing = initialHistoryBrowsing(
      {
        actorId: "a",
        incidentId: "incident-a",
        sessionIdentity: "s",
        epoch: 0,
      },
      subject.recordId,
      subject.viewSchemaId,
    );
    const first = beginHistoryRead(browsing, "initial");
    const second = beginHistoryRead(first, "refresh");
    expect(
      acceptHistoryPage(second, required(first.pending), page, {
        rowVersion: 4,
        deleted: false,
      }),
    ).toBe(second);
  });

  it("retains the accepted page and provenance through failed refresh", () => {
    const ready = historyPresentationFixture(subject, page);
    const requested = beginHistoryRead(required(ready.browsing), "refresh");
    const failed = rejectHistoryRead(requested, required(requested.pending), {
      kind: "retryable",
      message: "Offline",
    });
    const state = workbookRecordHistoryReducer(ready, {
      type: "browsing_changed",
      browsing: failed,
    });
    expect(workbookRecordHistoryLoadedData(state)).toBe(
      required(required(ready.browsing).accepted).data,
    );
    expect(required(required(state.browsing).accepted).provenance).toBe(
      required(required(ready.browsing).accepted).provenance,
    );
    expect(required(state.browsing).failure?.error.message).toBe("Offline");
    expect(state).not.toHaveProperty("result");
    expect(state).not.toHaveProperty("retainedData");
  });

  it("invalidates unsubmitted review on refresh or a newer source version", () => {
    const ready = historyPresentationFixture(subject, page);
    const review = workbookRecordHistoryReducer(ready, {
      type: "preview",
      pendingAction: pending,
    });
    const requested = beginHistoryRead(required(ready.browsing), "refresh");
    expect(
      workbookRecordHistoryPendingAction(
        workbookRecordHistoryReducer(review, {
          type: "browsing_changed",
          browsing: requested,
        }),
      ),
    ).toBeNull();
    expect(
      workbookRecordHistoryPendingAction(
        workbookRecordHistoryReducer(review, {
          type: "retarget",
          subject: { ...subject, rowVersion: 5 },
        }),
      ),
    ).toBeNull();
  });

  it("cancels review without weakening captured operation identity", () => {
    const ready = historyPresentationFixture(subject, page);
    const review = workbookRecordHistoryReducer(ready, {
      type: "preview",
      pendingAction: pending,
    });
    expect(
      workbookRecordHistoryPendingAction(
        workbookRecordHistoryReducer(review, { type: "cancel" }),
      ),
    ).toBeNull();
    const submitting = workbookRecordHistoryReducer(review, {
      type: "submit",
      operationId,
    });
    expect(
      workbookRecordHistoryReducer(submitting, { type: "submit", operationId }),
    ).toBe(submitting);
    for (const values of [
      {
        operationId: workbookRecordHistoryOperationId(2),
        recordId: "record-a",
        rowVersion: 5,
      },
      { operationId, recordId: "wrong", rowVersion: 5 },
      { operationId, recordId: "record-a", rowVersion: 0 },
      { operationId, recordId: "record-a", rowVersion: 5.5 },
    ]) {
      expect(
        workbookRecordHistoryReducer(submitting, {
          type: "operation_accepted",
          ...values,
        }),
      ).toBe(submitting);
    }
    const accepted = workbookRecordHistoryReducer(submitting, {
      type: "operation_accepted",
      operationId,
      recordId: "record-a",
      rowVersion: 5,
    });
    expect(accepted.subject).toMatchObject({ kind: "deleted", rowVersion: 5 });
    expect(accepted.submission).toBeUndefined();
    expect(accepted.browsing?.accepted?.data.paging).toEqual(page.paging);
  });

  it("derives restored subject state from the captured operation", () => {
    const deleted = {
      ...subject,
      kind: "deleted" as const,
      stateLabel: "Deleted",
    };
    const ready = historyPresentationFixture(deleted, {
      ...page,
      deleted: true,
    });
    const review = workbookRecordHistoryReducer(ready, {
      type: "preview",
      pendingAction: { ...pending, operation: "restore" },
    });
    const submitting = workbookRecordHistoryReducer(review, {
      type: "submit",
      operationId,
    });
    const accepted = workbookRecordHistoryReducer(submitting, {
      type: "operation_accepted",
      operationId,
      recordId: "record-a",
      rowVersion: 5,
    });
    expect(accepted.subject).toMatchObject({ kind: "live", rowVersion: 5 });
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
      diff_summary: historyDiffFixture("Changed"),
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

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Missing expected test value");
  return value;
}
