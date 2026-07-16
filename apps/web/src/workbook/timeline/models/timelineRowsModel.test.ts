import { describe, expect, it } from "vitest";
import {
  buildTimelineGridRows,
  decideWorkbookRecordFreshness,
} from "./timelineRowsModel";
import type { WorkbookRow } from "./workbookTimelineModel";

function workbookRow(
  key: string,
  recordId: string | null,
  rowVersion = 1,
): WorkbookRow {
  return {
    key,
    recordId,
    rowVersion,
    viewSchemaId: "cartulary.view.timeline.v2",
    captureState: "rough",
    values: {},
    committedValues: {},
    collectionValues: {
      hostRefs: [],
      identityRefs: [],
      tags: [],
    },
    collectionDrafts: {
      hostRefs: "",
      identityRefs: "",
      tags: "",
    },
    pendingSignature: null,
    rawRow: null,
  } as unknown as WorkbookRow;
}

describe("timelineRowsModel", () => {
  it("classifies stale row-version updates only when durable identity is comparable", () => {
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: 2 }, 3),
    ).toEqual({ comparable: true, stale: true });
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: 3 }, 3),
    ).toEqual({ comparable: true, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: null, rowVersion: 1 }, 3),
    ).toEqual({ comparable: false, stale: false });
    expect(
      decideWorkbookRecordFreshness({ recordId: "row-1", rowVersion: null }, 3),
    ).toEqual({ comparable: false, stale: false });
    expect(
      decideWorkbookRecordFreshness(
        { recordId: "row-1", rowVersion: 1 },
        undefined,
      ),
    ).toEqual({ comparable: false, stale: false });

    const rows = [
      workbookRow("draft", null, 0),
      workbookRow("row-a", "record-a", 2),
      workbookRow("row-b", "record-b", 3),
    ];
    const gridRows = buildTimelineGridRows({
      presenceForRow: (recordId) =>
        recordId === "record-b" ? ["presence-b"] : [],
      renderDraftGutterContent: (row) => `draft:${row.key}`,
      renderSavedGutterContent: ({ ordinal, presences, recordId }) =>
        `${ordinal}:${recordId}:${presences.join(",")}`,
      rows,
    });

    expect(gridRows.draftRow?.testId).toBe(
      "cartulary.view.timeline.v2-inline-draft-row",
    );
    expect(gridRows.recordRows.map((row) => row.testId)).toEqual([
      "grid-row-cartulary.view.timeline.v2-record-a",
      "grid-row-cartulary.view.timeline.v2-record-b",
    ]);
    expect(gridRows.draftRow).toMatchObject({
      kind: "draft",
      gutterLabel: "+",
    });
    expect(gridRows.recordRows.map((row) => row.gutterLabel)).toEqual([
      "2",
      "3",
    ]);
    expect(gridRows.recordRows[1]).toMatchObject({
      kind: "record",
      recordId: "record-b",
      rowVersion: 3,
      gutterContent: "3:record-b:presence-b",
      gutterTestId: "cartulary.view.timeline.v2-row-gutter-record-b",
    });
  });
});
