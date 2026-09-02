import {
  type InspectorConfig,
  type InspectorFeatureGroup,
  listViewContracts,
} from "@cartulary/view-contracts";

const canonicalFeatureByIdentity = buildCanonicalFeatureIndex();

export function admitCanonicalInspectorFeature(
  config: InspectorConfig,
  featureGroupKey: string,
): InspectorFeatureGroup | null {
  const addressed = config.featureGroups.filter(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (addressed.length !== 1) return null;
  const candidate = addressed[0];
  if (candidate === undefined) return null;
  return (
    canonicalFeatureByIdentity.get(
      inspectorFeatureIdentity(config.viewSchemaId, candidate),
    ) ?? null
  );
}

export function inspectorFeatureIdentity(
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

function buildCanonicalFeatureIndex(): ReadonlyMap<
  string,
  InspectorFeatureGroup
> {
  const result = new Map<string, InspectorFeatureGroup>();
  for (const contract of listViewContracts()) {
    for (const featureGroup of contract.inspectorConfig.featureGroups) {
      const identity = inspectorFeatureIdentity(
        contract.viewSchemaId,
        featureGroup,
      );
      if (result.has(identity)) {
        throw new Error(`Duplicate canonical inspector feature: ${identity}`);
      }
      result.set(identity, featureGroup);
    }
  }
  return result;
}
