import { decisionSupersessionTestId } from "@cartulary/ui-contracts";
import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { DecisionSupersessionOperation } from "./decisionSupersessionOperation";

export function WorkbookDecisionSupersessionRecovery({
  runtime,
}: {
  readonly runtime: WorkbookMutationRuntime;
}) {
  const owner = runtime.decisionSupersession;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const summary = useRef<HTMLElement>(null);
  const focusRequested = useRef(false);
  useLayoutEffect(() => {
    if (open && focusRequested.current) {
      focusRequested.current = false;
      summary.current?.focus({ preventScroll: true });
    }
  }, [open]);
  useEffect(() => {
    if (!snapshot.authority) setOpen(false);
  }, [snapshot.authority]);
  if (!snapshot.entries.length) return null;
  const close = () => {
    const restore = summary.current?.parentElement?.contains(
      document.activeElement,
    );
    setOpen(false);
    if (restore) trigger.current?.focus({ preventScroll: true });
  };
  const acknowledged = snapshot.entries.filter(
    (entry) => entry.receipt !== null,
  ).length;
  const unknown = snapshot.entries.filter(
    (entry) => entry.phase === "uncertain",
  ).length;
  const refreshRequired = snapshot.entries.some(
    (entry) => entry.receipt && entry.reconciliation !== "complete",
  );
  const pending = snapshot.entries.filter(
    (entry) => entry.phase === "preparing" || entry.phase === "submitting",
  ).length;
  const rejected = snapshot.entries.filter(
    (entry) => entry.phase === "rejected",
  ).length;
  const status = [
    unknown
      ? `${unknown} supersession outcome${unknown === 1 ? "" : "s"} unknown.`
      : null,
    acknowledged
      ? `${acknowledged} supersession${acknowledged === 1 ? "" : "s"} completed.${refreshRequired ? " Refresh still required." : ""}`
      : null,
    pending
      ? `${pending} supersession${pending === 1 ? "" : "s"} in progress.`
      : null,
    rejected
      ? `${rejected} supersession${rejected === 1 ? "" : "s"} rejected.`
      : null,
  ]
    .filter(Boolean)
    .join(" ");
  return (
    <div style={{ position: "relative" }}>
      <WorkbookInspectorActionButton
        ref={trigger}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            focusRequested.current = true;
            setOpen(true);
          }
        }}
      >
        Decision actions ({snapshot.entries.length})
      </WorkbookInspectorActionButton>
      <span
        role="status"
        style={{
          marginInlineStart: "var(--ct-spacing-xs)",
          fontSize: "var(--ct-typography-compact-metadata-fontSize)",
        }}
      >
        {status}
      </span>
      {open ? (
        <section
          data-testid={decisionSupersessionTestId("recovery")}
          aria-label="Decision action recovery"
          style={{
            position: "absolute",
            zIndex: 30,
            insetInlineEnd: 0,
            inlineSize: "min(38rem, 90vw)",
            maxBlockSize: "75vh",
            overflow: "auto",
            overflowWrap: "anywhere",
            padding: "var(--ct-spacing-md)",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            boxShadow: "var(--ct-elevation-popover)",
          }}
          onKeyDown={(event) => {
            if (
              event.key === "Escape" &&
              !(
                event.target instanceof Element &&
                event.target.closest('[role="alertdialog"]')
              )
            ) {
              event.stopPropagation();
              close();
            }
          }}
        >
          <section
            ref={summary}
            tabIndex={-1}
            aria-label="Decision action recovery summary"
          >
            <strong>Decision actions</strong>
            <p>
              Admitted requests remain here when the inspector closes. A timeout
              or panel closure does not cancel server work.
            </p>
          </section>
          <WorkbookInspectorActionButton onClick={close}>
            Close Decision actions
          </WorkbookInspectorActionButton>
          {snapshot.entries.map((entry) => (
            <DecisionRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              runtime={runtime}
            />
          ))}
        </section>
      ) : null}
    </div>
  );
}

function DecisionRecoveryEntry({
  entry,
  runtime,
}: {
  readonly entry: DecisionSupersessionOperation;
  readonly runtime: WorkbookMutationRuntime;
}) {
  const owner = runtime.decisionSupersession;
  const { review } = entry.attempt;
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <article
      aria-label={`Supersession of ${review.target.label}`}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        paddingBlock: "var(--ct-spacing-md)",
        borderBlockStart: "var(--ct-border-hairline)",
      }}
    >
      <strong>Decision supersession</strong>
      <p>
        Target: {review.target.label} ({review.target.recordId})<br />
        Replacement: {review.replacement.label} ({review.replacement.recordId})
      </p>
      <p style={{ whiteSpace: "pre-wrap" }}>Reason: {review.reason}</p>
      <p role="status">
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Supersession accepted. Both Decisions and related projections refreshed."
            : entry.reconciliation === "refreshing"
              ? "Supersession accepted. Refresh in progress."
              : "Supersession accepted. Refresh is still required."
          : entry.phase === "preparing"
            ? "Waiting for earlier writes and checking the captured review."
            : entry.phase === "submitting"
              ? "Supersession submitted; waiting for acknowledgement."
              : entry.phase === "uncertain"
                ? "The outcome is unknown. The server may have committed. Exact replay uses the same transaction and request."
                : "Supersession rejected. Refresh and review again before a new attempt."}
      </p>
      {entry.failure ? <p>{entry.failure.message}</p> : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canSubmit()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay exact supersession request
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.transportPending && entry.phase === "uncertain" ? (
        <p>
          Transport has not settled. Closing this panel does not cancel a server
          commit.
        </p>
      ) : null}
      {entry.receipt ? (
        <>
          <p>
            Accepted change set: {entry.receipt.change_set_id}. Target status:{" "}
            {entry.receipt.target_status}.
          </p>
          <details>
            <summary>Decision receipt</summary>
            <p>
              View: {entry.receipt.view_schema_id}
              <br />
              Target: {entry.receipt.target_record_id}, version{" "}
              {entry.receipt.target_row_version}
              <br />
              Replacement: {entry.receipt.superseding_record_id}, version{" "}
              {entry.receipt.superseding_row_version}
            </p>
            <p style={{ whiteSpace: "pre-wrap" }}>
              Recorded reason: {entry.receipt.reason}
            </p>
          </details>
          {entry.reconciliation !== "complete" ? (
            <WorkbookInspectorActionButton
              disabled={pending || entry.reconciliation === "refreshing"}
              onClick={() => void owner.refresh(entry.attempt.id)}
            >
              Retry refresh
            </WorkbookInspectorActionButton>
          ) : null}
          <p>Review or reverse this change through existing record History.</p>
        </>
      ) : null}
      {!pending &&
      (entry.phase === "rejected" ||
        (entry.receipt && entry.reconciliation === "complete")) ? (
        <WorkbookInspectorActionButton
          onClick={() => owner.dismiss(entry.attempt.id)}
        >
          Dismiss Decision action
        </WorkbookInspectorActionButton>
      ) : null}
    </article>
  );
}
