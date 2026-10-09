import type { GridHandle } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { savedViewTestResource } from "../../testing/workbookSavedViewTestSupport";
import {
  buildSavedViewLayoutJson,
  emptyWorkbookQueryState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "../models/workbookQuery";
import {
  notesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  useWorkbookBrowsingRegistry,
  WorkbookQueryBrowsingProvider,
} from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookRecordLocatorPort } from "../query/WorkbookRecordLocatorPort";
import type {
  WorkbookViewQueryAccepted,
  WorkbookViewQueryPort,
} from "../query/WorkbookViewQueryPort";
import { useWorkbookWorkbench } from "./useWorkbookWorkbench";
import type { WorkbookNavigationHost } from "./WorkbookNavigationHost";
import {
  type WorkbookReturnOrigin,
  WorkbookSessionNavigation,
} from "./WorkbookSessionNavigation";

afterEach(cleanup);
const contract = requireViewContract(notesViewSchemaId);
const recordId = "00000000-0000-4000-8000-000000000001";
function page(
  queryState = emptyWorkbookQueryState(),
): WorkbookViewQueryAccepted {
  return {
    incidentId: "incident",
    viewSchemaId: notesViewSchemaId,
    rows: [
      fullWorkbookViewRow(contract, recordId, 1, {
        "note.title": "Saved note",
      }),
    ],
    canonicalQuery: { sort: queryState.sort, filters: queryState.filters },
    paging: { limit: 100, hasMore: false, nextCursor: null },
    producingRequest: { queryState, limit: 100, cursorToken: "ordinary-start" },
  };
}
function fixture() {
  const session = new WorkbookSessionNavigation("incident");
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
    kind: "accepted",
    value: page(),
  }));
  const locate = vi.fn<WorkbookRecordLocatorPort["locate"]>(async () => ({
    kind: "accepted",
    value: { outcome: "located", page: page() },
  }));
  const applyIdentity = vi.fn();
  const currentQuery: WorkbookQueryState = {
    ...emptyWorkbookQueryState(),
    filters: [
      {
        fieldKey: "note.created_by_user_id",
        op: "eq",
        arg: { value: recordId },
      },
    ],
  };
  // The hook consumes only these owner ports; no source operation is available.
  const host: WorkbookNavigationHost = {
    snapshot: {
      gridEntryFocusRequest: { kind: "idle" },
      surface: timelineViewSchemaId,
      startupSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      activeContract: requireViewContract(timelineViewSchemaId),
      selectedSavedView: null,
    },
    commands: {
      currentQueryStateForSurface: () => currentQuery,
      currentLayoutStateForSurface: () => ({}),
      cancelGridEntryFocus: vi.fn(),
      selectWorkbookSurface: vi.fn(),
      selectExtensionWorkspace: vi.fn(),
      applyQueryStateForSurface: vi.fn(),
      applyLayoutStateForSurface: vi.fn(),
      applyWorkbookIdentity: applyIdentity,
    },
    savedViews: {
      retainNavigation: vi.fn(),
      read: vi.fn(async () => null),
      isUnavailable: vi.fn(() => false),
      acceptResource: vi.fn(),
    },
  };
  const admitsPage = vi.fn(() => true);
  const options = {
    session,
    incidentId: "incident",
    actorId: "actor",
    readable: true,
    host,
    query: { query },
    locator: { locate },
    extensionAvailable: () => false,
    authorityFailure: vi.fn(),
    admitsPage,
  };
  let registry!: ReturnType<typeof useWorkbookBrowsingRegistry>;
  const hook = renderHook(
    (props) => {
      registry = useWorkbookBrowsingRegistry();
      return useWorkbookWorkbench(props);
    },
    {
      initialProps: options,
      wrapper: WorkbookQueryBrowsingProvider,
    },
  );
  const detachPresentation = vi.fn();
  registry.updatePresentation(
    timelineViewSchemaId,
    Symbol(),
    true,
    detachPresentation,
  );
  const target = {
    sheetRef: { kind: "view_schema" as const, id: notesViewSchemaId },
    recordId,
  };
  return {
    session,
    query,
    locate,
    applyIdentity,
    admitsPage,
    options,
    hook,
    target,
    detachPresentation,
    registry,
  };
}

