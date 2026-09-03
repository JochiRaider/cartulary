import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  planTimelineBulkTag,
  timelineBulkTagSubmissionIsCurrent,
} from "./timelineBulkTagPlan";
import { normalizeTimelineFullRow, rowFromApi } from "./timelineRowModel";

const timeline = requireViewContract(timelineViewSchemaId);
const context = {
  authorized: true,
  capabilityAvailable: true,
  surfaceKey: "view_schema:timeline",
};

describe("Timeline bulk-tag plan", () => {
  it("builds one ordered versioned command and binds settlement to its subject", () => {
    const first = committedRow("record-1", 3);
    const second = committedRow("record-2", 4);
    const plan = planTimelineBulkTag({
      context,
      rows: [first, second],
      selectedRecordIds: new Set(["record-1", "record-2"]),
      tagName: "  triaged  ",
    });
    expect(plan).toMatchObject({
      kind: "dispatch",
      normalizedTagName: "triaged",
      targets: [
        { baseRowVersion: 3, recordId: "record-1" },
        { baseRowVersion: 4, recordId: "record-2" },
      ],
    });
    if (plan.kind !== "dispatch") return;
    expect(
      timelineBulkTagSubmissionIsCurrent({
        context,
        plan,
        selectedRecordIds: new Set(["record-2", "record-1"]),
        tagName: "triaged",
      }),
    ).toBe(true);
    expect(
      timelineBulkTagSubmissionIsCurrent({
        context,
        plan,
        selectedRecordIds: new Set(["record-1"]),
        tagName: "triaged",
      }),
    ).toBe(false);
  });

  it("rejects access loss, missing capability, partial selection, and stale targets", () => {
    const row = committedRow("record-1", 3);
    const base = {
      context,
      rows: [row],
      selectedRecordIds: new Set(["record-1"]),
      tagName: "tag",
    };
    expect(
      planTimelineBulkTag({
        ...base,
        context: { ...context, authorized: false },
      }),
    ).toEqual({ kind: "reject", reason: "authorization_lost" });
    expect(
      planTimelineBulkTag({
        ...base,
        context: { ...context, capabilityAvailable: false },
      }),
    ).toEqual({ kind: "reject", reason: "capability_unavailable" });
    expect(planTimelineBulkTag({ ...base, rows: [] })).toEqual({
      kind: "reject",
      reason: "partial_selection",
    });
    expect(
      planTimelineBulkTag({
        ...base,
        rows: [{ ...row, pendingSignature: "pending" }],
      }),
    ).toEqual({ kind: "reject", reason: "invalid_target" });
  });
});

function committedRow(recordId: string, rowVersion: number) {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timeline, recordId, rowVersion, {}),
      "bulk tag plan fixture",
    ),
  );
}
