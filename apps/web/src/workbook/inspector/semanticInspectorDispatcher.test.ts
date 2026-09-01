import {
  type InspectorConfig,
  type InspectorFeatureGroup,
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  inspectorFeatureDisabledTokens,
  resolveSemanticInspectorFeature,
  type SemanticInspectorDisposition,
} from "./semanticInspectorDispatcher";

describe("semantic inspector dispatcher", () => {
  it("resolves every current projected tuple once to its canonical object and disposition", () => {
    const contracts = listViewContracts();
    const semanticKeys = new Set<string>();
    let featureCount = 0;

    for (const contract of contracts) {
      for (const featureGroup of contract.inspectorConfig.featureGroups) {
        const resolution = resolveSemanticInspectorFeature(
          contract.inspectorConfig,
          featureGroup.featureGroupKey,
        );
        expect(resolution.kind).toBe("supported");
        if (resolution.kind === "unsupported") continue;
        expect(resolution.featureGroup).toBe(featureGroup);
        expect(resolution.disposition).toBe(expectedDisposition(featureGroup));
        expect(semanticKeys.has(resolution.semanticKey)).toBe(false);
        semanticKeys.add(resolution.semanticKey);
        featureCount += 1;
      }
    }

    expect(contracts).toHaveLength(17);
    expect(featureCount).toBe(247);
    expect(semanticKeys.size).toBe(247);
  });

  it("omits an unknown key and any additive or altered stable tuple", () => {
    const canonicalConfig = requireViewContract(
      "cartulary.view.timeline.v2",
    ).inspectorConfig;
    const canonical = canonicalConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "create_related.note",
    );
    expect(canonical).toBeDefined();
    if (!canonical) return;

    expect(
      resolveSemanticInspectorFeature(canonicalConfig, "future.unknown_action"),
    ).toEqual({ kind: "unsupported" });

    const additiveConfig = withFeatureGroups(canonicalConfig, [
      ...canonicalConfig.featureGroups,
      { ...canonical, featureGroupKey: "future.unknown_action" },
    ]);
    expect(
      resolveSemanticInspectorFeature(additiveConfig, "future.unknown_action"),
    ).toEqual({ kind: "unsupported" });

    for (const altered of [
      { ...canonical, featureGroupKey: "create_related.changed" },
      { ...canonical, panelId: "details" as const },
      {
        ...canonical,
        routeBinding: {
          ...canonical.routeBinding,
          kind: "record_patch" as const,
        },
      },
      {
        ...canonical,
        routeBinding: {
          ...canonical.routeBinding,
          owner: "record_patch_route" as const,
        },
      },
      {
        ...canonical,
        routeBinding: {
          ...canonical.routeBinding,
          actionKey: "create_related.changed",
        },
      },
    ]) {
      const alteredConfig = replaceFeature(canonicalConfig, canonical, altered);
      expect(
        resolveSemanticInspectorFeature(alteredConfig, altered.featureGroupKey),
      ).toEqual({ kind: "unsupported" });
    }
  });

  it("returns the canonical config-owned feature when presentation labels change", () => {
    const canonicalConfig = requireViewContract(
      "cartulary.view.timeline.v2",
    ).inspectorConfig;
    const canonical = canonicalConfig.featureGroups[0];
    expect(canonical).toBeDefined();
    if (!canonical) return;
    const localized = { ...canonical, label: "Localized label" };
    const localizedConfig = replaceFeature(
      canonicalConfig,
      canonical,
      localized,
    );
    const resolution = resolveSemanticInspectorFeature(
      localizedConfig,
      localized.featureGroupKey,
    );

    expect(resolution.kind).toBe("supported");
    if (resolution.kind === "unsupported") return;
    expect(resolution.featureGroup).toBe(localized);
  });

  it("adds authorization_lost without hiding role-restricted features", () => {
    const contract = requireViewContract("cartulary.view.hosts.v1");
    const merge = contract.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "entity.merge",
    );
    expect(merge).toBeDefined();
    if (!merge) return;
    expect(
      inspectorFeatureDisabledTokens({
        currentIncidentRole: "editor",
        featureGroup: merge,
        stateTokens: new Set(),
      }),
    ).toContain("authorization_lost");
    expect(
      inspectorFeatureDisabledTokens({
        currentIncidentRole: "reviewer",
        featureGroup: merge,
        stateTokens: new Set(),
      }),
    ).not.toContain("authorization_lost");
  });
});

function expectedDisposition(
  featureGroup: InspectorFeatureGroup,
): SemanticInspectorDisposition {
  if (featureGroup.routeBinding.kind === "panel_read") return "panel_read";
  if (
    featureGroup.routeBinding.kind === "view_row_create" ||
    featureGroup.routeBinding.kind === "surface_pivot"
  ) {
    return "contextual_workflow_or_pivot";
  }
  if (
    featureGroup.panelId === "history" &&
    ["record.delete", "record.restore", "history.rollback"].includes(
      featureGroup.featureGroupKey,
    )
  ) {
    return "direct_history_action";
  }
  return "existing_owner_control";
}

function withFeatureGroups(
  config: InspectorConfig,
  featureGroups: readonly InspectorFeatureGroup[],
): InspectorConfig {
  return { ...config, featureGroups };
}

function replaceFeature(
  config: InspectorConfig,
  current: InspectorFeatureGroup,
  replacement: InspectorFeatureGroup,
): InspectorConfig {
  return withFeatureGroups(
    config,
    config.featureGroups.map((feature) =>
      feature === current ? replacement : feature,
    ),
  );
}