describe("workbench navigation admission", () => {
  it("restores captured Return configuration against a freshly authorized changed saved view", async () => {
    const f = fixture();
    const capturedQuery: WorkbookQueryState = {
      ...emptyWorkbookQueryState(),
      sort: [{ fieldKey: "note.title", direction: "desc" }],
    };
    const capturedLayout = buildSavedViewLayoutJson(contract, {
      hiddenFieldKeys: ["note.body"],
    });
    const resource = savedViewTestResource({
      incident_id: "incident",
      view_schema_id: notesViewSchemaId,
      saved_view_version: 2,
      query_json: { sort: [], filters: [] },
      layout_json: buildSavedViewLayoutJson(contract),
    });
    const origin: WorkbookReturnOrigin = {
      incidentId: "incident",
      sheetRef: { kind: "saved_view", id: resource.saved_view_id },
      viewSchemaId: notesViewSchemaId,
      query: capturedQuery,
      layout: capturedLayout,
      savedViewVersion: 1,
      invoker: "view",
    };
    await act(async () => {
      await f.session.navigate("leave", origin, async () => "changed");
    });
    vi.mocked(f.options.host.savedViews.read).mockResolvedValue(resource);
    f.query.mockResolvedValue({ kind: "accepted", value: page(capturedQuery) });
    act(() => f.hook.result.current.returnToOrigin());
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledOnce());
    expect(f.options.host.savedViews.read).toHaveBeenCalledWith(
      resource.saved_view_id,
    );
    expect(f.options.host.savedViews.acceptResource).toHaveBeenCalledWith(
      resource,
    );
    expect(
      f.options.host.commands.applyQueryStateForSurface,
    ).toHaveBeenCalledWith(notesViewSchemaId, capturedQuery);
    expect(
      f.options.host.commands.applyLayoutStateForSurface,
    ).toHaveBeenCalledWith(
      notesViewSchemaId,
      workbookLayoutStateFromSavedViewLayoutJson(contract, capturedLayout),
    );
    expect(f.hook.result.current.message).toContain("saved view changed");
    expect(f.session.getSnapshot().trail).toEqual([]);
  });
  it("retains Return on failed or cancelled reads and uses an explicit base fallback", async () => {
    for (const outcome of ["failure", "cancel", "unavailable"] as const) {
      const f = fixture();
      const resource = savedViewTestResource({
        incident_id: "incident",
        view_schema_id: notesViewSchemaId,
        layout_json: buildSavedViewLayoutJson(contract),
      });
      const origin: WorkbookReturnOrigin = {
        incidentId: "incident",
        sheetRef: { kind: "saved_view", id: resource.saved_view_id },
        viewSchemaId: notesViewSchemaId,
        query: emptyWorkbookQueryState(),
        layout: buildSavedViewLayoutJson(contract),
        invoker: "view",
      };
      await act(async () => {
        await f.session.navigate("leave", origin, async () => "changed");
      });
      const pending = deferred<typeof resource | null>();
      vi.mocked(f.options.host.savedViews.read).mockImplementation(
        () => pending.promise,
      );
      if (outcome === "failure")
        f.query.mockResolvedValueOnce({
          kind: "rejected",
          failure: {
            kind: "invalid_contract",
            message: "Cannot verify restored page",
          },
        });
      act(() => f.hook.result.current.returnToOrigin());
      await waitFor(() =>
        expect(f.options.host.savedViews.read).toHaveBeenCalledOnce(),
      );
      expect(f.session.getSnapshot().trail).toEqual([origin]);
      expect(f.applyIdentity).not.toHaveBeenCalled();
      if (outcome === "cancel")
        act(() => f.hook.result.current.cancelNavigation());
      await act(async () =>
        pending.resolve(outcome === "unavailable" ? null : resource),
      );
      expect(f.session.getSnapshot().trail).toEqual([origin]);
      expect(f.applyIdentity).not.toHaveBeenCalled();
      expect(f.detachPresentation).not.toHaveBeenCalled();
      if (outcome === "unavailable") {
        expect(f.hook.result.current.openBase).not.toBeNull();
        act(() => f.hook.result.current.openBase?.());
        await waitFor(() =>
          expect(f.applyIdentity).toHaveBeenCalledWith(
            {
              sheetRef: { kind: "view_schema", id: notesViewSchemaId },
              viewSchemaId: notesViewSchemaId,
            },
            { focusFirstGridTarget: false },
          ),
        );
        expect(f.query).toHaveBeenCalledWith(
          expect.objectContaining({ queryState: emptyWorkbookQueryState() }),
        );
        expect(f.session.getSnapshot().trail).toEqual([]);
      } else
        expect(f.session.getSnapshot().outcome).toBe(
          outcome === "cancel" ? "cancelled" : "failed",
        );
      f.hook.unmount();
    }
  });
  it("retains origin while location is pending and rejects late cancelled attachment", async () => {
    for (const invalidation of [
      "interaction",
      "account",
      "authority",
      "unmount",
    ]) {
      const f = fixture();
      const pending =
        deferred<Awaited<ReturnType<WorkbookRecordLocatorPort["locate"]>>>();
      f.locate.mockReturnValueOnce(pending.promise);
      act(() => f.hook.result.current.open(f.target));
      await waitFor(() => expect(f.locate).toHaveBeenCalledTimes(1));
      expect(f.applyIdentity).not.toHaveBeenCalled();
      expect(f.detachPresentation).not.toHaveBeenCalled();
      expect(f.session.getSnapshot().trail).toEqual([]);
      act(() => {
        if (invalidation === "interaction")
          f.hook.result.current.cancelNavigation();
        else if (invalidation === "unmount") f.hook.unmount();
        else
          f.hook.rerender({
            ...f.options,
            ...(invalidation === "account"
              ? { actorId: "other" }
              : { readable: false }),
          });
      });
      await act(async () =>
        pending.resolve({
          kind: "accepted",
          value: { outcome: "located", page: page() },
        }),
      );
      expect(f.applyIdentity).not.toHaveBeenCalled();
      expect(f.session.getSnapshot().trail).toEqual([]);
      expect(f.locate.mock.calls[0]?.[0].signal.aborted).toBe(true);
      expect(f.detachPresentation).not.toHaveBeenCalled();
      if (invalidation !== "unmount") f.hook.unmount();
    }
  });
  it("keeps ordinary surface entry with its existing read and recovery owner", async () => {
    const f = fixture();
    act(() => f.hook.result.current.open({ sheetRef: f.target.sheetRef }));
    await waitFor(() =>
      expect(
        f.options.host.commands.selectWorkbookSurface,
      ).toHaveBeenCalledWith(notesViewSchemaId, { focusFirstGridTarget: true }),
    );
    expect(f.query).not.toHaveBeenCalled();
    expect(f.locate).not.toHaveBeenCalled();
    expect(f.applyIdentity).not.toHaveBeenCalled();
    await waitFor(() => expect(f.session.getSnapshot().trail).toHaveLength(1));
  });
  it("keeps an outside-query origin until an explicit base locator succeeds", async () => {
    const f = fixture();
    f.locate.mockResolvedValueOnce({
      kind: "accepted",
      value: { outcome: "outside_query" },
    });
    act(() => f.hook.result.current.open(f.target));
    await waitFor(() => expect(f.hook.result.current.openBase).not.toBeNull());
    expect(f.applyIdentity).not.toHaveBeenCalled();
    expect(f.session.getSnapshot().trail).toEqual([]);
    expect(f.detachPresentation).not.toHaveBeenCalled();
    act(() => f.hook.result.current.openBase?.());
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledTimes(1));
    expect(f.detachPresentation).toHaveBeenCalledOnce();
    expect(f.locate.mock.calls[1]?.[0].recordId).toBe(recordId);
    expect(f.locate.mock.calls[1]?.[0].queryState).toEqual(
      emptyWorkbookQueryState(),
    );
    expect(f.query).not.toHaveBeenCalled();
    await waitFor(() => expect(f.session.getSnapshot().trail).toHaveLength(1));
  });
  it("preserves the origin when the locator is unsupported without querying pages", async () => {
    const f = fixture();
    f.locate.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Location is unsupported." },
    });
    act(() => f.hook.result.current.open(f.target, true));
    await waitFor(() => expect(f.hook.result.current.retry).not.toBeNull());
    expect(f.locate).toHaveBeenCalledOnce();
    expect(f.query).not.toHaveBeenCalled();
    expect(f.applyIdentity).not.toHaveBeenCalled();
    expect(f.detachPresentation).not.toHaveBeenCalled();
    expect(f.session.getSnapshot().trail).toEqual([]);
  });
  it("opens record pins with defaults and preserves pins after operational failure", async () => {
    const f = fixture();
    const pin = {
      incidentId: "incident",
      ...f.target,
      label: "Authorized label",
    };
    act(() => f.session.pin(pin));
    f.locate.mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "invalid_contract", message: "Cannot verify response" },
    });
    act(() => f.hook.result.current.openPin(pin));
    await waitFor(() => expect(f.hook.result.current.retry).not.toBeNull());
    expect(f.locate.mock.calls[0]?.[0].queryState).toEqual(
      emptyWorkbookQueryState(),
    );
    expect(f.session.getSnapshot().pins[0]?.label).toBe("Authorized label");
    expect(f.applyIdentity).not.toHaveBeenCalled();
    expect(f.detachPresentation).not.toHaveBeenCalled();
    act(() => f.hook.result.current.retry?.());
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledTimes(1));
    f.detachPresentation.mockClear();
    f.locate.mockResolvedValueOnce({
      kind: "accepted",
      value: { outcome: "unavailable" },
    });
    act(() => f.hook.result.current.openPin(pin));
    await waitFor(() =>
      expect(f.hook.result.current.message).toBe("This record is unavailable."),
    );
    expect(f.detachPresentation).not.toHaveBeenCalled();
    expect(f.session.getSnapshot().pins[0]?.label).toBe("Unavailable item");
  });
  it("refuses observations older than owner evidence and conceals uncertain authority", async () => {
    const f = fixture();
    f.admitsPage.mockReturnValue(false);
    act(() => f.hook.result.current.open(f.target));
    await waitFor(() =>
      expect(f.hook.result.current.message).toContain(
        "changed during navigation",
      ),
    );
    expect(f.applyIdentity).not.toHaveBeenCalled();
    act(() =>
      f.session.pin({
        incidentId: "incident",
        ...f.target,
        label: "Saved label",
      }),
    );
    f.hook.rerender({ ...f.options, readable: false });
    expect(f.session.getSnapshot().pins).toEqual([]);
    f.hook.rerender(f.options);
    expect(f.session.getSnapshot().pins[0]?.label).toBe("Saved label");
  });
});

