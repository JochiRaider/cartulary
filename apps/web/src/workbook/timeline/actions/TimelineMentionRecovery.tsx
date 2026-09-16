import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
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
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: WorkbookRecoveryItem[] = [];
  if (current.authority) {
    for (const entry of current.creations) {
      const link = current.entries.find(
        (operation) => operation.key === entry.linkKey,
      );
      const completed =
        entry.receipt &&
        entry.refresh === "complete" &&
        link?.receipt &&
        link.refresh === "complete";
      const progress =
        entry.phase === "preparing" ||
        entry.phase === "submitting" ||
        link?.phase === "preparing" ||
        link?.phase === "submitting";
      items.push({
        id: `create:${entry.key}`,
        refreshViews: [
          ...(entry.receipt && entry.refresh !== "complete"
            ? [
                entry.attempt.review.subject.entityType === "host"
                  ? hostsViewSchemaId
                  : identitiesViewSchemaId,
              ]
            : []),
          ...(link?.receipt && link.refresh !== "complete"
            ? [timelineViewSchemaId]
            : []),
        ],
        label: "Mention creation and resolution",
        origin: entry.attempt.review.subject.rawText,
        sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        order: entry.key,
        summary: completed
          ? "Created and resolved"
          : progress
            ? "Creating or resolving"
            : "Review creation, resolution or refresh",
        attention: completed
          ? "completed"
          : progress
            ? "progress"
            : "attention",
      });
    }
    for (const entry of current.entries) {
      if (current.creations.some((creation) => creation.linkKey === entry.key))
        continue;
      items.push({
        id: `resolve:${entry.key}`,
        refreshViews:
          entry.receipt && entry.refresh !== "complete"
            ? [timelineViewSchemaId]
            : [],
        label: "Mention resolution",
        origin: entry.attempt.review.subject.rawText,
        sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
        order: entry.key,
        summary: entry.receipt
          ? entry.refresh === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : "Resolution pending or needs review",
        attention:
          entry.receipt && entry.refresh === "complete"
            ? "completed"
            : entry.phase === "preparing" || entry.phase === "submitting"
              ? "progress"
              : "attention",
      });
    }
  }
  const selected = useWorkbookRecoverySource("mention", items);
  const snapshot = {
    ...current,
    creations: current.creations.filter(
      (entry) => `create:${entry.key}` === selected,
    ),
    entries: current.entries.filter(
      (entry) =>
        `resolve:${entry.key}` === selected ||
        current.creations.some(
          (creation) =>
            `create:${creation.key}` === selected &&
            creation.linkKey === entry.key,
        ),
    ),
  };
  return (
    <WorkbookRecoveryDetail source="mention" item={selected}>
      <section aria-label="Retained mention operations">
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
    </WorkbookRecoveryDetail>
  );
}
