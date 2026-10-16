// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import { installAnchorTrace } from "./anchorTraceSupport";

afterEach(() => {
  window.cartularyAnchorTrace?.stop();
  delete window.cartularyAnchorTrace;
  document.body.replaceChildren();
  vi.unstubAllGlobals();
});

function fixture() {
  document.body.innerHTML =
    '<div role="grid"><div role="row" data-grid-record-id="a" data-grid-row-version="3"><input data-editor value="draft"></div><div role="row" data-grid-record-id="b" data-grid-row-version="1"></div></div>';
  let scheduled: FrameRequestCallback | undefined;
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    scheduled = callback;
    return 1;
  });
  vi.stubGlobal("cancelAnimationFrame", () => {});
  installAnchorTrace({
    recordId: "a",
    initialVersion: 3,
    gridSelector: '[role="grid"]',
    editorSelector: "[data-editor]",
    draftValue: "draft",
    totalRows: 20000,
  });
  const row = document.querySelector('[data-grid-record-id="a"]');
  const grid = document.querySelector('[role="grid"]');
  const editor = document.querySelector<HTMLInputElement>("[data-editor]");
  const trace = window.cartularyAnchorTrace;
  if (!row || !grid || !editor || !trace)
    throw new Error("Missing trace fixture");
  return {
    row,
    grid,
    editor,
    trace,
    advance: () => {
      if (!scheduled) throw new Error("Missing frame callback");
      scheduled(10);
    },
  };
}

it("retains frame evidence and detects retargeting, draft changes and version rollback", () => {
  const { row, editor, trace, advance } = fixture();
  expect(() => trace.checkpoint("draft_open")).toThrow("two observed frames");
  advance();
  advance();
  trace.checkpoint("draft_open");
  expect(trace.read().violations).toEqual([]);
  row.setAttribute("data-grid-row-version", "4");
  advance();
  row.setAttribute("data-grid-row-version", "3");
  advance();
  editor.value = "lost";
  advance();
  row.setAttribute("data-grid-record-id", "b");
  advance();
  expect(trace.read().violations).toEqual([
    "row_version_regressed",
    "draft_changed",
    "anchor_missing",
    "editor_retargeted",
  ]);
  trace.draft = false;
  row.setAttribute("data-grid-row-version", "invalid");
  advance();
  expect(trace.read().violations).toContain("row_binding_invalid");
  row.setAttribute("data-grid-row-version", "4");
  row.setAttribute("data-grid-record-id", "a");
  editor.value = "draft";
  const beforePending = trace.read().frame_count;
  advance();
  expect(trace.read().frame_count).toBe(beforePending + 1);
  expect(trace.read().checkpoints).toEqual([
    { step: "draft_open", frame_count: 2, row_version: 3, mounted_rows: 2 },
  ]);
  trace.stop();
  delete window.cartularyAnchorTrace;
  const pending = fixture();
  pending.trace.draft = false;
  pending.row.setAttribute("data-grid-record-id", "other");
  pending.advance();
  expect(pending.trace.read().violations).toEqual(["editor_retargeted"]);
});

it("follows viewport replacement and detects full-sheet mounting without changing the grid", () => {
  const { grid, trace, advance } = fixture();
  for (let i = 2; i < 20000; i++) {
    const row = document.createElement("div");
    row.setAttribute("role", "row");
    row.setAttribute("data-grid-record-id", String(i));
    row.setAttribute("data-grid-row-version", "1");
    grid.append(row);
  }
  advance();
  expect(trace.read().violations).toContain("full_sheet_mounted");
  grid.replaceWith(grid.cloneNode(true));
  advance();
  expect(trace.read().grid_replacements).toBe(1);
  expect(trace.read().violations).toEqual(["full_sheet_mounted"]);
});
