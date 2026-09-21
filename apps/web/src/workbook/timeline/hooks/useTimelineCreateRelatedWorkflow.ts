import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import { useCallback, useRef } from "react";
import { useContextualCreateAttachment } from "../../features/coordination/useContextualCreateAttachment";
import { useCoordinationCreateAttachment } from "../../features/coordination/useCoordinationCreateAttachment";
import { useTimelineRelatedEvidenceAttachment } from "../../features/evidence/useTimelineRelatedEvidenceAttachment";
import { useNoteCreateAttachment } from "../../features/notes/useNoteCreateAttachment";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { WorkbookInspectorLiveSubject } from "../../inspector/workbookInspectorSubject";
import type { WorkbookRow } from "../models/timelineRowModel";

type TimelineCreateRelatedWorkflowInput = {
  readonly isInspectorOpen?: boolean;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedSubject: WorkbookInspectorLiveSubject | null;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
};

/** Presentations attach to their retained owners; unknown additive actions are omitted. */
export function useTimelineCreateRelatedWorkflow(
  input: TimelineCreateRelatedWorkflowInput,
) {
  const note = useNoteCreateAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
    input.isInspectorOpen ?? true,
  );
  const coordination = useCoordinationCreateAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
    input.isInspectorOpen ?? true,
  );
  const evidence = useTimelineRelatedEvidenceAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
    input.isInspectorOpen ?? true,
  );
  const {
    workflow: contextualWorkflow,
    begin: contextualBegin,
    detach: contextualDetach,
    update: contextualUpdate,
  } = useContextualCreateAttachment(
    input.selectedSubject && input.selectedRow
      ? {
          subject: input.selectedSubject,
          cells: input.selectedRow.rawRow?.cells ?? {},
        }
      : null,
  );

  const current = useRef(input);
  current.current = input;
  const cancelWorkflow = useCallback(
    (reason: "owner_action" | "lifecycle" = "owner_action") => {
      note.detach();
      coordination.detach();
      contextualDetach();
      if (reason === "owner_action") evidence.detach();
    },
    [note.detach, coordination.detach, contextualDetach, evidence.detach],
  );
  const workflow =
    coordination.workflow ??
    note.workflow ??
    evidence.workflow ??
    contextualWorkflow;
  const beginWorkflow = useCallback(
    (feature: InspectorFeatureGroup) => {
      if (
        workflow &&
        workflow.featureGroup.featureGroupKey !== feature.featureGroupKey
      )
        cancelWorkflow();
      if (
        coordination.begin(feature) ||
        note.begin(feature) ||
        contextualBegin(feature)
      )
        return;
      if (evidence.begin(feature)) {
        const message = evidence.notice();
        if (message)
          current.current.setInspectorMessage({
            ...workbookInspectorMessageFeedback(message, "none"),
            destination: {
              kind: "feature",
              panel: feature.panelId,
              featureGroupKey: feature.featureGroupKey,
            },
            ...(current.current.selectedSubject
              ? { sourceRecordId: current.current.selectedSubject.recordId }
              : {}),
          });
      }
    },
    [
      workflow,
      cancelWorkflow,
      coordination.begin,
      note.begin,
      contextualBegin,
      evidence.begin,
      evidence.notice,
    ],
  );
  const updateWorkflowDraft = useCallback(
    (feature: string, field: string, value: string) => {
      if (coordination.workflow?.featureGroup.featureGroupKey === feature)
        coordination.update(field, value);
      else if (note.workflow?.featureGroup.featureGroupKey === feature)
        note.update(field, value);
      else if (contextualWorkflow?.featureGroup.featureGroupKey === feature)
        contextualUpdate(field, value);
    },
    [
      coordination.workflow,
      coordination.update,
      note.workflow,
      note.update,
      contextualWorkflow,
      contextualUpdate,
    ],
  );
  // Each target form dispatches directly through its owner and attachment lease.
  const submitWorkflow = useCallback(async () => {}, []);
  return {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow,
  };
}
