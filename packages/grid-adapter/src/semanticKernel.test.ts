import { describe, expect, it, vi } from "vitest";

import type {
  GridCellAnchor,
  GridColumn,
  GridDataRow,
  GridDataState,
  SemanticDataGridProps,
} from "./core";
import { decideSemanticActiveCellTransition } from "./semanticActiveCellPolicy";
import { resolveSemanticGridCapabilities } from "./semanticCapabilities";
import {
  mergeSemanticFillIntents,
  planSemanticCopy,
  planSemanticFillFromRange,
  planSemanticPaste,
} from "./semanticClipboardPolicy";
import {
  gridDataStatePresentsAuthorizedRows,
  gridDataStatePresentsDraft,
  resolveGridDataStatePresentation,
  resolveGridInteractionModePresentation,
} from "./semanticDataState";
import {
  decideSemanticGridKey,
  normalizeGridKey,
} from "./semanticKeyboardPolicy";
import { buildSemanticPresentationModel } from "./semanticPresentation";
import {
  nextSemanticSort,
  resolveSemanticBulkSelection,
  toggleAllSemanticRecords,
  toggleSemanticRecordRange,
} from "./semanticSelectionPolicy";

type Row = { readonly label: string; readonly state: string };

const surface = { kind: "view_schema", viewSchemaId: "test.view" } as const;
const editor = {
  clearDraftValue: "",
  commit: async () => ({ kind: "accepted" as const }),
  initialDraftValue: (row: Row) => row.label,
  renderEditor: () => null,
};
const columns: readonly GridColumn<Row>[] = [
  {
    contractWritable: true,
    editor,
    fieldKey: "label",
    getClipboardValue: (row) => row.label,
    label: "Label",
    renderCell: ({ row }) => row.label,
  },
  {
    fieldKey: "state",
    getClipboardValue: (row) => row.state,
    label: "State",
    renderCell: ({ row }) => row.state,
  },
];
const rows: readonly GridDataRow<Row>[] = [
  row("record-1", 1, "Alpha", "open"),
  row("record-2", 2, "Beta", "closed"),
  row("record-3", 3, "Gamma", "open"),
];
const model = buildSemanticPresentationModel({
  allowCreateRows: true,
  columns,
  dataRows: rows,
  fieldKeys: ["label", "state"],
  surface,
});
const start = anchor("record-1", "label");

