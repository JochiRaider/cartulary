import { describe, expect, it } from "vitest";
import {
  interpretNetworkFlowCollaborationEvent,
  interpretNetworkFlowCollaborationMessage,
} from "./networkFlowCollaborationInterpreter";

describe("interpretNetworkFlowCollaborationMessage", () => {
  it("preserves table and saved graph identities and rejects unrelated resources", () => {
    const envelope = {
      emitted_at: "2026-07-13T12:00:00Z",
      event_id: "event-1",
      incident_id: "incident-1",
    } as const;
    expect(
      interpretNetworkFlowCollaborationMessage({
        ...envelope,
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
      resourceKind: "network_flow_table",
      resourceId: "nft_a",
    });
    expect(
      interpretNetworkFlowCollaborationMessage({
        ...envelope,
        type: "extension_resource_changed",
        stream_seq: 2,
        payload: {
          extension_profile_id: "network_flow_activity",
          resource_kind: "network_flow_graph_view",
          resource_id: "nfgv_a",
          change_kind: "remove",
          reason_code: "soft_deleted",
        },
      }),
    ).toEqual({
      resourceKind: "network_flow_graph_view",
      resourceId: "nfgv_a",
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    expect(
      interpretNetworkFlowCollaborationMessage({
        ...envelope,
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

it("preserves normalized authorization scope in every Network Flow owner lifecycle", () => {
  for (const reasonCode of [
    "session_expired",
    "session_revoked",
    "concurrency_limit",
  ] as const) {
    expect(
      interpretNetworkFlowCollaborationEvent({
        kind: "authorization_revoked",
        incidentId: "incident",
        scope: "session",
        reasonCode,
      }),
    ).toMatchObject({
      changeKind: "remove",
      resourceKind: "*",
      reasonCode: "session_revoked",
    });
  }
  for (const event of [
    {
      kind: "authorization_revoked",
      incidentId: "incident",
      scope: "incident",
      reasonCode: "incident_access_revoked",
    },
    { kind: "authorization_lost" },
  ] as const) {
    expect(interpretNetworkFlowCollaborationEvent(event)).toMatchObject({
      changeKind: "remove",
      resourceKind: "*",
      reasonCode: "authorization_lost",
    });
  }
  expect(
    interpretNetworkFlowCollaborationEvent({ kind: "incident_closed" }),
  ).toMatchObject({ reasonCode: "incident_closed" });
  expect(
    interpretNetworkFlowCollaborationEvent({
      kind: "reset_required",
      generation: 1,
      reason: "sequence_gap",
    }),
  ).toBeNull();
});
