import { describe, expect, it } from "vitest";
import type { RecordChangedPayload } from "./workbookCollaborationMessages";
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
        message: { payload },
      },
    );

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
      createWorkbookSocketLifecycleState(),
      {
        type: "session_ack",
        messageType: "resume_ack",
        payload: {
          status: "reset_required",
        },
      },
    );

    expect(resetRequired.state.established).toBe(true);
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
    });
    const authorizationClose = reduceWorkbookSocketLifecycle(connected, {
      type: "authorization_closed",
    });

    expect(authorizationClose.state).toMatchObject({
      authPaused: true,
      established: false,
      reconnectSuppressed: true,
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
      },
    });
    expect(recovered.state).toMatchObject({
      authPaused: false,
      connectionId: "connection-2",
      established: true,
      reconnectSuppressed: false,
    });
    expect(recovered.effects).toEqual([{ kind: "resume_pending_replay" }]);
  });
});
