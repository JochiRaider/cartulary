import { describe, expect, it, vi } from "vitest";
import { metadataDeferred } from "../testing/incidentMetadataTestSupport";
import { reviewAppDeparture } from "./appDepartureReview";

describe("App departure review", () => {
  it("reviews all outstanding owners sequentially without overlapping dialogs", async () => {
    const membership = metadataDeferred<boolean>();
    const metadata = metadataDeferred<boolean>();
    let membershipDirty = true;
    let metadataDirty = true;
    const owners = {
      memberships: {
        hasWork: () => membershipDirty,
        requestLeave: vi.fn(() => membership.promise),
      },
      metadata: {
        hasWork: () => metadataDirty,
        requestLeave: vi.fn(() => metadata.promise),
      },
      deploymentUsers: {
        hasWork: () => false,
        requestLeave: vi.fn(async () => true),
      },
      isCurrent: () => true,
    };
    const result = reviewAppDeparture(owners);
    expect(owners.memberships.requestLeave).toHaveBeenCalledTimes(1);
    expect(owners.metadata.requestLeave).not.toHaveBeenCalled();
    membershipDirty = false;
    membership.resolve(true);
    await Promise.resolve();
    expect(owners.metadata.requestLeave).toHaveBeenCalledTimes(1);
    metadataDirty = false;
    metadata.resolve(true);
    expect(await result).toBe(true);
  });
  it("stops on Stay and preserves unreviewed work", async () => {
    const metadataLeave = vi.fn(async () => true);
    expect(
      await reviewAppDeparture({
        memberships: { hasWork: () => true, requestLeave: async () => false },
        metadata: { hasWork: () => true, requestLeave: metadataLeave },
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
        deploymentUsers: stillDirty,
        isCurrent: () => true,
      }),
    ).toBe(false);
  });
});
