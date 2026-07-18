import { describe, expect, it } from "vitest";
import {
  createTimelineCollaborationState,
  reduceTimelineCollaboration,
} from "./timelineCollaborationEffects";
import type { RecordChangedPayload } from "./workbookCollaborationMessages";

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

describe("Timeline collaboration effects", () => {
  it("applies row updates and requests stale-row requery without transport state", () => {
    const payload = recordChangedPayload();
    const rowUpdate = reduceTimelineCollaboration(
      createTimelineCollaborationState(),
      { type: "record_changed_received", payload },
    );

    expect(rowUpdate.effects).toEqual([
      { kind: "apply_record_change", payload },
    ]);
    expect(JSON.stringify(rowUpdate)).not.toMatch(
      /connection|reconnect|resume_token|socket/u,
    );

    const invalidate = reduceTimelineCollaboration(rowUpdate.state, {
      type: "record_change_result",
      applied: false,
    });
    expect(invalidate.effects).toEqual([{ kind: "request_record_refresh" }]);
  });

  it("pauses feature replay on access loss and resumes only after session establishment", () => {
    const authorizationLoss = reduceTimelineCollaboration(
      createTimelineCollaborationState(),
      { type: "authorization_lost" },
    );
    expect(authorizationLoss.state).toEqual({ authPaused: true });
    expect(authorizationLoss.effects).toEqual([
      { kind: "pause_for_auth_recovery" },
      { kind: "schedule_auth_recovery_probe" },
    ]);

    const recovered = reduceTimelineCollaboration(authorizationLoss.state, {
      type: "session_established",
    });
    expect(recovered.state).toEqual({ authPaused: false });
    expect(recovered.effects).toEqual([{ kind: "resume_pending_replay" }]);
  });
});
