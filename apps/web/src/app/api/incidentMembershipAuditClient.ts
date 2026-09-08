import { type APIError, fetchHTTPOperation } from "../../services/browserApi";
import {
  auditFilterFields,
  normalizeAuditInstant,
} from "../../shared/auditReadValues";
import type {
  MembershipAuditAuthority,
  MembershipAuditEvent,
  MembershipAuditPaging,
  MembershipAuditQuery,
} from "../incidentMembershipAuditModel";
import type { ListIncidentMembershipAuditEventsResponse } from "./publicHttpTypes";

export type MembershipAuditResult =
  | {
      readonly ok: true;
      readonly rows: readonly MembershipAuditEvent[];
      readonly paging: MembershipAuditPaging;
    }
  | { readonly ok: false; readonly status: number; readonly error: APIError };
export async function listIncidentMembershipAuditPage(request: {
  readonly authority: MembershipAuditAuthority;
  readonly query: MembershipAuditQuery;
  readonly limit: number;
  readonly cursor: string | null;
  readonly signal: AbortSignal;
}): Promise<MembershipAuditResult> {
  const query: Record<string, string | number> = { limit: request.limit };
  for (const field of auditFilterFields)
    if (request.query[field] !== "") query[field] = request.query[field];
  if (request.cursor !== null) query.cursor_token = request.cursor;
  const result =
    await fetchHTTPOperation<ListIncidentMembershipAuditEventsResponse>({
      operationID: "listIncidentMembershipAuditEvents",
      apiBase: request.authority.apiBase,
      pathParameters: { incident_id: request.authority.incidentId },
      query,
      init: { signal: request.signal },
    });
  if (!result.ok)
    return {
      ok: false,
      status: result.status,
      error: result.payload.error ?? { code: "unknown_public_error" },
    };
  const rows = result.payload.data.audit_events;
  const paging = result.payload.meta.paging;
  const invalid = (): MembershipAuditResult => ({
    ok: false,
    status: 502,
    error: { code: "invalid_public_contract_response" },
  });
  if (
    !paging ||
    result.status !== 200 ||
    paging.limit !== request.limit ||
    rows.length > paging.limit ||
    (paging.has_more
      ? !paging.next_cursor ||
        paging.next_cursor === request.cursor ||
        rows.length === 0
      : paging.next_cursor !== null)
  )
    return invalid();
  const ids = new Set<string>();
  let previous: { time: bigint; id: string } | null = null;
  for (const event of rows) {
    const instant = normalizeAuditInstant(event.occurred_at);
    if (
      event.scope_kind !== "incident" ||
      event.scope_id !== request.authority.incidentId ||
      (event.actor_kind === "user"
        ? event.actor_user_id === null
        : event.actor_user_id !== null) ||
      !instant ||
      !event.occurred_at.endsWith("Z") ||
      ids.has(event.audit_event_id)
    )
      return invalid();
    if (
      previous &&
      (instant.nanoseconds > previous.time ||
        (instant.nanoseconds === previous.time &&
          event.audit_event_id >= previous.id))
    )
      return invalid();
    ids.add(event.audit_event_id);
    previous = { time: instant.nanoseconds, id: event.audit_event_id };
  }
  return { ok: true, rows, paging };
}
