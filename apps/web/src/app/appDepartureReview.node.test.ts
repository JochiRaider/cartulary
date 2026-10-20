import { describe, expect, it, vi } from "vitest";
import { metadataDeferred } from "../testing/incidentMetadataTestSupport";
import { reviewAppDeparture } from "./appDepartureReview";

describe("App departure review", () => {
  it("reviews all outstanding owners sequentially without overlapping dialogs", async () => {
    const work = Array.from({ length: 5 }, () => ({
      dirty: true,
      decision: metadataDeferred<boolean>(),
    }));
    const participants = work.map((item) => ({
      hasWork: () => item.dirty,
      requestLeave: vi.fn(() => item.decision.promise),
    }));
    const [memberships, metadata, lifecycle, preferences, deploymentUsers] =
      participants;
    if (
      !memberships ||
      !metadata ||
      !lifecycle ||
      !preferences ||
      !deploymentUsers
    )
      throw new Error("Missing departure fixture");
    const result = reviewAppDeparture({
      memberships,
      metadata,
      lifecycle,
      preferences,
      deploymentUsers,
      isCurrent: () => true,
    });
    for (const [index, item] of work.entries()) {
      expect(participants[index]?.requestLeave).toHaveBeenCalledTimes(1);
      for (const later of participants.slice(index + 1))
        expect(later.requestLeave).not.toHaveBeenCalled();
      item.dirty = false;
      item.decision.resolve(true);
      await Promise.resolve();
    }
    expect(await result).toBe(true);
  });
  it("stops on Stay and preserves unreviewed work", async () => {
    const metadataLeave = vi.fn(async () => true);
    expect(
      await reviewAppDeparture({
        memberships: { hasWork: () => true, requestLeave: async () => false },
        metadata: { hasWork: () => true, requestLeave: metadataLeave },
        lifecycle: { hasWork: () => false, requestLeave: async () => true },
        preferences: { hasWork: () => false, requestLeave: async () => true },
        deploymentUsers: {
          hasWork: () => false,
          requestLeave: async () => true,
        },
        isCurrent: () => true,
      }),
    ).toBe(false);
    expect(metadataLeave).not.toHaveBeenCalled();
  });
  it("rejects retirement or newly outstanding work before navigation", async () => {
    let current = true;
    const metadataLeave = vi.fn(async () => true);
    expect(
      await reviewAppDeparture({
        memberships: {
          hasWork: () => true,
          requestLeave: async () => {
            current = false;
            return true;
          },
        },
        metadata: { hasWork: () => true, requestLeave: metadataLeave },
        lifecycle: { hasWork: () => false, requestLeave: async () => true },
        preferences: { hasWork: () => false, requestLeave: async () => true },
        deploymentUsers: {
          hasWork: () => false,
          requestLeave: async () => true,
        },
        isCurrent: () => current,
      }),
    ).toBe(false);
    expect(metadataLeave).not.toHaveBeenCalled();
    const stillDirty = { hasWork: () => true, requestLeave: async () => true };
    expect(
      await reviewAppDeparture({
        memberships: stillDirty,
        metadata: stillDirty,
        lifecycle: stillDirty,
        preferences: stillDirty,
        deploymentUsers: stillDirty,
        isCurrent: () => true,
      }),
    ).toBe(false);
  });
});
