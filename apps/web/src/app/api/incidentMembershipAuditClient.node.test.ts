import { afterEach, describe, expect, it, vi } from "vitest";
import {
  auditEnvelope,
  auditEvent,
} from "../../testing/administrativeAuditTestSupport";
import { emptyMembershipAuditQuery } from "../incidentMembershipAuditModel";
import { listIncidentMembershipAuditPage } from "./incidentMembershipAuditClient";

const incidentId = "00000000-0000-4000-8000-000000001001";
const event = () =>
  auditEvent({
    scope_kind: "incident",
    scope_id: incidentId,
    action_code: "membership_role_changed",
    target_kind: "incident_membership",
  });
const request = () => ({
  authority: {
    incidentId,
    actorId: "actor",
    lifetime: "one",
    apiBase: "https://cartulary.test",
  },
  query: emptyMembershipAuditQuery,
  limit: 100,
  cursor: null,
  signal: new AbortController().signal,
});
afterEach(() => vi.unstubAllGlobals());
describe("Incident membership audit typed client", () => {
  it("uses the generated incident operation with only exact filters and captured cursor context", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(
      async () =>
        new Response(JSON.stringify(auditEnvelope([event()])), {
          headers: { "Content-Type": "application/json" },
        }),
    );
    vi.stubGlobal("fetch", fetch);
    const result = await listIncidentMembershipAuditPage({
      ...request(),
      cursor: "opaque",
      query: {
        ...emptyMembershipAuditQuery,
        target_kind: "incident_membership",
        target_id: "member",
        occurred_at_gte: "2026-05-24T00:00:00.000000001Z",
      },
    });
    expect(result.ok).toBe(true);
    const url = new URL(String(fetch.mock.calls[0]?.[0]));
    expect(url.pathname).toBe(
      `/api/v1/incidents/${incidentId}/membership-audit-events`,
    );
    expect(Object.fromEntries(url.searchParams)).toEqual({
      limit: "100",
      cursor_token: "opaque",
      target_kind: "incident_membership",
      target_id: "member",
      occurred_at_gte: "2026-05-24T00:00:00.000000001Z",
    });
  });
  it("accepts additive read vocabulary and safe redaction without weakening response validation", async () => {
    const row = {
      ...event(),
      action_code: "future_action",
      target_kind: "future_target",
      changes: [
        {
          field_path: "credential",
          value_state: "redacted" as const,
          before: null,
          after: null,
        },
      ],
    };
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(JSON.stringify(auditEnvelope([row])), {
            headers: { "Content-Type": "application/json" },
          }),
      ),
    );
    expect(await listIncidentMembershipAuditPage(request())).toMatchObject({
      ok: true,
      rows: [row],
    });
  });
  it("rejects malformed wrong scope unordered duplicate and inconsistent paging responses", async () => {
    const initial = auditEnvelope([event()], "next");
    const cases: unknown[] = [
      { ...initial, data: {} },
      { ...initial, meta: { request_id: "missing-paging" } },
      auditEnvelope([
        { ...event(), scope_id: "00000000-0000-4000-8000-000000001002" },
      ]),
      auditEnvelope([auditEvent()]),
      auditEnvelope([event(), event()]),
      auditEnvelope([
        event(),
        { ...event(), audit_event_id: "00000000-0000-4000-8000-000000009999" },
      ]),
      auditEnvelope([
        {
          ...event(),
          actor_kind: "system",
          actor_user_id: event().actor_user_id,
        },
      ]),
      auditEnvelope([{ ...event(), changes: [] }]),
      auditEnvelope([
        {
          ...event(),
          changes: [
            {
              field_path: "password",
              value_state: "redacted",
              before: "forbidden",
              after: null,
            },
          ],
        },
      ]),
      auditEnvelope([
        {
          ...event(),
          changes: [
            {
              field_path: "value",
              value_state: "visible",
              before: { password: "forbidden" },
              after: null,
            },
          ],
        },
      ]),
      {
        ...initial,
        meta: {
          request_id: "bad-limit",
          paging: { limit: 50, has_more: true, next_cursor: "next" },
        },
      },
      {
        ...initial,
        meta: {
          request_id: "bad-paging",
          paging: { limit: 100, has_more: false, next_cursor: "next" },
        },
      },
      auditEnvelope([], "next"),
    ];
    for (const payload of cases) {
      vi.stubGlobal(
        "fetch",
        vi.fn(
          async () =>
            new Response(JSON.stringify(payload), {
              headers: { "Content-Type": "application/json" },
            }),
        ),
      );
      expect(await listIncidentMembershipAuditPage(request())).toMatchObject({
        ok: false,
        status: 502,
        error: { code: "invalid_public_contract_response" },
      });
    }
  });
});
