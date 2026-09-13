import {
  type RefObject,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { CoordinationCreateForm } from "./CoordinationCreateForm";
import type { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";

export function CoordinationCreateRecovery({
  owner,
  fallbackFocusRef,
}: {
  readonly owner: WorkbookCoordinationCreateOwner;
  readonly fallbackFocusRef?: RefObject<HTMLElement | null>;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const attachment = useRef(Symbol("coordination-recovery")).current;
  const details = useRef<HTMLDetailsElement>(null);
  const summary = useRef<HTMLElement>(null);
  const focused = useRef(false);
  const visible =
    !!state.authority &&
    (!!state.draft ||
      state.entries.some(
        (entry) => entry.phase !== "rejected" && entry.refresh !== "complete",
      ));
  useLayoutEffect(() => {
    if (!visible && focused.current && state.authority) {
      focused.current = false;
      fallbackFocusRef?.current?.focus({ preventScroll: true });
    }
  }, [visible, state.authority, fallbackFocusRef]);
  if (!visible) return null;
  const actions = owner.captureDraftActions(state.attachment);
  return (
    <details
      ref={details}
      onFocusCapture={() => {
        focused.current = true;
      }}
      onBlurCapture={(event) => {
        if (
          event.relatedTarget &&
          !event.currentTarget.contains(event.relatedTarget)
        )
          focused.current = false;
      }}
      style={{ position: "relative", flexShrink: 0 }}
      onToggle={(event) => {
        if (!event.currentTarget.open) owner.detach(attachment);
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          if (details.current) details.current.open = false;
          owner.detach(attachment);
          summary.current?.focus();
        }
      }}
    >
      <summary ref={summary} style={{ whiteSpace: "nowrap" }}>
        <span
          style={{
            display: "inline-block",
            maxWidth: "11rem",
            overflow: "hidden",
            textOverflow: "ellipsis",
            verticalAlign: "middle",
          }}
        >
          {state.entries.some((entry) => entry.phase === "uncertain")
            ? "Coordination recovery"
            : state.draft
              ? "Coordination draft"
              : "Coordination refresh"}
        </span>
      </summary>
      <section
        aria-label="Retained Coordination authoring"
        style={{
          position: "fixed",
          right: "1rem",
          top: "4rem",
          zIndex: 30,
          width: "min(28rem, calc(100vw - 2rem))",
          maxHeight: "calc(100dvh - 5rem)",
          overflow: "auto",
          padding: "1rem",
          background: "var(--ct-colors-surface-1)",
          border: "var(--ct-border-hairline)",
          boxSizing: "border-box",
        }}
      >
        {state.attachment === attachment ? (
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
    </details>
  );
}
