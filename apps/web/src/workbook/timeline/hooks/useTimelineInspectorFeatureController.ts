import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import {
  type IndicatorInspectorHandler,
  resolveIndicatorInspectorHandler,
} from "../../features/indicators/indicatorInspectorHandlers";
import type { InspectorContextualCapability } from "../../inspector/inspectorCapabilityResolver";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  type WorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "../../inspector/workbookInspectorSubject";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";

export type TimelineInspectorFeatureLifecycle = {
  readonly authorizationKey: string;
  readonly invalidationGeneration: number;
  readonly lifecycleKey: string;
  readonly subject: WorkbookInspectorSubject | null;
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
  readonly cancelCreateRelatedWorkflow: (
    reason?: "owner_action" | "lifecycle",
  ) => void;
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
    const previousLifecycle = previousLifecycleRef.current;
    if (sameTimelineInspectorFeatureLifecycle(previousLifecycle, lifecycle)) {
      return;
    }
    previousLifecycleRef.current = lifecycle;
    setIndicatorHandler(null);
    cancelCreateRelatedWorkflow("lifecycle");
    setInspectorMessage(null);
  }, [cancelCreateRelatedWorkflow, lifecycle, setInspectorMessage]);

  const handleFeatureAction = useCallback(
    (capability: InspectorContextualCapability) => {
      switch (capability.kind) {
        case "indicator": {
          const handler = resolveIndicatorInspectorHandler(
            timelineViewSchemaId,
            capability.featureGroup,
          );
          cancelCreateRelatedWorkflow();
          setIndicatorHandler(handler);
          setInspectorMessage(
            handler === null
              ? workbookInspectorMessageFeedback(
                  "Inspector action is unavailable.",
                  "none",
                )
              : null,
          );
          return;
        }
        case "create_related":
          setIndicatorHandler(null);
          beginCreateRelatedWorkflow(capability.featureGroup);
          return;
      }
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
    workbookInspectorSubjectsEqual(left.subject, right.subject) &&
    left.surfaceKey === right.surfaceKey
  );
}
