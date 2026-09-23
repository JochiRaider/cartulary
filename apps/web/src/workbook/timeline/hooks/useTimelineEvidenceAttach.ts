import { useCallback, useLayoutEffect, useRef } from "react";
import type { TimelineFileSource } from "../../features/evidence/timelineFileOperation";
import type { WorkbookTimelineFileOwner } from "../../features/evidence/WorkbookTimelineFileOwner";
import {
  targetWorkbookInspectorFeedback,
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { TimelineEvidenceActionContext } from "../models/timelineEvidenceAttachmentPlan";

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
    (source: TimelineFileSource, files: FileList | readonly File[]) => {
      const { actionContext, owner, setInspectorMessage } = current.current;
      if (!actionContext.authorized || !actionContext.capabilityAvailable)
        return;
      const message = owner.begin(source, files);
      setInspectorMessage(
        message && source.recordId
          ? targetWorkbookInspectorFeedback(
              workbookInspectorMessageFeedback(message, "none"),
              source.recordId,
              {
                kind: "region",
                panel: "evidence",
                regionId: "evidence-metadata",
              },
            )
          : null,
      );
    },
    [],
  );
  return {
    handleTimelineEvidenceFiles,
  };
}
