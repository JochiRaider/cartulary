import {
  type HTTPOperationErrorCode,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  expectTypeOf,
  it,
  vi,
} from "vitest";

import { incidentResource } from "../testing/appShellTestSupport";
import {
  apiPath,
  csrfHeaderName,
  fetchHTTPOperation,
  fetchJSON,
  fetchMultipartHTTPOperation,
  readCookie,
} from "./browserApi";

describe("fetchJSON", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("adds JSON content type and CSRF header for mutating requests with included credentials", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=test-csrf",
    );
    fetchMock.mockImplementation(() =>
      Promise.resolve(jsonResponse({ data: { ok: true } })),
    );

    await fetchJSON("/api/v1/incidents", {
      method: "POST",
      body: JSON.stringify({ title: "IR-201" }),
    });

    const request = capturedRequest(fetchMock);
    expect(request.credentials).toBe("include");
    expect(request.headers.get("Content-Type")).toBe("application/json");
    expect(request.headers.get(csrfHeaderName)).toBe("test-csrf");
  });

  it("does not expose a private timing request protocol", async () => {
    fetchMock.mockResolvedValue(jsonResponse({ data: { ok: true } }));

    await fetchJSON(
      "/api/v1/incidents/incident-1/views/cartulary.view.timeline.v2/rows",
      { method: "POST", body: "{}" },
    );

    expect(
      capturedRequest(fetchMock).headers.get("X-Cartulary-Timing-Debug"),
    ).toBeNull();
  });

  it.each([
    {
      cookie: "other_cookie=value",
      label: "the CSRF cookie is absent",
    },
    {
      cookie: "cartulary_csrf=",
      label: "the CSRF cookie value is empty",
    },
  ])("omits the CSRF header when $label", async ({ cookie }) => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(cookie);
    fetchMock.mockResolvedValue(jsonResponse({ data: { ok: true } }));

    await fetchJSON("/api/v1/incidents", {
      method: "POST",
      body: JSON.stringify({ title: "IR-201" }),
    });

    const request = capturedRequest(fetchMock);
    expect(request.headers.get("Content-Type")).toBe("application/json");
    expect(request.headers.get(csrfHeaderName)).toBeNull();
  });

  it("omits the CSRF header for mutating requests that do not include credentials", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=test-csrf",
    );
    fetchMock.mockResolvedValue(jsonResponse({ data: { ok: true } }));

    await fetchJSON("/api/v1/auth/mfa/totp/begin", {
      method: "POST",
      credentials: "omit",
      body: JSON.stringify({ client_txn_id: "txn-1" }),
    });

    const request = capturedRequest(fetchMock);
    expect(request.credentials).toBe("omit");
    expect(request.headers.get("Content-Type")).toBe("application/json");
    expect(request.headers.get(csrfHeaderName)).toBeNull();
  });

  it("leaves GET requests without JSON content type or CSRF header", async () => {
    const cookieSpy = vi
      .spyOn(document, "cookie", "get")
      .mockReturnValue("cartulary_csrf=test-csrf");
    fetchMock.mockImplementation(() =>
      Promise.resolve(jsonResponse({ data: { ok: true } })),
    );

    await fetchJSON("/api/v1/incidents");

    const request = capturedRequest(fetchMock);
    expect(request.credentials).toBe("include");
    expect(request.headers.get("Content-Type")).toBeNull();
    expect(request.headers.get(csrfHeaderName)).toBeNull();
    expect(cookieSpy).not.toHaveBeenCalled();
  });

  it("route-boundary CSRF headers follow cookie-backed mutating request rules", async () => {
    const cookieSpy = vi
      .spyOn(document, "cookie", "get")
      .mockReturnValue("cartulary_csrf=test-csrf");
    fetchMock.mockImplementation(() =>
      Promise.resolve(jsonResponse({ data: { ok: true } })),
    );

    await fetchJSON("/api/v1/incidents", {
      method: "POST",
      body: JSON.stringify({ client_txn_id: "txn-include" }),
    });

    cookieSpy.mockReturnValue("other_cookie=value");
    await fetchJSON("/api/v1/incidents", {
      method: "POST",
      body: JSON.stringify({ client_txn_id: "txn-no-cookie" }),
    });

    cookieSpy.mockReturnValue("cartulary_csrf=test-csrf");
    await fetchJSON("/api/v1/auth/mfa/totp/begin", {
      method: "POST",
      credentials: "omit",
      body: JSON.stringify({ client_txn_id: "txn-bootstrap" }),
    });

    await fetchJSON("/api/v1/auth/session");

    expect(fetchMock).toHaveBeenCalledTimes(4);
    expect(requestAt(fetchMock, 0).credentials).toBe("include");
    expect(requestAt(fetchMock, 0).headers.get("Content-Type")).toBe(
      "application/json",
    );
    expect(requestAt(fetchMock, 0).headers.get(csrfHeaderName)).toBe(
      "test-csrf",
    );

    expect(requestAt(fetchMock, 1).credentials).toBe("include");
    expect(requestAt(fetchMock, 1).headers.get("Content-Type")).toBe(
      "application/json",
    );
    expect(requestAt(fetchMock, 1).headers.get(csrfHeaderName)).toBeNull();

    expect(requestAt(fetchMock, 2).credentials).toBe("omit");
    expect(requestAt(fetchMock, 2).headers.get("Content-Type")).toBe(
      "application/json",
    );
    expect(requestAt(fetchMock, 2).headers.get(csrfHeaderName)).toBeNull();

    expect(requestAt(fetchMock, 3).credentials).toBe("include");
    expect(requestAt(fetchMock, 3).headers.get("Content-Type")).toBeNull();
    expect(requestAt(fetchMock, 3).headers.get(csrfHeaderName)).toBeNull();
  });
});

