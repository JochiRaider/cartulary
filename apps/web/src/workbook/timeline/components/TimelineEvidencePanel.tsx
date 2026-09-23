import {
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import {
  type RefCallback,
  useContext,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { EvidenceAttachmentEntry } from "../../features/evidence/EvidenceAttachmentEntry";
import type { TimelineFileSnapshot } from "../../features/evidence/WorkbookTimelineFileOwner";
import { TimelineAttachmentDetails } from "./TimelineAttachmentFeedback";

const noFileSubscription = () => () => {};
const emptyFiles: readonly TimelineFileSnapshot[] = [];
const noFiles = () => emptyFiles;

import type { TimelineFileSource } from "../../features/evidence/timelineFileOperation";
import { captureTimelineFileSource } from "../models/timelineEvidenceAttachmentPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import { bodyStyle, inspectorSectionStyle } from "./TimelineWorkbookStyles";

type TimelineEvidenceCountDisplay = {
  readonly displayCount: string;
  readonly stateKey: string;
};

type TimelineEvidencePanelProps = {
  readonly countDisplay: TimelineEvidenceCountDisplay;
  readonly elementRef?: RefCallback<HTMLElement> | undefined;
  readonly row: WorkbookRow;
  readonly onFilesSelected: (
    source: TimelineFileSource,
    files: FileList | readonly File[],
  ) => void;
};

export function TimelineEvidencePanel({
  countDisplay,
  elementRef,
  row,
  onFilesSelected,
}: TimelineEvidencePanelProps) {
  const owner = useContext(TimelineFileContext);
  const panel = useRef<HTMLElement>(null);
  const files = useSyncExternalStore(
    owner?.subscribe ?? noFileSubscription,
    owner?.getSnapshot ?? noFiles,
  );
  const recordId = row.recordId;
  // Keep the local fallback ref attached while the Inspector replaces its registry callback.
  useLayoutEffect(() => {
    const cleanup = elementRef?.(recordId === null ? null : panel.current);
    return () => {
      if (typeof cleanup === "function") cleanup();
      else elementRef?.(null);
    };
  }, [elementRef, recordId]);
  const disabledReason =
    owner?.attachmentDisabledReason() ??
    (owner ? null : "File attachment is unavailable.");
  const source = captureTimelineFileSource(row);
  const attach = (files: readonly File[]) => {
    if (!owner || owner.attachmentDisabledReason() !== null) return;
    onFilesSelected(source, files);
  };
  if (recordId === null) {
    return null;
  }

  return (
    <section
      ref={panel}
      tabIndex={-1}
      data-testid={timelineInspectorSectionTestId("evidence")}
      data-evidence-count-state={countDisplay.stateKey}
      style={inspectorSectionStyle}
      aria-label="Timeline evidence attachment"
    >
      <section aria-label="Evidence information">
        <p style={bodyStyle}>
          Attached evidence count:{" "}
          {countDisplay.stateKey === "inconsistent"
            ? "Unavailable"
            : countDisplay.displayCount}
        </p>
        <p style={bodyStyle}>
          Linked Evidence records. File access is checked separately on each
          Evidence record.
        </p>
      </section>
      <EvidenceAttachmentEntry
        title="this Timeline record"
        regionTestId={timelineEvidenceAttachSectionTestId(recordId)}
        testId={timelineEvidenceFileInputTestId(recordId)}
        disabledReason={disabledReason}
        busy={false}
        onAttach={attach}
      />
      {files.some((entry) => entry.recordId === recordId) ? (
        <section aria-label="Attachment progress and recovery">
          {owner ? (
            <TimelineAttachmentDetails
              owner={owner}
              files={files.filter((entry) => entry.recordId === recordId)}
              presentation="inspector"
              fallbackRef={panel}
            />
          ) : null}
        </section>
      ) : null}
    </section>
  );
}
