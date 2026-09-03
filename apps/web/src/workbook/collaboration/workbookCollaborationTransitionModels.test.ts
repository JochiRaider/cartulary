import { describe, expect, it } from "vitest";
import type { IncidentCollaborationEvent } from "../../collaboration/IncidentCollaborationSession";
import {
  beginWorkbookAuthorizationRecovery,
  completeWorkbookAuthorizationRecovery,
  initialWorkbookAuthorizationRecoveryMachine,
  planWorkbookAuthorizationRecoveryResult,
  scheduleWorkbookAuthorizationRecovery,
  terminateWorkbookAuthorizationRecovery,
} from "./workbookAuthorizationRecoveryMachine";
import { planWorkbookCollaborationEvent } from "./workbookCollaborationEventPlan";
import { planWorkbookCollaborationInvalidation } from "./workbookCollaborationInvalidationPlan";
import {
  beginWorkbookCollaborationReset,
  cancelWorkbookCollaborationReset,
  initialWorkbookCollaborationResetMachine,
  workbookCollaborationResetIsCurrent,
} from "./workbookCollaborationResetMachine";
import {
  activeWorkbookPresenceRecords,
  applyWorkbookPresenceDelta,
  initialWorkbookPresenceProjection,
  replaceWorkbookPresenceSnapshot,
} from "./workbookPresenceProjection";
import {
  initialWorkbookPresencePublicationMachine,
  scheduleWorkbookPresencePublication,
  settleWorkbookPresencePublication,
} from "./workbookPresencePublicationMachine";

function presence(connectionId: string, displayName = connectionId) {
  return {
    connection_id: connectionId,
    display_name: displayName,
    expires_at: "2026-09-03T14:00:00Z",
    mode: "viewing",
    observed_at: "2026-09-03T13:59:00Z",
    sheet_ref: { kind: "view_schema", id: "timeline" },
    user_id: `user-${connectionId}`,
  };
}

