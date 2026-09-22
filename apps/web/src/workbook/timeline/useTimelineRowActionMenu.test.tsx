import type {
  GridHandle,
  GridPresentationSnapshot,
} from "@cartulary/grid-adapter";
import { act, cleanup, renderHook } from "@testing-library/react";
import type { MouseEvent } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useTimelineRowActionMenu } from "./hooks/useTimelineRowActionMenu";
import { createDraftRow } from "./models/timelineRowModel";

afterEach(() => {
  cleanup();
  document.body.replaceChildren();
});

function harness() {
  const grid = document.createElement("div");
  grid.tabIndex = -1;
  const row = document.createElement("div");
  row.dataset.gridRecordId = "saved";
  const cell = document.createElement("div");
  cell.dataset.gridFieldKey = "summary";
  cell.tabIndex = -1;
  row.append(cell);
  grid.append(row);
  document.body.append(grid);
  const listeners = new Set<() => void>();
  let snapshot: GridPresentationSnapshot = {
    surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
    rowIdentities: [{ kind: "core_record", recordId: "saved" }],
    fieldKeys: ["summary", "source"],
    revision: 1,
  };
  const requestFocus = vi
    .fn<GridHandle["requestFocus"]>()
    .mockResolvedValue("focused");
  const gridHandleRef = {
    current: {
      requestFocus,
      getScrollElement: () => grid,
      getAnchorRect: () => new DOMRect(0, 0, 100, 40),
      presentation: {
        getSnapshot: () => snapshot,
        subscribe: (listener: () => void) => {
          listeners.add(listener);
          return () => listeners.delete(listener);
        },
      },
    } as unknown as GridHandle,
  };
  const initialProps = {
    readable: true,
    scopeKey: "scope",
    rows: [{ ...createDraftRow(0), recordId: "saved", rowVersion: 1 }],
  };
  const hook = renderHook(
    (props) => useTimelineRowActionMenu({ ...props, gridHandleRef }),
    { initialProps },
  );
  const open = () =>
    act(() => {
      hook.result.current.commands.handleTimelineGridContextMenu({
        target: cell,
        nativeEvent: new globalThis.MouseEvent("contextmenu"),
        defaultPrevented: false,
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        clientX: 20,
        clientY: 20,
      } as unknown as MouseEvent<HTMLDivElement>);
    });
  const focusMenu = () => {
    const menu = document.createElement("div");
    menu.tabIndex = -1;
    document.body.append(menu);
    const state = hook.result.current.menu;
    if (!state) throw new Error("menu missing");
    state.menuRef.current = menu;
    state.onFocusChange(true);
    menu.focus();
  };
  const publish = (patch: Partial<GridPresentationSnapshot>) =>
    act(() => {
      snapshot = { ...snapshot, ...patch, revision: snapshot.revision + 1 };
      for (const listener of listeners) listener();
    });
  return {
    ...hook,
    grid,
    cell,
    initialProps,
    requestFocus,
    open,
    focusMenu,
    publish,
  };
}

describe("Timeline row menu lifetime", () => {
  it("keeps compatible refresh and invalidates hidden anchors and authority without stealing outside focus", async () => {
    const h = harness();
    h.open();
    h.focusMenu();
    h.rerender({
      ...h.initialProps,
      rows: h.initialProps.rows.map((row) => ({ ...row, rowVersion: 2 })),
    });
    expect(h.result.current.menu?.row?.rowVersion).toBe(2);
    expect(h.requestFocus).not.toHaveBeenCalled();
    const queuedScroll = new Event("scroll");
    Object.defineProperty(queuedScroll, "target", { value: h.grid });
    act(() => h.result.current.menu?.onExternalScroll(queuedScroll));
    expect(h.result.current.menu).not.toBeNull();
    const input = document.createElement("input");
    h.cell.append(input);
    const editorScroll = new Event("scroll");
    Object.defineProperty(editorScroll, "target", { value: input });
    act(() => h.result.current.menu?.onExternalScroll(editorScroll));
    expect(h.result.current.menu).not.toBeNull();
    h.grid.scrollTop = 80;
    act(() => h.result.current.menu?.onExternalScroll(queuedScroll));
    expect(h.result.current.menu).toBeNull();
    h.open();
    h.focusMenu();
    h.requestFocus.mockClear();
    h.publish({ fieldKeys: ["source"] });
    expect(h.result.current.menu).toBeNull();
    expect(h.requestFocus).toHaveBeenCalledWith(
      { kind: "cell", anchor: expect.objectContaining({ fieldKey: "source" }) },
      expect.objectContaining({ preserveSelection: true }),
    );
    h.publish({ fieldKeys: ["summary"] });
    h.open();
    h.focusMenu();
    const outside = document.createElement("button");
    document.body.append(outside);
    outside.focus();
    h.requestFocus.mockClear();
    h.rerender({ ...h.initialProps, scopeKey: "new authority" });
    expect(h.result.current.menu).toBeNull();
    expect(h.requestFocus).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(outside);
    await act(async () => {});
  });

  it("cancels the entire semantic fallback chain for newer input focus and scope", async () => {
    for (const intent of ["pointer", "keyboard", "focus", "scope"] as const) {
      const h = harness();
      h.open();
      let unavailable = () => {};
      let signal: AbortSignal | undefined;
      h.requestFocus.mockImplementationOnce((_target, options) => {
        signal = options?.signal;
        return new Promise((resolve) => {
          unavailable = () => resolve("unavailable");
        });
      });
      act(() => {
        h.result.current.menu?.onRestoreFocus();
        h.result.current.menu?.onClose();
      });
      const other = document.createElement("button");
      h.grid.append(other);
      if (intent === "focus") other.focus();
      if (intent === "pointer")
        other.dispatchEvent(new Event("pointerdown", { bubbles: true }));
      if (intent === "keyboard")
        other.dispatchEvent(
          new KeyboardEvent("keydown", { key: "Tab", bubbles: true }),
        );
      if (intent === "scope")
        h.rerender({ ...h.initialProps, scopeKey: "replaced" });
      expect(signal?.aborted).toBe(true);
      await act(async () => unavailable());
      expect(h.requestFocus).toHaveBeenCalledTimes(1);
      h.unmount();
      h.grid.remove();
    }
  });

  it("uses current semantic row then root fallbacks when the invoking field or row disappears", async () => {
    const h = harness();
    h.open();
    const restore = h.result.current.menu?.onRestoreFocus;
    act(() => h.result.current.menu?.onClose());
    h.publish({ fieldKeys: ["source"] });
    await act(async () => restore?.());
    expect(h.requestFocus).toHaveBeenLastCalledWith(
      { kind: "cell", anchor: expect.objectContaining({ fieldKey: "source" }) },
      expect.objectContaining({ preserveSelection: true }),
    );
    // Source removal can render before the adapter publishes its next snapshot.
    h.rerender({ ...h.initialProps, rows: [] });
    await act(async () => restore?.());
    expect(h.requestFocus).toHaveBeenLastCalledWith(
      { kind: "root" },
      expect.anything(),
    );
    h.publish({ rowIdentities: [] });
    await act(async () => restore?.());
    expect(h.requestFocus).toHaveBeenLastCalledWith(
      { kind: "root" },
      expect.anything(),
    );
  });
});
