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
  buildEvidenceMetadataPresentation,
  evidenceOperationFeedback,
} from "../../evidence/evidenceAccessPresentation";
import {
  EvidenceHandleButtons,
  evidenceButtonStyle,
  evidenceMessageStyle,
} from "../../features/evidence/EvidenceAccessActions";
import { TimelineFileContext } from "../../features/evidence/EvidenceAttachmentContext";
import { EvidenceAttachmentEntry } from "../../features/evidence/EvidenceAttachmentEntry";
import type { TimelineFileSnapshot } from "../../features/evidence/WorkbookTimelineFileOwner";
import type { useTimelineEvidenceMetadata } from "../hooks/useTimelineEvidenceMetadata";
import type { useTimelineLinkedEvidenceReview } from "../hooks/useTimelineLinkedEvidenceReview";
import {
  readTimelineEvidenceLinks,
  timelineEvidenceLinkKey,
} from "../models/timelineEvidenceLinks";
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

export type TimelineEvidenceReview = ReturnType<
  typeof useTimelineLinkedEvidenceReview
> & {
  readonly metadata: ReturnType<typeof useTimelineEvidenceMetadata>;
};

type TimelineEvidencePanelProps = {
  readonly countDisplay: TimelineEvidenceCountDisplay;
  readonly elementRef?: RefCallback<HTMLElement> | undefined;
  readonly row: WorkbookRow;
  readonly review?: TimelineEvidenceReview | undefined;
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
  const metadata =
    review?.metadata.sourceKey === timelineEvidenceLinkKey(row)
      ? review.metadata
      : null;
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
        <>
          <p style={bodyStyle}>Linked evidence: {links.items.length}</p>
          {metadata?.kind === "loading" ? (
            <p role="status" style={bodyStyle}>
              Loading evidence information…
            </p>
          ) : null}
          {metadata?.kind === "failed" ? (
            <p role="alert" style={bodyStyle}>
              Evidence information is unavailable.{" "}
              <button
                type="button"
                style={evidenceButtonStyle}
                onClick={metadata.retry}
              >
                Retry evidence information
              </button>
            </p>
          ) : null}
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
            {(metadata?.links ?? []).map((link) => {
              const state =
                review?.access.operations[link.evidenceRecordId]?.state;
              const feedback = state ? evidenceOperationFeedback(state) : null;
              const item =
                metadata?.kind === "ready"
                  ? metadata.items.find(
                      (item) => item.recordId === link.evidenceRecordId,
                    )
                  : undefined;
              const availability = item
                ? buildEvidenceMetadataPresentation(item, state ?? null)
                : null;
              const title = item?.title ?? link.title;
              return (
                <li
                  key={link.itemRef}
                  data-evidence-record-id={link.evidenceRecordId}
                  data-evidence-row-version={item?.rowVersion}
                  data-evidence-metadata-state={
                    item ? "ready" : (metadata?.kind ?? "unavailable")
                  }
                  style={{ minInlineSize: 0, overflowWrap: "anywhere" }}
                >
                  <p style={bodyStyle}>{title}</p>
                  {availability ? (
                    <p style={bodyStyle}>
                      {availability.lifecycleLabel} · {availability.uploadLabel}
                      {!availability.canPreview
                        ? ` · ${availability.message}`
                        : ""}
                    </p>
                  ) : null}
                  <EvidenceHandleButtons
                    canRead={!!review && !review.access.accessLost && !!item}
                    canPreview={availability?.canPreview ?? false}
                    canDownload={availability?.canDownload ?? false}
                    retryKind={availability?.retryKind ?? null}
                    busyKind={availability?.busyKind ?? null}
                    context="inspector"
                    labelSuffix={title}
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
          {metadata && metadata.total > metadata.links.length ? (
            <nav
              aria-label="Linked Evidence pages"
              style={{
                display: "flex",
                alignItems: "center",
                gap: "var(--ct-spacing-sm)",
              }}
            >
              <button
                type="button"
                style={evidenceButtonStyle}
                disabled={!metadata.previous}
                onClick={() => metadata.previous?.()}
              >
                Previous evidence
              </button>
              <span style={bodyStyle}>
                {metadata.start + 1}–{metadata.start + metadata.links.length} of{" "}
                {metadata.total}
              </span>
              <button
                type="button"
                style={evidenceButtonStyle}
                disabled={!metadata.next}
                onClick={() => metadata.next?.()}
              >
                Next evidence
              </button>
            </nav>
          ) : null}
        </>
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
