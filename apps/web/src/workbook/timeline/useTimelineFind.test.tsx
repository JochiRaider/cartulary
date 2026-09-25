import type {
  GridEditCommitOutcome,
  GridHandle,
  GridPresentationSnapshot,
} from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { WorkbookQueryBrowser } from "../query/WorkbookQueryBrowser";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineFindSource } from "./hooks/useTimelineFindSource";
import { createDraftRow, rowFromApi } from "./models/timelineRowModel";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});
async function fixture() {
  const contract = requireViewContract(timelineViewSchemaId);
  const state = emptyWorkbookQueryState();
  const rows = ["one", "two", "pin"].map((id) =>
    rowFromApi({
      ...fullWorkbookViewRow(contract, id, 1, {
        "timeline.activity_synopsis_text": "committed needle",
      }),
      view_schema_id: timelineViewSchemaId,
    }),
  );
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
    kind: "accepted",
    value: {
      incidentId: "incident",
      rows: rows
        .slice(0, 2)
        .map(
          (row) =>
            row.rawRow ?? { record_id: row.key, row_version: 1, cells: {} },
        ),
      viewSchemaId: timelineViewSchemaId,
      canonicalQuery: { filters: [], sort: [] },
      paging: { limit: 100, hasMore: false, nextCursor: null },
      producingRequest: { queryState: state, limit: 100 },
    },
  }));
  const browser = new WorkbookQueryBrowser({ query }, timelineViewSchemaId);
  const response = await browser.query({
    contract,
    queryState: state,
    signal: new AbortController().signal,
  });
  if (response.kind !== "accepted") throw new Error("fixture query failed");
  browser.accept(response.value);
  const listeners = new Set<() => void>();
  let authority = true;
  const input: Parameters<typeof useTimelineFindSource>[0] = {
    runtime: {
      incident: {
        id: "incident",
        continuityResetKey: "incident",
        sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        currentRole: "editor",
      },
      query: { state },
      layout: {
        snapshot: {
          state: {
            columnOrder: [],
            columnWidths: {},
            hiddenFieldKeys: [],
            frozenThroughFieldKey: null,
          },
        },
      },
      entities: { index: {} },
      collaborationProjection: {
        getReadAuthorization: () => authority,
        subscribe: (listener) => {
          listeners.add(listener);
          return () => {
            listeners.delete(listener);
          };
        },
      },
    },
    browser,
    rows: [...rows, createDraftRow(1)],
    registry: createTimelineEditorDraftRegistry(),
    accessLost: false,
    stale: false,
    interruptViewportContinuity: vi.fn(),
    queueScalarSave: vi.fn(),
    queueCollectionSave: vi.fn(),
  };
  const presentation: GridPresentationSnapshot = {
    surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
    rowIdentities: ["two", "pin", "one"].map((recordId) => ({
      kind: "core_record",
      recordId,
    })),
    fieldKeys: ["timeline.activity_synopsis_text"],
    revision: 1,
  };
  const navigateToCell = vi.fn<NonNullable<GridHandle["navigateToCell"]>>(
    async (_anchor, options) => {
      options?.beforeFocus?.();
      return "focused";
    },
  );
  const grid = {
    presentation: {
      getSnapshot: () => presentation,
      subscribe: () => () => {},
    },
    navigateToCell,
    getActiveCell: () => ({
      surface: presentation.surface,
      rowIdentity: { kind: "core_record", recordId: "one" },
      fieldKey: "timeline.activity_synopsis_text",
    }),
    requestFocus: vi.fn(async () => "focused"),
  } as unknown as GridHandle;
  const hook = renderHook((props) => useTimelineFindSource(props), {
    initialProps: input,
  });
  act(() => hook.result.current.bindGrid(grid));
  const open = async () => {
    act(() => {
      hook.result.current.control.open();
      hook.result.current.control.changeTerm("needle");
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
  };
  return {
    ...hook,
    input,
    rows,
    query,
    open,
    navigateToCell,
    grid,
    recover: () => {
      authority = true;
      for (const listener of listeners) listener();
    },
    retire: () => {
      authority = false;
      for (const listener of listeners) listener();
    },
  };
}
describe("Timeline Find integration", () => {
  it("intersects accepted membership with presented order without querying or searching creation pins and local drafts", async () => {
    vi.useFakeTimers();
    const fixtureValue = await fixture();
    const { result, query, open, input } = fixtureValue;
    await open();
    expect(
      result.current.control.snapshot.matches.map((match) => match.rowIdentity),
    ).toEqual([
      { kind: "core_record", recordId: "two" },
      { kind: "core_record", recordId: "one" },
    ]);
    expect(query).toHaveBeenCalledTimes(1);
    expect(input.queueScalarSave).not.toHaveBeenCalled();
    expect(input.queueCollectionSave).not.toHaveBeenCalled();
    fixtureValue.rerender({ ...input, stale: true });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(result.current.control.snapshot.stale).toBe(true);
    expect(result.current.control.snapshot.matches).toHaveLength(2);
    act(fixtureValue.retire);
    expect(result.current.control.snapshot.term).toBe("");
    expect(result.current.control.snapshot.matches).toHaveLength(0);
    await act(async () => {
      await result.current.control.close();
    });
    expect(fixtureValue.grid.requestFocus).toHaveBeenLastCalledWith(
      { kind: "root" },
      { signal: expect.any(AbortSignal) },
    );
    act(fixtureValue.recover);
    expect(result.current.control.available).toBe(true);
    expect(result.current.control.snapshot.term).toBe("");
  });
  it("borrows ordinary inspector authoring without submitting it during Find navigation", async () => {
    vi.useFakeTimers();
    const f = await fixture();
    const editor = document.createElement("textarea");
    editor.dataset.inspectorEditorField = "timeline.activity_synopsis_text";
    document.body.append(editor);
    editor.value = "exact unfinished draft";
    editor.focus();
    await f.open();
    await act(async () => {
      await f.result.current.control.navigate(1);
    });
    expect(f.input.queueScalarSave).not.toHaveBeenCalled();
    expect(f.input.queueCollectionSave).not.toHaveBeenCalled();
    expect(f.navigateToCell).toHaveBeenCalledOnce();
    expect(editor.value).toBe("exact unfinished draft");
    await act(async () => {
      await f.result.current.control.close();
    });
    expect(f.input.queueScalarSave).not.toHaveBeenCalled();
    editor.remove();
  });
  it("retains rejected collection drafts and never departs recordless authoring through creation", async () => {
    vi.useFakeTimers();
    const f = await fixture();
    const editor = document.createElement("input");
    document.body.append(editor);
    const identity = {
      rowKey: "one",
      field: "tags" as const,
      surface: "grid" as const,
    };
    f.input.registry.registerInput(identity, editor);
    f.input.registry.setDraft(identity, "bad tag");
    editor.focus();
    vi.mocked(f.input.queueCollectionSave).mockImplementation(
      (_row, _field, _draft, _value, _surface, callback) =>
        callback?.({ kind: "validation_error", message: "invalid" }),
    );
    await f.open();
    await act(async () => {
      await f.result.current.control.navigate(1);
    });
    expect(f.navigateToCell).not.toHaveBeenCalled();
    expect(f.input.registry.draftValue(identity)).toBe("bad tag");
    expect(document.activeElement).toBe(editor);
    await act(async () => {
      await f.result.current.control.close();
    });
    f.input.registry.registerInput(identity, null);
    const draft = f.input.rows.find((row) => row.recordId === null);
    if (!draft) throw new Error("Missing fixture draft");
    f.input.registry.registerInput({ ...identity, rowKey: draft.key }, editor);
    editor.focus();
    await f.open();
    await act(async () => {
      await f.result.current.control.navigate(1);
    });
    expect(f.input.queueCollectionSave).toHaveBeenCalledTimes(1);
    expect(f.input.queueScalarSave).not.toHaveBeenCalled();
    expect(f.navigateToCell).toHaveBeenCalledTimes(1);
    editor.remove();
  });
  it("cancels a borrowed collection destination on keyboard focus departure while its write settles", async () => {
    vi.useFakeTimers();
    for (const { key, outcome } of [
      { key: "Tab", outcome: { kind: "accepted" } },
      {
        key: "Shift+Tab",
        outcome: { kind: "validation_error", message: "invalid" },
      },
    ] as const) {
      const f = await fixture();
      const host = document.createElement("div");
      const findInput = document.createElement("textarea");
      const outside = document.createElement("button");
      const editor = document.createElement("input");
      host.append(findInput);
      document.body.append(host, outside, editor);
      f.result.current.control.hostRef.current = host;
      f.result.current.control.inputRef.current = findInput;
      const identity = {
        rowKey: "one",
        field: "tags" as const,
        surface: "grid" as const,
      };
      f.input.registry.registerInput(identity, editor);
      f.input.registry.setDraft(identity, "pending tag");
      editor.focus();
      let settle: ((outcome: GridEditCommitOutcome) => void) | undefined;
      vi.mocked(f.input.queueCollectionSave).mockImplementation(
        (_row, _field, _draft, _value, _surface, callback) => {
          settle = callback;
        },
      );
      await f.open();
      let navigation: Promise<unknown> | undefined;
      act(() => {
        navigation = f.result.current.control.navigate(1);
      });
      await act(async () => {
        await Promise.resolve();
      });
      expect(f.input.queueCollectionSave).toHaveBeenCalledOnce();
      expect(f.result.current.control.snapshot.navigating).toBe(true);
      act(() => {
        findInput.dispatchEvent(
          new KeyboardEvent("keydown", {
            key: "Tab",
            shiftKey: key === "Shift+Tab",
            bubbles: true,
          }),
        );
        outside.focus();
      });
      expect(document.activeElement).toBe(outside);
      expect(f.result.current.control.snapshot.navigating).toBe(false);
      await act(async () => {
        settle?.(outcome);
        await navigation;
      });
      expect(f.navigateToCell).not.toHaveBeenCalled();
      expect(document.activeElement).toBe(outside);
      expect(f.input.queueCollectionSave).toHaveBeenCalledOnce();
      if (outcome.kind !== "accepted")
        expect(f.input.registry.draftValue(identity)).toBe("pending tag");
      f.input.registry.registerInput(identity, null);
      f.unmount();
      host.remove();
      outside.remove();
      editor.remove();
    }
  });
});
