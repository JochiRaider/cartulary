import { describe, expect, it, vi } from "vitest";

import { atJsonOrigin, requestPublicJson } from "./publicJsonClient";

describe("public JSON transport", () => {
  it("preserves non-2xx JSON envelopes, status, and headers", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({ "x-request-id": "request-1" }),
      json: async () => ({
        error: { code: "row_version_conflict" },
        meta: { request_id: "request-1" },
      }),
      ok: () => false,
      status: () => 409,
    }));

    await expect(
      requestPublicJson({
        body: { base_row_version: 2 },
        headers: { "X-CSRF-Token": "csrf" },
        method: "PATCH",
        path: "/api/v1/incidents/incident-1/records/record-1",
        request: { fetch },
      }),
    ).resolves.toEqual({
      body: {
        error: { code: "row_version_conflict" },
        meta: { request_id: "request-1" },
      },
      headers: { "x-request-id": "request-1" },
      ok: false,
      status: 409,
    });
    expect(fetch).toHaveBeenCalledWith(
      "/api/v1/incidents/incident-1/records/record-1",
      {
        data: { base_row_version: 2 },
        headers: { "X-CSRF-Token": "csrf" },
        method: "PATCH",
      },
    );
  });

  it("preserves body omission and rejects absolute or protocol-relative paths", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({}),
      json: async () => ({ data: [] }),
      ok: () => true,
      status: () => 200,
    }));

    await requestPublicJson({
      method: "GET",
      path: "/api/v1/incidents?limit=10",
      request: { fetch },
    });

    expect(fetch).toHaveBeenCalledWith("/api/v1/incidents?limit=10", {
      method: "GET",
    });
    await expect(
      requestPublicJson({
        method: "GET",
        path: "https://attacker.test/api/v1/incidents",
        request: { fetch },
      }),
    ).rejects.toThrow(/relative same-origin path/);
    await expect(
      requestPublicJson({
        method: "GET",
        path: "//attacker.test/api/v1/incidents",
        request: { fetch },
      }),
    ).rejects.toThrow(/relative same-origin path/);
  });

  it("binds relative paths to an explicitly approved request origin", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({}),
      json: async () => ({ data: [] }),
      ok: () => true,
      status: () => 200,
    }));
    const request = atJsonOrigin({ fetch }, "http://127.0.0.1:8080");

    await request.fetch("/api/v1/incidents?limit=10", { method: "GET" });

    expect(fetch).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/api/v1/incidents?limit=10",
      { method: "GET" },
    );
    expect(() => atJsonOrigin({ fetch }, "file:///tmp/control")).toThrow(
      /absolute HTTP origin/,
    );
  });
});
