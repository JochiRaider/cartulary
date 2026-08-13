// @vitest-environment node

import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  interactiveMeasurementSamplePolicy,
  parseServerTiming,
  percentile95,
} from "../../measurement/timingSupport";
import { reconcileSuiteAdminTotpState } from "../auth/suiteAdmin";
import { verifyOwnedHarnessRuntime } from "./readiness";

function clearHarnessTokenEnv() {
  delete process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE;
  delete process.env.CARTULARY_TEST_ROUTE_TOKEN;
  delete process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
  delete process.env.CARTULARY_WEB_E2E_SERVER_LOG;
}

beforeEach(() => {
  clearHarnessTokenEnv();
});

afterEach(() => {
  clearHarnessTokenEnv();
  vi.unstubAllGlobals();
});

describe("reconcileSuiteAdminTotpState", () => {
  it("reuses the stored suite-admin secret when it still authenticates", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({ kind: "success" as const })),
      provisionTotpFromBootstrap: vi.fn(async () => "unused"),
    };

    const secret = await reconcileSuiteAdminTotpState(
      client,
      "JBSWY3DPEHPK3PXP",
      {
        externalServerMode: true,
        sharedStateDir: "/tmp/cartulary-shared",
        stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
      },
    );

    expect(secret).toBe("JBSWY3DPEHPK3PXP");
    expect(client.loginLocal).toHaveBeenCalledOnce();
    expect(client.provisionTotpFromBootstrap).not.toHaveBeenCalled();
  });

  it("enrolls a new suite-admin secret when the backend requires TOTP setup", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({
        kind: "error" as const,
        status: 401,
        code: "mfa_setup_required",
        details: { bootstrap_token: "bootstrap-token-123" },
      })),
      provisionTotpFromBootstrap: vi.fn(async () => "JBSWY3DPEHPK3PXP"),
    };

    const secret = await reconcileSuiteAdminTotpState(client, null, {
      externalServerMode: true,
      sharedStateDir: "/tmp/cartulary-shared",
      stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
    });

    expect(secret).toBe("JBSWY3DPEHPK3PXP");
    expect(client.provisionTotpFromBootstrap).toHaveBeenCalledWith(
      "bootstrap-token-123",
    );
  });

  it("fails explicitly when shared harness state is missing for an active MFA setup", async () => {
    const client = {
      loginLocal: vi.fn(async () => ({
        kind: "error" as const,
        status: 401,
        code: "mfa_required",
        details: {},
      })),
      provisionTotpFromBootstrap: vi.fn(async () => "unused"),
    };

    await expect(
      reconcileSuiteAdminTotpState(client, null, {
        externalServerMode: true,
        sharedStateDir: "/tmp/cartulary-shared",
        stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
      }),
    ).rejects.toThrow(
      "suite admin MFA is already active but harness authentication state is unavailable",
    );
    const error = await reconcileSuiteAdminTotpState(client, null, {
      externalServerMode: true,
      sharedStateDir: "/tmp/cartulary-shared",
      stateFilePath: "/tmp/cartulary-shared/admin-totp.txt",
    }).catch((caught: unknown) => caught);
    const message = error instanceof Error ? error.message : String(error);
    expect(message).not.toContain("/tmp/cartulary-shared");
    expect(message).not.toContain("CARTULARY_PLAYWRIGHT_STATE_DIR");
  });
});

