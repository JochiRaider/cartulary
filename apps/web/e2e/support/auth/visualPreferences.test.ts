import type { Route } from "@playwright/test";
import { describe, expect, it, vi } from "vitest";
import { VisualPreferences } from "./visualPreferences";

describe("visual preference isolation", () => {
  it("starts each page from no override despite previous variants and failures", () => {
    for (const density of ["compact", "comfortable", "default"] as const) {
      const failedPage = new VisualPreferences("owned-test-account");
      failedPage.select(density);
      const nextPage = new VisualPreferences("owned-test-account");
      expect(nextPage.read().density_mode).toBeNull();
      expect(nextPage.read().preferences_version).toBe(1);
      const copy = nextPage.read();
      copy.density_mode = "comfortable";
      expect(nextPage.read().density_mode).toBeNull();
    }
  });

  it("blocks undeclared writes without forwarding requests to persistence", async () => {
    const resource = new VisualPreferences("owned-test-account");
    const abort = vi.fn();
    const route = {
      request: () => ({ method: () => "PUT" }),
      abort,
    } as unknown as Route;
    await expect(resource.route(route)).rejects.toMatchObject({
      name: "CartularyVisualCaptureError",
    });
    expect(abort).toHaveBeenCalledWith("blockedbyclient");
    expect(resource.unexpectedWrites).toEqual(["PUT"]);
    expect(resource.read().density_mode).toBeNull();
  });

  it("has one presentation write owner and immutable read snapshots", async () => {
    const resource = new VisualPreferences("owned-test-account");
    const writer = vi.fn(async () => {
      resource.select("comfortable");
    });
    resource.handleWrites(writer);
    expect(() => resource.handleWrites(writer)).toThrow(
      "already have a write owner",
    );
    const before = resource.read();
    await resource.route({
      request: () => ({ method: () => "PUT" }),
    } as unknown as Route);
    expect(writer).toHaveBeenCalledOnce();
    expect(before.density_mode).toBeNull();
    expect(resource.read().density_mode).toBe("comfortable");
  });
});
