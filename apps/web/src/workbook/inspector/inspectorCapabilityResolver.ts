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

type RelatedCreationOwner =
  | "contextual"
  | "coordination"
  | "note"
  | "timeline_evidence"
  | "assessment";

/** A route alone never grants a generic mutation lifetime. */
export function inspectorRelatedCreationOwner(
  viewSchemaId: string,
  feature: InspectorFeatureGroup,
): RelatedCreationOwner | null {
  const route = feature.routeBinding;
  if (feature.featureGroupKey === "create_related.note")
    return route.kind === "record_action" &&
      route.owner === "record_linked_note_create_route" &&
      route.actionKey === feature.featureGroupKey
      ? "note"
      : null;
  if (
    route.kind !== "view_row_create" ||
    route.owner !== "view_row_create_route"
  )
    return null;
  const targets = {
    "create_related.task_request": [
      "cartulary.view.task_requests.v1",
      "contextual",
    ],
    "create_related.decision": ["cartulary.view.decisions.v1", "contextual"],
    "create_related.comm_log": ["cartulary.view.comm_log.v1", "coordination"],
    "create_related.handoff": ["cartulary.view.handoff.v1", "coordination"],
    "create_related.status_review": [
      "cartulary.view.status_review.v1",
      "coordination",
    ],
    "create_related.lesson": ["cartulary.view.lesson.v1", "coordination"],
    "create_related.evidence": [
      "cartulary.view.evidence.v1",
      "timeline_evidence",
    ],
    "create_related.assessment": [
      "cartulary.view.assessments.v1",
      "assessment",
    ],
  } as const satisfies Readonly<
    Record<string, readonly [string, RelatedCreationOwner]>
  >;
  const binding = Object.entries(targets).find(
    ([key]) => key === feature.featureGroupKey,
  )?.[1];
  if (!binding || binding[0] !== route.targetViewSchemaId) return null;
  if (
    binding[1] === "timeline_evidence" &&
    viewSchemaId !== "cartulary.view.timeline.v2"
  )
    return null;
  if (
    binding[1] === "assessment" &&
    viewSchemaId !== "cartulary.view.assessments.v1"
  )
    return null;
  return binding[1];
}

export type InspectorContextualCapability =
  | {
      readonly kind: "timeline_capture";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "decision_supersede";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "note_create";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
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
    if (canonical === null) return [];
    const capability = contextualCapability(config.viewSchemaId, canonical);
    if (
      canonical.requiresConfirmation &&
      capability?.kind !== "decision_supersede" &&
      capability?.kind !== "timeline_capture"
    )
      return [];
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
    viewSchemaId === "cartulary.view.timeline.v2" &&
    featureGroup.routeBinding.kind === "record_action" &&
    featureGroup.routeBinding.actionKey === featureGroup.featureGroupKey &&
    ((featureGroup.featureGroupKey === "timeline.mark_reviewed" &&
      featureGroup.routeBinding.owner === "record_mark_reviewed_route") ||
      (featureGroup.featureGroupKey === "timeline.supersede" &&
        featureGroup.routeBinding.owner === "record_supersede_route"))
  )
    return { kind: "timeline_capture", featureGroup, semanticKey };
  if (
    viewSchemaId === "cartulary.view.decisions.v1" &&
    featureGroup.featureGroupKey === "decision.supersede" &&
    featureGroup.routeBinding.kind === "record_action" &&
    featureGroup.routeBinding.owner === "record_supersede_route" &&
    featureGroup.routeBinding.actionKey === "decision.supersede"
  ) {
    return { kind: "decision_supersede", featureGroup, semanticKey };
  }
  if (
    featureGroup.routeBinding.kind === "indicator_observations" ||
    featureGroup.routeBinding.kind === "indicator_lifecycle"
  ) {
    return { featureGroup, kind: "indicator", semanticKey };
  }
  if (
    featureGroup.featureGroupKey === "create_related.note" &&
    featureGroup.routeBinding.kind === "record_action" &&
    featureGroup.routeBinding.owner === "record_linked_note_create_route" &&
    featureGroup.routeBinding.actionKey === "create_related.note"
  )
    return { featureGroup, kind: "note_create", semanticKey };
  if (
    featureGroup.routeBinding.kind === "view_row_create" &&
    featureGroup.routeBinding.owner === "view_row_create_route" &&
    inspectorRelatedCreationOwner(viewSchemaId, featureGroup) !== null
  ) {
    return { featureGroup, kind: "create_related", semanticKey };
  }
  return null;
}

const recordHistoryActionsByOwner = {
  current_row_projection: null,
  note_associations_route: null,
  entity_mention_resolve_route: null,
  evidence_attach_blob_route: null,
  evidence_download_handle_route: null,
  evidence_preview_handle_route: null,
  indicator_lifecycle_route: null,
  indicator_observations_route: null,
  record_delete_route: "delete",
  record_history_route: null,
  record_linked_note_create_route: null,
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
