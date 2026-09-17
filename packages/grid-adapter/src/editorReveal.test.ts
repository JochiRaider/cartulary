// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  bindGridEditorReveal,
  editorRevealDelta,
  revealGridEditorTarget,
} from "./editorReveal";
import { visibleGridViewport } from "./viewportGeometry";

afterEach(() => {
  document.body.replaceChildren();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

function mountedEditor(scale = 1) {
  const root = document.createElement("div");
  const editor = document.createElement("span");
  const input = document.createElement("input");
  root.style.cssText =
    "width:388px;height:300px;box-sizing:border-box;overflow:auto";
  editor.style.cssText =
    "display:block;width:220px;height:32px;box-sizing:border-box";
  input.style.cssText =
    "width:220px;height:32px;box-sizing:border-box;outline:none";
  editor.append(input);
  root.append(editor);
  document.body.append(root);
  Object.defineProperties(root, {
    clientWidth: { value: 388 },
    offsetWidth: { value: 388 },
    scrollWidth: { value: 2000 },
    clientHeight: { value: 300 },
    offsetHeight: { value: 300 },
    scrollHeight: { value: 900 },
  });
  vi.spyOn(root, "getBoundingClientRect").mockImplementation(
    () => new DOMRect(scale, 0, 388 * scale, 300 * scale),
  );
  const cellRect = () =>
    new DOMRect(
      (176 - root.scrollLeft) * scale,
      (32 - root.scrollTop) * scale,
      220 * scale,
      32 * scale,
    );
  vi.spyOn(editor, "getBoundingClientRect").mockImplementation(cellRect);
  vi.spyOn(input, "getBoundingClientRect").mockImplementation(cellRect);
  const scroll = vi.fn((options?: ScrollToOptions | number) => {
    if (typeof options === "object") {
      root.scrollLeft = options.left ?? root.scrollLeft;
      root.scrollTop = options.top ?? root.scrollTop;
    }
  });
  root.scrollTo = scroll;
  return { root, editor, input, scroll };
}

describe("mounted editor reveal", () => {
  it("uses minimal translations and leaves an oversized spanning control stable", () => {
    expect(editorRevealDelta(176, 396, 1, 389)).toBe(7);
    expect(editorRevealDelta(-5, 215, 1, 389)).toBe(-6);
    expect(editorRevealDelta(1, 221, 1, 389)).toBe(0);
    expect(editorRevealDelta(-100, 500, 1, 389)).toBe(0);
    expect(editorRevealDelta(50, 650, 1, 389)).toBe(49);
    expect(editorRevealDelta(-650, -50, 1, 389)).toBe(-439);
    expect(editorRevealDelta(0, 20, 10, 10)).toBe(0);
  });

  it("reveals the actual control in local scroll units at CSS zoom without touching text or focus", () => {
    const { root, editor, input, scroll } = mountedEditor(2);
    input.value = "  unfinished timestamp  ";
    input.focus();
    input.setSelectionRange(3, 7);
    revealGridEditorTarget(root, editor, input);
    expect(root.scrollLeft).toBe(7);
    expect(input.getBoundingClientRect().right).toBe(778);
    expect(input.value).toBe("  unfinished timestamp  ");
    expect([input.selectionStart, input.selectionEnd]).toEqual([3, 7]);
    expect(document.activeElement).toBe(input);
    expect(window.scrollX).toBe(0);
    revealGridEditorTarget(root, editor, input);
    expect(scroll).toHaveBeenCalledTimes(1);
  });

  it("intersects ancestor client clipping and accounts for grid borders and scrollbars", () => {
    const { root, editor, input } = mountedEditor();
    const pane = document.createElement("div");
    pane.style.cssText =
      "overflow-x:hidden;overflow-y:hidden;width:350px;height:250px;border:2px solid;box-sizing:border-box";
    document.body.append(pane);
    pane.append(root);
    Object.defineProperties(pane, {
      offsetWidth: { value: 350 },
      clientWidth: { value: 331 },
      offsetHeight: { value: 250 },
      clientHeight: { value: 246 },
    });
    vi.spyOn(pane, "getBoundingClientRect").mockReturnValue(
      new DOMRect(10, 5, 350, 250),
    );
    expect(visibleGridViewport(root)).toEqual({
      left: 12,
      right: 343,
      top: 7,
      bottom: 253,
    });
    revealGridEditorTarget(root, editor, input);
    expect(root.scrollLeft).toBe(53);
  });

  it("avoids sticky headers and frozen columns and clamps unreachable scroll destinations", () => {
    const { root, editor, input } = mountedEditor();
    const header = document.createElement("div");
    header.setAttribute("role", "columnheader");
    header.setAttribute("aria-colindex", "1");
    header.style.cssText = "position:sticky;inset-inline-start:0px";
    root.prepend(header);
    vi.spyOn(header, "getBoundingClientRect").mockReturnValue(
      new DOMRect(1, 0, 48, 40),
    );
    root.scrollLeft = 170;
    root.scrollTop = 20;
    revealGridEditorTarget(root, editor, input);
    expect(root.scrollLeft).toBe(127);
    expect(root.scrollTop).toBe(0);
    expect(input.getBoundingClientRect().left).toBe(49);
    const draftRow = document.createElement("div");
    draftRow.dataset.cartularyGridDraftRow = "true";
    const draftCell = document.createElement("div");
    draftCell.setAttribute("role", "gridcell");
    draftCell.style.position = "sticky";
    draftRow.append(draftCell);
    root.append(draftRow);
    vi.spyOn(draftCell, "getBoundingClientRect").mockReturnValue(
      new DOMRect(1, 260, 388, 40),
    );
    vi.spyOn(input, "getBoundingClientRect").mockReturnValue(
      new DOMRect(50, 250, 220, 32),
    );
    revealGridEditorTarget(root, editor, input);
    expect(root.scrollTop).toBe(22);
  });

  it("keeps RTL scroll offsets within the negative horizontal range", () => {
    const { root, editor, input } = mountedEditor();
    root.style.direction = "rtl";
    vi.spyOn(input, "getBoundingClientRect").mockReturnValue(
      new DOMRect(-20, 32, 220, 32),
    );
    revealGridEditorTarget(root, editor, input);
    expect(root.scrollLeft).toBe(-21);
  });

  it("coalesces layout work and cancels detached superseded or unfocused editor lifetimes", () => {
    const { root, editor, input, scroll } = mountedEditor();
    const frames = new Map<number, FrameRequestCallback>();
    let sequence = 0;
    vi.spyOn(globalThis, "requestAnimationFrame").mockImplementation(
      (callback) => {
        frames.set(++sequence, callback);
        return sequence;
      },
    );
    vi.spyOn(globalThis, "cancelAnimationFrame").mockImplementation((id) => {
      frames.delete(id);
    });
    const disconnect = vi.fn();
    let resize!: ResizeObserverCallback;
    vi.stubGlobal(
      "ResizeObserver",
      class {
        constructor(callback: ResizeObserverCallback) {
          resize = callback;
        }
        observe() {}
        disconnect = disconnect;
      },
    );
    const flush = () => {
      const pending = [...frames.values()];
      frames.clear();
      for (const callback of pending) callback(0);
    };
    let current = true;
    const binding = bindGridEditorReveal(root, editor, () => current);
    input.focus();
    binding.refresh();
    binding.refresh();
    expect(frames.size).toBe(1);
    current = false;
    flush();
    expect(scroll).not.toHaveBeenCalled();
    current = true;
    resize([], {} as ResizeObserver);
    flush();
    expect(scroll).toHaveBeenCalledTimes(1);
    root.scrollLeft = 0;
    input.blur();
    binding.refresh();
    flush();
    expect(scroll).toHaveBeenCalledTimes(1);
    input.focus();
    editor.remove();
    flush();
    expect(scroll).toHaveBeenCalledTimes(1);
    binding.refresh();
    binding.dispose();
    expect(frames.size).toBe(0);
    expect(disconnect).toHaveBeenCalledOnce();
    binding.refresh();
    expect(frames.size).toBe(0);
  });
});
