import type {
  GridCellNavigationOptions,
  GridHandle,
  GridPresentationSnapshot,
} from "@cartulary/grid-adapter";
import {
  hostsViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { WorkbookQueryBrowser } from "../query/WorkbookQueryBrowser";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { useWorkbookFind, type WorkbookFindInput } from "./useWorkbookFind";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});
async function fixture() {
  const contract = requireViewContract(hostsViewSchemaId);
  const state = emptyWorkbookQueryState();
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
    kind: "accepted",
    value: {
      incidentId: "incident",
      viewSchemaId: hostsViewSchemaId,
      rows: ["one", "two", "collapsed"].map((id) =>
        fullWorkbookViewRow(contract, id, 1, {}),
      ),
      canonicalQuery: { filters: [], sort: [] },
      paging: { limit: 100, hasMore: false, nextCursor: null },
      producingRequest: { queryState: state, limit: 100 },
    },
  }));
  const browser = new WorkbookQueryBrowser({ query }, hostsViewSchemaId);
  const response = await browser.query({
    contract,
    queryState: state,
    signal: new AbortController().signal,
  });
  if (response.kind !== "accepted") throw new Error("fixture");
  browser.accept(response.value);
  let presentation: GridPresentationSnapshot = {
    surface: { kind: "view_schema", viewSchemaId: hostsViewSchemaId },
    rowIdentities: ["two", "creation-pin", "one"].map((recordId) => ({
      kind: "core_record",
      recordId,
    })),
    fieldKeys: ["host.display_name", "host.aliases"],
    revision: 1,
  };
  const gridListeners = new Set<() => void>();
  const authListeners = new Set<() => void>();
  let authority = true;
  const restore = vi.fn(() => true);
  const input: WorkbookFindInput = {
    lifetimeKey: "incident:hosts",
    configurationKey: "view1",
    browser,
    authorization: {
      getReadAuthorization: () => authority,
      subscribe: (listener) => {
        authListeners.add(listener);
        return () => {
          authListeners.delete(listener);
        };
      },
    },
    readable: true,
    stale: false,
    readText: () => ["needle"],
    captureFocus: () => ({ restore }),
  };
  const hook = renderHook((props) => useWorkbookFind(props), {
    initialProps: input,
  });
  const navigateToCell = vi.fn<NonNullable<GridHandle["navigateToCell"]>>(
    async (_anchor, options) => {
      // React replaces imperative handles during a still-current grid commit.
      hook.result.current.bindGrid(null);
      expect(options?.isCurrent?.()).toBe(true);
      hook.result.current.bindGrid(handle);
      options?.beforeFocus?.();
      expect(hook.result.current.isApplyingFocus()).toBe(true);
      return "focused";
    },
  );
  const handle = {
    presentation: {
      getSnapshot: () => presentation,
      subscribe: (listener: () => void) => {
        gridListeners.add(listener);
        return () => {
          gridListeners.delete(listener);
        };
      },
    },
    navigateToCell,
    requestFocus: vi.fn(async () => "focused"),
    getActiveCell: () => null,
  } as unknown as GridHandle;
  act(() => hook.result.current.bindGrid(handle));
  const flush = async () => {
    await act(async () => {
      await vi.runAllTimersAsync();
    });
  };
  const open = async () => {
    act(() => {
      hook.result.current.control.open();
      hook.result.current.control.changeTerm("needle");
    });
    await flush();
  };
  return {
    ...hook,
    input,
    browser,
    query,
    handle,
    navigateToCell,
    restore,
    open,
    flush,
    changePresentation: () => {
      presentation = {
        ...presentation,
        revision: presentation.revision + 1,
        rowIdentities: presentation.rowIdentities.slice(1),
      };
      for (const listener of gridListeners) listener();
    },
    setAuthority: (value: boolean) => {
      authority = value;
      for (const listener of authListeners) listener();
    },
  };
}
describe("Workbook Find binding", () => {
  it("intersects accepted and presented membership without reads and keeps focus loans independent", async () => {
    vi.useFakeTimers();
    const f = await fixture();
    await f.open();
    expect(
      f.result.current.control.snapshot.matches.map((a) => [
        a.rowIdentity,
        a.fieldKey,
      ]),
    ).toEqual(
      ["two", "one"].flatMap((recordId) =>
        ["host.display_name", "host.aliases"].map((field) => [
          { kind: "core_record", recordId },
          field,
        ]),
      ),
    );
    expect(f.query).toHaveBeenCalledTimes(1);
    expect(f.navigateToCell).not.toHaveBeenCalled();
    await act(async () => {
      await f.result.current.control.close();
    });
    expect(f.restore).toHaveBeenCalledOnce();
    await f.open();
    await act(async () => {
      await f.result.current.control.navigate(1);
    });
    expect(f.result.current.isApplyingFocus()).toBe(false);
    expect(f.result.current.control.snapshot.current?.rowIdentity).toEqual({
      kind: "core_record",
      recordId: "two",
    });
  });
  it("retires on schema or authority loss but retains intent on same-schema configuration replacement", async () => {
    vi.useFakeTimers();
    const f = await fixture();
    await f.open();
    f.rerender({ ...f.input, configurationKey: "saved-view2", stale: true });
    await f.flush();
    expect(f.result.current.control.snapshot.term).toBe("needle");
    expect(f.result.current.control.snapshot.stale).toBe(true);
    f.rerender({ ...f.input, lifetimeKey: "incident:identities" });
    expect(f.result.current.control.snapshot.term).toBe("");
    await f.open();
    act(() => f.setAuthority(false));
    expect(f.result.current.control.snapshot.matches).toEqual([]);
    expect(f.result.current.control.snapshot.term).toBe("");
    act(() => f.setAuthority(true));
    expect(f.result.current.control.snapshot.term).toBe("");
  });
  it("fences pending destinations synchronously at owner notifications without cancelling settlement", async () => {
    vi.useFakeTimers();
    const f = await fixture();
    await f.open();
    let options: GridCellNavigationOptions | undefined;
    let accept: ((result: "focused") => void) | undefined;
    f.navigateToCell.mockImplementation(async (_anchor, supplied) => {
      options = supplied;
      return new Promise((resolve) => {
        accept = resolve;
      });
    });
    let navigation: Promise<unknown> | undefined;
    act(() => {
      navigation = f.result.current.control.navigate(1);
    });
    await act(async () => {
      await Promise.resolve();
    });
    expect(options?.isCurrent?.()).toBe(true);
    act(() => {
      f.changePresentation();
      expect(options?.signal?.aborted).toBe(true);
      expect(options?.isCurrent?.()).toBe(false);
    });
    await act(async () => {
      accept?.("focused");
      await navigation;
    });
    expect(f.result.current.control.snapshot.current).toBeNull();
  });
});
