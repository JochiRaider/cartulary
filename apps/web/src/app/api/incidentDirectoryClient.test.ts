import { afterEach, describe, expect, it, vi } from "vitest";
import { incidentResource } from "../../testing/appShellTestSupport";
import { jsonResponse } from "../../testing/fetchMockTestSupport";
import { listVisibleIncidents } from "./appShellClient";

describe("incident directory HTTP adapter", () => {
  afterEach(() => vi.unstubAllGlobals());

  const incident = incidentResource(
    "00000000-0000-4000-8000-000000004001",
    "IR-HTTP",
    "Directory contract",
  );
  const envelope = () => ({
    data: { incidents: [incident] },
    meta: {
      request_id: "request-list",
      paging: { limit: 100, has_more: false, next_cursor: null },
    },
  });

  it("uses declared query members and preserves opaque cursor and search values", async () => {
    const fetch = vi
      .fn()
      .mockImplementation(() => Promise.resolve(jsonResponse(envelope())));
    vi.stubGlobal("fetch", fetch);
    const controller = new AbortController();
    const result = await listVisibleIncidents({
      search: "  Alpha + β  ",
      status: "closed",
      cursorToken: "a +/%",
      signal: controller.signal,
    });
    expect(result.ok).toBe(true);
    const [input, init] = fetch.mock.calls[0] ?? [];
    const url = new URL(String(input), "http://cartulary.test");
    expect(url.pathname).toBe("/api/v1/incidents");
    expect(Object.fromEntries(url.searchParams)).toEqual({
      limit: "100",
      search: "  Alpha + β  ",
      status: "closed",
      cursor_token: "a +/%",
    });
    expect(init.signal).toBe(controller.signal);
    await listVisibleIncidents();
    expect(
      new URL(String(fetch.mock.calls[1]?.[0]), "http://cartulary.test").search,
    ).toBe("?limit=100");
  });

  it("rejects malformed resource and paging successes without filling defaults", async () => {
    const { status: _status, ...missingStatus } = incident;
    const { created_by_user_id: _creator, ...missingCreator } = incident;
    for (const payload of [
      { ...envelope(), data: { incidents: [missingStatus] } },
      { ...envelope(), data: { incidents: [missingCreator] } },
      {
        ...envelope(),
        data: { incidents: [{ ...incident, status: "archived" }] },
      },
      {
        ...envelope(),
        data: { incidents: [{ ...incident, incident_id: "bad-id" }] },
      },
      { ...envelope(), meta: { request_id: "request-list" } },
    ]) {
      vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(payload)));
      const result = await listVisibleIncidents();
      expect(result).toMatchObject({
        ok: false,
        status: 502,
        payload: { error: { code: "invalid_public_contract_response" } },
      });
    }
  });

  it("preserves public failures and propagates cancellation to the caller", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ error: { code: "session_required" } }, 401),
        ),
    );
    expect(await listVisibleIncidents()).toMatchObject({
      ok: false,
      status: 401,
      payload: { error: { code: "session_required" } },
    });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new DOMException("Aborted", "AbortError")),
    );
    await expect(listVisibleIncidents()).rejects.toMatchObject({
      name: "AbortError",
    });
  });
});
