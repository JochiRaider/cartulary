import {
  act,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { type Dispatch, useReducer } from "react";
import { expect, it, vi } from "vitest";
import { historyPresentationFixture } from "../../testing/workbookHistoryTestSupport";
import type { RecordHistoryData } from "../adapters/workbookHistoryResponse";
import type {
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../continuity/workbookContinuityPort";
import { WorkbookHistoryContext } from "../history/WorkbookHistoryContext";
import { WorkbookRecordHistoryOwner } from "../history/WorkbookRecordHistoryOwner";
import {
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryState,
  workbookRecordHistoryReducer,
} from "../inspector/workbookRecordHistoryModel";
import { workbookInspectorStateIsOpen } from "../models/workbookInspectorModel";
import { useTimelineGridEnvironment } from "./composition/useTimelineGridEnvironment";
import { useTimelineInspectorStateComposition } from "./composition/useTimelineInspectorStateComposition";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineHistoryActions } from "./hooks/useTimelineHistoryActions";
import { createDraftRow } from "./models/timelineRowModel";

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
  const continuityScope = {
    getSnapshot: () => ({ key: "continuity-1", readable: true }),
    subscribe: () => () => {},
  };

  function GridEnvironmentHarness() {
    const grid = useTimelineGridEnvironment({
      continuityResetKey: "continuity-1",
      continuityScope,
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
  const restore = vi.fn(async () => true);
  const continuity: WorkbookContinuityPort = {
    capture,
    clear: vi.fn(),
    dispose: vi.fn(),
    focus: vi.fn(async () => true),
    restore,
    select: vi.fn(),
    snapshot: vi.fn(() => ({ anchor: null })),
  };
  const committedRow = {
    ...createDraftRow(2),
    key: "record-1",
    recordId: "record-1",
    rowVersion: 3,
    rawRow: {
      record_id: "record-1",
      row_version: 3,
      view_schema_id: "core.timeline.v1",
      cells: {},
      observation: {
        recordId: "record-1",
        rowVersion: 3,
        scope: {
          actorId: "actor",
          sessionIdentity: "session",
          incidentId: "incident",
          epoch: 0,
        },
      },
    },
  };
  const workbookFocusAnchorRef = {
    current: {
      fieldKey: "timeline.activity_synopsis_text",
      recordId: "record-1",
      viewSchemaId: "core.timeline.v1",
    },
  };
  const { result, rerender } = renderHook(
    ({ inspectorResetKey, currentIncidentRole }) =>
      useTimelineInspectorStateComposition({
        continuity,
        currentIncidentRole,
        dismissedMentionsByRow: {},
        observedMentions: [],
        inspectorResetKey,
        readScope: {
          actorId: "actor",
          sessionIdentity: "session",
          incidentId: "incident",
          epoch: 0,
        },
        rows: [committedRow],
        selectedMentionRef: null,
        workbookFocusAnchorRef,
      }),
    {
      initialProps: {
        inspectorResetKey: "inspector-1",
        currentIncidentRole: "reviewer",
      },
    },
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
  rerender({ inspectorResetKey: "inspector-1", currentIncidentRole: "editor" });
  expect(result.current.snapshot.selection.selectedRow).toBe(committedRow);
  expect(result.current.snapshot.selection.selectedRowId).toBe("record-1");

  act(() => {
    result.current.commands.setOpen(false);
  });
  expect(capture).toHaveBeenLastCalledWith(workbookFocusAnchorRef.current);

  rerender({ inspectorResetKey: "inspector-2", currentIncidentRole: "editor" });
  expect(result.current.snapshot.selection.selectedRowId).toBeNull();
  act(() => {
    result.current.commands.selectRow("record-1");
    result.current.commands.setOpen(true);
  });
  expect(result.current.snapshot.selection.selectedRow).toBe(committedRow);

  rerender({ inspectorResetKey: "inspector-2", currentIncidentRole: "" });
  expect(result.current.snapshot.selection.selectedRow).toBeNull();
  rerender({ inspectorResetKey: "inspector-3", currentIncidentRole: "editor" });
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
    representation_generation: "cartulary.history.1",
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
  let historyState: WorkbookRecordHistoryState = {
    ...historyPresentationFixture(subject, {
      ...data,
      paging: { limit: 100, has_more: false, next_cursor: null },
    }),
    pendingAction,
  };
  let publish: Dispatch<WorkbookRecordHistoryEvent> | undefined;
  const dispatchRowHistory = (event: WorkbookRecordHistoryEvent) => {
    trace.push(`history:${event.type}`);
    historyState = workbookRecordHistoryReducer(historyState, event);
    publish?.(event);
    return historyState;
  };
  let queuedWork: (() => Promise<void>) | null = null;
  const owner = new WorkbookRecordHistoryOwner("incident-1", {
    create: () => {
      trace.push("txn:next");
      return "txn-1";
    },
  });
  owner.setAuthority({
    actorId: "reviewer",
    incidentId: "incident-1",
    role: "reviewer",
    closed: false,
  });
  let committed = false;
  owner.configure({
    load: async () => {
      trace.push("route:load");
      return {
        kind: "accepted",
        value: {
          paging: { limit: 100, has_more: false, next_cursor: null },
          ...data,
          deleted: committed,
          row_version: committed ? 5 : 4,
        },
      };
    },
    send: async () => {
      trace.push("route:delete");
      committed = true;
      return {
        kind: "acknowledged",
        receipt: {
          kind: "delete",
          incidentId: "incident-1",
          recordId: subject.recordId,
          rowVersion: 5,
          deleted: true,
          deletedAt: "2026-09-10T00:00:00Z",
          deletedByUserId: "reviewer",
          changeSetId: "change-1",
        },
      };
    },
  });
  const { result } = renderHook(
    () => {
      const [state, dispatch] = useReducer(
        workbookRecordHistoryReducer,
        historyState,
      );
      publish = dispatch;
      return useTimelineHistoryActions({
        acceptTimelineRecordVersion: (_recordId, rowVersion) =>
          trace.push(`version:${rowVersion}`),
        activeHistorySubject: subject,
        dispatchRowHistory,
        enqueueSaveWork: (work) => {
          trace.push("save:enqueue");
          queuedWork = work;
        },
        loadRows: async () => {
          trace.push("rows:load");
        },
        rowHistory: state,
        setIsInspectorOpen: vi.fn(),
        setSelectedRowId: vi.fn(),
        waitForCommittedRecordIdle: async () => {
          trace.push("record:idle");
          return { row: null, rowVersion: subject.rowVersion };
        },
      });
    },
    {
      wrapper: ({ children }) => (
        <WorkbookHistoryContext.Provider
          value={{ history: owner, coordinateHistory: async () => 4 }}
        >
          {children}
        </WorkbookHistoryContext.Provider>
      ),
    },
  );

  trace.length = 0;
  act(() => {
    result.current.confirmRowHistoryPendingAction();
    result.current.confirmRowHistoryPendingAction();
  });
  expect(queuedWork).not.toBeNull();
  await act(async () => {
    await queuedWork?.();
  });

  await waitFor(() =>
    expect(owner.getSnapshot()[0]?.reconciliation).toBe("complete"),
  );
  expect(trace).toEqual([
    "history:submit",
    "txn:next",
    "save:enqueue",
    "record:idle",
    "route:load",
    "route:delete",
    "history:operation_accepted",
    "version:5",
    "route:load",
    "history:browsing_changed",
    "rows:load",
  ]);
});

it("Timeline history rejects queued version changes without replacing the confirmed base", async () => {
  const subject = {
    kind: "live" as const,
    label: "Row",
    recordId: "record-queued",
    rowVersion: 4,
    surfaceLabel: "Timeline",
    viewSchemaId: "cartulary.view.timeline.v2",
  };
  const data: RecordHistoryData = {
    deleted: false,
    incident_id: "incident-queued",
    representation_generation: "cartulary.history.1",
    items: [],
    record_id: subject.recordId,
    row_version: 4,
  };
  let snapshot: WorkbookRecordHistoryState = {
    ...historyPresentationFixture(subject, {
      ...data,
      paging: { limit: 100, has_more: false, next_cursor: null },
    }),
    pendingAction: {
      kind: "destructive",
      operation: "delete",
      recordId: subject.recordId,
      rowVersion: 4,
    },
  };
  let queuedWork: (() => Promise<void>) | undefined;
  const owner = new WorkbookRecordHistoryOwner("incident-queued", {
    create: () => "queued-action",
  });
  owner.setAuthority({
    actorId: "reviewer",
    incidentId: "incident-queued",
    role: "reviewer",
    closed: false,
  });
  const send = vi.fn(async () => ({ kind: "uncertain" as const }));
  owner.configure({
    load: async () => ({
      kind: "accepted",
      value: {
        paging: { limit: 100, has_more: false, next_cursor: null },
        ...data,
        row_version: 6,
      },
    }),
    send,
  });
  const { result } = renderHook(
    () =>
      useTimelineHistoryActions({
        activeHistorySubject: subject,
        rowHistory: snapshot,
        dispatchRowHistory: (event) => {
          snapshot = workbookRecordHistoryReducer(snapshot, event);
          return snapshot;
        },
        acceptTimelineRecordVersion: vi.fn(),
        enqueueSaveWork: (work) => {
          queuedWork = work;
        },
        waitForCommittedRecordIdle: async () => ({ row: null, rowVersion: 6 }),
        loadRows: vi.fn(async () => {}),
        setIsInspectorOpen: vi.fn(),
        setSelectedRowId: vi.fn(),
      }),
    {
      wrapper: ({ children }) => (
        <WorkbookHistoryContext.Provider
          value={{ history: owner, coordinateHistory: async () => 6 }}
        >
          {children}
        </WorkbookHistoryContext.Provider>
      ),
    },
  );
  act(() => result.current.confirmRowHistoryPendingAction());
  expect(queuedWork).toBeDefined();
  await act(async () => {
    await queuedWork?.();
  });
  await waitFor(() => expect(owner.getSnapshot()[0]?.phase).toBe("rejected"));
  expect(send).not.toHaveBeenCalled();
  expect(owner.getSnapshot()[0]?.dispatched).toBe(false);
  expect(
    JSON.parse(owner.getSnapshot()[0]?.attempt.body ?? "{}").base_row_version,
  ).toBe(4);
});
