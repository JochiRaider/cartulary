import { describe, expect, it } from "vitest";
import type { RecordChangedPayload } from "./workbookShellPhase4";
import {
  createWorkbookSocketLifecycleState,
  reduceWorkbookSocketLifecycle,
} from "./workbookSocketLifecycle";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

function recordChangedPayload(
  options: {
    readonly affectedViews?: RecordChangedPayload["affected_views"];
    readonly clientTxnId?: string;
    readonly recordId?: string;
    readonly rowVersion?: number;
  } = {},
): RecordChangedPayload {
  const recordId = options.recordId ?? "record-1";
  const rowVersion = options.rowVersion ?? 2;
  return {
    record_id: recordId,
    row_version: rowVersion,
    change_set_id: `change-set-${rowVersion}`,
    client_txn_id: options.clientTxnId ?? `client-txn-${rowVersion}`,
    actor_user_id: "user-1",
    changed_field_keys: ["timeline.activity_synopsis_text"],
    affected_views: options.affectedViews ?? [
      {
        view_schema_id: timelineViewSchemaId,
        change_kind: "patch",
        patch_cells: {
          record_id: recordId,
          row_version: rowVersion,
          cells: {
            "timeline.activity_synopsis_text": { value: "Remote summary" },
          },
        },
      },
    ],
  };
}

describe("Sprint 2 WebSocket reducer lifecycle coverage", () => {
  it("FE-U-P7-01 applies row update and invalidate lifecycle decisions without visible-index state", () => {
    const payload = recordChangedPayload();
    const rowUpdate = reduceWorkbookSocketLifecycle(
      createWorkbookSocketLifecycleState(),
      {
        type: "record_changed_received",
        message: { stream_seq: 1, payload },
      },
    );

    expect(rowUpdate.state.lastSeenStreamSeq).toBe(1);
    expect(rowUpdate.state.appliedStreamSeqs.has(1)).toBe(true);
    expect(rowUpdate.effects).toEqual([
      { kind: "apply_record_change", payload },
    ]);
    expect(JSON.stringify(rowUpdate.effects)).not.toContain("visible");

    const invalidate = reduceWorkbookSocketLifecycle(rowUpdate.state, {
      type: "record_change_result",
      applied: false,
    });

    expect(invalidate.effects).toEqual([
      { kind: "request_refresh", reason: "record_change_requery" },
    ]);
  });

  it("FE-U-P7-01 requests reset and stale-row requery for reset-required and stale record changes", () => {
    const resetRequired = reduceWorkbookSocketLifecycle(
      createWorkbookSocketLifecycleState({
        resumeToken: "resume-before-reset",
      }),
      {
        type: "session_ack",
        messageType: "resume_ack",
        payload: {
          resume_token: "resume-after-reset",
          status: "reset_required",
        },
      },
    );

    expect(resetRequired.state.established).toBe(true);
    expect(resetRequired.state.resumeToken).toBe("resume-after-reset");
    expect(resetRequired.effects).toEqual([
      { kind: "request_refresh", reason: "reset_required" },
    ]);

    const staleRequery = reduceWorkbookSocketLifecycle(resetRequired.state, {
      type: "record_change_result",
      applied: false,
    });
    expect(staleRequery.effects).toEqual([
      { kind: "request_refresh", reason: "record_change_requery" },
    ]);
  });

  it("FE-U-P7-01 pauses replay and suppresses reconnect after authorization close or session revocation", () => {
    const connected = createWorkbookSocketLifecycleState({
      connectionId: "connection-1",
      established: true,
      resumeToken: "resume-live",
    });
    const authorizationClose = reduceWorkbookSocketLifecycle(connected, {
      type: "authorization_closed",
    });

    expect(authorizationClose.state).toMatchObject({
      authPaused: true,
      established: false,
      reconnectSuppressed: true,
      resumeToken: null,
    });
    expect(authorizationClose.effects).toEqual([
      { kind: "pause_for_auth_recovery" },
      { kind: "suppress_reconnect" },
      { kind: "schedule_auth_recovery_probe" },
      { kind: "close_socket" },
    ]);

    const sessionRevoked = reduceWorkbookSocketLifecycle(connected, {
      type: "session_revoked",
    });
    expect(sessionRevoked.state.reconnectSuppressed).toBe(true);
    expect(sessionRevoked.effects).toContainEqual({
      kind: "pause_for_auth_recovery",
    });

    const recovered = reduceWorkbookSocketLifecycle(sessionRevoked.state, {
      type: "session_ack",
      messageType: "hello_ack",
      payload: {
        connection_id: "connection-2",
        resume_token: "resume-recovered",
      },
    });
    expect(recovered.state).toMatchObject({
      authPaused: false,
      connectionId: "connection-2",
      established: true,
      reconnectSuppressed: false,
      resumeToken: "resume-recovered",
    });
    expect(recovered.effects).toEqual([{ kind: "resume_pending_replay" }]);
  });

  it("FE-U-P7-01 ignores duplicate stream sequences and refreshes on sequence gaps", () => {
    const state = createWorkbookSocketLifecycleState({
      appliedStreamSeqs: new Set([4]),
      lastSeenStreamSeq: 4,
    });
    const duplicate = reduceWorkbookSocketLifecycle(state, {
      type: "record_changed_received",
      message: {
        stream_seq: 4,
        payload: recordChangedPayload({ rowVersion: 4 }),
      },
    });

    expect(duplicate.state).toBe(state);
    expect(duplicate.effects).toEqual([
      { kind: "ignore_duplicate_sequence", streamSeq: 4 },
    ]);

    const gap = reduceWorkbookSocketLifecycle(state, {
      type: "record_changed_received",
      message: {
        stream_seq: 6,
        payload: recordChangedPayload({ rowVersion: 6 }),
      },
    });

    expect(gap.state.lastSeenStreamSeq).toBe(6);
    expect(gap.state.appliedStreamSeqs.has(6)).toBe(true);
    expect(gap.effects).toEqual([
      { kind: "request_refresh", reason: "sequence_gap", streamSeq: 6 },
    ]);

    const next = reduceWorkbookSocketLifecycle(gap.state, {
      type: "record_changed_received",
      message: {
        stream_seq: 7,
        payload: recordChangedPayload({ rowVersion: 7 }),
      },
    });
    expect(next.effects[0]?.kind).toBe("apply_record_change");
  });
});
