import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { timelineCreateRelatedTargetContracts } from "./timelineWorkbookFeaturePolicy";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const createRelatedFeatures = timelineFeatures.filter(
  (feature) =>
    feature.routeBinding.kind === "view_row_create" &&
    ![
      "create_related.task_request",
      "create_related.decision",
      "create_related.evidence",
    ].includes(feature.featureGroupKey),
);

describe("timelineWorkbookFeaturePolicy", () => {
  it("owns the contracts for Timeline create-related targets outside retained authoring", () => {
    expect(createRelatedFeatures).toHaveLength(5);
    expect(timelineCreateRelatedTargetContracts.size).toBe(5);
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
