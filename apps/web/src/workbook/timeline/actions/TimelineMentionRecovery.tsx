import { useRef, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  MentionCreationRefreshFeedback,
  MentionOperationFeedback,
} from "../components/TimelineMentionActionControls";
import type { WorkbookTimelineMentionOperationOwner } from "./WorkbookTimelineMentionOperationOwner";
export function TimelineMentionRecovery({
  owner,
}: {
  readonly owner: WorkbookTimelineMentionOperationOwner;
}) {
  const disclosure = useRef<HTMLDetailsElement>(null);
  const trigger = useRef<HTMLElement>(null);
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const total = snapshot.entries.length + snapshot.creations.length;
  if (!snapshot.authority || !total) return null;
  return (
    <details ref={disclosure} style={{ position: "relative", minWidth: 0 }}>
      <summary ref={trigger}>Mention operations ({total})</summary>
      <section
        aria-label="Retained mention operations"
        tabIndex={-1}
        onKeyDown={(event) => {
          if (event.key === "Escape" && disclosure.current) {
            event.preventDefault();
            event.stopPropagation();
            disclosure.current.open = false;
            trigger.current?.focus({ preventScroll: true });
          }
        }}
        style={{
          position: "fixed",
          zIndex: 20,
          insetInlineEnd: "1rem",
          boxSizing: "border-box",
          background: "var(--ct-colors-surface-1)",
          border: "var(--ct-border-hairline)",
          padding: "var(--ct-spacing-md)",
          width: "min(28rem, calc(100vw - 2rem))",
          maxHeight: "65vh",
          overflow: "auto",
          overflowWrap: "anywhere",
        }}
      >
        {snapshot.creations.map((entry) => (
          <section key={`create:${entry.key}`}>
            <p>
              Create {entry.attempt.review.subject.entityType} from “
              {entry.attempt.review.subject.rawText}”
            </p>
            <p role="status">
              {entry.receipt
                ? `Entity saved: ${String(entry.receipt.data.row.cells[`${entry.attempt.review.subject.entityType}.display_name`]?.value ?? entry.receipt.data.row.record_id)}. ${entry.linkKey === null ? "Select this mention in Timeline to review and finish resolving it." : "Mention resolution has a separate result below."}`
                : entry.phase === "uncertain"
                  ? "Creation outcome is uncertain. The entity may have been saved."
                  : (entry.failure?.message ?? "Creating entity…")}
            </p>
            <MentionCreationRefreshFeedback
              creation={entry}
              refresh={() => void owner.refreshCreation(entry.key)}
            />
            {entry.phase === "uncertain" ? (
              <WorkbookInspectorActionButton
                tone="secondary"
                disabled={
                  !owner.canCreate(entry.attempt.review.subject.entityType)
                }
                onClick={() => void owner.replayCreation(entry.key)}
              >
                Replay entity creation
              </WorkbookInspectorActionButton>
            ) : null}
          </section>
        ))}
        {snapshot.entries.map((entry) => (
          <section key={entry.key}>
            <p>“{entry.attempt.review.subject.rawText}”</p>
            <MentionOperationFeedback
              operation={entry}
              replay={() => void owner.replay(entry.key)}
              refresh={() => void owner.refresh(entry.key)}
              canReplay={owner.canSubmit(entry.attempt.review.intent.action)}
            />
          </section>
        ))}
      </section>
    </details>
  );
}
