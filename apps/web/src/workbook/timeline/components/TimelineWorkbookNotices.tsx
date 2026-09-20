import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import type { AutoResolutionNotice } from "../models/workbookMentionChips";
import { actionButtonStyle } from "./TimelineWorkbookStyles";

export function TimelineWorkbookNotices({
  autoResolutionNotices,
  canManageMentions,
  entityIndex,
  inspectorOpen = false,
  onReviewAutoResolution,
  onUndoAutoResolution,
}: {
  readonly canManageMentions: boolean;
  readonly autoResolutionNotices: readonly AutoResolutionNotice[];
  readonly entityIndex: Record<string, { label: string }>;
  readonly inspectorOpen?: boolean | undefined;
  readonly onReviewAutoResolution: (
    rowRecordId: string,
    itemRef: string,
  ) => void;
  readonly onUndoAutoResolution: (notice: AutoResolutionNotice) => void;
}) {
  if (autoResolutionNotices.length === 0) return null;

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
            {canManageMentions ? (
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
            ) : null}
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
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
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

const noticeTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
  fontWeight: 600,
} satisfies CSSProperties;
