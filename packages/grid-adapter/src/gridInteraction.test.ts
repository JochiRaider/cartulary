import { afterEach, describe, expect, it, vi } from "vitest";
import type { GridCellAnchor, GridCellRange, GridCellTarget } from "./core";
import {
  createGridInteractionController,
  crossesGridDragThreshold,
  type GridInteractionSnapshot,
  gridEdgeScrollDelta,
} from "./gridInteractionController";
import {
  bindGridInteractionDom,
  type RegisteredGridCell,
} from "./gridInteractionDom";
import { extendSemanticCellRange } from "./semanticSelectionPolicy";

const surface = { kind: "view_schema", viewSchemaId: "test.range" } as const;
const anchor = (recordId: string, fieldKey = "summary"): GridCellAnchor => ({
  surface,
  rowIdentity: { kind: "core_record", recordId },
  fieldKey,
});
const a = anchor("a"),
  b = anchor("b"),
  c = anchor("c", "detail");
const input = { x: 100, y: 100, pointerId: 1, shiftKey: false };
const point = { x: 150, y: 180 };
const prior = { start: a, end: b };

function harness() {
  let snapshot: GridInteractionSnapshot = {
    active: b,
    editor: null,
    range: prior,
    model: {
      surface,
      fieldKeys: ["summary", "detail"],
      rowIdentities: [a, b, c].map((cell) => cell.rowIdentity),
    },
    enabled: true,
    available: true,
    scopeKey: "accepted-query",
    authorityKey: "editable",
  };
  const accept = vi.fn((range: GridCellRange) => {
    snapshot = { ...snapshot, range, active: range.end };
  });
  const changeRange = vi.fn((range: GridCellRange | null) => {
    snapshot = { ...snapshot, range };
  });
  const preview = vi.fn(),
    announce = vi.fn(),
    stopped = vi.fn();
  const controller = createGridInteractionController({
    read: () => snapshot,
    accept,
    changeRange,
    preview,
    announce,
    stopped,
  });
  controller.reconcile();
  return {
    controller,
    accept,
    changeRange,
    preview,
    announce,
    stopped,
    read: () => snapshot,
    patch: (patch: Partial<GridInteractionSnapshot>) => {
      snapshot = { ...snapshot, ...patch };
    },
  };
}

