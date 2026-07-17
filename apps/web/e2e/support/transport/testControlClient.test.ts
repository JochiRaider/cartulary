import { describe, expect, it, vi } from "vitest";

import { TestControlClient } from "./testControlClient";

function requestContext() {
  const fetch = vi.fn(async () => ({
    headers: () => ({ "x-request-id": "request-1" }),
    json: async () => ({ data: { accepted: true } }),
    ok: () => true,
    status: () => 202,
  }));
  return { fetch };
}

describe("test-control transport", () => {
  it("requires a token and approved exact origin before issuing a request", async () => {
    const request = requestContext();

    expect(
      () =>
        new TestControlClient({
          approvedEndpointOrigins: ["http://127.0.0.1:8080"],
          approvedRequestOrigins: ["http://127.0.0.1:4173"],
          endpointOrigin: "http://127.0.0.1:8080",
          request,
          requestOrigin: "http://127.0.0.1:4173",
          token: "",
        }),
    ).toThrow(/token must be non-empty/);
    expect(
      () =>
        new TestControlClient({
          approvedEndpointOrigins: ["http://127.0.0.1:8080"],
          approvedRequestOrigins: ["http://127.0.0.1:4173"],
          endpointOrigin: "http://localhost:8080",
          request,
          requestOrigin: "http://127.0.0.1:4173",
          token: "control-token",
        }),
    ).toThrow(/endpoint host is not approved/);
    expect(
      () =>
        new TestControlClient({
          approvedEndpointOrigins: ["http://127.0.0.1:8080"],
          approvedRequestOrigins: ["http://127.0.0.1:4173"],
          endpointOrigin: "http://127.0.0.1:8080",
          request,
          requestOrigin: "http://attacker.test",
          token: "control-token",
        }),
    ).toThrow(/request origin is not approved/);
    expect(request.fetch).not.toHaveBeenCalled();
  });

  it("sends privileged authorization only through the control facade", async () => {
    const request = requestContext();
    const client = new TestControlClient({
      approvedEndpointOrigins: ["http://127.0.0.1:8080"],
      approvedRequestOrigins: ["http://127.0.0.1:4173"],
      endpointOrigin: "http://127.0.0.1:8080",
      request,
      requestOrigin: "http://127.0.0.1:4173",
      token: "control-token",
    });

    await expect(
      client.request({
        body: { now: "2026-07-17T12:00:00Z" },
        method: "PUT",
        path: "/api/v1/test/clock",
      }),
    ).resolves.toMatchObject({ ok: true, status: 202 });
    expect(request.fetch).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/api/v1/test/clock",
      {
        data: { now: "2026-07-17T12:00:00Z" },
        headers: {
          Origin: "http://127.0.0.1:4173",
          "X-Cartulary-Test-Route-Token": "control-token",
        },
        method: "PUT",
      },
    );
    expect(JSON.stringify(client)).not.toContain("control-token");
  });

  it("rejects public, cross-origin, and traversal-shaped paths before mutation", async () => {
    const request = requestContext();
    const client = new TestControlClient({
      approvedEndpointOrigins: ["http://127.0.0.1:8080"],
      approvedRequestOrigins: ["http://127.0.0.1:4173"],
      endpointOrigin: "http://127.0.0.1:8080",
      request,
      requestOrigin: "http://127.0.0.1:4173",
      token: "control-token",
    });

    for (const path of [
      "/api/v1/incidents",
      "https://attacker.test/api/v1/test/clock",
      "//attacker.test/api/v1/test/clock",
    ]) {
      await expect(client.request({ method: "GET", path })).rejects.toThrow(
        /test-control path/,
      );
    }
    expect(request.fetch).not.toHaveBeenCalled();
  });
});