describe("shared semantic grid kernel", () => {
  it("admits only surface-valid capabilities and derives one effective mode", () => {
    const editable = resolveSemanticGridCapabilities({
      columns,
      dataRows: rows,
      surface,
    } as SemanticDataGridProps<Row>);
    expect(editable).toMatchObject({
      editable: true,
      interactionMode: { kind: "editable" },
    });

    const extension = {
      columns,
      dataRows: [],
      surface: {
        extensionProfileId: "network_flow_activity",
        gridSchemaId: "network_flow.rows.v1",
        kind: "extension_grid",
        workspaceKey: "incident-1",
      },
    } as const;
    expect(resolveSemanticGridCapabilities(extension)).toMatchObject({
      editable: false,
      interactionMode: { kind: "read_only", label: "Read only" },
    });
    expect(() =>
      resolveSemanticGridCapabilities({
        ...extension,
        interactionMode: { kind: "editable" },
      } as unknown as SemanticDataGridProps<Row>),
    ).toThrow(/cannot enable Core mutation/i);
  });

  it("resolves every data-state branch through one precedence table", () => {
    const states: readonly [
      GridDataState,
      string | null,
      boolean,
      boolean,
      boolean,
    ][] = [
      [{ kind: "ready" }, null, false, true, true],
      [
        { generationKey: 1, kind: "initial_loading", surfaceLabel: "Rows" },
        "Loading Rows…",
        true,
        false,
        false,
      ],
      [
        { kind: "refreshing", surfaceLabel: "Rows" },
        "Refreshing Rows…",
        false,
        true,
        true,
      ],
      [{ kind: "empty", message: "No rows." }, "No rows.", false, false, true],
      [
        {
          action: { label: "Clear", onInvoke: vi.fn() },
          kind: "filtered_empty",
        },
        "No rows match the current filters.",
        false,
        false,
        true,
      ],
      [
        { kind: "stale_error", message: "Failed." },
        "Failed. Previously loaded rows may be stale.",
        false,
        true,
        true,
      ],
      [
        { kind: "unavailable", message: "Offline." },
        "Offline.",
        true,
        false,
        false,
      ],
      [
        { kind: "permission_denied" },
        "You no longer have access to this workbook.",
        true,
        false,
        false,
      ],
    ];
    for (const [
      state,
      message,
      blocking,
      presentsRows,
      presentsDraft,
    ] of states) {
      const resolved = resolveGridDataStatePresentation(state);
      expect(resolved.message).toBe(message);
      expect(resolved.blocking).toBe(blocking);
      expect(resolved.semanticStateId).toBe(state.kind);
      expect(gridDataStatePresentsAuthorizedRows(state)).toBe(presentsRows);
      expect(gridDataStatePresentsDraft(state)).toBe(presentsDraft);
    }
    expect(
      resolveGridDataStatePresentation(
        { generationKey: 1, kind: "initial_loading", surfaceLabel: "Rows" },
        "delayed",
      ).message,
    ).toBe("Still loading this surface");
    expect(
      resolveGridInteractionModePresentation({
        kind: "read_only",
        label: "Closed, read-only",
      }),
    ).toMatchObject({
      live: "polite",
      message: "Closed, read-only",
      visible: true,
    });
  });

  it("returns closed keyboard decisions for edit, rejection, exit, navigation, range, fill, and ignore", () => {
    const firstRow = rows[0];
    if (firstRow === undefined)
      throw new Error("Kernel fixture requires one row.");
    const base = {
      anchor: start,
      column: columns[0],
      editable: true,
      model,
      pageSize: 2,
      range: null,
      readOnlyLabel: "Workbook locked",
      row: firstRow,
    };
    expect(decideSemanticGridKey({ ...base, input: key("x") })).toMatchObject({
      kind: "begin_edit",
      seed: { hasValue: true, value: "x" },
    });
    expect(
      decideSemanticGridKey({ ...base, editable: false, input: key("Enter") }),
    ).toEqual({ announcement: "Workbook locked", kind: "reject" });
    expect(
      decideSemanticGridKey({ ...base, input: key("Tab", { shiftKey: true }) }),
    ).toEqual({
      backwards: true,
      kind: "exit_grid",
    });
    expect(
      decideSemanticGridKey({ ...base, input: key("ArrowRight") }),
    ).toMatchObject({
      kind: "navigate",
      range: null,
      target: { fieldKey: "state" },
    });
    expect(
      decideSemanticGridKey({
        ...base,
        input: key("ArrowDown", { shiftKey: true }),
      }),
    ).toMatchObject({
      kind: "navigate",
      range: { start, end: { rowIdentity: { recordId: "record-2" } } },
    });
    expect(
      decideSemanticGridKey({ ...base, input: key("d", { ctrlKey: true }) }),
    ).toEqual({
      kind: "fill",
      range: null,
    });
    expect(decideSemanticGridKey({ ...base, input: key("Escape") })).toEqual({
      kind: "ignore",
    });
    expect(decideSemanticGridKey({ ...base, input: key("ArrowLeft") })).toEqual(
      { kind: "ignore" },
    );
  });

  it("plans copy, paste, and fill without binding coordinates or effects", () => {
    const range = { start, end: anchor("record-2", "label") };
    expect(
      planSemanticCopy({
        anchor: start,
        columns,
        dataRows: rows,
        model,
        range,
      }),
    ).toMatchObject({
      text: "Alpha\nBeta",
      intent: { range },
    });
    expect(
      planSemanticPaste({
        input: {
          format: "tsv",
          kind: "table",
          rawText: "A\nB",
          values: [["A"], ["B"]],
        },
        model,
        target: {
          ...start,
          mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
        },
      }),
    ).toMatchObject({
      range,
      targetResolution: { columns: ["label"] },
    });
    const fill = planSemanticFillFromRange({
      columns,
      dataRows: rows,
      model,
      range,
      surface,
    });
    expect(fill).toMatchObject({
      targets: [{ rowIdentity: { recordId: "record-2" } }],
    });
    expect(fill === null ? null : mergeSemanticFillIntents(null, fill)).toBe(
      fill,
    );
    expect(
      planSemanticFillFromRange({
        columns,
        dataRows: rows,
        model,
        range: { start, end: anchor("record-2", "state") },
        surface,
      }),
    ).toBeNull();
  });

  it("plans sorting and bulk selection with reference-preserving rejected transitions", () => {
    expect(nextSemanticSort([], "label", false)).toEqual([
      { direction: "asc", fieldKey: "label" },
    ]);
    expect(
      nextSemanticSort(
        [{ direction: "asc", fieldKey: "label" }],
        "label",
        false,
      ),
    ).toEqual([{ direction: "desc", fieldKey: "label" }]);
    expect(
      nextSemanticSort(
        [{ direction: "desc", fieldKey: "label" }],
        "label",
        false,
      ),
    ).toEqual([]);

    const selected = new Set(["record-1"]);
    const selection = resolveSemanticBulkSelection(
      rows,
      selected,
      (candidate) =>
        candidate.rowIdentity.kind === "core_record" &&
        candidate.rowIdentity.recordId !== "record-3",
    );
    expect(selection).toMatchObject({
      allSelected: false,
      partiallySelected: true,
      selectableIds: ["record-1", "record-2"],
    });
    expect(toggleAllSemanticRecords(selection)).toEqual(
      new Set(["record-1", "record-2"]),
    );
    expect(
      toggleSemanticRecordRange({
        anchorRecordId: "record-1",
        recordId: "record-2",
        selectableRows: selection.selectableRows,
        selectedRecordIds: selected,
        shiftKey: true,
      }),
    ).toEqual(new Set(["record-1", "record-2"]));
    expect(
      toggleSemanticRecordRange({
        anchorRecordId: null,
        recordId: "missing",
        selectableRows: selection.selectableRows,
        selectedRecordIds: selected,
        shiftKey: false,
      }),
    ).toBe(selected);
    expect(decideSemanticActiveCellTransition(start, start)).toEqual({
      kind: "no_change",
    });
    expect(decideSemanticActiveCellTransition(null, start)).toEqual({
      anchor: start,
      kind: "publish",
    });
  });
});

function row(
  recordId: string,
  baseRowVersion: number,
  label: string,
  state: string,
): GridDataRow<Row> {
  return {
    data: { label, state },
    kind: "data",
    mutationIdentity: { baseRowVersion, kind: "core_row_version" },
    rowIdentity: { kind: "core_record", recordId },
  };
}

function anchor(recordId: string, fieldKey: string): GridCellAnchor {
  return {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface,
  };
}

function key(
  value: string,
  modifiers: Partial<{
    readonly altKey: boolean;
    readonly ctrlKey: boolean;
    readonly metaKey: boolean;
    readonly shiftKey: boolean;
  }> = {},
) {
  return normalizeGridKey({
    altKey: modifiers.altKey ?? false,
    ctrlKey: modifiers.ctrlKey ?? false,
    key: value,
    metaKey: modifiers.metaKey ?? false,
    shiftKey: modifiers.shiftKey ?? false,
  });
}
