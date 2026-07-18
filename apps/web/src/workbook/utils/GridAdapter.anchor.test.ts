import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridDataRow,
  type GridDraftRow,
  navigateGridCellAnchor,
  resolveGridCellAnchor,
} from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";

type AnchorRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function savedRow(
  recordId: string,
  state: string | null | undefined,
): GridDataRow<AnchorRow> {
  return {
    kind: "data",
    mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
    rowIdentity: { kind: "core_record", recordId },
    data: {
      label: recordId,
      state,
    },
  };
}

function draftRow(key: string): GridDraftRow<AnchorRow> {
  return {
    kind: "draft",
    data: {
      label: key,
      state: "draft",
    },
  };
}

const columns: readonly GridColumn<AnchorRow>[] = [
  { fieldKey: "task.title", label: "Title", renderCell: () => null },
  { fieldKey: "task.status", label: "Status", renderCell: () => null },
  { fieldKey: "task.priority", label: "Priority", renderCell: () => null },
];

const surface = { kind: "view_schema", viewSchemaId: "test.view" } as const;

function anchor(recordId: string, fieldKey: string): GridCellAnchor {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface,
  };
}

describe("adapter-owned Cartulary anchor evidence", () => {
  it("resolves vendor coordinates to stable record_id and field_key anchors", () => {
    const rows = [savedRow("record-1", "open"), savedRow("record-2", "done")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface,
        selection: { rowIndex: 1, fieldKey: "task.status" },
      }),
    ).toEqual(anchor("record-2", "task.status"));
  });

  it("clears invalid row, field, group row, and recordless draft targets", () => {
    const rows = [
      savedRow("record-1", "open"),
      savedRow("record-2", "done"),
      draftRow("draft-1"),
    ];
    const presentationRows = buildGridPresentationRows({
      grouping: {
        fieldKey: "state",
        formatLabel: (value) => (value === null ? null : String(value)),
        getValue: (row) => row.state ?? null,
      },
      rows,
    });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface,
        selection: { rowIndex: -1, fieldKey: "task.status" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface,
        selection: { rowIndex: 1, fieldKey: "__cartulary_actions__" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface,
        selection: { rowIndex: 0, fieldKey: "task.title" },
      }),
    ).toBeNull();
    expect(
      presentationRows.every(
        (row) =>
          row.kind !== "data" ||
          row.gridRow.rowIdentity.kind !== "core_record" ||
          row.gridRow.rowIdentity.recordId !== "draft-1",
      ),
    ).toBe(true);
  });

  it("resolves Arrow, Tab, Enter, and Shift+Enter navigation through adapter anchors", () => {
    const rows = [
      savedRow("record-1", "open"),
      savedRow("record-2", "open"),
      savedRow("record-3", "open"),
    ];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      navigateGridCellAnchor({
        columns,
        current: anchor("record-1", "task.title"),
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual(anchor("record-1", "task.status"));
    expect(
      navigateGridCellAnchor({
        columns,
        current: anchor("record-1", "task.status"),
        intent: { key: "Tab" },
        presentationRows,
      }),
    ).toEqual(anchor("record-1", "task.priority"));
    expect(
      navigateGridCellAnchor({
        columns,
        current: anchor("record-1", "task.status"),
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual(anchor("record-2", "task.status"));
    expect(
      navigateGridCellAnchor({
        columns,
        current: anchor("record-2", "task.status"),
        intent: { key: "Enter", shiftKey: true },
        presentationRows,
      }),
    ).toEqual(anchor("record-1", "task.status"));
  });

  it("keeps vendor selection separate until translated by the adapter contract", () => {
    const rows = [savedRow("record-1", "open"), savedRow("record-2", "done")];
    const presentationRows = buildGridPresentationRows({ rows });
    const currentAnchor = anchor("record-1", "task.title");
    const vendorSelection = { fieldKey: "task.status", rowIndex: 1 };

    expect(currentAnchor).toEqual(anchor("record-1", "task.title"));
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface,
        selection: vendorSelection,
      }),
    ).toEqual(anchor("record-2", "task.status"));
  });
});
