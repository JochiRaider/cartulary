import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import type { IndicatorInspectorHandler } from "../../features/indicators/indicatorInspectorHandlers";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { resolveTimelineWorkbookFeature } from "../models/timelineWorkbookFeaturePolicy";

export type TimelineInspectorFeatureLifecycle = {
  readonly authorizationKey: string;
  readonly invalidationGeneration: number;
  readonly lifecycleKey: string;
  readonly subjectKey: string;
  readonly surfaceKey: string;
};

export function useTimelineInspectorFeatureController({
  beginCreateRelatedWorkflow,
  cancelCreateRelatedWorkflow,
  lifecycle,
  setInspectorMessage,
}: {
  readonly beginCreateRelatedWorkflow: (
    featureGroup: InspectorFeatureGroup,
  ) => void;
  readonly cancelCreateRelatedWorkflow: () => void;
  readonly lifecycle: TimelineInspectorFeatureLifecycle;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
}) {
  const [indicatorHandler, setIndicatorHandler] =
    useState<IndicatorInspectorHandler | null>(null);
  const previousLifecycleRef = useRef(lifecycle);

  const cancelFeatureAction = useCallback(() => {
    setIndicatorHandler(null);
    cancelCreateRelatedWorkflow();
    setInspectorMessage(null);
  }, [cancelCreateRelatedWorkflow, setInspectorMessage]);

  useLayoutEffect(() => {
    if (
      sameTimelineInspectorFeatureLifecycle(
        previousLifecycleRef.current,
        lifecycle,
      )
    ) {
      return;
    }
    previousLifecycleRef.current = lifecycle;
    setIndicatorHandler(null);
    cancelCreateRelatedWorkflow();
    setInspectorMessage(null);
  }, [cancelCreateRelatedWorkflow, lifecycle, setInspectorMessage]);

  const handleFeatureAction = useCallback(
    (featureGroup: InspectorFeatureGroup) => {
      const resolution = resolveTimelineWorkbookFeature(
        timelineViewSchemaId,
        featureGroup.featureGroupKey,
      );
      if (resolution.kind === "indicator") {
        cancelCreateRelatedWorkflow();
        setIndicatorHandler(resolution.handler);
        setInspectorMessage(null);
        return;
      }
      setIndicatorHandler(null);
      if (resolution.kind === "create_related") {
        beginCreateRelatedWorkflow(resolution.featureGroup);
        return;
      }
      if (resolution.kind === "panel_owned") {
        cancelCreateRelatedWorkflow();
        setInspectorMessage(null);
        return;
      }
      cancelCreateRelatedWorkflow();
      setInspectorMessage("Inspector action is unavailable.");
    },
    [
      beginCreateRelatedWorkflow,
      cancelCreateRelatedWorkflow,
      setInspectorMessage,
    ],
  );

  return {
    commands: {
      cancelFeatureAction,
      handleFeatureAction,
    },
    snapshot: {
      indicatorHandler,
    },
  };
}

function sameTimelineInspectorFeatureLifecycle(
  left: TimelineInspectorFeatureLifecycle,
  right: TimelineInspectorFeatureLifecycle,
): boolean {
  return (
    left.authorizationKey === right.authorizationKey &&
    left.invalidationGeneration === right.invalidationGeneration &&
    left.lifecycleKey === right.lifecycleKey &&
    left.subjectKey === right.subjectKey &&
    left.surfaceKey === right.surfaceKey
  );
}