function deferred() {
  let resolve!: (value: boolean) => void;
  const promise = new Promise<boolean>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

afterEach(() => {
  vi.restoreAllMocks();
  document.body.replaceChildren();
});

describe("semantic range interaction", () => {
  it("classifies tolerance once and shares pointer and keyboard anchors through reversal", () => {
    const h = harness();
    expect(crossesGridDragThreshold(input, { x: 104, y: 96 })).toBe(false);
    expect(crossesGridDragThreshold(input, { x: 104.01, y: 100 })).toBe(true);
    h.controller.begin(a, input);
    h.controller.move(point, c);
    expect(h.read().range).toEqual(prior);
    expect(h.preview).toHaveBeenLastCalledWith({ start: a, end: c });
    h.controller.move(input, a);
    h.controller.release(input, a);
    expect(h.accept).toHaveBeenLastCalledWith(
      { start: a, end: a },
      false,
      false,
    );
    h.controller.begin(c, { ...input, shiftKey: true });
    h.controller.release(input, c);
    expect(h.read().range).toEqual(
      extendSemanticCellRange(h.read().model, a, { start: a, end: a }, c),
    );
    h.controller.begin(b, { ...input, shiftKey: true });
    h.controller.release(input, b);
    expect(h.read().range).toEqual(prior);
  });
  it("edits stationary clicks immediately and uses active or destination fallback for extension", () => {
    const h = harness();
    h.controller.begin(a, input);
    h.controller.release({ x: 103, y: 99 }, a);
    expect(h.accept).toHaveBeenCalledWith({ start: a, end: a }, true, true);
    h.patch({ range: null, active: b });
    h.controller.begin(c, { ...input, shiftKey: true });
    h.controller.release(input, c);
    expect(h.read().range).toEqual({ start: b, end: c });
    h.patch({ range: null, active: anchor("missing") });
    h.controller.begin(c, { ...input, shiftKey: true });
    h.controller.release(input, c);
    expect(h.read().range).toEqual({ start: c, end: c });
  });
  it("requires pointer adoption and rejects missing structural or hidden destinations", () => {
    const h = harness();
    h.patch({ enabled: false });
    expect(h.controller.begin(a, { ...input, shiftKey: true })).toBe(false);
    h.controller.begin(a, input);
    h.controller.move(point, c);
    h.controller.release(point, c);
    expect(h.accept).not.toHaveBeenCalled();
    h.patch({ enabled: true });
    expect(h.controller.begin(anchor("draft"), input)).toBe(false);
    expect(h.controller.begin(anchor("a", "hidden"), input)).toBe(false);
    h.controller.begin(a, input);
    h.controller.move(point, c);
    h.controller.release(point, anchor("draft"));
    expect(h.read().range).toEqual({ start: a, end: c });
  });
  it("keeps pending editor selection tentative and retains the rejected draft attachment", async () => {
    const h = harness(),
      pending = deferred();
    const editor = {
      target: {
        ...a,
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
      } as GridCellTarget,
      requestCommit: vi.fn(() => pending.promise),
      focus: vi.fn(),
      cancel: vi.fn(),
      detach: vi.fn(),
    };
    h.patch({ editor });
    h.controller.begin(b, input);
    h.controller.move(point, c);
    expect(h.controller.scrolling).toBe(false);
    h.controller.release(point, c);
    expect(h.accept).not.toHaveBeenCalled();
    expect(h.read().range).toEqual(prior);
    expect(editor.requestCommit).toHaveBeenCalledTimes(1);
    pending.resolve(false);
    await pending.promise;
    expect(editor.focus).toHaveBeenCalledTimes(1);
    expect(editor.cancel).not.toHaveBeenCalled();
    expect(h.preview).toHaveBeenLastCalledWith(null);
    expect(h.read().range).toEqual(prior);
  });
  it("applies only a current accepted destination and never discards superseded writes", async () => {
    const h = harness(),
      pending = deferred();
    const editor = {
      target: {
        ...a,
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
      } as GridCellTarget,
      requestCommit: vi.fn(() => pending.promise),
      focus: vi.fn(),
      cancel: vi.fn(),
      detach: vi.fn(),
    };
    h.patch({ editor });
    h.controller.begin(b, input);
    h.controller.move(point, c);
    h.controller.release(point, c);
    h.controller.begin(c, { ...input, shiftKey: true });
    h.controller.release(input, c);
    h.patch({ editor: null });
    pending.resolve(true);
    await pending.promise;
    expect(h.accept).toHaveBeenCalledTimes(1);
    expect(h.read().range).toEqual({ start: a, end: c });
    expect(editor.cancel).not.toHaveBeenCalled();
  });
  it("resumes scrolling after acceptance and cancels a released destination on newer authoring", async () => {
    const h = harness(),
      pending = deferred();
    const editor = {
      target: {
        ...a,
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 1 },
      } as GridCellTarget,
      requestCommit: () => pending.promise,
      focus: vi.fn(),
      cancel: vi.fn(),
      detach: vi.fn(),
    };
    h.patch({ editor });
    h.controller.begin(b, input);
    h.controller.move(point, c);
    h.patch({ editor: null });
    pending.resolve(true);
    await pending.promise;
    expect(h.controller.scrolling).toBe(true);
    expect(h.accept).not.toHaveBeenCalled();
    h.controller.release(point, c);
    expect(h.accept).toHaveBeenCalledWith({ start: b, end: c }, false, false);
    const next = deferred();
    h.patch({ editor: { ...editor, requestCommit: () => next.promise } });
    h.controller.begin(b, input);
    h.controller.release(input, b);
    h.controller.cancel();
    next.resolve(true);
    await next.promise;
    expect(h.accept).toHaveBeenCalledTimes(1);
  });
  it("restores prior selection on cancellation without retaining invalid or substituted membership", () => {
    const h = harness();
    h.controller.begin(a, input);
    h.controller.move(point, c);
    h.controller.cancel(true);
    expect(h.read().range).toEqual(prior);
    expect(h.announce).toHaveBeenLastCalledWith("Selection canceled.");
    h.controller.begin(a, input);
    const model = h.read().model;
    h.patch({
      model: {
        ...model,
        rowIdentities: [
          a.rowIdentity,
          anchor("inserted").rowIdentity,
          b.rowIdentity,
          c.rowIdentity,
        ],
      },
    });
    h.controller.reconcile();
    expect(h.controller.active).toBe(false);
    expect(h.read().range).toBeNull();
    h.patch({ model, range: prior });
    h.controller.reconcile();
    h.controller.begin(a, input);
    h.patch({ available: false });
    h.controller.reconcile();
    expect(h.controller.active).toBe(false);
    expect(h.read().range).toBeNull();
  });
  it("preserves exact members on append and invalidates accepted scope and authority intentions", () => {
    const h = harness();
    h.controller.begin(a, input);
    h.patch({
      model: {
        ...h.read().model,
        rowIdentities: [
          ...h.read().model.rowIdentities,
          anchor("appended").rowIdentity,
        ],
      },
    });
    h.controller.reconcile();
    h.controller.move(point, anchor("appended"));
    h.controller.release(point, null);
    expect(h.read().range).toEqual({ start: a, end: a });
    h.patch({ scopeKey: "replacement-query" });
    h.controller.reconcile();
    expect(h.read().range).toBeNull();
    h.controller.begin(a, input);
    h.patch({ authorityKey: "read-only" });
    h.controller.reconcile();
    expect(h.controller.active).toBe(false);
    expect(h.controller.begin(a, input)).toBe(true);
    const replacement = { start: b, end: c };
    h.patch({ range: replacement });
    h.controller.reconcile();
    expect(h.controller.active).toBe(false);
    h.controller.release(point, a);
    expect(h.read().range).toEqual(replacement);
  });
  it("bounds edge speed and elapsed work on both axes including narrow viewports", () => {
    expect(gridEdgeScrollDelta(132, 100, 500, 16)).toBe(0);
    expect(gridEdgeScrollDelta(116, 100, 500, 16)).toBeCloseTo(-5.76);
    expect(gridEdgeScrollDelta(700, 100, 500, 1000)).toBeCloseTo(23.04);
    expect(gridEdgeScrollDelta(0, 100, 500, 1000)).toBeCloseTo(-23.04);
    expect(gridEdgeScrollDelta(110, 100, 120, 32)).toBe(0);
    expect(gridEdgeScrollDelta(100, 100, 100, 32)).toBe(0);
  });
});

