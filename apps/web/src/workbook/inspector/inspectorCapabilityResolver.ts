import type {
  InspectorConfig,
  InspectorFeatureGroup,
  InspectorRouteBindingOwner,
} from "@cartulary/view-contracts";
import {
  admitCanonicalInspectorFeature,
  inspectorFeatureIdentity,
} from "./canonicalInspectorAdmission";

export type InspectorRecordHistoryAction = "delete" | "restore" | "rollback";

export type InspectorContextualCapability =
  | {
      readonly kind: "create_related";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "indicator";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    };

export function inspectorContextualCapabilities({
  config,
  panelId,
}: {
  readonly config: InspectorConfig;
  readonly panelId: InspectorFeatureGroup["panelId"];
}): readonly InspectorContextualCapability[] {
  return config.featureGroups.flatMap((candidate) => {
    if (candidate.panelId !== panelId) return [];
    const canonical = admitCanonicalInspectorFeature(
      config,
      candidate.featureGroupKey,
    );
    if (canonical === null || canonical.requiresConfirmation) return [];
    const capability = contextualCapability(config.viewSchemaId, canonical);
    return capability === null ? [] : [capability];
  });
}

export function inspectorRecordHistoryActions(
  config: InspectorConfig,
): ReadonlySet<InspectorRecordHistoryAction> {
  const actions = new Set<InspectorRecordHistoryAction>();
  for (const candidate of config.featureGroups) {
    const canonical = admitCanonicalInspectorFeature(
      config,
      candidate.featureGroupKey,
    );
    if (canonical === null) continue;
    const action = recordHistoryAction(canonical);
    if (action !== null) actions.add(action);
  }
  return actions;
}

function contextualCapability(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): InspectorContextualCapability | null {
  const semanticKey = inspectorFeatureIdentity(viewSchemaId, featureGroup);
  if (
    featureGroup.routeBinding.kind === "indicator_observations" ||
    featureGroup.routeBinding.kind === "indicator_lifecycle"
  ) {
    return { featureGroup, kind: "indicator", semanticKey };
  }
  if (
    featureGroup.routeBinding.kind === "view_row_create" &&
    featureGroup.routeBinding.owner === "view_row_create_route"
  ) {
    return { featureGroup, kind: "create_related", semanticKey };
  }
  return null;
}

const recordHistoryActionsByOwner = {
  current_row_projection: null,
  entity_mention_resolve_route: null,
  evidence_attach_blob_route: null,
  evidence_download_handle_route: null,
  evidence_preview_handle_route: null,
  indicator_lifecycle_route: null,
  indicator_observations_route: null,
  record_delete_route: "delete",
  record_history_route: null,
  record_mark_reviewed_route: null,
  record_merge_route: null,
  record_patch_route: null,
  record_restore_route: "restore",
  record_rollback_route: "rollback",
  record_supersede_route: null,
  view_query_route: null,
  view_row_create_route: null,
} satisfies Readonly<
  Record<InspectorRouteBindingOwner, InspectorRecordHistoryAction | null>
>;

function recordHistoryAction(
  featureGroup: InspectorFeatureGroup,
): InspectorRecordHistoryAction | null {
  if (
    featureGroup.routeBinding.kind !== "record_action" ||
    featureGroup.routeBinding.actionKey !== featureGroup.featureGroupKey
  ) {
    return null;
  }
  return recordHistoryActionsByOwner[featureGroup.routeBinding.owner];
}
