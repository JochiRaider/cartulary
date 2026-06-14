import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { csrfHeaderName } from "./browserApi";
import {
  beginLatestQuery,
  fetchJSON,
  handleWorkbookLoadFailure,
  type LatestQueryRuntime,
  parseErrorMessage,
} from "./workbookApi";

describe("workbookApi", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    window.__cartularyWorkbookTimingProbe = undefined;
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("preserves WorkbookShell JSON and CSRF request semantics", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=test-csrf",
    );
    fetchMock.mockResolvedValue(jsonResponse({ data: { ok: true } }));

    await fetchJSON("/api/v1/workbook", {
      method: "POST",
      body: JSON.stringify({ client_txn_id: "txn-1" }),
    });

    const request = capturedRequest(fetchMock);
    expect(request.credentials).toBe("include");
    expect(request.headers.get("Content-Type")).toBe("application/json");
    expect(request.headers.get(csrfHeaderName)).toBe("test-csrf");
  });

  it("adds the timing debug header only for timeline row POSTs when the probe is installed", async () => {
    window.__cartularyWorkbookTimingProbe = { events: [] };
    fetchMock.mockResolvedValue(jsonResponse({ data: { ok: true } }));

    await fetchJSON(
      "/api/v1/incidents/incident-1/views/cartulary.view.timeline.v1/rows",
      { method: "POST", body: "{}" },
    );

    expect(
      capturedRequest(fetchMock).headers.get("X-Cartulary-Timing-Debug"),
    ).toBe("1");
  });

  it("keeps latest-query requests exclusive and aborts superseded controllers", () => {
    const runtime: { current: LatestQueryRuntime } = {
      current: { controller: null, sequence: 0 },
    };

    const first = beginLatestQuery(runtime);
    const second = beginLatestQuery(runtime);

    expect(first.signal.aborted).toBe(true);
    expect(first.isCurrent()).toBe(false);
    expect(second.signal.aborted).toBe(false);
    expect(second.isCurrent()).toBe(true);
  });

  it("keeps public error parsing and incident access-loss callbacks stable", () => {
    expect(
      parseErrorMessage({
        error: {
          code: "authorization_denied",
          details: { reason_code: "incident_not_found" },
        },
      }),
    ).toBe("authorization_denied: incident_not_found");

    const accessLost = vi.fn();
    expect(
      handleWorkbookLoadFailure(
        "authorization_denied",
        "Fallback.",
        accessLost,
      ),
    ).toBe("authorization_denied");
    expect(accessLost).toHaveBeenCalledTimes(1);
  });
});

function capturedRequest(fetchMock: ReturnType<typeof vi.fn>) {
  expect(fetchMock).toHaveBeenCalledTimes(1);
  return new Request("http://localhost/api", fetchMock.mock.calls[0]?.[1]);
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
