import { describe, expect, it, vi } from "vitest";

import {
  publicHttpOperation,
  publicHttpOperationObserved,
} from "./publicHttpOperationClient";
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

describe("public HTTP operation client", () => {
  it("derives methods, paths, and encoded queries from generated bindings", async () => {
    const fetch = vi.fn(async () => ({
      headers: () => ({ "x-request-id": "request-denied" }),
      json: async () => ({ error: { code: "authorization_denied" } }),
      ok: () => false,
      status: () => 403,
    }));

    await expect(
      publicHttpOperation({
        operationID: "listDeploymentUsers",
        query: { search: "alpha beta", limit: 10 },
        request: { fetch },
      }),
    ).resolves.toEqual({
      ok: false,
      payload: { error: { code: "authorization_denied" } },
      status: 403,
    });
    expect(fetch).toHaveBeenCalledWith(
      "/api/v1/users?limit=10&search=alpha%20beta",
      { method: "GET" },
    );
  });

  it("accepts valid success envelopes and rejects malformed success responses", async () => {
    const validSession = {
      data: {
        absolute_expires_at: "2026-08-05T01:00:00Z",
        authenticated_at: "2026-08-05T00:00:00Z",
        display_name: "Operator",
        idle_expires_at: "2026-08-05T00:30:00Z",
        is_deployment_admin: true,
        memberships: [],
        mfa_state: "satisfied",
        provider_type: "local",
        session_expires_at: "2026-08-05T00:30:00Z",
        user_id: "00000000-0000-4000-8000-000000000001",
      },
      meta: { request_id: "request-1" },
    };
    const fetch = vi
      .fn()
      .mockResolvedValueOnce({
        headers: () => ({}),
        json: async () => validSession,
        ok: () => true,
        status: () => 200,
      })
      .mockResolvedValueOnce({
        headers: () => ({}),
        json: async () => ({ data: {} }),
        ok: () => true,
        status: () => 200,
      });
    const request = { fetch };

    await expect(
      publicHttpOperation({ operationID: "getCurrentSession", request }),
    ).resolves.toEqual({ ok: true, payload: validSession, status: 200 });
    const malformed = await publicHttpOperation({
      operationID: "getCurrentSession",
      request,
    });

    expect(malformed.ok).toBe(false);
    expect(malformed.status).toBe(502);
    expect(malformed.payload).toEqual({
      error: expect.objectContaining({
        code: "invalid_public_contract_response",
        details: expect.objectContaining({
          operation_id: "getCurrentSession",
          schema_id: "cartulary.core_http.SessionEnvelope.v1",
        }),
      }),
    });
  });

  it("exposes repeated headers only after observed success validation", async () => {
    const validSession = {
      data: {
        absolute_expires_at: "2026-08-05T01:00:00Z",
        authenticated_at: "2026-08-05T00:00:00Z",
        display_name: "Operator",
        idle_expires_at: "2026-08-05T00:30:00Z",
        is_deployment_admin: true,
        memberships: [],
        mfa_state: "satisfied",
        provider_type: "local",
        session_expires_at: "2026-08-05T00:30:00Z",
        user_id: "00000000-0000-4000-8000-000000000001",
      },
      meta: { request_id: "request-login" },
    };
    const repeatedHeaders = [
      { name: "set-cookie", value: "cartulary_session=session; HttpOnly" },
      { name: "set-cookie", value: "cartulary_csrf=csrf" },
    ];
    const fetch = vi
      .fn()
      .mockResolvedValueOnce({
        headers: () => ({ "set-cookie": "cartulary_csrf=csrf" }),
        headersArray: () => repeatedHeaders,
        json: async () => validSession,
        ok: () => true,
        status: () => 200,
      })
      .mockResolvedValueOnce({
        headers: () => ({ "set-cookie": "cartulary_csrf=csrf" }),
        headersArray: () => repeatedHeaders,
        json: async () => ({ data: {} }),
        ok: () => true,
        status: () => 200,
      });
    const request = { fetch };

    const observed = await publicHttpOperationObserved({
      body: { password: "password", username: "operator@example.test" },
      operationID: "loginLocalUser",
      request,
    });
    expect(observed.ok).toBe(true);
    if (!observed.ok) {
      throw new Error("expected validated observed login success");
    }
    expect(observed.response.headersArray()).toEqual(repeatedHeaders);
    expect(fetch).toHaveBeenNthCalledWith(1, "/api/v1/auth/login", {
      data: { password: "password", username: "operator@example.test" },
      method: "POST",
    });

    const malformed = await publicHttpOperationObserved({
      body: { password: "password", username: "operator@example.test" },
      operationID: "loginLocalUser",
      request,
    });
    expect(malformed).toEqual({
      ok: false,
      payload: {
        error: expect.objectContaining({
          code: "invalid_public_contract_response",
        }),
      },
      status: 502,
    });
    expect("response" in malformed).toBe(false);
  });
});
