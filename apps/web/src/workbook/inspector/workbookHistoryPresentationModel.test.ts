import { describe, expect, it } from "vitest";
import {
  workbookHistoryEventPresentation,
  workbookHistoryPendingTechnicalFields,
  workbookHistoryRollbackLabel,
} from "./workbookHistoryPresentationModel";

describe("workbook history presentation model", () => {
  it("keeps actor identifiers technical and orders fields consistently", () => {
    const item = {
      actor_user_id: "40000000-0000-4000-8000-000000000001",
      available_rollback_actions: ["history_entry" as const],
      change_set_id: "change-set-1",
      committed_at: "2026-08-30T20:00:00Z",
      diff_summary: { summary: "Changed title", units: [] },
      history_entry_ref: "entry-1",
      history_item_ref: "history-1",
      operation: "patch",
      reversible: true,
      revision_no: 4,
    };

    expect(workbookHistoryEventPresentation(item)).toEqual({
      committedAt: "2026-08-30T20:00:00Z",
      key: "history-1",
      operation: "patch",
      summary: "Changed title",
      technicalFields: [
        { label: "Actor ID", value: item.actor_user_id },
        { label: "History reference", value: "history-1" },
        { label: "Change set ID", value: "change-set-1" },
        { label: "History entry", value: "entry-1" },
        { label: "Revision", value: "4" },
      ],
    });
    expect(
      workbookHistoryEventPresentation(item, "Alex Analyst").actorLabel,
    ).toBe("Alex Analyst");
  });

  it("shares rollback labels and pending field order", () => {
    expect(workbookHistoryRollbackLabel("history_entry")).toBe(
      "Rollback entry",
    );
    expect(workbookHistoryRollbackLabel("change_set")).toBe(
      "Rollback change set",
    );
    expect(workbookHistoryRollbackLabel("row_restore")).toBe(
      "Restore row fields",
    );
    expect(
      workbookHistoryPendingTechnicalFields({
        recordId: "record-1",
        rowVersion: null,
      }),
    ).toEqual([
      { label: "Record ID", value: "record-1" },
      { label: "Row version", value: "unknown" },
    ]);
  });
});
