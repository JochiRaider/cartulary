import { describe, expect, it } from "vitest";
import type {
  GridCellAnchor,
  GridCellRange,
  GridColumn,
  GridDataRow,
} from "./core";
import { captureSemanticClear, isClearNavigationKey } from "./semanticClear";
import {
  decideSemanticGridKey,
  decideSpreadsheetNavigation,
  normalizeGridKey,
} from "./semanticKeyboardPolicy";
import {
  buildSemanticPresentationModel,
  retainGridCellRange,
} from "./semanticPresentation";

function required<T>(value: T | undefined): T {
  if (value === undefined) throw new Error("Missing semantic fixture");
  return value;
}
const surface = { kind: "view_schema", viewSchemaId: "range-test" } as const;
const cell = (recordId: string, fieldKey: string): GridCellAnchor => ({
  surface,
  rowIdentity: { kind: "core_record", recordId },
  fieldKey,
});
const columns: GridColumn<string>[] = ["a", "b", "c"].map((fieldKey) => ({
  fieldKey,
  label: fieldKey,
  contractWritable: fieldKey !== "c",
  renderCell: ({ row }) => row,
  editor:
    fieldKey === "c"
      ? undefined
      : {
          clearDraftValue: "",
          initialDraftValue: (row) => row,
          commit: async () => ({ kind: "accepted" }),
          renderEditor: () => null,
        },
}));
const dataRows: GridDataRow<string>[] = ["1", "2", "3"].map((recordId) => ({
  kind: "data",
  rowIdentity: { kind: "core_record", recordId },
  mutationIdentity: { kind: "core_row_version", recordId, baseRowVersion: 1 },
  data: recordId,
}));
const model = buildSemanticPresentationModel({
  surface,
  columns,
  dataRows,
  fieldKeys: columns.map((column) => column.fieldKey),
  allowCreateRows: true,
});
const rectangle = { start: cell("1", "a"), end: cell("2", "b") };
const key = (value: string, shiftKey = false) =>
  normalizeGridKey({
    key: value,
    shiftKey,
    altKey: false,
    ctrlKey: false,
    metaKey: false,
  });

