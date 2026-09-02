import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import {
  type SemanticInspectorDisposition,
  semanticInspectorRegistrations,
} from "./semanticInspectorRegistry";

type SemanticInspectorFeatureResolution =
  | {
      readonly kind: "supported";
      readonly disposition: SemanticInspectorDisposition;
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | { readonly kind: "unsupported" };

export type InspectorRecordHistoryAction = "delete" | "restore" | "rollback";

export type InspectorOwnerCapability =
  | {
      readonly kind: "create_related";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "indicator";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "record_history";
      readonly action: InspectorRecordHistoryAction;
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | {
      readonly kind: "existing_owner_control";
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | { readonly kind: "unsupported" };

export type InspectorContextualCapability = Exclude<
  InspectorOwnerCapability,
  { readonly kind: "existing_owner_control" } | { readonly kind: "unsupported" }
>;

const registeredInspectorDispositionByKey = buildRegistrationLookup();

const roleRank: Readonly<Record<Exclude<WorkbookIncidentRole, "">, number>> = {
  viewer: 0,
  editor: 1,
  reviewer: 2,
  admin: 3,
};

function semanticInspectorFeatureKey(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): string {
  return JSON.stringify([
    viewSchemaId,
    featureGroup.featureGroupKey,
    featureGroup.panelId,
    featureGroup.routeBinding.kind,
    featureGroup.routeBinding.owner,
    featureGroup.routeBinding.actionKey ?? null,
  ]);
}

export function resolveSemanticInspectorFeature(
  config: InspectorConfig,
  featureGroupKey: string,
): SemanticInspectorFeatureResolution {
  const matchingFeatures = config.featureGroups.filter(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (matchingFeatures.length !== 1) {
    return { kind: "unsupported" };
  }
  const featureGroup = matchingFeatures[0];
  if (featureGroup === undefined) return { kind: "unsupported" };
  const semanticKey = semanticInspectorFeatureKey(
    config.viewSchemaId,
    featureGroup,
  );
  const disposition = registeredInspectorDispositionByKey.get(semanticKey);
  if (disposition === undefined) return { kind: "unsupported" };
  return {
    kind: "supported",
    disposition,
    featureGroup,
    semanticKey,
  };
}

export function resolveInspectorOwnerCapability(
  config: InspectorConfig,
  featureGroupKey: string,
): InspectorOwnerCapability {
  const semantic = resolveSemanticInspectorFeature(config, featureGroupKey);
  if (semantic.kind === "unsupported") return semantic;
  const { featureGroup, semanticKey } = semantic;
  if (semantic.disposition === "direct_history_action") {
    const action = recordHistoryAction(featureGroup);
    return action === null
      ? { kind: "unsupported" }
      : { action, featureGroup, kind: "record_history", semanticKey };
  }
  if (
    featureGroup.routeBinding.kind === "indicator_observations" ||
    featureGroup.routeBinding.kind === "indicator_lifecycle"
  ) {
    return { featureGroup, kind: "indicator", semanticKey };
  }
  if (
    semantic.disposition === "contextual_workflow_or_pivot" &&
    featureGroup.routeBinding.kind === "view_row_create" &&
    featureGroup.routeBinding.owner === "view_row_create_route"
  ) {
    return { featureGroup, kind: "create_related", semanticKey };
  }
  return { featureGroup, kind: "existing_owner_control", semanticKey };
}

export function inspectorContextualCapabilities({
  config,
  panelId,
  recordHistoryActions = [],
}: {
  readonly config: InspectorConfig;
  readonly panelId: InspectorFeatureGroup["panelId"];
  readonly recordHistoryActions?: readonly InspectorRecordHistoryAction[];
}): readonly InspectorContextualCapability[] {
  const allowedHistoryActions = new Set(recordHistoryActions);
  return config.featureGroups.flatMap((featureGroup) => {
    if (featureGroup.panelId !== panelId) return [];
    const capability = resolveInspectorOwnerCapability(
      config,
      featureGroup.featureGroupKey,
    );
    if (
      capability.kind === "unsupported" ||
      capability.kind === "existing_owner_control" ||
      (capability.kind === "record_history" &&
        !allowedHistoryActions.has(capability.action))
    ) {
      return [];
    }
    return [capability];
  });
}

export function inspectorRecordHistoryActions(
  config: InspectorConfig,
): ReadonlySet<InspectorRecordHistoryAction> {
  const actions = new Set<InspectorRecordHistoryAction>();
  for (const featureGroup of config.featureGroups) {
    const capability = resolveInspectorOwnerCapability(
      config,
      featureGroup.featureGroupKey,
    );
    if (capability.kind === "record_history") {
      actions.add(capability.action);
    }
  }
  return actions;
}

function recordHistoryAction(
  featureGroup: InspectorFeatureGroup,
): InspectorRecordHistoryAction | null {
  if (
    featureGroup.routeBinding.kind !== "record_action" ||
    featureGroup.routeBinding.actionKey !== featureGroup.featureGroupKey
  ) {
    return null;
  }
  switch (featureGroup.routeBinding.owner) {
    case "record_delete_route":
      return "delete";
    case "record_restore_route":
      return "restore";
    case "record_rollback_route":
      return "rollback";
    default:
      return null;
  }
}

export function inspectorFeatureDisabledTokens({
  currentIncidentRole,
  featureGroup,
  stateTokens,
}: {
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly featureGroup: InspectorFeatureGroup;
  readonly stateTokens: ReadonlySet<InspectorDisabledCondition>;
}): ReadonlySet<InspectorDisabledCondition> {
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

function buildRegistrationLookup(): ReadonlyMap<
  string,
  SemanticInspectorDisposition
> {
  const result = new Map<string, SemanticInspectorDisposition>();
  for (const registration of semanticInspectorRegistrations) {
    const [
      viewSchemaId,
      featureGroupKey,
      panelId,
      routeKind,
      routeOwner,
      actionKey,
      disposition,
    ] = registration;
    const key = JSON.stringify([
      viewSchemaId,
      featureGroupKey,
      panelId,
      routeKind,
      routeOwner,
      actionKey,
    ]);
    if (result.has(key)) {
      throw new Error(`Duplicate semantic inspector registration: ${key}`);
    }
    result.set(key, disposition);
  }
  return result;
}
