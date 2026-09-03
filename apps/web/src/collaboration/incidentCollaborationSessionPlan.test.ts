import type { IncidentStreamMessage } from "@cartulary/protocol-ts/collaboration";
import { describe, expect, it } from "vitest";
import { planIncidentCollaborationSessionMessage } from "./incidentCollaborationSessionPlan";

function replayable(streamSeq: number): IncidentStreamMessage {
  return {
    actor_user_id: "user-2",
    affected_views: undefined,
    emitted_at: "2026-09-03T13:00:00Z",
    event_id: `event-${streamSeq}`,
    incident_id: "incident-1",
    payload: {
      actor_user_id: "user-2",
      affected_views: [
        {
          change_kind: "invalidate",
          view_schema_id: "cartulary.view.timeline.v2",
        },
      ],
      change_set_id: "change-1",
      changed_field_keys: [],
      client_txn_id: "txn-1",
      record_id: "record-1",
      row_version: 2,
    },
    stream_seq: streamSeq,
    type: "record_changed",
  } as IncidentStreamMessage;
}

describe("incident collaboration session plan", () => {
  it("classifies handshake, heartbeat, and terminal messages exactly", () => {
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 4,
        message: {
          emitted_at: "2026-09-03T13:00:00Z",
          event_id: "event-1",
          incident_id: "incident-1",
          payload: {},
          type: "ping",
        } as IncidentStreamMessage,
        resetting: false,
      }),
    ).toEqual({ kind: "pong", nextStreamSeq: 4 });
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 4,
        message: {
          emitted_at: "2026-09-03T13:00:00Z",
          event_id: "event-2",
          incident_id: "incident-1",
          payload: {
            resume_token: "opaque",
            server_high_water_stream_seq: 8,
            status: "reset_required",
          },
          type: "resume_ack",
        } as IncidentStreamMessage,
        resetting: false,
      }),
    ).toMatchObject({
      kind: "established",
      nextStreamSeq: 8,
      resetRequired: true,
      resumeToken: "opaque",
    });
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 8,
        message: {
          emitted_at: "2026-09-03T13:00:00Z",
          event_id: "event-3",
          incident_id: "incident-1",
          payload: { reason_code: "session_revoked" },
          type: "session_revoked",
        } as IncidentStreamMessage,
        resetting: false,
      }),
    ).toEqual({
      kind: "terminate",
      nextStreamSeq: 8,
      reason: "session_revoked",
    });
  });

  it("deduplicates, detects gaps, and suppresses replay while resetting", () => {
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 3,
        message: replayable(3),
        resetting: false,
      }).kind,
    ).toBe("ignore");
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 3,
        message: replayable(5),
        resetting: false,
      }),
    ).toMatchObject({ kind: "reset", nextStreamSeq: 5 });
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 5,
        message: replayable(6),
        resetting: true,
      }),
    ).toMatchObject({ kind: "ignore", nextStreamSeq: 6 });
    expect(
      planIncidentCollaborationSessionMessage({
        lastSeenStreamSeq: 6,
        message: replayable(7),
        resetting: false,
      }),
    ).toMatchObject({ kind: "message", nextStreamSeq: 7 });
  });
});
