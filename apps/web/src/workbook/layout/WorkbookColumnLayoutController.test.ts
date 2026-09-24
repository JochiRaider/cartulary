import type {
  GridColumnMeasurement,
  GridColumnSizingPort,
} from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it, vi } from "vitest";
import { parseWorkbookColumnWidth } from "../models/workbookColumnSizing";
import {
  buildSavedViewLayoutJson,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "../models/workbookQuery";
import { WorkbookColumnLayoutController } from "./WorkbookColumnLayoutController";
import {
  applyWorkbookLayoutToColumns,
  workbookFrozenDataColumnPrefix,
} from "./workbookColumnLayout";

const id = "cartulary.view.timeline.v2";
const field = "timeline.activity_synopsis_text";
const other = "timeline.raw_activity_text";
const contract = requireViewContract(id);
function harness() {
  const owner = new WorkbookColumnLayoutController();
  const requests: {
    resolve: (result: GridColumnMeasurement) => void;
    signal: AbortSignal;
  }[] = [];
  const port: GridColumnSizingPort = {
    unavailableReason: () => null,
    subscribe: () => () => undefined,
    measureVisibleContent: (_field, { signal }) =>
      new Promise((resolve) => requests.push({ resolve, signal })),
  };
  const unbind = owner.bind(id, { defaultWidth: () => 300, port });
  owner.activate("timeline:base");
  const fit = () => owner.fitVisible(id, field);
  return { owner, requests, unbind, fit, port };
}
describe("Workbook column sizing", () => {
  it("preserves sparse defaults and unrelated layout while enforcing portable bounds", () => {
    const { owner } = harness();
    owner.move(id, field, "earlier");
    owner.hide(id, other, true);
    owner.onIntent(id, { kind: "set_width", fieldKey: other, widthPx: 4096 });
    owner.onIntent(id, { kind: "set_width", fieldKey: field, widthPx: 300 });
    expect(owner.read(id, field).overridden).toBe(true);
    const before = owner.currentLayoutStateForSurface(id);
    for (const widthPx of [
      39,
      4097,
      40.5,
      Number.NaN,
      Number.POSITIVE_INFINITY,
    ])
      owner.onIntent(id, { kind: "set_width", fieldKey: field, widthPx });
    expect(owner.currentLayoutStateForSurface(id)).toEqual(before);
    owner.restoreDefault(id, field);
    expect(owner.currentLayoutStateForSurface(id)).toEqual({
      ...before,
      columnWidths: { [other]: 4096 },
    });
    owner.onIntent(id, { kind: "set_width", fieldKey: field, widthPx: 40 });
    const columns = applyWorkbookLayoutToColumns(
      contract,
      [
        {
          fieldKey: field,
          label: "Summary",
          renderCell: () => null,
          width: 300,
        },
      ],
      owner.currentLayoutStateForSurface(id),
    );
    expect(columns[0]).toMatchObject({
      width: 40,
      minWidth: 40,
      maxWidth: 4096,
    });
    expect(
      buildSavedViewLayoutJson(contract, owner.currentLayoutStateForSurface(id))
        .column_widths,
    ).toContainEqual({ field_key: field, width_px: 40 });
  });
  it("validates numeric input without truncating saved fractional widths", () => {
    for (const text of ["", " ", "39", "4097", "40.5", "abc", "Infinity"])
      expect(parseWorkbookColumnWidth(text)).toBeNull();
    expect(parseWorkbookColumnWidth("40")).toBe(40);
    expect(parseWorkbookColumnWidth("4096")).toBe(4096);
    const decoded = workbookLayoutStateFromSavedViewLayoutJson(contract, {
      column_widths: [{ field_key: field, width_px: 100.7 }],
    });
    expect(decoded).toBeNull();
  });
  it("fits once and reports header-only and capped results", async () => {
    const h = harness();
    const completed = h.fit();
    h.requests[0]?.resolve({
      kind: "measured",
      widthPx: 4096,
      capped: true,
      cellCount: 0,
    });
    expect(await completed).toEqual({ kind: "completed", widthPx: 4096 });
    expect(h.owner.currentLayoutStateForSurface(id).columnWidths[field]).toBe(
      4096,
    );
    expect(h.owner.getSnapshot().notice).toContain("using the header");
    expect(h.owner.getSnapshot().notice).toContain("Maximum width reached");
    expect(h.owner.getSnapshot().pendingField).toBeNull();
    h.owner.refresh();
    expect(h.requests).toHaveLength(1);
  });
  it("rejects obsolete results after newer commands configuration changes and departure", async () => {
    for (const change of [
      (h: ReturnType<typeof harness>) => h.owner.applyWidth(id, field, 480),
      (h: ReturnType<typeof harness>) =>
        h.owner.onIntent(id, {
          kind: "set_width",
          fieldKey: field,
          widthPx: 480,
        }),
      (h: ReturnType<typeof harness>) => h.owner.restoreDefault(id, field),
      (h: ReturnType<typeof harness>) => h.owner.hide(id, field, true),
      (h: ReturnType<typeof harness>) => h.owner.move(id, field, "earlier"),
      (h: ReturnType<typeof harness>) => h.owner.reset(id),
      (h: ReturnType<typeof harness>) =>
        h.owner.applyLayoutStateForSurface(
          id,
          h.owner.currentLayoutStateForSurface(id),
        ),
      (h: ReturnType<typeof harness>) =>
        h.owner.activate("timeline:replacement-saved-view"),
      (h: ReturnType<typeof harness>) => h.owner.cancel(),
      (h: ReturnType<typeof harness>) => h.unbind(),
      (h: ReturnType<typeof harness>) => h.owner.dispose(),
    ]) {
      const h = harness();
      const completed = h.fit();
      change(h);
      const before = h.owner.currentLayoutStateForSurface(id);
      expect(h.requests[0]?.signal.aborted).toBe(true);
      h.requests[0]?.resolve({
        kind: "measured",
        widthPx: 800,
        capped: false,
        cellCount: 4,
      });
      expect(await completed).toEqual({ kind: "cancelled" });
      expect(h.owner.currentLayoutStateForSurface(id)).toEqual(before);
    }
  });
  it("repeated fitting admits only the newest result and unavailable content changes nothing", async () => {
    const h = harness();
    const first = h.fit();
    const second = h.fit();
    h.requests[0]?.resolve({
      kind: "measured",
      widthPx: 800,
      capped: false,
      cellCount: 2,
    });
    h.requests[1]?.resolve({
      kind: "unavailable",
      reason: "Fonts are loading.",
    });
    expect(await first).toEqual({ kind: "cancelled" });
    expect(await second).toEqual({
      kind: "unavailable",
      reason: "Fonts are loading.",
    });
    expect(h.owner.currentLayoutStateForSurface(id).columnWidths).toEqual({});
    expect(h.owner.getSnapshot().notice).toBe("Fonts are loading.");
    h.owner.hide(id, field, true);
    const measure = vi.spyOn(h.port, "measureVisibleContent");
    h.fit();
    expect(measure).not.toHaveBeenCalled();
  });
  it("returns applied and restored widths with one owner notice", () => {
    const h = harness();
    expect(h.owner.applyWidth(id, field, 300)).toEqual({
      kind: "completed",
      widthPx: 300,
    });
    expect(h.owner.read(id, field).overridden).toBe(true);
    expect(h.owner.getSnapshot().notice).toContain("width set to 300 px");
    expect(h.owner.restoreDefault(id, field)).toEqual({
      kind: "completed",
      widthPx: 300,
    });
    expect(h.owner.read(id, field).overridden).toBe(false);
    expect(h.owner.getSnapshot().notice).toContain("default width restored");
    h.owner.onIntent(id, { kind: "set_width", fieldKey: field, widthPx: 440 });
    expect(h.owner.getSnapshot().notice).toBeNull();
    expect(h.owner.applyWidth(id, field, 39)).toEqual({
      kind: "unavailable",
      reason: "Enter a whole number from 40 to 4096.",
    });
    expect(h.owner.read(id, field).width).toBe(440);
  });
});

it("Workbook frozen layout preserves semantic boundaries through hidden reorder replacement and reset", () => {
  const { owner } = harness();
  const read = () => owner.currentLayoutStateForSurface(id);
  const [first, second] = read().columnOrder;
  if (!first || !second) throw new Error("Missing semantic column fixture");
  owner.freeze(id, second);
  expect(read().frozenThroughFieldKey).toBe(second);
  expect(workbookFrozenDataColumnPrefix(read())).toEqual(
    read()
      .columnOrder.slice(0, 2)
      .filter((key) => !read().hiddenFieldKeys.includes(key)),
  );
  const saved = read();
  owner.hide(id, second, true);
  expect(read().frozenThroughFieldKey).toBe(second);
  expect(workbookFrozenDataColumnPrefix(read())).not.toContain(second);
  owner.move(id, second, "earlier");
  expect(read().columnOrder[0]).toBe(second);
  expect(workbookFrozenDataColumnPrefix(read())).toEqual([]);
  owner.hide(id, second, false);
  expect(workbookFrozenDataColumnPrefix(read())).toEqual([second]);
  owner.onIntent(id, { kind: "set_width", fieldKey: second, widthPx: 4096 });
  expect(read().frozenThroughFieldKey).toBe(second);
  owner.freeze(id, "unknown.field");
  expect(read().frozenThroughFieldKey).toBe(second);
  owner.freeze(id, null);
  expect(workbookFrozenDataColumnPrefix(read())).toEqual([]);
  expect(read().columnWidths[second]).toBe(4096);
  owner.applyLayoutStateForSurface(id, saved);
  expect(read()).toEqual(saved);
  for (const key of read().columnOrder) owner.hide(id, key, true);
  expect(workbookFrozenDataColumnPrefix(read())).toEqual([]);
  expect(read().frozenThroughFieldKey).toBe(second);
  owner.reset(id);
  expect(read().frozenThroughFieldKey).toBeNull();
  expect(read().columnWidths).toEqual({});
  expect(read().columnOrder[0]).toBe(first);
});
