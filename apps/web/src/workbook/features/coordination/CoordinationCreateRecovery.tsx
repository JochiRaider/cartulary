import { useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { CoordinationCreateForm } from "./CoordinationCreateForm";
import { coordinationRecoveryItems } from "./coordinationRecoveryItems";
import type { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";
export function CoordinationCreateRecovery({
  owner,
}: {
  readonly owner: WorkbookCoordinationCreateOwner;
}) {
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const attachment = useRef(Symbol("coordination-recovery")).current;
  const selected = useWorkbookRecoverySource(
    "coordination",
    coordinationRecoveryItems(current),
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
  const actions = owner.captureDraftActions(state.attachment);
  return (
    <WorkbookRecoveryDetail source="coordination" item={selected}>
      <section aria-label="Retained Coordination authoring">
        {state.draft && state.attachment === attachment ? (
          <CoordinationCreateForm
            owner={owner}
            attachment={attachment}
            onSubmit={() => void owner.submit(attachment)}
          />
        ) : state.draft ? (
          <>
            <p>Your {state.draft.target.title} draft is retained.</p>
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                gap: "var(--ct-spacing-sm)",
              }}
            >
              <Button
                tone="secondary"
                type="button"
                onClick={() => actions.resume(attachment)}
              >
                Resume Coordination draft
              </Button>
              <Button
                tone="secondary"
                type="button"
                disabled={owner.busy}
                onClick={() => actions.discard()}
              >
                Discard Coordination draft
              </Button>
            </div>
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
              aria-label="Coordination submission recovery"
            >
              <p role="status">{entry.message ?? "Creating Coordination…"}</p>
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
                  <p>Coordination created.</p>
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
                    Coordination: {entry.receipt.data.row.record_id}
                  </p>
                ) : null}
              </details>
            </section>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
