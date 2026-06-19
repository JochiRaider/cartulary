import { describe, expect, it } from "vitest";
import {
  beginTimelineEntityRefreshBarrier,
  settleTimelineViewportContinuityBarrier,
  type TimelineEntityCatalogInput,
  timelineEntityCatalogKey,
  timelineEntityRefreshExpectationForMention,
  timelineViewportContinuityBarrierSatisfied,
  withTimelineEntityRefreshExpectation,
} from "./timelineViewportContinuityModel";

const baselineCatalog: TimelineEntityCatalogInput = {
  hostEntities: [
    {
      entityType: "host",
      recordId: "host-1",
      rowVersion: 2,
    },
  ],
  identityEntities: [
    {
      entityType: "identity",
      recordId: "identity-1",
      rowVersion: 3,
    },
  ],
};

describe("timelineViewportContinuityModel", () => {
  it("builds stable semantic catalog keys independent of list identity and order", () => {
    expect(timelineEntityCatalogKey(baselineCatalog)).toBe(
      "host:host-1:2|identity:identity-1:3",
    );
    expect(
      timelineEntityCatalogKey({
        hostEntities: [
          {
            entityType: "host",
            recordId: "host-1",
            rowVersion: 2,
          },
        ],
        identityEntities: [
          {
            entityType: "identity",
            recordId: "identity-1",
            rowVersion: 3,
          },
        ],
      }),
    ).toBe(timelineEntityCatalogKey(baselineCatalog));
  });

  it("holds an entity refresh barrier until refresh settlement and catalog progress", () => {
    const barrier = beginTimelineEntityRefreshBarrier(baselineCatalog);

    expect(
      timelineViewportContinuityBarrierSatisfied(barrier, baselineCatalog),
    ).toBe(false);
    expect(
      timelineViewportContinuityBarrierSatisfied(
        settleTimelineViewportContinuityBarrier(barrier, "complete"),
        baselineCatalog,
      ),
    ).toBe(false);
    expect(
      timelineViewportContinuityBarrierSatisfied(
        settleTimelineViewportContinuityBarrier(barrier, "complete"),
        {
          ...baselineCatalog,
          identityEntities: [
            ...baselineCatalog.identityEntities,
            {
              entityType: "identity",
              recordId: "identity-2",
              rowVersion: 1,
            },
          ],
        },
      ),
    ).toBe(true);
  });

  it("settles when the expected create-from-mention entity becomes visible", () => {
    const expected = timelineEntityRefreshExpectationForMention(
      {
        collectionValues: {
          hostRefs: [],
          identityRefs: [
            {
              entityType: "identity",
              itemRef: "mention-1",
              resolvedRecordId: "identity-2",
            },
          ],
        },
      },
      "mention-1",
    );
    const barrier = settleTimelineViewportContinuityBarrier(
      withTimelineEntityRefreshExpectation(
        beginTimelineEntityRefreshBarrier(baselineCatalog),
        expected,
      ),
      "complete",
    );

    expect(
      timelineViewportContinuityBarrierSatisfied(barrier, baselineCatalog),
    ).toBe(false);
    expect(
      timelineViewportContinuityBarrierSatisfied(barrier, {
        ...baselineCatalog,
        identityEntities: [
          {
            entityType: "identity",
            recordId: "identity-2",
            rowVersion: 1,
          },
        ],
      }),
    ).toBe(true);
  });

  it("lets terminal refresh failures release continuity to row-local fallback restoration", () => {
    const barrier = settleTimelineViewportContinuityBarrier(
      beginTimelineEntityRefreshBarrier(baselineCatalog),
      "terminal",
    );

    expect(
      timelineViewportContinuityBarrierSatisfied(barrier, baselineCatalog),
    ).toBe(true);
  });
});