describe("range pointer binding", () => {
  it("scrolls both axes with bounded frame work and stops after outside release", () => {
    const removeListener = vi.spyOn(document, "removeEventListener");
    const h = harness();
    const root = document.createElement("div");
    root.style.width = "200px";
    root.style.boxSizing = "border-box";
    document.body.append(root);
    vi.spyOn(root, "getBoundingClientRect").mockReturnValue(
      new DOMRect(0, 0, 200, 100),
    );
    const frames = new Map<number, FrameRequestCallback>();
    let sequence = 0;
    vi.spyOn(window, "requestAnimationFrame").mockImplementation((callback) => {
      frames.set(++sequence, callback);
      return sequence;
    });
    vi.spyOn(window, "cancelAnimationFrame").mockImplementation((id) => {
      frames.delete(id);
    });
    const advance = (now: number) => {
      const current = [...frames.values()];
      frames.clear();
      current.forEach((callback) => {
        callback(now);
      });
    };
    Object.defineProperties(root, {
      clientWidth: { value: 200 },
      clientHeight: { value: 100 },
      scrollWidth: { value: 400 },
      scrollHeight: { value: 160 },
    });
    const scroll = vi.fn((options: ScrollToOptions) => {
      root.scrollLeft = options.left ?? 0;
      root.scrollTop = options.top ?? 0;
    });
    Object.assign(root, {
      scrollTo: scroll,
      setPointerCapture: vi.fn(),
      hasPointerCapture: () => true,
      releasePointerCapture: vi.fn(),
    });
    const cells = new Map<string, RegisteredGridCell>();
    [a, b, c].forEach((anchor, index) => {
      const cell = document.createElement("div");
      cell.setAttribute("role", "gridcell");
      root.append(cell);
      vi.spyOn(cell, "getBoundingClientRect").mockImplementation(
        () => new DOMRect(0, 40 + index * 40 - root.scrollTop, 200, 40),
      );
      cells.set(String(index), { anchor, cell, token: {} });
    });
    const binding = bindGridInteractionDom(root, h.controller, () => cells);
    const pointer = (
      type: string,
      x: number,
      y: number,
      target: HTMLElement = root,
    ) => {
      const event = new MouseEvent(type, {
        bubbles: true,
        cancelable: true,
        clientX: x,
        clientY: y,
        buttons: 1,
      });
      Object.defineProperties(event, {
        pointerId: { value: 7 },
        pointerType: { value: "mouse" },
      });
      target.dispatchEvent(event);
    };
    const firstCell = cells.get("0");
    if (!firstCell) throw new Error("Missing source cell");
    pointer("pointerdown", 10, 50, firstCell.cell);
    pointer("pointermove", 220, 120);
    advance(100);
    advance(1100);
    expect(scroll).toHaveBeenCalledTimes(1);
    expect(root.scrollLeft).toBeCloseTo(23.04);
    expect(root.scrollTop).toBeCloseTo(23.04);
    pointer("pointerup", 220, 120, document.body);
    expect(h.accept).toHaveBeenCalledTimes(1);
    expect(h.accept.mock.calls[0]?.[0].start).toEqual(a);
    expect(frames.size).toBe(0);
    for (const event of ["pointermove", "pointerup", "pointercancel"])
      expect(removeListener).toHaveBeenCalledWith(
        event,
        expect.any(Function),
        true,
      );
    advance(2100);
    expect(scroll).toHaveBeenCalledTimes(1);
    binding.dispose();
  });
  it("owns capture cleanup and compatibility clicks while preserving native controls", () => {
    const h = harness();
    const root = document.createElement("div"),
      cell = document.createElement("div"),
      button = document.createElement("button");
    cell.setAttribute("role", "gridcell");
    cell.append(button);
    root.append(cell);
    document.body.append(root);
    const cells = new Map<string, RegisteredGridCell>([
      ["a", { cell, anchor: a, token: {} }],
    ]);
    const capture = vi.fn(),
      release = vi.fn();
    Object.assign(root, {
      setPointerCapture: capture,
      hasPointerCapture: () => true,
      releasePointerCapture: release,
    });
    const raf = vi.spyOn(window, "requestAnimationFrame").mockReturnValue(12);
    const cancelFrame = vi
      .spyOn(window, "cancelAnimationFrame")
      .mockImplementation(() => {});
    const binding = bindGridInteractionDom(root, h.controller, () => cells);
    const pointer = (target: HTMLElement, type: string) => {
      const event = new MouseEvent(type, {
        bubbles: true,
        cancelable: true,
        clientX: 100,
        clientY: 100,
        button: 0,
        buttons: 1,
      });
      Object.defineProperties(event, {
        pointerId: { value: 1 },
        pointerType: { value: "mouse" },
      });
      target.dispatchEvent(event);
    };
    pointer(button, "pointerdown");
    expect(capture).not.toHaveBeenCalled();
    pointer(cell, "pointerdown");
    expect(capture).toHaveBeenCalledWith(1);
    expect(raf).toHaveBeenCalledTimes(1);
    pointer(root, "lostpointercapture");
    expect(h.controller.active).toBe(false);
    expect(cancelFrame).toHaveBeenCalledWith(12);
    const click = new MouseEvent("click", {
      bubbles: true,
      cancelable: true,
      detail: 1,
    });
    cell.dispatchEvent(click);
    expect(click.defaultPrevented).toBe(true);
    const draft = document.createElement("div");
    draft.dataset.cartularyGridDraftRow = "true";
    const input = document.createElement("input");
    draft.append(input);
    root.append(draft);
    h.patch({ range: { start: a, end: b } });
    input.focus();
    expect(h.read().range).toBeNull();
    pointer(cell, "pointerdown");
    binding.dispose();
    expect(release).toHaveBeenCalled();
    const count = capture.mock.calls.length;
    pointer(cell, "pointerdown");
    expect(capture).toHaveBeenCalledTimes(count);
    expect(h.accept).not.toHaveBeenCalled();
  });
});
