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
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorLiveRowBinding,
} from "./workbookInspectorSubject";

const timeline = requireViewContract("cartulary.view.timeline.v2");

function liveTimelineSubject(): WorkbookInspectorLiveRowBinding {
  const subject = buildWorkbookInspectorSubject({
    config: timeline.inspectorConfig,
    kind: "live",
    label: "Timeline row",
    recordId: "11111111-1111-4111-8111-111111111111",
    rowVersion: 3,
    surfaceLabel: "Timeline",
  });
  if (subject?.kind !== "live") throw new Error("Live subject fixture");
  return { cells: {}, subject };
}

const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

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

  it("omits unbound additive creation and never dispatches a fallback", async () => {
    const feature = timeline.inspectorConfig.featureGroups.find(
      (f) => f.featureGroupKey === "create_related.task_request",
    );
    if (!feature) throw new Error("Missing authored feature");
    const unknown = { ...feature, featureGroupKey: "create_related.future" };
    const fetch = vi.spyOn(globalThis, "fetch");
    const selectedSubject = liveTimelineSubject();
    const { result, unmount } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({ selectedSubject }),
    );
    act(() => {
      expect(result.current.commands.begin(unknown)).toBe(false);
      result.current.commands.cancel();
    });
    expect(
      inspectorRelatedCreationOwner(timeline.viewSchemaId, unknown),
    ).toBeNull();
    expect(result.current.snapshot.workflow).toBeNull();
    await act(settle);
    expect(fetch).not.toHaveBeenCalled();
    unmount();
    fetch.mockRestore();
  });

  it("does not replace a missing retained owner with a generic form or transport", async () => {
    const fetch = vi.spyOn(globalThis, "fetch");
    const selectedSubject = liveTimelineSubject();
    const { result, unmount } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({ selectedSubject }),
    );
    const claimed: string[] = [];
    for (const feature of timeline.inspectorConfig.featureGroups) {
      act(() => {
        if (result.current.commands.begin(feature))
          claimed.push(feature.featureGroupKey);
      });
      expect(
        result.current.snapshot.workflow,
        feature.featureGroupKey,
      ).toBeNull();
    }
    expect(claimed.sort()).toEqual([
      "create_related.comm_log",
      "create_related.decision",
      "create_related.handoff",
      "create_related.lesson",
      "create_related.note",
      "create_related.status_review",
      "create_related.task_request",
    ]);
    await act(settle);
    expect(result.current.snapshot.workflow).toBeNull();
    expect(fetch).not.toHaveBeenCalled();
    expect(Object.keys(result.current.commands).sort()).toEqual([
      "begin",
      "cancel",
    ]);
    unmount();
    fetch.mockRestore();
  });
});
