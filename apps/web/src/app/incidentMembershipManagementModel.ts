import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import type {
  CreateIncidentMembershipRequest,
  DeleteIncidentMembershipRequest,
  ListIncidentMembershipsResponse,
  PatchIncidentMembershipRequest,
} from "./api/publicHttpTypes";

export type Membership =
  ListIncidentMembershipsResponse["data"]["memberships"][number];
export type MembershipRole = Membership["role"];
export type MembershipPaging = NonNullable<
  ListIncidentMembershipsResponse["meta"]["paging"]
>;
export type MembershipAuthority = {
  readonly incidentId: string;
  readonly actorId: string;
  readonly lifetime: string;
  readonly apiBase?: string | undefined;
  readonly role: WorkbookIncidentRole;
};
export const membershipPageLimit = 100;
export const membershipHistoryLimit = 20;
export const membershipRoles = [
  "viewer",
  "editor",
  "reviewer",
  "admin",
] as const satisfies readonly MembershipRole[];
export type MembershipPosition = {
  readonly cursor: string | null;
  readonly number: number;
};
export type MembershipPage = {
  readonly rows: readonly Membership[];
  readonly paging: MembershipPaging;
  readonly position: MembershipPosition;
};
export type MembershipEditor = {
  readonly id: number;
  readonly revision: number;
  readonly kind: "add" | "role" | "remove";
  readonly base: Membership | null;
  readonly email: string;
  readonly role: MembershipRole;
  readonly reviewRequired: boolean;
};
export type MembershipInput =
  | {
      readonly kind: "add";
      readonly payload: Extract<
        CreateIncidentMembershipRequest,
        { email: string }
      >;
    }
  | {
      readonly kind: "role";
      readonly userId: string;
      readonly payload: PatchIncidentMembershipRequest;
    }
  | {
      readonly kind: "remove";
      readonly userId: string;
      readonly payload: DeleteIncidentMembershipRequest;
    };
export type MembershipAttempt = {
  readonly id: number;
  readonly authority: MembershipAuthority;
  readonly input: MembershipInput;
  readonly editorId: number;
  readonly draftRevision: number;
  readonly base: Membership | null;
  readonly label: string;
};
export const membershipProblemCodes = [
  "user_not_found",
  "user_inactive",
  "membership_exists_use_patch",
  "last_incident_admin",
  "membership_not_found",
  "membership_version_conflict",
  "client_txn_conflict",
  "authorization_denied",
  "invalid_mutation_payload",
  "csrf_verification_failed",
  "session_required",
  "incident_not_found",
  "invalid_pagination_request",
  "service_unavailable",
  "internal_error",
  "invalid_public_contract_response",
  "transport_uncertain",
  "access_unavailable",
  "unknown_public_error",
] as const;
export type MembershipProblem = {
  readonly code: (typeof membershipProblemCodes)[number];
  readonly field?: string | undefined;
};
export type MembershipRefresh = "idle" | "pending" | "failed" | "done";
export type MembershipOperation =
  | { readonly kind: "idle" }
  | {
      readonly kind: "pending";
      readonly attempt: MembershipAttempt;
      readonly stage: "authorization" | "write";
    }
  | {
      readonly kind: "confirmed";
      readonly attempt: MembershipAttempt;
      readonly status: 200 | 201 | 204;
      readonly receipt: Membership | null;
      readonly listRefresh: MembershipRefresh;
      readonly accessRefresh: MembershipRefresh;
    }
  | {
      readonly kind: "rejected" | "conflicted" | "uncertain";
      readonly attempt: MembershipAttempt;
      readonly problem: MembershipProblem;
      readonly observed: Membership | null;
    }
  | { readonly kind: "retired"; readonly message: string };
export type MembershipRead =
  | { readonly kind: "idle" }
  | { readonly kind: "pending"; readonly position: MembershipPosition }
  | {
      readonly kind: "failed";
      readonly position: MembershipPosition;
      readonly history: readonly MembershipPosition[];
      readonly reason: "read" | "access";
    }
  | { readonly kind: "cursor_rejected" };
export type MembershipDeparture =
  | { readonly kind: "editor"; readonly next: MembershipEditor | null }
  | { readonly kind: "route" };
export type MembershipManagementState = {
  readonly authority: MembershipAuthority | null;
  readonly active: boolean;
  readonly access: "ready" | "checking" | "unavailable";
  readonly page: MembershipPage | null;
  readonly history: readonly MembershipPosition[];
  readonly read: MembershipRead;
  readonly editor: MembershipEditor | null;
  readonly operation: MembershipOperation;
  readonly fieldError: string | null;
  readonly departure: MembershipDeparture | null;
  readonly transportPending: boolean;
};
export function initialMembershipManagementState(): MembershipManagementState {
  return {
    authority: null,
    active: false,
    access: "checking",
    page: null,
    history: [],
    read: { kind: "idle" },
    editor: null,
    operation: { kind: "idle" },
    fieldError: null,
    departure: null,
    transportPending: false,
  };
}
export function membershipBinding(a: MembershipAuthority) {
  return JSON.stringify([a.incidentId, a.actorId, a.lifetime, a.apiBase ?? ""]);
}
export function membershipEditorDirty(editor: MembershipEditor | null) {
  return (
    editor !== null &&
    (editor.kind === "add"
      ? editor.email !== "" || editor.role !== "viewer"
      : editor.kind === "remove" ||
        editor.role !== editor.base?.role ||
        editor.reviewRequired)
  );
}
export function membershipProblemText(problem: MembershipProblem): string {
  switch (problem.code) {
    case "user_not_found":
      return "No existing account has this email address. Check the address; this action does not create an account or send an invitation.";
    case "user_inactive":
      return "This is an inactive account. An active existing account is required.";
    case "membership_exists_use_patch":
      return "This account already belongs to the incident with a different role. Find the member and explicitly change their role.";
    case "last_incident_admin":
      return "The incident must retain an administrator. Another current administrator is required before this change.";
    case "membership_not_found":
      return "This membership no longer exists. No membership was changed by this request.";
    case "membership_version_conflict":
      return "The membership changed after it was reviewed. Observe current membership and review a new action.";
    case "client_txn_conflict":
      return "The request identity was rejected. Review the original action before starting a new request.";
    case "authorization_denied":
      return "Current incident administrator access is required.";
    case "invalid_mutation_payload":
      return "The membership request was rejected. Review the email, role, and current membership before trying again.";
    case "csrf_verification_failed":
      return "The request was rejected. Check current access before trying again.";
    case "access_unavailable":
      return "Current access could not be checked. No write was sent. Check access before trying again.";
    default:
      return "The operation outcome is unavailable. Review its recovery options.";
  }
}
export function membershipConfirmedText(
  operation: Extract<MembershipOperation, { kind: "confirmed" }>,
) {
  if (operation.attempt.input.kind === "remove")
    return "Incident membership removed.";
  if (operation.attempt.input.kind === "add")
    return operation.status === 201
      ? "Membership added."
      : "Membership confirmed: existing membership or exact request replay.";
  return operation.receipt?.membership_version ===
    operation.attempt.base?.membership_version
    ? "Membership role already matched; no change was needed."
    : "Membership role saved.";
}