describe("range keyboard entry", () => {
  it("cycles every member and boundary in both orders and directions without changing reversed geometry", () => {
    const shapes = [
      {
        range: rectangle,
        tab: [cell("1", "a"), cell("1", "b"), cell("2", "a"), cell("2", "b")],
        enter: [cell("1", "a"), cell("2", "a"), cell("1", "b"), cell("2", "b")],
      },
      {
        range: { start: cell("1", "a"), end: cell("1", "c") },
        tab: [cell("1", "a"), cell("1", "b"), cell("1", "c")],
        enter: [cell("1", "a"), cell("1", "b"), cell("1", "c")],
      },
      {
        range: { start: cell("1", "b"), end: cell("3", "b") },
        tab: [cell("1", "b"), cell("2", "b"), cell("3", "b")],
        enter: [cell("1", "b"), cell("2", "b"), cell("3", "b")],
      },
    ];
    for (const { range: original, tab, enter } of shapes) {
      for (const range of [
        original,
        { start: original.end, end: original.start },
      ]) {
        for (const [chord, ordered] of [
          ["Tab", tab],
          ["Enter", enter],
        ] as const) {
          for (const reverse of [false, true]) {
            ordered.forEach((active, index) => {
              const decision = decideSpreadsheetNavigation(
                model,
                active,
                key(chord, reverse),
                ["a"],
                range,
              );
              expect(decision).toMatchObject({
                kind: "navigate",
                range,
                target:
                  ordered[
                    (index + (reverse ? -1 : 1) + ordered.length) %
                      ordered.length
                  ],
              });
              if (decision.kind === "navigate")
                expect(decision.range).toBe(range);
            });
          }
        }
      }
    }
  });

  it("uses presented membership including read-only fields and respects hidden reordered and removed members", () => {
    const presentation = {
      ...model,
      fieldKeys: ["c", "a"],
      rowIdentities: [cell("3", "c").rowIdentity, cell("1", "a").rowIdentity],
    };
    const range = { start: cell("3", "c"), end: cell("1", "a") };
    expect(
      decideSpreadsheetNavigation(
        presentation,
        range.start,
        key("Tab"),
        [],
        range,
      ),
    ).toMatchObject({ target: cell("3", "a"), range });
    expect(
      decideSpreadsheetNavigation(
        presentation,
        range.start,
        key("Enter"),
        [],
        range,
      ),
    ).toMatchObject({ target: cell("1", "c"), range });
    expect(retainGridCellRange(model, presentation, rectangle)).toBeNull();
    const refreshed = {
      ...model,
      dataRows: dataRows.map((row) => ({ ...row, data: "changed" })),
    };
    expect(retainGridCellRange(model, refreshed, rectangle)).toBe(rectangle);
    expect(
      retainGridCellRange(
        model,
        {
          ...model,
          rowIdentities: [
            ...model.rowIdentities,
            { kind: "core_record", recordId: "4" },
          ],
        },
        rectangle,
      ),
    ).toBe(rectangle);
    expect(
      retainGridCellRange(
        model,
        { ...model, rowIdentities: [...model.rowIdentities].reverse() },
        rectangle,
      ),
    ).toBeNull();
  });

  it("preserves ordinary draft shell and region navigation unless a valid multi-cell range is adopted", () => {
    const last = cell("3", "c");
    for (const range of [
      null,
      { start: last, end: last },
      { start: cell("missing", "a"), end: last },
      rectangle,
    ]) {
      expect(
        decideSpreadsheetNavigation(model, last, key("Tab"), ["a"], range),
      ).toEqual({ kind: "focus_draft", fieldKey: "a" });
      expect(
        decideSpreadsheetNavigation(model, last, key("Tab"), [], range),
      ).toEqual({ kind: "exit_grid", backwards: false });
    }
    const base = {
      anchor: rectangle.end,
      column: columns[1],
      editable: true,
      model,
      pageSize: 10,
      range: rectangle,
      readOnlyLabel: "Read only",
      row: required(dataRows[1]),
    };
    expect(decideSemanticGridKey({ ...base, input: key("Tab") })).toEqual({
      kind: "exit_grid",
      backwards: false,
    });
    expect(
      decideSemanticGridKey({
        ...base,
        keyboardNavigation: "spreadsheet",
        input: key("Tab"),
      }),
    ).toMatchObject({ kind: "navigate", range: null, target: cell("2", "c") });
    expect(decideSemanticGridKey({ ...base, input: key("F2") })).toEqual({
      kind: "ignore",
    });
    expect(
      decideSemanticGridKey({
        ...base,
        keyboardNavigation: "spreadsheet",
        rangeKeyboardEntry: "cycle",
        input: key("Tab"),
      }),
    ).toMatchObject({
      kind: "navigate",
      range: rectangle,
      target: rectangle.start,
    });
  });

  it("retains geometry only for eligible typing and F2 and extends from the traversed active cell", () => {
    const base = {
      anchor: rectangle.start,
      column: columns[0],
      editable: true,
      model,
      pageSize: 10,
      range: rectangle as GridCellRange | null,
      rangeKeyboardEntry: "cycle" as const,
      keyboardNavigation: "spreadsheet" as const,
      readOnlyLabel: "Read only",
      row: required(dataRows[0]),
    };
    expect(decideSemanticGridKey({ ...base, input: key("x") })).toMatchObject({
      kind: "begin_edit",
      range: rectangle,
      seed: { hasValue: true, value: "x" },
    });
    expect(decideSemanticGridKey({ ...base, input: key("F2") })).toMatchObject({
      kind: "begin_edit",
      range: rectangle,
      seed: {
        hasValue: false,
        activation: { source: "f2", initialSelection: "end" },
      },
    });
    expect(
      decideSemanticGridKey({ ...base, range: null, input: key("F2") }),
    ).toMatchObject({ kind: "begin_edit", range: null });
    expect(
      decideSemanticGridKey({ ...base, editable: false, input: key("F2") }),
    ).toEqual({ kind: "reject", announcement: "Read only" });
    for (const chord of ["Delete", "Backspace"])
      expect(
        decideSemanticGridKey({ ...base, input: key(chord) }),
      ).toMatchObject({ kind: "begin_edit", range: null, seed: { value: "" } });
    expect(decideSemanticGridKey({ ...base, input: key("Escape") })).toEqual({
      kind: "collapse_range",
      target: rectangle.start,
    });
    expect(
      decideSemanticGridKey({ ...base, input: key("ArrowRight", true) }),
    ).toMatchObject({
      kind: "navigate",
      range: { start: rectangle.start, end: cell("1", "b") },
    });
    expect(
      decideSemanticGridKey({ ...base, input: key("ArrowDown") }),
    ).toMatchObject({ kind: "navigate", range: null });
  });
});

describe("selected-cell clear capture", () => {
  it("captures reversed visible membership including offscreen rows without filtering fields", () => {
    const range = { start: cell("3", "c"), end: cell("1", "a") };
    const delivery = {};
    const captured = captureSemanticClear(model, range.start, range, delivery);
    expect(captured?.targets).toHaveLength(9);
    expect(captured?.expandedRange).toEqual({
      fieldKeys: ["a", "b", "c"],
      rowIdentities: dataRows.map((row) => row.rowIdentity),
    });
    expect(captured?.targets[0]).toMatchObject({
      rowIdentity: { recordId: "1" },
      fieldKey: "a",
      mutationIdentity: { baseRowVersion: 1 },
    });
    expect(captured?.anchor).toBe(range.start);
    expect(captured?.delivery).toBe(delivery);
    const reordered = { ...model, fieldKeys: ["c", "a"] };
    expect(
      captureSemanticClear(reordered, range.start, range)?.expandedRange
        .fieldKeys,
    ).toEqual(["c", "a"]);
    expect(
      captureSemanticClear(
        {
          ...model,
          rowIdentities: dataRows.slice(1).map((row) => row.rowIdentity),
        },
        range.start,
        range,
      ),
    ).toBeNull();
    expect(
      captureSemanticClear({ ...model, dataRows: [] }, range.start, range),
    ).toBeNull();
    expect(captureSemanticClear(model, null, null)).toBeNull();
  });
  it("admits only unmodified Delete as the optional clear chord", () => {
    const event = {
      key: "Delete",
      altKey: false,
      ctrlKey: false,
      metaKey: false,
      shiftKey: false,
    };
    expect(isClearNavigationKey(event)).toBe(true);
    for (const modifier of ["altKey", "ctrlKey", "metaKey", "shiftKey"])
      expect(isClearNavigationKey({ ...event, [modifier]: true })).toBe(false);
    for (const key of ["Backspace", "Enter", "F2"])
      expect(isClearNavigationKey({ ...event, key })).toBe(false);
  });
});
