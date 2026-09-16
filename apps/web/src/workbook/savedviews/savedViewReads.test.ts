import { describe, expect, it, vi } from "vitest";
import { observeAsyncOperation } from "../../services/asyncObservation";
import { deferred } from "../../testing/fetchMockTestSupport";
import { savedViewTestResource } from "../../testing/workbookSavedViewTestSupport";
import type { WorkbookSavedViewPort } from "../ports/WorkbookSavedViewPort";
import { SavedViewDiscovery } from "./SavedViewDiscovery";
import { SavedViewResourceObserver } from "./SavedViewResourceObserver";

const schema = "cartulary.view.timeline.v2";
const accepted = (id = "saved-1", version = 1) => ({
  kind: "accepted" as const,
  value: savedViewTestResource({
    saved_view_id: id,
    saved_view_version: version,
  }),
});
const failure = {
  kind: "rejected" as const,
  failure: { kind: "transport" as const, message: "Read failed" },
};
function port(): WorkbookSavedViewPort {
  return {
    listPage: vi.fn(async () => ({
      kind: "accepted" as const,
      value: { savedViews: [accepted().value], nextCursor: "next" },
    })),
    getResource: vi.fn(async ({ savedViewId }) => accepted(savedViewId)),
    create: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  };
}
async function settle() {
  for (let i = 0; i < 20; i++) await Promise.resolve();
}

