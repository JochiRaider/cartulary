import { describe, expect, it, vi } from "vitest";
import {
  buildTimelineGridRows,
  ensureTimelineDraftRow,
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
  it("allocates one draft row only when the Timeline row set needs one", () => {
    const nextDraftIndex = vi.fn(() => 7);
    const committedRows = [workbookRow("row-a", "record-a", 2)];
    const added = ensureTimelineDraftRow({
      nextDraftIndex,
      rows: committedRows,
    });

    expect(nextDraftIndex).toHaveBeenCalledOnce();
    expect(added.rows.map((row) => row.key)).toEqual(["row-a", "draft-7"]);
    expect(added.draftFocusKey).toBe("draft-7:activitySynopsisText:grid");

    const existingRows = [
      ...committedRows,
      workbookRow("draft-existing", null, 0),
    ];
    const retained = ensureTimelineDraftRow({
      nextDraftIndex,
      rows: existingRows,
    });
    expect(nextDraftIndex).toHaveBeenCalledOnce();
    expect(retained).toEqual({ rows: existingRows, draftFocusKey: null });
  });

  it("builds draft and committed rows with stable semantic identities", () => {
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
      kind: "data",
      mutationIdentity: {
        kind: "core_row_version",
        baseRowVersion: 3,
      },
      rowIdentity: { kind: "core_record", recordId: "record-b" },
      gutterContent: "3:record-b:presence-b",
      gutterTestId: "cartulary.view.timeline.v2-row-gutter-record-b",
    });
  });
});
