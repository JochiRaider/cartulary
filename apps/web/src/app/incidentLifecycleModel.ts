import type { IncidentResource } from "../shared/incidentResource";
import type { WorkbookIncidentRole } from "../shared/workbookShellContracts";
import type { CloseIncidentRequest } from "./api/publicHttpTypes";

export type LifecycleAction = "close" | "reopen";
export type LifecycleAuthority = Readonly<{
  incidentId: string;
  actorId: string;
  lifetime: string;
  role: WorkbookIncidentRole;
  apiBase?: string | undefined;
}>;
export const lifecycleBinding = (a: LifecycleAuthority) =>
  JSON.stringify([a.incidentId, a.actorId, a.lifetime, a.apiBase ?? ""]);
export type LifecycleDraft = Readonly<{
  reason: string;
  action: LifecycleAction | null;
  revision: number;
}>;
export type LifecycleReview = Readonly<{
  resource: IncidentResource;
  revision: number;
  role: WorkbookIncidentRole;
  action: LifecycleAction;
}>;
export type LifecycleAttempt = Readonly<{
  id: number;
  authority: LifecycleAuthority;
  action: LifecycleAction;
  resource: IncidentResource;
  revision: number;
  payload: Readonly<CloseIncidentRequest>;
}>;
export const lifecycleProblemCodes = [
  "invalid_incident_lifecycle_request",
  "incident_version_conflict",
  "illegal_transition",
  "client_txn_conflict",
  "session_required",
  "authorization_denied",
  "csrf_verification_failed",
  "incident_not_found",
  "credential_bootstrap_rejected",
  "internal_error",
  "invalid_public_contract_response",
  "unknown_public_error",
] as const;
export const lifecycleReasons = [
  "request_not_object",
  "missing_required_field",
  "field_not_nullable",
  "invalid_base_incident_version",
  "invalid_client_txn_id",
  "invalid_reason",
  "reason_empty_after_normalization",
  "reason_too_long",
  "control_character_not_allowed",
  "unknown_field",
  "incident_already_closed",
  "incident_not_closed",
] as const;
export type LifecycleProblem = Readonly<{
  code: (typeof lifecycleProblemCodes)[number];
  field?: "reason" | "base_incident_version" | "client_txn_id" | undefined;
  reason?: (typeof lifecycleReasons)[number] | undefined;
}>;
export type LifecycleOutcome =
  | { readonly kind: "idle" }
  | {
      readonly kind: "pending";
      readonly attempt: LifecycleAttempt;
      readonly replay: boolean;
      readonly stage: "authorization" | "write";
    }
  | {
      readonly kind: "uncertain";
      readonly attempt: LifecycleAttempt;
      readonly problem?: LifecycleProblem | undefined;
    }
  | {
      readonly kind: "rejected";
      readonly attempt: LifecycleAttempt;
      readonly problem: LifecycleProblem;
    }
  | {
      readonly kind: "confirmed";
      readonly attempt: LifecycleAttempt;
      readonly receipt: IncidentResource;
      readonly replay: boolean;
    };
export type LifecycleState = Readonly<{
  authority: LifecycleAuthority | null;
  active: boolean;
  access: "checking" | "ready" | "unavailable";
  resource: IncidentResource | null;
  read: "idle" | "loading" | "refreshing" | "ready" | "failed";
  draft: LifecycleDraft;
  review: LifecycleReview | null;
  operation: LifecycleOutcome;
  transportPending: boolean;
  fieldError: { readonly revision: number; readonly message: string } | null;
  notice: string;
  departure: boolean;
}>;
export const initialLifecycleState = (): LifecycleState => ({
  authority: null,
  active: false,
  access: "checking",
  resource: null,
  read: "idle",
  draft: { reason: "", action: null, revision: 0 },
  review: null,
  operation: { kind: "idle" },
  transportPending: false,
  fieldError: null,
  notice: "",
  departure: false,
});
export const lifecycleAllowed = (
  action: LifecycleAction,
  resource: IncidentResource,
) =>
  action === "close"
    ? resource.status === "active"
    : resource.status === "closed";
export function lifecycleProblemMessage(problem: LifecycleProblem): string {
  switch (problem.code) {
    case "invalid_incident_lifecycle_request":
      switch (problem.reason) {
        case "reason_too_long":
          return "Use no more than 4096 Unicode characters after normalization.";
        case "control_character_not_allowed":
          return "Remove unsupported control characters from the reason.";
        case "reason_empty_after_normalization":
          return "Enter a reason containing more than whitespace.";
        default:
          return problem.field === "reason"
            ? "Enter a valid reason for this action."
            : "The lifecycle request was rejected as invalid. Refresh and review before trying again.";
      }
    case "incident_version_conflict":
      return "The incident version changed. Review current incident state before a new action.";
    case "illegal_transition":
      return problem.reason === "incident_already_closed"
        ? "Close was rejected because the incident was already closed."
        : problem.reason === "incident_not_closed"
          ? "Reopen was rejected because the incident was not closed."
          : "The requested transition does not match the incident's current state. Refresh to review it.";
    case "client_txn_conflict":
      return "Transaction conflict: this action key is associated with a different request. Refresh current state and explicitly forget this recovery before reviewing a new action.";
    case "authorization_denied":
      return "Administrator authorization was denied. Check current access before another action.";
    case "csrf_verification_failed":
      return "Session verification failed. Check current access before another action.";
    case "credential_bootstrap_rejected":
      return "A full authenticated session is required for this action.";
    default:
      return "The action result could not be confirmed. Keep its recovery record and check current access.";
  }
}
