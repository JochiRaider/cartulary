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
import {
  evidenceAccessActionAvailability,
  evidenceOperationFeedback,
} from "../../evidence/evidenceAccessPresentation";
import {
  EvidenceHandleButtons,
  evidenceMessageStyle,
} from "../../features/evidence/EvidenceAccessActions";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { EvidenceAttachmentEntry } from "../../features/evidence/EvidenceAttachmentEntry";
import type { TimelineFileSnapshot } from "../../features/evidence/WorkbookTimelineFileOwner";
import type { useTimelineLinkedEvidenceReview } from "../hooks/useTimelineLinkedEvidenceReview";
import { readTimelineEvidenceLinks } from "../models/timelineEvidenceLinks";
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
  readonly review?:
    | ReturnType<typeof useTimelineLinkedEvidenceReview>
    | undefined;
  readonly registerList?:
    | ((recordId: string, element: HTMLElement | null) => void)
    | undefined;
  readonly onFilesSelected: (
    source: TimelineFileSource,
    files: FileList | readonly File[],
  ) => void;
};

export function TimelineEvidencePanel({
  countDisplay,
  elementRef,
  row,
  review,
  registerList,
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
  const links = readTimelineEvidenceLinks(row);
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
      aria-label="Timeline evidence"
    >
      <section aria-label="Evidence information">
        <p style={bodyStyle}>
          Attached evidence count:{" "}
          {countDisplay.stateKey === "inconsistent"
            ? "Unavailable"
            : countDisplay.displayCount}
        </p>
      </section>
      {links.kind === "available" ? (
        <ul
          aria-label="Linked Evidence"
          tabIndex={-1}
          ref={(element) => registerList?.(recordId, element)}
          style={{
            listStyle: "none",
            margin: 0,
            padding: 0,
            display: "grid",
            gap: "var(--ct-spacing-sm)",
          }}
        >
          {links.items.map((link) => {
            const state =
              review?.access.operations[link.evidenceRecordId]?.state;
            const feedback = state ? evidenceOperationFeedback(state) : null;
            const availability = evidenceAccessActionAvailability(
              state ?? null,
              true,
            );
            return (
              <li
                key={link.itemRef}
                style={{ minInlineSize: 0, overflowWrap: "anywhere" }}
              >
                <p style={bodyStyle}>{link.title}</p>
                <EvidenceHandleButtons
                  canRead={!!review && !review.access.accessLost}
                  canPreview={availability.canPreview}
                  canDownload={availability.canDownload}
                  retryKind={availability.retryKind}
                  busyKind={availability.busyKind}
                  context="inspector"
                  labelSuffix={link.title}
                  recordId={link.evidenceRecordId}
                  onIssue={(kind, invoker) => {
                    if (review) void review.issue(row, link, kind, invoker);
                  }}
                />
                {feedback ? (
                  <p
                    role={
                      state?.kind === "rejected" || state?.kind === "deadline"
                        ? "alert"
                        : "status"
                    }
                    style={evidenceMessageStyle}
                  >
                    {feedback.message}
                  </p>
                ) : null}
              </li>
            );
          })}
        </ul>
      ) : (
        <p
          tabIndex={-1}
          ref={(element) => registerList?.(recordId, element)}
          style={bodyStyle}
        >
          {links.kind === "empty"
            ? "No linked Evidence records."
            : "Linked Evidence is unavailable."}
        </p>
      )}
      {review?.spaceStatus === "checking" ? (
        <p role="status" style={evidenceMessageStyle}>
          Checking linked previews…
        </p>
      ) : null}
      {review?.spaceStatus === "indeterminate" ? (
        <p role="alert" style={evidenceMessageStyle}>
          Preview availability could not be fully checked. Choose an item to
          retry.
        </p>
      ) : null}
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
