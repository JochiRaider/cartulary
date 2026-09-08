import {
  type AuditFieldErrors,
  type AuditInputs,
  auditFilterFields,
  normalizeAuditFilters,
} from "../shared/auditReadValues";
import type { ListIncidentMembershipAuditEventsResponse } from "./api/publicHttpTypes";

export {
  type AuditField,
  type AuditInputs,
  auditFilterFields,
} from "../shared/auditReadValues";
export const membershipAuditActions = [
  "membership_created",
  "membership_deleted",
  "membership_role_changed",
] as const;
export const membershipAuditTargets = ["incident_membership"] as const;
export const membershipAuditLimit = 100;
export const membershipAuditHistoryLimit = 20;
export const membershipAuditDeadline = 30_000;
export const emptyMembershipAuditQuery: MembershipAuditQuery = {
  actor_user_id: "",
  action_code: "",
  target_kind: "",
  target_id: "",
  occurred_at_gte: "",
  occurred_at_lt: "",
};
export type MembershipAuditQuery = Omit<
  AuditInputs,
  "action_code" | "target_kind"
> & {
  readonly action_code: "" | (typeof membershipAuditActions)[number];
  readonly target_kind: "" | (typeof membershipAuditTargets)[number];
};
export type MembershipAuditEvent =
  ListIncidentMembershipAuditEventsResponse["data"]["audit_events"][number];
export type MembershipAuditPaging = NonNullable<
  ListIncidentMembershipAuditEventsResponse["meta"]["paging"]
>;
export type MembershipAuditAuthority = {
  readonly incidentId: string;
  readonly actorId: string;
  readonly lifetime: string;
  readonly apiBase?: string | undefined;
};
export type MembershipAuditPosition = {
  readonly cursor: string | null;
  readonly number: number;
};
export type MembershipAuditPage = {
  readonly binding: string;
  readonly query: MembershipAuditQuery;
  readonly position: MembershipAuditPosition;
  readonly rows: readonly MembershipAuditEvent[];
  readonly paging: MembershipAuditPaging;
};
export type MembershipAuditIntent = {
  readonly kind: "initial" | "refresh" | "replace" | "page" | "access";
  readonly query: MembershipAuditQuery;
  readonly position: MembershipAuditPosition;
  readonly history: readonly MembershipAuditPosition[];
};
export type MembershipAuditProblem = {
  readonly kind: "unavailable" | "query" | "access";
  readonly message: string;
};
export type MembershipAuditRead =
  | { readonly kind: "idle" }
  | { readonly kind: "pending"; readonly intent: MembershipAuditIntent }
  | {
      readonly kind: "failed";
      readonly intent: MembershipAuditIntent;
      readonly problem: MembershipAuditProblem;
    }
  | { readonly kind: "cursor_rejected" }
  | {
      readonly kind: "denied";
      readonly reason: "role" | "incident" | "session";
    };
export type MembershipAuditState = {
  readonly authority: MembershipAuditAuthority | null;
  readonly active: boolean;
  readonly inputs: AuditInputs;
  readonly applied: MembershipAuditQuery;
  readonly errors: AuditFieldErrors;
  readonly page: MembershipAuditPage | null;
  readonly history: readonly MembershipAuditPosition[];
  readonly read: MembershipAuditRead;
  readonly expandedId: string | null;
  readonly announcement: string;
  readonly announcementRole: "status" | "alert";
};
export function initialMembershipAuditState(): MembershipAuditState {
  return {
    authority: null,
    active: false,
    inputs: emptyMembershipAuditQuery,
    applied: emptyMembershipAuditQuery,
    errors: {},
    page: null,
    history: [],
    read: { kind: "idle" },
    expandedId: null,
    announcement: "",
    announcementRole: "status",
  };
}
export function validateMembershipAuditInputs(inputs: AuditInputs) {
  const { values, errors } = normalizeAuditFilters(inputs);
  const action = membershipAuditActions.find(
    (value) => value === values.action_code,
  );
  const target = membershipAuditTargets.find(
    (value) => value === values.target_kind,
  );
  if (values.action_code && !action)
    errors.action_code = "Choose a current membership action code.";
  if (values.target_kind && !target)
    errors.target_kind = "Choose incident_membership for a membership target.";
  const query: MembershipAuditQuery = {
    ...values,
    action_code: action ?? "",
    target_kind: target ?? "",
  };
  return { query, errors };
}
export function sameMembershipAuditQuery(a: AuditInputs, b: AuditInputs) {
  return auditFilterFields.every((field) => a[field] === b[field]);
}
export function membershipAuditBinding(
  authority: MembershipAuditAuthority,
  query: MembershipAuditQuery,
) {
  return JSON.stringify([
    authority.lifetime,
    authority.actorId,
    authority.incidentId,
    authority.apiBase ?? "",
    "listIncidentMembershipAuditEvents",
    membershipAuditLimit,
    "occurred_at_desc,audit_event_id_desc",
    ...auditFilterFields.map((field) => query[field]),
  ]);
}
export function membershipAuditPageSummary(page: MembershipAuditPage) {
  return `Page ${page.position.number}: ${page.rows.length} event${page.rows.length === 1 ? "" : "s"}. ${page.paging.has_more ? "More events available." : "No further events on this continuation."}`;
}
