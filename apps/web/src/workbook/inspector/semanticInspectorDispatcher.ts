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

export type SemanticInspectorFeatureResolution =
  | {
      readonly kind: "supported";
      readonly disposition: SemanticInspectorDisposition;
      readonly featureGroup: InspectorFeatureGroup;
      readonly semanticKey: string;
    }
  | { readonly kind: "unsupported" };

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
