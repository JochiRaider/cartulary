import { describe, expect, it, vi } from "vitest";

import {
  assertGridRows,
  buildGridPresentationRows,
  cleanupGridAdapters,
  type GridPresentationRow,
  type GridRow,
  isGridColumnEditable,
  navigateGridCellAnchor,
  resolveGridCellAnchor,
  resolveGridPasteTargets,
  resolveGridRenderer,
} from "./core";

type HarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function gridRow(
  key: string,
  state: string | null | undefined,
): GridRow<HarnessRow> {
  return {
    key,
    recordId: key,
    data: {
      label: key,
      state,
    },
  };
}

function draftRow(
  key: string,
  state: string | null | undefined,
): GridRow<HarnessRow> {
  return {
    key,
    recordId: null,
    data: {
      label: key,
      state,
    },
  };
}

function summarizeRows(
  rows: readonly GridPresentationRow<HarnessRow>[],
): readonly string[] {
  return rows.map((row) =>
    row.kind === "group"
      ? `group:${row.groupLabel ?? "null"}:${row.key}:${row.testId ?? "no-test-id"}`
      : `data:${row.key}`,
  );
}

describe("grid presentation rows", () => {
  it("maps rows directly to data presentation rows without grouping", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];

    expect(
      buildGridPresentationRows({
        groupBy: null,
        rows,
      }),
    ).toEqual([
      {
        gridRow: rows[0],
        key: "record-1",
        kind: "data",
      },
      {
        gridRow: rows[1],
        key: "record-2",
        kind: "data",
      },
    ]);

    expect(
      buildGridPresentationRows({
        groupBy: "state",
        rows,
      }),
    ).toEqual([
      {
        gridRow: rows[0],
        key: "record-1",
        kind: "data",
      },
      {
        gridRow: rows[1],
        key: "record-2",
        kind: "data",
      },
    ]);
  });

  it("builds one group bucket per normalized committed label", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", "open"),
      gridRow("record-3", "reviewed"),
      gridRow("record-4", "reviewed"),
      gridRow("record-5", "open"),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row, fieldKey) =>
        fieldKey === "state" ? row.state : null,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "data:record-5",
      "group:reviewed:group:state:reviewed:0:group-state-reviewed",
      "data:record-3",
      "data:record-4",
    ]);
  });

  it("normalizes committed empty group labels to the unassigned group without test IDs", () => {
    const rows = [
      gridRow("record-1", null),
      gridRow("record-2", undefined),
      gridRow("record-3", "   "),
      gridRow("record-4", ""),
      gridRow("record-5", "closed"),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:null:group:state:empty:0:no-test-id",
      "data:record-1",
      "data:record-2",
      "data:record-3",
      "data:record-4",
      "group:closed:group:state:closed:0:group-state-closed",
      "data:record-5",
    ]);
  });

  it("creates an unassigned group when empty labels follow a non-empty group", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", " "),
      gridRow("record-3", null),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "group:null:group:state:empty:0:no-test-id",
      "data:record-2",
      "data:record-3",
    ]);
  });

  it("keeps recordless draft rows outside grouped committed result buckets", () => {
    const rows = [
      gridRow("record-1", "open"),
      draftRow("draft-1", "rough"),
      gridRow("record-2", "reviewed"),
      gridRow("record-3", null),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "group:reviewed:group:state:reviewed:0:group-state-reviewed",
      "data:record-2",
      "group:null:group:state:empty:0:no-test-id",
      "data:record-3",
      "data:draft-1",
    ]);
  });

  it("trims labels before comparing groups and generating keys", () => {
    const rows = [
      gridRow("record-1", " open "),
      gridRow("record-2", "open"),
      gridRow("record-3", " reviewed "),
    ];

    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "group:reviewed:group:state:reviewed:0:group-state-reviewed",
      "data:record-3",
    ]);
  });
});

