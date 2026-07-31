import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export type ExtensionUnavailabilityReason =
  | "missing_availability"
  | "malformed_availability"
  | "support_mismatch"
  | "secure_random_unavailable"
  | "stale_generation"
  | "authorization_lost"
  | "capability_activation_failed";

export type WorkbookInvalidationReason =
  | { readonly kind: "startup_superseded" }
  | { readonly kind: "session_unavailable" }
  | { readonly kind: "incident_access_lost" }
  | {
      readonly kind: "incident_role_changed";
      readonly role: WorkbookIncidentRole;
    }
  | { readonly kind: "incident_closed" }
  | {
      readonly kind: "extension_unavailable";
      readonly extensionProfileId: string;
      readonly reason: ExtensionUnavailabilityReason;
    }
  | { readonly kind: "collaboration_reset_required" }
  | { readonly kind: "incident_changed"; readonly nextIncidentId: string }
  | { readonly kind: "runtime_disposed" };

export type WorkbookQueryInvalidationReason = Extract<
  WorkbookInvalidationReason,
  | { readonly kind: "session_unavailable" }
  | { readonly kind: "incident_access_lost" }
  | { readonly kind: "collaboration_reset_required" }
  | { readonly kind: "incident_closed" }
  | { readonly kind: "incident_changed" }
  | { readonly kind: "runtime_disposed" }
>;

export type WorkbookMutationInvalidationReason = Extract<
  WorkbookInvalidationReason,
  | { readonly kind: "session_unavailable" }
  | { readonly kind: "incident_access_lost" }
  | { readonly kind: "incident_role_changed" }
  | { readonly kind: "incident_closed" }
  | { readonly kind: "incident_changed" }
  | { readonly kind: "runtime_disposed" }
>;

export type WorkbookDependentPresentationInvalidationReason = Extract<
  WorkbookInvalidationReason,
  | { readonly kind: "session_unavailable" }
  | { readonly kind: "incident_access_lost" }
  | { readonly kind: "incident_role_changed" }
  | { readonly kind: "incident_closed" }
  | { readonly kind: "incident_changed" }
  | { readonly kind: "runtime_disposed" }
>;
