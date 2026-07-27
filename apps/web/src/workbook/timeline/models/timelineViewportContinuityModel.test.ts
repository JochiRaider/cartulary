import { describe, expect, it } from "vitest";
import {
  advanceTimelineContinuityRender,
  beginTimelineContinuityLifecycle,
  requireTimelineSourceRecord,
  settleTimelineContinuityRequirement,
  timelineContinuityRequirementsSettled,
  timelineSourceRecordRequirementSatisfied,
  transitionTimelineContinuity,
} from "./timelineViewportContinuityModel";

describe("timelineViewportContinuityModel", () => {
  it("holds continuity until the authoritative source row reaches its response version", () => {
    const required = requireTimelineSourceRecord(
      beginTimelineContinuityLifecycle({
        semanticFocusTarget: {
          kind: "row-inspect",
          recordId: "record-1",
        },
        userInterruptionGeneration: 4,
      }),
      { recordId: "record-1", minimumRowVersion: 8 },
    );

    const stale = advanceTimelineContinuityRender(required, {
      sourceRecord: { recordId: "record-1", rowVersion: 7 },
    });
    expect(timelineContinuityRequirementsSettled(stale)).toBe(false);

    const exact = advanceTimelineContinuityRender(stale, {
      sourceRecord: { recordId: "record-1", rowVersion: 8 },
    });
    expect(timelineContinuityRequirementsSettled(exact)).toBe(true);

    const concurrent = advanceTimelineContinuityRender(required, {
      sourceRecord: { recordId: "record-1", rowVersion: 9 },
    });
    expect(timelineContinuityRequirementsSettled(concurrent)).toBe(true);
  });

  it("rejects source row identity drift and never treats another row as convergence", () => {
    const lifecycle = beginTimelineContinuityLifecycle({
      semanticFocusTarget: {
        kind: "row-inspect",
        recordId: "record-1",
      },
      userInterruptionGeneration: 0,
    });

    expect(() =>
      requireTimelineSourceRecord(lifecycle, {
        recordId: "record-2",
        minimumRowVersion: 2,
      }),
    ).toThrow(/does not match target/);
    expect(
      timelineSourceRecordRequirementSatisfied(
        { recordId: "record-1", minimumRowVersion: 2 },
        { recordId: "record-2", rowVersion: 9 },
      ),
    ).toBe(false);
  });

  it("tracks named follow-ups, render generations, interruption generation, and terminal state", () => {
    const lifecycle = beginTimelineContinuityLifecycle({
      semanticFocusTarget: {
        kind: "row-inspect",
        recordId: "record-1",
      },
      userInterruptionGeneration: 12,
      requirements: ["entity-refresh"],
    });
    expect(lifecycle).toMatchObject({
      followUpRequirements: { "entity-refresh": "pending" },
      renderGeneration: 0,
      state: "pending",
      userInterruptionGeneration: 12,
    });
    expect(transitionTimelineContinuity(lifecycle, "cancelled")).toMatchObject({
      followUpRequirements: { "entity-refresh": "pending" },
      state: "cancelled",
      userInterruptionGeneration: 12,
    });

    const settled = settleTimelineContinuityRequirement(
      lifecycle,
      "entity-refresh",
      "settled",
    );
    expect(settled.renderGeneration).toBe(1);
    expect(timelineContinuityRequirementsSettled(settled)).toBe(true);
    expect(transitionTimelineContinuity(settled, "completed").state).toBe(
      "completed",
    );

    const failed = transitionTimelineContinuity(lifecycle, "failed");
    expect(failed).toMatchObject({
      followUpRequirements: { "entity-refresh": "terminal" },
      state: "failed",
    });
  });
});
