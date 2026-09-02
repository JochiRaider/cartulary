import {
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  type InspectorContextualCapability,
  inspectorContextualCapabilities,
} from "../inspector/inspectorCapabilityResolver";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  type TimelineInspectorFeatureLifecycle,
  useTimelineInspectorFeatureController,
} from "./hooks/useTimelineInspectorFeatureController";

const timelineFeatures =
  requireViewContract(timelineViewSchemaId).inspectorConfig.featureGroups;
const indicatorFeature = requireTimelineFeature(
  "indicator.observations.manage",
);
const createRelatedFeature = requireTimelineFeature("create_related.note");
const indicatorCapability = requireTimelineCapability(
  "indicator.observations.manage",
);
const createRelatedCapability = requireTimelineCapability(
  "create_related.note",
);
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
    const config = requireViewContract(timelineViewSchemaId).inspectorConfig;
    const createRelatedCapabilities = config.panels
      .flatMap((panel) =>
        inspectorContextualCapabilities({ config, panelId: panel.panelId }),
      )
      .filter(
        (
          capability,
        ): capability is Extract<
          InspectorContextualCapability,
          { readonly kind: "create_related" }
        > => capability.kind === "create_related",
      );
    for (const capability of createRelatedCapabilities) {
      act(() => result.current.commands.handleFeatureAction(capability));
    }
    act(() => result.current.commands.handleFeatureAction(indicatorCapability));

    expect(
      mocks.beginCreateRelatedWorkflow.mock.calls.map(([feature]) => feature),
    ).toEqual(
      createRelatedCapabilities.map((capability) => capability.featureGroup),
    );
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
    act(() => result.current.commands.handleFeatureAction(indicatorCapability));
    expect(result.current.snapshot.indicatorHandler).not.toBeNull();
    act(() =>
      result.current.commands.handleFeatureAction(createRelatedCapability),
    );
    expect(result.current.snapshot.indicatorHandler).toBeNull();
    expect(mocks.beginCreateRelatedWorkflow).toHaveBeenLastCalledWith(
      createRelatedFeature,
    );

    expect(mocks.beginCreateRelatedWorkflow).toHaveBeenCalledTimes(1);
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledOnce();

    act(() => result.current.commands.cancelFeatureAction());
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenCalledTimes(2);
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
      act(() =>
        result.current.commands.handleFeatureAction(indicatorCapability),
      );
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

function requireTimelineCapability(
  featureGroupKey: string,
): InspectorContextualCapability {
  const config = requireViewContract(timelineViewSchemaId).inspectorConfig;
  const capability = config.panels
    .flatMap((panel) =>
      inspectorContextualCapabilities({ config, panelId: panel.panelId }),
    )
    .find(
      (candidate) => candidate.featureGroup.featureGroupKey === featureGroupKey,
    );
  if (capability === undefined) {
    throw new Error(`Missing Timeline capability ${featureGroupKey}`);
  }
  return capability;
}