describe("Saved-view independent reads", () => {
  it("publishes the first page without traversal and retains it through delayed failed continuation and retry", async () => {
    const p = port();
    const discovery = new SavedViewDiscovery({
      port: () => p,
      observe: observeAsyncOperation,
      changed: () => {},
      failed: () => {},
    });
    discovery.setSchema(schema);
    expect(p.listPage).not.toHaveBeenCalled();
    discovery.open();
    await settle();
    expect(discovery.getSnapshot().candidates).toHaveLength(1);
    expect(p.listPage).toHaveBeenCalledTimes(1);
    const pending =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["listPage"]>>>();
    vi.mocked(p.listPage).mockReturnValueOnce(pending.promise);
    const next = discovery.next();
    expect(discovery.getSnapshot().pending).toBe(true);
    expect(discovery.getSnapshot().candidates[0]?.saved_view_id).toBe(
      "saved-1",
    );
    pending.resolve(failure);
    await next;
    expect(discovery.getSnapshot()).toMatchObject({
      accepted: true,
      stale: true,
      cursor: null,
      previous: [],
    });
    vi.mocked(p.listPage).mockResolvedValueOnce({
      kind: "accepted",
      value: { savedViews: [accepted("saved-2").value], nextCursor: null },
    });
    await discovery.retry();
    expect(discovery.getSnapshot()).toMatchObject({
      cursor: "next",
      previous: [null],
      stale: false,
      nextCursor: null,
    });
    expect(p.listPage).toHaveBeenLastCalledWith(
      expect.objectContaining({
        viewSchemaId: schema,
        limit: 50,
        cursorToken: "next",
      }),
    );
    discovery.clear();
  });
  it("bounds candidates and checkpoints across repeated forward and backward browsing", async () => {
    const p = port();
    vi.mocked(p.listPage).mockImplementation(async ({ cursorToken }) => {
      const n = Number(cursorToken ?? 0);
      return {
        kind: "accepted",
        value: {
          savedViews: Array.from(
            { length: 50 },
            (_, i) => accepted(`${n}-${i}`).value,
          ),
          nextCursor: String(n + 1),
        },
      };
    });
    const d = new SavedViewDiscovery({
      port: () => p,
      observe: observeAsyncOperation,
      changed: () => {},
      failed: () => {},
    });
    d.setSchema(schema);
    d.open();
    await settle();
    for (let i = 0; i < 30; i++) {
      await d.next();
      expect(d.getSnapshot().candidates).toHaveLength(50);
      expect(d.getSnapshot().previous.length).toBeLessThanOrEqual(10);
    }
    for (let i = 0; i < 10; i++) await d.previous();
    expect(d.getSnapshot().cursor).toBe("20");
    expect(d.getSnapshot().previous).toHaveLength(0);
    await d.first();
    expect(d.getSnapshot().cursor).toBeNull();
    expect(d.getSnapshot().previous).toHaveLength(0);
    d.clear();
  });
  it("preserves committed position on dismissal and fences schema replacement and invalidation", async () => {
    const p = port();
    const d = new SavedViewDiscovery({
      port: () => p,
      observe: observeAsyncOperation,
      changed: () => {},
      failed: () => {},
    });
    d.setSchema(schema);
    d.open();
    await settle();
    const late =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["listPage"]>>>();
    vi.mocked(p.listPage).mockReturnValueOnce(late.promise);
    void d.next();
    d.close();
    late.resolve({
      kind: "accepted",
      value: { savedViews: [accepted("late").value], nextCursor: null },
    });
    await settle();
    expect(d.getSnapshot()).toMatchObject({ open: false, cursor: null });
    d.open();
    await settle();
    expect(d.getSnapshot().candidates[0]?.saved_view_id).toBe("saved-1");
    d.invalidate("saved-1");
    expect(d.getSnapshot().candidates).toHaveLength(0);
    d.setSchema("cartulary.view.notes.v1");
    expect(d.getSnapshot()).toMatchObject({
      accepted: false,
      previous: [],
      candidates: [],
    });
    d.clear();
  });
  it("rejects cyclic cursors and duplicate candidates without losing the accepted page", async () => {
    const p = port();
    const failed = vi.fn();
    const d = new SavedViewDiscovery({
      port: () => p,
      observe: observeAsyncOperation,
      changed: () => {},
      failed,
    });
    d.setSchema(schema);
    d.open();
    await settle();
    await d.next();
    expect(d.getSnapshot().problem?.kind).toBe("invalid_contract");
    expect(d.getSnapshot().candidates).toHaveLength(1);
    expect(failed).toHaveBeenCalled();
    d.clear();
  });
  it("retains selected resources independently and releases unused observations", async () => {
    const p = port();
    const owner = new SavedViewResourceObserver({
      port: () => p,
      observe: observeAsyncOperation,
      visible: () => true,
      changed: () => {},
      failed: () => {},
    });
    owner.retain("selected", "saved-1");
    owner.accept(accepted().value);
    vi.mocked(p.getResource).mockResolvedValueOnce(failure);
    await owner.read("saved-1");
    expect(owner.get("saved-1")).toMatchObject({
      status: "ready",
      resource: { saved_view_id: "saved-1" },
      problem: { kind: "transport" },
    });
    for (let i = 0; i < 40; i++) {
      owner.retain("activation", `candidate-${i}`);
      await owner.read(`candidate-${i}`);
      expect(owner.getSnapshot().size).toBe(2);
    }
    owner.retain("activation", null);
    expect(owner.getSnapshot().size).toBe(1);
    owner.clear();
  });
  it("fences older resource reads after receipts deletion and retirement", async () => {
    const p = port();
    const owner = new SavedViewResourceObserver({
      port: () => p,
      observe: observeAsyncOperation,
      visible: () => true,
      changed: () => {},
      failed: () => {},
    });
    owner.retain("selected", "saved-1");
    owner.accept(accepted().value);
    const late =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["getResource"]>>>();
    vi.mocked(p.getResource).mockReturnValueOnce(late.promise);
    void owner.read("saved-1");
    owner.accept(accepted("saved-1", 3).value);
    late.resolve(accepted("saved-1", 2));
    await settle();
    expect(owner.get("saved-1")?.resource?.saved_view_version).toBe(3);
    const deletion =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["getResource"]>>>();
    vi.mocked(p.getResource).mockReturnValueOnce(deletion.promise);
    void owner.read("saved-1");
    owner.invalidate();
    owner.unavailable("saved-1", {
      kind: "unavailable_target",
      message: "Unavailable",
    });
    deletion.resolve(accepted());
    await settle();
    expect(owner.get("saved-1")?.status).toBe("unavailable");
    owner.clear();
    expect(owner.getSnapshot().size).toBe(0);
  });
  it("distinguishes unavailable resources from failed observations without inferring receipts", async () => {
    const p = port();
    const failed = vi.fn();
    const owner = new SavedViewResourceObserver({
      port: () => p,
      observe: observeAsyncOperation,
      visible: () => true,
      changed: () => {},
      failed,
    });
    owner.retain("operation", "saved-1");
    owner.accept(accepted().value);
    vi.mocked(p.getResource).mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "unavailable_target", message: "Unavailable" },
    });
    await owner.read("saved-1");
    expect(owner.get("saved-1")).toMatchObject({
      status: "unavailable",
      resource: null,
    });
    expect(failed).toHaveBeenCalledWith(
      "saved-1",
      expect.objectContaining({ kind: "unavailable_target" }),
    );
    expect(p.create).not.toHaveBeenCalled();
    owner.clear();
  });
});
