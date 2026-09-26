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
import { workbookInspectorSubjectsEqual } from "../../inspector/workbookInspectorSubject";
import {
  type WorkbookInspectorState,
  workbookInspectorStateIsOpen,
} from "../../models/workbookInspectorModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";

export type TimelineInspectorFeatureLifecycle = {
  readonly authorizationKey: string;
  readonly inspector: WorkbookInspectorState;
  readonly lifecycleKey: string;
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
    const previousInspector = previousLifecycle.inspector;
    const inspector = lifecycle.inspector;
    // Observation drafts validate committed versions themselves. A new version
    // of the same live source must not dismiss the Relationships workflow.
    const sameObservationSource =
      workbookInspectorStateIsOpen(previousInspector) &&
      workbookInspectorStateIsOpen(inspector) &&
      previousInspector.subject?.kind === "live" &&
      inspector.subject?.kind === "live" &&
      previousInspector.subject.recordId === inspector.subject.recordId &&
      previousInspector.subject.viewSchemaId ===
        inspector.subject.viewSchemaId &&
      previousInspector.attachmentGeneration ===
        inspector.attachmentGeneration &&
      previousLifecycle.authorizationKey === lifecycle.authorizationKey &&
      previousLifecycle.lifecycleKey === lifecycle.lifecycleKey &&
      previousLifecycle.surfaceKey === lifecycle.surfaceKey &&
      inspector.invalidationCause === "record_updated";
    setIndicatorHandler((handler) =>
      handler?.action === "indicator.observations.manage" &&
      sameObservationSource
        ? handler
        : null,
    );
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
              ? {
                  ...workbookInspectorMessageFeedback(
                    "Inspector action is unavailable.",
                    "none",
                  ),
                  destination: {
                    kind: "panel",
                    panel: capability.featureGroup.panelId,
                  },
                }
              : null,
          );
          return;
        }
        case "note_create":
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
    left.inspector.reviewGeneration === right.inspector.reviewGeneration &&
    left.inspector.attachmentGeneration ===
      right.inspector.attachmentGeneration &&
    left.lifecycleKey === right.lifecycleKey &&
    left.inspector.phase === right.inspector.phase &&
    workbookInspectorSubjectsEqual(
      left.inspector.subject,
      right.inspector.subject,
    ) &&
    left.surfaceKey === right.surfaceKey
  );
}
