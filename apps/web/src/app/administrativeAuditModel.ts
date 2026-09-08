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
export const auditFilterFields = [
  "actor_user_id",
  "action_code",
  "target_kind",
  "target_id",
  "occurred_at_gte",
  "occurred_at_lt",
] as const;
export type AuditField = (typeof auditFilterFields)[number];
export type AuditInputs = Readonly<Record<AuditField, string>>;
export type AuditQuery = Omit<AuditInputs, "action_code" | "target_kind"> & {
  readonly action_code: "" | (typeof auditActionCodes)[number];
  readonly target_kind: "" | (typeof auditTargetKinds)[number];
};
export type AuditFieldErrors = Partial<Record<AuditField, string>>;
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
const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/iu;

/** Match the route's UTC RFC3339Nano normalization without millisecond rounding. */
export function normalizeAuditInstant(
  input: string,
): { text: string; nanoseconds: bigint } | null {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?(Z|([+-])(\d{2}):(\d{2}))$/u.exec(
      input,
    );
  if (!match) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const hour = Number(match[4]);
  const minute = Number(match[5]);
  const second = Number(match[6]);
  const offsetHour = Number(match[10] ?? 0);
  const offsetMinute = Number(match[11] ?? 0);
  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    hour > 23 ||
    minute > 59 ||
    second > 59 ||
    offsetHour > 23 ||
    offsetMinute > 59
  )
    return null;
  const date = new Date(0);
  date.setUTCFullYear(year, month - 1, day);
  date.setUTCHours(hour, minute, second, 0);
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  )
    return null;
  const offset = (offsetHour * 60 + offsetMinute) * (match[9] === "-" ? -1 : 1);
  date.setTime(date.getTime() - offset * 60_000);
  if (date.getUTCFullYear() < 0 || date.getUTCFullYear() > 9999) return null;
  // The existing Go route admits fractional digits at nanosecond precision.
  const fraction = (match[7] ?? "").slice(0, 9).padEnd(9, "0");
  const significant = fraction.replace(/0+$/u, "");
  return {
    text: `${date.toISOString().slice(0, 19)}${significant ? `.${significant}` : ""}Z`,
    nanoseconds: BigInt(date.getTime()) * 1_000_000n + BigInt(fraction),
  };
}

export function validateAuditInputs(inputs: AuditInputs): {
  query: AuditQuery;
  errors: AuditFieldErrors;
} {
  const values = { ...inputs };
  const errors: AuditFieldErrors = {};
  for (const field of auditFilterFields) {
    values[field] = inputs[field].trim();
    const value = values[field];
    if (
      value &&
      (value === "null" ||
        value.includes(",") ||
        value.startsWith("[") ||
        value.endsWith("]"))
    )
      errors[field] = "Enter one exact value, or leave this filter blank.";
  }
  if (values.actor_user_id && !uuidPattern.test(values.actor_user_id))
    errors.actor_user_id = "Enter a user ID in UUID format.";
  values.actor_user_id = values.actor_user_id.toLowerCase();
  const action = auditActionCodes.find((value) => value === values.action_code);
  const target = auditTargetKinds.find((value) => value === values.target_kind);
  if (values.action_code && !action)
    errors.action_code = "Choose a current deployment action code.";
  if (values.target_kind && !target)
    errors.target_kind = "Choose a current deployment target kind.";
  if (values.target_id && !values.target_kind)
    errors.target_kind = "Choose a target kind when supplying a target ID.";
  const lower = values.occurred_at_gte
    ? normalizeAuditInstant(values.occurred_at_gte)
    : null;
  const upper = values.occurred_at_lt
    ? normalizeAuditInstant(values.occurred_at_lt)
    : null;
  for (const [field, parsed] of [
    ["occurred_at_gte", lower],
    ["occurred_at_lt", upper],
  ] as const) {
    if (values[field] && !parsed)
      errors[field] =
        "Enter a valid RFC 3339 timestamp with Z or a numeric timezone offset.";
    if (parsed) values[field] = parsed.text;
  }
  if (lower && upper && lower.nanoseconds >= upper.nanoseconds)
    errors.occurred_at_lt =
      "The exclusive upper bound must be later than the inclusive lower bound.";
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
