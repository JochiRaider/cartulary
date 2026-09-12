import { useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import { ContextualCreateForm } from "./ContextualCreateForm";
import type { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";

/** Compact workbook attachment; opening it preserves navigation and query state. */
export function ContextualCreateRecovery({
  owner,
}: {
  readonly owner: WorkbookContextualTaskDecisionCreateOwner;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const token = useRef(Symbol("contextual-create-recovery")).current;
  const details = useRef<HTMLDetailsElement>(null),
    trigger = useRef<HTMLElement>(null);
  const [open, setOpen] = useState(false);
  if (!snapshot.authority || (!snapshot.draft && !snapshot.entries.length))
    return null;
  return (
    <details
      ref={details}
      onToggle={(event) => {
        const opened = event.currentTarget.open;
        setOpen(opened);
        if (!opened) owner.detach(token);
      }}
      style={{ minWidth: 0 }}
    >
      <summary ref={trigger}>
        Task / Decision creation
        {owner.blockedCount
          ? ` (${owner.blockedCount} need recovery)`
          : snapshot.draft
            ? " (draft retained)"
            : ""}
      </summary>
      {open ? (
        <section
          tabIndex={-1}
          aria-label="Retained contextual creation"
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
          {snapshot.draft && snapshot.attachment !== token ? (
            <div>
              <p>
                {snapshot.draft.target.title} draft retained from{" "}
                {snapshot.draft.presentation.surfaceLabel}.
              </p>
              <WorkbookInspectorActionButton
                tone="secondary"
                type="button"
                onClick={() => owner.resume(token)}
              >
                Resume contextual draft
              </WorkbookInspectorActionButton>{" "}
              <WorkbookInspectorActionButton
                tone="secondary"
                type="button"
                disabled={owner.busy}
                onClick={() => owner.discard()}
              >
                Discard draft
              </WorkbookInspectorActionButton>
            </div>
          ) : null}
          <ContextualCreateForm
            owner={owner}
            attachment={token}
            disabled={owner.busy}
            onSubmit={() => void owner.submit(token)}
          />
          <ul>
            {snapshot.entries.map((entry) => (
              <li key={entry.attempt.clientTxnId}>
                <p>
                  {entry.receipt
                    ? `${entry.attempt.review.draft.target.title} created.`
                    : entry.phase === "uncertain"
                      ? "Creation outcome unconfirmed."
                      : entry.phase === "rejected"
                        ? "Creation rejected; the draft is retained."
                        : "Creating…"}{" "}
                  {entry.message}
                </p>
                {entry.receipt ? (
                  <details>
                    <summary>Creation receipt</summary>
                    <p>
                      Record: {entry.receipt.data.row.record_id}. Version:{" "}
                      {entry.receipt.data.row.row_version}. Change set:{" "}
                      {entry.receipt.data.change_set_id}. Request:{" "}
                      {entry.receipt.meta.request_id}.
                    </p>
                  </details>
                ) : null}
                {entry.phase === "uncertain" ? (
                  <WorkbookInspectorActionButton
                    tone="secondary"
                    type="button"
                    disabled={
                      entry.transportPending ||
                      snapshot.preparing ||
                      !owner.canReplay()
                    }
                    onClick={() => void owner.replay(entry.attempt.clientTxnId)}
                  >
                    Recover original creation
                  </WorkbookInspectorActionButton>
                ) : null}
                {entry.receipt && entry.refresh !== "complete" ? (
                  <WorkbookInspectorActionButton
                    tone="secondary"
                    type="button"
                    disabled={entry.refresh === "refreshing"}
                    onClick={() =>
                      void owner.retryRefresh(entry.attempt.clientTxnId)
                    }
                  >
                    Retry creation refresh
                  </WorkbookInspectorActionButton>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
      <span role="status" style={visuallyHiddenStyle}>
        {snapshot.message}
      </span>
    </details>
  );
}
