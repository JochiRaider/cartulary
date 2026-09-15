import { useCallback, useLayoutEffect, useRef } from "react";
import type { WorkbookTimelineFileOwner } from "../../features/evidence/WorkbookTimelineFileOwner";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { TimelineEvidenceActionContext } from "../models/timelineEvidenceAttachmentPlan";
import type { WorkbookRow } from "../models/timelineRowModel";

export function useTimelineEvidenceAttach(input: {
  readonly actionContext: TimelineEvidenceActionContext;
  readonly owner: WorkbookTimelineFileOwner;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
}) {
  const current = useRef(input);
  current.current = input;
  useLayoutEffect(() => {
    input.owner.setPresentation(
      `${input.actionContext.surfaceKey}:${input.actionContext.selectedRowKey ?? "draft"}`,
    );
    return () => input.owner.setPresentation(null);
  }, [
    input.owner,
    input.actionContext.surfaceKey,
    input.actionContext.selectedRowKey,
  ]);
  const handleTimelineEvidenceFiles = useCallback(
    (row: WorkbookRow, files: FileList | File[]) => {
      const { actionContext, owner, setInspectorMessage } = current.current;
      if (!actionContext.authorized || !actionContext.capabilityAvailable)
        return;
      const message = owner.begin(
        {
          label:
            row.values.activitySynopsisText ||
            row.values.rawActivityText ||
            "Timeline draft",
          key: row.key,
          recordId: row.recordId,
          rowVersion: row.rowVersion,
        },
        files,
      );
      setInspectorMessage(
        message ? workbookInspectorMessageFeedback(message, "none") : null,
      );
    },
    [],
  );
  return {
    handleTimelineEvidenceFiles,
    attachEvidenceFileToTimeline: (row: WorkbookRow, file: File) =>
      handleTimelineEvidenceFiles(row, [file]),
  };
}
