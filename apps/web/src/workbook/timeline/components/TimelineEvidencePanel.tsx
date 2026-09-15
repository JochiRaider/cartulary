import {
  timelineEvidenceAttachSectionTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
} from "@cartulary/ui-contracts";
import { type RefCallback, useContext, useSyncExternalStore } from "react";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { EvidenceFileRecovery } from "../../features/evidence/EvidenceFileRecovery";
import type { TimelineFileSnapshot } from "../../features/evidence/WorkbookTimelineFileOwner";

const noFileSubscription = () => () => {};
const emptyFiles: readonly TimelineFileSnapshot[] = [];
const noFiles = () => emptyFiles;

import type { WorkbookRow } from "../models/timelineRowModel";
import {
  bodyStyle,
  inputStyle,
  inspectorSectionStyle,
  labelStyle,
} from "./TimelineWorkbookStyles";

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
      onDragOver={(event) => {
        event.preventDefault();
      }}
      onDrop={(event) => {
        event.preventDefault();
        event.stopPropagation();
        onFilesSelected(row, event.dataTransfer.files);
      }}
      onPaste={(event) => {
        if (event.clipboardData.files.length > 0) {
          event.preventDefault();
          event.stopPropagation();
          onFilesSelected(row, event.clipboardData.files);
        }
      }}
    >
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
      <div data-testid={timelineEvidenceAttachSectionTestId(recordId)}>
        <p style={bodyStyle}>
          Attached evidence count: {countDisplay.displayCount}
        </p>
        <label style={labelStyle}>
          Attach file
          <input
            data-testid={timelineEvidenceFileInputTestId(recordId)}
            style={inputStyle}
            type="file"
            accept="image/*,.txt,.pdf,text/plain,application/pdf"
            onChange={(event) => {
              onFilesSelected(row, event.currentTarget.files ?? []);
              event.currentTarget.value = "";
            }}
          />
        </label>
      </div>
    </section>
  );
}
