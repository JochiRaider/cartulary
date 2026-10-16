import type { ReactNode } from "react";
import {
  evidenceButtonStyle,
  evidenceMessageStyle,
} from "./EvidenceAccessActions";
import type { EvidenceUploadSession } from "./EvidenceUploadSession";
import type { EvidenceWorkAttention } from "./evidenceWorkAttention";

/** Local recovery actions operate on retained identity, independent of selection. */
export function EvidenceFileRecovery({
  filename,
  verification,
  onCancelVerification,
  attention,
  presentation = "grid",
  source,
  message,
  busy,
  needsReview,
  canResume,
  canFreshSlot,
  canNewId,
  canDiscard,
  refreshRequired,
  onReview,
  onResume,
  onFreshSlot,
  onNewId,
  onDiscard,
  onRefresh,
  reviewText,
  onConfirmReview,
}: {
  readonly attention?: EvidenceWorkAttention | null;
  readonly presentation?: "grid" | "inspector";
  readonly reviewText: string | null;
  readonly onConfirmReview: () => void;
  readonly filename: string;
  readonly verification: EvidenceUploadSession["status"]["verification"];
  readonly onCancelVerification: () => void;
  readonly source: ReactNode;
  readonly message: string;
  readonly busy: boolean;
  readonly needsReview: boolean;
  readonly canResume: boolean;
  readonly canFreshSlot: boolean;
  readonly canNewId: boolean;
  readonly canDiscard: boolean;
  readonly refreshRequired: boolean;
  readonly onReview: () => void;
  readonly onResume: () => void;
  readonly onFreshSlot: () => void;
  readonly onNewId: () => void;
  readonly onDiscard: () => void;
  readonly onRefresh: () => void;
}) {
  return (
    <fieldset
      tabIndex={-1}
      data-evidence-work-id={attention?.workId}
      aria-label={`${presentation === "inspector" ? "Inspector file recovery" : "File recovery"}: ${filename}`}
      style={{
        border: 0,
        margin: 0,
        minInlineSize: 0,
        overflowWrap: "anywhere",
        display: "grid",
        gap: "var(--ct-spacing-xs)",
        padding: "var(--ct-spacing-xs)",
      }}
    >
      <legend>{filename}</legend>
      <span>Attachment to {source}</span>
      {/* Retained state can appear in two places; runtime owns live save acknowledgement. */}
      <span role="status" aria-live="off" style={evidenceMessageStyle}>
        {message}
      </span>
      <EvidenceVerificationProgress
        filename={filename}
        verification={verification}
        onCancel={onCancelVerification}
      />
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "var(--ct-spacing-xs)",
          alignItems: "center",
        }}
      >
        {needsReview && canResume ? (
          <button type="button" style={evidenceButtonStyle} onClick={onReview}>
            Review original source
          </button>
        ) : null}
        {reviewText && canResume ? (
          <span>
            {reviewText}{" "}
            <button
              type="button"
              style={evidenceButtonStyle}
              onClick={onConfirmReview}
            >
              Use reviewed source
            </button>
          </span>
        ) : null}
        {canResume ? (
          <button type="button" style={evidenceButtonStyle} onClick={onResume}>
            Resume
          </button>
        ) : null}
        {canFreshSlot ? (
          <button
            type="button"
            style={evidenceButtonStyle}
            onClick={onFreshSlot}
          >
            Start fresh upload
          </button>
        ) : null}
        {canNewId ? (
          <button type="button" style={evidenceButtonStyle} onClick={onNewId}>
            Use new request
          </button>
        ) : null}
        {refreshRequired ? (
          <button
            type="button"
            disabled={busy}
            style={evidenceButtonStyle}
            onClick={onRefresh}
          >
            Refresh
          </button>
        ) : null}
        {canDiscard ? (
          <button type="button" style={evidenceButtonStyle} onClick={onDiscard}>
            Discard retained file work
          </button>
        ) : null}
      </div>
    </fieldset>
  );
}

/** One source-owned Job view, shared by expanded recovery and compact progress. */
export function EvidenceVerificationProgress({
  filename,
  verification,
  onCancel,
}: {
  readonly filename: string;
  readonly verification: EvidenceUploadSession["status"]["verification"];
  readonly onCancel: () => void;
}) {
  if (verification.phase === "idle" || verification.phase === "verified")
    return null;
  return (
    <fieldset
      aria-label={`File verification: ${filename}`}
      data-job-id={verification.job?.job_id}
      data-job-blob-id={verification.blobId ?? undefined}
      data-job-status={verification.job?.status}
      data-job-purpose="evidence_verification"
      data-job-cancelable={verification.canCancel ? "true" : "false"}
      style={{
        border: 0,
        margin: 0,
        minInlineSize: 0,
        overflowWrap: "anywhere",
        display: "flex",
        flexWrap: "wrap",
        alignItems: "center",
        gap: "var(--ct-spacing-xs)",
        padding: "var(--ct-spacing-xs)",
      }}
    >
      <span role="status" data-job-progress="true">
        {filename}:{" "}
        {verification.phase === "admitting"
          ? "Starting file verification."
          : verification.phase === "uncertain"
            ? "Verification admission is uncertain. Resume the same request."
            : verification.phase === "interrupted"
              ? "Verification observation stopped. Resume to check its status."
              : verification.phase === "rejected"
                ? "Verification could not start. Review the file before retrying."
                : verification.phase === "failed"
                  ? "File verification failed. Review the file before retrying."
                  : verification.phase === "canceled"
                    ? "File verification canceled."
                    : verification.job?.status === "cancel_requested"
                      ? "Canceling file verification."
                      : "Verifying uploaded file."}
      </span>
      {verification.job &&
      !["succeeded", "failed", "canceled"].includes(verification.job.status) ? (
        <progress
          aria-label="File verification progress"
          max={verification.job.progress.total ?? undefined}
          value={
            verification.job.progress.total === null
              ? undefined
              : verification.job.progress.completed
          }
        />
      ) : null}
      {verification.canCancel || verification.cancelUncertain ? (
        <button
          type="button"
          data-job-cancel="true"
          style={evidenceButtonStyle}
          onClick={onCancel}
        >
          {verification.cancelUncertain
            ? "Recover cancellation"
            : "Cancel verification"}
        </button>
      ) : null}
    </fieldset>
  );
}
