import { describe, expect, it } from "vitest";
import { reconcileWorkbookRecordRows } from "./workbookRowReconciliation";

describe("workbook row reconciliation", () => {
  it("Verify sparse patches preserve unchanged row object references and intentionally replace changed rows by record_id.", () => {
    const stable = { recordId: "record-1", state: "open" };
    const changed = { recordId: "record-2", state: "open" };
    const removed = { recordId: "record-3", state: "closed" };
    const nextChanged = { recordId: "record-2", state: "closed" };

    const reconciled = reconcileWorkbookRecordRows(
      [stable, changed, removed],
      [
        { recordId: "record-1", state: "open" },
        nextChanged,
        { recordId: "record-4", state: "reviewed" },
      ],
    );

    expect(reconciled.map((row) => row.recordId)).toEqual([
      "record-1",
      "record-2",
      "record-4",
    ]);
    expect(reconciled[0]).toBe(stable);
    expect(reconciled[1]).toBe(nextChanged);
    expect(reconciled).not.toContain(removed);
    const previousDraft = { recordId: null, state: "local" };
    const nextDraft = { recordId: null, state: "incoming" };
    expect(reconcileWorkbookRecordRows([previousDraft], [nextDraft])[0]).toBe(
      nextDraft,
    );
    const stableVersioned = {
      recordId: "record-versioned",
      rowVersion: 7,
      cells: { title: { value: "Stable" } },
    };
    const sameVersion = {
      recordId: "record-versioned",
      rowVersion: 7,
      cells: { title: { value: "Stable" } },
    };
    const nextVersion = {
      recordId: "record-versioned",
      rowVersion: 8,
      cells: { title: { value: "Changed" } },
    };

    expect(
      reconcileWorkbookRecordRows([stableVersioned], [sameVersion])[0],
    ).toBe(stableVersioned);
    expect(
      reconcileWorkbookRecordRows([stableVersioned], [nextVersion])[0],
    ).toBe(nextVersion);
  });
});
