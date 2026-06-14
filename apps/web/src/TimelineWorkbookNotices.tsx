import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { TimelinePendingQueueSnapshot } from "./useTimelinePendingSaves";
import type { AutoResolutionNotice } from "./workbookMentionChips";

export function TimelineWorkbookNotices({
  autoResolutionNotices,
  entityIndex,
  onReviewAutoResolution,
  onUndoAutoResolution,
  pendingQueueSnapshot,
}: {
  readonly autoResolutionNotices: readonly AutoResolutionNotice[];
  readonly entityIndex: Record<string, { label: string }>;
  readonly onReviewAutoResolution: (
    rowRecordId: string,
    itemRef: string,
  ) => void;
  readonly onUndoAutoResolution: (notice: AutoResolutionNotice) => void;
  readonly pendingQueueSnapshot: TimelinePendingQueueSnapshot;
}) {
  const hasPendingQueueNotice =
    pendingQueueSnapshot.overflowMessage !== null ||
    pendingQueueSnapshot.haltedMessage !== null ||
    pendingQueueSnapshot.authPaused ||
    pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount > 0;

  return (
    <>
      {autoResolutionNotices.length > 0 ? (
        <aside style={noticeStackStyle}>
          {autoResolutionNotices.map((notice) => (
            <div
              key={notice.itemRef}
              data-testid={autoResolutionNoticeTestId(notice.itemRef)}
              style={noticeCardStyle}
            >
              <p style={noticeTitleStyle}>Auto-resolved mention</p>
              <p style={bodyStyle}>
                Raw token <strong>{notice.rawText}</strong> matched{" "}
                <strong>
                  {entityIndex[notice.resolvedRecordId]?.label ??
                    notice.rawText}
                </strong>
                {notice.matchedAliasText ? (
                  <>
                    {" "}
                    via alias <strong>{notice.matchedAliasText}</strong>
                  </>
                ) : null}
                .
              </p>
              <div style={inlineButtonRowStyle}>
                <button
                  data-testid={autoResolutionUndoButtonTestId(notice.itemRef)}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    onUndoAutoResolution(notice);
                  }}
                >
                  Undo
                </button>
                <button
                  data-testid={autoResolutionReviewButtonTestId(notice.itemRef)}
                  style={secondaryActionButtonStyle}
                  type="button"
                  onClick={() => {
                    onReviewAutoResolution(notice.rowRecordId, notice.itemRef);
                  }}
                >
                  Review
                </button>
              </div>
            </div>
          ))}
        </aside>
      ) : null}

      {hasPendingQueueNotice ? (
        <aside
          data-testid={pendingQueueNoticeTestId()}
          role="status"
          style={noticeCardStyle}
        >
          <p style={noticeTitleStyle}>Queued edits</p>
          <p style={bodyStyle}>
            {pendingQueueSnapshot.overflowMessage ??
              pendingQueueSnapshot.haltedMessage ??
              (pendingQueueSnapshot.authPaused
                ? "Authentication is required before queued edits can replay."
                : "Queued edits are waiting to replay.")}
          </p>
          <p data-testid={pendingQueueCountTestId()} style={bodyStyle}>
            Pending units:{" "}
            {pendingQueueSnapshot.queuedCount +
              pendingQueueSnapshot.inFlightCount}
          </p>
        </aside>
      ) : null}
    </>
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
} satisfies CSSProperties;

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
} satisfies CSSProperties;

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap",
} satisfies CSSProperties;

const noticeStackStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
} satisfies CSSProperties;

const noticeCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
} satisfies CSSProperties;

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
} satisfies CSSProperties;