it("workbench navigation reconciles a superseded inspector request until committed attachment and fences cancellation", async () => {
  for (const cancelled of [false, true]) {
    const f = fixture();
    const binding = f.registry.prepare(f.query, notesViewSchemaId, true);
    f.registry.commit(binding);
    const grid = {
      requestFocus: vi.fn(async () => "focused"),
      presentation: {
        getSnapshot: () => ({
          surface: { kind: "view_schema", viewSchemaId: notesViewSchemaId },
          rowIdentities: [{ kind: "core_record", recordId }],
          fieldKeys: ["note.title"],
          revision: 1,
        }),
        subscribe: () => () => {},
      },
    } as unknown as GridHandle;
    f.registry.bindGrid(
      notesViewSchemaId,
      { current: grid },
      (_id, committed) => committed(),
    );
    f.registry.updatePresentation(notesViewSchemaId, Symbol(), true, () => {});
    const wrongSubjectFocus = vi.fn(() => true);
    act(() =>
      f.hook.result.current.registerInspectorFocus(
        notesViewSchemaId,
        "unrelated-record",
        wrongSubjectFocus,
      ),
    );
    const supersededOpen = vi.fn();
    act(() =>
      f.hook.result.current.registerInspector(
        notesViewSchemaId,
        supersededOpen,
      ),
    );
    act(() => f.hook.result.current.open(f.target, true));
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledOnce());
    // The staged page still needs explicit acceptance by its read owner.
    await act(async () => {
      const browser = binding.currentBrowser();
      if (!browser) throw new Error("Missing destination browser");
      const read = await browser.query({
        contract,
        queryState: emptyWorkbookQueryState(),
        signal: new AbortController().signal,
      });
      if (read.kind !== "accepted")
        throw new Error("Missing accepted destination");
      browser.accept(read.value);
    });
    await waitFor(() => expect(supersededOpen).toHaveBeenCalledOnce());
    expect(grid.requestFocus).toHaveBeenCalledOnce();
    expect(wrongSubjectFocus).not.toHaveBeenCalled();

    // A committed lifecycle replacement can discard the first open request.
    // Its registration is a new readiness event, not evidence of attachment.
    if (cancelled) act(() => f.hook.result.current.cancelNavigation());
    const replacementOpen = vi.fn();
    act(() =>
      f.hook.result.current.registerInspector(
        notesViewSchemaId,
        replacementOpen,
      ),
    );
    await act(async () => {});
    expect(replacementOpen).toHaveBeenCalledTimes(cancelled ? 0 : 1);
    const focus = vi.fn(() => true);
    act(() =>
      f.hook.result.current.registerInspectorFocus(
        notesViewSchemaId,
        recordId,
        focus,
      ),
    );
    await act(async () => {});
    expect(focus).toHaveBeenCalledTimes(cancelled ? 0 : 1);
    act(() => f.registry.resumeNavigation());
    await act(async () => {});
    expect(replacementOpen).toHaveBeenCalledTimes(cancelled ? 0 : 1);
    expect(focus).toHaveBeenCalledTimes(cancelled ? 0 : 1);
    expect(wrongSubjectFocus).not.toHaveBeenCalled();
    f.hook.unmount();
  }
});

