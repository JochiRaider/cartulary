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
import { useWorkbookInspectorCoordinator } from "../inspector/useWorkbookInspectorCoordinator";
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
const initialLifecycle = {
  authorizationKey: "editor:authorized",
  inspector: {
    reviewGeneration: 0,
    attachmentGeneration: 0,
    invalidationCause: null,
    phase: "open_ready",
    lifecycleKey: "inspector-1",
    subject: timelineSubject("row-a", 1),
  },
  lifecycleKey: "inspector-1:continuity-1",
  surfaceKey: "view_schema:cartulary.view.timeline.v2",
} satisfies TimelineInspectorFeatureLifecycle;

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
          { readonly kind: "create_related" | "note_create" }
        > =>
          capability.kind === "create_related" ||
          capability.kind === "note_create",
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
      inspector: {
        ...initialLifecycle.inspector,
        phase: "open_no_subject",
        subject: null,
      },
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

  it("resets Indicator and generic workflows on subject, surface, lifecycle, and authorization changes", () => {
    const { mocks, rerender, result } = controller();
    const lifecycleChanges: readonly TimelineInspectorFeatureLifecycle[] = [
      {
        ...initialLifecycle,
        inspector: {
          ...initialLifecycle.inspector,
          subject: timelineSubject("row-b", 1),
        },
      },
      {
        ...initialLifecycle,
        inspector: {
          ...initialLifecycle.inspector,
          subject: timelineSubject("row-c", 2),
        },
      },
      { ...initialLifecycle, surfaceKey: "saved_view:saved-view-2" },
      {
        ...initialLifecycle,
        lifecycleKey: "inspector-2:continuity-2",
      },
      {
        ...initialLifecycle,
        inspector: { ...initialLifecycle.inspector, reviewGeneration: 1 },
      },
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
  it("retains observation management across a same-source committed version while close and deletion retire it", () => {
    const { rerender, result, mocks } = controller();
    act(() => result.current.commands.handleFeatureAction(indicatorCapability));
    const updated = {
      ...initialLifecycle,
      inspector: {
        ...initialLifecycle.inspector,
        subject: timelineSubject("row-a", 2),
        reviewGeneration: 1,
        invalidationCause: "record_updated",
      },
    } satisfies TimelineInspectorFeatureLifecycle;
    rerender({ activeLifecycle: updated });
    expect(result.current.snapshot.indicatorHandler?.action).toBe(
      "indicator.observations.manage",
    );
    expect(mocks.cancelCreateRelatedWorkflow).toHaveBeenLastCalledWith(
      "lifecycle",
    );
    rerender({
      activeLifecycle: {
        ...updated,
        inspector: {
          ...updated.inspector,
          phase: "closed",
          reviewGeneration: 2,
          attachmentGeneration: 1,
          invalidationCause: "close",
        },
      },
    });
    expect(result.current.snapshot.indicatorHandler).toBeNull();
    rerender({ activeLifecycle: updated });
    act(() => result.current.commands.handleFeatureAction(indicatorCapability));
    rerender({
      activeLifecycle: {
        ...updated,
        inspector: {
          ...updated.inspector,
          subject: {
            ...timelineSubject("row-a", 2),
            kind: "deleted",
            stateLabel: "Deleted",
          },
          reviewGeneration: 3,
          attachmentGeneration: 1,
        },
      },
    });
    expect(result.current.snapshot.indicatorHandler).toBeNull();
  });

  it("retains observation management through the coordinator's committed snapshot transition", () => {
    const mocks = {
      beginCreateRelatedWorkflow: vi.fn(),
      cancelCreateRelatedWorkflow: vi.fn(),
      setInspectorMessage: vi.fn(),
    };
    const { result, rerender } = renderHook(
      ({ recordId, rowVersion }) => {
        const coordinator = useWorkbookInspectorCoordinator({
          config: requireViewContract(timelineViewSchemaId).inspectorConfig,
          lifecycleKey: "source",
          subject: timelineSubject(recordId, rowVersion),
          actionPorts: { resetOwnerState: () => {}, restoreFocus: () => {} },
        });
        const feature = useTimelineInspectorFeatureController({
          ...mocks,
          lifecycle: { ...initialLifecycle, inspector: coordinator.snapshot },
        });
        return { coordinator, feature };
      },
      { initialProps: { recordId: "row-a", rowVersion: 1 } },
    );
    act(() => result.current.coordinator.commands.open());
    act(() =>
      result.current.feature.commands.handleFeatureAction(indicatorCapability),
    );
    const attached = result.current.coordinator.snapshot.attachmentGeneration;
    rerender({ recordId: "row-a", rowVersion: 2 });
    expect(result.current.coordinator.snapshot.invalidationCause).toBe(
      "record_updated",
    );
    expect(result.current.coordinator.snapshot.attachmentGeneration).toBe(
      attached,
    );
    expect(result.current.feature.snapshot.indicatorHandler?.action).toBe(
      "indicator.observations.manage",
    );
    rerender({ recordId: "row-b", rowVersion: 1 });
    expect(result.current.feature.snapshot.indicatorHandler).toBeNull();
  });
});

function timelineSubject(recordId: string, rowVersion: number) {
  return {
    kind: "live" as const,
    label: "Timeline row",
    recordId,
    rowVersion,
    surfaceLabel: "Timeline",
    viewSchemaId: timelineViewSchemaId,
  };
}

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
