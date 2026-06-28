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

    const selected: string[] = [];
    const rows = [
      workbookRow("draft", null, 0),
      workbookRow("row-a", "record-a", 2),
      workbookRow("row-b", "record-b", 3),
    ];
    const gridRows = buildTimelineGridRows({
      onSelectRow: (recordId) => {
        selected.push(recordId);
      },
      presenceForRow: (recordId) =>
        recordId === "record-b" ? ["presence-b"] : [],
      renderDraftGutterContent: (row) => `draft:${row.key}`,
      renderSavedGutterContent: ({ ordinal, presences, recordId }) =>
        `${ordinal}:${recordId}:${presences.join(",")}`,
      rows,
      selectedRowId: "record-b",
    });

    expect(gridRows.map((row) => row.testId)).toEqual([
      "cartulary.view.timeline.v2-inline-draft-row",
      "grid-row-cartulary.view.timeline.v2-record-a",
      "grid-row-cartulary.view.timeline.v2-record-b",
    ]);
    expect(gridRows.map((row) => row.gutterLabel)).toEqual(["+", "2", "3"]);
    expect(gridRows[0]).toMatchObject({
      recordId: null,
      selected: false,
      variant: "draft",
    });
    expect(gridRows[2]).toMatchObject({
      recordId: "record-b",
      selected: true,
      gutterContent: "3:record-b:presence-b",
      gutterTestId: "cartulary.view.timeline.v2-row-gutter-record-b",
      variant: "default",
    });
    gridRows[0]?.onSelect?.({} as never);
    gridRows[1]?.onSelect?.({} as never);
    expect(selected).toEqual(["record-a"]);
  });
});
