import {
  type RefObject,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookWorkAreaOverlay } from "../../layout/WorkbookWorkAreaOverlay";
import { NoteCreateForm } from "./NoteCreateForm";
import type { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";

export function NoteCreateRecovery({
  owner,
  fallbackFocusRef,
}: {
  readonly owner: WorkbookNoteCreateOwner;
  readonly fallbackFocusRef?: RefObject<HTMLElement | null>;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const attachment = useRef(Symbol("note-recovery")).current;
  const details = useRef<HTMLDetailsElement>(null);
  const summary = useRef<HTMLElement>(null);
  const focused = useRef(false);
  const [open, setOpen] = useState(false);
  const panel = useRef<HTMLElement>(null);
  const visible =
    !!state.authority &&
    (!!state.draft ||
      state.entries.some(
        (entry) => entry.phase !== "rejected" && entry.refresh !== "complete",
      ));
  useLayoutEffect(() => {
    if (!visible) setOpen(false);
    if (!visible && focused.current && state.authority) {
      focused.current = false;
      fallbackFocusRef?.current?.focus({ preventScroll: true });
    }
  }, [visible, state.authority, fallbackFocusRef]);
  if (!visible) return null;
  return (
    <details
      ref={details}
      onFocusCapture={() => {
        focused.current = true;
      }}
      onBlurCapture={(event) => {
        if (
          event.relatedTarget &&
          !event.currentTarget.contains(event.relatedTarget) &&
          !panel.current?.contains(event.relatedTarget)
        )
          focused.current = false;
      }}
      style={{ position: "relative" }}
      onToggle={(event) => {
        setOpen(event.currentTarget.open);
        if (!event.currentTarget.open) owner.detach(attachment);
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          if (details.current) details.current.open = false;
          setOpen(false);
          owner.detach(attachment);
          summary.current?.focus();
        }
      }}
    >
      <summary ref={summary}>
        {state.entries.some((entry) => entry.phase === "uncertain")
          ? "Note recovery"
          : state.draft
            ? "Note draft"
            : "Note refresh"}
      </summary>
      {open ? (
        <WorkbookWorkAreaOverlay ref={panel} label="Retained Note authoring">
          {state.attachment === attachment ? (
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
        </WorkbookWorkAreaOverlay>
      ) : null}
    </details>
  );
}
