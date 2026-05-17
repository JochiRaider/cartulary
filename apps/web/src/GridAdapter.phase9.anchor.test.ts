import {
  buildGridPresentationRows,
  type GridColumn,
  type GridRow,
  navigateGridCellAnchor,
  resolveGridCellAnchor,
} from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";

type Phase9AnchorRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function savedRow(
  recordId: string,
  state: string | null | undefined,
): GridRow<Phase9AnchorRow> {
  return {
    key: recordId,
    recordId,
    data: {
      label: recordId,
      state,
    },
  };
}

function draftRow(key: string): GridRow<Phase9AnchorRow> {
  return {
    key,
    recordId: null,
    data: {
      label: key,
      state: "draft",
    },
  };
}

const columns: readonly GridColumn<Phase9AnchorRow>[] = [
  { fieldKey: "task.title", label: "Title", renderCell: () => null },
  { fieldKey: "task.status", label: "Status", renderCell: () => null },
  { fieldKey: "task.priority", label: "Priority", renderCell: () => null },
];

describe("Phase 9 U-9-GRID-01 adapter-owned Cartulary anchor evidence", () => {
  it("Phase 9 U-9-GRID-01 resolves vendor coordinates to stable record_id and field_key anchors", () => {
    const rows = [savedRow("record-1", "open"), savedRow("record-2", "done")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: 1, fieldKey: "task.status" },
      }),
    ).toEqual({
      fieldKey: "task.status",
      recordId: "record-2",
    });
  });

  it("Phase 9 U-9-GRID-01 clears invalid row, field, group row, and recordless draft targets", () => {
    const rows = [
      savedRow("record-1", "open"),
      savedRow("record-2", "done"),
      draftRow("draft-1"),
    ];
    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      groupBy: "state",
      rows,
    });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: -1, fieldKey: "task.status" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: 1, fieldKey: "__cartulary_actions__" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: 0, fieldKey: "task.title" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: {
          rowIndex: presentationRows.length - 1,
          fieldKey: "task.title",
        },
      }),
    ).toBeNull();
  });

  it("Phase 9 U-9-GRID-01 resolves Arrow, Tab, Enter, and Shift+Enter navigation through adapter anchors", () => {
    const rows = [
      savedRow("record-1", "open"),
      savedRow("record-2", "open"),
      savedRow("record-3", "open"),
    ];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "task.title" },
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "task.status", recordId: "record-1" });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "task.status" },
        intent: { key: "Tab" },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "task.priority", recordId: "record-1" });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "task.status" },
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "task.status", recordId: "record-2" });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-2", fieldKey: "task.status" },
        intent: { key: "Enter", shiftKey: true },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "task.status", recordId: "record-1" });
  });

  it("Phase 9 U-9-GRID-01 keeps vendor selection separate until translated by the adapter contract", () => {
    const rows = [savedRow("record-1", "open"), savedRow("record-2", "done")];
    const presentationRows = buildGridPresentationRows({ rows });
    const currentAnchor = { fieldKey: "task.title", recordId: "record-1" };
    const vendorSelection = { fieldKey: "task.status", rowIndex: 1 };

    expect(currentAnchor).toEqual({
      fieldKey: "task.title",
      recordId: "record-1",
    });
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: vendorSelection,
      }),
    ).toEqual({
      fieldKey: "task.status",
      recordId: "record-2",
    });
  });
});
