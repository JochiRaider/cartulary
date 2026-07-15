import { describe, expect, it } from "vitest";

import {
  assertGridRows,
  buildGridPresentationRows,
  type GridDraftRow,
  type GridPresentationRow,
  type GridRecordRow,
  isGridColumnEditable,
  navigateGridCellAnchor,
  resolveGridCellAnchor,
  resolveGridPasteTargets,
} from "./core";

type HarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function gridRow(
  key: string,
  state: string | null | undefined,
): GridRecordRow<HarnessRow> {
  return {
    kind: "record",
    recordId: key,
    rowVersion: 1,
    data: {
      label: key,
      state,
    },
  };
}

function draftRow(
  key: string,
  state: string | null | undefined,
): GridDraftRow<HarnessRow> {
  return {
    kind: "draft",
    data: {
      label: key,
      state,
    },
  };
}

function stateGrouping(withTestId = true) {
  return {
    fieldKey: "state",
    formatLabel: (value: boolean | number | string | null) =>
      value === null ? null : String(value),
    getTestId: withTestId
      ? (
          fieldKey: string,
          _value: boolean | number | string | null,
          label: string | null,
        ) => (label === null ? undefined : `group-${fieldKey}-${label}`)
      : undefined,
    getValue: (row: HarnessRow) => {
      const value = row.state?.trim();
      return value === undefined || value === "" ? null : value;
    },
  } as const;
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
      grouping: stateGrouping(),
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:s:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "data:record-5",
      "group:reviewed:group:state:s:reviewed:0:group-state-reviewed",
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
      grouping: stateGrouping(),
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:null:group:state:n:null:0:no-test-id",
      "data:record-1",
      "data:record-2",
      "data:record-3",
      "data:record-4",
      "group:closed:group:state:s:closed:0:group-state-closed",
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
      grouping: stateGrouping(),
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:s:open:0:group-state-open",
      "data:record-1",
      "group:null:group:state:n:null:0:no-test-id",
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
      grouping: stateGrouping(),
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:s:open:0:group-state-open",
      "data:record-1",
      "group:reviewed:group:state:s:reviewed:0:group-state-reviewed",
      "data:record-2",
      "group:null:group:state:n:null:0:no-test-id",
      "data:record-3",
    ]);
  });

  it("trims labels before comparing groups and generating keys", () => {
    const rows = [
      gridRow("record-1", " open "),
      gridRow("record-2", "open"),
      gridRow("record-3", " reviewed "),
    ];

    const presentationRows = buildGridPresentationRows({
      grouping: stateGrouping(),
      rows,
    });

    expect(summarizeRows(presentationRows)).toEqual([
      "group:open:group:state:s:open:0:group-state-open",
      "data:record-1",
      "data:record-2",
      "group:reviewed:group:state:s:reviewed:0:group-state-reviewed",
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
      grouping: stateGrouping(false),
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
        viewSchemaId: "test.view",
        selection: { rowIndex: groupIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        viewSchemaId: "test.view",
        selection: { rowIndex: draftIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-2",
          fieldKey: "summary",
        },
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
        viewSchemaId: "test.view",
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual({
      viewSchemaId: "test.view",
      fieldKey: "state",
      recordId: "record-1",
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual({
      fieldKey: "state",
      recordId: "record-1",
      viewSchemaId: "test.view",
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({
      fieldKey: "summary",
      recordId: "record-2",
      viewSchemaId: "test.view",
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-1",
          viewSchemaId: "test.view",
        },
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-2",
          viewSchemaId: "test.view",
        },
      ],
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "state",
        },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-1",
          viewSchemaId: "test.view",
        },
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-2",
          viewSchemaId: "test.view",
        },
      ],
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: { viewSchemaId: "test.view", recordId: "", fieldKey: "state" },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });

  it("FE-U-P3-03 Resolve grid editability from explicit editor adapters and contract writeability only.", () => {
    const writableColumn = {
      contractWritable: true,
      fieldKey: "summary",
      label: "Summary",
      renderCell: ({ row }: { readonly row: HarnessRow }) => row.label,
      renderEditCell: () => null,
    };
    const readOnlyColumn = {
      contractWritable: false,
      fieldKey: "state",
      label: "State",
      renderCell: ({ row }: { readonly row: HarnessRow }) => row.state,
      renderEditCell: () => null,
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
    const renderCell = ({ row }: { readonly row: HarnessRow }) => row.label;
    const renderEditCell = () => null;
    const column = {
      contractWritable: true,
      fieldKey: "summary",
      label: "Summary",
      renderCell,
      renderEditCell,
    };

    expect(column.renderCell({ row: { label: "Alpha", state: "open" } })).toBe(
      "Alpha",
    );
    expect(isGridColumnEditable(column)).toBe(true);
  });

  it("translates valid presentation coordinates into stable record_id and field_key anchors", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        viewSchemaId: "test.view",
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual({
      recordId: "record-2",
      fieldKey: "state",
      viewSchemaId: "test.view",
    });
  });

  it("clears anchors for invalid row, field, group, and recordless targets", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", "closed"),
      draftRow("draft-1", "rough"),
    ];
    const presentationRows = buildGridPresentationRows({
      grouping: stateGrouping(),
      rows,
    });

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        viewSchemaId: "test.view",
        selection: { rowIndex: -1, fieldKey: "state" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        viewSchemaId: "test.view",
        selection: { rowIndex: 1, fieldKey: "__cartulary_actions__" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        viewSchemaId: "test.view",
        selection: { rowIndex: 0, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      presentationRows.some(
        (row) => row.kind === "data" && row.gridRow.recordId === "draft-1",
      ),
    ).toBe(false);
  });

  it("updates anchors for keyboard navigation and clears on presentation-only targets", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      navigateGridCellAnchor({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual({
      recordId: "record-1",
      fieldKey: "state",
      viewSchemaId: "test.view",
    });

    expect(
      navigateGridCellAnchor({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "state",
        },
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({
      recordId: "record-2",
      fieldKey: "state",
      viewSchemaId: "test.view",
    });

    const groupedRows = buildGridPresentationRows({
      grouping: stateGrouping(false),
      rows,
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
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
        viewSchemaId: "test.view",
        selection: vendorSelection,
      }),
    ).toEqual({
      recordId: "record-2",
      fieldKey: "state",
      viewSchemaId: "test.view",
    });
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
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-1",
          viewSchemaId: "test.view",
        },
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-2",
          viewSchemaId: "test.view",
        },
      ],
    });
  });

  it("maps filtered overflow to explicit create-row anchors", () => {
    const rows = [gridRow("record-2", "reviewed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridPasteTargets({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-2",
          fieldKey: "state",
        },
        pastedColumnCount: 2,
        pastedRowCount: 3,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        {
          baseRowVersion: 1,
          kind: "record",
          recordId: "record-2",
          viewSchemaId: "test.view",
        },
        { createIndex: 0, kind: "create", viewSchemaId: "test.view" },
        { createIndex: 1, kind: "create", viewSchemaId: "test.view" },
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
      grouping: stateGrouping(false),
      rows,
    });

    expect(
      resolveGridPasteTargets({
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-1",
          fieldKey: "summary",
        },
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: {
          viewSchemaId: "test.view",
          recordId: "record-2",
          fieldKey: "summary",
        },
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
        current: {
          viewSchemaId: "test.view",
          recordId: "",
          fieldKey: vendorSelection.fieldKey,
        },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });
});
