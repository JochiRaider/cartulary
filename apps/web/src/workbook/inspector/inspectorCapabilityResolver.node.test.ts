import {
  type InspectorConfig,
  type InspectorFeatureGroup,
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  admitCanonicalInspectorFeature,
  inspectorFeatureIdentity,
} from "./canonicalInspectorAdmission";
import {
  inspectorContextualCapabilities,
  inspectorRecordHistoryActions,
} from "./inspectorCapabilityResolver";

afterEach(() => {
  vi.doUnmock("@cartulary/view-contracts");
  vi.resetModules();
});

describe("inspector capability resolver", () => {
  it("admits all 17 schemas and 247 exact tuples as canonical objects", () => {
    const identities = new Set<string>();
    let featureCount = 0;

    for (const contract of listViewContracts()) {
      for (const featureGroup of contract.inspectorConfig.featureGroups) {
        expect(
          admitCanonicalInspectorFeature(
            contract.inspectorConfig,
            featureGroup.featureGroupKey,
          ),
        ).toBe(featureGroup);
        identities.add(
          inspectorFeatureIdentity(contract.viewSchemaId, featureGroup),
        );
        featureCount += 1;
      }
    }

    expect(listViewContracts()).toHaveLength(17);
    expect(featureCount).toBe(247);
    expect(identities.size).toBe(247);
  });

  it("classifies exactly 37 create, 4 Note, 4 Indicator, 51 history, 1 Decision action, 2 Timeline actions, and 148 non-contextual features", () => {
    const counts = {
      timeline_capture: 0,
      create_related: 0,
      note_create: 0,
      decision_supersede: 0,
      indicator: 0,
      non_contextual: 0,
      record_history: 0,
    };

    for (const contract of listViewContracts()) {
      const contextual = contract.inspectorConfig.panels.flatMap((panel) =>
        inspectorContextualCapabilities({
          config: contract.inspectorConfig,
          panelId: panel.panelId,
        }),
      );
      for (const capability of contextual) {
        counts[capability.kind] += 1;
        expect(
          contract.inspectorConfig.featureGroups.includes(
            capability.featureGroup,
          ),
        ).toBe(true);
      }
      const contextualIdentities = new Set(
        contextual.map((capability) =>
          inspectorFeatureIdentity(
            contract.viewSchemaId,
            capability.featureGroup,
          ),
        ),
      );
      let historyFeatures = 0;
      for (const featureGroup of contract.inspectorConfig.featureGroups) {
        const identity = inspectorFeatureIdentity(
          contract.viewSchemaId,
          featureGroup,
        );
        const isContextual = contextualIdentities.has(identity);
        const historyActions = inspectorRecordHistoryActions(
          withFeatureGroups(contract.inspectorConfig, [featureGroup]),
        ).size;
        expect(historyActions, identity).toBeLessThanOrEqual(1);
        const isHistory = historyActions === 1;
        expect(isContextual && isHistory, identity).toBe(false);
        if (isHistory) historyFeatures += 1;
        if (!isContextual && !isHistory) counts.non_contextual += 1;
      }
      expect(contextualIdentities.size, contract.viewSchemaId).toBe(
        contextual.length,
      );
      expect(historyFeatures, contract.viewSchemaId).toBe(
        inspectorRecordHistoryActions(contract.inspectorConfig).size,
      );
      counts.record_history += historyFeatures;
    }

    expect(counts).toEqual({
      create_related: 37,
      note_create: 4,
      indicator: 4,
      non_contextual: 148,
      timeline_capture: 2,
      decision_supersede: 1,
      record_history: 51,
    });
  });

  it("isolates additive, altered, and duplicate addressed tuples without losing siblings", () => {
    const canonicalConfig = requireViewContract(
      "cartulary.view.timeline.v2",
    ).inspectorConfig;
    const canonical = requireFeature(canonicalConfig, "create_related.note");
    const sibling = requireFeature(canonicalConfig, "details.read");
    const additive = { ...canonical, featureGroupKey: "future.unknown_action" };
    const additiveConfig = withFeatureGroups(canonicalConfig, [
      ...canonicalConfig.featureGroups,
      additive,
    ]);

    expect(
      admitCanonicalInspectorFeature(additiveConfig, additive.featureGroupKey),
    ).toBeNull();
    expect(
      admitCanonicalInspectorFeature(additiveConfig, canonical.featureGroupKey),
    ).toBe(canonical);

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
        admitCanonicalInspectorFeature(alteredConfig, altered.featureGroupKey),
      ).toBeNull();
      expect(
        admitCanonicalInspectorFeature(alteredConfig, sibling.featureGroupKey),
      ).toBe(sibling);
    }

    const duplicateConfig = withFeatureGroups(canonicalConfig, [
      ...canonicalConfig.featureGroups,
      canonical,
    ]);
    expect(
      admitCanonicalInspectorFeature(
        duplicateConfig,
        canonical.featureGroupKey,
      ),
    ).toBeNull();
    expect(
      admitCanonicalInspectorFeature(duplicateConfig, sibling.featureGroupKey),
    ).toBe(sibling);
  });

  it("admits label-only variants but returns the canonical projected feature", () => {
    const canonicalConfig = requireViewContract(
      "cartulary.view.timeline.v2",
    ).inspectorConfig;
    const canonical = requireFeature(canonicalConfig, "create_related.note");
    const localizedConfig = replaceFeature(canonicalConfig, canonical, {
      ...canonical,
      label: "Localized label",
    });

    expect(
      admitCanonicalInspectorFeature(
        localizedConfig,
        canonical.featureGroupKey,
      ),
    ).toBe(canonical);
    expect(
      inspectorContextualCapabilities({
        config: localizedConfig,
        panelId: canonical.panelId,
      }).find(
        (capability) =>
          capability.featureGroup.featureGroupKey === canonical.featureGroupKey,
      )?.featureGroup,
    ).toBe(canonical);
  });

  it("omits a canonically declared future confirmed capability", async () => {
    const actual = await vi.importActual<
      typeof import("@cartulary/view-contracts")
    >("@cartulary/view-contracts");
    const contract = actual.requireViewContract("cartulary.view.timeline.v2");
    const confirmedFeature = {
      ...requireFeature(contract.inspectorConfig, "create_related.note"),
      requiresConfirmation: true,
    };
    const mockedContract = {
      ...contract,
      inspectorConfig: replaceFeature(
        contract.inspectorConfig,
        requireFeature(contract.inspectorConfig, "create_related.note"),
        confirmedFeature,
      ),
    };
    vi.resetModules();
    vi.doMock("@cartulary/view-contracts", () => ({
      ...actual,
      listViewContracts: () => [mockedContract],
    }));
    const resolver = await import("./inspectorCapabilityResolver");

    expect(
      resolver
        .inspectorContextualCapabilities({
          config: mockedContract.inspectorConfig,
          panelId: confirmedFeature.panelId,
        })
        .some(
          (capability) =>
            capability.featureGroup.featureGroupKey ===
            confirmedFeature.featureGroupKey,
        ),
    ).toBe(false);
  });
});

function requireFeature(
  config: InspectorConfig,
  featureGroupKey: string,
): InspectorFeatureGroup {
  const featureGroup = config.featureGroups.find(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (featureGroup === undefined) {
    throw new Error(`Missing inspector feature ${featureGroupKey}`);
  }
  return featureGroup;
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
