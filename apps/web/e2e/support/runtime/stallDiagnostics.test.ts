// @vitest-environment node

import { mkdtempSync, readFileSync, rmSync, statSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { recordStallDiagnostics } from "./stallDiagnostics";

const origin = "https://127.0.0.1:9622";
const jobId = "e7befad9-7ff5-4d9c-83fc-90bff3c85f34";
const subject = { kind: "job", job_id: jobId, last_status: "running" } as const;

function testInfo(directory: string, attach = vi.fn(async () => {})) {
  return { outputPath: (name: string) => join(directory, name), attach };
}

function response(status: number, body: unknown) {
  return {
    headers: () => ({}),
    json: async () => body,
    ok: () => status >= 200 && status < 300,
    status: () => status,
  };
}

describe("stall diagnostics", () => {
  let directory = "";
  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "cartulary-stall-"));
    vi.stubEnv("CARTULARY_TEST_ROUTE_TOKEN", "a".repeat(43));
    vi.stubEnv("CARTULARY_TEST_ROUTE_TOKEN_FILE", "");
    vi.stubEnv("CARTULARY_WEB_E2E_PUBLIC_ORIGIN", origin);
  });
  afterEach(() => {
    vi.unstubAllEnvs();
    rmSync(directory, { recursive: true, force: true });
  });

  it("records the server's view of a live stall beside the failed test", async () => {
    const data = { schema_id: "cartulary.test.runtime_stall_diagnostics.v1" };
    const fetch = vi.fn(async () => response(200, { data, meta: {} }));
    const info = testInfo(directory);
    await recordStallDiagnostics({ fetch }, origin, info, subject);

    expect(fetch).toHaveBeenCalledWith(
      `${origin}/api/v1/test/runtime/stall-diagnostics`,
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({
          "X-Cartulary-Test-Route-Token": "a".repeat(43),
        }),
      }),
    );
    const file = join(directory, `stall-diagnostics-${jobId}.json`);
    expect(JSON.parse(readFileSync(file, "utf8"))).toEqual({
      subject,
      diagnostics: data,
    });
    expect(statSync(file).mode & 0o777).toBe(0o600);
    expect(info.attach).toHaveBeenCalledWith("stall-diagnostics", {
      path: file,
      contentType: "application/json",
    });
  });

  it("never throws and records why the server view is unavailable", async () => {
    const refused = vi.fn(async () => response(403, { error: {} }));
    await recordStallDiagnostics(
      { fetch: refused },
      origin,
      testInfo(directory),
      subject,
    );
    const file = join(directory, `stall-diagnostics-${jobId}.json`);
    expect(JSON.parse(readFileSync(file, "utf8")).diagnostics).toEqual({
      unavailable: "status_403",
    });

    const lost = vi.fn(async () => {
      throw new Error("socket hang up");
    });
    const unnamed = { ...subject, job_id: "../escape" };
    await recordStallDiagnostics(
      { fetch: lost },
      origin,
      testInfo(directory),
      unnamed,
    );
    expect(
      JSON.parse(
        readFileSync(join(directory, "stall-diagnostics-unknown.json"), "utf8"),
      ),
    ).toEqual({
      subject: unnamed,
      diagnostics: { unavailable: "request_failed" },
    });

    const attach = vi.fn(async () => {
      throw new Error("attachment failed");
    });
    await expect(
      recordStallDiagnostics(
        { fetch: refused },
        origin,
        testInfo(directory, attach),
        subject,
      ),
    ).resolves.toBeUndefined();
  });
});
