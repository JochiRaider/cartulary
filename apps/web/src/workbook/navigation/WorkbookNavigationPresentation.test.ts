import type { GridHandle } from "@cartulary/grid-adapter";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import type { WorkbookQueryAcceptance } from "../query/WorkbookQueryBrowser";
import { WorkbookNavigationPresentation } from "./WorkbookNavigationPresentation";

async function settle() {
  for (let i = 0; i < 12; i++) await Promise.resolve();
}
function fixture() {
  const listeners = new Set<() => void>();
  const emit = () => {
    for (const listener of listeners) listener();
  };
  const focus = vi.fn<GridHandle["requestFocus"]>(async () => "focused");
  const grid = {
    requestFocus: focus,
    revealRecord: vi.fn(),
    presentation: {
      getSnapshot: () => ({
        surface: { kind: "view_schema", viewSchemaId: "view" },
        rowIdentities: [{ kind: "core_record", recordId: "record" }],
        fieldKeys: ["first", "second"],
        revision: 1,
      }),
      subscribe: () => () => {},
    },
  } as unknown as GridHandle;
  let committed = () => {};
  const select = vi.fn((_id: string, callback: () => void) => {
    committed = callback;
  });
  const state = {
    grid: grid as GridHandle | null,
    selection: { token: Symbol(), select },
    mount: { token: Symbol(), ready: true } as
      | { token: symbol; ready: boolean }
      | undefined,
    acceptance: "pending" as ReturnType<WorkbookQueryAcceptance["state"]>,
    recordPresent: true,
    admitted: true,
    identity: true,
  };
  const release = vi.fn();
  const acceptance: WorkbookQueryAcceptance = {
    state: () => state.acceptance,
    hasRecord: () => state.recordPresent,
    release,
  };
  const owner = new WorkbookNavigationPresentation({
    subscribePresentation: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    grid: () => state.grid,
    selectionBinding: () => state.selection,
    presentationState: () => state.mount,
  });
  const disconnect = owner.connect();
  const abort = new AbortController();
  const done = vi.fn();
  const start = (inspect = false) =>
    owner.start({
      view: "view",
      recordId: "record",
      fieldKey: "second",
      mode: "record",
      inspect,
      acceptance,
      signal: abort.signal,
      admitted: () => state.admitted,
      committed: () => state.identity,
      entryReady: () => true,
      settled: done,
    });
  return {
    state,
    owner,
    emit,
    focus,
    select,
    release,
    disconnect,
    abort,
    done,
    start,
    commit: () => committed(),
    currentCommit: () => committed,
  };
}

