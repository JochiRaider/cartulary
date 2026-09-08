import { type APIError, fetchHTTPOperation } from "../../services/browserApi";
import { validatedPublicErrorReason } from "../../services/publicErrorIdentity";
import { resolvePublicErrorPresentation } from "../../shared/publicErrorPresentation";
import {
  type AuditPage,
  type AuditQuery,
  auditFilterFields,
  normalizeAuditInstant,
} from "../administrativeAuditModel";
import type { ListAdministrativeAuditEventsResponse } from "./publicHttpTypes";

export type AuditReadResult =
  | {
      readonly ok: true;
      readonly rows: AuditPage["rows"];
      readonly paging: AuditPage["paging"];
    }
  | { readonly ok: false; readonly status: number; readonly error: APIError };

export async function listAdministrativeAuditPage(request: {
  readonly query: AuditQuery;
  readonly limit: number;
  readonly cursor: string | null;
  readonly signal: AbortSignal;
}): Promise<AuditReadResult> {
  const query: Record<string, string | number> = { limit: request.limit };
  for (const field of auditFilterFields)
    if (request.query[field] !== "") query[field] = request.query[field];
  if (request.cursor !== null) query.cursor_token = request.cursor;
  const result =
    await fetchHTTPOperation<ListAdministrativeAuditEventsResponse>({
      operationID: "listAdministrativeAuditEvents",
      query,
      init: { signal: request.signal },
    });
  if (!result.ok)
    return {
      ok: false,
      status: result.status,
      error: result.payload.error ?? { code: "unknown_public_error" },
    };
  const { audit_events: rows } = result.payload.data;
  const paging = result.payload.meta.paging;
  const invalid = (): AuditReadResult => ({
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
      event.scope_kind !== "deployment" ||
      event.scope_id !== null ||
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

export function auditFailureMessage(
  error: APIError,
  status: number,
  retained: boolean,
  paging = false,
) {
  const reason = validatedPublicErrorReason(
    error.code,
    error.details?.reason_code,
  );
  if (error.code === "invalid_list_query") {
    if (reason === "invalid_filter_range")
      return "The applied time range was rejected. Check both timestamps, their timezones, and that the lower bound precedes the upper bound.";
    if (reason === "invalid_filter_value")
      return "An applied filter was rejected. Check the actor ID, selected action and target kind, and the exact target ID.";
    return "The applied filters could not be accepted. Review the filters and apply them again.";
  }
  const presentation = resolvePublicErrorPresentation({
    code: error.code,
    reasonCode: reason,
    status,
    hasAuthorizedMaterialization: retained,
    operationFamily: retained ? "surface_refresh" : "surface_load",
  });
  switch (presentation.family) {
    case "authentication_required":
      return "Your session has ended. Sign in again to read the audit.";
    case "permission_or_incident_access_loss":
      return "Deployment administrator access is required to read this audit.";
    case "local_validation":
      return "The audit read was rejected. Review the applied filters and try again.";
    case "stale_refresh":
      return paging
        ? "The requested audit page could not be loaded. Displayed events were kept. Try the page again."
        : "Audit refresh failed. Displayed events may be stale. Try the read again.";
    default:
      return "Administrative audit is unavailable. Try the read again.";
  }
}
