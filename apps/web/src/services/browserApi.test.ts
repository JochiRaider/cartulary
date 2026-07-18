import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { apiPath, csrfHeaderName, fetchJSON, readCookie } from "./browserApi";

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

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
