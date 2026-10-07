import type { GridHandle } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
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
import { WorkbookSessionNavigation } from "./WorkbookSessionNavigation";

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
  const currentQuery = {
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
  const runtime = {
    snapshot: {
      surface: timelineViewSchemaId,
      startupSheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      activeContract: requireViewContract(timelineViewSchemaId),
      savedViewsResource: { selectedSavedView: null },
    },
    commands: {
      currentQueryStateForSurface: () => currentQuery,
      currentLayoutStateForSurface: () => ({}),
      cancelGridEntryFocus: vi.fn(),
      selectWorkbookSurface: vi.fn(),
      applyQueryStateForSurface: vi.fn(),
      applyLayoutStateForSurface: vi.fn(),
      applyWorkbookIdentity: applyIdentity,
    },
  } as unknown as Parameters<typeof useWorkbookWorkbench>[0]["runtime"];
  const admitsPage = vi.fn(() => true);
  const options = {
    session,
    incidentId: "incident",
    actorId: "actor",
    readable: true,
    runtime,
    query: { query },
    locator: { locate },
    extensionAvailable: () => false,
    authorityFailure: vi.fn(),
    admitsPage,
  };
  const hook = renderHook((props) => useWorkbookWorkbench(props), {
    initialProps: options,
    wrapper: WorkbookQueryBrowsingProvider,
  });
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
  };
}

describe("workbench navigation admission", () => {
  it("retains origin while location is pending and rejects late cancelled attachment", async () => {
    const f = fixture();
    const pending =
      deferred<Awaited<ReturnType<WorkbookRecordLocatorPort["locate"]>>>();
    f.locate.mockReturnValueOnce(pending.promise);
    act(() => f.hook.result.current.open(f.target));
    await waitFor(() => expect(f.locate).toHaveBeenCalledTimes(1));
    expect(f.applyIdentity).not.toHaveBeenCalled();
    expect(f.session.getSnapshot().trail).toEqual([]);
    act(() => f.hook.result.current.cancelNavigation());
    await act(async () =>
      pending.resolve({
        kind: "accepted",
        value: { outcome: "located", page: page() },
      }),
    );
    expect(f.applyIdentity).not.toHaveBeenCalled();
    expect(f.session.getSnapshot().trail).toEqual([]);
    expect(f.locate.mock.calls[0]?.[0].signal.aborted).toBe(true);
  });
  it("keeps ordinary surface entry with its existing read and recovery owner", async () => {
    const f = fixture();
    act(() => f.hook.result.current.open({ sheetRef: f.target.sheetRef }));
    await waitFor(() =>
      expect(
        f.options.runtime.commands.selectWorkbookSurface,
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
    act(() => f.hook.result.current.openBase?.());
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledTimes(1));
    expect(f.locate.mock.calls[1]?.[0].recordId).toBe(recordId);
    expect(f.locate.mock.calls[1]?.[0].queryState).toEqual(
      emptyWorkbookQueryState(),
    );
    expect(f.query).not.toHaveBeenCalled();
    await waitFor(() => expect(f.session.getSnapshot().trail).toHaveLength(1));
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
    act(() => f.hook.result.current.retry?.());
    await waitFor(() => expect(f.applyIdentity).toHaveBeenCalledTimes(1));
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

it("navigation selects the accepted record through its owner before inspection, and ignores cancelled focus", async () => {
  const { result } = renderHook(() => useWorkbookBrowsingRegistry(), {
    wrapper: WorkbookQueryBrowsingProvider,
  });
  const registry = result.current;
  const read = vi.fn<WorkbookViewQueryPort["query"]>();
  const binding = registry.prepare(read, notesViewSchemaId, true);
  const unbind = registry.commit(binding);
  const browser = binding.currentBrowser();
  if (!browser) throw new Error("Missing browser");
  const select = vi.fn();
  const inspect = vi.fn();
  let changed = () => {};
  const grid = {
    revealRecord: vi.fn(() => true),
    requestFocus: vi.fn(async () => "focused" as const),
    presentation: {
      getSnapshot: () => ({
        surface: { kind: "view_schema", viewSchemaId: notesViewSchemaId },
        rowIdentities: [{ kind: "core_record", recordId }],
        fieldKeys: ["note.title"],
        revision: 1,
      }),
      subscribe: (listener: () => void) => {
        changed = listener;
        return () => {};
      },
    },
  } as unknown as GridHandle;
  const detach = registry.bindGrid(
    notesViewSchemaId,
    { current: grid },
    select,
  );
  const destination = page();
  registry.stageNavigation(destination, read, { recordId }, inspect);
  await act(async () => {});
  expect(select).not.toHaveBeenCalled();
  expect(grid.requestFocus).not.toHaveBeenCalled();
  // Only the source owner accepting the staged read permits selection and focus.
  const staged = await browser.query({
    contract,
    queryState: emptyWorkbookQueryState(),
    signal: new AbortController().signal,
  });
  if (staged.kind !== "accepted") throw new Error("Expected staged read");
  browser.accept(staged.value);
  await act(async () => changed());
  expect(select).toHaveBeenCalledWith(recordId);
  expect(inspect).toHaveBeenCalledOnce();
  expect(select.mock.invocationCallOrder[0]).toBeLessThan(
    inspect.mock.invocationCallOrder[0] ?? 0,
  );
  registry.stageNavigation(page(), read, { recordId }, inspect);
  registry.cancelNavigationFocus();
  await act(async () => changed());
  expect(select).toHaveBeenCalledTimes(1);
  detach();
  unbind?.();
});