it("navigation selects the accepted record through its owner before inspection, and ignores cancelled focus", async () => {
  for (const [fieldKeys, requested, expectedField] of [
    [["note.title", "note.body"], "note.body", "note.body"],
    [["note.title"], "note.body", "note.title"],
    [[], "note.body", undefined],
  ] as const) {
    const { result, unmount } = renderHook(
      () => useWorkbookBrowsingRegistry(),
      {
        wrapper: WorkbookQueryBrowsingProvider,
      },
    );
    const registry = result.current;
    const read = vi.fn<WorkbookViewQueryPort["query"]>();
    const binding = registry.prepare(read, notesViewSchemaId, true);
    registry.commit(binding);
    const browser = binding.currentBrowser();
    if (!browser) throw new Error("Missing browser");
    let commitSelection = () => {};
    const select = vi.fn((_id: string, committed: () => void) => {
      commitSelection = committed;
    });
    const inspect = vi.fn();
    const focus = deferred<"focused">();
    const grid = {
      revealRecord: vi.fn(() => true),
      requestFocus: vi.fn(() => focus.promise),
      presentation: {
        getSnapshot: () => ({
          surface: { kind: "view_schema", viewSchemaId: notesViewSchemaId },
          rowIdentities: [{ kind: "core_record", recordId }],
          fieldKeys,
          revision: 1,
        }),
        subscribe: () => () => {},
      },
    } as unknown as GridHandle;
    registry.bindGrid(notesViewSchemaId, { current: grid }, select);
    const token = Symbol();
    registry.updatePresentation(notesViewSchemaId, token, false, () => {});
    registry.stageNavigation(
      page(),
      read,
      { recordId, fieldKey: requested },
      inspect,
    );
    await act(async () => {});
    expect(select).not.toHaveBeenCalled();
    const staged = await browser.query({
      contract,
      queryState: emptyWorkbookQueryState(),
      signal: new AbortController().signal,
    });
    if (staged.kind !== "accepted") throw new Error("Expected staged read");
    browser.accept(staged.value);
    await act(async () => {});
    // Read acceptance cannot focus an inert presentation or prematurely select.
    expect(select).not.toHaveBeenCalled();
    act(() =>
      registry.updatePresentation(notesViewSchemaId, token, true, () => {}),
    );
    await act(async () => {});
    expect(select).toHaveBeenCalledWith(recordId, expect.any(Function));
    expect(grid.requestFocus).not.toHaveBeenCalled();
    await act(async () => commitSelection());
    expect(grid.requestFocus).toHaveBeenCalledWith(
      expectedField
        ? {
            kind: "cell",
            anchor: {
              surface: { kind: "view_schema", viewSchemaId: notesViewSchemaId },
              rowIdentity: { kind: "core_record", recordId },
              fieldKey: expectedField,
            },
          }
        : { kind: "root" },
      expect.anything(),
    );
    expect(inspect).not.toHaveBeenCalled();
    await act(async () => focus.resolve("focused"));
    expect(inspect).toHaveBeenCalledOnce();
    unmount();
  }
});

