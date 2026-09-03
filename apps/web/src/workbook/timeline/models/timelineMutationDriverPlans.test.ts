import { describe, expect, it } from "vitest";
import { createWorkbookPendingQueueRuntime } from "../../runtime/workbookPendingReplayRuntime";
import type {
  PendingQueueSnapshot,
  PendingReplaySettlement,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";
import {
  planTimelineAcceptedProjection,
  planTimelineDiscard,
  planTimelineRejectedSettlement,
  planTimelineReplayAdmission,
} from "./timelineMutationDriverPlans";
import { timelinePendingSavesRefsFor } from "./timelinePendingSaves";

const unit: PendingReplayUnitState = {
  clientInstanceId: "client-1",
  clientTxnId: "txn-1",
  coalesceKey: "record-1",
  enqueueOrder: 1,
  id: "unit-1",
  identity: {
    base_row_version: 3,
    changes: [{ field_key: "timeline.activity_synopsis_text", value: "Local" }],
    client_txn_id: "txn-1",
    kind: "patch",
    record_id: "record-1",
    route_scope: { record_id: "record-1" },
    view_schema_id: "cartulary.view.timeline.v2",
  },
  incidentId: "incident-1",
  kind: "patch",
  mutationSignature: "signature-1",
  operationClass: "hot_path",
  payloadIntent: { base_row_version: 3, changes: [] },
  recordId: "record-1",
  rowKey: "record-1",
  source: "autosave",
  status: "queued",
  viewSchemaId: "cartulary.view.timeline.v2",
};

function snapshot(
  overrides: Partial<PendingQueueSnapshot> = {},
): PendingQueueSnapshot {
  return {
    authPaused: false,
    capacity: 64,
    halted: null,
    inFlightCount: 0,
    overflow: null,
    primarySaveStateInput: "Syncing",
    queuedCount: 1,
    sameFieldConflicts: [],
    saveStatePresentation: {
      conflictAnchors: [],
      primaryLabel: "Syncing",
      secondaryKind: "queued",
      secondaryMessage: "Queued",
    },
    scope: { clientInstanceId: "client-1", incidentId: "incident-1" },
    units: [unit],
    ...overrides,
  };
}

function admission(
  overrides: {
    candidate?: PendingReplayUnitState | null;
    currentRowVersion?: number | null;
    envelopeViewSchemaId?: string;
    expectedUnitId?: string;
    hasLocalConflict?: boolean;
    hasMetadata?: boolean;
    refreshBlocked?: boolean;
    snapshot?: PendingQueueSnapshot;
  } = {},
) {
  return planTimelineReplayAdmission({
    candidate: overrides.candidate === undefined ? unit : overrides.candidate,
    currentRowVersion:
      "currentRowVersion" in overrides ? overrides.currentRowVersion : 3,
    envelopeViewSchemaId:
      overrides.envelopeViewSchemaId ?? "cartulary.view.timeline.v2",
    expectedUnitId: overrides.expectedUnitId ?? unit.id,
    hasLocalConflict: overrides.hasLocalConflict ?? false,
    hasMetadata: overrides.hasMetadata ?? true,
    refreshBlocked: overrides.refreshBlocked ?? false,
    snapshot: overrides.snapshot ?? snapshot(),
  });
}

describe("Timeline mutation driver plans", () => {
  it("admits only the exact current owner unit with a committed patch version", () => {
    expect(admission()).toEqual({
      kind: "dispatch",
      committedRowVersion: 3,
    });
    expect(admission({ expectedUnitId: "other" })).toEqual({
      kind: "pause",
      reason: "owner_mismatch",
    });
    expect(admission({ hasMetadata: false })).toEqual({
      kind: "reject_missing_metadata",
    });
    expect(admission({ currentRowVersion: null })).toEqual({
      kind: "retry_missing_committed_row",
    });
    expect(admission({ refreshBlocked: true })).toEqual({
      kind: "pause",
      reason: "refresh",
    });
  });

  it("pauses for authentication, terminal, and conflict gates", () => {
    expect(admission({ snapshot: snapshot({ authPaused: true }) })).toEqual({
      kind: "pause",
      reason: "authentication",
    });
    expect(
      admission({
        snapshot: snapshot({
          halted: {
            anchor: { kind: "surface" },
            error_code: "terminal",
            message: "Stopped",
            unit_id: unit.id,
          },
        }),
      }),
    ).toEqual({ kind: "pause", reason: "halted" });
    expect(admission({ hasLocalConflict: true })).toEqual({
      kind: "pause",
      reason: "conflict",
    });
    expect(admission({ candidate: null })).toEqual({ kind: "idle" });
  });

  it("keeps Timeline owner queue context across surface remounts", () => {
    const mutationRuntime = {};
    const pendingQueue = createWorkbookPendingQueueRuntime({
      clientInstanceId: "client-1",
      incidentId: "incident-1",
    });
    const firstMount = timelinePendingSavesRefsFor(
      mutationRuntime,
      pendingQueue,
    );
    firstMount.pendingReplayOrderRef.current = 9;
    firstMount.pendingSignaturesRef.current.set("row-1", "signature-1");

    const remount = timelinePendingSavesRefsFor(mutationRuntime, pendingQueue);
    expect(remount).toBe(firstMount);
    expect(remount.pendingReplayOrderRef.current).toBe(9);
    expect(remount.pendingSignaturesRef.current.get("row-1")).toBe(
      "signature-1",
    );

    expect(timelinePendingSavesRefsFor({}, pendingQueue)).not.toBe(firstMount);
  });

  it("projects every rejected settlement into an explicit driver action", () => {
    const retrySettlement: PendingReplaySettlement = {
      outcome: "retryable_failure",
      snapshot: snapshot(),
      unit,
    };
    expect(
      planTimelineRejectedSettlement(retrySettlement, {
        kind: "retryable",
        message: "Retry",
      }),
    ).toEqual({ kind: "retry" });

    const authSettlement: PendingReplaySettlement = {
      outcome: "auth_paused",
      snapshot: snapshot({ authPaused: true }),
      unit,
    };
    expect(
      planTimelineRejectedSettlement(authSettlement, {
        kind: "authentication_required",
        message: "Authenticate",
      }),
    ).toEqual({ kind: "request_authorization" });

    const haltedSettlement: PendingReplaySettlement = {
      halt: {
        anchor: { kind: "surface" },
        error_code: "terminal",
        message: "Safe public halt",
        unit_id: unit.id,
      },
      outcome: "halted",
      snapshot: snapshot(),
      unit,
    };
    expect(
      planTimelineRejectedSettlement(haltedSettlement, {
        kind: "terminal",
        message: "Raw failure",
      }),
    ).toEqual({ kind: "halt", message: "Safe public halt" });
  });

  it("makes stale accepted projection and discard effects explicit", () => {
    expect(
      planTimelineAcceptedProjection({
        currentRowVersion: 5,
        postMutationQueryRefreshRequired: true,
        responseRowVersion: 4,
      }),
    ).toEqual({
      kind: "apply_committed_result",
      preserveKnownCommittedRow: true,
      refreshAfterApply: true,
    });
    expect(planTimelineDiscard({ hasMetadata: true, recovered: true })).toEqual(
      { kind: "reconcile", clearViewportContinuity: true },
    );
    expect(
      planTimelineDiscard({ hasMetadata: false, recovered: false }),
    ).toEqual({ kind: "refused" });
  });
});
