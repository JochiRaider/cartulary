import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  planTimelineEvidenceTarget,
  timelineEvidenceTargetIdentity,
} from "./timelineEvidenceAttachmentPlan";
import {
  createDraftRowForKey,
  normalizeTimelineFullRow,
  rowFromApi,
} from "./timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const row = rowFromApi(
  normalizeTimelineFullRow(
    fullWorkbookViewRow(timeline, "record-1", 3, {}),
    "evidence target plan fixture",
  ),
);
const context = {
  authorized: true,
  capabilityAvailable: true,
  selectedRowKey: row.key,
  surfaceKey: "view_schema:timeline",
};
const identity = timelineEvidenceTargetIdentity(row, context.surfaceKey);

describe("Timeline Evidence attachment target plan", () => {
  it("re-reads the current row version before each dispatch", () => {
    const current = { ...row, rowVersion: 5 };
    expect(
      planTimelineEvidenceTarget({ context, identity, rows: [current] }),
    ).toEqual({ kind: "dispatch", target: current });
  });

  it("treats the unselected draft control as the active draft subject", () => {
    const draft = createDraftRowForKey("draft-1");
    if (draft === null) throw new Error("expected draft fixture");
    expect(
      planTimelineEvidenceTarget({
        context: { ...context, selectedRowKey: null },
        identity: timelineEvidenceTargetIdentity(draft, context.surfaceKey),
        rows: [draft],
      }),
    ).toEqual({ kind: "dispatch", target: draft });
  });

  it("rejects invalid action contexts", () => {
    for (const [nextContext, reason] of [
      [{ ...context, authorized: false }, "authorization_lost"],
      [{ ...context, capabilityAvailable: false }, "capability_unavailable"],
      [{ ...context, surfaceKey: "saved_view:other" }, "surface_changed"],
      [{ ...context, selectedRowKey: "other" }, "selection_changed"],
    ] as const) {
      expect(
        planTimelineEvidenceTarget({
          context: nextContext,
          identity,
          rows: [row],
        }),
      ).toEqual({ kind: "reject", reason });
    }
  });

  it("rejects deleted, replaced, and pending targets", () => {
    expect(planTimelineEvidenceTarget({ context, identity, rows: [] })).toEqual(
      {
        kind: "reject",
        reason: "target_missing",
      },
    );
    expect(
      planTimelineEvidenceTarget({
        context,
        identity,
        rows: [{ ...row, recordId: "replacement" }],
      }),
    ).toEqual({ kind: "reject", reason: "target_identity_changed" });
    expect(
      planTimelineEvidenceTarget({
        context,
        identity,
        rows: [{ ...row, pendingSignature: "pending" }],
      }),
    ).toEqual({ kind: "reject", reason: "target_not_dispatchable" });
  });
});
