import { useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { TimelineRelatedEvidenceForm } from "./TimelineRelatedEvidenceForm";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

export function TimelineRelatedEvidenceRecovery({
  owner,
}: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot),
    token = useRef(Symbol("related-evidence-recovery")).current;
  const details = useRef<HTMLDetailsElement>(null),
    trigger = useRef<HTMLElement>(null);
  const [open, setOpen] = useState(false);
  if (!state.authority || (!state.draft && !state.checkpoints.length))
    return null;
  return (
    <details
      ref={details}
      onToggle={(event) => {
        setOpen(event.currentTarget.open);
        if (!event.currentTarget.open) owner.detach(token);
      }}
      style={{ minWidth: 0 }}
    >
      <summary ref={trigger}>
        Timeline Evidence creation
        {owner.blockedCount
          ? ` (${owner.blockedCount} need recovery)`
          : state.draft
            ? " (draft retained)"
            : state.checkpoints.some((checkpoint) =>
                  owner.linkComplete(checkpoint),
                )
              ? " (created and linked)"
              : ""}
      </summary>
      {open ? (
        <section
          aria-label="Retained Timeline Evidence creation"
          onKeyDown={(event) => {
            if (event.key === "Escape" && details.current) {
              event.preventDefault();
              event.stopPropagation();
              details.current.open = false;
              owner.detach(token);
              trigger.current?.focus({ preventScroll: true });
            }
          }}
          style={{
            position: "fixed",
            zIndex: 20,
            insetInlineEnd: "var(--ct-spacing-md)",
            width: "min(28rem, calc(100vw - 2rem))",
            maxHeight: "65vh",
            boxSizing: "border-box",
            overflow: "auto",
            overflowWrap: "anywhere",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            padding: "var(--ct-spacing-md)",
          }}
        >
          {state.draft && state.attachment !== token ? (
            <div>
              <p>
                Evidence metadata is retained for its original Timeline record.
              </p>
              <Button
                type="button"
                tone="secondary"
                onClick={() => owner.resume(token)}
              >
                Resume Evidence draft
              </Button>
              <Button
                type="button"
                tone="secondary"
                disabled={owner.busy}
                onClick={() => owner.discard()}
              >
                Discard Evidence draft
              </Button>
            </div>
          ) : null}
          <TimelineRelatedEvidenceForm
            owner={owner}
            attachment={token}
            onSubmit={() => void owner.submit(token)}
            onReview={() => void owner.review()}
          />
          {state.checkpoints.map((checkpoint) => (
            <section
              key={checkpoint.id}
              aria-label="Evidence creation result"
              style={{
                borderBlockStart: "var(--ct-border-hairline)",
                paddingBlock: "var(--ct-spacing-sm)",
              }}
            >
              <p>
                Timeline source:{" "}
                {checkpoint.create.attempt.review.draft.presentation.label}
              </p>
              <details>
                <summary>Original Timeline record</summary>
                <p>{checkpoint.create.attempt.review.draft.source.recordId}</p>
              </details>
              <p role="status">
                {checkpoint.create.receipt
                  ? owner.linkComplete(checkpoint)
                    ? "Evidence created and linked to the original Timeline record."
                    : "Evidence created; Timeline link incomplete."
                  : checkpoint.create.phase === "rejected"
                    ? state.draft?.id ===
                      checkpoint.create.attempt.review.draft.id
                      ? "Evidence creation rejected. Your draft is retained."
                      : "Evidence creation rejected."
                    : checkpoint.create.phase === "uncertain"
                      ? "Evidence creation outcome unconfirmed."
                      : "Creating Evidence…"}
              </p>
              {checkpoint.message &&
              checkpoint.message !==
                "Evidence created; Timeline link incomplete." &&
              checkpoint.message !==
                "Evidence created and linked to the original Timeline record." ? (
                <p>{checkpoint.message}</p>
              ) : null}
              {checkpoint.sourceUnavailable ? (
                <p>
                  The original Timeline record is unavailable or superseded. The
                  created Evidence is retained.
                </p>
              ) : null}
              {checkpoint.evidenceUnavailable ? (
                <p>
                  The created Evidence is currently unavailable. Its accepted
                  creation is retained.
                </p>
              ) : null}
              {[checkpoint.create, ...checkpoint.links].map((entry) => (
                <div key={entry.attempt.clientTxnId}>
                  <p>
                    {entry.attempt.stage === "create"
                      ? "Creation"
                      : "Timeline link"}
                    :{" "}
                    {entry.phase === "accepted"
                      ? `saved; ${entry.refresh === "complete" ? "views refreshed" : "views need refresh"}`
                      : entry.phase === "uncertain"
                        ? "outcome unconfirmed"
                        : entry.phase === "rejected"
                          ? "rejected"
                          : "submitting"}
                    .
                  </p>
                  {entry.failure ? <p>{entry.failure.message}</p> : null}
                  {entry.receipt ? (
                    <details>
                      <summary>
                        {entry.attempt.stage === "create"
                          ? "Evidence creation receipt"
                          : "Timeline link receipt"}
                      </summary>
                      <p>
                        Record: {entry.receipt.data.row.record_id}. Version:{" "}
                        {entry.receipt.data.row.row_version}. Change set:{" "}
                        {entry.receipt.data.change_set_id}. Request:{" "}
                        {entry.receipt.meta.request_id}.
                      </p>
                    </details>
                  ) : null}
                  {entry.phase === "uncertain" ? (
                    <Button
                      type="button"
                      tone="primary"
                      disabled={
                        state.preparing ||
                        entry.transportPending ||
                        !owner.canReplay()
                      }
                      onClick={() =>
                        void owner.replay(
                          checkpoint.id,
                          entry.attempt.clientTxnId,
                        )
                      }
                    >
                      {entry.attempt.stage === "create"
                        ? "Recover Evidence creation"
                        : "Recover Timeline link"}
                    </Button>
                  ) : null}
                  {entry.phase === "rejected" &&
                  entry.failure?.kind === "client_txn_conflict" ? (
                    <Button
                      type="button"
                      tone="secondary"
                      disabled={owner.busy || !owner.canSubmit()}
                      onClick={() =>
                        owner.reviewNewRequestId(
                          checkpoint.id,
                          entry.attempt.clientTxnId,
                        )
                      }
                    >
                      Retry with a new request ID
                    </Button>
                  ) : null}
                </div>
              ))}
              {checkpoint.resolution ? (
                <details>
                  <summary>Collection review receipt</summary>
                  <p>
                    Record: {checkpoint.resolution.receipt.data.row.record_id}.
                    Version:{" "}
                    {checkpoint.resolution.receipt.data.row.row_version}.
                    Request: {checkpoint.resolution.receipt.meta.request_id}.
                  </p>
                </details>
              ) : null}
              {checkpoint.create.receipt && !owner.linkComplete(checkpoint) ? (
                <div>
                  <Button
                    type="button"
                    tone="secondary"
                    disabled={!owner.canSubmit()}
                    aria-disabled={owner.busy || !owner.canSubmit()}
                    onClick={() => {
                      if (!owner.busy) void owner.reviewLink(checkpoint.id);
                    }}
                  >
                    Review original Timeline link
                  </Button>
                  {checkpoint.review ? (
                    <Button
                      type="button"
                      tone="primary"
                      disabled={owner.busy || !owner.canSubmit()}
                      onClick={() => void owner.link(checkpoint.id)}
                    >
                      Link created Evidence
                    </Button>
                  ) : null}
                </div>
              ) : null}
              {checkpoint.create.receipt ? (
                <Button
                  type="button"
                  tone="secondary"
                  disabled={
                    state.preparing ||
                    [checkpoint.create, ...checkpoint.links].some(
                      (entry) => entry.refresh === "refreshing",
                    )
                  }
                  onClick={() => void owner.retryRefresh(checkpoint.id)}
                >
                  Refresh Evidence result
                </Button>
              ) : null}
            </section>
          ))}
        </section>
      ) : null}
    </details>
  );
}
