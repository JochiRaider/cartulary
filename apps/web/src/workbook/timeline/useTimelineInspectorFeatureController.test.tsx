import {
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  type TimelineInspectorFeatureLifecycle,
  useTimelineInspectorFeatureController,
} from "./hooks/useTimelineInspectorFeatureController";
import { resolveTimelineWorkbookFeature } from "./models/timelineWorkbookFeaturePolicy";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const indicatorFeature = requireTimelineFeature(
  "indicator.observations.manage",
);
const createRelatedFeature = requireTimelineFeature("create_related.note");
const initialLifecycle: TimelineInspectorFeatureLifecycle = {
  authorizationKey: "editor:authorized",
  invalidationGeneration: 0,
  lifecycleKey: "inspector-1:continuity-1",
  subjectKey: "row-a:1",
  surfaceKey: "view_schema:cartulary.view.timeline.v2",
};

function controller(
  lifecycle: TimelineInspectorFeatureLifecycle = initialLifecycle,
) {
  const mocks = {
    beginCreateRelatedWorkflow: vi.fn(),
    cancelCreateRelatedWorkflow: vi.fn(),
    setInspectorMessage: vi.fn(),
  };
  const rendered = renderHook(
    ({ activeLifecycle }) =>
      useTimelineInspectorFeatureController({
        ...mocks,
        lifecycle: activeLifecycle,
      }),
    { initialProps: { activeLifecycle: lifecycle } },
  );
  return { ...rendered, mocks };
}

describe("useTimelineInspectorFeatureController", () => {
  it("routes canonical features with Indicator precedence and no generic fallback", () => {
    const { mocks, result } = controller();
    const supportedFeatures = timelineFeatures.filter(
      (featureGroup) =>
        resolveTimelineWorkbookFeature(
          timelineViewSchemaId,
          featureGroup.featureGroupKey,
        ).kind !== "unsupported",
    );

    expect(supportedFeatures).toHaveLength(timelineFeatures.length);
    const createRelatedFeatures = supportedFeatures.filter(
      (featureGroup) => featureGroup.routeBinding.kind === "view_row_create",
    );
    for (const featureGroup of createRelatedFeatures) {
      act(() => result.current.commands.handleFeatureAction(featureGroup));
    }
    act(() => result.current.commands.handleFeatureAction(indicatorFeature));

    expect(
      mocks.beginCreateRelatedWorkflow.mock.calls.map(([feature]) => feature),
    ).toEqual(createRelatedFeatures);
    expect(mocks.beginCreateRelatedWorkflow).not.toHaveBeenCalledWith(
      indicatorFeature,
    );
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledOnce();
    expect(result.current.snapshot.indicatorHandler).toEqual({
      action: "indicator.observations.manage",
      panelId: "relationships",
    });
  });

  it("fails closed while owning cancellation, unavailable messages, and mutually exclusive selection", () => {
    const { mocks, result } = controller({
      ...initialLifecycle,
      authorizationKey: "viewer:authorized",
      subjectKey: "",
    });
    act(() => result.current.commands.handleFeatureAction(indicatorFeature));
    expect(result.current.snapshot.indicatorHandler).not.toBeNull();
    act(() =>
      result.current.commands.handleFeatureAction(createRelatedFeature),
    );
    expect(result.current.snapshot.indicatorHandler).toBeNull();
    expect(mocks.beginCreateRelatedWorkflow).toHaveBeenLastCalledWith(
      createRelatedFeature,
    );

    const unsupportedFeature = {
      ...createRelatedFeature,
      featureGroupKey: "create_related.unknown",
    } satisfies InspectorFeatureGroup;
    act(() => result.current.commands.handleFeatureAction(unsupportedFeature));
    expect(
      resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        unsupportedFeature.featureGroupKey,
      ),
    ).toEqual({ kind: "unsupported" });
    expect(mocks.beginCreateRelatedWorkflow).toHaveBeenCalledTimes(1);
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledTimes(2);
    expect(mocks.setInspectorMessage).toHaveBeenLastCalledWith(
      "Inspector action is unavailable.",
    );

    act(() => result.current.commands.cancelFeatureAction());
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledTimes(3);
    expect(mocks.setInspectorMessage).toHaveBeenLastCalledWith(null);
  });

  it("resets Indicator and generic workflows on subject, version, surface, lifecycle, and authorization changes", () => {
    const { mocks, rerender, result } = controller();
    const lifecycleChanges: readonly TimelineInspectorFeatureLifecycle[] = [
      { ...initialLifecycle, subjectKey: "row-b:1" },
      { ...initialLifecycle, subjectKey: "row-b:2" },
      { ...initialLifecycle, surfaceKey: "saved_view:saved-view-2" },
      {
        ...initialLifecycle,
        lifecycleKey: "inspector-2:continuity-2",
      },
      { ...initialLifecycle, invalidationGeneration: 1 },
      { ...initialLifecycle, authorizationKey: "none:access-lost" },
    ];

    for (const activeLifecycle of lifecycleChanges) {
      act(() => result.current.commands.handleFeatureAction(indicatorFeature));
      expect(result.current.snapshot.indicatorHandler).not.toBeNull();
      rerender({ activeLifecycle });
      expect(result.current.snapshot.indicatorHandler).toBeNull();
    }
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledTimes(
      lifecycleChanges.length * 2,
    );
    expect(mocks.setInspectorMessage).toHaveBeenLastCalledWith(null);
  });
});

function requireTimelineFeature(
  featureGroupKey: string,
): InspectorFeatureGroup {
  const featureGroup = timelineFeatures.find(
    (candidate) => candidate.featureGroupKey === featureGroupKey,
  );
  if (featureGroup === undefined) {
    throw new Error(`Missing Timeline feature ${featureGroupKey}`);
  }
  return featureGroup;
}