describe("fetchHTTPOperation public error decoding", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  const createIncident = () =>
    fetchHTTPOperation({
      operationID: "createIncident",
      init: { method: "POST", body: "{}" },
    });

  it("surfaces a declared code typed by the operation and status", async () => {
    const error = publicError("incident_key_conflict", 409);
    fetchMock.mockResolvedValue(jsonResponse({ error }, 409));

    const result = await createIncident();

    expect(result).toEqual({ ok: false, status: 409, payload: { error } });
    if (result.ok) return;
    expectTypeOf(result.payload.error.code).toEqualTypeOf<
      | HTTPOperationErrorCode<"createIncident">
      | "invalid_public_contract_response"
    >();
    if (result.payload.error.status === 409) {
      expectTypeOf(result.payload.error.code).toEqualTypeOf<
        | "client_txn_conflict"
        | "credential_bootstrap_rejected"
        | "incident_key_conflict"
      >();
    }
    if (result.payload.error.status === 500) {
      expectTypeOf(result.payload.error.code).toEqualTypeOf<"internal_error">();
    }
    if (result.payload.error.status === 503) {
      expectTypeOf(
        result.payload.error.code,
      ).toEqualTypeOf<"service_unavailable">();
    }
  });

  it("decodes the declared 500 internal_error", async () => {
    const error = publicError("internal_error", 500);
    fetchMock.mockResolvedValue(jsonResponse({ error }, 500));

    const result = await fetchHTTPOperation({
      operationID: "activateReferencePackVersion",
      pathParameters: { pack_key: "type_registry.host", pack_version: "1" },
      init: { method: "POST", body: "{}" },
    });

    expect(result).toEqual({ ok: false, status: 500, payload: { error } });
  });

  it("decodes a 503 service_unavailable as a declared public error, never a contract failure", async () => {
    // The single REQ-03-086a rule needs both codes declared everywhere.
    for (const binding of Object.values(httpOperationBindings)) {
      const declared: Readonly<Record<string, readonly string[]>> =
        binding.error_codes_by_status;
      expect(declared["500"]).toContain("internal_error");
      expect(declared["503"]).toContain("service_unavailable");
    }
    const error = {
      ...publicError("service_unavailable", 503),
      details: { reason_code: "admission_paused" },
      retryable: true,
    };
    fetchMock.mockImplementation(() =>
      Promise.resolve(jsonResponse({ error }, 503)),
    );

    const results = [
      await createIncident(),
      await fetchHTTPOperation({ operationID: "listVisibleIncidents" }),
      await fetchHTTPOperation({ operationID: "getCurrentSession" }),
      await fetchMultipartHTTPOperation({
        operationID: "importIncidentBundle",
        body: new FormData(),
      }),
    ];

    for (const result of results)
      expect(result).toEqual({ ok: false, status: 503, payload: { error } });
    expect(JSON.stringify(results)).not.toContain(
      "invalid_public_contract_response",
    );
  });

  it.each([
    {
      body: { error: publicError("authorization_denied", 403) },
      instancePath: "/error/code",
      label: "an undeclared code",
      status: 403,
    },
    {
      body: { error: publicError("client_txn_conflict", 400) },
      instancePath: "/error/status",
      label: "a mismatched envelope status",
      status: 409,
    },
    {
      body: { error: publicError("internal_error", 503) },
      instancePath: "/error/code",
      label: "an internal_error outside status 500",
      status: 503,
    },
    {
      body: { error: publicError("service_unavailable", 500) },
      instancePath: "/error/code",
      label: "a service_unavailable at status 500",
      status: 500,
    },
    {
      body: { error: { code: "incident_key_conflict", status: 409 } },
      instancePath: "/error/message",
      label: "an invalid error envelope",
      status: 409,
    },
  ])(
    "turns $label into the client contract failure",
    async ({ body, instancePath, status }) => {
      fetchMock.mockResolvedValue(jsonResponse(body, status));

      const result = await createIncident();

      expect(result).toEqual({
        ok: false,
        status: 502,
        payload: {
          error: {
            code: "invalid_public_contract_response",
            details: {
              instance_path: instancePath,
              operation_id: "createIncident",
              response_status: status,
              schema_id: "cartulary.core_http.ErrorEnvelope.v1",
            },
            message: "The server returned an invalid public contract response.",
            retryable: true,
            status: 502,
          },
        },
      });
      expect(JSON.stringify(result)).not.toContain(body.error.code);
    },
  );

  it("turns a non-JSON error body into the client contract failure", async () => {
    fetchMock.mockResolvedValue(
      new Response("<html>Bad gateway</html>", {
        status: 502,
        headers: { "Content-Type": "text/html" },
      }),
    );

    const result = await createIncident();

    expect(result).toMatchObject({
      ok: false,
      status: 502,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          details: { operation_id: "createIncident", response_status: 502 },
        },
      },
    });
  });

  it("validates success payloads against the received success status", async () => {
    const envelope = {
      data: incidentResource(
        "00000000-0000-4000-8000-000000004101",
        "IR-A",
        "Alpha",
      ),
      meta: { request_id: "request-test" },
    };
    fetchMock
      .mockResolvedValueOnce(jsonResponse(envelope, 201))
      .mockResolvedValueOnce(jsonResponse(envelope, 202));

    expect(await createIncident()).toEqual({
      ok: true,
      status: 201,
      payload: envelope,
    });
    const result = await createIncident();

    expect(result).toMatchObject({
      ok: false,
      status: 502,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          details: { operation_id: "createIncident", response_status: 202 },
        },
      },
    });
  });

  it("decodes multipart operation errors against the same contract", async () => {
    const declared = publicError("incident_bundle_import_rejected", 409);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ error: declared }, 409))
      .mockResolvedValueOnce(
        jsonResponse({ error: publicError("incident_key_conflict", 409) }, 409),
      );
    const send = () =>
      fetchMultipartHTTPOperation({
        operationID: "importIncidentBundle",
        body: new FormData(),
      });

    expect(await send()).toEqual({
      ok: false,
      status: 409,
      payload: { error: declared },
    });
    expect(await send()).toMatchObject({
      ok: false,
      status: 502,
      payload: {
        error: {
          code: "invalid_public_contract_response",
          details: {
            instance_path: "/error/code",
            operation_id: "importIncidentBundle",
            response_status: 409,
          },
        },
      },
    });
  });
});

