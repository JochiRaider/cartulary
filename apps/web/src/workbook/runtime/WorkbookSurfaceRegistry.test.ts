import { describe, expect, it, vi } from "vitest";
import { WorkbookSurfaceRegistry } from "./WorkbookSurfaceRegistry";

function deferred() {
  let resolve!: () => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<void>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}
describe("Workbook surface refresh debt", () => {
  it("requires a later read when another acceptance arrives during refresh", async () => {
    const first = deferred();
    const second = deferred();
    const refresh = vi
      .fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const registry = new WorkbookSurfaceRegistry(vi.fn());
    registry.register("view", refresh);
    const a = registry.refreshRequired("view");
    await Promise.resolve();
    const b = registry.refreshRequired("view");
    first.resolve();
    await vi.waitFor(() => expect(refresh).toHaveBeenCalledTimes(2));
    expect(registry.requiresRefresh("view")).toBe(true);
    second.resolve();
    await Promise.all([a, b]);
    expect(registry.requiresRefresh("view")).toBe(false);
  });
  it("retains debt through failed reads and obsolete surface registrations", async () => {
    const first = deferred();
    const registry = new WorkbookSurfaceRegistry(vi.fn());
    const unregister = registry.register("view", () => first.promise);
    const running = registry.refreshRequired("view");
    await Promise.resolve();
    unregister();
    first.resolve();
    await expect(running).rejects.toThrow("surface changed");
    expect(registry.requiresRefresh("view")).toBe(true);
    const current = deferred();
    const stop = registry.register("view", () => current.promise);
    const failed = registry.refreshRequired("view");
    current.reject(new Error("read failed"));
    await expect(failed).rejects.toThrow("read failed");
    expect(registry.requiresRefresh("view")).toBe(true);
    stop();
    const refresh = vi.fn(async () => {});
    registry.register("view", refresh);
    await vi.waitFor(() =>
      expect(registry.requiresRefresh("view")).toBe(false),
    );
    expect(refresh).toHaveBeenCalledOnce();
    const read = deferred();
    registry.register("view", () => read.promise);
    const beforeAuthorizationChange = registry.refreshRequired("view");
    await Promise.resolve();
    registry.invalidateAuthority();
    read.resolve();
    await expect(beforeAuthorizationChange).rejects.toThrow(
      "Authorization changed",
    );
    expect(registry.requiresRefresh("view")).toBe(true);
  });
});
