import type { ReactNode } from "react";
import {
  evidenceButtonStyle,
  evidenceMessageStyle,
} from "./EvidenceAccessActions";

/** Local recovery actions operate on retained identity, independent of selection. */
export function EvidenceFileRecovery({
  filename,
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
  readonly presentation?: "grid" | "inspector";
  readonly reviewText: string | null;
  readonly onConfirmReview: () => void;
  readonly filename: string;
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
