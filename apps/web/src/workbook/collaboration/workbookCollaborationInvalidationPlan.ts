import type {
  WorkbookDependentPresentationInvalidationReason,
  WorkbookInvalidationReason,
  WorkbookMutationInvalidationReason,
  WorkbookQueryInvalidationReason,
} from "../lifecycle/workbookInvalidation";

export type WorkbookCollaborationInvalidationEffect =
  | {
      readonly kind: "mutation";
      readonly reason: WorkbookMutationInvalidationReason;
    }
  | {
      readonly kind: "query" | "active_surface";
      readonly reason: WorkbookQueryInvalidationReason;
    }
  | {
      readonly kind: "inspector" | "continuity" | "evidence";
      readonly reason: WorkbookDependentPresentationInvalidationReason;
    }
  | {
      readonly kind: "extension";
      readonly reason: Extract<
        WorkbookInvalidationReason,
        | { readonly kind: "session_unavailable" }
        | { readonly kind: "incident_access_lost" }
      >;
    }
  | { readonly kind: "presence" };

export function planWorkbookCollaborationInvalidation(
  reason: WorkbookInvalidationReason,
): readonly WorkbookCollaborationInvalidationEffect[] {
  switch (reason.kind) {
    case "session_unavailable":
    case "incident_access_lost":
      return [
        { kind: "mutation", reason },
        { kind: "query", reason },
        { kind: "active_surface", reason },
        { kind: "extension", reason },
        { kind: "presence" },
        { kind: "inspector", reason },
        { kind: "continuity", reason },
        { kind: "evidence", reason },
      ];
    case "incident_closed":
      return [
        { kind: "mutation", reason },
        { kind: "query", reason },
        { kind: "active_surface", reason },
        { kind: "presence" },
        { kind: "inspector", reason },
        { kind: "continuity", reason },
        { kind: "evidence", reason },
      ];
    case "collaboration_reset_required":
      return [
        { kind: "presence" },
        { kind: "query", reason },
        { kind: "active_surface", reason },
      ];
    case "incident_role_changed":
      return [
        { kind: "mutation", reason },
        { kind: "inspector", reason },
      ];
    case "runtime_disposed":
      return [
        { kind: "query", reason },
        { kind: "active_surface", reason },
        { kind: "inspector", reason },
        { kind: "continuity", reason },
        { kind: "evidence", reason },
      ];
    default:
      return [];
  }
}