it("cancels destination attachment before delayed mounting, selection commit and inspector activation", async () => {
  for (const phase of [
    "mount",
    "selection",
    "focus",
    "inspector",
    "unmount",
  ] as const) {
    const { result, unmount } = renderHook(
      () => useWorkbookBrowsingRegistry(),
      { wrapper: WorkbookQueryBrowsingProvider },
    );
    const registry = result.current;
    const read = vi.fn<WorkbookViewQueryPort["query"]>();
    const binding = registry.prepare(read, notesViewSchemaId, true);
    registry.commit(binding);
    const browser = binding.currentBrowser();
    if (!browser) throw new Error("Missing browser");
    const focus = deferred<"focused">();
    let selectionCommitted = () => {};
    const select = vi.fn((_id: string, committed: () => void) => {
      selectionCommitted = committed;
    });
    const inspect = vi.fn(() => false);
    const grid = {
      requestFocus: vi.fn(() => focus.promise),
      presentation: {
        getSnapshot: () => ({
          surface: { kind: "view_schema", viewSchemaId: notesViewSchemaId },
          rowIdentities: [{ kind: "core_record", recordId }],
          fieldKeys: ["note.title"],
          revision: 1,
        }),
        subscribe: () => () => {},
      },
    } as unknown as GridHandle;
    const handle = { current: phase === "mount" ? null : grid };
    registry.bindGrid(notesViewSchemaId, handle, select);
    const token = Symbol();
    registry.updatePresentation(notesViewSchemaId, token, true, () => {});
    const intent = new AbortController();
    registry.stageNavigation(page(), read, { recordId }, inspect, {
      signal: intent.signal,
      navigationOnly: true,
      onUnavailable: vi.fn(),
    });
    const staged = await browser.query({
      contract,
      queryState: emptyWorkbookQueryState(),
      signal: intent.signal,
    });
    if (staged.kind !== "accepted") throw new Error("Expected staged read");
    browser.accept(staged.value);
    await act(async () => {});
    if (phase === "focus" || phase === "unmount" || phase === "inspector")
      await act(async () => selectionCommitted());
    if (phase === "inspector") await act(async () => focus.resolve("focused"));
    if (phase === "unmount") unmount();
    else intent.abort();
    handle.current = grid;
    registry.refreshGridBinding(notesViewSchemaId);
    await act(async () => {
      selectionCommitted();
      focus.resolve("focused");
    });
    expect(inspect).toHaveBeenCalledTimes(phase === "inspector" ? 1 : 0);
    expect(select).toHaveBeenCalledTimes(phase === "mount" ? 0 : 1);
    expect(grid.requestFocus).toHaveBeenCalledTimes(
      phase === "focus" || phase === "unmount" || phase === "inspector" ? 1 : 0,
    );
    if (phase !== "unmount") unmount();
  }
});
