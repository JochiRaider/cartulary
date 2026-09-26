import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  workbookGridRowHeightPx,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useLayoutEffect,
  useSyncExternalStore,
} from "react";
import type {
  AutoResolutionDisclosure,
  WorkbookTimelineMentionOperationOwner,
} from "../actions/WorkbookTimelineMentionOperationOwner";
import {
  type DisclosureReviewFeedback,
  disclosureReviewKey,
} from "../models/workbookMentionChips";
import { actionButtonStyle } from "./TimelineWorkbookStyles";

/** This leaf observes retained disclosure, independently of grid composition. */
export function TimelineWorkbookNotices({
  owner,
  density,
  entityIndex,
  reviewFeedback,
  onReviewAutoResolution,
  onUndoAutoResolution,
}: {
  readonly owner: WorkbookTimelineMentionOperationOwner;
  readonly entityIndex: Readonly<Record<string, { label: string }>>;
  readonly density: Parameters<typeof workbookGridRowHeightPx>[0];
  readonly reviewFeedback: DisclosureReviewFeedback | null;
  readonly onReviewAutoResolution: (notice: AutoResolutionDisclosure) => void;
  readonly onUndoAutoResolution: (notice: AutoResolutionDisclosure) => void;
}) {
  const notices = useSyncExternalStore(
    owner.subscribe,
    owner.getDisclosureSnapshot,
  );
  const actions = useSyncExternalStore(
    owner.subscribe,
    owner.getActionSnapshot,
  );
  useLayoutEffect(() => {
    if (notices.length > 0) owner.updateDisclosureLabels(entityIndex);
  }, [owner, entityIndex, notices]);
  if (notices.length === 0) return null;
  const batches = new Set<string>();
  return (
    <aside
      aria-label="Auto-resolution disclosures"
      style={{
        ...regionStyle,
        maxBlockSize: `min(calc(${3 * workbookGridRowHeightPx(density)}px + 2 * var(--ct-spacing-sm)), 25cqh)`,
      }}
    >
      <ul style={listStyle}>
        {notices.map((notice) => {
          const reviewKey = disclosureReviewKey(notice);
          const localReview =
            reviewFeedback?.key === reviewKey ? reviewFeedback : null;
          const entry = [...actions.entries]
            .reverse()
            .find(
              (entry) =>
                entry.attempt.review.subject.mentionId ===
                notice.entityMentionId,
            );
          const batchKey = JSON.stringify(notice.operation);
          const showBatch =
            notice.operation.kind === "batch" && !batches.has(batchKey);
          batches.add(batchKey);
          return (
            <li
              key={notice.identity}
              data-testid={autoResolutionNoticeTestId(notice.itemRef)}
              style={itemStyle}
            >
              <div style={detailsStyle}>
                {showBatch ? (
                  <strong>
                    {notice.acceptedCount} mentions auto-resolved in this change
                    set.{" "}
                  </strong>
                ) : null}
                <span>
                  Auto-resolved: raw token <strong>{notice.rawText}</strong>{" "}
                  matched{" "}
                  <strong>
                    {entityIndex[notice.resolvedRecordId]?.label ||
                      (notice.displayText !== notice.rawText
                        ? notice.displayText
                        : null) ||
                      notice.resolvedRecordId}
                  </strong>
                  {notice.matchedAliasText ? (
                    <>
                      {" "}
                      via alias <strong>{notice.matchedAliasText}</strong>
                    </>
                  ) : (
                    " (no alias supplied)"
                  )}
                  .
                </span>
                {entry?.failure ? (
                  <span role="status"> {entry.failure.message}</span>
                ) : null}
                {entry && ["preparing", "submitting"].includes(entry.phase) ? (
                  <span role="status"> Undo pending.</span>
                ) : null}
                {localReview?.phase === "pending" ? (
                  <span role="status"> Opening source…</span>
                ) : null}
                {localReview?.phase === "failure" ? (
                  <span role="alert"> {localReview.message}</span>
                ) : null}
              </div>
              <div style={actionsStyle}>
                {owner.canSubmit("revert_to_unresolved") ? (
                  <button
                    data-testid={autoResolutionUndoButtonTestId(notice.itemRef)}
                    disabled={!owner.canUndoDisclosure(notice)}
                    style={buttonStyle}
                    type="button"
                    onClick={() => onUndoAutoResolution(notice)}
                  >
                    Undo
                  </button>
                ) : (
                  <span>Read only</span>
                )}
                <button
                  data-testid={autoResolutionReviewButtonTestId(notice.itemRef)}
                  data-disclosure-review-key={reviewKey}
                  aria-busy={localReview?.phase === "pending"}
                  style={buttonStyle}
                  type="button"
                  onClick={() => onReviewAutoResolution(notice)}
                >
                  Review
                </button>
                {entry?.phase === "uncertain" ? (
                  <button
                    type="button"
                    disabled={!owner.canSubmit("revert_to_unresolved")}
                    style={buttonStyle}
                    onClick={() => void owner.replay(entry.key)}
                  >
                    Retry Undo
                  </button>
                ) : null}
              </div>
            </li>
          );
        })}
      </ul>
    </aside>
  );
}
const regionStyle = {
  minHeight: 0,
  minWidth: 0,
  overflowY: "auto",
  overflowAnchor: "none",
  boxSizing: "border-box",
  background: "var(--ct-colors-surface-2)",
  borderBlockEnd: "var(--ct-border-hairline)",
  padding: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const listStyle = {
  listStyle: "none",
  margin: 0,
  padding: 0,
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const itemStyle = {
  display: "flex",
  flexWrap: "wrap",
  alignItems: "start",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
} satisfies CSSProperties;
const detailsStyle = {
  flex: "1 1 0",
  minWidth: 0,
  overflowWrap: "anywhere",
  lineHeight: "inherit",
} satisfies CSSProperties;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const buttonStyle = {
  ...actionButtonStyle,
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
} satisfies CSSProperties;
