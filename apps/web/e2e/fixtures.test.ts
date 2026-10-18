// @vitest-environment node

import type { SessionResource } from "@cartulary/protocol-ts/http";
import type { APIResponse } from "@playwright/test";
import { afterEach, describe, expect, it, vi } from "vitest";
import { workerAdminIndexForParallelIndex } from "./fixtures";
import { rewriteSessionPresentation } from "./support/auth/sessionPresentation";

function sessionResource(
  overrides: Partial<SessionResource> = {},
): SessionResource {
  return {
    user_id: "00000000-0000-4000-8000-000000000001",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "not_required",
    is_deployment_admin: false,
    authenticated_at: "2026-04-20T12:00:00Z",
    idle_expires_at: "2026-04-20T12:30:00Z",
    absolute_expires_at: "2026-04-20T20:00:00Z",
    session_expires_at: "2026-04-20T12:30:00Z",
    memberships: [],
    ...overrides,
  };
}

describe("session presentation", () => {
  function fixture(body: unknown, status = 200, malformedJSON = false) {
    const response = {
      status: () => status,
      ok: () => status >= 200 && status < 300,
      json: async () => {
        if (malformedJSON) throw new Error("private response text");
        return body;
      },
    } as APIResponse;
    const route = {
      fetch: vi.fn(async () => response),
      fulfill: vi.fn(),
      abort: vi.fn(),
    };
    return { route, response, report: vi.fn() };
  }

  it("validates successful sessions before preserving unrelated fields and rewriting presentation", async () => {
    const body = { data: sessionResource(), meta: { request_id: "original" } };
    const { route, response, report } = fixture(body);
    await rewriteSessionPresentation(
      route,
      (session) => ({ ...session, display_name: "Audit operator" }),
      report,
    );
    expect(route.fulfill).toHaveBeenCalledWith({
      response,
      json: {
        ...body,
        data: { ...body.data, display_name: "Audit operator" },
      },
    });
    expect(report).not.toHaveBeenCalled();
  });

  it("forwards upstream failures and malformed sessions without fabricating memberships or exposing private text", async () => {
    for (const [status, body, malformed] of [
      [
        500,
        {
          error: {
            code: "internal_error",
            message: "private diagnostic",
            request_id: "request-500",
          },
        },
        false,
      ],
      [401, { error: { code: "session_required" } }, false],
      [200, { data: { memberships: null } }, false],
      [200, { data: sessionResource({ memberships: [] }), meta: null }, false],
      [200, null, true],
    ] as const) {
      const { route, response, report } = fixture(body, status, malformed);
      const rewrite = vi.fn();
      await rewriteSessionPresentation(route, rewrite, report);
      expect(rewrite).not.toHaveBeenCalled();
      expect(route.fulfill).toHaveBeenCalledWith({ response });
      expect(report).toHaveBeenCalledOnce();
      expect(JSON.stringify(report.mock.calls)).not.toContain("private");
      expect(report.mock.calls[0]?.[0].status).toBe(status);
      if (status === 500)
        expect(report.mock.calls[0]?.[0]).toMatchObject({
          code: "internal_error",
          request_id: "request-500",
        });
    }
  });

  it("reports transport failure safely and aborts the intercepted request", async () => {
    const { route, report } = fixture(null);
    route.fetch.mockRejectedValue(new Error("private transport detail"));
    await rewriteSessionPresentation(route, vi.fn(), report);
    expect(route.abort).toHaveBeenCalledWith("failed");
    expect(route.fulfill).not.toHaveBeenCalled();
    expect(report).toHaveBeenCalledWith({
      operation: "getCurrentSession",
      status: null,
      reason: "transport_failure",
    });
  });
});

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
  it("rejects missing owned worker bindings", () => {
    clearWorkerEnv();
    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "owned browser worker offset must be exactly zero",
    );
  });
  it("rejects nonzero worker offsets", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "4";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "1";
    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "owned browser worker offset must be exactly zero",
    );
  });
  it("requires an owned worker count", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "0";
    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "owned browser worker count must be exactly one",
    );
  });
  it("rejects widened worker counts", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "0";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "2";
    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "owned browser worker count must be exactly one",
    );
  });
  it("rejects malformed worker offsets", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "01";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "1";
    expect(() => workerAdminIndexForParallelIndex(0)).toThrow(
      "owned browser worker offset must be exactly zero",
    );
  });
  it("rejects parallel indices outside the single owned slot", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "0";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "1";
    for (const index of [-1, 1, 0.5, Number.NaN])
      expect(() => workerAdminIndexForParallelIndex(index)).toThrow(
        "owned browser parallel index must be exactly zero",
      );
  });
  it("uses the single owned worker account", () => {
    clearWorkerEnv();
    process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET = "0";
    process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT = "1";
    expect(workerAdminIndexForParallelIndex(0)).toBe(0);
  });
});
