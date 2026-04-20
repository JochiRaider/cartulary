// @vitest-environment node

import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { afterEach, describe, expect, it, vi } from "vitest";

const originalExternalServer = process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER;
const originalStateDir = process.env.CARTULARY_PLAYWRIGHT_STATE_DIR;
const cleanupPaths: string[] = [];

afterEach(() => {
  vi.resetModules();
  if (originalExternalServer === undefined) {
    delete process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER;
  } else {
    process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER = originalExternalServer;
  }
  if (originalStateDir === undefined) {
    delete process.env.CARTULARY_PLAYWRIGHT_STATE_DIR;
  } else {
    process.env.CARTULARY_PLAYWRIGHT_STATE_DIR = originalStateDir;
  }
  for (const cleanupPath of cleanupPaths.splice(0)) {
    rmSync(cleanupPath, { force: true, recursive: true });
  }
});

describe("harnessState", () => {
  it("falls back to tmpdir paths when shared state is not configured", async () => {
    delete process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER;
    delete process.env.CARTULARY_PLAYWRIGHT_STATE_DIR;

    const module = await import("./harnessState");

    expect(module.isExternalServerHarnessMode()).toBeFalsy();
    expect(module.sharedPlaywrightStateDir()).toBeNull();
    expect(module.usesSharedPlaywrightState()).toBeFalsy();
    expect(module.resolvePlaywrightStateFile("state.txt")).toBe(
      join(tmpdir(), "state.txt"),
    );
    expect(module.resolvePlaywrightStateDirectory("workers")).toBe(
      join(tmpdir(), "workers"),
    );
  });

  it("uses the configured shared-state directory in external-server mode", async () => {
    process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER = "1";
    process.env.CARTULARY_PLAYWRIGHT_STATE_DIR = mkdtempSync(
      join(tmpdir(), "cartulary-playwright-state-test-"),
    );
    cleanupPaths.push(process.env.CARTULARY_PLAYWRIGHT_STATE_DIR);

    const module = await import("./harnessState");

    expect(module.isExternalServerHarnessMode()).toBeTruthy();
    expect(module.sharedPlaywrightStateDir()).toBe(
      process.env.CARTULARY_PLAYWRIGHT_STATE_DIR,
    );
    expect(module.usesSharedPlaywrightState()).toBeTruthy();
    expect(module.resolvePlaywrightStateFile("state.txt")).toBe(
      join(process.env.CARTULARY_PLAYWRIGHT_STATE_DIR, "state.txt"),
    );
    expect(module.resolvePlaywrightStateDirectory("workers")).toBe(
      join(process.env.CARTULARY_PLAYWRIGHT_STATE_DIR, "workers"),
    );
  });
});
