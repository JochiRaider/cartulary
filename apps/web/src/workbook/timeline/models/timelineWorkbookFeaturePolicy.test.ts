import {
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  resolveTimelineWorkbookFeature,
  timelineCreateRelatedTargetContracts,
} from "./timelineWorkbookFeaturePolicy";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const createRelatedFeatures = timelineFeatures.filter(
  (feature) => feature.routeBinding.kind === "view_row_create",
);
const indicatorFeature = requireTimelineFeature(
  "indicator.observations.manage",
);

describe("timelineWorkbookFeaturePolicy", () => {
  it("resolves every canonical Timeline create-related and Indicator tuple", () => {
    expect(createRelatedFeatures).toHaveLength(8);
    for (const featureGroup of createRelatedFeatures) {
      const result = resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        featureGroup.featureGroupKey,
      );
      expect(result).toEqual({ kind: "create_related", featureGroup });
      if (featureGroup.routeBinding.kind !== "view_row_create") {
        throw new Error("Expected a create-related route binding");
      }
      expect(
        timelineCreateRelatedTargetContracts.get(
          featureGroup.routeBinding.targetViewSchemaId ?? "",
        )?.viewSchemaId,
      ).toBe(featureGroup.routeBinding.targetViewSchemaId);
    }

    expect(
      resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        indicatorFeature.featureGroupKey,
      ),
    ).toEqual({
      kind: "indicator",
      handler: {
        action: "indicator.observations.manage",
        panelId: "relationships",
      },
    });
  });

  it("fails closed for unsupported schemas and altered semantic tuple members", () => {
    const relatedFeature = requireTimelineFeature("create_related.note");
    expect(
      resolveTimelineWorkbookFeature(
        "cartulary.view.indicators.v1",
        indicatorFeature.featureGroupKey,
      ),
    ).toEqual({ kind: "unsupported" });
    expect(
      resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        "create_related.unknown",
      ),
    ).toEqual({ kind: "unsupported" });

    expect(
      resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        relatedFeature.featureGroupKey,
      ),
    ).toEqual({ kind: "create_related", featureGroup: relatedFeature });
  });
});

function requireTimelineFeature(
  featureGroupKey: string,
): InspectorFeatureGroup {
  const featureGroup = timelineFeatures.find(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (featureGroup === undefined) {
    throw new Error(`Missing Timeline feature ${featureGroupKey}`);
  }
  return featureGroup;
}