describe("verifyOwnedHarnessRuntime", () => {
  it("skips the identity probe when the owned-stack token file is not configured", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await verifyOwnedHarnessRuntime();

    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("verifies the token-protected harness identity before suite admin login", async () => {
    const tempDir = mkdtempSync(join(tmpdir(), "cartulary-identity-"));
    try {
      const tokenFile = join(tempDir, "token");
      writeFileSync(tokenFile, "0123456789abcdef0123456789abcdef\n", "utf8");
      process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE = tokenFile;
      const fetchMock = vi.fn(async () => ({
        headers: new Headers(),
        ok: true,
        status: 200,
        json: async () => ({
          data: {
            schema_id: "cartulary.test.runtime_identity.v1",
            runtime_marker: "harness-owned",
            server_pid: 12345,
            test_routes_enabled: true,
          },
        }),
      }));
      vi.stubGlobal("fetch", fetchMock);

      await verifyOwnedHarnessRuntime();

      expect(fetchMock).toHaveBeenCalledWith(
        "http://127.0.0.1:8080/api/v1/test/runtime/identity",
        expect.objectContaining({
          headers: {
            Origin: "http://127.0.0.1:8080",
            "X-Cartulary-Test-Route-Token": "0123456789abcdef0123456789abcdef",
          },
        }),
      );
    } finally {
      rmSync(tempDir, { force: true, recursive: true });
    }
  });

  it("fails before admin login when the identity endpoint is not the harness-owned runtime", async () => {
    const tempDir = mkdtempSync(join(tmpdir(), "cartulary-identity-"));
    try {
      const tokenFile = join(tempDir, "token");
      writeFileSync(tokenFile, "0123456789abcdef0123456789abcdef\n", "utf8");
      process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE = tokenFile;
      process.env.CARTULARY_WEB_E2E_SERVER_LOG = "/tmp/server.log";
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => ({
          headers: new Headers(),
          ok: false,
          status: 404,
          json: async () => ({}),
        })),
      );

      const error = await verifyOwnedHarnessRuntime().catch(
        (caught: unknown) => caught,
      );
      const message = error instanceof Error ? error.message : String(error);
      expect(message).toContain("browser harness identity check failed");
      expect(message).not.toContain(tokenFile);
      expect(message).not.toContain("/tmp/server.log");
    } finally {
      rmSync(tempDir, { force: true, recursive: true });
    }
  });
});

describe("qualified measurement helpers", () => {
  it("uses the nearest-rank boundary for exactly 100 measured samples", () => {
    const samples = [
      ...Array.from(
        { length: interactiveMeasurementSamplePolicy.measuredSamples - 5 },
        () => 100,
      ),
      ...Array.from({ length: 5 }, () => 1_000),
    ];

    expect(percentile95(samples, { sampleLabel: "single-outlier" })).toBe(100);
  });

  it("moves p95 to the high sample when six samples exceed the envelope", () => {
    const samples = [
      ...Array.from(
        { length: interactiveMeasurementSamplePolicy.measuredSamples - 6 },
        () => 100,
      ),
      ...Array.from({ length: 6 }, () => 1_000),
    ];

    expect(percentile95(samples, { sampleLabel: "six-outliers" })).toBe(1_000);
  });

  it("rejects p95 calculations below the typed measurement sample floor", () => {
    expect(() =>
      percentile95([100, 101, 102], { sampleLabel: "undersampled" }),
    ).toThrow("expected at least 100 samples, got 3");
  });

  it("rejects non-finite and negative samples", () => {
    expect(() =>
      percentile95([...Array.from({ length: 99 }, () => 1), Number.NaN], {
        sampleLabel: "nan",
      }),
    ).toThrow("finite non-negative");
    expect(() =>
      percentile95([...Array.from({ length: 99 }, () => 1), -1], {
        sampleLabel: "negative",
      }),
    ).toThrow("finite non-negative");
  });

  it("parses Server-Timing diagnostics while preserving raw metric entries", () => {
    expect(
      parseServerTiming(
        'store_base_insert;dur=212.522, store_commit;dur=20.312;desc="commit"',
      ),
    ).toEqual([
      {
        attributes: { dur: "212.522" },
        durationMs: 212.522,
        name: "store_base_insert",
        raw: "store_base_insert;dur=212.522",
      },
      {
        attributes: { desc: "commit", dur: "20.312" },
        durationMs: 20.312,
        name: "store_commit",
        raw: 'store_commit;dur=20.312;desc="commit"',
      },
    ]);
  });
});
