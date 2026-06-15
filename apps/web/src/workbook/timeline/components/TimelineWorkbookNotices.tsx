import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { TimelinePendingQueueSnapshot } from "../hooks/useTimelinePendingSaves";
import type { AutoResolutionNotice } from "../models/workbookMentionChips";

export function timelinePendingQueueMessage(
  pendingQueueSnapshot: TimelinePendingQueueSnapshot,
): string | null {
  if (pendingQueueSnapshot.overflowMessage !== null) {
    return pendingQueueSnapshot.overflowMessage;
  }
  if (pendingQueueSnapshot.haltedMessage !== null) {
    return pendingQueueSnapshot.haltedMessage;
  }
  if (pendingQueueSnapshot.authPaused) {
    return "Authentication is required before queued edits can replay.";
  }
  if (
    pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount >
    0
  ) {
    return "Queued edits are waiting to replay.";
  }
  return null;
}

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
  const pendingQueueMessage = timelinePendingQueueMessage(pendingQueueSnapshot);
  const pendingQueueCount =
    pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount;

  if (autoResolutionNotices.length === 0 && pendingQueueMessage === null) {
    return null;
  }

  return (
    <aside aria-label="Workbook notices" style={noticeStackStyle}>
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
              {entityIndex[notice.resolvedRecordId]?.label ?? notice.rawText}
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

      {pendingQueueMessage !== null ? (
        <div
          data-testid={pendingQueueNoticeTestId()}
          role="status"
          style={pendingQueueNoticeCardStyle}
        >
          <strong style={pendingQueueTitleStyle}>Queued edits</strong>
          <span style={noticeMessageStyle}>{pendingQueueMessage}</span>
          <span data-testid={pendingQueueCountTestId()} style={queueCountStyle}>
            Pending {pendingQueueCount}
          </span>
        </div>
      ) : null}
    </aside>
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;

const noticeMessageStyle = {
  ...bodyStyle,
  flex: "1 1 auto",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
} satisfies CSSProperties;

const queueCountStyle = {
  ...bodyStyle,
  flex: "0 0 auto",
  whiteSpace: "nowrap" as const,
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
  gap: "0.5rem",
  marginBottom: "0.5rem",
  minWidth: 0,
  maxBlockSize: "min(8rem, 20vh)",
  overflowY: "auto",
} satisfies CSSProperties;

const noticeCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
  minWidth: 0,
  alignSelf: "start",
} satisfies CSSProperties;

const pendingQueueNoticeCardStyle = {
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.45rem 0.75rem",
  display: "flex",
  alignItems: "center",
  gap: "0.75rem",
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const pendingQueueTitleStyle = {
  color: "var(--ct-colors-ink)",
  fontSize: "0.85rem",
  fontWeight: 650,
  whiteSpace: "nowrap",
} satisfies CSSProperties;

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
} satisfies CSSProperties;
