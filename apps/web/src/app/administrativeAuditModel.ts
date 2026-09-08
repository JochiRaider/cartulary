import {
  type AuditField,
  type AuditFieldErrors,
  type AuditInputs,
  auditFilterFields,
  normalizeAuditFilters,
} from "../shared/auditReadValues";

export {
  type AuditField,
  type AuditFieldErrors,
  type AuditInputs,
  auditFilterFields,
  normalizeAuditInstant,
} from "../shared/auditReadValues";

import type { ListAdministrativeAuditEventsResponse } from "./api/publicHttpTypes";

// Current request vocabulary projected from module.auth OpenAPI, parity-tested
// against its authored JSON. Response vocabulary remains additive.
export const auditActionCodes = [
  "account_preferences_updated",
  "auth_binding_created",
  "auth_binding_retired",
  "auth_binding_rotated",
  "backup_created",
  "bootstrap_admin_created",
  "deployment_admin_granted",
  "deployment_admin_revoked",
  "password_changed",
  "password_reset",
  "restore_completed",
  "restore_failed",
  "restore_started",
  "restore_verification_completed",
  "sessions_revoked",
  "totp_enrollment_begun",
  "totp_enrollment_completed",
  "totp_reset",
  "user_created",
  "user_profile_updated",
  "user_status_changed",
] as const;
export const auditTargetKinds = [
  "account_preferences",
  "auth_binding",
  "backup_set",
  "restore_operation",
  "user",
] as const;
export type AuditQuery = Omit<AuditInputs, "action_code" | "target_kind"> & {
  readonly action_code: "" | (typeof auditActionCodes)[number];
  readonly target_kind: "" | (typeof auditTargetKinds)[number];
};
export type AuditEvent =
  ListAdministrativeAuditEventsResponse["data"]["audit_events"][number];
export type AuditPaging = NonNullable<
  ListAdministrativeAuditEventsResponse["meta"]["paging"]
>;
export type AuditAuthority = {
  readonly lifetime: string;
  readonly actorId: string;
};
export const auditPageLimit = 100;
export const auditHistoryLimit = 20;
export const auditReadTimeout = 30_000;
export const emptyAuditQuery: AuditQuery = {
  actor_user_id: "",
  action_code: "",
  target_kind: "",
  target_id: "",
  occurred_at_gte: "",
  occurred_at_lt: "",
};
export const auditFieldLabels: Record<AuditField, string> = {
  actor_user_id: "Actor user ID",
  action_code: "Action code",
  target_kind: "Target kind",
  target_id: "Target ID",
  occurred_at_gte: "Occurred at or after",
  occurred_at_lt: "Occurred before",
};
export function validateAuditInputs(inputs: AuditInputs): {
  query: AuditQuery;
  errors: AuditFieldErrors;
} {
  const { values, errors } = normalizeAuditFilters(inputs);
  const action = auditActionCodes.find((value) => value === values.action_code);
  const target = auditTargetKinds.find((value) => value === values.target_kind);
  if (values.action_code && !action)
    errors.action_code = "Choose a current deployment action code.";
  if (values.target_kind && !target)
    errors.target_kind = "Choose a current deployment target kind.";
  return {
    query: { ...values, action_code: action ?? "", target_kind: target ?? "" },
    errors,
  };
}
export function sameAuditQuery(a: AuditInputs, b: AuditInputs) {
  return auditFilterFields.every((field) => a[field] === b[field]);
}
export function auditBinding(authority: AuditAuthority, query: AuditQuery) {
  return JSON.stringify([
    authority.lifetime,
    authority.actorId,
    "/api/v1/administrative-audit-events",
    "deployment",
    null,
    auditPageLimit,
    "occurred_at_desc,audit_event_id_desc",
    ...auditFilterFields.map((field) => query[field]),
  ]);
}
export type AuditPageDescriptor = {
  readonly binding: string;
  readonly cursor: string | null;
  readonly number: number;
};
export type AuditPage = {
  readonly descriptor: AuditPageDescriptor;
  readonly rows: readonly AuditEvent[];
  readonly paging: AuditPaging;
};
export type AuditReadKind = "initial" | "refresh" | "page" | "reactivate";
export type AuditProblem = {
  readonly kind: "initial" | "refresh" | "page" | "query" | "cursor" | "access";
  readonly message: string;
};
export type AuditState = {
  readonly authority: AuditAuthority | null;
  readonly active: boolean;
  readonly inputs: AuditInputs;
  readonly applied: AuditQuery;
  readonly fieldErrors: AuditFieldErrors;
  readonly page: AuditPage | null;
  readonly history: readonly AuditPageDescriptor[];
  readonly continuationValid: boolean;
  readonly activity: AuditReadKind | null;
  readonly problem: AuditProblem | null;
  readonly expandedId: string | null;
  readonly scrollTop: number;
  readonly tableScrollLeft: number;
  readonly announcement: string;
  readonly announcementRole: "status" | "alert";
  readonly announcementSequence: number;
};
export function initialAuditState(): AuditState {
  return {
    authority: null,
    active: false,
    inputs: emptyAuditQuery,
    applied: emptyAuditQuery,
    fieldErrors: {},
    page: null,
    history: [],
    continuationValid: false,
    activity: null,
    problem: null,
    expandedId: null,
    scrollTop: 0,
    tableScrollLeft: 0,
    announcement: "",
    announcementRole: "status",
    announcementSequence: 0,
  };
}
export function auditPageSummary(page: AuditPage) {
  return `Page ${page.descriptor.number}: ${page.rows.length} event${page.rows.length === 1 ? "" : "s"}. ${page.paging.has_more ? "More events available." : "No further events on this continuation."}`;
}
