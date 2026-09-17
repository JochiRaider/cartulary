import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  createColumnSizingPort,
  elementCssScale,
  normalizeMeasuredColumnWidth,
} from "./columnSizing";
import { SemanticDataGrid } from "./SemanticDataGrid";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});
describe("Grid column sizing", () => {
  it("normalizes finite measurements once with owner bounds and CSS zoom", () => {
    const column = { fieldKey: "summary", minWidth: 40, maxWidth: 4096 };
    expect(normalizeMeasuredColumnWidth(39.8, column)).toBe(40);
    expect(normalizeMeasuredColumnWidth(400.4, column)).toBe(400);
    expect(normalizeMeasuredColumnWidth(400.4, column, true)).toBe(401);
    expect(normalizeMeasuredColumnWidth(9000, column, true)).toBe(4096);
    for (const value of [Number.NaN, Number.POSITIVE_INFINITY])
      expect(normalizeMeasuredColumnWidth(value, column)).toBeNull();
    expect(normalizeMeasuredColumnWidth(-200, column)).toBe(40);
    expect(normalizeMeasuredColumnWidth(0, column, true)).toBeNull();
    expect(
      normalizeMeasuredColumnWidth(40, {
        fieldKey: "extension",
        minWidth: 120,
      }),
    ).toBe(120);
    const element = document.createElement("div");
    element.style.width = "100px";
    element.style.boxSizing = "border-box";
    document.body.append(element);
    vi.spyOn(element, "getBoundingClientRect").mockReturnValue(
      new DOMRect(0, 0, 200, 40),
    );
    expect(elementCssScale(element)).toBe(2);
    element.remove();
  });
  it("rejects unavailable columns and cancels pending measurement on abort or invalidation", async () => {
    const root = document.createElement("div");
    const header = document.createElement("div");
    header.setAttribute("role", "columnheader");
    header.dataset.gridFieldKey = "summary";
    root.append(header);
    document.body.append(root);
    vi.spyOn(root, "getBoundingClientRect").mockReturnValue(
      new DOMRect(0, 0, 800, 400),
    );
    vi.spyOn(header, "getBoundingClientRect").mockReturnValue(
      new DOMRect(0, 0, 200, 30),
    );
    const sizing = createColumnSizingPort(() => ({
      root,
      columns: [{ fieldKey: "summary" }],
      presentationKey: "one",
    }));
    expect(sizing.port.unavailableReason("hidden")).toContain("Show");
    const fontDescriptor = Object.getOwnPropertyDescriptor(document, "fonts");
    Object.defineProperty(document, "fonts", {
      configurable: true,
      value: { status: "loading" },
    });
    expect(sizing.port.unavailableReason("summary")).toContain("Fonts");
    if (fontDescriptor)
      Object.defineProperty(document, "fonts", fontDescriptor);
    else Reflect.deleteProperty(document, "fonts");
    const pane = document.createElement("div");
    pane.style.overflowX = "hidden";
    document.body.append(pane);
    pane.append(root);
    vi.spyOn(pane, "getBoundingClientRect").mockReturnValue(
      new DOMRect(300, 0, 400, 400),
    );
    expect(sizing.port.unavailableReason("summary")).toContain("Scroll");
    document.body.append(root);
    pane.remove();
    const abort = new AbortController();
    const first = sizing.port.measureVisibleContent("summary", {
      signal: abort.signal,
    });
    abort.abort();
    expect(await first).toEqual({ kind: "cancelled" });
    const second = sizing.port.measureVisibleContent("summary", {
      signal: new AbortController().signal,
    });
    sizing.invalidate();
    expect(await second).toEqual({ kind: "cancelled" });
    root.remove();
    sizing.dispose();
  });
  it("routes production header keyboard and double-click intents without sorting", () => {
    const onIntent = vi.fn();
    const onSort = vi.fn();
    render(
      <SemanticDataGrid
        surface={{ kind: "view_schema", viewSchemaId: "test.sizing" }}
        columns={[
          {
            fieldKey: "summary",
            label: "Summary",
            width: 100,
            minWidth: 40,
            maxWidth: 4096,
            sortableFieldKey: "summary",
            renderCell: ({ row }: { row: { value: string } }) => row.value,
          },
        ]}
        dataRows={[]}
        onColumnSizingIntent={onIntent}
        onSortChange={onSort}
      />,
    );
    const header = screen.getByRole("columnheader", { name: "Summary" });
    header.style.width = "100px";
    header.style.boxSizing = "border-box";
    vi.spyOn(header, "getBoundingClientRect").mockReturnValue(
      new DOMRect(0, 0, 200, 30),
    );
    fireEvent.keyDown(header, { key: "ArrowRight", ctrlKey: true });
    expect(onIntent).toHaveBeenLastCalledWith({
      kind: "set_width",
      fieldKey: "summary",
      widthPx: 110,
    });
    const handle = header.querySelector('[data-grid-column-resize="summary"]');
    expect(handle).not.toBeNull();
    if (handle) fireEvent.doubleClick(handle);
    expect(onIntent).toHaveBeenLastCalledWith({
      kind: "fit_visible",
      fieldKey: "summary",
    });
    expect(onSort).not.toHaveBeenCalled();
  });
});
