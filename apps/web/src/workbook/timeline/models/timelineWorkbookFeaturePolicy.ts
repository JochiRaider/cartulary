import {
  getViewContract,
  type InspectorFeatureGroup,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type IndicatorInspectorHandler,
  resolveIndicatorInspectorHandler,
} from "../../features/indicators/indicatorInspectorHandlers";
import { resolveSemanticInspectorFeature } from "../../inspector/semanticInspectorDispatcher";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";

type TimelineWorkbookFeatureResolution =
  | {
      readonly kind: "indicator";
      readonly handler: IndicatorInspectorHandler;
    }
  | {
      readonly kind: "create_related";
      readonly featureGroup: InspectorFeatureGroup;
    }
  | {
      readonly kind: "panel_owned";
      readonly featureGroup: InspectorFeatureGroup;
    }
  | { readonly kind: "unsupported" };

const unsupportedTimelineWorkbookFeature = { kind: "unsupported" } as const;
const timelineContract = requireViewContract(timelineViewSchemaId);

const createRelatedTargetContracts = new Map<string, ViewContract>();
for (const featureGroup of timelineContract.inspectorConfig.featureGroups) {
  if (
    featureGroup.routeBinding.kind !== "view_row_create" ||
    featureGroup.routeBinding.owner !== "view_row_create_route" ||
    featureGroup.routeBinding.targetViewSchemaId === undefined
  ) {
    continue;
  }
  const targetContract = getViewContract(
    featureGroup.routeBinding.targetViewSchemaId,
  );
  if (targetContract !== undefined) {
    createRelatedTargetContracts.set(
      targetContract.viewSchemaId,
      targetContract,
    );
  }
}

export const timelineCreateRelatedTargetContracts: ReadonlyMap<
  string,
  ViewContract
> = createRelatedTargetContracts;

export function resolveTimelineWorkbookFeature(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): TimelineWorkbookFeatureResolution {
  if (viewSchemaId !== timelineViewSchemaId) {
    return unsupportedTimelineWorkbookFeature;
  }
  const semanticResolution = resolveSemanticInspectorFeature(
    timelineContract.inspectorConfig,
    featureGroup,
  );
  if (semanticResolution.kind === "unsupported") {
    return unsupportedTimelineWorkbookFeature;
  }
  const canonicalFeatureGroup = semanticResolution.featureGroup;

  const indicatorHandler = resolveIndicatorInspectorHandler(
    timelineViewSchemaId,
    canonicalFeatureGroup,
  );
  if (indicatorHandler !== null) {
    return { kind: "indicator", handler: indicatorHandler };
  }

  if (
    canonicalFeatureGroup.routeBinding.kind !== "view_row_create" ||
    canonicalFeatureGroup.routeBinding.owner !== "view_row_create_route" ||
    canonicalFeatureGroup.routeBinding.targetViewSchemaId === undefined ||
    !createRelatedTargetContracts.has(
      canonicalFeatureGroup.routeBinding.targetViewSchemaId,
    )
  ) {
    return {
      kind: "panel_owned",
      featureGroup: canonicalFeatureGroup,
    };
  }
  return {
    kind: "create_related",
    featureGroup: canonicalFeatureGroup,
  };
}