describe("navigation presentation", () => {
  it("waits for acceptance committed identity and selection before focusing the matching inspector", async () => {
    const f = fixture();
    const inspectorFocus = vi.fn(() => true);
    const open = vi.fn(() => {
      if (f.state.mount) f.state.mount.ready = false;
      f.emit();
    });
    f.owner.registerInspector("view", open);
    f.owner.registerInspectorFocus("view", "wrong-record", inspectorFocus);
    f.start(true);
    await settle();
    expect(f.select).not.toHaveBeenCalled();
    f.state.acceptance = "accepted";
    f.state.identity = false;
    f.emit();
    await settle();
    expect(f.select).not.toHaveBeenCalled();
    f.state.identity = true;
    f.emit();
    await settle();
    expect(f.select).toHaveBeenCalledWith("record", expect.any(Function));
    expect(f.focus).not.toHaveBeenCalled();
    f.commit();
    await settle();
    expect(f.focus).toHaveBeenCalledWith(
      expect.objectContaining({
        kind: "cell",
        anchor: expect.objectContaining({ fieldKey: "second" }),
      }),
      expect.anything(),
    );
    expect(open).toHaveBeenCalledOnce();
    expect(inspectorFocus).not.toHaveBeenCalled();
    expect(f.done).not.toHaveBeenCalled();
    f.owner.registerInspectorFocus("view", "record", inspectorFocus);
    await settle();
    expect(f.done).toHaveBeenCalledExactlyOnceWith("succeeded");
    expect(f.release).toHaveBeenCalledOnce();
    expect(f.owner.isCurrent()).toBe(true);
    f.emit();
    await settle();
    expect(inspectorFocus).toHaveBeenCalledOnce();
    f.disconnect();
  });

  it("fences replaced selection and grid bindings and never acknowledges their late focus", async () => {
    const f = fixture();
    f.state.acceptance = "accepted";
    f.start();
    await settle();
    const oldCommit = f.currentCommit();
    const oldFocus = deferred<"focused">();
    f.focus.mockReturnValueOnce(oldFocus.promise);
    f.commit();
    await settle();
    const oldSignal = f.focus.mock.calls[0]?.[1]?.signal;
    const newFocus = vi.fn<GridHandle["requestFocus"]>(async () => "focused");
    f.state.grid = { ...f.state.grid, requestFocus: newFocus } as GridHandle;
    f.state.selection = { token: Symbol(), select: f.select };
    f.emit();
    await settle();
    expect(oldSignal?.aborted).toBe(true);
    oldCommit();
    oldFocus.resolve("focused");
    await settle();
    expect(newFocus).not.toHaveBeenCalled();
    expect(f.done).not.toHaveBeenCalled();
    f.commit();
    await settle();
    expect(newFocus).toHaveBeenCalledOnce();
    expect(f.done).toHaveBeenCalledExactlyOnceWith("succeeded");
    f.disconnect();
  });

  it("settles failure cancellation and unmount once without late attachment", async () => {
    for (const reason of [
      "failed",
      "cancelled",
      "missing",
      "focus",
      "exhausted",
      "abort",
      "unmount",
    ] as const) {
      const f = fixture();
      f.state.acceptance =
        reason === "failed" || reason === "cancelled" ? reason : "accepted";
      f.state.recordPresent = reason !== "missing";
      if (reason === "focus")
        f.focus.mockRejectedValue(new Error("Focus unavailable"));
      if (reason === "exhausted") f.focus.mockResolvedValue("unavailable");
      f.start();
      await settle();
      if (reason === "abort") f.abort.abort();
      if (reason === "unmount") {
        f.state.mount = undefined;
        f.emit();
      }
      f.commit();
      await settle();
      expect(f.done).toHaveBeenCalledExactlyOnceWith(
        ["cancelled", "abort", "unmount"].includes(reason)
          ? "cancelled"
          : "failed",
      );
      f.emit();
      f.commit();
      await settle();
      expect(f.done).toHaveBeenCalledOnce();
      expect(f.release).toHaveBeenCalledOnce();
      if (reason === "exhausted") expect(f.focus).toHaveBeenCalledTimes(3);
      f.disconnect();
    }
  });

  it("waits for delayed mounts and falls back through eligible fields to the grid root", async () => {
    const f = fixture();
    f.state.acceptance = "accepted";
    f.state.mount = undefined;
    f.start();
    await settle();
    expect(f.focus).not.toHaveBeenCalled();
    f.state.mount = { token: Symbol(), ready: true };
    f.focus
      .mockResolvedValueOnce("unavailable")
      .mockResolvedValueOnce("unavailable");
    f.emit();
    await settle();
    f.commit();
    await settle();
    expect(f.focus.mock.calls.map(([target]) => target.kind)).toEqual([
      "cell",
      "cell",
      "root",
    ]);
    expect(f.done).toHaveBeenCalledExactlyOnceWith("succeeded");
    f.disconnect();
  });

  it("ignores stale inspector cleanup and reconnects after strict lifecycle cleanup", async () => {
    const f = fixture();
    f.disconnect();
    const disconnect = f.owner.connect();
    f.state.acceptance = "accepted";
    const old = f.owner.registerInspector("view", vi.fn());
    const open = vi.fn();
    f.owner.registerInspector("view", open);
    old();
    f.start(true);
    await settle();
    f.commit();
    await settle();
    expect(open).toHaveBeenCalledOnce();
    const focus = vi.fn(() => true);
    const oldFocus = f.owner.registerInspectorFocus(
      "view",
      "record",
      vi.fn(() => false),
    );
    f.owner.registerInspectorFocus("view", "record", focus);
    oldFocus();
    await settle();
    expect(focus).toHaveBeenCalledOnce();
    expect(f.done).toHaveBeenCalledExactlyOnceWith("succeeded");
    disconnect();
  });
});
