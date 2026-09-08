import type { AuditEvent } from "../app/administrativeAuditModel";
import type { AuditReadResult } from "../app/api/administrativeAuditClient";

export const auditActorId = "00000000-0000-4000-8000-000000000001";
export function auditEvent(overrides: Partial<AuditEvent> = {}): AuditEvent {
  return {
    audit_event_id: "00000000-0000-4000-8000-000000002001",
    scope_kind: "deployment",
    scope_id: null,
    occurred_at: "2026-05-24T00:00:00Z",
    actor_kind: "user",
    actor_user_id: auditActorId,
    source: "ui",
    action_code: "user_created",
    target_kind: "user",
    target_id: "00000000-0000-4000-8000-000000000002",
    reason_code: null,
    changes: [
      {
        field_path: "display_name",
        value_state: "visible",
        before: null,
        after: "Target User",
      },
    ],
    ...overrides,
  };
}
export function auditPageResult(
  rows: AuditEvent[] = [auditEvent()],
  cursor: string | null = null,
): Extract<AuditReadResult, { ok: true }> {
  return {
    ok: true,
    rows,
    paging:
      cursor === null
        ? { limit: 100, has_more: false, next_cursor: null }
        : { limit: 100, has_more: true, next_cursor: cursor },
  };
}
export function auditEnvelope(
  rows: AuditEvent[] = [auditEvent()],
  cursor: string | null = null,
) {
  return {
    data: { audit_events: rows },
    meta: {
      request_id: "audit-test",
      paging: auditPageResult(rows, cursor).paging,
    },
  };
}
