import { afterEach, describe, expect, it, vi } from "vitest";
import {
  auditEnvelope,
  auditEvent,
} from "../../testing/administrativeAuditTestSupport";
import { jsonResponse } from "../../testing/fetchMockTestSupport";
import { emptyAuditQuery } from "../administrativeAuditModel";
import { listAdministrativeAuditPage } from "./administrativeAuditClient";

const request = () => ({
  query: emptyAuditQuery,
  limit: 100,
  cursor: null,
  signal: new AbortController().signal,
});
describe("Administrative audit typed client", () => {
  afterEach(() => vi.unstubAllGlobals());
  it("uses only the existing deployment operation and exact query members", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValue(jsonResponse(auditEnvelope()));
    vi.stubGlobal("fetch", fetch);
    const captured = {
      ...request(),
      query: {
        ...emptyAuditQuery,
        target_kind: "user" as const,
        target_id: "Exact + target",
      },
      cursor: "opaque +/%",
    };
    expect((await listAdministrativeAuditPage(captured)).ok).toBe(true);
    const url = new URL(
      String(fetch.mock.calls[0]?.[0]),
      "http://cartulary.test",
    );
    expect(url.pathname).toBe("/api/v1/administrative-audit-events");
    expect(Object.fromEntries(url.searchParams)).toEqual({
      limit: "100",
      target_kind: "user",
      target_id: "Exact + target",
      cursor_token: "opaque +/%",
    });
    expect(fetch.mock.calls[0]?.[1]?.signal).toBe(captured.signal);
  });
  it("accepts additive vocabulary while enforcing safe complete deployment pages", async () => {
    for (const action_code of [
      "future_action",
      "constructor",
      "toString",
      "__proto__",
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(
          jsonResponse(
            auditEnvelope([
              auditEvent({
                action_code,
                target_kind: "future_target",
                target_id: null,
              }),
            ]),
          ),
        ),
      );
      expect((await listAdministrativeAuditPage(request())).ok).toBe(true);
    }
    const event = auditEvent();
    for (const body of [
      { ...auditEnvelope(), meta: { request_id: "missing-paging" } },
      auditEnvelope([{ ...event, scope_kind: "incident" }]),
      auditEnvelope([{ ...event, changes: [] }]),
      auditEnvelope([event, event]),
      auditEnvelope([
        {
          ...event,
          changes: [
            {
              field_path: "password",
              value_state: "redacted",
              before: null,
              after: "must-not-leak",
            },
          ],
        },
      ]),
      auditEnvelope([
        {
          ...event,
          changes: [
            {
              field_path: "details",
              value_state: "visible",
              before: null,
              after: { nested_secret: "must-not-leak" },
            },
          ],
        },
      ]),
      {
        ...auditEnvelope(),
        meta: {
          request_id: "wrong-limit",
          paging: { limit: 20, has_more: false, next_cursor: null },
        },
      },
      {
        ...auditEnvelope(),
        meta: {
          request_id: "missing-cursor",
          paging: { limit: 100, has_more: true, next_cursor: null },
        },
      },
    ]) {
      vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(body)));
      const result = await listAdministrativeAuditPage(request());
      expect(result).toMatchObject({ ok: false, status: 502 });
      expect(JSON.stringify(result)).not.toContain("must-not-leak");
    }
  });
});
