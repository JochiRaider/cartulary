import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorNoticeView } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import {
  type NoteAssociationEntry,
  noteAssociationView,
} from "./noteAssociationOperation";
import type { WorkbookNoteAssociationOwner } from "./WorkbookNoteAssociationOwner";

export function NoteAssociationRecovery({
  owner,
}: {
  readonly owner: WorkbookNoteAssociationOwner;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items = state.entries
    .filter(
      (entry) => entry.phase !== "rejected" && entry.refresh !== "complete",
    )
    .map((entry) => ({
      id: entry.attempt.clientTxnId,
      label:
        entry.attempt.review.kind === "source"
          ? "Note source associations"
          : entry.attempt.review.kind === "evidence"
            ? "Note evidence associations"
            : "Related Note associations",
      origin: entry.attempt.review.label,
      summary: entry.receipt
        ? "Saved; refresh required"
        : entry.phase === "uncertain"
          ? "Outcome unconfirmed"
          : "Saving associations",
      sheetRef: entry.attempt.review.sheetRef,
      attention:
        entry.receipt || entry.phase === "uncertain"
          ? ("attention" as const)
          : ("progress" as const),
      order: entry.order,
      refreshViews: entry.receipt
        ? [noteAssociationView, "cartulary.view.evidence.v1"]
        : [],
    }));
  const selected = useWorkbookRecoverySource("note-associations", items);
  return (
    <WorkbookRecoveryDetail source="note-associations" item={selected}>
      {state.entries
        .filter((entry) => entry.attempt.clientTxnId === selected)
        .map((entry) => (
          <NoteAssociationResult
            key={entry.attempt.clientTxnId}
            owner={owner}
            entry={entry}
          />
        ))}
    </WorkbookRecoveryDetail>
  );
}

export function NoteAssociationResult({
  owner,
  entry,
}: {
  readonly owner: WorkbookNoteAssociationOwner;
  readonly entry: NoteAssociationEntry;
}) {
  const notice = owner.inspectorNotice(
    entry.attempt.review.row.record_id,
    entry.attempt.review.kind,
    entry.attempt.clientTxnId,
    owner.outcomeTransition(entry.attempt.clientTxnId),
    entry.failure
      ? {
          kind: "error",
          error: workbookInspectorErrorPresentation(entry.failure),
        }
      : {
          kind: "message",
          announcement: "none",
          message: entry.receipt
            ? entry.refresh === "complete"
              ? "Associations saved."
              : "Associations saved. Reads need refresh."
            : entry.phase === "uncertain"
              ? "The outcome is unconfirmed. Recover the original submission before changing these associations."
              : "Saving associations…",
        },
    entry.receipt ? "none" : entry.failure ? "assertive" : "polite",
  );
  return (
    <section aria-label="Note association result">
      <WorkbookInspectorNoticeView
        notice={notice}
        consume={owner.inspectorNotices.consume}
      />
      {entry.phase === "uncertain" ? (
        <Button
          tone="primary"
          disabled={
            !owner.canReplay() ||
            owner
              .getSnapshot()
              .preparing.includes(entry.attempt.review.row.record_id) ||
            entry.transportPending
          }
          onClick={() => void owner.replay(entry.attempt.clientTxnId)}
        >
          Recover submission
        </Button>
      ) : null}
      {entry.receipt && entry.refresh !== "complete" ? (
        <Button
          tone="secondary"
          disabled={entry.refresh === "refreshing"}
          onClick={() => void owner.retryRefresh(entry.attempt.clientTxnId)}
        >
          Retry refresh
        </Button>
      ) : null}
    </section>
  );
}
