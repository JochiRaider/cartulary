import { describe, expect, it } from "vitest";

import {
  assertGridRows,
  buildGridPresentationRows,
  formatGridClipboardTSV,
  type GridCellAnchor,
  type GridDataRow,
  type GridDraftRow,
  type GridPresentationRow,
  isGridColumnEditable,
  navigateGridCellAnchor,
  parseGridClipboardTable,
  resolveGridCellAnchor,
  resolveGridCellRange,
  resolveGridPasteTargets,
} from "./core";
import {
  gridSemanticStateClassNames,
  mergeGridSemanticState,
  resolveGridSemanticState,
} from "./semanticState";

type HarnessRow = {
  readonly label: string;
  readonly state: string | null | undefined;
};

function gridRow(
  key: string,
  state: string | null | undefined,
): GridDataRow<HarnessRow> {
  return {
    kind: "data",
    mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
    rowIdentity: { kind: "core_record", recordId: key },
    data: {
      label: key,
      state,
    },
  };
}

const testSurface = { kind: "view_schema", viewSchemaId: "test.view" } as const;

function gridAnchor(recordId: string, fieldKey: string): GridCellAnchor {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface: testSurface,
  };
}

function pasteRecordTarget(recordId: string) {
  return {
    kind: "record" as const,
    mutationIdentity: {
      kind: "core_row_version" as const,
      baseRowVersion: 1,
    },
    rowIdentity: { kind: "core_record" as const, recordId },
    surface: testSurface,
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

    const presentationRows = buildGridPresentationRows<HarnessRow>({
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

describe("extension grid semantic identities", () => {
  const extensionSurface = {
    kind: "extension_grid",
    extensionProfileId: "network_flow_activity",
    workspaceKey: "incident-1",
    gridSchemaId: "network_flow.accepted_rows.v1",
  } as const;
  const extensionRows = [
    {
      data: { label: "Flow 1", state: "accepted" },
      kind: "data",
      rowIdentity: {
        kind: "extension_resource",
        extensionProfileId: "network_flow_activity",
        resourceKind: "accepted_flow_row",
        resourceId: "flow-row-1",
      },
    },
    {
      data: { label: "Flow 2", state: "accepted" },
      kind: "data",
      rowIdentity: {
        kind: "extension_resource",
        extensionProfileId: "network_flow_activity",
        resourceKind: "accepted_flow_row",
        resourceId: "flow-row-2",
      },
    },
  ] as const satisfies readonly GridDataRow<HarnessRow>[];
  const extensionColumns = [
    { fieldKey: "label", label: "Label", renderCell: () => null },
    { fieldKey: "state", label: "State", renderCell: () => null },
  ] as const;

  it("resolves extension anchors, ranges, grouping, and navigation without Core aliases", () => {
    const presentationRows = buildGridPresentationRows<HarnessRow>({
      grouping: stateGrouping(false),
      rows: extensionRows,
    });
    const first = resolveGridCellAnchor({
      columns: extensionColumns,
      presentationRows,
      selection: { fieldKey: "label", rowIndex: 1 },
      surface: extensionSurface,
    });
    expect(first).toEqual({
      fieldKey: "label",
      rowIdentity: extensionRows[0].rowIdentity,
      surface: extensionSurface,
    });
    if (first === null) throw new Error("Expected extension anchor");
    expect(
      navigateGridCellAnchor({
        columns: extensionColumns,
        current: first,
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual({
      fieldKey: "label",
      rowIdentity: extensionRows[1].rowIdentity,
      surface: extensionSurface,
    });
    expect(
      resolveGridCellRange({
        columns: extensionColumns,
        presentationRows,
        range: {
          end: {
            fieldKey: "state",
            rowIdentity: extensionRows[1].rowIdentity,
            surface: extensionSurface,
          },
          start: first,
        },
      }),
    ).toEqual({
      fieldKeys: ["label", "state"],
      rowIdentities: extensionRows.map((row) => row.rowIdentity),
    });
  });

  it("rejects duplicate extension identities and all mutation target compilation", () => {
    expect(() => assertGridRows([extensionRows[0], extensionRows[0]])).toThrow(
      /duplicate semantic row identity/i,
    );
    const presentationRows = buildGridPresentationRows<HarnessRow>({
      rows: extensionRows,
    });
    expect(
      resolveGridPasteTargets({
        columns: extensionColumns,
        current: {
          fieldKey: "label",
          rowIdentity: extensionRows[0].rowIdentity,
          surface: extensionSurface,
        },
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });
});

describe("semantic grid state precedence", () => {
  it("resolves the design priority while retaining declared co-displays", () => {
    const resolved = resolveGridSemanticState(
      {
        active: true,
        bulkSelected: true,
        conflicted: true,
        inspectorActive: true,
        invalid: { message: "Required" },
        pending: true,
        readOnlyOrDerived: true,
        saved: true,
        stale: true,
      },
      "Summary",
    );

    expect(resolved.primary).toBe("conflicted");
    expect(resolved.stateIds).toEqual([
      "conflicted",
      "active",
      "bulk-selected",
      "inspector-active",
      "read-only",
      "stale",
    ]);
    expect(resolved.markers).toEqual([
      {
        accessibleLabel: "Conflict on Summary",
        glyph: "!",
        kind: "conflicted",
      },
      {
        accessibleLabel: "Stale Summary; refresh required",
        glyph: "↻",
        kind: "stale",
      },
    ]);
    expect(gridSemanticStateClassNames("cell", resolved)).toContain(
      "cartulary-grid-cell-state-conflicted",
    );
    expect(gridSemanticStateClassNames("cell", resolved)).not.toContain("rdg-");
  });

  it("orders invalid, pending, active, selected, read-only, and saved states", () => {
    expect(
      resolveGridSemanticState(
        { invalid: { message: "Choose an allowed value" }, pending: true },
        "Status",
      ).primary,
    ).toBe("invalid");
    expect(
      resolveGridSemanticState({ active: true, pending: true }, "Status")
        .primary,
    ).toBe("pending");
    expect(
      resolveGridSemanticState({ active: true, bulkSelected: true }, "Status")
        .primary,
    ).toBe("active");
    expect(
      resolveGridSemanticState(
        { bulkSelected: true, readOnlyOrDerived: true },
        "Status",
      ).primary,
    ).toBe("bulk-selected");
    expect(
      resolveGridSemanticState({ readOnlyOrDerived: true }, "Status").primary,
    ).toBe("read-only");
    expect(resolveGridSemanticState({ saved: true }, "Status").primary).toBe(
      "saved",
    );
  });

  it("merges owner state with adapter-derived context without weakening state", () => {
    expect(
      mergeGridSemanticState(
        {
          conflicted: true,
          invalid: { message: "Owner validation" },
          saved: false,
        },
        {
          active: true,
          bulkSelected: true,
          readOnlyOrDerived: true,
          saved: true,
          stale: true,
        },
      ),
    ).toEqual({
      active: true,
      bulkSelected: true,
      conflicted: true,
      inspectorActive: false,
      invalid: { message: "Owner validation" },
      pending: false,
      readOnlyOrDerived: true,
      saved: false,
      stale: true,
    });
  });
});

describe("grid Cartulary anchors", () => {
  const columns = [
    { fieldKey: "summary", label: "Summary", renderCell: () => null },
    { fieldKey: "state", label: "State", renderCell: () => null },
  ] as const;

  it("Reject unsafe record identity and keep presentation rows from mutation-capable anchors.", () => {
    expect(() =>
      assertGridRows([gridRow("record-1", "open"), gridRow(" ", "open")]),
    ).toThrow(/invalid semantic identity/i);
    expect(() =>
      assertGridRows([
        gridRow("record-1", "open"),
        gridRow("record-1", "closed"),
      ]),
    ).toThrow(/duplicate semantic row identity/i);

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
    const draftIndex = -1;

    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface: testSurface,
        selection: { rowIndex: groupIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface: testSurface,
        selection: { rowIndex: draftIndex, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        columns,
        current: gridAnchor("record-1", "summary"),
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: gridAnchor("record-2", "summary"),
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toBeNull();
  });

  it("Translate vendor row and column coordinates to stable record_id and field_key anchors.", () => {
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
        surface: testSurface,
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual(gridAnchor("record-1", "state"));
    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "summary"),
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-1", "state"));
    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "summary"),
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-2", "summary"));
    expect(
      resolveGridPasteTargets({
        columns,
        current: gridAnchor("record-1", "summary"),
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        pasteRecordTarget("record-1"),
        pasteRecordTarget("record-2"),
      ],
    });
    expect(
      resolveGridCellRange({
        columns,
        presentationRows,
        range: {
          start: gridAnchor("record-2", "state"),
          end: gridAnchor("record-1", "summary"),
        },
      }),
    ).toEqual({
      fieldKeys: ["summary", "state"],
      rowIdentities: [
        { kind: "core_record", recordId: "record-1" },
        { kind: "core_record", recordId: "record-2" },
      ],
    });
    const clipboardText = formatGridClipboardTSV([
      ["plain", "has\ttab"],
      ["line\nbreak", 'has "quote"'],
    ]);
    expect(clipboardText).toBe(
      'plain\t"has\ttab"\n"line\nbreak"\t"has ""quote"""',
    );
    expect(parseGridClipboardTable(clipboardText)).toEqual([
      ["plain", "has\ttab"],
      ["line\nbreak", 'has "quote"'],
    ]);
    expect(
      resolveGridPasteTargets({
        columns,
        current: gridAnchor("record-1", "state"),
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        pasteRecordTarget("record-1"),
        pasteRecordTarget("record-2"),
      ],
    });
    expect(
      resolveGridPasteTargets({
        columns,
        current: gridAnchor("", "state"),
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });

  it("Resolve grid editability from explicit editor adapters and contract writeability only.", () => {
    const writableColumn = {
      contractWritable: true,
      fieldKey: "summary",
      label: "Summary",
      renderCell: ({ row }: { readonly row: HarnessRow }) => row.label,
      editor: {
        commit: async () => ({ kind: "accepted" as const }),
        initialDraftValue: () => "",
        renderEditor: () => null,
      },
    };
    const readOnlyColumn = {
      contractWritable: false,
      fieldKey: "state",
      label: "State",
      renderCell: ({ row }: { readonly row: HarnessRow }) => row.state,
      editor: {
        commit: async () => ({ kind: "accepted" as const }),
        initialDraftValue: () => "",
        renderEditor: () => null,
      },
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

  it("Resolve renderers and editors deterministically and clean adapter-owned resources.", () => {
    const renderCell = ({ row }: { readonly row: HarnessRow }) => row.label;
    const editor = {
      commit: async () => ({ kind: "accepted" as const }),
      initialDraftValue: () => "",
      renderEditor: () => null,
    };
    const column = {
      contractWritable: true,
      fieldKey: "summary",
      label: "Summary",
      renderCell,
      editor,
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
        surface: testSurface,
        selection: { rowIndex: 1, fieldKey: "state" },
      }),
    ).toEqual(gridAnchor("record-2", "state"));
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
        surface: testSurface,
        selection: { rowIndex: -1, fieldKey: "state" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface: testSurface,
        selection: { rowIndex: 1, fieldKey: "__cartulary_actions__" },
      }),
    ).toBeNull();
    expect(
      resolveGridCellAnchor({
        columns,
        presentationRows,
        surface: testSurface,
        selection: { rowIndex: 0, fieldKey: "summary" },
      }),
    ).toBeNull();
    expect(
      presentationRows.some(
        (row) =>
          row.kind === "data" &&
          row.gridRow.rowIdentity.kind === "core_record" &&
          row.gridRow.rowIdentity.recordId === "draft-1",
      ),
    ).toBe(false);
  });

  it("updates anchors for keyboard navigation and clears on presentation-only targets", () => {
    const rows = [gridRow("record-1", "open"), gridRow("record-2", "closed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "summary"),
        intent: { key: "ArrowRight" },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-1", "state"));

    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "state"),
        intent: { key: "Enter" },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-2", "state"));

    const groupedRows = buildGridPresentationRows({
      grouping: stateGrouping(false),
      rows,
    });
    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "summary"),
        intent: { key: "ArrowUp" },
        presentationRows: groupedRows,
      }),
    ).toBeNull();

    expect(
      navigateGridCellAnchor({
        columns,
        current: gridAnchor("record-1", "summary"),
        intent: { key: "ArrowDown" },
        presentationRows: groupedRows,
      }),
    ).toEqual(gridAnchor("record-2", "summary"));
  });

  it("supports semantic page, row-edge, and grid-edge navigation", () => {
    const rows = [
      gridRow("record-1", "open"),
      gridRow("record-2", "open"),
      gridRow("record-3", "closed"),
    ];
    const presentationRows = buildGridPresentationRows({ rows });
    const current = gridAnchor("record-2", "summary");

    expect(
      navigateGridCellAnchor({
        columns,
        current,
        intent: { key: "PageDown", pageSize: 8 },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-3", "summary"));
    expect(
      navigateGridCellAnchor({
        columns,
        current,
        intent: { key: "Home" },
        presentationRows,
      }),
    ).toEqual(current);
    expect(
      navigateGridCellAnchor({
        columns,
        current,
        intent: { key: "End", ctrlOrMetaKey: true },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-3", "state"));
    expect(
      navigateGridCellAnchor({
        columns,
        current,
        intent: { key: "Home", ctrlOrMetaKey: true },
        presentationRows,
      }),
    ).toEqual(gridAnchor("record-1", "summary"));
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
        surface: testSurface,
        selection: vendorSelection,
      }),
    ).toEqual(gridAnchor("record-2", "state"));
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
        current: gridAnchor("record-1", "summary"),
        pastedColumnCount: 2,
        pastedRowCount: 2,
        presentationRows,
      }),
    ).toEqual({
      columns: ["summary", "state"],
      rowTargets: [
        pasteRecordTarget("record-1"),
        pasteRecordTarget("record-2"),
      ],
    });
  });

  it("maps filtered overflow to explicit create-row anchors", () => {
    const rows = [gridRow("record-2", "reviewed")];
    const presentationRows = buildGridPresentationRows({ rows });

    expect(
      resolveGridPasteTargets({
        columns,
        current: gridAnchor("record-2", "state"),
        pastedColumnCount: 1,
        pastedRowCount: 3,
        presentationRows,
      }),
    ).toEqual({
      columns: ["state"],
      rowTargets: [
        pasteRecordTarget("record-2"),
        { createIndex: 0, kind: "create", surface: testSurface },
        { createIndex: 1, kind: "create", surface: testSurface },
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
        current: gridAnchor("record-1", "summary"),
        pastedColumnCount: 1,
        pastedRowCount: 2,
        presentationRows: groupedRows,
      }),
    ).toBeNull();
    expect(
      resolveGridPasteTargets({
        allowCreateRows: false,
        columns,
        current: gridAnchor("record-2", "summary"),
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
        current: gridAnchor("", vendorSelection.fieldKey),
        pastedColumnCount: 1,
        pastedRowCount: 1,
        presentationRows,
      }),
    ).toBeNull();
  });
});