describe("grid Cartulary anchors", () => {
  const columns = [
    { fieldKey: "summary", label: "Summary", renderCell: () => null },
    { fieldKey: "state", label: "State", renderCell: () => null },
  ] as const;

  it("FE-U-P3-01 Reject unsafe record identity and keep presentation rows from mutation-capable anchors.", () => {
    expect(() =>
      assertGridRows([{ recordId: "record-1" }, { recordId: " " }]),
    ).toThrow(/missing record_id/i);
    expect(() =>
      assertGridRows([{ recordId: "record-1" }, { recordId: "record-1" }]),
    ).toThrow(/duplicate record_id/i);

    const rows = [
      gridRow("record-1", "open"),
      draftRow("draft-1", "rough"),
      gridRow("record-2", "reviewed"),
    ];
    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      groupBy: "state",
      rows,
    });
    const groupIndex = presentationRows.findIndex(
      (row) => row.kind === "group",
    );
    const draftIndex = presentationRows.findIndex(
      (row) => row.kind === "data" && row.gridRow.recordId === null,
    );

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: groupIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: draftIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: { recordId: "record-2", fieldKey: "summary" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toBeNull();
  });

  it("FE-U-P3-02 Translate vendor row and column coordinates to stable record_id and field_key anchors.", () => {
    const rows = [
      gridRow("record-3", "closed"),
      gridRow("record-1", "open"),
      gridRow("record-2", "reviewed"),
    ];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual({
      fieldKey: "state",
      recordId: "record-1",
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "state", recordId: "record-1" });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({ fieldKey: "summary", recordId: "record-2" });
    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        { kind: "record", recordId: "record-1" },
        { kind: "record", recordId: "record-2" },
      ],
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-1", fieldKey: "state" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        { kind: "record", recordId: "record-1" },
        { kind: "record", recordId: "record-2" },
      ],
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "", fieldKey: "state" },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });

  it("FE-U-P3-03 Resolve grid editability from explicit editor adapters and contract writeability only.", () => {
    const editorAdapter = {
      renderEditor: (row: HarnessRow) => row.label,
    };
    const writableColumn = {
      contractWritable: true,
      editorAdapter,
      fieldKey: "summary",
      label: "Summary",
      renderCell: (row: HarnessRow) => row.label,
    };
    const readOnlyColumn = {
      contractWritable: false,
      editorAdapter,
      fieldKey: "state",
      label: "State",
      renderCell: (row: HarnessRow) => row.state,
    };
    const adapterlessColumn = {
      contractWritable: true,
      fieldKey: "details",
      label: "Details",
      renderCell: () => null,
    };
    const vendorEditableColumn = {
      editable: true,
      fieldKey: "vendor",
      label: "Vendor",
      renderCell: () => null,
    } as typeof adapterlessColumn & { readonly editable: true };

    expect(isGridColumnEditable(writableColumn)).toBe(true);
    expect(isGridColumnEditable(readOnlyColumn)).toBe(false);
    expect(isGridColumnEditable(adapterlessColumn)).toBe(false);
    expect(isGridColumnEditable(vendorEditableColumn)).toBe(false);
  });

  it("FE-U-P3-04 Resolve renderers and editors deterministically and clean adapter-owned resources.", () => {
    const row = { label: "Alpha", state: "open" };
    const fallbackRenderer = { renderCell: (value: HarnessRow) => value.label };
    const valueKindRenderer = {
      renderCell: (value: HarnessRow) => `kind:${value.label}`,
    };
    const fieldRenderer = {
      renderCell: (value: HarnessRow) => `field:${value.label}`,
    };
    const columnRenderer = {
      renderCell: (value: HarnessRow) => `column:${value.label}`,
    };
    const registry = {
      fallbackRenderer,
      fieldRenderers: {
        summary: fieldRenderer,
      },
      valueKindRenderers: new Map([["short_text", valueKindRenderer]]),
    };

    expect(
      resolveGridRenderer({
        column: {
          fieldKey: "summary",
          label: "Summary",
          renderCell: () => null,
          rendererAdapter: columnRenderer,
          valueKind: "short_text",
        },
        registry,
      }).renderCell(row),
    ).toBe("column:Alpha");
    expect(
      resolveGridRenderer({
        column: {
          fieldKey: "summary",
          label: "Summary",
          renderCell: () => null,
          valueKind: "short_text",
        },
        registry,
      }).renderCell(row),
    ).toBe("field:Alpha");
    expect(
      resolveGridRenderer({
        column: {
          fieldKey: "details",
          label: "Details",
          renderCell: () => null,
          valueKind: "short_text",
        },
        registry,
      }).renderCell(row),
    ).toBe("kind:Alpha");
    expect(
      resolveGridRenderer({
        column: {
          fieldKey: "details",
          label: "Details",
          renderCell: () => null,
        },
        registry,
      }),
    ).toBe(fallbackRenderer);

    const sharedCleanup = vi.fn();
    const editorCleanup = vi.fn();
    const rendererCleanup = vi.fn();
    const sharedAdapter = { cleanup: sharedCleanup, renderCell: () => null };
    cleanupGridAdapters([
      sharedAdapter,
      sharedAdapter,
      {
        editorAdapter: { cleanup: editorCleanup, renderEditor: () => null },
        fieldKey: "summary",
        label: "Summary",
        renderCell: () => null,
        rendererAdapter: {
          cleanup: rendererCleanup,
          renderCell: () => null,
        },
      },
    ]);

    expect(sharedCleanup).toHaveBeenCalledTimes(1);
    expect(editorCleanup).toHaveBeenCalledTimes(1);
    expect(rendererCleanup).toHaveBeenCalledTimes(1);
  });

  it("translates valid presentation coordinates into stable record_id and field_key anchors", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual({
      recordId: "record-2",
      fieldKey: "state",
    });
  });

  it("clears anchors for invalid row, field, group, and recordless targets", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", "closed"),
      draftRow("draft-1", "rough"),
    ];
    const presentationRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      getGroupRowTestId: (fieldKey, value) => `group-${fieldKey}-${value}`,
      groupBy: "state",
      rows,
    });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: { rowIndex: -1, fieldKey: "state" },
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
        selection: { rowIndex: 0, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: {
          rowIndex: presentationRows.length - 1,
          fieldKey: "summary",
        },
      }),
    ).toBeNull();
  });

  it("updates anchors for keyboard navigation and clears on presentation-only targets", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual({ recordId: "record-1", fieldKey: "state" });

    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "state" },
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({ recordId: "record-2", fieldKey: "state" });

    const groupedRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      groupBy: "state",
      rows,
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        intent: { key: "ArrowUp" },
        presentationRows: groupedRows,
      }),
    ).toBeNull();
  });

  it("does not treat vendor selection changes alone as anchor updates", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });
    const current = { recordId: "record-1", fieldKey: "summary" };
    const vendorSelection = { rowIndex: 1, fieldKey: "state" };

    expect(current).toEqual({ recordId: "record-1", fieldKey: "summary" });
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        selection: vendorSelection,
      }),
    ).toEqual({ recordId: "record-2", fieldKey: "state" });
  });

  it("targets sorted paste rows by stable visible record identities", () => {
    const rows = [
      gridRow("record-3", "closed"),
      gridRow("record-1", "open"),
      gridRow("record-2", "reviewed"),
    ];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        { kind: "record", recordId: "record-1" },
        { kind: "record", recordId: "record-2" },
      ],
    });
  });

  it("maps filtered overflow to explicit create-row anchors", () => {
    const rows = [gridRow("record-2", "reviewed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-2", fieldKey: "state" },
        pastedColumnCount: 2,
        pastedRowCount: 3,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        { kind: "record", recordId: "record-2" },
        { createIndex: 0, kind: "create" },
        { createIndex: 1, kind: "create" },
      ],
    });
  });

  it("rejects group and presentation-only paste anchors", () => {
    const rows = [
      gridRow("record-1", "open"),
      draftRow("draft-1", "rough"),
      gridRow("record-2", "reviewed"),
    ];
    const groupedRows = buildGridPresentationRows({
      getGroupLabel: (row) => row.state,
      groupBy: "state",
      rows,
    });

    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "record-1", fieldKey: "summary" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: { recordId: "record-2", fieldKey: "summary" },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
  });

  it("requires a Cartulary anchor instead of vendor coordinates alone", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });
    const vendorSelection = { rowIndex: 1, fieldKey: "state" };

    expect(
      resolveGridPasteTargets({
        columns,
        current: { recordId: "", fieldKey: vendorSelection.fieldKey },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });
});
