// @vitest-environment node

import { afterEach, describe, expect, it } from "vitest";

import { workerAdminIndexForParallelIndex } from "./fixtures";

const envNames = [
  "CARTULARY_BROWSER_GROUP_KIND",
  "CARTULARY_BROWSER_SESSION_GROUP",
  "CARTULARY_PLAYWRIGHT_WORKER_COUNT",
  "CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET",
] as const;

const originalEnv = Object.fromEntries(
  envNames.map((name) => [name, process.env[name]]),
);

afterEach(() => {
  for (const name of envNames) {
    const original = originalEnv[name];
    if (original === undefined) {
      delete process.env[name];
    } else {
      process.env[name] = original;
    }
  }
});

function clearWorkerEnv() {
  for (const name of envNames) {
    delete process.env[name];
  }
}

describe("workerAdminIndexForParallelIndex", () => {
  it("keeps the direct isolated Playwright default at worker-admin slot zero", () => {
    clearWorkerEnv();

    expect(workerAdminIndexForParallelIndex(0)).toBe(0);
    expect(workerAdminIndexForParallelIndex(2)).toBe(2);
  });

  it("uses an explicit direct offset without requiring a worker count", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "4";

    expect(workerAdminIndexForParallelIndex(1)).toBe(5);
  });

  it("requires scheduled browser groups to declare an offset", () => {
    clearWorkerEnv();
    process.env.CARTULARY_BROWSER_GROUP_KIND = "visual";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "8";

    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET is required for scheduled browser groups",
    );
  });

  it("requires scheduled browser groups to declare a worker count", () => {
    clearWorkerEnv();
    process.env.CARTULARY_BROWSER_SESSION_GROUP =
      "default-check-browser-shared";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "7";

    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "CARTULARY_PLAYWRIGHT_WORKER_COUNT is required for scheduled browser groups",
    );
  });

  it("rejects invalid scheduled offsets before product tests run", () => {
    clearWorkerEnv();
    process.env.CARTULARY_BROWSER_GROUP_KIND = "a11y";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "8";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "01";

    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET must be a non-negative integer",
    );
  });

  it("rejects scheduled worker ranges outside the provisioned slot pool", () => {
    clearWorkerEnv();
    process.env.CARTULARY_BROWSER_GROUP_KIND = "stateful";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "8";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "7";

    expect(() => workerAdminIndexForParallelIndex(1)).toThrow(
      "scheduled browser group worker slot 8 is outside CARTULARY_PLAYWRIGHT_WORKER_COUNT=8",
    );
  });

  it("accepts scheduled worker ranges inside the provisioned slot pool", () => {
    clearWorkerEnv();
    process.env.CARTULARY_BROWSER_GROUP_KIND = "visual";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "10";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "8";

    expect(workerAdminIndexForParallelIndex(1)).toBe(9);
  });
});
