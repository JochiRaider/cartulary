import { networkFlowActivityProfileId } from "./networkFlowClient";

export type NetworkFlowExtensionResourceChange = {
  readonly changeKind: "invalidate" | "remove";
  readonly reasonCode: string;
  readonly resourceId: string;
};

export type ExtensionResourceMessage = {
  readonly type?: string;
  readonly stream_seq?: number;
  readonly payload?: Readonly<Record<string, unknown>>;
};

// This pure interpreter is the only Network Flow component that understands
// the generic collaboration wire event. Reconnect lifecycle remains shared by
// the hook while feature code receives a closed domain event.
export function interpretNetworkFlowCollaborationMessage(
  message: ExtensionResourceMessage,
): NetworkFlowExtensionResourceChange | null {
  const payload = message.payload;
  if (
    message.type !== "extension_resource_changed" ||
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
