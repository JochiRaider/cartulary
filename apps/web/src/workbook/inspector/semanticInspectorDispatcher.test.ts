import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  assertCurrentInspectorDispatchCompleteness,
  inspectorFeatureDisabledTokens,
  resolveSemanticInspectorFeature,
  semanticInspectorFeatureKey,
} from "./semanticInspectorDispatcher";

describe("semantic inspector dispatcher", () => {
  it("resolves every current projected feature exactly once", () => {
    expect(assertCurrentInspectorDispatchCompleteness).not.toThrow();
    const features = listViewContracts().flatMap((contract) =>
      contract.inspectorConfig.featureGroups.map((featureGroup) =>
        semanticInspectorFeatureKey(contract.viewSchemaId, featureGroup),
      ),
    );
    expect(new Set(features).size).toBe(features.length);
  });

  it("omits an unknown additive feature instead of inferring from its label", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const canonical = contract.inspectorConfig.featureGroups[0];
    expect(canonical).toBeDefined();
    if (!canonical) return;
    expect(
      resolveSemanticInspectorFeature(contract.inspectorConfig, {
        ...canonical,
        featureGroupKey: "future.unknown_action",
        label: canonical.label,
      }),
    ).toEqual({ kind: "unsupported" });
  });

  it("matches the complete semantic contract while treating labels as presentation", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const canonical = contract.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "create_related.note",
    );
    expect(canonical).toBeDefined();
    if (!canonical) return;
    expect(
      resolveSemanticInspectorFeature(contract.inspectorConfig, {
        ...canonical,
        label: "Localized label",
      }).kind,
    ).toBe("action");
    for (const altered of [
      { ...canonical, minimumIncidentRole: "admin" as const },
      { ...canonical, mutates: !canonical.mutates },
      { ...canonical, requiresConfirmation: !canonical.requiresConfirmation },
      {
        ...canonical,
        disabledWhen: [...canonical.disabledWhen, "record_deleted" as const],
      },
      {
        ...canonical,
        seedBindings: canonical.seedBindings.slice(1),
      },
      {
        ...canonical,
        successResultBehavior: "clear_to_no_row_selected" as const,
      },
      {
        ...canonical,
        failureResultBehavior:
          "show_same_shell_error_invalidate_pending_action" as const,
      },
    ]) {
      expect(
        resolveSemanticInspectorFeature(contract.inspectorConfig, altered),
      ).toEqual({ kind: "unsupported" });
    }
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
