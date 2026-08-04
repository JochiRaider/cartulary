import {
  type InspectorFeatureGroup,
  type InspectorSpecializedActionKey,
  listViewContracts,
} from "@cartulary/view-contracts";

export type IndicatorInspectorAction = InspectorSpecializedActionKey;

export type IndicatorInspectorHandler = {
  readonly action: IndicatorInspectorAction;
  readonly panelId: "relationships" | "history";
};

const indicatorInspectorActionKeys: ReadonlySet<string> = new Set(
  listViewContracts().flatMap((contract) =>
    contract.inspectorConfig.featureGroups.flatMap((feature) =>
      (feature.routeBinding.kind === "indicator_observations" ||
        feature.routeBinding.kind === "indicator_lifecycle") &&
      feature.routeBinding.actionKey !== undefined
        ? [feature.routeBinding.actionKey]
        : [],
    ),
  ),
);

const indicatorInspectorBindings = listViewContracts().flatMap((contract) =>
  contract.inspectorConfig.featureGroups.flatMap((feature) => {
    const action = feature.routeBinding.actionKey;
    if (
      (feature.routeBinding.kind !== "indicator_observations" &&
        feature.routeBinding.kind !== "indicator_lifecycle") ||
      (feature.panelId !== "relationships" && feature.panelId !== "history") ||
      !isIndicatorInspectorAction(action)
    ) {
      return [];
    }
    return [
      {
        action,
        featureGroupKey: feature.featureGroupKey,
        kind: feature.routeBinding.kind,
        owner: feature.routeBinding.owner,
        panelId: feature.panelId,
        viewSchemaId: contract.viewSchemaId,
      },
    ];
  }),
);

export function resolveIndicatorInspectorHandler(
  viewSchemaId: string,
  featureGroup: InspectorFeatureGroup,
): IndicatorInspectorHandler | null {
  const binding = indicatorInspectorBindings.find(
    (candidate) =>
      candidate.viewSchemaId === viewSchemaId &&
      candidate.featureGroupKey === featureGroup.featureGroupKey &&
      candidate.panelId === featureGroup.panelId &&
      candidate.kind === featureGroup.routeBinding.kind &&
      candidate.owner === featureGroup.routeBinding.owner &&
      candidate.action === featureGroup.routeBinding.actionKey,
  );
  return binding === undefined
    ? null
    : { action: binding.action, panelId: binding.panelId };
}

export function isIndicatorInspectorAction(
  value: string | undefined,
): value is IndicatorInspectorAction {
  return value !== undefined && indicatorInspectorActionKeys.has(value);
}
