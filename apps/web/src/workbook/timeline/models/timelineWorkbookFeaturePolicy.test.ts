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
        featureGroup,
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
      resolveTimelineWorkbookFeature(timelineViewSchemaId, indicatorFeature),
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
    const alterations: readonly InspectorFeatureGroup[] = [
      {
        ...relatedFeature,
        panelId: "relationships",
      },
      {
        ...relatedFeature,
        minimumIncidentRole: "admin",
      },
      {
        ...relatedFeature,
        mutates: false,
      },
      {
        ...relatedFeature,
        requiresConfirmation: true,
      },
      {
        ...relatedFeature,
        routeBinding: {
          ...relatedFeature.routeBinding,
          owner: "record_patch_route",
        },
      },
      {
        ...relatedFeature,
        routeBinding: {
          ...relatedFeature.routeBinding,
          targetViewSchemaId: "cartulary.view.unknown.v1",
        },
      },
      {
        ...relatedFeature,
        seedBindings: [
          {
            source: { kind: "selected_record_id" },
            targetFieldKey: "unexpected.field",
          },
        ],
      },
      {
        ...relatedFeature,
        disabledWhen: [...relatedFeature.disabledWhen, "record_deleted"],
      },
      {
        ...indicatorFeature,
        routeBinding: {
          actionKey: "create_related.note",
          kind: "view_row_create",
          owner: "view_row_create_route",
          targetViewSchemaId: "cartulary.view.notes.v1",
        },
      },
    ];

    for (const alteredFeature of alterations) {
      expect(
        resolveTimelineWorkbookFeature(timelineViewSchemaId, alteredFeature),
      ).toEqual({ kind: "unsupported" });
    }
    expect(
      resolveTimelineWorkbookFeature(
        "cartulary.view.indicators.v1",
        indicatorFeature,
      ),
    ).toEqual({ kind: "unsupported" });
    expect(
      resolveTimelineWorkbookFeature(timelineViewSchemaId, {
        ...relatedFeature,
        featureGroupKey: "create_related.unknown",
      }),
    ).toEqual({ kind: "unsupported" });

    expect(
      resolveTimelineWorkbookFeature(timelineViewSchemaId, {
        ...relatedFeature,
        label: "A label that is deliberately not an identity",
      }),
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
