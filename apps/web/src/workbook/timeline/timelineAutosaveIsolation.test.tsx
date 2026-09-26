import type { SemanticDataGrid } from "@cartulary/grid-adapter";
import {
  rowCellTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  notesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { act, fireEvent, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { renderTimelineWorkbook } from "../../testing/timelineWorkbookRenderTestSupport";
import {
  cleanupTimelineWorkbookTestGlobals,
  installTimelineWorkbookTestGlobals,
  successEnvelope,
  timelineRow,
} from "../../testing/timelineWorkbookTestSupport";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { TimelineCollectionCell } from "./components/TimelineCollectionCell";

const observed = vi.hoisted(() => ({
  runtime: null as WorkbookMutationRuntime | null,
  grids: [] as { columns: unknown; rows: unknown }[],
  cells: new Map<string, number>(),
}));

// Observe the production composition at public component boundaries. No fiber
// flags, global browser handles, or test-specific memoization affect scheduling.
vi.mock("../runtime/createWorkbookMutationRuntime", async (original) => {
  const actual =
    await original<typeof import("../runtime/createWorkbookMutationRuntime")>();
  return {
    ...actual,
    createWorkbookMutationRuntime: (
      ...args: Parameters<typeof actual.createWorkbookMutationRuntime>
    ) => {
      const runtime = actual.createWorkbookMutationRuntime(...args);
      observed.runtime = runtime;
      return runtime;
    },
  };
});
vi.mock("@cartulary/grid-adapter", async () => {
  const actual = await import("@cartulary/grid-adapter/test-support");
  const { createElement, useLayoutEffect } = await import("react");
  return {
    ...actual,
    SemanticDataGrid: function ObservedGrid(
      props: ComponentProps<typeof SemanticDataGrid>,
    ) {
      useLayoutEffect(() => {
        observed.grids.push({ columns: props.columns, rows: props.dataRows });
      });
      return createElement(actual.SemanticDataGrid, props);
    },
  };
});
vi.mock("./components/TimelineCollectionCell", async (original) => {
  const actual =
    await original<typeof import("./components/TimelineCollectionCell")>();
  const { createElement, Profiler } = await import("react");
  return {
    ...actual,
    TimelineCollectionCell: function ObservedCollection(
      props: ComponentProps<typeof TimelineCollectionCell>,
    ) {
      const key = `${props.row.key}:${props.binding.fieldKey}:${props.surface}`;
      return createElement(
        Profiler,
        {
          id: key,
          onRender: () =>
            observed.cells.set(key, (observed.cells.get(key) ?? 0) + 1),
        },
        createElement(actual.TimelineCollectionCell, props),
      );
    },
  };
});

const recordId = "20000000-0000-4000-8000-000000000001";
let fetchMock: ReturnType<typeof installTimelineWorkbookTestGlobals>;
beforeEach(() => {
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  fetchMock = installTimelineWorkbookTestGlobals();
  observed.runtime = null;
  observed.grids = [];
  observed.cells.clear();
  fetchMock.mockResolvedValue(
    successEnvelope({
      incident_id: "10000000-0000-4000-8000-000000000001",
      view_schema_id: timelineViewSchemaId,
      rows: [
        timelineRow({
          recordId,
          rowVersion: 1,
          summary: "Original",
          captureState: "rough",
        }),
      ],
    }),
  );
});
afterEach(() => {
  cleanupTimelineWorkbookTestGlobals();
  vi.useRealTimers();
});

async function mounted() {
  renderTimelineWorkbook({
    layoutState: {
      ...defaultWorkbookLayoutState(requireViewContract(timelineViewSchemaId)),
      hiddenFieldKeys: [],
    },
  });
  await act(async () => {
    await vi.advanceTimersByTimeAsync(0);
  });
  expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
  const runtime = observed.runtime;
  if (!runtime) throw new Error("Fixture did not construct the runtime");
  return runtime;
}
function cellCommits() {
  return new Map(observed.cells);
}
function assertIsolated(
  cells: Map<string, number>,
  grid: { columns: unknown; rows: unknown },
  start: number,
) {
  expect(
    [...observed.cells].filter(([key, count]) => count !== cells.get(key)),
  ).toEqual([]);
  for (const next of observed.grids.slice(start)) {
    expect(next.columns).toBe(grid.columns);
    expect(next.rows).toBe(grid.rows);
  }
}
function statusRow(id: string) {
  return {
    record_id: id,
    row_version: 1,
    view_schema_id: notesViewSchemaId,
    cells: Object.fromEntries(
      requireViewContract(notesViewSchemaId).fields.map((field) => [
        field.fieldKey,
        { value: null },
      ]),
    ),
  };
}
function configureNotes(runtime: WorkbookMutationRuntime) {
  runtime.registerSurface(notesViewSchemaId, async () => {});
  runtime.explicitPatches.configure(
    {
      send: async (request) => ({
        kind: "acknowledged",
        receipt: {
          changeSetId: request.clientTxnId,
          viewSchemaId: notesViewSchemaId,
          row: { ...statusRow(request.recordId), row_version: 2 },
        },
      }),
    },
    undefined,
    async (_view, id) => statusRow(id),
  );
}
function admit(runtime: WorkbookMutationRuntime, id: string) {
  const preparation = deferred<void>();
  const submitted = runtime.explicitPatches.submit(
    {
      baseline: statusRow(id),
      viewSchemaId: notesViewSchemaId,
      changes: [{ field_key: "note.body", value: "Authored" }],
      purpose: "generic-patch",
      sheetRef: { kind: "view_schema", id: notesViewSchemaId },
      surfaceLabel: "Notes",
    },
    [{ prepare: () => preparation.promise }],
  );
  return async () => {
    preparation.resolve();
    await submitted;
  };
}

it("Timeline composition isolates unchanged publications and overlapping save status from grid and inspector cells", async () => {
  const runtime = await mounted();
  fireEvent.click(
    screen.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  );
  fireEvent.click(
    screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  );
  await act(async () => {
    await vi.advanceTimersByTimeAsync(0);
  });
  expect(
    [...observed.cells.keys()].some((key) => key.endsWith(":inspector")),
  ).toBe(true);
  act(() => configureNotes(runtime));
  const grid = observed.grids.at(-1);
  if (!grid) throw new Error("Timeline grid did not mount");
  const cells = cellCommits(),
    start = observed.grids.length;
  const snapshot = runtime.getSnapshot();
  const published = vi.fn();
  const unsubscribe = runtime.subscribe(published);
  await act(async () => runtime.notifyPendingChanged());
  expect(runtime.getSnapshot()).toBe(snapshot);
  expect(published).toHaveBeenCalledOnce();
  assertIsolated(cells, grid, start);
  let first!: () => Promise<void>, second!: () => Promise<void>;
  await act(async () => {
    first = admit(runtime, "note-1");
  });
  const one = runtime.getSnapshot();
  await act(async () => {
    second = admit(runtime, "note-2");
  });
  const two = runtime.getSnapshot();
  expect(one.primaryLabel).toBe("Syncing");
  expect(two.primaryLabel).toBe("Syncing");
  expect(two.explicitInFlightCount).toBe(2);
  expect(two).not.toBe(one);
  assertIsolated(cells, grid, start);
  await act(first);
  expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
  expect(two.explicitInFlightCount).toBe(2);
  assertIsolated(cells, grid, start);
  await act(second);
  expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
  assertIsolated(cells, grid, start);
  unsubscribe();
});

it("Timeline composition observes editor activation drafts inspector changes and replacement conflicts independently", async () => {
  const runtime = await mounted();
  const key = `${recordId}:timeline.tags:grid`;
  const before = observed.cells.get(key) ?? 0;
  const buttons = screen.getAllByRole("button", {
    name: "Add tags token",
  });
  const add = buttons[0];
  if (!add) throw new Error("Tags editor activation is absent");
  fireEvent.click(add);
  const input = screen.getByTestId(
    timelineCollectionInputTestId(recordId, "timeline.tags", "grid"),
  ) as HTMLInputElement;
  expect(observed.cells.get(key)).toBeGreaterThan(before);
  const active = observed.cells.get(key) ?? 0;
  const editorPublication = deferred<void>();
  const delayedEditor = editorPublication.promise.then(() =>
    fireEvent.input(input, { target: { value: "raw Ω draft" } }),
  );
  await act(async () => runtime.notifyPendingChanged());
  expect(observed.cells.get(key)).toBe(active);
  await act(async () => {
    editorPublication.resolve();
    await delayedEditor;
  });
  expect(observed.cells.get(key)).toBeGreaterThan(active);
  input.setSelectionRange(2, 5, "backward");
  await act(async () => runtime.notifyPendingChanged());
  expect(
    screen.getByTestId(
      timelineCollectionInputTestId(recordId, "timeline.tags", "grid"),
    ),
  ).toBe(input);
  expect(input.value).toBe("raw Ω draft");
  expect([
    input.selectionStart,
    input.selectionEnd,
    input.selectionDirection,
  ]).toEqual([2, 5, "backward"]);
  fireEvent.click(
    screen.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  );
  fireEvent.click(
    screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  );
  expect(
    [...observed.cells.keys()].some((cell) => cell.endsWith(":inspector")),
  ).toBe(true);
  const register = (token: string, version: number) =>
    runtime.registerConflict({
      viewSchemaId: timelineViewSchemaId,
      rowLabel: "Row",
      surfaceLabel: "Timeline",
      conflict: {
        record_id: recordId,
        field_key: "timeline.activity_synopsis_text",
        base_row_version: 1,
        current_row_version: version,
        client_value: "draft",
        server_value: "saved",
        conflict_token: token,
        conflict_resolution_class: "text_compare_merge",
      },
    });
  const initialGrid = observed.grids.at(-1);
  await act(async () => register("first", 2));
  const first = runtime.getSnapshot();
  expect(first.primaryLabel).toBe("Conflict");
  expect(observed.grids.at(-1)).not.toBe(initialGrid);
  const previous = cellCommits();
  await act(async () => register("replacement", 3));
  expect(runtime.getSnapshot().conflicts[0]?.conflict.conflict_token).toBe(
    "replacement",
  );
  expect(first.conflicts[0]?.conflict.conflict_token).toBe("first");
  expect(observed.cells).not.toEqual(previous);
  await act(async () => {
    for (const conflict of runtime.getSnapshot().conflicts)
      runtime.clearConflict(conflict.key);
  });
  const rowsBeforeRefresh = observed.grids.at(-1)?.rows;
  const cellsBeforeRefresh = cellCommits();
  fetchMock.mockResolvedValue(
    successEnvelope({
      incident_id: "10000000-0000-4000-8000-000000000001",
      view_schema_id: timelineViewSchemaId,
      rows: [
        timelineRow({
          recordId,
          rowVersion: 4,
          summary: "Accepted refresh",
          captureState: "rough",
          tags: ["saved-token"],
        }),
      ],
    }),
  );
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await vi.advanceTimersByTimeAsync(0);
  });
  expect(observed.grids.at(-1)?.rows).not.toBe(rowsBeforeRefresh);
  expect(observed.cells).not.toEqual(cellsBeforeRefresh);
  expect(
    screen.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ).textContent,
  ).toContain("Accepted refresh");
});

