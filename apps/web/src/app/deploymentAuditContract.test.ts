import { describe, expect, it } from "vitest";

import { administrativeAuditEventsFromPayload } from "./deploymentAuditContract";

describe("deployment administrative-audit contract", () => {
  it("accepts the canonical response key and forward-compatible action and target values", () => {
    const auditEvents = [
      {
        audit_event_id: "20000000-0000-0000-0000-000000000001",
        scope_kind: "deployment",
        scope_id: null,
        occurred_at: "2026-07-25T19:30:00Z",
        actor_kind: "system",
        actor_user_id: null,
        source: "system",
        action_code: "future_administrative_action",
        target_kind: "future_administrative_target",
        target_id: "target-1",
        changes: [],
        reason_code: null,
      },
    ];

    expect(
      administrativeAuditEventsFromPayload({
        data: { audit_events: auditEvents },
        meta: { request_id: "request-1" },
      }),
    ).toEqual(auditEvents);
  });

  it("rejects the obsolete response alias and mixed response shapes", () => {
    expect(
      administrativeAuditEventsFromPayload({
        data: { administrative_audit_events: [] },
      }),
    ).toBeNull();
    expect(
      administrativeAuditEventsFromPayload({
        data: { audit_events: [], administrative_audit_events: [] },
      }),
    ).toBeNull();
  });

  it("rejects malformed event shapes while preserving redacted null values", () => {
    const redacted = {
      audit_event_id: "20000000-0000-0000-0000-000000000002",
      scope_kind: "deployment",
      scope_id: null,
      occurred_at: "2026-07-25T19:31:00Z",
      actor_kind: "operator",
      actor_user_id: null,
      source: "operator",
      action_code: "password_reset",
      target_kind: "user",
      target_id: "10000000-0000-0000-0000-000000000001",
      changes: [
        {
          field_path: "password",
          value_state: "redacted",
          before: null,
          after: null,
        },
      ],
      reason_code: null,
    };
    expect(
      administrativeAuditEventsFromPayload({
        data: { audit_events: [redacted] },
      }),
    ).toEqual([redacted]);
    expect(
      administrativeAuditEventsFromPayload({
        data: {
          audit_events: [
            {
              ...redacted,
              changes: [{ ...redacted.changes[0], before: "secret" }],
            },
          ],
        },
      }),
    ).toBeNull();
    expect(
      administrativeAuditEventsFromPayload({
        data: {
          audit_events: [{ ...redacted, unexpected: true }],
        },
      }),
    ).toBeNull();
  });
});
