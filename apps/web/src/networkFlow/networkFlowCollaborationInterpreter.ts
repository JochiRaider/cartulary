import type { IncidentCollaborationMessage } from "../collaboration/IncidentCollaborationSession";
import { networkFlowActivityProfileId } from "./networkFlowClient";

export type NetworkFlowExtensionResourceChange = {
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
    payload.resource_kind !== "network_flow_table" ||
    typeof payload.resource_id !== "string" ||
    (payload.change_kind !== "invalidate" &&
      payload.change_kind !== "remove") ||
    typeof payload.reason_code !== "string"
  ) {
    return null;
  }
  return {
    changeKind: payload.change_kind,
    reasonCode: payload.reason_code,
    resourceId: payload.resource_id,
  };
}