it("Timeline inspector isolates unrelated drafts and observes selected retained work and authority", async () => {
  const runtime = await mounted();
  fireEvent.click(
    screen.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  );
  fireEvent.click(
    screen.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
  );
  await act(async () => {
    await vi.advanceTimersByTimeAsync(0);
  });
  const before = cellCommits();
  const update = (id: string) =>
    runtime.inspectorDrafts.update(
      {
        viewSchemaId: timelineViewSchemaId,
        recordId: id,
        fieldKey: "timeline.activity_synopsis_text",
        action: "value",
      },
      {
        ...timelineRow({
          recordId: id,
          rowVersion: 1,
          summary: "Original",
          captureState: "rough",
        }),
      },
      "retained draft",
      "detached-inspector",
    );
  await act(async () => update("unrelated-record"));
  expect(observed.cells).toEqual(before);
  await act(async () => update(recordId));
  expect(observed.cells).not.toEqual(before);
  expect(screen.getAllByText(/Unsaved draft:/).length).toBeGreaterThan(0);
  await act(async () => runtime.setAuthority(null));
  expect(screen.queryAllByText(/Unsaved draft:/)).toHaveLength(0);
  expect(
    runtime.inspectorDrafts.readRecord(timelineViewSchemaId, recordId),
  ).toEqual([]);
});
