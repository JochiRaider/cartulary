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
import {
  type RecordHistoryData,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "../inspector/workbookRecordHistoryModel";
import { workbookInspectorStateIsOpen } from "../models/workbookInspectorModel";
import { useTimelineGridEnvironment } from "./composition/useTimelineGridEnvironment";
import { useTimelineInspectorStateComposition } from "./composition/useTimelineInspectorStateComposition";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineHistoryActions } from "./hooks/useTimelineHistoryActions";
import { createDraftRow } from "./models/timelineRowModel";
import type { TimelineHistoryPort } from "./ports/TimelineHistoryPort";

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
  const token: WorkbookContinuityToken = { sequence: 1 };
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
  expect(workbookInspectorStateIsOpen(result.current.snapshot.lifecycle)).toBe(
    true,
  );
  expect(capture).toHaveBeenCalledWith();

  act(() => {
    result.current.commands.setOpen(false);
  });
  expect(capture).toHaveBeenLastCalledWith(workbookFocusAnchorRef.current);

  rerender({ inspectorResetKey: "inspector-2" });
  expect(result.current.snapshot.selection.selectedRowId).toBeNull();
});

it("useTimelineHistoryActions preserves the committed delete ordering trace", async () => {
  const trace: string[] = [];
  const subject = {
    kind: "live" as const,
    label: "Timeline row",
    recordId: "record-1",
    rowVersion: 4,
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  };
  const data: RecordHistoryData = {
    deleted: false,
    incident_id: "incident-1",
    items: [],
    record_id: subject.recordId,
    row_version: subject.rowVersion,
  };
  const pendingAction = {
    kind: "destructive" as const,
    operation: "delete" as const,
    recordId: subject.recordId,
    rowVersion: subject.rowVersion,
  };
  const operationId = workbookRecordHistoryOperationId(1);
  let historyState: WorkbookRecordHistoryState = {
    pendingAction,
    phase: "ready",
    result: { data, kind: "loaded" },
    subject,
  };
  const dispatchRowHistory = (event: WorkbookRecordHistoryEvent) => {
    trace.push(`history:${event.type}`);
    historyState = workbookRecordHistoryReducer(historyState, event);
    return historyState;
  };
  let queuedWork: (() => Promise<void>) | null = null;
  const historyPort: TimelineHistoryPort = {
    deleteOrRestore: vi.fn(async () => {
      trace.push("route:delete");
      return {
        kind: "accepted" as const,
        value: { recordId: subject.recordId, rowVersion: 5 },
      };
    }),
    load: vi.fn(async () => {
      trace.push("route:load");
      return {
        kind: "accepted" as const,
        value: { ...data, deleted: true, row_version: 5 },
      };
    }),
    rollback: vi.fn(),
  };
  const { result } = renderHook(() =>
    useTimelineHistoryActions({
      acceptTimelineRecordVersion: (_recordId, rowVersion) =>
        trace.push(`version:${rowVersion}`),
      activeHistoryLiveRecordId: subject.recordId,
      activeHistorySubject: subject,
      beginRowHistoryOperation: () => operationId,
      beginRowHistoryRequest: () => {
        trace.push("history:request");
        return workbookRecordHistoryRequestId(2);
      },
      beginSave: () => trace.push("save:begin"),
      beginViewportContinuity: () => {
        trace.push("viewport:begin");
        return 41;
      },
      clearViewportContinuity: () => trace.push("viewport:clear"),
      currentHistoryRecordId: subject.recordId,
      currentHistoryRecordIdMatches: (recordId) =>
        recordId === subject.recordId,
      currentHistoryRowVersion: subject.rowVersion,
      dispatchRowHistory,
      enqueueSaveWork: (work) => {
        trace.push("save:enqueue");
        queuedWork = work;
      },
      finishSave: (state) => trace.push(`save:finish:${state}`),
      historyPort,
      loadRows: async () => {
        trace.push("rows:load");
      },
      nextClientTxnId: () => {
        trace.push("txn:next");
        return "txn-1";
      },
      resolvePendingSocketTxn: () => trace.push("socket:resolve"),
      retargetRowHistory: (nextSubject) => {
        dispatchRowHistory({ subject: nextSubject, type: "retarget" });
      },
      rowHistory: historyState,
      rowHistoryPendingAction: pendingAction,
      rowHistoryRequestIsCurrent: () => true,
      selectedRowRecordId: subject.recordId,
      setIsInspectorOpen: vi.fn(),
      setSelectedRowId: vi.fn(),
      trackPendingSocketTxn: () => trace.push("socket:track"),
      waitForCommittedRecordIdle: async () => {
        trace.push("record:idle");
        return { row: null, rowVersion: subject.rowVersion };
      },
    }),
  );

  act(() => result.current.confirmRowHistoryPendingAction());
  expect(queuedWork).not.toBeNull();
  await act(async () => {
    await queuedWork?.();
  });

  expect(trace).toEqual([
    "history:submit",
    "txn:next",
    "viewport:begin",
    "save:begin",
    "save:enqueue",
    "record:idle",
    "socket:track",
    "route:delete",
    "version:5",
    "history:operation_accepted",
    "history:request",
    "history:retarget",
    "history:load_requested",
    "route:load",
    "version:5",
    "history:load_accepted",
    "rows:load",
    "save:finish:Saved",
  ]);
});
