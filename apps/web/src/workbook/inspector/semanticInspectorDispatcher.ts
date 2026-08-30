import {
  type InspectorConfig,
  type InspectorFeatureGroup,
  listViewContracts,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export type InspectorDisabledToken =
  | "no_row_selected"
  | "incident_closed"
  | "authorization_lost"
  | "row_version_changed"
  | "record_deleted"
  | "record_merged"
  | "evidence_preview_unavailable"
  | "merge_target_unavailable"
  | "record_not_deleted"
  | "rollback_target_unavailable"
  | "party_text_unavailable"
  | "pivot_target_unavailable";

export type SemanticInspectorFeatureResolution =
  | {
      readonly kind: "panel_read";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "action";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | { readonly kind: "unsupported" };

const supportedRouteOwners = new Map<string, ReadonlySet<string>>([
  ["panel_read", new Set(["current_row_projection", "record_history_route"])],
  ["view_row_create", new Set(["view_row_create_route"])],
  ["record_patch", new Set(["record_patch_route"])],
  [
    "record_action",
    new Set([
      "record_delete_route",
      "record_restore_route",
      "record_rollback_route",
      "record_merge_route",
      "record_supersede_route",
      "record_mark_reviewed_route",
      "evidence_attach_blob_route",
    ]),
  ],
  ["entity_mention_action", new Set(["entity_mention_resolve_route"])],
  [
    "evidence_access",
    new Set([
      "evidence_preview_handle_route",
      "evidence_download_handle_route",
    ]),
  ],
  ["surface_pivot", new Set(["view_query_route"])],
  ["indicator_observations", new Set(["indicator_observations_route"])],
  ["indicator_lifecycle", new Set(["indicator_lifecycle_route"])],
]);

const roleRank: Readonly<Record<Exclude<WorkbookIncidentRole, "">, number>> = {
  viewer: 0,
  editor: 1,
  reviewer: 2,
  admin: 3,
};

export function semanticInspectorFeatureKey(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): string {
  return [
    viewSchemaId,
    featureGroup.featureGroupKey,
    featureGroup.routeBinding.kind,
    featureGroup.routeBinding.owner,
    featureGroup.routeBinding.actionKey ?? "",
  ].join("|");
}

export function resolveSemanticInspectorFeature(
  config: InspectorConfig,
  requested: InspectorFeatureGroup,
): SemanticInspectorFeatureResolution {
  const featureGroup = config.featureGroups.find(
    (candidate) => candidate.featureGroupKey === requested.featureGroupKey,
  );
  if (
    featureGroup === undefined ||
    !sameSemanticInspectorFeature(featureGroup, requested) ||
    !supportedRouteOwners
      .get(featureGroup.routeBinding.kind)
      ?.has(featureGroup.routeBinding.owner)
  ) {
    return { kind: "unsupported" };
  }
  const semanticKey = semanticInspectorFeatureKey(
    config.viewSchemaId,
    featureGroup,
  );
  return featureGroup.routeBinding.kind === "panel_read"
    ? { kind: "panel_read", featureGroup, semanticKey }
    : { kind: "action", featureGroup, semanticKey };
}

export function inspectorFeatureDisabledTokens({
  currentIncidentRole,
  featureGroup,
  stateTokens,
}: {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly stateTokens: ReadonlySet<InspectorDisabledToken>;
}): ReadonlySet<InspectorDisabledToken> {
  const tokens = new Set(stateTokens);
  const minimumRole = featureGroup.minimumIncidentRole;
  const currentRole = currentIncidentRole || null;
  if (
    minimumRole !== null &&
    (currentRole === null || roleRank[currentRole] < roleRank[minimumRole])
  ) {
    tokens.add("authorization_lost");
  }
  return tokens;
}

export function assertCurrentInspectorDispatchCompleteness(): void {
  const keys = new Set<string>();
  for (const contract of listViewContracts()) {
    for (const featureGroup of contract.inspectorConfig.featureGroups) {
      const resolution = resolveSemanticInspectorFeature(
        contract.inspectorConfig,
        featureGroup,
      );
      if (resolution.kind === "unsupported") {
        throw new Error(
          `Unsupported current inspector feature ${contract.viewSchemaId}/${featureGroup.featureGroupKey}`,
        );
      }
      if (keys.has(resolution.semanticKey)) {
        throw new Error(
          `Duplicate inspector semantic key ${resolution.semanticKey}`,
        );
      }
      keys.add(resolution.semanticKey);
    }
  }
}

function sameSemanticInspectorFeature(
  left: InspectorFeatureGroup,
  right: InspectorFeatureGroup,
): boolean {
  return (
    left.featureGroupKey === right.featureGroupKey &&
    left.panelId === right.panelId &&
    left.routeBinding.kind === right.routeBinding.kind &&
    left.routeBinding.owner === right.routeBinding.owner &&
    left.routeBinding.actionKey === right.routeBinding.actionKey &&
    left.routeBinding.targetViewSchemaId ===
      right.routeBinding.targetViewSchemaId &&
    left.minimumIncidentRole === right.minimumIncidentRole &&
    left.mutates === right.mutates &&
    left.requiresConfirmation === right.requiresConfirmation &&
    left.successResultBehavior === right.successResultBehavior &&
    left.failureResultBehavior === right.failureResultBehavior &&
    sameStringArray(left.disabledWhen, right.disabledWhen) &&
    sameSeedBindings(left.seedBindings, right.seedBindings)
  );
}

function sameStringArray(
  left: readonly string[],
  right: readonly string[],
): boolean {
  return (
    left.length === right.length &&
    left.every((value, index) => value === right[index])
  );
}

function sameSeedBindings(
  left: InspectorFeatureGroup["seedBindings"],
  right: InspectorFeatureGroup["seedBindings"],
): boolean {
  return (
    left.length === right.length &&
    left.every((binding, index) => {
      const candidate = right[index];
      return (
        candidate !== undefined &&
        binding.targetFieldKey === candidate.targetFieldKey &&
        binding.source.kind === candidate.source.kind &&
        binding.source.sourceFieldKey === candidate.source.sourceFieldKey &&
        sameLiteralValue(binding.source.value, candidate.source.value)
      );
    })
  );
}

function sameLiteralValue(left: unknown, right: unknown): boolean {
  if (Object.is(left, right)) return true;
  if (left === null || right === null) return false;
  if (Array.isArray(left) && Array.isArray(right)) {
    return (
      left.length === right.length &&
      left.every((value, index) => sameLiteralValue(value, right[index]))
    );
  }
  if (
    typeof left !== "object" ||
    typeof right !== "object" ||
    Array.isArray(left) ||
    Array.isArray(right)
  ) {
    return false;
  }
  const leftRecord = left as Readonly<Record<string, unknown>>;
  const rightRecord = right as Readonly<Record<string, unknown>>;
  const leftKeys = Object.keys(leftRecord).sort();
  const rightKeys = Object.keys(rightRecord).sort();
  return (
    sameStringArray(leftKeys, rightKeys) &&
    leftKeys.every((key) => sameLiteralValue(leftRecord[key], rightRecord[key]))
  );
}
