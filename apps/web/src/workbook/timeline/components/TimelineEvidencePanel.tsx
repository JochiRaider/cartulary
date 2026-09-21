import {
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type RefCallback, useContext, useSyncExternalStore } from "react";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { EvidenceAttachmentEntry } from "../../features/evidence/EvidenceAttachmentEntry";
import { EvidenceFileRecovery } from "../../features/evidence/EvidenceFileRecovery";
import type { TimelineFileSnapshot } from "../../features/evidence/WorkbookTimelineFileOwner";

const noFileSubscription = () => () => {};
const emptyFiles: readonly TimelineFileSnapshot[] = [];
const noFiles = () => emptyFiles;

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
    row: WorkbookRow,
    files: FileList | File[],
  ) => void;
};

export function TimelineEvidencePanel({
  countDisplay,
  elementRef,
  row,
  onFilesSelected,
}: TimelineEvidencePanelProps) {
  const owner = useContext(TimelineFileContext);
  const files = useSyncExternalStore(
    owner?.subscribe ?? noFileSubscription,
    owner?.getSnapshot ?? noFiles,
  );
  const recordId = row.recordId;
  const disabledReason =
    owner?.attachmentDisabledReason() ??
    (owner ? null : "File attachment is unavailable.");
  const attach = (files: FileList | File[]) => {
    if (!owner || owner.attachmentDisabledReason() !== null) return;
    onFilesSelected(row, files);
  };
  if (recordId === null) {
    return null;
  }

  return (
    <section
      ref={elementRef}
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
        onAttach={(selected) => attach(Array.from(selected))}
      />
      {files.some((entry) => entry.recordId === recordId) ? (
        <section aria-label="Attachment progress and recovery">
          {owner
            ? files
                .filter((entry) => entry.recordId === recordId)
                .map((entry) => (
                  <EvidenceFileRecovery
                    {...entry}
                    key={entry.key}
                    presentation="inspector"
                    source={entry.sourceLabel}
                    onConfirmReview={() => owner.confirmReview(entry.key)}
                    onReview={() => void owner.review(entry.key)}
                    onResume={() => void owner.resume(entry.key)}
                    onFreshSlot={() => owner.freshSlot(entry.key)}
                    onNewId={() => owner.newRequestId(entry.key)}
                    onDiscard={() => owner.discard(entry.key)}
                    onRefresh={() => void owner.refresh(entry.key)}
                  />
                ))
            : null}
        </section>
      ) : null}
    </section>
  );
}
