import { describe, expect, it } from "vitest";
import { interpretNetworkFlowCollaborationMessage } from "./networkFlowCollaborationInterpreter";

describe("interpretNetworkFlowCollaborationMessage", () => {
  it("admits only closed Network Flow table changes", () => {
    expect(
      interpretNetworkFlowCollaborationMessage({
        type: "extension_resource_changed",
        stream_seq: 1,
        payload: {
          extension_profile_id: "network_flow_activity",
          resource_kind: "network_flow_table",
          resource_id: "nft_a",
          change_kind: "invalidate",
          reason_code: "renamed",
        },
      }),
    ).toEqual({
      changeKind: "invalidate",
      reasonCode: "renamed",
      resourceId: "nft_a",
    });
    expect(
      interpretNetworkFlowCollaborationMessage({
        type: "extension_resource_changed",
        stream_seq: 1,
        payload: {
          extension_profile_id: "another_profile",
          resource_kind: "network_flow_table",
          resource_id: "nft_a",
          change_kind: "remove",
          reason_code: "deleted",
        },
      }),
    ).toBeNull();
  });
});
