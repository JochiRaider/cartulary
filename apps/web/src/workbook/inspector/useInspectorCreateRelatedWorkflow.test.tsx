import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { isContextualCreateFeature } from "../features/coordination/contextualCreateModel";
import { coordinationVariant } from "../features/coordination/coordinationCreateModel";
import { relatedEvidenceFeature } from "../features/evidence/timelineRelatedEvidenceModel";
import { noteCreateFeature } from "../features/notes/noteCreateModel";
import {
  inspectorContextualCapabilities,
  inspectorRelatedCreationOwner,
} from "./inspectorCapabilityResolver";
import { useInspectorCreateRelatedWorkflow } from "./useInspectorCreateRelatedWorkflow";

const timeline = requireViewContract("cartulary.view.timeline.v2");

describe("useInspectorCreateRelatedWorkflow", () => {
  it("binds every authored creation feature to exactly one retained operation owner", () => {
    const owners = new Set<string>();
    for (const contract of listViewContracts()) {
      const declared = contract.inspectorConfig.featureGroups.filter(
        (feature) =>
          feature.routeBinding.kind === "view_row_create" ||
          feature.featureGroupKey === noteCreateFeature,
      );
      for (const feature of declared) {
        // Independent feature-owner predicates must partition the authored projection.
        const matches = [
          isContextualCreateFeature(feature.featureGroupKey) && "contextual",
          coordinationVariant(feature.featureGroupKey) && "coordination",
          feature.featureGroupKey === noteCreateFeature && "note",
          contract.viewSchemaId === timeline.viewSchemaId &&
            feature.featureGroupKey === relatedEvidenceFeature &&
            "timeline_evidence",
          contract.viewSchemaId === "cartulary.view.assessments.v1" &&
            feature.featureGroupKey === "create_related.assessment" &&
            "assessment",
        ].filter(Boolean);
        expect(
          matches,
          `${contract.viewSchemaId}:${feature.featureGroupKey}`,
        ).toHaveLength(1);
        expect(
          inspectorRelatedCreationOwner(contract.viewSchemaId, feature),
        ).toBe(matches[0]);
        const capabilities = inspectorContextualCapabilities({
          config: contract.inspectorConfig,
          panelId: feature.panelId,
        });
        expect(
          capabilities.filter((c) => c.featureGroup === feature),
        ).toHaveLength(1);
        owners.add(String(matches[0]));
      }
    }
    expect([...owners].sort()).toEqual([
      "assessment",
      "contextual",
      "coordination",
      "note",
      "timeline_evidence",
    ]);
  });

  it("omits unbound additive creation and never dispatches a fallback", () => {
    const feature = timeline.inspectorConfig.featureGroups.find(
      (f) => f.featureGroupKey === "create_related.task_request",
    );
    if (!feature) throw new Error("Missing authored feature");
    const unknown = { ...feature, featureGroupKey: "create_related.future" };
    const fetch = vi.spyOn(globalThis, "fetch");
    const { result, unmount } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({ selectedSubject: null }),
    );
    act(() => {
      expect(result.current.commands.begin(unknown)).toBe(false);
      result.current.commands.cancel();
    });
    expect(
      inspectorRelatedCreationOwner(timeline.viewSchemaId, unknown),
    ).toBeNull();
    expect(result.current.snapshot.workflow).toBeNull();
    expect(fetch).not.toHaveBeenCalled();
    unmount();
    fetch.mockRestore();
  });

  it("does not replace a missing retained owner with a generic form or transport", () => {
    const { result } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({ selectedSubject: null }),
    );
    for (const feature of timeline.inspectorConfig.featureGroups) {
      act(() => result.current.commands.begin(feature));
      expect(result.current.snapshot.workflow).toBeNull();
    }
    expect(Object.keys(result.current.commands).sort()).toEqual([
      "begin",
      "cancel",
    ]);
  });
});
