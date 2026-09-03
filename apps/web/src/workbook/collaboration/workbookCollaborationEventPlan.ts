import type { IncidentCollaborationEvent } from "../../collaboration/IncidentCollaborationSession";
import type { RecordChangedPayload } from "./workbookCollaborationMessages";

export type WorkbookCollaborationEventPlan =
  | { readonly kind: "established"; readonly mayResume: boolean }
  | {
      readonly eventGeneration: number;
      readonly kind: "reset";
      readonly reason: "resume_reset" | "sequence_gap";
    }
  | { readonly kind: "recover_authorization" }
  | { readonly kind: "incident_closed" }
  | { readonly kind: "presence_snapshot"; readonly payload: unknown }
  | { readonly kind: "presence_delta"; readonly payload: unknown }
  | { readonly kind: "record_changed"; readonly payload: RecordChangedPayload }
  | { readonly kind: "ignore" };

export function planWorkbookCollaborationEvent(
  event: IncidentCollaborationEvent,
): WorkbookCollaborationEventPlan {
  switch (event.kind) {
    case "established":
      return {
        kind: "established",
        mayResume: event.payload.status !== "reset_required",
      };
    case "reset_required":
      return {
        eventGeneration: event.generation,
        kind: "reset",
        reason: event.reason,
      };
    case "authorization_lost":
    case "session_revoked":
      return { kind: "recover_authorization" };
    case "incident_closed":
      return { kind: "incident_closed" };
    case "message":
      switch (event.message.type) {
        case "presence_snapshot":
          return { kind: "presence_snapshot", payload: event.message.payload };
        case "presence_delta":
          return { kind: "presence_delta", payload: event.message.payload };
        case "record_changed":
          return { kind: "record_changed", payload: event.message.payload };
        default:
          return { kind: "ignore" };
      }
  }
}
