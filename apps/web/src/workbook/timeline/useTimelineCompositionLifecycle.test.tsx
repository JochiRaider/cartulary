import {
  act,
  fireEvent,
  render,
  renderHook,
  screen,
} from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type {
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../continuity/workbookContinuityPort";
import { useTimelineGridEnvironment } from "./composition/useTimelineGridEnvironment";
import { useTimelineInspectorStateComposition } from "./composition/useTimelineInspectorStateComposition";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { createDraftRow } from "./models/workbookTimelineModel";

it("useTimelineGridEnvironment owns rounded measurement and observer cleanup", () => {
  let clientWidth = 413.8;
  const disconnect = vi.fn();
  const observe = vi.fn();
  class TestResizeObserver {
    readonly disconnect = disconnect;
    readonly observe = observe;
    readonly unobserve = vi.fn();
  }
  vi.stubGlobal("ResizeObserver", TestResizeObserver);
  const clientWidthSpy = vi
    .spyOn(HTMLElement.prototype, "clientWidth", "get")
    .mockImplementation(() => clientWidth);
  const requestAnimationFrameSpy = vi
    .spyOn(window, "requestAnimationFrame")
    .mockImplementation((callback) => {
      callback(0);
      return 1;
    });
  const editorDraftRegistry = createTimelineEditorDraftRegistry();

  function GridEnvironmentHarness() {
    const grid = useTimelineGridEnvironment({
      continuityResetKey: "continuity-1",
      editorDraftRegistry,
      rowsRef: { current: [] },
    });
    return (
      <div ref={grid.refs.gridShell}>
        <output>{grid.snapshot.gridShellWidth}</output>
      </div>
    );
  }

  const view = render(<GridEnvironmentHarness />);
  expect(screen.getByText("413")).toBeTruthy();
  expect(observe).toHaveBeenCalledTimes(2);

  clientWidth = 519.9;
  fireEvent(window, new Event("resize"));
  expect(screen.getByText("519")).toBeTruthy();

  view.unmount();
  expect(disconnect).toHaveBeenCalledTimes(1);

  requestAnimationFrameSpy.mockRestore();
  clientWidthSpy.mockRestore();
  vi.unstubAllGlobals();
});

it("useTimelineInspectorStateComposition preserves continuity and resets selection lifecycle", () => {
  const token = "continuity-1" as WorkbookContinuityToken;
  const capture = vi.fn(() => token);
  const restore = vi.fn(() => true);
  const continuity: WorkbookContinuityPort = {
    capture,
    clear: vi.fn(),
    dispose: vi.fn(),
    focus: vi.fn(() => true),
    restore,
    select: vi.fn(),
    snapshot: vi.fn(() => ({ anchor: null })),
  };
  const committedRow = {
    ...createDraftRow(2),
    key: "record-1",
    recordId: "record-1",
    rowVersion: 3,
  };
  const workbookFocusAnchorRef = {
    current: {
      fieldKey: "timeline.activity_synopsis_text",
      recordId: "record-1",
      viewSchemaId: "core.timeline.v1",
    },
  };
  const { result, rerender } = renderHook(
    ({ inspectorResetKey }) =>
      useTimelineInspectorStateComposition({
        continuity,
        currentIncidentRole: "editor",
        dismissedMentionsByRow: {},
        inspectorResetKey,
        rows: [committedRow],
        selectedMentionRef: null,
        workbookFocusAnchorRef,
      }),
    { initialProps: { inspectorResetKey: "inspector-1" } },
  );

  act(() => {
    result.current.commands.selectRow("record-1");
    result.current.commands.setOpen(true);
  });
  expect(result.current.snapshot.selection.selectedRowId).toBe("record-1");
  expect(result.current.snapshot.lifecycle.isOpen).toBe(true);
  expect(capture).toHaveBeenCalledWith();

  act(() => {
    result.current.commands.setOpen(false);
  });
  expect(capture).toHaveBeenLastCalledWith(workbookFocusAnchorRef.current);

  rerender({ inspectorResetKey: "inspector-2" });
  expect(result.current.snapshot.selection.selectedRowId).toBeNull();
});
