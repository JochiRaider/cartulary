import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { timelineCreateRelatedTargetContracts } from "./timelineWorkbookFeaturePolicy";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const createRelatedFeatures = timelineFeatures.filter(
  (feature) =>
    feature.routeBinding.kind === "view_row_create" &&
    !["create_related.task_request", "create_related.decision"].includes(
      feature.featureGroupKey,
    ),
);

describe("timelineWorkbookFeaturePolicy", () => {
  it("owns the contracts for Timeline create-related targets outside retained Task and Decision authoring", () => {
    expect(createRelatedFeatures).toHaveLength(6);
    expect(timelineCreateRelatedTargetContracts.size).toBe(6);
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