describe("Workbook collaboration transition models", () => {
  it("routes the closed event family without interpreting ignored message owners", () => {
    expect(
      planWorkbookCollaborationEvent({
        kind: "established",
        messageType: "resume_ack",
        payload: { status: "reset_required" },
      }),
    ).toEqual({ kind: "established", mayResume: false });
    expect(
      planWorkbookCollaborationEvent({
        generation: 4,
        kind: "reset_required",
        reason: "sequence_gap",
      }),
    ).toEqual({ eventGeneration: 4, kind: "reset", reason: "sequence_gap" });
    expect(planWorkbookCollaborationEvent({ kind: "session_revoked" })).toEqual(
      { kind: "recover_authorization" },
    );
    expect(
      planWorkbookCollaborationEvent({
        kind: "message",
        message: {
          emitted_at: "2026-09-03T13:00:00Z",
          event_id: "event-1",
          incident_id: "incident-1",
          payload: {},
          type: "ping",
        },
      } as IncidentCollaborationEvent),
    ).toEqual({ kind: "ignore" });
  });

  it("plans ordered protected-state, reset, role, closure, and disposal invalidation", () => {
    expect(
      planWorkbookCollaborationInvalidation({
        kind: "session_unavailable",
      }).map((effect) => effect.kind),
    ).toEqual([
      "mutation",
      "query",
      "active_surface",
      "extension",
      "presence",
      "inspector",
      "continuity",
      "evidence",
    ]);
    expect(
      planWorkbookCollaborationInvalidation({
        kind: "collaboration_reset_required",
      }).map((effect) => effect.kind),
    ).toEqual(["presence", "query", "active_surface"]);
    expect(
      planWorkbookCollaborationInvalidation({
        kind: "incident_role_changed",
        role: "viewer",
      }).map((effect) => effect.kind),
    ).toEqual(["mutation", "inspector"]);
    expect(
      planWorkbookCollaborationInvalidation({ kind: "incident_closed" }).map(
        (effect) => effect.kind,
      ),
    ).toEqual([
      "mutation",
      "query",
      "active_surface",
      "presence",
      "inspector",
      "continuity",
      "evidence",
    ]);
    expect(
      planWorkbookCollaborationInvalidation({ kind: "runtime_disposed" }).map(
        (effect) => effect.kind,
      ),
    ).toEqual([
      "query",
      "active_surface",
      "inspector",
      "continuity",
      "evidence",
    ]);
  });

  it("projects presence as an exact keyed collection and rejects malformed snapshots", () => {
    const initial = initialWorkbookPresenceProjection();
    const snapshot = replaceWorkbookPresenceSnapshot(initial, {
      presences: [
        presence("self", "Self"),
        presence("b", "Beta"),
        presence("a", "Alpha"),
      ],
    });
    expect(
      activeWorkbookPresenceRecords({
        activeSheetRef: { kind: "view_schema", id: "timeline" },
        connectionId: "self",
        projection: snapshot,
      }).map((entry) => entry.connection_id),
    ).toEqual(["a", "b"]);
    expect(
      replaceWorkbookPresenceSnapshot(snapshot, {
        presences: [presence("duplicate"), presence("duplicate")],
      }),
    ).toBe(snapshot);
    const removed = applyWorkbookPresenceDelta(snapshot, {
      delta_kind: "remove",
      presence: { connection_id: "b" },
    });
    expect(removed.has("b")).toBe(false);
    expect(
      applyWorkbookPresenceDelta(removed, {
        delta_kind: "future",
        presence: presence("future"),
      }),
    ).toBe(removed);
  });

  it("settles only the current reset subject", () => {
    const started = beginWorkbookCollaborationReset(
      initialWorkbookCollaborationResetMachine(),
      { eventGeneration: 3, sessionGeneration: 2, sheetKey: "view:timeline" },
    );
    expect(
      workbookCollaborationResetIsCurrent(started.machine, started.admission, {
        sessionGeneration: 2,
        sheetKey: "view:timeline",
      }),
    ).toBe(true);
    expect(
      workbookCollaborationResetIsCurrent(started.machine, started.admission, {
        sessionGeneration: 3,
        sheetKey: "view:timeline",
      }),
    ).toBe(false);
    expect(
      workbookCollaborationResetIsCurrent(
        cancelWorkbookCollaborationReset(started.machine),
        started.admission,
        { sessionGeneration: 2, sheetKey: "view:timeline" },
      ),
    ).toBe(false);
  });

  it("rejects stale recovery settlement and caps coalesced presence publication", () => {
    const scheduled = scheduleWorkbookAuthorizationRecovery(
      initialWorkbookAuthorizationRecoveryMachine(),
      100,
    );
    expect(scheduled.scheduledForMs).toBe(1_100);
    const begun = beginWorkbookAuthorizationRecovery(
      scheduled,
      scheduled.generation,
    );
    if (begun.kind !== "recover") throw new Error("expected recovery");
    const authorized = planWorkbookAuthorizationRecoveryResult(
      begun.machine,
      begun.admission,
      { kind: "authorized", role: "editor", userId: "user-1" },
      1_100,
    );
    if (authorized.kind !== "authorized") {
      throw new Error("expected authorized plan");
    }
    expect(
      completeWorkbookAuthorizationRecovery(
        authorized.machine,
        authorized.admission,
      ),
    ).toMatchObject({ canResumeMutations: true, kind: "complete" });
    expect(
      planWorkbookAuthorizationRecoveryResult(
        terminateWorkbookAuthorizationRecovery(authorized.machine),
        begun.admission,
        { kind: "authorized", role: "admin", userId: "user-1" },
        1_200,
      ).kind,
    ).toBe("stale");

    const first = scheduleWorkbookPresencePublication(
      initialWorkbookPresencePublicationMachine(),
      0,
    );
    const nearLimit = scheduleWorkbookPresencePublication(first.machine, 950);
    expect(first.dueAtMs).toBe(150);
    expect(nearLimit.dueAtMs).toBe(1_000);
    expect(
      settleWorkbookPresencePublication(nearLimit.machine, nearLimit.generation)
        .kind,
    ).toBe("publish");
    expect(
      settleWorkbookPresencePublication(nearLimit.machine, first.generation)
        .kind,
    ).toBe("stale");
  });
});
