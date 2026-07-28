import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, RefObject } from "react";
import type { WorkbookPendingQueueSnapshot } from "../../runtime/workbookPendingReplayRuntime";
import type { AutoResolutionNotice } from "../models/workbookMentionChips";

export function timelinePendingQueueMessage(
  pendingQueueSnapshot: WorkbookPendingQueueSnapshot,
): string | null {
  if (pendingQueueSnapshot.overflowMessage !== null) {
    return pendingQueueSnapshot.overflowMessage;
  }
  if (pendingQueueSnapshot.haltedMessage !== null) {
    return pendingQueueSnapshot.haltedMessage;
  }
  if (
    pendingQueueSnapshot.authPaused &&
    pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount > 0
  ) {
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
  inspectorOpen = false,
  onReviewAutoResolution,
  onUndoAutoResolution,
  pendingQueueSnapshot,
  recoveryPanelRef,
}: {
  readonly autoResolutionNotices: readonly AutoResolutionNotice[];
  readonly entityIndex: Record<string, { label: string }>;
  readonly inspectorOpen?: boolean | undefined;
  readonly onReviewAutoResolution: (
    rowRecordId: string,
    itemRef: string,
  ) => void;
  readonly onUndoAutoResolution: (notice: AutoResolutionNotice) => void;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly recoveryPanelRef: RefObject<HTMLDivElement | null>;
}) {
  const blockedEdit = pendingQueueSnapshot.blockedEdit;
  const noticeIsFocusable =
    blockedEdit !== null || pendingQueueSnapshot.overflowMessage !== null;
  const pendingQueueMessage = timelinePendingQueueMessage(pendingQueueSnapshot);
  const pendingQueueCount =
    pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount;

  if (autoResolutionNotices.length === 0 && pendingQueueMessage === null) {
    return null;
  }

  return (
    <aside
      aria-label="Workbook notices"
      style={{
        ...noticeStackStyle,
        ...(inspectorOpen ? noticeStackWithInspectorStyle : null),
      }}
    >
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
          ref={recoveryPanelRef}
          role={blockedEdit === null ? "status" : undefined}
          style={
            blockedEdit === null
              ? pendingQueueNoticeCardStyle
              : pendingQueueRecoveryCardStyle
          }
          tabIndex={noticeIsFocusable ? -1 : undefined}
        >
          <strong
            id={
              blockedEdit === null ? undefined : "pending-queue-recovery-title"
            }
            style={pendingQueueTitleStyle}
          >
            Queued edits
          </strong>
          <span aria-live="polite" style={noticeMessageStyle}>
            {pendingQueueMessage}
          </span>
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
  pointerEvents: "auto",
} satisfies CSSProperties;

const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap",
} satisfies CSSProperties;

const noticeStackStyle = {
  position: "absolute",
  zIndex: 6,
  insetBlockStart:
    "calc(var(--ct-layout-viewBarHeight) + var(--ct-spacing-sm))",
  insetInlineEnd: "var(--ct-spacing-sm)",
  display: "grid",
  gap: "0.5rem",
  inlineSize: "min(34rem, calc(100% - var(--ct-spacing-xl)))",
  minWidth: 0,
  maxBlockSize: "min(14rem, 32vh)",
  overflowY: "auto",
  pointerEvents: "none",
} satisfies CSSProperties;

const noticeStackWithInspectorStyle = {
  insetInlineEnd:
    "calc(var(--ct-layout-inspectorDefaultWidth) + var(--ct-spacing-sm))",
  inlineSize: "min(28rem, 50vw)",
} satisfies CSSProperties;

const noticeCardStyle = {
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem 1rem",
  display: "grid",
  gap: "0.5rem",
  minWidth: 0,
  alignSelf: "start",
  boxShadow: "var(--ct-elevation-popover)",
  pointerEvents: "none",
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
  boxShadow: "var(--ct-elevation-popover)",
  pointerEvents: "none",
} satisfies CSSProperties;

const pendingQueueRecoveryCardStyle = {
  ...pendingQueueNoticeCardStyle,
  alignItems: "start",
  display: "grid",
  overflow: "visible",
  pointerEvents: "auto",
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
