import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { timelineCreateRelatedTargetContracts } from "./timelineWorkbookFeaturePolicy";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const createRelatedFeatures = timelineFeatures.filter(
  (feature) => feature.routeBinding.kind === "view_row_create",
);

describe("timelineWorkbookFeaturePolicy", () => {
  it("owns the exact contracts needed by every Timeline create-related target", () => {
    expect(createRelatedFeatures).toHaveLength(8);
    expect(timelineCreateRelatedTargetContracts.size).toBe(8);
    for (const featureGroup of createRelatedFeatures) {
      if (featureGroup.routeBinding.kind !== "view_row_create") {
        throw new Error("Expected a create-related route binding");
      }
      expect(
        timelineCreateRelatedTargetContracts.get(
          featureGroup.routeBinding.targetViewSchemaId ?? "",
        )?.viewSchemaId,
      ).toBe(featureGroup.routeBinding.targetViewSchemaId);
    }
  });
});