describe("readCookie", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("decodes the named cookie value", () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "other_cookie=ignore; cartulary_csrf=test%20csrf",
    );

    expect(readCookie("cartulary_csrf")).toBe("test csrf");
  });

  it("distinguishes an empty cookie from a missing cookie", () => {
    const cookieSpy = vi
      .spyOn(document, "cookie", "get")
      .mockReturnValue("cartulary_csrf=");

    expect(readCookie("cartulary_csrf")).toBe("");

    cookieSpy.mockReturnValue("other_cookie=value");
    expect(readCookie("cartulary_csrf")).toBeNull();
  });
});

describe("apiPath", () => {
  it("returns the path unchanged when the base is empty", () => {
    expect(apiPath(undefined, "/api/v1/incidents")).toBe("/api/v1/incidents");
    expect(apiPath("  ", "/api/v1/incidents")).toBe("/api/v1/incidents");
  });

  it("joins a trimmed base with the path", () => {
    expect(apiPath(" /cartulary ", "/api/v1/incidents")).toBe(
      "/cartulary/api/v1/incidents",
    );
    expect(apiPath("/cartulary/", "/api/v1/incidents")).toBe(
      "/cartulary/api/v1/incidents",
    );
  });
});

function capturedRequest(fetchMock: ReturnType<typeof vi.fn>) {
  expect(fetchMock).toHaveBeenCalledTimes(1);
  return new Request("http://localhost/api", fetchMock.mock.calls[0]?.[1]);
}

function requestAt(fetchMock: ReturnType<typeof vi.fn>, index: number) {
  return new Request("http://localhost/api", fetchMock.mock.calls[index]?.[1]);
}

function publicError(code: string, status: number) {
  return {
    code,
    details: {},
    message: "Request failed.",
    request_id: "request-test",
    retryable: false,
    status,
  };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
