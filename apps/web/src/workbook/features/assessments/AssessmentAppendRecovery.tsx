import { useRef, useState, useSyncExternalStore } from "react";
import { secondaryButtonStyle } from "../../components/workbookGridControlStyles";
import type { WorkbookAssessmentAuthoringOwner } from "./WorkbookAssessmentAuthoringOwner";

/** Retained results remain keyboard reachable after leaving the Assessment surface. */
export function AssessmentAppendRecovery({
  owner,
}: {
  readonly owner: WorkbookAssessmentAuthoringOwner;
}) {
  const disclosure = useRef<HTMLDetailsElement>(null);
  const trigger = useRef<HTMLElement>(null);
  const [open, setOpen] = useState(false);
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  if (!snapshot.authority || !snapshot.entries.length) return null;
  const recoverable = snapshot.entries.filter(
    (entry) =>
      entry.phase === "uncertain" ||
      (entry.receipt && entry.refresh !== "complete"),
  );
  return (
    <details
      ref={disclosure}
      onToggle={(event) => setOpen(event.currentTarget.open)}
      style={{ position: "relative", minWidth: 0 }}
    >
      <summary ref={trigger}>
        Assessment appends
        {recoverable.length ? ` (${recoverable.length} need recovery)` : ""}
      </summary>
      {open ? (
        <section
          tabIndex={-1}
          aria-label="Retained Assessment appends"
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
            insetInlineEnd: "var(--ct-spacing-md)",
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
          <p>
            Closing the inspector retains dispatched appends and their results
            for this incident session.
          </p>
          <ul>
            {snapshot.entries.map((entry) => (
              <li key={entry.attempt.clientTxnId}>
                <p>
                  {entry.receipt
                    ? "Assessment created."
                    : entry.phase === "uncertain"
                      ? "Append result unconfirmed."
                      : entry.phase === "rejected"
                        ? "Append rejected; the editable draft is retained."
                        : "Appending assessment…"}{" "}
                  {entry.message}
                </p>
                {entry.receipt ? (
                  <p>
                    Record: {entry.receipt.data.row.record_id}. Version:{" "}
                    {entry.receipt.data.row.row_version}. Change set:{" "}
                    {entry.receipt.data.change_set_id}.
                  </p>
                ) : null}
                {entry.phase === "uncertain" ? (
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    disabled={
                      entry.transportPending ||
                      snapshot.preparing ||
                      !owner.canReplay()
                    }
                    onClick={() => void owner.replay(entry.attempt.clientTxnId)}
                  >
                    Recover assessment append
                  </button>
                ) : null}
                {entry.receipt && entry.refresh !== "complete" ? (
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    disabled={entry.refresh === "refreshing"}
                    onClick={() =>
                      void owner.retryRefresh(entry.attempt.clientTxnId)
                    }
                  >
                    Retry assessment refresh
                  </button>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </details>
  );
}
