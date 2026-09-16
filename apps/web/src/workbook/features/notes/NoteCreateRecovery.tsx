import { useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { NoteCreateForm } from "./NoteCreateForm";
import { noteRecoveryItems } from "./noteRecoveryItems";
import type { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";
export function NoteCreateRecovery({
  owner,
}: {
  readonly owner: WorkbookNoteCreateOwner;
}) {
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const attachment = useRef(Symbol("notes-recovery")).current;
  const selected = useWorkbookRecoverySource(
    "notes",
    noteRecoveryItems(current),
    { detach: () => owner.detach(attachment) },
  );
  const state = {
    ...current,
    draft:
      current.draft && String(current.draft.id) === selected
        ? current.draft
        : null,
    entries: current.entries.filter(
      (entry) => String(entry.attempt.review.draft.id) === selected,
    ),
  };
  return (
    <WorkbookRecoveryDetail source="notes" item={selected}>
      <section aria-label="Retained Note authoring">
        {state.draft && state.attachment === attachment ? (
          <NoteCreateForm
            owner={owner}
            attachment={attachment}
            onSubmit={() => void owner.submit(attachment)}
          />
        ) : state.draft ? (
          <>
            <p>Your Note draft is retained.</p>
            <Button
              tone="secondary"
              type="button"
              onClick={() => owner.resume(attachment)}
            >
              Resume Note draft
            </Button>
            <Button
              tone="secondary"
              type="button"
              disabled={owner.busy}
              onClick={() => owner.discard()}
            >
              Discard Note draft
            </Button>
          </>
        ) : null}
        {state.entries
          .filter(
            (entry) =>
              entry.phase !== "rejected" && entry.refresh !== "complete",
          )
          .map((entry) => (
            <section
              key={entry.attempt.clientTxnId}
              aria-label="Note submission recovery"
            >
              <p role="status">{entry.message ?? "Creating Note…"}</p>
              {entry.phase === "uncertain" ? (
                <Button
                  tone="primary"
                  type="button"
                  disabled={
                    !owner.canReplay() ||
                    state.preparing ||
                    entry.transportPending
                  }
                  onClick={() => void owner.replay(entry.attempt.clientTxnId)}
                >
                  Recover submission
                </Button>
              ) : null}
              {entry.receipt ? (
                <>
                  <p>Note created.</p>
                  <Button
                    tone="secondary"
                    type="button"
                    disabled={entry.refresh === "refreshing"}
                    onClick={() =>
                      void owner.retryRefresh(entry.attempt.clientTxnId)
                    }
                  >
                    Retry refresh
                  </Button>
                </>
              ) : null}
              <details>
                <summary>Submission details</summary>
                <p style={{ overflowWrap: "anywhere" }}>
                  Transaction: {entry.attempt.clientTxnId}
                </p>
                {entry.receipt ? (
                  <p style={{ overflowWrap: "anywhere" }}>
                    Note: {entry.receipt.data.row.record_id}
                  </p>
                ) : null}
              </details>
            </section>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
