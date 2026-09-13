import type {
  IncidentCollaborationEvent,
  IncidentCollaborationMessage,
} from "../collaboration/IncidentCollaborationSession";
import { networkFlowActivityProfileId } from "./networkFlowClient";

export type NetworkFlowExtensionResourceChange = {
  readonly resourceKind: "network_flow_table" | "network_flow_graph_view" | "*";
  readonly changeKind: "invalidate" | "remove";
  readonly reasonCode: string;
  readonly resourceId: string;
};

// This pure interpreter is the only Network Flow component that understands
// the generic collaboration wire event. The incident session owns reconnect,
// resume, decoding, and sequence state while feature code receives a closed
// domain event.
export function interpretNetworkFlowCollaborationMessage(
  message: IncidentCollaborationMessage,
): NetworkFlowExtensionResourceChange | null {
  if (message.type !== "extension_resource_changed") {
    return null;
  }
  const payload = message.payload;
  if (
    payload?.extension_profile_id !== networkFlowActivityProfileId ||
    (payload.resource_kind !== "network_flow_table" &&
      payload.resource_kind !== "network_flow_graph_view") ||
    typeof payload.resource_id !== "string" ||
    (payload.change_kind !== "invalidate" &&
      payload.change_kind !== "remove") ||
    typeof payload.reason_code !== "string"
  ) {
    return null;
  }
  return {
    resourceKind: payload.resource_kind,
    changeKind: payload.change_kind,
    reasonCode: payload.reason_code,
    resourceId: payload.resource_id,
  };
}

// Translate the normalized authorization scope once into existing owner lifecycle
// effects. Consumers never infer account-session loss from a wire reason string.
export function interpretNetworkFlowCollaborationEvent(
  event: IncidentCollaborationEvent,
): NetworkFlowExtensionResourceChange | null {
  if (event.kind === "message")
    return interpretNetworkFlowCollaborationMessage(event.message);
  if (
    event.kind !== "authorization_revoked" &&
    event.kind !== "authorization_lost" &&
    event.kind !== "incident_closed"
  )
    return null;
  return {
    resourceKind: "*",
    resourceId: "*",
    changeKind: "remove",
    reasonCode:
      event.kind === "authorization_revoked"
        ? event.scope === "session"
          ? "session_revoked"
          : "authorization_lost"
        : event.kind,
  };
}
